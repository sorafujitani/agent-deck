package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sorafujitani/agent-deck/internal/deck"
	"github.com/sorafujitani/agent-deck/internal/storage"
)

func TestAppWorkflow(t *testing.T) {
	store := storage.NewJSONStore(filepath.Join(t.TempDir(), "deck.json"))
	service := deck.NewService(
		store,
		deck.WithClock(func() time.Time {
			return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
		}),
		deck.WithIDGenerator(sequenceIDGenerator("tsk_1", "run_1")),
	)
	app := NewApp(service, store.Path())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := app.Run(
		[]string{"new", "review", "PR", "--repo", "/tmp/repo", "--context", "pull request"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("new exit = %d, stderr = %s", code, stderr.String())
	}
	id := strings.Fields(stdout.String())[1]

	stdout.Reset()
	stderr.Reset()
	code = app.Run(
		[]string{"run", id, "--agent", "codex", "--summary", "checked diff"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"inbox"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("inbox exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "needs-review") {
		t.Fatalf("inbox output = %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"done", id}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("done exit = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"inbox"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("inbox exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No tasks.") {
		t.Fatalf("inbox output = %s", stdout.String())
	}
}

func TestAppLatestAndJSON(t *testing.T) {
	store := storage.NewJSONStore(filepath.Join(t.TempDir(), "deck.json"))
	service := deck.NewService(
		store,
		deck.WithClock(func() time.Time {
			return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
		}),
		deck.WithIDGenerator(sequenceIDGenerator("tsk_1", "run_1")),
	)
	app := NewApp(service, store.Path())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := app.Run([]string{"new", "review", "PR", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("new exit = %d, stderr = %s", code, stderr.String())
	}
	var created deck.Task
	if err := json.Unmarshal(stdout.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID != "tsk_1" {
		t.Fatalf("created id = %s, want tsk_1", created.ID)
	}

	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"show", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("show exit = %d, stderr = %s", code, stderr.String())
	}
	var shown deck.Task
	if err := json.Unmarshal(stdout.Bytes(), &shown); err != nil {
		t.Fatal(err)
	}
	if shown.ID != "tsk_1" {
		t.Fatalf("shown id = %s, want tsk_1", shown.ID)
	}

	stdout.Reset()
	stderr.Reset()
	code = app.Run(
		[]string{"run", "latest", "--agent", "codex", "--summary", "checked diff", "--json"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr = %s", code, stderr.String())
	}
	var runResult struct {
		Task deck.Task      `json:"task"`
		Run  deck.RunRecord `json:"run"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &runResult); err != nil {
		t.Fatal(err)
	}
	if runResult.Task.Status != deck.StatusNeedsReview {
		t.Fatalf("status = %s, want %s", runResult.Task.Status, deck.StatusNeedsReview)
	}

	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"artifact", "./review.md", "--kind", "report", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("artifact exit = %d, stderr = %s", code, stderr.String())
	}
	var artifactResult struct {
		Task     deck.Task     `json:"task"`
		Artifact deck.Artifact `json:"artifact"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &artifactResult); err != nil {
		t.Fatal(err)
	}
	if artifactResult.Artifact.Path != "./review.md" {
		t.Fatalf("artifact path = %s, want ./review.md", artifactResult.Artifact.Path)
	}
}

func sequenceIDGenerator(ids ...string) deck.IDGenerator {
	index := 0
	return func(prefix string) string {
		if index >= len(ids) {
			return prefix + "_extra"
		}
		id := ids[index]
		index++
		return id
	}
}
