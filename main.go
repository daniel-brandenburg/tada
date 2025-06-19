package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func findTadaDir(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".tada")); err == nil {
			return filepath.Join(dir, ".tada"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(".tada folder not found")
}

func main() {
	cwd, _ := os.Getwd()
	tadaDir, err := findTadaDir(cwd)
	if err != nil {
		fmt.Println("No .tada folder found in this or any parent directory.")
		os.Exit(1)
	}
	store := NewFileStore(tadaDir)
	var rootCmd = &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "A terminal-based todo application\n\nTada is a simple yet powerful todo application with both CLI and TUI interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			RunTUI()
		},
	}

	rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store), NewCompleteCmd(store), NewTuiCmd(store), NewEditCmd(store), NewDeleteCmd(store))

	if err := fang.Execute(context.TODO(), rootCmd); err != nil {
		os.Exit(1)
	}
}
