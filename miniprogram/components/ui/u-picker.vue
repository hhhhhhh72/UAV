<template>
  <u-popup :show="show" position="bottom" round @close="onCancel">
    <view class="u-picker">
      <view class="u-picker-bar">
        <text class="u-picker-btn" @click="onCancel">取消</text>
        <text class="u-picker-title">{{ title }}</text>
        <text class="u-picker-btn u-picker-btn--confirm" @click="onConfirm">确定</text>
      </view>
      <picker-view :value="[index]" class="u-picker-view" @change="onChange">
        <picker-view-column>
          <view v-for="(item, i) in columns" :key="i" class="u-picker-item">{{ item }}</view>
        </picker-view-column>
      </picker-view>
    </view>
  </u-popup>
</template>

<script setup>
import { ref, watch, computed, defineEmits } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  columns: { type: Array, default: () => [] },
  modelValue: { type: [String, Number], default: '' },
  title: { type: String, default: '请选择' }
})
const emit = defineEmits(['update:modelValue', 'update:show', 'confirm', 'cancel'])
const index = ref(0)
watch(() => props.modelValue, v => {
  const i = props.columns.indexOf(v)
  if (i >= 0) index.value = i
}, { immediate: true })
function onChange(e) {
  index.value = e.detail.value[0]
}
function onConfirm() {
  const v = props.columns[index.value]
  emit('update:modelValue', v)
  emit('confirm', v)
  emit('update:show', false)
}
function onCancel() {
  emit('cancel')
  emit('update:show', false)
}
</script>

<style scoped>
.u-picker { padding-bottom: env(safe-area-inset-bottom); }
.u-picker-bar { display: flex; align-items: center; justify-content: space-between; padding: 24rpx 32rpx; border-bottom: 1rpx solid var(--color-divider, #ebedf0); }
.u-picker-title { font-size: 30rpx; font-weight: 600; }
.u-picker-btn { font-size: 28rpx; color: var(--ui-color-text-secondary, #969799); }
.u-picker-btn--confirm { color: var(--color-primary, #0A66C2); font-weight: 600; }
.u-picker-view { height: 400rpx; }
.u-picker-item { display: flex; align-items: center; justify-content: center; font-size: 30rpx; color: var(--color-text, #1a1a1a); }
</style>
