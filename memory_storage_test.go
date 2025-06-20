package main

import (
	"sync"
	"time"
)

// MemoryStorage is an in-memory implementation of Storage for testing.
type MemoryStorage struct {
	mu    sync.Mutex
	tasks []*TaskWithPath
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (m *MemoryStorage) LoadAllTasks() (map[string][]*TaskWithPath, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string][]*TaskWithPath)
	for _, t := range m.tasks {
		topic := t.Topic
		result[topic] = append(result[topic], t)
	}
	return result, nil
}

func (m *MemoryStorage) AddTask(task Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	m.tasks = append(m.tasks, &TaskWithPath{Task: &task, Topic: "", FilePath: ""})
	return nil
}
