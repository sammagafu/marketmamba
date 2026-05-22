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
    <p class="login-title">Admin login (email)</p>
    <div class="field">
      <label>Email</label>
      <input v-model="email" type="email" autocomplete="username" placeholder="you@example.com" />
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
.email-login { margin-top: 1rem; text-align: left; }
.login-title { font-weight: 600; margin-bottom: 0.75rem; text-align: center; }
.field { margin-bottom: 0.75rem; }
button { width: 100%; margin-top: 0.25rem; }
</style>
