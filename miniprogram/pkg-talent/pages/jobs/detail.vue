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
          <text v-if="job.job_type" class="type-tag">{{ job.job_type }}</text>
          <text v-if="job.status === 'published'" class="st-tag">招聘中</text>
        </view>
        <view class="job-sub">
          <view v-if="job.location" class="job-loc"><text>{{ job.location }}</text></view>
          <text>{{ formatDate(job.created_at) }} 发布</text>
        </view>
      </view>

      <!-- 职位描述 -->
      <view class="card section-card">
        <text class="section-title">职位描述</text>
        <text class="job-desc">{{ job.description || '暂无描述' }}</text>
      </view>

      <!-- 底部操作 -->
      <view class="bottom-bar">
        <view class="bottom-btn" :class="{ applied: applied }" hover-class="bottom-press" :hover-stay-time="100" @tap="applyJob">{{ applied ? '已投递' : '投递简历' }}</view>
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
const loading = ref(true)
const errorMsg = ref('')

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
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

// 投递：登录 → 简历检查 → 提交
const applyJob = async () => {
  if (!job.value || applied.value) return
  if (!requireLogin()) return
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
  background: #fff;
  padding-bottom: 120rpx;
}

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 卡片：白上白 ===== */
.card {
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.head-card { margin: 12px; padding: 16px; animation: cardIn .22s ease-out backwards; animation-delay: 60ms; }
.section-card { margin: 0 12px 12px; padding: 16px; animation: cardIn .22s ease-out backwards; animation-delay: 120ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.job-title { font-size: 18px; font-weight: 700; color: #17212B; display: block; line-height: 1.4; }
.job-meta { display: flex; align-items: center; gap: 8px; margin-top: 10px; flex-wrap: wrap; }
.salary { font-size: 16px; font-weight: 700; color: #C2410C; }
.type-tag {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 4px;
  background: #EAF3FB;
  color: #0A66C2;
  font-weight: 600;
}
.st-tag {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 4px;
  background: #E9F7F0;
  color: #0B6B41;
  font-weight: 600;
}
.job-sub { display: flex; align-items: center; gap: 12px; font-size: 12px; color: #667085; margin-top: 10px; }
.job-loc { font-size: 12px; color: #667085; }

.section-title { font-size: 14px; font-weight: 700; color: #17212B; display: block; margin-bottom: 10px; }
.job-desc { font-size: 13px; color: #344054; line-height: 1.7; white-space: pre-wrap; }

/* ===== 底部操作栏 ===== */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1px solid #EEF1F4;
  box-shadow: 0 -2px 12px rgba(16, 24, 40, 0.04);
}
.bottom-btn {
  height: 46px;
  border-radius: 23px;
  background: #0A66C2;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 14px rgba(10, 102, 194, 0.28);
}
.bottom-btn.applied { background: #EEF1F4; color: #667085; box-shadow: none; }
.bottom-press { transform: scale(0.98); opacity: 0.9; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
</style>
