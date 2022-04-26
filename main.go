package main

import (
	"os"

	"elastiq/commands"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:          "elastiq",
		Short:        "CLI for ElastricSearch",
		SilenceUsage: true,
	}

	commands.AddQueryCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
