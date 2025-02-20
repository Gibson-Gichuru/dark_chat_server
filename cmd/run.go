package cmd

import (
	"context"
	"darkchat/server"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the server and listen for incoming connections",
	Long:  "This command starts the server and listens for incoming connections. It takes optional flags for the address and port to listen on.",

	Run: func(cmd *cobra.Command, args []string) {

		serverAddress, _ := cmd.Flags().GetString("address")
		serverPort, _ := cmd.Flags().GetString("port")
		serverctx := context.Background()

		connectionBuilder := server.ConnectionBuilder{ConnectionType: "tcp", Address: serverAddress, Port: serverPort}

		server.ServerStart(serverctx, connectionBuilder)

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("address", "localhost", "The address to listen on")
	runCmd.Flags().String("port", "8080", "The port to listen on")
}
