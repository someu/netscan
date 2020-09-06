package main

import (
	"fakescan/scanner"
	"github.com/spf13/cobra"
	"log"
)

var globalScanner = scanner.NewScanner()

var rootCmd = &cobra.Command{
	Use:   "fakescan",
	Short: "FakeScan is a web scanner",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(rootCmd.Help())
	}
}
