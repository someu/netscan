package main

import (
	"context"
	"log"
	"net"
	"time"
)

type PortScanner struct {
	ctx            context.Context
	cancel         context.CancelFunc
	sender         *Sender
	receiver       *Receiver
	arpTimeout     time.Duration
	sessionTimeout time.Duration
}

func NewPortScanner() (*PortScanner, error) {
	sender, err := NewSender()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &PortScanner{
		ctx:      ctx,
		cancel:   cancel,
		sender:   sender,
		receiver: NewReceiver(ctx),
	}, nil
}

func (scanner *PortScanner) Scan(ip net.IP, port uint16) error {
	route, err := scanner.sender.send(ip, port)
	if err != nil {
		return err
	}

	scanner.receiver.startReceive(route.iface.Name, route.handle, scanner.ctx)

	return nil
}

func main() {
	scanner, err := NewPortScanner()
	if err != nil {
		log.Fatalln("Create scanner error", err.Error())
	}
	ip := net.IP{10, 0, 8, 92}
	ports := []uint16{21, 22, 80, 8080, 443, 27017, 13443}
	for _, port := range ports {
		err = scanner.Scan(ip, port)
		if err != nil {
			log.Printf("Scan %s:%d error %s", ip.String(), port, err.Error())
		}
	}
	log.Println("Send all")
	time.Sleep(time.Second * 10)
	scanner.cancel()
	log.Println("finished")
	time.Sleep(time.Second * 1)
}
