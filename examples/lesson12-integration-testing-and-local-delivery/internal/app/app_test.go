package app

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAppIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run integration test")
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

	application := New(db, NewTokenManager("test-secret", time.Hour))
	email := "integration-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	user, err := application.Register(ctx, email, "super-secret")
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	token, err := application.Login(ctx, email, "super-secret")
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("Login() returned empty token")
	}

	task, err := application.CreateTask(ctx, user, "Integration task")
	if err != nil {
		t.Fatalf("CreateTask() unexpected error: %v", err)
	}

	tasks, err := application.ListTasks(ctx, user)
	if err != nil {
		t.Fatalf("ListTasks() unexpected error: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("ListTasks() returned no tasks")
	}

	updated, err := application.UpdateTaskStatus(ctx, user, task.ID, "done")
	if err != nil {
		t.Fatalf("UpdateTaskStatus() unexpected error: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("Status = %q, want done", updated.Status)
	}
}
