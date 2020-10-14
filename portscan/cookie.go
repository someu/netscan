package main

import (
	"github.com/dchest/siphash"
	"go.artemisc.eu/godium/random"
	"net"
	"strconv"
)

var k0 uint64
var k1 uint64

func init() {
	k0 = random.New().UInt64()
	k1 = random.New().UInt64()
}

func generateCookie(srcIp net.IP, srcPort uint16, dstIp net.IP, dstPort uint16) uint32 {
	var data []byte
	data = append(data, srcIp...)
	data = append(data, []byte(strconv.Itoa(int(srcPort)))...)
	data = append(data, dstIp...)
	data = append(data, []byte(strconv.Itoa(int(dstPort)))...)

	return uint32(siphash.Hash(k0, k1, data))
}
