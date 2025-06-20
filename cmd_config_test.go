package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndMerge(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tada-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up XDG_CONFIG_HOME for global config
	xdg := filepath.Join(tempDir, "xdg")
	os.MkdirAll(filepath.Join(xdg, "tada"), 0755)
	os.Setenv("XDG_CONFIG_HOME", xdg)
	globalPath := filepath.Join(xdg, "tada", "config.yaml")
	os.WriteFile(globalPath, []byte("default_sort: priority\ntheme: dark\n"), 0644)

	// Set up local config
	localDir := filepath.Join(tempDir, "proj")
	os.MkdirAll(filepath.Join(localDir, ".tada"), 0755)
	localPath := filepath.Join(localDir, ".tada", "config.yaml")
	os.WriteFile(localPath, []byte("theme: light\ndefault_status: done\n"), 0644)

	// Change working dir to localDir
	os.Chdir(localDir)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.DefaultSort != "priority" {
		t.Errorf("Expected default_sort 'priority', got '%s'", cfg.DefaultSort)
	}
	if cfg.Theme != "light" {
		t.Errorf("Expected theme 'light' (local override), got '%s'", cfg.Theme)
	}
	if cfg.DefaultStatus != "done" {
		t.Errorf("Expected default_status 'done' (local), got '%s'", cfg.DefaultStatus)
	}
}
