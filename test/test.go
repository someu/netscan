package main

import (
	"encoding/json"
	"fmt"
)

type S struct {
	A string
}

func main() {
	s := S{
		A: "123",
	}
	bytes, _ := json.Marshal(s)
	fmt.Println(string(bytes))
}
