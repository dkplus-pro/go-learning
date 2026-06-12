package task

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func NewHandler(service *Service, logger *slog.Logger, timeout time.Duration) http.Handler {
	h := &Handler{service: service, logger: logger}
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(timeoutMiddleware(timeout))
	r.Use(loggingMiddleware(logger))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/tasks", h.createTask)
		r.Get("/tasks/{taskID}", h.getTask)
	})

	return r
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	task, err := h.service.CreateTask(r.Context(), req.Title)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.service.GetTask(r.Context(), chi.URLParam(r, "taskID"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "validation_error", publicMessage(err))
	case errors.Is(err, ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", publicMessage(err))
	case errors.Is(err, context.DeadlineExceeded):
		writeAPIError(w, http.StatusGatewayTimeout, "request_timeout", "request timed out")
	default:
		h.logger.Error("unhandled application error", "error", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			logger.Info(
				"request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(startedAt).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	r.Header().Set("Content-Length", strconv.Itoa(len(body)))
	return r.ResponseWriter.Write(body)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
