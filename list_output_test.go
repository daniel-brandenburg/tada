package main

import (
	"os"
	"strings"
	"testing"
)

func TestListCmd_OutputFormats(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-list-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "FormatMe", Status: StatusTodo}
	store.SaveTask("", task)
	cfg := &Config{}
	cmd := NewListCmd(store, cfg)
	cmd.SetArgs([]string{"--output", "json"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "FormatMe") || !strings.Contains(out.String(), "Title") {
		t.Errorf("Expected JSON output, got: %s", out.String())
	}

	cmd = NewListCmd(store, cfg)
	cmd.SetArgs([]string{"--output", "yaml"})
	out.Reset()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "FormatMe") || !strings.Contains(out.String(), "title:") {
		t.Errorf("Expected YAML output, got: %s", out.String())
	}
}
