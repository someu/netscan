package grab

import (
	"errors"
	"net"
	"regexp"
	"strings"

	"time"

	"github.com/sirupsen/logrus"
	"runtime/debug"
)

// ReadAvaiable reads what it can without blocking for more than
// defaultReadTimeout per read, or defaultTotalTimeout for the whole session.
// Reads at most defaultMaxReadSize bytes.
func ReadAvailable(conn net.Conn) ([]byte, error) {
	const defaultReadTimeout = 10 * time.Millisecond
	const defaultMaxReadSize = 1024 * 512
	// if the buffer size exactly matches the number of bytes returned, we hit
	// a corner case where we attempt to read even though there is nothing
	// available. Otherwise we should be able to return without blocking at all.
	// So -- it's better to be large than small, but the worst case is getting
	// the exact right number of bytes.
	const defaultBufferSize = 8209

	return ReadAvailableWithOptions(conn, defaultBufferSize, defaultReadTimeout, 0, defaultMaxReadSize)
}

// ReadAvailableWithOptions reads whatever can be read (up to maxReadSize) from
// conn without blocking for longer than readTimeout per read, or totalTimeout
// for the entire session. A totalTimeout of 0 means attempt to use the
// connection's timeout (or, failing that, 1 second).
// On failure, returns anything it was able to read along with the error.
func ReadAvailableWithOptions(conn net.Conn, bufferSize int, readTimeout time.Duration, totalTimeout time.Duration, maxReadSize int) ([]byte, error) {
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	var totalDeadline time.Time
	if totalTimeout == 0 {
		// Would be nice if this could be taken from the SetReadDeadline(), but that's not possible in general
		const defaultTotalTimeout = 1 * time.Second
		totalTimeout = defaultTotalTimeout
		timeoutConn, isTimeoutConn := conn.(*TimeoutConnection)
		if isTimeoutConn {
			totalTimeout = timeoutConn.Timeout
		}
	}
	if totalTimeout > 0 {
		totalDeadline = time.Now().Add(totalTimeout)
	}

	buf := make([]byte, bufferSize)
	ret := make([]byte, 0)

	// The first read will use any pre-assigned deadlines.
	n, err := conn.Read(buf[0:min(bufferSize, maxReadSize)])
	ret = append(ret, buf[0:n]...)
	if err != nil || n >= maxReadSize {
		return ret, err
	}
	maxReadSize -= n

	// If there were more than bufSize -1 bytes available, read whatever is
	// available without blocking longer than timeout, and do not treat timeouts
	// as an error.
	// Keep reading until we time out or get an error.
	for totalDeadline.IsZero() || totalDeadline.After(time.Now()) {
		deadline := time.Now().Add(readTimeout)
		conn.SetReadDeadline(deadline)
		n, err := conn.Read(buf[0:min(maxReadSize, bufferSize)])
		maxReadSize -= n
		ret = append(ret, buf[0:n]...)
		if err != nil {
			if IsTimeoutError(err) {
				err = nil
			}
			return ret, err
		}
		if err != nil {
			return ret, err
		}
		if n >= maxReadSize {
			return ret, err
		}
	}
	return ret, ErrTotalTimeout
}

var InsufficientBufferError = errors.New("not enough buffer space")

// ReadUntilRegex calls connection.Read() until it returns an error, or the cumulatively-read data matches the given regexp
func ReadUntilRegex(connection net.Conn, res []byte, expr *regexp.Regexp) (int, error) {
	buf := res[0:]
	length := 0
	for finished := false; !finished; {
		n, err := connection.Read(buf)
		length += n
		if err != nil {
			return length, err
		}
		if expr.Match(res[0:length]) {
			finished = true
		}
		if length == len(res) {
			return length, InsufficientBufferError
		}
		buf = res[length:]
	}
	return length, nil
}

// TLDMatches checks for a strict TLD match
func TLDMatches(host1 string, host2 string) bool {
	splitStr1 := strings.Split(stripPortNumber(host1), ".")
	splitStr2 := strings.Split(stripPortNumber(host2), ".")

	tld1 := splitStr1[len(splitStr1)-1]
	tld2 := splitStr2[len(splitStr2)-1]

	return tld1 == tld2
}

func stripPortNumber(host string) string {
	return strings.Split(host, ":")[0]
}

type timeoutError interface {
	Timeout() bool
}

// IsTimeoutError checks if the given error corresponds to a timeout (of any type).
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if cast, ok := err.(timeoutError); ok {
		return cast.Timeout()
	}
	if cast, ok := err.(*ScanError); ok {
		return cast.Status == SCAN_IO_TIMEOUT || cast.Status == SCAN_CONNECTION_TIMEOUT
	}

	return false
}

// LogPanic is intended to be called from within defer -- if there was no panic, it returns without
// doing anything. Otherwise, it logs the stacktrace, the panic error, and the provided message
// before re-raising the original panic.
// Example:
//     defer zgrab2.LogPanic("Error decoding body '%x'", body)
func LogPanic(format string, args ...interface{}) {
	err := recover()
	if err == nil {
		return
	}
	logrus.Errorf("Uncaught panic at %s: %v", string(debug.Stack()), err)
	logrus.Errorf(format, args...)
	panic(err)
}
