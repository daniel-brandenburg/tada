package main

import (
	"testing"
)

func TestEditText(t *testing.T) {
	m := &model{}
	m.editForm.field = 0
	m.editForm.title = "foo"
	m.editText("")
	if m.editForm.title != "fo" {
		t.Errorf("Expected backspace to remove last char, got '%s'", m.editForm.title)
	}
	m.editText("b")
	if m.editForm.title != "fob" {
		t.Errorf("Expected to append 'b', got '%s'", m.editForm.title)
	}
}

func TestCycleStatus(t *testing.T) {
	m := &model{}
	m.editForm.status = StatusTodo
	m.cycleStatus(1)
	if m.editForm.status != StatusInProgress {
		t.Errorf("Expected status to cycle to in-progress, got '%s'", m.editForm.status)
	}
	m.cycleStatus(-1)
	if m.editForm.status != StatusTodo {
		t.Errorf("Expected status to cycle back to todo, got '%s'", m.editForm.status)
	}
}

func TestCyclePriority(t *testing.T) {
	m := &model{}
	m.editForm.priority = "3"
	m.cyclePriority(1)
	if m.editForm.priority != "4" {
		t.Errorf("Expected priority to increase to 4, got '%s'", m.editForm.priority)
	}
	m.cyclePriority(-1)
	if m.editForm.priority != "3" {
		t.Errorf("Expected priority to decrease to 3, got '%s'", m.editForm.priority)
	}
}

func TestBuildItemsAndAddTopic(t *testing.T) {
	m := &model{
		tasks: map[string][]*TaskWithPath{
			"":        {{Task: &Task{Title: "root task"}}},
			"Project": {{Task: &Task{Title: "proj task"}}},
		},
		expanded: map[string]bool{"Project": true},
	}
	m.buildItems()
	foundRoot := false
	foundTopic := false
	for _, itm := range m.items {
		if itm.isTopic && itm.topic == "Project" {
			foundTopic = true
		}
		if !itm.isTopic && itm.topic == "" && itm.task != nil && itm.task.Task.Title == "root task" {
			foundRoot = true
		}
	}
	if !foundRoot || !foundTopic {
		t.Errorf("Expected to find both root task and Project topic in items")
	}
}

func TestGetStatusIcon(t *testing.T) {
	if getStatusIcon(StatusTodo) != "○" {
		t.Errorf("Expected icon for todo")
	}
	if getStatusIcon(StatusDone) != "●" {
		t.Errorf("Expected icon for done")
	}
}
