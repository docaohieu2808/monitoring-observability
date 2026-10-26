#!/usr/bin/env bash
# End-to-end smoke test for the compose stack. Run after:
#   docker compose up -d --build --wait
# Checks: Prometheus targets are up, rules are loaded, a log line reaches Loki (via Alloy and via
# the OTel Collector's OTLP path), a trace reaches Tempo through the collector, Grafana datasources work.
set -euo pipefail

PROM=${PROM_URL:-http://127.0.0.1:9090}
PROM_GLOBAL=${PROM_GLOBAL_URL:-http://127.0.0.1:9091}
LOKI=${LOKI_URL:-http://127.0.0.1:3100}
TEMPO=${TEMPO_URL:-http://127.0.0.1:3200}
OTLP=${OTLP_URL:-http://127.0.0.1:4318}
GRAFANA=${GRAFANA_URL:-http://127.0.0.1:3000}
GRAFANA_AUTH=${GRAFANA_AUTH:-admin:${GRAFANA_ADMIN_PASSWORD:-changeme}}
TIMEOUT=${SMOKE_TIMEOUT:-180}

pass() { echo "PASS  $*"; }
fail() { echo "FAIL  $*" >&2; exit 1; }

# retry <description> <command...>: run until it succeeds or TIMEOUT seconds pass.
retry() {
  local desc=$1; shift
  local deadline=$((SECONDS + TIMEOUT))
  until "$@" >/dev/null 2>&1; do
    if ((SECONDS >= deadline)); then fail "$desc (timed out after ${TIMEOUT}s)"; fi
    sleep 3
  done
  pass "$desc"
}

# prom_up <base-url> <job>: succeeds when up{job="<job>"} == 1 for at least one target.
prom_up() {
  curl -fsG "$1/api/v1/query" --data-urlencode "query=up{job=\"$2\"} == 1" |
    jq -e '.data.result | length > 0'
}

now_ns() { date +%s%N; }

# --- 0. Images without a shell (loki, tempo, collector, alloy) have no compose healthcheck ---
retry "loki ready" curl -fs "$LOKI/ready"
retry "tempo ready" curl -fs "$TEMPO/ready"
retry "otel collector healthy" curl -fs http://127.0.0.1:13133/
retry "alloy ready" curl -fs http://127.0.0.1:12345/-/ready

# --- 1. Prometheus targets ------------------------------------------------
for job in prometheus alertmanager loki tempo otel-collector alloy grafana custom-exporter node-exporter; do
  retry "prometheus target up: $job" prom_up "$PROM" "$job"
done
retry "federation target up on prometheus-global" prom_up "$PROM_GLOBAL" federate-primary

# --- 2. Rules loaded (alert rules from l2, recording rules from l6) -------
rules=$(curl -fs "$PROM/api/v1/rules")
for group in service.alerts node.alerts postgresql.alerts redis.alerts node.recording postgresql.recording redis.recording; do
  echo "$rules" | jq -e --arg g "$group" '[.data.groups[].name] | index($g) != null' >/dev/null ||
    fail "rule group not loaded: $group"
done
pass "alert and recording rule groups loaded"

# --- 3. Log line via Alloy (docker log tailing) ---------------------------
token="smoke-alloy-$(now_ns)"
# Write into the exporter's stdout (PID 1) so the Docker json-file driver captures it.
docker exec custom-exporter sh -c "echo '{\"level\":\"info\",\"msg\":\"$token\"}' > /proc/1/fd/1"
loki_has() { # loki_has <logql>
  curl -fsG "$LOKI/loki/api/v1/query_range" --data-urlencode "query=$1" \
    --data-urlencode "start=$(( $(date +%s) - 600 ))000000000" --data-urlencode "limit=5" |
    jq -e '.data.result | length > 0'
}
retry "log line reaches Loki via Alloy" loki_has "{container=\"custom-exporter\"} |= \"$token\""

# --- 4. Log line via OTel Collector -> otlphttp -> Loki /otlp -------------
otlp_token="smoke-otlp-$(now_ns)"
curl -fs -X POST "$OTLP/v1/logs" -H 'Content-Type: application/json' -d "{
  \"resourceLogs\": [{
    \"resource\": {\"attributes\": [{\"key\": \"service.name\", \"value\": {\"stringValue\": \"smoke-test\"}}]},
    \"scopeLogs\": [{\"logRecords\": [{\"timeUnixNano\": \"$(now_ns)\", \"severityText\": \"INFO\", \"body\": {\"stringValue\": \"$otlp_token\"}}]}]
  }]}" >/dev/null || fail "OTLP log POST to collector"
retry "log line reaches Loki via OTel Collector" loki_has "{service_name=\"smoke-test\"} |= \"$otlp_token\""

# --- 5. Trace via OTel Collector -> Tempo ---------------------------------
trace_id=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
span_id=$(od -An -N8 -tx1 /dev/urandom | tr -d ' \n')
start=$(now_ns)
curl -fs -X POST "$OTLP/v1/traces" -H 'Content-Type: application/json' -d "{
  \"resourceSpans\": [{
    \"resource\": {\"attributes\": [{\"key\": \"service.name\", \"value\": {\"stringValue\": \"smoke-test\"}}]},
    \"scopeSpans\": [{\"spans\": [{\"traceId\": \"$trace_id\", \"spanId\": \"$span_id\", \"name\": \"smoke-span\", \"kind\": 2,
      \"startTimeUnixNano\": \"$start\", \"endTimeUnixNano\": \"$((start + 1000000))\"}]}]
  }]}" >/dev/null || fail "OTLP trace POST to collector"
tempo_has() { curl -fs "$TEMPO/api/traces/$trace_id" | jq -e '.. | .name? // empty | select(. == "smoke-span")'; }
retry "trace $trace_id reaches Tempo via OTel Collector" tempo_has

# --- 6. Grafana datasources reach their backends --------------------------
for uid in prometheus loki tempo; do
  retry "grafana datasource healthy: $uid" \
    curl -fs -u "$GRAFANA_AUTH" "$GRAFANA/api/datasources/uid/$uid/health"
done

echo "ALL SMOKE CHECKS PASSED"
