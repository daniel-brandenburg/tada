package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// Setup test environment
func setupTestEnv(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "tada-cmd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	testTadaDir := filepath.Join(tempDir, ".tada")
	fs := NewFileStore(testTadaDir)
	if err := fs.ensureDirectories(); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return testTadaDir, cleanup
}

// TestAddCmd tests the add command
func TestAddCmd(t *testing.T) {
	testTadaDir, cleanup := setupTestEnv(t)
	defer cleanup()
	store := NewFileStore(testTadaDir)

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Save original stdout and replace it with our buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout when done
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test adding a simple task
	testCases := []struct {
		args        []string
		flags       map[string]string
		expectTitle string
		expectTopic string
	}{
		{
			args:        []string{"Test Task"},
			flags:       map[string]string{},
			expectTitle: "Test Task",
			expectTopic: "",
		},
		{
			args:        []string{"work/Test Work Task"},
			flags:       map[string]string{},
			expectTitle: "Test Work Task",
			expectTopic: "work",
		},
		{
			args:        []string{"Test Task with Description"},
			flags:       map[string]string{"description": "This is a test description"},
			expectTitle: "Test Task with Description",
			expectTopic: "",
		},
		{
			args:        []string{"Test Task with Priority"},
			flags:       map[string]string{"priority": "1"},
			expectTitle: "Test Task with Priority",
			expectTopic: "",
		},
		{
			args:        []string{"Test Task with Tags"},
			flags:       map[string]string{"tags": "test,unit-test"},
			expectTitle: "Test Task with Tags",
			expectTopic: "",
		},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Create a new command
			cmd := NewAddCmd(store)
			for k, v := range tc.flags {
				cmd.Flags().Set(k, v)
			}
			cmd.SetArgs(tc.args)
			cmd.SetOut(&buf)
			cmd.Execute()

			// Flush stdout to our buffer
			w.Close()
			io.Copy(&buf, r)

			// Check if task was created
			tasks, err := store.LoadAllTasks()
			if err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			// Check if task exists in the expected topic
			found := false
			for topic, taskList := range tasks {
				if topic == tc.expectTopic {
					for _, task := range taskList {
						if task.Task.Title == tc.expectTitle {
							found = true
							break
						}
					}
				}
			}

			if !found {
				t.Errorf("Task '%s' not found in topic '%s'", tc.expectTitle, tc.expectTopic)
			}
		})
	}
}

// TestListCmd tests the list command
func TestListCmd(t *testing.T) {
	testTadaDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create some test tasks
	store := NewFileStore(testTadaDir)
	tasks := []*Task{
		{
			Title:       "Task 1",
			Description: "Description 1",
			Priority:    1,
			Status:      StatusTodo,
			Tags:        []string{"test", "high"},
			CreatedAt:   time.Now(),
		},
		{
			Title:       "Task 2",
			Description: "Description 2",
			Priority:    2,
			Status:      StatusInProgress,
			Tags:        []string{"test", "medium"},
			CreatedAt:   time.Now(),
		},
		{
			Title:       "Task 3",
			Description: "Description 3",
			Priority:    3,
			Status:      StatusDone,
			Tags:        []string{"test", "low"},
			CreatedAt:   time.Now(),
			CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
		},
	}

	// Save tasks
	for _, task := range tasks {
		if err := store.SaveTask("", task); err != nil {
			t.Fatalf("Failed to save test task: %v", err)
		}
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Save original stdout and replace it with our buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout when done
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test listing tasks
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		expectTasks  []string
		unexpectTask string
	}{
		{
			name:        "List all tasks",
			args:        []string{},
			flags:       map[string]string{},
			expectTasks: []string{"Task 1", "Task 2", "Task 3"},
		},
		{
			name:         "List todo tasks",
			args:         []string{},
			flags:        map[string]string{"status": "todo"},
			expectTasks:  []string{"Task 1"},
			unexpectTask: "Task 2",
		},
		{
			name:         "List in-progress tasks",
			args:         []string{},
			flags:        map[string]string{"status": "in-progress"},
			expectTasks:  []string{"Task 2"},
			unexpectTask: "Task 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Create a new command
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().StringP("status", "s", "", "")
			cmd.Flags().String("sort", "created", "")

			// Set flags
			for k, v := range tc.flags {
				cmd.Flags().Set(k, v)
			}

			// Execute the command function
			listCmd := NewListCmd(store)
			listCmd.SetArgs(tc.args)
			for k, v := range tc.flags {
				listCmd.Flags().Set(k, v)
			}
			listCmd.SetOut(&buf)
			listCmd.Execute()

			// Flush stdout to our buffer
			w.Close()
			io.Copy(&buf, r)
			output := buf.String()

			// Check if expected tasks are in the output
			for _, taskTitle := range tc.expectTasks {
				if !strings.Contains(output, taskTitle) {
					t.Errorf("Expected output to contain task '%s', but it didn't", taskTitle)
				}
			}

			// Check if unexpected task is not in the output
			if tc.unexpectTask != "" && strings.Contains(output, tc.unexpectTask) {
				t.Errorf("Expected output to not contain task '%s', but it did", tc.unexpectTask)
			}
		})
	}
}

// TestCompleteCmd tests the complete command
func TestCompleteCmd(t *testing.T) {
	testTadaDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a test task
	store := NewFileStore(testTadaDir)
	task := &Task{
		Title:       "Test Complete Task",
		Description: "This is a task to test completion",
		Priority:    1,
		Status:      StatusTodo,
		Tags:        []string{"test", "complete"},
		CreatedAt:   time.Now(),
	}

	// Save task
	if err := store.SaveTask("", task); err != nil {
		t.Fatalf("Failed to save test task: %v", err)
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Save original stdout and replace it with our buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout when done
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test completing the task
	cmd := NewCompleteCmd(store)
	cmd.SetArgs([]string{"Test Complete Task"})
	cmd.SetOut(&buf)
	cmd.Execute()

	// Flush stdout to our buffer
	w.Close()
	io.Copy(&buf, r)
	output := buf.String()

	// Check if the output indicates successful completion
	if !strings.Contains(output, "Task completed and archived") {
		t.Errorf("Expected output to indicate task completion, got: %s", output)
	}

	// Check if task was moved to archive
	tasks, err := store.LoadAllTasks()
	if err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	// Task should no longer be in the active tasks
	for _, taskList := range tasks {
		for _, taskWithPath := range taskList {
			if taskWithPath.Task.Title == "Test Complete Task" {
				t.Errorf("Task should have been archived but is still active")
			}
		}
	}

	// Get a new store instance
	store = NewFileStore(testTadaDir)
	// Check archive directory
	archiveFiles, err := os.ReadDir(filepath.Join(store.basePath, ArchiveDir))
	if err != nil {
		t.Fatalf("Failed to read archive directory: %v", err)
	}

	if len(archiveFiles) != 1 {
		t.Errorf("Expected 1 file in archive directory, got %d", len(archiveFiles))
	}
}
