package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestHandler() http.Handler {
	return NewHandler(NewService(NewTokenManager("test-secret", time.Minute)))
}

func TestProtectedTaskFlow(t *testing.T) {
	handler := newTestHandler()

	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"email":"ada@example.com","password":"super-secret"}`))
	registerRec := httptest.NewRecorder()
	handler.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", registerRec.Code, http.StatusCreated)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"ada@example.com","password":"super-secret"}`))
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginRec.Code, http.StatusOK)
	}

	var login loginResponse
	if err := json.NewDecoder(loginRec.Body).Decode(&login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	taskReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"Protect task API"}`))
	taskReq.Header.Set("Authorization", "Bearer "+login.Token)
	taskRec := httptest.NewRecorder()
	handler.ServeHTTP(taskRec, taskReq)

	if taskRec.Code != http.StatusCreated {
		t.Fatalf("task status = %d, want %d", taskRec.Code, http.StatusCreated)
	}

	var task Task
	if err := json.NewDecoder(taskRec.Body).Decode(&task); err != nil {
		t.Fatalf("decode task response: %v", err)
	}
	if task.OwnerID == "" {
		t.Fatal("task owner id is empty")
	}
}

func TestProtectedTaskRequiresToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"No token"}`))
	rec := httptest.NewRecorder()

	newTestHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
