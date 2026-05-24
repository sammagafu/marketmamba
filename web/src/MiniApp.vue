<script setup>
import { ref, computed, onMounted } from 'vue'
import { API, saveTelegramSession, api } from './api'
import { initTelegramWebApp, isTelegramMiniApp, telegramInitData } from './telegramWebApp'
import BrandLogo from './components/BrandLogo.vue'
import { VALUE_PROPOSITION, PAYMENT_NOTE } from './brand'

const loading = ref(true)
const error = ref('')
const trades = ref([])
const positions = ref([])
const subscription = ref(null)
const pricing = ref(null)
const instructions = ref(null)
const pendingOrder = ref(null)
const txRef = ref('')
const paying = ref(false)
const dailyStats = ref(null)
const connectUrl = ref('')
const valueProposition = ref('')
const contactUrl = ref('')
const contactLabel = ref('Contact us')
const paymentNote = ref('')

const canTrade = computed(() => subscription.value?.can_trade === true)
const planLabel = computed(() => {
  const sub = subscription.value?.subscription
  if (!sub) return 'Trial'
  return String(sub.plan || 'trial').replace(/^\w/, (c) => c.toUpperCase())
})
const planStatus = computed(() => subscription.value?.subscription?.status || 'trial')
const expiresLabel = computed(() => {
  const exp = subscription.value?.expires_at
  if (!exp) return '—'
  return new Date(exp).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
})
const daysLeft = computed(() => subscription.value?.days_left ?? '—')
const priceUsdt = computed(() => pricing.value?.price_usdt ?? 10)
const trialDays = computed(() => pricing.value?.trial_days ?? 5)
const subscribeLabel = computed(() => `Subscribe · ${priceUsdt.value} USDT / month`)

const netProfit = computed(() => {
  const n = dailyStats.value?.net_profit
  if (n == null) return null
  return Number(n)
})

const closedCount = computed(() => trades.value.filter((t) => t.status === 'CLOSED').length)

async function authMiniApp() {
  const initData = telegramInitData()
  if (!initData) {
    throw new Error('Open this page from the Telegram bot menu (Dashboard).')
  }
  const res = await fetch(`${API}/auth/telegram/webapp`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ init_data: initData }),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  saveTelegramSession(data.session_token, data.telegram_id)
  return data
}

async function loadDashboard() {
  const data = await api('/miniapp/dashboard')
  trades.value = data.trades || []
  positions.value = data.positions || []
  subscription.value = data.subscription || {}
  pricing.value = data.pricing || {}
  dailyStats.value = data.daily_stats || null
  connectUrl.value = data.connect_url || ''
  valueProposition.value = data.value_proposition || VALUE_PROPOSITION
  contactUrl.value = data.contact_us_url || ''
  contactLabel.value = data.contact_us_label || 'Contact us'
  paymentNote.value = data.payment_note || PAYMENT_NOTE
}

async function startPayment() {
  paying.value = true
  error.value = ''
  try {
    const data = await api('/payments/binance/order', { method: 'POST' })
    pendingOrder.value = data.order
    instructions.value = data.instructions || {}
    if (data.order?.checkout_url && window.Telegram?.WebApp?.openLink) {
      window.Telegram.WebApp.openLink(data.order.checkout_url)
    }
  } catch (e) {
    error.value = e.message
  } finally {
    paying.value = false
  }
}

async function confirmPayment() {
  if (!pendingOrder.value?.id || !txRef.value.trim()) {
    error.value = 'Enter your Binance transaction ID'
    return
  }
  paying.value = true
  error.value = ''
  try {
    const data = await api('/payments/binance/confirm', {
      method: 'POST',
      body: { order_id: pendingOrder.value.id, tx_reference: txRef.value.trim() },
    })
    pendingOrder.value = data.order
    subscription.value = data.subscription
    await loadDashboard()
  } catch (e) {
    error.value = e.message
  } finally {
    paying.value = false
  }
}

function openLink(url) {
  if (!url) return
  if (window.Telegram?.WebApp?.openLink) {
    window.Telegram.WebApp.openLink(url)
  } else {
    window.open(url, '_blank', 'noopener')
  }
}

function fmtProfit(t) {
  if (t.profit == null) return '—'
  const n = Number(t.profit)
  return `${n >= 0 ? '+' : ''}$${n.toFixed(2)}`
}

function fmtTime(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function plClass(n) {
  if (n == null) return ''
  return Number(n) >= 0 ? 'pos' : 'neg'
}

onMounted(async () => {
  if (!isTelegramMiniApp()) {
    loading.value = false
    error.value = 'Open from the Telegram bot menu: Dashboard'
    return
  }
  initTelegramWebApp()
  try {
    await authMiniApp()
    await loadDashboard()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="mini-app">
    <header class="corp-header">
      <div class="corp-header-main">
        <BrandLogo variant="icon" class="corp-logo" />
        <div class="corp-brand">
          <span class="corp-eyebrow">Client portal</span>
          <h1 class="corp-title">Market Mamba</h1>
        </div>
      </div>
      <span v-if="!loading" class="status-pill" :class="canTrade ? 'status-active' : 'status-pending'">
        {{ canTrade ? 'Active' : 'Limited' }}
      </span>
    </header>

    <div v-if="loading" class="state-panel">
      <div class="spinner" aria-hidden="true" />
      <p class="state-title">Loading your account</p>
      <p class="state-sub">Secure session via Telegram</p>
    </div>

    <div v-else-if="error && !trades.length && !positions.length" class="state-panel state-error">
      <p class="state-title">Unable to load</p>
      <p class="state-sub">{{ error }}</p>
    </div>

    <main v-else class="corp-main">
      <p class="corp-tagline">{{ valueProposition }}</p>

      <!-- Subscription -->
      <section class="corp-card corp-card-highlight">
        <div class="card-head">
          <h2 class="card-title">Membership</h2>
          <span class="plan-chip">{{ planLabel }}</span>
        </div>

        <div class="metric-grid">
          <div class="metric">
            <span class="metric-label">Status</span>
            <span class="metric-value">{{ planStatus }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Renews / ends</span>
            <span class="metric-value">{{ expiresLabel }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Days remaining</span>
            <span class="metric-value">{{ daysLeft }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Billing</span>
            <span class="metric-value">{{ priceUsdt }} USDT/mo</span>
          </div>
        </div>

        <p v-if="!canTrade" class="alert alert-warn">{{ subscription?.message }}</p>
        <p v-else class="alert alert-ok">Account in good standing — trading enabled</p>

        <div class="pricing-block">
          <p class="pricing-lead">
            {{ trialDays }}-day evaluation period, then
            <strong>{{ priceUsdt }} USDT</strong> per month via Binance.
          </p>
          <p class="pricing-note">{{ paymentNote }}</p>
        </div>

        <div class="btn-stack">
          <button type="button" class="btn btn-primary" :disabled="paying" @click="startPayment">
            {{ paying ? 'Processing…' : subscribeLabel }}
          </button>
          <button
            v-if="connectUrl"
            type="button"
            class="btn btn-secondary"
            @click="openLink(connectUrl)"
          >
            Connect broker account
          </button>
          <button
            v-if="contactUrl"
            type="button"
            class="btn btn-ghost"
            @click="openLink(contactUrl)"
          >
            {{ contactLabel }} · Enterprise &amp; Pro
          </button>
        </div>
      </section>

      <!-- Payment in progress -->
      <section v-if="pendingOrder" class="corp-card">
        <div class="card-head">
          <h2 class="card-title">Payment confirmation</h2>
          <span class="plan-chip plan-chip-warn">Pending</span>
        </div>
        <dl class="detail-list">
          <div class="detail-row">
            <dt>Reference</dt>
            <dd><code>{{ pendingOrder.merchant_trade_no }}</code></dd>
          </div>
        </dl>
        <ol v-if="instructions?.step1" class="steps-list">
          <li v-if="instructions.step1">{{ instructions.step1 }}</li>
          <li v-if="instructions.step2">{{ instructions.step2 }}</li>
          <li v-if="instructions.step3">{{ instructions.step3 }}</li>
        </ol>
        <button
          v-if="pendingOrder.checkout_url"
          type="button"
          class="btn btn-secondary"
          @click="openLink(pendingOrder.checkout_url)"
        >
          Open Binance Pay
        </button>
        <label class="field">
          <span class="field-label">Transaction ID</span>
          <span class="field-hint">After sending USDT on Binance</span>
          <input
            v-model="txRef"
            type="text"
            class="field-input"
            placeholder="Tx hash or order ID"
            autocomplete="off"
          />
        </label>
        <button type="button" class="btn btn-primary" :disabled="paying" @click="confirmPayment">
          Confirm payment
        </button>
      </section>

      <!-- Performance snapshot -->
      <section class="corp-card">
        <div class="card-head">
          <h2 class="card-title">Today&apos;s performance</h2>
        </div>
        <div class="metric-grid metric-grid-3">
          <div class="metric">
            <span class="metric-label">Trades</span>
            <span class="metric-value">{{ dailyStats?.trade_count ?? 0 }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Net P/L</span>
            <span class="metric-value" :class="plClass(netProfit)">
              {{ netProfit != null ? (netProfit >= 0 ? '+' : '') + '$' + netProfit.toFixed(2) : '—' }}
            </span>
          </div>
          <div class="metric">
            <span class="metric-label">Closed (all time)</span>
            <span class="metric-value">{{ closedCount }}</span>
          </div>
        </div>
      </section>

      <!-- Positions -->
      <section class="corp-card">
        <div class="card-head">
          <h2 class="card-title">Open positions</h2>
          <span class="count-badge">{{ positions.length }}</span>
        </div>
        <div v-if="!positions.length" class="empty-state">
          <p>No open positions on your connected account.</p>
        </div>
        <div v-else class="data-table">
          <div class="data-row data-head">
            <span>Instrument</span>
            <span>Side</span>
            <span class="align-right">P/L</span>
          </div>
          <div v-for="p in positions" :key="p.id || p.symbol + p.type" class="data-row">
            <span class="cell-primary">{{ p.symbol }}</span>
            <span class="side-tag" :class="p.type === 'BUY' ? 'buy' : 'sell'">{{ p.type }}</span>
            <span class="align-right" :class="plClass(p.profit)">
              ${{ Number(p.profit || 0).toFixed(2) }}
            </span>
          </div>
        </div>
      </section>

      <!-- Trade history -->
      <section class="corp-card">
        <div class="card-head">
          <h2 class="card-title">Trade history</h2>
          <span class="count-badge">{{ trades.length }}</span>
        </div>
        <div v-if="!trades.length" class="empty-state">
          <p>No trades recorded yet. Connect a broker and enable automation in the bot.</p>
        </div>
        <ul v-else class="trade-list">
          <li v-for="t in trades" :key="t.id" class="trade-item">
            <div class="trade-item-top">
              <div>
                <span class="cell-primary">{{ t.symbol }}</span>
                <span class="side-tag" :class="t.type === 'BUY' ? 'buy' : 'sell'">{{ t.type }}</span>
              </div>
              <span class="status-tag" :class="t.status?.toLowerCase()">{{ t.status }}</span>
            </div>
            <div class="trade-item-meta">
              <span>Entry {{ Number(t.entry_price).toFixed(5) }}</span>
              <span>{{ fmtTime(t.created_at) }}</span>
            </div>
            <div class="trade-item-foot">
              <span :class="plClass(t.profit)">P/L {{ fmtProfit(t) }}</span>
              <span v-if="t.closure_reason" class="closure">{{ t.closure_reason }}</span>
            </div>
          </li>
        </ul>
      </section>

      <p v-if="error" class="inline-error">{{ error }}</p>
    </main>

    <footer class="corp-footer">
      <p>Market Mamba · Controlled automation · Not a broker</p>
      <p class="corp-footer-sub">Forex trading involves substantial risk. USDT billing via Binance only.</p>
    </footer>
  </div>
</template>

<style scoped>
.mini-app {
  --corp-bg: #0c1117;
  --corp-surface: #151b26;
  --corp-surface-2: #1c2433;
  --corp-border: #2a3548;
  --corp-border-light: #3d4d66;
  --corp-text: #f1f5f9;
  --corp-text-soft: #cbd5e1;
  --corp-muted: #94a3b8;
  --corp-accent: #10b981;
  --corp-accent-dim: #059669;
  --corp-accent-soft: rgba(16, 185, 129, 0.12);
  --corp-warn: #f59e0b;
  --corp-warn-soft: rgba(245, 158, 11, 0.12);
  --corp-danger: #f87171;
  --corp-danger-soft: rgba(248, 113, 113, 0.1);

  max-width: 520px;
  margin: 0 auto;
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--corp-bg);
  color: var(--corp-text);
  font-family: 'Inter', 'Segoe UI', system-ui, -apple-system, sans-serif;
  font-size: 15px;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

.corp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.25rem;
  padding-top: max(1rem, env(safe-area-inset-top));
  background: var(--corp-surface);
  border-bottom: 1px solid var(--corp-border);
}

.corp-header-main {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.corp-logo {
  width: 40px !important;
  height: 40px !important;
  flex-shrink: 0;
}

.corp-eyebrow {
  display: block;
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--corp-muted);
}

.corp-title {
  margin: 0.1rem 0 0;
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.status-pill {
  flex-shrink: 0;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--corp-border);
}

.status-active {
  color: var(--corp-accent);
  background: var(--corp-accent-soft);
  border-color: rgba(16, 185, 129, 0.35);
}

.status-pending {
  color: var(--corp-warn);
  background: var(--corp-warn-soft);
  border-color: rgba(245, 158, 11, 0.35);
}

.corp-main {
  flex: 1;
  padding: 1.25rem;
  padding-bottom: 0.5rem;
}

.corp-tagline {
  margin: 0 0 1.25rem;
  font-size: 0.875rem;
  line-height: 1.55;
  color: var(--corp-text-soft);
}

.corp-card {
  background: var(--corp-surface);
  border: 1px solid var(--corp-border);
  border-radius: 12px;
  padding: 1.15rem 1.25rem;
  margin-bottom: 1rem;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset;
}

.corp-card-highlight {
  border-color: var(--corp-border-light);
  background: linear-gradient(180deg, var(--corp-surface-2) 0%, var(--corp-surface) 100%);
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.card-title {
  margin: 0;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--corp-muted);
}

.plan-chip {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  background: var(--corp-accent-soft);
  color: var(--corp-accent);
  border: 1px solid rgba(16, 185, 129, 0.25);
}

.plan-chip-warn {
  background: var(--corp-warn-soft);
  color: var(--corp-warn);
  border-color: rgba(245, 158, 11, 0.3);
}

.metric-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem 1rem;
  margin-bottom: 1rem;
}

.metric-grid-3 {
  grid-template-columns: repeat(3, 1fr);
  margin-bottom: 0;
}

@media (max-width: 380px) {
  .metric-grid-3 {
    grid-template-columns: 1fr;
  }
}

.metric-label {
  display: block;
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--corp-muted);
  margin-bottom: 0.2rem;
}

.metric-value {
  font-size: 0.95rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--corp-text);
}

.alert {
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  border-radius: 8px;
  font-size: 0.85rem;
  line-height: 1.45;
}

.alert-ok {
  background: var(--corp-accent-soft);
  color: #6ee7b7;
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.alert-warn {
  background: var(--corp-warn-soft);
  color: #fcd34d;
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.pricing-block {
  padding: 0.85rem 0;
  border-top: 1px solid var(--corp-border);
  border-bottom: 1px solid var(--corp-border);
  margin-bottom: 1rem;
}

.pricing-lead {
  margin: 0 0 0.35rem;
  font-size: 0.875rem;
  color: var(--corp-text-soft);
}

.pricing-note {
  margin: 0;
  font-size: 0.75rem;
  color: var(--corp-muted);
}

.btn-stack {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.btn {
  width: 100%;
  min-height: 48px;
  padding: 0.65rem 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, opacity 0.15s;
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.btn-primary {
  border: none;
  background: var(--corp-accent);
  color: #042f1e;
}

.btn-primary:not(:disabled):active {
  background: var(--corp-accent-dim);
}

.btn-secondary {
  border: 1px solid var(--corp-border-light);
  background: transparent;
  color: var(--corp-text);
}

.btn-ghost {
  border: none;
  background: transparent;
  color: var(--corp-muted);
  min-height: 40px;
  font-weight: 500;
}

.detail-list {
  margin: 0 0 1rem;
}

.detail-row {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.5rem 1rem;
  font-size: 0.85rem;
}

.detail-row dt {
  color: var(--corp-muted);
  font-weight: 500;
}

.detail-row dd {
  margin: 0;
  text-align: right;
}

.detail-row code {
  font-size: 0.8rem;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  background: var(--corp-bg);
  border: 1px solid var(--corp-border);
}

.steps-list {
  margin: 0 0 1rem;
  padding-left: 1.2rem;
  font-size: 0.85rem;
  color: var(--corp-text-soft);
  line-height: 1.55;
}

.field {
  display: block;
  margin-bottom: 0.75rem;
}

.field-label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--corp-text-soft);
}

.field-hint {
  display: block;
  font-size: 0.72rem;
  color: var(--corp-muted);
  margin-bottom: 0.35rem;
}

.field-input {
  width: 100%;
  padding: 0.7rem 0.85rem;
  border-radius: 8px;
  border: 1px solid var(--corp-border);
  background: var(--corp-bg);
  color: var(--corp-text);
  font-size: 0.9rem;
  font-family: inherit;
}

.field-input:focus {
  outline: none;
  border-color: var(--corp-accent);
  box-shadow: 0 0 0 2px var(--corp-accent-soft);
}

.count-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
  background: var(--corp-bg);
  border: 1px solid var(--corp-border);
  color: var(--corp-muted);
  font-variant-numeric: tabular-nums;
}

.empty-state {
  padding: 1.25rem 0.5rem;
  text-align: center;
  font-size: 0.85rem;
  color: var(--corp-muted);
}

.empty-state p {
  margin: 0;
}

.data-table {
  border: 1px solid var(--corp-border);
  border-radius: 8px;
  overflow: hidden;
}

.data-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 0.5rem 0.75rem;
  align-items: center;
  padding: 0.65rem 0.85rem;
  font-size: 0.85rem;
  border-bottom: 1px solid var(--corp-border);
}

.data-row:last-child {
  border-bottom: none;
}

.data-head {
  background: var(--corp-bg);
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--corp-muted);
}

.cell-primary {
  font-weight: 600;
  color: var(--corp-text);
}

.side-tag {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  padding: 0.12rem 0.4rem;
  border-radius: 4px;
}

.side-tag.buy {
  color: #6ee7b7;
  background: var(--corp-accent-soft);
}

.side-tag.sell {
  color: var(--corp-muted);
  background: var(--corp-bg);
  border: 1px solid var(--corp-border);
}

.align-right {
  text-align: right;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.pos {
  color: #6ee7b7;
}

.neg {
  color: var(--corp-danger);
}

.trade-list {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: 360px;
  overflow-y: auto;
}

.trade-item {
  padding: 0.85rem 0;
  border-bottom: 1px solid var(--corp-border);
}

.trade-item:last-child {
  border-bottom: none;
}

.trade-item-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
}

.trade-item-top > div {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.status-tag {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: var(--corp-bg);
  border: 1px solid var(--corp-border);
  color: var(--corp-muted);
}

.status-tag.open {
  color: #6ee7b7;
  border-color: rgba(16, 185, 129, 0.3);
}

.trade-item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 1rem;
  font-size: 0.78rem;
  color: var(--corp-muted);
  margin-bottom: 0.35rem;
}

.trade-item-foot {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 0.25rem;
  font-size: 0.85rem;
  font-weight: 600;
}

.closure {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--corp-muted);
}

.state-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1.5rem;
  text-align: center;
}

.state-title {
  margin: 1rem 0 0.35rem;
  font-size: 1rem;
  font-weight: 600;
}

.state-sub {
  margin: 0;
  font-size: 0.875rem;
  color: var(--corp-muted);
  max-width: 280px;
}

.state-error .state-sub {
  color: var(--corp-danger);
}

.spinner {
  width: 36px;
  height: 36px;
  border: 2px solid var(--corp-border);
  border-top-color: var(--corp-accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.inline-error {
  margin: 0;
  padding: 0.75rem;
  font-size: 0.85rem;
  color: var(--corp-danger);
  background: var(--corp-danger-soft);
  border-radius: 8px;
  border: 1px solid rgba(248, 113, 113, 0.25);
}

.corp-footer {
  padding: 1.25rem;
  padding-bottom: max(1.25rem, env(safe-area-inset-bottom));
  border-top: 1px solid var(--corp-border);
  text-align: center;
}

.corp-footer p {
  margin: 0;
  font-size: 0.72rem;
  color: var(--corp-muted);
  letter-spacing: 0.02em;
}

.corp-footer-sub {
  margin-top: 0.35rem !important;
  font-size: 0.68rem !important;
  opacity: 0.85;
}
</style>
