# Monitoring & Observability

Full-stack observability setup covering metrics (Prometheus), dashboards (Grafana), log aggregation (Loki), distributed tracing (Tempo), and telemetry collection (OpenTelemetry).

## Tools Covered

| Tool | Purpose |
|------|---------|
| **Prometheus** | Time-series metrics scraping & alerting |
| **Grafana** | Visualization dashboards, alerting UI |
| **Loki** | Log aggregation (label-based, no full-text index) |
| **Tempo** | Distributed tracing backend (Jaeger-compatible) |
| **OpenTelemetry** | Vendor-neutral telemetry collection pipeline |
| **Custom Exporter** | Pattern for writing your own Prometheus exporter |

## Skill Levels

| Level | Description |
|-------|-------------|
| L1 | Basic scrape config, Grafana provisioning |
| L2 | Alerting rules, dashboard templating |
| L3 | Loki config, promtail, log queries |
| L4 | Tempo tracing, OTel Collector pipeline |
| L5 | Alertmanager routing, recording rules |
| L6 | Federation, remote write |
| L-max | Custom Grafana plugin, custom exporter |

## Prerequisites

- Docker & Docker Compose >= v2
- Ports 3000 (Grafana), 9090 (Prometheus), 3100 (Loki), 3200 (Tempo) available
- `curl` for API testing

## Quick Start

```bash
# Spin up the full observability stack
cd /home/hieudc/monitoring-observability
docker compose -f docker-compose-full-stack.yml up -d

# Access UIs
open http://localhost:3000   # Grafana (admin/admin)
open http://localhost:9090   # Prometheus
```

## Directory Structure

```
monitoring-observability/
├── docker-compose-full-stack.yml   ← full stack in one command
├── prometheus/
│   ├── l1-basic-scrape-config/
│   ├── l2-alerting-rules/
│   ├── l5-alertmanager/
│   └── l6-recording-rules-federation/
├── grafana/
│   ├── dashboards/
│   ├── l2-provisioning/
│   └── l-max-custom-plugin/
├── loki/
│   ├── l3-loki-config/
│   └── promtail/
├── tempo/
│   └── l4-tempo-config/
├── opentelemetry/
│   └── l4-otel-collector/
└── custom-exporter/
```

## Usage

- **Prometheus**: Edit `prometheus/l1-basic-scrape-config/prometheus.yml`, add targets under `scrape_configs`.
- **Grafana**: Dashboards auto-provisioned from `grafana/dashboards/`. Login at `:3000`.
- **Loki**: Ship logs via promtail or OTel Collector. Query in Grafana with LogQL.
- **Tempo**: Send traces via OTel (gRPC `:4317`). Visualize in Grafana Tempo datasource.
- **OTel Collector**: Configure pipelines in `opentelemetry/l4-otel-collector/otel-config.yaml`.
