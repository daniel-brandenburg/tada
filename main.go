package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	store := NewFileStore()
	var rootCmd = &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "A terminal-based todo application\n\nTada is a simple yet powerful todo application with both CLI and TUI interfaces",
	}

	rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store), NewCompleteCmd(store), NewTuiCmd(store))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
