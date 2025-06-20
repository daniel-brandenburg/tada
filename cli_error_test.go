package main

import (
	"os"
	"strings"
	"testing"
)

func TestEditCmd_TaskNotFound_Error(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	cmd := NewEditCmd(store)
	cmd.SetArgs([]string{"nonexistent"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task not found") {
		t.Errorf("Expected 'Task not found' error, got: %s", out.String())
	}
}

func TestDeleteCmd_TaskNotFound_Error(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	cmd := NewDeleteCmd(store)
	cmd.SetArgs([]string{"nonexistent"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task not found") {
		t.Errorf("Expected 'Task not found' error, got: %s", out.String())
	}
}

func TestMoveCmd_TaskNotFound_Error(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	cmd := NewMoveCmd(store)
	cmd.SetArgs([]string{"nonexistent", "newtopic"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task not found") {
		t.Errorf("Expected 'Task not found' error, got: %s", out.String())
	}
}

func TestCopyCmd_TaskNotFound_Error(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	cmd := NewCopyCmd(store)
	cmd.SetArgs([]string{"nonexistent", "newtopic"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task not found") {
		t.Errorf("Expected 'Task not found' error, got: %s", out.String())
	}
}

func TestBulkCmd_NoMatch_Error(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	cmd := NewBulkCmd(store)
	cmd.SetArgs([]string{"--delete", "--search", "nonexistent"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "No matching tasks found") {
		t.Errorf("Expected 'No matching tasks found' error, got: %s", out.String())
	}
}
