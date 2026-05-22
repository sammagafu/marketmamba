<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api'

const props = defineProps({
  canTrade: { type: Boolean, default: true },
})

const emit = defineEmits(['message'])

const loading = ref(true)
const saving = ref(false)
const available = ref([])
const rows = ref([])

async function load() {
  loading.value = true
  try {
    const data = await api('/trading-pairs')
    available.value = data.available_symbols || []
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
  saving.value = true
  try {
    const data = await api('/trading-pairs', {
      method: 'PUT',
      body: { pairs: rows.value },
    })
    rows.value = (data.pairs || []).map((p) => ({
      symbol: p.symbol,
      receive_signals: !!p.receive_signals,
      auto_trade: !!p.auto_trade,
    }))
    emit('message', {
      text: `Pairs saved — signals: ${(data.signal_symbols || []).join(', ') || 'none'}`,
      ok: true,
    })
  } catch (e) {
    emit('message', { text: e.message, ok: false })
  } finally {
    saving.value = false
  }
}

function enableAll() {
  rows.value = rows.value.map((r) => ({
    ...r,
    receive_signals: true,
    auto_trade: true,
  }))
}

onMounted(load)
</script>

<template>
  <section class="card card-bull wide pairs-card">
    <div class="pairs-head">
      <h2>Trading pairs</h2>
      <button type="button" class="btn-secondary" :disabled="loading" @click="load">Refresh</button>
    </div>
    <p class="muted pairs-lede">
      Choose which pairs you receive signals for and which auto-trade when
      <code>/autostart</code> is on.
    </p>
    <p v-if="loading" class="muted">Loading…</p>
    <template v-else>
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
            <tr v-for="row in rows" :key="row.symbol">
              <td><strong>{{ row.symbol }}</strong></td>
              <td>
                <label class="pair-check">
                  <input
                    v-model="row.receive_signals"
                    type="checkbox"
                    :disabled="!canTrade"
                  />
                  <span>Telegram alerts</span>
                </label>
              </td>
              <td>
                <label class="pair-check">
                  <input
                    v-model="row.auto_trade"
                    type="checkbox"
                    :disabled="!canTrade"
                  />
                  <span>With /autostart</span>
                </label>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="pairs-actions">
        <button type="button" class="btn-secondary" :disabled="!canTrade" @click="enableAll">
          Enable all
        </button>
        <button
          type="button"
          class="btn-primary"
          :disabled="!canTrade || saving"
          @click="save"
        >
          {{ saving ? 'Saving…' : 'Save pairs' }}
        </button>
      </div>
      <p class="muted pairs-hint">
        Telegram: <code>/pairs EURUSD BTCUSD</code> · <code>/pairs</code> to view
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
  margin: 0 0 1rem;
  font-size: 0.9rem;
}
.pairs-table-wrap {
  overflow-x: auto;
  margin-bottom: 1rem;
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
}
.pairs-hint {
  margin: 0.75rem 0 0;
  font-size: 0.8rem;
}
</style>
