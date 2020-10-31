package portscan

import "sync"

var (
	runningScan    *PortScan
	waitingScans   []*PortScan
	scheduleLocker = sync.Mutex{}
)

func scheduleNext() {
	scheduleLocker.Lock()
	defer scheduleLocker.Unlock()
	if runningScan != nil && runningScan.Status == STATUS_RUNNING {
		return
	} else if runningScan != nil && runningScan.Status == STATUS_WAITING {
		runningScan.Run()
		return
	} else {
		if len(waitingScans) == 0 {
			runningScan = nil
		} else {
			runningScan = waitingScans[0]
			waitingScans = waitingScans[1:]
			runningScan.Run()
		}
	}
}

func addPortScan(scan *PortScan) {
	scheduleLocker.Lock()
	defer scheduleLocker.Unlock()
	if runningScan != nil {
		waitingScans = append(waitingScans, scan)
	} else {
		runningScan = scan
	}
}
