# Deploy / TLS

| File | Purpose |
|------|---------|
| `nginx-marketmamba*.conf.template` | Host nginx — rendered by `scripts/render-nginx.sh` |
| `Caddyfile.template` | Docker Caddy automatic HTTPS |
| `certbot-renew-hook.sh` | Reload nginx after certificate renewal |

**Default (Docker):** `docker compose -p marketmamba up -d --build` — Caddy on 80/443, needs `SSL_EMAIL` in `.env`

**Optional host nginx:** `sudo -E bash scripts/setup-ssl.sh` (stop/disable Docker `caddy` first)
