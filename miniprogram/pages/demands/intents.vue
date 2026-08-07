<template>
  <view class="intents-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">收到的对接意向</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">等待处理</text>
      <text class="list-count">共 {{ intents.length }} 条</text>
    </view>

    <!-- 空状态 -->
    <view v-if="intents.length === 0" class="state-panel">
      <view class="state-mark">⌁</view>
      <text class="state-title">暂无对接意向</text>
      <text class="state-desc">发布的信息收到登记后，会第一时间展示在这里</text>
    </view>

    <!-- 意向列表 -->
    <view v-else class="intent-list">
      <view v-for="intent in intents" :key="intent.id" class="intent-card">
        <view class="intent-head">
          <view class="intent-avatar"><text>{{ intent.initial }}</text></view>
          <view class="intent-copy">
            <text class="intent-name">{{ intent.name }}</text>
            <text class="tag" :class="intentStatusClass(intent.status)">{{ intent.status }}</text>
          </view>
        </view>
        <text class="intent-detail">对接项目：{{ intent.target }}</text>
        <text class="intent-note">对方说明：{{ intent.note }}</text>
        <view class="intent-actions">
          <template v-if="intent.status === '待处理'">
            <view class="intent-btn" @tap="markViewed(intent)">标记已查看</view>
            <view class="intent-btn primary" @tap="startChat(intent)">发起洽谈</view>
          </template>
          <template v-else-if="intent.status === '洽谈中'">
            <view class="intent-btn primary" @tap="startChat(intent)">继续洽谈</view>
          </template>
          <template v-else>
            <view class="intent-btn" @tap="toastArchived">查看记录</view>
          </template>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { safeNavigateTo } from '../../utils/nav'
import { getReceivedIntents, saveReceivedIntents } from '../../utils/hallData'

const intents = ref(getReceivedIntents())

function intentStatusClass(s) {
  return s === '已合作' ? 'green' : s === '未合作' ? 'gray' : 'orange'
}

function markViewed(intent) {
  intent.status = '已查看'
  saveReceivedIntents(intents.value)
  uni.showToast({ title: '已标记为已查看', icon: 'success' })
}

function startChat(intent) {
  intent.status = '洽谈中'
  saveReceivedIntents(intents.value)
  uni.setStorageSync('hall_chat_partner', JSON.stringify({ name: intent.name, initial: intent.initial }))
  safeNavigateTo('/pages/demands/chat')
}

const toastArchived = () => {
  uni.showToast({ title: '该意向记录已归档', icon: 'none' })
}

const goBack = () => uni.navigateBack()
</script>

<style scoped>
.intents-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 40rpx;
}

.page-header {
  height: 56px;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-spacer { width: 72rpx; }

.list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

.intent-list { padding: 0 32rpx 32rpx; }
.intent-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 26rpx;
  border: 1px solid #EEF1F4;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.intent-card + .intent-card { margin-top: 20rpx; }

.intent-head { display: flex; gap: 20rpx; align-items: center; }
.intent-avatar {
  width: 76rpx;
  height: 76rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 28rpx;
  font-weight: 750;
  flex-shrink: 0;
}
.intent-copy { display: flex; flex-direction: column; gap: 8rpx; }
.intent-name { font-size: 26rpx; font-weight: 700; color: #17212B; }
.tag {
  align-self: flex-start;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.gray { color: #667085; background: #F1F3F5; }

.intent-detail { display: block; font-size: 24rpx; color: #17212B; margin-top: 18rpx; }
.intent-note { display: block; font-size: 22rpx; color: #667085; line-height: 1.6; margin-top: 8rpx; }

.intent-actions {
  display: flex;
  gap: 14rpx;
  border-top: 1px solid #EEF1F4;
  margin-top: 22rpx;
  padding-top: 20rpx;
}
.intent-btn {
  flex: 1;
  height: 62rpx;
  border-radius: 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  color: #0A66C2;
  border: 1px solid #C7DEF1;
  background: #fff;
}
.intent-btn.primary { color: #fff; background: #0A66C2; border-color: #0A66C2; }

.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }
</style>
