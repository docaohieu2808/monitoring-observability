package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// AppCollector collects custom application metrics: request count, latency, error rate.
type AppCollector struct {
	mu sync.Mutex

	requestsTotal *prometheus.CounterVec
	requestErrors *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	activeRequests  *prometheus.GaugeVec
	buildInfo       *prometheus.GaugeVec
}

// NewAppCollector creates and registers a new AppCollector.
func NewAppCollector(namespace string) *AppCollector {
	c := &AppCollector{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed, partitioned by method and status code.",
			},
			[]string{"method", "path", "status_code"},
		),
		requestErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "request_errors_total",
				Help:      "Total number of HTTP request errors, partitioned by method and path.",
			},
			[]string{"method", "path", "error_type"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency distributions in seconds.",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"method", "path"},
		),
		activeRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_requests",
				Help:      "Number of HTTP requests currently being processed.",
			},
			[]string{"method", "path"},
		),
		buildInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "build_info",
				Help:      "Build information about the application.",
			},
			[]string{"version", "go_version", "git_commit"},
		),
	}

	// Set static build info
	c.buildInfo.WithLabelValues("1.0.0", "go1.22", "abc1234").Set(1)

	return c
}

// Describe implements prometheus.Collector.
func (c *AppCollector) Describe(ch chan<- *prometheus.Desc) {
	c.requestsTotal.Describe(ch)
	c.requestErrors.Describe(ch)
	c.requestDuration.Describe(ch)
	c.activeRequests.Describe(ch)
	c.buildInfo.Describe(ch)
}

// Collect implements prometheus.Collector.
func (c *AppCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestsTotal.Collect(ch)
	c.requestErrors.Collect(ch)
	c.requestDuration.Collect(ch)
	c.activeRequests.Collect(ch)
	c.buildInfo.Collect(ch)
}

// SimulateTraffic generates synthetic request metrics for demonstration.
func (c *AppCollector) SimulateTraffic() {
	endpoints := []struct{ method, path string }{
		{"GET", "/api/v1/users"},
		{"POST", "/api/v1/users"},
		{"GET", "/api/v1/products"},
		{"GET", "/health"},
	}
	statusCodes := []string{"200", "201", "400", "404", "500"}
	errorTypes := []string{"timeout", "connection_refused", "internal"}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		ep := endpoints[rand.Intn(len(endpoints))]
		status := statusCodes[rand.Intn(len(statusCodes))]
		latency := rand.Float64() * 2.0

		c.mu.Lock()
		c.requestsTotal.WithLabelValues(ep.method, ep.path, status).Inc()
		c.requestDuration.WithLabelValues(ep.method, ep.path).Observe(latency)

		if status == "500" || status == "400" {
			errType := errorTypes[rand.Intn(len(errorTypes))]
			c.requestErrors.WithLabelValues(ep.method, ep.path, errType).Inc()
		}

		active := float64(rand.Intn(20))
		c.activeRequests.WithLabelValues(ep.method, ep.path).Set(active)
		c.mu.Unlock()
	}
}
