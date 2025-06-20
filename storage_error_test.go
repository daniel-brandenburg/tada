package main

import (
	"os"
	"testing"
)

func TestFileStore_SaveTask_Error(t *testing.T) {
	fs := NewFileStore("/root/forbidden-dir")
	task := &Task{Title: "Test", Status: StatusTodo}
	err := fs.SaveTask("", task)
	if err == nil {
		t.Errorf("Expected error when saving to forbidden directory")
	}
}

func TestFileStore_LoadAllTasks_Error(t *testing.T) {
	fs := NewFileStore("/root/forbidden-dir")
	_, err := fs.LoadAllTasks()
	if err == nil {
		t.Errorf("Expected error when loading from forbidden directory")
	}
}

func TestFileStore_CompleteTask_NotFound(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tada-storage-test-*")
	defer os.RemoveAll(tempDir)
	fs := NewFileStore(tempDir)
	err := fs.CompleteTask("", "nonexistent")
	if err == nil || err.Error() != "task not found: nonexistent" {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}
