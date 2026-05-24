<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, API, apiTargetLabel, loadSession, clearSession } from './api'
import { Perm, can, applyProfile } from './acl'
import { SLOGAN_SHORT, TAGLINE } from './brand'
import TelegramLogin from './components/TelegramLogin.vue'
import EmailAdminLogin from './components/EmailAdminLogin.vue'
import LandingHero from './components/LandingHero.vue'
import AdminPanel from './components/AdminPanel.vue'
import UserDashboard from './components/UserDashboard.vue'
import BrandLogo from './components/BrandLogo.vue'
import TradingPairs from './components/TradingPairs.vue'
import BrokerConnectWizard from './components/BrokerConnectWizard.vue'
import FilterStack from './components/FilterStack.vue'
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
const brands = ref([])
const metaapiSharedToken = ref(false)
const brokerConnection = ref(null)
const adminBrokerConnections = ref({})
const adminEnabledBrands = ref([])
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
const telegramId = ref('')

const botUsername = computed(() => config.value?.telegram_bot_username || 'market_mamba_bot')
const publicSiteUrl = computed(() => config.value?.public_site_url || '')

function onLoggedIn(data) {
  loggedIn.value = true
  applyProfile({ role, isAdmin, permissions, isBlocked, canTrade, tradeMessage }, data)
  telegramId.value = data.telegram_id || localStorage.getItem('mm_telegram_id') || ''
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

function onBrokerWizardMessage(m) {
  message.value = m.text
  messageOk.value = m.ok
}

function maybeShowCommunityUnlock() {
  const cfg = config.value
  if (!cfg?.asset_phase_unlocked || !cfg?.community_unlock_message) return
  if (localStorage.getItem('mm_community_unlock_seen') === '1') return
  localStorage.setItem('mm_community_unlock_seen', '1')
  message.value = cfg.community_unlock_message
  messageOk.value = true
}

async function refresh() {
  if (!loggedIn.value) return
  message.value = ''
  try {
    config.value = await fetch(`${API}/config`).then((r) => r.json())
    maybeShowCommunityUnlock()
    const me = await api('/auth/me')
    applyProfile({ role, isAdmin, permissions, isBlocked, canTrade, tradeMessage }, me)
    telegramId.value = me.telegram_id || ''
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
    brands.value = bt.brands || []
    metaapiSharedToken.value = bt.metaapi_shared_token === true
    const conn = await api('/brokers/connection')
    brokerConnection.value = conn.connection
    if (can(permissions.value, Perm.adminStats)) {
      const adminRes = await api('/admin/stats')
      adminStats.value = adminRes.stats || adminRes
      adminBrokerConnections.value = adminRes.broker_connections || {}
      adminEnabledBrands.value = adminRes.enabled_broker_brands || []
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
    if (!res.ok) {
      throw new Error(res.status === 500 ? 'Server error — is the API running on port 8090?' : `API ${res.status}`)
    }
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
    <a v-else class="header-cta" href="#login-portal">Sign in</a>
  </header>

  <p v-if="loggedIn && isBlocked" class="err banner-msg blocked-banner">
    Your account is restricted. Contact us for assistance.
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
      :broker-connections="adminBrokerConnections"
      :enabled-broker-brands="adminEnabledBrands"
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
    <UserDashboard
      :status="status"
      :account="account"
      :subscription="subscription"
      :config="config"
      :positions="positions"
      :trades="trades"
      :telegram-id="telegramId"
      :can-trade="canTrade"
      @refresh="refresh"
    />

    <TradingPairs
      :config="config"
      :can-trade="canTrade && !isBlocked"
      @message="(m) => { message = m.text; messageOk = m.ok }"
    />

    <FilterStack :symbols="config?.signal_symbols || ['EURUSD', 'BTCUSD']" />

    <BrokerConnectWizard
      id="connect"
      :brands="brands"
      :brokers="brokers"
      :connection="brokerConnection"
      :public-site-url="publicSiteUrl"
      :metaapi-shared-token="metaapiSharedToken"
      @saved="refresh"
      @message="onBrokerWizardMessage"
    />
    </template>
  </div>
  </main>

  <AppFooter
    :landing="!loggedIn"
    :bot-username="botUsername"
    :contact-url="config?.contact_us_url || ''"
    :contact-label="config?.contact_us_label || 'Contact us'"
  />
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
