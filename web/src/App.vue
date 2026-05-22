<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, API, apiTargetLabel, loadSession, clearSession } from './api'
import { Perm, can, applyProfile } from './acl'
import { SLOGAN_SHORT, TAGLINE } from './brand'
import TelegramLogin from './components/TelegramLogin.vue'
import EmailAdminLogin from './components/EmailAdminLogin.vue'
import LandingHero from './components/LandingHero.vue'
import AdminPanel from './components/AdminPanel.vue'
import BrandLogo from './components/BrandLogo.vue'
import TradingPairs from './components/TradingPairs.vue'
import AppFooter from './components/AppFooter.vue'

const loggedIn = ref(false)
const userName = ref('')
const apiOffline = ref(false)
const config = ref(null)
const status = ref(null)
const account = ref(null)
const subscription = ref(null)
const positions = ref([])
const trades = ref([])
const adminTrades = ref([])
const brokers = ref([])
const provider = ref('mock')
const dynamicFields = ref([])
const brokerLabel = ref('')
const brokerConnection = ref(null)
const credentials = ref({})
const message = ref('')
const messageOk = ref(true)
const adminStats = ref(null)
const recentUsers = ref([])
const activateTarget = ref('')
const activateDays = ref(30)
const isAdmin = ref(false)
const role = ref('user')
const permissions = ref([])
const isBlocked = ref(false)
const canTrade = ref(true)
const tradeMessage = ref('')

const selectedBroker = computed(() => brokers.value.find((b) => b.id === provider.value))
const brokerIsLive = computed(() => selectedBroker.value?.status === 'live')
const botUsername = computed(() => config.value?.telegram_bot_username || 'market_mamba_bot')

function onLoggedIn(data) {
  loggedIn.value = true
  applyProfile({ role, isAdmin, permissions, isBlocked, canTrade, tradeMessage }, data)
  userName.value =
    [data.user?.first_name, data.user?.last_name].filter(Boolean).join(' ') ||
    data.email ||
    String(data.telegram_id)
  if (isBlocked.value) {
    message.value = 'Your account is blocked. Contact support.'
    messageOk.value = false
  } else {
    message.value = `Welcome, ${userName.value}!`
    messageOk.value = true
  }
  refresh()
}

function logout() {
  clearSession()
  loggedIn.value = false
  status.value = null
  account.value = null
  trades.value = []
  adminTrades.value = []
  adminStats.value = null
  role.value = 'user'
  permissions.value = []
  isBlocked.value = false
  canTrade.value = true
  tradeMessage.value = ''
  message.value = 'Logged out'
  messageOk.value = true
}

function onProviderChange() {
  dynamicFields.value = selectedBroker.value?.fields || []
  const next = {}
  for (const f of dynamicFields.value) {
    if (f.type === 'boolean') {
      next[f.key] = credentials.value[f.key] === 'true' ? 'true' : 'false'
    } else {
      next[f.key] = credentials.value[f.key] ?? ''
    }
  }
  credentials.value = next
}

function fmtProfit(t) {
  if (t.profit == null) return '—'
  const n = Number(t.profit)
  return `${n >= 0 ? '+' : ''}$${n.toFixed(2)}`
}

async function refresh() {
  if (!loggedIn.value) return
  message.value = ''
  try {
    config.value = await fetch(`${API}/config`).then((r) => r.json())
    const me = await api('/auth/me')
    applyProfile({ role, isAdmin, permissions, isBlocked, canTrade, tradeMessage }, me)
    userName.value = [me.user?.first_name, me.user?.last_name].filter(Boolean).join(' ')
    if (isBlocked.value) {
      message.value = 'Your account is blocked. Contact support.'
      messageOk.value = false
      return
    }
    status.value = await api('/status')
    subscription.value = await api('/subscription')
    try {
      account.value = await api('/account')
    } catch {
      account.value = null
    }
    try {
      const pos = await api('/positions')
      positions.value = pos.positions || []
    } catch {
      positions.value = []
    }
    try {
      const tr = await api('/trades')
      trades.value = tr.trades || []
    } catch (e) {
      trades.value = []
      console.warn('trades', e)
    }
    const bt = await api('/brokers/types')
    brokers.value = bt.brokers || []
    onProviderChange()
    const conn = await api('/brokers/connection')
    brokerConnection.value = conn.connection
    if (conn.connection) {
      provider.value = conn.connection.provider
      brokerLabel.value = conn.connection.label || ''
      onProviderChange()
    }
    if (can(permissions.value, Perm.adminStats)) {
      adminStats.value = await api('/admin/stats')
      const u = await api('/admin/users')
      recentUsers.value = u.users || []
      try {
        const at = await api('/admin/trades')
        adminTrades.value = at.trades || []
      } catch {
        adminTrades.value = []
      }
    }
  } catch (e) {
    if (e.message.includes('session') || e.message.includes('log in')) {
      logout()
    }
    message.value = e.message
    messageOk.value = false
  }
}

async function connectMockDemo() {
  provider.value = 'mock'
  brokerLabel.value = 'Demo account'
  credentials.value = { initial_balance: '10000' }
  onProviderChange()
  await saveBroker()
}

async function testBroker() {
  try {
    const creds = { ...credentials.value }
    const r = await api('/brokers/test', {
      method: 'POST',
      body: { provider: provider.value, label: brokerLabel.value, credentials: creds },
    })
    message.value = `Connection OK — balance $${r.balance}`
    messageOk.value = true
  } catch (e) {
    message.value = e.message
    messageOk.value = false
  }
}

async function saveBroker() {
  if (!brokerIsLive.value) {
    message.value = 'This broker is not available yet — use Mock (Demo)'
    messageOk.value = false
    return
  }
  try {
    const creds = { ...credentials.value }
    await api('/brokers/connection', {
      method: 'POST',
      body: { provider: provider.value, label: brokerLabel.value, credentials: creds },
    })
    message.value = 'Broker saved'
    messageOk.value = true
    refresh()
  } catch (e) {
    message.value = e.message
    messageOk.value = false
  }
}

async function adminBlockUser(telegramId, blocked) {
  try {
    await api('/admin/users/block', { method: 'POST', body: { telegram_id: Number(telegramId), blocked } })
    message.value = blocked ? 'User blocked' : 'User unblocked'
    messageOk.value = true
    refresh()
  } catch (e) {
    message.value = e.message
    messageOk.value = false
  }
}

async function adminRevoke(telegramId) {
  try {
    await api('/admin/users/revoke', { method: 'POST', body: { telegram_id: Number(telegramId) } })
    message.value = 'Subscription revoked'
    messageOk.value = true
    refresh()
  } catch (e) {
    message.value = e.message
    messageOk.value = false
  }
}

async function adminBroadcastSignal() {
  try {
    const r = await api('/admin/signals/broadcast', { method: 'POST', body: { generate: true } })
    message.value = `Signal sent to ${r.sent} subscribers (${r.signal?.symbol} ${r.signal?.type})`
    messageOk.value = true
    refresh()
  } catch (e) {
    message.value = e.message
    messageOk.value = false
  }
}

async function adminActivate() {
  try {
    await api('/admin/activate', {
      method: 'POST',
      body: {
        telegram_id: Number(activateTarget.value),
        days: Number(activateDays.value),
        plan: 'manual',
        notes: 'Activated from web admin',
      },
    })
    message.value = 'Subscription activated'
    messageOk.value = true
    refresh()
  } catch (e) {
    message.value = e.message
    messageOk.value = false
  }
}

async function loadConfig() {
  try {
    const res = await fetch(`${API}/config`)
    if (!res.ok) throw new Error('API error')
    config.value = await res.json()
    apiOffline.value = false
  } catch {
    apiOffline.value = true
    config.value = {
      telegram_bot_username: 'market_mamba_bot',
      telegram_login_enabled: true,
    }
  }
}

async function tryRestoreSession() {
  const s = loadSession()
  if (!s.sessionToken) return
  try {
    await loadConfig()
    loggedIn.value = true
    await refresh()
  } catch {
    clearSession()
    loggedIn.value = false
  }
}

onMounted(async () => {
  await loadConfig()
  await tryRestoreSession()
})
</script>

<template>
  <div class="app-shell">
  <header class="header" :class="{ 'header-landing': !loggedIn }">
    <div class="brand">
      <BrandLogo :variant="loggedIn ? 'icon' : 'landscape'" />
      <div v-if="loggedIn" class="brand-text">
        <h1 class="brand-title">Market Mamba</h1>
        <p class="muted brand-tag">{{ TAGLINE }}</p>
      </div>
      <p v-else class="muted brand-tag brand-tag-landing">{{ SLOGAN_SHORT }}</p>
    </div>
    <div v-if="loggedIn" class="header-actions">
      <span v-if="userName" class="user-pill">Hi, <strong>{{ userName }}</strong></span>
      <span v-if="isAdmin" class="admin-badge">Admin</span>
      <span v-else class="user-badge">Trader</span>
      <button type="button" class="btn-secondary" @click="logout">Log out</button>
    </div>
    <a v-else class="header-cta" href="#login-portal">Get started</a>
  </header>

  <p v-if="loggedIn && isBlocked" class="err banner-msg blocked-banner">
    Account blocked — you cannot use trading features. Contact support.
  </p>
  <p v-else-if="loggedIn && !canTrade && tradeMessage" class="err banner-msg blocked-banner">
    {{ tradeMessage }}
  </p>
  <p v-if="message" :class="messageOk ? 'ok banner-msg' : 'err banner-msg'">{{ message }}</p>

  <main class="app-main">
  <LandingHero
    v-if="!loggedIn"
    :config="config"
    :api-offline="apiOffline"
    :api-target="apiTargetLabel()"
    :bot-username="botUsername"
    @error="(m) => { message = m; messageOk = false }"
  >
    <TelegramLogin
      v-if="config && config.telegram_login_enabled !== false"
      :bot-username="botUsername"
      :client-id="config.telegram_client_id || ''"
      :login-domain="config.telegram_login_domain || 'marketmamba.kkooapp.co.tz'"
      :public-site-url="config.public_site_url || 'https://marketmamba.kkooapp.co.tz'"
      @logged-in="onLoggedIn"
      @error="(m) => { message = m; messageOk = false }"
    />
    <p v-else-if="config && config.telegram_login_enabled === false" class="muted portal-no-tg">
      Telegram login is not configured on the server (set TELEGRAM_BOT_TOKEN).
    </p>
    <p v-else class="muted portal-no-tg">Loading sign-in…</p>
    <hr class="divider" />
    <EmailAdminLogin
      @logged-in="onLoggedIn"
      @error="(m) => { message = m; messageOk = false }"
    />
  </LandingHero>

  <div v-else class="grid dashboard">
    <AdminPanel
      v-if="can(permissions, Perm.adminStats) && adminStats"
      :stats="adminStats"
      :users="recentUsers"
      :trades="adminTrades"
      :activate-target="activateTarget"
      :activate-days="activateDays"
      @update:activate-target="activateTarget = $event"
      @update:activate-days="activateDays = $event"
      @broadcast-signal="adminBroadcastSignal"
      @activate="adminActivate"
      @block-user="adminBlockUser"
      @revoke-user="adminRevoke"
    />

    <template v-if="!isBlocked">
    <section class="card card-bull">
      <h2>Status</h2>
      <template v-if="status">
        <p>Broker: <strong>{{ status.provider }}</strong></p>
        <p>Can trade: {{ status.can_trade ? 'yes' : 'no' }}</p>
        <p v-if="status.trade_message" class="muted">{{ status.trade_message }}</p>
        <p>Auto: {{ status.auto_trading ? 'on' : 'off' }}</p>
      </template>
    </section>

    <section class="card card-bull">
      <h2>Account</h2>
      <template v-if="account">
        <p>Balance: <strong>${{ account.balance?.toFixed(2) }}</strong></p>
        <p>Equity: ${{ account.equity?.toFixed(2) }}</p>
      </template>
      <p v-else class="muted">Connect Mock broker below</p>
    </section>

    <TradingPairs
      :can-trade="canTrade && !isBlocked"
      @message="(m) => { message = m.text; messageOk = m.ok }"
    />

    <section class="card card-win">
      <h2>Subscription</h2>
      <template v-if="subscription">
        <p v-if="subscription.subscription">
          Plan: {{ subscription.subscription.plan }} · {{ subscription.subscription.status }}
        </p>
        <p class="muted">{{ config?.subscription_message }}</p>
      </template>
    </section>

    <section class="card wide trade-log-card">
      <div class="trade-log-head">
        <h2>📋 Your trade log</h2>
        <button type="button" class="btn-secondary" @click="refresh">Refresh</button>
      </div>
      <div class="table-wrap">
        <table v-if="trades.length">
          <thead>
            <tr>
              <th>Time</th><th>Symbol</th><th>Side</th><th>Qty</th><th>Entry</th><th>Status</th><th>P/L</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in trades" :key="t.id">
              <td>{{ new Date(t.created_at).toLocaleString() }}</td>
              <td><strong>{{ t.symbol }}</strong></td>
              <td :class="t.type === 'BUY' ? 'buy' : 'sell'">{{ t.type }}</td>
              <td>{{ t.quantity }}</td>
              <td>{{ Number(t.entry_price).toFixed(5) }}</td>
              <td>
                <span class="status-chip" :class="t.status?.toLowerCase()">{{ t.status }}</span>
                <span v-if="t.closure_reason" class="muted"> {{ t.closure_reason }}</span>
              </td>
              <td :class="{ profit: t.profit > 0, loss: t.profit < 0 }">{{ fmtProfit(t) }}</td>
            </tr>
          </tbody>
        </table>
        <p v-else class="muted empty-hint">
          No trades logged for your account yet.<br />
          Telegram: <code>/broker connect</code> then <code>/open EURUSD BUY 0.1 1.08 1.10</code> or <code>/autostart</code>
        </p>
      </div>
    </section>

    <section class="card wide card-bull">
      <h2>Broker connection</h2>
      <p v-if="brokerConnection" class="ok broker-connected">
        Connected: <strong>{{ brokerConnection.provider }}</strong>
        <span v-if="brokerConnection.label"> — {{ brokerConnection.label }}</span>
      </p>
      <p v-else class="muted">No broker — use quick connect or Telegram <code>/broker connect</code></p>
      <button
        v-if="!brokerConnection || brokerConnection.provider !== 'mock'"
        type="button"
        class="btn-primary broker-quick"
        @click="connectMockDemo"
      >
        Connect Mock Demo ($10,000)
      </button>
      <hr class="divider broker-divider" />
      <div class="field">
        <label>Broker</label>
        <select v-model="provider" @change="onProviderChange">
          <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.name }} ({{ b.status }})</option>
        </select>
      </div>
      <div v-for="f in dynamicFields" :key="f.key" class="field">
        <label>{{ f.label }}</label>
        <label v-if="f.type === 'boolean'" class="checkbox-row">
          <input
            type="checkbox"
            :checked="credentials[f.key] === 'true'"
            @change="credentials[f.key] = $event.target.checked ? 'true' : 'false'"
          />
          <span>Enabled</span>
        </label>
        <input v-else v-model="credentials[f.key]" :type="f.type === 'password' ? 'password' : 'text'" />
      </div>
      <div class="broker-actions">
        <button type="button" class="btn-secondary" :disabled="!brokerIsLive" @click="testBroker">Test</button>
        <button type="button" class="btn-primary" :disabled="!brokerIsLive" @click="saveBroker">Save</button>
      </div>
    </section>

    <section class="card wide">
      <h2>Open positions</h2>
      <table v-if="positions.length">
        <thead><tr><th>Symbol</th><th>Type</th><th>Qty</th><th>P/L</th></tr></thead>
        <tbody>
          <tr v-for="p in positions" :key="p.id">
            <td>{{ p.symbol }}</td>
            <td>{{ p.type }}</td>
            <td>{{ p.quantity }}</td>
            <td>{{ p.profit?.toFixed(2) }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="muted">No open positions</p>
    </section>
    </template>
  </div>
  </main>

  <AppFooter :landing="!loggedIn" :bot-username="botUsername" />
  </div>
</template>

<style scoped>
.header.header-landing {
  background: var(--header-bg);
  border-bottom-color: var(--border);
  backdrop-filter: blur(12px);
}
.header-cta {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: var(--touch-min);
  padding: 0.625rem 1.25rem;
  border-radius: 999px;
  font-weight: 700;
  font-size: 0.95rem;
  text-decoration: none;
  color: var(--on-brand);
  background: linear-gradient(180deg, var(--brand), var(--brand-dim));
  box-shadow: 0 0 20px var(--win-glow);
  transition: transform 0.2s, box-shadow 0.2s;
}
.header-cta:hover {
  transform: translateY(-2px);
  box-shadow: 0 0 32px var(--win-glow);
}
.header-cta:active {
  transform: translateY(0);
}
.brand {
  display: flex;
  align-items: center;
  gap: 0.5rem 0.75rem;
  flex-wrap: wrap;
  flex: 1 1 auto;
  min-width: 0;
  max-width: 100%;
}
.brand-text { min-width: 0; }
.brand-title {
  margin: 0;
  font-size: clamp(1rem, 4vw, 1.25rem);
  font-weight: 800;
  overflow-wrap: anywhere;
}
.brand-tag {
  margin: 0;
  font-size: 0.75rem;
  overflow-wrap: anywhere;
}
.brand-tag-landing {
  display: none;
}
.header-landing .brand {
  width: 100%;
}
.header-landing .header-cta {
  width: 100%;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem 0.75rem;
  flex-wrap: wrap;
  width: 100%;
  justify-content: flex-start;
}
@media (min-width: 640px) {
  .header-landing .brand {
    width: auto;
    flex: 1 1 auto;
  }
  .header-landing .header-cta {
    width: auto;
    flex-shrink: 0;
  }
  .header-actions {
    width: auto;
    justify-content: flex-end;
  }
}
@media (min-width: 720px) {
  .header-landing .brand-tag-landing {
    display: block;
  }
}
.user-pill { font-size: 0.95rem; }
.admin-badge,
.user-badge {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
}
.admin-badge {
  background: var(--brand-muted);
  color: var(--brand);
  border: 1px solid var(--brand-dim);
}
.user-badge {
  background: var(--brand-muted);
  color: var(--brand-dim);
  border: 1px solid var(--border-strong);
}
.dashboard {
  width: 100%;
  max-width: 1200px;
}
.trade-log-head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}
.broker-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.trade-log-card {
  position: relative;
  overflow: hidden;
}
.trade-log-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--brand), var(--brand-deep));
}
.trade-log-head h2 { margin: 0; font-size: clamp(1rem, 4vw, 1.25rem); }
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
.empty-hint { padding: 1.5rem; text-align: center; line-height: 1.6; }
.divider { border: none; border-top: 1px solid var(--border); margin: 1.25rem 0; }
.broker-connected { margin-bottom: 0.75rem; }
.broker-quick { margin-bottom: 0.75rem; }
.broker-divider { margin: 1rem 0; }

</style>
