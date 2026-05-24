/** Market Mamba brand lines — keep in sync across landing, header, Telegram bio. */

import logoPortrait from './assets/images/Logo-potrait.svg'
import logoLandscape from './assets/images/Logo-landscape.svg'
import logoIcon from './assets/images/favcon.svg'

/** Primary slogan (hero, marketing). */
export const SLOGAN = 'When the market moves, the mamba strikes'

/** Short tag for header / mobile. */
export const SLOGAN_SHORT = 'Market moves. Mamba strikes.'

/** One-line descriptor under the logo. */
export const TAGLINE = 'Controlled automation on your broker'

/** Core product promise. */
export const VALUE_PROPOSITION =
  'Automate with discipline: built-in risk limits, qualified signals, and execution on the MT broker you already use.'

/** Billing — no card processors. */
export const PAYMENT_NOTE = 'USDT via Binance only. No cards or Stripe.'

/** Hero rotating emphasis words (lowercase in UI). */
export const HERO_FOCUS_WORDS = ['discipline', 'risk limits', 'automation', 'control']

/** Login portal */
export const PORTAL_TITLE = 'Client sign-in'
export const PORTAL_SUB = 'Telegram login · USDT billing in the bot'

export const ASSETS = {
  logoPortrait,
  logoLandscape,
  logoIcon,
  favicon: logoIcon,
}
