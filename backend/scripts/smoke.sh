#!/usr/bin/env bash
# Smoke test — verifies the API is up and responding correctly.
# Run from the repo root or backend/ directory.
# Usage: bash scripts/smoke.sh [BASE_URL]

set -euo pipefail

BASE_URL="${1:-${BASE_URL:-http://localhost:8080}}"
PASS=0
FAIL=0

check() {
  local label="$1"
  local expected="$2"
  local actual="$3"

  if [ "$actual" = "$expected" ]; then
    echo "  ✓  $label"
    PASS=$((PASS + 1))
  else
    echo "  ✗  $label  (expected $expected, got $actual)"
    FAIL=$((FAIL + 1))
  fi
}

echo ""
echo "Smoke test — $BASE_URL"
echo "─────────────────────────────────────"

# ── Health ────────────────────────────────────────────────────────────────────
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
check "GET /health → 200" "200" "$status"

body=$(curl -s "$BASE_URL/health")
echo "$body" | grep -q '"status":"ok"' && \
  check 'GET /health body has "status":"ok"' "ok" "ok" || \
  check 'GET /health body has "status":"ok"' "ok" "missing"

# ── 404 on unknown route ──────────────────────────────────────────────────────
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/no-such-route")
check "GET /no-such-route → 404" "404" "$status"

# ─────────────────────────────────────────────────────────────────────────────
echo "─────────────────────────────────────"
echo "  Passed: $PASS  Failed: $FAIL"
echo ""

[ "$FAIL" -eq 0 ] || exit 1
