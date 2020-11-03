package main

import (
	"fmt"
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

func main() {

	fmt.Println(uniq([]string{"3.3.4", "3.3.4", "3.3.4"}))
}
