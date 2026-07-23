#!/bin/bash
# E2E test script for Drone Platform API
# Usage: bash Test/e2e.sh [base_url]
BASE="${1:-http://localhost:8080}"
PASS=0; FAIL=0

code() { curl -s -o /dev/null -w '%{http_code}' "$@"; }

check() {
  local desc="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "  ✅ $desc ($actual)"
    PASS=$((PASS+1))
  else
    echo "  ❌ $desc (expected $expected, got $actual)"
    FAIL=$((FAIL+1))
  fi
}

echo "=== Drone Platform E2E Tests ==="
echo "Base: $BASE"
echo ""

# 1. Health check
echo "[1] Health Check"
CODE=$(code "$BASE/healthz")
check "GET /healthz → 200" 200 "$CODE"

# 2. Home page
echo "[2] Home"
CODE=$(code "$BASE/api/v1/home")
check "GET /api/v1/home → 200" 200 "$CODE"

# 3. Demand flow
echo "[3] Demand Flow"
TOKEN=$(curl -s -X POST "$BASE/api/v1/admin/token" -H 'Content-Type: application/json' -d '{"role":"enterprise"}' | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
AUTH="Authorization: Bearer $TOKEN"

check "POST /demands → 201" 201 "$(code -X POST "$BASE/api/v1/demands" -H 'Content-Type: application/json' -H "$AUTH" -d '{"publisher_name":"E2E","contact":"13800000000","title":"E2E2","biz_type":"other"}')"

# 4. List demands with pagination
echo "[4] Pagination"
CODE=$(code "$BASE/api/v1/demands?page=1&page_size=5")
check "GET /demands?page=1&page_size=5 → 200" 200 "$CODE"

# 5. Admin list with pagination
echo "[5] Admin List"
ADMIN_TOKEN=$(curl -s -X POST "$BASE/api/v1/admin/token" -H 'Content-Type: application/json' -d '{"role":"platform_admin"}' | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
ADMIN_AUTH="Authorization: Bearer $ADMIN_TOKEN"
CODE=$(code -H "$ADMIN_AUTH" "$BASE/api/v1/admin/demands?page=1&page_size=5")
check "GET /admin/demands?page=1&page_size=5 → 200 (pagination)" 200 "$CODE"

# 6. Contract flow
echo "[6] Contract Flow"
CID=$(curl -s -X POST "$BASE/api/v1/contracts" -H 'Content-Type: application/json' -H "$AUTH" -d '{"template_id":"tpl-001"}' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
CODE=$(code -X POST "$BASE/api/v1/contracts" -H 'Content-Type: application/json' -H "$AUTH" -d '{"template_id":"tpl-001"}')
check "POST /contracts → 201" 201 "$CODE"

# 7. Webhook
echo "[7] Webhook"
TS=$(date +%s)
CODE=$(code -X POST "$BASE/api/v1/webhooks/signing" -H 'Content-Type: application/json' -d "{\"event_id\":\"e2e-$(date +%s%N)\",\"contract_id\":\"$CID\",\"status\":\"sent\",\"timestamp\":$TS}")
check "POST /webhooks/signing → 200" 200 "$CODE"

# 8. Employment list with pagination
echo "[8] Employment Pagination"
CODE=$(code -H "$AUTH" "$BASE/api/v1/employment-requests?page=1&page_size=5")
check "GET /employment-requests?page=1&page_size=5 → 200" 200 "$CODE"

# 9. New Business Module APIs
echo "[9] New Biz APIs"
check "GET /experts → 200" 200 "$(code "$BASE/api/v1/experts")"
check "GET /cases → 200" 200 "$(code "$BASE/api/v1/cases")"
check "GET /compliance-docs → 200" 200 "$(code "$BASE/api/v1/compliance-docs")"
check "GET /achievements → 200" 200 "$(code "$BASE/api/v1/achievements")"
check "GET /competitions → 200" 200 "$(code "$BASE/api/v1/competitions")"
check "GET /events → 200" 200 "$(code "$BASE/api/v1/events")"
check "GET /industry-reports → 200" 200 "$(code "$BASE/api/v1/industry-reports")"
check "GET /industry-resources → 200" 200 "$(code "$BASE/api/v1/industry-resources")"
check "GET /emergency-resources → 200" 200 "$(code "$BASE/api/v1/emergency-resources")"
check "GET /recommendations → 200" 200 "$(code -H "$ADMIN_AUTH" "$BASE/api/v1/recommendations?biz_type=cable_inspection")"

# 10. Admin creates + Rate limiting
echo "[10] Admin Ops"
check "POST /admin/experts → 201" 201 "$(code -X POST "$BASE/api/v1/admin/experts" -H 'Content-Type: application/json' -H "$ADMIN_AUTH" -d '{"Name":"E2E专家","Title":"教授","Org":"重大","Field":"无人机","Bio":"简介"}')"

# Rate limiting
echo "[11] Rate Limiting"
for i in $(seq 1 210); do curl -s -o /dev/null "$BASE/healthz" 2>/dev/null; done
echo "  ℹ️  Rate limit test sent 210 requests (manual verification needed)"

# Summary
echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && echo "🎉 ALL E2E TESTS PASSED" || echo "💥 SOME TESTS FAILED"
exit $FAIL
