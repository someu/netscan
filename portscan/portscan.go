package main

import (
	"context"
	"log"
	"time"
)

type ScanConfig struct {
	Timeout         time.Duration
	PacketPerSecond uint
}

type Scan struct {
	scanner *PortScanner
	target  *TargetRange
	config  *ScanConfig
	cancel  context.CancelFunc
	wait    func()
}

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

func (scanner *PortScanner) CreateScan(target *TargetRange, config *ScanConfig) *Scan {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		// send a packet spend time, Nanosecond
		nsPerPacket := 1000000000 / config.PacketPerSecond
		if nsPerPacket == 0 {
			nsPerPacket = 1
		}

		var i uint = 1
		for ; ; i++ {
			if !target.hasNext() {
				break
			}
			ip, port, err := target.nextTarget()
			if err != nil {
				log.Printf("get %d ip port error %s\n", i, err)
				continue
			}
			route, err := scanner.sender.send(ip, port)
			if err != nil {
				log.Printf("send packet to %s:%d error %s\n", ip, port, err)
			}

			scanner.receiver.startReceive(route.iface.Name, route.handle, ctx)
			time.Sleep(time.Duration(nsPerPacket) * time.Nanosecond)
		}
		log.Println("send all packet", i)
		// time out
		timeout, _ := context.WithTimeout(ctx, config.Timeout)
		<-timeout.Done()
	}()

	wait := func() {
		<-ctx.Done()
	}

	return &Scan{
		scanner: scanner,
		target:  target,
		config:  config,
		cancel:  cancel,
		wait:    wait,
	}
}

func main() {
	scanner, err := NewPortScanner()
	if err != nil {
		log.Fatalln("Create scanner error", err.Error())
	}
	//ip := net.IP{122, 51, 121, 205}
	//ports := []uint16{21, 22, 80, 8080, 443, 27017, 13443}

	ipSegment, _ := parseIpSegment("113.160.0.218/16")
	portSegment, _ := parsePortSegment("80")
	target := NewTargetRange(
		Segments{ipSegment}, Segments{portSegment},
	)
	conf := &ScanConfig{
		Timeout:         time.Second * 10,
		PacketPerSecond: 200,
	}

	scan := scanner.CreateScan(target, conf)

	log.Println("created scan")

	scan.wait()

	log.Println("finished")
	time.Sleep(time.Second * 1)
}
