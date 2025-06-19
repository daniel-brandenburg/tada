package main

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func NewListCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  "List all tasks with optional filtering and sorting",
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := store.LoadAllTasks()
			if err != nil {
				fmt.Printf("Error loading tasks: %v\n", err)
				return
			}

			// Filter by status if specified
			status, _ := cmd.Flags().GetString("status")
			if status != "" {
				filtered := make(map[string][]*TaskWithPath)
				for path, taskList := range tasks {
					for _, task := range taskList {
						if string(task.Task.Status) == status {
							filtered[path] = append(filtered[path], task)
						}
					}
				}
				tasks = filtered
			}

			// Sort tasks
			sortBy, _ := cmd.Flags().GetString("sort")
			for path := range tasks {
				sortTasks(tasks[path], sortBy)
			}

			// Display in table format
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TOPIC\tTITLE\tPRIORITY\tSTATUS\tTAGS\tCREATED")

			for topic, taskList := range tasks {
				if len(taskList) == 0 {
					continue
				}

				topicDisplay := topic
				if topicDisplay == "" {
					topicDisplay = "."
				}

				for _, taskWithPath := range taskList {
					task := taskWithPath.Task
					tagsStr := strings.Join(task.Tags, ",")
					if tagsStr == "" {
						tagsStr = "-"
					}

					priority := fmt.Sprintf("%d", task.Priority)

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
						topicDisplay,
						task.Title,
						priority,
						task.Status,
						tagsStr,
						task.CreatedAt.Format("2006-01-02 15:04"),
					)
				}
			}
			w.Flush()
		},
	}
	cmd.Flags().StringP("status", "s", "", "Filter by status (todo, in-progress, done, cancelled, paused)")
	cmd.Flags().String("sort", "created", "Sort by: created, priority, title, status")
	return cmd
}

func sortTasks(tasks []*TaskWithPath, sortBy string) {
	sort.Slice(tasks, func(i, j int) bool {
		switch sortBy {
		case "priority":
			// Lower numbers = higher priority
			return tasks[i].Task.Priority < tasks[j].Task.Priority
		case "title":
			return tasks[i].Task.Title < tasks[j].Task.Title
		case "status":
			return tasks[i].Task.Status < tasks[j].Task.Status
		default: // created
			return tasks[i].Task.CreatedAt.Before(tasks[j].Task.CreatedAt)
		}
	})
}
