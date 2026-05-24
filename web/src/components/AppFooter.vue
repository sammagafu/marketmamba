<script setup>
import { computed } from 'vue'
import BrandLogo from './BrandLogo.vue'
import { TAGLINE, SLOGAN_SHORT, PAYMENT_NOTE } from '../brand'
import { RISK_DISCLAIMER } from '../howItWorks'

const props = defineProps({
  landing: { type: Boolean, default: false },
  botUsername: { type: String, default: 'market_mamba_bot' },
  contactUrl: { type: String, default: '' },
  contactLabel: { type: String, default: 'Contact us' },
})

const year = computed(() => new Date().getFullYear())
const telegramUrl = computed(() => `https://t.me/${props.botUsername}`)
</script>

<template>
  <footer class="site-footer">
    <div class="footer-inner">
      <div class="footer-brand">
        <BrandLogo variant="icon" class="footer-logo" alt="" />
        <div>
          <p class="footer-name">Market Mamba</p>
          <p class="footer-tagline">{{ TAGLINE }}</p>
          <p class="footer-slogan muted">{{ SLOGAN_SHORT }}</p>
        </div>
      </div>

      <nav class="footer-nav" aria-label="Footer">
        <p class="footer-nav-label">Resources</p>
        <ul class="footer-links">
          <li v-if="landing">
            <a href="#how-we-trade">How it works</a>
          </li>
          <li v-if="landing">
            <a href="#login-portal">Sign in</a>
          </li>
          <li>
            <a :href="telegramUrl" target="_blank" rel="noopener">Telegram @{{ botUsername }}</a>
          </li>
          <li v-if="contactUrl">
            <a :href="contactUrl" target="_blank" rel="noopener">{{ contactLabel }}</a>
          </li>
        </ul>
        <p class="footer-pay muted">{{ PAYMENT_NOTE }}</p>
      </nav>

      <div class="footer-meta">
        <p class="footer-copy">© {{ year }} Market Mamba</p>
        <p class="footer-risk">{{ RISK_DISCLAIMER }}</p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  position: relative;
  z-index: 2;
  margin-top: auto;
  width: 100%;
  border-top: 1px solid var(--border);
  background: var(--header-bg);
  backdrop-filter: blur(12px);
  padding: 2rem var(--page-pad-right) max(2.5rem, env(safe-area-inset-bottom)) var(--page-pad);
}

.footer-inner {
  max-width: 1100px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.75rem;
}

@media (min-width: 640px) {
  .footer-inner {
    gap: 2rem;
  }
}

@media (min-width: 768px) {
  .footer-inner {
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 0.6fr) minmax(0, 1.4fr);
    align-items: start;
    gap: 2.5rem;
  }
}

.footer-brand {
  display: flex;
  gap: 0.85rem;
  align-items: flex-start;
  min-width: 0;
}

.footer-logo {
  flex-shrink: 0;
}

.footer-name {
  margin: 0 0 0.2rem;
  font-size: 1rem;
  font-weight: 800;
}

.footer-tagline {
  margin: 0 0 0.35rem;
  font-size: 0.85rem;
  color: var(--win-bright);
  font-weight: 600;
}

.footer-slogan {
  margin: 0;
  font-size: 0.8rem;
}

.footer-nav-label {
  margin: 0 0 0.65rem;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.footer-links {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.footer-links a {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-soft);
  text-decoration: none;
  transition: color 0.2s;
}

.footer-links a:hover {
  color: var(--win-bright);
}

.footer-pay {
  margin: 0.75rem 0 0;
  font-size: 0.75rem;
  line-height: 1.4;
}

.footer-meta {
  min-width: 0;
}

.footer-copy {
  margin: 0 0 0.75rem;
  font-size: 0.8rem;
  color: var(--muted);
}

.footer-risk {
  margin: 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--muted);
  max-width: 36rem;
}
</style>
