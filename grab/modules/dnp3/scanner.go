// Package dnp3 provides a grab module that scans for dnp3.
// Default port: 20000 (TCP)
//
// Copied unmodified from the original zgrab.
// Connects, and reads the banner. Returns the raw response.
package dnp3

import (
	"github.com/mcuadros/go-defaults"
	"grab"
)

// Flags holds the command-line configuration for the dnp3 scan module.
// Populated by the framework.
type Flags struct {
	grab.BaseFlags
	// TODO: Support UDP?
	Verbose bool `long:"verbose" description:"More verbose logging, include debug fields in the scan results"`
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
	flags.BaseFlags.Name = "dnp3"
	flags.BaseFlags.Port = 20000
	return flags
}

// NewScanner returns a new Scanner instance.
func (module *Module) NewScanner() grab.Scanner {
	return new(Scanner)
}

// Description returns an overview of this module.
func (module *Module) Description() string {
	return "Probe for DNP3, a SCADA protocol"
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
	return "dnp3"
}

// Scan probes for a DNP3 service.
// Connects to the configured TCP port (default 20000) and reads the banner.
func (scanner *Scanner) Scan(target grab.ScanTarget) (grab.ScanStatus, interface{}, error) {
	// TODO: Allow UDP?
	conn, err := target.Open(&scanner.config.BaseFlags)
	if err != nil {
		return grab.TryGetScanStatus(err), nil, err
	}
	defer conn.Close()
	ret := new(DNP3Log)
	if err := GetDNP3Banner(ret, conn); err != nil {
		return grab.TryGetScanStatus(err), nil, err
	}
	return grab.SCAN_SUCCESS, ret, nil
}