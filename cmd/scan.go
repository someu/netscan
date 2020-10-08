package main

import (
	"fakescan/scanner"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var (
	ip          string
	port        string
	masscanPath string
	masscanRate int
	level       int
	timeout     int
	concurrent  int
	input       string
	output      string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "start a scan",
	Long:  `start a scan`,
	Run: func(cmd *cobra.Command, args []string) {
		startAt := time.Now()
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
		if concurrent > 0 {
			globalScanner.SetConcurrent(concurrent)
		}

		var file *os.File
		if len(output) != 0 {
			var err error
			if file, err = os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0666); err != nil {
				log.Fatalf("Open file failed, %s", err.Error())
			}
			defer file.Close()
		}
		ips := strings.Split(ip, ",")

		if len(input) != 0 {
			var inputFile *os.File
			var err error
			if inputFile, err = os.OpenFile(input, os.O_RDONLY, 0666); err != nil {
				log.Fatalf("Open input file failed, %s", err.Error())
			}
			defer inputFile.Close()

			var inputContent []byte
			if inputContent, err = ioutil.ReadAll(inputFile); err != nil {
				log.Fatalf("Read input file failed, %s", err.Error())
			}
			ips = append(ips, scanner.ParseMassScanResult(string(inputContent))...)
		}

		ports := strings.Split(port, ",")
		if len(ips) == 0 || len(port) == 0 {
			fmt.Println("No scan target")
			cmd.Help()
			return
		}

		log.Println("Start scan", ip, port)

		var i int
		wg := globalScanner.Scan(ips, ports, func(result *scanner.MatchedResult) {
			var appsStr string
			for _, app := range result.Apps {
				appsStr += strings.TrimSpace(fmt.Sprintf("%s %s, ", app.Name, strings.Join(app.Versions, ", ")))
			}
			if len(appsStr) != 0 {
				if file != nil {
					file.WriteString(result.String())
				} else {
					log.Println(fmt.Sprintf("[ %d scan %s finished] %s", i, result.Url, appsStr))
				}
				i++
			}
		})
		wg.Wait()
		log.Printf("Finished scan，spend %d s", int(time.Now().UnixNano()-startAt.UnixNano())/(1000*1000*1000))
	},
}

func init() {
	scanCmd.Flags().StringVarP(&ip, "ip", "i", "", "scan ips, multi target split by ','")
	scanCmd.Flags().StringVarP(&port, "port", "p", "80", "scan ports, multi target split by ','")
	scanCmd.Flags().IntVarP(&masscanRate, "masscanRate", "r", 1000, "masscan rate")
	scanCmd.Flags().StringVarP(&masscanPath, "masscanPath", "m", "masscan", "masscan path")
	scanCmd.Flags().IntVarP(&level, "level", "l", 1, "web finger match level, set -1 to use all level")
	scanCmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "request timeout, set 0 to never timeout")
	scanCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 100, "scan concurrent")
	scanCmd.Flags().StringVarP(&output, "output", "o", "", "output file")
	scanCmd.Flags().StringVarP(&input, "input", "", "", "input masscan result file")
	rootCmd.AddCommand(scanCmd)
}
