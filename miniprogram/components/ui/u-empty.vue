<template>
  <view class="u-empty">
    <view class="u-empty-art">
      <view class="u-empty-circle" />
      <view class="u-empty-line u-empty-line--1" />
      <view class="u-empty-line u-empty-line--2" />
    </view>
    <text v-if="description" class="u-empty-desc">{{ description }}</text>
    <view v-if="$slots.default" class="u-empty-action"><slot /></view>
  </view>
</template>

<script setup>
defineProps({
  description: { type: String, default: '暂无数据' },
  image: { type: String, default: '' }
})
</script>

<style scoped>
.u-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80rpx 40rpx; gap: 24rpx; }
.u-empty-art { position: relative; width: 160rpx; height: 120rpx; }
.u-empty-circle { position: absolute; top: 0; left: 40rpx; width: 80rpx; height: 80rpx; border-radius: 50%; background: var(--color-primary-light, #E8F2FC); }
.u-empty-line { position: absolute; height: 12rpx; border-radius: 6rpx; background: var(--color-primary-light, #E8F2FC); }
.u-empty-line--1 { bottom: 16rpx; left: 20rpx; width: 120rpx; }
.u-empty-line--2 { bottom: 0; left: 50rpx; width: 60rpx; }
.u-empty-desc { font-size: var(--ui-font-size-md, 30rpx); color: var(--ui-color-text-secondary, #969799); }
.u-empty-action { margin-top: 8rpx; }
</style>
