package grab

import (
	"fmt"
	"net"
	log "github.com/sirupsen/logrus"
)

// ScanTarget is the host that will be scanned
type ScanTarget struct {
	IP     net.IP
	Domain string
	Tag    string
	Port   *uint
}

func (target ScanTarget) String() string {
	if target.IP == nil && target.Domain == "" {
		return "<empty target>"
	}
	res := ""
	if target.IP != nil && target.Domain != "" {
		res = target.Domain + "(" + target.IP.String() + ")"
	} else if target.IP != nil {
		res = target.IP.String()
	} else {
		res = target.Domain
	}
	if target.Tag != "" {
		res += " tag:" + target.Tag
	}
	return res
}

// Host gets the host identifier as a string: the IP address if it is available,
// or the domain if not.
func (target *ScanTarget) Host() string {
	if target.IP != nil {
		return target.IP.String()
	} else if target.Domain != "" {
		return target.Domain
	}
	log.Fatalf("Bad target %s: no IP/Domain", target.String())
	panic("unreachable")
}

// Open connects to the ScanTarget using the configured flags, and returns a net.Conn that uses the configured timeouts for Read/Write operations.
func (target *ScanTarget) Open(flags *BaseFlags) (net.Conn, error) {
	var port uint
	// If the port is supplied in ScanTarget, let that override the cmdline option
	if target.Port != nil {
		port = *target.Port
	} else {
		port = flags.Port
	}

	address := net.JoinHostPort(target.Host(), fmt.Sprintf("%d", port))
	return DialTimeoutConnection("tcp", address, flags.Timeout, flags.BytesReadLimit)
}

// OpenTLS connects to the ScanTarget using the configured flags, then performs
// the TLS handshake. On success error is nil, but the connection can be non-nil
// even if there is an error (this allows fetching the handshake log).
func (target *ScanTarget) OpenTLS(baseFlags *BaseFlags, tlsFlags *TLSFlags) (*TLSConnection, error) {
	conn, err := tlsFlags.Connect(target, baseFlags)
	if err != nil {
		return conn, err
	}
	err = conn.Handshake()
	return conn, err
}

// OpenUDP connects to the ScanTarget using the configured flags, and returns a net.Conn that uses the configured timeouts for Read/Write operations.
// Note that the UDP "connection" does not have an associated timeout.
func (target *ScanTarget) OpenUDP(flags *BaseFlags, udp *UDPFlags) (net.Conn, error) {
	var port uint
	// If the port is supplied in ScanTarget, let that override the cmdline option
	if target.Port != nil {
		port = *target.Port
	} else {
		port = flags.Port
	}
	address := net.JoinHostPort(target.Host(), fmt.Sprintf("%d", port))
	var local *net.UDPAddr
	if udp != nil && (udp.LocalAddress != "" || udp.LocalPort != 0) {
		local = &net.UDPAddr{}
		if udp.LocalAddress != "" && udp.LocalAddress != "*" {
			local.IP = net.ParseIP(udp.LocalAddress)
		}
		if udp.LocalPort != 0 {
			local.Port = int(udp.LocalPort)
		}
	}
	remote, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", local, remote)
	if err != nil {
		return nil, err
	}
	return NewTimeoutConnection(nil, conn, flags.Timeout, 0, 0, flags.BytesReadLimit), nil
}
