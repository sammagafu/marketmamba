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
      <div class="how-block">
        <p class="section-eyebrow">Technical filters</p>
        <h3 class="how-block-title">What we measure</h3>
        <ul class="how-card-list">
          <li v-for="ind in INDICATORS" :key="ind.name" class="how-card">
            <span class="how-tag">{{ ind.tag }}</span>
            <strong class="how-card-name">{{ ind.name }}</strong>
            <p class="how-card-detail">{{ ind.detail }}</p>
          </li>
        </ul>
      </div>

      <div class="how-block">
        <p class="section-eyebrow">Risk framework</p>
        <h3 class="how-block-title">What we enforce</h3>
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
  margin: 3rem auto 0;
  padding-top: 2.5rem;
  border-top: 1px solid var(--border);
}

.how-head {
  margin-bottom: 2.5rem;
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
  margin-bottom: 1rem;
}

.how-pipeline {
  list-style: none;
  margin: 0 0 3rem;
  padding: 0;
  display: grid;
  gap: 1rem;
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
  gap: 0.85rem;
  padding: 1.15rem 1.2rem;
  border-radius: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  min-width: 0;
}

.how-pipe-n {
  flex-shrink: 0;
  font-size: 0.6875rem;
  font-weight: 800;
  color: var(--win);
  letter-spacing: 0.08em;
}

.how-pipe-title {
  margin: 0 0 0.4rem;
  font-size: 0.9375rem;
  font-weight: 700;
}

.how-pipe-body {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.55;
  color: var(--muted);
}

.how-columns {
  display: grid;
  gap: 2.5rem;
  margin-bottom: 2.5rem;
}

@media (min-width: 900px) {
  .how-columns {
    grid-template-columns: 1fr 1fr;
    gap: 3rem;
  }
}

.how-block-title {
  margin: 0 0 1.15rem;
  font-size: 1.0625rem;
  font-weight: 700;
}

.how-card-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}

.how-card {
  padding: 1rem 1.1rem;
  border-radius: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
}

.how-tag {
  display: inline-block;
  margin-bottom: 0.4rem;
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--win-bright);
}

.how-card-name {
  display: block;
  font-size: 0.9rem;
  margin-bottom: 0.3rem;
}

.how-card-detail {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--muted);
}

.how-risk-top {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.35rem 0.75rem;
  margin-bottom: 0.3rem;
}

.how-risk-value {
  font-size: 0.875rem;
  font-weight: 800;
  color: var(--brand);
  font-variant-numeric: tabular-nums;
}

.how-sources {
  display: grid;
  gap: 1rem;
  margin-bottom: 2rem;
}

@media (min-width: 640px) {
  .how-sources {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.how-source {
  padding: 1.15rem 1.2rem;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--surface);
}

.how-source-title {
  margin: 0 0 0.55rem;
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1.35;
}

.how-source-body {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.55;
  color: var(--muted);
}

.how-disclaimer {
  margin: 0;
  padding: 1.15rem 1.2rem;
  font-size: 0.75rem;
  line-height: 1.55;
  color: var(--muted);
  border-radius: 10px;
  background: var(--warn-bg);
  border: 1px solid var(--warn-border);
}
</style>
