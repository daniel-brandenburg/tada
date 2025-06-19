package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Show a detailed view of a single task by topic/title or ID
func NewShowCmd(store *FileStore) *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "show [topic/]title|id",
		Short: "Show details for a task",
		Long:  "Show a detailed view of a single task by topic/title or ID.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := strings.Join(args, " ")
			var topic, title, id string
			if len(args) == 1 && len(args[0]) == 5 {
				id = args[0]
			}
			if id == "" {
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
					base := filepath.Base(t.FilePath)
					if (id != "" && strings.HasPrefix(base, id)) || (t.Task.Title == title && t.Topic == topic) {
						found = t
					}
				}
			}
			if found == nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render("Task not found.")
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}

			if outputFormat == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				enc.Encode(found)
				return
			}
			if outputFormat == "yaml" {
				enc := yaml.NewEncoder(cmd.OutOrStdout())
				enc.Encode(found)
				return
			}

			// Pretty print task details
			header := lipgloss.NewStyle().Bold(true).Foreground(cliPrimary).Render(found.Task.Title)
			topicStyle := lipgloss.NewStyle().Foreground(cliSecondary)
			meta := fmt.Sprintf("Topic: %s\nPriority: %d\nStatus: %s\nTags: %s\nCreated: %s",
				found.Topic,
				found.Task.Priority,
				found.Task.Status,
				strings.Join(found.Task.Tags, ", "),
				found.Task.CreatedAt.Format("2006-01-02 15:04"),
			)
			fmt.Fprintln(cmd.OutOrStdout(), header)
			fmt.Fprintln(cmd.OutOrStdout(), topicStyle.Render(meta))
			if found.Task.Description != "" {
				descStyle := lipgloss.NewStyle().Foreground(cliMuted)
				fmt.Fprintln(cmd.OutOrStdout(), descStyle.Render("\n"+found.Task.Description))
			}
		},
	}
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: json, yaml, or pretty (default)")
	return cmd
}
