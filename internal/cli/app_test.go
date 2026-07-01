package cli

import (
	"bytes"
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
