package main

import (
	"fakescan/scanner"
	"log"
	"regexp"
)

type Rule struct {
	Reg    string         `json:"reg"`
	regexp *regexp.Regexp `json:"regexp"`
}

func main() {
	url := "118.68.0.183"
	globalScanner := scanner.NewScanner()
	log.Println("start scan", url)
	apps := globalScanner.ScanUrl(url)
	log.Println("start scan", apps)

}
