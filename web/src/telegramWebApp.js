/** Telegram Mini App helpers — https://core.telegram.org/bots/webapps */

export function isTelegramMiniApp() {
  const tg = window.Telegram?.WebApp
  return !!(tg && tg.initData)
}

export function initTelegramWebApp() {
  const tg = window.Telegram?.WebApp
  if (!tg) return null
  tg.ready()
  tg.expand()
  tg.setHeaderColor('#151b26')
  tg.setBackgroundColor('#0c1117')
  const p = tg.themeParams
  if (p?.bg_color) {
    document.documentElement.style.setProperty('--tg-bg', p.bg_color)
  }
  if (p?.text_color) {
    document.documentElement.style.setProperty('--tg-text', p.text_color)
  }
  return tg
}

export function telegramInitData() {
  return window.Telegram?.WebApp?.initData || ''
}
