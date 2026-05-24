<script setup>
import { computed, ref, watch } from 'vue'
import { api } from '../api'

const props = defineProps({
  brands: { type: Array, default: () => [] },
  brokers: { type: Array, default: () => [] },
  connection: { type: Object, default: null },
  publicSiteUrl: { type: String, default: '' },
  metaapiSharedToken: { type: Boolean, default: false },
})

const emit = defineEmits(['saved', 'message'])

const step = ref(1)
const selectedBrandId = ref('mock')
const credentials = ref({})
const label = ref('')
const testing = ref(false)
const saving = ref(false)
const testBalance = ref(null)

const selectedBrand = computed(() =>
  props.brands.find((b) => b.id === selectedBrandId.value),
)

const liveBrands = computed(() =>
  props.brands.filter((b) => b.status === 'live'),
)

const metaapiBrands = computed(() => {
  const list = liveBrands.value.filter((b) => b.uses_metaapi && b.adapter_id === 'metaapi')
  const order = ['any_mt', 'deriv', 'exness', 'tickmill', 'icmarkets']
  return [...list].sort((a, b) => {
    const ia = order.indexOf(a.id)
    const ib = order.indexOf(b.id)
    const pa = ia === -1 ? 99 : ia
    const pb = ib === -1 ? 99 : ib
    return pa - pb
  })
})

const otherBrands = computed(() =>
  liveBrands.value.filter((b) => !b.uses_metaapi || b.adapter_id !== 'metaapi'),
)

const fields = computed(() => {
  const raw = selectedBrand.value?.fields || []
  if (!props.metaapiSharedToken) return raw
  return raw.filter((f) => f.key !== 'metaapi_token')
})

const isLive = computed(() => selectedBrand.value?.status === 'live')

function applyPresets() {
  const b = selectedBrand.value
  if (!b) return
  const next = { ...credentials.value }
  for (const f of b.fields || []) {
    if (next[f.key] === undefined) {
      next[f.key] = f.type === 'boolean' ? 'false' : ''
    }
  }
  for (const [k, v] of Object.entries(b.credential_preset || {})) {
    if (!next[k] || next[k] === '') next[k] = v
  }
  credentials.value = next
  if (!label.value) label.value = b.display_name
}

watch(selectedBrandId, () => {
  applyPresets()
})

function selectBrand(id) {
  selectedBrandId.value = id
  step.value = 2
  testBalance.value = null
  applyPresets()
}

async function connectMock() {
  selectedBrandId.value = 'mock'
  credentials.value = { initial_balance: '10000' }
  label.value = 'Demo account'
  await save()
}

async function testConnection() {
  testing.value = true
  testBalance.value = null
  try {
    const body = {
      brand_id: selectedBrandId.value,
      label: label.value,
      credentials: { ...credentials.value },
    }
    const r = await api('/brokers/test', { method: 'POST', body })
    testBalance.value = r.balance
    emit('message', { text: `Connection OK — balance $${r.balance}`, ok: true })
  } catch (e) {
    emit('message', { text: e.message, ok: false })
  } finally {
    testing.value = false
  }
}

async function save() {
  if (!isLive.value) {
    emit('message', { text: 'This broker is not available yet — use Demo', ok: false })
    return
  }
  saving.value = true
  try {
    await api('/brokers/connection', {
      method: 'POST',
      body: {
        brand_id: selectedBrandId.value,
        label: label.value,
        credentials: { ...credentials.value },
      },
    })
    emit('message', { text: 'Broker connected successfully', ok: true })
    emit('saved')
    step.value = 3
  } catch (e) {
    emit('message', { text: e.message, ok: false })
  } finally {
    saving.value = false
  }
}

const docsUrl = computed(() => {
  const base = props.publicSiteUrl || ''
  return `${base}/docs/BROKER_CONNECT.md`
})
</script>

<template>
  <section class="card wide broker-wizard card-bull">
    <h2>Connect broker</h2>
    <p class="muted disclaimer">
      Market Mamba is not a broker. You connect your own account. Credentials are encrypted on the server.
    </p>
    <p v-if="connection" class="ok">
      Connected: <strong>{{ connection.label || connection.provider }}</strong>
      ({{ connection.provider }})
    </p>

    <div v-if="step === 1" class="wizard-step">
      <h3>1. Choose your broker</h3>
      <p class="metaapi-intro">
        <template v-if="metaapiSharedToken">
          Enter your broker <strong>MT login</strong>, <strong>password</strong>, and
          <strong>server name</strong> — the platform MetaAPI connection is already configured.
        </template>
        <template v-else>
          Live MT brokers use the <strong>MetaAPI MT bridge</strong> — you need a
          <a href="https://app.metaapi.cloud/" target="_blank" rel="noopener">MetaAPI token</a>
          plus your MT login, password, and server name.
        </template>
      </p>

      <h4 v-if="metaapiBrands.length" class="group-title">MT brokers (MetaAPI)</h4>
      <div v-if="metaapiBrands.length" class="brand-grid">
        <button
          v-for="b in metaapiBrands"
          :key="b.id"
          type="button"
          class="brand-card"
          :class="{ active: selectedBrandId === b.id }"
          @click="selectBrand(b.id)"
        >
          <span class="badge" :class="{ highlight: b.id === 'any_mt' }">
            {{ b.id === 'any_mt' ? 'Any broker' : 'MetaAPI MT' }}
          </span>
          <strong>{{ b.display_name }}</strong>
          <span class="muted">{{ b.description }}</span>
        </button>
      </div>

      <h4 v-if="otherBrands.length" class="group-title">Demo &amp; other</h4>
      <div v-if="otherBrands.length" class="brand-grid">
        <button
          v-for="b in otherBrands"
          :key="b.id"
          type="button"
          class="brand-card"
          :class="{ active: selectedBrandId === b.id }"
          @click="selectBrand(b.id)"
        >
          <strong>{{ b.display_name }}</strong>
          <span class="muted">{{ b.description }}</span>
        </button>
      </div>
      <button type="button" class="btn-secondary quick-mock" @click="connectMock">
        Quick: Demo account ($10,000)
      </button>
    </div>

    <div v-else-if="step === 2 && selectedBrand" class="wizard-step">
      <h3>2. Enter credentials — {{ selectedBrand.display_name }}</h3>
      <button type="button" class="link-back" @click="step = 1">← Change broker</button>

      <p v-if="selectedBrand.uses_metaapi" class="metaapi-step">
        Connecting via <strong>MetaAPI</strong> to your {{ selectedBrand.display_name }} MT account.
        <span v-if="metaapiSharedToken">Only your MT credentials are required.</span>
        First save may take 1–3 minutes while MetaAPI deploys.
      </p>

      <ul v-if="selectedBrand.warnings?.length" class="warnings">
        <li v-for="(w, i) in selectedBrand.warnings" :key="i">{{ w }}</li>
      </ul>
      <p v-if="selectedBrand.server_examples?.length" class="hint">
        Example servers:
        <code>{{ selectedBrand.server_examples.join(', ') }}</code>
      </p>
      <p v-if="selectedBrand.help_url" class="hint">
        <a :href="selectedBrand.help_url" target="_blank" rel="noopener">MetaAPI / broker signup</a>
      </p>

      <div class="field">
        <label>Account label</label>
        <input v-model="label" type="text" :placeholder="selectedBrand.display_name" />
      </div>
      <div v-for="f in fields" :key="f.key" class="field">
        <label>{{ f.label }}</label>
        <label v-if="f.type === 'boolean'" class="checkbox-row">
          <input
            type="checkbox"
            :checked="credentials[f.key] === 'true'"
            @change="credentials[f.key] = $event.target.checked ? 'true' : 'false'"
          />
          <span>Enabled</span>
        </label>
        <input
          v-else
          v-model="credentials[f.key]"
          :type="f.type === 'password' ? 'password' : 'text'"
          :placeholder="f.placeholder"
        />
      </div>

      <p class="warn isolation">
        Do not manual-trade the same account while auto-trade is on. The bot only tracks its own orders.
      </p>

      <div class="broker-actions">
        <button type="button" class="btn-secondary" :disabled="testing || saving" @click="testConnection">
          {{ testing ? 'Testing…' : 'Test connection' }}
        </button>
        <button type="button" class="btn-primary" :disabled="testing || saving" @click="save">
          {{ saving ? 'Saving…' : 'Save & connect' }}
        </button>
      </div>
      <p v-if="testBalance != null" class="ok">Test balance: ${{ testBalance }}</p>
      <p class="muted small">First MetaAPI connect may take 1–3 minutes.</p>
    </div>

    <div v-else-if="step === 3" class="wizard-step">
      <h3>3. Ready</h3>
      <p class="ok">Broker connected. Try <code>/balance</code> and <code>/autostart</code> in Telegram.</p>
      <button type="button" class="btn-secondary" @click="step = 1">Connect a different broker</button>
    </div>

    <p class="muted small">
      <a :href="docsUrl" target="_blank" rel="noopener">Full connection guide</a>
    </p>
  </section>
</template>

<style scoped>
.broker-wizard .disclaimer {
  margin-bottom: 1rem;
}
.brand-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem;
  margin: 1rem 0;
}
.brand-card {
  text-align: left;
  padding: 1rem;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--surface);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.brand-card.active,
.brand-card:hover {
  border-color: var(--brand);
  box-shadow: 0 0 12px var(--win-glow);
}
.brand-card strong {
  color: var(--text);
}
.badge {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--brand);
  align-self: flex-start;
}
.badge.highlight {
  color: var(--on-brand);
  background: var(--brand);
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}
.group-title {
  margin: 1.25rem 0 0.5rem;
  font-size: 0.95rem;
  color: var(--text-muted, var(--muted));
}
.metaapi-intro,
.metaapi-step {
  font-size: 0.92rem;
  margin-bottom: 1rem;
  line-height: 1.5;
}
.metaapi-intro a {
  color: var(--brand);
}
.warnings {
  color: var(--warn, #e6a700);
  font-size: 0.9rem;
  margin: 0.75rem 0;
  padding-left: 1.2rem;
}
.hint {
  font-size: 0.9rem;
  margin-bottom: 0.75rem;
}
.isolation {
  font-size: 0.85rem;
  margin: 1rem 0;
  padding: 0.75rem;
  border-radius: 8px;
  background: rgba(230, 167, 0, 0.08);
  border: 1px solid rgba(230, 167, 0, 0.25);
}
.link-back {
  background: none;
  border: none;
  color: var(--brand);
  cursor: pointer;
  margin-bottom: 0.75rem;
  padding: 0;
}
.quick-mock {
  margin-top: 0.5rem;
}
.broker-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1rem;
}
.small {
  font-size: 0.85rem;
  margin-top: 1rem;
}
</style>
