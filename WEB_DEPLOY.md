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

```bash
git pull
docker compose -p marketmamba up -d --build
```

## Web login (Telegram)

1. In @BotFather: `/setdomain` → choose your bot → enter `marketmamba.kkooapp.co.tz`
2. Set `TELEGRAM_BOT_USERNAME=market_mamba_bot` and `WEB_SESSION_SECRET` in `.env`
3. Open the site → **Log in with Telegram**
4. Admins see stats after login (if your Telegram ID is in `TELEGRAM_ADMIN_USER_IDS`)

**Local dev:** Telegram widget often fails on `localhost` — use “manual login (dev)” with API key + Telegram ID.

## Telegram admin

- `/admin stats` — user counts
- `/admin activate <telegram_id> <days>` — extend subscription after manual payment

## Local Vue dev

```bash
cd web && npm install && npm run dev
# Proxies API to localhost:8090
```

## Build frontend for Go embed

```bash
make web-build
go build -o forex-bot ./cmd/server
```
