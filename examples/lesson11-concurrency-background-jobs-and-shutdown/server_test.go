package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dkplus-pro/go-learning/examples/lesson11-concurrency-background-jobs-and-shutdown/internal/jobs"
)

func TestCreateJob(t *testing.T) {
	queue := jobs.NewQueue(1, time.Millisecond)
	handler := NewHandler(queue)

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"payload":"send digest"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var job jobs.Job
	if err := json.NewDecoder(rec.Body).Decode(&job); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if job.ID == "" || job.Status != jobs.StatusQueued {
		t.Fatalf("job = %+v, want queued job with id", job)
	}
}

func TestGetJobNotFound(t *testing.T) {
	queue := jobs.NewQueue(1, time.Millisecond)
	handler := NewHandler(queue)

	req := httptest.NewRequest(http.MethodGet, "/jobs/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
