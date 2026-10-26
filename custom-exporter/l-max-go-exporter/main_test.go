package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, h http.Handler, path string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	return rec.Code, string(body)
}

func TestHealth(t *testing.T) {
	h, _ := newHandler("app")
	code, body := get(t, h, "/health")
	if code != http.StatusOK || body != "ok" {
		t.Fatalf("got %d %q, want 200 \"ok\"", code, body)
	}
}

func TestMetricsExposeApplicationSeries(t *testing.T) {
	h, c := newHandler("testns")
	// Record one request directly instead of waiting for the background simulator.
	c.requestsTotal.WithLabelValues("GET", "/api/v1/users", "200").Inc()
	c.requestDuration.WithLabelValues("GET", "/api/v1/users").Observe(0.2)

	code, body := get(t, h, "/metrics")
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	for _, want := range []string{
		`testns_requests_total{method="GET",path="/api/v1/users",status_code="200"} 1`,
		`testns_request_duration_seconds_count{method="GET",path="/api/v1/users"} 1`,
		`testns_build_info{`,
		"go_goroutines",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q", want)
		}
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("EXPORTER_TEST_KEY", "")
	if got := envOrDefault("EXPORTER_TEST_KEY", "fallback"); got != "fallback" {
		t.Fatalf("got %q", got)
	}
	t.Setenv("EXPORTER_TEST_KEY", "set")
	if got := envOrDefault("EXPORTER_TEST_KEY", "fallback"); got != "set" {
		t.Fatalf("got %q", got)
	}
}
