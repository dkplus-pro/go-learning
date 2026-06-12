package task

import (
	"context"
	"errors"
	"testing"
)

func TestServiceCreateTaskRequiresTitle(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateTask(context.Background(), " ")

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("CreateTask() error = %v, want ErrValidation", err)
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "todo to doing", from: StatusTodo, to: StatusDoing, want: true},
		{name: "doing to done", from: StatusDoing, to: StatusDone, want: true},
		{name: "done to todo rejected", from: StatusDone, to: StatusTodo, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransition(tt.from, tt.to); got != tt.want {
				t.Fatalf("CanTransition() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeRepository struct{}

func (fakeRepository) Create(context.Context, Task) (Task, error) {
	return Task{}, nil
}

func (fakeRepository) Find(context.Context, string) (Task, error) {
	return Task{}, ErrNotFound
}

func (fakeRepository) List(context.Context) ([]Task, error) {
	return nil, nil
}

func (fakeRepository) Update(context.Context, Task) (Task, error) {
	return Task{}, nil
}
