# Real-time sniper decision support

Live prices drive **TAKE / SKIP / WAIT** decisions with confidence and reasons. No multi-year history archive is required—a rolling buffer of live samples builds indicators (EMA, RSI, ATR).

## Price feeds

| Symbol | Default source |
|--------|----------------|
| EURUSD | [Frankfurter](https://www.frankfurter.app/) (ECB rates) |
| BTCUSD | [CoinGecko](https://www.coingecko.com/) |

Optional: set `MARKET_DATA_API_KEY` (Twelve Data) for bid/ask quotes and 1-minute bar seeding (faster warm-up).

## Modes (`DECISION_MODE`)

| Value | Advisory Telegram | Assisted auto (`/autostart`) |
|-------|-------------------|------------------------------|
| `both` (default) | Yes | Yes, if confidence ≥ `SNIPER_MIN_CONFIDENCE` |
| `advisory` | Yes | No |
| `auto` | No | Yes (when confidence threshold met) |

## Env vars

```env
DECISION_ENABLED=true
DECISION_INTERVAL_SEC=60
SNIPER_MIN_CONFIDENCE=0.75
SNIPER_COOLDOWN_MIN=45
DECISION_MODE=both
MARKET_DATA_API_KEY=              # Twelve Data — recommended on VPS
SUBSCRIPTION_REQUIRED=true        # production
AUTO_TRADE_REQUIRES_APPROVAL=true # production — admin /approveauto
```

## OANDA (practice → live)

Dashboard or Telegram: connect with `api_token`, `account_id`, `practice=true` for fxTrade Practice.

API docs: https://developer.oanda.com/rest-live-v20/introduction/

## Telegram

- `/analyze EURUSD` — instant live decision for you
- Subscribers receive **TAKE** sniper broadcasts only (not every SKIP)
- `/autostart` users get personal sniper DMs on **TAKE** when advisory is on

## Warm-up

After deploy, the bot needs ~35 live samples per symbol (`DECISION_INTERVAL_SEC` ticks) before **TAKE** is possible. Until then you will see **WAIT** with “Building live history…”.
