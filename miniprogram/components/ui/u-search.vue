<template>
  <view class="u-search">
    <view class="u-search-box">
      <u-icon name="search" size="28rpx" color="#969799" />
      <input
        class="u-search-input"
        :value="modelValue"
        :placeholder="placeholder"
        :placeholder-class="'u-search-ph'"
        confirm-type="search"
        @input="onInput"
        @confirm="onConfirm"
      />
      <view v-if="modelValue" class="u-search-clear" @click="onClear"><u-icon name="close" size="24rpx" color="#c8c9cc" /></view>
    </view>
  </view>
</template>

<script setup>
import { defineEmits } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '搜索' },
  disabled: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'search', 'change'])
function onInput(e) {
  emit('update:modelValue', e.detail.value)
  emit('change', e.detail.value)
}
function onConfirm() {
  emit('search', props.modelValue)
}
function onClear() {
  emit('update:modelValue', '')
  emit('change', '')
}
</script>

<style scoped>
.u-search { padding: 16rpx 24rpx; background: #f5f6f8; }
.u-search-box { display: flex; align-items: center; gap: 12rpx; background: #fff; border-radius: 50rpx; padding: 12rpx 24rpx; }
.u-search-input { flex: 1; font-size: 28rpx; }
.u-search-ph { color: #c8c9cc; }
.u-search-clear { padding: 4rpx; }
</style>
