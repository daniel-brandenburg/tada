package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestListCommandDefaultSort(t *testing.T) {
	cmd := exec.Command("./tada", "list", "--output", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Errorf("list command failed: %v", err)
	}
	if !strings.Contains(out.String(), "title") {
		t.Errorf("list output missing expected fields")
	}
}

func TestConfigShow(t *testing.T) {
	cmd := exec.Command("./tada", "config", "show")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Errorf("config show failed: %v", err)
	}
	if !strings.Contains(out.String(), "default_sort") {
		t.Errorf("config show output missing expected fields")
	}
}
