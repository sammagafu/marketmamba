# How Market Mamba trades

Public-facing summary; implementation lives in `internal/trading`, `internal/risk`, and `internal/signals`.

## Pipeline

1. **Scan & filter** — Spread, ATR volatility, EMA trend stack (20 / 50 / 200), RSI band checks (`signal_generator.go`).
2. **Qualify** — Valid BUY/SELL, SL/TP, minimum strength, risk–reward vs settings (`signals/qualify.go`, `risk/risk.go`).
3. **Size & execute** — Lot from `MaxRiskPerTrade × balance` and stop distance; broker market order with SL/TP (`trading/executor.go`).
4. **Monitor & log** — TP/SL checks; Postgres trade log with risk/reward (`trading/monitor.go`, `trading/tradelog.go`).

## Default risk (new users / env)

| Setting | Default |
|---------|---------|
| Risk per trade | 0.5% of balance |
| Max daily loss | 2% of balance |
| Max open trades | 2 |
| Max trades per day | 10 |
| Min risk–reward | 1:1 |

Override via `.env` (`MAX_RISK_PER_TRADE`, etc.) or per-user `risk_settings` in the database.

## Signals

- **Broadcast**: Admin qualifies a setup and sends to Telegram subscribers.
- **Auto**: Users with auto-trading on and an active subscription get execution through the multi-user coordinator.

## Web copy

Landing section: `web/src/components/HowWeTrade.vue` (data in `web/src/howItWorks.js`). Update those files when behavior changes.
