<template>
  <Layout :current="3">
    <view class="mine-page">
      <van-nav-bar title="我的" fixed placeholder />

      <!-- 用户信息卡片 -->
      <view class="user-card">
        <view class="user-header" @tap="handleUserClick">
          <van-image
            v-if="user && user.avatar"
            round
            width="60"
            height="60"
            :src="user.avatar"
          />
          <view v-else class="default-avatar">
            <van-icon name="contact" size="36" color="#bdc3c7" />
          </view>
          <view class="user-info">
            <text class="user-name">{{ user ? (user.name || user.nickname || '用户') : '点击登录' }}</text>
            <view class="user-sub-row">
              <text class="user-phone">{{ user ? (user.phone || '') : '登录后享受更多服务' }}</text>
              <van-tag v-if="user && roleLabel" type="primary" size="small">{{ roleLabel }}</van-tag>
            </view>
          </view>
          <van-icon v-if="!user" name="arrow" size="16" color="#8e8e93" />
        </view>

        <!-- 统计数据 -->
        <view class="stats-grid">
          <view class="stat-item">
            <text class="stat-value">{{ totalCount }}</text>
            <text class="stat-label">总申请</text>
          </view>
          <view class="stat-item">
            <text class="stat-value">{{ processingCount }}</text>
            <text class="stat-label">处理中</text>
          </view>
          <view class="stat-item">
            <text class="stat-value">{{ completedCount }}</text>
            <text class="stat-label">已完成</text>
          </view>
        </view>
      </view>

      <!-- 第一组功能菜单 -->
      <view class="menu-section">
        <van-cell-group inset>
          <van-cell
            v-if="user && isAdmin"
            title="管理后台"
            icon="setting-o"
            is-link
            @tap="goAdmin"
          />
          <van-cell
            title="企业入驻"
            icon="shop-o"
            is-link
            @tap="goProfile"
          />
          <van-cell
            title="我的需求"
            icon="records-o"
            is-link
            @tap="goApplications"
          />
          <van-cell
            title="我的证书"
            icon="award-o"
            is-link
            @tap="showComingSoon"
          />
          <van-cell
            title="我的合同"
            icon="description"
            is-link
            @tap="showComingSoon"
          />
          <van-cell
            title="钱包余额"
            icon="gold-coin-o"
            is-link
            @tap="showComingSoon"
          />
        </van-cell-group>
      </view>

      <!-- 第二组功能菜单 -->
      <view class="menu-section">
        <van-cell-group inset>
          <van-cell
            title="服务指南"
            icon="guide-o"
            is-link
            @tap="showGuide"
          />
          <van-cell
            title="常见问题"
            icon="question-o"
            is-link
            @tap="showFAQ"
          />
          <van-cell
            title="联系客服"
            icon="service-o"
            is-link
            @tap="showContact"
          />
          <van-cell
            title="关于我们"
            icon="info-o"
            is-link
            @tap="showAbout"
          />
        </van-cell-group>
      </view>

      <!-- 退出登录 -->
      <view class="menu-section" v-if="user">
        <van-cell-group inset>
          <van-cell
            title="退出登录"
            icon="close"
            is-link
            @tap="handleLogout"
          />
        </van-cell-group>
      </view>

      <view class="safe-area-bottom"></view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { getStoredUser, request, authStorage } from '../../utils/request'

const user = ref(null)
const totalCount = ref(0)
const processingCount = ref(0)
const completedCount = ref(0)

const roleLabels = {
  platform_admin: '平台管理员',
  association_admin: '协会管理员',
  enterprise: '企业用户',
  individual: '个人用户',
}

const roleLabel = computed(() => {
  if (!user.value) return ''
  return roleLabels[user.value.role] || ''
})

const isAdmin = computed(() => {
  if (!user.value) return false
  return user.value.role === 'platform_admin' || user.value.role === 'association_admin'
})

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
    if (meRes?.user || meRes?.data?.user) {
      const u = meRes.user || meRes.data.user
      user.value = u
      uni.setStorageSync('user', JSON.stringify(u))
    }
  } catch (e) { /* use cached user */ }

  try {
    const res = await request({ url: '/api/list', data: { userId: currentUser.id } })
    const list = Array.isArray(res) ? res : (res?.data || [])
    totalCount.value = list.length
    processingCount.value = list.filter(i => i.status === '处理中').length
    completedCount.value = list.filter(i => i.status === '已完成').length
  } catch (e) {
    totalCount.value = 0
    processingCount.value = 0
    completedCount.value = 0
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
const goApplications = () => uni.navigateTo({ url: '/pages/applications/index' })
const goProfile = () => {
  if (!user.value) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  uni.navigateTo({ url: '/pages/mine/profile' })
}

const showComingSoon = () => {
  uni.showToast({ title: '即将上线', icon: 'none', duration: 1500 })
}

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
    content: '客服电话：0577-55558188\n工作时间：工作日 9:00-18:00',
    confirmText: '拨打电话',
    success: (res) => {
      if (res.confirm) {
        uni.makePhoneCall({ phoneNumber: '0577-55558188' })
      }
    }
  })
}

const showAbout = () => {
  uni.showModal({
    title: '关于我们',
    content: '无人机产业综合服务平台\n版本：v1.1.0\n\n专注于提供专业、高效、安全的低空服务',
    showCancel: false
  })
}

const handleLogout = () => {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await request({ url: '/api/auth/logout', method: 'POST' })
        } catch (e) { /* ignore */ }
        authStorage.clearTokens()
        uni.removeStorageSync('user')
        user.value = null
        totalCount.value = 0
        processingCount.value = 0
        completedCount.value = 0
        uni.switchTab({ url: '/pages/home/index' })
      }
    }
  })
}
</script>

<style scoped>
.mine-page {
  background: #f7f8fa;
  min-height: 100vh;
}

.user-card {
  background: #ffffff;
  margin: 12px;
  padding: 16px;
  border-radius: 18px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.default-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: #f2f2f7;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 20px;
  font-weight: 600;
  color: #1d1d1f;
  display: block;
  margin-bottom: 6px;
}

.user-sub-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-phone {
  font-size: 14px;
  color: #86868b;
}

.stats-grid {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.stat-item {
  flex: 1;
  text-align: center;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: #1d1d1f;
  display: block;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 13px;
  color: #86868b;
  display: block;
}

.menu-section {
  margin: 12px 0;
}

.safe-area-bottom {
  height: 20px;
}
</style>
