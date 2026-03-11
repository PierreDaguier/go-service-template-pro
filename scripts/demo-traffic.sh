#!/usr/bin/env bash
set -euo pipefail

API_BASE=${API_BASE:-http://localhost:8080}
API_KEY=${API_KEY:-client-demo-key-2026}

for i in {1..25}; do
  curl -sS -H "X-API-Key: ${API_KEY}" "${API_BASE}/api/v1/overview" >/dev/null
  curl -sS -H "X-API-Key: ${API_KEY}" "${API_BASE}/api/v1/metrics/live?window=15m" >/dev/null
  if (( i % 5 == 0 )); then
    curl -sS -X POST "${API_BASE}/api/v1/errors" \
      -H "X-API-Key: ${API_KEY}" \
      -H "Content-Type: application/json" \
      -d '{"severity":"high","status":"open","message":"Synthetic timeout during checkout path"}' >/dev/null
  fi
  sleep 0.3
done

echo "Demo traffic generated."
