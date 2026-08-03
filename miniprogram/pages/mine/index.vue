<template>
  <Layout :current="4">
    <view class="mine-page">

      <!-- ===== 头部区域（仿参考图：浅灰蓝渐变 + 头像 + 三个小图标） ===== -->
      <view class="header-section">
        <!-- 装饰背景：右上方光斑 -->
        <view class="header-light header-light-1"></view>
        <view class="header-light header-light-2"></view>

        <view class="header-main">
          <!-- 左侧头像 + 姓名 + 认证徽章 -->
          <view class="user-block" @tap="handleUserClick">
            <view class="avatar-wrap">
              <image
                v-if="user && user.avatar"
                class="avatar"
                :src="avatarSrc(user.avatar)"
                mode="aspectFill"
              />
              <view v-else class="avatar avatar-placeholder">
                <text class="avatar-text">{{ userInitial }}</text>
              </view>
            </view>

            <view class="user-info-col">
              <text class="user-name">{{ user ? (user.name || user.phone || '微信用户') : '点击登录' }}</text>
              <view v-if="user" class="user-meta">
                <view class="cert-pill">
                  <text class="cert-pill-icon">飞</text>
                  <text class="cert-pill-text">飞手</text>
                </view>
                <text v-if="userRoleLabel" class="role-text">· {{ userRoleLabel }}</text>
              </view>
              <text v-else class="login-hint">登录后享受更多服务</text>
            </view>

            <text v-if="!user" class="login-arrow">›</text>
          </view>

          <!-- 右侧三个图标按钮（参考图：APP / 通知 / 设置） -->
          <view class="header-icon-row">
            <view class="hdr-icon-btn" @tap="openApp" v-if="user">
              <view class="hdr-icon-circle">
                <text class="hdr-icon-glyph">APP</text>
              </view>
            </view>
            <view class="hdr-icon-btn" @tap="goMessages">
              <view class="hdr-icon-circle">
                <text class="hdr-icon-glyph">信</text>
                <view v-if="unreadCount > 0" class="hdr-dot"></view>
              </view>
            </view>
            <view class="hdr-icon-btn" @tap="goSettings">
              <view class="hdr-icon-circle">
                <text class="hdr-icon-glyph">设</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- ===== 我的订单 - 5 状态栏 ===== -->
      <view class="card section-card">
        <view class="card-header" @tap="goOrderList">
          <text class="card-title">我的订单</text>
          <view class="card-more">
            <text class="more-text">查看全部</text>
            <text class="more-arrow">›</text>
          </view>
        </view>

        <view class="order-row">
          <view
            v-for="tab in orderTabs"
            :key="tab.key"
            class="order-cell"
            @tap="goOrderListWithStatus(tab.key)"
          >
            <view class="order-icon-wrap">
              <text :class="['order-icon', tab.iconClass]">{{ tab.icon }}</text>
            </view>
            <text class="order-label">{{ tab.label }}</text>
          </view>
        </view>
      </view>

      <!-- ===== 飞手推广卡片（参考图风格：暖橙渐变） ===== -->
      <view class="card promo-card" @tap="goPilotPromotion">
        <view class="promo-inner">
          <view class="promo-icon">
            <text class="promo-icon-text">推</text>
          </view>
          <view class="promo-content">
            <text class="promo-title">飞手推广</text>
            <text class="promo-desc">
              <text class="promo-segment">邀请飞手认证</text>
              <text class="promo-dot">·</text>
              <text class="promo-segment promo-highlight">认证得积分</text>
              <text class="promo-dot">·</text>
              <text class="promo-segment promo-highlight">积分可提现</text>
            </text>
          </view>
          <view class="promo-cta">
            <text>去推广</text>
            <text class="promo-cta-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- ===== 我的服务 - 4 宫格 ===== -->
      <view class="card section-card">
        <view class="card-header">
          <text class="card-title">我的服务</text>
        </view>

        <view class="grid-row">
          <view class="grid-cell" @tap="goMyBids">
            <view class="grid-icon-wrap grid-icon-coupon">
              <text class="grid-icon-text">¥</text>
            </view>
            <text class="grid-label">我的竞标</text>
          </view>
          <view class="grid-cell" @tap="goMyContracts">
            <view class="grid-icon-wrap grid-icon-delivery">
              <text class="grid-icon-text">同</text>
            </view>
            <text class="grid-label">我的合同</text>
          </view>
          <view class="grid-cell" @tap="goMyPublish">
            <view class="grid-icon-wrap grid-icon-bind">
              <text class="grid-icon-text">发</text>
            </view>
            <text class="grid-label">我的发布</text>
          </view>
          <view class="grid-cell" @tap="goAddress">
            <view class="grid-icon-wrap grid-icon-address">
              <text class="grid-icon-text">址</text>
            </view>
            <text class="grid-label">地址管理</text>
          </view>
        </view>
      </view>

      <!-- ===== 认证与工具 - 8 项 2 行 ===== -->
      <view class="card section-card">
        <view class="card-header">
          <text class="card-title">认证与工具</text>
        </view>

        <view class="grid-row">
          <view class="grid-cell" @tap="goAuth">
            <view class="grid-icon-wrap grid-icon-auth">
              <text class="grid-icon-text">证</text>
            </view>
            <text class="grid-label">实名认证</text>
          </view>
          <view class="grid-cell" @tap="goPilotCert">
            <view class="grid-icon-wrap grid-icon-pilot">
              <text class="grid-icon-text">飞</text>
            </view>
            <text class="grid-label">飞手认证</text>
          </view>
          <view class="grid-cell" @tap="goEnterpriseCert">
            <view class="grid-icon-wrap grid-icon-enterprise">
              <text class="grid-icon-text">商</text>
            </view>
            <text class="grid-label">商家认证</text>
          </view>
          <view class="grid-cell" @tap="goMyResume">
            <view class="grid-icon-wrap grid-icon-send">
              <text class="grid-icon-text">简</text>
            </view>
            <text class="grid-label">我的发布</text>
          </view>
        </view>

        <view class="grid-row grid-row-secondary">
          <view class="grid-cell" @tap="goMyPoints">
            <view class="grid-icon-wrap grid-icon-points">
              <text class="grid-icon-text">分</text>
            </view>
            <text class="grid-label">我的积分</text>
          </view>
          <view class="grid-cell" @tap="goDeviceBinding">
            <view class="grid-icon-wrap grid-icon-device">
              <text class="grid-icon-text">绑</text>
            </view>
            <text class="grid-label">设备绑定</text>
          </view>
          <view class="grid-cell" @tap="goOfficialService">
            <view class="grid-icon-wrap grid-icon-service">
              <text class="grid-icon-text">服</text>
            </view>
            <text class="grid-label">官方客服</text>
          </view>
          <view class="grid-cell" @tap="goAbout">
            <view class="grid-icon-wrap grid-icon-about">
              <text class="grid-icon-text">介</text>
            </view>
            <text class="grid-label">公司简介</text>
          </view>
        </view>
      </view>

      <!-- ===== 退出登录 ===== -->
      <view class="card logout-card" v-if="user" @tap="doLogout">
        <text class="logout-text">退出登录</text>
      </view>

      <!-- 底部留白 -->
      <view class="bottom-spacer"></view>
    </view>
  </Layout>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import Layout from '@/components/Layout.vue'
import { getStoredUser, request, authStorage, BASE_URL } from '../../utils/request'

const user = ref(null)
const unreadCount = ref(0)
const userInitial = ref('?')
const userRoleLabel = ref('')

const roleLabels = {
  individual: '个人用户',
  enterprise: '企业用户',
  association_admin: '协会管理',
  platform_admin: '平台管理'
}

// 我的订单 6 状态：参考图 5 项 + 待评价
const orderTabs = [
  { key: 'pending_payment', icon: '付', iconClass: 'icon-pay',    label: '待付款' },
  { key: 'pending_ship',    icon: '发', iconClass: 'icon-send',   label: '待发货' },
  { key: 'pending_receipt', icon: '收', iconClass: 'icon-truck',  label: '待收货' },
  { key: 'pending_review',  icon: '评', iconClass: 'icon-review', label: '待评价' },
  { key: 'completed',       icon: '完', iconClass: 'icon-done',   label: '已完成' },
  { key: 'refund',          icon: '退', iconClass: 'icon-refund', label: '退款/售后' }
]

const fetchData = async () => {
  const currentUser = getStoredUser()
  user.value = currentUser

  if (currentUser) {
    userInitial.value = (currentUser.name || currentUser.phone || '微').charAt(0).toUpperCase()
    userRoleLabel.value = roleLabels[currentUser.role] || ''

    // 刷新服务端用户信息（合并而非覆盖：me 响应缺 name/phone 时保留本地值）
    try {
      const meRes = await request({ url: '/api/auth/me' })
      if (meRes?.user) {
        const merged = { ...currentUser, ...meRes.user }
        user.value = merged
        uni.setStorageSync('user', JSON.stringify(merged))
        userInitial.value = (merged.name || merged.phone || '微').charAt(0).toUpperCase()
        userRoleLabel.value = roleLabels[merged.role] || ''
      }
    } catch (e) { /* fallback to cache */ }

    // 未读消息数
    try {
      const msgRes = await request({ url: '/api/v1/messages/unread-count' })
      unreadCount.value = msgRes?.data?.count || msgRes?.count || 0
    } catch (e) { unreadCount.value = 0 }
  } else {
    userInitial.value = '?'
    userRoleLabel.value = ''
    unreadCount.value = 0
  }
}

onShow(() => {
  fetchData()
})

// ── 导航 ──
const goLogin = () => uni.navigateTo({ url: '/pages/login/index' })

// 头像 URL：相对路径（/uploads/...）拼上后端地址，完整 URL 原样使用
const avatarSrc = (u) => {
  if (!u) return ''
  return u.startsWith('http') ? u : BASE_URL + u
}

// 退出登录：清 token 与用户信息，回到未登录态（此时点登录可进登录页）
const doLogout = () => {
  uni.showModal({
    title: '提示',
    content: '确定退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        authStorage.clearTokens()
        uni.removeStorageSync('user')
        user.value = null
        userInitial.value = '?'
        userRoleLabel.value = ''
        uni.showToast({ title: '已退出登录', icon: 'none' })
      }
    }
  })
}

const handleUserClick = () => {
  if (!user.value) {
    goLogin()
  } else {
    uni.navigateTo({ url: '/pages/mine/profile' })
  }
}

const openApp = () => {
  uni.showToast({ title: '请在微信中打开', icon: 'none' })
}

const goMessages = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/messages/index' })
}

const goSettings = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/profile' })
}

const goOrderList = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/demands/mine' })
}

const goOrderListWithStatus = (status) => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: `/pages/demands/mine?status=${status}` })
}

const goPilotPromotion = () => {
  uni.showToast({ title: '飞手推广功能即将上线', icon: 'none' })
}

const goMyBids = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/demands/bid?mine=1' })
}

const goMyContracts = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/applications/index' })
}

const goMyPublish = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/publish/index' })
}

const goAddress = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/profile' })
}

const goAuth = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/auth' })
}

const goPilotCert = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/training/certificates' })
}

const goEnterpriseCert = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/enterprise/register' })
}

const goMyResume = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/jobs/resume' })
}

const goMyPoints = () => {
  if (!user.value) return goLogin()
  uni.showToast({ title: '我的积分即将上线', icon: 'none' })
}

const goDeviceBinding = () => {
  if (!user.value) return goLogin()
  uni.showToast({ title: '设备绑定即将上线', icon: 'none' })
}

const goOfficialService = () => {
  uni.showModal({
    title: '官方客服',
    content: '客服电话：023-55550500\n工作日 9:00 - 18:00',
    showCancel: false
  })
}

const goAbout = () => {
  uni.showModal({
    title: '公司简介',
    content: '重庆市无人机产业协会\n低空综合服务平台 v1.0.0',
    showCancel: false
  })
}
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background: #f3f4f6;
  padding-bottom: 20rpx;
}

/* ===== 头部区域 ===== */
.header-section {
  position: relative;
  height: 320rpx;
  background: linear-gradient(180deg, #aab2c3 0%, #c2c9d6 45%, #d8dce5 100%);
  overflow: hidden;
}

.header-light {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}

.header-light-1 {
  top: -80rpx;
  right: -60rpx;
  width: 280rpx;
  height: 280rpx;
  background: radial-gradient(circle, rgba(255,255,255,0.32) 0%, transparent 70%);
}

.header-light-2 {
  top: 80rpx;
  right: 220rpx;
  width: 200rpx;
  height: 200rpx;
  background: radial-gradient(circle, rgba(255,255,255,0.18) 0%, transparent 70%);
}

.header-main {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 100rpx 32rpx 40rpx;
  height: 100%;
  box-sizing: border-box;
}

/* 用户块 */
.user-block {
  display: flex;
  align-items: center;
  gap: 24rpx;
  flex: 1;
  min-width: 0;
}

.avatar-wrap {
  position: relative;
  flex-shrink: 0;
  width: 120rpx;
  height: 120rpx;
  border-radius: 60rpx;
  background: var(--color-success);
  border: 4rpx solid #fff;
  box-shadow: 0 6rpx 20rpx rgba(0, 0, 0, 0.12);
  overflow: hidden;
}

.avatar {
  width: 100%;
  height: 100%;
}

.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-text {
  font-size: 56rpx;
  font-weight: 600;
  color: #fff;
}

.user-info-col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.user-name {
  font-size: 38rpx;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 2rpx 6rpx rgba(0, 0, 0, 0.08);
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.cert-pill {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  padding: 4rpx 14rpx;
  background: rgba(52, 199, 89, 0.18);
  border: 1rpx solid rgba(52, 199, 89, 0.4);
  border-radius: 20rpx;
}

.cert-pill-icon {
  font-size: 18rpx;
  color: var(--color-success);
  font-weight: 700;
}

.cert-pill-text {
  font-size: 20rpx;
  color: #fff;
  font-weight: 500;
}

.role-text {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.85);
}

.login-hint {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.9);
}

.login-arrow {
  font-size: 48rpx;
  color: #fff;
  font-weight: 300;
  margin-left: 4rpx;
}

/* 头部图标行 */
.header-icon-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.hdr-icon-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rpx;
  padding: 4rpx;
}

.hdr-icon-circle {
  position: relative;
  width: 72rpx;
  height: 72rpx;
  border-radius: 36rpx;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.hdr-icon-glyph {
  font-size: 28rpx;
  color: #4a5568;
  font-weight: 700;
}

.hdr-dot {
  position: absolute;
  top: 14rpx;
  right: 14rpx;
  width: 16rpx;
  height: 16rpx;
  background: var(--color-danger);
  border-radius: 8rpx;
  border: 2rpx solid #fff;
}

/* ===== 通用卡片 ===== */
.card {
  background: #fff;
  border-radius: 20rpx;
  margin: 0 24rpx 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.section-card {
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 28rpx 8rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: 700;
  color: var(--color-text);
  letter-spacing: 0.5rpx;
}

.card-more {
  display: flex;
  align-items: center;
  gap: 4rpx;
}

.more-text {
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

.more-arrow {
  font-size: 26rpx;
  color: var(--color-text-placeholder);
}

/* ===== 我的订单 5 状态栏 ===== */
.order-row {
  display: flex;
  justify-content: space-around;
  padding: 16rpx 16rpx 36rpx;
}

.order-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10rpx;
  padding: 4rpx 0;
  min-width: 90rpx;
  flex: 1;
}

.order-icon-wrap {
  width: 72rpx;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.order-icon {
  font-size: 48rpx;
  line-height: 1;
}

.icon-pay    { color: var(--color-primary); }
.icon-send   { color: #5a6b85; }
.icon-truck  { color: #6b7a99; }
.icon-review { color: var(--color-warning); }
.icon-done   { color: var(--color-success); }
.icon-refund { color: #5a6b85; }

.order-label {
  font-size: 24rpx;
  color: #4a5568;
}

/* ===== 飞手推广卡片 ===== */
.promo-card {
  background: linear-gradient(135deg, #fff5ec 0%, #fff0e2 50%, #ffeadd 100%);
  border: 1rpx solid rgba(255, 122, 51, 0.12);
}

.promo-inner {
  display: flex;
  align-items: center;
  padding: 32rpx 28rpx;
  gap: 20rpx;
}

.promo-icon {
  flex-shrink: 0;
  width: 88rpx;
  height: 88rpx;
  border-radius: 44rpx;
  background: rgba(255, 122, 51, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}

.promo-icon-text {
  font-size: 44rpx;
}

.promo-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.promo-title {
  font-size: 30rpx;
  font-weight: 700;
  color: var(--color-text);
}

.promo-desc {
  font-size: 22rpx;
  color: #6b7280;
  line-height: 1.5;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4rpx;
}

.promo-segment {
  color: #6b7280;
}

.promo-highlight {
  color: var(--color-warning);
  font-weight: 500;
}

.promo-dot {
  color: #c8c9cc;
  margin: 0 4rpx;
}

.promo-cta {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 2rpx;
  padding: 12rpx 24rpx;
  background: linear-gradient(135deg, var(--color-warning), var(--color-danger));
  border-radius: 40rpx;
  box-shadow: 0 6rpx 16rpx rgba(255, 90, 31, 0.28);
}

.promo-cta text {
  font-size: 24rpx;
  color: #fff;
  font-weight: 600;
}

.promo-cta-arrow {
  font-size: 22rpx;
  margin-left: 2rpx;
}

/* ===== 我的服务 / 认证工具 4 宫格 ===== */
.grid-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  padding: 16rpx 16rpx 16rpx;
}

.grid-row-secondary {
  padding-top: 0;
  padding-bottom: 36rpx;
}

.grid-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14rpx;
  padding: 12rpx 0;
}

.grid-icon-wrap {
  width: 80rpx;
  height: 80rpx;
  border-radius: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.grid-icon-coupon,
.grid-icon-delivery,
.grid-icon-bind,
.grid-icon-address,
.grid-icon-auth,
.grid-icon-pilot,
.grid-icon-enterprise,
.grid-icon-send,
.grid-icon-points,
.grid-icon-device,
.grid-icon-service,
.grid-icon-about { background: linear-gradient(135deg, var(--color-warning), var(--color-danger)); }

.grid-icon-text {
  font-size: 36rpx;
  color: #fff;
  font-weight: 700;
  line-height: 1;
}

.grid-label {
  font-size: 24rpx;
  color: #4a5568;
  text-align: center;
}

.bottom-spacer {
  height: 40rpx;
}

.logout-card {
  margin: 0 24rpx;
  padding: 28rpx 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.logout-text {
  font-size: 28rpx;
  color: var(--color-danger);
}

</style>
