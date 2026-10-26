#!/usr/bin/env python3
"""Fail unless every alert rule has a promtool test that fires and one that must stay quiet.

Reads prometheus/l2-alerting-rules/alert-rules-*.yml and tests/*.test.yml. For each `alert:` it needs
at least one alert_rule_test with non-empty exp_alerts and at least one with empty exp_alerts.
`promtool test rules` proves the assertions hold; this script proves the assertions exist.
"""
import glob
import sys

import yaml

BASE = "prometheus/l2-alerting-rules"


def load(path):
    with open(path, encoding="utf-8") as handle:
        return yaml.safe_load(handle)


alerts = sorted(
    rule["alert"]
    for path in glob.glob(f"{BASE}/alert-rules-*.yml")
    for group in load(path)["groups"]
    for rule in group["rules"]
    if "alert" in rule
)

fires, quiet = set(), set()
for path in glob.glob(f"{BASE}/tests/*.test.yml"):
    for test in load(path).get("tests", []):
        for case in test.get("alert_rule_test", []):
            (fires if case.get("exp_alerts") else quiet).add(case["alertname"])

missing = False
for name in alerts:
    gaps = [label for label, seen in (("firing case", fires), ("not-firing case", quiet)) if name not in seen]
    if gaps:
        missing = True
        print(f"FAIL  {name}: missing {' and '.join(gaps)}", file=sys.stderr)
    else:
        print(f"ok    {name}")

unknown = (fires | quiet) - set(alerts)
if unknown:
    missing = True
    print(f"FAIL  tests reference unknown alerts: {sorted(unknown)}", file=sys.stderr)

sys.exit(1 if missing else 0)
