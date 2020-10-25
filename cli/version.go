package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"netscan/appscan"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  `Print the version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("FakeScan %s", appscan.Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
