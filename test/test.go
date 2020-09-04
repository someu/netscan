package main

import (
	"fakescan/scanner"
	"log"
)

func main() {
	url := "10.12.21.154"
	globalScanner := scanner.NewScanner()
	log.Println("start scan", url)
	apps, _ := globalScanner.Scan(url)
	log.Println("start scan", apps)
}
