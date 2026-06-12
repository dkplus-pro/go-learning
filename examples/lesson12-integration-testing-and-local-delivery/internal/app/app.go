package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrValidation   = errors.New("validation error")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

type App struct {
	db     *sql.DB
	tokens *TokenManager
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Task struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func New(db *sql.DB, tokens *TokenManager) *App {
	return &App{db: db, tokens: tokens}
}

func (a *App) Register(ctx context.Context, email string, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") {
		return User{}, fmt.Errorf("%w: email must be valid", ErrValidation)
	}
	if len(password) < 8 {
		return User{}, fmt.Errorf("%w: password must be at least 8 characters", ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{ID: newID("user"), Email: email}
	result, err := a.db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`, user.ID, user.Email, hash)
	if err != nil {
		return User{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return User{}, err
	}
	if affected == 0 {
		return User{}, fmt.Errorf("%w: email already registered", ErrConflict)
	}

	return user, nil
}

func (a *App) Login(ctx context.Context, email string, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var user User
	var hash []byte
	err := a.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUnauthorized
	}
	if err != nil {
		return "", err
	}

	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return "", ErrUnauthorized
	}

	return a.tokens.Generate(user)
}

func (a *App) CreateTask(ctx context.Context, user User, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, fmt.Errorf("%w: title is required", ErrValidation)
	}

	task := Task{
		ID:        newID("task"),
		OwnerID:   user.ID,
		Title:     title,
		Status:    "todo",
		CreatedAt: time.Now().UTC(),
	}

	_, err := a.db.ExecContext(ctx, `
		INSERT INTO tasks (id, owner_id, title, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, task.ID, task.OwnerID, task.Title, task.Status, task.CreatedAt)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (a *App) ListTasks(ctx context.Context, user User) ([]Task, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, owner_id, title, status, created_at
		FROM tasks
		WHERE owner_id = $1
		ORDER BY created_at ASC
	`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.OwnerID, &task.Title, &task.Status, &task.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (a *App) UpdateTaskStatus(ctx context.Context, user User, taskID string, status string) (Task, error) {
	if !validStatus(status) {
		return Task{}, fmt.Errorf("%w: invalid status", ErrValidation)
	}

	var task Task
	err := a.db.QueryRowContext(ctx, `
		SELECT id, owner_id, title, status, created_at
		FROM tasks
		WHERE id = $1 AND owner_id = $2
	`, taskID, user.ID).Scan(&task.ID, &task.OwnerID, &task.Title, &task.Status, &task.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}

	if !CanTransition(task.Status, status) {
		return Task{}, fmt.Errorf("%w: cannot move %s to %s", ErrValidation, task.Status, status)
	}

	task.Status = status
	_, err = a.db.ExecContext(ctx, `
		UPDATE tasks
		SET status = $3
		WHERE id = $1 AND owner_id = $2
	`, task.ID, user.ID, task.Status)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func validStatus(status string) bool {
	return status == "todo" || status == "doing" || status == "done"
}

func CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "todo":
		return to == "doing" || to == "done"
	case "doing":
		return to == "done"
	default:
		return false
	}
}

func newID(prefix string) string {
	return prefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
