# Telegram Mini App — Market Mamba

Opens from the bot menu **📊 Dashboard** ([Telegram Web Apps](https://core.telegram.org/bots/webapps)).

## Features

- Auto-login via signed `initData` (no separate Telegram Login widget)
- **Activity dashboard**: all trades, open positions, today’s stats
- **Subscription**: 5-day free trial, then **10 USDT / month** via Binance
- **Binance Pay** checkout (if `BINANCE_PAY_*` merchant keys are set)
- **Manual USDT** to `BINANCE_PAY_UID` + submit transaction ID

## VPS configuration

```env
SUBSCRIPTION_REQUIRED=true
FREE_TRIAL_DAYS=5
SUBSCRIPTION_PRICE_USDT=10
SUBSCRIPTION_DAYS=30
MINI_APP_URL=https://marketmamba.kkooapp.co.tz

# Option A — Binance Pay merchant API
BINANCE_PAY_API_KEY=
BINANCE_PAY_SECRET=
BINANCE_PAY_CERT_SN=

# Option B — USDT transfer to your Binance account
BINANCE_PAY_UID=your_binance_uid
BINANCE_PAY_NETWORK=TRC20
```

@BotFather → **Bot Settings → Menu Button** → Web App URL = `MINI_APP_URL` (also set automatically on app start via API).

Webhook for Binance Pay (optional): `POST https://your-domain/api/v1/payments/binance/webhook`

## API

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `POST /api/v1/auth/telegram/webapp` | init_data | Mini App session |
| `GET /api/v1/miniapp/dashboard` | session | Trades, positions, subscription |
| `POST /api/v1/payments/binance/order` | session | Create 10 USDT order |
| `POST /api/v1/payments/binance/confirm` | session | Submit tx reference |
