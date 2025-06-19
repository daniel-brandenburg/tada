package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegration tests the entire application workflow
func TestIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tada-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a FileStore with a custom base path for testing
	testTadaDir := filepath.Join(tempDir, ".tada")

	// Create a FileStore with custom path
	fs := &FileStore{
		basePath: testTadaDir,
	}
	if err := fs.ensureDirectories(); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test the complete workflow:
	// 1. Add tasks
	// 2. List tasks
	// 3. Complete tasks
	// 4. Verify archive

	// Step 1: Add tasks
	t.Run("AddTasks", func(t *testing.T) {
		tasks := []*Task{
			{
				Title:       "Integration Test Task 1",
				Description: "Description for task 1",
				Priority:    1,
				Status:      StatusTodo,
				Tags:        []string{"integration", "high"},
				CreatedAt:   time.Now(),
			},
			{
				Title:       "Integration Test Task 2",
				Description: "Description for task 2",
				Priority:    2,
				Status:      StatusInProgress,
				Tags:        []string{"integration", "medium"},
				CreatedAt:   time.Now(),
			},
			{
				Title:       "Integration Test Task 3",
				Description: "Description for task 3",
				Priority:    3,
				Status:      StatusTodo,
				Tags:        []string{"integration", "project"},
				CreatedAt:   time.Now(),
			},
		}

		for i, task := range tasks {
			var topic string
			if i == 2 {
				topic = "Project"
			}

			if err := fs.SaveTask(topic, task); err != nil {
				t.Fatalf("Failed to save task: %v", err)
			}
		}

		// Verify tasks were saved
		savedTasks, err := fs.LoadAllTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		// Check if we have tasks in the expected topics
		if len(savedTasks[""]) != 2 {
			t.Errorf("Expected 2 tasks in root topic, got %d", len(savedTasks[""]))
		}

		if len(savedTasks["Project"]) != 1 {
			t.Errorf("Expected 1 task in Project topic, got %d", len(savedTasks["Project"]))
		}

		// Verify task content
		for _, taskList := range savedTasks {
			for _, task := range taskList {
				if !strings.Contains(task.Task.Title, "Integration Test Task") {
					t.Errorf("Unexpected task title: %s", task.Task.Title)
				}

				if !strings.Contains(task.Task.Description, "Description for task") {
					t.Errorf("Unexpected task description: %s", task.Task.Description)
				}

				if len(task.Task.Tags) == 0 || task.Task.Tags[0] != "integration" {
					t.Errorf("Expected 'integration' tag, got: %v", task.Task.Tags)
				}
			}
		}
	})

	// Step 2: Complete a task
	t.Run("CompleteTask", func(t *testing.T) {
		// Complete the first task
		if err := fs.CompleteTask("", "Integration Test Task 1"); err != nil {
			t.Fatalf("Failed to complete task: %v", err)
		}

		// Verify task was moved to archive
		tasks, err := fs.LoadAllTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		// Check if we now have one less task in the root topic
		if len(tasks[""]) != 1 {
			t.Errorf("Expected 1 task in root topic after completion, got %d", len(tasks[""]))
		}

		// Check if task is in archive
		archiveFiles, err := os.ReadDir(filepath.Join(fs.basePath, ArchiveDir))
		if err != nil {
			t.Fatalf("Failed to read archive directory: %v", err)
		}

		if len(archiveFiles) != 1 {
			t.Errorf("Expected 1 file in archive directory, got %d", len(archiveFiles))
		}
	})

	// Step 3: Complete a task in a topic
	t.Run("CompleteTopicTask", func(t *testing.T) {
		// Complete the task in the Project topic
		if err := fs.CompleteTask("Project", "Integration Test Task 3"); err != nil {
			t.Fatalf("Failed to complete task in topic: %v", err)
		}

		// Verify task was moved to archive
		tasks, err := fs.LoadAllTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		// Check if the Project topic is now empty
		if len(tasks["Project"]) != 0 {
			t.Errorf("Expected 0 tasks in Project topic after completion, got %d", len(tasks["Project"]))
		}

		// Check if task is in archive
		topicArchiveDir := filepath.Join(fs.basePath, ArchiveDir, "Project")
		topicArchiveFiles, err := os.ReadDir(topicArchiveDir)
		if err != nil {
			t.Fatalf("Failed to read topic archive directory: %v", err)
		}

		if len(topicArchiveFiles) != 1 {
			t.Errorf("Expected 1 file in topic archive directory, got %d", len(topicArchiveFiles))
		}
	})

	// Step 4: Verify final state
	t.Run("FinalState", func(t *testing.T) {
		// Verify we have one active task left
		tasks, err := fs.LoadAllTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		totalTasks := 0
		for _, taskList := range tasks {
			totalTasks += len(taskList)
		}

		if totalTasks != 1 {
			t.Errorf("Expected 1 active task remaining, got %d", totalTasks)
		}

		// Verify we have two archived tasks
		archiveFiles1, err := filepath.Glob(filepath.Join(fs.basePath, ArchiveDir, "*.md"))
		if err != nil {
			t.Fatalf("Failed to find archive files: %v", err)
		}
		archiveFiles2, err := filepath.Glob(filepath.Join(fs.basePath, ArchiveDir, "**/*.md"))
		if err != nil {
			t.Fatalf("Failed to find archive files: %v", err)
		}
		archivedCount := len(archiveFiles1) + len(archiveFiles2)
		if archivedCount != 2 {
			t.Errorf("Expected 2 archived tasks, got %d", archivedCount)
		}
	})
}
