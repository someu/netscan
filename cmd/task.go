package cmd

import (
	"bufio"
	"fmt"
	"github.com/someu/netscan/appscan"
	"github.com/someu/netscan/portscan"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var (
	packetPerSecond   uint     // --pps
	appScanConcurrent uint     // --thread
	timeout           uint     // --timeout
	ipStr             string   // -i, --ip
	portStr           string   // -p, --port
	urls              []string // -u, --url
	input             string   // --file
	outputListFile    string   // --output-list
	outputJSONFile    string   // --output-json
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "task mode",
	Run: func(cmd *cobra.Command, args []string) {

		if (len(ipStr) == 0 || len(portStr) == 0) && len(urls) == 0 && len(input) == 0 {
			cmd.Help()
			os.Exit(1)
		}

		var err error
		var outputFile *os.File
		var outputTargetFile *os.File

		//if len(output) > 0 {
		//	outputFile, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0666)
		//	if err != nil {
		//		fmt.Println(err)
		//		os.Exit(1)
		//	}
		//	defer outputFile.Close()
		//}
		//if len(outputTarget) > 0 {
		//	outputTargetFile, err = os.OpenFile(outputTarget, os.O_CREATE|os.O_WRONLY, 0666)
		//	if err != nil {
		//		fmt.Println(err)
		//		os.Exit(1)
		//	}
		//	defer outputTargetFile.Close()
		//}

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
	taskCmd.Flags().UintVarP(&packetPerSecond, "pps", "", 100, "packet per second send in syn port scan stage")
	taskCmd.Flags().UintVarP(&appScanConcurrent, "thread", "", 100, "request thread in grab stage")
	taskCmd.Flags().UintVarP(&timeout, "timeout", "", 10, "request timeout")
	taskCmd.Flags().StringVarP(&ipStr, "ip", "i", "", "ip")
	taskCmd.Flags().StringVarP(&portStr, "port", "p", "", "port")
	taskCmd.Flags().StringArrayVarP(&urls, "url", "u", nil, "urls")
	taskCmd.Flags().StringVarP(&input, "input", "", "", "input file")
	taskCmd.Flags().StringVarP(&outputListFile, "output-list", "", "", "output list")
	taskCmd.Flags().StringVarP(&outputJSONFile, "output-json", "", "", "output json")
	RootCmd.AddCommand(taskCmd)
}
