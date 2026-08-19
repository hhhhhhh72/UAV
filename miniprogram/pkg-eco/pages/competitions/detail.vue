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
        <!-- ① Hero 区（海报优先，装饰仅在无图时兜底） -->
        <view class="hero" :style="{ paddingTop: (statusBarHeight + 10) + 'px' }">
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

          <!-- 三段渐变蒙层：顶暗→中透→底暗 -->
          <view class="hero-mask-top" />
          <view class="hero-mask-bottom" />

          <!-- 导航层 -->
          <view class="hero-nav">
            <view class="back-btn" hover-class="press-feedback" :hover-stay-time="120" @click="goBack">
              <view class="back-arrow" />
            </view>
            <button class="share-btn" open-type="share" hover-class="press-feedback" :hover-stay-time="120">
              <text class="share-text">分享</text>
            </button>
          </view>

          <!-- 左下角赛事首字徽章（半嵌在图片底部） -->
          <view class="hero-emblem"><text class="emblem-char">{{ emblemChar(detail) }}</text></view>

          <!-- 内容层：徽章 + 标题 + 标签 -->
          <view class="hero-content">
            <view class="status-badge" :class="statusClass(derivedStatus(detail))">{{ statusTextOf(detail) }}</view>
            <text class="hero-title">{{ detail.title || detail.name || '未知赛事' }}</text>
            <view class="hero-tags">
              <text v-for="tag in compTags(detail)" :key="tag" class="hero-tag">{{ tag }}</text>
            </view>
          </view>
        </view>

        <!-- ② 基本信息：时间轴 -->
        <view class="card info-timeline">
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
          <view class="card intro-text">{{ detail.intro || detail.description }}</view>
        </view>

        <!-- ④ 报名条件 -->
        <view class="section-block">
          <view class="section-title">报名条件</view>
          <view class="card requirements-card">
            <view v-if="requirements(detail).length === 0" class="req-empty">以主办方公布为准</view>
            <view v-for="req in requirements(detail)" :key="req.name" class="req-item">
              <view class="req-icon">
                <text class="req-icon-text">{{ req.icon }}</text>
              </view>
              <view class="req-body">
                <text class="req-name">{{ req.name }}</text>
                <text class="req-desc">{{ req.desc }}</text>
              </view>
              <view class="req-badge" :class="req.level === '必满足' ? 'req-badge--must' : 'req-badge--advise'">{{ req.level }}</view>
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
              <view class="prize-medal">{{ p.medal }}</view>
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
            <view class="org-avatar">{{ orgInitial(detail) }}</view>
            <view class="org-info">
              <text class="org-name">{{ detail.organizer || detail.sponsor || '待定' }}</text>
              <text class="org-sub">{{ detail.organizer_sub || '主办单位' }}</text>
            </view>
          </view>
        </view>

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
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onShareAppMessage } from '@dcloudio/uni-app'
import { useReduceMotion } from '../../../utils/motion'
import { request } from '../../../utils/request'
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
function emblemChar(item) {
  const t = item.title || ''
  if (t.indexOf('FPV') >= 0 || t.indexOf('竞速') >= 0) return '竞'
  if (t.indexOf('创新') >= 0 || t.indexOf('应用') >= 0) return '创'
  if (t.indexOf('青少年') >= 0) return '青'
  if (t.indexOf('国际') >= 0) return '际'
  if (t.indexOf('全国') >= 0) return '国'
  if (t.indexOf('贵州') >= 0) return '贵'
  return '赛'
}

function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
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

function eventList(item) {
  if (Array.isArray(item.events) && item.events.length > 0) return item.events
  return []
}

function prizes(item) {
  if (Array.isArray(item.prizes) && item.prizes.length > 0) return item.prizes
  return []
}

function compMinFee(item) {
  if (item.min_fee != null) return item.min_fee
  if (item.fee != null) return item.fee
  var evts = eventList(item)
  if (evts.length > 0) {
    var fees = evts.map(function (e) { return e.fee }).filter(function (f) { return f != null })
    if (fees.length > 0) return Math.min.apply(null, fees)
  }
  return null
}

function orgInitial(item) {
  var name = item.organizer || item.sponsor || '中'
  return name.charAt(0)
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

function goBack() { uni.navigateBack({ delta: 1 }) }

function goRegister() {
  if (detail.value && isClosed(detail.value)) return
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
  min-height: 100vh;
  background: #fff;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== 通用卡片：白上白（对齐设计语言 token） ===== */
.card {
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}

/* ================================================================= */
/* ① Hero 区                                                          */
/* ================================================================= */
.hero {
  min-height: 240px;
  background: linear-gradient(135deg, #074D92 0%, #0A66C2 100%);
  position: relative;
  overflow: hidden;
  padding: 40px 16px 20px;
}

/* 真实海报图 */
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity .25s ease-out;
}

/* 三段渐变蒙层：顶暗→中透→底暗 */
.hero-mask-top {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 45px;
  background: linear-gradient(180deg, rgba(10, 31, 68, 0.65) 0%, rgba(10, 31, 68, 0) 100%);
  pointer-events: none;
  z-index: 1;
}

.hero-mask-bottom {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 55px;
  background: linear-gradient(0deg, rgba(10, 31, 68, 0.85) 0%, rgba(10, 31, 68, 0) 100%);
  pointer-events: none;
  z-index: 1;
}

/* 左下角首字徽章：完整收纳在 Hero 内，不与下方卡片重叠 */
.hero-emblem {
  position: absolute;
  left: 16px;
  bottom: 10px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.emblem-char {
  font-size: 17px;
  font-weight: 700;
  color: #0A66C2;
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

.hero-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 2;
}

.back-btn {
  width: 44px;
  height: 44px;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* CSS 绘制返回箭头（左向 V 形） */
.back-arrow {
  width: 10px;
  height: 10px;
  border-left: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: rotate(45deg);
  margin-left: 4px;
}

/* 分享：文字标签按钮（button 去默认样式） */
.share-btn {
  width: 56px;
  height: 32px;
  margin: 0;
  padding: 0;
  line-height: 32px;
  text-align: center;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 999px;
  font-size: 12px;
}
.share-btn::after { border: none; }
.share-text { color: #fff; font-weight: 500; }

/* 状态徽章（扁平 tint，无脉冲） */
.status-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  z-index: 2;
  margin-bottom: 10px;
}

.badge--enrolling { background: #E9F7F0; color: #0B6B41; }
.badge--ongoing { background: #EAF3FB; color: #0A66C2; }
.badge--closed { background: #EEF1F4; color: #5D6B82; }

/* ===== Hero 内容层 ===== */
.hero-content {
  position: relative;
  z-index: 2;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}

.hero-title {
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.35;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 8px;
}

.hero-tags { display: flex; flex-wrap: wrap; gap: 6px; }

.hero-tag {
  padding: 3px 9px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.15);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

/* ===== 信息时间轴 ===== */
.info-timeline {
  margin: 20px 12px 0;
  padding: 4px 12px;
  animation: cardIn .22s ease-out backwards;
  animation-delay: 80ms;
}

.info-timeline::before {
  content: '';
  position: absolute;
  left: 17px;
  top: 12px;
  bottom: 12px;
  width: 2px;
  background: #0A66C2;
}

.tl-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 0;
  position: relative;
}

.tl-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-top: 8px;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
  background: #0A66C2;
  box-shadow: 0 0 0 3px rgba(10, 102, 194, 0.15);
}

.tl-dot--deadline { background: #EF4444; box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.15); }

.tl-content { flex: 1; }

.tl-label { font-size: 12px; color: #667085; display: block; margin-bottom: 2px; }
.tl-label--deadline { color: #EF4444; font-weight: 500; }

.tl-value { font-size: 15px; color: #17212B; font-weight: 500; display: block; }
.tl-value--deadline { color: #EF4444; font-weight: 600; }

.tl-countdown {
  display: inline-block;
  margin-top: 3px;
  padding: 1px 6px;
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
  font-size: 11px;
  font-weight: 600;
  border-radius: 999px;
}

/* ===== 章节 ===== */
.section-block { margin: 16px 12px 0; animation: cardIn .22s ease-out backwards; }
.section-block:nth-of-type(1) { animation-delay: 100ms; }
.section-block:nth-of-type(2) { animation-delay: 120ms; }
.section-block:nth-of-type(3) { animation-delay: 140ms; }
.section-block:nth-of-type(4) { animation-delay: 160ms; }
.section-block:nth-of-type(5) { animation-delay: 180ms; }

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.3;
  margin-bottom: 10px;
}

/* 简介 */
.intro-text {
  font-size: 14px;
  color: #344054;
  line-height: 1.8;
  white-space: pre-line;
  padding: 12px;
}

/* ===== 报名条件 ===== */
.requirements-card { padding: 2px 12px; }

.req-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid #F0F1F3;
}

.req-item:last-child { border-bottom: none; }

.req-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.req-icon-text { font-size: 14px; color: #ffffff; font-weight: 600; }

.req-body { flex: 1; }
.req-name { font-size: 14px; font-weight: 500; color: #17212B; display: block; margin-bottom: 2px; }
.req-desc { font-size: 12px; color: #667085; line-height: 1.5; }

.req-badge {
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.req-badge--must { background: #FDECEC; color: #B42318; }
.req-badge--advise { background: #EAF3FB; color: #0A66C2; }

.req-empty {
  padding: 16px 4px;
  font-size: 13px;
  color: #667085;
  text-align: center;
}

/* ===== 参赛项目（无左缘彩条） ===== */
.event-list { display: flex; flex-direction: column; gap: 8px; }

.event-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
}

.event-info { min-width: 0; }
.event-name-row { display: flex; align-items: center; gap: 6px; }
.event-name { font-size: 15px; font-weight: 500; color: #17212B; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-meta { font-size: 12px; color: #667085; margin-top: 2px; display: block; }
.event-price { font-size: 16px; font-weight: 700; color: #C2410C; flex-shrink: 0; }
.event-price--pending { font-size: 12px; font-weight: 500; color: #98A2B3; }

.event-empty {
  padding: 16px 4px;
  font-size: 13px;
  color: #667085;
  text-align: center;
}

.hot-badge {
  padding: 1px 6px;
  background: #FFF4EC;
  color: #E96012;
  font-size: 10px;
  font-weight: 600;
  border-radius: 999px;
}

/* ===== 奖项 ===== */
.prize-row { display: flex; gap: 8px; }

.prize-card {
  flex: 1;
  padding: 12px 6px;
  border-radius: 10px;
  text-align: center;
  border: 1px solid #E4E7EC;
  background: #fff;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}

.prize-card--gold { background: linear-gradient(135deg, #FFF9C4, #FFE082); }
.prize-card--silver { background: linear-gradient(135deg, #F5F5F5, #E0E0E0); }
.prize-card--bronze { background: linear-gradient(135deg, #FFF3E0, #FFCC80); }

.prize-medal {
  width: 28px;
  height: 28px;
  margin: 0 auto 4px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
}

.prize-card--gold .prize-medal { background: #92400E; }
.prize-card--silver .prize-medal { background: #4B5563; }
.prize-card--bronze .prize-medal { background: #9A3412; }

.prize-level { font-size: 13px; font-weight: 500; color: #17212B; display: block; margin-bottom: 2px; }

.prize-amount-row { display: flex; align-items: baseline; justify-content: center; }
.prize-symbol { font-size: 11px; color: #344054; font-weight: 700; }
.prize-amount { font-size: 15px; font-weight: 800; color: #344054; }

/* ===== 主办单位 ===== */
.organizer-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
}

.org-avatar {
  width: 48px;
  height: 48px;
  background: #0A66C2;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  flex-shrink: 0;
}

.org-info { flex: 1; min-width: 0; }
.org-name { font-size: 14px; font-weight: 500; color: #17212B; display: block; margin-bottom: 2px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.org-sub { font-size: 12px; color: #98A2B3; }

/* ===== 底部 CTA ===== */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #ffffff;
  border-top: 1px solid #EEF1F4;
  box-shadow: 0 -2px 12px rgba(16, 24, 40, 0.04);
  padding: 10px 16px calc(10px + env(safe-area-inset-bottom));
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 10;
}

.fee-label { font-size: 11px; color: #98A2B3; display: block; margin-bottom: 2px; }

.fee-price { display: flex; align-items: baseline; }
.fee-symbol { font-size: 12px; color: #C2410C; font-weight: 700; }
.fee-value { font-size: 22px; font-weight: 800; color: #C2410C; line-height: 1; }
.fee-value--pending { font-size: 14px; font-weight: 600; color: #98A2B3; line-height: 1.4; }
.fee-unit { font-size: 11px; color: #98A2B3; margin-left: 2px; }

.free-badge {
  padding: 2px 10px;
  background: #E9F7F0;
  border: 1px solid #C9EEDC;
  color: #0B6B41;
  font-size: 13px;
  font-weight: 600;
  border-radius: 999px;
}

.bottom-actions { display: flex; gap: 10px; }

.btn-primary {
  padding: 10px 24px;
  border-radius: 999px;
  background: #0A66C2;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.24);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}

.btn-primary.disabled {
  background: #EEF1F4;
  color: #667085;
  box-shadow: none;
  pointer-events: none;
}

.bottom-spacer { height: 140rpx; }

/* ===== 动效 ===== */
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.press-feedback { transform: scale(0.98); opacity: 0.92; }

/* ===== 减弱动效（无障碍）：装饰动画全关，保留淡入 ===== */
.page.no-motion .hero-content,
.page.no-motion .card,
.page.no-motion .section-block { animation: none; }
</style>
