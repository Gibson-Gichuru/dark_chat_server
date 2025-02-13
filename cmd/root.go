package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "darkchat",
	Short: "This is a chat server",
	Long:  "This is a chat server",

	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal("Please use a subcommand")
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}
