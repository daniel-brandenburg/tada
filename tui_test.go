package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitAddFormWithTopic(t *testing.T) {
	m := &model{}
	m.initAddFormWithTopic("")
	if m.editForm.title != "" {
		t.Errorf("Expected empty title for root, got '%s'", m.editForm.title)
	}

	topic := "Project"
	m.initAddFormWithTopic(topic)
	if m.editForm.title != topic+"/" {
		t.Errorf("Expected title to be prefilled with topic, got '%s'", m.editForm.title)
	}
}

func TestInitAddFormWithTopicFromTask(t *testing.T) {
	m := &model{}
	// Simulate selecting a task within a topic
	topic := "Project"
	itm := item{isTopic: false, topic: topic, task: &TaskWithPath{}}
	m.items = []item{itm}
	m.selected = 0
	m.mode = listView

	// Simulate pressing 'a' (supported fields only)
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mModel := m2.(model)

	if mModel.editForm.title != topic+"/" {
		t.Errorf("Expected title to be prefilled with topic, got '%s'", mModel.editForm.title)
	}
}
