package main

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func makeTaskWithPath(title, topic string, status TaskStatus) *TaskWithPath {
	return &TaskWithPath{
		Task:     &Task{Title: title, Status: status},
		Topic:    topic,
		FilePath: "testfile.md",
	}
}

func TestYankAndPaste(t *testing.T) {
	m := model{}
	task := makeTaskWithPath("TestTask", "", StatusTodo)
	m.items = []item{{task: task}}
	m.selected = 0
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m2m := m2.(model)
	if m2m.yankedTask == nil || m2m.yankedTask.Task.Title != "TestTask" {
		t.Errorf("Expected yanked task to be set")
	}
	m2m.yankedTask = task
	_, _ = m2m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Clean up any copied files (match slugified/lowercase names)
	tasksDir := ".tada/tasks"
	entries, _ := os.ReadDir(tasksDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Name()), "testtask-copy") {
			_ = os.Remove(tasksDir + "/" + entry.Name())
		}
	}
}

func TestStatusCycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tada-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &TaskWithPath{Task: &Task{Title: "TestTask", Status: StatusTodo}}
	_ = store.SaveTask("", task.Task)
	// Find the file path actually used
	loaded, _ := store.LoadAllTasks()
	var realTask *TaskWithPath
	for _, tasks := range loaded {
		for _, t := range tasks {
			if t.Task.Title == "TestTask" {
				realTask = t
			}
		}
	}
	if realTask == nil {
		t.Fatalf("Task not found after saving")
	}
	defer os.Remove(realTask.FilePath)
	m := model{tasks: map[string][]*TaskWithPath{"": {realTask}}, items: []item{{task: realTask}}, selected: 0}
	m.cycleTaskStatus(realTask, 1) // cycle forward
	loaded, _ = store.LoadAllTasks()
	var foundStatus TaskStatus
	for _, tasks := range loaded {
		for _, t := range tasks {
			if t.Task.Title == "TestTask" {
				foundStatus = t.Task.Status
			}
		}
	}
	if foundStatus != StatusInProgress {
		t.Errorf("Expected status to be in-progress after cycling forward, got %s", foundStatus)
	}
	m.cycleTaskStatus(realTask, -1) // cycle backward
	loaded2, _ := store.LoadAllTasks()
	var foundStatus2 TaskStatus
	for _, tasks := range loaded2 {
		for _, t := range tasks {
			if t.Task.Title == "TestTask" {
				foundStatus2 = t.Task.Status
			}
		}
	}
	if foundStatus2 != StatusTodo {
		t.Errorf("Expected status to be todo after cycling backward, got %s", foundStatus2)
	}
}

func TestDeleteTask(t *testing.T) {
	f, _ := os.CreateTemp("", "testfile-*.md")
	defer os.Remove(f.Name())
	task := &TaskWithPath{Task: &Task{Title: "DelTask"}, FilePath: f.Name()}
	m := model{items: []item{{task: task}}, selected: 0}
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, _ = m2.(model).updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		t.Errorf("Expected file to be deleted")
	}
}

func TestEditShortcut(t *testing.T) {
	m := model{}
	task := makeTaskWithPath("EditTask", "", StatusTodo)
	m.items = []item{{task: task}}
	m.selected = 0
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if m2.(model).mode != editView {
		t.Errorf("Expected to enter edit mode")
	}
}

func TestSearchMode(t *testing.T) {
	m := model{searchMode: false}
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m2m := m2.(model)
	if !m2m.searchMode {
		t.Errorf("Expected search mode to be enabled")
	}
	m2m.searchQuery = "Test"
	task := makeTaskWithPath("TestTask", "", StatusTodo)
	m2m.items = []item{{task: task, text: "○ TestTask"}}
	out := m2m.viewList()
	if !strings.Contains(out, "TestTask") {
		t.Errorf("Expected search results to include TestTask")
	}
}

func TestDetailsPopup(t *testing.T) {
	m := model{}
	task := makeTaskWithPath("DetailTask", "", StatusTodo)
	m.items = []item{{task: task}}
	m.selected = 0
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m2m := m2.(model)
	if !m2m.showDetails {
		t.Errorf("Expected details popup to be shown")
	}
	out := m2m.viewList()
	if !strings.Contains(out, "DetailTask") {
		t.Errorf("Expected details popup to show task title")
	}
}

func TestCompletedTaskVisibleUntilExit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tada-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	store := NewFileStore(tempDir)
	task := &TaskWithPath{Task: &Task{Title: "CompleteMe", Status: StatusTodo}}
	_ = store.SaveTask("", task.Task)
	// Find the file path actually used
	loaded, _ := store.LoadAllTasks()
	var realTask *TaskWithPath
	for _, tasks := range loaded {
		for _, t := range tasks {
			if t.Task.Title == "CompleteMe" {
				realTask = t
			}
		}
	}
	if realTask == nil {
		t.Fatalf("Task not found after saving")
	}
	defer os.Remove(realTask.FilePath)
	m := model{tasks: map[string][]*TaskWithPath{"": {realTask}}, selected: 0}
	m.buildItems()
	if len(m.items) == 0 || m.items[0].task.Task.Title != "CompleteMe" {
		t.Fatalf("Task should be visible before completion")
	}
	// Simulate cycling to done
	m.cycleTaskStatus(realTask, 1) // Todo -> InProgress
	m.cycleTaskStatus(realTask, 1) // InProgress -> Done
	m.toArchive = append(m.toArchive, realTask)
	m.buildItems()
	found := false
	for _, it := range m.items {
		if it.task != nil && it.task.Task.Title == "CompleteMe" {
			found = true
		}
	}
	if !found {
		t.Errorf("Completed task should still be visible until exit")
	}
	// Simulate TUI exit (archive clears toArchive)
	m.toArchive = nil
	m.buildItems()
	for _, it := range m.items {
		if it.task != nil && it.task.Task.Title == "CompleteMe" {
			t.Errorf("Completed task should be hidden after exit")
		}
	}
}
