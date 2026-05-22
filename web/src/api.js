const API = '/api/v1'

export function loadSession() {
  return {
    apiKey: localStorage.getItem('mm_api_key') || '',
    telegramId: localStorage.getItem('mm_telegram_id') || '',
  }
}

export function saveSession(apiKey, telegramId) {
  localStorage.setItem('mm_api_key', apiKey)
  localStorage.setItem('mm_telegram_id', telegramId)
}

export async function api(path, { method = 'GET', body } = {}) {
  const { apiKey, telegramId } = loadSession()
  const headers = { 'Content-Type': 'application/json' }
  if (apiKey) headers['X-API-Key'] = apiKey
  if (telegramId) headers['X-Telegram-User-Id'] = telegramId
  const res = await fetch(API + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}
