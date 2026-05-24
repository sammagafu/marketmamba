# Subscription tiers & usage limits

Market Mamba enforces **per-plan quotas** each calendar month (UTC).

## Plans

| Plan | Broker accounts | Signals / month | Long (BUY) | Short (SELL) |
|------|-----------------|-----------------|------------|--------------|
| **trial** | 1 | 30 | 5 | 5 |
| **monthly** | 2 | 200 | 30 | 30 |
| **pro** | 5 | 1000 | 100 | 100 |
| **manual** | 10 | 10000 | 1000 | 1000 |

- **Broker accounts** — active connections saved in the dashboard (multiple allowed on higher tiers; one is *primary* for auto-trade).
- **Signals** — Telegram signal/sniper alerts delivered to the user.
- **Long / short** — auto-executed or manual `/open` trades counted when opened.

Telegram IDs in `TELEGRAM_ADMIN_USER_IDS` use **manual** limits (effectively unlimited for normal use).

## Assigning a plan

| Method | How |
|--------|-----|
| **Trial** | Automatic on `/start` (`plan=trial`) |
| **Monthly** | Binance USDT payment → `plan=monthly` |
| **Pro** | Admin: `/admin activate <telegram_id> <days>` with plan in web API, or `POST /api/v1/admin/activate` body `{ "telegram_id": 123, "days": 30, "plan": "pro" }` |
| **Manual** | Admin activate with `"plan": "manual"` for VIP / staff |

## API

- `GET /api/v1/subscription` — includes `tier.limits` and `tier.usage`
- `GET /api/v1/tiers` — public list of plan limits
- `GET /api/v1/brokers/connection` — `connections[]` and primary `connection`

## Migration

Run `migrations/008_subscription_tiers.sql` (also applied on app startup if migrations run).

## Multiple broker accounts

Each new save adds a connection and sets it **primary** (used for `/balance`, auto-trade). Older connections stay active until you exceed the tier account limit.
