#!/bin/bash
# Copy local .env to VPS and rebuild Market Mamba.
# Run from your Mac (SSH key for sammy@kkooapp.co.tz must work):
#   bash scripts/copy-env-deploy.sh
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${REPO_DIR}"

SSH_HOST="${DEPLOY_SSH:-sammy@kkooapp.co.tz}"
REMOTE_DIR="${DEPLOY_DIR:-/home/sammy/marketmamba}"
ENV_SRC="${ENV_FILE:-${REPO_DIR}/.env}"

if [ ! -f "${ENV_SRC}" ]; then
  echo "ERROR: Missing ${ENV_SRC}" >&2
  exit 1
fi

# VPS tweaks: production + Docker postgres hostname
TMP_ENV="$(mktemp)"
trap 'rm -f "${TMP_ENV}"' EXIT
sed -e 's/^APP_ENV=development/APP_ENV=production/' \
    -e 's|@localhost:5432|@postgres:5432|' \
    "${ENV_SRC}" > "${TMP_ENV}"

echo "→ Uploading .env to ${SSH_HOST}:${REMOTE_DIR}/.env"
scp "${TMP_ENV}" "${SSH_HOST}:${REMOTE_DIR}/.env"

echo "→ Pull, migrate, build, seed admin"
ssh "${SSH_HOST}" bash -s <<EOF
set -euo pipefail
cd "${REMOTE_DIR}"
if ! git remote get-url origin &>/dev/null; then
  echo "ERROR: git remote origin required" >&2
  exit 1
fi
git pull origin main
# Host nginx/apache blocks Caddy on 80/443 — stop before compose up
if command -v sudo >/dev/null 2>&1; then
  sudo bash scripts/free-web-ports.sh 2>/dev/null || bash scripts/free-web-ports.sh 2>/dev/null || true
else
  bash scripts/free-web-ports.sh 2>/dev/null || true
fi
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/006_auto_trade_approval.sql 2>/dev/null || true
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/007_binance_payment_orders.sql 2>/dev/null || true
docker compose -p marketmamba up -d --build
sleep 8
docker compose -p marketmamba exec app ./seedadmin || true
echo ""
echo "Health:"
curl -sf "https://marketmamba.kkooapp.co.tz/health" && echo "" || echo "WARN: health check failed"
echo "Config:"
curl -sf "https://marketmamba.kkooapp.co.tz/api/v1/config" | head -c 600
echo ""
docker compose -p marketmamba ps
EOF

echo "Done."
