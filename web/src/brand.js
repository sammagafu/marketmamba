/** Market Mamba brand lines — keep in sync across landing, header, Telegram bio. */

import logoPortrait from './assets/images/Logo-potrait.svg'
import logoLandscape from './assets/images/Logo-landscape.svg'
import logoIcon from './assets/images/favcon.svg'

/** Primary slogan (hero, marketing). */
export const SLOGAN = 'When the market moves, the mamba strikes'

/** Short tag for header / mobile (same idea, tighter). */
export const SLOGAN_SHORT = 'Market moves. Mamba strikes.'

/** Supporting line under the name. */
export const TAGLINE = 'Forex automation on Telegram'

/** Brand SVGs (Vite resolves URLs; favicon copied to public on build). */
export const ASSETS = {
  logoPortrait,
  logoLandscape,
  logoIcon,
  favicon: logoIcon,
}
