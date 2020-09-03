package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var url string

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "start a scan",
	Long:  `start a scan`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("[start scan]:", url)
		if apps, err := globalScanner.Scan(url); err == nil {
			log.Println(fmt.Sprintf("[scan %s finished]: %s", url, strings.Join(apps, ", ")))
		} else {
			log.Println(err)
		}
	},
}

func init() {
	scanCmd.Flags().StringVarP(&url, "url", "u", "", "scan target")
	rootCmd.AddCommand(scanCmd)
}
