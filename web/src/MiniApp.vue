<script setup>
import { ref, computed, onMounted } from 'vue'
import { API, saveTelegramSession, api } from './api'
import { initTelegramWebApp, isTelegramMiniApp, telegramInitData } from './telegramWebApp'
import BrandLogo from './components/BrandLogo.vue'

const loading = ref(true)
const error = ref('')
const trades = ref([])
const positions = ref([])
const payments = ref([])
const subscription = ref(null)
const pricing = ref(null)
const instructions = ref(null)
const pendingOrder = ref(null)
const txRef = ref('')
const paying = ref(false)
const dailyStats = ref(null)
const connectUrl = ref('')

const canTrade = computed(() => subscription.value?.can_trade === true)
const planLabel = computed(() => {
  const sub = subscription.value?.subscription
  if (!sub) return `Trial · ${pricing.value?.trial_days || 5} days`
  return `${sub.plan} · ${sub.status}`
})
const expiresLabel = computed(() => subscription.value?.expires_at || '—')
const daysLeft = computed(() => subscription.value?.days_left ?? '—')

async function authMiniApp() {
  const initData = telegramInitData()
  if (!initData) {
    throw new Error('Open this page from Telegram (📊 Dashboard menu)')
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
  payments.value = data.payments || []
  subscription.value = data.subscription || {}
  pricing.value = data.pricing || {}
  dailyStats.value = data.daily_stats || null
  connectUrl.value = data.connect_url || ''
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

function fmtProfit(t) {
  if (t.profit == null) return '—'
  const n = Number(t.profit)
  return `${n >= 0 ? '+' : ''}$${n.toFixed(2)}`
}

function fmtTime(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

onMounted(async () => {
  if (!isTelegramMiniApp()) {
    loading.value = false
    error.value = 'Open from Telegram bot menu: 📊 Dashboard'
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
    <header class="mini-header">
      <BrandLogo variant="icon" />
      <div>
        <h1>Market Mamba</h1>
        <p class="muted">Trades · Activity · Subscription</p>
      </div>
    </header>

    <p v-if="loading" class="muted center">Loading…</p>
    <p v-else-if="error && !trades.length" class="err banner">{{ error }}</p>

    <template v-else>
      <section class="card sub-card">
        <h2>Subscription</h2>
        <p><strong>{{ planLabel }}</strong></p>
        <p class="muted">Expires: {{ expiresLabel }} · {{ daysLeft }} days left</p>
        <p v-if="!canTrade" class="warn">{{ subscription?.message }}</p>
        <p v-else class="ok">✓ Active access</p>
        <p class="price-line">
          {{ pricing?.trial_days ?? 5 }}-day trial · then
          <strong>{{ pricing?.price_usdt ?? 10 }} USDT</strong> / month (Binance)
        </p>
        <button type="button" class="btn-primary" :disabled="paying" @click="startPayment">
          {{ paying ? 'Please wait…' : 'Subscribe — 10 USDT / month' }}
        </button>
        <a
          v-if="connectUrl"
          class="link-btn broker-link"
          :href="connectUrl"
          target="_blank"
          rel="noopener"
        >Connect broker (Deriv, Exness, Tickmill…)</a>
      </section>

      <section v-if="pendingOrder" class="card pay-card">
        <h2>Payment</h2>
        <p>Reference: <code>{{ pendingOrder.merchant_trade_no }}</code></p>
        <p v-if="instructions?.step1">{{ instructions.step1 }}</p>
        <p v-if="instructions?.step2">{{ instructions.step2 }}</p>
        <p v-if="instructions?.step3">{{ instructions.step3 }}</p>
        <a
          v-if="pendingOrder.checkout_url"
          class="link-btn"
          :href="pendingOrder.checkout_url"
          target="_blank"
          rel="noopener"
        >Open Binance Pay</a>
        <label class="field">
          <span>Transaction ID (after USDT sent)</span>
          <input v-model="txRef" type="text" placeholder="Binance tx hash or order id" />
        </label>
        <button type="button" class="btn-secondary" :disabled="paying" @click="confirmPayment">
          Confirm payment
        </button>
      </section>

      <section class="card stats-card" v-if="dailyStats">
        <h2>Today</h2>
        <div class="stat-row">
          <span>Trades</span><strong>{{ dailyStats.trade_count ?? 0 }}</strong>
          <span>Net P/L</span><strong>{{ dailyStats.net_profit != null ? '$' + Number(dailyStats.net_profit).toFixed(2) : '—' }}</strong>
        </div>
      </section>

      <section class="card">
        <h2>Your open positions ({{ positions.length }})</h2>
        <p v-if="!positions.length" class="muted">No open positions</p>
        <ul v-else class="activity-list">
          <li v-for="p in positions" :key="p.id || p.symbol + p.type">
            <span class="sym">{{ p.symbol }} {{ p.type }}</span>
            <span :class="Number(p.profit) >= 0 ? 'ok' : 'err'">${{ Number(p.profit || 0).toFixed(2) }}</span>
          </li>
        </ul>
      </section>

      <section class="card">
        <h2>Your trades ({{ trades.length }})</h2>
        <p v-if="!trades.length" class="muted">No trades yet</p>
        <ul v-else class="activity-list trades-scroll">
          <li v-for="t in trades" :key="t.id">
            <div class="trade-top">
              <span class="sym">{{ t.symbol }} {{ t.type }}</span>
              <span class="badge" :class="t.status">{{ t.status }}</span>
            </div>
            <div class="trade-meta muted">
              Entry {{ t.entry_price }} · {{ fmtTime(t.created_at) }}
            </div>
            <div class="trade-bottom">
              <span>P/L {{ fmtProfit(t) }}</span>
              <span v-if="t.closure_reason">{{ t.closure_reason }}</span>
            </div>
          </li>
        </ul>
      </section>

      <p v-if="error" class="err small">{{ error }}</p>
    </template>
  </div>
</template>

<style scoped>
.mini-app {
  max-width: 480px;
  margin: 0 auto;
  padding: 1rem;
  padding-bottom: calc(1rem + env(safe-area-inset-bottom));
  min-height: 100vh;
  background: var(--bg, #0a0a0f);
  color: var(--text, #f0f0f5);
  font-family: var(--font-sans, 'Poppins', sans-serif);
}

.mini-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.mini-header h1 {
  margin: 0;
  font-size: 1.25rem;
}

.card {
  background: var(--surface-raised, #14141c);
  border: 1px solid var(--border, #2a2a35);
  border-radius: 14px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.card h2 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}

.price-line {
  font-size: 0.9rem;
  margin: 0.75rem 0;
}

.btn-primary {
  width: 100%;
  min-height: 48px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #00c853, #00a844);
  color: #fff;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
}

.btn-secondary {
  width: 100%;
  margin-top: 0.5rem;
  min-height: 44px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text);
  font-weight: 600;
  cursor: pointer;
}

.link-btn {
  display: block;
  text-align: center;
  margin: 0.75rem 0;
  color: #f0b429;
  font-weight: 600;
}

.field {
  display: block;
  margin-top: 0.75rem;
}

.field input {
  width: 100%;
  margin-top: 0.35rem;
  padding: 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
}

.activity-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.activity-list li {
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--border);
}

.trades-scroll {
  max-height: 320px;
  overflow-y: auto;
}

.trade-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sym {
  font-weight: 700;
  color: var(--win-bright, #00e676);
}

.badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  background: #333;
}

.badge.OPEN {
  background: #1b5e20;
}

.badge.CLOSED {
  background: #37474f;
}

.stat-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.center {
  text-align: center;
}

.banner {
  padding: 1rem;
  border-radius: 10px;
  background: var(--warn-bg, #3d2a00);
}

.small {
  font-size: 0.85rem;
}

.muted {
  color: var(--muted, #9a9aad);
  font-size: 0.85rem;
}

.ok {
  color: #00e676;
}

.err {
  color: #ff5252;
}

.warn {
  color: #ffb74d;
  font-size: 0.9rem;
}
</style>
