package main

import (
	"fakescan/scanner"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	ips   []string
	ports []string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "start a scan",
	Long:  `start a scan`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(ips) == 0 || len(ports) == 0 {
			log.Fatalln("no target")
		}
		log.Println("Start scan", ips, ports)
		wg := globalScanner.Scan(ips, ports, func(result *scanner.MatchedResult) {
			spend := float64(result.EndAt.UnixNano()-result.StartAt.UnixNano()) / (1000 * 1000)
			var appsStr string
			for _, app := range result.Apps {
				appsStr += strings.TrimSpace(fmt.Sprintf("%s %s, ", app.Name, strings.Join(app.Versions, ", ")))
			}
			log.Println(fmt.Sprintf("[scan %s finished] spend %f ms, result: %s", result.Url, spend, appsStr))
		})
		wg.Wait()
		log.Println("Finished scan")
	},
}

func init() {
	scanCmd.Flags().StringArrayVarP(&ips, "ips", "i", nil, "scan ips")
	scanCmd.Flags().StringArrayVarP(&ports, "ports", "p", nil, "scan ports")
	rootCmd.AddCommand(scanCmd)
}
