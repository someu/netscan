package appscan

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"regexp"
	"strings"
	"sync"
	"time"
)

const Version = "0.0.1"
const ScanConcurrent = 100
const DefaultRequestConcurrent = 100
const DefaultRequestTimeout = 30 * time.Second

var (
	requestPool   *ants.Pool
	appScanInited = false
)

type AppScan struct {
	ctx      context.Context
	wg       sync.WaitGroup
	locker   sync.Mutex
	scanPool *ants.PoolWithFunc
	errors   []error
	client   *RequestClient
	StartAt  time.Time
	EndAt    time.Time
	Results  []*AppScanResult
	cancel   func()
	Config   *AppScanConfig
}

type AppScanResult struct {
	Url             string
	ResponseData    *ResponseData
	MatchedFeatures []*MatchedFeature
}

type AppScanConfig struct {
	Urls           []string
	Features       []*Feature
	Callback       func(result *AppScanResult)
	ScanTimeout    time.Duration
	RequestTimeout time.Duration
}

func SetRequestConcurrent(concurrent int) {
	requestPool.Tune(concurrent)
}

func (res AppScanResult) String() string {
	var featureStrs []string
	for _, feature := range res.MatchedFeatures {
		featureStr := feature.Feature.Name
		if len(feature.Versions) > 0 {
			featureStr += fmt.Sprintf("[versions: %s]", strings.Join(feature.Versions, ", "))
		}

		//if len(feature.Proofs) > 0 {
		//	featureStr += fmt.Sprintf("(proofs: %s)", strings.Join(feature.Proofs, ", "))
		//}
		featureStrs = append(featureStrs, featureStr)
		if len(feature.Feature.Implies) > 0 {
			featureStrs = append(featureStrs, feature.Feature.Implies...)
		}
	}

	return fmt.Sprintf("%s: %s", res.Url, strings.Join(featureStrs, ", "))
}

func CreateScan(config *AppScanConfig) (*AppScan, error) {
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
		scanTimeout = time.Duration(requestCount)*DefaultRequestTimeout + 5*time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)

	// create scan
	scan := &AppScan{
		ctx:     ctx,
		wg:      sync.WaitGroup{},
		locker:  sync.Mutex{},
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
		if scan.Config.Callback != nil {
			scan.Config.Callback(result)
		}
	}
	var err error
	scan.scanPool, err = ants.NewPoolWithFunc(ScanConcurrent, scanFunc)
	if err != nil {
		return nil, err
	}

	for _, url := range config.Urls {
		scan.wg.Add(1)
		scan.scanPool.Invoke(url)
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
	currentScanWg := sync.WaitGroup{}
	for path, features := range featuresMap {
		target := tailTrimedUrl
		if path != "/" {
			target += path
		}
		func(target string) {
			currentScanWg.Add(1)
			requestPool.Submit(func() {
				defer currentScanWg.Done()
				select {
				case <-scan.ctx.Done():
					fmt.Printf("scan task has been finished, reason: %s\n", scan.ctx.Err())
					return
				default:
					break
				}
				response, err := scan.client.Get(target, scan.ctx)
				if err != nil {
					fmt.Printf("get %s error %s\n", target, err)
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
			})
		}(target)
	}
	currentScanWg.Wait()
	result := &AppScanResult{
		Url:             url,
		MatchedFeatures: results,
	}

	return result
}

func (scan *AppScan) Wait() {
	scan.wg.Wait()
	scan.Stop()
}

func (scan *AppScan) Stop() {
	scan.cancel()
	scan.scanPool.Release()
	scan.EndAt = time.Now()
}

func (scan AppScan) Errors() []error {
	return scan.errors
}

func initRequestPool() error {
	var err error
	requestPool, err = ants.NewPool(DefaultRequestConcurrent)
	return err
}

func InitAppScan() error {
	if appScanInited == false {
		if err := initRequestPool(); err != nil {
			return err
		}
		if err := initFeatureRegexps(); err != nil {
			return err
		}
		appScanInited = true
	}

	return nil
}
