#!/bin/bash
# Render nginx configs from templates. Usage:
#   DOMAIN=marketmamba.kkooapp.co.tz bash scripts/render-nginx.sh http|ssl|both
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DOMAIN="${DOMAIN:-marketmamba.kkooapp.co.tz}"
MODE="${1:-both}"

render() {
  local template="$1"
  local out="$2"
  sed "s/__DOMAIN__/${DOMAIN}/g" "${template}" > "${out}"
  echo "Wrote ${out}"
}

case "${MODE}" in
  http)
    render "${REPO_DIR}/deploy/nginx-marketmamba-http.conf.template" \
      "${REPO_DIR}/deploy/nginx-marketmamba-http.conf.generated"
    ;;
  ssl)
    render "${REPO_DIR}/deploy/nginx-marketmamba.conf.template" \
      "${REPO_DIR}/deploy/nginx-marketmamba.conf.generated"
    ;;
  both)
    bash "$(dirname "$0")/render-nginx.sh" http
    bash "$(dirname "$0")/render-nginx.sh" ssl
    ;;
  *)
    echo "Usage: DOMAIN=... $0 http|ssl|both" >&2
    exit 1
    ;;
esac
