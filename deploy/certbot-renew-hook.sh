#!/bin/bash
# Reload nginx after certbot renew (installed by scripts/setup-ssl.sh)
set -e
if command -v nginx >/dev/null 2>&1; then
  nginx -t && systemctl reload nginx
fi
