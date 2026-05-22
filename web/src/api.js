/** Relative path — Vite dev proxies to VITE_API_PROXY_TARGET; production serves API same-origin. */
export const API = '/api/v1'

export function apiTargetLabel() {
  return import.meta.env.VITE_API_PROXY_TARGET || 'http://localhost:8090'
}

export function loadSession() {
  return {
    sessionToken: localStorage.getItem('mm_session') || '',
    apiKey: localStorage.getItem('mm_api_key') || '',
    telegramId: localStorage.getItem('mm_telegram_id') || '',
  }
}

export function saveTelegramSession(sessionToken, telegramId) {
  localStorage.setItem('mm_session', sessionToken)
  localStorage.setItem('mm_telegram_id', String(telegramId))
  localStorage.removeItem('mm_api_key')
}

export function saveLegacySession(apiKey, telegramId) {
  localStorage.setItem('mm_api_key', apiKey)
  localStorage.setItem('mm_telegram_id', telegramId)
  localStorage.removeItem('mm_session')
}

export function clearSession() {
  localStorage.removeItem('mm_session')
  localStorage.removeItem('mm_api_key')
  localStorage.removeItem('mm_telegram_id')
}

export function isLoggedIn() {
  const s = loadSession()
  return !!(s.sessionToken || (s.telegramId && s.apiKey))
}

export async function loginWithEmail(email, password) {
  const res = await fetch(`${API}/auth/email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  saveTelegramSession(data.session_token, data.telegram_id)
  return data
}

export async function loginWithTelegramOIDC(idToken) {
  const res = await fetch(`${API}/auth/telegram/oidc`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id_token: idToken }),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  saveTelegramSession(data.session_token, data.telegram_id)
  return data
}

export async function loginWithTelegram(user) {
  const res = await fetch(`${API}/auth/telegram`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(user),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  saveTelegramSession(data.session_token, data.telegram_id)
  return data
}

export async function api(path, { method = 'GET', body } = {}) {
  const { sessionToken, apiKey, telegramId } = loadSession()
  const headers = { 'Content-Type': 'application/json' }
  if (sessionToken) {
    headers['Authorization'] = `Bearer ${sessionToken}`
  } else {
    if (apiKey) headers['X-API-Key'] = apiKey
    if (telegramId) headers['X-Telegram-User-Id'] = telegramId
  }
  const res = await fetch(API + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}
