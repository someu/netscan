package schedule

import (
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"grab"
	"grab/modules"
)

var taskPool *ants.Pool
var ModuleSet grab.ModuleSet

func init() {
	ModuleSet = modules.NewModuleSetWithDefaults()
	var err error
	if taskPool, err = ants.NewPool(100); err != nil {
		logrus.Panicf("init request pool error: %s", err)
	}
}

func SetThreadCount(n int) {
	taskPool.Tune(n)
}

func AddScan(protocol string, target grab.ScanTarget, callback func(status grab.ScanStatus, res interface{}, err error)) error {
	if module, ok := ModuleSet[protocol]; ok {
		taskPool.Submit(func() {
			scanner := module.NewScanner()
			flags := module.NewFlags()
			scanner.Init(flags.(grab.ScanFlags))
			status, res, err := scanner.Scan(target)
			callback(status, res, err)
		})
	} else {
		return errors.New(fmt.Sprintf("no module for protocol: %s", protocol))
	}
	return nil
}
