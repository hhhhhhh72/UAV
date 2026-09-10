<template>
  <view class="mine-page">
    <u-nav-bar title="我的招聘" show-back @back="goBack" />

    <!-- 非企业账号：引导入驻 -->
    <view v-if="!isEnterprise" class="state-panel">
      <view class="state-mark">聘</view>
      <text class="state-title">仅企业账号可管理招聘</text>
      <text class="state-desc">完成企业入驻并审核通过后，即可发布与管理招聘职位</text>
      <view class="state-btn" @tap="goRegister">去企业入驻</view>
    </view>

    <template v-else>
      <!-- 头部：统计 + 发布入口 -->
      <view class="list-head">
        <view class="head-left">
          <text class="list-title">职位记录</text>
          <text class="list-count">共 {{ list.length }} 个</text>
        </view>
        <view class="head-btn" hover-class="head-btn--press" :hover-stay-time="100" @tap="goPublish">
          <u-icon name="plus" size="24rpx" color="#FFFFFF" />
          <text>发布新职位</text>
        </view>
      </view>

      <!-- 状态筛选：下划线 tab（带计数，单维度 tab 即维度） -->
      <view class="stage-wrap">
        <view class="stages">
          <view
            v-for="t in statusTabs"
            :key="t.value"
            class="stg"
            :class="{ on: activeStatus === t.value }"
            @tap="activeStatus = t.value"
          >
            <text>{{ t.label }}</text>
            <text v-if="t.count" class="stg-num">{{ t.count }}</text>
          </view>
        </view>
      </view>

      <!-- 草稿提示：仅在有草稿时出现，避免常驻噪音 -->
      <view v-if="draftCount" class="note">草稿职位不对外展示，点「发布上线」后求职者才能看到</view>

      <!-- Loading -->
      <view v-if="loading" class="loading-state">
        <u-loading size="28rpx" />
        <text class="loading-text">加载中...</text>
      </view>

      <!-- 空态：未发布 / 筛选无结果 -->
      <view v-else-if="!list.length" class="state-panel">
        <view class="state-mark">聘</view>
        <text class="state-title">还没有发布职位</text>
        <text class="state-desc">发布职位后，求职者可在招聘大厅查看并投递简历</text>
        <view class="state-btn" @tap="goPublish">发布新职位</view>
      </view>
      <view v-else-if="!filtered.length" class="state-panel">
        <view class="state-mark">聘</view>
        <text class="state-title">当前状态下暂无职位</text>
        <text class="state-desc">换个状态看看，或返回「全部」</text>
        <view class="state-btn" @tap="activeStatus = ''">查看全部</view>
      </view>

      <!-- 职位列表 -->
      <view v-else class="job-list">
        <view
          v-for="item in filtered"
          :key="item.id"
          class="job-card"
          hover-class="tap-scale"
          :hover-stay-time="100"
          @tap="openDetail(item)"
        >
          <view class="tag-row">
            <text class="tag" :class="statusMeta(item.status).cls">{{ statusMeta(item.status).label }}</text>
            <text v-if="item.job_type" class="tag tag-blue">{{ item.job_type }}</text>
          </view>
          <text class="job-title">{{ item.title }}</text>
          <view class="job-meta">
            <view v-if="item.location" class="meta-item">
              <u-icon name="location" size="24rpx" color="#667085" />
              <text>{{ item.location }}</text>
            </view>
            <text class="meta-item salary">{{ salaryText(item) }}</text>
          </view>
          <text class="job-time">{{ timeText(item) }}</text>
          <view class="action-row">
            <view v-if="item.status !== 'draft'" class="action-link" @tap.stop="goApplicants(item)">查看投递</view>
            <view v-if="item.status === 'draft'" class="action-link" @tap.stop="publishJob(item)">发布上线</view>
            <view v-if="item.status === 'published'" class="action-link danger" @tap.stop="closeJob(item)">关闭职位</view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
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
const activeStatus = ref('')

// 状态文案/配色：与全站标签色板同源（绿=进行中 橙=待处理 灰=已结束）
const STATUS_META = {
  draft: { label: '草稿', cls: 'tag-orange' },
  published: { label: '招聘中', cls: 'tag-green' },
  closed: { label: '已关闭', cls: 'tag-gray' },
}
const statusMeta = (s) => STATUS_META[s] || { label: s || '未知', cls: 'tag-gray' }

const filtered = computed(() => {
  if (!activeStatus.value) return list.value
  return list.value.filter((j) => j.status === activeStatus.value)
})
const countOf = (s) => list.value.filter((j) => j.status === s).length
const draftCount = computed(() => countOf('draft'))
const statusTabs = computed(() => [
  { value: '', label: '全部', count: list.value.length },
  { value: 'draft', label: '草稿', count: countOf('draft') },
  { value: 'published', label: '招聘中', count: countOf('published') },
  { value: 'closed', label: '已关闭', count: countOf('closed') },
])

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

const salaryText = (item) => {
  if (item && item.salary_fen) return '¥' + (item.salary_fen / 100).toLocaleString('zh-CN') + '/月'
  return '薪资面议'
}

const formatDate = (iso) => {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const p = (n) => String(n).padStart(2, '0')
  return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate())
}

// 时间行只用真实字段：创建时间恒有，更新时间与创建不同才追加（不编造发布/关闭时间）
const timeText = (item) => {
  const created = formatDate(item.created_at)
  const updated = formatDate(item.updated_at)
  let text = created ? '创建 ' + created : ''
  if (updated && updated !== created) text += (text ? ' · ' : '') + '更新 ' + updated
  return text
}

// 卡片点击 → 职位详情：仅"招聘中"职位对公众开放（草稿/已关闭公开详情返回 404）
const openDetail = (item) => {
  if (item.status === 'published') {
    uni.navigateTo({ url: '/pkg-talent/pages/jobs/detail?id=' + encodeURIComponent(item.id) })
    return
  }
  uni.showToast({ title: item.status === 'draft' ? '草稿职位不对外展示' : '已关闭职位不对外展示', icon: 'none' })
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

// 关闭不可逆（状态机只允许 draft→published→closed，无重新开启路径）→ 二次确认
const closeJob = async (item) => {
  const ok = await new Promise((resolve) => {
    uni.showModal({
      title: '关闭该职位？',
      content: '关闭后职位将从招聘大厅移除，且无法重新开启；如需再次招聘请重新发布。',
      confirmText: '关闭职位',
      confirmColor: '#D92D20',
      success: (r) => resolve(!!r.confirm),
    })
  })
  if (!ok) return
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
.mine-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* ===== 头部：标题 + 计数 + 发布入口 ===== */
.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  padding: 32rpx 32rpx 16rpx;
}
.head-left { display: flex; align-items: baseline; gap: 16rpx; min-width: 0; }
.list-title { font-size: 36rpx; font-weight: 700; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }
.head-btn {
  display: flex;
  align-items: center;
  gap: 8rpx;
  height: 72rpx;
  padding: 0 28rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 26rpx;
  font-weight: 600;
  box-shadow: 0 6rpx 16rpx rgba(10, 102, 194, 0.22);
}
.head-btn--press { transform: scale(0.96); opacity: 0.9; }

/* ===== 状态筛选：下划线 tab（对齐成果库/我的发布） ===== */
.stage-wrap { background: #fff; border-bottom: 1px solid #EEF1F4; }
.stages { display: flex; gap: 40rpx; padding: 4rpx 32rpx 0; white-space: nowrap; }
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 0 8rpx;
  font-size: 26rpx;
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; }
.stg.on::after {
  content: '';
  position: absolute;
  left: 8rpx;
  right: 8rpx;
  bottom: 12rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #074D92;
  animation: toc-in .22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.stg-num { font-size: 20rpx; color: #0A66C2; background: #EAF3FB; border-radius: 8rpx; padding: 2rpx 8rpx; line-height: 1.4; }

/* 草稿提示条 */
.note {
  margin: 20rpx 32rpx 0;
  padding: 18rpx 24rpx;
  border-radius: 12rpx;
  background: #FFF0E6;
  color: #DB5F0D;
  font-size: 22rpx;
  line-height: 1.5;
}

/* ===== 职位卡片（对齐「我的发布」卡片规范） ===== */
.job-list { display: flex; flex-direction: column; gap: 20rpx; padding: 24rpx 32rpx 0; }
.job-card {
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
.tag-gray { color: #667085; background: #F1F3F5; }

.job-title {
  display: block;
  margin: 16rpx 0 10rpx;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
}
.job-meta { display: flex; flex-wrap: wrap; align-items: center; gap: 8rpx 24rpx; }
.meta-item { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: #667085; }
.meta-item.salary { color: #E96012; font-weight: 600; }
.job-time { display: block; margin-top: 12rpx; font-size: 22rpx; color: #667085; }

.action-row {
  display: flex;
  gap: 40rpx;
  margin-top: 22rpx;
  padding-top: 20rpx;
  border-top: 1px solid #EEF1F4;
}
.action-link { min-height: 64rpx; display: inline-flex; align-items: center; font-size: 24rpx; font-weight: 600; color: #0A66C2; }
.action-link.danger { color: #D92D20; }

/* ===== 状态面板（未入驻 / 空态，全局同款） ===== */
.state-panel {
  min-height: 560rpx;
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
  .stg.on::after, .head-btn { animation: none !important; transition: none !important; }
}
</style>
