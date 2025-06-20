package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestOnboardingMessage(t *testing.T) {
	msg := onboardingMessage()
	if !strings.Contains(msg, "tada tui") || !strings.Contains(msg, "completion") {
		t.Errorf("Onboarding message missing expected content")
	}
}

func TestShellCompletion(t *testing.T) {
	if _, err := os.Stat("./tada"); os.IsNotExist(err) {
		t.Skip("tada binary not found; skipping shell completion test")
	}
	cmd := exec.Command("./tada", "completion", "bash")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Errorf("completion command failed: %v", err)
	}
	if !strings.Contains(out.String(), "_tada_completion") {
		t.Errorf("completion output missing expected function")
	}
}
