<script setup>
import { ref } from 'vue'
import { loginWithEmail } from '../api'

const emit = defineEmits(['logged-in', 'error'])

const email = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    const data = await loginWithEmail(email.value.trim(), password.value)
    emit('logged-in', data)
  } catch (e) {
    emit('error', e.message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="email-login">
    <p class="section-eyebrow login-eyebrow">Operations</p>
    <p class="login-title">Administrator sign-in</p>
    <div class="field">
      <label>Email</label>
      <input v-model="email" type="email" autocomplete="username" placeholder="magafu317@gmail.com" />
    </div>
    <div class="field">
      <label>Password</label>
      <input v-model="password" type="password" autocomplete="current-password" />
    </div>
    <button class="btn-primary" type="button" :disabled="loading" @click="submit">
      {{ loading ? 'Signing in…' : 'Sign in as admin' }}
    </button>
  </div>
</template>

<style scoped>
.email-login {
  margin-top: 1rem;
  text-align: left;
  width: 100%;
  max-width: 100%;
  min-width: 0;
}
.login-eyebrow { text-align: center; margin-bottom: 0.35rem; }
.login-title { font-weight: 700; margin: 0 0 1rem; text-align: center; font-size: 1.05rem; }
.field { margin-bottom: 0.75rem; }
button { width: 100%; margin-top: 0.25rem; }
</style>
