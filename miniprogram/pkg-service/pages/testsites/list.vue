<template>
  <view class="ts-page">
    <u-nav-bar title="测试场地" show-back @back="goBack" />

    <u-search v-model="keyword" placeholder="搜索场地名称或位置" @change="onSearch" />

    <!-- 类型筛选 -->
    <view class="filter-bar">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-tabs">
          <view
            v-for="p in typePills"
            :key="p.value"
            class="filter-tab"
            :class="{ active: activeType === p.value }"
            @tap="selectType(p.value)"
          >{{ p.label }}</view>
        </view>
      </scroll-view>
    </view>

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && filteredSites.length === 0"
      empty-text="暂无测试场地"
      @retry="fetchList"
    >
      <view class="list-body">
        <view
          v-for="site in filteredSites"
          :key="site.id"
          class="site-card"
          hover-class="tap-fade"
          @click="goDetail(site)"
        >
          <!-- 类型图标块（低饱和色块） -->
          <view class="card-thumb" :class="'thumb--' + typeKey(site.site_type)">
            <text class="thumb-char">{{ typeChar(site.site_type) }}</text>
          </view>

          <view class="card-body">
            <view class="card-top">
              <text class="card-name">{{ site.name }}</text>
              <text class="card-status" :class="'status--' + site.status">{{ statusLabel(site.status) }}</text>
            </view>
            <text class="card-location">{{ site.location || '位置待定' }}</text>

            <view v-if="facilityTags(site.facilities).length > 0" class="tags-row">
              <text v-for="(f, i) in facilityTags(site.facilities)" :key="i" class="fac-tag">{{ f }}</text>
            </view>

            <view class="card-bottom">
              <view class="price-row">
                <text class="price-num">{{ formatPrice(site.price_fen) }}</text>
                <text class="price-unit">参考</text>
              </view>
              <text class="card-type">{{ typeLabel(site.site_type) }}</text>
            </view>
          </view>
        </view>
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const SITE_TYPE_MAP = {
  flying_field: '飞行场地',
  lab: '实验室',
  anechoic_chamber: '暗室',
  wind_tunnel: '风洞',
}
const STATUS_MAP = { available: '可预约', maintenance: '维护中', reserved: '已预约' }
const FACILITY_MAP = { '5G': '5G', RTK: 'RTK', radar: '雷达', spectrum_analyzer: '频谱分析' }
const TYPE_KEYS = { flying_field: 'fly', lab: 'lab', anechoic_chamber: 'chamber', wind_tunnel: 'tunnel' }
const TYPE_CHARS = { flying_field: '飞', lab: '实', anechoic_chamber: '暗', wind_tunnel: '风' }

const keyword = ref('')
const activeType = ref('all')
const loading = ref(false)
const errorMsg = ref('')
const sites = ref([])

const typePills = [
  { label: '全部', value: 'all' },
  { label: '飞行场地', value: 'flying_field' },
  { label: '实验室', value: 'lab' },
  { label: '暗室', value: 'anechoic_chamber' },
  { label: '风洞', value: 'wind_tunnel' },
]

// 关键字在前端过滤（后端列表接口仅支持 site_type 参数）
const filteredSites = computed(() => {
  const q = (keyword.value || '').trim().toLowerCase()
  if (!q) return sites.value
  return sites.value.filter((s) => {
    const name = (s.name || '').toLowerCase()
    const loc = (s.location || '').toLowerCase()
    const facs = (s.facilities || []).join(',').toLowerCase()
    return name.indexOf(q) >= 0 || loc.indexOf(q) >= 0 || facs.indexOf(q) >= 0
  })
})

function typeLabel(t) { return SITE_TYPE_MAP[t] || t || '测试场地' }
function typeKey(t) { return TYPE_KEYS[t] || 'fly' }
function typeChar(t) { return TYPE_CHARS[t] || '测' }
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

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  try {
    const params = {}
    if (activeType.value !== 'all') params.site_type = activeType.value
    const res = await request({ url: '/api/v1/test-sites', data: params })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    sites.value = Array.isArray(data) ? data : (data && data.items) || []
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function onSearch() {
  // 客户端过滤由 computed 自动完成
}

function selectType(value) {
  if (activeType.value === value) return
  activeType.value = value
  fetchList()
}

function goDetail(site) {
  uni.navigateTo({ url: '/pkg-service/pages/testsites/detail?id=' + encodeURIComponent(site.id) })
}

function goBack() {
  uni.navigateBack()
}

onLoad(() => fetchList())
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

/* 类型筛选 */
.filter-bar {
  background: #fff;
  padding: 8px 12px;
  border-bottom: 1px solid #EEF1F4;
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

.site-card {
  display: flex;
  gap: 12px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.tap-fade {
  opacity: 0.7;
}

.card-thumb {
  width: 64px;
  height: 64px;
  border-radius: 8px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.thumb--fly { background: #EAF3FB; }
.thumb--lab { background: #E9F7F0; }
.thumb--chamber { background: #F6F4FF; }
.thumb--tunnel { background: #FFF0E6; }

.thumb-char {
  font-size: 22px;
  font-weight: 700;
  color: #0A66C2;
}

.thumb--lab .thumb-char { color: #168A55; }
.thumb--chamber .thumb-char { color: #7A5AF8; }
.thumb--tunnel .thumb-char { color: #E96012; }

.card-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.card-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.card-name {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-status {
  flex-shrink: 0;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status--available { color: #168A55; background: #E9F7F0; }
.status--maintenance { color: #667085; background: #F2F4F7; }
.status--reserved { color: #E96012; background: #FFF0E6; }

.card-location {
  font-size: 12px;
  color: #667085;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.fac-tag {
  font-size: 10px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 2px 8px;
  border-radius: 4px;
}

.card-bottom {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-top: auto;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.price-num {
  font-size: 16px;
  font-weight: 700;
  color: #E96012;
}

.price-unit {
  font-size: 10px;
  color: #98A2B3;
}

.card-type {
  font-size: 11px;
  color: #98A2B3;
}
</style>
