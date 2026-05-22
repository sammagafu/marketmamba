<script setup>
import { computed } from 'vue'
import { ASSETS } from '../brand'

const props = defineProps({
  /** portrait | landscape | icon */
  variant: { type: String, default: 'portrait' },
  alt: { type: String, default: 'Market Mamba' },
})

const src = computed(() => {
  if (props.variant === 'landscape') return ASSETS.logoLandscape
  if (props.variant === 'icon') return ASSETS.logoIcon
  return ASSETS.logoPortrait
})

const hasWordmark = computed(() => props.variant === 'landscape' || props.variant === 'portrait')
</script>

<template>
  <img
    :src="src"
    :alt="alt"
    :class="['brand-logo', `brand-logo--${variant}`, { 'brand-logo--wordmark': hasWordmark }]"
  />
</template>

<style scoped>
.brand-logo {
  display: block;
  object-fit: contain;
}

.brand-logo--icon {
  width: 40px;
  height: 40px;
}

.brand-logo--portrait {
  width: 72px;
  height: 72px;
}

.brand-logo--landscape {
  height: clamp(40px, 11vw, 52px);
  width: auto;
  max-width: min(100%, 320px);
}

@media (min-width: 640px) {
  .brand-logo--landscape {
    max-width: min(340px, calc(100vw - 6rem));
  }
}

.brand-logo--wordmark {
  flex-shrink: 0;
}
</style>
