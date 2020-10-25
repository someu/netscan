package portscan

import (
	"context"
	"net"
	"sync"
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
	ctx          context.Context
	cancel       context.CancelFunc
	locker       sync.Mutex
	sender       *Sender
	receiver     *Receiver
	config       *PortScannerConfig
	runningScan  *PortScan
	waitingScans []*PortScan
	scaning      bool
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
		locker:   sync.Mutex{},
		receiver: NewReceiver(ctx),
		config:   config,
		scaning:  false,
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

	// add scan
	scanner.locker.Lock()
	if scanner.runningScan != nil {
		scanner.waitingScans = append(scanner.waitingScans, scan)
	} else {
		scanner.runningScan = scan
		scanner.run()
	}
	scanner.locker.Unlock()

	go func() {
		defer cancel()

		for {
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

func (scanner *PortScanner) run() {
	go func() {
		if scanner.runningScan == nil {
			return
		}
		scan := scanner.runningScan

		for {
			if !scan.Target.hasNext() {
				break
			}
			ip, port, err := scan.Target.nextTarget()
			if err != nil {
				scan.Errors = append(scan.Errors, err)
				continue
			}
			route, err := scanner.sender.send(ip, port)
			if err != nil {
				scan.Errors = append(scan.Errors, err)
				continue
			}

			scanner.receiver.startReceive(route.iface.Name, route.handle, scan.ctx)
			time.Sleep(scan.perPacketInterval)
		}

		// triger next scan
		scanner.locker.Lock()
		if len(scanner.waitingScans) > 0 {
			scanner.runningScan = scanner.waitingScans[0]
			scanner.waitingScans = scanner.waitingScans[1:]
			scanner.run()
		} else {
			scanner.runningScan = nil
		}
		scanner.locker.Unlock()
	}()
}

func (scan *PortScan) Wait() {
	<-scan.ctx.Done()
}

func (scan *PortScan) Cancel() {
	scan.cancel()
}
