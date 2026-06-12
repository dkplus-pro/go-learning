package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Report struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ReportService struct {
	Delay time.Duration
}

type errorResponse struct {
	Error string `json:"error"`
}

// Build 模拟一个耗时操作。
// select 同时等待业务结果和取消信号，避免请求超时后继续浪费资源。
func (s ReportService) Build(ctx context.Context, id string) (Report, error) {
	select {
	case <-time.After(s.Delay):
		return Report{ID: id, Status: "ready"}, nil
	case <-ctx.Done():
		return Report{}, ctx.Err()
	}
}

func NewHandler(service ReportService, timeout time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /reports/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleReport(w, r, service, timeout)
	})
	return mux
}

func handleReport(w http.ResponseWriter, r *http.Request, service ReportService, timeout time.Duration) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "report id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	report, err := service.Build(ctx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			writeJSON(w, http.StatusGatewayTimeout, errorResponse{Error: "report generation timed out"})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
