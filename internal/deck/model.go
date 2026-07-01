package deck

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrArtifactPathRequired = errors.New("artifact path is required")
	ErrGoalRequired         = errors.New("goal is required")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrTaskNotFound         = errors.New("task not found")
)

type Status string

const (
	StatusInbox       Status = "inbox"
	StatusReady       Status = "ready"
	StatusRunning     Status = "running"
	StatusNeedsReview Status = "needs-review"
	StatusBlocked     Status = "blocked"
	StatusFailed      Status = "failed"
	StatusDone        Status = "done"
)

var validStatuses = map[Status]struct{}{
	StatusInbox:       {},
	StatusReady:       {},
	StatusRunning:     {},
	StatusNeedsReview: {},
	StatusBlocked:     {},
	StatusFailed:      {},
	StatusDone:        {},
}

type Deck struct {
	Version int    `json:"version"`
	Tasks   []Task `json:"tasks"`
}

type Task struct {
	ID          string      `json:"id"`
	Goal        string      `json:"goal"`
	Status      Status      `json:"status"`
	Repo        string      `json:"repo,omitempty"`
	Issue       string      `json:"issue,omitempty"`
	PR          string      `json:"pr,omitempty"`
	Context     []string    `json:"context,omitempty"`
	Runs        []RunRecord `json:"runs,omitempty"`
	Artifacts   []Artifact  `json:"artifacts,omitempty"`
	NextAction  string      `json:"next_action,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

type RunRecord struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent"`
	Summary   string    `json:"summary,omitempty"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

type Artifact struct {
	Kind      string    `json:"kind"`
	Path      string    `json:"path"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type NewTaskInput struct {
	Goal       string
	Repo       string
	Issue      string
	PR         string
	Context    []string
	NextAction string
}

type UpdateTaskInput struct {
	Status     string
	NextAction *string
	Context    []string
}

type AddRunInput struct {
	Agent   string
	Summary string
}

type AddArtifactInput struct {
	Kind string
	Path string
	Note string
}

func NewDeck() Deck {
	return Deck{Version: 1, Tasks: []Task{}}
}

func (d *Deck) AddTask(id string, input NewTaskInput, now time.Time) (Task, error) {
	goal := strings.TrimSpace(input.Goal)
	if goal == "" {
		return Task{}, ErrGoalRequired
	}

	task := Task{
		ID:         strings.TrimSpace(id),
		Goal:       goal,
		Status:     StatusInbox,
		Repo:       strings.TrimSpace(input.Repo),
		Issue:      strings.TrimSpace(input.Issue),
		PR:         strings.TrimSpace(input.PR),
		Context:    compactStrings(input.Context),
		NextAction: strings.TrimSpace(input.NextAction),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	d.Tasks = append(d.Tasks, task)
	return task, nil
}

func (d *Deck) UpdateTask(id string, input UpdateTaskInput, now time.Time) (Task, error) {
	task, ok := d.FindTask(id)
	if !ok {
		return Task{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	if input.Status != "" {
		status, err := ParseStatus(input.Status)
		if err != nil {
			return Task{}, err
		}
		task.Status = status
		if status == StatusDone {
			completedAt := now
			task.CompletedAt = &completedAt
		} else {
			task.CompletedAt = nil
		}
	}
	if input.NextAction != nil {
		task.NextAction = strings.TrimSpace(*input.NextAction)
	}
	task.Context = append(task.Context, compactStrings(input.Context)...)
	task.UpdatedAt = now
	d.replaceTask(task)
	return task, nil
}

func (d *Deck) AddRun(id string, runID string, input AddRunInput, now time.Time) (Task, RunRecord, error) {
	task, ok := d.FindTask(id)
	if !ok {
		return Task{}, RunRecord{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	agent := strings.TrimSpace(input.Agent)
	if agent == "" {
		agent = "agent"
	}

	run := RunRecord{
		ID:        strings.TrimSpace(runID),
		Agent:     agent,
		Summary:   strings.TrimSpace(input.Summary),
		StartedAt: now,
		EndedAt:   now,
	}
	task.Runs = append(task.Runs, run)
	task.Status = StatusNeedsReview
	task.UpdatedAt = now
	d.replaceTask(task)
	return task, run, nil
}

func (d *Deck) AddArtifact(id string, input AddArtifactInput, now time.Time) (Task, Artifact, error) {
	task, ok := d.FindTask(id)
	if !ok {
		return Task{}, Artifact{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	path := strings.TrimSpace(input.Path)
	if path == "" {
		return Task{}, Artifact{}, ErrArtifactPathRequired
	}
	kind := strings.TrimSpace(input.Kind)
	if kind == "" {
		kind = "file"
	}

	artifact := Artifact{
		Kind:      kind,
		Path:      path,
		Note:      strings.TrimSpace(input.Note),
		CreatedAt: now,
	}
	task.Artifacts = append(task.Artifacts, artifact)
	task.UpdatedAt = now
	d.replaceTask(task)
	return task, artifact, nil
}

func (d *Deck) FindTask(id string) (Task, bool) {
	id = strings.TrimSpace(id)
	for _, task := range d.Tasks {
		if task.ID == id {
			return task, true
		}
	}
	return Task{}, false
}

func (d *Deck) replaceTask(next Task) {
	for i, task := range d.Tasks {
		if task.ID == next.ID {
			d.Tasks[i] = next
			return
		}
	}
}

func ParseStatus(value string) (Status, error) {
	status := Status(strings.TrimSpace(value))
	if _, ok := validStatuses[status]; !ok {
		return "", fmt.Errorf("%w: %s", ErrInvalidStatus, value)
	}
	return status, nil
}

func compactStrings(values []string) []string {
	next := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			next = append(next, value)
		}
	}
	return next
}
