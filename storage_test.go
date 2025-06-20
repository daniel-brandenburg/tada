package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFileStore tests the FileStore functionality
func TestFileStore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tada-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a custom FileStore with the test directory
	fs := &FileStore{
		basePath: filepath.Join(tempDir, ".tada"),
	}

	// Test ensureDirectories
	t.Run("EnsureDirectories", func(t *testing.T) {
		err := fs.ensureDirectories()
		if err != nil {
			t.Fatalf("Failed to ensure directories: %v", err)
		}

		// Check if directories were created
		dirs := []string{
			fs.basePath,
			filepath.Join(fs.basePath, TasksDir),
			filepath.Join(fs.basePath, ArchiveDir),
		}

		for _, dir := range dirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Directory was not created: %s", dir)
			}
		}
	})

	// Test SaveTask
	t.Run("SaveTask", func(t *testing.T) {
		task := &Task{
			Title:       "Test Task",
			Description: "This is a test task",
			Priority:    1,
			Status:      StatusTodo,
			Tags:        []string{"test", "unit-test"},
			CreatedAt:   time.Now(),
		}

		// Save task without topic
		err := fs.SaveTask("", task)
		if err != nil {
			t.Fatalf("Failed to save task: %v", err)
		}

		// Check if task file was created in the tasks directory
		files, err := os.ReadDir(filepath.Join(fs.basePath, TasksDir))
		if err != nil {
			t.Fatalf("Failed to read tasks directory: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 file, got %d", len(files))
		}

		// Save task with topic
		taskWithTopic := &Task{
			Title:       "Test Task with Topic",
			Description: "This is a test task with a topic",
			Priority:    2,
			Status:      StatusInProgress,
			Tags:        []string{"test", "topic"},
			CreatedAt:   time.Now(),
		}

		err = fs.SaveTask("test-topic", taskWithTopic)
		if err != nil {
			t.Fatalf("Failed to save task with topic: %v", err)
		}

		// Check if topic directory was created
		topicDir := filepath.Join(fs.basePath, TasksDir, "test-topic")
		if _, err := os.Stat(topicDir); os.IsNotExist(err) {
			t.Errorf("Topic directory was not created: %s", topicDir)
		}

		// Check if task file was created in the topic directory
		topicFiles, err := os.ReadDir(topicDir)
		if err != nil {
			t.Fatalf("Failed to read topic directory: %v", err)
		}

		if len(topicFiles) != 1 {
			t.Errorf("Expected 1 file in topic directory, got %d", len(topicFiles))
		}
	})

	// Test LoadAllTasks
	t.Run("LoadAllTasks", func(t *testing.T) {
		tasks, err := fs.LoadAllTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		// Check if we got the expected number of tasks
		if len(tasks) != 2 { // "" topic and "test-topic"
			t.Errorf("Expected 2 topics, got %d", len(tasks))
		}

		if len(tasks[""]) != 1 {
			t.Errorf("Expected 1 task in root topic, got %d", len(tasks[""]))
		}

		if len(tasks["test-topic"]) != 1 {
			t.Errorf("Expected 1 task in test-topic, got %d", len(tasks["test-topic"]))
		}

		// Verify task content
		rootTask := tasks[""][0].Task
		if rootTask.Title != "Test Task" {
			t.Errorf("Expected title 'Test Task', got '%s'", rootTask.Title)
		}

		topicTask := tasks["test-topic"][0].Task
		if topicTask.Title != "Test Task with Topic" {
			t.Errorf("Expected title 'Test Task with Topic', got '%s'", topicTask.Title)
		}
	})

	// Test CompleteTask
	t.Run("CompleteTask", func(t *testing.T) {
		// Complete the task in the root topic
		err := fs.CompleteTask("", "Test Task")
		if err != nil {
			t.Fatalf("Failed to complete task: %v", err)
		}

		// Check if task was moved to archive
		archiveFiles, err := os.ReadDir(filepath.Join(fs.basePath, ArchiveDir))
		if err != nil {
			t.Fatalf("Failed to read archive directory: %v", err)
		}

		if len(archiveFiles) != 1 {
			t.Errorf("Expected 1 file in archive directory, got %d", len(archiveFiles))
		}

		// Check if task was removed from tasks directory
		rootFiles, err := os.ReadDir(filepath.Join(fs.basePath, TasksDir))
		if err != nil {
			t.Fatalf("Failed to read tasks directory: %v", err)
		}

		// We should only have the test-topic directory now
		if len(rootFiles) != 1 || rootFiles[0].Name() != "test-topic" {
			t.Errorf("Expected only test-topic directory in tasks directory")
		}

		// Complete the task in the test-topic
		err = fs.CompleteTask("test-topic", "Test Task with Topic")
		if err != nil {
			t.Fatalf("Failed to complete task in topic: %v", err)
		}

		// Check if task was moved to archive
		topicArchiveDir := filepath.Join(fs.basePath, ArchiveDir, "test-topic")
		topicArchiveFiles, err := os.ReadDir(topicArchiveDir)
		if err != nil {
			t.Fatalf("Failed to read topic archive directory: %v", err)
		}

		if len(topicArchiveFiles) != 1 {
			t.Errorf("Expected 1 file in topic archive directory, got %d", len(topicArchiveFiles))
		}

		// Check if all tasks are now completed
		tasks, err := fs.LoadAllTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		if len(tasks) != 0 {
			t.Errorf("Expected 0 active tasks, got %d topics with tasks", len(tasks))
		}
	})
}

// TestGenerateFileName tests the generateFileName function
func TestGenerateFileName(t *testing.T) {
	fs := NewFileStore()

	testCases := []struct {
		title    string
		expected string // partial match because of timestamp
	}{
		{"Test Task", "-test-task.md"},
		{"Test Task with Spaces", "-test-task-with-spaces.md"},
		{"Test/With/Slashes", "-testwithslashes.md"},
		{"Test with special chars: !@#$%^&*()", "-test-with-special-chars.md"},
	}

	for _, tc := range testCases {
		filename := fs.generateFileName(tc.title)
		if !strings.HasSuffix(filename, tc.expected) {
			t.Errorf("For title '%s', expected filename ending with '%s', got '%s'", tc.title, tc.expected, filename)
		}
	}
}

// TestTaskToMarkdown tests the taskToMarkdown function
func TestTaskToMarkdown(t *testing.T) {
	fs := NewFileStore()
	createdAt := time.Date(2025, 6, 18, 10, 0, 0, 0, time.UTC)

	task := &Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    1,
		Status:      StatusTodo,
		Tags:        []string{"test", "unit-test"},
		CreatedAt:   createdAt,
	}

	markdown := fs.taskToMarkdown(task)

	// Check if markdown contains expected elements
	expectedElements := []string{
		"---", // YAML frontmatter start
		"title: Test Task",
		"description: This is a test task",
		"priority: 1",
		"status: todo",
		"tags:",
		"- test",
		"- unit-test",
		"created_at:",         // followed by timestamp
		"---",                 // YAML frontmatter end
		"# Test Task",         // Title heading
		"This is a test task", // Description
	}

	for _, expected := range expectedElements {
		if !contains(markdown, expected) {
			t.Errorf("Expected markdown to contain '%s', but it didn't", expected)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
