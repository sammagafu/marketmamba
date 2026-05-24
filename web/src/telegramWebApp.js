/** Telegram Mini App helpers — https://core.telegram.org/bots/webapps */

const BRAND_BG = '#000000'
const BRAND_HEADER = '#0a0a0a'
const BRAND_ACCENT = '#3dff7a'

export function isTelegramMiniApp() {
  const tg = window.Telegram?.WebApp
  return !!(tg && tg.initData)
}

export function initTelegramWebApp() {
  const tg = window.Telegram?.WebApp
  if (!tg) return null
  tg.ready()
  tg.expand()
  if (typeof tg.disableVerticalSwipes === 'function') {
    tg.disableVerticalSwipes()
  }
  tg.setHeaderColor(BRAND_HEADER)
  tg.setBackgroundColor(BRAND_BG)
  if (typeof tg.setBottomBarColor === 'function') {
    tg.setBottomBarColor(BRAND_BG)
  }
  const p = tg.themeParams
  if (p?.bg_color) {
    document.documentElement.style.setProperty('--tg-bg', p.bg_color)
  }
  if (p?.text_color) {
    document.documentElement.style.setProperty('--tg-text', p.text_color)
  }
  document.documentElement.classList.add('tg-mini-app')
  return tg
}

export function telegramInitData() {
  return window.Telegram?.WebApp?.initData || ''
}

export function hapticLight() {
  try {
    window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light')
  } catch {
    /* ignore */
  }
}

export function hapticSuccess() {
  try {
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch {
    /* ignore */
  }
}

let mainButtonHandler = null

export function hideMainButton() {
  const tg = window.Telegram?.WebApp
  if (!tg?.MainButton) return
  if (mainButtonHandler && tg.MainButton.offClick) {
    tg.MainButton.offClick(mainButtonHandler)
    mainButtonHandler = null
  }
  tg.MainButton.hide()
}

export function showMainButton(text, onClick) {
  const tg = window.Telegram?.WebApp
  if (!tg?.MainButton || !onClick) return
  hideMainButton()
  mainButtonHandler = onClick
  tg.MainButton.setText(text)
  if (tg.MainButton.setParams) {
    tg.MainButton.setParams({ color: BRAND_ACCENT, text_color: '#041a0c' })
  }
  tg.MainButton.onClick(mainButtonHandler)
  tg.MainButton.show()
}
