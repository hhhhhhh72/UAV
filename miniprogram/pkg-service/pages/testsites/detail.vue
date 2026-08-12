<template>
  <view class="tsd-page">
    <u-nav-bar title="场地详情" show-back @back="goBack" />

    <!-- 加载中 -->
    <view v-if="loading" class="state-inline">
      <u-loading size="48rpx" color="#0A66C2" />
      <text class="state-text">加载中...</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="errorMsg" class="state-inline">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchDetail">重新加载</view>
    </view>

    <template v-else-if="site">
      <!-- 决策摘要 -->
      <view class="summary-card">
        <view class="summary-tags">
          <text class="tag-type">{{ typeLabel(site.site_type) }}</text>
          <text class="tag-status" :class="'status--' + site.status">{{ statusLabel(site.status) }}</text>
        </view>
        <text class="summary-title">{{ site.name }}</text>
        <view class="summary-meta">
          <view class="meta-block">
            <text class="meta-label">参考价格</text>
            <text class="meta-value price">{{ formatPrice(site.price_fen) }}</text>
          </view>
          <view class="meta-block">
            <text class="meta-label">位置</text>
            <text class="meta-value">{{ site.location || '位置待定' }}</text>
          </view>
        </view>
      </view>

      <!-- 场地设施 -->
      <view class="section-block">
        <text class="section-title">场地设施</text>
        <view v-if="facilityTags(site.facilities).length > 0" class="fac-grid">
          <text v-for="(f, i) in facilityTags(site.facilities)" :key="i" class="fac-item">{{ f }}</text>
        </view>
        <text v-else class="section-empty">暂无设施信息</text>
      </view>

      <!-- 位置 -->
      <view class="section-block">
        <text class="section-title">位置</text>
        <text class="section-text">{{ site.location || '位置待定' }}</text>
      </view>

      <!-- 预约规则 -->
      <view class="section-block">
        <text class="section-title">预约规则</text>
        <text class="section-text">{{ site.booking_rule || '以场地方实际安排为准' }}</text>
      </view>

      <!-- 费用说明 -->
      <view class="notice-block">
        <text class="notice-title">费用说明</text>
        <text class="notice-line">· 场地费用在线下向场地方支付，平台不参与资金流转</text>
        <text class="notice-line warn">· 预约提交后，请与场地方确认费用标准与支付方式</text>
      </view>
    </template>

    <!-- 底部操作栏 -->
    <view v-if="site" class="bottom-bar">
      <view class="bb-info">
        <text class="bb-label">参考价格</text>
        <text class="bb-price">{{ formatPrice(site.price_fen) }}</text>
      </view>
      <view
        class="bb-btn"
        :class="{ disabled: site.status === 'maintenance' }"
        @tap="goBooking"
      >{{ site.status === 'maintenance' ? '维护中暂不可约' : '预约测试' }}</view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { safeBack } from '../../../utils/nav'

const SITE_TYPE_MAP = {
  flying_field: '飞行场地',
  lab: '实验室',
  anechoic_chamber: '暗室',
  wind_tunnel: '风洞',
}
const STATUS_MAP = { available: '可预约', maintenance: '维护中', reserved: '已预约' }
const FACILITY_MAP = { '5G': '5G', RTK: 'RTK', radar: '雷达', spectrum_analyzer: '频谱分析' }

const loading = ref(false)
const errorMsg = ref('')
const site = ref(null)
let siteId = ''

function typeLabel(t) { return SITE_TYPE_MAP[t] || t || '测试场地' }
function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function facilityTags(list) {
  return (list || []).map((f) => FACILITY_MAP[f] || f)
}
function formatPrice(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const text = Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
  return '¥' + text
}

async function fetchDetail() {
  if (!siteId) return
  loading.value = true
  errorMsg.value = ''
  site.value = null
  try {
    const res = await request({ url: '/api/v1/test-sites/' + encodeURIComponent(siteId) })
    const d = (res && res.data) || res
    if (d && d.id) {
      site.value = d
    } else {
      errorMsg.value = '场地不存在或已下架'
    }
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goBooking() {
  if (site.value && site.value.status === 'maintenance') {
    uni.showToast({ title: '场地维护中，暂不可约', icon: 'none' })
    return
  }
  uni.navigateTo({ url: '/pkg-service/pages/testsites/booking?id=' + encodeURIComponent(siteId) })
}

function goBack() {
  safeBack()
}

onLoad((options) => {
  siteId = (options && options.id) || ''
  fetchDetail()
})
</script>

<style scoped>
.tsd-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 88px;
}

.state-inline {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80px 20px;
  gap: 12px;
}

.state-text {
  font-size: 13px;
  color: #667085;
}

.retry-btn {
  padding: 8px 24px;
  background: #0A66C2;
  color: #fff;
  border-radius: 8px;
  font-size: 13px;
}

/* 决策摘要 */
.summary-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 16px;
  margin: 12px;
}

.summary-tags {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.tag-type {
  font-size: 11px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 3px 8px;
  border-radius: 4px;
}

.tag-status {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
}

.status--available { color: #168A55; background: #E9F7F0; }
.status--maintenance { color: #667085; background: #F2F4F7; }
.status--reserved { color: #E96012; background: #FFF0E6; }

.summary-title {
  font-size: 19px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.35;
  display: block;
}

.summary-meta {
  display: flex;
  gap: 24px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #EEF1F4;
}

.meta-block {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.meta-label {
  font-size: 11px;
  color: #98A2B3;
}

.meta-value {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
}

.meta-value.price {
  color: #E96012;
  font-size: 17px;
}

/* 分组区块 */
.section-block {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px 16px;
  margin: 0 12px 8px;
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  display: block;
  margin-bottom: 10px;
}

.section-text {
  font-size: 14px;
  color: #344054;
  line-height: 1.7;
  white-space: pre-wrap;
}

.section-empty {
  font-size: 13px;
  color: #98A2B3;
}

/* 设施标签 */
.fac-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.fac-item {
  font-size: 12px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 4px 12px;
  border-radius: 4px;
}

/* 费用说明 */
.notice-block {
  background: #F4F8FC;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 0 12px 8px;
}

.notice-title {
  font-size: 13px;
  font-weight: 600;
  color: #0A66C2;
  display: block;
  margin-bottom: 6px;
}

.notice-line {
  display: block;
  font-size: 12px;
  color: #344054;
  line-height: 1.7;
}

.notice-line.warn {
  color: #B54708;
}

/* 底部操作栏 */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 12px;
  background: #fff;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: center;
  gap: 12px;
}

.bb-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.bb-label {
  font-size: 11px;
  color: #98A2B3;
}

.bb-price {
  font-size: 17px;
  font-weight: 700;
  color: #E96012;
}

.bb-btn {
  height: 46px;
  padding: 0 28px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
}

.bb-btn.disabled {
  background: #98A2B3;
}
</style>
