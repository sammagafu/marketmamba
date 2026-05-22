#!/bin/sh
# Render Caddyfile and run Caddy (automatic Let's Encrypt when ENABLE_SSL=true).
set -eu

DOMAIN="${TELEGRAM_LOGIN_DOMAIN:-marketmamba.kkooapp.co.tz}"
EMAIL="${SSL_EMAIL:-}"
ENABLE_SSL="${ENABLE_SSL:-true}"

CADDYFILE="/etc/caddy/Caddyfile"

if [ "${ENABLE_SSL}" = "false" ] || [ "${APP_ENV:-production}" = "development" ]; then
  cat > "${CADDYFILE}" <<EOF
:80 {
	encode gzip
	header Cross-Origin-Opener-Policy "same-origin-allow-popups"
	reverse_proxy app:8090
}
EOF
  echo "Caddy: HTTP only (ENABLE_SSL=false or APP_ENV=development)"
  exec caddy run --config "${CADDYFILE}" --adapter caddyfile
fi

if [ -z "${EMAIL}" ]; then
  echo "ERROR: Set SSL_EMAIL in .env for automatic HTTPS" >&2
  exit 1
fi

cat > "${CADDYFILE}" <<EOF
{
	email ${EMAIL}
}

${DOMAIN} {
	encode gzip
	header Cross-Origin-Opener-Policy "same-origin-allow-popups"
	reverse_proxy app:8090
}
EOF

echo "Caddy: automatic HTTPS for ${DOMAIN}"
exec caddy run --config "${CADDYFILE}" --adapter caddyfile
