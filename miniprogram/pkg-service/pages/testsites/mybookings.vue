<template>
  <view class="ts-page" :class="{ 'no-motion': noMotion }">
    <u-nav-bar title="我的预约" show-back @back="goBack" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && bookings.length === 0"
      empty-text="还没有预约记录"
      :show-create="!loading && !errorMsg"
      create-text="去预约测试场地"
      @retry="fetchList"
      @create="goList"
    >
      <view class="list-body">
        <view
          v-for="bk in bookings"
          :key="bk.id"
          class="bk-card"
          hover-class="tap-fade"
          @click="goDetail(bk)"
        >
          <!-- 顶部：场地名 + 状态 -->
          <view class="bk-top">
            <text class="bk-site">{{ siteName(bk.site_id) }}</text>
            <text class="bk-status" :class="'status--' + bk.status">{{ statusLabel(bk.status) }}</text>
          </view>

          <!-- 预约时间 -->
          <view class="bk-time">{{ timeText(bk) }}</view>

          <!-- 用途 -->
          <view class="bk-purpose">
            <text class="bk-purpose-label">预约用途</text>
            <text class="bk-purpose-text">{{ bk.purpose || '—' }}</text>
          </view>

          <!-- 联系人 -->
          <view class="bk-meta">
            <text class="bk-meta-item">联系人 {{ bk.contact_name || '—' }}</text>
            <text class="bk-meta-item" v-if="bk.contact_phone">{{ bk.contact_phone }}</text>
          </view>

          <!-- 审核备注（rejected 时按需求单 6.3 语义展示为「拒绝原因」） -->
          <view v-if="bk.review_note" class="bk-note">{{ bk.status === 'rejected' ? '拒绝原因：' : '审核备注：' }}{{ bk.review_note }}</view>

          <!-- 按状态动作行：预约本体可行动（整卡点击仍去场地详情） -->
          <view v-if="bk.status === 'pending'" class="bk-act-row pending-hint">预计 24 小时内场地方将与您联系确认</view>
          <view v-else-if="bk.status === 'rejected'" class="bk-act-row">
            <view class="bk-rebook" @click.stop="goRebook(bk)">重新预约</view>
          </view>
          <view v-else-if="!bk.site_id" class="bk-act-row site-gone">场地已下架，无法查看详情</view>
        </view>
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { safeBack } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'
import StateView from '../../../components/StateView.vue'

const STATUS_MAP = { pending: '待审核', approved: '已确认', rejected: '已驳回', completed: '已完成' }

const { noMotion, checkMotion } = useReduceMotion()

const loading = ref(false)
const errorMsg = ref('')
const bookings = ref([])
const siteNames = ref({}) // site_id → 场地名（列表接口全量拼名）

function statusLabel(s) { return STATUS_MAP[s] || s || '其他状态' }
function siteName(id) { return siteNames.value[id] || '测试场地' }

// start_time/end_time 形如 "2026-08-20T09:00:00+08:00" → 口语日期 + HH:MM-HH:MM（评审 Minor：Date 解析防 Z 格式错钟点）
function fmtTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return String(iso).slice(11, 16)
  const p = (n) => (n < 10 ? '0' + n : '' + n)
  return p(d.getHours()) + ':' + p(d.getMinutes())
}
function dateText(iso) {
  if (!iso) return ''
  const m = String(iso).match(/^\d{4}-(\d{2})-(\d{2})/)
  return m ? Number(m[1]) + '月' + Number(m[2]) + '日' : ''
}
function timeText(bk) {
  const d = dateText(bk.start_time)
  const s = fmtTime(bk.start_time)
  const e = fmtTime(bk.end_time)
  if (!d) return '—'
  return s && e ? d + ' ' + s + '-' + e : d
}

async function fetchList(silent) {
  // 双轨（评审 P1）：有数据时静默刷新不清 loading（列表不闪断）；首次才显示骨架
  if (!silent) {
    loading.value = true
    errorMsg.value = ''
  }
  let bkOk = false
  try {
    const bkRes = await request({ url: '/api/v1/test-sites/bookings/mine' })
    const list = Array.isArray(bkRes) ? bkRes : (bkRes && bkRes.data) || []
    bookings.value = Array.isArray(list) ? list : []
    bkOk = true
  } catch (e) {
    if (!silent) errorMsg.value = '网络异常，请稍后重试'
  }
  // 场地名接口失败降级（评审 P1）：不整页报错，siteName 兜底 '测试场地'
  try {
    const siteRes = await request({ url: '/api/v1/test-sites' })
    const sites = Array.isArray(siteRes) ? siteRes : (siteRes && siteRes.data) || []
    const map = {}
    for (const s of sites) if (s && s.id) map[s.id] = s.name || ''
    siteNames.value = map
  } catch (e) {
    siteNames.value = {}
  }
  if (!silent && !bkOk) errorMsg.value = errorMsg.value || '网络异常，请稍后重试'
  loading.value = false
}

function goDetail(bk) {
  if (!bk.site_id) return
  uni.navigateTo({ url: '/pkg-service/pages/testsites/detail?id=' + encodeURIComponent(bk.site_id) })
}

function goRebook(bk) {
  // 已驳回重新预约：带原场地与日期回预约页（booking 页按 ?date= 预填）
  if (!bk.site_id) return
  const d = String(bk.start_time || '').slice(0, 10)
  const url = '/pkg-service/pages/testsites/booking?id=' + encodeURIComponent(bk.site_id) + (d ? '&date=' + encodeURIComponent(d) : '')
  uni.navigateTo({ url })
}

function goList() {
  uni.navigateTo({ url: '/pkg-service/pages/testsites/list' })
}

function goBack() {
  safeBack()
}

// onShow 而非 onLoad：预约提交返回后立即看到最新记录（评审 P1：已有数据静默刷新，列表不闪断）
onLoad(() => { checkMotion() })
onShow(() => {
  checkMotion()
  fetchList(bookings.value.length > 0)
})
onPullDownRefresh(() => {
  fetchList(true).finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.ts-page {
  min-height: 100vh;
  background: #fff; /* 白上白（对齐组）：与 list/detail/booking/pay/result 同范式 */
  padding-bottom: env(safe-area-inset-bottom);
}

.list-body {
  padding: 24rpx;
}

/* 白上白卡片（对齐组配方） */
.bk-card {
  background: #fff;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
  padding: 24rpx 32rpx;
  margin-bottom: 16rpx;
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.04),
    0 12rpx 32rpx rgba(10, 30, 60, 0.05);
  /* 按压弹簧：按下 .1s linear 即时到位，松手 .35s ios-pop 回弹（list 卡片同手感） */
  transition: transform 0.35s cubic-bezier(0.34, 1.8, 0.64, 1), opacity 0.15s ease;
}

.tap-fade {
  opacity: 0.85;
  transform: scale(0.96);
  /* 按压物理感：阴影随卡片收拢（list 同款） */
  box-shadow:
    0 1rpx 3rpx rgba(10, 30, 60, 0.03),
    0 6rpx 16rpx rgba(10, 30, 60, 0.04);
  transition-duration: 0.1s;
  transition-timing-function: linear;
}

.bk-top {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.bk-site {
  flex: 1;
  min-width: 0;
  font-size: 30rpx; /* 对齐组例外：卡片标题 */
  font-weight: 600;
  color: #17212B;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bk-status {
  flex-shrink: 0;
  font-size: 24rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  font-weight: 700;
}

/* 状态角标：对齐组 AA 暗变体（与 detail 状态徽章同族） */
.status--pending { color: #B45309; background: #FFF4E5; }
.status--approved { color: #0B6B41; background: #E9F7F0; }
.status--rejected { color: #B42318; background: #FEF0EF; }
.status--completed { color: #5D6B82; background: #EEF1F4; }

.bk-time {
  margin-top: 12rpx;
  font-size: 26rpx;
  font-weight: 600;
  color: #344054;
}

.bk-purpose {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
  margin-top: 8rpx;
}

.bk-purpose-label {
  flex-shrink: 0;
  font-size: 24rpx;
  color: #667085;
}

.bk-purpose-text {
  font-size: 26rpx;
  color: #344054;
  word-break: break-word; /* 硬化：长用途文本卡片内断行 */
}

.bk-meta {
  display: flex;
  align-items: center;
  gap: 24rpx;
  margin-top: 8rpx;
}

.bk-meta-item {
  font-size: 24rpx;
  color: #667085;
}

.bk-note {
  margin-top: 12rpx;
  padding-top: 12rpx;
  border-top: 2rpx solid #f0f1f3;
  font-size: 24rpx;
  color: #C2410C; /* 拒绝原因深橙：E96012 白底 ≈3.4:1 → 提深过 AA（评审 P2） */
  word-break: break-word; /* 硬化：长拒绝原因卡片内断行 */
}

/* 按状态动作行（Solid-Line 实线分隔） */
.bk-act-row {
  margin-top: 12rpx;
  padding-top: 12rpx;
  border-top: 2rpx solid #f0f1f3;
  display: flex;
  align-items: center;
}

.pending-hint {
  font-size: 24rpx;
  color: #667085;
}

.site-gone {
  font-size: 24rpx;
  color: #98A2B3;
}

/* 小号主行动按钮（对齐组按钮配方缩放；视觉 56rpx，padding 上下外扩热区至 88rpx——list b-sclr 同款技巧） */
.bk-rebook {
  margin-left: auto;
  height: 56rpx;
  padding: 16rpx 32rpx;
  margin-top: -16rpx;
  margin-bottom: -16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  font-weight: 600;
  border-radius: 16rpx;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.15s ease-out;
}
.bk-rebook:active { transform: scale(0.96); }

/* 减弱动效（无障碍）：关按压缩放，保留状态色与淡入（list 同款） */
.no-motion .tap-fade { transform: none !important; }
.no-motion .bk-rebook { transition: none; }
</style>
