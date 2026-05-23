#!/bin/bash
# Free ports 80/443 for Docker Caddy (host nginx/apache conflict).
# On VPS: sudo bash scripts/free-web-ports.sh
set -euo pipefail

echo "Checking ports 80 and 443..."
if command -v ss >/dev/null 2>&1; then
  ss -tlnp | grep -E ':80 |:443 ' || true
fi

for svc in nginx apache2 httpd; do
  if systemctl is-active --quiet "${svc}" 2>/dev/null; then
    echo "Stopping ${svc}..."
    systemctl stop "${svc}" || true
    systemctl disable "${svc}" 2>/dev/null || true
  fi
done

# Old Caddy on host (not in Docker)
if systemctl is-active --quiet caddy 2>/dev/null; then
  echo "Stopping host caddy service..."
  systemctl stop caddy || true
fi

echo "Port check after stop:"
if command -v ss >/dev/null 2>&1; then
  ss -tlnp | grep -E ':80 |:443 ' || echo "Ports 80/443 appear free."
fi

echo "Start Docker Caddy: docker compose -p marketmamba up -d caddy"
