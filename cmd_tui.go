package main

import (
	"github.com/spf13/cobra"
)

func NewTuiCmd(store *FileStore) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Start the TUI interface",
		Long:  "Start the interactive terminal user interface",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := loadConfig()
			tadaDir := ""
			if store != nil {
				tadaDir = store.basePath
			}
			showWelcomeIfNeeded(cfg, tadaDir)
			RunTUI()
		},
	}
}
