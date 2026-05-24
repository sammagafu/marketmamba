# Market Mamba — Agent / operator guide

Use this file as the index for deployment and admin setup.

## Brand assets

Logos: `web/src/assets/images/` (`Logo-landscape.svg`, `Logo-potrait.svg`, `favcon.svg`). Favicon: `web/public/favicon.svg`. Rebuild with `make web-build`.

## Documentation

| Doc | Purpose |
|-----|---------|
| [VPS_DEPLOY.md](./VPS_DEPLOY.md) | **Start here** — `.env` variables, copy-paste VPS commands, nginx, SSL, admin seed |
| [WEB_DEPLOY.md](./WEB_DEPLOY.md) | Web dashboard, Telegram login, local dev proxy |
| [README.md](./README.md) | Project overview |
| [docs/HOW_WE_TRADE.md](./docs/HOW_WE_TRADE.md) | Signals, indicators, risk — aligns with landing copy |
| [docs/MINI_APP.md](./docs/MINI_APP.md) | Telegram Mini App, 5-day trial, 10 USDT/month Binance |
| [docs/BROKER_CONNECT.md](./docs/BROKER_CONNECT.md) | Connect Deriv, Exness, Tickmill, OANDA, Mock |
| [docs/SUBSCRIPTION_TIERS.md](./docs/SUBSCRIPTION_TIERS.md) | Plan limits: accounts, signals, long/short trades |

## Quick VPS commands

```bash
ssh sammy@kkooapp.co.tz
cd /home/sammy/marketmamba
cp .env.example .env && nano .env          # or scp .env from Mac (see VPS_DEPLOY.md)
# Set SSL_EMAIL in .env — HTTPS is automatic (Caddy on ports 80/443)
docker compose -p marketmamba up -d --build
# Or: bash scripts/vps-deploy.sh (stops host nginx, pulls, builds, checks https)
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/004_web_admins.sql
docker compose -p marketmamba exec app ./seedadmin
docker compose -p marketmamba logs -f app
```

**Copy local `.env` to VPS and rebuild** (from your Mac, SSH key required):

```bash
bash scripts/copy-env-deploy.sh
```

## Admin login (email)

1. On the VPS `.env` set `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_TELEGRAM_ID` (your Telegram user id).
2. Run `seed-admin` (see above).
3. Open the site → **Admin login (email)**.

**Do not commit real passwords to git.** Set them only in `.env` on the server.

## Telegram admin

Same Telegram ID must be in `TELEGRAM_ADMIN_USER_IDS`. Commands: `/admin stats`, `/admin activate <id> <days>`.

**Auto-trade (production):** set `SUBSCRIPTION_REQUIRED=true` and `AUTO_TRADE_REQUIRES_APPROVAL=true`, then `/approveauto <telegram_id>` before users can run `/auto on`.

## Access control (admin vs trader)

| Role | How assigned | API |
|------|----------------|-----|
| **admin** | Telegram ID in `TELEGRAM_ADMIN_USER_IDS` (+ email seed for web) | `/api/v1/admin/*` |
| **user** (trader) | Anyone who logs in via Telegram | `/api/v1/status`, brokers, trades, etc. |

- `/auth/me` returns `role`, `permissions`, `is_blocked`, `can_trade`.
- Blocked users get **403** on protected routes (admins are never blocked).
- Web dashboard hides admin panel unless `admin:stats` permission; trader badge vs admin badge in header.
- Details: [docs/ACL.md](./docs/ACL.md)
