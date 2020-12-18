package cmd

import "github.com/spf13/cobra"

var (
	listenHost  string
	listenPort  string
	redisUrl    string
	mongodbUrl  string
	pushBackUrl string
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start a net scan server",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	apiCmd.Flags().StringVarP(&listenHost, "listen-host", "", "127.0.0.1", "set which add the server bind")
	apiCmd.Flags().StringVarP(&listenPort, "listen-port", "", "8989", "set port to bind")
	apiCmd.Flags().StringVarP(&redisUrl, "redis", "", "", "")
	apiCmd.Flags().StringVarP(&mongodbUrl, "mongo", "", "", "")
	apiCmd.Flags().StringVarP(&pushBackUrl, "push-back", "", "", "")
	RootCmd.AddCommand(apiCmd)
}
