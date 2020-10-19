package portscan

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"io"
	"log"
)

type Receiver struct {
	ctx       context.Context
	handleMap map[string]*pcap.Handle
}

func NewReceiver(ctx context.Context) *Receiver {
	return &Receiver{
		ctx:       ctx,
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
		count := 0
		parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &layers.Ethernet{}, ipv4, tcp)
		for {
			select {
			case <-ctx.Done():
				log.Println("receive exit")
				return
			default:
				break
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
							log.Printf("%d %s:%d is open", count, ipv4.SrcIP, tcp.SrcPort)
							count++
						}
					}
				}
			}
		}
	}()
}
