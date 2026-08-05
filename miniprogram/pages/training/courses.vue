<template>
  <view class="courses-page">
    <!-- 搜索 -->
    <u-sticky>
      <u-search
        v-model="keyword"
        placeholder="搜索培训机构"
        @search="onSearch"
      />
    </u-sticky>

    <!-- 区县筛选 + 证书类型筛选 -->
    <view class="filter-bar">
      <picker
        mode="selector"
        :range="chongqingDistricts"
        :value="districtIndex"
        @change="onDistrictChange"
      >
        <view class="district-row">
          <text class="district-city">重庆市</text>
          <text class="district-divider"></text>
          <text class="district-value" :class="{ placeholder: !selectedDistrict }">{{ selectedDistrict || '全部区县' }}</text>
          <text class="district-arrow">▾</text>
        </view>
      </picker>
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-tabs">
          <view
            v-for="p in certPills"
            :key="p.value"
            class="filter-tab"
            :class="{ active: activeCertType === p.value }"
            @tap="selectCertType(p.value)"
          >{{ p.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 状态区 -->
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && list.length === 0"
      empty-text="暂无培训机构"
      @retry="fetchList"
    >
      <!-- 机构横卡列表 -->
      <view class="list-body">
        <view
          v-for="item in list"
          :key="item.id"
          class="org-card"
          hover-class="tap-fade"
          @click="goEnroll(item)"
        >
          <image
            v-if="item.cover_image || item.image"
            :src="item.cover_image || item.image"
            mode="aspectFill"
            class="org-image"
          />
          <view v-else class="org-image org-image-placeholder">
            <text class="org-image-text">培</text>
          </view>

          <view class="org-body">
            <text class="org-name">{{ item.title || item.name || '未知机构' }}</text>
            <text class="org-region">{{ shortRegion(item) }}</text>

            <view v-if="itemTags(item).length > 0" class="tags-row">
              <text
                v-for="(t, i) in itemTags(item)"
                :key="i"
                class="tag-item"
              >{{ t }}</text>
            </view>

            <view v-if="itemPrices(item).length > 0" class="price-line">
              <text class="price-name">{{ itemPrices(item)[0].name }}</text>
              <text class="price-value">¥{{ itemPrices(item)[0].price }}</text>
              <text v-if="itemPrices(item).length > 1" class="price-more">+{{ itemPrices(item).length - 1 }} 项</text>
            </view>

            <view class="card-actions">
              <view class="act-btn act-phone" @click.stop="callPhone(item)">电话</view>
              <view class="act-btn act-primary" @click.stop="consult(item)">咨询报名</view>
            </view>
          </view>
        </view>

        <!-- 加载更多 -->
        <view v-if="list.length > 0" class="load-more">
          <view v-if="loadingMore" class="loading-inline">
            <u-loading size="24rpx" />
            <text>加载更多...</text>
          </view>
          <text v-else-if="!hasMore" class="no-more">没有更多了</text>
        </view>
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
const selectedDistrict = ref('') // 空 = 全部区县
const districtIndex = ref(0)
const activeCertType = ref('all')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

/* ===== 证书类型筛选 ===== */
const certPills = [
  { label: '全部', value: 'all' },
  { label: 'CAAC', value: 'caac' },
  { label: 'UTC', value: 'utc_dji' },
  { label: '人社等级', value: 'gov_level' },
]

/* ===== 重庆 38 区县 ===== */
const chongqingDistricts = [
  '渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区',
  '南岸区', '北碚区', '渝北区', '巴南区', '万州区',
  '涪陵区', '黔江区', '长寿区', '江津区', '合川区',
  '永川区', '南川区', '綦江区', '大足区', '璧山区',
  '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区',
  '武隆区', '城口县', '丰都县', '垫江县', '忠县',
  '云阳县', '奉节县', '巫山县', '巫溪县',
  '石柱土家族自治县', '秀山土家族苗族自治县',
  '酉阳土家族苗族自治县', '彭水苗族土家族自治县',
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
    if (activeCertType.value !== 'all') params.cert_type = activeCertType.value
    if (selectedDistrict.value) params.region = selectedDistrict.value
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
  if (loadingMore.value || !hasMore.value) return
  page.value++
  fetchList(false)
}

function onSearch() {
  fetchList(true)
}

function onDistrictChange(e) {
  const idx = Number(e.detail.value)
  districtIndex.value = idx
  selectedDistrict.value = chongqingDistricts[idx]
  fetchList(true)
}

function selectCertType(value) {
  if (activeCertType.value === value) return
  activeCertType.value = value
  fetchList(true)
}

function itemTags(item) {
  if (Array.isArray(item.tags) && item.tags.length >= 3) return item.tags.slice(0, 6)
  const tags = []
  if (item.district) tags.push(item.district)
  else {
    const loc = item.location || ''
    const match = loc.match(/^([一-龥]+省)?([一-龥]+市)?([一-龥]+区|县)?/)
    if (match && match[0]) tags.push(match[0])
  }
  if (item.cert_type) tags.push(item.cert_type)
  if (item.category) tags.push(item.category)
  return tags.slice(0, 4)
}

function itemPrices(item) {
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    return item.courses.map(function (c) {
      return {
        name: c.name || c.title || c.cert_type || '课程',
        price: c.price != null ? c.price : (c.price_fen ? (c.price_fen / 100) : 0),
      }
    })
  }
  if (item.price != null) {
    return [{ name: item.course_name || '课程', price: item.price }]
  }
  if (item.price_fen) {
    return [{ name: item.course_name || '课程', price: item.price_fen / 100 }]
  }
  return []
}

function shortRegion(item) {
  if (item.province && item.city) return item.province + ' ' + item.city
  if (item.province || item.city) return item.province || item.city
  const loc = item.location || ''
  const match = loc.match(/^([一-龥]+省)?([一-龥]+市)?/)
  if (match && match[0]) return match[0]
  return item.district || '重庆'
}

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

onLoad(() => fetchList(true))
onPullDownRefresh(() => {
  fetchList(true).finally(() => uni.stopPullDownRefresh())
})
onReachBottom(loadMore)
</script>

<style scoped>
.courses-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 筛选 */
.filter-bar {
  background: #fff;
  padding: 8px 12px;
}

/* 区县筛选行 */
.district-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0 10px;
}

.district-city {
  font-size: 13px;
  font-weight: 600;
  color: #17212B;
}

.district-divider {
  width: 1px;
  height: 12px;
  background: #E4E7EC;
}

.district-value {
  font-size: 13px;
  color: #0A66C2;
  font-weight: 500;
}

.district-value.placeholder {
  color: #98A2B3;
  font-weight: 400;
}

.district-arrow {
  font-size: 10px;
  color: #98A2B3;
}

.filter-scroll {
  white-space: nowrap;
}

.filter-tabs {
  display: inline-flex;
  gap: 8px;
}

.filter-tab {
  flex-shrink: 0;
  padding: 6px 16px;
  border-radius: 8px;
  font-size: 13px;
  color: #344054;
  background: #F4F6F8;
  border: 1px solid #EEF1F4;
  transition: all 0.2s;
}

.filter-tab.active {
  color: #fff;
  background: #0A66C2;
  border-color: #0A66C2;
}

/* 列表 */
.list-body {
  padding: 12px;
}

.org-card {
  display: flex;
  gap: 12px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.org-image {
  width: 112px;
  height: 112px;
  border-radius: 8px;
  background: #F4F6F8;
  flex-shrink: 0;
}

.org-image-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}

.org-image-text {
  font-size: 22px;
  font-weight: 700;
  color: #98A2B3;
}

.org-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.org-name {
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.org-region {
  font-size: 11px;
  color: #98A2B3;
}

.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-item {
  font-size: 10px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 2px 8px;
  border-radius: 4px;
}

.price-line {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.price-name {
  font-size: 12px;
  color: #667085;
}

.price-value {
  font-size: 16px;
  font-weight: 700;
  color: #E96012;
}

.price-more {
  font-size: 11px;
  color: #98A2B3;
}

.card-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.act-btn {
  flex: 1;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 500;
}

.act-phone {
  background: #fff;
  color: #0A66C2;
  border: 1px solid #0A66C2;
}

.act-primary {
  background: #0A66C2;
  color: #fff;
}

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.loading-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #667085;
}

.no-more {
  color: #98A2B3;
  font-size: 12px;
}

.tap-fade {
  opacity: 0.7;
}
</style>
