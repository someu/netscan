package scanner

import (
	"encoding/json"
	"fakescan/util"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"log"
	"reflect"
	"regexp"
	"sync"
	"time"
)

type Scanner struct {
	FeatureCollection []*Feature
	RequestClient     *RequestClient
	Level             int
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
		Level: 1,
	}
	scanner.LoadFeatures()
	scanner.RequestClient = NewRequestClient()
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
			app := rule.MatchResponse(response)
			if app != nil {
				result.Apps = append(result.Apps, app)
			}
		}
	}
	result.EndAt = time.Now()
	return result
}

// async scan a list of url, return a waitgroup
func (scanner *Scanner) ScanUrls(urls []string, callback func(*MatchedResult)) func() {
	wg := sync.WaitGroup{}
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			result := scanner.ScanUrl(url)
			callback(result)
		}(url)
	}
	return func() {
		wg.Wait()
	}
}

// sync scan a list of url
func (scanner *Scanner) ScanUrlsSync(urls []string) []*MatchedResult {
	var results []*MatchedResult
	var lock sync.Mutex
	scanner.ScanUrls(urls, func(result *MatchedResult) {
		defer lock.Unlock()
		lock.Lock()
		results = append(results, result)
	})
	return results
}

func (result *MatchedResult) String() string {
	return util.Stringify(result)
}
