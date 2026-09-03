<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="赛事不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- 顶部导航（对齐培训详情：独立导航条，白圆钮返回 + 居中标题 + 右侧分享） -->
        <view class="detail-nav" :style="{ paddingTop: statusBarHeight + 'px' }">
          <view class="detail-nav-back" hover-class="detail-nav-back--press" :hover-stay-time="120" aria-role="button" aria-label="返回" @click="goBack">
            <text>‹</text>
          </view>
          <text class="detail-nav-title">赛事详情</text>
          <button class="share-btn" open-type="share" hover-class="detail-nav-back--press" :hover-stay-time="120">
            <text class="share-text">分享</text>
          </button>
        </view>

        <!-- ① Hero（对齐培训详情：内嵌圆角卡片，封面 + 蒙层 + 状态徽章 + 标题贴底） -->
        <view class="hero">
          <!-- 真实赛事海报图（有则覆盖兜底；加载失败回退兜底） -->
          <image
            v-if="heroPoster(detail) && !imgError"
            :src="heroPoster(detail)"
            mode="aspectFill"
            class="hero-img"
            lazy-load
            :style="{ opacity: imgLoaded ? 1 : 0 }"
            @load="onHeroImgLoad"
            @error="onHeroImgError"
          />

          <!-- 兜底装饰层：仅在无海报/海报加载失败时渲染 -->
          <view v-if="!heroPoster(detail) || imgError" class="hero-deco">
            <view class="deco-grid" />
            <view class="deco-radar" />
            <view class="deco-trail" />
          </view>

          <view class="hero-mask" />

          <!-- 状态徽章（对齐培训页：Hero 左上，白底胶囊 + 彩色文字） -->
          <view class="status-badge" :class="statusClass(derivedStatus(detail))">{{ statusTextOf(detail) }}</view>

          <!-- Hero 底部信息区（对齐培训页 hero-bottom：标题 + 机构副标 + 周期 meta 行 + 标签） -->
          <view class="hero-bottom">
            <text class="hero-title">{{ detail.title || detail.name || '未知赛事' }}</text>
            <text class="hero-org">{{ detail.organizer || detail.sponsor || '待定' }}</text>
            <view class="hero-meta-row">
              <view class="meta-ico meta-ico--cal"><view class="cal-top" /><view class="cal-body"><view class="cal-line l1" /><view class="cal-line l2" /><view class="cal-line l3" /></view></view>
              <text class="hero-meta-text">{{ compPeriod(detail) }}</text>
            </view>
            <view class="hero-tags">
              <text v-for="tag in compTags(detail)" :key="tag" class="hero-tag">{{ tag }}</text>
            </view>
          </view>
        </view>

        <!-- ② 内容区（对齐培训详情页：直接在页面底色上排布卡片） -->
        <view class="content">

        <!-- ② 基本信息：时间轴 -->
        <view class="card info-timeline">
          <view class="info-head">
            <text class="info-head-title">赛事信息</text>
            <text class="info-head-status" :class="statusClass(derivedStatus(detail))">{{ statusTextOf(detail) }}</text>
          </view>
          <view class="tl-item">
            <view class="tl-dot" />
            <view class="tl-content">
              <text class="tl-label">比赛时间</text>
              <text class="tl-value">{{ fmtDate(detail.start_date) }} - {{ fmtDate(detail.end_date) }}</text>
            </view>
          </view>
          <view class="tl-item">
            <view class="tl-dot" />
            <view class="tl-content">
              <text class="tl-label">比赛地点</text>
              <text class="tl-value">{{ detail.location || '待定' }}</text>
            </view>
          </view>
          <view class="tl-item">
            <view class="tl-dot tl-dot--deadline" />
            <view class="tl-content">
              <text class="tl-label tl-label--deadline">报名截止</text>
              <text class="tl-value tl-value--deadline">{{ fmtDate(detail.deadline || detail.enroll_deadline) }}</text>
              <text v-if="!isClosed(detail)" class="tl-countdown">{{ countdownText(detail) }}</text>
            </view>
          </view>
        </view>

        <!-- ③ 赛事简介 -->
        <view v-if="detail.intro || detail.description" class="section-block">
          <view class="section-title">赛事简介</view>
          <rich-text v-if="((detail.intro || detail.description) || '').indexOf('<') >= 0" class="card intro-text" :nodes="detail.intro || detail.description"></rich-text>
          <view v-else class="card intro-text">{{ detail.intro || detail.description }}</view>
        </view>

        <!-- ④ 报名条件 -->
        <view class="section-block">
          <view class="section-title">报名条件</view>
          <view class="card requirements-card">
            <view v-if="requirements(detail).length === 0" class="req-empty">以主办方公布为准</view>
            <view v-for="req in requirements(detail)" :key="req.name" class="req-item">
              <view class="req-icon" />
              <view class="req-body">
                <text class="req-name">{{ req.name }}</text>
                <text class="req-desc">{{ req.desc }}</text>
              </view>
              <view class="req-badge" :class="reqBadgeClass(req.level)">{{ req.level }}</view>
            </view>
          </view>
        </view>

        <!-- ⑤ 参赛项目 -->
        <view class="section-block">
          <view class="section-title">参赛项目</view>
          <view class="event-list">
            <view v-if="eventList(detail).length === 0" class="card event-empty">以主办方公布为准</view>
            <view
              v-for="(ev, i) in eventList(detail)"
              :key="ev.name"
              class="card event-item"
              :class="i === 0 ? 'event-item--hot' : ''"
            >
              <view class="event-info">
                <view class="event-name-row">
                  <text class="event-name">{{ ev.name }}</text>
                  <view v-if="i === 0" class="hot-badge">热门</view>
                </view>
                <text class="event-meta">{{ ev.type }} · {{ ev.format }}</text>
              </view>
              <text v-if="ev.fee != null" class="event-price">¥{{ ev.fee.toLocaleString() }}</text>
              <text v-else class="event-price event-price--pending">费用待定</text>
            </view>
          </view>
        </view>

        <!-- ⑥ 奖项 -->
        <view v-if="prizes(detail).length > 0" class="section-block">
          <view class="section-title">奖项设置</view>
          <view class="prize-row">
            <view v-for="p in prizes(detail)" :key="p.level" class="prize-card" :class="'prize-card--' + p.metal">
              <view class="prize-medal" />
              <text class="prize-level">{{ p.level }}</text>
              <view v-if="p.amount != null" class="prize-amount-row">
                <text class="prize-symbol">¥</text>
                <text class="prize-amount">{{ p.amount.toLocaleString() }}</text>
              </view>
            </view>
          </view>
        </view>

        <!-- ⑦ 主办单位 -->
        <view class="section-block">
          <view class="section-title">主办单位</view>
          <view class="card organizer-row">
            <view class="org-avatar" />
            <view class="org-info">
              <text class="org-name">{{ detail.organizer || detail.sponsor || '待定' }}</text>
              <text class="org-sub">{{ detail.organizer_sub || '主办单位' }}</text>
            </view>
          </view>
        </view>

        </view><!-- /content -->

        <!-- ⑧ 底部 CTA -->
        <view class="bottom-bar">
          <view class="bottom-left">
            <text class="fee-label">报名费</text>
            <view class="fee-price">
              <template v-if="compMinFee(detail) === 0">
                <text class="free-badge">免费</text>
              </template>
              <template v-else-if="compMinFee(detail) != null">
                <text class="fee-symbol">¥</text>
                <text class="fee-value">{{ compMinFee(detail) }}</text>
                <text class="fee-unit">起/人</text>
              </template>
              <text v-else class="fee-value fee-value--pending">以主办方公布为准</text>
            </view>
          </view>
          <view class="bottom-actions">
            <view
              class="btn-primary"
              :class="{ disabled: isClosed(detail) }"
              hover-class="press-feedback"
              :hover-stay-time="120"
              @click="goRegister"
            >{{ isClosed(detail) ? '已截止' : '立即报名' }}</view>
          </view>
        </view>
        <view class="bottom-spacer" />
      </template>
    </StateView>
  </view>
</template>

<script setup>
import { safeBack } from '../../../utils/nav'
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onShareAppMessage } from '@dcloudio/uni-app'
import { useReduceMotion } from '../../../utils/motion'
import { request, requireLogin } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const { noMotion, checkMotion } = useReduceMotion()
const statusBarHeight = ref(20)
const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
const imgLoaded = ref(false)
const imgError = ref(false)

const statusMap = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已截止', full: '已截止' }

/* ===== 单一状态源：status 与报名截止时间统一判定，徽章/倒计时/CTA 全部由此派生 ===== */
function deadlineDate(item) {
  const d = item.deadline || item.enroll_deadline
  if (!d) return null
  const t = Date.parse(String(d).replace(/-/g, '/'))
  return isNaN(t) ? null : new Date(t)
}

function derivedStatus(item) {
  if (item.status === 'closed' || item.status === 'full') return 'closed'
  const dl = deadlineDate(item)
  if (dl && dl.getTime() < Date.now()) return 'closed'
  return item.status || 'enrolling'
}

function isClosed(item) {
  return derivedStatus(item) === 'closed'
}

function statusTextOf(item) {
  return statusMap[derivedStatus(item)] || '报名中'
}

function statusClass(status) {
  if (status === 'ongoing') return 'badge--ongoing'
  if (status === 'closed' || status === 'full') return 'badge--closed'
  return 'badge--enrolling'
}

function countdownText(item) {
  const dl = deadlineDate(item)
  if (!dl) return ''
  const days = Math.floor((dl.getTime() - Date.now()) / 86400000)
  if (days < 0) return '已截止'
  if (days === 0) return '今天截止'
  return '剩余 ' + days + ' 天'
}

/** Hero 海报图：兼容 poster / cover / image / banner */
function heroPoster(item) {
  const u = item.poster || item.cover || item.image || item.banner
  return u ? u : ''
}

/** 海报图加载完成淡入 */
function onHeroImgLoad() {
  imgLoaded.value = true
}

/** 海报加载失败：隐藏图片走兜底装饰 */
function onHeroImgError() {
  imgError.value = true
}

/** Hero 左下角首字徽章 */
function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}

function compPeriod(item) {
  var s = fmtDate(item.start_date)
  var e = fmtDate(item.end_date)
  if (s === '待定' && e === '待定') return '赛期待定'
  return e && e !== s ? s + ' ~ ' + e : s
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (item.category) return [item.category]
  return ['多旋翼', '国家级']
}

function requirements(item) {
  if (Array.isArray(item.requirements) && item.requirements.length > 0) return item.requirements
  return []
}

/* 报名条件 badge 三态：必满足(红) / 建议(蓝) / 等级(灰蓝-初/中/高级) */
function reqBadgeClass(level) {
  if (level === '必满足') return 'req-badge--must'
  if (level === '建议') return 'req-badge--advise'
  return 'req-badge--level'
}

function eventList(item) {
  if (Array.isArray(item.events) && item.events.length > 0) return item.events
  return []
}

function prizes(item) {
  if (Array.isArray(item.prizes) && item.prizes.length > 0) return item.prizes
  return []
}

function compMinFee(item) {
  // min_fee 为 int 零值=未配置，视为未填（否则后台未配置时显示 ¥0 起/人与列表不一致）
  if (item.min_fee != null && item.min_fee > 0) return item.min_fee
  if (item.fee != null) return item.fee
  var evts = eventList(item)
  if (evts.length > 0) {
    var fees = evts.map(function (e) { return e.fee }).filter(function (f) { return f != null })
    if (fees.length > 0) return Math.min.apply(null, fees)
  }
  return null
}

async function loadDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    var res = await request({ url: '/api/v1/competitions/' + encodeURIComponent(id.value) })
    detail.value = res
    if (!res) errorMsg.value = '赛事不存在'
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goBack() { safeBack() }

function goRegister() {
  if (detail.value && isClosed(detail.value)) return
  // 登录拦截：未登录点「立即报名」→ 跳登录页（与其他详情页报名按钮一致）
  if (!requireLogin('请先登录后再报名')) return
  uni.navigateTo({ url: '/pkg-eco/pages/competitions/register?id=' + encodeURIComponent(id.value) })
}

onShareAppMessage(function () {
  const item = detail.value || {}
  return {
    title: '赛事报名：' + (item.title || item.name || '无人机赛事'),
    path: '/pkg-eco/pages/competitions/detail?id=' + encodeURIComponent(id.value),
  }
})

onLoad(function (options) {
  checkMotion()
  try {
    const sys = uni.getSystemInfoSync()
    if (sys && sys.statusBarHeight) statusBarHeight.value = sys.statusBarHeight
  } catch (e) {}
  id.value = options.id || ''
  loadDetail()
})

onPullDownRefresh(function () {
  loadDetail().then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: calc(140rpx + env(safe-area-inset-bottom));
  overflow-x: hidden;
}

/* ===== 通用卡片（对齐培训详情页：白卡 + 16px 圆角 + 1rpx 边框 + 柔和投影） ===== */
.card {
  position: relative;
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 16px;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
}

/* ================================================================= */
/* ① Hero 区（对齐培训详情页：内嵌圆角卡片 + 单层蓝蒙层） */
/* ================================================================= */
.hero {
  position: relative;
  width: auto;
  height: 348rpx;
  margin: 0 24rpx;
  border-radius: 24rpx;
  overflow: hidden;
  background: linear-gradient(160deg, #0a5897 0%, #074D92 100%);
  box-shadow: 0 14rpx 34rpx rgba(31, 89, 169, 0.2);
  box-sizing: border-box;
}

/* 真实海报图 */
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity .25s ease-out;
}

/* 单层渐变蒙层（对齐培训页 hero-mask：上淡 → 中透 → 底深蓝） */
.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(4, 30, 68, 0.08) 0%, rgba(4, 30, 68, 0.05) 34%, rgba(4, 30, 68, 0.8) 100%);
  pointer-events: none;
}

/* ===== 兜底装饰层（仅无海报时渲染，静态无动画） ===== */
.hero-deco { position: absolute; inset: 0; pointer-events: none; }

/* 网格点阵 */
.deco-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.12) 2rpx, transparent 2rpx);
  background-size: 40rpx 40rpx;
  opacity: 0.6;
}

/* 雷达同心圆（静态） */
.deco-radar {
  position: absolute;
  right: -80rpx;
  top: -80rpx;
  width: 300rpx;
  height: 300rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.15);
  border-radius: 50%;
}

.deco-radar::before,
.deco-radar::after {
  content: '';
  position: absolute;
  inset: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

.deco-radar::after {
  inset: 90rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.08);
}

/* 飞行轨迹虚线（静态） */
.deco-trail {
  position: absolute;
  left: 20px;
  bottom: 30px;
  width: 100px;
  height: 1px;
  background: repeating-linear-gradient(
    90deg,
    rgba(255, 255, 255, 0.5) 0 8px,
    transparent 8px 14px
  );
  transform: rotate(-15deg);
  opacity: 0.5;
}

/* ===== 顶部导航（对齐培训详情页：白圆钮返回 + 居中标题 + 右侧分享） ===== */
.detail-nav {
  position: relative;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-left: 24rpx;
  padding-right: 24rpx;
  box-sizing: content-box;
  background: #F5F8FC;
}

.detail-nav-back {
  width: 60rpx;
  height: 60rpx;
  flex: 0 0 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #ffffff;
  box-shadow: 0 6rpx 16rpx rgba(31, 89, 169, 0.13);
}
.detail-nav-back--press { transform: scale(0.94); opacity: 0.86; }
.detail-nav-back text { margin-top: -4rpx; color: #1A3353; font-size: 42rpx; line-height: 1; }

.detail-nav-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  max-width: 56%;
  display: block;
  color: #17212B;
  font-size: 34rpx;
  font-weight: 700;
  text-align: center;
  line-height: 88rpx;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 右侧分享（button 去默认样式） */
.share-btn {
  height: 60rpx;
  margin: 0;
  padding: 0 8rpx;
  line-height: 60rpx;
  text-align: center;
  background: transparent;
  border: none;
  font-size: 24rpx;
}
.share-btn::after { border: none; }
.share-text { color: #0A66C2; font-weight: 600; }

/* 状态徽章（对齐培训页：Hero 左上，白底胶囊 + 彩色文字） */
.status-badge {
  position: absolute;
  top: 18rpx;
  left: 18rpx;
  padding: 7rpx 16rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: none;
  z-index: 4;
  font-size: 20rpx;
  font-weight: 650;
  color: #0A66C2;
}

.badge--enrolling { background: #E9F7F0 !important; color: #0B6B41; }
.badge--ongoing { background: #EAF3FB !important; color: #0A66C2; }
.badge--closed { background: #EEF1F4 !important; color: #5D6B82; }

/* ===== Hero 底部信息区（对齐培训页 hero-bottom：标题 + 标签列贴底） ===== */
.hero-bottom {
  position: absolute;
  left: 24rpx;
  right: 24rpx;
  bottom: 24rpx;
  z-index: 3;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}

.hero-org {
  display: block;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.78);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hero-meta-row { display: flex; align-items: center; gap: 8rpx; }
.hero-meta-text { font-size: 22rpx; color: rgba(255, 255, 255, 0.85); }

.meta-ico { width: 26rpx; height: 26rpx; flex-shrink: 0; position: relative; }
.meta-ico--cal { border: 2rpx solid rgba(255, 255, 255, 0.9); border-radius: 4rpx; box-sizing: border-box; }
.meta-ico--cal::before, .meta-ico--cal::after { content: ''; position: absolute; top: 4rpx; width: 2rpx; height: 5rpx; background: rgba(255, 255, 255, 0.9); }
.meta-ico--cal::before { left: 6rpx; }
.meta-ico--cal::after { right: 6rpx; }
.meta-ico--cal .cal-top { position: absolute; left: 3rpx; right: 3rpx; top: 8rpx; height: 2rpx; background: rgba(255, 255, 255, 0.9); }
.meta-ico--cal .cal-line { position: absolute; left: 5rpx; right: 5rpx; height: 2rpx; background: rgba(255, 255, 255, 0.9); opacity: 0.6; }
.meta-ico--cal .cal-line.l1 { top: 14rpx; }
.meta-ico--cal .cal-line.l2 { top: 18rpx; }
.meta-ico--cal .cal-line.l3 { top: 22rpx; }

.hero-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.35;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.32);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.hero-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }

.hero-tag {
  padding: 5rpx 12rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
  border: 1rpx solid rgba(255, 255, 255, 0.3);
}

/* ===== 内容区：对齐培训详情页，直接在页面底色上排布卡片 ===== */
.content {
  position: relative;
  background: transparent;
  border-radius: 0;
  margin-top: 8rpx;
  padding: 0;
  box-shadow: none;
}

/* ===== 信息时间轴 ===== */
.info-timeline {
  margin: 20rpx 24rpx 0;
  padding: 8rpx 24rpx;
  animation: cardIn .22s ease-out backwards;
  animation-delay: 80ms;
}

.info-timeline::before {
  content: '';
  position: absolute;
  left: 32rpx;
  top: 88rpx;
  bottom: 16rpx;
  width: 2rpx;
  background: #0A66C2;
  opacity: 0.35;
}

/* 标题行：赛事信息 + 状态 pill（对齐培训详情页 course-info-head） */
.info-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14rpx;
  padding: 6rpx 0 10rpx;
}
.info-head-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.info-head-status { flex-shrink: 0; padding: 5rpx 14rpx; border-radius: 999rpx; font-size: 20rpx; font-weight: 650; background: #EEF1F4; color: #5D6B82; }

.tl-item {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
  padding: 16rpx 0;
  position: relative;
}

.tl-dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  margin-top: 10rpx;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
  background: #0A66C2;
  box-shadow: 0 0 0 6rpx rgba(10, 102, 194, 0.15);
}

.tl-dot--deadline { background: #EF4444; box-shadow: 0 0 0 6rpx rgba(239, 68, 68, 0.15); }

.tl-content { flex: 1; }

.tl-label { font-size: 22rpx; color: #7A8798; display: block; margin-bottom: 4rpx; }
.tl-label--deadline { color: #EF4444; font-weight: 500; }

.tl-value { font-size: 28rpx; color: #17212B; font-weight: 600; display: block; line-height: 1.35; }
.tl-value--deadline { color: #EF4444; font-weight: 600; }

.tl-countdown {
  display: inline-block;
  margin-top: 6rpx;
  padding: 4rpx 12rpx;
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
  font-size: 20rpx;
  font-weight: 600;
  border-radius: 999rpx;
}

/* ===== 章节（对齐培训详情页：30rpx 粗体 + 左侧 6rpx 深蓝条） ===== */
.section-block { margin: 28rpx 24rpx 0; animation: cardIn .22s ease-out backwards; }
.section-block:nth-of-type(1) { animation-delay: 100ms; }
.section-block:nth-of-type(2) { animation-delay: 120ms; }
.section-block:nth-of-type(3) { animation-delay: 140ms; }
.section-block:nth-of-type(4) { animation-delay: 160ms; }
.section-block:nth-of-type(5) { animation-delay: 180ms; }

.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.3;
  margin-bottom: 14rpx;
}

/* 简介 */
.intro-text {
  font-size: 26rpx;
  color: #344054;
  line-height: 1.8;
  white-space: pre-line;
  padding: 24rpx;
}

/* ===== 报名条件 ===== */
.requirements-card { padding: 4rpx 24rpx; }

.req-item {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #F0F1F3;
}

.req-item:last-child { border-bottom: none; }

.req-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 14rpx;
  background: linear-gradient(135deg, #0A66C2, #2889DD);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4rpx 12rpx rgba(10, 102, 194, 0.22);
}
.req-icon::before {
  content: '';
  width: 26rpx;
  height: 26rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23ffffff' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M5 12.5l4.5 4.5L19 7.5'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

.req-body { flex: 1; }
.req-name { font-size: 26rpx; font-weight: 600; color: #17212B; display: block; margin-bottom: 4rpx; }
.req-desc { font-size: 22rpx; color: #667085; line-height: 1.5; }

.req-badge {
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.req-badge--must { background: #FDECEC; color: #B42318; }
.req-badge--advise { background: #EAF3FB; color: #0A66C2; }
.req-badge--level { background: #EEF2F7; color: #5D6B82; }

.req-empty {
  padding: 28rpx 8rpx;
  font-size: 24rpx;
  color: #667085;
  text-align: center;
}

/* ===== 参赛项目（无左缘彩条） ===== */
.event-list { display: flex; flex-direction: column; gap: 16rpx; }

.event-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
}

/* 热门项（首项）：淡橙强调，突出推荐 */
.event-item--hot { border-color: #F5D3B3; background: linear-gradient(180deg, #FFF7EF 0%, #ffffff 100%); }

.event-info { min-width: 0; }
.event-name-row { display: flex; align-items: center; gap: 10rpx; }
.event-name { font-size: 28rpx; font-weight: 600; color: #17212B; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-meta { font-size: 22rpx; color: #667085; margin-top: 4rpx; display: block; }
.event-price { font-size: 32rpx; font-weight: 700; color: #E96012; flex-shrink: 0; }
.event-price--pending { font-size: 22rpx; font-weight: 500; color: #98A2B3; }

.event-empty {
  padding: 28rpx 8rpx;
  font-size: 24rpx;
  color: #667085;
  text-align: center;
}

.hot-badge {
  padding: 2rpx 12rpx;
  background: #FFF4EC;
  color: #E96012;
  font-size: 20rpx;
  font-weight: 600;
  border-radius: 999rpx;
}

/* ===== 奖项 ===== */
.prize-row { display: flex; gap: 16rpx; }

.prize-card {
  flex: 1;
  padding: 24rpx 12rpx;
  border-radius: 16px;
  text-align: center;
  border: 1rpx solid #EEF1F4;
  background: #fff;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
}

.prize-card--gold { background: linear-gradient(135deg, #FFF9C4, #FFE082); }
.prize-card--silver { background: linear-gradient(135deg, #F5F5F5, #E0E0E0); }
.prize-card--bronze { background: linear-gradient(135deg, #FFF3E0, #FFCC80); }

.prize-medal {
  width: 56rpx;
  height: 56rpx;
  margin: 0 auto 8rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.prize-medal::before {
  content: '';
  width: 30rpx;
  height: 30rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23ffffff' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M7 4h10v5a5 5 0 0 1-10 0z'/%3E%3Cpath d='M7 5H4a2.5 2.5 0 0 0 0 5h3M17 5h3a2.5 2.5 0 0 1 0 5h-3'/%3E%3Cpath d='M12 14v3M8.5 17h7M8.5 20h7'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

.prize-card--gold .prize-medal { background: #92400E; }
.prize-card--silver .prize-medal { background: #4B5563; }
.prize-card--bronze .prize-medal { background: #9A3412; }

.prize-level { font-size: 24rpx; font-weight: 600; color: #17212B; display: block; margin-bottom: 4rpx; }

.prize-amount-row { display: flex; align-items: baseline; justify-content: center; }
.prize-symbol { font-size: 22rpx; color: #344054; font-weight: 700; }
.prize-amount { font-size: 30rpx; font-weight: 800; color: #344054; }

/* ===== 主办单位 ===== */
.organizer-row {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
}

.org-avatar {
  width: 88rpx;
  height: 88rpx;
  background: linear-gradient(135deg, #0A66C2, #2889DD);
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4rpx 12rpx rgba(10, 102, 194, 0.22);
}
.org-avatar::before {
  content: '';
  width: 40rpx;
  height: 40rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23ffffff' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M3 21h18M5 21V8l7-4 7 4v13M9 21v-4h6v4'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

.org-info { flex: 1; min-width: 0; }
.org-name { font-size: 28rpx; font-weight: 600; color: #17212B; display: block; margin-bottom: 4rpx; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.org-sub { font-size: 22rpx; color: #98A2B3; }

/* ===== 底部 CTA（对齐培训详情页：毛白底 + 上投影 + 大价格 + 大按钮 76rpx 高） ===== */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid #E8EDF3;
  box-shadow: 0 -6rpx 18rpx rgba(16, 24, 40, 0.06);
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
  z-index: 50;
}

.fee-label { font-size: 20rpx; color: #667085; display: block; margin-bottom: 4rpx; }

.fee-price { display: flex; align-items: baseline; }
.fee-symbol { font-size: 22rpx; color: #E96012; font-weight: 700; }
.fee-value { font-size: 44rpx; font-weight: 700; color: #E96012; line-height: 1; }
.fee-value--pending { font-size: 26rpx; font-weight: 600; color: #98A2B3; line-height: 1.4; }
.fee-unit { font-size: 20rpx; color: #667085; margin-left: 4rpx; }

.free-badge {
  padding: 6rpx 20rpx;
  background: #E9F7F0;
  border: 1rpx solid #C9EEDC;
  color: #0B6B41;
  font-size: 26rpx;
  font-weight: 600;
  border-radius: 999rpx;
}

.bottom-actions { display: flex; gap: 12rpx; flex: 1; justify-content: flex-end; }

.btn-primary {
  height: 76rpx;
  padding: 0 40rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28);
  transition: transform 300ms var(--ease), opacity .15s ease;
}

.btn-primary:active { background: #0759AA; }

.btn-primary.disabled {
  background: #C9CDD4;
  color: #ffffff;
  box-shadow: none;
  pointer-events: none;
}

.bottom-spacer { height: 20rpx; }

/* ===== 动效 ===== */
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.press-feedback { transform: scale(0.98); opacity: 0.92; }

/* ===== 减弱动效（无障碍）：装饰动画全关，保留淡入 ===== */
.page.no-motion .hero-bottom,
.page.no-motion .card,
.page.no-motion .section-block { animation: none; }
</style>
