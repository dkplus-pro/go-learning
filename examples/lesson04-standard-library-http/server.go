package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type greetingRequest struct {
	Name string `json:"name"`
}

type greetingResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewHandler 创建应用的 HTTP 入口。
// 让 main.go 只负责启动，可以让测试直接复用同一个 handler。
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("POST /greetings", greetingsHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func greetingsHandler(w http.ResponseWriter, r *http.Request) {
	var req greetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}

	writeJSON(w, http.StatusCreated, greetingResponse{Message: "Hello, " + name + "!"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Encode 失败通常意味着响应对象本身不可编码。
	// 本 demo 的响应结构都可编码，所以这里不把错误暴露给客户端。
	_ = json.NewEncoder(w).Encode(value)
}
