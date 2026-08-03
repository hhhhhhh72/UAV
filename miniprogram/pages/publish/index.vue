<template>
<Layout :current="2">
<view class="page">
  <view class="header">发布</view>
  <view class="grid">
    <view class="btn" @tap="publish('demand')"><text class="btn-ico">求</text><text>发布需求</text></view>
    <view class="btn" @tap="publish('product')"><text class="btn-ico">品</text><text>发布商品</text></view>
    <view class="btn" @tap="publish('job')"><text class="btn-ico">聘</text><text>发布招聘</text></view>
    <view class="btn" @tap="publish('case')"><text class="btn-ico">例</text><text>发布案例</text></view>
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
.page { min-height: 100vh; background: var(--color-bg); padding: 20px; }
.header { font-size: 20px; font-weight: 700; color: var(--color-text); margin-bottom: 20px; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.btn { background: #fff; border-radius: 12px; padding: 24px; text-align: center; font-size: 16px; font-weight: 500; color: var(--color-text); box-shadow: 0 1px 3px rgba(0,0,0,.05); display: flex; align-items: center; justify-content: center; gap: 10rpx; }
.btn-ico { display: inline-flex; align-items: center; justify-content: center; width: 44rpx; height: 44rpx; border-radius: 50%; background: var(--color-primary-light); color: var(--color-primary); font-size: 20rpx; font-weight: 600; }
</style>
