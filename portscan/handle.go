package portscan

import (
	"github.com/google/gopacket/pcap"
)

var handleMap = make(map[string]*pcap.Handle)

// 获取接口handle
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
