// Package banner provides simple banner grab and matching implementation of the grab.Module.
// It sends a customizble probe (default to "\n") and filters the results based on custom regexp (--pattern)

package banner

import (
	"errors"
	"fmt"
	"github.com/mcuadros/go-defaults"
	"io"
	"net"
	"regexp"
	"strconv"

	"grab"
)

// Flags give the command-line flags for the banner module.
type Flags struct {
	grab.BaseFlags
	Probe    string `long:"probe" default:"\\n" description:"Probe to send to the server. Use triple slashes to escape, for example \\\\\\n is literal \\n" `
	Pattern  string `long:"pattern" description:"Pattern to match, must be valid regexp."`
	MaxTries int    `long:"max-tries" default:"1" description:"Number of tries for timeouts and connection errors before giving up."`
}

// Module is the implementation of the grab.Module interface.
type Module struct {
}

// Scanner is the implementation of the grab.Scanner interface.
type Scanner struct {
	config *Flags
	regex  *regexp.Regexp
	probe  []byte
}

type Results struct {
	Banner string `json:"banner,omitempty"`
	Length int    `json:"length,omitempty"`
}

// NewFlags returns a new default flags object.
func (m *Module) NewFlags() interface{} {
	flags := new(Flags)
	defaults.SetDefaults(flags)
	flags.BaseFlags.Name = "banner"
	flags.BaseFlags.Port = 80
	return flags
}

// GetName returns the Scanner name defined in the Flags.
func (scanner *Scanner) GetName() string {
	return scanner.config.Name
}

// GetTrigger returns the Trigger defined in the Flags.
func (scanner *Scanner) GetTrigger() string {
	return scanner.config.Trigger
}

// Protocol returns the protocol identifier of the scan.
func (scanner *Scanner) Protocol() string {
	return "banner"
}

// InitPerSender initializes the scanner for a given sender.
func (scanner *Scanner) InitPerSender(senderID int) error {
	return nil
}

// NewScanner returns a new Scanner object.
func (m *Module) NewScanner() grab.Scanner {
	return new(Scanner)
}

// Validate validates the flags and returns nil on success.
func (f *Flags) Validate(args []string) error {
	return nil
}

// Description returns an overview of this module.
func (module *Module) Description() string {
	return "Fetch a raw banner by sending a static probe and checking the result against a regular expression"
}

// Help returns the module's help string.
func (f *Flags) Help() string {
	return ""
}

// Init initializes the Scanner with the command-line flags.
func (scanner *Scanner) Init(flags grab.ScanFlags) error {
	f, _ := flags.(*Flags)
	scanner.config = f
	scanner.regex = regexp.MustCompile(scanner.config.Pattern)
	probe, err := strconv.Unquote(fmt.Sprintf(`"%s"`, scanner.config.Probe))
	if err != nil {
		panic("Probe error")
	}
	scanner.probe = []byte(probe)
	return nil
}

var NoMatchError = errors.New("pattern did not match")

func (scanner *Scanner) Scan(target grab.ScanTarget) (grab.ScanStatus, interface{}, error) {
	try := 0
	var (
		conn    net.Conn
		err     error
		readerr error
	)
	for try < scanner.config.MaxTries {
		try += 1
		conn, err = target.Open(&scanner.config.BaseFlags)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return grab.TryGetScanStatus(err), nil, err
	}
	defer conn.Close()

	var ret []byte
	try = 0
	for try < scanner.config.MaxTries {
		try += 1
		_, err = conn.Write(scanner.probe)
		ret, readerr = grab.ReadAvailable(conn)
		if err != nil {
			continue
		}
		if readerr != io.EOF && readerr != nil {
			continue
		}
		break
	}
	if err != nil {
		return grab.TryGetScanStatus(err), nil, err
	}
	if readerr != io.EOF && readerr != nil {
		return grab.TryGetScanStatus(readerr), nil, readerr
	}
	results := Results{Banner: string(ret), Length: len(ret)}
	if scanner.regex.Match(ret) {
		return grab.SCAN_SUCCESS, &results, nil
	}

	return grab.SCAN_PROTOCOL_ERROR, &results, NoMatchError

}
