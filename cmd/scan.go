package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var url string

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "start a scan",
	Long:  `start a scan`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("[start scan]:", url)
		if apps, err := globalScanner.Scan(url); err == nil {
			var result string
			for _, app := range apps {
				result += fmt.Sprintf("%s %s,", app.Name, strings.Join(app.Versions, "/"))
			}
			log.Println(fmt.Sprintf("[scan %s finished]: %s", url, result))
		} else {
			log.Println(err)
		}
	},
}

func init() {
	scanCmd.Flags().StringVarP(&url, "url", "u", "", "scan target")
	rootCmd.AddCommand(scanCmd)
}
