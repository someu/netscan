package main

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/mostlygeek/arp"
	"github.com/phayes/freeport"
	"io"
	"log"
	"net"
	"time"
)

type PortScanner struct {
	router           routing.Router
	srcPort          int
	serializeOptions gopacket.SerializeOptions
	handleMap        map[string]*pcap.Handle
}

// cache gateway's mac address
var gatewayMacMap = map[string]string{}

func NewPortScanner() (*PortScanner, error) {
	var err error
	var router routing.Router
	if router, err = routing.New(); err != nil {
		return nil, err
	}
	var sport int
	if sport, err = freeport.GetFreePort(); err != nil {
		return nil, err
	}
	return &PortScanner{
		router:  router,
		srcPort: sport,
		serializeOptions: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
	}, nil
}

func (s *PortScanner) getHandle(device string) (*pcap.Handle, error) {
	if s.handleMap[device] == nil {
		var err error
		s.handleMap[device], err = pcap.OpenLive(device, 65535, true, pcap.BlockForever)
		if err != nil {
			return nil, err
		}
	}
	return s.handleMap[device], nil
}

func (s *PortScanner) sendSynPacket(dstMac net.HardwareAddr, dstIp net.IP, dstPort layers.TCPPort) {

}

func sendArpPacket() {

}

func (s *PortScanner) Scan(ip net.IP, port int) error {
	var err error
	device, gateway, srcIp, err := s.router.Route(ip)
	if err != nil {
		return err
	}

	handle, err := pcap.OpenLive(device.Name, 65535, true, pcap.BlockForever)
	if err != nil {
		return nil
	}
	//defer handle.Close()

	// get mac address
	dstMac, err := s.getMac(ip, gateway, srcIp, device)
	if err != nil {
		return err
	}

	// send packet
	ethernet := layers.Ethernet{
		SrcMAC:       device.HardwareAddr,
		DstMAC:       dstMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipv4 := layers.IPv4{
		SrcIP:    srcIp,
		DstIP:    ip,
		Version:  4,
		TTL:      255,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(s.srcPort),
		DstPort: layers.TCPPort(port),
		SYN:     true,
	}
	tcp.SetNetworkLayerForChecksum(&ipv4)

	packet := gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(packet, s.serializeOptions, &ethernet, &ipv4, &tcp)
	if err != nil {
		return err
	}

	go func() {
		defer handle.Close()
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
				if decode == layers.LayerTypeTCP && ipv4.SrcIP.Equal(ip) && tcp.SrcPort == layers.TCPPort(port) {
					if tcp.SYN && tcp.ACK {
						log.Printf("%s:%d is open", ip, port)
					} else if tcp.RST {
						log.Printf("%s:%d is close", ip, port)
					}
					return
				}
			}
		}
	}()

	handle.WritePacketData(packet.Bytes())
	return nil
}

func (s *PortScanner) getMac(ip net.IP, gateway net.IP, srcIp net.IP, device *net.Interface) (net.HardwareAddr, error) {
	macStr := arp.Search(ip.String())
	macStr = "00:00:00:00:00:00"
	if macStr != "00:00:00:00:00:00" {
		if mac, err := net.ParseMAC(macStr); err == nil {
			return mac, nil
		}
	}

	arpDst := ip
	if gateway != nil {
		arpDst = gateway
	}

	handle, err := pcap.OpenLive(device.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	start := time.Now()

	// Prepare the layers to send for an ARP request.
	eth := layers.Ethernet{
		SrcMAC:       device.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(device.HardwareAddr),
		SourceProtAddress: []byte(srcIp),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(arpDst),
	}

	packet := gopacket.NewSerializeBuffer()

	if err := gopacket.SerializeLayers(packet, s.serializeOptions, &eth, &arp); err != nil {
		return nil, err
	}
	if err := handle.WritePacketData(packet.Bytes()); err != nil {
		return nil, err
	}

	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*3 {
			return nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(arpDst) {
				return arp.SourceHwAddress, nil
			}
		}
	}
}

func main() {
	scanner, err := NewPortScanner()
	if err != nil {
		log.Fatalln("Create scanner error", err.Error())
	}
	for i := 0; i < 5; i++ {
		err = scanner.Scan(net.IP{10, 0, 8, 94}, 8080)
		if err != nil {
			log.Fatalln("Scan error", err.Error())
		} else {
			log.Println("Success scan")
		}
		err = scanner.Scan(net.IP{10, 0, 8, 91}, 80)
		if err != nil {
			log.Fatalln("Scan error", err.Error())
		} else {
			log.Println("Success scan")
		}
		time.Sleep(time.Millisecond * 100)
	}
	log.Println("Send all")
	time.Sleep(time.Second * 100)
}
