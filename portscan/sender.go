package portscan

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/phayes/freeport"
	"net"
	"time"
)

var arpCache map[string]net.HardwareAddr

func init() {
	arpCache = make(map[string]net.HardwareAddr)
}

type SenderConfig struct {
	arpTimeout time.Duration
}

type Sender struct {
	router routing.Router
	config *SenderConfig
}

func NewSender(config *SenderConfig) (*Sender, error) {
	router, err := routing.New()
	if err != nil {
		return nil, err
	}
	return &Sender{
		router: router,
		config: config,
	}, nil
}

func (sender *Sender) send(dstIp net.IP, dstPort uint16) (*TargetRoute, error) {
	iface, gateway, srcIp, err := sender.router.Route(dstIp)
	if err != nil {
		return nil, err
	}

	handle, err := getHandle(iface.Name)
	if err != nil {
		return nil, err
	}

	srcPort, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	// TODO lo network scan
	if iface.HardwareAddr == nil {
		iface.HardwareAddr = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}

	route := &TargetRoute{
		srcIp:   srcIp,
		srcPort: uint16(srcPort),
		srcMac:  iface.HardwareAddr,
		dstIp:   dstIp,
		dstPort: dstPort,
		gateway: gateway,
		iface:   iface,
		handle:  handle,
	}

	if err := route.arpMac(sender.config.arpTimeout); err != nil {
		return nil, err
	}

	if err := route.sendSynPacket(); err != nil {
		return nil, err
	}

	return route, nil
}

type TargetRoute struct {
	srcIp      net.IP
	srcPort    uint16
	srcMac     net.HardwareAddr
	dstIp      net.IP
	dstPort    uint16
	dstMac     net.HardwareAddr
	gateway    net.IP
	gatewayMac net.HardwareAddr
	iface      *net.Interface
	handle     *pcap.Handle
}

func (route *TargetRoute) arpMac(timeout time.Duration) error {
	// fix lo interface
	if route.iface.HardwareAddr.String() == "00:00:00:00:00:00" {
		route.dstMac = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		return nil
	}
	route.dstMac, _ = searchIPMacInArpTable(route.dstIp)
	gatewayStr := route.gateway.String()
	if route.gateway != nil && arpCache[gatewayStr] != nil {
		route.gatewayMac, _ = arpCache[gatewayStr]
	} else {
		route.gatewayMac, _ = searchIPMacInArpTable(route.gateway)
	}

	if route.dstMac != nil || route.gatewayMac != nil {
		return nil
	}

	// send arp packet when dst arp not in arp table
	arpDstErr := route.sendArpPacket(route.dstIp)
	arpGatewayErr := route.sendArpPacket(route.gateway)

	if arpDstErr != nil && arpGatewayErr != nil {
		return errors.New(fmt.Sprintf("end arp packet error %s, %s", arpDstErr, arpGatewayErr))
	}

	// wait arp packet
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			break
		}
		data, _, err := route.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			arpSrcIp := net.IP(arp.SourceProtAddress)
			if arpSrcIp.Equal(route.gateway) {
				route.gatewayMac = arp.SourceHwAddress
				arpCache[gatewayStr] = route.gatewayMac
				return nil
			} else if arpSrcIp.Equal(route.srcIp) {
				route.dstMac = arp.SourceHwAddress
				return nil
			}
		}
	}

	return errors.New("arp mac address timeout")
}

func (route *TargetRoute) sendArpPacket(dstIP net.IP) error {
	// broadcast arp packet
	ethernet := layers.Ethernet{
		SrcMAC:       route.srcMac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(route.srcMac),
		SourceProtAddress: []byte(route.srcIp),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(dstIP),
	}
	return route.writePacketLayers(&ethernet, &arp)
}

func (route *TargetRoute) sendSynPacket() error {
	// send packet
	cookie := generateCookie(route.srcIp, route.srcPort, route.dstIp, route.dstPort)

	if route.dstMac == nil {
		route.dstMac = route.gatewayMac
	}

	ethernet := layers.Ethernet{
		SrcMAC:       route.srcMac,
		DstMAC:       route.dstMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipv4 := layers.IPv4{
		SrcIP:    route.srcIp,
		DstIP:    route.dstIp,
		Version:  4,
		TTL:      255,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(route.srcPort),
		DstPort: layers.TCPPort(route.dstPort),
		SYN:     true,
		Seq:     cookie,
	}
	tcp.SetNetworkLayerForChecksum(&ipv4)

	return route.writePacketLayers(&ethernet, &ipv4, &tcp)
}

func (route *TargetRoute) writePacketLayers(layers ...gopacket.SerializableLayer) error {
	packet := gopacket.NewSerializeBuffer()

	err := gopacket.SerializeLayers(packet, gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}, layers...)
	if err != nil {
		return err
	}
	return route.handle.WritePacketData(packet.Bytes())
}
