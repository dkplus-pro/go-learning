package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	addr := ":8080"
	service := ReportService{Delay: 120 * time.Millisecond}
	timeout := 300 * time.Millisecond

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, NewHandler(service, timeout)))
}
