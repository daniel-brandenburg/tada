package main

import (
	"os"
	"strings"
	"testing"
)

func TestEditCmd_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	// Create a task file
	task := &Task{Title: "EditMe", Status: StatusTodo}
	store.SaveTask("", task)
	cmd := NewEditCmd(store)
	cmd.SetArgs([]string{"EditMe", "--description", "Updated!", "--priority", "2", "--tags", "foo,bar", "--status", "done"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task updated") {
		t.Errorf("Expected success message, got: %s", out.String())
	}
}

func TestDeleteCmd_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "DeleteMe", Status: StatusTodo}
	store.SaveTask("", task)
	cmd := NewDeleteCmd(store)
	cmd.SetArgs([]string{"DeleteMe"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task deleted") {
		t.Errorf("Expected success message, got: %s", out.String())
	}
}

func TestMoveCmd_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "MoveMe", Status: StatusTodo}
	store.SaveTask("", task)
	cmd := NewMoveCmd(store)
	cmd.SetArgs([]string{"MoveMe", "NewTopic"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task moved to topic: NewTopic") {
		t.Errorf("Expected success message, got: %s", out.String())
	}
}

func TestCopyCmd_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "CopyMe", Status: StatusTodo}
	store.SaveTask("", task)
	cmd := NewCopyCmd(store)
	cmd.SetArgs([]string{"CopyMe", "NewTopic"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Task copied to topic: NewTopic") {
		t.Errorf("Expected success message, got: %s", out.String())
	}
}

func TestBulkCmd_Complete_Success(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-cli-test-*")
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &Task{Title: "BulkMe", Status: StatusTodo}
	store.SaveTask("", task)
	cmd := NewBulkCmd(store)
	cmd.SetArgs([]string{"--complete", "--search", "BulkMe"})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute()
	if !strings.Contains(out.String(), "Bulk operation complete") {
		t.Errorf("Expected success message, got: %s", out.String())
	}
}
