<template>
  <view class="page-container">
    <u-nav-bar title="我的招聘" show-back @back="goBack" />

    <!-- 非企业账号：引导入驻 -->
    <view v-if="!isEnterprise" class="gate">
      <view class="gate-ico">聘</view>
      <text class="gate-title">仅企业账号可管理招聘</text>
      <text class="gate-note">请先完成企业入驻，审核通过后即可发布与管理招聘职位。</text>
      <u-button type="primary" size="medium" round @tap="goRegister">去企业入驻</u-button>
    </view>

    <template v-else>
      <!-- 顶部操作 -->
      <view class="toolbar">
        <text class="toolbar-hint">职位发布后为草稿，需点击「发布上线」才会公开</text>
        <u-button type="primary" size="small" round @tap="goPublish">发布新职位</u-button>
      </view>

      <!-- Loading -->
      <view v-if="loading" class="loading-state">
        <u-loading size="28rpx" />
        <text class="loading-text">加载中...</text>
      </view>

      <!-- Empty -->
      <view v-else-if="!list.length" class="empty-state">
        <u-empty description="暂无招聘职位" />
      </view>

      <!-- 职位列表 -->
      <view v-else class="job-list">
        <view v-for="item in list" :key="item.id" class="job-card">
          <view class="job-row1">
            <text class="job-title">{{ item.title }}</text>
            <u-tag :type="statusTag(item.status)" size="mini">{{ statusLabel[item.status] || item.status }}</u-tag>
          </view>
          <view class="job-meta">
            <text v-if="item.location">{{ item.location }}</text>
            <text v-if="item.salary_fen">¥{{ (item.salary_fen / 100).toLocaleString() }}/月</text>
          </view>
          <view class="job-actions">
            <u-button v-if="item.status === 'published'" type="primary" size="mini" round @tap="goApplicants(item)">查看投递</u-button>
            <u-button v-if="item.status === 'draft'" type="primary" size="mini" round @tap="publishJob(item)">发布上线</u-button>
            <u-button v-if="item.status === 'published'" type="warning" size="mini" round @tap="closeJob(item)">关闭</u-button>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../../utils/request'

const goBack = () => uni.navigateBack()
const goRegister = () => uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })
const goPublish = () => uni.navigateTo({ url: '/pkg-talent/pages/publish/job' })
const goApplicants = (item) => uni.navigateTo({ url: '/pkg-talent/pages/jobs/applicants?job_id=' + encodeURIComponent(item.id) })

const user = getStoredUser()
const isEnterprise = !!(user && (user.role === 'enterprise' || user.role === 'platform_admin'))

const list = ref([])
const loading = ref(false)

const statusLabel = { draft: '草稿', published: '招聘中', closed: '已关闭' }
const statusTag = (s) => ({ draft: 'info', published: 'success', closed: 'default' }[s] || 'info')

const load = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/jobs/mine' })
    list.value = Array.isArray(res) ? res : ((res && res.data) || [])
  } catch (e) {
    uni.showToast({ title: '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const publishJob = async (item) => {
  try {
    await request({ url: '/api/v1/jobs/' + encodeURIComponent(item.id) + '/publish', method: 'POST' })
    uni.showToast({ title: '已发布上线', icon: 'success' })
    load()
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '操作失败', icon: 'none' })
  }
}

const closeJob = async (item) => {
  try {
    await request({ url: '/api/v1/jobs/' + encodeURIComponent(item.id) + '/close', method: 'POST' })
    uni.showToast({ title: '已关闭', icon: 'success' })
    load()
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '操作失败', icon: 'none' })
  }
}

// onShow 而非 onMounted：发布/关闭等操作返回后立即看到最新状态
onShow(load)
</script>

<style scoped>
.page-container { min-height: 100vh; background: var(--color-bg); padding-bottom: 20px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 16rpx 32rpx; }
.toolbar-hint { font-size: 22rpx; color: var(--color-text-placeholder); flex: 1; }
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 60px 0; }
.loading-text { font-size: 24rpx; color: var(--color-text-secondary); }
.empty-state { padding-top: 60px; }
.job-list { padding: 8px 12px; }
.job-card { background: var(--color-bg-card); border-radius: 12rpx; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 1px 3px rgba(0,0,0,.05); }
.job-row1 { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; }
.job-title { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.job-meta { display: flex; gap: 24rpx; font-size: 24rpx; color: var(--color-text-secondary); margin-top: 12rpx; }
.job-actions { display: flex; justify-content: flex-end; margin-top: 16rpx; }
.gate { padding: 120rpx 48rpx; display: flex; flex-direction: column; align-items: center; gap: 24rpx; }
.gate-ico { width: 120rpx; height: 120rpx; border-radius: 50%; background: var(--color-primary-light); color: var(--color-primary); font-size: 48rpx; font-weight: 700; display: flex; align-items: center; justify-content: center; }
.gate-title { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.gate-note { font-size: 24rpx; color: var(--color-text-secondary); text-align: center; line-height: 1.6; }
</style>
