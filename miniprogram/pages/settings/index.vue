<template>
  <view class="settings-page">
    <u-nav-bar title="设置" show-back @back="goBack" />

    <!-- 账号与安全 -->
    <view class="page-section">
      <view class="section-title"><text class="section-title-text">账号与安全</text></view>
      <view class="card">
        <MineCellGroup :items="accountItems" @select="onAccountSelect" />
      </view>
    </view>

    <!-- 通知与隐私 -->
    <view class="page-section">
      <view class="section-title"><text class="section-title-text">通知与隐私</text></view>
      <view class="card">
        <MineCellGroup :items="privacyItems" @select="onPrivacySelect" />
      </view>
    </view>

    <!-- 关于平台 -->
    <view class="about-foot">
      <view class="about-logo">
        <image class="about-logo-img" :src="icons.drone" mode="aspectFit" />
      </view>
      <view class="about-copy">
        <text class="about-name">无人机产业综合服务平台</text>
        <text class="about-meta">版本 1.0.0 · 运营方：重庆市无人机产业协会</text>
        <text class="about-copy-right">© 2026 重庆市无人机产业协会</text>
      </view>
    </view>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import MineCellGroup from '@/components/mine/MineCellGroup.vue'

const icons = {
  drone: '/static/mine-icons/drone.svg',
}

const goBack = () => uni.navigateBack()

// 导航目标：全部为真实页面（无"未开放"占位）
const accountItems = computed(() => [
  { icon: '/static/mine-icons/account.svg', tone: 'primary', label: '账号信息', desc: '手机号、微信绑定与个人资料', go: goProfile },
  { icon: '/static/mine-icons/certification.svg', tone: 'primary', label: '实名认证', desc: '实名信息与认证状态', go: goAuth },
])

const privacyItems = computed(() => [
  { icon: '/static/mine-icons/message.svg', tone: 'primary', label: '消息通知', desc: '系统公告与服务通知', go: goMessages },
  { icon: '/static/mine-icons/doc.svg', tone: 'gray', label: '用户协议', desc: '平台服务条款', go: goAgreement },
  { icon: '/static/mine-icons/doc.svg', tone: 'gray', label: '隐私政策', desc: '个人信息保护说明', go: goPrivacy },
])

const onAccountSelect = (i) => { const it = accountItems.value[i]; if (it && it.go) it.go() }
const onPrivacySelect = (i) => { const it = privacyItems.value[i]; if (it && it.go) it.go() }

const goProfile = () => uni.navigateTo({ url: '/pages/mine/profile' })
const goAuth = () => uni.navigateTo({ url: '/pages/mine/auth' })
const goMessages = () => uni.navigateTo({ url: '/pages/messages/index' })
const goAgreement = () => uni.navigateTo({ url: '/pages/agreement/index' })
const goPrivacy = () => uni.navigateTo({ url: '/pages/agreement/index?type=privacy' })
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(24px + env(safe-area-inset-bottom));
}

.page-section {
  margin: 20rpx 24rpx 0;
}
.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 32px;
  padding: 0 8rpx 16rpx;
}
.section-title-text {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
}

.card {
  background: #fff;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  box-shadow: 0 8rpx 32rpx rgba(16,24,40,.06);
  overflow: hidden;
}

/* ===== 关于平台 ===== */
.about-foot {
  margin-top: 56rpx;
  padding: 0 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.about-logo {
  width: 104rpx;
  height: 104rpx;
  border-radius: 28rpx;
  background: linear-gradient(145deg, #E8F2FC, #DCEBFB);
  display: flex;
  align-items: center;
  justify-content: center;
}
.about-logo-img {
  width: 60rpx;
  height: 60rpx;
}
.about-copy {
  margin-top: 20rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10rpx;
}
.about-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
}
.about-meta {
  font-size: 22rpx;
  color: #667085;
}
.about-copy-right {
  font-size: 20rpx;
  color: #98A2B3;
}

.bottom-spacer { height: 8px; }
</style>
