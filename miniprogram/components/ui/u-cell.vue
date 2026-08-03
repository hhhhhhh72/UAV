<template>
  <view class="u-cell" :class="{ 'u-cell--clickable': isClickable || isLink }" @click="onClick">
    <view v-if="$slots.icon || icon" class="u-cell-icon"><slot name="icon"><u-icon v-if="icon" :name="icon" :size="iconSize" /></slot></view>
    <view class="u-cell-body">
      <view class="u-cell-title-row">
        <!-- 微信 WXML 中 <text> 内不能含 <view>，标题插槽改为 view 以支持富内容插槽 -->
        <view class="u-cell-title" :style="{ color: titleColor }"><slot name="title">{{ title }}</slot></view>
        <u-tag v-if="tag" size="mini" :type="tagType" plain>{{ tag }}</u-tag>
      </view>
      <text v-if="label" class="u-cell-label">{{ label }}</text>
    </view>
    <view v-if="value || $slots.value" class="u-cell-value"><slot name="value">{{ value }}</slot></view>
    <u-icon v-if="isLink" name="arrow" size="26rpx" color="#c8c9cc" />
  </view>
</template>

<script setup>
import { defineEmits } from 'vue'

const props = defineProps({
  title: { type: String, default: '' },
  value: { type: String, default: '' },
  label: { type: String, default: '' },
  icon: { type: String, default: '' },
  iconSize: { type: String, default: '36rpx' },
  isLink: { type: Boolean, default: false },
  isClickable: { type: Boolean, default: false },
  titleColor: { type: String, default: '' },
  tag: { type: String, default: '' },
  tagType: { type: String, default: 'primary' }
})
const emit = defineEmits(['click'])
function onClick() {
  if (props.isClickable || props.isLink) emit('click')
}
</script>

<style scoped>
.u-cell { display: flex; align-items: center; padding: 24rpx var(--ui-space-card, 24rpx); background: #fff; gap: 16rpx; }
.u-cell--clickable { cursor: pointer; }
.u-cell-icon { flex-shrink: 0; }
.u-cell-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.u-cell-title-row { display: flex; align-items: center; gap: 12rpx; }
.u-cell-title { font-size: var(--ui-font-size-md, 30rpx); color: var(--color-text, #1a1a1a); }
.u-cell-label { font-size: var(--ui-font-size-sm, 26rpx); color: var(--ui-color-text-secondary, #969799); }
.u-cell-value { font-size: var(--ui-font-size-md, 30rpx); color: var(--ui-color-text-secondary, #969799); flex-shrink: 0; }
</style>
