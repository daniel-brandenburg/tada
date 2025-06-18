package main

import (
	"fmt"
	fstore "io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	TadaDir    = ".tada"
	ArchiveDir = "archive"
	TasksDir   = "tasks"
)

type FileStore struct {
	basePath string
}

func NewFileStore(basePath ...string) *FileStore {
	path := TadaDir
	if len(basePath) > 0 && basePath[0] != "" {
		path = basePath[0]
	}
	return &FileStore{
		basePath: path,
	}
}

func (fs *FileStore) ensureDirectories() error {
	dirs := []string{
		fs.basePath,
		filepath.Join(fs.basePath, TasksDir),
		filepath.Join(fs.basePath, ArchiveDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (fs *FileStore) generateFileName(title string) string {
	// Create slug from title
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "") // Remove slashes entirely

	// Remove special characters
	var result strings.Builder
	prevDash := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
			prevDash = false
		} else if r == '-' && !prevDash {
			result.WriteRune(r)
			prevDash = true
		} // else skip
	}
	slug = result.String()

	// Remove leading/trailing dashes
	slug = strings.Trim(slug, "-")

	// Add timestamp
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s.md", timestamp, slug)
}

func (fs *FileStore) SaveTask(topic string, task *Task) error {
	if err := fs.ensureDirectories(); err != nil {
		return err
	}

	// Set creation time if not set
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Create topic directory if needed
	topicPath := filepath.Join(fs.basePath, TasksDir, topic)
	if topic != "" {
		if err := os.MkdirAll(topicPath, 0755); err != nil {
			return fmt.Errorf("failed to create topic directory: %w", err)
		}
	} else {
		topicPath = filepath.Join(fs.basePath, TasksDir)
	}

	// Generate filename
	filename := fs.generateFileName(task.Title)
	filePath := filepath.Join(topicPath, filename)

	// Create markdown content
	content := fs.taskToMarkdown(task)

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

func (fs *FileStore) taskToMarkdown(task *Task) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	yamlData, _ := yaml.Marshal(task)
	content.Write(yamlData)
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(fmt.Sprintf("# %s\n\n", task.Title))

	if task.Description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", task.Description))
	}

	return content.String()
}

func (fs *FileStore) LoadAllTasks() (map[string][]*TaskWithPath, error) {
	if err := fs.ensureDirectories(); err != nil {
		return nil, err
	}

	tasks := make(map[string][]*TaskWithPath)
	tasksPath := filepath.Join(fs.basePath, TasksDir)

	err := filepath.WalkDir(tasksPath, func(path string, d fstore.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Calculate relative topic path
		relPath, err := filepath.Rel(tasksPath, path)
		if err != nil {
			return err
		}

		topic := filepath.Dir(relPath)
		if topic == "." {
			topic = ""
		}

		// Load task
		task, err := fs.loadTaskFromFile(path)
		if err != nil {
			fmt.Printf("Warning: failed to load task from %s: %v\n", path, err)
			return nil
		}

		taskWithPath := &TaskWithPath{
			Task:     task,
			FilePath: path,
			Topic:    topic,
		}

		tasks[topic] = append(tasks[topic], taskWithPath)
		return nil
	})

	return tasks, err
}

func (fs *FileStore) loadTaskFromFile(filePath string) (*Task, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML frontmatter
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		return nil, fmt.Errorf("invalid task file format: missing YAML frontmatter")
	}

	parts := strings.SplitN(contentStr[4:], "\n---\n", 2)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid task file format: malformed YAML frontmatter")
	}

	var task Task
	if err := yaml.Unmarshal([]byte(parts[0]), &task); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	return &task, nil
}

func (fs *FileStore) CompleteTask(topic, title string) error {
	// Find the task file
	tasks, err := fs.LoadAllTasks()
	if err != nil {
		return err
	}

	var targetTask *TaskWithPath
	for t, taskList := range tasks {
		if t == topic {
			for _, task := range taskList {
				if task.Task.Title == title {
					targetTask = task
					break
				}
			}
		}
	}

	if targetTask == nil {
		return fmt.Errorf("task not found: %s", title)
	}

	// Update task status
	targetTask.Task.Status = StatusDone
	now := time.Now()
	targetTask.Task.CompletedAt = &now

	// Move to archive
	archivePath := filepath.Join(fs.basePath, ArchiveDir, topic)
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	filename := filepath.Base(targetTask.FilePath)
	newPath := filepath.Join(archivePath, filename)

	// Save updated task to archive
	content := fs.taskToMarkdown(targetTask.Task)
	if err := os.WriteFile(newPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write archived task: %w", err)
	}

	// Remove original file
	if err := os.Remove(targetTask.FilePath); err != nil {
		return fmt.Errorf("failed to remove original task file: %w", err)
	}

	return nil
}
