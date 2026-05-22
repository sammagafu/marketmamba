#!/bin/bash
# Full VPS deploy: pull, build, start with automatic HTTPS (Caddy in Docker).
#   cd /home/sammy/marketmamba && bash scripts/vps-deploy.sh
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${REPO_DIR}"

if [ ! -f .env ]; then
  echo "ERROR: Missing .env — copy .env.example and set SSL_EMAIL, tokens, secrets." >&2
  exit 1
fi

git pull

# Host nginx conflicts with Caddy on ports 80/443
if [ "$(id -u)" -eq 0 ]; then
  systemctl stop nginx 2>/dev/null || true
else
  echo "Tip: stop host nginx if ports 80/443 are in use: sudo systemctl stop nginx"
fi

docker compose -p marketmamba up -d --build

DOMAIN="${TELEGRAM_LOGIN_DOMAIN:-marketmamba.kkooapp.co.tz}"
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
  DOMAIN="${TELEGRAM_LOGIN_DOMAIN:-${DOMAIN}}"
fi

echo "Waiting for TLS (first boot may take ~30s)..."
sleep 5
if curl -sfI "https://${DOMAIN}/health" >/dev/null 2>&1; then
  echo "HTTPS OK: https://${DOMAIN}"
else
  echo "WARN: https://${DOMAIN}/health not ready yet — check DNS, SSL_EMAIL, and: docker compose -p marketmamba logs caddy"
fi

echo "Deploy complete."
