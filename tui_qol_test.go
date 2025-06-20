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

func TestUndoDeleteTask(t *testing.T) {
	f, _ := os.CreateTemp("", "testfile-*.md")
	defer os.Remove(f.Name())
	task := &TaskWithPath{Task: &Task{Title: "UndoDelTask"}, FilePath: f.Name()}
	m := model{items: []item{{task: task}}, selected: 0}
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m3, _ := m2.(model).updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		t.Errorf("Expected file to be deleted")
	}
	// Now undo
	m4, _ := m3.(model).updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	// File should be restored
	if _, err := os.Stat(f.Name()); err != nil {
		t.Errorf("Expected file to be restored after undo, got error: %v", err)
	}
	if !strings.Contains(m4.(model).undoMsg, "restored") {
		t.Errorf("Expected undo message after undo, got: %s", m4.(model).undoMsg)
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

func TestBulkSelectAndDelete(t *testing.T) {
	f1, _ := os.CreateTemp("", "testfile1-*.md")
	f2, _ := os.CreateTemp("", "testfile2-*.md")
	defer os.Remove(f1.Name())
	defer os.Remove(f2.Name())

	t1 := &TaskWithPath{Task: &Task{Title: "Bulk1"}, FilePath: f1.Name()}
	t2 := &TaskWithPath{Task: &Task{Title: "Bulk2"}, FilePath: f2.Name()}
	tasks := map[string][]*TaskWithPath{"": {t1, t2}}
	m := model{tasks: tasks, items: []item{{task: t1}, {task: t2}}, selected: 0}
	m.buildItems()
	m.selectedItems = map[int]struct{}{0: {}, 1: {}}

	// Debug: print file paths and selection
	t.Logf("File 1: %s, File 2: %s", f1.Name(), f2.Name())
	for idx, it := range m.items {
		if it.task != nil {
			t.Logf("Item %d: %s", idx, it.task.FilePath)
		}
	}
	m.selectedItems = map[int]struct{}{0: {}, 1: {}}
	if len(m.selectedItems) != 2 {
		t.Errorf("Expected 2 items selected, got %d", len(m.selectedItems))
	}
	// Bulk delete
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = m2.(model)
	// Debug: check file existence
	if _, err := os.Stat(f1.Name()); err == nil {
		t.Errorf("Expected file 1 to be deleted, but it exists")
	} else {
		t.Logf("File 1 deleted as expected: %v", err)
	}
	if _, err := os.Stat(f2.Name()); err == nil {
		t.Errorf("Expected file 2 to be deleted, but it exists")
	} else {
		t.Logf("File 2 deleted as expected: %v", err)
	}
}

func TestBulkExportSelectedTasks(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tada-export-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	f1, _ := os.CreateTemp(tempDir, "task1-*.md")
	f2, _ := os.CreateTemp(tempDir, "task2-*.md")
	t1 := &TaskWithPath{Task: &Task{Title: "Export1", Description: "Desc1", Priority: 2, Status: StatusTodo}, FilePath: f1.Name()}
	t2 := &TaskWithPath{Task: &Task{Title: "Export2", Description: "Desc2", Priority: 3, Status: StatusInProgress}, FilePath: f2.Name()}
	tasks := map[string][]*TaskWithPath{"": {t1, t2}}
	m := model{tasks: tasks, items: []item{{task: t1}, {task: t2}}, selected: 0}
	m.buildItems()
	m.selectedItems = map[int]struct{}{0: {}, 1: {}}

	// Simulate pressing 'x' to start export
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = m2.(model)
	if m.exportPrompt == nil {
		t.Fatalf("Expected export prompt to be active after 'x' key")
	}

	// Simulate choosing JSON format ('2')
	m2, _ = m.updateExportPrompt(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = *m2.(*model)
	if m.exportPrompt == nil || m.exportPrompt.format != "json" {
		t.Fatalf("Expected export format to be set to json")
	}

	// Simulate entering file path and pressing enter
	exportPath := tempDir + "/bulk_export.json"
	for _, ch := range exportPath {
		m2, _ = m.updateExportPrompt(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = *m2.(*model)
	}
	m2, _ = m.updateExportPrompt(tea.KeyMsg{Type: tea.KeyEnter})
	m = *m2.(*model)

	// Check export message
	if !strings.Contains(m.exportMsg, "Exported selected tasks") {
		t.Errorf("Expected export success message, got: %s", m.exportMsg)
	}

	// Check file exists and content is valid JSON
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}
	if !strings.Contains(string(data), "Export1") || !strings.Contains(string(data), "Export2") {
		t.Errorf("Exported file missing expected task data: %s", string(data))
	}
}

func TestRangeSelectionVimStyle(t *testing.T) {
	t1 := makeTaskWithPath("Task1", "", StatusTodo)
	t2 := makeTaskWithPath("Task2", "", StatusTodo)
	t3 := makeTaskWithPath("Task3", "", StatusTodo)
	items := []item{{task: t1}, {task: t2}, {task: t3}}
	m := model{items: items, selected: 0, selectedItems: make(map[int]struct{}), lastSelect: -1}

	// Range select down (0 to 2)
	m.selected = 0
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	m = m2.(model)
	m.selected = 2
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	m = m2.(model)
	if len(m.selectedItems) != 3 {
		t.Errorf("Expected 3 items selected, got %d", len(m.selectedItems))
	}
	for i := 0; i < 3; i++ {
		if _, ok := m.selectedItems[i]; !ok {
			t.Errorf("Expected index %d to be selected", i)
		}
	}

	// Range select up (2 to 0)
	m.selectedItems = make(map[int]struct{})
	m.lastSelect = -1
	m.selected = 2
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	m = m2.(model)
	m.selected = 0
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	m = m2.(model)
	if len(m.selectedItems) != 3 {
		t.Errorf("Expected 3 items selected (reverse), got %d", len(m.selectedItems))
	}
	for i := 0; i < 3; i++ {
		if _, ok := m.selectedItems[i]; !ok {
			t.Errorf("Expected index %d to be selected (reverse)", i)
		}
	}

	// Single-item range
	m.selectedItems = make(map[int]struct{})
	m.lastSelect = -1
	m.selected = 1
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	m = m2.(model)
	m.selected = 1
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	m = m2.(model)
	if len(m.selectedItems) != 1 {
		t.Errorf("Expected 1 item selected (single), got %d", len(m.selectedItems))
	}
	if _, ok := m.selectedItems[1]; !ok {
		t.Errorf("Expected index 1 to be selected (single)")
	}

	// Clear selection with esc
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyEsc})
	m = m2.(model)
	if len(m.selectedItems) != 0 {
		t.Errorf("Expected selection to be cleared with esc")
	}
}
