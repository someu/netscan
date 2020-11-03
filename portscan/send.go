package portscan

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/mostlygeek/arp"
	"github.com/phayes/freeport"
	"net"
	"time"
)

var router routing.Router
var hardwareAddrCache = make(map[string]net.HardwareAddr)

func init() {
	router, _ = routing.New()
}

type TargetRoute struct {
	srcIp               net.IP
	srcPort             uint16
	srcHardwareAddr     net.HardwareAddr
	dstIp               net.IP
	dstPort             uint16
	dstHardwareAddr     net.HardwareAddr
	gateway             net.IP
	gatewayHardwareAddr net.HardwareAddr
	iface               *net.Interface
	handle              *pcap.Handle
}

func resolveIPHardwareAddr(ip net.IP) (net.HardwareAddr, error) {
	macStr := arp.Search(ip.String())
	if macStr != "00:00:00:00:00:00" {
		return net.ParseMAC(macStr)
	}
	return nil, errors.New("not matched any mac address")
}

func (route *TargetRoute) resolveDstHardwareAddr(timeout time.Duration) error {
	// fix lo interface
	if route.iface.HardwareAddr.String() == "00:00:00:00:00:00" {
		route.dstHardwareAddr = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		return nil
	}
	route.dstHardwareAddr, _ = resolveIPHardwareAddr(route.dstIp)
	gatewayStr := route.gateway.String()
	if route.gateway != nil && hardwareAddrCache[gatewayStr] != nil {
		route.gatewayHardwareAddr, _ = hardwareAddrCache[gatewayStr]
	} else {
		route.gatewayHardwareAddr, _ = resolveIPHardwareAddr(route.gateway)
	}

	if route.dstHardwareAddr != nil || route.gatewayHardwareAddr != nil {
		return nil
	}

	// send arp packet when dst arp not in arp table
	arpDstErr := route.sendArpPacket(route.dstIp)
	arpGatewayErr := route.sendArpPacket(route.gateway)

	if arpDstErr != nil && arpGatewayErr != nil {
		return errors.New(fmt.Sprintf("send arp packet error %s, %s", arpDstErr, arpGatewayErr))
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
				route.gatewayHardwareAddr = arp.SourceHwAddress
				hardwareAddrCache[gatewayStr] = route.gatewayHardwareAddr
				return nil
			} else if arpSrcIp.Equal(route.srcIp) {
				route.dstHardwareAddr = arp.SourceHwAddress
				return nil
			}
		}
	}

	return errors.New("resolve dstIp's hardware address timeout")
}

func (route *TargetRoute) sendArpPacket(dstIP net.IP) error {
	// broadcast arp packet
	ethernet := layers.Ethernet{
		SrcMAC:       route.srcHardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(route.srcHardwareAddr),
		SourceProtAddress: []byte(route.srcIp),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(dstIP),
	}
	return route.writePacketLayers(&ethernet, &arp)
}

func (route *TargetRoute) sendSynPacket() error {
	// send packet
	cookie, err := generateCookie(route.srcIp, route.srcPort, route.dstIp, route.dstPort)
	if err != nil {
		return err
	}

	if route.dstHardwareAddr == nil {
		route.dstHardwareAddr = route.gatewayHardwareAddr
	}

	ethernet := layers.Ethernet{
		SrcMAC:       route.srcHardwareAddr,
		DstMAC:       route.dstHardwareAddr,
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

// 往 ip:port 发送syn包
func send(dstIp net.IP, dstPort uint16) error {
	if router == nil {
		return errors.New("route is nil")
	}

	iface, gateway, srcIp, err := router.Route(dstIp)
	if err != nil {
		return err
	}

	handle, err := getHandle(iface.Name)
	if err != nil {
		return err
	}

	srcPort, err := freeport.GetFreePort()
	if err != nil {
		return err
	}

	// TODO lo network scan
	if iface.HardwareAddr == nil {
		iface.HardwareAddr = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}

	route := &TargetRoute{
		srcIp:           srcIp,
		srcPort:         uint16(srcPort),
		srcHardwareAddr: iface.HardwareAddr,
		dstIp:           dstIp,
		dstPort:         dstPort,
		gateway:         gateway,
		iface:           iface,
		handle:          handle,
	}

	if err := route.resolveDstHardwareAddr(time.Second * 5); err != nil {
		return err
	}

	if err := route.sendSynPacket(); err != nil {
		return err
	}

	return nil
}
