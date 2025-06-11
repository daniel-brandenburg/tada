package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "Tada is a simple yet powerful todo application with both CLI and TUI interfaces",
	}

	rootCmd.AddCommand(addCmd, listCmd, completeCmd, tuiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
