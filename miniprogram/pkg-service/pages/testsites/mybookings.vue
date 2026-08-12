<template>
  <view class="ts-page">
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

          <!-- 审核备注 -->
          <view v-if="bk.review_note" class="bk-note">审核备注：{{ bk.review_note }}</view>
        </view>
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { safeBack } from '../../../utils/nav'
import StateView from '../../../components/StateView.vue'

const STATUS_MAP = { pending: '待审核', approved: '已通过', rejected: '已驳回', completed: '已完成' }

const loading = ref(false)
const errorMsg = ref('')
const bookings = ref([])
const siteNames = ref({}) // site_id → 场地名（列表接口全量拼名）

function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function siteName(id) { return siteNames.value[id] || '测试场地' }

// start_time/end_time 形如 "2026-08-20T09:00:00+08:00" → 日期 + HH:MM-HH:MM
function fmtTime(iso) {
  if (!iso) return ''
  return String(iso).slice(11, 16)
}
function timeText(bk) {
  const d = String(bk.start_time || '').slice(0, 10)
  const s = fmtTime(bk.start_time)
  const e = fmtTime(bk.end_time)
  if (!d) return '—'
  return s && e ? d + ' ' + s + '-' + e : d
}

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  try {
    const [bkRes, siteRes] = await Promise.all([
      request({ url: '/api/v1/test-sites/bookings/mine' }),
      request({ url: '/api/v1/test-sites' }),
    ])
    const list = Array.isArray(bkRes) ? bkRes : (bkRes && bkRes.data) || []
    bookings.value = Array.isArray(list) ? list : []
    // 场地名映射：site_id → name
    const sites = Array.isArray(siteRes) ? siteRes : (siteRes && siteRes.data) || []
    const map = {}
    for (const s of sites) if (s && s.id) map[s.id] = s.name || ''
    siteNames.value = map
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goDetail(bk) {
  if (!bk.site_id) return
  uni.navigateTo({ url: '/pkg-service/pages/testsites/detail?id=' + encodeURIComponent(bk.site_id) })
}

function goList() {
  uni.navigateTo({ url: '/pkg-service/pages/testsites/list' })
}

function goBack() {
  safeBack()
}

// onShow 而非 onLoad：预约提交返回后立即看到最新记录
onShow(() => fetchList())
onPullDownRefresh(() => {
  fetchList().finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.ts-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

.list-body {
  padding: 12px;
}

.bk-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.tap-fade {
  opacity: 0.7;
}

.bk-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.bk-site {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bk-status {
  flex-shrink: 0;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status--pending { color: #E96012; background: #FFF0E6; }
.status--approved { color: #168A55; background: #E9F7F0; }
.status--rejected { color: #D92D20; background: #FEF0EF; }
.status--completed { color: #667085; background: #F2F4F7; }

.bk-time {
  margin-top: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #344054;
}

.bk-purpose {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-top: 6px;
}

.bk-purpose-label {
  flex-shrink: 0;
  font-size: 11px;
  color: #98A2B3;
}

.bk-purpose-text {
  font-size: 13px;
  color: #344054;
}

.bk-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
}

.bk-meta-item {
  font-size: 11px;
  color: #667085;
}

.bk-note {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #F2F4F7;
  font-size: 11px;
  color: #E96012;
}
</style>
