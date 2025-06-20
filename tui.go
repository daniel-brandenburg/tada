// tui.go
// Package main implements the Tada terminal-based todo manager TUI.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UndoActionType represents the type of action that can be undone.
type UndoActionType int

const (
	UndoNone UndoActionType = iota
	UndoDelete
	UndoComplete
)

type UndoEntry struct {
	Action UndoActionType
	Task   *TaskWithPath
}

type viewMode int

const (
	listView viewMode = iota
	editView
	addView
)

// model represents the TUI state and logic for the todo manager.
type model struct {
	tasks    map[string][]*TaskWithPath
	items    []item
	selected int
	expanded map[string]bool
	mode     viewMode
	editTask *TaskWithPath
	editForm form
	err      error
	height   int // track terminal height
	// QoL features
	yankedTask    *TaskWithPath
	searchQuery   string
	searchMode    bool
	showDetails   bool
	toArchive     []*TaskWithPath // tasks to archive on exit
	store         *FileStore      // injected for testability, optional
	confirmDelete bool            // show confirm dialog
	pendingDelete *TaskWithPath   // task to delete if confirmed

	// Undo stack (single-level for now)
	undoStack []UndoEntry
	undoMsg   string // status message for undo

	// Bulk selection support
	selectedItems map[int]struct{} // index-based selection for bulk actions
	lastSelect    int              // for range selection (V)

	// Export prompt state
	exportPrompt *exportPromptState // nil unless prompting for export
	exportMsg    string             // status message for export
}

type item struct {
	text    string
	isTopic bool
	topic   string
	task    *TaskWithPath
}

type form struct {
	field    int
	title    string
	desc     string
	priority string
	status   TaskStatus
	tags     string
}

type exportPromptState struct {
	step     int // 0: format, 1: path
	format   string
	filePath string
}

// Simple color scheme
var (
	accent = lipgloss.Color("12") // bright blue
	// mutedStyle is used for help text and muted UI elements
	muted      = lipgloss.Color("8") // gray
	mutedStyle = lipgloss.NewStyle().Foreground(muted)
	warning    = lipgloss.Color("11") // bright yellow

	selectedStyle = lipgloss.NewStyle().Reverse(true)
	topicStyle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	focusStyle    = lipgloss.NewStyle().Foreground(warning).Bold(true)
)

func initialModel() model {
	return model{
		expanded:      make(map[string]bool),
		mode:          listView,
		selectedItems: make(map[int]struct{}),
		lastSelect:    -1,
	}
}

func (m model) Init() tea.Cmd {
	return loadTasks
}

func loadTasks() tea.Msg {
	store := NewFileStore()
	tasks, err := store.LoadAllTasks()
	return struct {
		tasks map[string][]*TaskWithPath
		err   error
	}{tasks, err}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.exportPrompt != nil {
			return m.updateExportPrompt(msg)
		}
		if m.searchMode {
			if msg.String() == "esc" {
				m.searchMode = false
				m.searchQuery = ""
				m.buildItems()
				return m, nil
			}
			if msg.String() == "backspace" && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				return m, nil
			}
			if len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				return m, nil
			}
		}
		if m.mode == editView {
			return m.updateEditView(msg)
		} else if m.mode == addView {
			return m.updateAddView(msg)
		}
		return m.updateListView(msg)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case struct {
		tasks map[string][]*TaskWithPath
		err   error
	}:
		m.tasks = msg.tasks
		m.err = msg.err
		m.buildItems()
		return m, nil
	}
	return m, nil
}

// updateListView handles key events and actions in the main list view.
func (m model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmDelete {
		switch msg.String() {
		case "y":
			if m.pendingDelete != nil {
				// Save undo info before deleting
				m.undoStack = append(m.undoStack, UndoEntry{Action: UndoDelete, Task: m.pendingDelete})
				_ = os.Remove(m.pendingDelete.FilePath)
				m.undoMsg = "Task deleted. Press 'u' to undo."
			}
			m.confirmDelete = false
			m.pendingDelete = nil
			return m, loadTasks
		case "n", "esc":
			m.confirmDelete = false
			m.pendingDelete = nil
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "ctrl+c", "q":
		// Archive any completed tasks before quitting
		if len(m.toArchive) > 0 {
			for _, task := range m.toArchive {
				store := m.store
				if store == nil {
					store = NewFileStore()
				}
				_ = store.CompleteTask(task.Topic, task.Task.Title)
			}
		}
		return m, tea.Quit
	case "/":
		m.searchMode = true
		m.searchQuery = ""
		return m, nil
	case "i":
		if m.showDetails {
			m.showDetails = false
			return m, nil
		}
		m.showDetails = true
		return m, nil
	case "y":
		if m.selected < len(m.items) && m.items[m.selected].task != nil {
			m.yankedTask = m.items[m.selected].task
		}
	case "p":
		if m.yankedTask != nil {
			copy := *m.yankedTask
			orig := m.yankedTask.Task
			copy.Task = &Task{
				Title:       orig.Title + " (Copy)",
				Description: orig.Description,
				Status:      orig.Status,
				Priority:    orig.Priority,
				Tags:        append([]string{}, orig.Tags...),
			}
			store := NewFileStore()
			_ = store.SaveTask(copy.Topic, copy.Task)
			return m, loadTasks
		}
	case "d":
		if len(m.selectedItems) > 0 {
			// Bulk delete
			for idx := range m.selectedItems {
				if idx < len(m.items) && m.items[idx].task != nil {
					_ = os.Remove(m.items[idx].task.FilePath)
					m.undoStack = append(m.undoStack, UndoEntry{Action: UndoDelete, Task: m.items[idx].task})
				}
			}
			m.undoMsg = "Bulk delete complete. Press 'u' to undo last."
			m.selectedItems = make(map[int]struct{})
			return m, loadTasks
		}
		if m.selected < len(m.items) && m.items[m.selected].task != nil {
			m.confirmDelete = true
			m.pendingDelete = m.items[m.selected].task
			return m, nil
		}
	case "e":
		if m.selected < len(m.items) && m.items[m.selected].task != nil {
			m.mode = editView
			m.editTask = m.items[m.selected].task
			m.initForm()
			return m, nil
		}
	case "s":
		if len(m.selectedItems) > 0 {
			for idx := range m.selectedItems {
				if idx < len(m.items) && m.items[idx].task != nil {
					task := m.items[idx].task
					m.cycleTaskStatus(task, 1)
					if task.Task.Status == StatusDone {
						m.toArchive = append(m.toArchive, task)
						m.undoStack = append(m.undoStack, UndoEntry{Action: UndoComplete, Task: task})
					}
				}
			}
			m.undoMsg = "Bulk status cycle complete. Press 'u' to undo last."
			return m, loadTasks
		}
		if m.selected < len(m.items) && m.items[m.selected].task != nil {
			task := m.items[m.selected].task
			m.cycleTaskStatus(task, 1)
			if task.Task.Status == StatusDone {
				m.toArchive = append(m.toArchive, task)
				m.undoStack = append(m.undoStack, UndoEntry{Action: UndoComplete, Task: task})
				m.undoMsg = "Task completed. Press 'u' to undo."
			}
			return m, loadTasks
		}
	case "S":
		if m.selected < len(m.items) && m.items[m.selected].task != nil {
			task := m.items[m.selected].task
			m.cycleTaskStatus(task, -1)
			return m, loadTasks
		}
	case "j", "down":
		if m.selected < len(m.items)-1 {
			m.selected++
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
	case "space", "enter":
		if m.selected < len(m.items) {
			item := m.items[m.selected]
			if item.isTopic {
				m.expanded[item.topic] = !m.expanded[item.topic]
				m.buildItems()
			} else if item.task != nil {
				m.mode = editView
				m.editTask = item.task
				m.initForm()
			}
		}
	case "a":
		// If a topic or a task within a topic is selected, prefill the title with the topic
		topic := ""
		if m.selected < len(m.items) {
			item := m.items[m.selected]
			if item.isTopic {
				topic = item.topic
			} else if item.task != nil && item.topic != "" {
				topic = item.topic
			}
		}
		m.mode = addView
		m.initAddFormWithTopic(topic)
	case "r":
		return m, loadTasks
	case "u":
		if len(m.undoStack) > 0 {
			entry := m.undoStack[len(m.undoStack)-1]
			m.undoStack = m.undoStack[:len(m.undoStack)-1]
			switch entry.Action {
			case UndoDelete:
				// Restore deleted task at original file path
				if entry.Task != nil && entry.Task.FilePath != "" && entry.Task.Task != nil {
					content := "---\n"
					if entry.Task.Task != nil {
						data, _ := yaml.Marshal(entry.Task.Task)
						content += string(data)
					}
					content += "---\n\n# " + entry.Task.Task.Title + "\n"
					_ = os.WriteFile(entry.Task.FilePath, []byte(content), 0644)
					m.undoMsg = "Undo: Task restored."
					return m, loadTasks
				}
				m.undoMsg = "Undo failed."
				return m, nil
			case UndoComplete:
				// Mark as not completed
				entry.Task.Task.Status = StatusTodo
				entry.Task.Task.CompletedAt = nil
				store := m.store
				if store == nil {
					store = NewFileStore()
				}
				_ = store.SaveTask(entry.Task.Topic, entry.Task.Task)
				m.undoMsg = "Undo: Task marked as not completed."
				return m, loadTasks
			}
		}
	case "x":
		if len(m.selectedItems) > 0 {
			m.exportPrompt = &exportPromptState{step: 0}
			return m, nil
		}
	case "V":
		if m.lastSelect == -1 {
			m.lastSelect = m.selected
			m.selectedItems = map[int]struct{}{m.selected: {}}
		} else {
			start, end := m.lastSelect, m.selected
			if start > end {
				start, end = end, start
			}
			for i := start; i <= end; i++ {
				m.selectedItems[i] = struct{}{}
			}
			m.lastSelect = -1 // reset after range select
		}
		return m, nil
	case "v":
		if m.selectedItems == nil {
			m.selectedItems = make(map[int]struct{})
		}
		if _, ok := m.selectedItems[m.selected]; ok {
			delete(m.selectedItems, m.selected)
		} else {
			m.selectedItems[m.selected] = struct{}{}
		}
		m.lastSelect = -1 // clear range mode if toggling single
		return m, nil
	case "esc":
		if len(m.selectedItems) > 0 {
			m.selectedItems = make(map[int]struct{})
			m.lastSelect = -1
			return m, nil
		}
		if m.searchMode {
			m.searchMode = false
			m.searchQuery = ""
			m.buildItems()
			return m, nil
		}
		if m.showDetails {
			m.showDetails = false
			return m, nil
		}
	}
	return m, nil
}

// updateEditView handles key events and actions in the edit view.
func (m model) updateEditView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.mode = listView
		return m, nil
	case "tab":
		m.editForm.field = (m.editForm.field + 1) % 7
	case "shift+tab":
		m.editForm.field = (m.editForm.field - 1 + 7) % 7
	case "enter":
		if m.editForm.field == 5 { // Save
			return m.saveTask()
		} else if m.editForm.field == 6 { // Cancel
			m.mode = listView
			return m, nil
		}
	case "left":
		if m.editForm.field == 3 {
			m.cycleStatus(-1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(-1)
		}
	case "right":
		if m.editForm.field == 3 {
			m.cycleStatus(1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(1)
		}
	case "h":
		if m.editForm.field == 3 {
			m.cycleStatus(-1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(-1)
		} else {
			m.editText("h")
		}
	case "l":
		if m.editForm.field == 3 {
			m.cycleStatus(1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(1)
		} else {
			m.editText("l")
		}
	case "backspace":
		m.editText("")
	case "space":
		m.editText(" ")
	default:
		if len(msg.String()) == 1 {
			m.editText(msg.String())
		}
	}
	return m, nil
}

// updateAddView handles key events and actions in the add view.
func (m model) updateAddView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.mode = listView
		return m, nil
	case "tab":
		m.editForm.field = (m.editForm.field + 1) % 7
	case "shift+tab":
		m.editForm.field = (m.editForm.field - 1 + 7) % 7
	case "enter":
		if m.editForm.field == 5 { // Save
			return m.addTask()
		} else if m.editForm.field == 6 { // Cancel
			m.mode = listView
			return m, nil
		}
	case "left":
		if m.editForm.field == 3 {
			m.cycleStatus(-1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(-1)
		}
	case "right":
		if m.editForm.field == 3 {
			m.cycleStatus(1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(1)
		}
	case "h":
		if m.editForm.field == 3 {
			m.cycleStatus(-1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(-1)
		} else {
			m.editText("h")
		}
	case "l":
		if m.editForm.field == 3 {
			m.cycleStatus(1)
		} else if m.editForm.field == 2 {
			m.cyclePriority(1)
		} else {
			m.editText("l")
		}
	case "backspace":
		m.editText("")
	case "space":
		m.editText(" ")
	default:
		if len(msg.String()) == 1 {
			m.editText(msg.String())
		}
	}
	return m, nil
}

func (m *model) editText(char string) {
	switch m.editForm.field {
	case 0: // Title
		if char == "" && len(m.editForm.title) > 0 {
			m.editForm.title = m.editForm.title[:len(m.editForm.title)-1]
		} else {
			m.editForm.title += char
		}
	case 1: // Description
		if char == "" && len(m.editForm.desc) > 0 {
			m.editForm.desc = m.editForm.desc[:len(m.editForm.desc)-1]
		} else {
			m.editForm.desc += char
		}
	case 2: // Priority
		if char == "" && len(m.editForm.priority) > 0 {
			m.editForm.priority = m.editForm.priority[:len(m.editForm.priority)-1]
		} else if char >= "0" && char <= "9" {
			m.editForm.priority += char
		}
	case 4: // Tags
		if char == "" && len(m.editForm.tags) > 0 {
			m.editForm.tags = m.editForm.tags[:len(m.editForm.tags)-1]
		} else {
			m.editForm.tags += char
		}
	}
}

func (m *model) cycleStatus(direction int) {
	statuses := []TaskStatus{StatusTodo, StatusInProgress, StatusDone, StatusPaused, StatusCancelled}
	current := 0
	for i, status := range statuses {
		if status == m.editForm.status {
			current = i
			break
		}
	}
	current = (current + direction + len(statuses)) % len(statuses)
	m.editForm.status = statuses[current]
}

func (m *model) cyclePriority(direction int) {
	// Default priority range: 1-5
	p := 3
	if m.editForm.priority != "" {
		fmt.Sscanf(m.editForm.priority, "%d", &p)
	}
	p += direction
	if p < 1 {
		p = 1
	}
	if p > 5 {
		p = 5
	}
	m.editForm.priority = fmt.Sprintf("%d", p)
}

func (m *model) initForm() {
	task := m.editTask.Task
	m.editForm = form{
		field:    0,
		title:    task.Title,
		desc:     task.Description,
		priority: fmt.Sprintf("%d", task.Priority),
		status:   task.Status,
		tags:     strings.Join(task.Tags, ", "),
	}
}

func (m *model) initAddFormWithTopic(topic string) {
	m.editForm = form{
		field:    0,
		title:    "",
		desc:     "",
		priority: "3",
		status:   StatusTodo,
		tags:     "",
	}
	if topic != "" {
		m.editForm.title = topic + "/"
	}
}

// saveTask saves the current task edits to the file.
func (m model) saveTask() (tea.Model, tea.Cmd) {
	task := m.editTask.Task
	task.Title = m.editForm.title
	task.Description = m.editForm.desc

	if m.editForm.priority != "" {
		fmt.Sscanf(m.editForm.priority, "%d", &task.Priority)
	}

	task.Status = m.editForm.status

	if m.editForm.tags != "" {
		tags := strings.Split(m.editForm.tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		task.Tags = tags
	} else {
		task.Tags = []string{}
	}

	store := NewFileStore()
	content := store.taskToMarkdown(task)

	if err := os.WriteFile(m.editTask.FilePath, []byte(content), 0644); err != nil {
		m.err = fmt.Errorf("failed to save: %w", err)
	}

	m.mode = listView
	return m, loadTasks
}

// addTask adds a new task from the add view form.
func (m model) addTask() (tea.Model, tea.Cmd) {
	if m.editForm.title == "" {
		m.err = fmt.Errorf("title cannot be empty")
		return m, nil
	}
	title := m.editForm.title
	var topic, taskTitle string
	if strings.Contains(title, "/") {
		parts := strings.Split(title, "/")
		if len(parts) >= 2 {
			topic = filepath.Join(parts[:len(parts)-1]...)
			taskTitle = parts[len(parts)-1]
		} else {
			taskTitle = title
		}
	} else {
		taskTitle = title
	}

	// Create new task
	task := &Task{
		Title:       taskTitle,
		Description: m.editForm.desc,
		Status:      m.editForm.status,
		Priority:    3, // default
	}

	if m.editForm.priority != "" {
		fmt.Sscanf(m.editForm.priority, "%d", &task.Priority)
	}

	if m.editForm.tags != "" {
		tags := strings.Split(m.editForm.tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		task.Tags = tags
	}

	store := NewFileStore()
	if err := store.SaveTask(topic, task); err != nil {
		m.err = fmt.Errorf("error adding task: %v", err)
		return m, nil
	}

	m.mode = listView
	return m, loadTasks
}

// Cycles the status of a given task (for list view status cycling)
func (m *model) cycleTaskStatus(task *TaskWithPath, direction int) {
	statuses := []TaskStatus{StatusTodo, StatusInProgress, StatusDone, StatusPaused, StatusCancelled}
	current := 0
	for i, status := range statuses {
		if status == task.Task.Status {
			current = i
			break
		}
	}
	current = (current + direction + len(statuses)) % len(statuses)
	task.Task.Status = statuses[current]
	// Save the updated status to the original file path
	store := NewFileStore()
	content := store.taskToMarkdown(task.Task)
	_ = os.WriteFile(task.FilePath, []byte(content), 0644)
}

// buildItems constructs the visible list of items for the current state.
func (m *model) buildItems() {
	m.items = []item{}

	// Helper to check if a task is in toArchive
	inToArchive := func(task *TaskWithPath) bool {
		for _, t := range m.toArchive {
			if t.FilePath == task.FilePath {
				return true
			}
		}
		return false
	}

	// Add topics first (excluding root)
	for topic, tasks := range m.tasks {
		if topic != "" {
			m.addTopic(topic, tasks)
		}
	}

	// Add root tasks directly (not under a 'Root' group)
	if tasks, exists := m.tasks[""]; exists {
		for _, task := range tasks {
			if task.Task.Status == StatusDone && !inToArchive(task) {
				continue // hide completed tasks unless just completed
			}
			title := task.Task.Title
			if task.Task.Priority != 3 {
				title = fmt.Sprintf("[%d] %s", task.Task.Priority, title)
			}
			m.items = append(m.items, item{
				text:    getStatusIcon(task.Task.Status) + " " + title,
				isTopic: false,
				topic:   "",
				task:    task,
			})
		}
	}
}

func (m *model) addTopic(topic string, tasks []*TaskWithPath) {
	name := topic
	if name == "" {
		name = "Root"
	}

	m.items = append(m.items, item{
		text:    name,
		isTopic: true,
		topic:   topic,
	})

	inToArchive := func(task *TaskWithPath) bool {
		for _, t := range m.toArchive {
			if t.FilePath == task.FilePath {
				return true
			}
		}
		return false
	}

	if m.expanded[topic] {
		for _, task := range tasks {
			if task.Task.Status == StatusDone && !inToArchive(task) {
				continue // hide completed tasks unless just completed
			}
			title := task.Task.Title
			if task.Task.Priority != 3 {
				title = fmt.Sprintf("[%d] %s", task.Task.Priority, title)
			}

			m.items = append(m.items, item{
				text:    "  " + getStatusIcon(task.Task.Status) + " " + title,
				isTopic: false,
				topic:   topic,
				task:    task,
			})
		}
	}
}

func getStatusIcon(status TaskStatus) string {
	switch status {
	case StatusTodo:
		return "○"
	case StatusInProgress:
		return "◐"
	case StatusDone:
		return "●"
	case StatusPaused:
		return "⏸"
	case StatusCancelled:
		return "✗"
	default:
		return "○"
	}
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}

	if m.mode == editView {
		return m.viewEdit()
	} else if m.mode == addView {
		return m.viewAdd()
	}
	return m.viewList()
}

func (m model) viewList() string {
	s := "TADA - Todo Manager\n"
	s += mutedStyle.Render("j/k: move • space: expand • enter: edit • a: add • r: refresh • d: delete • q: quit") + "\n\n"

	if m.confirmDelete && m.pendingDelete != nil {
		msg := focusStyle.Render("Delete task '") + m.pendingDelete.Task.Title + focusStyle.Render("'? (y/n)")
		popup := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(warning).Background(lipgloss.Color("0")).Padding(1, 2).Width(50).Align(lipgloss.Left).Render(msg)
		return s + popup + "\n"
	}

	if len(m.items) == 0 {
		return s + mutedStyle.Render("No tasks found. Press 'a' to add a task.")
	}

	height := m.height
	if height == 0 {
		height = 24 // fallback default
	}
	reserved := 4 // header + help + some padding
	visibleLines := height - reserved
	if visibleLines < 1 {
		visibleLines = 1
	}

	start := 0
	end := len(m.items)
	if len(m.items) > visibleLines {
		// Center selected if possible
		start = m.selected - visibleLines/2
		if start < 0 {
			start = 0
		}
		end = start + visibleLines
		if end > len(m.items) {
			end = len(m.items)
			start = end - visibleLines
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		item := m.items[i]
		line := item.text

		if item.isTopic {
			icon := "▶"
			if m.expanded[item.topic] {
				icon = "▼"
			}
			line = icon + " " + line
			line = topicStyle.Render(line)
		}

		if i == m.selected {
			line = selectedStyle.Render(line)
		}
		if m.selectedItems != nil {
			if _, ok := m.selectedItems[i]; ok {
				line = focusStyle.Render("[✔] ") + line
			} else {
				line = "    " + line
			}
		}

		s += line + "\n"

		// Insert the popup directly below the selected item
		if m.showDetails && i == m.selected && item.task != nil {
			task := item.task.Task
			detail := lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Task Details") + "\n"
			detail += focusStyle.Render("Title: ") + task.Title + "\n"
			detail += focusStyle.Render("Description: ") + task.Description + "\n"
			detail += focusStyle.Render("Priority: ") + fmt.Sprintf("%d", task.Priority) + "\n"
			detail += focusStyle.Render("Status: ") + string(task.Status) + "\n"
			detail += focusStyle.Render("Tags: ") + strings.Join(task.Tags, ", ") + "\n"
			detail += mutedStyle.Render("(Press esc/i to close)")

			popupStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accent).
				Background(lipgloss.Color("0"))
			popup := popupStyle.Padding(1, 2).Width(50).Align(lipgloss.Left).Render(detail)

			s += popup + "\n"
		}
	}

	// Show undo status message if present
	if m.undoMsg != "" {
		s += focusStyle.Render(m.undoMsg) + "\n"
	}
	if m.exportMsg != "" {
		s += focusStyle.Render(m.exportMsg) + "\n"
	}
	return s
}

func (m model) viewEdit() string {
	s := "Edit Task\n"
	s += mutedStyle.Render("tab: next field • enter: save/cancel • esc: back") + "\n\n"

	fields := []struct {
		label string
		value string
		help  string
	}{
		{"Title:", m.editForm.title, ""},
		{"Description:", m.editForm.desc, ""},
		{"Priority:", m.editForm.priority, "(1-5, default 3)"},
		{"Status:", string(m.editForm.status), "(h/l to change)"},
		{"Tags:", m.editForm.tags, "(comma separated)"},
	}

	for i, field := range fields {
		label := field.label
		value := field.value

		if i == m.editForm.field {
			label = focusStyle.Render(label)
			value += "█" // cursor
		}

		s += fmt.Sprintf("%s %s", label, value)
		if field.help != "" {
			s += " " + mutedStyle.Render(field.help)
		}
		s += "\n"
	}

	s += "\n"

	// Buttons
	save := "Save"
	cancel := "Cancel"

	if m.editForm.field == 5 {
		save = focusStyle.Render("[" + save + "]")
	} else {
		save = "[" + save + "]"
	}

	if m.editForm.field == 6 {
		cancel = focusStyle.Render("[" + cancel + "]")
	} else {
		cancel = "[" + cancel + "]"
	}

	s += save + " " + cancel

	return s
}

func (m model) viewAdd() string {
	s := "Add New Task\n"
	s += mutedStyle.Render("tab: next field • enter: add/cancel • esc: back") + "\n\n"

	fields := []struct {
		label string
		value string
		help  string
	}{
		{"Title:", m.editForm.title, "(required)"},
		{"Description:", m.editForm.desc, ""},
		{"Priority:", m.editForm.priority, "(1-5, default 3)"},
		{"Status:", string(m.editForm.status), "(h/l to change)"},
		{"Tags:", m.editForm.tags, "(comma separated)"},
	}

	for i, field := range fields {
		label := field.label
		value := field.value

		if i == m.editForm.field {
			label = focusStyle.Render(label)
			value += "█" // cursor
		}

		s += fmt.Sprintf("%s %s", label, value)
		if field.help != "" {
			s += " " + mutedStyle.Render(field.help)
		}
		s += "\n"
	}

	s += "\n"

	// Buttons
	add := "Add"
	cancel := "Cancel"

	if m.editForm.field == 5 {
		add = focusStyle.Render("[" + add + "]")
	} else {
		add = "[" + add + "]"
	}

	if m.editForm.field == 6 {
		cancel = focusStyle.Render("[" + cancel + "]")
	} else {
		cancel = "[" + cancel + "]"
	}

	s += add + " " + cancel

	return s
}

// RunTUI launches the Tada TUI application.
func RunTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error: %v", err))
		fmt.Fprintln(os.Stderr, styledErr)
	}
}

// RunTUIWithConfig launches the Tada TUI with a custom config.
func RunTUIWithConfig(cfg *Config) {
	m := initialModel()
	m.applyConfig(cfg)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error: %v", err))
		fmt.Fprintln(os.Stderr, styledErr)
	}
}

func (m *model) applyConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	// Example: apply theme (expand as needed)
	if cfg.Theme == "dark" {
		accent = lipgloss.Color("4")
		muted = lipgloss.Color("8")
		warning = lipgloss.Color("11")
	} else if cfg.Theme == "light" {
		accent = lipgloss.Color("12")
		muted = lipgloss.Color("7")
		warning = lipgloss.Color("11")
	}
	// Add more config-driven settings as needed
}

func showWelcomeIfNeeded(cfg *Config, tadaDir string) {
	// Check config: if ShowWelcome is set to false, skip
	if cfg != nil && cfg.ShowWelcome != nil && !*cfg.ShowWelcome {
		return
	}
	flagPath := filepath.Join(tadaDir, ".welcome_shown")
	if _, err := os.Stat(flagPath); err == nil {
		return // already shown for this project
	}
	fmt.Println(onboardingMessage())
	// Mark as shown
	_ = os.WriteFile(flagPath, []byte("shown\n"), 0644)
}

func (m *model) updateExportPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prompt := m.exportPrompt
	switch prompt.step {
	case 0: // format
		if msg.String() == "1" {
			prompt.format = "csv"
			prompt.step = 1
			return m, nil
		} else if msg.String() == "2" {
			prompt.format = "json"
			prompt.step = 1
			return m, nil
		} else if msg.String() == "3" {
			prompt.format = "md"
			prompt.step = 1
			return m, nil
		} else if msg.String() == "esc" {
			m.exportPrompt = nil
			return m, nil
		}
	case 1: // file path (accept any input, enter to confirm)
		if msg.String() == "esc" {
			m.exportPrompt = nil
			return m, nil
		} else if msg.String() == "backspace" && len(prompt.filePath) > 0 {
			prompt.filePath = prompt.filePath[:len(prompt.filePath)-1]
			return m, nil
		} else if msg.String() == "enter" && prompt.filePath != "" {
			// Do export
			err := m.bulkExportSelected(prompt.format, prompt.filePath)
			if err != nil {
				m.exportMsg = "Export failed: " + err.Error()
			} else {
				m.exportMsg = "Exported selected tasks to " + prompt.filePath
			}
			m.exportPrompt = nil
			return m, nil
		} else if len(msg.String()) == 1 {
			prompt.filePath += msg.String()
			return m, nil
		}
	}
	return m, nil
}

func (m *model) bulkExportSelected(format, filePath string) error {
	var tasks []*TaskWithPath
	for idx := range m.selectedItems {
		if idx < len(m.items) && m.items[idx].task != nil {
			tasks = append(tasks, m.items[idx].task)
		}
	}
	if len(tasks) == 0 {
		return fmt.Errorf("no tasks selected")
	}
	// Use export logic from cmd_export.go (refactor if needed)
	return ExportTasksToFile(tasks, format, filePath)
}
