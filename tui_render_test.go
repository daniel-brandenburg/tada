package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewListRendersVisibleItems(t *testing.T) {
	m := model{
		height:   8, // simulate small terminal
		items:    []item{},
		selected: 2,
	}
	for i := 0; i < 10; i++ {
		m.items = append(m.items, item{text: "Task " + string(rune('A'+i))})
	}
	out := m.viewList()
	// With height=8, reserved=4, visibleLines=4, selected=2, expect Task A-D
	if !strings.Contains(out, "Task C") || !strings.Contains(out, "Task D") {
		t.Errorf("Expected visible window to include Task C and Task D, got: %s", out)
	}
}

func TestViewEditRendersFields(t *testing.T) {
	m := model{}
	m.editForm.title = "TestTitle"
	m.editForm.desc = "TestDesc"
	m.editForm.priority = "2"
	m.editForm.status = StatusInProgress
	m.editForm.tags = "tag1,tag2"
	out := m.viewEdit()
	if !strings.Contains(out, "TestTitle") || !strings.Contains(out, "TestDesc") || !strings.Contains(out, "2") || !strings.Contains(out, "in-progress") || !strings.Contains(out, "tag1,tag2") {
		t.Errorf("Expected all fields in viewEdit output, got: %s", out)
	}
}

func TestViewAddRendersFields(t *testing.T) {
	m := model{}
	m.editForm.title = "AddTitle"
	m.editForm.desc = "AddDesc"
	m.editForm.priority = "4"
	m.editForm.status = StatusPaused
	m.editForm.tags = "tag3,tag4"
	out := m.viewAdd()
	if !strings.Contains(out, "AddTitle") || !strings.Contains(out, "AddDesc") || !strings.Contains(out, "4") || !strings.Contains(out, "paused") || !strings.Contains(out, "tag3,tag4") {
		t.Errorf("Expected all fields in viewAdd output, got: %s", out)
	}
}

func TestViewSwitchesModes(t *testing.T) {
	m := model{err: nil}
	m.mode = editView
	m.editForm.title = "EditMode"
	if !strings.Contains(m.View(), "EditMode") {
		t.Errorf("Expected edit view in View()")
	}
	m.mode = addView
	m.editForm.title = "AddMode"
	if !strings.Contains(m.View(), "AddMode") {
		t.Errorf("Expected add view in View()")
	}
	m.mode = listView
	m.items = []item{{text: "ListMode"}}
	m.height = 10
	if !strings.Contains(m.View(), "ListMode") {
		t.Errorf("Expected list view in View()")
	}
}

func TestUpdateListViewKeys(t *testing.T) {
	m := model{items: []item{{}, {}, {}}, selected: 1}
	m2, _ := m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m2.(model).selected != 2 {
		t.Errorf("Expected selected to move down")
	}
	m2, _ = m.updateListView(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m2.(model).selected != 0 {
		t.Errorf("Expected selected to move up")
	}
}

func TestUpdateEditViewTabAndEnter(t *testing.T) {
	m := model{mode: editView}
	m.editForm.field = 0
	m2, _ := m.updateEditView(tea.KeyMsg{Type: tea.KeyTab})
	if m2.(model).editForm.field != 1 {
		t.Errorf("Expected tab to move to next field")
	}
}

func TestUpdateAddViewTabAndEnter(t *testing.T) {
	m := model{mode: addView}
	m.editForm.field = 0
	m2, _ := m.updateAddView(tea.KeyMsg{Type: tea.KeyTab})
	if m2.(model).editForm.field != 1 {
		t.Errorf("Expected tab to move to next field")
	}
}
