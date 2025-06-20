package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAddTask_EmptyTitle_Error(t *testing.T) {
	m := model{editForm: form{title: ""}}
	m2, _ := m.addTask()
	if m2.(model).err == nil || !strings.Contains(m2.(model).err.Error(), "title cannot be empty") {
		t.Errorf("Expected error for empty title, got: %v", m2.(model).err)
	}
}

func TestSaveTask_FileWriteError(t *testing.T) {
	// Use a file path that cannot be written to
	m := model{editTask: &TaskWithPath{Task: &Task{Title: "Test"}, FilePath: "/root/forbidden.md"}, editForm: form{title: "Test"}}
	m2, _ := m.saveTask()
	if m2.(model).err == nil || !strings.Contains(m2.(model).err.Error(), "failed to save") {
		t.Errorf("Expected file write error, got: %v", m2.(model).err)
	}
}

func TestTUIErrorRendering(t *testing.T) {
	m := model{err: errors.New("something went wrong")}
	out := m.View()
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("Expected error message in View output, got: %s", out)
	}
}

func TestTUIArchiveOnQuit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tada-tui-archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	tasksDir := tempDir + "/tasks"
	os.MkdirAll(tasksDir, 0755)
	task := &Task{Title: "ArchiveMe", Status: StatusDone}
	// Write a valid markdown file with YAML frontmatter

	taskWithPath := &TaskWithPath{Task: task, Topic: "", FilePath: tasksDir + "/archiveme.md"}
	content := store.taskToMarkdown(task)
	os.WriteFile(taskWithPath.FilePath, []byte(content), 0644)
	m := model{toArchive: []*TaskWithPath{taskWithPath}, store: store}
	m.updateListView(keyMsg("q"))
	archiveDir := tempDir + "/archive"
	entries, _ := os.ReadDir(archiveDir)
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "archiveme") {
			found = true
		}
	}
	if !found {
		// Print debug info
		files, _ := os.ReadDir(archiveDir)
		var names []string
		for _, f := range files {
			names = append(names, f.Name())
		}
		t.Errorf("Expected archived file in archive directory, found: %v", names)
	}
}

// keyMsg is a helper to create a tea.KeyMsg for testing
func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
