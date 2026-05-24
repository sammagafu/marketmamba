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

const activeTypeCount = computed(() => activeGroupIds.value.length)

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

const signalCount = computed(
  () => visibleRows.value.filter((r) => r.receive_signals).length,
)

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

const typeCards = [
  { id: 'forex', title: 'Forex', hint: 'Majors & crosses', icon: 'FX', accent: 'fx' },
  { id: 'indexes', title: 'Indexes', hint: 'US 500, NAS, volatility', icon: 'IX', accent: 'ix' },
  { id: 'crypto', title: 'Bitcoin & crypto', hint: 'BTC/USD, ETH/USD', icon: '₿', accent: 'cr' },
]

const groupAccent = { forex: 'fx', indexes: 'ix', crypto: 'cr' }

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
      text: `Saved — ${formatTypes()} · ${(data.signal_symbols || []).length} pairs with signals`,
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

onMounted(load)
</script>

<template>
  <section class="card card-bull wide pairs-card">
    <div class="pairs-head">
      <div>
        <p class="section-eyebrow">Automation scope</p>
        <h2 class="section-title section-title-sm">Signals &amp; pairs</h2>
      </div>
      <button type="button" class="btn-secondary btn-sm" :disabled="loading" @click="load">
        Refresh
      </button>
    </div>

    <p class="muted pairs-lede">
      Pick <strong>forex</strong>, <strong>indexes</strong>, or <strong>bitcoin/crypto</strong>, then choose which
      symbols get Telegram alerts and auto-trade with <code>/autostart</code>.
    </p>

    <div v-if="loading" class="pairs-loading" aria-busy="true">
      <div class="skel skel-wide" />
      <div class="skel-row">
        <div class="skel skel-card" />
        <div class="skel skel-card" />
        <div class="skel skel-card" />
      </div>
    </div>

    <template v-else>
      <div class="pairs-summary" role="status">
        <span class="summary-chip">{{ activeTypeCount }} type{{ activeTypeCount === 1 ? '' : 's' }} on</span>
        <span class="summary-chip">{{ visibleRows.length }} pairs visible</span>
        <span class="summary-chip summary-chip-accent">{{ signalCount }} receiving signals</span>
      </div>

      <p class="section-eyebrow types-label">1 · Signal types</p>
      <div class="type-grid" role="group" aria-label="Signal asset types">
        <button
          v-for="t in typeCards"
          :key="t.id"
          type="button"
          class="type-card"
          :class="[`type-card--${t.accent}`, { active: signalTypes[t.id] }]"
          :disabled="!canTrade"
          :aria-pressed="signalTypes[t.id]"
          @click="toggleType(t.id)"
        >
          <span class="type-icon" aria-hidden="true">{{ t.icon }}</span>
          <span class="type-body">
            <strong class="type-title">{{ t.title }}</strong>
            <span class="type-hint">{{ t.hint }}</span>
          </span>
          <span class="type-pill">{{ signalTypes[t.id] ? 'On' : 'Off' }}</span>
        </button>
      </div>

      <p class="section-eyebrow pairs-label">2 · Pairs</p>

      <p v-if="!rowsByGroup.length" class="pairs-empty">
        Enable at least one signal type above to see available pairs.
      </p>

      <div
        v-for="bucket in rowsByGroup"
        :key="bucket.group.id"
        class="pair-group"
        :class="`pair-group--${groupAccent[bucket.group.id] || 'fx'}`"
      >
        <header class="pair-group-head">
          <h3 class="pair-group-title">{{ bucket.group.label }}</h3>
          <span class="pair-group-count">{{ bucket.rows.length }} pairs</span>
        </header>

        <div class="pair-cards-mobile">
          <div v-for="row in bucket.rows" :key="row.symbol" class="pair-row-card">
            <strong class="pair-symbol">{{ row.symbol }}</strong>
            <div class="pair-toggles">
              <label class="toggle-pill" :class="{ on: row.receive_signals }">
                <input v-model="row.receive_signals" type="checkbox" :disabled="!canTrade" />
                <span>Signals</span>
              </label>
              <label class="toggle-pill" :class="{ on: row.auto_trade }">
                <input v-model="row.auto_trade" type="checkbox" :disabled="!canTrade" />
                <span>Auto</span>
              </label>
            </div>
          </div>
        </div>

        <div class="pairs-table-wrap pair-table-desktop table-wrap">
          <table class="pairs-table">
            <thead>
              <tr>
                <th>Pair</th>
                <th>Signals</th>
                <th>Auto-trade</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in bucket.rows" :key="`t-${row.symbol}`">
                <td><strong>{{ row.symbol }}</strong></td>
                <td>
                  <label class="pair-check">
                    <input v-model="row.receive_signals" type="checkbox" :disabled="!canTrade" />
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
        <button type="button" class="btn-primary" :disabled="!canTrade || saving" @click="save">
          {{ saving ? 'Saving…' : 'Save preferences' }}
        </button>
      </div>
      <p class="muted pairs-hint">
        Telegram: <code>/signaltypes forex crypto</code> · <code>/pairs EURUSD</code>
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
.btn-sm {
  min-height: 2.25rem;
  padding: 0.4rem 0.85rem;
  font-size: 0.8125rem;
}
.pairs-lede {
  margin: 0 0 1rem;
  font-size: 0.9rem;
  line-height: 1.55;
}
.pairs-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}
.summary-chip {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  color: var(--muted);
}
.summary-chip-accent {
  color: var(--brand);
  border-color: var(--brand-muted);
  background: var(--brand-soft);
}
.types-label,
.pairs-label {
  margin: 0 0 0.75rem;
}
.type-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1.75rem;
}
.type-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.95rem 1rem;
  border-radius: 14px;
  border: 1px solid var(--border);
  background: var(--surface-raised);
  cursor: pointer;
  text-align: left;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;
}
.type-card:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: var(--border-strong);
}
.type-card.active {
  box-shadow: 0 0 20px var(--win-glow);
}
.type-card--fx.active { border-color: #3dff7a; }
.type-card--ix.active { border-color: #5eb3ff; }
.type-card--cr.active { border-color: #fbbf24; }
.type-card:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.type-icon {
  flex-shrink: 0;
  width: 2.35rem;
  height: 2.35rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  font-size: 0.8rem;
  font-weight: 800;
  border: 1px solid var(--border);
  background: var(--bg);
}
.type-card--fx.active .type-icon {
  background: rgba(61, 255, 122, 0.15);
  color: var(--brand);
  border-color: transparent;
}
.type-card--ix.active .type-icon {
  background: rgba(94, 179, 255, 0.15);
  color: #5eb3ff;
  border-color: transparent;
}
.type-card--cr.active .type-icon {
  background: rgba(251, 191, 36, 0.15);
  color: #fbbf24;
  border-color: transparent;
}
.type-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.12rem;
}
.type-title {
  font-size: 0.9rem;
  color: var(--text);
}
.type-hint {
  font-size: 0.72rem;
  color: var(--muted);
  line-height: 1.35;
}
.type-pill {
  flex-shrink: 0;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 0.25rem 0.45rem;
  border-radius: 6px;
  background: var(--bg);
  color: var(--muted);
}
.type-card.active .type-pill {
  background: var(--brand);
  color: var(--on-brand);
}
.pair-group {
  margin-bottom: 1.5rem;
  padding: 1rem;
  border-radius: 14px;
  border: 1px solid var(--border);
  background: var(--surface);
}
.pair-group--fx { border-left: 3px solid #3dff7a; }
.pair-group--ix { border-left: 3px solid #5eb3ff; }
.pair-group--cr { border-left: 3px solid #fbbf24; }
.pair-group-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.pair-group-title {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 700;
}
.pair-group-count {
  font-size: 0.75rem;
  color: var(--muted);
}
.pair-cards-mobile {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.pair-table-desktop {
  display: none;
}
@media (min-width: 720px) {
  .pair-cards-mobile {
    display: none;
  }
  .pair-table-desktop {
    display: block;
  }
}
.pair-row-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem 0.85rem;
  border-radius: 10px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
}
.pair-symbol {
  font-size: 0.9rem;
  font-variant-numeric: tabular-nums;
}
.pair-toggles {
  display: flex;
  gap: 0.4rem;
}
.toggle-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  border: 1px solid var(--border);
  color: var(--muted);
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}
.toggle-pill input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.toggle-pill.on {
  border-color: var(--brand);
  background: var(--brand-soft);
  color: var(--brand);
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
.pairs-empty {
  margin: 0 0 1.25rem;
  padding: 1rem;
  text-align: center;
  font-size: 0.875rem;
  color: var(--muted);
  border-radius: 10px;
  border: 1px dashed var(--border-strong);
}
.pairs-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.25rem;
}
.pairs-hint {
  margin: 0.75rem 0 0;
  font-size: 0.8rem;
}
.pairs-loading {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.skel {
  border-radius: 10px;
  background: linear-gradient(
    90deg,
    var(--surface) 0%,
    var(--surface-raised) 50%,
    var(--surface) 100%
  );
  background-size: 200% 100%;
  animation: shimmer 1.2s ease-in-out infinite;
}
.skel-wide {
  height: 2.5rem;
}
.skel-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}
.skel-card {
  height: 4.5rem;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}
</style>
