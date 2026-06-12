#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="smoke-$(date +%s)@example.com"
PASSWORD="super-secret"

curl -fsS "$BASE_URL/healthz" >/dev/null

curl -fsS -X POST "$BASE_URL/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" >/dev/null

LOGIN_RESPONSE="$(curl -fsS -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")"

TOKEN="$(printf '%s' "$LOGIN_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"
if [ -z "$TOKEN" ]; then
  echo "login response did not contain token" >&2
  exit 1
fi

TASK_RESPONSE="$(curl -fsS -X POST "$BASE_URL/api/v1/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Smoke test task"}')"

TASK_ID="$(printf '%s' "$TASK_RESPONSE" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
if [ -z "$TASK_ID" ]; then
  echo "create task response did not contain id" >&2
  exit 1
fi

curl -fsS "$BASE_URL/api/v1/tasks" -H "Authorization: Bearer $TOKEN" >/dev/null

curl -fsS -X PATCH "$BASE_URL/api/v1/tasks/$TASK_ID/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"done"}' >/dev/null

echo "smoke test passed"
