package deck

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Repository interface {
	Load() (Deck, error)
	Save(Deck) error
}

type Clock func() time.Time

type IDGenerator func(prefix string) string

type Service struct {
	repo  Repository
	now   Clock
	newID IDGenerator
}

type ServiceOption func(*Service)

func NewService(repo Repository, opts ...ServiceOption) *Service {
	service := &Service{
		repo:  repo,
		now:   time.Now,
		newID: RandomID,
	}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

func WithClock(clock Clock) ServiceOption {
	return func(service *Service) {
		if clock != nil {
			service.now = clock
		}
	}
}

func WithIDGenerator(generator IDGenerator) ServiceOption {
	return func(service *Service) {
		if generator != nil {
			service.newID = generator
		}
	}
}

func RandomID(prefix string) string {
	var bytes [4]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(bytes[:])
}

func (s *Service) Inbox(includeDone bool) ([]Task, error) {
	state, err := s.repo.Load()
	if err != nil {
		return []Task{}, fmt.Errorf("load deck: %w", err)
	}
	return SortByAttention(state.Tasks, includeDone), nil
}

func (s *Service) CreateTask(input NewTaskInput) (Task, error) {
	state, err := s.repo.Load()
	if err != nil {
		return Task{}, fmt.Errorf("load deck: %w", err)
	}

	task, err := state.AddTask(s.newID("tsk"), input, s.now())
	if err != nil {
		return Task{}, err
	}
	if err := s.repo.Save(state); err != nil {
		return Task{}, fmt.Errorf("save deck: %w", err)
	}
	return task, nil
}

func (s *Service) GetTask(id string) (Task, error) {
	id, err := s.ResolveTaskID(id)
	if err != nil {
		return Task{}, err
	}

	state, err := s.repo.Load()
	if err != nil {
		return Task{}, fmt.Errorf("load deck: %w", err)
	}
	task, ok := state.FindTask(id)
	if !ok {
		return Task{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	return task, nil
}

func (s *Service) UpdateTask(id string, input UpdateTaskInput) (Task, error) {
	id, err := s.ResolveTaskID(id)
	if err != nil {
		return Task{}, err
	}

	state, err := s.repo.Load()
	if err != nil {
		return Task{}, fmt.Errorf("load deck: %w", err)
	}

	task, err := state.UpdateTask(id, input, s.now())
	if err != nil {
		return Task{}, err
	}
	if err := s.repo.Save(state); err != nil {
		return Task{}, fmt.Errorf("save deck: %w", err)
	}
	return task, nil
}

func (s *Service) MarkDone(id string, nextAction *string) (Task, error) {
	return s.UpdateTask(id, UpdateTaskInput{
		Status:     string(StatusDone),
		NextAction: nextAction,
	})
}

func (s *Service) AddRun(id string, input AddRunInput) (Task, RunRecord, error) {
	id, err := s.ResolveTaskID(id)
	if err != nil {
		return Task{}, RunRecord{}, err
	}

	state, err := s.repo.Load()
	if err != nil {
		return Task{}, RunRecord{}, fmt.Errorf("load deck: %w", err)
	}

	task, run, err := state.AddRun(id, s.newID("run"), input, s.now())
	if err != nil {
		return Task{}, RunRecord{}, err
	}
	if err := s.repo.Save(state); err != nil {
		return Task{}, RunRecord{}, fmt.Errorf("save deck: %w", err)
	}
	return task, run, nil
}

func (s *Service) AddArtifact(id string, input AddArtifactInput) (Task, Artifact, error) {
	id, err := s.ResolveTaskID(id)
	if err != nil {
		return Task{}, Artifact{}, err
	}

	state, err := s.repo.Load()
	if err != nil {
		return Task{}, Artifact{}, fmt.Errorf("load deck: %w", err)
	}

	task, artifact, err := state.AddArtifact(id, input, s.now())
	if err != nil {
		return Task{}, Artifact{}, err
	}
	if err := s.repo.Save(state); err != nil {
		return Task{}, Artifact{}, fmt.Errorf("save deck: %w", err)
	}
	return task, artifact, nil
}

func (s *Service) ResolveTaskID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id != "" && id != "latest" {
		return id, nil
	}

	tasks, err := s.Inbox(false)
	if err != nil {
		return "", err
	}
	if len(tasks) == 0 {
		return "", ErrNoOpenTasks
	}
	return tasks[0].ID, nil
}

func SortByAttention(tasks []Task, includeDone bool) []Task {
	next := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if !includeDone && task.Status == StatusDone {
			continue
		}
		next = append(next, task)
	}

	sort.SliceStable(next, func(i, j int) bool {
		left := attentionRank(next[i].Status)
		right := attentionRank(next[j].Status)
		if left != right {
			return left < right
		}
		return next[i].UpdatedAt.After(next[j].UpdatedAt)
	})
	return next
}

func attentionRank(status Status) int {
	switch status {
	case StatusNeedsReview:
		return 0
	case StatusBlocked, StatusFailed:
		return 1
	case StatusRunning:
		return 2
	case StatusInbox, StatusReady:
		return 3
	case StatusDone:
		return 4
	default:
		return 5
	}
}
