package portscan

import (
	"errors"
	"github.com/dchest/siphash"
	"go.artemisc.eu/godium/random"
	"net"
	"strconv"
)

var (
	k0 uint64
	k1 uint64
)

func updateCookieKey() {
	k0 = random.New().UInt64()
	k1 = random.New().UInt64()
	if k0 == 0 || k1 == 0 {
		updateCookieKey()
	}
}

// 生成cookie
func generateCookie(srcIp net.IP, srcPort uint16, dstIp net.IP, dstPort uint16) (uint32, error) {
	if k0 == 0 || k1 == 0 {
		return 0, errors.New("invalid cookie key")
	}

	var data []byte
	data = append(data, srcIp...)
	data = append(data, []byte(strconv.Itoa(int(srcPort)))...)
	data = append(data, dstIp...)
	data = append(data, []byte(strconv.Itoa(int(dstPort)))...)

	return uint32(siphash.Hash(k0, k1, data)), nil
}
