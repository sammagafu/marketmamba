# Connecting your broker — Market Mamba

Market Mamba is **not a broker**. You connect **your own** trading account. The bot sends signals and can auto-execute trades on that account with your risk settings.

## Recommended: MetaAPI MT bridge (live traders)

For **Deriv, Exness, Tickmill**, and other **MT4/MT5** brokers, Market Mamba uses the **[MetaAPI](https://metaapi.cloud/) cloud bridge**. You do not install MetaTrader on the server — you provide:

1. **MetaAPI API token** (from [app.metaapi.cloud](https://app.metaapi.cloud/) → API access)
2. Your broker’s **MT login**, **password**, and **exact server name** (e.g. `Deriv-Demo`, `Exness-MT5Trial`)

The dashboard wizard saves these encrypted; the backend talks to MetaAPI, which connects to your broker.

### Operator: shared MetaAPI token (easiest for clients)

If you set `METAAPI_SHARED_TOKEN` in the server `.env` (your MetaAPI API token from [app.metaapi.cloud](https://app.metaapi.cloud/)), **clients do not need their own MetaAPI account**. They only enter:

- MT login (account number)
- MT password
- MT server name (exact string from their broker)

Redeploy after setting the variable. The connect wizard hides the MetaAPI token field when this is enabled.

## Quick paths

| Goal | How |
|------|-----|
| Try the bot with no risk | Telegram: `/broker connect` or web **Demo (Mock)** |
| Any MT4/MT5 broker | Web → **Connect broker** → **Any MT broker** (enter your server name) |
| Deriv / Exness / Tickmill (MT) | Same wizard — named tiles with preset hints |
| OANDA (optional) | Enable `oanda` in `ENABLED_BROKER_BRANDS`, then choose OANDA in wizard |

---

## Demo (Mock)

1. Open the web dashboard or Telegram.
2. Run `/broker connect` in Telegram, or click **Connect Mock Demo** on the web.
3. Check balance with `/balance`.

No API keys required. Simulated $10,000 balance.

---

## Deriv, Exness, Tickmill (via MetaAPI)

These brokers use **MetaTrader (MT4/MT5)**. Market Mamba connects through [MetaAPI](https://metaapi.cloud/) — you need:

1. A **MetaAPI** account and **API token** ([app.metaapi.cloud](https://app.metaapi.cloud/) → API access).
2. Your broker’s **MT login**, **password**, and **exact server name** (from the broker’s client area).

### Steps

1. Log in to the Market Mamba web dashboard.
2. Open **Connect broker** (or go to `#/connect`).
3. Choose **Deriv**, **Exness**, or **Tickmill**.
4. Fill in:
   - **MetaAPI token**
   - **MT login** (account number)
   - **MT password**
   - **MT server** — must match exactly, e.g. `Deriv-Demo`, `Exness-MT5Trial`, `Tickmill-Demo`
5. Click **Test connection** — wait up to 1–3 minutes on first connect while MetaAPI deploys your account.
6. Click **Save**.
7. In Telegram: `/balance` then `/autostart` (requires active subscription; production may need admin `/approveauto`).

### Server name examples

| Broker | Example demo servers |
|--------|----------------------|
| Deriv | `Deriv-Demo`, `Deriv-Server` |
| Exness | `Exness-MT5Trial`, `Exness-Trial` |
| Tickmill | `Tickmill-Demo`, `Tickmill-Live` |

Wrong server names are the most common connection failure.

---

## OANDA

1. Create an OANDA account and generate a **v20 API token**.
2. Web dashboard → **Connect broker** → **OANDA**.
3. Enter API token, account ID, and enable **Practice** if using a demo account.
4. Test → Save.

OANDA is not available in all countries. Use Mock or MetaAPI if signup is blocked.

---

## After connecting

- **Subscription:** Active plan required when `SUBSCRIPTION_REQUIRED=true`.
- **Auto-trade:** `/autostart` in Telegram; may require admin approval in production.
- **Pairs:** `/pairs` to choose symbols for signals and automation.
- **Important:** Do not place manual trades on the same MT account while auto-trade is on — the bot tracks its own positions in the database but the broker API sees the full account.

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| MetaAPI “not deployed” / timeout | Wait 1–3 min after first save; run **Test** again |
| Invalid credentials | Re-copy MT login, password, server from broker |
| `BROKER_ENCRYPTION_KEY` error on VPS | Set 32+ char key in `.env` and rebuild |
| Symbol not found | Add pair via `/pairs`; some symbols differ on MT (e.g. `frxEURUSD`) |
| Subscription blocked | `/subscribe` or pay in Mini App |

---

## Security

- Credentials are encrypted at rest (`BROKER_ENCRYPTION_KEY` on the server).
- API never returns your passwords after save.
- Use demo accounts until you trust the setup.
