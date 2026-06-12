package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrValidation        = errors.New("validation error")
	ErrNotFound          = errors.New("task not found")
	ErrInvalidTransition = errors.New("invalid task status transition")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTask(ctx context.Context, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, fmt.Errorf("%w: title is required", ErrValidation)
	}

	return s.repo.Create(ctx, Task{
		Title:  title,
		Status: StatusTodo,
	})
}

func (s *Service) GetTask(ctx context.Context, id string) (Task, error) {
	if strings.TrimSpace(id) == "" {
		return Task{}, fmt.Errorf("%w: task id is required", ErrValidation)
	}

	return s.repo.Find(ctx, id)
}

func (s *Service) ListTasks(ctx context.Context) ([]Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) MoveTask(ctx context.Context, id string, next Status) (Task, error) {
	if !next.Valid() {
		return Task{}, fmt.Errorf("%w: unknown status %q", ErrValidation, next)
	}

	task, err := s.GetTask(ctx, id)
	if err != nil {
		return Task{}, err
	}

	if !CanTransition(task.Status, next) {
		return Task{}, fmt.Errorf("%w: %s to %s", ErrInvalidTransition, task.Status, next)
	}

	task.Status = next
	return s.repo.Update(ctx, task)
}
