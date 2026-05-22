const API = '/api/v1'

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
