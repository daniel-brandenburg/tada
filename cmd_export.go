package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func NewExportCmd(store Storage) *cobra.Command {
	var format, output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all tasks to a single file (csv, json, or md)",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := store.LoadAllTasks()
			if err != nil {
				return fmt.Errorf("failed to load tasks: %w", err)
			}
			var out *os.File
			if output == "" || output == "-" {
				out = os.Stdout
			} else {
				out, err = os.Create(output)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer out.Close()
			}
			switch format {
			case "json":
				all := flattenTasks(tasks)
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(all)
			case "csv":
				return exportCSV(out, tasks)
			case "md", "markdown":
				return exportMarkdown(out, tasks)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Export format: csv, json, md")
	cmd.Flags().StringVarP(&output, "output", "o", "-", "Output file (default: stdout)")
	return cmd
}

func flattenTasks(tasks map[string][]*TaskWithPath) []*Task {
	var all []*Task
	for _, list := range tasks {
		for _, t := range list {
			all = append(all, t.Task)
		}
	}
	return all
}

func exportCSV(out *os.File, tasks map[string][]*TaskWithPath) error {
	w := csv.NewWriter(out)
	defer w.Flush()
	w.Write([]string{"Title", "Description", "Priority", "Status", "Tags", "CreatedAt", "CompletedAt"})
	for _, list := range tasks {
		for _, t := range list {
			task := t.Task
			w.Write([]string{
				task.Title,
				task.Description,
				fmt.Sprintf("%d", task.Priority),
				string(task.Status),
				strings.Join(task.Tags, ","),
				task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				formatCompletedAt(task.CompletedAt),
			})
		}
	}
	return w.Error()
}

func formatCompletedAt(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z07:00")
}

func exportMarkdown(out *os.File, tasks map[string][]*TaskWithPath) error {
	for _, list := range tasks {
		for _, t := range list {
			task := t.Task
			fmt.Fprintf(out, "# %s\n\n", task.Title)
			fmt.Fprintf(out, "- **Description:** %s\n", task.Description)
			fmt.Fprintf(out, "- **Priority:** %d\n", task.Priority)
			fmt.Fprintf(out, "- **Status:** %s\n", task.Status)
			fmt.Fprintf(out, "- **Tags:** %s\n", strings.Join(task.Tags, ", "))
			fmt.Fprintf(out, "- **Created At:** %s\n", task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
			if task.CompletedAt != nil {
				fmt.Fprintf(out, "- **Completed At:** %s\n", task.CompletedAt.Format("2006-01-02T15:04:05Z07:00"))
			}
			fmt.Fprintln(out)
		}
	}
	return nil
}

// ExportTasksToFile exports a flat slice of tasks to the given file in the specified format (csv, json, md)
func ExportTasksToFile(tasks []*TaskWithPath, format, filePath string) error {
	var out *os.File
	var err error
	if filePath == "" || filePath == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
	}
	switch format {
	case "json":
		var all []*Task
		for _, t := range tasks {
			all = append(all, t.Task)
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	case "csv":
		w := csv.NewWriter(out)
		defer w.Flush()
		w.Write([]string{"Title", "Description", "Priority", "Status", "Tags", "CreatedAt", "CompletedAt"})
		for _, t := range tasks {
			w.Write([]string{
				t.Task.Title,
				t.Task.Description,
				fmt.Sprintf("%d", t.Task.Priority),
				string(t.Task.Status),
				strings.Join(t.Task.Tags, ","),
				t.Task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				formatCompletedAt(t.Task.CompletedAt),
			})
		}
		return w.Error()
	case "md", "markdown":
		for _, t := range tasks {
			task := t.Task
			fmt.Fprintf(out, "# %s\n\n", task.Title)
			fmt.Fprintf(out, "- **Description:** %s\n", task.Description)
			fmt.Fprintf(out, "- **Priority:** %d\n", task.Priority)
			fmt.Fprintf(out, "- **Status:** %s\n", task.Status)
			fmt.Fprintf(out, "- **Tags:** %s\n", strings.Join(task.Tags, ", "))
			fmt.Fprintf(out, "- **Created At:** %s\n", task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
			if task.CompletedAt != nil {
				fmt.Fprintf(out, "- **Completed At:** %s\n", task.CompletedAt.Format("2006-01-02T15:04:05Z07:00"))
			}
			fmt.Fprintln(out)
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
