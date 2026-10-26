package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// newHandler wires a private registry (Go runtime, process and application collectors) behind
// /metrics, /health and / . It returns the app collector so callers can drive it.
func newHandler(namespace string) (http.Handler, *AppCollector) {
	// Custom registry to avoid default Go metrics collision
	reg := prometheus.NewRegistry()

	// Standard Go runtime and process metrics
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		Namespace: namespace,
	}))

	// Application-specific collector
	appCollector := NewAppCollector(namespace)
	reg.MustRegister(appCollector)

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body><h1>Custom Exporter</h1><a href="/metrics">Metrics</a></body></html>`))
	})

	return mux, appCollector
}

func main() {
	listenAddr := envOrDefault("LISTEN_ADDR", ":9200")
	namespace := envOrDefault("METRICS_NAMESPACE", "app")

	handler, appCollector := newHandler(namespace)

	// Simulate traffic in background goroutine
	go appCollector.SimulateTraffic()

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Custom exporter listening on %s (namespace: %s)", listenAddr, namespace)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
