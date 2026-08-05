<template>
  <view class="en-page">
    <u-nav-bar title="课程报名" show-back @back="goBack" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="课程不存在"
      @retry="fetchDetail"
    >
      <template v-if="detail">
        <view class="en-content">
          <!-- 实景图（4:3） -->
          <view class="hero-wrap" @tap="previewHero">
            <image v-if="heroImage" :src="heroImage" mode="aspectFill" class="hero-img" />
            <view v-else class="hero-img hero-placeholder">
              <text class="hero-placeholder-text">暂无实景图</text>
            </view>
          </view>

          <!-- 决策摘要：机构名 / 类型标签 / 价格 -->
          <view class="summary-card">
            <view v-if="tags.length > 0" class="summary-tags">
              <text v-for="(t, i) in tags" :key="i" class="tag-blue">{{ t }}</text>
              <text v-if="statusText" class="tag-status">{{ statusText }}</text>
            </view>
            <text class="summary-title">{{ titleText }}</text>
            <view class="summary-meta">
              <view class="meta-block">
                <text class="meta-label">课程费用</text>
                <text class="meta-value price">{{ priceText }}</text>
              </view>
              <view v-if="locationText" class="meta-block">
                <text class="meta-label">培训地点</text>
                <text class="meta-value">{{ locationText }}</text>
              </view>
              <view v-if="startDateText" class="meta-block">
                <text class="meta-label">开课时间</text>
                <text class="meta-value">{{ startDateText }}</text>
              </view>
            </view>
            <view v-if="quotaText" class="quota-row">
              <text class="quota-text">{{ quotaText }}</text>
            </view>
          </view>

          <!-- 课程内容 -->
          <view v-if="descriptionText" class="section-block">
            <text class="section-title">课程内容</text>
            <text class="desc-text">{{ descriptionText }}</text>
          </view>

          <!-- 费用明细（真实价格，price_fen → 元） -->
          <view v-if="priceRows.length > 0" class="section-block">
            <text class="section-title">费用明细</text>
            <text class="section-sub">元 / 人 · 仅供参考，以机构确认为准</text>
            <view class="price-list">
              <view v-for="(p, i) in priceRows" :key="i" class="price-item">
                <text class="price-name">{{ p.name }}</text>
                <view class="price-right">
                  <text class="price-symbol">¥</text>
                  <text class="price-value">{{ p.price }}</text>
                  <text class="price-unit">/{{ p.unit || '人' }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 培训环境 -->
          <view v-if="envImgs.length > 0" class="section-block">
            <text class="section-title">培训环境</text>
            <view class="photo-grid">
              <image
                v-for="(img, idx) in envImgs"
                :key="idx"
                :src="img"
                mode="aspectFill"
                class="photo"
                @tap="previewEnv(idx)"
              />
            </view>
          </view>

          <!-- 联系信息 -->
          <view class="section-block">
            <text class="section-title">联系信息</text>
            <view class="contact-item" @tap="openMap">
              <view class="contact-icon ic-blue"><text class="contact-icon-text">址</text></view>
              <view class="contact-content">
                <text class="contact-label">培训地点</text>
                <text class="contact-value">{{ locationText || '暂无' }}</text>
              </view>
              <text class="contact-arrow">›</text>
            </view>
            <view class="contact-item" @tap="callPhone">
              <view class="contact-icon ic-orange"><text class="contact-icon-text">话</text></view>
              <view class="contact-content">
                <text class="contact-label">联系电话</text>
                <text class="contact-value link">{{ phoneText }}</text>
              </view>
              <text class="contact-arrow">›</text>
            </view>
            <view v-if="businessHours" class="contact-item">
              <view class="contact-icon ic-green"><text class="contact-icon-text">时</text></view>
              <view class="contact-content">
                <text class="contact-label">营业时间</text>
                <text class="contact-value">{{ businessHours }}</text>
              </view>
            </view>
          </view>

          <!-- 报名须知 -->
          <view class="notice-block">
            <text class="notice-title">报名须知</text>
            <text class="notice-line">· 报名需提交姓名、手机号、身份证号及证件照片</text>
            <text class="notice-line">· 课程费用与开班安排以机构确认为准，线下签约缴费</text>
            <text class="notice-line warn">· 请通过本平台官方入口报名，切勿向个人账户转账</text>
          </view>
        </view>

        <view class="bottom-placeholder" />
      </template>
    </StateView>

    <!-- 底部操作栏 -->
    <view v-if="detail" class="bottom-bar">
      <view class="bb-btn bb-secondary" @tap="handleConsult">联系咨询</view>
      <view class="bb-btn bb-primary" @tap="handleEnroll">立即报名</view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const HOTLINE = '400-116-0851'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

/* 真实字段映射（只展示后端存在的字段，不做编造兜底） */
const CERT_LABELS = { caac: 'CAAC执照', utc_dji: '大疆UTC认证', gov_level: '人社等级证书' }
const STATUS_LABELS = { draft: '待上架', recruiting: '招生中', full: '已满', upcoming: '即将开课' }

const titleText = computed(function () {
  const item = detail.value
  return item ? (item.title || item.name || '未知机构') : ''
})

const heroImage = computed(function () {
  const item = detail.value
  const raw = item && (item.banner || item.cover_image || item.image || '')
  if (!raw) return ''
  return raw.startsWith('http') ? raw : BASE_URL + raw
})

/* 课程类型标签：course_types 数组优先，否则真实 cert_type 单标签 */
const tags = computed(function () {
  const item = detail.value
  if (!item) return []
  if (Array.isArray(item.course_types) && item.course_types.length > 0) {
    return item.course_types.slice(0, 4)
  }
  if (item.cert_type && CERT_LABELS[item.cert_type]) return [CERT_LABELS[item.cert_type]]
  return []
})

const statusText = computed(function () {
  const s = detail.value && detail.value.status
  return (s && STATUS_LABELS[s]) || ''
})

const priceText = computed(function () {
  const rows = priceRows.value
  if (rows.length > 0) return '¥' + rows[0].price
  return '面议'
})

/* 费用明细：courses 数组（若后端下发）逐条展示，否则单条 price_fen → 元 */
const priceRows = computed(function () {
  const item = detail.value
  if (!item) return []
  const rows = []
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    item.courses.forEach(function (c) {
      const price = c.price != null ? c.price : (c.price_fen ? c.price_fen / 100 : 0)
      if (price <= 0) return
      rows.push({ name: c.name || c.title || '课程', price: formatPrice(price), unit: c.unit || '人' })
    })
  } else {
    const price = item.price != null ? item.price : (item.price_fen ? item.price_fen / 100 : 0)
    if (price > 0) rows.push({ name: '课程费用', price: formatPrice(price), unit: '人' })
  }
  return rows
})

const locationText = computed(function () {
  const item = detail.value
  return item ? (item.location || item.district || '') : ''
})

const startDateText = computed(function () {
  const s = detail.value && detail.value.start_date
  return s ? String(s).slice(0, 10) : ''
})

const descriptionText = computed(function () {
  const item = detail.value
  return item ? (item.description || '') : ''
})

const quotaText = computed(function () {
  const item = detail.value
  if (!item || !item.max_students) return ''
  const enrolled = item.enrolled_count || 0
  return '本班已报 ' + enrolled + ' 人 / 共 ' + item.max_students + ' 人'
})

const phoneText = computed(function () {
  const item = detail.value
  return (item && (item.phone || item.contact_phone)) || HOTLINE
})

const businessHours = computed(function () {
  const item = detail.value
  return item ? (item.business_hours || '') : ''
})

const envImgs = computed(function () {
  const item = detail.value
  if (!item) return []
  let arr = []
  if (Array.isArray(item.environment) && item.environment.length > 0) arr = item.environment
  else if (Array.isArray(item.env_images) && item.env_images.length > 0) arr = item.env_images
  else if (Array.isArray(item.images) && item.images.length > 0) arr = item.images
  return arr.map(function (u) {
    return u && u.startsWith('http') ? u : BASE_URL + u
  })
})

function formatPrice(n) {
  return Number(n).toLocaleString()
}

/* === 数据获取：列表接口按 id 匹配（公开接口无单条详情） === */
async function fetchDetail() {
  loading.value = true
  errorMsg.value = ''

  try {
    const res = await request({ url: '/api/v1/training-courses' })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || data || []

    let found = null
    const targetId = String(id.value)
    for (let i = 0; i < items.length; i++) {
      if (String(items[i].id) === targetId) { found = items[i]; break }
    }
    detail.value = found
    if (!found) errorMsg.value = '课程不存在'
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

/* === 交互 === */
function goBack() { uni.navigateBack({ delta: 1 }) }

function previewHero() {
  if (heroImage.value) uni.previewImage({ urls: [heroImage.value] })
}

function openMap() {
  const addr = locationText.value
  uni.showToast({ title: addr ? '导航到：' + addr : '暂无地址信息', icon: 'none' })
}

function callPhone() {
  uni.makePhoneCall({ phoneNumber: phoneText.value })
}

function handleConsult() {
  uni.showToast({ title: '请联系客服 ' + HOTLINE, icon: 'none' })
}

function handleEnroll() {
  uni.navigateTo({ url: '/pages/training/register?id=' + encodeURIComponent(id.value) })
}

function previewEnv(idx) {
  const imgs = envImgs.value
  if (imgs.length > 0) uni.previewImage({ urls: imgs, current: imgs[idx] || imgs[0] })
}

onLoad(function (options) {
  id.value = options.id || ''
  fetchDetail()
})

onPullDownRefresh(function () {
  fetchDetail().then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.en-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
}

.en-content {
  padding: 20rpx 24rpx 0;
}

/* ===== 实景图 4:3 ===== */
.hero-wrap {
  position: relative;
  width: 100%;
  height: 0;
  padding-bottom: 75%;
  border-radius: 16rpx;
  overflow: hidden;
  background: #F4F6F8;
  margin-bottom: 20rpx;
}

.hero-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.hero-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}

.hero-placeholder-text {
  font-size: 26rpx;
  color: #98A2B3;
}

/* ===== 白卡 ===== */
.summary-card,
.section-block {
  background: #FFFFFF;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

/* 决策摘要 */
.summary-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 14rpx;
}

.tag-blue {
  font-size: 22rpx;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}

.tag-status {
  font-size: 22rpx;
  color: #168A55;
  background: #E9F7F0;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}

.summary-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
  display: block;
}

.summary-meta {
  display: flex;
  gap: 48rpx;
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #EEF1F4;
}

.meta-block {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  flex: 1;
}

.meta-label {
  font-size: 22rpx;
  color: #98A2B3;
}

.meta-value {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
  word-break: break-all;
}

.meta-value.price {
  color: #E96012;
  font-size: 34rpx;
}

.quota-row {
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #EEF1F4;
}

.quota-text {
  font-size: 24rpx;
  color: #E96012;
}

/* 分组区块 */
.section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  display: block;
  margin-bottom: 16rpx;
}

.section-sub {
  font-size: 22rpx;
  color: #98A2B3;
  display: block;
  margin-bottom: 16rpx;
}

.desc-text {
  font-size: 28rpx;
  color: #344054;
  line-height: 1.7;
  white-space: pre-wrap;
}

/* 费用明细 */
.price-list {
  border-top: 1rpx solid #EEF1F4;
}

.price-item {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #EEF1F4;
}

.price-item:last-child { border-bottom: none; }

.price-name {
  font-size: 28rpx;
  color: #344054;
  font-weight: 500;
}

.price-right { display: flex; align-items: baseline; }
.price-symbol { font-size: 24rpx; color: #E96012; font-weight: 600; }
.price-value { font-size: 34rpx; color: #E96012; font-weight: 700; margin: 0 4rpx; }
.price-unit { font-size: 22rpx; color: #98A2B3; }

/* 培训环境 */
.photo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12rpx;
}

.photo {
  width: 100%;
  height: 210rpx;
  border-radius: 12rpx;
  background: #F4F6F8;
}

/* 联系信息 */
.contact-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #EEF1F4;
}

.contact-item:last-child { border-bottom: none; }

.contact-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ic-blue { background: #EAF3FB; }
.ic-orange { background: #FFF0E6; }
.ic-green { background: #E9F7F0; }

.contact-icon-text {
  font-size: 28rpx;
  font-weight: 600;
}

.ic-blue .contact-icon-text { color: #0A66C2; }
.ic-orange .contact-icon-text { color: #E96012; }
.ic-green .contact-icon-text { color: #168A55; }

.contact-content { flex: 1; }

.contact-label {
  font-size: 22rpx;
  color: #98A2B3;
  display: block;
  margin-bottom: 4rpx;
}

.contact-value {
  font-size: 28rpx;
  color: #344054;
}

.contact-value.link {
  color: #0A66C2;
  font-weight: 600;
}

.contact-arrow {
  color: #98A2B3;
  font-size: 32rpx;
}

/* 报名须知 */
.notice-block {
  background: #EAF3FB;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.notice-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #0A66C2;
  display: block;
  margin-bottom: 10rpx;
}

.notice-line {
  display: block;
  font-size: 24rpx;
  color: #344054;
  line-height: 1.7;
}

.notice-line.warn {
  color: #E96012;
}

.bottom-placeholder { height: 20rpx; }

/* ===== 底部操作栏 ===== */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #FFFFFF;
  border-top: 1rpx solid #EEF1F4;
  display: flex;
  gap: 20rpx;
  z-index: 100;
}

.bb-btn {
  flex: 1;
  height: 88rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
}

.bb-primary {
  background: #0A66C2;
  color: #FFFFFF;
}

.bb-secondary {
  background: #FFFFFF;
  color: #0A66C2;
  border: 2rpx solid #0A66C2;
}
</style>
