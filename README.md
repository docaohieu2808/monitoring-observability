# monitoring-observability

Full observability stack: metrics (Prometheus + Alertmanager), logs (Loki + Promtail), traces (Tempo), unified collection (OTel Collector), dashboards (Grafana), and a custom Go exporter.

Mirrors the actual stack running on server3 (prometheus/grafana/loki/tempo in production use).

## Tier map

| Tier | Path | What it covers |
|------|------|----------------|
| l1 | `prometheus/l1-basic-scrape-config/` | Prometheus scrape config — node, postgres, redis exporters |
| l2 | `prometheus/l2-alerting-rules/` | Alert rules for node, database, and service SLOs |
| l3 | `loki/l3-loki-config/` | Loki 3.x config — TSDB schema v13, compactor, retention |
| l4 | `tempo/l4-tempo-config/` + `opentelemetry/l4-otel-collector/` | Tempo 2.4 distributed tracing + OTel Collector pipeline |
| l5 | `prometheus/l5-alertmanager/` | Alertmanager routing — Slack / email / PagerDuty receivers |
| l6 | `prometheus/l6-recording-rules-federation/` | Recording rules + Prometheus federation config |
| l-max | `grafana/l-max-custom-plugin/` + `custom-exporter/l-max-go-exporter/` | Grafana plugin skeleton + custom Go exporter with Dockerfile |

Grafana provisioning (datasources + 4 dashboards) lives in `grafana/l2-provisioning/` and `grafana/dashboards/`.

## Run full stack

```bash
# Copy and fill in your target host
cp .env.example .env          # set SERVER2_IP
docker compose -f docker-compose-full-stack.yml up -d
```

Grafana: http://localhost:3000 (admin / changeme)
Prometheus: http://localhost:9090
Alertmanager: http://localhost:9093

To run only Prometheus + alerting:
```bash
docker run -v ./prometheus/l1-basic-scrape-config/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus:v3.4.1
```

## Interview talking points

- **Trace correlation**: Grafana derived fields link Loki log lines to Tempo traces via `traceId`; exemplars bridge Prometheus metrics to traces.
- **Loki 3.x schema v13 + TSDB**: faster label queries vs BoltDB-shipper; compactor handles retention and deletion.
- **OTel Collector fan-out**: single receiver pipeline exports to Prometheus (metrics), Loki (logs), and Tempo (traces) — decouples instrumentation from backends.
- **Federation**: `l6` shows how a global Prometheus scrapes recording rules from a remote instance to reduce cardinality at the top level.
- **Custom exporter**: `l-max` demonstrates building a `/metrics` endpoint in Go with `prometheus/client_golang`, multi-stage Docker build.

## Notes

- `SERVER2_IP` env var replaces the real server IP in scrape targets and Grafana domain; set before running.
- Prometheus bumped to v3.4.1 (3.x LTS); `--web.console.*` flags removed (dropped in 3.0).
- Alertmanager uses `matchers` syntax (deprecated `match`/`match_re` removed in 0.27+).
