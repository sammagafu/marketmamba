<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, loadSession, saveSession } from './api'

const apiKey = ref(loadSession().apiKey)
const telegramId = ref(loadSession().telegramId)
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

function saveAuth() {
  saveSession(apiKey.value.trim(), telegramId.value.trim())
  message.value = 'Session saved'
  messageOk.value = true
  refresh()
}

function onProviderChange() {
  dynamicFields.value = selectedBroker.value?.fields || []
}

async function refresh() {
  message.value = ''
  try {
    config.value = await fetch('/api/v1/config').then((r) => r.json())
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
    try {
      adminStats.value = await api('/admin/stats')
      const u = await api('/admin/users')
      recentUsers.value = u.users || []
      isAdmin.value = true
    } catch {
      isAdmin.value = false
      adminStats.value = null
    }
  } catch (e) {
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

onMounted(refresh)
</script>

<template>
  <header class="header">
    <div>
      <h1>🐍 Market Mamba</h1>
      <p class="muted">Public bot · manual subscriptions · per-user brokers</p>
    </div>
    <div class="card" style="display:flex;gap:0.5rem;flex-wrap:wrap;align-items:end">
      <div class="field" style="margin:0">
        <label>API key</label>
        <input v-model="apiKey" type="password" placeholder="WEB_API_KEY" />
      </div>
      <div class="field" style="margin:0">
        <label>Your Telegram ID</label>
        <input v-model="telegramId" placeholder="from @userinfobot" />
      </div>
      <button class="btn-primary" @click="saveAuth">Save</button>
    </div>
  </header>

  <p v-if="message" :class="messageOk ? 'ok' : 'err'" style="text-align:center">{{ message }}</p>

  <div class="grid">
    <section class="card">
      <h2>Status</h2>
      <template v-if="status">
        <p>Broker: <strong>{{ status.provider }}</strong></p>
        <p>Can trade: {{ status.can_trade ? 'yes' : 'no' }}</p>
        <p v-if="status.trade_message" class="muted">{{ status.trade_message }}</p>
        <p>Auto: {{ status.auto_trading ? 'on' : 'off' }}</p>
      </template>
      <p v-else class="muted">Enter API key + Telegram ID</p>
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
        <p v-else class="muted">No plan — /start in Telegram</p>
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
        <thead><tr><th>ID</th><th>Name</th><th>Username</th><th>Last seen</th></tr></thead>
        <tbody>
          <tr v-for="u in recentUsers" :key="u.telegram_id">
            <td>{{ u.telegram_id }}</td>
            <td>{{ u.first_name }} {{ u.last_name }}</td>
            <td>@{{ u.username }}</td>
            <td>{{ new Date(u.last_seen_at).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="card wide">
      <h2>Broker connection</h2>
      <p class="muted">Mock works now. Other brokers appear when adapters are added.</p>
      <div class="field">
        <label>Broker</label>
        <select v-model="provider" @change="onProviderChange">
          <option v-for="b in brokers" :key="b.id" :value="b.id">
            {{ b.name }} ({{ b.status }})
          </option>
        </select>
      </div>
      <div v-for="f in dynamicFields" :key="f.key" class="field">
        <label>{{ f.label }}</label>
        <input :data-cred="f.key" :type="f.type === 'password' ? 'password' : 'text'" :placeholder="f.placeholder" />
      </div>
      <div class="field">
        <label>Label</label>
        <input v-model="brokerLabel" placeholder="My demo account" />
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
