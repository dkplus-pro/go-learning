package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dkplus-pro/go-learning/examples/lesson10-users-passwords-and-jwt/internal/auth"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	manager := auth.NewTokenManager(secret, 2*time.Hour)
	service := auth.NewService(manager)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, auth.NewHandler(service)))
}
