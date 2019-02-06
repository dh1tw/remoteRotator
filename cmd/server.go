package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the serve command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "remoteRotator Server",
	Long: `Run a remoteRotator server

Start a remoteRotator server using a specific transport protocol.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please select the server type (--help for available options)")
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
