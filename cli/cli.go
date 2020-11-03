package main

import (
	"fmt"
	"github.com/someu/netscan/appscan"
	"github.com/someu/netscan/portscan"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

var (
	packetPerSecond uint
	timeout         uint
	ipStr           string
	portStr         string
	urls            []string
)

var rootCmd = &cobra.Command{
	Use:   "netscan",
	Short: "app feature scanner",
	Run: func(cmd *cobra.Command, args []string) {
		if len(ipStr) > 0 && len(portStr) > 0 {
			ips := strings.Split(ipStr, ",")
			ports := strings.Split(portStr, ",")
			var ipSegs portscan.Segments
			var portSegs portscan.Segments
			for _, ip := range ips {
				seg, err := portscan.ParseIpSegment(ip)
				if err != nil {
					log.Panicf("Parse ip segment error: %s\n", err)
				}
				ipSegs = append(ipSegs, seg)
			}
			for _, port := range ports {
				seg, err := portscan.ParsePortSegment(port)
				if err != nil {
					log.Panicf("Parse port segment error: %s\n", err)
				}
				portSegs = append(portSegs, seg)
			}
			count := 0
			handle := func(res portscan.PortScanResult) {
				var url string
				if res.Port == 443 {
					url = fmt.Sprintf("https://%s:%d", res.IP.String(), res.Port)
				} else {
					url = fmt.Sprintf("http://%s:%d", res.IP.String(), res.Port)
				}
				urls = append(urls, url)
				log.Printf("%d %s:%d is opened", count, res.IP.String(), res.Port)
				count++
			}

			portScan, err := portscan.CreatePortScan(&portscan.PortScanConfig{
				IPSegments:      ipSegs,
				PortSegments:    portSegs,
				PacketPerSecond: packetPerSecond,
				Timeout:         time.Duration(timeout) * time.Second,
				Callback:        handle,
			})
			if err != nil {
				log.Panic(err)
			}
			portScan.Wait()
		}

		if len(urls) == 0 {
			log.Panic("no scan target")
		}

		var features []*appscan.Feature
		for _, feature := range appscan.Features {
			if feature.Path == "/" {
				features = append(features, feature)
			}
		}

		appScan, err := appscan.CreateScan(&appscan.AppScanConfig{
			Urls:        urls,
			Features:    features,
			ScanTimeout: time.Duration(timeout) * time.Second,
			Callback: func(result *appscan.AppScanResult) {
				if len(result.MatchedFeatures) > 0 {
					log.Println(result.String())
				}
			},
		})
		if err != nil {
			log.Panic(err)
		}
		appScan.Wait()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&ipStr, "ip", "i", "", "ip, split by ',', eg: 10.0.8.1,10.0.8.1/24")
	rootCmd.Flags().StringVarP(&portStr, "port", "p", "80,443", "port, split by ', eg: 80,80-8080")
	rootCmd.Flags().UintVarP(&packetPerSecond, "pps", "", 100, "packet per second")
	rootCmd.Flags().UintVarP(&timeout, "timeout", "", 10, "timeout, unit second")
	rootCmd.Flags().StringArrayVarP(&urls, "url", "u", nil, "url")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(rootCmd.Help())
	}
}
