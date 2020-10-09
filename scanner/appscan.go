package scanner

import (
	"encoding/json"
	"fakescan/util"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/panjf2000/ants/v2"
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

const Version = "0.0.1"

type Scanner struct {
	FeatureCollection []*Feature
	RequestClient     *RequestClient
	Level             int
	MasscanPath       string
	MasscanRate       int
	Concurrent        int
	pool              *ants.Pool
}

type MatchedApp struct {
	Name     string
	Versions []string
}

type MatchedResult struct {
	Url     string
	StartAt time.Time
	EndAt   time.Time
	Apps    []*MatchedApp
}

func NewScanner() *Scanner {
	scanner := &Scanner{
		Level:         1,
		MasscanPath:   "masscan",
		MasscanRate:   1000,
		Concurrent:    100,
		RequestClient: NewRequestClient(),
	}
	var err error
	if scanner.pool, err = ants.NewPool(scanner.Concurrent); err != nil {
		log.Fatalf("Init goroutine pool failed, %s", err.Error())
	}
	scanner.LoadFeatures()
	return scanner
}

func (scanner *Scanner) LoadFeatures() {
	sources := packr.New(`sources`, `../sources`)

	featureBytes, err := sources.Find(`features.json`)
	if err != nil {
		log.Panic("load features failed!")
	} else {
		json.Unmarshal(featureBytes, &scanner.FeatureCollection)
		featureRuleItemSliceType := reflect.TypeOf([]*FeatureRuleItem{})
		featureRuleItemMapType := reflect.TypeOf(make(map[string][]*FeatureRuleItem))
		for _, feature := range scanner.FeatureCollection {
			for _, rule := range feature.Rules {
				ruleValue := reflect.ValueOf(*rule)
				for i := 0; i < ruleValue.NumField(); i++ {
					ruleFieldValue := ruleValue.Field(i)
					ruleFieldValueType := ruleFieldValue.Type()

					if ruleFieldValueType == featureRuleItemSliceType {
						for _, ruleItem := range ruleFieldValue.Interface().([]*FeatureRuleItem) {
							if len(ruleItem.Regexp) > 0 {
								ruleItem.regexp, err = regexp.Compile(fmt.Sprintf("(?i)%s", ruleItem.Regexp))
								if err != nil {
									fmt.Println(ruleItem.Regexp)
									recover()
								}
							}
						}
					} else if ruleFieldValueType == featureRuleItemMapType {
						for _, ruleItems := range ruleFieldValue.Interface().(map[string][]*FeatureRuleItem) {
							for _, ruleItem := range ruleItems {
								if len(ruleItem.Regexp) > 0 {
									ruleItem.regexp, err = regexp.Compile(fmt.Sprintf("(?i)%s", ruleItem.Regexp))
									if err != nil {
										fmt.Println(ruleItem.Regexp)
										recover()
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func (scanner *Scanner) SetTimeout(timeout int) {
	scanner.RequestClient.HttpClient.Timeout = time.Second * time.Duration(timeout)
}

func (scanner *Scanner) SetConcurrent(concurrent int) int {
	scanner.Concurrent = concurrent
	scanner.pool.Tune(scanner.Concurrent)
	return scanner.Concurrent
}

func (scanner *Scanner) ScanUrl(url string) *MatchedResult {
	var result = &MatchedResult{Url: url, StartAt: time.Now()}
	for _, feature := range scanner.FeatureCollection[:scanner.Level] {
		var (
			target   = url + feature.Path
			response *Response
			err      error
		)
		if response, err = scanner.RequestClient.Get(target); err != nil {
			continue
		}

		for _, rule := range feature.Rules {
			app := rule.MatchResponseData(response.Data)
			if app != nil {
				result.Apps = append(result.Apps, app)
			}
		}
	}
	result.EndAt = time.Now()
	return result
}

// async scan a list of url, return a waitgroup
func (scanner *Scanner) scanUrls(urls []string, callback func(*MatchedResult), wg *sync.WaitGroup) {
	var lock sync.Mutex
	for _, url := range urls {
		wg.Add(1)
		func(url string) {
			scanner.pool.Submit(func() {
				defer wg.Done()
				result := scanner.ScanUrl(url)
				lock.Lock()
				defer lock.Unlock()
				callback(result)
			})
		}(url)
	}
}

func (scanner *Scanner) masscan(targets string, ports string) []string {
	var urls []string
	if len(targets) == 0 || len(ports) == 0 {
		return urls
	}

	log.Printf("Start masscan %s %s", targets, ports)
	masscan := NewMasscan(targets, ports)
	masscan.SetRate(scanner.MasscanRate)
	masscan.SetProgramPath(scanner.MasscanPath)
	var err error
	if urls, err = masscan.Scan(); err != nil {
		log.Printf("Masscan failed: %s", err.Error())
	} else {
		log.Printf("Masscan finished, find %d urls", len(urls))
	}
	return urls
}

func (scanner *Scanner) Scan(targets []string, ports []string, callback func(*MatchedResult)) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	go func() {
		var urls []string
		var cidrs []string
		for _, target := range targets {
			if util.IsIP(target) || util.IsCIDR(target) {
				cidrs = append(cidrs, target)
			} else {
				urls = append(urls, target)
			}
		}
		u := scanner.masscan(strings.Join(cidrs, ","), strings.Join(ports, ","))
		urls = append(u, urls...)
		scanner.scanUrls(urls, callback, wg)
	}()
	return wg

}

func (scanner *Scanner) ScanSync(targets []string, ports []string) []*MatchedResult {
	var results []*MatchedResult
	wg := scanner.Scan(targets, ports, func(result *MatchedResult) {
		results = append(results, result)
	})
	wg.Wait()
	return results
}

func (result *MatchedResult) String() string {
	return util.Stringify(result)
}
