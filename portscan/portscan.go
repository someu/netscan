package portscan

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
	Scanner *PortScanner
	Target  *TargetRange
	Config  *ScanConfig
	Cancel  context.CancelFunc
	Wait    func()
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
		Scanner: scanner,
		Target:  target,
		Config:  config,
		Cancel:  cancel,
		Wait:    wait,
	}
}
