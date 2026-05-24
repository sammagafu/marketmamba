<script setup>
import {
  TRADE_PIPELINE,
  INDICATORS,
  RISK_CONTROLS,
  SIGNAL_SOURCES,
  RISK_DISCLAIMER,
} from '../howItWorks'
</script>

<template>
  <section id="how-we-trade" class="how-trade" aria-labelledby="how-trade-title">
    <header class="how-head copy-block">
      <p class="section-eyebrow">How it works</p>
      <h2 id="how-trade-title" class="section-title how-title">Controlled execution, end to end</h2>
      <p class="section-lead how-lead">
        Qualified signals, technical filters, and firm risk limits on the broker you already use.
        We do not copy third-party Telegram channels.
      </p>
    </header>

    <p class="section-eyebrow pipeline-label">The pipeline</p>
    <ol class="how-pipeline">
      <li v-for="item in TRADE_PIPELINE" :key="item.step" class="how-pipe-card">
        <span class="how-pipe-n">{{ item.step }}</span>
        <div>
          <h3 class="how-pipe-title">{{ item.title }}</h3>
          <p class="how-pipe-body">{{ item.body }}</p>
        </div>
      </li>
    </ol>

    <div class="how-columns">
      <div class="how-block how-block-filters">
        <header class="how-block-head">
          <p class="section-eyebrow">Technical filters</p>
          <h3 class="how-block-title">What we measure</h3>
        </header>
        <ul class="how-card-list">
          <li
            v-for="ind in INDICATORS"
            :key="ind.name"
            class="how-card"
            :class="ind.tone ? `how-card--${ind.tone}` : ''"
          >
            <div class="how-card-head">
              <span class="how-tag">{{ ind.tag }}</span>
              <strong class="how-card-name">{{ ind.name }}</strong>
            </div>
            <p class="how-card-detail">{{ ind.detail }}</p>
          </li>
        </ul>
      </div>

      <div class="how-block how-block-risk">
        <header class="how-block-head">
          <p class="section-eyebrow">Risk framework</p>
          <h3 class="how-block-title">What we enforce</h3>
        </header>
        <ul class="how-card-list">
          <li v-for="rule in RISK_CONTROLS" :key="rule.name" class="how-card how-card-risk">
            <div class="how-risk-top">
              <strong class="how-card-name">{{ rule.name }}</strong>
              <span class="how-risk-value">{{ rule.value }}</span>
            </div>
            <p class="how-card-detail">{{ rule.detail }}</p>
          </li>
        </ul>
      </div>
    </div>

    <p class="section-eyebrow sources-label">Platform</p>
    <ul class="how-sources">
      <li v-for="src in SIGNAL_SOURCES" :key="src.title" class="how-source">
        <strong class="how-source-title">{{ src.title }}</strong>
        <span class="how-source-body">{{ src.body }}</span>
      </li>
    </ul>

    <p class="how-disclaimer">{{ RISK_DISCLAIMER }}</p>
  </section>
</template>

<style scoped>
.how-trade {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 1100px;
  margin: 3rem auto 0;
  padding-top: 2.5rem;
  border-top: 1px solid var(--border);
}

.how-head {
  margin-bottom: 2rem;
}

.how-title {
  font-size: clamp(1.5rem, 5vw, 2.125rem);
  font-weight: 800;
}

.how-lead {
  margin-bottom: 0;
}

.pipeline-label,
.sources-label {
  margin-bottom: 0.75rem;
}

.how-pipeline {
  list-style: none;
  margin: 0 0 2.5rem;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}

@media (min-width: 640px) {
  .how-pipeline {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 960px) {
  .how-pipeline {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.how-pipe-card {
  display: flex;
  gap: 0.75rem;
  padding: 1rem 1.1rem;
  border-radius: 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  min-width: 0;
}

.how-pipe-n {
  flex-shrink: 0;
  font-size: 0.6875rem;
  font-weight: 800;
  color: var(--brand);
  letter-spacing: 0.06em;
  padding-top: 0.15rem;
}

.how-pipe-title {
  margin: 0 0 0.35rem;
  font-size: 0.9rem;
  font-weight: 700;
}

.how-pipe-body {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--muted);
}

.how-columns {
  display: grid;
  gap: 0;
  margin-bottom: 2.25rem;
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
  background: var(--surface);
}

@media (min-width: 900px) {
  .how-columns {
    grid-template-columns: 1fr 1fr;
  }
}

.how-block {
  padding: 1.35rem 1.25rem 1.25rem;
  min-width: 0;
}

@media (max-width: 899px) {
  .how-block-filters {
    border-bottom: 1px solid var(--border);
  }
}

@media (min-width: 900px) {
  .how-block-filters {
    border-right: 1px solid var(--border);
  }
}

.how-block-filters {
  background: linear-gradient(180deg, var(--brand-muted) 0%, transparent 28%);
}

.how-block-risk {
  background: var(--surface-raised);
}

.how-block-head {
  margin-bottom: 1.1rem;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid var(--border);
}

.how-block-filters .section-eyebrow {
  color: var(--brand);
}

.how-block-title {
  margin: 0.2rem 0 0;
  font-size: 1.0625rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.how-card-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.5rem;
}

.how-card {
  padding: 0.85rem 0.95rem;
  border-radius: 10px;
  background: var(--bg);
  border: 1px solid var(--border);
  transition: border-color 0.15s ease;
}

.how-block-risk .how-card {
  background: var(--surface);
}

.how-card:hover {
  border-color: var(--border-strong);
}

.how-card--trend { border-left: 2px solid rgba(61, 255, 122, 0.55); }
.how-card--momentum { border-left: 2px solid rgba(94, 179, 255, 0.45); }
.how-card--volatility { border-left: 2px solid rgba(192, 132, 252, 0.4); }
.how-card--cost { border-left: 2px solid rgba(251, 191, 36, 0.45); }

.how-card-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem 0.65rem;
  margin-bottom: 0.4rem;
}

.how-tag {
  display: inline-block;
  padding: 0.18rem 0.45rem;
  font-size: 0.5625rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--brand);
  background: var(--brand-soft);
  border-radius: 4px;
  flex-shrink: 0;
}

.how-card-name {
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.3;
}

.how-card-detail {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.55;
  color: var(--muted);
}

.how-card-risk {
  border-left: 2px solid var(--brand);
}

.how-risk-top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem 0.75rem;
  margin-bottom: 0.35rem;
}

.how-risk-value {
  flex-shrink: 0;
  padding: 0.15rem 0.5rem;
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--brand);
  font-variant-numeric: tabular-nums;
  border: 1px solid var(--brand-muted);
  border-radius: 6px;
  background: var(--brand-soft);
}

.how-sources {
  list-style: none;
  margin: 0 0 1.5rem;
  padding: 0;
  display: grid;
  gap: 0;
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
}

.how-source {
  display: grid;
  gap: 0.25rem;
  padding: 0.85rem 1rem;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
}

.how-source:last-child {
  border-bottom: none;
}

@media (min-width: 768px) {
  .how-source {
    grid-template-columns: minmax(9rem, 34%) 1fr;
    gap: 1rem;
    align-items: baseline;
  }
}

.how-source-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text);
  line-height: 1.35;
}

.how-source-body {
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--muted);
}

.how-disclaimer {
  margin: 0;
  padding: 1rem 1.1rem;
  font-size: 0.75rem;
  line-height: 1.55;
  color: var(--muted);
  border-radius: 10px;
  background: var(--surface);
  border: 1px solid var(--border);
}
</style>
