<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { loginWithTelegram } from '../api'

const props = defineProps({
  botUsername: { type: String, required: true },
})
const emit = defineEmits(['logged-in', 'error'])

const container = ref(null)
let scriptEl = null

async function onTelegramAuth(user) {
  try {
    const data = await loginWithTelegram(user)
    emit('logged-in', data)
  } catch (e) {
    emit('error', e.message)
  }
}

onMounted(() => {
  window.onTelegramAuth = onTelegramAuth
  if (!props.botUsername || !container.value) return
  scriptEl = document.createElement('script')
  scriptEl.async = true
  scriptEl.src = 'https://telegram.org/js/telegram-widget.js?22'
  scriptEl.setAttribute('data-telegram-login', props.botUsername)
  scriptEl.setAttribute('data-size', 'large')
  scriptEl.setAttribute('data-radius', '8')
  scriptEl.setAttribute('data-onauth', 'onTelegramAuth(user)')
  scriptEl.setAttribute('data-request-access', 'write')
  container.value.appendChild(scriptEl)
})

onUnmounted(() => {
  if (window.onTelegramAuth === onTelegramAuth) {
    delete window.onTelegramAuth
  }
})
</script>

<template>
  <div class="telegram-login">
    <p class="muted">Sign in with your Telegram account (same as the bot).</p>
    <div ref="container" class="widget-wrap"></div>
    <p class="hint muted">
      Domain must be set in @BotFather → /setdomain for production.
      Localhost may not work — use manual login below for dev.
    </p>
  </div>
</template>

<style scoped>
.telegram-login { text-align: center; }
.widget-wrap {
  display: flex;
  justify-content: center;
  min-height: 48px;
  margin: 1rem 0;
}
.hint { font-size: 0.8rem; max-width: 420px; margin: 0 auto; }
</style>
