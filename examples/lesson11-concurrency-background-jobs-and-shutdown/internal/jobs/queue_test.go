package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestQueueProcessesJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := NewQueue(1, time.Millisecond)
	queue.Start(ctx)

	job, err := queue.Enqueue(ctx, "send digest")
	if err != nil {
		t.Fatalf("Enqueue() unexpected error: %v", err)
	}

	waitFor(t, 100*time.Millisecond, func() bool {
		got, err := queue.Find(ctx, job.ID)
		return err == nil && got.Status == StatusDone && got.ProcessedAt != nil
	})
}

func TestQueueRejectsEmptyPayload(t *testing.T) {
	_, err := NewQueue(1, time.Millisecond).Enqueue(context.Background(), " ")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("Enqueue() error = %v, want ErrValidation", err)
	}
}

func TestQueueStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := NewQueue(1, time.Hour)
	queue.Start(ctx)

	job, err := queue.Enqueue(ctx, "long job")
	if err != nil {
		t.Fatalf("Enqueue() unexpected error: %v", err)
	}

	cancel()
	time.Sleep(5 * time.Millisecond)

	got, err := queue.Find(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("Find() unexpected error: %v", err)
	}
	if got.Status == StatusDone {
		t.Fatal("job was completed after context cancellation")
	}
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition was not met before timeout")
}
