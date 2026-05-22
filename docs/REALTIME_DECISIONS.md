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

## Brokers (execution)

| Option | Use when |
|--------|----------|
| **Mock (Demo)** | Testing auto-trade, risk limits, Telegram flow — no real money |
| **Twelve Data** | Live prices for signals/sniper (`MARKET_DATA_API_KEY`) — not execution |
| **OANDA** | Only if they accept your country of residence |
| **MetaAPI + MT5** | Real trades via a broker you *can* open (Exness, IC Markets, etc.) — integration planned |

OANDA docs (if eligible): https://developer.oanda.com/rest-live-v20/introduction/

## Telegram

- `/analyze EURUSD` — instant live decision for you
- Subscribers receive **TAKE** sniper broadcasts only (not every SKIP)
- `/autostart` users get personal sniper DMs on **TAKE** when advisory is on

## Warm-up

After deploy, the bot needs ~35 live samples per symbol (`DECISION_INTERVAL_SEC` ticks) before **TAKE** is possible. Until then you will see **WAIT** with “Building live history…”.
