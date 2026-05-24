<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { api, API } from '../api'

const props = defineProps({
  symbols: { type: Array, default: () => ['EURUSD', 'BTCUSD'] },
})

const symbol = ref('EURUSD')
const loading = ref(false)
const error = ref('')
const report = ref(null)
const catalog = ref([])

const verdictClass = computed(() => {
  const v = report.value?.report?.verdict
  if (v === 'pass') return 'verdict-pass'
  if (v === 'warn') return 'verdict-warn'
  return 'verdict-fail'
})

async function loadCatalog() {
  try {
    const res = await fetch(`${API}/filters/catalog`)
    const data = await res.json()
    catalog.value = data.filters || []
  } catch {
    catalog.value = []
  }
}

async function runReport() {
  loading.value = true
  error.value = ''
  try {
    const data = await api(`/filters/report?symbol=${encodeURIComponent(symbol.value)}`)
    report.value = data
  } catch (e) {
    error.value = e.message
    report.value = null
  } finally {
    loading.value = false
  }
}

function statusIcon(status) {
  if (status === 'pass') return '✓'
  if (status === 'warn') return '◐'
  if (status === 'skip') return '—'
  return '✗'
}

function statusClass(status) {
  return `step-${status || 'fail'}`
}

onMounted(() => {
  if (props.symbols?.length) {
    symbol.value = props.symbols[0]
  }
  loadCatalog()
  runReport()
})

watch(symbol, runReport)
</script>

<template>
  <section class="card wide filter-stack card-bull">
    <div class="filter-head">
      <div>
        <p class="section-eyebrow">Transparency</p>
        <h2 class="section-title">Filter stack</h2>
        <p class="section-lead filter-lead">
          See exactly which gates a symbol passes before broadcast or auto-trade — same pipeline as production.
        </p>
      </div>
      <div class="filter-controls">
        <label class="sym-label">
          <span>Symbol</span>
          <select v-model="symbol" class="sym-select">
            <option v-for="s in symbols" :key="s" :value="s">{{ s }}</option>
          </select>
        </label>
        <button type="button" class="btn-secondary" :disabled="loading" @click="runReport">
          {{ loading ? 'Running…' : 'Refresh' }}
        </button>
      </div>
    </div>

    <p v-if="error" class="err">{{ error }}</p>

    <div v-if="report?.report" class="filter-body">
      <div class="verdict-bar" :class="verdictClass">
        <div>
          <span class="verdict-label">Verdict</span>
          <strong class="verdict-text">{{ report.report.verdict?.toUpperCase() }}</strong>
          <span v-if="report.report.qualified" class="qual-badge">Qualified</span>
        </div>
        <p class="verdict-summary">{{ report.report.summary }}</p>
        <p class="verdict-meta muted">
          {{ report.report.data_source }}
          <span v-if="report.report.trend"> · {{ report.report.trend }}</span>
          <span v-if="report.report.live_ready"> · live ready</span>
          <span v-else-if="report.report.bar_count"> · {{ report.report.bar_count }}/{{ report.report.min_bars }} bars</span>
        </p>
      </div>

      <div v-if="report.report.signal_side" class="signal-chip">
        <span>{{ report.report.signal_side }}</span>
        <span v-if="report.report.strength">Strength {{ Math.round(report.report.strength * 100) }}%</span>
        <span v-if="report.report.risk_reward">R:R {{ report.report.risk_reward?.toFixed(2) }}</span>
      </div>

      <div v-for="layer in report.report.layers" :key="layer.id" class="filter-layer">
        <h3 class="layer-title">{{ layer.title }}</h3>
        <ul class="step-list">
          <li v-for="step in layer.steps" :key="step.id" :class="statusClass(step.status)">
            <span class="step-icon" aria-hidden="true">{{ statusIcon(step.status) }}</span>
            <div class="step-body">
              <span class="step-name">{{ step.name }}</span>
              <span class="step-msg">{{ step.message }}</span>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <details v-if="catalog.length" class="catalog-details">
      <summary>Filter reference ({{ catalog.length }} gates)</summary>
      <ul class="catalog-list">
        <li v-for="f in catalog" :key="f.id">
          <strong>{{ f.name }}</strong>
          <span class="cat-tag">{{ f.category }}</span>
          <p class="muted">{{ f.description }}</p>
          <p v-if="f.threshold" class="thresh">Threshold: {{ f.threshold }}</p>
        </li>
      </ul>
    </details>
  </section>
</template>

<style scoped>
.filter-stack {
  margin-top: 0.5rem;
}

.filter-head {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.filter-lead {
  margin-bottom: 0;
  max-width: 36rem;
}

.filter-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: flex-end;
}

.sym-label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.75rem;
  color: var(--muted);
}

.sym-select {
  min-height: 40px;
  padding: 0.4rem 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  font-family: inherit;
}

.verdict-bar {
  padding: 1rem 1.15rem;
  border-radius: 10px;
  margin-bottom: 1.25rem;
  border: 1px solid var(--border);
}

.verdict-pass {
  background: var(--brand-muted);
  border-color: rgba(61, 255, 122, 0.25);
}

.verdict-fail {
  background: rgba(248, 113, 113, 0.08);
  border-color: rgba(248, 113, 113, 0.25);
}

.verdict-warn {
  background: var(--warn-bg);
  border-color: var(--warn-border);
}

.verdict-label {
  font-size: 0.65rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
  margin-right: 0.5rem;
}

.verdict-text {
  font-size: 1.1rem;
}

.qual-badge {
  margin-left: 0.5rem;
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: var(--brand);
  color: var(--on-brand);
  font-weight: 700;
}

.verdict-summary {
  margin: 0.5rem 0 0.25rem;
  font-size: 0.95rem;
}

.verdict-meta {
  margin: 0;
  font-size: 0.8rem;
}

.signal-chip {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.25rem;
  padding: 0.65rem 1rem;
  margin-bottom: 1.25rem;
  border-radius: 8px;
  background: var(--surface);
  border: 1px solid var(--border);
  font-size: 0.85rem;
  font-weight: 600;
}

.filter-layer {
  margin-bottom: 1.25rem;
}

.layer-title {
  margin: 0 0 0.65rem;
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.step-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.step-list li {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem 0.85rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
}

.step-icon {
  flex-shrink: 0;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  font-weight: 800;
  font-size: 0.85rem;
}

.step-pass .step-icon {
  background: var(--brand-muted);
  color: var(--brand);
}

.step-fail .step-icon {
  background: rgba(248, 113, 113, 0.15);
  color: #f87171;
}

.step-warn .step-icon {
  background: var(--warn-bg);
  color: var(--warn);
}

.step-body {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.step-name {
  font-weight: 600;
  font-size: 0.9rem;
}

.step-msg {
  font-size: 0.8rem;
  color: var(--muted);
  line-height: 1.45;
}

.catalog-details {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}

.catalog-details summary {
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-soft);
}

.catalog-list {
  list-style: none;
  margin: 1rem 0 0;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}

@media (min-width: 720px) {
  .catalog-list {
    grid-template-columns: 1fr 1fr;
  }
}

.catalog-list li {
  padding: 0.85rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--chart-bg);
}

.cat-tag {
  margin-left: 0.35rem;
  font-size: 0.65rem;
  text-transform: uppercase;
  color: var(--brand);
}

.catalog-list p {
  margin: 0.35rem 0 0;
  font-size: 0.8rem;
}

.thresh {
  font-size: 0.75rem !important;
  color: var(--muted) !important;
}
</style>
