<template>
  <view class="page">
    <u-nav-bar title="更多功能" show-back @back="goBack" />
    <view class="grid">
      <view v-for="f in funcs" :key="f.path" class="item" @tap="go(f.path)">
        <view class="icon" :style="{ background: f.bg }">{{ f.emoji }}</view>
        <text class="name">{{ f.name }}</text>
        <text class="desc">{{ f.desc }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
const goBack = () => uni.navigateBack()
const go = (path) => uni.navigateTo({ url: path })

// 更多页 = 金刚区外的补充入口（金刚区已含的入口不重复展示）
const funcs = [
  { name: '赛事活动', desc: '竞赛报名与详情', emoji: '赛', bg: '#fff3e0', path: '/pages/competitions/list' },
  { name: '院校列表', desc: '合作院校信息', emoji: '校', bg: '#e8f5e9', path: '/pages/colleges/list' },
  { name: '应急资源', desc: '应急物资调度', emoji: '急', bg: '#fce4ec', path: '/pages/emergency/resources' },
  { name: '会员资源', desc: '产业资源台账', emoji: '资', bg: '#f3e5f5', path: '/pages/resources/list' },
]
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); }
.grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; padding: 16px; }
.item { background: var(--color-bg-card); border-radius: 8px; padding: 20px; display: flex; flex-direction: column; align-items: center; gap: 6px; }
.icon { width: 52px; height: 52px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 26px; font-weight: 600; color: #fff; }
.name { font-size: 15px; font-weight: 600; color: var(--color-text); }
.desc { font-size: 12px; color: var(--color-text-secondary); }
</style>
