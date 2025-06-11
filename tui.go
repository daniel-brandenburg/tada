package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tuiModel struct {
	tasks    map[string][]*TaskWithPath
	topics   []string
	selected int
	expanded map[string]bool
	items    []tuiItem
	err      error
}

type tuiItem struct {
	title   string
	isTopic bool
	topic   string
	task    *TaskWithPath
	indent  int
}

type loadTasksMsg struct {
	tasks map[string][]*TaskWithPath
	err   error
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	topicStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	taskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#874BFD")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	statusStyles = map[TaskStatus]lipgloss.Style{
		StatusTodo:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")),
		StatusInProgress: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")),
		StatusDone:       lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
		StatusCancelled:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
		StatusPaused:     lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
	}
)

func initialModel() tuiModel {
	return tuiModel{
		expanded: make(map[string]bool),
		items:    []tuiItem{},
	}
}

func (m tuiModel) Init() tea.Cmd {
	return loadTasks
}

func loadTasks() tea.Msg {
	store := NewFileStore()
	tasks, err := store.LoadAllTasks()
	return loadTasksMsg{tasks: tasks, err: err}
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case "enter", "space":
			if m.selected < len(m.items) {
				item := m.items[m.selected]
				if item.isTopic {
					m.expanded[item.topic] = !m.expanded[item.topic]
					m.rebuildItems()
				}
			}
		case "r":
			return m, loadTasks
		}

	case loadTasksMsg:
		m.tasks = msg.tasks
		m.err = msg.err
		m.rebuildItems()

	case tea.WindowSizeMsg:
		// Handle window resize if needed
	}

	return m, nil
}

func (m *tuiModel) rebuildItems() {
	m.items = []tuiItem{}

	// Sort topics
	topics := make([]string, 0, len(m.tasks))
	for topic := range m.tasks {
		topics = append(topics, topic)
	}

	// Add root topic first if it exists
	if _, exists := m.tasks[""]; exists {
		m.addTopicItems("", 0)
	}

	// Add other topics
	for _, topic := range topics {
		if topic != "" {
			m.addTopicItems(topic, 0)
		}
	}
}

func (m *tuiModel) addTopicItems(topic string, indent int) {
	displayTopic := topic
	if displayTopic == "" {
		displayTopic = "Root"
	}

	// Add topic item
	m.items = append(m.items, tuiItem{
		title:   displayTopic,
		isTopic: true,
		topic:   topic,
		indent:  indent,
	})

	// Add tasks if topic is expanded
	if m.expanded[topic] {
		tasks := m.tasks[topic]
		for _, task := range tasks {
			title := task.Task.Title
			if task.Task.Priority != "" {
				title = fmt.Sprintf("[%s] %s", task.Task.Priority, title)
			}

			m.items = append(m.items, tuiItem{
				title:   title,
				isTopic: false,
				topic:   topic,
				task:    task,
				indent:  indent + 1,
			})
		}
	}
}

func (m tuiModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("TADA - Todo Manager"))
	s.WriteString("\n\n")
	s.WriteString("j/k: navigate • enter/space: expand/collapse • r: refresh • q: quit\n\n")

	if len(m.items) == 0 {
		s.WriteString("No tasks found. Use 'tada add' to create tasks.\n")
		return s.String()
	}

	for i, item := range m.items {
		prefix := strings.Repeat("  ", item.indent)

		var line string
		if item.isTopic {
			expanded := m.expanded[item.topic]
			icon := "▶"
			if expanded {
				icon = "▼"
			}
			line = fmt.Sprintf("%s%s %s", prefix, icon, item.title)
			line = topicStyle.Render(line)
		} else {
			statusIcon := "○"
			switch item.task.Task.Status {
			case StatusTodo:
				statusIcon = "○"
			case StatusInProgress:
				statusIcon = "◐"
			case StatusDone:
				statusIcon = "●"
			case StatusCancelled:
				statusIcon = "✗"
			case StatusPaused:
				statusIcon = "⏸"
			}

			line = fmt.Sprintf("%s%s %s", prefix, statusIcon, item.title)
			if style, ok := statusStyles[item.task.Task.Status]; ok {
				line = style.Render(line)
			} else {
				line = taskStyle.Render(line)
			}
		}

		if i == m.selected {
			line = selectedStyle.Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

func RunTUI() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v", err)
	}
}
