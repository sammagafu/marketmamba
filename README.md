# Market Mamba

**Controlled forex automation on Telegram** — built-in risk limits, qualified signals, and execution on **the broker you already use** (Deriv, Exness, Tickmill, or any MT4/MT5 via [MetaAPI](https://metaapi.cloud/)).

Market Mamba is **not** a broker and **does not** copy trades from third-party Telegram signal channels. It runs **your** rules on **your** account.

## What you get

| Capability | Description |
|------------|-------------|
| **Controlled automation** | Qualified signals, position sizing from risk %, SL/TP on every trade |
| **Risk limits** | Per-trade risk, daily loss cap, max open trades, tier quotas (signals/trades/brokers) |
| **Any MT broker** | Connect via MetaAPI — brand presets or **Any MT broker** with your server name |
| **Telegram + web** | Bot commands, Mini App dashboard, web connect wizard |
| **Subscriptions** | Free trial, then **USDT via Binance** (no cards / no Stripe) |

## Risk disclaimer

Forex trading carries substantial risk of loss. This software is provided for educational and operational use. Test on **demo (mock)** before live credentials. The operators are not responsible for trading losses.

## Quick start

```bash
git clone git@github.com:sammagafu/marketmamba.git forex-bot
cd forex-bot
cp .env.example .env   # edit TELEGRAM_BOT_TOKEN, DATABASE_URL, secrets
```

### Docker (recommended)

```bash
docker compose up -d
# Apply migrations (001–009) — see make vps-migrate or run SQL under migrations/
make web-build         # embed Vue dashboard in Go binary
go build -o forex-bot cmd/server/main.go
```

### Local dev

```bash
go mod download
createdb forexbot
# Run all migrations in migrations/ in order
go run cmd/server/main.go
```

Web UI dev server: `cd web && npm install && npm run dev` (API on `HTTP_PORT`, default 8090).

## Configuration (highlights)

```env
# Telegram
TELEGRAM_BOT_TOKEN=
TELEGRAM_BOT_USERNAME=market_mamba_bot
PUBLIC_SITE_URL=https://your-domain.example

# Brokers (live traders)
ENABLED_BROKER_BRANDS=mock,deriv,exness,tickmill,any_mt
METAAPI_SHARED_TOKEN=          # optional — users skip MetaAPI token field
BROKER_ENCRYPTION_KEY=         # required in production (32+ chars)

# Risk (defaults for new users)
MAX_RISK_PER_TRADE=0.005
MAX_DAILY_LOSS=0.02
MAX_OPEN_TRADES=2
MAX_TRADES_PER_DAY=10

# Subscriptions — USDT only
SUBSCRIPTION_REQUIRED=true
FREE_TRIAL_DAYS=5
SUBSCRIPTION_PRICE_USDT=10
SUBSCRIPTION_DAYS=30
SUBSCRIPTION_CONTACT=Pay in USDT via Binance only (no cards). Pro plans? Contact us on Telegram.
CONTACT_US_URL=https://t.me/your_bot
CONTACT_US_LABEL=Contact us
VALUE_PROPOSITION=Controlled automation with built-in risk limits — connect the broker you already use.

# Community launch (Bitcoin-first; unlock is internal — no quota UI)
ASSET_PHASE=bitcoin
SIGNAL_BROADCAST_SYMBOLS=BTCUSD,ETHUSD
UNLOCK_MIN_PAID_SUBSCRIBERS=100
AI_TRAINING_NOTE=We're training our bots with AI for more precise entries.

# Binance USDT checkout (Mini App)
BINANCE_PAY_API_KEY=
BINANCE_PAY_SECRET=
```

Full template: [`.env.example`](./.env.example).

## Broker connection

| Brand | Path |
|-------|------|
| **Demo** | Telegram `/broker connect` or web wizard → Mock |
| **Deriv / Exness / Tickmill** | Web **Connect broker** → MetaAPI token + MT login, password, **exact server name** |
| **Any MT broker** | Same wizard — enter your broker’s MT server (e.g. IC Markets, Pepperstone) |

Guide: [`docs/BROKER_CONNECT.md`](./docs/BROKER_CONNECT.md)

First MetaAPI deploy often takes **1–3 minutes**. Market Mamba is not a broker — funds stay at your broker.

## Subscription tiers

Plans limit broker accounts, signals per month, and long/short auto-trades. See [`docs/SUBSCRIPTION_TIERS.md`](./docs/SUBSCRIPTION_TIERS.md).

| Plan | Typical use |
|------|-------------|
| **trial** | Auto on `/start` |
| **monthly** | USDT payment via Mini App / Binance |
| **pro / manual** | Admin activate — contact operator |

**Payments:** USDT via Binance only. No Stripe or card billing in-app.

## Telegram commands (traders)

| Command | Purpose |
|---------|---------|
| `/start` | Register, trial, help |
| `/subscribe` | Plans & USDT payment info |
| `/myplan` | Current subscription |
| `/broker connect web` | Link to connect wizard |
| `/pairs` | Signal & auto-trade symbols |
| `/balance` `/positions` `/trades` | Account view |
| `/open` `/close` | Manual trades |
| `/autostart` `/autostop` | Automation toggle |
| `/analyze SYMBOL` | Sniper decision (TAKE/SKIP/WAIT) |

Admins: `/admin activate`, signal broadcast, user block — see [`Agent.md`](./Agent.md).

## How trading works

1. **Scan & filter** — EMA stack, RSI, ATR, spread checks  
2. **Qualify** — SL/TP, strength, minimum R:R  
3. **Size & execute** — Risk-based lots on your connected broker  
4. **Monitor & log** — TP/SL, per-user trade history  

Details: [`docs/HOW_WE_TRADE.md`](./docs/HOW_WE_TRADE.md) · Filter pipeline: [`docs/FILTER_FRAMEWORK.md`](./docs/FILTER_FRAMEWORK.md)

**Inspect filters:** Dashboard → **Filter stack**, or `GET /api/v1/filters/report?symbol=EURUSD` (authenticated).

## Project layout

```
cmd/server/           # Entry point
internal/broker/      # Adapters, catalog, MetaAPI, connect flow
internal/tier/        # Plan limits & usage
internal/trading/     # Coordinator, executor, monitor
internal/signals/     # Broadcast & qualification
internal/telegram/    # Bot handlers
internal/api/         # REST + embedded web/dist
web/                  # Vue dashboard & Mini App
migrations/           # 001–009 PostgreSQL schema
docs/                 # Operator & user guides
```

## Testing

```bash
go test ./...
make web-build && go build ./cmd/server/
```

## Production deploy

**Guides:** [VPS_DEPLOY.md](./VPS_DEPLOY.md) · [Agent.md](./Agent.md) · [WEB_DEPLOY.md](./WEB_DEPLOY.md)

```bash
make vps-up
make vps-migrate    # applies migrations
make vps-seed-admin
```

After deploy, run migrations **008** (`broker multi-account`) and **009** (`tier usage`) if not already applied.

## What Market Mamba does *not* do

- **Copy external Telegram signal channels** (no channel listener / AI copier)  
- **Hold client funds** (execution is at the user’s broker)  
- **Accept card/Stripe payments** (USDT via Binance + manual admin activate)

## Security

- Encrypt broker credentials (`BROKER_ENCRYPTION_KEY`)  
- Never commit `.env`  
- Restrict `TELEGRAM_ADMIN_USER_IDS`  
- Use HTTPS in production (Caddy/nginx — see deploy docs)

## License

Provided as-is. See repository for terms.

---

**Support:** set `CONTACT_US_URL` (defaults to `https://t.me/<bot username>`). Pro, teams, or payment help → contact operator on Telegram.
