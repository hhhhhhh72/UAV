<template>
  <view class="page">
    <!-- ① 搜索栏（顶部直达，小程序胶囊标题） -->
    <view class="search-container">
      <view class="search-box">
        <u-icon name="search" size="28rpx" color="#ADB8C7" />
        <input
          class="search-input"
          v-model="keyword"
          placeholder="搜索研学活动"
          placeholder-class="search-ph"
          confirm-type="search"
          @confirm="onSearch"
          @input="onSearch"
        />
        <view v-if="keyword" class="search-clear" @click="clearKeyword"><text class="search-clear-x">×</text></view>
      </view>
    </view>

    <!-- ② 位置筛选条（左右分栏 + 副信息） -->
    <picker
      mode="selector"
      :range="chongqingDistricts"
      :value="districtIndex"
      @change="onDistrictChange"
      class="district-picker"
    >
      <view class="district-bar">
        <view class="district-left">
          <u-icon name="location" size="26rpx" color="#1E5EFF" />
          <text class="district-city">重庆市</text>
          <view class="district-divider" />
          <text class="district-value" :class="{ placeholder: !selectedDistrict }">{{ selectedDistrict || '全部区县' }}</text>
          <text class="district-arrow">▾</text>
        </view>
        <text class="district-count">{{ displayList.length }} 个活动</text>
      </view>
    </picker>

    <!-- ③ 主题分类 Pills（胶囊） -->
    <view class="filter-row">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-inner">
          <view
            v-for="pill in themePills"
            :key="pill.value"
            class="filter-pill"
            :class="{ active: activeTheme === pill.value }"
            @click="switchTheme(pill.value)"
          >{{ pill.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- ④ 研学卡片列表 -->
    <StateView
      class="state-fill"
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && displayList.length === 0"
      empty-text="暂无研学活动"
      @retry="loadData"
    >
      <view class="list-wrap">
        <view
          v-for="(item, i) in displayList"
          :key="item.id"
          class="study-card"
          hover-class="card-hover"
          :hover-stay-time="120"
          :style="{ animationDelay: (i * 60) + 'ms' }"
          @click="openDetail(item)"
        >
          <!-- 3.1 封面：有真实图显示图，无图兜底主题渐变 -->
          <view class="card-cover" :style="!item.cover_image ? { background: themeInfo(item).gradient } : {}">
            <!-- 真实封面图 -->
            <image
              v-if="item.cover_image"
              :src="item.cover_image"
              mode="aspectFill"
              class="cover-img"
              lazy-load
            />
            <!-- 无图兜底：本地主题封面图 -->
            <image v-else :src="coverByTheme(item)" mode="aspectFill" class="cover-img" lazy-load />
            <view class="status-badge" :style="{ background: statusStyle(item).bg, color: statusStyle(item).color }">
              {{ statusStyle(item).text }}
            </view>
          </view>

          <!-- 3.2 卡片信息区 -->
          <view class="card-body">
            <text class="card-name">{{ item.title || '未知活动' }}</text>

            <view class="card-info">
              <view class="info-row">
                <text class="info-icon">📅</text>
                <text class="info-label">时间</text>
                <text class="info-value">{{ dateRange(item) }}</text>
              </view>
              <view class="info-row">
                <text class="info-icon">📍</text>
                <text class="info-label">地点</text>
                <text class="info-value ellipsis">{{ item.location || '待定' }}</text>
              </view>
              <view class="info-row">
                <text class="info-icon">⏱️</text>
                <text class="info-label">时长</text>
                <text class="info-value">{{ item.duration || '待定' }}</text>
              </view>
              <view class="info-row">
                <text class="info-icon">👥</text>
                <text class="info-label">名额</text>
                <text class="info-value">
                  <text class="capacity-value">{{ capacityText(item) }}</text>
                </text>
              </view>
            </view>
          </view>

          <!-- 3.3 卡片底部 -->
          <view class="card-footer">
            <text class="detail-hint">点击查看详情</text>
            <text class="footer-arrow">›</text>
          </view>
        </view>

        <!-- 列表底部 -->
        <view v-if="!loading && displayList.length > 0" class="list-footer">
          <text class="footer-line" />
          <text class="footer-text">没有更多了</text>
          <text class="footer-line" />
        </view>
        <view style="height: 40rpx" />
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 50
const hasMore = ref(true)

// ── 地区筛选（重庆 38 个区县） ──────────────────────────
const chongqingDistricts = ['全部区县', '渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区', '南岸区', '北碚区', '渝北区', '巴南区', '两江新区', '长寿区', '江津区', '合川区', '永川区', '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区', '武隆区', '万州区', '黔江区', '涪陵区', '奉节县', '云阳县', '忠县', '垫江县', '丰都县', '城口县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县']
const districtIndex = ref(0)
const selectedDistrict = computed(() => chongqingDistricts[districtIndex.value])
const onDistrictChange = (e) => { districtIndex.value = Number(e.detail.value); page.value = 1; loadData(true) }

// ── 主题分类 Pills ────────────────────
const themePills = [
  { label: '全部', value: 'all' },
  { label: '科普研学', value: 'science' },
  { label: '产业研学', value: 'industry' },
  { label: '实践研学', value: 'practice' },
  { label: '职业研学', value: 'career' },
]
const activeTheme = ref('all')
const switchTheme = (v) => { activeTheme.value = v; page.value = 1; loadData(true) }

// ── 主题推断（title/description 关键词）──
const THEMES = [
  { key: ['职业', '院校', '学院', '专业', '开放日'], value: 'career', label: '职业研学', short: '职业', icon: '职', gradient: 'linear-gradient(135deg,#FF8E3C,#FFB36B)' },
  { key: ['实践', '实训', '训练营', '穿越机', '巡检', '应急救援', '测绘', '实战', '试飞'], value: 'practice', label: '实践研学', short: '实践', icon: '实', gradient: 'linear-gradient(135deg,#065F46,#06B6D4)' },
  { key: ['科普', '科技', '科学', '航模', '体验', '组装'], value: 'science', label: '科普研学', short: '科普', icon: '学', gradient: 'linear-gradient(135deg,#0A1F44,#1E5EFF)' },
  { key: ['产业', '低空经济', '企业', '龙头'], value: 'industry', label: '产业研学', short: '产业', icon: '产', gradient: 'linear-gradient(135deg,#6D28D9,#DB2777)' },
]
const themeInfo = (item) => {
  const title = item.title || ''
  const desc = item.description || ''
  // 先匹配标题（更权威），再匹配描述
  for (const t of THEMES) {
    if (t.key.some((k) => title.includes(k))) return t
  }
  for (const t of THEMES) {
    if (t.key.some((k) => desc.includes(k))) return t
  }
  return { value: 'science', label: '科普研学', short: '研学', icon: '研', gradient: 'linear-gradient(135deg,#0A1F44,#1E5EFF)' }
}

// ── 本地兜底封面：无 cover_image 时按主题映射静态图 ──
const COVER_BY_THEME = { career: 1, practice: 2, science: 3, industry: 1 }
const coverByTheme = (item) => '/static/images/study/cover-' + (COVER_BY_THEME[themeInfo(item).value] || 1) + '.jpg'

// ── 状态徽章 ──────────────────────────
const statusStyle = (item) => {
  const s = item.status || 'active'
  if (s === 'closed') return { text: '已结束', bg: '#EEF0F4', color: '#ADB8C7' }
  if (s === 'draft') return { text: '即将开始', bg: 'rgba(239,68,68,.1)', color: '#EF4444' }
  return { text: '招募中', bg: 'rgba(255,142,60,.12)', color: '#FF8E3C' }
}

// ── 列表：过滤 + 筛选 ──────────────────
const baseList = computed(() => list.value.filter((it) => it.status === 'active' || it.status === ''))
const filteredByTheme = computed(() => {
  if (activeTheme.value === 'all') return baseList.value
  return baseList.value.filter((it) => themeInfo(it).value === activeTheme.value)
})
const displayList = computed(() => {
  if (!selectedDistrict.value || selectedDistrict.value === '全部区县') return filteredByTheme.value
  return filteredByTheme.value.filter((it) => (it.location || '').includes(selectedDistrict.value.slice(0, 2)))
})

// ── 卡片字段 ──────────────────────────
const capacityText = (item) => {
  const c = item.capacity
  if (c == null || c <= 0) return '不限'
  return `${c} 人`
}

// 时间范围：start_date ~ end_date（ISO 时间戳；零值时间视为未设置）
function dateRange(item) {
  const s = item.start_date
  const e = item.end_date
  const fmt = (v) => {
    if (!v) return ''
    const d = new Date(v)
    if (isNaN(d.getTime())) return ''
    if (d.getFullYear() <= 1) return ''
    const p = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}年${p(d.getMonth() + 1)}月${p(d.getDate())}日`
  }
  const ss = fmt(s)
  const ee = fmt(e)
  if (ss && ee && ss !== ee) return `${ss}-${ee.replace(/^\d{4}年/, '')}`
  return ss || ee || '待定'
}

function openDetail(item) {
  uni.setStorageSync('study_tour_detail', item)
  uni.navigateTo({ url: '/pkg-talent/pages/study/detail?id=' + encodeURIComponent(item.id) })
}

// ── 搜索 ──────────────────────────────
const clearKeyword = () => { keyword.value = ''; onSearch() }
var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { page.value = 1; loadData(true) }, 300)
}

// ── 数据加载 ──────────────────────────
async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''
  try {
    var params = { page: page.value, page_size: pageSize }
    if (keyword.value) params.keyword = keyword.value
    var res = await request({ url: '/api/v1/study/tours', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []
    var total = (data && data.total) != null ? data.total : items.length
    if (reset) { list.value = items } else { list.value = list.value.concat(items) }
    hasMore.value = list.value.length < total
  } catch (e) {
    if (reset) errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

onLoad(() => loadData(true))
onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: 40rpx;
}

/* ═══ ① 搜索框 ═══ */
.search-container {
  padding: 20rpx 24rpx 0;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 14rpx;
  background: #F5F8FC;
  border-radius: 999rpx;
  padding: 18rpx 26rpx;
  border: 1rpx solid #EDF0F5;
}
.search-input {
  flex: 1;
  font-size: 26rpx;
  color: #17212B;
}
.search-ph {
  color: #ADB8C7;
}
.search-clear {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #EDF0F5;
  display: flex;
  align-items: center;
  justify-content: center;
}
.search-clear-x {
  font-size: 28rpx;
  color: #6B7B95;
  line-height: 1;
}

/* ═══ ② 区县筛选条 ═══ */
.district-picker {
  display: block;
  margin: 20rpx 24rpx 0;
}
.district-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 0;
  border-bottom: 1rpx solid #EEF0F5;
}
.district-left {
  display: flex;
  align-items: center;
  gap: 10rpx;
}
.district-city {
  font-size: 26rpx;
  font-weight: 700;
  color: #1E5EFF;
}
.district-divider {
  width: 1rpx;
  height: 24rpx;
  background: #E0E6EF;
}
.district-value {
  font-size: 24rpx;
  color: #0A1F44;
}
.district-value.placeholder {
  color: #ADB8C7;
}
.district-arrow {
  font-size: 20rpx;
  color: #6B7B95;
}
.district-count {
  font-size: 22rpx;
  color: #6B7B95;
}

/* ═══ ③ 主题分类 Pills ═══ */
.filter-row {
  margin: 20rpx 24rpx 0;
}
.filter-scroll {
  width: 100%;
  white-space: nowrap;
}
.filter-inner {
  display: inline-flex;
  gap: 12rpx;
  padding-bottom: 4rpx;
}
.filter-pill {
  display: inline-block;
  padding: 10rpx 26rpx;
  border-radius: 999rpx;
  background: transparent;
  color: #6B7B95;
  font-size: 24rpx;
  border: 1rpx solid transparent;
  transition: all 0.2s;
}
.filter-pill.active {
  background: #0A1F44;
  color: #ffffff;
  font-weight: 600;
}

/* ═══ ④ 研学卡片 ═══ */
.state-fill {
  margin-top: 20rpx;
}
.list-wrap {
  padding: 0 24rpx;
}
.study-card {
  background: #ffffff;
  border-radius: 16rpx;
  overflow: hidden;
  margin-bottom: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
  transition: background 0.2s;
  animation: cardIn 0.4s ease both;
}
.card-hover {
  background: #EEF3FA;
  transform: scale(0.98);
}

/* 3.1 主题色块 */
.card-cover {
  position: relative;
  height: 140rpx;
  overflow: hidden;
}
.cover-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.cover-icon {
  position: absolute;
  left: 24rpx;
  top: 50%;
  transform: translateY(-50%);
  font-size: 56rpx;
  font-weight: 700;
  color: rgba(255,255,255,0.18);
}
.cover-theme {
  position: absolute;
  left: 24rpx;
  bottom: 16rpx;
  font-size: 30rpx;
  font-weight: 700;
  color: #ffffff;
}
.status-badge {
  position: absolute;
  top: 16rpx;
  right: 16rpx;
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 700;
  animation: badgePulse 2s ease-in-out infinite;
}

/* 3.2 信息区 */
.card-body {
  padding: 24rpx;
}
.card-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #0A1F44;
  display: block;
  margin-bottom: 16rpx;
}
.card-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}
.info-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
}
.info-icon {
  width: 36rpx;
  font-size: 26rpx;
  text-align: center;
  flex-shrink: 0;
}
.info-label {
  font-size: 24rpx;
  color: #6B7B95;
  width: 56rpx;
  flex-shrink: 0;
}
.info-value {
  font-size: 24rpx;
  color: #2C3E50;
}
.info-value.ellipsis {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  flex: 1;
}
.capacity-value {
  color: #00C896;
  font-weight: 700;
  font-size: 28rpx;
}

/* 3.3 卡片底部 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 24rpx;
  padding: 16rpx 0;
  border-top: 1rpx solid #E8EEF7;
}
.detail-hint {
  font-size: 26rpx;
  color: #1E5EFF;
  font-weight: 600;
}
.footer-arrow {
  font-size: 36rpx;
  color: #1E5EFF;
  line-height: 1;
  font-weight: 300;
}

/* ═══ 列表底部 ═══ */
.list-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20rpx;
  padding: 32rpx 0;
}
.footer-line {
  width: 100rpx;
  height: 1rpx;
  background: linear-gradient(90deg, rgba(107,123,149,0), rgba(107,123,149,.4), rgba(107,123,149,0));
}
.footer-text {
  font-size: 22rpx;
  color: #ADB8C7;
}

/* ═══ 微动效 ═══ */
@keyframes cardIn {
  from {
    transform: translateY(20rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
@keyframes badgePulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}
</style>
