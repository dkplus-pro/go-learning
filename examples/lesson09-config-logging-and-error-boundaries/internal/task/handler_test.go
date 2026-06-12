package task

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(NewService(), logger, time.Second)
}

func TestCreateTask(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"Ship logs"}`))
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var task Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if task.Title != "Ship logs" {
		t.Fatalf("title = %q, want Ship logs", task.Title)
	}
}

func TestCreateTaskValidationError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":" "}`))
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "validation_error" {
		t.Fatalf("error code = %q, want validation_error", body.Error.Code)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/missing", nil)
	rec := httptest.NewRecorder()

	testHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
