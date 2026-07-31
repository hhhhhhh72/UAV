<template>
  <view class="tabbar-root">
    <van-tabbar
      :active="current"
      active-color="#1a1a1a"
      inactive-color="#969799"
      @change="handleTabChange"
      safe-area-inset-bottom
    >
      <van-tabbar-item icon="home-o" :name="0"><text>首页</text></van-tabbar-item>
      <van-tabbar-item icon="shopping-cart-o" :name="1"><text>商城</text></van-tabbar-item>
      <van-tabbar-item icon="add-o" :name="2"><text>发布</text></van-tabbar-item>
      <van-tabbar-item icon="shop-o" :name="3"><text>商家</text></van-tabbar-item>
      <van-tabbar-item icon="user-o" :name="4"><text>我的</text></van-tabbar-item>
    </van-tabbar>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps(['current'])

const tabPages = [
  '/pages/home/index',
  '/pages/mall/index',
  '/pages/publish/index',
  '/pages/shops/index',
  '/pages/mine/index',
]

const handleTabChange = (e) => {
  const index = typeof e.detail === 'object' ? e.detail.index : e.detail
  const idx = Number(index)
  if (idx === Number(props.current)) return
  uni.switchTab({ url: tabPages[idx] })
}
</script>

<style scoped>
.tabbar-root { position: fixed; bottom: 0; left: 0; right: 0; z-index: 999; }
</style>
