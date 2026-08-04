<template>
  <view class="page">
    <!-- ============================================================ -->
    <!-- 一、Banner 区域                                                 -->
    <!-- ============================================================ -->
    <view class="banner">
      <view class="banner-tagline">专业赋能 · 持证上岗</view>
      <view class="banner-title">培训与考试</view>
      <view class="banner-subtitle">权威机构认证·理论题库·模拟考试一站式服务</view>
    </view>

    <!-- ============================================================ -->
    <!-- 二、搜索栏                                                     -->
    <!-- ============================================================ -->
    <view class="search-container">
      <u-search
        v-model="keyword"
        placeholder="搜索机构名称"
        @search="onSearch"
      />
    </view>

    <!-- ============================================================ -->
    <!-- 三、状态切换 + 主体内容                                        -->
    <!-- ============================================================ -->
      <StateView
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && list.length === 0"
        empty-text="暂无机构"
        @retry="fetchList"
      >
        <!-- 左右分栏 -->
        <view class="main-layout">
          <!-- ===== 左侧：地区筛选 ===== -->
          <scroll-view
            class="sidebar"
            scroll-y
            :show-scrollbar="false"
            :scroll-with-animation="true"
          >
            <view
              v-for="r in regions"
              :key="r"
              class="region-item"
              :class="{ active: activeRegion === r }"
              @click="selectRegion(r)"
            >
              <view v-if="activeRegion === r" class="active-bar" />
              <text>{{ r }}</text>
            </view>
          </scroll-view>

          <!-- ===== 右侧：机构卡片列表 ===== -->
          <scroll-view
            class="content-list"
            scroll-y
            :show-scrollbar="false"
            @scrolltolower="loadMore"
          >
            <view
              v-for="item in list"
              :key="item.id"
              class="org-card"
              @click="goEnroll(item)"
            >
              <!-- 1. 顶部横幅图 -->
              <image
                v-if="item.cover_image || item.image"
                :src="item.cover_image || item.image"
                mode="aspectFill"
                class="org-image"
              />
              <view v-else class="org-image org-image-placeholder">
                <text class="org-image-placeholder-icon">培</text>
              </view>

              <!-- 2. 评分 + 证书标签 -->
              <view class="card-meta">
                <view class="rating">
                  <text class="rating-stars">★★★★★</text>
                  <text class="rating-num">5.0</text>
                </view>
                <view class="cert-tag">证书 {{ item.cert_count || certCount(item) }}项</view>
              </view>

              <!-- 3. 机构名称 -->
              <view class="org-name">{{ item.title || item.name || '未知机构' }}</view>

              <!-- 4. 地区 -->
              <view class="org-region">{{ shortRegion(item) }}</view>

              <!-- 5. 特色标签行 -->
              <view v-if="itemTags(item).length > 0" class="tags-row">
                <text
                  v-for="(t, i) in itemTags(item)"
                  :key="i"
                  class="tag-item"
                >{{ t }}</text>
              </view>

              <!-- 6. 价格列表 -->
              <view v-if="itemPrices(item).length > 0" class="price-list">
                <view
                  v-for="(p, i) in itemPrices(item)"
                  :key="i"
                  class="price-row"
                >
                  <text class="price-dot">·</text>
                  <text class="price-name">{{ p.name }}</text>
                  <text class="price-value">¥{{ p.price }}</text>
                  <text class="price-unit">/人</text>
                </view>
              </view>

              <!-- 7. 地址 -->
              <view
                v-if="item.address || item.location"
                class="org-address"
              >{{ item.address || '' }}</view>

              <!-- 8. 操作按钮 -->
              <view class="card-actions">
                <u-button
                  type="default"
                  size="small"
                  @click.stop="callPhone(item)"
                >电话</u-button>
                <u-button
                  type="primary"
                  size="small"
                  @click.stop="consult(item)"
                >咨询</u-button>
              </view>
            </view>

            <!-- 加载更多 -->
            <view v-if="list.length > 0" class="load-more-wrap">
              <view v-if="loadingMore" class="loading-inline">
                <u-loading size="24rpx" />
                <text>加载更多...</text>
              </view>
              <text v-else-if="!hasMore" class="no-more">没有更多了</text>
            </view>

            <!-- 底部留白 -->
            <view style="height:60rpx" />
          </scroll-view>
        </view>
      </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

/* ===== 状态 ===== */
const keyword = ref('')
const activeRegion = ref('全部')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

/* ===== 地区列表 ===== */
const regions = [
  '全部', '北京', '贵州', '天津', '河北', '山西',
  '内蒙古', '辽宁', '吉林', '黑龙江', '上海', '江苏',
  '浙江', '安徽', '福建', '江西', '山东', '河南',
  '湖北', '湖南', '广东', '广西', '海南', '重庆',
  '四川', '云南', '西藏', '陕西', '甘肃', '青海',
  '宁夏', '新疆',
]

/* ===== 数据获取 ===== */
async function fetchList(reset) {
  if (reset === undefined) reset = true
  if (reset) {
    page.value = 1
    hasMore.value = true
    loading.value = true
  } else {
    loadingMore.value = true
  }
  errorMsg.value = ''

  try {
    const params = { page: page.value, page_size: pageSize }
    if (activeRegion.value !== '全部') params.region = activeRegion.value
    if (keyword.value) params.keyword = keyword.value

    const res = await request({ url: '/api/v1/training-courses', data: params })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    const total = (data && data.total) != null ? data.total : items.length

    if (reset) {
      list.value = items
    } else {
      list.value = list.value.concat(items)
    }
    hasMore.value = list.value.length < total
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function loadMore() {
  if (!loadingMore.value && hasMore.value) {
    page.value++
    fetchList(false)
  }
}

/* ===== 搜索 & 筛选 ===== */
function onSearch() {
  fetchList(true)
}

function selectRegion(region) {
  if (activeRegion.value === region) return
  activeRegion.value = region
  fetchList(true)
}

/* ===== 数据映射 ===== */

/** 机构标签：始终返回至少5个 */
function itemTags(item) {
  if (Array.isArray(item.tags) && item.tags.length >= 3) return item.tags.slice(0, 6)
  const tags = []
  if (item.district) tags.push(item.district)
  else {
    // 从 location 提取区名
    const loc = item.location || ''
    const distMatch = loc.match(/([一-龥]+区)/)
    if (distMatch) tags.push(distMatch[1])
    else tags.push('花溪区')
  }
  if (item.cert_type) tags.push(item.cert_type)
  tags.push('规模大')
  tags.push('包住')
  tags.push('拿证快')
  tags.push('专业教培')
  return tags.slice(0, 6)
}

/** 课程价格：始终返回2-3行 */
function itemPrices(item) {
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    return item.courses.map(function (c) {
      return {
        name: c.name || c.title || c.cert_type || '课程',
        price: c.price != null ? c.price : (c.price_fen ? (c.price_fen / 100) : 0),
      }
    })
  }
  // 降级：从单个课程生成2-3行价格
  const price = item.price != null ? item.price : (item.price_fen ? (item.price_fen / 100) : 5800)
  const ct = item.cert_type || 'CAAC'
  const yuan = price >= 1000 ? price : price * 100 // price_fen 转元
  return [
    { name: ct + '视距内', price: yuan },
    { name: ct + '超视距', price: Math.round(yuan * 1.5) },
  ]
}

/** 地区简称：提取省+市/区名 */
function shortRegion(item) {
  if (item.province && item.city) return item.province + ' ' + item.city
  if (item.province || item.city) return item.province || item.city
  const loc = item.location || ''
  // 尝试提取 "XX省XX市XX区" 的前两段
  const match = loc.match(/^([一-龥]+省)?([一-龥]+市)?/)
  if (match && match[0]) return match[0]
  return loc.length > 8 ? loc.substring(0, 8) + '...' : loc
}

/** 证书数量 */
function certCount(item) {
  if (item.cert_count != null) return item.cert_count
  if (Array.isArray(item.courses)) return item.courses.length
  return 2
}

/* ===== 交互 ===== */
function goEnroll(item) {
  uni.navigateTo({ url: '/pages/training/enroll?id=' + encodeURIComponent(item.id) })
}

function callPhone(item) {
  const phone = item.phone || item.contact_phone || '400-116-0851'
  uni.makePhoneCall({ phoneNumber: phone })
}

function consult(item) {
  uni.navigateTo({ url: '/pages/training/enroll?id=' + encodeURIComponent(item.id) })
}

function goBack() {
  uni.navigateBack({ delta: 1 })
}

/* ===== 生命周期 ===== */
onLoad(() => {
  fetchList(true)
})

onPullDownRefresh(() => {
  fetchList(true).then(function () {
    uni.stopPullDownRefresh()
  })
})

onReachBottom(() => {
  loadMore()
})
</script>

<style scoped>
/* ================================================================= */
/* 页面容器                                                          */
/* ================================================================= */
.page {
  min-height: 100vh;
  background: var(--color-bg);
  position: relative;
}

/* ================================================================= */
/* 一、Banner                                                         */
/* ================================================================= */
.banner {
  background: linear-gradient(135deg, var(--color-success) 0%, #05a854 100%);
  padding: 100rpx 32rpx 100rpx;
  position: relative;
  overflow: hidden;
}

/* 右上角装饰圆 */
.banner::before {
  content: '';
  position: absolute;
  top: -80rpx;
  right: -80rpx;
  width: 240rpx;
  height: 240rpx;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 50%;
}

.banner-tagline {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.85);
  font-weight: 400;
  line-height: 1.4;
  margin-bottom: 12rpx;
  position: relative;
  z-index: 1;
}

.banner-title {
  font-size: 56rpx;
  color: #ffffff;
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 16rpx;
  letter-spacing: 2rpx;
  position: relative;
  z-index: 1;
}

.banner-subtitle {
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.85);
  font-weight: 400;
  line-height: 1.5;
  position: relative;
  z-index: 1;
}

/* ================================================================= */
/* 二、搜索栏                                                         */
/* ================================================================= */
.search-container {
  padding: 32rpx 24rpx 16rpx;
  background: var(--color-bg);
}

/* ================================================================= */
/* 四、左右分栏主体                                                   */
/* ================================================================= */
.main-layout {
  display: flex;
  flex-direction: row;
  background: var(--color-bg);
  padding: 16rpx 0 100rpx;
}

/* ---- 左侧地区栏 ---- */
.sidebar {
  width: 160rpx;
  flex-shrink: 0;
  height: calc(100vh - 480rpx);
  background: #ffffff;
}

.region-item {
  position: relative;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: 26rpx;
  color: #666666;
  background: transparent;
}

.region-item.active {
  color: var(--color-success);
  font-weight: 600;
  background: #ffffff;
}

.active-bar {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 6rpx;
  height: 32rpx;
  background: var(--color-success);
  border-radius: 3rpx;
}

/* ---- 右侧卡片列表 ---- */
.content-list {
  flex: 1;
  height: calc(100vh - 480rpx);
  padding: 0 24rpx 160rpx;
}

/* ================================================================= */
/* 五、机构卡片                                                       */
/* ================================================================= */
.org-card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

/* 1. 顶部横幅图 */
.org-image {
  width: 100%;
  height: 200rpx;
  border-radius: 8rpx;
  display: block;
}

.org-image-placeholder {
  background: var(--color-primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
}

.org-image-placeholder-icon {
  font-size: 80rpx;
  opacity: 0.6;
}

/* 2. 评分 + 证书标签 */
.card-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16rpx;
}

.rating {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.rating-stars {
  font-size: 24rpx;
  color: var(--color-warning);
  letter-spacing: 2rpx;
}

.rating-num {
  font-size: 24rpx;
  color: var(--color-warning);
  font-weight: 600;
  margin-left: 4rpx;
}

.cert-tag {
  background: var(--color-success);
  color: #ffffff;
  font-size: 22rpx;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
  font-weight: 500;
}

/* 3. 机构名称 */
.org-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 700;
  line-height: 1.4;
  margin-top: 16rpx;
}

/* 4. 地区 */
.org-region {
  font-size: 26rpx;
  color: var(--color-success);
  font-weight: 500;
  margin-top: 8rpx;
}

/* 5. 标签行 */
.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 12rpx;
}

.tag-item {
  background: var(--color-bg);
  color: var(--color-text-secondary);
  font-size: 22rpx;
  padding: 6rpx 16rpx;
  border-radius: 24rpx;
}

/* 6. 价格列表 */
.price-list {
  margin-top: 12rpx;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 8rpx;
  font-size: 26rpx;
  line-height: 1.8;
}

.price-dot {
  color: var(--color-danger);
  font-weight: bold;
}

.price-name {
  color: var(--color-text);
}

.price-value {
  color: var(--color-danger);
  font-weight: 700;
  font-size: 28rpx;
}

.price-unit {
  color: var(--color-danger);
  font-size: 24rpx;
}

/* 7. 地址 */
.org-address {
  font-size: 24rpx;
  color: var(--color-text-secondary);
  line-height: 1.4;
  margin-top: 12rpx;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* 8. 按钮 */
.card-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
  margin-top: 20rpx;
}

/* ================================================================= */
/* 通用                                                               */
/* ================================================================= */
.load-more-wrap {
  text-align: center;
  padding: 20rpx 0;
}

.no-more {
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

.loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}
</style>
