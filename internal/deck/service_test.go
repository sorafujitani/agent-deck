package deck

import (
	"slices"
	"testing"
	"time"
)

func TestServiceWorkflow(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	service := NewService(
		repo,
		WithClock(func() time.Time { return now }),
		WithIDGenerator(sequenceIDGenerator("tsk_1", "run_1")),
	)

	task, err := service.CreateTask(NewTaskInput{
		Goal:       "review PR",
		Repo:       "/tmp/repo",
		Context:    []string{"pull request"},
		NextAction: "read diff",
	})
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != "tsk_1" {
		t.Fatalf("task id = %s, want tsk_1", task.ID)
	}

	updated, run, err := service.AddRun(task.ID, AddRunInput{
		Agent:   "codex",
		Summary: "checked diff",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.ID != "run_1" {
		t.Fatalf("run id = %s, want run_1", run.ID)
	}
	if updated.Status != StatusNeedsReview {
		t.Fatalf("status = %s, want %s", updated.Status, StatusNeedsReview)
	}

	tasks, err := service.Inbox(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("tasks = %#v", tasks)
	}

	done, err := service.MarkDone(task.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if done.CompletedAt == nil {
		t.Fatal("completed_at should be set")
	}

	tasks, err = service.Inbox(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Fatalf("tasks = %#v, want none", tasks)
	}
}

func TestServiceResolvesLatestTask(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	service := NewService(
		repo,
		WithClock(func() time.Time { return now }),
		WithIDGenerator(sequenceIDGenerator("tsk_1", "tsk_2")),
	)
	if _, err := service.CreateTask(NewTaskInput{Goal: "first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateTask(NewTaskInput{Goal: "second"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateTask("tsk_1", UpdateTaskInput{Status: string(StatusNeedsReview)}); err != nil {
		t.Fatal(err)
	}

	id, err := service.ResolveTaskID("latest")
	if err != nil {
		t.Fatal(err)
	}
	if id != "tsk_1" {
		t.Fatalf("id = %s, want tsk_1", id)
	}

	task, err := service.GetTask("")
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != "tsk_1" {
		t.Fatalf("task id = %s, want tsk_1", task.ID)
	}
}

func TestSortByAttention(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	tasks := []Task{
		{ID: "done", Status: StatusDone, UpdatedAt: now.Add(3 * time.Minute)},
		{ID: "running", Status: StatusRunning, UpdatedAt: now.Add(2 * time.Minute)},
		{ID: "review", Status: StatusNeedsReview, UpdatedAt: now},
		{ID: "blocked", Status: StatusBlocked, UpdatedAt: now.Add(time.Minute)},
	}

	got := SortByAttention(tasks, false)
	ids := make([]string, 0, len(got))
	for _, task := range got {
		ids = append(ids, task.ID)
	}
	want := []string{"review", "blocked", "running"}
	if !slices.Equal(ids, want) {
		t.Fatalf("ids = %#v, want %#v", ids, want)
	}
}

type memoryRepository struct {
	state Deck
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{state: NewDeck()}
}

func (r *memoryRepository) Load() (Deck, error) {
	return r.state, nil
}

func (r *memoryRepository) Save(state Deck) error {
	r.state = state
	return nil
}

func sequenceIDGenerator(ids ...string) IDGenerator {
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
