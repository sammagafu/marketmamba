/**
 * Public copy for how Market Mamba trades — keep aligned with
 * internal/trading, internal/risk, and user defaults.
 */

export const TRADE_PIPELINE = [
  {
    step: '01',
    title: 'Scan markets',
    body:
      'Major pairs are screened for spread, volatility (ATR), EMA trend alignment, and RSI — before any setup is considered.',
  },
  {
    step: '02',
    title: 'Qualify setups',
    body:
      'Only trades with stop loss, take profit, minimum strength, and acceptable risk–reward pass through to broadcast or auto-execution.',
  },
  {
    step: '03',
    title: 'Size & route',
    body:
      'Position size follows your risk-per-trade setting. Orders are sent to your connected broker with SL and TP attached.',
  },
  {
    step: '04',
    title: 'Monitor & record',
    body:
      'Open trades are tracked against TP and SL. Every fill is logged with entry, exit, P/L, and reason — for a clear audit trail.',
  },
]

export const INDICATORS = [
  {
    tag: 'Trend',
    name: 'EMA 20 · 50 · 200',
    detail:
      'Stacked averages define trend direction and pullback zones near EMA 20 or 50.',
  },
  {
    tag: 'Momentum',
    name: 'RSI',
    detail:
      'Filters exhausted moves by avoiding extreme overbought and oversold readings.',
  },
  {
    tag: 'Volatility',
    name: 'ATR',
    detail:
      'Skips flat markets; helps set stop distance and take-profit versus your minimum R:R.',
  },
  {
    tag: 'Cost',
    name: 'Spread',
    detail:
      'Wide spreads on majors are rejected so execution cost does not erode the edge.',
  },
]

export const RISK_CONTROLS = [
  {
    name: 'Risk per trade',
    value: '0.5%',
    detail: 'Default maximum account risk per position (adjustable per user).',
  },
  {
    name: 'Daily loss cap',
    value: '2%',
    detail: 'Automation pauses when realized daily loss reaches this level.',
  },
  {
    name: 'Open exposure',
    value: '2 max',
    detail: 'Caps concurrent positions so risk cannot stack without limit.',
  },
  {
    name: 'Daily trade count',
    value: '10 max',
    detail: 'Reduces over-trading in noisy sessions.',
  },
  {
    name: 'Minimum R:R',
    value: '≥ 1:1',
    detail: 'Reward must justify stop distance before any order is placed.',
  },
  {
    name: 'Your control',
    value: 'Pause anytime',
    detail:
      'Turn automation off in Telegram or the dashboard. Active membership required; plan quotas apply monthly.',
  },
]

export const SIGNAL_SOURCES = [
  {
    title: 'Our signals, not channel copy',
    body:
      'Market Mamba does not mirror third-party Telegram VIP channels. You receive qualified setups from our engine, scoped to pairs you enable and limits on your plan.',
  },
  {
    title: 'Your broker, your capital',
    body:
      'Connect Deriv, Exness, Tickmill, or any MT4/MT5 server via MetaAPI. We are not a broker — funds and execution remain with you.',
  },
  {
    title: 'Automation on your terms',
    body:
      'Enable auto-trading when you are ready. Hard limits apply: per-trade risk, daily loss, open trades, and monthly tier quotas.',
  },
  {
    title: 'Simple USDT billing',
    body:
      'Start on a free trial, then pay in USDT through Binance in the Telegram app. Pro and team plans — contact us directly.',
  },
]

export const RISK_DISCLAIMER =
  'Forex and CFD trading involve substantial risk of loss. Past results and platform statistics do not guarantee future performance. Only risk capital you can afford to lose. Market Mamba is not a broker and does not provide investment advice.'
