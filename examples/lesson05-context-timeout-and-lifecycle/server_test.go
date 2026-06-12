package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReportHandlerReturnsReport(t *testing.T) {
	handler := NewHandler(ReportService{Delay: time.Millisecond}, 50*time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/reports/frontend-architecture", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var report Report
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if report.ID != "frontend-architecture" || report.Status != "ready" {
		t.Fatalf("report = %+v, want ready frontend-architecture", report)
	}
}

func TestReportHandlerTimesOut(t *testing.T) {
	handler := NewHandler(ReportService{Delay: 50 * time.Millisecond}, time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/reports/slow", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}
}

func TestReportServiceRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ReportService{Delay: time.Hour}.Build(ctx, "cancelled")

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Build() error = %v, want context.Canceled", err)
	}
}
