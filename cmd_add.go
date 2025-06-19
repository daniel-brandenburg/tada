package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Use CLI color palette from cmd_list.go

func NewAddCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [topic/]title",
		Short: "Add a new task",
		Long:  "Add a new task with optional topic path and description, priority, tags, and status.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			title := strings.Join(args, " ")

			description, _ := cmd.Flags().GetString("description")
			priority, _ := cmd.Flags().GetInt("priority")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			status, _ := cmd.Flags().GetString("status")

			// Parse topic from title if contains "/"
			var topic, taskTitle string
			if strings.Contains(title, "/") {
				parts := strings.Split(title, "/")
				if len(parts) >= 2 {
					topic = filepath.Join(parts[:len(parts)-1]...)
					taskTitle = parts[len(parts)-1]
				} else {
					taskTitle = title
				}
			} else {
				taskTitle = title
			}

			ts := StatusTodo
			if status != "" {
				ts = TaskStatus(status)
			}

			task := &Task{
				Title:       taskTitle,
				Description: description,
				Priority:    priority,
				Tags:        tags,
				Status:      ts,
			}

			if err := store.SaveTask(topic, task); err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error adding task: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}

			successStyle := lipgloss.NewStyle().Foreground(cliPrimary).Bold(true)
			fmt.Fprintln(cmd.OutOrStdout(), successStyle.Render(fmt.Sprintf("Task added: %s", task.Title)))
			if topic != "" {
				topicStyle := lipgloss.NewStyle().Foreground(cliSecondary)
				fmt.Fprintln(cmd.OutOrStdout(), topicStyle.Render(fmt.Sprintf("Topic: %s", topic)))
			}
		},
	}
	cmd.Flags().StringP("description", "d", "", "Task description")
	cmd.Flags().IntP("priority", "p", 3, "Task priority (0, 1, 2, ...)")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Task tags")
	cmd.Flags().String("status", "", "Task status (todo, in-progress, done, cancelled, paused)")
	return cmd
}
