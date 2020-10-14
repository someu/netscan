package main

import (
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"net"
)

var handleMap map[string]*pcap.Handle

func init() {
	handleMap = make(map[string]*pcap.Handle)
}

type NetInterface struct {
	name       string
	iface      *net.Interface
	ip         net.IP
	mac        net.HardwareAddr
	gateway    net.IP
	gatewayMac net.HardwareAddr
	handle     *pcap.Handle
}

type Router struct {
	router routing.Router
}

func NewRouter() (*Router, error) {
	router, err := routing.New()
	if err != nil {
		return nil, err
	}
	return &Router{
		router: router,
	}, nil
}

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

func (r *Router) routeIp(ip net.IP) (*NetInterface, error) {
	iface, gateway, ip, err := r.router.Route(ip)
	if err != nil {
		return nil, err
	}
	handle, err := getHandle(iface.Name)
	return &NetInterface{
		name:    iface.Name,
		iface:   iface,
		gateway: gateway,
		ip:      ip,
		mac:     iface.HardwareAddr,
		handle:  handle,
	}, nil
}
