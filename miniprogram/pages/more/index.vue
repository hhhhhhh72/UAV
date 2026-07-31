<template>
  <view class="page">
    <van-nav-bar title="更多功能" left-arrow @click-left="goBack" />
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

const funcs = [
  { name: '赛事活动', desc: '竞赛报名与详情', emoji: '🏆', bg: 'linear-gradient(135deg,#fff3e0,#ffb74d)', path: '/pages/events/list' },
  { name: '院校列表', desc: '合作院校信息', emoji: '🎓', bg: 'linear-gradient(135deg,#e8f5e9,#81c784)', path: '/pages/colleges/list' },
  { name: '应急资源', desc: '应急物资调度', emoji: '🆘', bg: 'linear-gradient(135deg,#fce4ec,#e57373)', path: '/pages/emergency/resources' },
  { name: '低空研学', desc: '研学课程报名', emoji: '✈️', bg: 'linear-gradient(135deg,#e3f2fd,#64b5f6)', path: '/pages/study/index' },
  { name: '会员资源', desc: '产业资源台账', emoji: '📋', bg: 'linear-gradient(135deg,#f3e5f5,#ba68c8)', path: '/pages/resources/list' },
  { name: '合规政策', desc: '政策法规文库', emoji: '📜', bg: 'linear-gradient(135deg,#e0f2f1,#4db6ac)', path: '/pages/compliance/news' },
]
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f6f7; }
.grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; padding: 16px; }
.item { background: #fff; border-radius: 12px; padding: 20px; display: flex; flex-direction: column; align-items: center; gap: 6px; }
.icon { width: 52px; height: 52px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 26px; }
.name { font-size: 15px; font-weight: 600; color: #1a1a1a; }
.desc { font-size: 12px; color: #999; }
</style>
