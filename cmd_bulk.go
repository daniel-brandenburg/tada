package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Bulk operations: delete, complete, move multiple tasks by query or tag
func NewBulkCmd(store *FileStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk operations on tasks",
		Long:  "Perform bulk operations (delete, complete, move) on tasks by search, tag, or status.",
	}

	var (
		bulkDelete   bool
		bulkComplete bool
		bulkMove     string
		bulkQuery    string
		bulkTag      string
		bulkStatus   string
	)

	cmd.Flags().BoolVar(&bulkDelete, "delete", false, "Delete matching tasks")
	cmd.Flags().BoolVar(&bulkComplete, "complete", false, "Complete (and archive) matching tasks")
	cmd.Flags().StringVar(&bulkMove, "move", "", "Move matching tasks to this topic")
	cmd.Flags().StringVar(&bulkQuery, "search", "", "Search query (title, description, tags, topic)")
	cmd.Flags().StringVar(&bulkTag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&bulkStatus, "status", "", "Filter by status")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		tasks, err := store.LoadAllTasks()
		if err != nil {
			styledErr := lipgloss.NewStyle().Foreground(cliError).Render(fmt.Sprintf("Error loading tasks: %v", err))
			fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
			return
		}
		var toProcess []*TaskWithPath
		for _, taskList := range tasks {
			for _, t := range taskList {
				match := true
				if bulkQuery != "" {
					q := strings.ToLower(bulkQuery)
					if !(strings.Contains(strings.ToLower(t.Task.Title), q) ||
						strings.Contains(strings.ToLower(t.Task.Description), q) ||
						strings.Contains(strings.ToLower(strings.Join(t.Task.Tags, ",")), q) ||
						strings.Contains(strings.ToLower(t.Topic), q)) {
						match = false
					}
				}
				if bulkTag != "" && !containsString(t.Task.Tags, bulkTag) {
					match = false
				}
				if bulkStatus != "" && string(t.Task.Status) != bulkStatus {
					match = false
				}
				if match {
					toProcess = append(toProcess, t)
				}
			}
		}
		if len(toProcess) == 0 {
			styledErr := lipgloss.NewStyle().Foreground(cliError).Render("No matching tasks found.")
			fmt.Fprintln(cmd.ErrOrStderr(), styledErr)
			return
		}
		for _, t := range toProcess {
			if bulkDelete {
				os.Remove(t.FilePath)
			}
			if bulkComplete {
				store.CompleteTask(t.Topic, t.Task.Title)
			}
			if bulkMove != "" {
				newDir := store.basePath + "/tasks/" + bulkMove
				os.MkdirAll(newDir, 0755)
				newPath := newDir + "/" + filepath.Base(t.FilePath)
				os.Rename(t.FilePath, newPath)
			}
		}
		successStyle := lipgloss.NewStyle().Foreground(cliPrimary).Bold(true)
		fmt.Fprintln(cmd.OutOrStdout(), successStyle.Render(fmt.Sprintf("Bulk operation complete on %d tasks.", len(toProcess))))
	}
	return cmd
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
