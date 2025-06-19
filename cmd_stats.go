package main

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Stats command: show counts by status, topic, tag
func NewStatsCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show statistics about your tasks",
		Long:  "Show statistics about your tasks: counts by status, topic, and tag.",
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := store.LoadAllTasks()
			if err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error loading tasks: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
				return
			}
			statusCounts := make(map[string]int)
			topicCounts := make(map[string]int)
			tagCounts := make(map[string]int)
			for topic, taskList := range tasks {
				topicCounts[topic] += len(taskList)
				for _, t := range taskList {
					statusCounts[string(t.Task.Status)]++
					for _, tag := range t.Task.Tags {
						tagCounts[tag]++
					}
				}
			}
			headStyle := lipgloss.NewStyle().Bold(true).Foreground(cliPrimary)
			fmt.Fprintln(cmd.OutOrStdout(), headStyle.Render("Task Statistics"))
			fmt.Fprintln(cmd.OutOrStdout(), "\nBy Status:")
			for _, status := range []string{"todo", "in-progress", "done", "paused", "cancelled"} {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n", status, statusCounts[status])
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nBy Topic:")
			topics := make([]string, 0, len(topicCounts))
			for topic := range topicCounts {
				topics = append(topics, topic)
			}
			sort.Strings(topics)
			for _, topic := range topics {
				if topic == "" {
					topic = "."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n", topic, topicCounts[topic])
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nBy Tag:")
			tags := make([]string, 0, len(tagCounts))
			for tag := range tagCounts {
				tags = append(tags, tag)
			}
			sort.Strings(tags)
			for _, tag := range tags {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n", tag, tagCounts[tag])
			}
		},
	}
	return cmd
}
