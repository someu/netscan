package main

import (
	"log"
	"os"
	"time"
)

func uniq(arr []string) []string {
	valueMap := make(map[string]bool)
	for _, v := range arr {
		valueMap[v] = true
	}
	var newArr []string
	for value, _ := range valueMap {
		newArr = append(newArr, value)
	}
	return newArr
}

func ioExample() {
	file, err := os.OpenFile("test", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	file.WriteString("12666s3")
	go func() {
		time.Sleep(time.Second)
		os.Exit(0)
	}()
	time.Sleep(time.Second * 3)
}

func main() {

	ioExample()
}
