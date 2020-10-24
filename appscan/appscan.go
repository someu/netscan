package appscan

import (
	"context"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

const Version = "0.0.1"
const DefaultConcurrent = 100
const DefaultAppScanTimeoutPerUrl = time.Duration(15) * time.Second

type AppScanner struct {
	Features      []*Feature
	RequestClient *RequestClient
	pool          *ants.Pool
}

type AppScan struct {
	ctx     context.Context
	Scanner *AppScanner
	StartAt time.Time
	EndAt   time.Time
	Results []*AppScanResult
	Wait    func()
	Cancel  func()
	Config  *AppScanConfig
}

type AppScanResult struct {
	Url             string
	MatchedFeatures []*MatchedFeature
}

type AppScanConfig struct {
	Urls     []string
	Features []*Feature
	Callback func(v interface{})
	Timeout  time.Duration
}

func NewAppScanner() (*AppScanner, error) {
	pool, err := ants.NewPool(DefaultConcurrent)
	if err != nil {
		return nil, err
	}
	scanner := &AppScanner{
		RequestClient: NewRequestClient(),
		pool:          pool,
		Features:      Features,
	}
	return scanner, nil
}

func (scanner *AppScanner) SetRequestTimeout(timeout time.Duration) {
	scanner.RequestClient.HttpClient.Timeout = timeout
}

func (scanner *AppScanner) SetConcurrent(concurrent int) {
	scanner.pool.Tune(concurrent)
}

func (scanner *AppScanner) CreateScan(config *AppScanConfig) *AppScan {
	timeout := time.Duration(len(config.Urls)) * DefaultAppScanTimeoutPerUrl
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var locker sync.Mutex
	scan := &AppScan{
		ctx:     ctx,
		Scanner: scanner,
		StartAt: time.Now(),
		Config:  config,
		Cancel:  cancel,
	}

	wg := sync.WaitGroup{}
	for _, url := range config.Urls {
		func(url string) {
			wg.Add(1)
			scanner.pool.Submit(func() {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				default:
					result := scan.ScanUrl(url)
					locker.Lock()
					defer locker.Unlock()
					scan.Results = append(scan.Results, result)
					if config.Callback != nil {
						config.Callback(result)
					}
				}
			})
		}(url)
	}

	scan.Wait = func() {
		wg.Wait()
	}

	return scan
}

func (scan *AppScan) ScanUrl(url string) *AppScanResult {
	var results []*MatchedFeature
	var featuresMap = make(map[string][]*Feature)
	for _, feature := range scan.Config.Features {
		featuresMap[feature.Path] = append(featuresMap[feature.Path], feature)
	}
	for path, features := range featuresMap {
		target := url
		if path != "/" {
			target += path
		}
		response, err := scan.Scanner.RequestClient.Get(target, scan.ctx)
		if err != nil {
			continue
		}
		for _, feature := range features {
			result := feature.MatchResponseData(response.Data)
			if result != nil {
				results = append(results, result)
			}
		}
	}

	return &AppScanResult{
		Url:             url,
		MatchedFeatures: results,
	}
}
