package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const userKey contextKey = "user"

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type taskRequest struct {
	Title string `json:"title"`
}

type statusRequest struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (a *App) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", a.handleRegister)
		r.Post("/auth/login", a.handleLogin)

		r.Group(func(r chi.Router) {
			r.Use(a.authenticate)
			r.Get("/tasks", a.handleListTasks)
			r.Post("/tasks", a.handleCreateTask)
			r.Patch("/tasks/{taskID}/status", a.handleUpdateTaskStatus)
		})
	})

	return r
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	user, err := a.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	token, err := a.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (a *App) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())
	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	task, err := a.CreateTask(r.Context(), user, req.Title)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (a *App) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := a.ListTasks(r.Context(), currentUser(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (a *App) handleUpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	task, err := a.UpdateTaskStatus(r.Context(), currentUser(r.Context()), chi.URLParam(r, "taskID"), req.Status)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (a *App) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		user, err := a.tokens.Verify(strings.TrimSpace(token))
		if err != nil {
			writeError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func currentUser(ctx context.Context) User {
	user, _ := ctx.Value(userKey).(User)
	return user
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrConflict):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	case errors.Is(err, ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
