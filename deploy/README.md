# Deploy / TLS

| File | Purpose |
|------|---------|
| `nginx-marketmamba*.conf.template` | Host nginx — rendered by `scripts/render-nginx.sh` |
| `Caddyfile.template` | Docker Caddy automatic HTTPS |
| `certbot-renew-hook.sh` | Reload nginx after certificate renewal |

**Host (automatic):** `sudo -E bash scripts/setup-ssl.sh` (reads `SSL_EMAIL`, `TELEGRAM_LOGIN_DOMAIN` from `.env`)

**Docker (automatic):** `docker compose -f docker-compose.yml -f docker-compose.ssl.yml up -d`
