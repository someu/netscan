package main

import (
	"github.com/spf13/cobra"
	"log"
	"netscan/appscan"
)

var globalScanner = appscan.NewScanner()

var rootCmd = &cobra.Command{
	Use:   "netscan",
	Short: "netscan is a web scanner",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(rootCmd.Help())
	}
}
