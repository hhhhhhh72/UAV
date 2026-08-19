<template>
  <view class="page-container">
    <!-- ① 胶囊导航 + 标题 -->
    <u-nav-bar title="认证飞手" show-back @back="goBack" />

    <!-- 右下角浮动申请按钮（避开微信胶囊，弹性入场 + 微光呼吸） -->
    <view class="apply-fab" hover-class="apply-fab-hover" :hover-stay-time="80" @tap="applyPilot">
      <view class="fab-icon"><text class="fab-icon-char">飞</text></view>
      <text class="fab-text">{{ applyText }}</text>
    </view>

    <!-- ② 统计横幅（第一视觉锚点） -->
    <view class="stats-banner">
      <view class="stat-cell">
        <view class="stat-icon"><view class="icon-cert" /></view>
        <text class="stat-num" :class="{ 'stat-anim': loaded }">{{ displayStats.totalCerts }}</text>
        <text class="stat-label">证书</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-cell">
        <view class="stat-icon"><view class="icon-plane" /></view>
        <text class="stat-num" :class="{ 'stat-anim': loaded }">{{ displayStats.totalHours }}</text>
        <text class="stat-label">飞行小时</text>
      </view>
      <template v-if="displayStats.hasRating">
        <view class="stat-divider" />
        <view class="stat-cell">
          <view class="stat-icon"><text class="icon-star">★</text></view>
          <text class="stat-num" :class="{ 'stat-anim': loaded }">{{ displayStats.avgRating }}</text>
          <text class="stat-label">平均评分</text>
        </view>
      </template>
    </view>

    <!-- ③ 搜索框 -->
    <view class="search-bar">
      <u-icon name="search" size="32rpx" color="#1E5EFF" />
      <input
        class="search-input"
        v-model="searchText"
        placeholder="搜索认证飞手姓名/编号"
        placeholder-class="search-ph"
        confirm-type="search"
        @confirm="onSearch"
      />
      <view v-if="searchText" class="search-clear" @tap="clearSearch">
        <u-icon name="close" size="24rpx" color="#ADB8C7" />
      </view>
    </view>

    <!-- ④ 飞手卡片列表 -->
    <view v-for="(item, i) in list" :key="item.id" class="card card-anim" :style="{ animationDelay: (i * 60) + 'ms' }" hover-class="card-hover" :hover-stay-time="120" @tap="goDetail(item)">
      <!-- 4.1 卡片头部：头像 + 认证徽章 + 名字/编号 + 评分 -->
      <view class="card-head">
        <view class="avatar-wrap">
          <image
            v-if="item.avatar"
            :src="item.avatar"
            mode="aspectFill"
            class="avatar"
            lazy-load
          />
          <view v-else class="avatar avatar-fallback" :style="{ background: avatarBg(item.real_name) }">
            <text class="avatar-char">{{ firstChar(item.real_name) }}</text>
          </view>
          <view class="cert-badge" />
        </view>
        <view class="head-main">
          <view class="name-row">
            <text class="name">{{ item.real_name || '认证飞手' }}</text>
            <text class="pilot-id">{{ idLabel(item) }}</text>
          </view>
          <view class="sub-row">
            <text class="cert-count">{{ (item.cert_ids || []).length }} 项认证</text>
          </view>
        </view>
        <view class="rating-wrap">
          <view class="star">★</view>
          <text class="rating-num">{{ ratingText(item) }}</text>
          <text class="rating-sub">{{ ratingSub(item) }}</text>
        </view>
      </view>

      <!-- 4.3 数据行（两行四列） -->
      <view class="data-grid">
        <view class="data-item">
          <view class="data-icon data-icon-blue"><view class="icon-cert" /></view>
          <view class="data-body">
            <text class="data-label">证书</text>
            <text class="data-value">{{ (item.cert_ids || []).length }}</text>
          </view>
        </view>
        <view class="data-item">
          <view class="data-icon data-icon-purple"><view class="icon-plane" /></view>
          <view class="data-body">
            <text class="data-label">飞行</text>
            <text class="data-value">{{ item.flight_hours || 0 }} 小时</text>
          </view>
        </view>
        <view class="data-item">
          <view class="data-icon data-icon-green"><view class="icon-check" /></view>
          <view class="data-body">
            <text class="data-label">作业</text>
            <text class="data-value">{{ item.completed_jobs || 0 }}</text>
          </view>
        </view>
        <view class="data-item">
          <view class="data-icon data-icon-orange"><view class="icon-target" /></view>
          <view class="data-body">
            <text class="data-label">擅长</text>
            <text class="data-value ellipsis">{{ mainSkill(item) }}</text>
          </view>
        </view>
      </view>

      <!-- 4.4 作业类型标签（7 类分色） -->
      <view v-if="jobTags(item).length > 0" class="tag-row">
        <text
          v-for="(t, ti) in shownTags(item)"
          :key="ti"
          class="job-tag tag-pop"
          :style="tagStyle(t)"
        >{{ t }}</text>
        <text v-if="jobTags(item).length > 3" class="more-tag">+{{ jobTags(item).length - 3 }}</text>
      </view>

      <!-- 4.5 底部：查看飞手档案（整行可点击） -->
      <view class="card-footer" hover-class="footer-hover" :hover-stay-time="120" @tap.stop="goDetail(item)">
        <text class="card-hint">查看飞手档案</text>
        <text class="footer-arrow">›</text>
      </view>
    </view>

    <!-- 列表底部 -->
    <view v-if="!loading && list.length > 0" class="list-footer">
      <text class="footer-line" />
      <text class="footer-text">没有更多了</text>
      <text class="footer-line" />
    </view>

    <!-- 错误态 -->
    <view v-if="!loading && errorMsg" class="empty-wrap">
      <view class="empty-icon"><text>飞</text></view>
      <text class="empty-title">{{ errorMsg }}</text>
      <view class="empty-btn" @tap="fetchData">重新加载</view>
    </view>

    <!-- 空态 -->
    <view v-if="!loading && !errorMsg && !list.length" class="empty-wrap">
      <view class="empty-icon"><text>飞</text></view>
      <text class="empty-title">暂无认证飞手</text>
      <text class="empty-sub">成为协会认证飞手，即可展示在此名录</text>
      <view class="empty-btn" @tap="applyPilot">申请认证</view>
    </view>

    <!-- 加载态 -->
    <view v-if="loading && !list.length" class="loading-wrap">
      <u-loading size="32rpx" />
      <text class="loading-text">加载中...</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'

const searchText = ref('')
const list = ref([])
const loading = ref(false)
const loaded = ref(false)
const errorMsg = ref('')
const goBack = () => uni.navigateBack()

// 右上按钮文案：随我的认证状态变化，入口永远存在
const applyText = ref('申请认证')
const refreshMineLabel = async () => {
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = res && res.data ? res.data : res
    if (mine && mine.id) {
      applyText.value = { pending: '审核中', approved: '我的档案', rejected: '重新申请' }[mine.status] || '申请认证'
    } else {
      applyText.value = '申请认证'
    }
  } catch (e) { applyText.value = '申请认证' }
}

// ── 统计数字滚动计数 ─────────────────────
const displayStats = ref({ totalCerts: 0, totalHours: 0, avgRating: '—', hasRating: false })
const countUp = (target, duration = 800) => {
  const from = { totalCerts: 0, totalHours: 0 }
  const to = { totalCerts: target.totalCerts, totalHours: target.totalHours }
  const start = Date.now()
  // 微信小程序无 requestAnimationFrame，用 setTimeout 模拟帧（16ms）
  const tick = () => {
    const p = Math.min(1, (Date.now() - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3) // easeOutCubic
    displayStats.value = {
      totalCerts: Math.round(from.totalCerts + (to.totalCerts - from.totalCerts) * ease),
      totalHours: Math.round(from.totalHours + (to.totalHours - from.totalHours) * ease),
      avgRating: target.avgRating,
      hasRating: target.hasRating,
    }
    if (p < 1) setTimeout(tick, 16)
  }
  tick()
}

// ── 统计横幅（真实聚合）──────────────────────
const stats = computed(() => {
  const totalCerts = list.value.reduce((s, p) => s + (p.cert_ids || []).length, 0)
  const totalHours = list.value.reduce((s, p) => s + (p.flight_hours || 0), 0)
  const rated = list.value.filter((p) => p.rating > 0)
  const avg = rated.length ? (rated.reduce((s, p) => s + p.rating, 0) / rated.length) : 0
  const hasRating = avg > 0
  return {
    totalCerts,
    totalHours,
    hasRating,
    // 无真实评分时不展示评分项（横幅由 hasRating 控制显隐）
    avgRating: hasRating ? avg.toFixed(1) : '—',
  }
})

// ── 头像兜底：姓名首字 + 姓名哈希选渐变（每人不同）──
const AVATAR_GRADIENTS = [
  'linear-gradient(135deg,#0A1F44,#1E5EFF)',
  'linear-gradient(135deg,#6D28D9,#DB2777)',
  'linear-gradient(135deg,#0EA5E9,#06B6D4)',
  'linear-gradient(135deg,#FF8E3C,#F97316)',
  'linear-gradient(135deg,#00C896,#34c759)',
]
const firstChar = (name) => String(name || '飞').charAt(0)
const avatarBg = (name) => {
  const n = String(name || '')
  if (!n) return AVATAR_GRADIENTS[0]
  let h = 0
  for (let i = 0; i < n.length; i++) h = (h * 31 + n.charCodeAt(i)) >>> 0
  return AVATAR_GRADIENTS[h % AVATAR_GRADIENTS.length]
}

// ── 卡片字段映射 ──────────────────────────
// 编号：后端无编号字段，展示"协会认证 · N 项证书"
const idLabel = (item) => {
  const n = (item.cert_ids || []).length
  return `协会认证 · ${n} 项证书`
}
// 评分：rating>0 显示实际评分；无评分显示"—"（不伪造满分）
const ratingText = (item) => (item.rating > 0 ? item.rating.toFixed(1) : '—')
// 评价数：≥10 才显示"XX 人评价"，否则显示"暂无评价"（弱化）
const ratingSub = (item) => {
  const n = item.completed_jobs || 0
  return n >= 10 ? `${n} 人评价` : '暂无评价'
}
// 擅长（替代地址）：bio 拆出的第一个标签
const bioList = (bio) => String(bio || '').split(/[/，,、\s]+/).filter(Boolean)
const mainSkill = (item) => bioList(item.bio)[0] || '综合'

// ── 作业类型标签（7 类分色）──────────────
const JOB_TAG_MAP = [
  { key: ['电力巡检', '巡检'], color: '#1E5EFF', bg: 'rgba(30,94,255,.08)' },
  { key: ['测绘', '航拍', '拍摄'], color: '#8B5CF6', bg: 'rgba(139,92,246,.08)' },
  { key: ['植保', '喷洒'], color: '#00C896', bg: 'rgba(0,200,150,.08)' },
  { key: ['应急', '救援', '侦察'], color: '#EF4444', bg: 'rgba(239,68,68,.08)' },
  { key: ['吊运', '吊装', '实操'], color: '#FF8E3C', bg: 'rgba(255,142,60,.08)' },
  { key: ['物流', '运输', '投送'], color: '#06B6D4', bg: 'rgba(6,182,212,.08)' },
  { key: ['航拍', '宣传'], color: '#EC4899', bg: 'rgba(236,72,153,.08)' },
]
const matchTag = (tag) => {
  for (const m of JOB_TAG_MAP) {
    if (m.key.some((k) => tag.includes(k))) return m
  }
  return { color: '#6B7B95', bg: 'rgba(107,123,149,.08)' }
}
const jobTags = (item) => bioList(item.bio)
const shownTags = (item) => jobTags(item).slice(0, 3)
const tagStyle = (t) => {
  const m = matchTag(t)
  return { color: m.color, background: m.bg, borderColor: m.color }
}

// ── 数据加载 ────────────────────────────
const goDetail = (item) => {
  uni.setStorageSync('pilot_detail', item)
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(item.id) })
}
const onSearch = () => fetchData()
const clearSearch = () => { searchText.value = ''; fetchData() }

const fetchData = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const kw = searchText.value.trim()
    const res = await request({ url: '/api/v1/certified-pilots', data: { page: 1, page_size: 100, keyword: kw } })
    list.value = (Array.isArray(res) ? res : (res.data || []))
  } catch {
    list.value = []
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
    loaded.value = true
    // 统计数字滚动动画（数据就绪后触发）
    countUp(stats.value)
  }
}

// ---- 申请认证 / 我的状态 ----
const applyPilot = async () => {
  if (!requireLogin()) return
  let mine = null
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    mine = res && res.data ? res.data : res
  } catch (e) {}
  if (mine && mine.id) {
    const label = { pending: '待审核', approved: '已认证', rejected: '未通过' }[mine.status] || mine.status
    // 已认证 → 查看我的档案（detail 页按 id 回源）；驳回 → 重新提交（后端支持覆盖重提）；待审核 → 仅提示
    if (mine.status === 'approved') {
      uni.showModal({
        title: '我的飞手认证',
        content: `当前状态：${label}，已展示在认证飞手名录中`,
        confirmText: '查看档案',
        success: (r) => {
          if (r.confirm) {
            uni.removeStorageSync('pilot_detail')
            uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(mine.id) })
          }
        },
      })
      return
    }
    if (mine.status === 'rejected') {
      uni.showModal({
        title: '我的飞手认证',
        content: `当前状态：${label}，可修改资料后重新提交\n${mine.real_name || ''}`,
        confirmText: '重新提交',
        success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' }) },
      })
      return
    }
    uni.showModal({ title: '我的飞手认证', content: `当前状态：${label}\n${mine.real_name || ''}`, showCancel: false, confirmText: '知道了' })
    return
  }
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

onLoad(() => { fetchData(); refreshMineLabel() })
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: 200rpx; /* 给右下角浮动按钮留空间 */
}

/* ═══ 右下角浮动申请按钮 ═══ */
.apply-fab {
  position: fixed;
  right: 32rpx;
  bottom: calc(48rpx + env(safe-area-inset-bottom));
  z-index: 60;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 14rpx 32rpx 14rpx 14rpx;
  background: linear-gradient(135deg, #0A66C2 0%, #0E7AE0 100%);
  border-radius: 999rpx;
  box-shadow: 0 10rpx 30rpx rgba(10, 102, 194, 0.35);
  animation: fab-in 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
/* 微光呼吸（青绿辅色光晕，呼应品牌辅色） */
.apply-fab::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 999rpx;
  animation: fab-glow 2.6s ease-in-out infinite;
  pointer-events: none;
}
.apply-fab-hover {
  transform: scale(0.94);
  opacity: 0.92;
}
.fab-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}
.fab-icon-char {
  font-size: 28rpx;
  font-weight: 700;
  color: #fff;
}
.fab-text {
  font-size: 28rpx;
  font-weight: 600;
  color: #fff;
}
/* 弹性入场：上浮 + 缩放回弹 */
@keyframes fab-in {
  from { opacity: 0; transform: translateY(60rpx) scale(0.6); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
/* 呼吸光晕 */
@keyframes fab-glow {
  0%, 100% { box-shadow: 0 0 18rpx rgba(29, 212, 168, 0.35); }
  50% { box-shadow: 0 0 36rpx rgba(29, 212, 168, 0.6); }
}

/* ═══ ② 统计横幅 ═══ */
.stats-banner {
  display: flex;
  align-items: center;
  margin: 20rpx 24rpx 0;
  background: linear-gradient(135deg, #EEF3FA 0%, #F0F6FF 100%);
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
}
.stat-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
}
.stat-icon {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: #DCE9FF;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4rpx;
}
.stat-num {
  font-size: 40rpx;
  font-weight: 700;
  color: #0A1F44;
  line-height: 1.2;
}
.stat-anim {
  animation: statPop 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}
.stat-label {
  font-size: 22rpx;
  color: #6B7B95;
}
.stat-divider {
  width: 1rpx;
  height: 72rpx;
  background: linear-gradient(180deg, rgba(30,94,255,0), rgba(30,94,255,0.1) 50%, rgba(30,94,255,0));
}

/* ═══ ③ 搜索框 ═══ */
.search-bar {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin: 20rpx 24rpx;
  background: #F5F8FC;
  border-radius: 999rpx;
  padding: 18rpx 26rpx;
  border: 1rpx solid #EDF0F5;
}
.search-input {
  flex: 1;
  font-size: 26rpx;
  color: #17212B;
}
.search-ph {
  color: #ADB8C7;
}
.search-clear {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #EDF0F5;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ═══ ④ 飞手卡片 ═══ */
.card {
  margin: 16rpx 24rpx;
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
  transition: background 0.2s, transform 0.2s ease;
}
.card-anim {
  animation: cardIn 0.45s ease both;
}
.card-hover {
  background: #EEF3FA;
  transform: scale(0.98);
}

/* 4.1 卡片头部 */
.card-head {
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.avatar-wrap {
  position: relative;
  flex-shrink: 0;
}
.avatar {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  animation: fadeIn 0.4s ease;
}
.avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-char {
  font-size: 36rpx;
  font-weight: 700;
  color: #ffffff;
}
.cert-badge {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: #00C896;
  border: 3rpx solid #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-badge::after {
  content: '';
  position: absolute;
  left: 9rpx;
  top: 7rpx;
  width: 10rpx;
  height: 14rpx;
  border: solid #fff;
  border-width: 0 3rpx 3rpx 0;
  transform: rotate(45deg);
}
.head-main {
  flex: 1;
  min-width: 0;
}
.name-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}
.name {
  font-size: 32rpx;
  font-weight: 700;
  color: #0A1F44;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.pilot-id {
  font-size: 22rpx;
  color: #ADB8C7;
  flex-shrink: 0;
}
.sub-row {
  margin-top: 8rpx;
}
.cert-count {
  font-size: 22rpx;
  color: #6B7B95;
}
.rating-wrap {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  flex-shrink: 0;
}
.star {
  font-size: 26rpx;
  color: #FBBF24;
  line-height: 1;
}
.rating-num {
  font-size: 30rpx;
  font-weight: 700;
  color: #0A1F44;
  margin-top: 2rpx;
}
.rating-sub {
  font-size: 20rpx;
  color: #ADB8C7;
  margin-top: 2rpx;
}

/* 4.3 数据行 */
.data-grid {
  display: flex;
  flex-wrap: wrap;
  margin-top: 20rpx;
  background: #FAFBFC;
  border-radius: 12rpx;
  padding: 16rpx 12rpx;
  gap: 12rpx 0;
}
.data-item {
  width: 50%;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 6rpx 12rpx;
}
.data-icon {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.data-icon-blue { background: #DCE9FF; }
.data-icon-purple { background: #F3E8FF; }
.data-icon-green { background: #D1FAE5; }
.data-icon-orange { background: #FED7AA; }

/* ═══ CSS 绘制的符号图标（禁 emoji 规范）═══ */

/* 证书：圆角方块 + 中缝 */
.icon-cert {
  width: 16rpx;
  height: 20rpx;
  border-radius: 3rpx;
  background: #1E5EFF;
  position: relative;
}
.icon-cert::after {
  content: '';
  position: absolute;
  left: 3rpx;
  right: 3rpx;
  top: 8rpx;
  height: 2rpx;
  background: #DCE9FF;
  box-shadow: 0 4rpx 0 #DCE9FF;
}

/* 纸飞机：三角翼 */
.icon-plane {
  width: 0;
  height: 0;
  border-top: 5rpx solid transparent;
  border-bottom: 5rpx solid transparent;
  border-right: 18rpx solid #8B5CF6;
  position: relative;
  transform: rotate(-8deg);
}
.icon-plane::after {
  content: '';
  position: absolute;
  right: -16rpx;
  top: -4rpx;
  width: 0;
  height: 0;
  border-left: 4rpx solid transparent;
  border-right: 4rpx solid transparent;
  border-bottom: 8rpx solid #8B5CF6;
}

/* 对勾：作业完成 */
.icon-check {
  width: 14rpx;
  height: 24rpx;
  border: solid #00C896;
  border-width: 0 4rpx 4rpx 0;
  transform: rotate(45deg);
  margin-top: -6rpx;
}

/* 目标：圆环 + 中心点 */
.icon-target {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  border: 4rpx solid #FF8E3C;
  position: relative;
}
.icon-target::after {
  content: '';
  position: absolute;
  left: 4rpx;
  top: 4rpx;
  width: 4rpx;
  height: 4rpx;
  border-radius: 50%;
  background: #FF8E3C;
}

/* 星星（统计横幅） */
.icon-star {
  font-size: 24rpx;
  color: #FBBF24;
  line-height: 1;
}
.data-body {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.data-label {
  font-size: 20rpx;
  color: #6B7B95;
}
.data-value {
  font-size: 26rpx;
  font-weight: 700;
  color: #0A1F44;
}
.data-value.ellipsis {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 200rpx;
}

/* 4.4 作业标签 */
.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 16rpx;
}
.job-tag {
  font-size: 20rpx;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  border: 1rpx solid;
  font-weight: 600;
  max-width: 200rpx;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.more-tag {
  font-size: 20rpx;
  color: #6B7B95;
  background: rgba(107,123,149,.08);
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  font-weight: 600;
}

/* 4.5 卡片底部：整行可点击 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16rpx;
  padding: 0 16rpx;
  height: 48rpx;
  border-top: 1rpx solid #E8EEF7;
  border-radius: 0 0 16rpx 16rpx;
  transition: background 0.2s;
}
.footer-hover {
  background: #EEF3FA;
}
.card-hint {
  font-size: 26rpx;
  color: #1E5EFF;
  font-weight: 600;
}
.footer-arrow {
  font-size: 36rpx;
  color: #1E5EFF;
  line-height: 1;
  font-weight: 300;
}

/* ═══ ⑤ 列表底部 ═══ */
.list-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20rpx;
  padding: 48rpx 0;
}
.footer-line {
  width: 100rpx;
  height: 1rpx;
  background: linear-gradient(90deg, rgba(107,123,149,0), rgba(107,123,149,.4), rgba(107,123,149,0));
}
.footer-text {
  font-size: 22rpx;
  color: #ADB8C7;
}

/* ═══ 空态 ═══ */
.empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120rpx;
}
.empty-icon {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: #DCE9FF;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 56rpx;
  color: #1E5EFF;
  font-weight: 700;
}
.empty-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #0A1F44;
  margin-top: 24rpx;
}
.empty-sub {
  font-size: 24rpx;
  color: #6B7B95;
  margin-top: 10rpx;
}
.empty-btn {
  margin-top: 32rpx;
  padding: 16rpx 56rpx;
  background: #1E5EFF;
  border-radius: 999rpx;
  color: #ffffff;
  font-size: 26rpx;
  font-weight: 600;
}

/* ═══ 加载态 ═══ */
.loading-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 120rpx 0;
}
.loading-text {
  font-size: 24rpx;
  color: #6B7B95;
}

/* ═══ 微动效 ═══ */
@keyframes statPop {
  from {
    transform: scale(0.8);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes cardIn {
  from {
    transform: translateY(24rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
/* 标签浮起微动效 */
.tag-pop {
  transition: transform 0.2s ease;
  animation: tagIn 0.4s ease both;
}
@keyframes tagIn {
  from {
    transform: scale(0.7);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
</style>
