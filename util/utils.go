package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var ipSplice = "([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"
var cidrMask = "([1-9]|[1-2]\\d|3[0-2])"
var cidrReString = fmt.Sprintf("^(%s\\.){3}%s\\/%s$", ipSplice, ipSplice, cidrMask)
var CidrRe = regexp.MustCompile(cidrReString)
var ipReString = fmt.Sprintf("^(%s\\.){3}%s$", ipSplice, ipSplice)
var IPRe = regexp.MustCompile(ipReString)

func ReadFileLines(filepath string) ([]string, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	reader := bufio.NewReader(file)
	for {
		if line, err := reader.ReadString('\n'); err != nil {
			break
		} else {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func Stringify(v interface{}) string {
	outputBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(outputBuffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	encoder.Encode(v)
	return outputBuffer.String()
}

func IsIP(ip string) bool {
	return IPRe.MatchString(ip)
}

func IsCIDR(cidr string) bool {
	return CidrRe.MatchString(cidr)
}

func CIDRToIpList(cidr string) ([]string, error) {
	var ips []string
	startIp, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ipLong := (uint(startIp[12]) << 24) + (uint(startIp[13]) << 16) + (uint(startIp[14]) << 8) + uint(startIp[15])
	slashIndex := strings.LastIndex(cidr, "/")
	mask, err := strconv.Atoi(cidr[slashIndex+1:])
	if err != nil {
		return nil, err
	}
	endIpLong := ipLong + (1 << uint(32-mask))
	for ; ipLong < endIpLong; ipLong++ {
		ip := net.IPv4(byte(ipLong>>24), byte(ipLong>>16), byte(ipLong>>8), byte(ipLong)).String()
		ips = append(ips, ip)
	}
	return ips, nil
}
