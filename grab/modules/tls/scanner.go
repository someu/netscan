package tls

import (
	"github.com/mcuadros/go-defaults"
	"grab"
)

type TLSFlags struct {
	grab.BaseFlags
	grab.TLSFlags
}

type Module struct {
}

type TLSScanner struct {
	config *TLSFlags
}

func (m *Module) NewFlags() interface{} {
	flags := new(TLSFlags)
	defaults.SetDefaults(flags)
	flags.BaseFlags.Name = "tls"
	flags.BaseFlags.Port = 443
	return flags
}

func (m *Module) NewScanner() grab.Scanner {
	return new(TLSScanner)
}

// Description returns an overview of this module.
func (m *Module) Description() string {
	return "Perform a TLS handshake"
}

func (f *TLSFlags) Validate(args []string) error {
	return nil
}

func (f *TLSFlags) Help() string {
	return ""
}

func (s *TLSScanner) Init(flags grab.ScanFlags) error {
	f, ok := flags.(*TLSFlags)
	if !ok {
		return grab.ErrMismatchedFlags
	}
	s.config = f
	return nil
}

func (s *TLSScanner) GetName() string {
	return s.config.Name
}

func (s *TLSScanner) GetTrigger() string {
	return s.config.Trigger
}

func (s *TLSScanner) InitPerSender(senderID int) error {
	return nil
}

// Scan opens a TCP connection to the target (default port 443), then performs
// a TLS handshake. If the handshake gets past the ServerHello stage, the
// handshake log is returned (along with any other TLS-related logs, such as
// heartbleed, if enabled).
func (s *TLSScanner) Scan(t grab.ScanTarget) (grab.ScanStatus, interface{}, error) {
	conn, err := t.OpenTLS(&s.config.BaseFlags, &s.config.TLSFlags)
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		if conn != nil {
			if log := conn.GetLog(); log != nil {
				if log.HandshakeLog.ServerHello != nil {
					// If we got far enough to get a valid ServerHello, then
					// consider it to be a positive TLS detection.
					return grab.TryGetScanStatus(err), log, err
				}
				// Otherwise, detection failed.
			}
		}
		return grab.TryGetScanStatus(err), nil, err
	}
	return grab.SCAN_SUCCESS, conn.GetLog(), nil
}

// Protocol returns the protocol identifer for the scanner.
func (s *TLSScanner) Protocol() string {
	return "tls"
}