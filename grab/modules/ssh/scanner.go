package ssh

import (
	"github.com/mcuadros/go-defaults"
	log "github.com/sirupsen/logrus"
	"grab"
	"grab/lib/ssh"
	"net"
	"strconv"
	"strings"
)

type SSHFlags struct {
	grab.BaseFlags
	ClientID          string `long:"client" description:"Specify the client ID string to use" default:"SSH-2.0-Go"`
	KexAlgorithms     string `long:"kex-algorithms" description:"Set SSH Key Exchange Algorithms"`
	HostKeyAlgorithms string `long:"host-key-algorithms" description:"Set SSH Host Key Algorithms"`
	Ciphers           string `long:"ciphers" description:"A comma-separated list of which ciphers to offer."`
	CollectUserAuth   bool   `long:"userauth" description:"Use the 'none' authentication request to see what userauth methods are allowed"`
	GexMinBits        uint   `long:"gex-min-bits" description:"The minimum number of bits for the DH GEX prime." default:"1024"`
	GexMaxBits        uint   `long:"gex-max-bits" description:"The maximum number of bits for the DH GEX prime." default:"8192"`
	GexPreferredBits  uint   `long:"gex-preferred-bits" description:"The preferred number of bits for the DH GEX prime." default:"2048"`
	HelloOnly         bool   `long:"hello-only" description:"Limit scan to the initial hello message"`
	Verbose           bool   `long:"verbose" description:"Output additional information, including SSH client properties from the SSH handshake."`
}

type Module struct {
}

type SSHScanner struct {
	config *SSHFlags
}

func (m *Module) NewFlags() interface{} {
	flags := new(SSHFlags)
	s := ssh.MakeSSHConfig()

	defaults.SetDefaults(flags)
	flags.BaseFlags.Port = 22
	flags.BaseFlags.Name = "ssh"
	flags.HostKeyAlgorithms = strings.Join(s.HostKeyAlgorithms, ",")
	flags.KexAlgorithms = strings.Join(s.KeyExchanges, ",")
	flags.Ciphers = strings.Join(s.Ciphers, ",")
	return flags
}

func (m *Module) NewScanner() grab.Scanner {
	return new(SSHScanner)
}

// Description returns an overview of this module.
func (m *Module) Description() string {
	return "Fetch an SSH server banner and collect key exchange information"
}

func (f *SSHFlags) Validate(args []string) error {
	return nil
}

func (f *SSHFlags) Help() string {
	return ""
}

func (s *SSHScanner) Init(flags grab.ScanFlags) error {
	f, _ := flags.(*SSHFlags)
	s.config = f
	return nil
}

func (s *SSHScanner) InitPerSender(senderID int) error {
	return nil
}

func (s *SSHScanner) GetName() string {
	return s.config.Name
}

func (s *SSHScanner) GetTrigger() string {
	return s.config.Trigger
}

func (s *SSHScanner) Scan(t grab.ScanTarget) (grab.ScanStatus, interface{}, error) {
	data := new(ssh.HandshakeLog)

	var port uint
	// If the port is supplied in ScanTarget, let that override the cmdline option
	if t.Port != nil {
		port = *t.Port
	} else {
		port = s.config.Port
	}
	portStr := strconv.FormatUint(uint64(port), 10)
	rhost := net.JoinHostPort(t.Host(), portStr)

	sshConfig := ssh.MakeSSHConfig()
	sshConfig.Timeout = s.config.Timeout
	sshConfig.ConnLog = data
	sshConfig.ClientVersion = s.config.ClientID
	sshConfig.HelloOnly = s.config.HelloOnly
	if err := sshConfig.SetHostKeyAlgorithms(s.config.HostKeyAlgorithms); err != nil {
		log.Fatal(err)
	}
	if err := sshConfig.SetKexAlgorithms(s.config.KexAlgorithms); err != nil {
		log.Fatal(err)
	}
	if err := sshConfig.SetCiphers(s.config.Ciphers); err != nil {
		log.Fatal(err)
	}
	sshConfig.Verbose = s.config.Verbose
	sshConfig.DontAuthenticate = s.config.CollectUserAuth
	sshConfig.GexMinBits = s.config.GexMinBits
	sshConfig.GexMaxBits = s.config.GexMaxBits
	sshConfig.GexPreferredBits = s.config.GexPreferredBits
	sshConfig.BannerCallback = func(banner string) error {
		data.Banner = strings.TrimSpace(banner)
		return nil
	}
	_, err := ssh.Dial("tcp", rhost, sshConfig)
	// TODO FIXME: Distinguish error types
	status := grab.TryGetScanStatus(err)
	return status, data, err
}

// Protocol returns the protocol identifer for the scanner.
func (s *SSHScanner) Protocol() string {
	return "ssh"
}
