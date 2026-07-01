package deck

import (
	"testing"
	"time"
)

func TestDeckLifecycle(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	deck := NewDeck()

	task, err := deck.AddTask(NewTaskInput{
		Goal:       "review PR",
		Repo:       "/tmp/repo",
		Context:    []string{"  issue #1  ", ""},
		NextAction: "check diff",
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != StatusInbox {
		t.Fatalf("status = %s, want %s", task.Status, StatusInbox)
	}
	if len(task.Context) != 1 || task.Context[0] != "issue #1" {
		t.Fatalf("context = %#v", task.Context)
	}

	updated, err := deck.UpdateTask(task.ID, UpdateTaskInput{Status: string(StatusRunning)}, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != StatusRunning {
		t.Fatalf("status = %s, want %s", updated.Status, StatusRunning)
	}

	updated, run, err := deck.AddRun(task.ID, AddRunInput{Agent: "codex", Summary: "looked at diff"}, now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if run.Agent != "codex" {
		t.Fatalf("agent = %s", run.Agent)
	}
	if updated.Status != StatusNeedsReview {
		t.Fatalf("status = %s, want %s", updated.Status, StatusNeedsReview)
	}

	updated, artifact, err := deck.AddArtifact(task.ID, AddArtifactInput{Kind: "diff", Path: "diff.patch"}, now.Add(3*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Kind != "diff" || len(updated.Artifacts) != 1 {
		t.Fatalf("artifact = %#v, task = %#v", artifact, updated.Artifacts)
	}

	done, err := deck.UpdateTask(task.ID, UpdateTaskInput{Status: string(StatusDone)}, now.Add(4*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if done.CompletedAt == nil {
		t.Fatal("completed_at should be set")
	}
}

func TestUpdateRejectsInvalidStatus(t *testing.T) {
	deck := NewDeck()
	task, err := deck.AddTask(NewTaskInput{Goal: "ship"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	_, err = deck.UpdateTask(task.ID, UpdateTaskInput{Status: "waiting"}, time.Now())
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}
