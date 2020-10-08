package main

import (
	"fakescan/scanner"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

var (
	ip          []string
	port        []string
	masscanPath string
	masscanRate int
	level       int
	timeout     int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "start a scan",
	Long:  `start a scan`,
	Run: func(cmd *cobra.Command, args []string) {
		startAt := time.Now()
		if len(ip) == 0 || len(port) == 0 {
			fmt.Println("No scan target")
			cmd.Help()
			return
		}
		log.Println("Start scan", ip, port)
		if masscanRate > 0 {
			globalScanner.MasscanRate = masscanRate
		}
		if len(masscanPath) > 0 {
			globalScanner.MasscanPath = masscanPath
		}
		if level > 0 {
			globalScanner.Level = level
		} else if level == -1 {
			globalScanner.Level = len(globalScanner.FeatureCollection)
		}
		if timeout >= 0 {
			globalScanner.SetTimeout(timeout)
		}

		wg := globalScanner.Scan(ip, port, func(result *scanner.MatchedResult) {
			spend := float64(result.EndAt.UnixNano()-result.StartAt.UnixNano()) / (1000 * 1000)
			var appsStr string
			for _, app := range result.Apps {
				appsStr += strings.TrimSpace(fmt.Sprintf("%s %s, ", app.Name, strings.Join(app.Versions, ", ")))
			}
			log.Println(fmt.Sprintf("[scan %s finished] spend %f ms, result: %s", result.Url, spend, appsStr))
		})
		wg.Wait()
		log.Printf("Finished scan，spend %d s", int(time.Now().UnixNano()-startAt.UnixNano())/(1000*1000*1000))
	},
}

func init() {
	scanCmd.Flags().StringArrayVarP(&ip, "ip", "i", nil, "scan ips")
	scanCmd.Flags().StringArrayVarP(&port, "port", "p", []string{"80"}, "scan ports")
	scanCmd.Flags().IntVarP(&masscanRate, "masscanRate", "r", 1000, "masscan rate")
	scanCmd.Flags().StringVarP(&masscanPath, "masscanPath", "m", "masscan", "masscan path")
	scanCmd.Flags().IntVarP(&level, "level", "l", 1, "web finger match level")
	scanCmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "request timeout")
	rootCmd.AddCommand(scanCmd)
}
