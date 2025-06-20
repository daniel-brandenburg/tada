package main

import (
	"os"
	"strings"
	"testing"
)

func TestStatsCmd_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-stats-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "StatsMe", Status: StatusTodo, Tags: []string{"foo"}}
	store.SaveTask("", task)
	cmd := NewStatsCmd(store)
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	output := out.String()
	if !strings.Contains(output, "Task Statistics") || !strings.Contains(output, "foo: 1") || !strings.Contains(output, "todo: 1") {
		t.Errorf("Expected stats output with correct tag and status counts, got: %s", output)
	}
}

func TestShowCmd_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-show-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "ShowMe", Status: StatusTodo, Description: "desc"}
	store.SaveTask("", task)
	cmd := NewShowCmd(store)
	cmd.SetArgs([]string{"ShowMe"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "ShowMe") || !strings.Contains(out.String(), "desc") {
		t.Errorf("Expected show output, got: %s", out.String())
	}
}
