#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"

echo "Starting smoke tests against $BASE_URL..."

# Check 1: health
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || true)
if [ "$code" != "200" ]; then
  echo "FAIL: health returned $code" >&2
  exit 1
fi
echo "health OK"

# Check 2: create an order
code=$(curl -s -o /tmp/order.json -w "%{http_code}" -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u1","amount":1999}' || true)
if [ "$code" != "201" ]; then
  echo "FAIL: create order returned $code" >&2
  exit 1
fi
order_id=$(jq -r .id /tmp/order.json)
echo "create order OK (id: $order_id)"

# Check 3: fetch the order and compare the id
code=$(curl -s -o /tmp/fetched.json -w "%{http_code}" "$BASE_URL/orders/$order_id" || true)
if [ "$code" != "200" ]; then
  echo "FAIL: get order returned $code" >&2
  exit 1
fi
fetched_id=$(jq -r .id /tmp/fetched.json)
if [ "$fetched_id" != "$order_id" ]; then
  echo "FAIL: expected id $order_id, got $fetched_id" >&2
  exit 1
fi
echo "get order OK"

echo "OK"
