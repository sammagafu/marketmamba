#!/bin/bash
# Run on VPS as root or with sudo:
#   cd /home/sammy/marketmamba && sudo bash scripts/setup-ssl.sh
set -e

DOMAIN="marketmamba.kkooapp.co.tz"
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
NGINX_SITE="/etc/nginx/sites-available/${DOMAIN}"

echo "=== Install certbot (Ubuntu/Debian) ==="
if ! command -v certbot >/dev/null 2>&1; then
  apt-get update
  apt-get install -y certbot python3-certbot-nginx
fi

echo "=== HTTP nginx (for certificate issue) ==="
cp "${REPO_DIR}/deploy/nginx-marketmamba-http.conf.example" "${NGINX_SITE}"
ln -sf "${NGINX_SITE}" "/etc/nginx/sites-enabled/${DOMAIN}"
nginx -t
systemctl reload nginx

echo "=== Obtain / renew certificate ==="
certbot --nginx -d "${DOMAIN}" --non-interactive --agree-tos --redirect \
  --email iammagafu@gmail.com \
  || certbot --nginx -d "${DOMAIN}"

echo "=== Apply full SSL nginx config (Telegram COOP header) ==="
cp "${REPO_DIR}/deploy/nginx-marketmamba.conf.example" "${NGINX_SITE}"
nginx -t
systemctl reload nginx

echo "=== Auto-renewal timer ==="
systemctl enable certbot.timer 2>/dev/null || true
systemctl start certbot.timer 2>/dev/null || true

echo "=== Test ==="
curl -sI "https://${DOMAIN}/health" | head -5
echo "Done: https://${DOMAIN}"
