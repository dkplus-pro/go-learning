package task

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	mu    sync.Mutex
	next  int
	tasks map[string]Task
}

func NewService() *Service {
	return &Service{tasks: make(map[string]Task)}
}

func (s *Service) CreateTask(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, validationError("title is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.next++
	task := Task{
		ID:        "task-" + strconvInt(s.next),
		Title:     title,
		Status:    "todo",
		CreatedAt: time.Now().UTC(),
	}
	s.tasks[task.ID] = task

	return task, nil
}

func (s *Service) GetTask(ctx context.Context, id string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, notFoundError("task %s not found", id)
	}

	return task, nil
}

func strconvInt(value int) string {
	return strconv.FormatInt(int64(value), 10)
}
