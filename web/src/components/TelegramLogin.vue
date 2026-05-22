<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
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
let scriptEl = null

const productionLoginUrl = () => props.publicSiteUrl.replace(/\/$/, '')

function currentHost() {
  return typeof window !== 'undefined' ? window.location.hostname : ''
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
  const id = Number(props.clientId)
  if (!id || !window.Telegram?.Login) {
    emit('error', 'Telegram sign-in is still loading — wait a moment and try again')
    return
  }
  window.Telegram.Login.auth(
    { client_id: id, request_access: ['write'] },
    handleOidc,
  )
}

function mountOidcWidget() {
  useOidc.value = true
  scriptEl = document.createElement('script')
  scriptEl.async = true
  scriptEl.src = 'https://oauth.telegram.org/js/telegram-login.js?5'
  scriptEl.setAttribute('data-client-id', props.clientId)
  scriptEl.setAttribute('data-onauth', 'mmTelegramOIDC(data)')
  scriptEl.setAttribute('data-request-access', 'write')
  scriptEl.setAttribute('data-lang', 'en')
  scriptEl.onload = () => {
    const id = Number(props.clientId)
    if (window.Telegram?.Login?.init) {
      window.Telegram.Login.init({ client_id: id, request_access: ['write'] }, handleOidc)
    }
  }
  scriptEl.onerror = () => {
    loadError.value =
      `Could not load Telegram sign-in. In @BotFather → Bot Settings → Web Login, add: https://${props.loginDomain}`
  }
  document.head.appendChild(scriptEl)
}

function mountLegacyWidget() {
  useOidc.value = false
  const el = widgetRef.value
  if (!el) return

  const callbackUrl = `${productionLoginUrl()}/api/v1/auth/telegram/callback`

  scriptEl = document.createElement('script')
  scriptEl.async = true
  scriptEl.src = 'https://telegram.org/js/telegram-widget.js?22'
  scriptEl.setAttribute('data-telegram-login', props.botUsername.replace('@', ''))
  scriptEl.setAttribute('data-size', 'large')
  scriptEl.setAttribute('data-radius', '10')
  scriptEl.setAttribute('data-userpic', 'false')
  scriptEl.setAttribute('data-request-access', 'write')
  scriptEl.setAttribute('data-auth-url', callbackUrl)
  scriptEl.onerror = () => {
    loadError.value = 'Could not load Telegram Login Widget.'
  }
  el.appendChild(scriptEl)
}

onMounted(async () => {
  window.mmTelegramOIDC = handleOidc
  window.mmTelegramLegacy = handleLegacy

  if (props.clientId) {
    mountOidcWidget()
  } else if (props.botUsername) {
    await nextTick()
    mountLegacyWidget()
  } else {
    loadError.value = 'Telegram bot not configured on server (TELEGRAM_BOT_TOKEN).'
  }
})

onUnmounted(() => {
  delete window.mmTelegramOIDC
  delete window.mmTelegramLegacy
  scriptEl?.remove()
  if (widgetRef.value) widgetRef.value.innerHTML = ''
})
</script>

<template>
  <div class="telegram-login">
    <button
      v-if="useOidc && clientId"
      type="button"
      class="telegram-signin-btn"
      :disabled="!!loadError"
      @click="openOidcPopup"
    >
      <svg class="telegram-signin-icon" viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.74-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.12.27z"
        />
      </svg>
      Log in with Telegram
    </button>

    <div v-else ref="widgetRef" class="telegram-widget-slot" />

    <p v-if="loadError" class="warn-box">{{ loadError }}</p>
    <p v-else-if="currentHost() && currentHost() !== loginDomain" class="warn-box">
      For production login, open
      <a :href="productionLoginUrl()" class="prod-link">{{ productionLoginUrl() }}</a>
      or add <code>https://{{ loginDomain }}</code> in @BotFather → Web Login.
    </p>

    <p v-if="!clientId && botUsername && !loadError" class="muted widget-hint">
      Official Telegram button · @{{ botUsername.replace('@', '') }}
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
  min-height: 52px;
  margin-bottom: 0.75rem;
}

/* Official OIDC button hook (SDK may replace/enhance) */
.telegram-widget-slot :deep(.tg-auth-button) {
  font-family: var(--font-sans);
  cursor: pointer;
}

/* Branded fallback / primary button (Telegram blue) */
.telegram-signin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  width: 100%;
  max-width: 100%;
  min-height: var(--touch-min);
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

.telegram-signin-btn:hover {
  background: #229ed9;
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(42, 171, 238, 0.45);
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
