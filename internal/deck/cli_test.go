package deck

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIWorkflow(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "deck.json"))
	now := func() time.Time {
		return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"new", "review", "PR", "--repo", "/tmp/repo", "--context", "pull request"}, &stdout, &stderr, store, now)
	if code != 0 {
		t.Fatalf("new exit = %d, stderr = %s", code, stderr.String())
	}
	id := strings.Fields(stdout.String())[1]

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"run", id, "--agent", "codex", "--summary", "checked diff"}, &stdout, &stderr, store, now)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"inbox"}, &stdout, &stderr, store, now)
	if code != 0 {
		t.Fatalf("inbox exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "needs-review") {
		t.Fatalf("inbox output = %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"done", id}, &stdout, &stderr, store, now)
	if code != 0 {
		t.Fatalf("done exit = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"inbox"}, &stdout, &stderr, store, now)
	if code != 0 {
		t.Fatalf("inbox exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No tasks.") {
		t.Fatalf("inbox output = %s", stdout.String())
	}
}
