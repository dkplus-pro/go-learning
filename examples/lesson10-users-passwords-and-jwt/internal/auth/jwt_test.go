package auth

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

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	_, err := NewTokenManager("test-secret", time.Minute).Verify("not-a-token")
	if err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}
