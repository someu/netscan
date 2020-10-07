package main

import (
	"fakescan/scanner"
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  `Print the version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("FakeScan %s", scanner.Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
