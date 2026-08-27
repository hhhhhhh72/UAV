<template>
  <view class="u-field" :class="{ 'u-field--textarea': type === 'textarea' }">
    <text v-if="label" class="u-field-label">{{ label }}</text>
    <input
      v-if="type !== 'textarea'"
      class="u-field-input"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :placeholder-class="'u-field-ph'"
      :disabled="disabled"
      @input="onInput"
    />
    <textarea
      v-else
      class="u-field-input u-field-textarea"
      :value="modelValue"
      :placeholder="placeholder"
      :placeholder-class="'u-field-ph'"
      :disabled="disabled"
      :auto-height="autoHeight"
      @input="onInput"
    />
  </view>
</template>

<script setup>
const props = defineProps({
  label: { type: String, default: '' },
  modelValue: { type: [String, Number], default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  autoHeight: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])
function onInput(e) {
  emit('update:modelValue', e.detail.value)
}
</script>

<style scoped>
.u-field { display: flex; align-items: center; background: #fafafa; border-radius: 24rpx; padding: 20rpx 24rpx; gap: 16rpx; }
.u-field-label { font-size: var(--ui-font-size-md, 30rpx); color: var(--color-text, #1a1a1a); flex-shrink: 0; }
.u-field-input { flex: 1; font-size: var(--ui-font-size-md, 30rpx); color: var(--color-text, #1a1a1a); }
.u-field-textarea { min-height: 120rpx; }
.u-field-ph { color: var(--color-text-placeholder, #c8c9cc); }
.u-field--textarea { align-items: flex-start; }
</style>
