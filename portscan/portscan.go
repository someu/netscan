package portscan

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type PortScanConfig struct {
	IPSegments      Segments
	PortSegments    Segments
	Timeout         time.Duration
	PacketPerSecond uint
	Callback        func(result PortScanResult)
}

type PortScanResult struct {
	IP   net.IP
	Port uint16
}

type PortScan struct {
	ctx                context.Context
	cancel             context.CancelFunc
	packetSendInterval time.Duration
	target             *TargetRange
	statusLocker       sync.Mutex
	pushResultLocker   sync.Mutex
	Config             *PortScanConfig
	Status             int
	CreatAt            time.Time
	StartAt            time.Time
	EndAt              time.Time
	Results            []PortScanResult
	Errors             []error
}

const (
	STATUS_WAITING = iota
	STATUS_RUNNING
	STATUS_FINISHED
)

const DefaultPortScanTimeout = time.Second * 10

var portScanInited = false

func CreatePortScan(config *PortScanConfig) (*PortScan, error) {
	if len(config.IPSegments) == 0 || len(config.PortSegments) == 0 {
		return nil, errors.New("invalid IPSegments or PortSegments\n")
	}

	if config.Timeout == 0 {
		config.Timeout = DefaultPortScanTimeout
	}

	ctx, cancel := context.WithCancel(context.Background())

	// calc packet send interval
	packetSendInterval := 1000000000 / config.PacketPerSecond
	if packetSendInterval == 0 {
		packetSendInterval = 1
	}

	scan := &PortScan{
		ctx:                ctx,
		cancel:             cancel,
		packetSendInterval: time.Duration(packetSendInterval) * time.Nanosecond,
		statusLocker:       sync.Mutex{},
		pushResultLocker:   sync.Mutex{},
		target:             NewTargetRange(config.IPSegments, config.PortSegments),
		Status:             STATUS_WAITING,
		Config:             config,
		CreatAt:            time.Now(),
	}

	addPortScan(scan)
	scheduleNext()

	return scan, nil
}

func (scan *PortScan) Run() {
	scan.statusLocker.Lock()
	defer scan.statusLocker.Unlock()

	if scan.Status != STATUS_WAITING {
		scan.Errors = append(scan.Errors, errors.New("current scan not in waiting"))
		scan.Stop()
		return
	}
	if scan.target == nil {
		scan.Errors = append(scan.Errors, errors.New("scan's target is nil"))
		scan.Stop()
		return
	}
	scan.Status = STATUS_RUNNING
	scan.StartAt = time.Now()

	updateCookieKey()

	go func() {
		defer scan.Stop()
		target := scan.target
		for {
			start := time.Now().UnixNano()
			select {
			case <-scan.ctx.Done():
				return
			default:
				break
			}
			if !target.hasNext() {
				break
			}
			ip, port, err := target.nextTarget()
			if err != nil {
				log.Println(err)
				scan.Errors = append(scan.Errors, err)
				continue
			}
			if err := send(ip, port); err != nil {
				scan.Errors = append(scan.Errors, err)
				log.Println(err)
				continue
			}
			sleep := scan.packetSendInterval.Nanoseconds() - (time.Now().UnixNano() - start)
			if sleep > 0 {
				time.Sleep(time.Duration(sleep) * time.Nanosecond)
			}
		}
		// log.Println("send all packet")
		// wait for timeout after send all packet
		timeout, _ := context.WithTimeout(scan.ctx, scan.Config.Timeout)
		<-timeout.Done()
	}()
}

func (scan *PortScan) pushResult(ip net.IP, port uint16) {
	scan.pushResultLocker.Lock()
	defer scan.pushResultLocker.Unlock()
	result := PortScanResult{
		IP:   ip,
		Port: port,
	}
	scan.Results = append(scan.Results, result)
	if scan.Config.Callback != nil {
		scan.Config.Callback(result)
	}
}

func (scan *PortScan) Stop() {
	scan.statusLocker.Lock()
	defer scan.statusLocker.Unlock()
	defer scheduleNext()
	scan.cancel()
	scan.Status = STATUS_FINISHED
	scan.EndAt = time.Now()
}

func (scan *PortScan) Wait() {
	<-scan.ctx.Done()
}

func InitPortScan() error {
	if portScanInited == false {
		if err := initRecvInterfaces(); err != nil {
			return err
		}
		portScanInited = true
	}

	return nil
}
