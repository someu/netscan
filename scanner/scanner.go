package scanner

import (
	"encoding/json"
	"github.com/gobuffalo/packr/v2"
	"log"
)

type Scanner struct {
	FeatureCollection []*Feature
	RequestClient     *RequestClient
	Level             int
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
		log.Panic("ad features failed!")
	} else {
		json.Unmarshal(featureBytes, &scanner.FeatureCollection)
	}
}

func (scanner *Scanner) Scan(url string) ([]string, error) {
	var apps []string
	for _, feature := range scanner.FeatureCollection[:scanner.Level] {
		var (
			target   = url + feature.Path
			response *Response
			err      error
		)
		if response, err = scanner.RequestClient.Get(target); err != nil {
			return nil, err
		}

		for _, rule := range feature.Rules {
			matched, _ := rule.MatchResponse(response)
			if matched {
				apps = append(apps, rule.Name)
			}
		}
	}
	return apps, nil
}
