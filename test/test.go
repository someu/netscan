package main

import (
	"github.com/someu/netscan/appscan"
	"log"
)

func main() {
	scanner, err := appscan.NewAppScanner()
	if err != nil {
		log.Panic(err)
	}
	scanConf := &appscan.AppScanConfig{
		Urls:     []string{"https://www.baidu.com"},
		Features: appscan.Features,
	}
	scan := scanner.CreateScan(scanConf)
	scan.Wait()
	log.Println(scan.Results)
}
