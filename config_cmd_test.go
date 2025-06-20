package main

import (
	"os"
	"strings"
	"testing"
)

func TestConfigCmd_Show(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-config-test-*")
	defer os.RemoveAll(tempDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	cfg := &Config{DefaultSort: "priority", Theme: "dark"}
	saveConfig(cfg, true)
	cmd := NewConfigCmd()
	cmd.SetArgs([]string{"show"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "priority") || !strings.Contains(out.String(), "dark") {
		t.Errorf("Expected config output, got: %s", out.String())
	}
}

func TestConfigCmd_Set(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-config-test-*")
	defer os.RemoveAll(tempDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	cmd := NewConfigCmd()
	cmd.SetArgs([]string{"set", "theme", "light", "--global"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Config updated") {
		t.Errorf("Expected config updated message, got: %s", out.String())
	}
}
