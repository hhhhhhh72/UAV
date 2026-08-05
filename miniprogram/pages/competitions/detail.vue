<template>
  <view class="page">
    <u-nav-bar title="赛事详情" show-back @back="goBack" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="赛事不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- ① 决策信息：状态 + 标题 + 标签 -->
        <view class="head-card">
          <view class="head-top">
            <text class="chip" :class="statusClass(detail.status)">{{ statusText[detail.status] || '报名中' }}</text>
          </view>
          <text class="head-title">{{ detail.title || detail.name || '未知赛事' }}</text>
          <view v-if="compTags(detail).length > 0" class="head-tags">
            <text v-for="tag in compTags(detail)" :key="tag" class="chip chip--tag">{{ tag }}</text>
          </view>
        </view>

        <!-- ② 关键信息 -->
        <view class="info-card">
          <view class="info-row">
            <text class="info-label">比赛时间</text>
            <text class="info-value">{{ fmtDate(detail.start_date) }} - {{ fmtDate(detail.end_date) }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">比赛地点</text>
            <view class="info-loc">
              <u-icon name="location" size="26rpx" color="#0A66C2" />
              <text class="info-value">{{ detail.location || '待定' }}</text>
            </view>
          </view>
          <view class="info-row info-row--deadline">
            <text class="info-label info-label--deadline">报名截止</text>
            <view class="info-deadline">
              <text class="info-value info-value--deadline">{{ fmtDate(detail.deadline || detail.enroll_deadline) }}</text>
              <text v-if="!isClosed(detail) && countdownText(detail)" class="countdown-chip">{{ countdownText(detail) }}</text>
            </view>
          </view>
        </view>

        <!-- ③ 赛事简介 -->
        <view v-if="detail.intro || detail.description" class="section">
          <text class="section-title">赛事简介</text>
          <view class="section-card">
            <text class="intro-text">{{ detail.intro || detail.description }}</text>
          </view>
        </view>

        <!-- ④ 报名条件 -->
        <view v-if="requirements(detail).length > 0" class="section">
          <text class="section-title">报名条件</text>
          <view class="section-card req-list">
            <view v-for="(req, i) in requirements(detail)" :key="req.name" class="req-item">
              <view class="req-icon" :class="'req-icon--' + (i % 4)">
                <text class="req-icon-text">{{ req.icon }}</text>
              </view>
              <view class="req-body">
                <text class="req-name">{{ req.name }}</text>
                <text class="req-desc">{{ req.desc }}</text>
              </view>
              <text class="req-badge" :class="req.level === '必满足' ? 'req-badge--must' : 'req-badge--advise'">{{ req.level }}</text>
            </view>
          </view>
        </view>

        <!-- ⑤ 参赛项目 -->
        <view v-if="eventList(detail).length > 0" class="section">
          <text class="section-title">参赛项目</text>
          <view class="event-list">
            <view
              v-for="(ev, i) in eventList(detail)"
              :key="ev.name"
              class="event-item"
              :style="{ borderLeftColor: eventColor(i) }"
            >
              <view class="event-info">
                <text class="event-name">{{ ev.name }}</text>
                <text v-if="eventMeta(ev)" class="event-meta">{{ eventMeta(ev) }}</text>
              </view>
              <text v-if="ev.fee != null" class="event-price">¥{{ Number(ev.fee).toLocaleString() }}</text>
            </view>
          </view>
        </view>

        <!-- ⑥ 奖项设置 -->
        <view v-if="prizes(detail).length > 0" class="section">
          <text class="section-title">奖项设置</text>
          <view class="prize-row">
            <view v-for="p in prizes(detail)" :key="p.level" class="prize-card" :class="'prize-card--' + p.metal">
              <view class="prize-medal">{{ p.medal }}</view>
              <text class="prize-level">{{ p.level }}</text>
              <view class="prize-amount-row">
                <text class="prize-symbol">¥</text>
                <text class="prize-amount">{{ Number(p.amount).toLocaleString() }}</text>
              </view>
            </view>
          </view>
        </view>

        <!-- ⑦ 主办单位 -->
        <view class="section">
          <text class="section-title">主办单位</text>
          <view class="section-card org-row">
            <view class="org-avatar">{{ orgInitial(detail) }}</view>
            <view class="org-info">
              <text class="org-name">{{ detail.organizer || detail.sponsor || '待定' }}</text>
              <text v-if="detail.organizer_sub" class="org-sub">{{ detail.organizer_sub }}</text>
            </view>
            <u-icon name="arrow" size="26rpx" color="#98A2B3" />
          </view>
        </view>

        <!-- ⑧ 底部 CTA -->
        <view class="bottom-bar">
          <view class="bottom-fee">
            <text class="fee-label">报名费</text>
            <view class="fee-price">
              <text class="fee-symbol">¥</text>
              <text class="fee-value">{{ compMinFee(detail) }}</text>
              <text class="fee-unit">起/人</text>
            </view>
          </view>
          <view class="bottom-actions">
            <view class="btn-outline" @click="handleConsult">咨询</view>
            <view
              class="btn-primary"
              :class="{ 'btn-primary--disabled': isClosed(detail) }"
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
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

const statusText = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已结束', full: '已满额' }

function isClosed(item) {
  return item.status === 'closed' || item.status === 'full'
}

function statusClass(status) {
  if (status === 'ongoing') return 'chip--ongoing'
  if (status === 'closed' || status === 'full') return 'chip--closed'
  return 'chip--enrolling'
}

function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (item.category) return [item.category]
  return []
}

/* 仅展示接口真实返回的内容，无本地假数据 */
function requirements(item) {
  return Array.isArray(item.requirements) ? item.requirements : []
}

function eventList(item) {
  return Array.isArray(item.events) ? item.events : []
}

function eventColor(i) {
  return ['#0A66C2', '#E96012', '#168A55'][i % 3]
}

function prizes(item) {
  return Array.isArray(item.prizes) ? item.prizes : []
}

function eventMeta(ev) {
  var parts = []
  if (ev.type) parts.push(ev.type)
  if (ev.format) parts.push(ev.format)
  return parts.join(' · ')
}

function compMinFee(item) {
  if (item.minFee != null) return item.minFee
  var evts = eventList(item)
  if (evts.length > 0) return Math.min.apply(null, evts.map(function (e) { return e.fee != null ? e.fee : 0 }))
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  return 0
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
  detail.value = null
  try {
    var res = await request({ url: '/api/v1/competitions' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { detail.value = items[i]; break }
    }
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goBack() { uni.navigateBack({ delta: 1 }) }

function goRegister() {
  if (detail.value && isClosed(detail.value)) return
  uni.navigateTo({ url: '/pages/competitions/register?id=' + encodeURIComponent(id.value) })
}

function handleConsult() {
  uni.showToast({ title: '咨询功能开发中', icon: 'none' })
}

onLoad(function (options) {
  id.value = options.id || ''
  loadDetail()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ① 决策信息卡 */
.head-card {
  margin: 20rpx 24rpx 0;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 28rpx;
}

.head-top { margin-bottom: 12rpx; }

.head-title {
  font-size: 38rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.35;
  display: block;
  margin-bottom: 16rpx;
}

.head-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }

/* 状态徽章（4px 圆角，非胶囊） */
.chip {
  display: inline-block;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 500;
  line-height: 1.6;
}

.chip--enrolling { background: #FFF0E6; color: #E96012; }
.chip--ongoing { background: #EAF3FB; color: #0A66C2; }
.chip--closed { background: #F4F6F8; color: #667085; }

.chip--tag { font-weight: 400; background: #EAF3FB; color: #0A66C2; }

/* ② 关键信息卡 */
.info-card {
  margin: 20rpx 24rpx 0;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 8rpx 28rpx;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  padding: 22rpx 0;
  border-bottom: 1px solid #F4F6F8;
}

.info-row:last-child { border-bottom: none; }

.info-label { font-size: 24rpx; color: #667085; flex-shrink: 0; }

.info-value {
  font-size: 28rpx;
  font-weight: 500;
  color: #17212B;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-loc { display: flex; align-items: center; gap: 8rpx; min-width: 0; }
.info-loc .info-value { flex: 1; }

.info-row--deadline { align-items: flex-start; }
.info-label--deadline { color: #E96012; font-weight: 500; padding-top: 6rpx; }
.info-deadline { display: flex; flex-direction: column; align-items: flex-end; gap: 8rpx; }
.info-value--deadline { color: #E96012; font-weight: 600; }

.countdown-chip {
  display: inline-block;
  padding: 4rpx 12rpx;
  background: #FFF0E6;
  color: #E96012;
  font-size: 22rpx;
  font-weight: 600;
  border-radius: 8rpx;
}

/* ③ 章节 */
.section { margin: 32rpx 24rpx 0; }

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #17212B;
  padding-left: 16rpx;
  border-left: 6rpx solid #0A66C2;
  line-height: 1.3;
  margin-bottom: 16rpx;
  display: block;
}

.section-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
}

.intro-text {
  font-size: 27rpx;
  color: #344054;
  line-height: 1.8;
  white-space: pre-line;
}

/* ④ 报名条件 */
.req-list { padding: 8rpx 24rpx; }

.req-item {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
  padding: 20rpx 0;
  border-bottom: 1px solid #F4F6F8;
}

.req-item:last-child { border-bottom: none; }

.req-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.req-icon--0 { background: #EAF3FB; }
.req-icon--1 { background: #FFF0E6; }
.req-icon--2 { background: #E9F7F0; }
.req-icon--3 { background: #F6F4FF; }

.req-icon--0 .req-icon-text { color: #0A66C2; }
.req-icon--1 .req-icon-text { color: #E96012; }
.req-icon--2 .req-icon-text { color: #168A55; }
.req-icon--3 .req-icon-text { color: #6E56CF; }

.req-icon-text { font-size: 28rpx; font-weight: 600; }

.req-body { flex: 1; min-width: 0; }

.req-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #17212B;
  display: block;
  margin-bottom: 4rpx;
}

.req-desc { font-size: 24rpx; color: #667085; line-height: 1.5; display: block; }

.req-badge {
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 500;
  flex-shrink: 0;
  margin-top: 8rpx;
}

.req-badge--must { background: #FFF0E6; color: #E96012; }
.req-badge--advise { background: #EAF3FB; color: #0A66C2; }

/* ⑤ 参赛项目 */
.event-list { display: flex; flex-direction: column; gap: 16rpx; }

.event-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
  padding: 24rpx;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-left: 6rpx solid #0A66C2;
  border-radius: 12rpx;
}

.event-info { flex: 1; min-width: 0; }

.event-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #17212B;
  display: block;
  margin-bottom: 4rpx;
}

.event-meta { font-size: 23rpx; color: #98A2B3; display: block; }

.event-price { font-size: 30rpx; font-weight: 700; color: #E96012; flex-shrink: 0; }

/* ⑥ 奖项 */
.prize-row { display: flex; gap: 16rpx; }

.prize-card {
  flex: 1;
  padding: 24rpx 12rpx;
  border: 1px solid #EEF1F4;
  border-radius: 12rpx;
  text-align: center;
}

.prize-card--gold { background: #FFF0E6; }
.prize-card--silver { background: #F4F6F8; }
.prize-card--bronze { background: #E9F7F0; }

.prize-medal {
  width: 56rpx;
  height: 56rpx;
  margin: 0 auto 10rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26rpx;
  font-weight: 700;
}

.prize-card--gold .prize-medal { background: #FFF0E6; color: #E96012; }
.prize-card--silver .prize-medal { background: #F4F6F8; color: #667085; }
.prize-card--bronze .prize-medal { background: #E9F7F0; color: #168A55; }

.prize-level { font-size: 26rpx; font-weight: 500; color: #17212B; display: block; margin-bottom: 6rpx; }

.prize-amount-row { display: flex; align-items: baseline; justify-content: center; }
.prize-symbol { font-size: 22rpx; color: #E96012; font-weight: 700; }
.prize-amount { font-size: 30rpx; font-weight: 700; color: #E96012; }

/* ⑦ 主办单位 */
.org-row { display: flex; align-items: center; gap: 20rpx; }

.org-avatar {
  width: 88rpx;
  height: 88rpx;
  background: #EAF3FB;
  color: #0A66C2;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.org-info { flex: 1; min-width: 0; }

.org-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #17212B;
  display: block;
  margin-bottom: 4rpx;
}

.org-sub { font-size: 23rpx; color: #98A2B3; display: block; }

/* ⑧ 底部 CTA */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #ffffff;
  border-top: 1px solid #EEF1F4;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 10;
}

.fee-label { font-size: 22rpx; color: #667085; display: block; margin-bottom: 4rpx; }

.fee-price { display: flex; align-items: baseline; }
.fee-symbol { font-size: 24rpx; color: #E96012; font-weight: 700; }
.fee-value { font-size: 40rpx; font-weight: 700; color: #E96012; line-height: 1; }
.fee-unit { font-size: 22rpx; color: #98A2B3; margin-left: 4rpx; }

.bottom-actions { display: flex; gap: 16rpx; }

.btn-outline {
  height: 76rpx;
  padding: 0 32rpx;
  border: 1px solid #0A66C2;
  border-radius: 12rpx;
  color: #0A66C2;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}

.btn-outline:active { background: #EAF3FB; }

.btn-primary {
  height: 76rpx;
  padding: 0 36rpx;
  background: #0A66C2;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 500;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 160ms ease;
}

.btn-primary:active { opacity: 0.85; }

.btn-primary--disabled { background: #98A2B3; }

.bottom-spacer { height: 140rpx; }
</style>
