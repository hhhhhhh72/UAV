<template>
<Layout :current="2">
<view class="page">
  <view class="header">发布</view>
  <view class="grid">
    <view class="btn" @tap="publish('demand')">📋 发布需求</view>
    <view class="btn" @tap="publish('product')">🛒 发布商品</view>
    <view class="btn" @tap="publish('job')">💼 发布招聘</view>
    <view class="btn" @tap="publish('case')">🏆 发布案例</view>
  </view>
</view>
</Layout>
</template>
<script setup>
import Layout from '../../components/Layout.vue'
const publish = (type) => {
  uni.showToast({ title: type + '发布开发中', icon: 'none' })
}
</script>
<style scoped>
.page { min-height: 100vh; background: #f2f5f7; padding: 20px; }
.header { font-size: 20px; font-weight: 700; color: #1a1a1a; margin-bottom: 20px; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.btn { background: #fff; border-radius: 12px; padding: 24px; text-align: center; font-size: 16px; font-weight: 500; color: #1a1a1a; box-shadow: 0 1px 3px rgba(0,0,0,.05); }
</style>
