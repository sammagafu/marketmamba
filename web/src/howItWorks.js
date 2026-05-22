/**
 * Public copy for how Market Mamba trades — keep aligned with
 * internal/trading/signal_generator.go, internal/risk, and user defaults.
 */

export const TRADE_PIPELINE = [
  {
    step: '01',
    title: 'Scan & filter',
    body:
      'We watch major pairs and run each setup through spread, volatility (ATR), EMA trend alignment, and RSI bands before a signal is even considered.',
  },
  {
    step: '02',
    title: 'Qualify the signal',
    body:
      'Only BUY/SELL ideas with defined stop loss and take profit, minimum strength, and a risk–reward ratio that meets platform rules are qualified for broadcast or auto-execution.',
  },
  {
    step: '03',
    title: 'Size & execute',
    body:
      'Lot size is calculated from your account balance and risk-per-trade settings. Orders go to your connected broker with SL and TP attached; retries handle transient broker errors.',
  },
  {
    step: '04',
    title: 'Monitor & log',
    body:
      'Open positions are checked against TP and SL. Every open and close is written to your trade log with entry, exit, P/L, and closure reason for a full audit trail.',
  },
]

export const INDICATORS = [
  {
    tag: 'Trend',
    name: 'EMA 20 · 50 · 200',
    detail:
      'Stacked moving averages define strong uptrend/downtrend and pullback entries near EMA20 or EMA50.',
  },
  {
    tag: 'Momentum',
    name: 'RSI',
    detail:
      'Avoids extreme overbought/oversold zones (roughly below 20 or above 80) so scalps are not taken into exhausted moves.',
  },
  {
    tag: 'Volatility',
    name: 'ATR',
    detail:
      'Filters dead markets and sets stop loss (~1× ATR) and take profit (ATR × your minimum risk–reward).',
  },
  {
    tag: 'Execution',
    name: 'Spread',
    detail:
      'Signals are rejected when spread is too wide (~3 pips on majors) so entries are not eaten by cost.',
  },
]

export const RISK_CONTROLS = [
  {
    name: 'Risk per trade',
    value: '0.5%',
    detail: 'Default cap on account balance risked per position (configurable per user).',
  },
  {
    name: 'Daily loss limit',
    value: '2%',
    detail: 'Trading stops for the day once realized loss hits this share of balance.',
  },
  {
    name: 'Open positions',
    value: '2 max',
    detail: 'Limits concurrent exposure so the book does not stack unchecked.',
  },
  {
    name: 'Trades per day',
    value: '10 max',
    detail: 'Prevents over-trading in choppy sessions.',
  },
  {
    name: 'Risk–reward',
    value: '≥ 1:1',
    detail: 'Every signal must meet minimum reward vs. stop distance before execution.',
  },
  {
    name: 'Pause & subscription',
    value: 'You control',
    detail:
      'Pause automation anytime. Active subscription (or trial) is required; admins can block accounts that breach policy.',
  },
]

export const SIGNAL_SOURCES = [
  {
    title: 'Qualified Telegram signals',
    body:
      'Pick your pairs (EURUSD, BTCUSD, …) in the dashboard or via /pairs. You only receive signals and TP/SL updates for pairs you enable.',
  },
  {
    title: 'Your broker, your book',
    body:
      'Connect a supported broker (or mock for testing). Execution and balances stay on your account; Market Mamba orchestrates rules and logging.',
  },
  {
    title: 'Automation you toggle',
    body:
      'Turn auto-trading on from the dashboard or Telegram (/autostart). When off, you still get signals and can manage trades manually.',
  },
]

export const RISK_DISCLAIMER =
  'Forex and CFD trading carry substantial risk of loss. Past performance and logged trade counts do not guarantee future results. Only trade with capital you can afford to lose.'
