package main

import (
	"errors"
	"fmt"
	"github.com/someu/netscan/portscan"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var (
	packetPerSecond uint
	portScanTimeout uint
	portScanIpStr   string
	portScanPortStr string
)

var portScan = &cobra.Command{
	Use:   "portscan",
	Short: "scan target's open port",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(portScanIpStr) == 0 {
			return errors.New("ip is required")
		}
		if len(portScanPortStr) == 0 {
			return errors.New("port is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		scanner, err := portscan.NewPortScanner()
		if err != nil {
			fmt.Printf("Create port scanner error: %s\n", err)
			os.Exit(1)
		}
		ips := strings.Split(portScanIpStr, ",")
		ports := strings.Split(portScanPortStr, ",")
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

		scan := scanner.CreatePortScan(portscan.NewTargetRange(ipSegs, portSegs), &portscan.PortScanConfig{
			PacketPerSecond: packetPerSecond,
			Timeout:         time.Duration(timeout) * time.Second,
		})
		scan.Wait()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&portScanIpStr, "ip", "i", "140.143.0.136/24", "ip, split by ',', eg: 80,80-8080")
	rootCmd.Flags().StringVarP(&portScanPortStr, "port", "p", "27222", "port, split by ', eg: 10.0.8.1,10.0.8.1/24")
	rootCmd.Flags().UintVarP(&packetPerSecond, "pps", "", 100, "packet per second")
	rootCmd.Flags().UintVarP(&portScanTimeout, "timeout", "", 10, "timeout, unit second")
	rootCmd.AddCommand(portScan)
}
