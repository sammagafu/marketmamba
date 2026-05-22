#!/bin/bash
# Run on VPS: cd /home/sammy/marketmamba && bash scripts/vps-fix.sh
set -e
cd "$(dirname "$0")/.."

echo "=== Stop old containers ==="
docker compose -p marketmamba down 2>/dev/null || true

echo "=== Free port 8090 if something else holds it ==="
sudo fuser -k 8090/tcp 2>/dev/null || true

echo "=== Build and start ==="
docker compose -p marketmamba up -d --build

echo "=== Wait for postgres ==="
sleep 5

echo "=== Migrations ==="
for f in migrations/002_broker_connections.sql migrations/003_users_subscriptions.sql migrations/004_web_admins.sql; do
  echo "Applying $f ..."
  docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < "$f" || true
done

echo "=== Seed admin (needs ADMIN_EMAIL in .env) ==="
docker compose -p marketmamba exec app ./server seed-admin || true

echo "=== Status ==="
docker compose -p marketmamba ps
curl -s http://127.0.0.1:8090/health || echo "health check failed"

echo "Done. Stop Market Mamba on your Mac: docker compose -p marketmamba down"
