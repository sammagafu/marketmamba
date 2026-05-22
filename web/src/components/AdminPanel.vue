<script setup>
defineProps({
  stats: { type: Object, required: true },
  users: { type: Array, default: () => [] },
  trades: { type: Array, default: () => [] },
  activateTarget: { type: String, default: '' },
  activateDays: { type: Number, default: 30 },
})

defineEmits([
  'broadcast-signal',
  'activate',
  'block-user',
  'revoke-user',
  'update:activateTarget',
  'update:activateDays',
])

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
      <h2>⚡ Admin command center</h2>
      <p class="muted">Users, subscriptions, signals &amp; platform-wide trade log</p>
    </div>

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
        📡 Broadcast signal
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

    <div class="admin-grid">
      <div class="admin-block">
        <h3>Recent users</h3>
        <div class="table-wrap">
          <table v-if="users.length">
            <thead>
              <tr>
                <th>ID</th><th>Name</th><th>Status</th><th>Last seen</th><th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.telegram_id">
                <td><code>{{ u.telegram_id }}</code></td>
                <td>{{ u.first_name }} {{ u.last_name }} <span class="muted">@{{ u.username }}</span></td>
                <td>
                  <span :class="u.is_blocked ? 'badge bad' : 'badge ok'">{{ u.is_blocked ? 'blocked' : 'active' }}</span>
                </td>
                <td>{{ fmtTime(u.last_seen_at) }}</td>
                <td class="admin-actions">
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

      <div class="admin-block admin-block-trades">
        <h3>Platform trade log <span class="muted">({{ trades.length }})</span></h3>
        <div class="table-wrap table-wrap-tall">
          <table v-if="trades.length">
            <thead>
              <tr>
                <th>Time</th><th>User</th><th>Pair</th><th>Side</th><th>Qty</th><th>Entry</th><th>Status</th><th>P/L</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in trades" :key="t.id">
                <td>{{ fmtTime(t.created_at) }}</td>
                <td><code>{{ t.user_id }}</code></td>
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
            No trades in database yet.<br />
            Users: connect broker → <code>/open</code> or <code>/autostart</code> on Telegram.
          </p>
        </div>
      </div>
    </div>
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
}

.admin-head h2 { margin: 0 0 0.35rem; }
.admin-head p { margin: 0 0 1.25rem; }

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
  margin-bottom: 1.25rem;
}
.admin-toolbar input {
  flex: 1 1 120px;
  min-width: 0;
  max-width: 100%;
}

@media (min-width: 560px) {
  .admin-toolbar input {
    max-width: 200px;
  }
}
.days-input { max-width: 80px !important; flex: 0 !important; }

.admin-grid {
  display: grid;
  gap: 1.25rem;
  grid-template-columns: 1fr;
}
@media (min-width: 900px) {
  .admin-grid { grid-template-columns: 1fr 1.2fr; }
}

.admin-block h3 { margin: 0 0 0.75rem; font-size: 1rem; }
.table-wrap { overflow-x: auto; }
.table-wrap-tall { max-height: 420px; overflow-y: auto; }

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

.admin-actions { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.admin-actions button { padding: 0.3rem 0.5rem; font-size: 0.75rem; }

.empty-trades { padding: 2rem 1rem; text-align: center; line-height: 1.6; }
</style>
