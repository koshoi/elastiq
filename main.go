package main

import (
	"fmt"
	"os"

	"github.com/koshoi/elastiq/commands"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "elastiq",
		Short: "CLI for ElastricSearch",
	}

	commands.AddQueryCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
