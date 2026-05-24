<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { API, saveTelegramSession, api } from './api'
import {
  initTelegramWebApp,
  isTelegramMiniApp,
  telegramInitData,
  hapticLight,
  hapticSuccess,
  showMainButton,
  hideMainButton,
} from './telegramWebApp'
import BrandLogo from './components/BrandLogo.vue'
import { VALUE_PROPOSITION, PAYMENT_NOTE, TAGLINE } from './brand'

const SPLASH_MIN_MS = 750
const SPLASH_FADE_MS = 380

const loading = ref(true)
const splashVisible = ref(true)
const splashFading = ref(false)
const refreshing = ref(false)
const error = ref('')
const trades = ref([])
const positions = ref([])
const subscription = ref(null)
const pricing = ref(null)
const packages = ref([])
const instructions = ref(null)
const pendingOrder = ref(null)
const txRef = ref('')
const paying = ref(false)
const dailyStats = ref(null)
const connectUrl = ref('')
const botUsername = ref('market_mamba_bot')
const valueProposition = ref('')
const contactUrl = ref('')
const contactLabel = ref('Contact us')
const paymentNote = ref('')
const signalTypes = ref({ forex: true, indexes: true, crypto: true })
const showAllTrades = ref(false)
const communityMessage = ref('')
const aiTrainingNote = ref('')
const assetPhaseUnlocked = ref(true)
const displayName = ref('')

const canTrade = computed(() => subscription.value?.can_trade === true)
const welcomeLabel = computed(() => {
  if (displayName.value) return `Welcome, ${displayName.value}`
  return 'Your dashboard'
})
const openPositionsCount = computed(() => positions.value.length)
const tierInfo = computed(() => subscription.value?.tier || null)

const currentPlanId = computed(() => {
  const plan =
    subscription.value?.subscription?.plan || tierInfo.value?.limits?.plan || 'trial'
  return String(plan).toLowerCase()
})

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
const subscribeLabel = computed(() => `Subscribe · ${priceUsdt.value} USDT`)

const netProfit = computed(() => {
  const n = dailyStats.value?.net_profit
  if (n == null) return null
  return Number(n)
})

const closedCount = computed(() => trades.value.filter((t) => t.status === 'CLOSED').length)
const visibleTrades = computed(() =>
  showAllTrades.value ? trades.value : trades.value.slice(0, 5),
)

const activeSignalLabels = computed(() => {
  const labels = []
  if (signalTypes.value.forex) labels.push('Forex')
  if (signalTypes.value.indexes) labels.push('Indexes')
  if (signalTypes.value.crypto) labels.push('Crypto')
  return labels
})

const botUrl = computed(() => `https://t.me/${botUsername.value}`)

function usagePct(used, max) {
  if (max == null || max <= 0) return 0
  return Math.min(100, Math.round((Number(used) / Number(max)) * 100))
}

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
  const u = data.user
  displayName.value =
    [u?.first_name, u?.last_name].filter(Boolean).join(' ') ||
    u?.username ||
    'Trader'
  return data
}

async function loadPairs() {
  try {
    const data = await api('/trading-pairs')
    if (data.signal_types) {
      signalTypes.value = {
        forex: !!data.signal_types.forex,
        indexes: !!data.signal_types.indexes,
        crypto: !!data.signal_types.crypto,
      }
    }
  } catch {
    /* optional */
  }
}

async function loadDashboard() {
  const data = await api('/miniapp/dashboard')
  trades.value = data.trades || []
  positions.value = data.positions || []
  subscription.value = data.subscription || {}
  pricing.value = data.pricing || {}
  packages.value = data.packages || []
  dailyStats.value = data.daily_stats || null
  connectUrl.value = data.connect_url || ''
  botUsername.value = data.telegram_bot_username || 'market_mamba_bot'
  valueProposition.value = data.value_proposition || VALUE_PROPOSITION
  contactUrl.value = data.contact_us_url || ''
  contactLabel.value = data.contact_us_label || 'Contact us'
  paymentNote.value = data.payment_note || PAYMENT_NOTE
  communityMessage.value = data.community_phase_message || ''
  aiTrainingNote.value = data.ai_training_note || ''
  assetPhaseUnlocked.value = data.asset_phase_unlocked !== false
}

async function refresh() {
  refreshing.value = true
  error.value = ''
  try {
    await Promise.all([loadDashboard(), loadPairs()])
    hapticLight()
  } catch (e) {
    error.value = e.message
  } finally {
    refreshing.value = false
  }
}

async function startPayment() {
  hapticLight()
  paying.value = true
  error.value = ''
  try {
    const data = await api('/payments/binance/order', { method: 'POST' })
    pendingOrder.value = data.order
    instructions.value = data.instructions || {}
    hideMainButton()
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
  hapticLight()
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
    await loadPairs()
    hapticSuccess()
  } catch (e) {
    error.value = e.message
  } finally {
    paying.value = false
  }
}

function openLink(url) {
  if (!url) return
  hapticLight()
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

function isCurrentPlan(pkg) {
  return pkg.id === currentPlanId.value
}

function onPackageAction(pkg) {
  if (pkg.contact_only) {
    openLink(contactUrl.value || botUrl.value)
    return
  }
  if (pkg.id === 'monthly' && !isCurrentPlan(pkg)) {
    startPayment()
  }
}

function syncMainButton() {
  if (loading.value || pendingOrder.value) {
    hideMainButton()
    return
  }
  if (!canTrade.value) {
    showMainButton(subscribeLabel.value, startPayment)
  } else {
    hideMainButton()
  }
}

watch([canTrade, pendingOrder, loading], syncMainButton)

function hideSplash() {
  return new Promise((resolve) => {
    splashFading.value = true
    setTimeout(() => {
      splashVisible.value = false
      resolve()
    }, SPLASH_FADE_MS)
  })
}

async function dismissSplash(startedAt) {
  const remain = SPLASH_MIN_MS - (Date.now() - startedAt)
  if (remain > 0) {
    await new Promise((r) => setTimeout(r, remain))
  }
  await hideSplash()
}

onMounted(async () => {
  const splashStartedAt = Date.now()
  if (!isTelegramMiniApp()) {
    loading.value = false
    error.value = 'Open from the Telegram bot menu: Dashboard'
    await dismissSplash(splashStartedAt)
    return
  }
  initTelegramWebApp()
  try {
    await authMiniApp()
    await Promise.all([loadDashboard(), loadPairs()])
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
    await dismissSplash(splashStartedAt)
    syncMainButton()
  }
})
</script>

<template>
  <div class="mini-app">
    <Transition name="splash">
      <div
        v-if="splashVisible"
        class="tg-splash"
        :class="{ 'tg-splash--fade': splashFading }"
        role="status"
        aria-live="polite"
        aria-busy="true"
        aria-label="Loading Market Mamba"
      >
        <div class="tg-splash-inner">
          <BrandLogo variant="portrait" class="splash-logo" />
          <h1 class="splash-title">Market Mamba</h1>
          <p class="splash-tagline">{{ TAGLINE }}</p>
          <div class="splash-loader" aria-hidden="true">
            <span class="splash-dot" />
            <span class="splash-dot" />
            <span class="splash-dot" />
          </div>
        </div>
      </div>
    </Transition>

    <header v-show="!splashVisible" class="tg-header">
      <div class="tg-header-main">
        <BrandLogo variant="icon" class="tg-logo" />
        <div class="tg-header-text">
          <span class="tg-eyebrow">{{ TAGLINE }}</span>
          <h1 class="tg-title">{{ welcomeLabel }}</h1>
        </div>
      </div>
      <div class="tg-header-actions">
        <button
          v-if="!loading"
          type="button"
          class="icon-btn"
          :disabled="refreshing"
          aria-label="Refresh dashboard"
          @click="refresh"
        >
          ↻
        </button>
        <span class="status-pill" :class="canTrade ? 'on' : 'off'">
          {{ canTrade ? 'Active' : 'Limited' }}
        </span>
      </div>
    </header>

    <div
      v-if="!splashVisible && error && !trades.length && !positions.length"
      class="state-panel state-error"
    >
      <p class="state-title">Could not load</p>
      <p class="state-sub">{{ error }}</p>
      <button type="button" class="btn btn-primary" @click="refresh">Try again</button>
    </div>

    <main v-else-if="!splashVisible" class="tg-main dash">
      <div
        v-if="!assetPhaseUnlocked && communityMessage"
        class="community-banner"
        role="status"
      >
        <p class="community-banner-body">{{ communityMessage }}</p>
        <p v-if="aiTrainingNote" class="community-banner-ai">{{ aiTrainingNote }}</p>
      </div>

      <section class="dash-actions" aria-label="Quick actions">
        <button
          v-if="connectUrl"
          type="button"
          class="action-tile action-tile--primary"
          @click="openLink(connectUrl)"
        >
          <span class="action-icon" aria-hidden="true">⚡</span>
          <span class="action-title">Connect broker</span>
          <span class="action-sub">Link MT4/MT5 for live execution</span>
        </button>
        <button type="button" class="action-tile" @click="openLink(botUrl)">
          <span class="action-icon" aria-hidden="true">💬</span>
          <span class="action-title">Open bot</span>
          <span class="action-sub">Signals &amp; /autostart</span>
        </button>
        <button
          v-if="contactUrl"
          type="button"
          class="action-tile"
          @click="openLink(contactUrl)"
        >
          <span class="action-icon" aria-hidden="true">✉</span>
          <span class="action-title">{{ contactLabel }}</span>
          <span class="action-sub">Support &amp; Pro plans</span>
        </button>
        <button
          v-if="!canTrade && !pendingOrder"
          type="button"
          class="action-tile action-tile--accent"
          :disabled="paying"
          @click="startPayment"
        >
          <span class="action-icon" aria-hidden="true">₮</span>
          <span class="action-title">Subscribe</span>
          <span class="action-sub">{{ priceUsdt }} USDT / month</span>
        </button>
      </section>

      <section class="dash-bento" aria-label="Overview">
        <article class="bento-card bento-card--hero">
          <span class="bento-k">Net P/L today</span>
          <span class="bento-v" :class="plClass(netProfit)">
            {{ netProfit != null ? (netProfit >= 0 ? '+' : '') + '$' + netProfit.toFixed(2) : '—' }}
          </span>
        </article>
        <article class="bento-card">
          <span class="bento-k">Trades today</span>
          <span class="bento-v">{{ dailyStats?.trade_count ?? 0 }}</span>
        </article>
        <article class="bento-card">
          <span class="bento-k">Open positions</span>
          <span class="bento-v">{{ openPositionsCount }}</span>
        </article>
        <article class="bento-card">
          <span class="bento-k">Plan</span>
          <span class="bento-v bento-v-sm">{{ planLabel }}</span>
          <span class="bento-meta">{{ daysLeft }} days left</span>
        </article>
      </section>

      <section class="tg-card tg-card-accent dash-membership">
        <div class="card-head">
          <h2 class="card-label">Membership</h2>
          <span class="chip">{{ planLabel }}</span>
        </div>
        <div class="membership-strip">
          <div class="membership-item">
            <span class="stat-k">Status</span>
            <span class="stat-v">{{ planStatus }}</span>
          </div>
          <div class="membership-item">
            <span class="stat-k">Renews</span>
            <span class="stat-v stat-v-sm">{{ expiresLabel }}</span>
          </div>
          <div class="membership-item">
            <span class="stat-k">Signals</span>
            <span class="stat-v stat-v-sm">{{ activeSignalLabels.join(' · ') || '—' }}</span>
          </div>
        </div>
        <p v-if="!canTrade" class="banner banner-warn">{{ subscription?.message }}</p>
        <p v-else class="banner banner-ok">Trading enabled on your plan</p>
      </section>

      <section v-if="pendingOrder" class="tg-card dash-panel-wide">
        <div class="card-head">
          <h2 class="card-label">Complete payment</h2>
          <span class="chip chip-warn">Pending</span>
        </div>
        <p class="ref-line">
          Ref <code>{{ pendingOrder.merchant_trade_no }}</code>
        </p>
        <ol v-if="instructions?.step1" class="steps">
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
          <input
            v-model="txRef"
            type="text"
            class="field-input"
            placeholder="Paste after sending USDT"
            autocomplete="off"
          />
        </label>
        <button type="button" class="btn btn-primary" :disabled="paying" @click="confirmPayment">
          Confirm payment
        </button>
      </section>

      <div class="dash-split">
      <!-- Positions -->
      <section class="tg-card dash-panel">
        <div class="card-head">
          <h2 class="card-label">Open positions</h2>
          <span class="count">{{ openPositionsCount }}</span>
        </div>
        <p v-if="!positions.length" class="empty">No open positions.</p>
        <ul v-else class="list-cards">
          <li v-for="p in positions" :key="p.id || p.symbol + p.type" class="list-card">
            <div class="list-card-main">
              <strong class="list-card-title">{{ p.symbol }}</strong>
              <span class="side" :class="p.type === 'BUY' ? 'buy' : 'sell'">{{ p.type }}</span>
            </div>
            <span class="list-card-value" :class="plClass(p.profit)">
              ${{ Number(p.profit || 0).toFixed(2) }}
            </span>
          </li>
        </ul>
      </section>

      <!-- Trades -->
      <section class="tg-card dash-panel">
        <div class="card-head">
          <h2 class="card-label">Recent trades</h2>
          <span class="count">{{ trades.length }}</span>
        </div>
        <p v-if="!trades.length" class="empty">No trades yet. Connect a broker and use /autostart in the bot.</p>
        <ul v-else class="list-cards">
          <li v-for="t in visibleTrades" :key="t.id" class="list-card list-card--stack">
            <div class="list-card-top">
              <div class="list-card-main">
                <strong class="list-card-title">{{ t.symbol }}</strong>
                <span class="side" :class="t.type === 'BUY' ? 'buy' : 'sell'">{{ t.type }}</span>
              </div>
              <span class="status-badge">{{ t.status }}</span>
            </div>
            <p class="list-card-meta">{{ fmtTime(t.created_at) }} · {{ Number(t.entry_price).toFixed(5) }}</p>
            <p class="list-card-pl" :class="plClass(t.profit)">P/L {{ fmtProfit(t) }}</p>
          </li>
        </ul>
        <button
          v-if="trades.length > 5"
          type="button"
          class="btn-text"
          @click="showAllTrades = !showAllTrades"
        >
          {{ showAllTrades ? 'Show less' : `Show all ${trades.length}` }}
        </button>
      </section>
      </div>

      <section v-if="tierInfo" class="tg-card dash-panel-wide">
        <div class="card-head">
          <h2 class="card-label">Plan usage</h2>
          <span class="chip chip-muted">{{ tierInfo.limits?.plan }}</span>
        </div>
        <ul class="usage-bars">
          <li>
            <div class="usage-label">
              <span>Signals</span>
              <span>{{ tierInfo.usage?.signals_received ?? 0 }} / {{ tierInfo.limits?.max_signals_per_period }}</span>
            </div>
            <div class="bar-track">
              <div
                class="bar-fill"
                :style="{ width: usagePct(tierInfo.usage?.signals_received, tierInfo.limits?.max_signals_per_period) + '%' }"
              />
            </div>
          </li>
          <li>
            <div class="usage-label">
              <span>Long trades</span>
              <span>{{ tierInfo.usage?.long_trades ?? 0 }} / {{ tierInfo.limits?.max_long_trades }}</span>
            </div>
            <div class="bar-track">
              <div
                class="bar-fill"
                :style="{ width: usagePct(tierInfo.usage?.long_trades, tierInfo.limits?.max_long_trades) + '%' }"
              />
            </div>
          </li>
          <li>
            <div class="usage-label">
              <span>Short trades</span>
              <span>{{ tierInfo.usage?.short_trades ?? 0 }} / {{ tierInfo.limits?.max_short_trades }}</span>
            </div>
            <div class="bar-track">
              <div
                class="bar-fill"
                :style="{ width: usagePct(tierInfo.usage?.short_trades, tierInfo.limits?.max_short_trades) + '%' }"
              />
            </div>
          </li>
        </ul>
      </section>

      <details v-if="packages.length" class="dash-details tg-card">
        <summary class="dash-details-summary">
          <span class="card-label">Plans &amp; pricing</span>
          <span class="dash-details-chevron" aria-hidden="true">›</span>
        </summary>
        <div class="dash-details-body">
        <p class="billing-note packages-note">{{ paymentNote }}</p>
        <ul class="package-list">
          <li
            v-for="pkg in packages"
            :key="pkg.id"
            class="package-card"
            :class="{
              current: isCurrentPlan(pkg),
              recommended: pkg.recommended,
            }"
          >
            <div class="package-head">
              <div>
                <h3 class="package-name">{{ pkg.name }}</h3>
                <p class="package-desc">{{ pkg.description }}</p>
              </div>
              <span v-if="isCurrentPlan(pkg)" class="package-badge">Current</span>
              <span v-else-if="pkg.recommended" class="package-badge package-badge-rec">Popular</span>
            </div>
            <p class="package-price">{{ pkg.price_label }}</p>
            <ul class="package-features">
              <li v-for="(feat, i) in pkg.features" :key="i">{{ feat }}</li>
            </ul>
            <button
              v-if="pkg.contact_only"
              type="button"
              class="btn btn-secondary package-btn"
              @click="onPackageAction(pkg)"
            >
              {{ contactLabel }}
            </button>
            <button
              v-else-if="pkg.id === 'monthly' && !isCurrentPlan(pkg) && !pendingOrder"
              type="button"
              class="btn btn-primary package-btn"
              :disabled="paying"
              @click="onPackageAction(pkg)"
            >
              {{ paying ? 'Please wait…' : subscribeLabel }}
            </button>
          </li>
        </ul>
        </div>
      </details>

      <p class="tg-lead">{{ valueProposition }}</p>

      <!-- Bot tips -->
      <section class="tg-card tg-card-dim">
        <h2 class="card-label solo">In Telegram</h2>
        <ul class="cmd-list">
          <li><code>/signaltypes</code> — forex, indexes, crypto</li>
          <li><code>/pairs</code> — choose symbols</li>
          <li><code>/autostart</code> — enable automation</li>
          <li><code>/balance</code> — account balance</li>
        </ul>
      </section>

      <p v-if="error" class="inline-err">{{ error }}</p>
    </main>

    <footer v-show="!splashVisible" class="tg-footer">
      <p>Not a broker · USDT via Binance only</p>
      <p class="tg-footer-sub">Forex trading involves substantial risk.</p>
    </footer>
  </div>
</template>

<style scoped>
.mini-app {
  --mm-bg: #050505;
  --mm-surface: #0c0c0c;
  --mm-raised: #141414;
  --mm-border: #232323;
  --mm-text: #f3f4f6;
  --mm-muted: #9ca3af;
  --mm-brand: #b8ff3d;
  --mm-brand-soft: rgba(184, 255, 61, 0.14);
  --mm-on-brand: #0a1404;
  --mm-warn: #e5b84a;
  --mm-warn-soft: rgba(229, 184, 74, 0.12);
  --mm-loss: #f87171;
  --dash-radius: 16px;
  --dash-pad: clamp(0.75rem, 3.5vw, 1.15rem);
  --dash-max: 520px;
  --dash-gap: 0.75rem;

  max-width: 100%;
  margin: 0 auto;
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--mm-bg);
  color: var(--mm-text);
  font-family: 'Poppins', system-ui, -apple-system, sans-serif;
  font-size: 15px;
  line-height: 1.45;
  -webkit-font-smoothing: antialiased;
}

.tg-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  padding-top: max(0.85rem, env(safe-area-inset-top));
  background: var(--mm-surface);
  border-bottom: 1px solid var(--mm-border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.tg-header-main {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
  flex: 1;
}

.tg-header-text {
  min-width: 0;
}

.tg-header-text .tg-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: min(52vw, 14rem);
}

@media (min-width: 380px) {
  .tg-header-text .tg-title {
    max-width: 16rem;
  }
}

.tg-logo {
  width: 36px !important;
  height: 36px !important;
  flex-shrink: 0;
}

.tg-eyebrow {
  display: block;
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--mm-muted);
}

.tg-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.tg-header-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-shrink: 0;
}

.icon-btn {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 8px;
  border: 1px solid var(--mm-border);
  background: var(--mm-raised);
  color: var(--mm-text);
  font-size: 1.1rem;
  cursor: pointer;
  line-height: 1;
}

.icon-btn:disabled {
  opacity: 0.5;
}

.status-pill {
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 0.3rem 0.5rem;
  border-radius: 999px;
  border: 1px solid var(--mm-border);
}

.status-pill.on {
  color: var(--mm-brand);
  background: var(--mm-brand-soft);
  border-color: rgba(61, 255, 122, 0.35);
}

.status-pill.off {
  color: var(--mm-warn);
  background: var(--mm-warn-soft);
  border-color: rgba(229, 184, 74, 0.35);
}

.tg-main {
  flex: 1;
  padding: var(--dash-pad);
  padding-bottom: max(0.5rem, env(safe-area-inset-bottom));
}

.tg-main.dash {
  width: 100%;
  max-width: var(--dash-max);
  margin-left: auto;
  margin-right: auto;
  display: flex;
  flex-direction: column;
  gap: var(--dash-gap);
}

.tg-lead {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--mm-muted);
  text-align: center;
}

.community-banner {
  margin: 0;
  padding: 0.85rem 1rem;
  border-radius: var(--dash-radius);
  border: 1px solid rgba(184, 255, 61, 0.28);
  background: var(--mm-brand-soft);
}
.community-banner-body,
.community-banner-ai {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.45;
  color: var(--mm-text);
}
.community-banner-ai {
  margin-top: 0.4rem;
  color: var(--mm-muted);
}

/* Action tiles — responsive grid */
.dash-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
}

@media (min-width: 420px) {
  .dash-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.action-tile {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.2rem;
  padding: 0.9rem 1rem;
  min-height: 5.5rem;
  border-radius: var(--dash-radius);
  border: 1px solid var(--mm-border);
  background: var(--mm-surface);
  color: var(--mm-text);
  font-family: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.action-tile:active {
  transform: scale(0.98);
}

.action-tile--primary {
  grid-column: 1 / -1;
  border-color: rgba(184, 255, 61, 0.45);
  background: linear-gradient(145deg, rgba(184, 255, 61, 0.22), rgba(184, 255, 61, 0.06));
}

.action-tile--accent {
  border-color: var(--mm-brand);
  background: var(--mm-brand-soft);
}

.action-tile--accent .action-title {
  color: var(--mm-brand);
}

.action-icon {
  font-size: 1.25rem;
  line-height: 1;
}

.action-title {
  font-size: 0.9rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.action-sub {
  font-size: 0.7rem;
  line-height: 1.35;
  color: var(--mm-muted);
}

/* Bento stats */
.dash-bento {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
}

.bento-card {
  padding: 0.85rem 1rem;
  border-radius: var(--dash-radius);
  border: 1px solid var(--mm-border);
  background: var(--mm-surface);
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-height: 4.5rem;
}

.bento-card--hero {
  grid-column: 1 / -1;
  min-height: 5.25rem;
  border-color: rgba(184, 255, 61, 0.35);
  background: linear-gradient(160deg, var(--mm-surface) 0%, rgba(184, 255, 61, 0.08) 100%);
}

.bento-k {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--mm-muted);
}

.bento-v {
  font-size: 1.35rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.03em;
}

.bento-card--hero .bento-v {
  font-size: 1.65rem;
}

.bento-v-sm {
  font-size: 1rem;
}

.bento-meta {
  font-size: 0.7rem;
  color: var(--mm-muted);
}

.dash-membership {
  margin-bottom: 0;
}

.membership-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem 0.65rem;
  margin-bottom: 0.75rem;
}

@media (max-width: 360px) {
  .membership-strip {
    grid-template-columns: 1fr;
  }
}

.membership-item {
  min-width: 0;
}

.dash-split {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--dash-gap);
}

@media (min-width: 480px) {
  .dash-split {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.dash-panel {
  margin-bottom: 0;
  height: fit-content;
}

.dash-panel-wide {
  margin-bottom: 0;
}

.dash-details {
  margin-bottom: 0;
  padding: 0;
  overflow: hidden;
}

.dash-details-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  cursor: pointer;
  list-style: none;
  user-select: none;
}

.dash-details-summary::-webkit-details-marker {
  display: none;
}

.dash-details-chevron {
  font-size: 1.25rem;
  color: var(--mm-muted);
  transition: transform 0.2s;
}

.dash-details[open] .dash-details-chevron {
  transform: rotate(90deg);
}

.dash-details-body {
  padding: 0 1rem 1rem;
  border-top: 1px solid var(--mm-border);
}

.list-cards {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.list-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
  padding: 0.75rem 0.85rem;
  border-radius: 12px;
  background: var(--mm-raised);
  border: 1px solid var(--mm-border);
}

.list-card--stack {
  flex-direction: column;
  align-items: stretch;
}

.list-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.list-card-main {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  min-width: 0;
}

.list-card-title {
  font-size: 0.9rem;
  letter-spacing: -0.02em;
}

.list-card-value {
  font-size: 0.9rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.list-card-meta {
  margin: 0;
  font-size: 0.7rem;
  color: var(--mm-muted);
}

.list-card-pl {
  margin: 0;
  font-size: 0.8125rem;
  font-weight: 600;
}

.status-badge {
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.2rem 0.45rem;
  border-radius: 6px;
  background: var(--mm-raised);
  color: var(--mm-muted);
  border: 1px solid var(--mm-border);
  flex-shrink: 0;
}

.tg-card {
  background: var(--mm-surface);
  border: 1px solid var(--mm-border);
  border-radius: var(--dash-radius);
  padding: 1rem;
  margin-bottom: 0;
}

.tg-card-accent {
  border-top: 2px solid var(--mm-brand);
}

.tg-card-dim {
  background: var(--mm-raised);
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.85rem;
}

.card-label {
  margin: 0;
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--mm-muted);
}

.card-label.solo {
  margin-bottom: 0.75rem;
}

.chip {
  font-size: 0.6875rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 6px;
  background: var(--mm-brand-soft);
  color: var(--mm-brand);
  border: 1px solid rgba(61, 255, 122, 0.25);
}

.chip-muted {
  background: var(--mm-raised);
  color: var(--mm-muted);
  border-color: var(--mm-border);
}

.chip-warn {
  background: var(--mm-warn-soft);
  color: var(--mm-warn);
  border-color: rgba(229, 184, 74, 0.3);
}

.stat-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem 0.75rem;
  margin-bottom: 0.85rem;
}

.stat-row-3 {
  grid-template-columns: repeat(3, 1fr);
  margin-bottom: 0;
}

@media (max-width: 360px) {
  .stat-row-3 {
    grid-template-columns: 1fr;
  }
}

.stat-k {
  display: block;
  font-size: 0.65rem;
  color: var(--mm-muted);
  margin-bottom: 0.1rem;
}

.stat-v {
  font-size: 0.9rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.stat-v-sm {
  font-size: 0.8rem;
}

.banner {
  margin: 0 0 0.75rem;
  padding: 0.55rem 0.7rem;
  border-radius: 8px;
  font-size: 0.8125rem;
  line-height: 1.4;
}

.banner-ok {
  background: var(--mm-brand-soft);
  color: var(--mm-brand);
  border: 1px solid rgba(61, 255, 122, 0.2);
}

.banner-warn {
  background: var(--mm-warn-soft);
  color: var(--mm-warn);
  border: 1px solid rgba(229, 184, 74, 0.25);
}

.billing-note {
  margin: 0;
  font-size: 0.75rem;
  line-height: 1.5;
  color: var(--mm-muted);
}

.packages-note {
  margin-bottom: 0.85rem;
}

.package-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.65rem;
}

.package-card {
  padding: 0.85rem;
  border-radius: 10px;
  border: 1px solid var(--mm-border);
  background: var(--mm-raised);
}

.package-card.recommended {
  border-color: rgba(61, 255, 122, 0.35);
}

.package-card.current {
  border-color: var(--mm-brand);
  box-shadow: 0 0 0 1px var(--mm-brand-soft);
}

.package-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
}

.package-name {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.package-desc {
  margin: 0.2rem 0 0;
  font-size: 0.75rem;
  line-height: 1.45;
  color: var(--mm-muted);
}

.package-badge {
  flex-shrink: 0;
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.2rem 0.45rem;
  border-radius: 6px;
  background: var(--mm-brand-soft);
  color: var(--mm-brand);
  border: 1px solid rgba(61, 255, 122, 0.25);
}

.package-badge-rec {
  background: var(--mm-raised);
  color: var(--mm-muted);
  border-color: var(--mm-border);
}

.package-price {
  margin: 0 0 0.55rem;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--mm-brand);
  font-variant-numeric: tabular-nums;
}

.package-features {
  margin: 0 0 0.65rem;
  padding-left: 1rem;
  font-size: 0.75rem;
  color: var(--mm-muted);
  line-height: 1.55;
}

.package-btn {
  min-height: 42px;
  font-size: 0.8125rem;
}

.btn-row {
  margin-top: 0.85rem;
}

.btn {
  width: 100%;
  min-height: 48px;
  padding: 0.65rem 1rem;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 700;
  font-family: inherit;
  cursor: pointer;
  border: none;
}

.btn:disabled {
  opacity: 0.55;
}

.btn-primary {
  background: linear-gradient(180deg, var(--mm-brand), #32d96a);
  color: var(--mm-on-brand);
}

.btn-secondary {
  background: transparent;
  border: 1px solid var(--mm-border);
  color: var(--mm-text);
  margin-bottom: 0.65rem;
}

.btn-text {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.5rem;
  border: none;
  background: none;
  color: var(--mm-brand);
  font-size: 0.8125rem;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
}

.type-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.5rem;
}

.type-chip {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.3rem 0.6rem;
  border-radius: 999px;
  background: var(--mm-brand-soft);
  color: var(--mm-brand);
  border: 1px solid rgba(61, 255, 122, 0.25);
}

.card-hint {
  margin: 0;
  font-size: 0.75rem;
  color: var(--mm-muted);
}

.card-hint code {
  font-size: 0.7rem;
  color: var(--mm-brand);
  background: var(--mm-brand-soft);
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
}

.usage-bars {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}

.usage-label {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: var(--mm-muted);
  margin-bottom: 0.35rem;
}

.bar-track {
  height: 6px;
  border-radius: 999px;
  background: var(--mm-raised);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 999px;
  background: var(--mm-brand);
  transition: width 0.3s ease;
}

.ref-line {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  color: var(--mm-muted);
}

.ref-line code {
  font-size: 0.75rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: var(--mm-raised);
}

.steps {
  margin: 0 0 0.85rem;
  padding-left: 1.1rem;
  font-size: 0.8125rem;
  color: var(--mm-muted);
  line-height: 1.5;
}

.field {
  display: block;
  margin-bottom: 0.65rem;
}

.field-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  margin-bottom: 0.35rem;
  color: var(--mm-muted);
}

.field-input {
  width: 100%;
  padding: 0.7rem 0.85rem;
  border-radius: 8px;
  border: 1px solid var(--mm-border);
  background: var(--mm-bg);
  color: var(--mm-text);
  font-size: 16px;
  font-family: inherit;
}

.field-input:focus {
  outline: none;
  border-color: var(--mm-brand);
  box-shadow: 0 0 0 2px var(--mm-brand-soft);
}

.count {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--mm-muted);
  font-variant-numeric: tabular-nums;
}

.empty {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--mm-muted);
  text-align: center;
  padding: 0.5rem 0;
}

.pos-list,
.trade-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.pos-item,
.trade-item {
  padding: 0.7rem 0;
  border-bottom: 1px solid var(--mm-border);
}

.pos-item:last-child,
.trade-item:last-child {
  border-bottom: none;
}

.pos-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.side {
  margin-left: 0.35rem;
  font-size: 0.625rem;
  font-weight: 700;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  vertical-align: middle;
}

.side.buy {
  color: var(--mm-brand);
  background: var(--mm-brand-soft);
}

.side.sell {
  color: var(--mm-muted);
  background: var(--mm-raised);
}

.trade-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.25rem;
}

.status {
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--mm-muted);
}

.trade-meta {
  margin: 0 0 0.2rem;
  font-size: 0.75rem;
  color: var(--mm-muted);
}

.trade-pl {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 700;
}

.pos {
  color: var(--mm-brand);
}

.neg {
  color: var(--mm-loss);
}

.cmd-list {
  margin: 0;
  padding-left: 1rem;
  font-size: 0.8125rem;
  color: var(--mm-muted);
  line-height: 1.65;
}

.cmd-list code {
  color: var(--mm-brand);
  font-size: 0.75rem;
}

.state-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2.5rem 1.25rem;
  text-align: center;
}

.state-title {
  margin: 1rem 0 0.35rem;
  font-weight: 700;
}

.state-sub {
  margin: 0 0 1rem;
  font-size: 0.875rem;
  color: var(--mm-muted);
  max-width: 260px;
}

.state-error .state-sub {
  color: var(--mm-loss);
}

.spinner {
  width: 32px;
  height: 32px;
  border: 2px solid var(--mm-border);
  border-top-color: var(--mm-brand);
  border-radius: 50%;
  animation: spin 0.65s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.inline-err {
  margin: 0.5rem 0 0;
  padding: 0.65rem;
  font-size: 0.8125rem;
  color: var(--mm-loss);
  background: rgba(248, 113, 113, 0.1);
  border-radius: 8px;
  border: 1px solid rgba(248, 113, 113, 0.25);
}

.tg-footer {
  padding: 1rem;
  padding-bottom: max(1rem, env(safe-area-inset-bottom));
  border-top: 1px solid var(--mm-border);
  text-align: center;
}

.tg-footer p {
  margin: 0;
  font-size: 0.6875rem;
  color: var(--mm-muted);
}

.tg-footer-sub {
  margin-top: 0.25rem !important;
  opacity: 0.85;
}

.tg-splash {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--mm-bg);
  padding: max(1.5rem, env(safe-area-inset-top)) 1.5rem max(1.5rem, env(safe-area-inset-bottom));
  opacity: 1;
  transition: opacity 0.38s ease;
}

.tg-splash--fade {
  opacity: 0;
  pointer-events: none;
}

.splash-enter-active,
.splash-leave-active {
  transition: opacity 0.38s ease;
}

.splash-enter-from,
.splash-leave-to {
  opacity: 0;
}

.tg-splash-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 280px;
}

.splash-logo {
  width: 88px !important;
  height: 88px !important;
  margin-bottom: 1.1rem;
  animation: splash-pulse 2.2s ease-in-out infinite;
}

.splash-title {
  margin: 0 0 0.35rem;
  font-size: 1.35rem;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.splash-tagline {
  margin: 0 0 1.5rem;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--mm-muted);
}

.splash-loader {
  display: flex;
  gap: 0.45rem;
  align-items: center;
  justify-content: center;
}

.splash-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--mm-brand);
  animation: splash-bounce 1s ease-in-out infinite;
}

.splash-dot:nth-child(2) {
  animation-delay: 0.15s;
}

.splash-dot:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes splash-pulse {
  0%,
  100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.04);
    opacity: 0.92;
  }
}

@keyframes splash-bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }
  40% {
    transform: translateY(-6px);
    opacity: 1;
  }
}
</style>
