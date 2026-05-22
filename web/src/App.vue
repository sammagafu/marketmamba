<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, API, apiTargetLabel, loadSession, saveLegacySession, clearSession } from './api'
import TelegramLogin from './components/TelegramLogin.vue'
import EmailAdminLogin from './components/EmailAdminLogin.vue'

const loggedIn = ref(false)
const userName = ref('')
const apiKey = ref(loadSession().apiKey)
const telegramId = ref(loadSession().telegramId)
const isLocalhost = ref(
  typeof window !== 'undefined' &&
    (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'),
)
const showManual = ref(isLocalhost.value)
const apiOffline = ref(false)
const config = ref(null)
const status = ref(null)
const account = ref(null)
const subscription = ref(null)
const positions = ref([])
const brokers = ref([])
const provider = ref('mock')
const dynamicFields = ref([])
const brokerLabel = ref('')
const message = ref('')
const messageOk = ref(true)
const adminStats = ref(null)
const recentUsers = ref([])
const activateTarget = ref('')
const activateDays = ref(30)
const isAdmin = ref(false)

const selectedBroker = computed(() => brokers.value.find((b) => b.id === provider.value))
const botUsername = computed(() => config.value?.telegram_bot_username || 'market_mamba_bot')

function onLoggedIn(data) {
  loggedIn.value = true
  userName.value =
    [data.user?.first_name, data.user?.last_name].filter(Boolean).join(' ') ||
    data.email ||
    String(data.telegram_id)
  message.value = `Welcome, ${userName.value}!`
  messageOk.value = true
  refresh()
}

function saveAuth() {
  saveLegacySession(apiKey.value.trim(), telegramId.value.trim())
  loggedIn.value = true
  message.value = 'Manual session saved'
  messageOk.value = true
  refresh()
}

function logout() {
  clearSession()
  loggedIn.value = false
  status.value = null
  account.value = null
  message.value = 'Logged out'
  messageOk.value = true
}

function onProviderChange() {
  dynamicFields.value = selectedBroker.value?.fields || []
}

async function refresh() {
  if (!loggedIn.value) return
  message.value = ''
  try {
    config.value = await fetch(`${API}/config`).then((r) => r.json())
    const me = await api('/auth/me')
    userName.value = [me.user?.first_name, me.user?.last_name].filter(Boolean).join(' ')
    isAdmin.value = me.is_admin
    status.value = await api('/status')
    subscription.value = await api('/subscription')
    try {
      account.value = await api('/account')
    } catch {
      account.value = null
    }
    const pos = await api('/positions')
    positions.value = pos.positions || []
    const bt = await api('/brokers/types')
    brokers.value = bt.brokers || []
    onProviderChange()
    const conn = await api('/brokers/connection')
    if (conn.connection) {
      provider.value = conn.connection.provider
      onProviderChange()
    }
    if (isAdmin.value) {
      adminStats.value = await api('/admin/stats')
      const u = await api('/admin/users')
      recentUsers.value = u.users || []
    }
  } catch (e) {
    if (e.message.includes('session') || e.message.includes('log in')) {
      logout()
    }
    message.value = e.message
    messageOk.value = false
  }
}

async function testBroker() {
  try {
    const creds = collectCreds()
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
  try {
    const creds = collectCreds()
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

function collectCreds() {
  const creds = {}
  document.querySelectorAll('[data-cred]').forEach((el) => {
    if (el.value) creds[el.dataset.cred] = el.value
  })
  return creds
}

async function adminBlockUser(telegramId, blocked) {
  try {
    await api('/admin/users/block', {
      method: 'POST',
      body: { telegram_id: Number(telegramId), blocked },
    })
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
    await api('/admin/users/revoke', {
      method: 'POST',
      body: { telegram_id: Number(telegramId) },
    })
    message.value = 'Subscription revoked'
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
    message.value = `API unreachable (proxy → ${apiTargetLabel()}). Start Docker: docker compose -p marketmamba up -d — or fix VPS nginx. Use manual login below.`
    messageOk.value = false
  }
}

async function tryRestoreSession() {
  const s = loadSession()
  if (!s.sessionToken && !(s.apiKey && s.telegramId)) return
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
  <header class="header">
    <div>
      <h1>🐍 Market Mamba</h1>
      <p class="muted">Forex automation · per-user brokers</p>
    </div>
    <div v-if="loggedIn" style="display:flex;align-items:center;gap:1rem">
      <span v-if="userName">Hi, <strong>{{ userName }}</strong></span>
      <button class="btn-secondary" @click="logout">Log out</button>
    </div>
  </header>

  <p v-if="message" :class="messageOk ? 'ok' : 'err'" style="text-align:center">{{ message }}</p>

  <section v-if="!loggedIn" class="card wide login-card">
    <h2>Welcome to Market Mamba</h2>
    <p v-if="apiOffline" class="api-offline">
      API not reachable at <strong>{{ apiTargetLabel() }}</strong>.
      Edit <code>web/.env.development</code> and restart <code>npm run dev</code>.
    </p>

    <TelegramLogin
      v-if="config?.telegram_client_id"
      :bot-username="botUsername"
      :client-id="config.telegram_client_id"
      :login-domain="config?.telegram_login_domain || 'marketmamba.kkooapp.co.tz'"
      :public-site-url="config?.public_site_url || 'https://marketmamba.kkooapp.co.tz'"
      @logged-in="onLoggedIn"
      @error="(m) => { message = m; messageOk = false }"
    />

    <hr class="divider" />

    <EmailAdminLogin
      @logged-in="onLoggedIn"
      @error="(m) => { message = m; messageOk = false }"
    />

    <hr class="divider" />

    <p class="muted manual-label">Manual login (API key)</p>
    <button class="btn-secondary" type="button" @click="showManual = !showManual">
      {{ showManual ? 'Hide manual login' : 'Show manual login' }}
    </button>
    <div v-if="showManual" class="manual-row">
      <div class="field">
        <label>API key</label>
        <input v-model="apiKey" type="password" />
      </div>
      <div class="field">
        <label>Telegram ID</label>
        <input v-model="telegramId" />
      </div>
      <button class="btn-primary" @click="saveAuth">Continue</button>
    </div>
  </section>

  <div v-else class="grid">
    <section class="card">
      <h2>Status</h2>
      <template v-if="status">
        <p>Broker: <strong>{{ status.provider }}</strong></p>
        <p>Can trade: {{ status.can_trade ? 'yes' : 'no' }}</p>
        <p v-if="status.trade_message" class="muted">{{ status.trade_message }}</p>
        <p>Auto: {{ status.auto_trading ? 'on' : 'off' }}</p>
      </template>
    </section>

    <section class="card">
      <h2>Account</h2>
      <template v-if="account">
        <p>Balance: <strong>${{ account.balance?.toFixed(2) }}</strong></p>
        <p>Equity: ${{ account.equity?.toFixed(2) }}</p>
      </template>
      <p v-else class="muted">Connect Mock broker below</p>
    </section>

    <section class="card">
      <h2>Subscription</h2>
      <template v-if="subscription">
        <p v-if="subscription.subscription">
          Plan: {{ subscription.subscription.plan }} · {{ subscription.subscription.status }}
        </p>
        <p class="muted">{{ config?.subscription_message }}</p>
      </template>
    </section>

    <section v-if="isAdmin && adminStats" class="card wide">
      <h2>Admin — users</h2>
      <div style="display:flex;gap:2rem;flex-wrap:wrap">
        <div><strong>{{ adminStats.total_users }}</strong><br /><span class="muted">Total users</span></div>
        <div><strong>{{ adminStats.active_subscriptions }}</strong><br /><span class="muted">Active subs</span></div>
        <div><strong>{{ adminStats.auto_trading_users }}</strong><br /><span class="muted">Auto trading</span></div>
        <div><strong>{{ adminStats.new_users_last_7_days }}</strong><br /><span class="muted">New (7d)</span></div>
      </div>
      <div style="margin-top:1rem;display:flex;gap:0.5rem;flex-wrap:wrap">
        <input v-model="activateTarget" placeholder="Telegram user ID" />
        <input v-model="activateDays" type="number" style="width:100px" />
        <button class="btn-primary" @click="adminActivate">Activate (manual pay)</button>
      </div>
      <table v-if="recentUsers.length" style="margin-top:1rem">
        <thead>
          <tr>
            <th>ID</th><th>Name</th><th>Status</th><th>Last seen</th><th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in recentUsers" :key="u.telegram_id">
            <td>{{ u.telegram_id }}</td>
            <td>{{ u.first_name }} {{ u.last_name }} @{{ u.username }}</td>
            <td>{{ u.is_blocked ? 'blocked' : 'active' }}</td>
            <td>{{ new Date(u.last_seen_at).toLocaleString() }}</td>
            <td class="admin-actions">
              <button
                class="btn-secondary"
                type="button"
                @click="adminBlockUser(u.telegram_id, !u.is_blocked)"
              >
                {{ u.is_blocked ? 'Unblock' : 'Block' }}
              </button>
              <button class="btn-secondary" type="button" @click="adminRevoke(u.telegram_id)">
                Revoke sub
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="card wide">
      <h2>Broker connection</h2>
      <div class="field">
        <label>Broker</label>
        <select v-model="provider" @change="onProviderChange">
          <option v-for="b in brokers" :key="b.id" :value="b.id">{{ b.name }} ({{ b.status }})</option>
        </select>
      </div>
      <div v-for="f in dynamicFields" :key="f.key" class="field">
        <label>{{ f.label }}</label>
        <input :data-cred="f.key" :type="f.type === 'password' ? 'password' : 'text'" />
      </div>
      <div class="field">
        <label>Label</label>
        <input v-model="brokerLabel" />
      </div>
      <div style="display:flex;gap:0.5rem">
        <button class="btn-secondary" @click="testBroker">Test</button>
        <button class="btn-primary" @click="saveBroker">Save & activate</button>
      </div>
    </section>

    <section class="card wide">
      <h2>Positions</h2>
      <table v-if="positions.length">
        <thead><tr><th>Symbol</th><th>Type</th><th>Qty</th><th>P/L</th></tr></thead>
        <tbody>
          <tr v-for="p in positions" :key="p.ID">
            <td>{{ p.Symbol }}</td>
            <td>{{ p.Type }}</td>
            <td>{{ p.Quantity }}</td>
            <td>{{ p.Profit?.toFixed(2) }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="muted">No open positions</p>
    </section>
  </div>
</template>

<style scoped>
.login-card { max-width: 480px; margin: 2rem auto; }
.api-offline {
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}
.divider { border: none; border-top: 1px solid var(--border); margin: 1.5rem 0; }
.manual-label { margin-bottom: 0.5rem; font-size: 0.9rem; }
.manual-row { display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: end; margin-top: 1rem; }
.manual-row .field { flex: 1; min-width: 140px; margin: 0; }
.admin-actions { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.admin-actions button { padding: 0.35rem 0.5rem; font-size: 0.8rem; }
</style>
