package task

import (
	"context"
	"errors"
	"testing"
)

func TestServiceCreateTask(t *testing.T) {
	service := NewService(NewMemoryRepository())

	task, err := service.CreateTask(context.Background(), "  Read Go docs  ")
	if err != nil {
		t.Fatalf("CreateTask() unexpected error: %v", err)
	}

	if task.ID == "" {
		t.Fatal("CreateTask() did not assign an id")
	}
	if task.Title != "Read Go docs" {
		t.Fatalf("Title = %q, want Read Go docs", task.Title)
	}
	if task.Status != StatusTodo {
		t.Fatalf("Status = %q, want %q", task.Status, StatusTodo)
	}
}

func TestServiceCreateTaskRequiresTitle(t *testing.T) {
	service := NewService(NewMemoryRepository())

	_, err := service.CreateTask(context.Background(), " ")

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("CreateTask() error = %v, want ErrValidation", err)
	}
}

func TestServiceMoveTask(t *testing.T) {
	service := NewService(NewMemoryRepository())
	task, err := service.CreateTask(context.Background(), "Write tests")
	if err != nil {
		t.Fatalf("CreateTask() unexpected error: %v", err)
	}

	moved, err := service.MoveTask(context.Background(), task.ID, StatusDoing)
	if err != nil {
		t.Fatalf("MoveTask() unexpected error: %v", err)
	}

	if moved.Status != StatusDoing {
		t.Fatalf("Status = %q, want %q", moved.Status, StatusDoing)
	}
}

func TestServiceRejectsInvalidTransition(t *testing.T) {
	service := NewService(NewMemoryRepository())
	task, err := service.CreateTask(context.Background(), "Ship API")
	if err != nil {
		t.Fatalf("CreateTask() unexpected error: %v", err)
	}

	done, err := service.MoveTask(context.Background(), task.ID, StatusDone)
	if err != nil {
		t.Fatalf("MoveTask() unexpected error: %v", err)
	}

	_, err = service.MoveTask(context.Background(), done.ID, StatusTodo)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("MoveTask() error = %v, want ErrInvalidTransition", err)
	}
}
