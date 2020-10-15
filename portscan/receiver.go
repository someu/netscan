package main

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"io"
	"log"
)

type Receiver struct {
	handleMap map[string]*pcap.Handle
}

func NewReceiver() *Receiver {
	return &Receiver{
		handleMap: make(map[string]*pcap.Handle),
	}
}

func (receiver *Receiver) startReceive(name string, handle *pcap.Handle, ctx context.Context) {
	if receiver.handleMap[name] != nil {
		return
	} else {
		receiver.handleMap[name] = handle
	}

	go func() {
		defer handle.Close()
		ipv4 := &layers.IPv4{}
		tcp := &layers.TCP{}

		parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &layers.Ethernet{}, ipv4, tcp)
		for {
			select {
			case <-ctx.Done():
				log.Println("receive timeout")
				return
			}

			data, _, err := handle.ReadPacketData()
			if err == pcap.NextErrorTimeoutExpired || err == io.EOF {
				break
			} else if err != nil {
				log.Println("Read packet error", err.Error())
				continue
			}
			decodes := []gopacket.LayerType{}
			if err := parser.DecodeLayers(data, &decodes); err != nil {
				continue
			}
			for _, decode := range decodes {
				if decode == layers.LayerTypeTCP {
					cookie := generateCookie(ipv4.DstIP, uint16(tcp.DstPort), ipv4.SrcIP, uint16(tcp.SrcPort))
					if cookie == tcp.Ack-1 {
						if tcp.SYN && tcp.ACK {
							log.Printf("%s:%d is open", ipv4.SrcIP, tcp.SrcPort)
						}
					}
				}
			}
		}
	}()
}
