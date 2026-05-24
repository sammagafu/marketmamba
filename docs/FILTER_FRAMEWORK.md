# Filter framework

Market Mamba uses a **layered filter pipeline** before any signal is broadcast or auto-executed. Every gate is auditable — operators and clients can inspect the same stack via the dashboard **Filter stack** or API.

## Layers

```text
Market quality → Technical filters → Setup & signal → Risk envelope
```

| Layer | Purpose |
|-------|---------|
| **Market** | Live history depth, spread width |
| **Technical** | ATR floor, RSI band, EMA trend classification |
| **Setup** | Pattern match, strength, R:R from ATR-based SL/TP |
| **Risk** | Structural validation (SL/TP, limits) |
| **Platform** | Broadcast minimum strength |

## API

### Catalog (public)

`GET /api/v1/filters/catalog`

Returns all documented gates with thresholds.

### Live report (authenticated)

`GET /api/v1/filters/report?symbol=EURUSD`

Returns a `report` object:

- `verdict`: `pass` | `fail` | `warn`
- `qualified`: whether the symbol would broadcast
- `layers[]`: steps with `status`, `message`, `metrics`
- `data_source`: live provider name or `simulated`

Uses live market data when `MARKET_DATA_API_KEY` / price feeds are available; otherwise a representative simulated snapshot.

## Code

| Package | Role |
|---------|------|
| `internal/filter` | Pipeline types, `RunTechnical`, `AppendRisk`, `Service` |
| `internal/signalgen` | Pattern generation after filters pass |
| `internal/signals/qualify.go` | `MeetsRequirements` for broadcast |
| `internal/decision` | Sniper TAKE/SKIP/WAIT on live bars |

## Design principles

1. **Fail fast** — spread and volatility reject before expensive setup logic.
2. **Explain every reject** — each step returns a human message + metrics map for UI.
3. **Same rules everywhere** — broadcast, sniper `/analyze`, and filter report share `signalgen` + `risk`.
4. **No channel copy** — filters apply to **our** engine only, not third-party Telegram parsers.

## Thresholds (defaults)

| Gate | Rule |
|------|------|
| Spread | ~3 pips on majors; wider on BTC/XAU |
| ATR | ≥ 0.05% of price |
| RSI | 20–80 |
| Strength | ≥ `SIGNAL_MIN_STRENGTH` (0.7) |
| R:R | ≥ `RISK_REWARD_RATIO` (1.0) |

Tune via `.env` and per-user `risk_settings` in the database.
