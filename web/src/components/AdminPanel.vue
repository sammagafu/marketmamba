<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  stats: { type: Object, required: true },
  users: { type: Array, default: () => [] },
  trades: { type: Array, default: () => [] },
  activateTarget: { type: String, default: '' },
  activateDays: { type: Number, default: 30 },
})

const emit = defineEmits([
  'broadcast-signal',
  'activate',
  'block-user',
  'revoke-user',
  'update:activateTarget',
  'update:activateDays',
])

const activeTab = ref('overview')
const tradeUserFilter = ref(null)

const userMap = computed(() => {
  const m = new Map()
  for (const u of props.users) {
    m.set(u.telegram_id, u)
  }
  return m
})

const filteredTrades = computed(() => {
  if (tradeUserFilter.value == null) return props.trades
  return props.trades.filter((t) => t.user_id === tradeUserFilter.value)
})

const filterLabel = computed(() => {
  if (tradeUserFilter.value == null) return 'All users'
  const u = userMap.value.get(tradeUserFilter.value)
  if (!u) return `User ${tradeUserFilter.value}`
  const name = [u.first_name, u.last_name].filter(Boolean).join(' ')
  return name || u.username || String(u.telegram_id)
})

function userLabel(id) {
  const u = userMap.value.get(id)
  if (!u) return String(id)
  const name = [u.first_name, u.last_name].filter(Boolean).join(' ')
  if (name) return name
  if (u.username) return `@${u.username}`
  return String(id)
}

function viewUserTrades(telegramId) {
  tradeUserFilter.value = telegramId
  activeTab.value = 'trades'
}

function clearTradeFilter() {
  tradeUserFilter.value = null
}

function quickActivate(telegramId) {
  emit('update:activateTarget', String(telegramId))
  activeTab.value = 'overview'
}

function fmtProfit(t) {
  if (t.profit == null) return '—'
  const n = Number(t.profit)
  const sign = n >= 0 ? '+' : ''
  return `${sign}$${n.toFixed(2)}`
}

function fmtTime(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}
</script>

<template>
  <section class="admin-panel wide">
    <div class="admin-head">
      <h2>Admin command center</h2>
      <p class="muted">Platform overview · manage users · inspect trades per client</p>
    </div>

    <nav class="admin-tabs">
      <button type="button" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">
        Overview
      </button>
      <button type="button" :class="{ active: activeTab === 'users' }" @click="activeTab = 'users'">
        Users ({{ users.length }})
      </button>
      <button type="button" :class="{ active: activeTab === 'trades' }" @click="activeTab = 'trades'">
        Trades ({{ trades.length }})
      </button>
    </nav>

    <template v-if="activeTab === 'overview'">
      <div class="admin-metrics">
        <div class="metric">
          <strong>{{ stats.total_users }}</strong>
          <span>Users</span>
        </div>
        <div class="metric">
          <strong>{{ stats.active_subscriptions }}</strong>
          <span>Active subs</span>
        </div>
        <div class="metric">
          <strong>{{ stats.auto_trading_users }}</strong>
          <span>Auto trading</span>
        </div>
        <div class="metric">
          <strong>{{ stats.new_users_last_7_days }}</strong>
          <span>New (7d)</span>
        </div>
        <div class="metric highlight">
          <strong>{{ stats.total_trades ?? 0 }}</strong>
          <span>Total trades</span>
        </div>
        <div class="metric">
          <strong>{{ stats.open_trades ?? 0 }}</strong>
          <span>Open</span>
        </div>
        <div class="metric">
          <strong>{{ stats.trades_last_24h ?? 0 }}</strong>
          <span>Last 24h</span>
        </div>
        <div class="metric" :class="{ profit: (stats.net_profit_closed ?? 0) >= 0, loss: (stats.net_profit_closed ?? 0) < 0 }">
          <strong>${{ Number(stats.net_profit_closed ?? 0).toFixed(2) }}</strong>
          <span>Net P/L (closed)</span>
        </div>
      </div>

      <div class="admin-toolbar">
        <button type="button" class="btn-primary" @click="$emit('broadcast-signal')">
          Broadcast signal
        </button>
        <input
          :value="activateTarget"
          placeholder="Telegram user ID"
          @input="$emit('update:activateTarget', $event.target.value)"
        />
        <input
          :value="activateDays"
          type="number"
          min="1"
          class="days-input"
          @input="$emit('update:activateDays', Number($event.target.value))"
        />
        <button type="button" class="btn-secondary" @click="$emit('activate')">Activate sub</button>
      </div>
    </template>

    <template v-else-if="activeTab === 'users'">
      <div class="admin-block">
        <h3>Recent users</h3>
        <p class="muted block-hint">Click <strong>Trades</strong> to see that client’s log only.</p>
        <div class="table-wrap">
          <table v-if="users.length">
            <thead>
              <tr>
                <th>Telegram ID</th><th>Name</th><th>Status</th><th>Last seen</th><th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.telegram_id">
                <td><code>{{ u.telegram_id }}</code></td>
                <td>
                  {{ u.first_name }} {{ u.last_name }}
                  <span v-if="u.username" class="muted">@{{ u.username }}</span>
                </td>
                <td>
                  <span :class="u.is_blocked ? 'badge bad' : 'badge ok'">{{ u.is_blocked ? 'blocked' : 'active' }}</span>
                </td>
                <td>{{ fmtTime(u.last_seen_at) }}</td>
                <td class="admin-actions">
                  <button type="button" class="btn-link" @click="viewUserTrades(u.telegram_id)">Trades</button>
                  <button type="button" class="btn-link" @click="quickActivate(u.telegram_id)">Activate</button>
                  <button type="button" class="btn-secondary" @click="$emit('block-user', u.telegram_id, !u.is_blocked)">
                    {{ u.is_blocked ? 'Unblock' : 'Block' }}
                  </button>
                  <button type="button" class="btn-secondary" @click="$emit('revoke-user', u.telegram_id)">Revoke</button>
                </td>
              </tr>
            </tbody>
          </table>
          <p v-else class="muted">No users yet</p>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="admin-block admin-block-trades">
        <div class="trades-head">
          <h3>Platform trade log</h3>
          <div class="filter-bar">
            <span class="filter-label">Showing: <strong>{{ filterLabel }}</strong></span>
            <button v-if="tradeUserFilter != null" type="button" class="btn-link" @click="clearTradeFilter">
              Show all
            </button>
          </div>
        </div>
        <div class="table-wrap table-wrap-tall">
          <table v-if="filteredTrades.length">
            <thead>
              <tr>
                <th>Time</th><th>Client</th><th>Pair</th><th>Side</th><th>Qty</th><th>Entry</th><th>Status</th><th>P/L</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in filteredTrades" :key="t.id">
                <td>{{ fmtTime(t.created_at) }}</td>
                <td>
                  <button type="button" class="btn-link client-link" @click="viewUserTrades(t.user_id)">
                    {{ userLabel(t.user_id) }}
                  </button>
                  <span class="muted small"><br /><code>{{ t.user_id }}</code></span>
                </td>
                <td><strong>{{ t.symbol }}</strong></td>
                <td :class="t.type === 'BUY' ? 'buy' : 'sell'">{{ t.type }}</td>
                <td>{{ t.quantity }}</td>
                <td>{{ Number(t.entry_price).toFixed(5) }}</td>
                <td>
                  <span class="badge" :class="t.status === 'OPEN' ? 'open' : 'closed'">{{ t.status }}</span>
                  <span v-if="t.closure_reason" class="muted small"> {{ t.closure_reason }}</span>
                </td>
                <td :class="{ profit: t.profit > 0, loss: t.profit < 0 }">{{ fmtProfit(t) }}</td>
              </tr>
            </tbody>
          </table>
          <p v-else class="muted empty-trades">
            <template v-if="tradeUserFilter != null">No trades for this client yet.</template>
            <template v-else>
              No trades in database yet.<br />
              Users: connect broker → <code>/open</code> or <code>/autostart</code> on Telegram.
            </template>
          </p>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.admin-panel {
  border: 1px solid var(--border);
  border-top: 2px solid var(--brand);
  background: var(--card);
  border-radius: 16px;
  padding: clamp(1rem, 3vw, 1.5rem);
  box-shadow: inset 0 0 60px var(--brand-muted);
  width: 100%;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  grid-column: 1 / -1;
}

.admin-head h2 { margin: 0 0 0.35rem; }
.admin-head p { margin: 0 0 1rem; }

.admin-tabs {
  display: flex;
  gap: 0.35rem;
  flex-wrap: wrap;
  margin-bottom: 1.25rem;
}

.admin-tabs button {
  font-size: 0.85rem;
  padding: 0.45rem 0.85rem;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--chart-bg);
  color: var(--muted);
  cursor: pointer;
  font-family: inherit;
  font-weight: 600;
}

.admin-tabs button.active {
  background: var(--brand-muted);
  color: var(--brand);
  border-color: var(--brand-dim);
}

.admin-metrics {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 100px), 1fr));
  gap: 0.65rem;
  margin-bottom: 1.25rem;
}

.metric {
  padding: 0.65rem 0.75rem;
  border-radius: 10px;
  background: var(--chart-bg);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}
.metric strong { font-size: 1.2rem; }
.metric span { font-size: 0.7rem; color: var(--muted); text-transform: uppercase; letter-spacing: 0.04em; }
.metric.highlight { border-color: var(--win); }
.metric.profit strong { color: var(--win-bright); }
.metric.loss strong { color: var(--loss-text); }

.admin-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.admin-toolbar input {
  flex: 1 1 120px;
  min-width: 0;
  max-width: 100%;
}

@media (min-width: 560px) {
  .admin-toolbar input { max-width: 200px; }
}
.days-input { max-width: 80px !important; flex: 0 !important; }

.admin-block h3 { margin: 0 0 0.35rem; font-size: 1rem; }
.block-hint { margin: 0 0 0.75rem; font-size: 0.85rem; }

.trades-head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.trades-head h3 { margin: 0; }

.filter-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
}

.filter-label strong { color: var(--brand); }

.table-wrap { overflow-x: auto; }
.table-wrap-tall { max-height: 480px; overflow-y: auto; }

.badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 6px;
  font-weight: 600;
}
.badge.ok, .badge.open { background: var(--win-dim); color: var(--win-bright); }
.badge.bad { background: rgba(120, 123, 134, 0.2); color: var(--muted); }
.badge.closed { background: rgba(120, 123, 134, 0.15); color: var(--muted); }

.buy { color: var(--win-bright); font-weight: 600; }
.sell { color: var(--down); font-weight: 600; }
.profit { color: var(--win-bright); }
.loss { color: var(--loss-text); }
.small { font-size: 0.75rem; }

.admin-actions {
  display: flex;
  gap: 0.35rem;
  flex-wrap: wrap;
  align-items: center;
}
.admin-actions button { padding: 0.3rem 0.5rem; font-size: 0.75rem; }

.btn-link {
  background: none;
  border: none;
  color: var(--brand);
  cursor: pointer;
  font-family: inherit;
  font-size: 0.8rem;
  font-weight: 600;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.client-link {
  text-align: left;
}

.empty-trades { padding: 2rem 1rem; text-align: center; line-height: 1.6; }
</style>
