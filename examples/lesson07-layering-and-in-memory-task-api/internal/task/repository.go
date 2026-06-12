package task

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Repository interface {
	Create(context.Context, Task) (Task, error)
	Find(context.Context, string) (Task, error)
	List(context.Context) ([]Task, error)
	Update(context.Context, Task) (Task, error)
}

// MemoryRepository 用 map 保存任务。
// map 不是并发安全的，所以 HTTP 服务里必须用 mutex 保护读写。
type MemoryRepository struct {
	mu    sync.Mutex
	next  int
	tasks map[string]Task
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{tasks: make(map[string]Task)}
}

func (r *MemoryRepository) Create(ctx context.Context, task Task) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.next++
	task.ID = fmt.Sprintf("task-%d", r.next)
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}

	r.tasks[task.ID] = task
	return task, nil
}

func (r *MemoryRepository) Find(ctx context.Context, id string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	return task, nil
}

func (r *MemoryRepository) List(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks, nil
}

func (r *MemoryRepository) Update(ctx context.Context, task Task) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.ID]; !ok {
		return Task{}, ErrNotFound
	}

	r.tasks[task.ID] = task
	return task, nil
}

var _ Repository = (*MemoryRepository)(nil)
