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

## Web login

1. Open the site
2. Enter `WEB_API_KEY` and your **Telegram user ID**
3. Save — dashboard loads your account
4. Admins (in `TELEGRAM_ADMIN_USER_IDS`) see user stats + manual activate

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
