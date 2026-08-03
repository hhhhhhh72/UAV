<template>
  <view class="u-cell-group" :class="{ 'u-cell-group--inset': inset }">
    <slot />
  </view>
</template>

<script setup>
defineProps({
  inset: { type: Boolean, default: false },
  border: { type: Boolean, default: true }
})
</script>

<style scoped>
.u-cell-group { background: #fff; }
.u-cell-group--inset { margin: var(--ui-space-card, 24rpx); border-radius: var(--ui-radius-card, 24rpx); overflow: hidden; box-shadow: var(--ui-shadow-card, 0 4rpx 16rpx rgba(0,0,0,0.06)); }
.u-cell-group .u-cell + .u-cell { border-top: 1rpx solid var(--color-divider, #ebedf0); }
</style>
