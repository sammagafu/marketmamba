<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import {
  SLOGAN,
  SLOGAN_SHORT,
  TAGLINE,
  VALUE_PROPOSITION,
  PAYMENT_NOTE,
  HERO_FOCUS_WORDS,
  PORTAL_TITLE,
  PORTAL_SUB,
} from '../brand'
import BrandLogo from './BrandLogo.vue'
import HowWeTrade from './HowWeTrade.vue'

const props = defineProps({
  config: { type: Object, default: null },
  apiOffline: { type: Boolean, default: false },
  apiTarget: { type: String, default: '' },
  botUsername: { type: String, default: 'market_mamba_bot' },
})

defineEmits(['error'])

const tickers = ['BTCUSD', 'EURUSD', 'GBPUSD', 'USDJPY', 'XAUUSD', 'AUDUSD']
const featuredPair = 'BTCUSD'
const headlineWord = ref(0)
const words = HERO_FOCUS_WORDS
const displayTrades = ref(0)
const displayUsers = ref(0)

const totalTrades = computed(() => props.config?.total_trades ?? 0)
const totalUsers = computed(() => props.config?.total_users ?? 0)

let timers = []

function animateCounter(target, setter, duration = 1200) {
  const start = performance.now()
  const from = 0
  const tick = (now) => {
    const p = Math.min(1, (now - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3)
    setter(Math.round(from + (target - from) * ease))
    if (p < 1) requestAnimationFrame(tick)
  }
  requestAnimationFrame(tick)
}

onMounted(() => {
  timers.push(setInterval(() => {
    headlineWord.value = (headlineWord.value + 1) % words.length
  }, 3200))

  runCounters()
})

watch(
  () => [props.config?.total_trades, props.config?.total_users],
  () => runCounters(),
)

function runCounters() {
  if (totalTrades.value) animateCounter(totalTrades.value, (v) => { displayTrades.value = v })
  if (totalUsers.value) animateCounter(totalUsers.value, (v) => { displayUsers.value = v })
}

onUnmounted(() => {
  timers.forEach(clearInterval)
})
</script>

<template>
  <div class="landing">
    <div class="landing-vignette" aria-hidden="true" />

    <div class="marquee" aria-hidden="true">
      <div class="marquee-track">
        <span v-for="n in 2" :key="n" class="marquee-inner">
          <span
            v-for="(pair, idx) in tickers"
            :key="`${n}-${pair}`"
            class="marquee-item bull"
          >{{ pair }}</span>
          <span class="marquee-dot">◆</span>
          <span class="marquee-item live bull">RISK-LIMITED AUTOMATION</span>
          <span class="marquee-dot">◆</span>
        </span>
      </div>
    </div>

    <div class="landing-grid">
      <section class="hero-col">
        <div class="logo-mark">
          <BrandLogo variant="portrait" class="logo-core" alt="" />
        </div>

        <p class="eyebrow">
          <span class="live-dot" /> Market Mamba · {{ TAGLINE }}
        </p>

        <p class="hero-slogan">{{ SLOGAN }}</p>

        <h1 class="mega-title">
          <span class="line-muted">Forex automation with</span>
          <span class="line-flip" :key="words[headlineWord]">{{ words[headlineWord] }}</span>
        </h1>

        <p class="hero-lede hero-lede-primary">
          {{ config?.value_proposition || VALUE_PROPOSITION }}
        </p>
        <p class="hero-lede hero-lede-secondary">
          <span class="hero-slogan-inline">{{ SLOGAN }}</span>
        </p>
        <p class="hero-meta">
          {{ PAYMENT_NOTE }}
          <span class="hero-meta-sep">·</span>
          <a class="hero-link" href="#how-we-trade">How we trade</a>
          <template v-if="config?.contact_us_url">
            <span class="hero-meta-sep">·</span>
            <a
              class="hero-link"
              :href="config.contact_us_url"
              target="_blank"
              rel="noopener"
            >{{ config.contact_us_label || 'Contact us' }}</a>
          </template>
        </p>

        <div class="pair-hero">
          <span class="pair-label">Example pair</span>
          <span class="pair-big">{{ featuredPair }}</span>
          <span class="pair-meta">
            Qualified setups · SL &amp; TP required
          </span>
        </div>

        <div class="stat-row">
          <div class="stat-card stat-bull">
            <span class="stat-num">{{ displayTrades }}</span>
            <span class="stat-cap">Trades on record</span>
          </div>
          <div class="stat-card stat-bull stat-alt">
            <span class="stat-num">{{ displayUsers }}</span>
            <span class="stat-cap">Registered clients</span>
          </div>
          <div class="stat-card stat-glow">
            <span class="stat-num">2%</span>
            <span class="stat-cap">Default daily loss cap</span>
          </div>
        </div>

        <div class="steps">
          <div class="step">
            <span class="step-n">01</span>
            <span>Sign in with Telegram</span>
          </div>
          <div class="step-line" />
          <div class="step">
            <span class="step-n">02</span>
            <span>Link your MT broker</span>
          </div>
          <div class="step-line" />
          <div class="step">
            <span class="step-n">03</span>
            <span>Run automation within limits</span>
          </div>
        </div>
      </section>

      <aside id="login-portal" class="portal-col">
        <div class="portal">
          <div class="portal-glow" aria-hidden="true" />
          <div class="portal-inner">
            <h2 class="portal-title">{{ PORTAL_TITLE }}</h2>
            <p class="portal-sub">
              {{ PORTAL_SUB }} · {{ config?.free_trial_days ?? 5 }}-day evaluation
            </p>

            <p v-if="apiOffline" class="api-offline">
              API unreachable at <strong>{{ apiTarget }}</strong> — start the Go server
              (<code>go run cmd/server/main.go</code>) or check the Vite proxy.
            </p>

            <div class="portal-slot">
              <slot />
            </div>

            <a
              class="telegram-fallback"
              :href="`https://t.me/${botUsername}`"
              target="_blank"
              rel="noopener"
            >
              Open Telegram bot → @{{ botUsername }}
            </a>
          </div>
        </div>

        <p class="scroll-hint">
          <span class="scroll-line" aria-hidden="true" />
          <a href="#how-we-trade">Our process</a>
        </p>
      </aside>
    </div>

    <HowWeTrade />
  </div>
</template>

<style scoped>
.landing {
  position: relative;
  min-height: calc(100vh - 80px);
  width: 100%;
  max-width: 100%;
  overflow-x: clip;
  overflow-x: hidden;
  overflow-y: visible;
  margin: 0;
  padding: 0 var(--page-pad-right) 3rem var(--page-pad);
  padding-bottom: max(2rem, env(safe-area-inset-bottom));
}

.landing-vignette {
  position: fixed;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  background:
    radial-gradient(ellipse 80% 60% at 30% 40%, var(--win-soft), transparent 55%),
    radial-gradient(ellipse 50% 40% at 75% 60%, var(--win-dim), transparent 50%),
    linear-gradient(180deg, transparent 0%, var(--bg) 100%);
}

.marquee {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 100%;
  overflow: hidden;
  border-bottom: 1px solid var(--border);
  background: var(--header-bg);
  backdrop-filter: blur(8px);
  padding: 0.5rem 0;
  margin: 0 0 1.75rem;
}

.marquee-track {
  display: flex;
  width: max-content;
  animation: marquee 28s linear infinite;
}

.marquee-inner {
  display: flex;
  align-items: center;
  gap: 2rem;
  padding-right: 2rem;
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--muted);
}

.marquee-item.bull {
  color: var(--win-bright);
  text-shadow: 0 0 14px var(--win-glow);
}

.marquee-item.live {
  animation: pulse-live 2s ease-in-out infinite;
}

@keyframes pulse-live {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.65; }
}

.marquee-dot {
  color: var(--muted);
  font-size: 0.5rem;
}

@keyframes marquee {
  to { transform: translateX(-50%); }
}

.landing-grid {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
  align-items: start;
}

/* Mobile: sign-in portal above the fold */
.portal-col {
  order: -1;
}

.hero-col,
.portal-col {
  min-width: 0;
  max-width: 100%;
}

@media (min-width: 900px) {
  .landing-grid {
    grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr);
    gap: 3rem;
    padding-top: 1rem;
  }
  .portal-col {
    order: unset;
    position: sticky;
    top: calc(var(--header-h) + 1rem);
  }
}

.logo-mark {
  width: 120px;
  height: 120px;
  margin-bottom: 1.5rem;
}

.logo-core {
  width: 100% !important;
  height: 100% !important;
  filter: drop-shadow(0 0 20px var(--brand-glow));
  animation: float 4s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-6px); }
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--win-bright);
  margin: 0 0 1rem;
}

.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--win);
  box-shadow: 0 0 12px var(--win-glow);
  animation: blink 1.5s ease-in-out infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.85); }
}

.hero-slogan {
  margin: 0 0 1rem;
  font-size: clamp(0.95rem, 2.5vw, 1.15rem);
  font-weight: 500;
  line-height: 1.4;
  letter-spacing: -0.01em;
  color: var(--text-soft);
  max-width: 100%;
  overflow-wrap: anywhere;
}

.hero-slogan-inline {
  font-style: italic;
  color: var(--text-soft);
}

.mega-title {
  margin: 0 0 1rem;
  font-size: clamp(1.75rem, 8vw, 3.75rem);
  font-weight: 800;
  line-height: 1.05;
  letter-spacing: -0.03em;
  overflow-wrap: anywhere;
}

.line-muted {
  display: block;
  color: var(--muted);
  font-size: 0.55em;
  font-weight: 600;
  margin-bottom: 0.15em;
}

.line-flip {
  display: block;
  font-size: clamp(1.35rem, 6vw, 3.2rem);
  background: linear-gradient(90deg, var(--win-bright) 0%, var(--win) 55%, var(--win-deep) 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  animation: word-in 0.6s cubic-bezier(0.22, 1, 0.36, 1);
}

@keyframes word-in {
  from {
    opacity: 0;
    transform: translateY(12px) skewX(-4deg);
    filter: blur(4px);
  }
  to {
    opacity: 1;
    transform: none;
    filter: none;
  }
}

.hero-lede-primary {
  margin: 0 0 0.85rem;
  max-width: 38rem;
  font-size: clamp(1rem, 2.6vw, 1.125rem);
  line-height: 1.6;
  color: var(--text-soft);
  font-weight: 500;
}

.hero-lede-secondary {
  margin: 0 0 1rem;
}

.hero-meta {
  margin: 0 0 1.75rem;
  max-width: 38rem;
  font-size: 0.875rem;
  line-height: 1.55;
  color: var(--muted);
}

.hero-meta-sep {
  margin: 0 0.35rem;
  opacity: 0.5;
}

.hero-link {
  color: var(--win-bright);
  font-weight: 600;
  text-decoration: none;
  border-bottom: 1px solid var(--win-dim);
  transition: color 0.2s, border-color 0.2s;
}
.hero-link:hover {
  color: var(--win);
  border-color: var(--win);
}

.pair-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.75rem 1.25rem;
  padding: 1rem 1.25rem;
  margin-bottom: 1.75rem;
  border-radius: 16px;
  background: var(--surface);
  border: 1px solid var(--border-strong);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  width: 100%;
  max-width: 100%;
}

.pair-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--muted);
  width: 100%;
}

.pair-big {
  font-size: clamp(1.5rem, 6vw, 2.25rem);
  font-weight: 900;
  font-variant-numeric: tabular-nums;
  color: var(--text);
  text-shadow: 0 0 24px var(--brand-glow);
}

.pair-meta {
  font-size: 0.8rem;
  font-weight: 600;
}

.pair-buy {
  color: var(--win-bright);
}

.pair-sell {
  color: var(--muted);
}

.stat-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.65rem;
  margin-bottom: 2rem;
  width: 100%;
}

@media (min-width: 420px) {
  .stat-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .stat-row .stat-glow {
    grid-column: 1 / -1;
  }
}

@media (min-width: 720px) {
  .stat-row {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .stat-row .stat-glow {
    grid-column: auto;
  }
}

.stat-card {
  padding: 1rem 0.75rem;
  border-radius: 14px;
  background: var(--surface);
  border: 1px solid var(--border);
  text-align: center;
  transition: transform 0.25s, border-color 0.25s;
}
.stat-card.stat-bull:hover,
.stat-card.stat-alt:hover { border-color: var(--win); }
.stat-card:hover { transform: translateY(-4px); }

.stat-bull .stat-num,
.stat-alt .stat-num { color: var(--brand); }

.stat-num {
  display: block;
  font-size: 1.75rem;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
}

.stat-cap {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--muted);
}

.stat-glow .stat-num {
  background: linear-gradient(90deg, var(--win-bright), var(--win), var(--win-deep));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.steps {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  font-size: 0.85rem;
  color: var(--muted);
}

.step {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.step-n {
  font-weight: 800;
  color: var(--win);
  font-size: 0.75rem;
}

.step-line {
  width: 24px;
  height: 1px;
  background: linear-gradient(90deg, var(--win), transparent);
}

/* Login portal */
.portal-col {
  overflow: visible;
}

.portal {
  position: relative;
  width: 100%;
  max-width: 100%;
  border-radius: 24px;
  padding: 2px;
  background: linear-gradient(
    135deg,
    var(--win-bright),
    var(--win),
    var(--win-deep),
    var(--win-bright)
  );
  background-size: 300% 300%;
  animation: border-flow 6s ease infinite;
  box-shadow:
    0 0 56px var(--win-glow),
    0 30px 60px rgba(0, 0, 0, 0.5);
}

@keyframes border-flow {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.portal-glow {
  position: absolute;
  inset: -20px;
  background: radial-gradient(circle, var(--win-dim), transparent 65%);
  filter: blur(30px);
  z-index: -1;
  animation: pulse-glow 3s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

.portal-inner {
  background: var(--surface-raised);
  border-radius: 22px;
  padding: clamp(1.25rem, 4vw, 1.75rem);
  backdrop-filter: blur(16px);
  max-width: 100%;
}

.portal-title {
  margin: 0 0 0.35rem;
  font-size: 1.5rem;
  font-weight: 800;
  background: linear-gradient(90deg, var(--text), var(--brand));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.portal-sub {
  margin: 0 0 1.25rem;
  font-size: 0.85rem;
  color: var(--muted);
}

.portal-slot :deep(hr.divider) {
  border-color: var(--border);
  margin: 1.25rem 0;
}

.telegram-fallback {
  display: block;
  margin-top: 1.25rem;
  text-align: center;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--win-bright);
  text-decoration: none;
  transition: color 0.2s, text-shadow 0.2s;
}
.telegram-fallback:hover {
  color: var(--win);
  text-shadow: 0 0 16px var(--win-glow);
}

.scroll-hint {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 1.5rem;
  font-size: 0.75rem;
  color: var(--muted);
  justify-content: center;
}

.scroll-hint a {
  font-weight: 600;
  color: var(--win-bright);
  text-decoration: none;
}
.scroll-hint a:hover {
  color: var(--win);
}

.scroll-line {
  width: 32px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--brand));
  animation: extend 2s ease-in-out infinite;
}

@keyframes extend {
  0%, 100% { transform: scaleX(0.3); opacity: 0.4; }
  50% { transform: scaleX(1); opacity: 1; }
}

.api-offline {
  color: var(--warn);
  background: var(--warn-bg);
  border: 1px solid var(--warn-border);
  padding: 0.75rem;
  border-radius: 10px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

@media (prefers-reduced-motion: reduce) {
  .marquee-track,
  .logo-core,
  .live-dot,
  .portal,
  .portal-glow,
  .scroll-line,
  .line-flip {
    animation: none !important;
  }
}
</style>
