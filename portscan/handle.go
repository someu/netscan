package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/mostlygeek/arp"
	"github.com/phayes/freeport"
	"io"
	"log"
	"net"
	"time"
)

type Handle struct {
	serializeOptions gopacket.SerializeOptions
	handleMap        map[string]*pcap.Handle
	router           *Router
	ipMacMap         map[string]net.HardwareAddr
}

func NewHandle() (*Handle, error) {
	router, err := NewRouter()
	if err != nil {
		return nil, err
	}
	return &Handle{
		serializeOptions: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		handleMap: make(map[string]*pcap.Handle),
		ipMacMap:  map[string]net.HardwareAddr{},
		router:    router,
	}, nil
}

func (h *Handle) getHandle(device string) (*pcap.Handle, error) {
	if h.handleMap[device] == nil {
		var err error
		h.handleMap[device], err = pcap.OpenLive(device, 65535, true, pcap.BlockForever)
		if err != nil {
			return nil, err
		}
		go h.capture(h.handleMap[device])
	}
	return h.handleMap[device], nil
}

func (h *Handle) capture(handle *pcap.Handle) {
	ipv4 := &layers.IPv4{}
	tcp := &layers.TCP{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &layers.Ethernet{}, ipv4, tcp)
	for {
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
}

func (h *Handle) writePacketData(iface *NetInterface, data []byte) error {
	handle, err := h.getHandle(iface.name)
	if err != nil {
		return err
	}
	return handle.WritePacketData(data)
}

func (h *Handle) writePacketLayers(iface *NetInterface, layers ...gopacket.SerializableLayer) error {
	packet := gopacket.NewSerializeBuffer()

	err := gopacket.SerializeLayers(packet, h.serializeOptions, layers...)
	if err != nil {
		return err
	}
	return h.writePacketData(iface, packet.Bytes())
}

func (h *Handle) getMacAddr(dstIp net.IP, iface *NetInterface) net.HardwareAddr {
	dstIpStr := dstIp.String()
	gatewayStr := iface.gateway.String()
	if h.ipMacMap[dstIpStr] != nil {
		return h.ipMacMap[dstIpStr]
	}
	if h.ipMacMap[gatewayStr] != nil {
		return h.ipMacMap[gatewayStr]
	}

	macStr := arp.Search(dstIpStr)
	if macStr != "00:00:00:00:00:00" {
		mac, err := net.ParseMAC(macStr)
		if err == nil {
			return mac
		}
	}

	h.sendArpPacket(dstIp, iface)

	// Wait 3 seconds for an ARP reply.
	start := time.Now()
	for {
		if time.Since(start) > time.Second*3 {
			return net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		}
		data, _, err := iface.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(iface.gateway) {
				return arp.SourceHwAddress
			}
		}
	}

	return net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
}

func (h *Handle) sendArpPacket(dstIp net.IP, iface *NetInterface) error {
	if iface.gateway != nil {
		dstIp = iface.gateway
	}
	// broadcast arp packet
	eth := layers.Ethernet{
		SrcMAC:       iface.mac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.mac),
		SourceProtAddress: []byte(iface.ip),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(dstIp),
	}
	return h.writePacketLayers(iface, &eth, &arp)
}

func (h *Handle) sendSynPacket(dstIp net.IP, dstPort uint16) error {
	var err error
	srcPort, err := freeport.GetFreePort()
	if err != nil {
		return err
	}
	iface, err := h.router.routeIp(dstIp)
	if err != nil {
		return err
	}
	// get mac address
	dstMac := h.getMacAddr(dstIp, iface)
	if err != nil {
		return err
	}

	// send packet
	cookie := generateCookie(iface.ip, uint16(srcPort), dstIp, dstPort)
	ethernet := layers.Ethernet{
		SrcMAC:       iface.mac,
		DstMAC:       dstMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipv4 := layers.IPv4{
		SrcIP:    iface.ip,
		DstIP:    dstIp,
		Version:  4,
		TTL:      255,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		SYN:     true,
		Seq:     cookie,
	}
	tcp.SetNetworkLayerForChecksum(&ipv4)

	return h.writePacketLayers(iface, &ethernet, &ipv4, &tcp)
}
