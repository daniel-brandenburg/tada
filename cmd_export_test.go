package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestExportCommand_JSON(t *testing.T) {
	// Setup: create tasks and a temp output file
	store := newTestStoreWithTasks([]Task{{Title: "Task 1"}, {Title: "Task 2"}})
	tmpDir := t.TempDir()
	outputFile := tmpDir + "/test_export.json"

	cmd := NewExportCmd(store)
	cmd.SetArgs([]string{"--format", "json", "--output", outputFile})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("export json failed: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(tasks) != 2 || tasks[0].Title != "Task 1" || tasks[1].Title != "Task 2" {
		t.Errorf("unexpected tasks: %+v", tasks)
	}
}

func TestExportCommand_CSV(t *testing.T) {
	store := newTestStoreWithTasks([]Task{{Title: "A"}, {Title: "B"}})
	tmpDir := t.TempDir()
	outputFile := tmpDir + "/test_export.csv"

	cmd := NewExportCmd(store)
	cmd.SetArgs([]string{"--format", "csv", "--output", outputFile})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("export csv failed: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("invalid csv: %v", err)
	}
	if len(records) < 3 || records[1][0] != "A" || records[2][0] != "B" {
		t.Errorf("unexpected csv records: %+v", records)
	}
}

func TestExportCommand_Markdown(t *testing.T) {
	store := newTestStoreWithTasks([]Task{{Title: "Foo"}, {Title: "Bar"}})
	tmpDir := t.TempDir()
	outputFile := tmpDir + "/test_export.md"

	cmd := NewExportCmd(store)
	cmd.SetArgs([]string{"--format", "md", "--output", outputFile})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("export md failed: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Foo") || !strings.Contains(content, "Bar") {
		t.Errorf("markdown missing tasks: %s", content)
	}
}

func TestExportCommand_InvalidFormat(t *testing.T) {
	store := newTestStoreWithTasks([]Task{{Title: "X"}})
	tmpDir := t.TempDir()
	outputFile := tmpDir + "/test_export.invalid"

	cmd := NewExportCmd(store)
	cmd.SetArgs([]string{"--format", "invalid", "--output", outputFile})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected invalid format error, got: %v", err)
	}
}

// newTestStoreWithTasks is a helper for test setup
func newTestStoreWithTasks(tasks []Task) Storage {
	mem := NewMemoryStorage()
	for _, t := range tasks {
		_ = mem.AddTask(t)
	}
	return mem
}
