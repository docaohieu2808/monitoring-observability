package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddr := envOrDefault("LISTEN_ADDR", ":9200")
	namespace := envOrDefault("METRICS_NAMESPACE", "app")

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

	// Simulate traffic in background goroutine
	go appCollector.SimulateTraffic()

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Custom Exporter</h1><a href="/metrics">Metrics</a></body></html>`))
	})

	log.Printf("Custom exporter listening on %s (namespace: %s)", listenAddr, namespace)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
