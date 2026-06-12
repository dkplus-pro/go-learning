package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("job not found")
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
)

type Job struct {
	ID          string     `json:"id"`
	Payload     string     `json:"payload"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

type Queue struct {
	mu           sync.Mutex
	next         int
	processDelay time.Duration
	jobs         map[string]Job
	pending      chan string
}

func NewQueue(buffer int, processDelay time.Duration) *Queue {
	return &Queue{
		processDelay: processDelay,
		jobs:         make(map[string]Job),
		pending:      make(chan string, buffer),
	}
}

func (q *Queue) Start(ctx context.Context) {
	go q.worker(ctx)
}

func (q *Queue) Enqueue(ctx context.Context, payload string) (Job, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return Job{}, fmt.Errorf("%w: payload is required", ErrValidation)
	}

	q.mu.Lock()
	q.next++
	job := Job{
		ID:        fmt.Sprintf("job-%d", q.next),
		Payload:   payload,
		Status:    StatusQueued,
		CreatedAt: time.Now().UTC(),
	}
	q.jobs[job.ID] = job
	q.mu.Unlock()

	select {
	case q.pending <- job.ID:
		return job, nil
	case <-ctx.Done():
		return Job{}, ctx.Err()
	}
}

func (q *Queue) Find(ctx context.Context, id string) (Job, error) {
	if err := ctx.Err(); err != nil {
		return Job{}, err
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[id]
	if !ok {
		return Job{}, ErrNotFound
	}

	return job, nil
}

func (q *Queue) worker(ctx context.Context) {
	for {
		select {
		case id := <-q.pending:
			q.process(ctx, id)
		case <-ctx.Done():
			return
		}
	}
}

func (q *Queue) process(ctx context.Context, id string) {
	q.updateStatus(id, StatusProcessing, nil)

	timer := time.NewTimer(q.processDelay)
	defer timer.Stop()

	select {
	case <-timer.C:
		now := time.Now().UTC()
		q.updateStatus(id, StatusDone, &now)
	case <-ctx.Done():
		return
	}
}

func (q *Queue) updateStatus(id string, status Status, processedAt *time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[id]
	if !ok {
		return
	}

	job.Status = status
	job.ProcessedAt = processedAt
	q.jobs[id] = job
}
