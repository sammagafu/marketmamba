# Market Mamba — Web & public bot

## Features

- **Public Telegram bot** — anyone can `/start` (creates user + 30-day trial)
- **Manual subscriptions** — no payment gateway yet; admin activates via Telegram or web
- **Vue dashboard** at `https://marketmamba.kkooapp.co.tz`
- **Per-user brokers** — Mock works; OANDA/MetaAPI/Alpaca listed for future adapters
- **Admin stats** — total users, active subs, auto-trading count

## Environment

```env
PUBLIC_MODE=true
SUBSCRIPTION_REQUIRED=false
FREE_TRIAL_DAYS=30
TELEGRAM_ADMIN_USER_IDS=5311857635
WEB_API_KEY=long_random_secret
BROKER_ENCRYPTION_KEY=long_random_secret
```

Set `SUBSCRIPTION_REQUIRED=true` when you start charging (after manual approval flow is working).

## Database migrations

```bash
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/002_broker_connections.sql
docker compose -p marketmamba exec -T postgres psql -U forexbot -d forexbot < migrations/003_users_subscriptions.sql
```

## Deploy

**Create `.env` on the VPS** (same folder as `docker-compose.yml`). Compose reads it automatically; without it you get empty `WEB_API_KEY` and web login breaks.

```bash
cd ~/forex-bot
cp .env.example .env
nano .env
```

Generate secrets on the server:

```bash
openssl rand -hex 32   # use for WEB_API_KEY
openssl rand -hex 32   # use for WEB_SESSION_SECRET
openssl rand -hex 32   # use for BROKER_ENCRYPTION_KEY (32+ chars)
```

Required in `.env`:

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_ADMIN_USER_IDS` (your Telegram user id)
- `WEB_API_KEY`, `WEB_SESSION_SECRET`, `BROKER_ENCRYPTION_KEY`

Then:

```bash
git pull
docker compose -p marketmamba up -d --build
```

## Web login

### Admin (email) — recommended on VPS

1. Set `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_TELEGRAM_ID` in VPS `.env`
2. Run `make vps-seed-admin` (see [VPS_DEPLOY.md](./VPS_DEPLOY.md))
3. Open the site → **Admin login (email)**

### Telegram Login (OIDC)

Uses the official [Telegram Login](https://core.telegram.org/bots/telegram-login) library (`telegram-login.js`).

1. @BotFather → **Bot Settings → Web Login** → add Allowed URL `https://marketmamba.kkooapp.co.tz`
2. Set `TELEGRAM_BOT_CLIENT_ID=8040019896` (same as BotFather Client ID)
3. Open the site → **Sign in with Telegram** (shine button)
4. Backend verifies JWT `id_token` via Telegram JWKS

### API key (fallback)

Manual login with `WEB_API_KEY` + Telegram user ID.

**Local dev:** widget often fails on `localhost` — use email admin or API key login.

## Telegram admin

- `/admin stats` — user counts
- `/admin activate <telegram_id> <days>` — extend subscription after manual payment

## Local Vue dev (use production API)

By default `web/.env.development` proxies `/api` to **production**:

```env
VITE_API_PROXY_TARGET=https://marketmamba.kkooapp.co.tz
```

```bash
cd web && npm install && npm run dev
# Open http://localhost:5173 — data comes from the VPS, not local Docker
```

To use a **local** backend instead, set `VITE_API_PROXY_TARGET=http://localhost:8090` and run `docker compose -p marketmamba up -d`.

**Manual login from localhost:** use the same `WEB_API_KEY` as on the VPS `.env` and your Telegram user ID.

**Telegram bot:** only run **one** instance per bot token. Stop local Docker if the VPS bot is running, or `/dailyreport` and trades will hit the wrong database.

## Build frontend for Go embed

```bash
make web-build
go build -o forex-bot ./cmd/server
```
