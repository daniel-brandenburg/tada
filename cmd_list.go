package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// CLI color palette
var (
	cliPrimary   = lipgloss.Color("12") // blue
	cliSecondary = lipgloss.Color("10") // green
	cliError     = lipgloss.Color("9")  // red
	cliMuted     = lipgloss.Color("8")  // gray
)

func NewListCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  "List all tasks with optional filtering, searching, and sorting",
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := store.LoadAllTasks()
			if err != nil {
				styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error loading tasks: %v", err))
				fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
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

			// Search by query if specified (now includes tags and topic)
			query, _ := cmd.Flags().GetString("search")
			if query != "" {
				filtered := make(map[string][]*TaskWithPath)
				q := strings.ToLower(query)
				for path, taskList := range tasks {
					for _, task := range taskList {
						if strings.Contains(strings.ToLower(task.Task.Title), q) ||
							strings.Contains(strings.ToLower(task.Task.Description), q) ||
							strings.Contains(strings.ToLower(strings.Join(task.Task.Tags, ",")), q) ||
							strings.Contains(strings.ToLower(path), q) {
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

			// Simple output flag
			simple, _ := cmd.Flags().GetBool("simple")

			if simple {
				for _, taskList := range tasks {
					for _, taskWithPath := range taskList {
						id := ""
						base := filepath.Base(taskWithPath.FilePath)
						if len(base) > 5 {
							id = base[:5]
						}
						statusStyle := lipgloss.NewStyle().Foreground(cliPrimary)
						titleStyle := lipgloss.NewStyle().Bold(true)
						fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", id, titleStyle.Render(taskWithPath.Task.Title), statusStyle.Render(string(taskWithPath.Task.Status)))
					}
				}
				return
			}

			// Pretty output
			headStyle := lipgloss.NewStyle().Bold(true).Foreground(cliPrimary)
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, headStyle.Render("ID\tTOPIC\tTITLE\tPRIORITY\tSTATUS\tTAGS\tCREATED"))

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
					id := ""
					base := filepath.Base(taskWithPath.FilePath)
					if len(base) > 5 {
						id = base[:5]
					}
					statusStyle := lipgloss.NewStyle().Foreground(cliPrimary)
					titleStyle := lipgloss.NewStyle().Bold(true)
					tagsStyle := lipgloss.NewStyle().Foreground(cliSecondary)
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						id,
						topicDisplay,
						titleStyle.Render(task.Title),
						priority,
						statusStyle.Render(string(task.Status)),
						tagsStyle.Render(tagsStr),
						task.CreatedAt.Format("2006-01-02 15:04"),
					)
				}
			}
			w.Flush()
		},
	}
	cmd.Flags().StringP("status", "s", "", "Filter by status (todo, in-progress, done, cancelled, paused)")
	cmd.Flags().String("sort", "created", "Sort by: created, priority, title, status")
	cmd.Flags().Bool("simple", false, "Print simple output (id, title, status)")
	cmd.Flags().StringP("search", "q", "", "Search for tasks by title, description, tags, or topic")
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
