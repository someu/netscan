package main

import (
	"github.com/someu/netscan/portscan"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"
)

var packetPerSecond uint
var timeout uint
var ipStr string
var portStr string

func parseArgs() {
	var rootCmd = &cobra.Command{
		Use:   "portscan",
		Short: "portscan is a port scan based syn",
		Run: func(cmd *cobra.Command, args []string) {
			if len(ipStr) == 0 || len(portStr) == 0 {
				cmd.Help()
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&ipStr, "ip", "i", "140.143.0.136/24", "ip, split by ',', eg: 80,80-8080")
	rootCmd.Flags().StringVarP(&portStr, "port", "p", "80", "port, split by ', eg: 10.0.8.1,10.0.8.1/24")
	rootCmd.Flags().UintVarP(&packetPerSecond, "pps", "", 100, "packet per second")
	rootCmd.Flags().UintVarP(&timeout, "timeout", "", 10, "timeout, unit second")

	if err := rootCmd.Execute(); err != nil {
		rootCmd.Help()
		os.Exit(1)
	}
}

func scan() {
	ips := strings.Split(ipStr, ",")
	ports := strings.Split(portStr, ",")
	var ipSegs portscan.Segments
	var portSegs portscan.Segments
	for _, ip := range ips {
		seg, err := portscan.ParseIpSegment(ip)
		if err != nil {
			log.Fatalf("Parse ip segment error: %s\n", err)
		}
		ipSegs = append(ipSegs, seg)
	}
	for _, port := range ports {
		seg, err := portscan.ParsePortSegment(port)
		if err != nil {
			log.Fatalf("Parse port segment error: %s\n", err)
		}
		portSegs = append(portSegs, seg)
	}

	scan, err := portscan.CreatePortScan(&portscan.PortScanConfig{
		PacketPerSecond: packetPerSecond,
		Timeout:         time.Duration(timeout) * time.Second,
		IPSegments:      ipSegs,
		PortSegments:    portSegs,
	})
	if err != nil {
		log.Panic(err)
	} else {
		log.Println("start port scan")
	}

	scan2IPortSeg, err := portscan.ParsePortSegment("22")
	if err != nil {
		log.Panic(err)
	} else {
		log.Println("start port scan")
	}
	scan2, err := portscan.CreatePortScan(&portscan.PortScanConfig{
		PacketPerSecond: packetPerSecond,
		Timeout:         time.Duration(timeout) * time.Second,
		IPSegments:      ipSegs,
		PortSegments:    portscan.Segments{scan2IPortSeg},
	})
	if err != nil {
		log.Panic(err)
	} else {
		log.Println("start port scan")
	}
	scan2.Wait()
	scan.Wait()
}

func main() {
	parseArgs()
	scan()
}
