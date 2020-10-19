package portscan

import (
	"encoding/binary"
	"errors"
	"github.com/dchest/siphash"
	"github.com/google/gopacket/pcap"
	"github.com/mostlygeek/arp"
	"go.artemisc.eu/godium/random"
	"net"
	"strconv"
)

var (
	k0        = random.New().UInt64()
	k1        = random.New().UInt64()
	handleMap = make(map[string]*pcap.Handle)
)

func generateCookie(srcIp net.IP, srcPort uint16, dstIp net.IP, dstPort uint16) uint32 {
	var data []byte
	data = append(data, srcIp...)
	data = append(data, []byte(strconv.Itoa(int(srcPort)))...)
	data = append(data, dstIp...)
	data = append(data, []byte(strconv.Itoa(int(dstPort)))...)

	return uint32(siphash.Hash(k0, k1, data))
}

func searchIPMacInArpTable(ip net.IP) (net.HardwareAddr, error) {
	macStr := arp.Search(ip.String())
	if macStr != "00:00:00:00:00:00" {
		return net.ParseMAC(macStr)
	}
	return nil, errors.New("not matched any mac address")
}

func getHandle(device string) (*pcap.Handle, error) {
	if handleMap[device] == nil {
		var err error
		handleMap[device], err = pcap.OpenLive(device, 65535, true, pcap.BlockForever)
		if err != nil {
			return nil, err
		}

	}
	return handleMap[device], nil
}

func longToIP(ipLong uint32) net.IP {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	return ipByte
}

func ipToLong(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}
