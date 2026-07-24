<template>
  <view class="layout">
    <slot></slot>
    <HomeFloatButton />
    <TabBar :current="current" />
  </view>
</template>

<script setup>
import { onMounted } from 'vue'
import TabBar from './TabBar.vue'
import HomeFloatButton from './HomeFloatButton.vue'
defineProps(['current'])

onMounted(() => {
  // 仅在主 tab 页面显示时调用
  uni.hideTabBar({ animation: false }).catch(() => {})
})
</script>

<style scoped>
.layout { min-height: 100vh; background: #f5f5f5; }
</style>
