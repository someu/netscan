package main

import (
	"fakescan/scanner"
	"fakescan/util"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	urls   []string
	inputs []string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "start a scan",
	Long:  `start a scan`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("[start scan]", strings.Join(urls, ", "))
		var targets []string
		for _, input := range inputs {
			if lines, err := util.ReadFileLines(input); err != nil {
				log.Panic("[read file error]", err)
			} else {
				for _, url := range lines {
					if len(strings.TrimSpace(url)) != 0 {
						urls = append(urls, strings.TrimSpace(url))
					}
				}
			}
		}
		for _, url := range urls {
			if util.IsCIDR(url) {
				if ips, err := util.CIDRToIpList(url); err == nil {
					targets = append(targets, ips...)
				} else {
					log.Panic("[unrecognized cidr]", url)
				}
			} else {
				targets = append(targets, url)
			}
		}
		wg := globalScanner.ScanUrls(targets, func(result *scanner.MatchedResult) {
			spend := float64(result.EndAt.UnixNano()-result.StartAt.UnixNano()) / (1000 * 1000)
			var appsStr string
			for _, app := range result.Apps {
				appsStr += strings.TrimSpace(fmt.Sprintf("%s %s, ", app.Name, strings.Join(app.Versions, ", ")))
			}
			log.Println(fmt.Sprintf("[scan %s finished] spend %f ms, result: %s", result.Url, spend, appsStr))
		})
		wg()
		log.Println("[finished scan]")
	},
}

func init() {
	scanCmd.Flags().StringArrayVarP(&urls, "url", "u", nil, "scan target")
	scanCmd.Flags().StringArrayVarP(&inputs, "input", "i", nil, "scan target")
	rootCmd.AddCommand(scanCmd)
}
