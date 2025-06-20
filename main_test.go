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

	cfg := &Config{DefaultSort: "created", Theme: "dark"} // Provide a default config for test

	rootCmd := &cobra.Command{
		Use:   "tada",
		Short: "A terminal-based todo application",
		Long:  "A terminal-based todo application\n\nTada is a simple yet powerful todo application with both CLI and TUI interfaces",
	}

	rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store, cfg), NewCompleteCmd(store), NewTuiCmd(store))

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

		cfg := &Config{DefaultSort: "created", Theme: "dark"} // Provide a default config for test
		rootCmd.AddCommand(NewAddCmd(store), NewListCmd(store, cfg), NewCompleteCmd(store), NewTuiCmd(store))

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

// TestMain_NoTadaDir tests main error exit when .tada directory is missing
func TestMain_NoTadaDir(t *testing.T) {
	// Create a temp dir without .tada
	tempDir, err := os.MkdirTemp("", "tada-no-tada-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Save original stderr and replace with buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	exitCode := 0
	// Patch os.Exit to capture exit code
	origExit := osExit
	osExit = func(code int) { exitCode = code }
	defer func() { osExit = origExit }()

	main()
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !bytes.Contains([]byte(output), []byte("No .tada folder found")) {
		t.Errorf("Expected error message about missing .tada folder, got: %s", output)
	}
}

// TestMain_InvalidCommand tests main error exit on invalid command
func TestMain_InvalidCommand(t *testing.T) {
	testTadaDir, cleanup := setupTestEnv(t)
	defer cleanup()
	os.Setenv("TADA_DIR", testTadaDir)

	oldArgs := os.Args
	os.Args = []string{"tada", "notacommand"}
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	exitCode := 0
	origExit := osExit
	osExit = func(code int) { exitCode = code }
	defer func() { osExit = origExit }()

	main()
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid command, got %d", exitCode)
	}
	if !bytes.Contains([]byte(output), []byte("unknown command")) && !bytes.Contains([]byte(output), []byte("notacommand")) {
		t.Errorf("Expected error message about unknown command, got: %s", output)
	}
}
