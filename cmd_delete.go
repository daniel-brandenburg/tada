package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [topic/]title",
		Short: "Delete a task",
		Long:  "Delete a task by topic/title.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := strings.Join(args, " ")
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
			if err := os.Remove(found.FilePath); err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Failed to delete: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}
			successStyle := lipgloss.NewStyle().Foreground(cliPrimary).Bold(true)
			fmt.Fprintln(cmd.OutOrStdout(), successStyle.Render("Task deleted."))
		},
	}
	return cmd
}
