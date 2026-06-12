package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, task Task) (Task, error) {
	now := time.Now().UTC()
	if task.ID == "" {
		task.ID = "task-" + strconv.FormatInt(now.UnixNano(), 36)
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, title, status, created_at)
		VALUES ($1, $2, $3, $4)
	`, task.ID, task.Title, task.Status, task.CreatedAt)
	if err != nil {
		return Task{}, fmt.Errorf("insert task: %w", err)
	}

	return task, nil
}

func (r *PostgresRepository) Find(ctx context.Context, id string) (Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, status, created_at
		FROM tasks
		WHERE id = $1
	`, id)

	task, err := scanTask(row)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, status, created_at
		FROM tasks
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Status, &task.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepository) Update(ctx context.Context, task Task) (Task, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = $2, status = $3
		WHERE id = $1
	`, task.ID, task.Title, task.Status)
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return Task{}, fmt.Errorf("read affected rows: %w", err)
	}
	if affected == 0 {
		return Task{}, ErrNotFound
	}

	return task, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (Task, error) {
	var task Task
	if err := row.Scan(&task.ID, &task.Title, &task.Status, &task.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("scan task: %w", err)
	}

	return task, nil
}

var _ Repository = (*PostgresRepository)(nil)
