package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dkplus-pro/go-learning/examples/lesson11-concurrency-background-jobs-and-shutdown/internal/jobs"
)

type createJobRequest struct {
	Payload string `json:"payload"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(queue *jobs.Queue) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", func(w http.ResponseWriter, r *http.Request) {
		var req createJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
			return
		}

		job, err := queue.Enqueue(r.Context(), req.Payload)
		if err != nil {
			writeJobError(w, err)
			return
		}

		writeJSON(w, http.StatusAccepted, job)
	})

	mux.HandleFunc("GET /jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		job, err := queue.Find(r.Context(), r.PathValue("id"))
		if err != nil {
			writeJobError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, job)
	})

	return mux
}

func writeJobError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, jobs.ErrValidation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, jobs.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "job not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
