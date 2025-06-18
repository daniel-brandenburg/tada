// tui.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int

const (
	listView viewMode = iota
	editView
	addView
)

type model struct {
	tasks    map[string][]*TaskWithPath
	items    []item
	selected int
	expanded map[string]bool
	mode     viewMode
	editTask *TaskWithPath
	editForm form
	err      error
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

// Simple color scheme
var (
	accent    = lipgloss.Color("12") // bright blue
	secondary = lipgloss.Color("10") // bright green
	muted     = lipgloss.Color("8")  // gray
	danger    = lipgloss.Color("9")  // bright red
	warning   = lipgloss.Color("11") // bright yellow

	selectedStyle = lipgloss.NewStyle().Reverse(true)
	topicStyle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	taskStyle     = lipgloss.NewStyle()
	mutedStyle    = lipgloss.NewStyle().Foreground(muted)
	focusStyle    = lipgloss.NewStyle().Foreground(warning).Bold(true)
)

func initialModel() model {
	return model{
		expanded: make(map[string]bool),
		mode:     listView,
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
		if m.mode == editView {
			return m.updateEditView(msg)
		} else if m.mode == addView {
			return m.updateAddView(msg)
		}
		return m.updateListView(msg)

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

func (m model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
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
		m.mode = addView
		m.initAddForm()
	case "r":
		return m, loadTasks
	}
	return m, nil
}

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
	case "left", "h":
		if m.editForm.field == 3 { // Status field
			m.cycleStatus(-1)
		}
	case "right", "l":
		if m.editForm.field == 3 { // Status field
			m.cycleStatus(1)
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
	case "left", "h":
		if m.editForm.field == 3 { // Status field
			m.cycleStatus(-1)
		}
	case "right", "l":
		if m.editForm.field == 3 { // Status field
			m.cycleStatus(1)
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

func (m *model) initAddForm() {
	m.editForm = form{
		field:    0,
		title:    "",
		desc:     "",
		priority: "3",
		status:   StatusTodo,
		tags:     "",
	}
}

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
		m.err = fmt.Errorf("Error adding task: %v\n", err)
		return m, nil
	}

	m.mode = listView
	return m, loadTasks
}

func (m *model) buildItems() {
	m.items = []item{}

	// Root tasks first
	if tasks, exists := m.tasks[""]; exists {
		m.addTopic("", tasks)
	}

	// Other topics
	for topic, tasks := range m.tasks {
		if topic != "" {
			m.addTopic(topic, tasks)
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

	if m.expanded[topic] {
		for _, task := range tasks {
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
	s += mutedStyle.Render("j/k: move • space: expand • enter: edit • a: add • r: refresh • q: quit") + "\n\n"

	if len(m.items) == 0 {
		return s + mutedStyle.Render("No tasks found. Press 'a' to add a task.")
	}

	for i, item := range m.items {
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

		s += line + "\n"
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

func RunTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
