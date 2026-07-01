package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sorafujitani/agent-deck/internal/deck"
)

func PrintInbox(w io.Writer, tasks []deck.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks.")
		return
	}

	for _, task := range tasks {
		fmt.Fprintf(w, "%s  %-12s  %s\n", task.ID, task.Status, task.Goal)
		if task.NextAction != "" {
			fmt.Fprintf(w, "  next: %s\n", task.NextAction)
		}
	}
}

func PrintTask(w io.Writer, task deck.Task) {
	fmt.Fprintf(w, "%s  %s\n", task.ID, task.Goal)
	fmt.Fprintf(w, "status: %s\n", task.Status)
	printIfSet(w, "repo", task.Repo)
	printIfSet(w, "issue", task.Issue)
	printIfSet(w, "pr", task.PR)
	printIfSet(w, "next", task.NextAction)
	fmt.Fprintf(w, "created: %s\n", formatTime(task.CreatedAt))
	fmt.Fprintf(w, "updated: %s\n", formatTime(task.UpdatedAt))
	if task.CompletedAt != nil {
		fmt.Fprintf(w, "completed: %s\n", formatTime(*task.CompletedAt))
	}

	if len(task.Context) > 0 {
		fmt.Fprintln(w, "\ncontext:")
		for _, item := range task.Context {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(task.Runs) > 0 {
		fmt.Fprintln(w, "\nruns:")
		for _, run := range task.Runs {
			fmt.Fprintf(w, "- %s  %s  %s\n", run.ID, run.Agent, formatTime(run.EndedAt))
			if run.Summary != "" {
				fmt.Fprintf(w, "  %s\n", run.Summary)
			}
		}
	}
	if len(task.Artifacts) > 0 {
		fmt.Fprintln(w, "\nartifacts:")
		for _, artifact := range task.Artifacts {
			line := fmt.Sprintf("- %s  %s", artifact.Kind, artifact.Path)
			if artifact.Note != "" {
				line += "  " + artifact.Note
			}
			fmt.Fprintln(w, line)
		}
	}
}

func PrintHelp(w io.Writer) {
	fmt.Fprint(w, strings.TrimSpace(`
deck is a task-first CLI for tracking agent work.

Usage:
  deck inbox [--all] [--repo PATH] [--json]
  deck new <goal> [--repo PATH] [--issue URL] [--pr URL] [--context TEXT] [--next TEXT] [--json]
  deck show [<task-id>|latest] [--repo PATH] [--json]
  deck update [<task-id>|latest] [--status STATUS] [--repo PATH] [--context TEXT] [--next TEXT] [--json]
  deck run [<task-id>|latest] [--agent NAME] [--repo PATH] [--summary TEXT] [--json]
  deck artifact [<task-id>|latest] <path> [--kind KIND] [--repo PATH] [--note TEXT] [--json]
  deck done [<task-id>|latest] [--repo PATH] [--next TEXT] [--json]
  deck path

Statuses:
  inbox, ready, running, needs-review, blocked, failed, done
`)+"\n")
}

func printIfSet(w io.Writer, label, value string) {
	if value != "" {
		fmt.Fprintf(w, "%s: %s\n", label, value)
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Local().Format("2006-01-02 15:04:05")
}

func printJSON(stdout, stderr io.Writer, value any) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintf(stderr, "encode JSON: %v\n", err)
		return 1
	}
	return 0
}
