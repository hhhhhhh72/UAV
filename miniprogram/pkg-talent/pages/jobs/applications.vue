<template>
  <view class="apply-page">
    <u-nav-bar title="我的投递" show-back @back="goBack" />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 空态 -->
    <view v-else-if="!list.length" class="state-panel">
      <view class="state-mark">投</view>
      <text class="state-title">还没有投递记录</text>
      <text class="state-desc">在招聘大厅找到合适的职位，投递后在这里跟踪进展</text>
      <view class="state-btn" @tap="goHall">去看职位</view>
    </view>

    <template v-else>
      <view class="list-head">
        <text class="list-title">投递记录</text>
        <text class="list-count">共 {{ list.length }} 条</text>
      </view>

      <view class="apply-list">
        <view
          v-for="item in list"
          :key="item.id"
          class="apply-card"
          hover-class="tap-scale"
          :hover-stay-time="100"
          @tap="openJob(item)"
        >
          <view class="tag-row">
            <text class="tag" :class="statusMeta(item.status).cls">{{ statusMeta(item.status).label }}</text>
            <text v-if="jobOf(item) && jobOf(item).job_type" class="tag tag-blue">{{ jobOf(item).job_type }}</text>
            <text v-if="jobTag(item)" class="tag" :class="jobTag(item).cls">{{ jobTag(item).text }}</text>
          </view>

          <text class="apply-job" :class="{ gone: !jobOf(item) }">{{ jobTitle(item) }}</text>

          <view class="apply-meta">
            <text v-if="item.enterprise_name" class="meta-item">{{ item.enterprise_name }}</text>
            <view v-if="jobOf(item) && jobOf(item).location" class="meta-item">
              <u-icon name="location" size="24rpx" color="#667085" />
              <text>{{ jobOf(item).location }}</text>
            </view>
            <text v-if="jobOf(item)" class="meta-item salary">{{ salaryText(jobOf(item)) }}</text>
          </view>

          <text class="apply-time">投递于 {{ formatTime(item.created_at) }}</text>

          <view class="action-row">
            <view v-if="canOpen(item)" class="action-link" @tap.stop="openJob(item)">查看职位</view>
            <view v-if="canWithdraw(item)" class="action-link danger" @tap.stop="withdraw(item)">
              {{ withdrawing === item.id ? '撤回中...' : '撤回投递' }}
            </view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'

const goBack = () => uni.navigateBack()
const goHall = () => uni.navigateTo({ url: '/pkg-talent/pages/jobs/list' })

const list = ref([])
const loading = ref(false)
const withdrawing = ref('')

const STATUS_META = {
  submitted: { label: '已投递', cls: 'tag-blue' },
  viewed: { label: '已查看', cls: 'tag-orange' },
  interviewing: { label: '面试中', cls: 'tag-purple' },
  offered: { label: '已录用', cls: 'tag-green' },
  rejected: { label: '未通过', cls: 'tag-red' },
  withdrawn: { label: '已撤回', cls: 'tag-gray' },
}
const statusMeta = (s) => STATUS_META[s] || { label: s || '未知', cls: 'tag-gray' }
const canWithdraw = (item) => item.status === 'submitted' || item.status === 'viewed'

// 职位快照由后端 /api/v1/applications 内联返回（job 为 null = 职位已删除）
const jobOf = (item) => item.job || null
const jobTitle = (item) => {
  const j = jobOf(item)
  return j && j.title ? j.title : '该职位已删除'
}
// 职位状态标签：招聘中不必强调（卡片可点即是在招），只标出"看不到详情"的两种情况
const jobTag = (item) => {
  const j = jobOf(item)
  if (!j) return { text: '已删除', cls: 'tag-gray' }
  if (j.status === 'closed') return { text: '已关闭', cls: 'tag-gray' }
  if (j.status !== 'published') return { text: '未发布', cls: 'tag-gray' }
  return null
}
// 职位详情接口只对"招聘中"职位开放（草稿/已关闭返回 404），据此决定能否跳转
const canOpen = (item) => {
  const j = jobOf(item)
  return !!(j && j.status === 'published')
}
const salaryText = (job) => {
  if (job && job.salary_fen) return '¥' + (job.salary_fen / 100).toLocaleString('zh-CN') + '/月'
  return '薪资面议'
}

const normalizeList = (res) => {
  if (Array.isArray(res)) return res
  if (res && Array.isArray(res.data)) return res.data
  if (res && Array.isArray(res.items)) return res.items
  return []
}

// 时间：今天/昨天用相对日，跨年补全年份（只用真实 created_at）
const formatTime = (iso) => {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const p = (n) => String(n).padStart(2, '0')
  const hm = p(d.getHours()) + ':' + p(d.getMinutes())
  const now = new Date()
  const day = (x) => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime()
  const diff = Math.round((day(now) - day(d)) / 86400000)
  if (diff === 0) return '今天 ' + hm
  if (diff === 1) return '昨天 ' + hm
  const md = p(d.getMonth() + 1) + '-' + p(d.getDate())
  if (d.getFullYear() === now.getFullYear()) return md + ' ' + hm
  return d.getFullYear() + '-' + md + ' ' + hm
}

const load = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/applications' })
    list.value = normalizeList(res)
  } catch (e) {
    uni.showToast({ title: '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

// 职位已删除/已关闭时点卡片给明确反馈，而不是静默无响应
const openJob = (item) => {
  if (canOpen(item)) {
    uni.navigateTo({ url: '/pkg-talent/pages/jobs/detail?id=' + encodeURIComponent(item.job_id) })
    return
  }
  uni.showToast({ title: jobOf(item) ? '该职位已关闭，无法查看详情' : '该职位已删除', icon: 'none' })
}

// 撤回投递：submitted/viewed 可撤回，成功刷新列表
const withdraw = async (item) => {
  if (withdrawing.value) return
  const confirm = await new Promise((resolve) => {
    uni.showModal({
      title: '撤回投递',
      content: '确定撤回该投递？撤回后企业将无法查看您的简历。',
      success: (r) => resolve(r.confirm),
    })
  })
  if (!confirm) return
  withdrawing.value = item.id
  try {
    await request({
      url: '/api/v1/applications/' + encodeURIComponent(item.id) + '/status',
      method: 'PATCH',
      data: { status: 'withdrawn' },
    })
    uni.showToast({ title: '已撤回', icon: 'success' })
    load()
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '撤回失败，请稍后重试', icon: 'none' })
  } finally {
    withdrawing.value = ''
  }
}

// onShow 而非 onMounted：操作返回后立即看到最新状态
onShow(load)
</script>

<style scoped>
.apply-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* ===== 头部：标题 + 计数 ===== */
.list-head {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
  padding: 32rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 700; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

/* ===== 投递卡片（对齐「我的发布」卡片规范） ===== */
.apply-list { display: flex; flex-direction: column; gap: 20rpx; padding: 0 32rpx; }
.apply-card {
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 26rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.tap-scale { transform: scale(0.98); opacity: 0.92; }

.tag-row { display: flex; flex-wrap: wrap; gap: 10rpx; }
.tag { border-radius: 8rpx; padding: 6rpx 12rpx; font-size: 20rpx; line-height: 1; }
.tag-blue { color: #0A66C2; background: #EAF3FB; }
.tag-green { color: #168A55; background: #E9F7F0; }
.tag-orange { color: #DB5F0D; background: #FFF0E6; }
.tag-purple { color: #7B61D1; background: #F0EDFF; }
.tag-red { color: #D92D20; background: #FEF3F2; }
.tag-gray { color: #667085; background: #F1F3F5; }

.apply-job {
  display: block;
  margin: 16rpx 0 10rpx;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
}
.apply-job.gone { color: #667085; font-weight: 600; }
.apply-meta { display: flex; flex-wrap: wrap; align-items: center; gap: 8rpx 24rpx; }
.meta-item { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: #667085; }
.meta-item.salary { color: #E96012; font-weight: 600; }
.apply-time { display: block; margin-top: 12rpx; font-size: 22rpx; color: #667085; }

.action-row {
  display: flex;
  gap: 40rpx;
  margin-top: 22rpx;
  padding-top: 20rpx;
  border-top: 1px solid #EEF1F4;
}
.action-link { min-height: 64rpx; display: inline-flex; align-items: center; font-size: 24rpx; font-weight: 600; color: #0A66C2; }
.action-link.danger { color: #D92D20; }

/* ===== 状态面板（空态，全局同款） ===== */
.state-panel {
  min-height: 620rpx;
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
  font-weight: 700;
}
.state-title { font-size: 30rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 14rpx 0 32rpx; font-size: 24rpx; color: #667085; line-height: 1.6; }
.state-btn {
  height: 76rpx;
  padding: 0 40rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 26rpx;
  font-weight: 600;
  line-height: 76rpx;
}

/* ===== 加载中 ===== */
.loading-state { display: flex; align-items: center; justify-content: center; gap: 16rpx; padding: 120rpx 0; }
.loading-text { font-size: 24rpx; color: #667085; }

/* ===== 减弱动效（无障碍） ===== */
@media (prefers-reduced-motion: reduce) {
  .apply-card { transition: none !important; }
}
</style>
