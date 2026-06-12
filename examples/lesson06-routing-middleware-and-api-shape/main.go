package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := ":8080"
	logger := log.New(os.Stdout, "api ", log.LstdFlags)

	logger.Printf("listening on %s", addr)
	logger.Fatal(http.ListenAndServe(addr, NewHandler(logger)))
}
