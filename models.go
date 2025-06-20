package main

import (
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in-progress"
	StatusDone       TaskStatus = "done"
	StatusCancelled  TaskStatus = "cancelled"
	StatusPaused     TaskStatus = "paused"
)

type Task struct {
	Title       string     `yaml:"title"`
	Description string     `yaml:"description,omitempty"`
	Priority    int        `yaml:"priority,omitempty"`
	Status      TaskStatus `yaml:"status"`
	Tags        []string   `yaml:"tags,omitempty"`
	CreatedAt   time.Time  `yaml:"created_at"`
	CompletedAt *time.Time `yaml:"completed_at,omitempty"`
}

type TaskWithPath struct {
	Task     *Task
	FilePath string
	Topic    string
}

// Storage is a minimal interface for testable task storage.
type Storage interface {
	LoadAllTasks() (map[string][]*TaskWithPath, error)
}
