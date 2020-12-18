package main

import (
	"fmt"
	"github.com/someu/netscan/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
