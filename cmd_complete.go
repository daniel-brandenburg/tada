package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func NewCompleteCmd(store *FileStore) *cobra.Command {
	return &cobra.Command{
		Use:   "complete [topic/]title",
		Short: "Mark a task as completed",
		Long:  "Mark a task as completed and archive it",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := strings.Join(args, " ")

			// Parse topic and title
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

			if err := store.CompleteTask(topic, title); err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error completing task: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}

			successStyle := lipgloss.NewStyle().Foreground(cliPrimary).Bold(true)
			fmt.Fprintln(cmd.OutOrStdout(), successStyle.Render(fmt.Sprintf("Task completed and archived: %s", title)))
			if topic != "" {
				topicStyle := lipgloss.NewStyle().Foreground(cliSecondary)
				fmt.Fprintln(cmd.OutOrStdout(), topicStyle.Render(fmt.Sprintf("Topic: %s", topic)))
			}
		},
	}
}
