<template>
  <view class="intents-page">
    <view class="page-head" :style="headStyle">
      <text class="back-btn" @tap="goBack">‹</text>
      <text class="page-title">我的对接意向</text>
      <text class="head-space"></text>
    </view>

    <view v-if="loading && list.length === 0" class="state-wrap">
      <u-loading size="28rpx" />
      <text class="state-text">加载中...</text>
    </view>
    <view v-else-if="list.length === 0" class="state-wrap">
      <u-empty description="暂无对接意向记录" />
      <text class="state-hint">在需求详情页点击「联系对接」即可登记意向</text>
    </view>
    <view v-else class="list-body">
      <view v-for="it in list" :key="it.id" class="intent-card" @tap="goDemand(it)">
        <view class="intent-line1">
          <text class="intent-demand-id">需求 #{{ shortId(it.demand_id) }}</text>
          <text class="intent-status" :class="'st-' + it.status">{{ statusLabel(it.status) }}</text>
        </view>
        <view class="intent-meta">
          <text class="meta-text">联系方式：{{ it.contact }}</text>
          <text class="meta-date">{{ formatDate(it.created_at) }}</text>
        </view>
        <text v-if="it.remark" class="intent-remark">{{ it.remark }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

// 自定义导航：状态栏留白（JS 方式）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const headStyle = computed(() => ({ paddingTop: statusBarH.value + 'px' }))

const list = ref([])
const loading = ref(false)

const goBack = () => uni.navigateBack()
const shortId = (id) => (id || '').length > 10 ? id.slice(-8) : (id || '-')
const statusLabel = (s) => ({ pending: '待联系', contacted: '已洽谈', done: '已成交', closed: '已关闭' }[s] || s || '')
const formatDate = (iso) => {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    const m = d.getMonth() + 1
    const day = d.getDate()
    return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
  } catch { return '' }
}
const goDemand = (it) => uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(it.demand_id) })

const fetchList = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/intents/mine' })
    const data = Array.isArray(res) ? res : (res && res.data) || []
    list.value = data
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

onLoad(fetchList)
onPullDownRefresh(() => {
  fetchList().finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.intents-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 14px 10px;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
}

.back-btn {
  font-size: 26px;
  color: #17212B;
  line-height: 1;
  width: 40px;
}

.page-title {
  font-size: 17px;
  font-weight: 600;
  color: #17212B;
}

.head-space {
  width: 40px;
}

.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 64px 0;
}

.state-text {
  font-size: 13px;
  color: #667085;
}

.state-hint {
  font-size: 12px;
  color: #98A2B3;
}

.list-body {
  padding: 12px;
}

.intent-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 10px;
}

.intent-line1 {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.intent-demand-id {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
}

.intent-status {
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 4px;
  font-weight: 500;
}

.st-pending {
  color: #B54708;
  background: #FFF7F1;
}

.st-contacted {
  color: #0A66C2;
  background: #EAF3FB;
}

.st-done {
  color: #168A55;
  background: #E9F7F0;
}

.st-closed {
  color: #667085;
  background: #F4F6F8;
}

.intent-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
}

.meta-text {
  font-size: 12px;
  color: #667085;
}

.meta-date {
  font-size: 11px;
  color: #98A2B3;
}

.intent-remark {
  display: block;
  font-size: 12px;
  color: #344054;
  margin-top: 8px;
  background: #F4F6F8;
  border-radius: 6px;
  padding: 8px 10px;
}
</style>
