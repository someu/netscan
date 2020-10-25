package main

import (
	"github.com/someu/netscan/appscan"
	"log"
)

func main() {
	features := []*appscan.Feature{}
	for _, feature := range appscan.Features {
		if feature.Path == "/" {
			features = append(features, feature)
		}
	}
	scanner, err := appscan.NewAppScanner()
	if err != nil {
		log.Panic(err)
	}
	scanConf := &appscan.AppScanConfig{
		Urls:     []string{"http://127.0.0.1:8080/", "https://be.scanv.com", "https://bigteds.ru/"},
		Features: features,
	}
	scan, err := scanner.CreateScan(scanConf)
	if err != nil {
		log.Panic(err)
	}
	scan.Wait()
	for _, result := range scan.Results {
		log.Println(result.Url)
		for _, mr := range result.MatchedFeatures {
			log.Println(mr.Feature.Name, mr.Versions, mr.Proofs)
		}

	}
	for _, err := range scan.Errors() {
		log.Println(err)
	}

}
