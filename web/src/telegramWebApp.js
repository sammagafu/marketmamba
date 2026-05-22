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
  if (tg.themeParams?.bg_color) {
    document.documentElement.style.setProperty('--tg-bg', tg.themeParams.bg_color)
  }
  return tg
}

export function telegramInitData() {
  return window.Telegram?.WebApp?.initData || ''
}
