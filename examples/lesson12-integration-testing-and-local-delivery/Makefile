DATABASE_URL ?= postgres://app:secret@localhost:5432/taskdb?sslmode=disable
JWT_SECRET ?= dev-secret-change-me

.PHONY: test integration-test run smoke compose-up compose-down

test:
	go test ./...

integration-test:
	DATABASE_URL='$(DATABASE_URL)' go test ./...

run:
	DATABASE_URL='$(DATABASE_URL)' JWT_SECRET='$(JWT_SECRET)' go run .

smoke:
	BASE_URL='http://localhost:8080' ./scripts/smoke.sh

compose-up:
	docker compose up -d

compose-down:
	docker compose down
