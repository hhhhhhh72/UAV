<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="职位详情" show-back :fixed="true" @back="goBack" />

    <!-- 骨架屏：加载中 -->
    <view v-if="loading && !job" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w60"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w80"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- 错误态 -->
    <view v-else-if="errorMsg && !job" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="load(id)">重新加载</view>
      </u-empty>
    </view>

    <template v-else-if="job">
      <!-- 职位信息 -->
      <view class="card head-card">
        <text class="job-title">{{ job.title }}</text>
        <view class="job-meta">
          <text v-if="job.salary_fen" class="salary">¥{{ (job.salary_fen / 100).toLocaleString() }}/月</text>
          <text v-if="job.job_type" class="tag tag-blue">{{ job.job_type }}</text>
          <text v-if="job.status === 'published'" class="tag tag-green">招聘中</text>
        </view>
        <view class="job-sub">
          <view v-if="job.location" class="job-loc">
            <u-icon name="location" size="24rpx" color="#667085" />
            <text>{{ job.location }}</text>
          </view>
          <text class="job-date">{{ formatDate(job.created_at) }} 发布</text>
        </view>
      </view>

      <!-- 职位描述 -->
      <view class="card section-card">
        <text class="section-title">职位描述</text>
        <rich-text v-if="(job.description || '').indexOf('<') >= 0" class="job-desc" :nodes="job.description"></rich-text>
        <text v-else class="job-desc">{{ job.description || '暂无描述' }}</text>
      </view>

      <!-- 底部操作 -->
      <view class="bottom-bar">
        <view v-if="job.status === 'published'" class="bottom-btn" :class="{ applied: applied }" hover-class="bottom-press" :hover-stay-time="100" @tap="applyJob">{{ applied ? '已投递' : '投递简历' }}</view>
        <view v-else class="bottom-btn bottom-btn--disabled">{{ job.status === 'closed' ? '已关闭' : '暂未开放' }}</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useReduceMotion } from '../../../utils/motion'
import { request, getStoredUser } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'

const goBack = () => uni.navigateBack()
const { noMotion, checkMotion } = useReduceMotion()
const statusBarHeight = ref(20)
const id = ref('')
const job = ref(null)
const applied = ref(false)
const submitting = ref(false)
const loading = ref(true)
const errorMsg = ref('')

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate())
}

const load = async (pid) => {
  if (!pid) return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/jobs/' + encodeURIComponent(pid) })
    job.value = (res && res.data) || res || null
    if (!job.value) errorMsg.value = '职位不存在或已下架'
  } catch (e) {
    errorMsg.value = '加载失败，请检查网络'
  } finally {
    loading.value = false
  }
  // 已投递状态：拉取我的投递检查
  if (getStoredUser()) {
    try {
      const apps = await request({ url: '/api/v1/applications' })
      const alist = Array.isArray(apps) ? apps : ((apps && apps.data) || [])
      applied.value = alist.some((a) => a.job_id === pid)
    } catch (e) {}
  }
}

// 投递：登录 → 简历检查 → 提交（submitting 防重复点击）
const applyJob = async () => {
  if (submitting.value || !job.value || applied.value || job.value.status !== 'published') return
  if (!requireLogin()) return
  submitting.value = true
  try {
    const resumes = await request({ url: '/api/v1/resumes/mine' })
    const rlist = Array.isArray(resumes) ? resumes : ((resumes && resumes.data) || [])
    if (!rlist.length) {
      uni.showModal({
        title: '需要简历',
        content: '投递职位需要一份简历，是否现在去创建？',
        success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pkg-talent/pages/jobs/resume' }) },
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
  } finally {
    submitting.value = false
  }
}

onLoad((opts) => {
  checkMotion()
  try {
    const sys = uni.getSystemInfoSync()
    if (sys && sys.statusBarHeight) statusBarHeight.value = sys.statusBarHeight
  } catch (e) {}
  if (opts && opts.id) {
    id.value = opts.id
    load(opts.id)
  }
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(180rpx + env(safe-area-inset-bottom));
}

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 20rpx; padding: 24rpx 32rpx 0; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 26rpx;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
}
.sk-row { display: flex; align-items: center; gap: 16rpx; }
.sk-tag { width: 112rpx; height: 36rpx; border-radius: 8rpx; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 16rpx; }
.sk-l { height: 24rpx; background: #EDF0F3; border-radius: 8rpx; animation: skPulse 1.4s linear infinite; }
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.stb { height: 72rpx; padding: 0 40rpx; border-radius: 12rpx; background: #0A66C2; color: #fff; font-size: 24rpx; line-height: 72rpx; }

/* ===== 卡片：对齐「我的发布/职位列表」卡片规范（16rpx 圆角 + 细描边 + 轻投影） ===== */
.card {
  position: relative;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.head-card { margin: 24rpx 32rpx 0; padding: 26rpx; animation: cardIn .22s ease-out backwards; animation-delay: 60ms; }
.section-card { margin: 20rpx 32rpx 0; padding: 26rpx; animation: cardIn .22s ease-out backwards; animation-delay: 120ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.job-title { font-size: 36rpx; font-weight: 700; color: #17212B; display: block; line-height: 1.4; }
.job-meta { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; flex-wrap: wrap; }
.salary { font-size: 32rpx; font-weight: 700; color: #E96012; }
.tag { border-radius: 8rpx; padding: 6rpx 12rpx; font-size: 20rpx; line-height: 1; font-weight: 600; }
.tag-blue { color: #0A66C2; background: #EAF3FB; }
.tag-green { color: #168A55; background: #E9F7F0; }
.job-sub { display: flex; align-items: center; flex-wrap: wrap; gap: 8rpx 24rpx; margin-top: 16rpx; padding-top: 18rpx; border-top: 1px solid #EEF1F4; }
.job-loc { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: #667085; }
.job-date { font-size: 22rpx; color: #667085; }

.section-title { font-size: 30rpx; font-weight: 700; color: #17212B; display: block; margin-bottom: 16rpx; }
.job-desc { font-size: 28rpx; color: #344054; line-height: 1.7; white-space: pre-wrap; }

/* ===== 底部操作栏 ===== */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1px solid #EEF1F4;
  box-shadow: 0 -2px 12px rgba(16, 24, 40, 0.04);
}
.bottom-btn {
  height: 88rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 14px rgba(10, 102, 194, 0.28);
}
.bottom-btn.applied { background: #F1F3F5; color: #667085; box-shadow: none; }
.bottom-btn--disabled { background: #F1F3F5; color: #667085; box-shadow: none; }
.bottom-press { transform: scale(0.98); opacity: 0.9; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }

@media (prefers-reduced-motion: reduce) {
  .card { animation: none !important; transition: none !important; }
}
</style>
