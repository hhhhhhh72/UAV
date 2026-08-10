<template>
  <Layout :current="4">
    <view class="mine-page">
      <!-- ═══════ 身份卡 ═══════ -->
      <view class="identity-card">
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
              <text v-if="userRoleLabel" class="role-text">{{ userRoleLabel }}</text>
              <text v-if="user.role === 'individual'" class="cert-pill">飞手</text>
            </view>
            <text v-else class="login-hint">登录后查看需求与对接意向</text>
          </view>
          <text v-if="!user" class="login-arrow">›</text>
        </view>

        <!-- 右侧：消息（红点）+ 设置 -->
        <view class="header-icon-row">
          <view class="hdr-icon-btn" @tap="goMessages">
            <image class="hdr-icon" :src="'/static/home/icons/message.svg'" mode="aspectFit" />
            <view v-if="unreadCount > 0" class="hdr-dot"></view>
          </view>
          <view class="hdr-icon-btn" @tap="goSettings">
            <text class="hdr-icon-text">设</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 我的需求 - 状态栏 ═══════ -->
      <view class="card section-card">
        <view class="card-header" @tap="goOrderList">
          <text class="card-title">我的需求</text>
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
            <text :class="['order-icon', tab.iconClass]">{{ tab.icon }}</text>
            <text class="order-label">{{ tab.label }}</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 业务 ═══════ -->
      <view class="card section-card">
        <view class="card-header">
          <text class="card-title">业务</text>
        </view>
        <view class="menu-list">
          <view class="menu-item" hover-class="tap-fade" @tap="goMyDemands">
            <text class="menu-label">我的需求</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goMyOrders">
            <text class="menu-label">我的订单</text>
            <text class="menu-desc">接单作业与验收</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goMyPublish">
            <text class="menu-label">我的发布</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goMyIntents">
            <text class="menu-label">对接意向</text>
            <text class="menu-desc">我登记过的对接记录</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goMyContracts">
            <text class="menu-label">我的申请</text>
            <text class="menu-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 认证与资料 ═══════ -->
      <view class="card section-card">
        <view class="card-header">
          <text class="card-title">认证与资料</text>
        </view>
        <view class="menu-list">
          <view class="menu-item" hover-class="tap-fade" @tap="goAuth">
            <text class="menu-label">实名认证</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goPilotCert">
            <text class="menu-label">飞手认证</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goEnterpriseCert">
            <text class="menu-label">商家认证</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goMyResume">
            <text class="menu-label">我的简历</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goMyCertificates">
            <text class="menu-label">我的证书</text>
            <text class="menu-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 服务与支持 ═══════ -->
      <view class="card section-card">
        <view class="card-header">
          <text class="card-title">服务与支持</text>
        </view>
        <view class="menu-list">
          <view class="menu-item" hover-class="tap-fade" @tap="goMyPoints">
            <text class="menu-label">我的积分</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goOfficialService">
            <text class="menu-label">官方客服</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" hover-class="tap-fade" @tap="goAbout">
            <text class="menu-label">公司简介</text>
            <text class="menu-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 退出登录 ═══════ -->
      <view class="logout-card" v-if="user" @tap="doLogout">
        <text class="logout-text">退出登录</text>
      </view>

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

// 我的需求状态栏：对应 DemandStatus，demands/mine 按 ?status= 过滤
const orderTabs = [
  { key: 'pending',   icon: '审', iconClass: 'icon-pay',    label: '待审核' },
  { key: 'published', icon: '公', iconClass: 'icon-send',   label: '已公开' },
  { key: 'completed', icon: '完', iconClass: 'icon-done',   label: '已完成' },
  { key: 'cancelled', icon: '取', iconClass: 'icon-refund', label: '已取消' },
  { key: 'rejected',  icon: '驳', iconClass: 'icon-review', label: '被驳回' }
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
  uni.navigateTo({ url: '/pages/demands/mine?status=' + encodeURIComponent(status) })
}

const goMyDemands = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/demands/mine' })
}

const goMyOrders = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/orders/mine' })
}

const goMyContracts = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/applications/index' })
}

const goMyPublish = () => {
  if (!user.value) return goLogin()
  // 发布页是 tabBar 页，只能 switchTab
  uni.switchTab({ url: '/pages/publish/index' })
}

const goAddress = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/profile' })
}

const goMyIntents = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/intents/mine' })
}

const goAuth = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/auth' })
}

const goPilotCert = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/auth' })
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
  uni.showToast({ title: '积分功能即将上线', icon: 'none' })
}

const goMyCertificates = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/training/certificates' })
}

const goOfficialService = () => {
  if (!user.value) return goLogin()
  uni.showToast({ title: '请联系协会秘书处', icon: 'none' })
}

const goAbout = () => {
  uni.showToast({ title: '重庆无人机产业协会', icon: 'none' })
}
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding: 12px 12px calc(24px + env(safe-area-inset-bottom));
}

/* 身份卡 */
.identity-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 8px;
}

.user-block {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.avatar-wrap {
  flex-shrink: 0;
}

.avatar {
  width: 52px;
  height: 52px;
  border-radius: 8px;
  display: block;
}

.avatar-placeholder {
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-text {
  font-size: 20px;
  font-weight: 700;
  color: #0A66C2;
}

.user-info-col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-name {
  font-size: 17px;
  font-weight: 700;
  color: #17212B;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-text {
  font-size: 12px;
  color: #667085;
}

.cert-pill {
  font-size: 11px;
  color: #168A55;
  background: #E9F7F0;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.login-hint {
  font-size: 12px;
  color: #98A2B3;
}

.login-arrow {
  font-size: 22px;
  color: #98A2B3;
}

/* 右侧图标 */
.header-icon-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.hdr-icon-btn {
  position: relative;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hdr-icon {
  width: 22px;
  height: 22px;
}

.hdr-icon-text {
  font-size: 13px;
  font-weight: 600;
  color: #344054;
}

.hdr-dot {
  position: absolute;
  top: 6px;
  right: 8px;
  width: 10px;
  height: 10px;
  border-radius: 5px;
  background: #F97316;
  border: 2rpx solid #fff;
}

/* 卡片 */
.card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  margin-bottom: 8px;
}

.section-card {
  padding: 14px 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
}

.card-more {
  display: flex;
  align-items: center;
  gap: 4px;
}

.more-text {
  font-size: 12px;
  color: #667085;
}

.more-arrow {
  font-size: 14px;
  color: #98A2B3;
}

/* 需求状态栏 */
.order-row {
  display: flex;
}

.order-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.order-icon {
  font-size: 16px;
  font-weight: 700;
  color: #0A66C2;
  background: #EAF3FB;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.order-icon.icon-send { color: #168A55; background: #E9F7F0; }
.order-icon.icon-done { color: #168A55; background: #E9F7F0; }
.order-icon.icon-refund { color: #98A2B3; background: #F4F6F8; }
.order-icon.icon-review { color: #D92D20; background: #FEF3F2; }

.order-label {
  font-size: 11px;
  color: #344054;
}

/* 菜单列表 */
.menu-list {
  display: flex;
  flex-direction: column;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 0;
  border-top: 1px solid #EEF1F4;
}

.menu-item:first-child {
  border-top: none;
}

.menu-label {
  font-size: 14px;
  color: #17212B;
}

.menu-desc {
  flex: 1;
  font-size: 11px;
  color: #98A2B3;
  text-align: right;
}

.menu-arrow {
  margin-left: auto;
  font-size: 16px;
  color: #98A2B3;
}

/* 退出登录 */
.logout-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px 0;
  text-align: center;
  margin-top: 16px;
}

.logout-text {
  font-size: 14px;
  color: #D92D20;
  font-weight: 500;
}

.bottom-spacer {
  height: 8px;
}

.tap-fade {
  opacity: 0.7;
}
</style>
