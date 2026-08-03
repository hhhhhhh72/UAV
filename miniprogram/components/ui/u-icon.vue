<template>
  <view class="u-icon" :class="'u-icon--' + name" :style="{ width: size, height: size, borderColor: color }">
    <view v-if="name === 'arrow'" class="u-icon-arrow" :style="{ borderColor: color }" />
    <view v-else-if="name === 'close'" class="u-icon-close" :style="{ background: color }" />
    <view v-else-if="name === 'plus'" class="u-icon-plus" :style="{ background: color }" />
    <view v-else-if="name === 'check'" class="u-icon-check" :style="{ borderColor: color }" />
    <view v-else-if="name === 'search'" class="u-icon-search" :style="{ borderColor: color }" />
    <view v-else-if="name === 'back'" class="u-icon-arrow u-icon-arrow--left" :style="{ borderColor: color }" />
    <view v-else-if="name === 'location'" class="u-icon-location" :style="{ borderColor: color, background: color }" />
    <view v-else-if="name === 'success'" class="u-icon-success" :style="{ borderColor: color }"><text class="u-icon-success-mark">✓</text></view>
    <text v-else class="u-icon-char" :style="{ color, fontSize: size }">{{ name }}</text>
  </view>
</template>

<script setup>
defineProps({
  name: { type: String, default: '' },
  size: { type: String, default: '32rpx' },
  color: { type: String, default: '#969799' }
})
</script>

<style scoped>
.u-icon { position: relative; display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; }
.u-icon-arrow { width: 14rpx; height: 14rpx; border-right: 4rpx solid; border-top: 4rpx solid; transform: rotate(45deg); border-radius: 2rpx; }
.u-icon-arrow--left { transform: rotate(-135deg); }
.u-icon-close { width: 28rpx; height: 4rpx; border-radius: 4rpx; transform: rotate(45deg); }
.u-icon-close::after { content: ''; position: absolute; width: 28rpx; height: 4rpx; border-radius: 4rpx; background: inherit; transform: rotate(90deg); }
.u-icon-plus { width: 28rpx; height: 4rpx; border-radius: 4rpx; }
.u-icon-plus::after { content: ''; position: absolute; width: 28rpx; height: 4rpx; border-radius: 4rpx; background: inherit; transform: rotate(90deg); }
.u-icon-check { width: 24rpx; height: 14rpx; border-left: 4rpx solid; border-bottom: 4rpx solid; transform: rotate(-45deg); border-radius: 2rpx; }
.u-icon-search { width: 20rpx; height: 20rpx; border: 4rpx solid; border-radius: 50%; }
.u-icon-search::after { content: ''; position: absolute; right: -10rpx; bottom: -6rpx; width: 14rpx; height: 4rpx; border-radius: 4rpx; border-bottom: 4rpx solid; border-bottom-color: inherit; transform: rotate(45deg); }
.u-icon-location { width: 20rpx; height: 20rpx; border: 4rpx solid; border-radius: 50% 50% 50% 0; transform: rotate(-45deg); }
.u-icon-success { width: 28rpx; height: 28rpx; border: 4rpx solid; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
.u-icon-success-mark { font-size: 18rpx; line-height: 1; }
.u-icon-char { line-height: 1; }
</style>
