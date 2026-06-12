package task

import "time"

type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusDoing, StatusDone:
		return true
	default:
		return false
	}
}

func CanTransition(from, to Status) bool {
	if from == to {
		return true
	}

	switch from {
	case StatusTodo:
		return to == StatusDoing || to == StatusDone
	case StatusDoing:
		return to == StatusDone
	default:
		return false
	}
}
