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
    <header class="how-head">
      <p class="how-eyebrow">Transparency</p>
      <h2 id="how-trade-title" class="how-title">How we take trades</h2>
      <p class="how-lead">
        Market Mamba is built around qualified signals, technical filters, and hard risk limits —
        not gut feel. Here is what happens from scan to close.
      </p>
    </header>

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
      <div class="how-block">
        <h3 class="how-block-title">Indicators &amp; filters</h3>
        <ul class="how-card-list">
          <li v-for="ind in INDICATORS" :key="ind.name" class="how-card">
            <span class="how-tag">{{ ind.tag }}</span>
            <strong class="how-card-name">{{ ind.name }}</strong>
            <p class="how-card-detail">{{ ind.detail }}</p>
          </li>
        </ul>
      </div>

      <div class="how-block">
        <h3 class="how-block-title">Risk &amp; limits</h3>
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

    <div class="how-sources">
      <article v-for="src in SIGNAL_SOURCES" :key="src.title" class="how-source">
        <h3 class="how-source-title">{{ src.title }}</h3>
        <p class="how-source-body">{{ src.body }}</p>
      </article>
    </div>

    <p class="how-disclaimer">{{ RISK_DISCLAIMER }}</p>
  </section>
</template>

<style scoped>
.how-trade {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 1100px;
  margin: 2rem auto 0;
  padding-top: 2rem;
  border-top: 1px solid var(--border);
}

.how-head {
  margin-bottom: 2rem;
  max-width: 42rem;
}

.how-eyebrow {
  margin: 0 0 0.5rem;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--win-bright);
}

.how-title {
  margin: 0 0 0.75rem;
  font-size: clamp(1.5rem, 5vw, 2rem);
  font-weight: 800;
  line-height: 1.15;
}

.how-lead {
  margin: 0;
  font-size: 1rem;
  line-height: 1.6;
  color: var(--muted);
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
  padding: 1rem;
  border-radius: 14px;
  background: var(--surface);
  border: 1px solid var(--border);
  min-width: 0;
}

.how-pipe-n {
  flex-shrink: 0;
  font-size: 0.7rem;
  font-weight: 800;
  color: var(--win);
  letter-spacing: 0.06em;
}

.how-pipe-title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 700;
}

.how-pipe-body {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.5;
  color: var(--muted);
}

.how-columns {
  display: grid;
  gap: 2rem;
  margin-bottom: 2rem;
}

@media (min-width: 900px) {
  .how-columns {
    grid-template-columns: 1fr 1fr;
    gap: 2.5rem;
  }
}

.how-block-title {
  margin: 0 0 1rem;
  font-size: 1.1rem;
  font-weight: 700;
}

.how-card-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.65rem;
}

.how-card {
  padding: 0.85rem 1rem;
  border-radius: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
}

.how-tag {
  display: inline-block;
  margin-bottom: 0.35rem;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--win-bright);
}

.how-card-name {
  display: block;
  font-size: 0.9rem;
  margin-bottom: 0.25rem;
}

.how-card-detail {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.45;
  color: var(--muted);
}

.how-risk-top {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.35rem 0.75rem;
  margin-bottom: 0.25rem;
}

.how-risk-value {
  font-size: 0.85rem;
  font-weight: 800;
  color: var(--brand);
  font-variant-numeric: tabular-nums;
}

.how-sources {
  display: grid;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

@media (min-width: 720px) {
  .how-sources {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.how-source {
  padding: 1rem 1.1rem;
  border-radius: 14px;
  border: 1px solid var(--border);
  background: linear-gradient(160deg, var(--surface), var(--surface-raised));
}

.how-source-title {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text);
}

.how-source-body {
  margin: 0;
  font-size: 0.82rem;
  line-height: 1.5;
  color: var(--muted);
}

.how-disclaimer {
  margin: 0;
  padding: 1rem 1.1rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--muted);
  border-radius: 10px;
  background: var(--warn-bg);
  border: 1px solid var(--warn-border);
}
</style>
