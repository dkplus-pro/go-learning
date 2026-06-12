package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRegisterAndLogin(t *testing.T) {
	service := NewService(NewTokenManager("test-secret", time.Minute))

	user, err := service.Register(context.Background(), "ADA@EXAMPLE.COM", "super-secret")
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}
	if user.Email != "ada@example.com" {
		t.Fatalf("Email = %q, want ada@example.com", user.Email)
	}

	token, err := service.Login(context.Background(), "ada@example.com", "super-secret")
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("Login() returned empty token")
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	service := NewService(NewTokenManager("test-secret", time.Minute))
	_, err := service.Register(context.Background(), "ada@example.com", "super-secret")
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, err = service.Register(context.Background(), "ada@example.com", "super-secret")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("Register() error = %v, want ErrConflict", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	service := NewService(NewTokenManager("test-secret", time.Minute))
	_, err := service.Register(context.Background(), "ada@example.com", "super-secret")
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, err = service.Login(context.Background(), "ada@example.com", "wrong-password")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Login() error = %v, want ErrUnauthorized", err)
	}
}
