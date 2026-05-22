#!/bin/bash
# Full VPS deploy: pull, build, start app, ensure automatic SSL (host nginx).
#   cd /home/sammy/marketmamba && sudo -E bash scripts/vps-deploy.sh
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${REPO_DIR}"

git pull

docker compose -p marketmamba up -d --build

if command -v nginx >/dev/null 2>&1 || [ "$(id -u)" -eq 0 ]; then
  if [ "$(id -u)" -eq 0 ]; then
    bash "${REPO_DIR}/scripts/setup-ssl.sh"
  else
    echo "Run SSL setup as root: sudo -E bash scripts/setup-ssl.sh"
  fi
else
  echo "Skip SSL: nginx not installed (or use docker-compose.ssl.yml + Caddy)."
fi

echo "Deploy complete."
