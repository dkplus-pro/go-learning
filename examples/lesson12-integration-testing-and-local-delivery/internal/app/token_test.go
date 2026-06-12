package app

import (
	"testing"
	"time"
)

func TestTokenManagerRoundTrip(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Minute)
	user := User{ID: "user-1", Email: "ada@example.com"}

	token, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}

	got, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if got != user {
		t.Fatalf("Verify() = %+v, want %+v", got, user)
	}
}

func TestCanTransition(t *testing.T) {
	if !CanTransition("todo", "done") {
		t.Fatal("todo should move to done")
	}
	if CanTransition("done", "todo") {
		t.Fatal("done should not move back to todo")
	}
}
