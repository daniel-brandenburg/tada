package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [topic/]title",
	Short: "Add a new task",
	Long:  "Add a new task with optional topic path and description",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")

		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetInt("priority")
		tags, _ := cmd.Flags().GetStringSlice("tags")

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

		task := &Task{
			Title:       taskTitle,
			Description: description,
			Priority:    priority,
			Tags:        tags,
			Status:      StatusTodo,
		}

		store := NewFileStore()
		if err := store.SaveTask(topic, task); err != nil {
			fmt.Printf("Error adding task: %v\n", err)
			return
		}

		fmt.Printf("Task added: %s\n", task.Title)
		if topic != "" {
			fmt.Printf("Topic: %s\n", topic)
		}
	},
}

func init() {
	addCmd.Flags().StringP("description", "d", "", "Task description")
	addCmd.Flags().IntP("priority", "p", 3, "Task priority (0, 1, 2, ...)")
	addCmd.Flags().StringSliceP("tags", "t", []string{}, "Task tags")
}
