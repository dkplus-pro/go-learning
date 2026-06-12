package main

import (
	"log"
	"net/http"

	"github.com/dkplus-pro/go-learning/examples/lesson07-layering-and-in-memory-task-api/internal/task"
)

func main() {
	repo := task.NewMemoryRepository()
	service := task.NewService(repo)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, task.NewHandler(service)))
}
