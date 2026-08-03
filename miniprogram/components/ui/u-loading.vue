<template>
  <view class="u-loading" :style="{ width: size, height: size, borderColor: color + '33', borderTopColor: color }" />
</template>

<script setup>
defineProps({
  size: { type: String, default: '32rpx' },
  color: { type: String, default: '#0A66C2' }
})
</script>

<style scoped>
.u-loading { border: 4rpx solid; border-radius: 50%; animation: u-loading-spin 0.8s linear infinite; }
@keyframes u-loading-spin { to { transform: rotate(360deg); } }
</style>
