package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd tests the root command
func TestRootCmd(t *testing.T) {
	var buf bytes.Buffer

	testTadaDir, cleanup := setupTestEnv(t)
	defer cleanup()
	store := NewFileStore(testTadaDir)

	rootCmd := &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "A terminal-based todo application\n\nTada is a simple yet powerful todo application with both CLI and TUI interfaces",
	}

	rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store), NewCompleteCmd(store), NewTuiCmd(store))

	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Error executing root command: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"A terminal-based todo application",
		"Available Commands:",
		"add",
		"complete",
		"list",
		"tui",
		"help",
	}

	for _, phrase := range expectedPhrases {
		if !bytes.Contains(buf.Bytes(), []byte(phrase)) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", phrase, output)
		}
	}
}

// TestMain tests the main function
func TestMainFunction(t *testing.T) {
	// Save original args
	oldArgs := os.Args

	// Restore original args when done
	defer func() {
		os.Args = oldArgs
	}()

	// Set args to test help
	os.Args = []string{"tada", "--help"}

	// Save original stdout and replace it with a buffer
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout when done
	defer func() {
		os.Stdout = oldStdout
	}()

	testTadaDir, cleanup := setupTestEnv(t)
	defer cleanup()
	store := NewFileStore(testTadaDir)

	exitCh := make(chan int, 1)
	go func() {
		var rootCmd = &cobra.Command{
			Use:   "tada",
			Short: "A terminal-based todo application",
			Long:  "A terminal-based todo application\n\nTada is a simple yet powerful todo application with both CLI and TUI interfaces",
		}

		rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store), NewCompleteCmd(store), NewTuiCmd(store))

		if err := rootCmd.Execute(); err != nil {
			exitCh <- 1
			return
		}
		exitCh <- 0
	}()

	// Wait for exit
	exitCode := <-exitCh
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
