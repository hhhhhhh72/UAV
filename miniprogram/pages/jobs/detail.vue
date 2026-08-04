<template>
  <view class="detail-page">
    <u-nav-bar title="职位详情" show-back @back="goBack" />

    <view v-if="!job" class="loading-state">
      <u-loading size="28rpx" />
      <text class="loading-text">加载中...</text>
    </view>

    <template v-else>
      <!-- 职位信息 -->
      <view class="head-card">
        <text class="job-title">{{ job.title }}</text>
        <view class="job-meta">
          <text v-if="job.salary_fen" class="salary">¥{{ (job.salary_fen / 100).toLocaleString() }}/月</text>
          <u-tag v-if="job.job_type" type="primary" size="mini">{{ job.job_type }}</u-tag>
          <u-tag v-if="job.status === 'published'" type="success" size="mini">招聘中</u-tag>
        </view>
        <view class="job-sub">
          <view v-if="job.location" class="job-loc"><u-icon name="location" size="26rpx" color="#969799" /><text>{{ job.location }}</text></view>
          <text>{{ formatDate(job.created_at) }} 发布</text>
        </view>
      </view>

      <!-- 职位描述 -->
      <view class="section-card">
        <text class="section-title">职位描述</text>
        <text class="job-desc">{{ job.description || '暂无描述' }}</text>
      </view>

      <!-- 底部操作 -->
      <view class="bottom-bar">
        <view class="bottom-btn" :class="{ applied: applied }" @tap="applyJob">{{ applied ? '已投递' : '投递简历' }}</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, onLoad } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../utils/request'

const goBack = () => uni.navigateBack()
const job = ref(null)
const applied = ref(false)

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}

const load = async (id) => {
  try {
    const res = await request({ url: '/api/v1/jobs/' + encodeURIComponent(id) })
    job.value = (res && res.data) || res || null
  } catch (e) {
    uni.showToast({ title: '加载失败', icon: 'none' })
  }
  // 已投递状态：拉取我的投递检查
  if (getStoredUser()) {
    try {
      const apps = await request({ url: '/api/v1/applications' })
      const alist = Array.isArray(apps) ? apps : ((apps && apps.data) || [])
      applied.value = alist.some((a) => a.job_id === id)
    } catch (e) {}
  }
}

// 投递：登录 → 简历检查 → 提交
const applyJob = async () => {
  if (!job.value || applied.value) return
  const user = getStoredUser()
  if (!user) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }
  try {
    const resumes = await request({ url: '/api/v1/resumes/mine' })
    const rlist = Array.isArray(resumes) ? resumes : ((resumes && resumes.data) || [])
    if (!rlist.length) {
      uni.showModal({
        title: '需要简历',
        content: '投递职位需要一份简历，是否现在去创建？',
        success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pages/jobs/resume' }) },
      })
      return
    }
    await request({
      url: '/api/v1/applications',
      method: 'POST',
      data: { job_id: job.value.id, resume_id: rlist[0].id },
    })
    applied.value = true
    uni.showModal({ title: '投递成功', content: '简历已投递，可在「我的投递」查看进展', showCancel: false })
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '投递失败', icon: 'none' })
  }
}

onLoad((opts) => { if (opts.id) load(opts.id) })
</script>

<style scoped>
.detail-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 120rpx; }
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 60px 0; }
.loading-text { font-size: 24rpx; color: var(--color-text-secondary); }
.head-card { background: var(--color-bg-card); margin: 20rpx; border-radius: 12rpx; padding: 32rpx; }
.job-title { font-size: 36rpx; font-weight: 700; color: var(--color-text); display: block; }
.job-meta { display: flex; align-items: center; gap: 16rpx; margin-top: 16rpx; flex-wrap: wrap; }
.salary { font-size: 32rpx; font-weight: 700; color: var(--color-accent-deep); }
.job-sub { display: flex; align-items: center; gap: 24rpx; font-size: 24rpx; color: var(--color-text-secondary); margin-top: 16rpx; }
.job-loc { display: flex; align-items: center; gap: 6rpx; }
.section-card { background: var(--color-bg-card); margin: 0 20rpx 20rpx; border-radius: 12rpx; padding: 32rpx; }
.section-title { font-size: 28rpx; font-weight: 700; color: var(--color-text); display: block; margin-bottom: 16rpx; }
.job-desc { font-size: 26rpx; color: var(--color-text); line-height: 1.7; white-space: pre-wrap; }
.bottom-bar { position: fixed; left: 0; right: 0; bottom: 0; padding: 16rpx 32rpx calc(16rpx + env(safe-area-inset-bottom)); background: var(--color-bg-card); box-shadow: 0 -2rpx 12rpx rgba(0,0,0,.04); }
.bottom-btn { height: 88rpx; border-radius: 44rpx; background: var(--color-primary); color: #fff; font-size: 30rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; }
.bottom-btn.applied { background: var(--color-divider); color: var(--color-text-secondary); }
</style>
