package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Use CLI color palette from cmd_list.go
// (Removed local color palette to resolve redeclaration errors)

func NewEditCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [topic/]title",
		Short: "Edit a task",
		Long:  "Edit a task's fields by topic/title.",
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

			description, _ := cmd.Flags().GetString("description")
			priority, _ := cmd.Flags().GetInt("priority")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			status, _ := cmd.Flags().GetString("status")

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

			if description != "" {
				found.Task.Description = description
			}
			if cmd.Flags().Changed("priority") {
				found.Task.Priority = priority
			}
			if len(tags) > 0 {
				found.Task.Tags = tags
			}
			if status != "" {
				found.Task.Status = TaskStatus(status)
			}

			content := store.taskToMarkdown(found.Task)
			if err := os.WriteFile(found.FilePath, []byte(content), 0644); err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Failed to save: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}
			successStyle := lipgloss.NewStyle().Foreground(cliPrimary).Bold(true)
			fmt.Fprintln(cmd.OutOrStdout(), successStyle.Render("Task updated."))
		},
	}
	cmd.Flags().StringP("description", "d", "", "Task description")
	cmd.Flags().IntP("priority", "p", 3, "Task priority (0, 1, 2, ...)")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Task tags")
	cmd.Flags().String("status", "", "Task status (todo, in-progress, done, cancelled, paused)")
	return cmd
}
