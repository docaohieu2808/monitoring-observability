# monitoring-observability

[![ci](https://github.com/docaohieu2808/monitoring-observability/actions/workflows/ci.yml/badge.svg)](https://github.com/docaohieu2808/monitoring-observability/actions/workflows/ci.yml)

A learning ladder for a full observability stack: metrics (Prometheus + Alertmanager), logs (Loki + Grafana Alloy),
traces (Tempo), unified collection (OpenTelemetry Collector), dashboards (Grafana) and a custom Go exporter.
Everything runs with one `docker compose up --wait`, and CI proves it weekly.

## Versions (verified 2026-10-26)

| Component | Version |
|-----------|---------|
| Prometheus | v3.15.1 |
| Alertmanager | v0.34.1 |
| Grafana | 13.2.4 |
| Loki | 3.7.9 |
| Grafana Alloy | v1.21.2 |
| Tempo | 3.1.1 (single binary, `-target=all`) |
| OpenTelemetry Collector contrib | 0.163.2 |
| node_exporter | v1.12.1 |
| Go / client_golang | 1.26 / v1.25.0 |

Renamed or replaced upstream, and what changed here:

- **Promtail is end of life (2026-03-02)** -> replaced by **Grafana Alloy** (`loki/alloy/config.alloy`).
- The collector's **`loki` exporter was removed** -> logs go through `otlphttp/loki` to Loki's native `/otlp` endpoint.
- **Tempo 3** dropped the ingester/compactor modules and the OpenCensus receiver; retention moved to
  `backend_scheduler.provider.compaction.compaction.block_retention`.
- Collector self-telemetry uses `service.telemetry.metrics.readers` (the old `address:` key is gone).
- **Prometheus never expands `${VAR}` in its config.** Per-environment targets live in `file_sd` files
  (`prometheus/l1-basic-scrape-config/targets/*.yml`), which Prometheus re-reads on change.
- Grafana datasources carry explicit `uid`s and use `tracesToLogsV2` (`lokiSearch` is deprecated).

## Tier map

| Tier | Path | What it teaches |
|------|------|-----------------|
| l1 | `prometheus/l1-basic-scrape-config/` | Scrape config, `file_sd` target files, relabeling |
| l2 | `prometheus/l2-alerting-rules/` | Alert rules for node, database and service health, with `promtool test rules` unit tests |
| l3 | `loki/l3-loki-config/` + `loki/alloy/` | Loki 3.x (TSDB schema v13, compactor retention) and log shipping with Alloy |
| l4 | `tempo/l4-tempo-config/` + `opentelemetry/l4-otel-collector/` | Tempo 3 tracing, OTel Collector fan-out to Prometheus, Loki and Tempo |
| l5 | `prometheus/l5-alertmanager/` | Alertmanager routing and inhibition (Slack / email / PagerDuty receivers with placeholder secrets) |
| l6 | `prometheus/l6-recording-rules-federation/` | Recording rules (loaded by the primary) + a second Prometheus federating them (`prometheus-global`) |
| l-max | `custom-exporter/l-max-go-exporter/` | Go `/metrics` exporter on `client_golang`, multi-stage non-root Dockerfile, unit tests |

Also: `grafana/l2-provisioning/` (datasources, dashboard provider) and `grafana/dashboards/` (4 JSON dashboards:
node, Docker, PostgreSQL, Redis; they need the matching exporters to show data).
`grafana/l-max-custom-plugin/plugin-skeleton.json` is only a plugin manifest, not a working plugin (no source).

## Run it

```bash
cp .env.example .env            # optional; every value has a default
docker compose up -d --build --wait
./scripts/smoke-test.sh         # same checks CI runs
docker compose down -v
```

| Service | URL (bound to 127.0.0.1) |
|---------|--------------------------|
| Grafana | http://localhost:3000 (admin / `changeme`, override with `GRAFANA_ADMIN_PASSWORD`) |
| Prometheus | http://localhost:9090 |
| Prometheus (federated) | http://localhost:9091 |
| Alertmanager | http://localhost:9093 |
| Loki | http://localhost:3100 |
| Tempo | http://localhost:3200 |
| OTel Collector OTLP | `localhost:4317` (gRPC), `localhost:4318` (HTTP) |
| Custom exporter | http://localhost:9200/metrics |

Monitor your own hosts: add `host:9100` to `prometheus/l1-basic-scrape-config/targets/node-exporter.yml`
(Postgres `:9187`, Redis `:9121` in the other two files). No restart needed.

Send a test trace/log through the collector: see `scripts/smoke-test.sh` for working OTLP/HTTP JSON payloads.

## What CI checks (`.github/workflows/ci.yml`)

Runs on push, pull request, weekly (Mon 02:17 UTC) and manually.

| Job | Checks |
|-----|--------|
| lint | actionlint, yamllint, shellcheck, dashboard JSON, hadolint |
| prometheus | `promtool check config` (l1, l6), `promtool check rules`, `promtool test rules` (alert + recording rule unit tests), `amtool check-config` |
| backends | `loki -verify-config`, `tempo -config.verify`, `otelcol validate`, `alloy validate` |
| exporter | `go mod tidy` clean, gofmt, vet, build, `go test -race` |
| stack-smoke | `docker compose up --wait`, then `scripts/smoke-test.sh`: every Prometheus target up, rule groups loaded, a log line reaches Loki via Alloy and via the collector, a trace reaches Tempo through the collector, Grafana datasources healthy |

## Honest limits

- The stack is a single-node lab: filesystem storage, no auth between components, Grafana default password.
  Do not expose it to the internet.
- Postgres and Redis alert rules, recording rules and dashboards are unit-tested with synthetic series only; no
  Postgres/Redis exporter runs in the compose stack.
- Alertmanager receivers use placeholder URLs and keys (`changeme`); notifications are not delivered anywhere.
- Alloy tails Docker containers through `/var/run/docker.sock` (read-only mount, but still privileged access).
- The Grafana plugin in `l-max` is a manifest skeleton only.

## Interview talking points

- **Trace correlation**: Tempo `tracesToLogsV2` links spans to Loki logs, Loki derived fields link `trace_id` back to Tempo, exemplars link Prometheus metrics to traces.
- **Loki 3 schema v13 + TSDB**: faster label queries than BoltDB-shipper; the compactor handles retention. Native OTLP ingestion removes the need for a Loki exporter.
- **OTel Collector fan-out**: one receiver pipeline exports metrics to Prometheus (remote write), logs to Loki (OTLP) and traces to Tempo, decoupling instrumentation from backends.
- **Federation + recording rules**: a global Prometheus pulls only pre-aggregated `instance:*` series, which keeps top-level cardinality low.
- **Alert hygiene**: one generic `TargetDown` instead of one `up == 0` rule per exporter; division guarded against `redis_memory_max_bytes == 0`; rules covered by `promtool test rules`.
- **Why Alloy**: Promtail reached end of life; Alloy is Grafana's OTel-compatible collector with the same Loki pipeline stages.
