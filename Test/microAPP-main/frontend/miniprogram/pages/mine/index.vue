<template>
  <Layout :current="3">
    <view class="mine-page">
      <view class="nav-bar">个人中心</view>

      <!-- 用户信息卡片 (同步 H5 渐变色) -->
      <view class="user-card">
        <view class="user-header" @tap="handleUserClick">
          <image v-if="user && user.avatar" class="avatar" :src="user.avatar" mode="aspectFill" />
          <view v-else class="default-avatar">
            <text class="avatar-icon">👤</text>
          </view>
          <view class="user-info">
            <view class="user-name">{{ user?.name || '点击登录' }}</view>
            <view class="user-phone">{{ user?.phone || '登录后享受更多服务' }}</view>
          </view>
          <text v-if="!user" class="arrow">›</text>
        </view>

        <!-- 统计数据 -->
        <view class="stats-grid">
          <view class="stat-item">
            <view class="stat-value">{{ totalCount }}</view>
            <view class="stat-label">总申请</view>
          </view>
          <view class="stat-item">
            <view class="stat-value">{{ processingCount }}</view>
            <view class="stat-label">处理中</view>
          </view>
          <view class="stat-item">
            <view class="stat-value">{{ completedCount }}</view>
            <view class="stat-label">已完成</view>
          </view>
        </view>
      </view>

      <!-- 第一组功能菜单 -->
      <view class="menu-section">
        <view class="menu-list">
          <view class="menu-item" v-if="user && (user.role === 'admin' || user.role === 'dsl_admin')" @tap="goAdmin">
            <text class="menu-icon">⚙️</text>
            <text class="menu-title">后台管理</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" @tap="goApplications">
            <text class="menu-icon">📋</text>
            <text class="menu-title">我的申请</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" @tap="goCases">
            <text class="menu-icon">🎬</text>
            <text class="menu-title">案例展示</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" @tap="goProfile">
            <text class="menu-icon">👤</text>
            <text class="menu-title">个人信息</text>
            <text class="menu-arrow">›</text>
          </view>
          <!-- #ifdef MP-WEIXIN -->
          <view class="menu-item" v-if="user && user.wxOpenid && !user.phone">
            <text class="menu-icon">📱</text>
            <view class="menu-title-group">
              <text class="menu-title">绑定手机号</text>
              <button class="inline-phone-btn" open-type="getPhoneNumber" @getphonenumber="handleBindPhone">
                去绑定
              </button>
            </view>
          </view>
          <!-- #endif -->
          <view class="menu-item" @tap="goAuth">
            <text class="menu-icon">🛡️</text>
            <view class="menu-title-group">
              <text class="menu-title">实名认证</text>
              <text class="menu-label">{{ user?.isAuth ? '已认证' : '未认证' }}</text>
            </view>
            <text class="menu-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- 第二组功能菜单 (同步 H5) -->
      <view class="menu-section">
        <view class="menu-list">
          <view class="menu-item" @tap="showGuide">
            <text class="menu-icon">📖</text>
            <text class="menu-title">服务指南</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" @tap="showFAQ">
            <text class="menu-icon">❓</text>
            <text class="menu-title">常见问题</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" @tap="showContact">
            <text class="menu-icon">🎧</text>
            <text class="menu-title">联系客服</text>
            <text class="menu-arrow">›</text>
          </view>
          <view class="menu-item" @tap="showAbout">
            <text class="menu-icon">ℹ️</text>
            <text class="menu-title">关于我们</text>
            <text class="menu-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- 退出登录 -->
      <view class="menu-section logout-section" v-if="user">
        <view class="menu-list">
          <view class="menu-item logout-item" @tap="handleLogout">
            <text class="menu-icon">🚪</text>
            <text class="menu-title">退出登录</text>
            <text class="menu-arrow">›</text>
          </view>
        </view>
      </view>
      
      <view class="safe-area-bottom"></view>
    </view>
  </Layout>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import Layout from '@/components/Layout.vue'
import { getStoredUser, request, authStorage } from '../../utils/request'

const user = ref(null)
const totalCount = ref(0)
const processingCount = ref(0)
const completedCount = ref(0)

const fetchData = async () => {
  const currentUser = getStoredUser()
  user.value = currentUser
  if (!currentUser) {
    totalCount.value = 0
    processingCount.value = 0
    completedCount.value = 0
    return
  }

  // Refresh user info from server
  try {
    const meRes = await request({ url: '/api/auth/me' })
    if (meRes?.user) {
      user.value = meRes.user
      uni.setStorageSync('user', JSON.stringify(meRes.user))
    }
  } catch (e) { /* use cached user */ }

  try {
    const res = await request({ url: '/api/list', data: { userId: currentUser.id } })
    const list = Array.isArray(res) ? res : (res?.data || [])
    totalCount.value = list.length
    processingCount.value = list.filter((i) => i.status === '处理中').length
    completedCount.value = list.filter((i) => i.status === '已完成').length
  } catch (e) {
    const mock = uni.getStorageSync('mock_applications') || []
    const list = mock.filter((a) => a.userId === currentUser.id)
    totalCount.value = list.length
    processingCount.value = list.filter((i) => i.status === '处理中').length
    completedCount.value = list.filter((i) => i.status === '已完成').length
  }
}

onShow(() => {
  fetchData()
})

const handleUserClick = () => {
  if (!user.value) {
    uni.navigateTo({ url: '/pages/login/index' })
  } else {
    uni.navigateTo({ url: '/pages/mine/profile' })
  }
}
const goAdmin = () => uni.navigateTo({ url: '/pages/admin/index' })
const goApplications = () => uni.switchTab({ url: '/pages/applications/index' })
const goCases = () => uni.navigateTo({ url: '/pages/cases/index' })
const goProfile = () => {
  if (!user.value) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  uni.navigateTo({ url: '/pages/mine/profile' })
}
const goAuth = () => {
  if (!user.value) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  uni.navigateTo({ url: '/pages/mine/auth' })
}
const showToast = (msg) => uni.showToast({ title: msg, icon: 'none' })

const showGuide = () => {
  uni.showModal({
    title: '服务指南',
    content: '1. 选择所需服务\n2. 填写申请表单\n3. 提交申请\n4. 等待客服联系\n5. 确认服务详情\n6. 服务执行',
    showCancel: false
  })
}

const showFAQ = () => {
  uni.showModal({
    title: '常见问题',
    content: 'Q: 申请后多久联系？\nA: 2小时内会有客服联系您\n\nQ: 如何修改申请？\nA: 请联系客服进行修改',
    showCancel: false
  })
}

const showContact = () => {
  uni.showModal({
    title: '联系客服',
    content: '客服电话：400-123-4567\n工作时间：工作日 9:00-18:00',
    showCancel: false
  })
}

const showAbout = () => {
  uni.showModal({
    title: '关于我们',
    content: '低空综合服务平台\n开发主体：温州低空经济发展有限公司\n版本：v1.1.0\n\n专注于提供专业、高效、安全的低空服务',
    showCancel: false
  })
}

const handleBindPhone = async (e) => {
  if (e.detail.errMsg !== 'getPhoneNumber:ok') return
  uni.showLoading({ title: '绑定中...' })
  try {
    const res = await request({
      url: '/api/auth/wx-phone',
      method: 'POST',
      data: { code: e.detail.code }
    })
    if (res?.success) {
      user.value = res.user
      uni.setStorageSync('user', JSON.stringify(res.user))
      uni.showToast({ title: '绑定成功' })
    } else {
      throw new Error(res?.message || '绑定失败')
    }
  } catch (err) {
    uni.showToast({ title: err?.message || '绑定失败', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

const handleLogout = () => {
  uni.showModal({
    title: '提示',
    content: '确定退出？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await request({ url: '/api/auth/logout', method: 'POST' })
        } catch (e) { /* ignore */ }
        authStorage.clearTokens()
        uni.removeStorageSync('user')
        user.value = null
        uni.switchTab({ url: '/pages/home/index' })
      }
    }
  })
}
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background: #f7f8fa;
}

.nav-bar {
  height: 44px;
  line-height: 44px;
  text-align: center;
  font-size: 17px;
  font-weight: 600;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;
}

.user-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  margin: 12px 16px 20px;
  padding: 24px 20px;
  border-radius: 16px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.avatar, .default-avatar {
  width: 60px;
  height: 60px;
  border-radius: 30px;
  background: #f5f6fa;
  border: 2px solid rgba(255, 255, 255, 0.3);
}

.default-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-icon {
  font-size: 30px;
  color: #bdc3c7;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 4px;
}

.user-phone {
  font-size: 14px;
  opacity: 0.8;
}

.arrow {
  font-size: 20px;
  opacity: 0.6;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 22px;
  font-weight: bold;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  opacity: 0.8;
}

.menu-section {
  background: #fff;
  margin: 0 16px 12px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.menu-list {
  display: flex;
  flex-direction: column;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f2f3f5;
  transition: background-color 0.2s;
}

.menu-item:active {
  background-color: #f2f3f5;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-icon {
  font-size: 18px;
  margin-right: 12px;
}

.menu-title {
  flex: 1;
  font-size: 15px;
  color: #323233;
}

.menu-title-group {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.menu-label {
  font-size: 12px;
  color: #969799;
  margin-right: 4px;
}

.menu-arrow {
  font-size: 16px;
  color: #969799;
}

.logout-item .menu-title {
  color: #ee0a24;
}

.inline-phone-btn {
  display: inline-block;
  font-size: 13px;
  color: #667eea;
  background: transparent;
  padding: 4px 12px;
  border: 1px solid #667eea;
  border-radius: 14px;
  line-height: 1.4;
  margin: 0;
}

.inline-phone-btn::after {
  border: none;
}

.safe-area-bottom {
  height: calc(constant(safe-area-inset-bottom) + 20px);
  height: calc(env(safe-area-inset-bottom) + 20px);
}
</style>
