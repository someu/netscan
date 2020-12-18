package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "netscan",
	Short: "use netscan to find the technology stack of any website",
}
