package modules

import (
	"grab"
	"grab/modules/bacnet"
	"grab/modules/banner"
	"grab/modules/dnp3"
	"grab/modules/fox"
	"grab/modules/ftp"
	"grab/modules/http"
	"grab/modules/imap"
	"grab/modules/ipp"
	"grab/modules/modbus"
	"grab/modules/mongodb"
	"grab/modules/mssql"
	"grab/modules/mysql"
	"grab/modules/ntp"
	"grab/modules/oracle"
	"grab/modules/pop3"
	"grab/modules/postgres"
	"grab/modules/redis"
	"grab/modules/siemens"
	"grab/modules/smb"
	"grab/modules/smtp"
	"grab/modules/ssh"
	"grab/modules/telnet"
	"grab/modules/tls"
)

var defaultModules grab.ModuleSet

func init() {
	defaultModules = grab.ModuleSet{
		"bacnet":   &bacnet.Module{},
		"banner":   &banner.Module{},
		"dnp3":     &dnp3.Module{},
		"fox":      &fox.Module{},
		"ftp":      &ftp.Module{},
		"http":     &http.Module{},
		"imap":     &imap.Module{},
		"ipp":      &ipp.Module{},
		"modbus":   &modbus.Module{},
		"mongodb":  &mongodb.Module{},
		"mssql":    &mssql.Module{},
		"mysql":    &mysql.Module{},
		"ntp":      &ntp.Module{},
		"oracle":   &oracle.Module{},
		"pop3":     &pop3.Module{},
		"postgres": &postgres.Module{},
		"redis":    &redis.Module{},
		"siemens":  &siemens.Module{},
		"smb":      &smb.Module{},
		"smtp":     &smtp.Module{},
		"ssh":      &ssh.Module{},
		"telnet":   &telnet.Module{},
		"tls":      &tls.Module{},
	}
}

func NewModuleSetWithDefaults() grab.ModuleSet {
	out := grab.ModuleSet{}
	defaultModules.CopyInto(out)
	return out
}
