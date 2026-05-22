#!/bin/sh
# Render Caddyfile from template + env, then run Caddy
set -eu

DOMAIN="${TELEGRAM_LOGIN_DOMAIN:-marketmamba.kkooapp.co.tz}"
EMAIL="${SSL_EMAIL:-}"

if [ -z "${EMAIL}" ]; then
  echo "WARN: SSL_EMAIL unset — Caddy may not register with Let's Encrypt" >&2
fi

sed -e "s/__DOMAIN__/${DOMAIN}/g" -e "s/__SSL_EMAIL__/${EMAIL}/g" \
  /etc/caddy/Caddyfile.template > /etc/caddy/Caddyfile

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
