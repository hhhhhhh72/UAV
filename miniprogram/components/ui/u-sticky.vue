<template>
  <view class="u-sticky" :style="{ top: offsetTop + 'px', zIndex }">
    <slot />
  </view>
</template>

<script setup>
defineProps({
  offsetTop: { type: [String, Number], default: 0 },
  zIndex: { type: Number, default: 99 }
})
</script>

<style scoped>
.u-sticky { position: sticky; }
</style>
