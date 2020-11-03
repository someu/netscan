package portscan

import (
	"encoding/binary"
	"net"
)

func longToIP(ipLong uint32) net.IP {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	return ipByte
}

func ipToLong(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}
