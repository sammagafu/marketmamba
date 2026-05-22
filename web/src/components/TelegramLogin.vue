<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { loginWithTelegramOIDC } from '../api'

const props = defineProps({
  botUsername: { type: String, default: 'market_mamba_bot' },
  clientId: { type: String, required: true },
  loginDomain: { type: String, default: 'marketmamba.kkooapp.co.tz' },
  publicSiteUrl: { type: String, default: 'https://marketmamba.kkooapp.co.tz' },
})
const emit = defineEmits(['logged-in', 'error'])

const container = ref(null)
const ready = ref(false)
const domainError = ref('')
let scriptEl = null

const currentHost = computed(() =>
  typeof window !== 'undefined' ? window.location.hostname : '',
)

const productionLoginUrl = computed(() => props.publicSiteUrl.replace(/\/$/, ''))

async function onTelegramAuth(data) {
  if (!data || data.error) {
    emit('error', data?.error || 'Telegram login cancelled')
    return
  }
  if (!data.id_token) {
    emit('error', 'No id_token from Telegram')
    return
  }
  try {
    const res = await loginWithTelegramOIDC(data.id_token)
    emit('logged-in', res)
  } catch (e) {
    emit('error', e.message)
  }
}

onMounted(() => {
  if (!props.clientId) {
    domainError.value = 'TELEGRAM_BOT_CLIENT_ID not configured on server.'
    return
  }

  window.onTelegramAuth = onTelegramAuth

  if (!container.value) return

  const btn = document.createElement('button')
  btn.type = 'button'
  btn.className = 'tg-auth-button'
  btn.setAttribute('data-style', 'shine')
  btn.textContent = 'Sign in with Telegram'
  container.value.appendChild(btn)

  scriptEl = document.createElement('script')
  scriptEl.async = true
  scriptEl.src = 'https://oauth.telegram.org/js/telegram-login.js?5'
  scriptEl.setAttribute('data-client-id', props.clientId)
  scriptEl.setAttribute('data-onauth', 'onTelegramAuth(data)')
  scriptEl.setAttribute('data-request-access', 'write phone')
  scriptEl.onload = () => {
    ready.value = true
  }
  scriptEl.onerror = () => {
    domainError.value =
      'Could not load Telegram login. In @BotFather → Bot Settings → Web Login, add: https://' +
      props.loginDomain
  }
  document.head.appendChild(scriptEl)
})

onUnmounted(() => {
  if (window.onTelegramAuth === onTelegramAuth) {
    delete window.onTelegramAuth
  }
  scriptEl?.remove()
})
</script>

<template>
  <div class="telegram-login">
    <p class="login-title">Sign in with Telegram</p>
    <p class="muted">@{{ botUsername }} · Client ID {{ clientId }}</p>

    <p v-if="domainError" class="warn-box">{{ domainError }}</p>
    <p v-if="currentHost !== loginDomain" class="warn-box">
      Add <code>https://{{ loginDomain }}</code> (and dev URL if needed) in
      @BotFather → <strong>Web Login</strong> → Allowed URLs.
      <a :href="productionLoginUrl" class="prod-link">Open production site</a>
    </p>

    <div ref="container" class="widget-wrap"></div>

    <a
      :href="`https://t.me/${botUsername}`"
      target="_blank"
      rel="noopener"
      class="btn-telegram-fallback"
    >
      Open @{{ botUsername }} in Telegram
    </a>
  </div>
</template>

<style scoped>
.telegram-login { text-align: center; }
.login-title { font-size: 1.15rem; font-weight: 600; margin: 0 0 0.25rem; }
.widget-wrap {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 48px;
  margin: 1rem 0;
}
.warn-box {
  font-size: 0.85rem;
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.08);
  border: 1px solid rgba(251, 191, 36, 0.25);
  border-radius: 8px;
  padding: 0.75rem;
  margin: 0.75rem 0;
  text-align: left;
}
.warn-box code { color: #fde68a; }
.prod-link { color: #38bdf8; }
.btn-telegram-fallback {
  display: inline-flex;
  margin-top: 0.5rem;
  padding: 0.65rem 1.25rem;
  border-radius: 8px;
  background: #2aabee;
  color: #fff;
  font-weight: 600;
  text-decoration: none;
}
</style>
