package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Move a task to a new topic
func NewMoveCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move [topic/]title newtopic",
		Short: "Move a task to a new topic",
		Long:  "Move a task to a new topic (changes the file location).",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			newTopic := args[1]
			var topic, title string
			if strings.Contains(input, "/") {
				parts := strings.Split(input, "/")
				if len(parts) >= 2 {
					topic = strings.Join(parts[:len(parts)-1], "/")
					title = parts[len(parts)-1]
				} else {
					title = input
				}
			} else {
				title = input
			}

			tasks, err := store.LoadAllTasks()
			if err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error loading tasks: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}
			var found *TaskWithPath
			for _, taskList := range tasks {
				for _, t := range taskList {
					if t.Task.Title == title && t.Topic == topic {
						found = t
					}
				}
			}
			if found == nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render("Task not found.")
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}

			// Move file to new topic directory
			oldPath := found.FilePath
			newDir := filepath.Join(store.basePath, TasksDir, newTopic)
			if err := os.MkdirAll(newDir, 0755); err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Failed to create new topic directory: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}
			newPath := filepath.Join(newDir, filepath.Base(oldPath))
			if err := os.Rename(oldPath, newPath); err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Failed to move task: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}
			successStyle := lipgloss.NewStyle().Foreground(cliPrimary).Bold(true)
			fmt.Fprintln(cmd.OutOrStdout(), successStyle.Render(fmt.Sprintf("Task moved to topic: %s", newTopic)))
		},
	}
	return cmd
}
