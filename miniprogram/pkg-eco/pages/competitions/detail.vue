<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="赛事不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- ① Hero 区（真实海报图 + 三段蒙层 + 兜底） -->
        <view class="hero">
          <!-- 真实赛事海报图（有则覆盖渐变） -->
          <image
            v-if="heroPoster(detail)"
            :src="heroPoster(detail)"
            mode="aspectFill"
            class="hero-img"
            lazy-load
            :style="{ opacity: imgLoaded ? 1 : 0 }"
            @load="onHeroImgLoad"
          />

          <!-- 装饰层：作为图片加载失败/缺失时的兜底背景 -->
          <view class="hero-deco">
            <view class="deco-grid" />
            <view class="deco-radar" />
            <view class="deco-radar-scanner" />
            <view class="deco-trail" />
            <view class="deco-star s1" />
            <view class="deco-star s2" />
            <view class="deco-star s3" />
            <view class="deco-star s4" />
          </view>

          <!-- 三段渐变蒙层：顶暗→中透→底暗 -->
          <view class="hero-mask-top" />
          <view class="hero-mask-bottom" />

          <!-- 导航层 -->
          <view class="hero-nav">
            <view class="back-btn" hover-class="press-feedback" :hover-stay-time="120" @click="goBack">
              <text class="back-icon">‹</text>
            </view>
            <view class="hero-actions">
              <view class="hero-action hero-action--fav" @click="handleFav"><text class="fav-icon">♥</text></view>
              <button class="hero-action hero-action--share" open-type="share" hover-class="press-feedback" :hover-stay-time="120"><text class="share-icon">↗</text></button>
            </view>
          </view>

          <!-- 左下角赛事奖杯校徽（半嵌在图片底部） -->
          <view class="hero-emblem"><text class="emblem-char">{{ emblemChar(detail) }}</text></view>

          <!-- 内容层：徽章 + 标题 + 标签 -->
          <view class="hero-content">
            <view class="status-badge" :class="statusClass(detail.status)">{{ statusText[detail.status] || '报名中' }}</view>
            <text class="hero-title">{{ detail.title || detail.name || '未知赛事' }}</text>
            <view class="hero-tags">
              <text v-for="tag in compTags(detail)" :key="tag" class="hero-tag">{{ tag }}</text>
            </view>
          </view>
        </view>

        <!-- ③ 基本信息：时间轴 -->
        <view class="info-timeline">
          <view class="tl-item">
            <view class="tl-dot tl-dot--solid" />
            <view class="tl-content">
              <text class="tl-label">比赛时间</text>
              <text class="tl-value">{{ fmtDate(detail.start_date) }} - {{ fmtDate(detail.end_date) }}</text>
            </view>
          </view>
          <view class="tl-item">
            <view class="tl-dot tl-dot--solid" />
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

        <!-- ④ 赛事简介 -->
        <view v-if="detail.intro || detail.description" class="section-block">
          <view class="section-title">赛事简介</view>
          <view class="intro-text">{{ detail.intro || detail.description }}</view>
        </view>

        <!-- ⑤ 报名条件 -->
        <view class="section-block">
          <view class="section-title">报名条件</view>
          <view class="requirements-card">
            <view v-if="requirements(detail).length === 0" class="req-empty">以主办方公布为准</view>
            <view v-for="(req, i) in requirements(detail)" :key="req.name" class="req-item">
              <view class="req-icon" :class="'req-icon--' + (i % 5)">
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

        <!-- ⑥ 参赛项目 -->
        <view class="section-block">
          <view class="section-title">参赛项目</view>
          <view class="event-list">
            <view v-if="eventList(detail).length === 0" class="event-empty">以主办方公布为准</view>
            <view
              v-for="(ev, i) in eventList(detail)"
              :key="ev.name"
              class="event-item"
              :class="{ 'event-item--hot': i === 0 }"
              :style="{ borderLeftColor: eventColor(i) }"
            >
              <view class="event-info">
                <view class="event-name-row">
                  <text class="event-name">{{ ev.name }}</text>
                  <view v-if="i === 0" class="hot-badge">热门</view>
                </view>
                <text class="event-meta">{{ ev.type }} · {{ ev.format }}</text>
              </view>
              <text v-if="ev.fee != null" class="event-price">¥{{ ev.fee.toLocaleString() }}</text>
              <text v-else class="event-price">费用待定</text>
            </view>
          </view>
        </view>

        <!-- ⑦ 奖项 -->
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

        <!-- ⑧ 主办单位 -->
        <view class="section-block">
          <view class="section-title">主办单位</view>
          <view class="organizer-row">
            <view class="org-avatar">{{ orgInitial(detail) }}</view>
            <view class="org-info">
              <text class="org-name">{{ detail.organizer || detail.sponsor || '待定' }}</text>
              <text class="org-sub">{{ detail.organizer_sub || '主办单位' }}</text>
            </view>
            <view class="org-arrow">›</view>
          </view>
        </view>

        <!-- ⑨ 底部 CTA -->
        <view class="bottom-bar">
          <view class="bottom-left">
            <text class="fee-label">报名费</text>
            <view class="fee-price">
              <template v-if="compMinFee(detail) != null">
                <text class="fee-symbol">¥</text>
                <text class="fee-value">{{ compMinFee(detail) }}</text>
                <text class="fee-unit">起/人</text>
              </template>
              <text v-else class="fee-value fee-value--pending">以主办方公布为准</text>
            </view>
          </view>
          <view class="bottom-actions">
            <view class="btn-outline" hover-class="press-feedback" :hover-stay-time="120" @click="handleConsult">咨询</view>
            <view class="btn-primary" :class="{ disabled: isClosed(detail) }" hover-class="press-feedback" :hover-stay-time="120" @click="goRegister">
              {{ isClosed(detail) ? '已截止' : '立即报名' }}
            </view>
          </view>
        </view>
        <view class="bottom-spacer" />
      </template>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
const imgLoaded = ref(false)

const statusText = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已结束', full: '已满额' }

function isClosed(item) {
  return item.status === 'closed' || item.status === 'full'
}

function statusClass(status) {
  if (status === 'ongoing') return 'badge--ongoing'
  if (status === 'closed' || status === 'full') return 'badge--closed'
  return 'badge--enrolling'
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

/** Hero 左下角奖杯/首字校徽 */
function emblemChar(item) {
  const t = item.title || ''
  if (t.indexOf('FPV') >= 0 || t.indexOf('竞速') >= 0) return '竞'
  if (t.indexOf('创新') >= 0 || t.indexOf('应用') >= 0) return '创'
  if (t.indexOf('青少年') >= 0) return '青'
  if (t.indexOf('国际') >= 0) return '际'
  if (t.indexOf('全国') >= 0) return '国'
  if (t.indexOf('贵州') >= 0) return '黔'
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

function eventColor(i) {
  return ['#0A66C2', '#8B5CF6', '#34c759'][i % 3]
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

function countdownText(item) {
  var d = item.deadline || item.enroll_deadline
  if (!d) return ''
  var days = deadlineDays(d)
  if (days == null) return ''
  if (days <= 0) return '已截止'
  return '剩余 ' + days + ' 天'
}

function deadlineDays(d) {
  var m = String(d).match(/(\d{4})[年.\-\/](\d{1,2})[月.\-\/](\d{1,2})/)
  if (!m) return null
  var target = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  var today = new Date()
  return Math.ceil((target - today) / 86400000)
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

function handleConsult() {
  uni.showToast({ title: '咨询功能开发中', icon: 'none' })
}

function handleFav() {
  // 收藏接口未就绪：不做假成功提示，保持原状态
  uni.showToast({ title: '收藏功能即将开放', icon: 'none' })
}

function handleShare() {
  // 分享按钮（open-type="share"）与右上角菜单均走 onShareAppMessage
  uni.showToast({ title: '请点击右上角菜单分享', icon: 'none' })
}

onShareAppMessage(function () {
  const item = detail.value || {}
  return {
    title: '赛事报名：' + (item.title || item.name || '无人机赛事'),
    path: '/pkg-eco/pages/competitions/detail?id=' + encodeURIComponent(id.value),
  }
})

onLoad(function (options) {
  id.value = options.id || ''
  loadDetail()
})
</script>

<style scoped>
.page {
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.25, 0.46, 0.45, 0.94);
  min-height: 100vh;
  background: linear-gradient(180deg, #f5f6f8 0%, #E8F2FC 100%);
  padding-bottom: env(safe-area-inset-bottom);
}

/* ================================================================= */
/* ① Hero 区                                                          */
/* ================================================================= */
.hero {
  height: 280px;
  background: linear-gradient(135deg, #074D92 0%, #0A66C2 100%);
  position: relative;
  overflow: hidden;
  padding: 88rpx 32rpx 40rpx;
}

/* 真实海报图 */
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity var(--anim-base) ease-out;
}

/* 三段渐变蒙层：顶暗→中透→底暗 */
.hero-mask-top {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 90rpx;
  background: linear-gradient(180deg, rgba(10, 31, 68, 0.65) 0%, rgba(10, 31, 68, 0) 100%);
  pointer-events: none;
  z-index: 1;
}

.hero-mask-bottom {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 110rpx;
  background: linear-gradient(0deg, rgba(10, 31, 68, 0.85) 0%, rgba(10, 31, 68, 0) 100%);
  pointer-events: none;
  z-index: 1;
}

/* 左下角赛事奖杯/首字校徽：半嵌在图片底部 */
.hero-emblem {
  position: absolute;
  left: 32rpx;
  bottom: -24rpx;
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: #ffffff;
  border: 2rpx solid rgba(255, 255, 255, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
}

.emblem-char {
  font-size: 34rpx;
  font-weight: 700;
  color: #0A66C2;
}

/* ===== 装饰层（纯 CSS） ===== */
.hero-deco { position: absolute; inset: 0; pointer-events: none; }

/* 网格点阵 */
.deco-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.12) 2rpx, transparent 2rpx);
  background-size: 40rpx 40rpx;
  opacity: 0.6;
}

/* 雷达同心圆 */
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

/* 雷达扫射弧线 */
.deco-radar-scanner {
  position: absolute;
  right: -80rpx;
  top: -80rpx;
  width: 300rpx;
  height: 300rpx;
  border-radius: 50%;
  background: conic-gradient(
    from 0deg,
    rgba(0, 229, 255, 0.35) 0deg,
    transparent 60deg
  );
  animation: radarRotate 4s linear infinite;
}

/* 飞行轨迹虚线 */
.deco-trail {
  position: absolute;
  left: 40rpx;
  bottom: 60rpx;
  width: 200rpx;
  height: 2rpx;
  background: repeating-linear-gradient(
    90deg,
    rgba(255, 255, 255, 0.5) 0 16rpx,
    transparent 16rpx 28rpx
  );
  transform: rotate(-15deg);
  opacity: 0.5;
}

/* 星点 */
.deco-star {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  animation: twinkle 2.5s ease-in-out infinite;
}
.s1 { left: 60rpx;  top: 80rpx;  width: 6rpx;  height: 6rpx;  animation-delay: 0s; }
.s2 { left: 200rpx; top: 120rpx; width: 8rpx;  height: 8rpx;  animation-delay: 0.6s; }
.s3 { left: 500rpx; top: 60rpx;  width: 5rpx;  height: 5rpx;  animation-delay: 1.2s; }
.s4 { left: 620rpx; top: 170rpx; width: 7rpx;  height: 7rpx;  animation-delay: 1.8s; }

.hero-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 2;
}

.back-btn {
  width: 88rpx;
  height: 88rpx;
  background: rgba(255, 255, 255, 0.15);
  border: 1rpx solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-icon { color: #ffffff; font-size: 44rpx; font-weight: 300; }

.hero-actions { display: flex; gap: 16rpx; }

.hero-action {
  width: 88rpx;
  height: 88rpx;
  background: rgba(255, 255, 255, 0.15);
  border: 1rpx solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fav-icon { color: #ffffff; font-size: 32rpx; }
.share-icon { color: #ffffff; font-size: 36rpx; font-weight: 300; }

/* 分享按钮（open-type="share"）沿用 hero-action 圆形玻璃样式，去掉 button 默认样式 */
.hero-action--share { padding: 0; margin: 0; line-height: 1; }
.hero-action--share::after { border: none; }

/* 状态徽章 */
.status-badge {
  display: inline-block;
  padding: 8rpx 20rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 600;
  color: #ffffff;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.2);
  z-index: 2;
  margin-bottom: 20rpx;
}

.badge--enrolling { background: linear-gradient(135deg, #F97316, #E96012); animation: badgePulse 2s ease-in-out infinite; }
.badge--ongoing { background: linear-gradient(135deg, #00E5FF, #0A66C2); }
.badge--closed { background: #64748B; }

/* ================================================================= */
/* ② Hero 内容层                                                       */
/* ================================================================= */
.hero-content {
  position: relative;
  z-index: 2;
  animation: pageIn var(--anim-slow) var(--ease-out) both;
}

.hero-title {
  font-size: 40rpx;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.3;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.3);
  display: block;
  margin-bottom: 16rpx;
}

.hero-tags { display: flex; flex-wrap: wrap; gap: 12rpx; }

.hero-tag {
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.15);
  color: #ffffff;
  border: 1rpx solid rgba(255, 255, 255, 0.3);
}

/* ================================================================= */
/* ③ 信息时间轴                                                       */
/* ================================================================= */
.info-timeline {
  margin: 24rpx 24rpx 0;
  background: #ffffff;
  border-radius: 16px;
  padding: 8rpx 24rpx;
  box-shadow: 0 4px 16px rgba(10, 31, 68, 0.06);
  position: relative;
}

.info-timeline::before {
  content: '';
  position: absolute;
  left: 34rpx;
  top: 24rpx;
  bottom: 24rpx;
  width: 2rpx;
  background: linear-gradient(180deg, #0A66C2, #00E5FF);
}

.tl-item {
  display: flex;
  align-items: flex-start;
  gap: 24rpx;
  padding: 20rpx 0;
  position: relative;
}

.tl-dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  margin-top: 8rpx;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.tl-dot--solid { background: #0A66C2; box-shadow: 0 0 0 6rpx rgba(10, 102, 194, 0.15); }
.tl-dot--deadline { background: #EF4444; box-shadow: 0 0 0 6rpx rgba(239, 68, 68, 0.15); }

.tl-content { flex: 1; }

.tl-label { font-size: 24rpx; color: #969799; display: block; margin-bottom: 4rpx; }
.tl-label--deadline { color: #EF4444; font-weight: 500; }

.tl-value { font-size: 30rpx; color: #17212B; font-weight: 500; display: block; }
.tl-value--deadline { color: #EF4444; font-weight: 600; }

.tl-countdown {
  display: inline-block;
  margin-top: 6rpx;
  padding: 2rpx 12rpx;
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
  font-size: 22rpx;
  font-weight: 600;
  border-radius: 999rpx;
}

/* ================================================================= */
/* 章节                                                               */
/* ================================================================= */
.section-block { margin: 36rpx 24rpx 0; animation: blockIn var(--anim-base) var(--ease-out) both; }
.section-block:nth-of-type(1) { animation-delay: 60ms; }
.section-block:nth-of-type(2) { animation-delay: 120ms; }
.section-block:nth-of-type(3) { animation-delay: 180ms; }
.section-block:nth-of-type(4) { animation-delay: 240ms; }
.section-block:nth-of-type(5) { animation-delay: 300ms; }

.section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  padding-left: 20rpx;
  border-left: 6rpx solid #0A66C2;
  line-height: 1.3;
  margin-bottom: 20rpx;
}

/* 简介 */
.intro-text {
  font-size: 28rpx;
  color: #17212B;
  line-height: 1.8;
  white-space: pre-line;
  margin-bottom: 8rpx;
  background: #ffffff;
  border-radius: 16px;
  padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(10, 31, 68, 0.04);
}

/* ================================================================= */
/* 报名条件                                                           */
/* ================================================================= */
.requirements-card {
  background: linear-gradient(180deg, #FFFBEB, #FEF3C7);
  border-radius: 16px;
  border: 1rpx solid #FDE68A;
  padding: 8rpx 24rpx;
}

.req-item {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #F5E6C8;
}

.req-item:last-child { border-bottom: none; }

.req-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.req-icon--0 { background: linear-gradient(135deg, #0A66C2, #0A66C2); }
.req-icon--1 { background: linear-gradient(135deg, #8B5CF6, #A78BFA); }
.req-icon--2 { background: linear-gradient(135deg, #34c759, #06B6D4); }
.req-icon--3 { background: linear-gradient(135deg, #22C55E, #4ADE80); }
.req-icon--4 { background: linear-gradient(135deg, #F97316, #E96012); }

.req-icon-text { font-size: 28rpx; color: #ffffff; font-weight: 600; }

.req-body { flex: 1; }
.req-name { font-size: 28rpx; font-weight: 500; color: #17212B; display: block; margin-bottom: 4rpx; }
.req-desc { font-size: 24rpx; color: #969799; line-height: 1.5; }

.req-badge {
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.req-badge--must { background: #EF4444; color: #ffffff; }
.req-badge--advise { background: #ffffff; color: #0A66C2; border: 1rpx solid #0A66C2; }

.req-empty {
  padding: 32rpx 8rpx;
  font-size: 26rpx;
  color: #969799;
  text-align: center;
}

/* ================================================================= */
/* 参赛项目                                                           */
/* ================================================================= */
.event-list { display: flex; flex-direction: column; gap: 16rpx; }

.event-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
  background: #ffffff;
  border-radius: 12px;
  border-left: 4rpx solid #0A66C2;
  box-shadow: 0 2rpx 12rpx rgba(10, 31, 68, 0.04);
}

.event-item--hot {
  box-shadow: 0 4rpx 16rpx rgba(255, 142, 60, 0.2);
}

.event-name-row { display: flex; align-items: center; gap: 8rpx; }
.event-name { font-size: 30rpx; font-weight: 500; color: #17212B; display: block; }
.event-meta { font-size: 24rpx; color: #969799; margin-top: 4rpx; display: block; }
.event-price { font-size: 34rpx; font-weight: 700; color: #E96012; }

.event-empty {
  padding: 32rpx 8rpx;
  font-size: 26rpx;
  color: #969799;
  text-align: center;
}

.hot-badge {
  padding: 2rpx 12rpx;
  background: linear-gradient(135deg, #F97316, #E96012);
  color: #ffffff;
  font-size: 20rpx;
  font-weight: 600;
  border-radius: 999rpx;
  animation: badgePulse 2s ease-in-out infinite;
}

/* ================================================================= */
/* 奖项                                                               */
/* ================================================================= */
.prize-row { display: flex; gap: 16rpx; }

.prize-card {
  flex: 1;
  padding: 24rpx 12rpx;
  border-radius: 16px;
  text-align: center;
}

.prize-card--gold { background: linear-gradient(135deg, #FFF9C4, #FFD54F); box-shadow: 0 4rpx 16rpx rgba(255, 193, 7, 0.3); }
.prize-card--silver { background: linear-gradient(135deg, #F5F5F5, #E0E0E0); box-shadow: 0 4rpx 16rpx rgba(158, 158, 158, 0.25); }
.prize-card--bronze { background: linear-gradient(135deg, #FFF3E0, #FFCC80); box-shadow: 0 4rpx 16rpx rgba(255, 152, 0, 0.25); }

.prize-medal {
  width: 56rpx;
  height: 56rpx;
  margin: 0 auto 8rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
}

.prize-card--gold .prize-medal { background: #FFB300; color: #fff; }
.prize-card--silver .prize-medal { background: #9E9E9E; color: #fff; }
.prize-card--bronze .prize-medal { background: #F57C00; color: #fff; }

.prize-level { font-size: 26rpx; font-weight: 500; color: #17212B; display: block; margin-bottom: 4rpx; }

.prize-amount-row { display: flex; align-items: baseline; justify-content: center; }
.prize-symbol { font-size: 22rpx; color: #17212B; font-weight: 700; }
.prize-amount { font-size: 30rpx; font-weight: 800; color: #17212B; }

/* ================================================================= */
/* 主办单位                                                           */
/* ================================================================= */
.organizer-row {
  display: flex;
  align-items: center;
  gap: 20rpx;
  background: #ffffff;
  border-radius: 16px;
  padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(10, 31, 68, 0.04);
}

.org-avatar {
  width: 96rpx;
  height: 96rpx;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 40rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.org-info { flex: 1; }
.org-name { font-size: 28rpx; font-weight: 500; color: #17212B; display: block; margin-bottom: 4rpx; }
.org-sub { font-size: 24rpx; color: #969799; }

.org-arrow { font-size: 36rpx; color: #98A2B3; flex-shrink: 0; }

/* ================================================================= */
/* 底部 CTA                                                           */
/* ================================================================= */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #ffffff;
  border-top: 1rpx solid rgba(10, 31, 68, 0.06);
  box-shadow: 0 -4rpx 16rpx rgba(0, 0, 0, 0.04);
  padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom));
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 10;
}

.fee-label { font-size: 22rpx; color: #969799; display: block; margin-bottom: 4rpx; }

.fee-price { display: flex; align-items: baseline; }
.fee-symbol { font-size: 24rpx; color: #E96012; font-weight: 700; }
.fee-value { font-size: 44rpx; font-weight: 800; color: #E96012; line-height: 1; }
.fee-value--pending { font-size: 28rpx; font-weight: 600; color: #969799; line-height: 1.4; }
.fee-unit { font-size: 22rpx; color: #969799; margin-left: 4rpx; }

.bottom-actions { display: flex; gap: 20rpx; }

.btn-outline {
  padding: 20rpx 36rpx;
  border-radius: 50rpx;
  border: 2rpx solid #0A66C2;
  color: #0A66C2;
  font-size: 28rpx;
  font-weight: 500;
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.btn-primary {
  padding: 20rpx 40rpx;
  border-radius: 50rpx;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 600;
  box-shadow: 0 4rpx 16rpx rgba(10, 102, 194, 0.3);
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.btn-primary.disabled { background: #CBD5E1; box-shadow: none; }
.bottom-spacer { height: 180rpx; }

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes blockIn {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes badgePulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(255, 142, 60, 0.4); }
  50% { box-shadow: 0 0 0 8rpx rgba(255, 142, 60, 0); }
}

@keyframes radarRotate {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.2; }
  50%      { opacity: 0.8; }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .hero-content, .section-block, .btn-primary, .btn-outline {
    animation: none !important;
    transition: none !important;
  }
}
</style>
