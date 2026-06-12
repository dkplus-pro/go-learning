package task

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresRepositoryLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run PostgreSQL integration test")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewPostgresRepository(db)
	created, err := repo.Create(ctx, Task{Title: "Persist task", Status: StatusTodo})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	found, err := repo.Find(ctx, created.ID)
	if err != nil {
		t.Fatalf("Find() unexpected error: %v", err)
	}
	if found.Title != created.Title {
		t.Fatalf("Find() title = %q, want %q", found.Title, created.Title)
	}

	found.Status = StatusDone
	updated, err := repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if updated.Status != StatusDone {
		t.Fatalf("Update() status = %q, want %q", updated.Status, StatusDone)
	}

	tasks, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("List() returned no tasks")
	}
}

func TestPostgresRepositoryFindNotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run PostgreSQL integration test")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	_, err = NewPostgresRepository(db).Find(context.Background(), "missing-task-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find() error = %v, want ErrNotFound", err)
	}
}
