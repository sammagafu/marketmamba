<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from '../api'

const props = defineProps({
  canTrade: { type: Boolean, default: true },
})

const emit = defineEmits(['message'])

const loading = ref(true)
const saving = ref(false)
const assetGroups = ref([])
const signalTypes = ref({ forex: true, indexes: true, crypto: true })
const rows = ref([])

const activeGroupIds = computed(() => {
  const ids = []
  if (signalTypes.value.forex) ids.push('forex')
  if (signalTypes.value.indexes) ids.push('indexes')
  if (signalTypes.value.crypto) ids.push('crypto')
  return ids
})

const visibleRows = computed(() => {
  const enabled = new Set(activeGroupIds.value)
  const symToGroup = {}
  for (const g of assetGroups.value) {
    for (const sym of g.symbols || []) {
      symToGroup[sym] = g.id
    }
  }
  return rows.value.filter((r) => {
    const g = symToGroup[r.symbol]
    return !g || enabled.has(g)
  })
})

const rowsByGroup = computed(() => {
  const symToGroup = {}
  for (const g of assetGroups.value) {
    for (const sym of g.symbols || []) {
      symToGroup[sym] = g.id
    }
  }
  const buckets = {}
  for (const g of assetGroups.value) {
    if (!activeGroupIds.value.includes(g.id)) continue
    buckets[g.id] = { group: g, rows: [] }
  }
  for (const row of visibleRows.value) {
    const gid = symToGroup[row.symbol] || 'other'
    if (!buckets[gid]) continue
    buckets[gid].rows.push(row)
  }
  return Object.values(buckets).filter((b) => b.rows.length > 0)
})

function syncTypesFromGroups() {
  const map = { forex: false, indexes: false, crypto: false }
  for (const g of assetGroups.value) {
    if (g.enabled && map[g.id] !== undefined) map[g.id] = true
  }
  if (map.forex || map.indexes || map.crypto) {
    signalTypes.value = map
  }
}

function toggleType(id) {
  if (!props.canTrade) return
  const next = { ...signalTypes.value }
  next[id] = !next[id]
  if (!next.forex && !next.indexes && !next.crypto) {
    emit('message', { text: 'Keep at least one signal type enabled', ok: false })
    return
  }
  signalTypes.value = next
}

async function load() {
  loading.value = true
  try {
    const data = await api('/trading-pairs')
    assetGroups.value = data.asset_groups || []
    if (data.signal_types) {
      signalTypes.value = {
        forex: !!data.signal_types.forex,
        indexes: !!data.signal_types.indexes,
        crypto: !!data.signal_types.crypto,
      }
    } else {
      syncTypesFromGroups()
    }
    rows.value = (data.pairs || []).map((p) => ({
      symbol: p.symbol,
      receive_signals: !!p.receive_signals,
      auto_trade: !!p.auto_trade,
    }))
  } catch (e) {
    emit('message', { text: e.message, ok: false })
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!activeGroupIds.value.length) {
    emit('message', { text: 'Enable at least one signal type', ok: false })
    return
  }
  saving.value = true
  try {
    const data = await api('/trading-pairs', {
      method: 'PUT',
      body: {
        signal_types: { ...signalTypes.value },
        pairs: rows.value,
      },
    })
    assetGroups.value = data.asset_groups || assetGroups.value
    rows.value = (data.pairs || []).map((p) => ({
      symbol: p.symbol,
      receive_signals: !!p.receive_signals,
      auto_trade: !!p.auto_trade,
    }))
    emit('message', {
      text: `Saved — types: ${formatTypes()} · signals: ${(data.signal_symbols || []).join(', ') || 'none'}`,
      ok: true,
    })
  } catch (e) {
    emit('message', { text: e.message, ok: false })
  } finally {
    saving.value = false
  }
}

function formatTypes() {
  const parts = []
  if (signalTypes.value.forex) parts.push('Forex')
  if (signalTypes.value.indexes) parts.push('Indexes')
  if (signalTypes.value.crypto) parts.push('Crypto')
  return parts.join(', ') || 'none'
}

function enableAllVisible() {
  const visible = new Set(visibleRows.value.map((r) => r.symbol))
  rows.value = rows.value.map((r) =>
    visible.has(r.symbol) ? { ...r, receive_signals: true, auto_trade: true } : r,
  )
}

const typeCards = [
  { id: 'forex', title: 'Forex', hint: 'Majors & crosses', icon: 'FX' },
  { id: 'indexes', title: 'Indexes', hint: 'US 500, NAS, vol indices', icon: 'IX' },
  { id: 'crypto', title: 'Bitcoin & crypto', hint: 'BTC/USD, ETH/USD', icon: '₿' },
]

onMounted(load)
</script>

<template>
  <section class="card card-bull wide pairs-card">
    <div class="pairs-head">
      <div>
        <p class="section-eyebrow">Automation scope</p>
        <h2 class="section-title section-title-sm">Signals &amp; pairs</h2>
      </div>
      <button type="button" class="btn-secondary" :disabled="loading" @click="load">Refresh</button>
    </div>

    <p class="muted pairs-lede">
      Choose the <strong>types of signals</strong> you want (forex, indexes, bitcoin/crypto), then fine-tune
      individual pairs for Telegram alerts and <code>/autostart</code>.
    </p>

    <p v-if="loading" class="muted">Loading…</p>
    <template v-else>
      <p class="section-eyebrow types-label">Signal types</p>
      <div class="type-grid" role="group" aria-label="Signal asset types">
        <button
          v-for="t in typeCards"
          :key="t.id"
          type="button"
          class="type-card"
          :class="{ active: signalTypes[t.id], disabled: !canTrade }"
          :disabled="!canTrade"
          :aria-pressed="signalTypes[t.id]"
          @click="toggleType(t.id)"
        >
          <span class="type-icon" aria-hidden="true">{{ t.icon }}</span>
          <span class="type-body">
            <strong class="type-title">{{ t.title }}</strong>
            <span class="type-hint">{{ t.hint }}</span>
          </span>
          <span class="type-check" aria-hidden="true">{{ signalTypes[t.id] ? 'On' : 'Off' }}</span>
        </button>
      </div>

      <p class="section-eyebrow pairs-label">Pairs in your types</p>
      <div v-for="bucket in rowsByGroup" :key="bucket.group.id" class="pair-group">
        <h3 class="pair-group-title">{{ bucket.group.label }}</h3>
        <div class="pairs-table-wrap">
          <table class="pairs-table">
            <thead>
              <tr>
                <th>Pair</th>
                <th>Signals</th>
                <th>Auto-trade</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in bucket.rows" :key="row.symbol">
                <td><strong>{{ row.symbol }}</strong></td>
                <td>
                  <label class="pair-check">
                    <input
                      v-model="row.receive_signals"
                      type="checkbox"
                      :disabled="!canTrade"
                    />
                    <span>Telegram</span>
                  </label>
                </td>
                <td>
                  <label class="pair-check">
                    <input v-model="row.auto_trade" type="checkbox" :disabled="!canTrade" />
                    <span>/autostart</span>
                  </label>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="pairs-actions">
        <button type="button" class="btn-secondary" :disabled="!canTrade" @click="enableAllVisible">
          Enable all visible
        </button>
        <button
          type="button"
          class="btn-primary"
          :disabled="!canTrade || saving"
          @click="save"
        >
          {{ saving ? 'Saving…' : 'Save preferences' }}
        </button>
      </div>
      <p class="muted pairs-hint">
        Telegram: <code>/signaltypes forex crypto</code> · <code>/pairs EURUSD</code> ·
        <code>/pairs</code> to view
      </p>
    </template>
  </section>
</template>

<style scoped>
.pairs-card.wide {
  grid-column: 1 / -1;
}
.pairs-head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}
.pairs-head h2 {
  margin: 0;
}
.pairs-lede {
  margin: 0 0 1.25rem;
  font-size: 0.9rem;
  line-height: 1.55;
}
.types-label,
.pairs-label {
  margin: 0 0 0.75rem;
}
.type-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1.75rem;
}
.type-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.9rem 1rem;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--surface);
  cursor: pointer;
  text-align: left;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.type-card:hover:not(:disabled) {
  border-color: var(--brand);
}
.type-card.active {
  border-color: var(--brand);
  box-shadow: 0 0 0 1px var(--brand), 0 0 16px var(--win-glow);
}
.type-card:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}
.type-icon {
  flex-shrink: 0;
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 800;
  background: var(--bg);
  color: var(--win-bright);
  border: 1px solid var(--border);
}
.type-card.active .type-icon {
  background: var(--brand);
  color: var(--on-brand);
  border-color: transparent;
}
.type-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.type-title {
  font-size: 0.9rem;
  color: var(--text);
}
.type-hint {
  font-size: 0.75rem;
  color: var(--muted);
  line-height: 1.35;
}
.type-check {
  flex-shrink: 0;
  font-size: 0.6875rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
}
.type-card.active .type-check {
  color: var(--brand);
}
.pair-group {
  margin-bottom: 1.25rem;
}
.pair-group-title {
  margin: 0 0 0.5rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--muted);
}
.pairs-table-wrap {
  overflow-x: auto;
  margin-bottom: 0.5rem;
}
.pairs-table {
  width: 100%;
  min-width: 280px;
}
.pair-check {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  cursor: pointer;
}
.pair-check input {
  width: 1.1rem;
  height: 1.1rem;
  min-height: auto;
  accent-color: var(--brand);
}
.pairs-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.pairs-hint {
  margin: 0.75rem 0 0;
  font-size: 0.8rem;
}
</style>
