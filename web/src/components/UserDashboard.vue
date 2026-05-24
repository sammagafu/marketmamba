<script setup>
import { computed, ref } from 'vue'
import { VALUE_PROPOSITION, PAYMENT_NOTE } from '../brand'

const props = defineProps({
  status: { type: Object, default: null },
  account: { type: Object, default: null },
  subscription: { type: Object, default: null },
  config: { type: Object, default: null },
  positions: { type: Array, default: () => [] },
  trades: { type: Array, default: () => [] },
  telegramId: { type: [String, Number], default: '' },
  canTrade: { type: Boolean, default: true },
})

defineEmits(['refresh'])

const tradeFilter = ref('all')

const openTrades = computed(() => props.trades.filter((t) => t.status === 'OPEN'))
const closedTrades = computed(() => props.trades.filter((t) => t.status === 'CLOSED'))

const filteredTrades = computed(() => {
  if (tradeFilter.value === 'open') return openTrades.value
  if (tradeFilter.value === 'closed') return closedTrades.value
  return props.trades
})

const netClosedPL = computed(() =>
  closedTrades.value.reduce((sum, t) => sum + (Number(t.profit) || 0), 0),
)

const subscriptionLabel = computed(() => {
  const sub = props.subscription?.subscription
  if (!sub) {
    const days = props.config?.free_trial_days ?? props.config?.trial_days ?? 5
    return `Free trial · ${days} days`
  }
  return `${sub.plan} · ${sub.status}`
})

const subscriptionExpires = computed(() => {
  const exp = props.subscription?.subscription?.expires_at
  return exp ? new Date(exp).toLocaleDateString() : '—'
})

const tierInfo = computed(() => props.subscription?.tier || null)

function usagePct(used, max) {
  if (!max || max <= 0) return 0
  return Math.min(100, Math.round((used / max) * 100))
}

const miniAppUrl = computed(() => props.config?.mini_app_url || props.config?.public_site_url || '')

const valueProp = computed(
  () => props.config?.value_proposition || VALUE_PROPOSITION,
)
const contactUrl = computed(() => props.config?.contact_us_url || '')
const contactLabel = computed(() => props.config?.contact_us_label || 'Contact us')
const priceUsdt = computed(() => props.config?.subscription_price_usdt ?? 10)
const trialDays = computed(
  () => props.config?.free_trial_days ?? props.config?.trial_days ?? 5,
)

function fmtProfit(t) {
  if (t.profit == null) return '—'
  const n = Number(t.profit)
  return `${n >= 0 ? '+' : ''}$${n.toFixed(2)}`
}

function fmtTime(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}
</script>

<template>
  <section class="user-dashboard wide">
    <header class="dash-intro copy-block">
      <p class="section-eyebrow">Client dashboard</p>
      <h2 class="section-title">Account overview</h2>
      <p class="section-lead dash-lead">{{ valueProp }}</p>
    </header>

    <p class="isolation-hint">
      Automation runs on your linked broker only. Pause auto-trading before placing manual trades on the same account.
    </p>

    <div class="dash-head">
      <div class="dash-title">
        <p class="section-eyebrow">At a glance</p>
      </div>
      <div class="dash-head-actions">
        <span v-if="telegramId" class="id-pill">ID <code>{{ telegramId }}</code></span>
        <button type="button" class="btn-secondary" @click="$emit('refresh')">Refresh</button>
      </div>
    </div>

    <div class="stat-grid">
      <div class="stat-card">
        <span class="stat-label">Subscription</span>
        <strong class="stat-value">{{ subscriptionLabel }}</strong>
        <span class="stat-sub muted">Expires {{ subscriptionExpires }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Balance</span>
        <strong class="stat-value">${{ account?.balance != null ? Number(account.balance).toFixed(2) : '—' }}</strong>
        <span class="stat-sub muted">Equity ${{ account?.equity != null ? Number(account.equity).toFixed(2) : '—' }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Open positions</span>
        <strong class="stat-value">{{ positions.length }}</strong>
        <span class="stat-sub muted">{{ openTrades.length }} open in log</span>
      </div>
      <div class="stat-card" :class="{ profit: netClosedPL >= 0, loss: netClosedPL < 0 }">
        <span class="stat-label">Your closed P/L</span>
        <strong class="stat-value">{{ netClosedPL >= 0 ? '+' : '' }}${{ netClosedPL.toFixed(2) }}</strong>
        <span class="stat-sub muted">{{ closedTrades.length }} closed trades</span>
      </div>
    </div>

    <div v-if="tierInfo" class="card card-bull tier-usage">
      <p class="section-eyebrow">Membership</p>
      <h3 class="section-title tier-title">Plan usage · {{ tierInfo.limits?.plan || 'trial' }}</h3>
      <p class="muted small tier-sub">Monthly allowance (UTC). Period starts {{ tierInfo.usage?.period_start || '—' }}</p>
      <ul class="usage-list">
        <li>
          <span>Broker accounts</span>
          <strong>{{ tierInfo.usage?.broker_accounts ?? 0 }} / {{ tierInfo.limits?.max_broker_accounts ?? '—' }}</strong>
        </li>
        <li>
          <span>Signals received</span>
          <strong>{{ tierInfo.usage?.signals_received ?? 0 }} / {{ tierInfo.limits?.max_signals_per_period ?? '—' }}</strong>
        </li>
        <li>
          <span>Long trades (BUY)</span>
          <strong>{{ tierInfo.usage?.long_trades ?? 0 }} / {{ tierInfo.limits?.max_long_trades ?? '—' }}</strong>
        </li>
        <li>
          <span>Short trades (SELL)</span>
          <strong>{{ tierInfo.usage?.short_trades ?? 0 }} / {{ tierInfo.limits?.max_short_trades ?? '—' }}</strong>
        </li>
      </ul>
    </div>

    <div class="status-row">
      <div class="status-chip-row">
        <span class="pill" :class="status?.can_trade ? 'ok' : 'warn'">
          {{ status?.can_trade ? 'Can trade' : 'Trading locked' }}
        </span>
        <span class="pill" :class="status?.auto_trading ? 'ok' : ''">
          Auto {{ status?.auto_trading ? 'ON' : 'OFF' }}
        </span>
        <span class="pill muted-pill">Broker: {{ status?.provider || '—' }}</span>
      </div>
      <p v-if="status?.trade_message && !canTrade" class="warn-text">{{ status.trade_message }}</p>
      <p v-else-if="subscription?.message" class="muted">{{ subscription.message }}</p>
    </div>

    <div v-if="miniAppUrl" class="mini-cta card-inline">
      <div>
        <p class="section-eyebrow">Billing</p>
        <strong class="cta-title">Manage subscription in Telegram</strong>
        <p class="muted cta-body">
          {{ trialDays }}-day evaluation, then <strong>{{ priceUsdt }} USDT</strong> per month via Binance.
          {{ PAYMENT_NOTE }}
        </p>
        <p v-if="contactUrl" class="muted contact-line">
          <a :href="contactUrl" target="_blank" rel="noopener">{{ contactLabel }}</a>
          for Pro, enterprise, or billing support.
        </p>
      </div>
      <a
        class="btn-primary mini-link"
        :href="`https://t.me/${config?.telegram_bot_username || 'market_mamba_bot'}`"
        target="_blank"
        rel="noopener"
      >Open @{{ config?.telegram_bot_username || 'market_mamba_bot' }}</a>
    </div>

    <section class="card card-bull dash-section">
      <p class="section-eyebrow">Live book</p>
      <h3 class="section-title section-title-sm">Open positions</h3>
      <div class="table-wrap">
        <table v-if="positions.length">
          <thead>
            <tr><th>Symbol</th><th>Side</th><th>Qty</th><th>Entry</th><th>P/L</th></tr>
          </thead>
          <tbody>
            <tr v-for="p in positions" :key="p.id || p.symbol + p.type">
              <td><strong>{{ p.symbol }}</strong></td>
              <td :class="p.type === 'BUY' ? 'buy' : 'sell'">{{ p.type }}</td>
              <td>{{ p.quantity }}</td>
              <td>{{ Number(p.entry_price || 0).toFixed(5) }}</td>
              <td :class="{ profit: p.profit > 0, loss: p.profit < 0 }">
                {{ p.profit != null ? '$' + Number(p.profit).toFixed(2) : '—' }}
              </td>
            </tr>
          </tbody>
        </table>
        <p v-else class="muted empty-hint">No open positions on your account.</p>
      </div>
    </section>

    <section class="card trade-log-card dash-section">
      <div class="trade-log-head">
        <div>
          <p class="section-eyebrow">History</p>
          <h3 class="section-title section-title-sm">Trade log</h3>
        </div>
        <div class="filter-tabs">
          <button type="button" :class="{ active: tradeFilter === 'all' }" @click="tradeFilter = 'all'">
            All ({{ trades.length }})
          </button>
          <button type="button" :class="{ active: tradeFilter === 'open' }" @click="tradeFilter = 'open'">
            Open ({{ openTrades.length }})
          </button>
          <button type="button" :class="{ active: tradeFilter === 'closed' }" @click="tradeFilter = 'closed'">
            Closed ({{ closedTrades.length }})
          </button>
        </div>
      </div>
      <div class="table-wrap">
        <table v-if="filteredTrades.length">
          <thead>
            <tr>
              <th>Time</th><th>Symbol</th><th>Side</th><th>Qty</th><th>Entry</th><th>Status</th><th>P/L</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in filteredTrades" :key="t.id">
              <td>{{ fmtTime(t.created_at) }}</td>
              <td><strong>{{ t.symbol }}</strong></td>
              <td :class="t.type === 'BUY' ? 'buy' : 'sell'">{{ t.type }}</td>
              <td>{{ t.quantity }}</td>
              <td>{{ Number(t.entry_price).toFixed(5) }}</td>
              <td>
                <span class="status-chip" :class="t.status?.toLowerCase()">{{ t.status }}</span>
                <span v-if="t.closure_reason" class="muted small"> {{ t.closure_reason }}</span>
              </td>
              <td :class="{ profit: t.profit > 0, loss: t.profit < 0 }">{{ fmtProfit(t) }}</td>
            </tr>
          </tbody>
        </table>
        <p v-else class="muted empty-hint">
          No trades in this view yet.<br />
          Link your broker, then use <code>/open</code> or <code>/autostart</code> in Telegram.
        </p>
      </div>
    </section>
  </section>
</template>

<style scoped>
.user-dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  width: 100%;
  grid-column: 1 / -1;
}

.dash-intro {
  margin-bottom: 0.25rem;
}

.dash-lead {
  margin-bottom: 0;
}

.isolation-hint {
  margin: 0;
  padding: 0.85rem 1rem;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--warn);
  background: var(--warn-bg);
  border: 1px solid var(--warn-border);
  border-radius: 10px;
}

.section-title-sm {
  font-size: 1.0625rem;
  margin-bottom: 0.75rem;
}

.tier-title {
  margin-bottom: 0.35rem;
}

.tier-sub {
  margin: 0 0 0.75rem;
}

.cta-title {
  display: block;
  font-size: 1rem;
  margin-bottom: 0.35rem;
}

.cta-body {
  margin: 0;
  line-height: 1.55;
}

.dash-head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: space-between;
  align-items: flex-start;
}

.dash-title h2 {
  margin: 0 0 0.25rem;
  font-size: clamp(1.1rem, 4vw, 1.35rem);
}

.contact-line {
  margin: 0.35rem 0 0;
  font-size: 0.82rem;
}

.contact-line a {
  color: var(--brand);
  font-weight: 600;
}

.dash-title p {
  margin: 0;
  font-size: 0.85rem;
}

.dash-head-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.id-pill {
  font-size: 0.8rem;
  padding: 0.25rem 0.5rem;
  border-radius: 8px;
  background: var(--chart-bg);
  border: 1px solid var(--border);
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 140px), 1fr));
  gap: 0.65rem;
}

.stat-card {
  padding: 0.85rem 1rem;
  border-radius: 12px;
  background: var(--card);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.stat-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--muted);
}

.stat-value {
  font-size: 1.25rem;
  line-height: 1.2;
}

.stat-sub {
  font-size: 0.75rem;
}

.stat-card.profit .stat-value { color: var(--brand); }
.stat-card.loss .stat-value { color: var(--loss-text); }

.status-row {
  padding: 0.75rem 1rem;
  border-radius: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
}

.status-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.pill {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: var(--down-dim);
  color: var(--muted);
}

.pill.ok {
  background: var(--brand-muted);
  color: var(--brand);
}

.pill.warn {
  background: var(--warn-bg);
  color: var(--warn);
  border: 1px solid var(--warn-border);
}

.pill.muted-pill {
  background: transparent;
  border: 1px solid var(--border);
}

.warn-text {
  margin: 0.5rem 0 0;
  color: var(--warn);
  font-size: 0.9rem;
}

.mini-cta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-radius: 12px;
  background: var(--brand-muted);
  border: 1px solid var(--border-strong);
}

.mini-cta p {
  margin: 0.35rem 0 0;
  font-size: 0.85rem;
  line-height: 1.5;
}

.mini-link {
  text-decoration: none;
  white-space: nowrap;
}

.card-inline {
  margin: 0;
}

.dash-section {
  margin: 0;
}

.tier-usage {
  margin-bottom: 1rem;
}
.usage-list {
  list-style: none;
  padding: 0;
  margin: 0.75rem 0 0;
}
.usage-list li {
  display: flex;
  justify-content: space-between;
  padding: 0.4rem 0;
  border-bottom: 1px solid var(--border);
  font-size: 0.9rem;
}
.usage-list li:last-child {
  border-bottom: none;
}
.dash-section h3 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}

.trade-log-head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.trade-log-head h3 {
  margin: 0;
}

.filter-tabs {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}

.filter-tabs button {
  font-size: 0.75rem;
  padding: 0.35rem 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--chart-bg);
  color: var(--muted);
  cursor: pointer;
  font-family: inherit;
}

.filter-tabs button.active {
  background: var(--brand-muted);
  color: var(--brand);
  border-color: var(--brand-dim);
}

.trade-log-card::before {
  content: '';
  display: block;
  height: 3px;
  margin: -1rem -1rem 1rem;
  border-radius: 12px 12px 0 0;
  background: linear-gradient(90deg, var(--brand), var(--brand-deep));
}

.buy { color: var(--brand); font-weight: 600; }
.sell { color: var(--down); font-weight: 600; }
.profit { color: var(--brand); font-weight: 600; }
.loss { color: var(--loss-text); font-weight: 600; }

.status-chip {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: var(--down-dim);
}

.status-chip.open { background: var(--brand-muted); color: var(--brand); }
.status-chip.closed { background: var(--down-dim); color: var(--muted); }

.empty-hint {
  padding: 1.5rem;
  text-align: center;
  line-height: 1.6;
}

.small { font-size: 0.75rem; }
</style>
