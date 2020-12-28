// Package telnet provides a grab module that scans for telnet daemons.
// Default Port: 23 (TCP)
//
// The --max-read-size flag allows setting a ceiling to the number of bytes
// that will be read for the banner.
//
// The scan negotiates the options and attempts to grab the banner, using the
// same behavior as the original zgrab.
//
// The output contains the banner and the negotiated options, in the same
// format as the original zgrab.
package telnet

import (
	"github.com/mcuadros/go-defaults"
	"grab"
)

// Flags holds the command-line configuration for the Telnet scan module.
// Populated by the framework.
type Flags struct {
	grab.BaseFlags
	MaxReadSize int  `long:"max-read-size" description:"Set the maximum number of bytes to read when grabbing the banner" default:"65536"`
	Banner      bool `long:"force-banner" description:"Always return banner if it has non-zero bytes"`
	Verbose     bool `long:"verbose" description:"More verbose logging, include debug fields in the scan results"`
}

// Module implements the grab.Module interface.
type Module struct {
}

// Scanner implements the grab.Scanner interface.
type Scanner struct {
	config *Flags
}

// NewFlags returns a default Flags object.
func (module *Module) NewFlags() interface{} {
	flags := new(Flags)
	defaults.SetDefaults(flags)
	flags.BaseFlags.Name = "telnet"
	flags.BaseFlags.Port = 23
	return flags
}

// NewScanner returns a new Scanner instance.
func (module *Module) NewScanner() grab.Scanner {
	return new(Scanner)
}

// Description returns an overview of this module.
func (module *Module) Description() string {
	return "Fetch a telnet banner"
}

// Validate checks that the flags are valid.
// On success, returns nil.
// On failure, returns an error instance describing the error.
func (flags *Flags) Validate(args []string) error {
	return nil
}

// Help returns the module's help string.
func (flags *Flags) Help() string {
	return ""
}

// Init initializes the Scanner.
func (scanner *Scanner) Init(flags grab.ScanFlags) error {
	f, _ := flags.(*Flags)
	scanner.config = f
	return nil
}

// InitPerSender initializes the scanner for a given sender.
func (scanner *Scanner) InitPerSender(senderID int) error {
	return nil
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
	return "telnet"
}

// Scan connects to the target (default port TCP 23) and attempts to grab the Telnet banner.
func (scanner *Scanner) Scan(target grab.ScanTarget) (grab.ScanStatus, interface{}, error) {
	conn, err := target.Open(&scanner.config.BaseFlags)
	if err != nil {
		return grab.TryGetScanStatus(err), nil, err
	}
	defer conn.Close()
	result := new(TelnetLog)
	if err := GetTelnetBanner(result, conn, scanner.config.MaxReadSize); err != nil {
		if scanner.config.Banner && len(result.Banner) > 0 {
			return grab.TryGetScanStatus(err), result, err
		} else {
			return grab.TryGetScanStatus(err), result.getResult(), err
		}
	}
	return grab.SCAN_SUCCESS, result, nil
}
