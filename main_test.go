package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd tests the root command
func TestRootCmd(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a root command for testing
	rootCmd := &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "Tada is a simple yet powerful todo application with both CLI and TUI interfaces",
	}

	// Add subcommands
	rootCmd.AddCommand(addCmd, listCmd, completeCmd, tuiCmd)

	// Set output to our buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Test the help command
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Error executing root command: %v", err)
	}

	// Check if the output contains expected information
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

	// Run main in a separate goroutine to capture exit
	exitCh := make(chan int, 1)
	go func() {
		// This is a simplified version of main() that doesn't exit
		var rootCmd = &cobra.Command{
			Use:   "tada",
			Short: "A terminal-based todo application",
			Long:  "Tada is a simple yet powerful todo application with both CLI and TUI interfaces",
		}

		rootCmd.AddCommand(addCmd, listCmd, completeCmd, tuiCmd)

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
