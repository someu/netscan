package appscan

import (
	"context"
	"github.com/panjf2000/ants/v2"
	"regexp"
	"sync"
	"time"
)

const Version = "0.0.1"
const DefaultConcurrent = 100
const DefaultRequestTimeout = 30 * time.Second

type AppScanner struct {
	Features    []*Feature
	Config      *AppScannerConfig
	requestPool *ants.Pool
}

type AppScannerConfig struct {
	Concurrent int
}

type AppScan struct {
	ctx      context.Context
	wg       sync.WaitGroup
	locker   sync.Mutex
	scanPoop *ants.PoolWithFunc
	errors   []error
	client   *RequestClient
	Scanner  *AppScanner
	StartAt  time.Time
	EndAt    time.Time
	Results  []*AppScanResult
	cancel   func()
	Config   *AppScanConfig
}

type AppScanResult struct {
	Url             string
	MatchedFeatures []*MatchedFeature
}

type AppScanConfig struct {
	Urls           []string
	Features       []*Feature
	Callback       func(v interface{})
	ScanTimeout    time.Duration
	RequestTimeout time.Duration
}

func NewAppScanner(config *AppScannerConfig) (*AppScanner, error) {
	if config.Concurrent <= 0 {
		config.Concurrent = DefaultConcurrent
	}
	pool, err := ants.NewPool(DefaultConcurrent)
	if err != nil {
		return nil, err
	}
	scanner := &AppScanner{
		requestPool: pool,
		Features:    Features,
		Config:      config,
	}
	return scanner, nil
}

func (scanner *AppScanner) SetConcurrent(concurrent int) {
	scanner.requestPool.Tune(concurrent)
}

func (scanner *AppScanner) CreateScan(config *AppScanConfig) (*AppScan, error) {
	// calc requests per url
	existFlag := make(map[string]bool)
	requestCountPerUrl := 0
	for _, feature := range config.Features {
		if !existFlag[feature.Path] {
			requestCountPerUrl += 1
		}
		existFlag[feature.Path] = true
	}

	if config.RequestTimeout == 0 {
		config.RequestTimeout = DefaultRequestTimeout
	}

	// total scan timeout
	scanTimeout := config.ScanTimeout
	if scanTimeout == 0 {
		requestCount := len(config.Urls) * requestCountPerUrl
		scanTimeout = time.Duration(requestCount)*DefaultRequestTimeout + time.Duration(requestCount)*time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)

	// create scan
	scan := &AppScan{
		ctx:     ctx,
		wg:      sync.WaitGroup{},
		locker:  sync.Mutex{},
		Scanner: scanner,
		StartAt: time.Now(),
		Config:  config,
		cancel:  cancel,
		client: NewRequestClient(&RequestClientConfig{
			Timeout: config.RequestTimeout,
		}),
	}

	// create scan pool
	scanFunc := func(url interface{}) {
		defer scan.wg.Done()
		result := scan.ScanUrl(url.(string))
		scan.locker.Lock()
		defer scan.locker.Unlock()
		scan.Results = append(scan.Results, result)
	}
	var err error
	scan.scanPoop, err = ants.NewPoolWithFunc(DefaultConcurrent, scanFunc)
	if err != nil {
		return nil, err
	}

	for _, url := range config.Urls {
		scan.wg.Add(1)
		scan.scanPoop.Invoke(url)
	}

	return scan, nil
}

func (scan *AppScan) ScanUrl(url string) *AppScanResult {
	var results []*MatchedFeature
	var featuresMap = make(map[string][]*Feature)
	for _, feature := range scan.Config.Features {
		featuresMap[feature.Path] = append(featuresMap[feature.Path], feature)
	}
	tailRe := regexp.MustCompile("\\/$")
	tailTrimedUrl := tailRe.ReplaceAllString(url, "")
	locker := sync.Mutex{}
	wg := sync.WaitGroup{}
	for path, features := range featuresMap {
		target := tailTrimedUrl
		if path != "/" {
			target += path
		}
		func(target string) {
			wg.Add(1)
			scan.Scanner.requestPool.Submit(func() {
				defer wg.Done()
				select {
				case <-scan.ctx.Done():
					return
				default:
					response, err := scan.client.Get(target, scan.ctx)
					if err != nil {
						scan.errors = append(scan.errors, err)
						return
					}
					for _, feature := range features {
						result := feature.MatchResponseData(response.Data)
						if result != nil {
							locker.Lock()
							results = append(results, result)
							locker.Unlock()
						}
					}
				}
			})
		}(target)
	}
	wg.Wait()
	result := &AppScanResult{
		Url:             url,
		MatchedFeatures: results,
	}
	return result
}

func (scan *AppScan) Wait() {
	scan.wg.Wait()
	scan.scanPoop.Release()
	scan.EndAt = time.Now()
}

func (scan *AppScan) Cancel() {
	scan.cancel()
	scan.Wait()
}

func (scan AppScan) Errors() []error {
	return scan.errors
}
