package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var osExit = os.Exit

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
	return "", fmt.Errorf(".tada folder not found") /*  */
}

func onboardingMessage() string {
	return `
Welcome to Tada! 🎉

- Use 'tada tui' for the interactive terminal UI.
- Use 'tada list', 'tada add', etc. for CLI commands.
- Run 'tada completion [bash|zsh|fish|powershell]' to enable shell completions.
- Run 'tada config show' to see or edit your config (global or per-project).
- See the README for more tips!

Happy tasking!
`
}

func main() {
	if os.Getenv("TADA_TEST_NO_TUI") == "1" {
		fmt.Fprintln(os.Stderr, "TUI disabled for test")
		osExit(0)
	}
	cwd, _ := os.Getwd()
	tadaDir, err := findTadaDir(cwd)
	if err != nil {
		styledErr := lipgloss.NewStyle().Foreground(cliError).Render("No .tada folder found in this or any parent directory.")
		fmt.Fprintln(os.Stderr, styledErr)
		osExit(1)
	}
	store := NewFileStore(tadaDir)
	cfg, _ := loadConfig()
	var rootCmd = &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "A terminal-based todo application\n\nTada is a simple yet powerful todo application with both CLI and TUI interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			RunTUIWithConfig(cfg)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = false

	rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store, cfg), NewCompleteCmd(store), NewTuiCmd(store), NewEditCmd(store), NewDeleteCmd(store), NewShowCmd(store), NewMoveCmd(store), NewCopyCmd(store), NewBulkCmd(store), NewConfigCmd(), NewVersionCmd())

	if err := fang.Execute(context.TODO(), rootCmd); err != nil {
		osExit(1)
	}
}
