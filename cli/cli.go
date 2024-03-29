package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/someu/netscan/appscan"
	"github.com/someu/netscan/portscan"
	"github.com/spf13/cobra"
)

var (
	packetPerSecond   uint
	appScanConcurrent uint
	timeout           uint
	ipStr             string
	portStr           string
	urls              []string
	input             string
	output            string
	outputTarget      string
)

var rootCmd = &cobra.Command{
	Use:   "netscan",
	Short: "use netscan to find the technology stack of any website",
	Run: func(cmd *cobra.Command, args []string) {
		if (len(ipStr) == 0 || len(portStr) == 0) && len(urls) == 0 && len(input) == 0 {
			cmd.Help()
			os.Exit(1)
		}

		var err error
		var outputFile *os.File
		var outputTargetFile *os.File

		if len(output) > 0 {
			outputFile, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer outputFile.Close()
		}
		if len(outputTarget) > 0 {
			outputTargetFile, err = os.OpenFile(outputTarget, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer outputTargetFile.Close()
		}

		if len(input) > 0 {
			file, err := os.OpenFile(input, os.O_RDONLY, 0666)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer file.Close()

			reader := bufio.NewReader(file)
			for {
				line, err := reader.ReadString('\n')
				url := strings.TrimRight(line, "\n")
				if err != nil && err.Error() != "EOF" {
					fmt.Println(err)
					os.Exit(1)
				} else if err != nil && err.Error() == "EOF" {
					break
				} else if len(url) > 0 {
					urls = append(urls, url)
				}
			}
		}

		if outputTargetFile != nil && len(urls) > 0 {
			outputTargetFile.WriteString(fmt.Sprintf("%s\n", strings.Join(urls, "\n")))
		}

		// port scan
		if len(ipStr) > 0 && len(portStr) > 0 {
			if err := portscan.InitPortScan(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			ips := strings.Split(ipStr, ",")
			ports := strings.Split(portStr, ",")
			var ipSegs portscan.Segments
			var portSegs portscan.Segments
			for _, ip := range ips {
				seg, err := portscan.ParseIpSegment(ip)
				if err != nil {
					fmt.Printf("Parse ip segment error: %s\n", err)
					os.Exit(1)
				}
				ipSegs = append(ipSegs, seg)
			}
			for _, port := range ports {
				seg, err := portscan.ParsePortSegment(port)
				if err != nil {
					fmt.Printf("Parse port segment error: %s\n", err)
					os.Exit(1)
				}
				portSegs = append(portSegs, seg)
			}

			handle := func(res portscan.PortScanResult) {
				var url string
				if res.Port == 443 {
					url = fmt.Sprintf("https://%s:%d", res.IP.String(), res.Port)
				} else {
					url = fmt.Sprintf("http://%s:%d", res.IP.String(), res.Port)
				}
				if outputTargetFile != nil {
					outputTargetFile.WriteString(fmt.Sprintf("%s\n", url))
				}
				urls = append(urls, url)
				fmt.Printf("%s:%d\n", res.IP.String(), res.Port)
			}

			portScan, err := portscan.CreatePortScan(&portscan.PortScanConfig{
				IPSegments:      ipSegs,
				PortSegments:    portSegs,
				PacketPerSecond: packetPerSecond,
				Timeout:         time.Duration(timeout) * time.Second,
				Callback:        handle,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			portScan.Wait()
		}

		if len(urls) == 0 {
			os.Exit(0)
		}

		// app scan
		if err := appscan.InitAppScan(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		appscan.SetRequestConcurrent(int(appScanConcurrent))
		var features []*appscan.Feature
		for _, feature := range appscan.Features {
			if feature.Path == "/" {
				features = append(features, feature)
			}
		}
		appScan, err := appscan.CreateScan(&appscan.AppScanConfig{
			Urls:     urls,
			Features: features,
			Callback: func(result *appscan.AppScanResult) {
				if len(result.MatchedFeatures) > 0 {
					if outputFile != nil {
						outputFile.WriteString(fmt.Sprintf("%s\n", result.String()))
					}
					fmt.Println(result.String())
				}
			},
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		appScan.Wait()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&ipStr, "ip", "i", "", "ip, split by ',', eg: 10.0.8.1,10.0.8.1/24. This parameter should be used in root permission")
	rootCmd.Flags().StringVarP(&portStr, "port", "p", "80,443", "port, split by ', eg: 80,80-8080")
	rootCmd.Flags().UintVarP(&packetPerSecond, "pps", "", 100, "packet per second send when use \"--ip\" or \"-i\" parameter")
	rootCmd.Flags().UintVarP(&timeout, "timeout", "", 10, "timeout, unit second")
	rootCmd.Flags().UintVarP(&appScanConcurrent, "request-concurrent", "", 100, "request concurrent")
	rootCmd.Flags().StringArrayVarP(&urls, "url", "u", nil, "url")
	rootCmd.Flags().StringVarP(&input, "input-file", "", "", "get target from a file")
	rootCmd.Flags().StringVarP(&outputTarget, "output-target-file", "", "", "output scan targets to a file")
	rootCmd.Flags().StringVarP(&output, "output-file", "", "", "output scan results to  a file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
