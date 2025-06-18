package main

import (
	"testing"
	"time"
)

// TestTaskStatus tests the TaskStatus constants
func TestTaskStatus(t *testing.T) {
	// Verify that task status constants are defined correctly
	if StatusTodo != "todo" {
		t.Errorf("Expected StatusTodo to be 'todo', got '%s'", StatusTodo)
	}

	if StatusInProgress != "in-progress" {
		t.Errorf("Expected StatusInProgress to be 'in-progress', got '%s'", StatusInProgress)
	}

	if StatusDone != "done" {
		t.Errorf("Expected StatusDone to be 'done', got '%s'", StatusDone)
	}

	if StatusCancelled != "cancelled" {
		t.Errorf("Expected StatusCancelled to be 'cancelled', got '%s'", StatusCancelled)
	}

	if StatusPaused != "paused" {
		t.Errorf("Expected StatusPaused to be 'paused', got '%s'", StatusPaused)
	}
}

// TestTask tests the Task struct functionality
func TestTask(t *testing.T) {
	// Create a task
	task := Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    1,
		Status:      StatusTodo,
		Tags:        []string{"test", "unit-test"},
		CreatedAt:   time.Now(),
	}

	// Verify task fields
	if task.Title != "Test Task" {
		t.Errorf("Expected title to be 'Test Task', got '%s'", task.Title)
	}

	if task.Description != "This is a test task" {
		t.Errorf("Expected description to be 'This is a test task', got '%s'", task.Description)
	}

	if task.Priority != 1 {
		t.Errorf("Expected priority to be 1, got %d", task.Priority)
	}

	if task.Status != StatusTodo {
		t.Errorf("Expected status to be '%s', got '%s'", StatusTodo, task.Status)
	}

	if len(task.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(task.Tags))
	}

	if task.Tags[0] != "test" || task.Tags[1] != "unit-test" {
		t.Errorf("Expected tags to be ['test', 'unit-test'], got %v", task.Tags)
	}

	if task.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be set, got zero time")
	}

	if task.CompletedAt != nil {
		t.Errorf("Expected CompletedAt to be nil for a new task")
	}

	// Mark the task as completed
	now := time.Now()
	task.Status = StatusDone
	task.CompletedAt = &now

	if task.Status != StatusDone {
		t.Errorf("Expected status to be '%s', got '%s'", StatusDone, task.Status)
	}

	if task.CompletedAt == nil {
		t.Errorf("Expected CompletedAt to be set after completion")
	}
}

// TestTaskWithPath tests the TaskWithPath struct functionality
func TestTaskWithPath(t *testing.T) {
	// Create a task
	task := &Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    1,
		Status:      StatusTodo,
		Tags:        []string{"test", "unit-test"},
		CreatedAt:   time.Now(),
	}

	// Create a TaskWithPath
	taskWithPath := TaskWithPath{
		Task:     task,
		FilePath: "/path/to/task.md",
		Topic:    "test-topic",
	}

	// Verify TaskWithPath fields
	if taskWithPath.Task != task {
		t.Errorf("Expected Task to be the same as the one we created")
	}

	if taskWithPath.FilePath != "/path/to/task.md" {
		t.Errorf("Expected FilePath to be '/path/to/task.md', got '%s'", taskWithPath.FilePath)
	}

	if taskWithPath.Topic != "test-topic" {
		t.Errorf("Expected Topic to be 'test-topic', got '%s'", taskWithPath.Topic)
	}
}

