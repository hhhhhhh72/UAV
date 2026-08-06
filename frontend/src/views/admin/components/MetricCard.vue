<template>
  <div class="metric-card">
    <div class="metric-label">{{ label }}</div>
    <div class="metric-value" :style="valueColor ? { color: valueColor } : {}">
      {{ displayValue }}
    </div>
    <div v-if="sub" class="metric-sub">{{ sub }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  label: { type: String, required: true },
  value: { type: [Number, String], default: 0 },
  sub: { type: String, default: '' },
  valueColor: { type: String, default: '' }
})

const displayValue = computed(() => {
  if (typeof props.value === 'number') {
    return props.value.toLocaleString()
  }
  return props.value
})
</script>

<style scoped>
.metric-card {
  background: var(--color-bg-2);
  border-radius: 4px;
  padding: 20px;
  box-shadow: none;
  transition: box-shadow 0.2s ease;
}

.metric-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.metric-label {
  font-size: 13px;
  color: var(--color-text-3);
  margin-bottom: 8px;
  font-weight: 500;
}

.metric-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--color-text-1);
  line-height: 1.2;
  letter-spacing: -0.5px;
}

.metric-sub {
  font-size: 12px;
  color: var(--text-tertiary, #aeaeb2);
  margin-top: 4px;
}

@media (max-width: 767px) {
  .metric-card {
    padding: 14px;
  }
  .metric-value {
    font-size: 22px;
  }
}
</style>
