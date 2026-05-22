#!/bin/bash
# Automatic Let's Encrypt SSL for host nginx (Ubuntu/Debian).
#
# Prerequisites:
#   - DNS A record for DOMAIN → this server
#   - Ports 80 and 443 open
#   - Docker app listening on 127.0.0.1:8090
#
# Usage:
#   cd /home/sammy/marketmamba
#   SSL_EMAIL=magafu317@gmail.com sudo -E bash scripts/setup-ssl.sh
#
# Reads DOMAIN from TELEGRAM_LOGIN_DOMAIN or SSL_DOMAIN; email from SSL_EMAIL or .env
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${REPO_DIR}"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

DOMAIN="${SSL_DOMAIN:-${TELEGRAM_LOGIN_DOMAIN:-marketmamba.kkooapp.co.tz}}"
EMAIL="${SSL_EMAIL:-}"
WEBROOT="/var/www/certbot"
NGINX_SITE="/etc/nginx/sites-available/${DOMAIN}"
CERT_DIR="/etc/letsencrypt/live/${DOMAIN}"

if [ -z "${EMAIL}" ]; then
  echo "ERROR: Set SSL_EMAIL in .env or environment (required for automatic renewal)." >&2
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Run as root: sudo -E bash scripts/setup-ssl.sh" >&2
  exit 1
fi

export DOMAIN
bash "${REPO_DIR}/scripts/render-nginx.sh" both

echo "=== Packages ==="
if ! command -v nginx >/dev/null 2>&1; then
  apt-get update
  apt-get install -y nginx
fi
if ! command -v certbot >/dev/null 2>&1; then
  apt-get update
  apt-get install -y certbot
fi

echo "=== Webroot for ACME ==="
mkdir -p "${WEBROOT}"
chown -R www-data:www-data "${WEBROOT}" 2>/dev/null || true

echo "=== HTTP nginx (${DOMAIN}) ==="
cp "${REPO_DIR}/deploy/nginx-marketmamba-http.conf.generated" "${NGINX_SITE}"
ln -sf "${NGINX_SITE}" "/etc/nginx/sites-enabled/${DOMAIN}"
# Drop default site if it steals port 80
rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
nginx -t
systemctl enable nginx
systemctl reload nginx

echo "=== Certificate (obtain or renew) ==="
if [ -d "${CERT_DIR}" ]; then
  certbot renew --quiet --webroot -w "${WEBROOT}" --deploy-hook "${REPO_DIR}/deploy/certbot-renew-hook.sh" || true
else
  certbot certonly --webroot -w "${WEBROOT}" \
    -d "${DOMAIN}" \
    --non-interactive --agree-tos --email "${EMAIL}" \
    --deploy-hook "${REPO_DIR}/deploy/certbot-renew-hook.sh"
fi

echo "=== HTTPS nginx ==="
cp "${REPO_DIR}/deploy/nginx-marketmamba.conf.generated" "${NGINX_SITE}"
nginx -t
systemctl reload nginx

echo "=== Renewal hook + timer ==="
HOOK_DIR="/etc/letsencrypt/renewal-hooks/deploy"
mkdir -p "${HOOK_DIR}"
install -m 755 "${REPO_DIR}/deploy/certbot-renew-hook.sh" "${HOOK_DIR}/reload-nginx.sh"
systemctl enable certbot.timer 2>/dev/null || true
systemctl start certbot.timer 2>/dev/null || true

echo "=== Test HTTPS ==="
sleep 2
curl -sfI "https://${DOMAIN}/health" | head -8 || {
  echo "WARN: HTTPS check failed — confirm DNS and docker app on :8090" >&2
  exit 1
}

echo ""
echo "Done: https://${DOMAIN}"
echo "Certificates renew automatically (certbot.timer + deploy hook)."
