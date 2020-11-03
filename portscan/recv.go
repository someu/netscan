package portscan

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"io"
	"net"
)

var recvControl = make(map[string]*Control)

type Control struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

// 初始化接口，每个接口开启一个收包 goroutine
func initRecvInterfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, iface := range ifaces {
		if ctx, err := recv(iface.Name); err == nil {
			recvControl[iface.Name] = ctx
		} else {
			return err
		}
	}
	return nil
}

// 开启接口的收包 goroutine
func recv(name string) (*Control, error) {
	handle, err := getHandle(name)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer handle.Close()
		ipv4 := &layers.IPv4{}
		tcp := &layers.TCP{}
		parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &layers.Ethernet{}, ipv4, tcp)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				break
			}

			data, _, err := handle.ReadPacketData()
			if err == pcap.NextErrorTimeoutExpired || err == io.EOF {
				// TODO reopen handle ?
				break
			} else if err != nil {
				// TODO write error to scan
				continue
			}
			decodes := []gopacket.LayerType{}
			if err := parser.DecodeLayers(data, &decodes); err != nil {
				continue
			}
			for _, decode := range decodes {
				if decode == layers.LayerTypeTCP {
					cookie, err := generateCookie(ipv4.DstIP, uint16(tcp.DstPort), ipv4.SrcIP, uint16(tcp.SrcPort))
					if err != nil {
						continue
					}
					if cookie == tcp.Ack-1 {
						if tcp.SYN && tcp.ACK {
							if runningScan != nil {
								runningScan.pushResult(ipv4.SrcIP, uint16(tcp.SrcPort))
							}
						}
					}
				}
			}
		}
	}()
	return &Control{Ctx: ctx, Cancel: cancel}, nil
}
