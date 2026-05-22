<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { loginWithTelegramOIDC, loginWithTelegram } from '../api'

const props = defineProps({
  botUsername: { type: String, default: 'market_mamba_bot' },
  clientId: { type: String, default: '' },
  loginDomain: { type: String, default: 'marketmamba.kkooapp.co.tz' },
  publicSiteUrl: { type: String, default: 'https://marketmamba.kkooapp.co.tz' },
})

const emit = defineEmits(['logged-in', 'error'])

const widgetRef = ref(null)
const loadError = ref('')
const useOidc = ref(false)
const widgetReady = ref(false)
let scriptEl = null

const productionLoginUrl = () => props.publicSiteUrl.replace(/\/$/, '')

function currentHost() {
  return typeof window !== 'undefined' ? window.location.hostname : ''
}

function hasOidcClient() {
  const id = String(props.clientId || '').trim()
  return id !== '' && id !== '0'
}

/** OIDC requires origin matching a URL registered in @BotFather → Web Login. */
function loginOrigin() {
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin
  }
  try {
    return new URL(productionLoginUrl()).origin
  } catch {
    return `https://${props.loginDomain}`
  }
}

function oidcInitOptions() {
  return {
    client_id: Number(props.clientId),
    request_access: ['write'],
    origin: loginOrigin(),
  }
}

async function handleOidc(data) {
  if (!data || data.error) {
    emit('error', data?.error || 'Telegram sign-in cancelled')
    return
  }
  if (!data.id_token) {
    emit('error', 'No token from Telegram — try again')
    return
  }
  try {
    const res = await loginWithTelegramOIDC(data.id_token)
    emit('logged-in', res)
  } catch (e) {
    emit('error', e.message)
  }
}

async function handleLegacy(user) {
  if (!user?.id) {
    emit('error', 'Telegram sign-in cancelled')
    return
  }
  try {
    const res = await loginWithTelegram({
      id: user.id,
      first_name: user.first_name,
      last_name: user.last_name,
      username: user.username,
      photo_url: user.photo_url,
      auth_date: user.auth_date,
      hash: user.hash,
    })
    emit('logged-in', res)
  } catch (e) {
    emit('error', e.message)
  }
}

function openOidcPopup() {
  const opts = oidcInitOptions()
  if (!opts.client_id || !window.Telegram?.Login) {
    emit('error', 'Telegram sign-in is still loading — wait a moment and try again')
    return
  }
  window.Telegram.Login.auth(opts, handleOidc)
}

function teardown() {
  scriptEl?.remove()
  scriptEl = null
  if (widgetRef.value) widgetRef.value.innerHTML = ''
  widgetReady.value = false
}

function mountOidcWidget() {
  useOidc.value = true
  teardown()
  scriptEl = document.createElement('script')
  scriptEl.async = true
  scriptEl.src = 'https://oauth.telegram.org/js/telegram-login.js?5'
  scriptEl.onload = () => {
    const id = Number(props.clientId)
    if (window.Telegram?.Login?.init) {
      window.Telegram.Login.init(oidcInitOptions(), handleOidc)
    }
    widgetReady.value = true
  }
  scriptEl.onerror = () => {
    loadError.value =
      `Could not load Telegram sign-in. In @BotFather → Web Login, add origin: ${loginOrigin()}`
  }
  document.head.appendChild(scriptEl)
}

async function mountLegacyWidget() {
  useOidc.value = false
  teardown()
  loadError.value = ''
  await nextTick()
  const el = widgetRef.value
  if (!el) return

  const bot = props.botUsername.replace('@', '').trim()
  if (!bot) {
    loadError.value = 'Telegram bot username not configured (TELEGRAM_BOT_USERNAME).'
    return
  }

  scriptEl = document.createElement('script')
  scriptEl.async = true
  scriptEl.src = 'https://telegram.org/js/telegram-widget.js?22'
  scriptEl.setAttribute('data-telegram-login', bot)
  scriptEl.setAttribute('data-size', 'large')
  scriptEl.setAttribute('data-radius', '10')
  scriptEl.setAttribute('data-userpic', 'false')
  scriptEl.setAttribute('data-request-access', 'write')
  scriptEl.setAttribute('data-onauth', 'mmTelegramLegacy(user)')
  scriptEl.onload = () => {
    widgetReady.value = true
    // If domain is not whitelisted, Telegram leaves the slot empty — keep branded button visible.
    setTimeout(() => {
      if (!el.querySelector('iframe, a, button')) {
        loadError.value =
          `Telegram button did not load. In @BotFather → your bot → Bot Settings → Domain, add: ${props.loginDomain}`
      }
    }, 2500)
  }
  scriptEl.onerror = () => {
    loadError.value = 'Could not load Telegram Login Widget (blocked or offline).'
  }
  el.appendChild(scriptEl)
}

async function setupLogin() {
  loadError.value = ''
  if (hasOidcClient()) {
    mountOidcWidget()
  } else if (props.botUsername) {
    await mountLegacyWidget()
  } else {
    loadError.value = 'Telegram bot not configured on server (TELEGRAM_BOT_TOKEN).'
  }
}

function onSignInClick() {
  if (useOidc.value) {
    openOidcPopup()
    return
  }
  const el = widgetRef.value
  const official =
    el?.querySelector('iframe') ||
    el?.querySelector('a') ||
    el?.querySelector('button')
  if (official) {
    official.click()
    return
  }
  emit(
    'error',
    loadError.value ||
      `Add https://${props.loginDomain} in @BotFather → Bot Settings → Domain, then refresh.`,
  )
}

onMounted(async () => {
  window.mmTelegramOIDC = handleOidc
  window.mmTelegramLegacy = handleLegacy
  await setupLogin()
})

watch(
  () => [props.clientId, props.botUsername],
  () => {
    setupLogin()
  },
)

onUnmounted(() => {
  delete window.mmTelegramOIDC
  delete window.mmTelegramLegacy
  teardown()
})
</script>

<template>
  <div class="telegram-login">
    <button
      type="button"
      class="telegram-signin-btn"
      :class="{ secondary: !useOidc && widgetReady }"
      @click="onSignInClick"
    >
      <svg class="telegram-signin-icon" viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M9.78 18.65l.28-4.23 7.68-6.92c.34-.31-.07-.46-.52-.19L7.74 13.3 3.64 12c-.88-.25-.89-.86.2-1.3l15.97-6.16c.73-.33 1.43.18 1.15 1.3l-2.72 12.81c-.19.91-.74 1.13-1.5.71L12.6 16.3l-1.99 1.93c-.23.23-.42.42-.83.42z"
        />
      </svg>
      Log in with Telegram
    </button>

    <div
      v-if="!useOidc"
      ref="widgetRef"
      class="telegram-widget-slot"
      :class="{ 'has-widget': widgetReady }"
    />

    <p v-if="loadError" class="warn-box">{{ loadError }}</p>
    <p v-else-if="currentHost() && currentHost() !== loginDomain" class="warn-box">
      For production login, open
      <a :href="productionLoginUrl()" class="prod-link">{{ productionLoginUrl() }}</a>
      or add <code>{{ loginOrigin() }}</code> in @BotFather → Web Login (and Domain for the widget).
    </p>

    <p v-if="!useOidc && botUsername && !loadError" class="muted widget-hint">
      Official Telegram button appears below when your domain is linked · @{{ botUsername.replace('@', '') }}
    </p>
  </div>
</template>

<style scoped>
.telegram-login {
  text-align: center;
  width: 100%;
  max-width: 100%;
}

.telegram-widget-slot {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 0;
  margin-top: 0.75rem;
}

.telegram-widget-slot.has-widget {
  min-height: 52px;
}

.telegram-widget-slot :deep(iframe) {
  max-width: 100%;
}

.telegram-signin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  width: 100%;
  max-width: 100%;
  min-height: var(--touch-min, 48px);
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 10px;
  background: #2aabee;
  color: #fff;
  font-family: var(--font-sans);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, transform 0.15s, box-shadow 0.2s;
  box-shadow: 0 4px 14px rgba(42, 171, 238, 0.35);
}

.telegram-signin-btn.secondary {
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--border);
  box-shadow: none;
  font-size: 0.9rem;
  min-height: 42px;
  margin-bottom: 0.25rem;
}

.telegram-signin-btn:hover {
  background: #229ed9;
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(42, 171, 238, 0.45);
}

.telegram-signin-btn.secondary:hover {
  background: var(--surface-raised);
  transform: none;
  box-shadow: none;
}

.telegram-signin-icon {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
}

.widget-hint {
  margin: 0.5rem 0 0;
  font-size: 0.8rem;
}

.warn-box {
  font-size: 0.85rem;
  color: var(--warn);
  background: var(--warn-bg);
  border: 1px solid var(--warn-border);
  border-radius: 8px;
  padding: 0.75rem;
  margin: 0.75rem 0 0;
  text-align: left;
}

.warn-box code {
  color: var(--text-soft);
  background: var(--surface);
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
}

.prod-link {
  color: var(--brand);
  font-weight: 600;
}
</style>
