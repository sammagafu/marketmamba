#!/bin/bash
# Run on VPS: cd /home/sammy/marketmamba && bash scripts/diagnose.sh
set -e
cd "$(dirname "$0")/.."

echo "=== .env present? ==="
test -f .env && echo "OK: .env exists" || echo "MISSING: copy .env to $(pwd)/.env"

echo "=== Required .env keys ==="
grep -E '^(TELEGRAM_BOT_TOKEN|WEB_API_KEY|WEB_SESSION_SECRET|BROKER_ENCRYPTION_KEY|TELEGRAM_ADMIN_USER_IDS)=' .env 2>/dev/null \
  | sed 's/=.*/=***/' || echo "Fix .env — missing keys"

echo "=== Docker status ==="
docker compose -p marketmamba ps -a

echo "=== App logs (last 40 lines) ==="
docker compose -p marketmamba logs app --tail=40 2>&1 || true

echo "=== Port 8090 ==="
curl -s -o /dev/null -w "health:%{http_code}\n" http://127.0.0.1:8090/health || echo "8090 not responding"
