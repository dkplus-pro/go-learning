package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type storedUser struct {
	User
	PasswordHash []byte
}

type Service struct {
	mu      sync.Mutex
	next    int
	users   map[string]storedUser
	manager *TokenManager
}

func NewService(manager *TokenManager) *Service {
	return &Service{
		users:   make(map[string]storedUser),
		manager: manager,
	}
}

func (s *Service) Register(ctx context.Context, email string, password string) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

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

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return User{}, fmt.Errorf("%w: email already registered", ErrConflict)
	}

	s.next++
	user := User{ID: fmt.Sprintf("user-%d", s.next), Email: email}
	s.users[email] = storedUser{User: user, PasswordHash: hash}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	email = strings.ToLower(strings.TrimSpace(email))

	s.mu.Lock()
	user, exists := s.users[email]
	s.mu.Unlock()

	if !exists {
		return "", ErrUnauthorized
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) != nil {
		return "", ErrUnauthorized
	}

	return s.manager.Generate(user.User)
}

func (s *Service) VerifyToken(raw string) (User, error) {
	return s.manager.Verify(raw)
}
