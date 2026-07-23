<template>
  <div class="token" :style="{ '--tile-bg': bg }" :aria-label="`${fruitLabel}-${value}`">
    <img class="icon" :src="iconSrc" alt="" draggable="false" />
    <div class="num" aria-hidden="true">{{ value }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: Number, required: true }
})

const v = computed(() => {
  const n = Number(props.value)
  if (!Number.isFinite(n)) return 1
  return Math.max(1, Math.min(10, Math.trunc(n)))
})

// 1 葡萄、2 草莓、3 橘子、4 西瓜、5 菠萝、6 苹果、7 香蕉、8 樱桃、9 柠檬、10 桃子
const FRUIT_BY_VALUE = {
  // Higher-saturation, cute/cartoon tile colors (paired with Twemoji icons)
  1: { label: '葡萄', code: '1f347', bg: '#C79BFF' },
  2: { label: '草莓', code: '1f353', bg: '#FF6B93' },
  3: { label: '橘子', code: '1f34a', bg: '#FF9B3D' },
  4: { label: '西瓜', code: '1f349', bg: '#FF6B7A' },
  5: { label: '菠萝', code: '1f34d', bg: '#FFD04A' },
  6: { label: '苹果', code: '1f34e', bg: '#FF5C5C' },
  7: { label: '香蕉', code: '1f34c', bg: '#FFE04D' },
  8: { label: '樱桃', code: '1f352', bg: '#FF4D6D' },
  9: { label: '柠檬', code: '1f34b', bg: '#C9F25C' },
  10: { label: '桃子', code: '1f351', bg: '#FF86B3' }
}

const fruit = computed(() => FRUIT_BY_VALUE[v.value] || FRUIT_BY_VALUE[1])
const fruitLabel = computed(() => fruit.value.label)
const bg = computed(() => fruit.value.bg)
const iconSrc = computed(() => new URL(`./icons/twemoji/${fruit.value.code}.svg`, import.meta.url).href)
</script>

<style scoped>
.token {
  width: 100%;
  height: 100%;
  border-radius: 14px; /* square-ish, not circular */
  background: var(--tile-bg);
  border: 1px solid rgba(17, 24, 39, 0.10);
  box-shadow: 0 12px 22px rgba(17, 24, 39, 0.12);
  display: grid;
  place-items: center;
  position: relative;
  overflow: hidden;
}

.token::before {
  content: "";
  position: absolute;
  inset: 0;
  /* glossy highlight */
  background:
    radial-gradient(circle at 28% 22%, rgba(255, 255, 255, 0.48) 0%, rgba(255, 255, 255, 0) 52%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.10), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.token::after {
  content: "";
  position: absolute;
  top: 6%;
  bottom: 6%;
  right: 6%;
  width: 18%;
  border-radius: 12px;
  /* right-edge shading (link-link "tile depth") */
  background: linear-gradient(90deg, rgba(0, 0, 0, 0) 0%, rgba(17, 24, 39, 0.18) 100%);
  filter: blur(0.3px);
  opacity: 0.85;
  pointer-events: none;
}

.icon {
  width: 72%;
  height: 72%;
  object-fit: contain;
  filter: drop-shadow(0 10px 18px rgba(17, 24, 39, 0.14));
  position: relative;
  z-index: 1;
  user-select: none;
  pointer-events: none;
}

.num {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  font-weight: 950;
  font-size: clamp(12px, 4.6vw, 20px);
  color: rgba(255, 255, 255, 0.98);
  text-shadow: 0 2px 10px rgba(17, 24, 39, 0.34);
  font-variant-numeric: tabular-nums;
  pointer-events: none;
  z-index: 2;
}
</style>
