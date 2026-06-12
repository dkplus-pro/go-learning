package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	service *Service
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type Task struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type userContextKey struct{}

func NewHandler(service *Service) http.Handler {
	h := &Handler{service: service}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.register)
		r.Post("/auth/login", h.login)

		r.Group(func(r chi.Router) {
			r.Use(h.authenticate)
			r.Get("/me", h.me)
			r.Post("/tasks", h.createTask)
		})
	})

	return r
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, ErrUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, ErrUnauthorized)
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, errors.Join(ErrValidation, errors.New("title is required")))
		return
	}

	writeJSON(w, http.StatusCreated, Task{
		ID:        "task-demo",
		OwnerID:   user.ID,
		Title:     strings.TrimSpace(req.Title),
		CreatedAt: time.Now().UTC(),
	})
}

func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimSpace(r.Header.Get("Authorization"))
		token, ok := strings.CutPrefix(raw, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			writeError(w, ErrUnauthorized)
			return
		}

		user, err := h.service.VerifyToken(strings.TrimSpace(token))
		if err != nil {
			writeError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey{}).(User)
	return user, ok
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrConflict):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
