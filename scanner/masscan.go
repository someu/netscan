package scanner

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

type MasScan struct {
	ProgramPath string
	Args        []string
	Ports       string
	Ranges      string
	Rate        string
	Exclude     string
}

func (scan *MasScan) SetProgramPath(program string) {
	scan.ProgramPath = program
}

func (scan *MasScan) SetPorts(ports string) {
	scan.Ports = ports
}

func (scan *MasScan) SetRanges(ranges string) {
	scan.Ranges = ranges
}

func (scan *MasScan) SetRate(rate string) {
	scan.Rate = rate
}

func (scan *MasScan) SetExclude(exclude string) {
	scan.Exclude = exclude
}

func (scan *MasScan) Run() ([]string, error) {
	if scan.Ranges != "" {
		scan.Args = append(scan.Args, "--range")
		scan.Args = append(scan.Args, scan.Ranges)
	}
	if scan.Ports != "" {
		scan.Args = append(scan.Args, "-p")
		scan.Args = append(scan.Args, scan.Ports)
	}
	if scan.Rate != "" {
		scan.Args = append(scan.Args, "--rate")
		scan.Args = append(scan.Args, scan.Rate)
	}
	if scan.Exclude != "" {
		scan.Args = append(scan.Args, "--exclude")
		scan.Args = append(scan.Args, scan.Exclude)
	}
	scan.Args = append(scan.Args, "-oL", "-")

	cmd := exec.Command(scan.ProgramPath, scan.Args...)
	log.Printf("Run masscan: %v\n", cmd.Args)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	error := cmd.Run()

	if error == nil {
		result := string(stdout.Bytes())
		return parseMassScanResult(result), nil
	} else {
		log.Printf("Run masscan failed: %s", error.Error())
		return nil, error
	}
}

func NewMasscan(ranges string, ports string) *MasScan {
	return &MasScan{
		ProgramPath: "masscan",
		Rate:        "5000",
		Ranges:      ranges,
		Ports:       ports,
	}
}

func parseMassScanResult(result string) []string {
	lines := strings.Split(result, "\n")
	var targets []string
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		items := strings.Split(line, " ")
		port, ip := items[2], items[3]
		if len(port) == 0 || len(ip) == 0 {
			continue
		}
		var target string
		switch port {
		case "443":
			target = "https://" + ip
		case "80":
			target = "http://" + ip
		default:
			target = "http://" + ip + ":" + port
		}
		targets = append(targets, target)
	}

	return targets
}
