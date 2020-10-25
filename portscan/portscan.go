package portscan

import (
	"context"
	"net"
	"time"
)

const DefaultArpTimeout = time.Second * 10

type PortScanConfig struct {
	Timeout         time.Duration
	PacketPerSecond uint
}

type PortScanResult struct {
	IP   net.IP
	Port uint8
}

type PortScan struct {
	ctx               context.Context
	cancel            context.CancelFunc
	perPacketInterval time.Duration
	Scanner           *PortScanner
	Target            *TargetRange
	Config            *PortScanConfig
	StartAt           time.Time
	EndAt             time.Time
	Results           []*PortScanResult
	Errors            []error
}

type PortScannerConfig struct {
	ArpTimeout      time.Duration
	PacketPerSecond uint
}

type PortScanner struct {
	ctx      context.Context
	cancel   context.CancelFunc
	sender   *Sender
	receiver *Receiver
	config   *PortScannerConfig
}

func NewPortScanner(config *PortScannerConfig) (*PortScanner, error) {
	if config.ArpTimeout == 0 {
		config.ArpTimeout = DefaultArpTimeout
	}
	sender, err := NewSender(&SenderConfig{
		arpTimeout: config.ArpTimeout,
	})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &PortScanner{
		ctx:      ctx,
		cancel:   cancel,
		sender:   sender,
		receiver: NewReceiver(ctx),
		config:   config,
	}, nil
}

func (scanner *PortScanner) CreatePortScan(target *TargetRange, config *PortScanConfig) *PortScan {
	ctx, cancel := context.WithCancel(context.Background())

	// send packet interval
	nsPerPacket := 1000000000 / config.PacketPerSecond
	if nsPerPacket == 0 {
		nsPerPacket = 1
	}

	scan := &PortScan{
		Scanner:           scanner,
		Target:            target,
		Config:            config,
		ctx:               ctx,
		cancel:            cancel,
		perPacketInterval: time.Duration(nsPerPacket) * time.Nanosecond,
	}

	go func() {
		defer cancel()

		var i uint = 1
		for ; ; i++ {
			if !target.hasNext() {
				break
			}
			ip, port, err := target.nextTarget()
			if err != nil {
				scan.Errors = append(scan.Errors, err)
				continue
			}
			route, err := scanner.sender.send(ip, port)
			if err != nil {
				scan.Errors = append(scan.Errors, err)
				continue
			}

			scanner.receiver.startReceive(route.iface.Name, route.handle, ctx)
			time.Sleep(scan.perPacketInterval)
		}

		// wait for timeout after send all packet
		timeout, _ := context.WithTimeout(ctx, config.Timeout)
		<-timeout.Done()
	}()

	return scan
}

func (scanner *PortScanner) Wait() {
	<-scanner.ctx.Done()
}

func (scanner *PortScanner) Cancel() {
	scanner.cancel()
}
