// Package ftp contains the grab Module implementation for FTP(S).
//
// Setting the --authtls flag will cause the scanner to attempt a upgrade the
// connection to TLS. Settings for the TLS handshake / probe can be set with
// the standard TLSFlags.
//
// The scan performs a banner grab and (optionally) a TLS handshake.
//
// The output is the banner, any responses to the AUTH TLS/AUTH SSL commands,
// and any TLS logs.
package ftp

import (
	"fmt"
	"github.com/mcuadros/go-defaults"
	"net"
	"regexp"
	"strings"

	"grab"
)

// ScanResults is the output of the scan.
// Identical to the original from zgrab, with the addition of TLSLog.
type ScanResults struct {
	// Banner is the initial data banner sent by the server.
	Banner string `json:"banner,omitempty"`

	// AuthTLSResp is the response to the AUTH TLS command.
	// Only present if the FTPAuthTLS flag is set.
	AuthTLSResp string `json:"auth_tls,omitempty"`

	// AuthSSLResp is the response to the AUTH SSL command.
	// Only present if the FTPAuthTLS flag is set and AUTH TLS failed.
	AuthSSLResp string `json:"auth_ssl,omitempty"`

	// ImplicitTLS is true if the connection is wrapped in TLS, as opposed
	// to via AUTH TLS or AUTH SSL.
	ImplicitTLS bool `json:"implicit_tls,omitempty"`

	// TLSLog is the standard shared TLS handshake log.
	// Only present if the FTPAuthTLS flag is set.
	TLSLog *grab.TLSLog `json:"tls,omitempty"`
}

// Flags are the FTP-specific command-line flags. Taken from the original zgrab.
// (TODO: should FTPAuthTLS be on by default?).
type Flags struct {
	grab.BaseFlags
	grab.TLSFlags

	Verbose     bool `long:"verbose" description:"More verbose logging, include debug fields in the scan results"`
	FTPAuthTLS  bool `long:"authtls" description:"Collect FTPS certificates in addition to FTP banners"`
	ImplicitTLS bool `long:"implicit-tls" description:"Attempt to connect via a TLS wrapped connection"`
}

// Module implements the grab.Module interface.
type Module struct {
}

// Scanner implements the grab.Scanner interface, and holds the state
// for a single scan.
type Scanner struct {
	config *Flags
}

// Connection holds the state for a single connection to the FTP server.
type Connection struct {
	// buffer is a temporary buffer for sending commands -- so, never interleave
	// sendCommand calls on a given connection
	buffer  [10000]byte
	config  *Flags
	results ScanResults
	conn    net.Conn
}

// NewFlags returns the default flags object to be filled
func (m *Module) NewFlags() interface{} {
	flags := new(Flags)
	defaults.SetDefaults(flags)
	flags.BaseFlags.Name = "ftp"
	flags.BaseFlags.Port = 21
	return flags
}

// NewScanner returns a new Scanner instance.
func (m *Module) NewScanner() grab.Scanner {
	return new(Scanner)
}

// Description returns an overview of this module.
func (m *Module) Description() string {
	return "Grab an FTP banner"
}

// Validate flags
func (f *Flags) Validate(args []string) (err error) {
	if f.FTPAuthTLS && f.ImplicitTLS {
		err = fmt.Errorf("Cannot specify both '--authtls' and '--implicit-tls' together")
	}
	return
}

// Help returns this module's help string.
func (f *Flags) Help() string {
	return ""
}

// Protocol returns the protocol identifer for the scanner.
func (s *Scanner) Protocol() string {
	return "ftp"
}

// Init initializes the Scanner instance with the flags from the command
// line.
func (s *Scanner) Init(flags grab.ScanFlags) error {
	f, _ := flags.(*Flags)
	s.config = f
	return nil
}

// InitPerSender does nothing in this module.
func (s *Scanner) InitPerSender(senderID int) error {
	return nil
}

// GetName returns the configured name for the Scanner.
func (s *Scanner) GetName() string {
	return s.config.Name
}

// GetTrigger returns the Trigger defined in the Flags.
func (scanner *Scanner) GetTrigger() string {
	return scanner.config.Trigger
}

// ftpEndRegex matches zero or more lines followed by a numeric FTP status code
// and linebreak, e.g. "200 OK\r\n"
var ftpEndRegex = regexp.MustCompile(`^(?:.*\r?\n)*([0-9]{3})( [^\r\n]*)?\r?\n$`)

// isOKResponse returns true iff and only if the given response code indicates
// success (e.g. 2XX)
func (ftp *Connection) isOKResponse(retCode string) bool {
	// TODO: This is the current behavior; should it check that it isn't
	// garbage that happens to start with 2 (e.g. it's only ASCII chars, the
	// prefix is 2[0-9]+, etc)?
	return strings.HasPrefix(retCode, "2")
}

// readResponse reads an FTP response chunk from the server.
// It returns the full response, as well as the status code alone.
func (ftp *Connection) readResponse() (string, string, error) {
	respLen, err := grab.ReadUntilRegex(ftp.conn, ftp.buffer[:], ftpEndRegex)
	if err != nil {
		return "", "", err
	}
	ret := string(ftp.buffer[0:respLen])
	retCode := ftpEndRegex.FindStringSubmatch(ret)[1]
	return ret, retCode, nil
}

// GetFTPBanner reads the data sent by the server immediately after connecting.
// Returns true if and only if the server returns a success status code.
// Taken over from the original zgrab.
func (ftp *Connection) GetFTPBanner() (bool, error) {
	banner, retCode, err := ftp.readResponse()
	if err != nil {
		return false, err
	}
	ftp.results.Banner = banner
	return ftp.isOKResponse(retCode), nil
}

// sendCommand sends a command and waits for / reads / returns the response.
func (ftp *Connection) sendCommand(cmd string) (string, string, error) {
	ftp.conn.Write([]byte(cmd + "\r\n"))
	return ftp.readResponse()
}

// SetupFTPS returns true if and only if the server reported support for FTPS.
// First attempt AUTH TLS; if that fails, try AUTH SSL.
// Taken over from the original zgrab.
func (ftp *Connection) SetupFTPS() (bool, error) {
	ret, retCode, err := ftp.sendCommand("AUTH TLS")
	if err != nil {
		return false, err
	}
	ftp.results.AuthTLSResp = ret
	if ftp.isOKResponse(retCode) {
		return true, nil
	}
	ret, retCode, err = ftp.sendCommand("AUTH SSL")
	if err != nil {
		return false, err
	}
	ftp.results.AuthSSLResp = ret

	if ftp.isOKResponse(retCode) {
		return true, nil
	}
	return false, nil
}

// GetFTPSCertificates attempts to perform a TLS handshake with the server so
// that the TLS certificates will end up in the TLSLog.
// First sends the AUTH TLS/AUTH SSL command to tell the server we want to
// do a TLS handshake. If that fails, break. Otherwise, perform the handshake.
// Taken over from the original zgrab.
func (ftp *Connection) GetFTPSCertificates() error {
	ftpsReady, err := ftp.SetupFTPS()

	if err != nil {
		return err
	}
	if !ftpsReady {
		return nil
	}
	var conn *grab.TLSConnection
	if conn, err = ftp.config.TLSFlags.GetTLSConnection(ftp.conn); err != nil {
		return err
	}
	ftp.results.TLSLog = conn.GetLog()

	if err = conn.Handshake(); err != nil {
		// NOTE: With the default config of vsftp (without ssl_ciphers=HIGH),
		// AUTH TLS succeeds, but the handshake fails, dumping
		// "error:1408A0C1:SSL routines:ssl3_get_client_hello:no shared cipher"
		// to the socket.
		return err
	}
	ftp.conn = conn
	return nil
}

// Scan performs the configured scan on the FTP server, as follows:
// * Read the banner into results.Banner (if it is not a 2XX response, bail)
// * If the FTPAuthTLS flag is not set, finish.
// * Send the AUTH TLS command to the server. If the response is not 2XX, then
//   send the AUTH SSL command. If the response is not 2XX, then finish.
// * Perform ths TLS handshake / any configured TLS scans, populating
//   results.TLSLog.
// * Return SCAN_SUCCESS, &results, nil
func (s *Scanner) Scan(t grab.ScanTarget) (status grab.ScanStatus, result interface{}, thrown error) {
	var err error
	conn, err := t.Open(&s.config.BaseFlags)
	if err != nil {
		return grab.TryGetScanStatus(err), nil, err
	}
	cn := conn
	defer func() {
		cn.Close()
	}()

	results := ScanResults{}
	if s.config.ImplicitTLS {
		tlsConn, err := s.config.TLSFlags.GetTLSConnection(conn)
		if err != nil {
			return grab.TryGetScanStatus(err), nil, err
		}
		results.ImplicitTLS = true
		results.TLSLog = tlsConn.GetLog()
		tlsConn.Handshake()
		cn = tlsConn
	}

	ftp := Connection{conn: cn, config: s.config, results: results}
	is200Banner, err := ftp.GetFTPBanner()
	if err != nil {
		return grab.TryGetScanStatus(err), &ftp.results, err
	}
	if s.config.FTPAuthTLS && is200Banner {
		if err := ftp.GetFTPSCertificates(); err != nil {
			return grab.SCAN_APPLICATION_ERROR, &ftp.results, err
		}
	}
	return grab.SCAN_SUCCESS, &ftp.results, nil
}