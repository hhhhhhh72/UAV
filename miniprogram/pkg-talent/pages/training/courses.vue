<template>
  <view class="page">
    <!-- ① 自定义导航栏（白底 + 返回 + 标题 + 胶囊占位） -->
    <view class="nav-wrap">
      <view class="nav-bar">
        <view class="nav-back" hover-class="nav-press" :hover-stay-time="100" @click="goBack">
          <text class="nav-back-icon">‹</text>
        </view>
        <text class="nav-title">培训认证</text>
        <view class="nav-capsule">
          <view class="capsule-dot" />
          <view class="capsule-divider" />
          <view class="capsule-arrow" />
        </view>
      </view>
      <view class="nav-meta">
        <view class="meta-sync">
          <view class="sync-dot" />
          <text class="sync-text">已同步 · CAAC / UTC 考试管理中心</text>
        </view>
        <view class="meta-cert">
          <text class="meta-cert-text">资质认证</text>
          <view class="meta-cert-arrow" />
        </view>
      </view>
    </view>

    <!-- ② 工具栏：搜索（无筛选按钮） + 城市卡 -->
    <view class="toolbar">
      <view class="search-box" :class="{ focus: searchFocus }">
        <view class="search-ico"><view class="search-ring" /><view class="search-bar-ico" /></view>
        <input
          class="search-input"
          v-model="keyword"
          placeholder="搜索机构或课程名称"
          placeholder-class="search-ph"
          confirm-type="search"
          @focus="searchFocus = true"
          @blur="searchFocus = false"
          @input="onSearchInput"
          @confirm="onSearch"
        />
        <text v-show="!keyword && !searchFocus" class="search-kbd">⌘K</text>
        <view v-if="keyword" class="search-clear" @click="clearKeyword"><text class="search-clear-x">×</text></view>
      </view>

      <picker
        mode="selector"
        :range="chongqingDistricts"
        :value="districtIndex"
        @change="onDistrictChange"
        class="city-picker"
      >
        <view class="city-card">
          <view class="city-left">
            <view class="city-ico"><view class="loc-pin" /></view>
            <view class="city-copy">
              <text class="city-name">重庆市</text>
              <text class="city-sub">{{ selectedDistrict || '全部区县' }}</text>
            </view>
            <view class="city-arrow" />
          </view>
          <view class="city-count">
            共 <text class="count-num">{{ filteredList.length }}</text> 家机构
          </view>
        </view>
      </picker>
    </view>

    <!-- ③ 证书类型 Tab（四态 + 数量徽标 + 主题色图标） -->
    <scroll-view scroll-x class="cert-tabs" :show-scrollbar="false">
      <view class="cert-tabs-inner">
        <view
          v-for="t in certPills"
          :key="t.value"
          class="cert-tab"
          :class="{ on: activeCertType === t.value }"
          @click="selectCertType(t.value)"
        >
          <view class="cert-tab-ico" :class="'cti--' + t.value" />
          <text class="cert-tab-label">{{ t.label }}</text>
          <view v-if="certCount(t.value)" class="cert-tab-count" :class="{ 'cert-tab-count--on': activeCertType === t.value }">
            {{ certCount(t.value) }}
          </view>
        </view>
      </view>
    </scroll-view>

    <!-- ④ 列表区 -->
    <view class="content">
      <StateView
        class="state-fill"
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && filteredList.length === 0"
        empty-text="暂无课程"
        @retry="fetchList(true)"
      >
        <scroll-view class="list-scroll" scroll-y :show-scrollbar="false">
          <!-- 副标题行 -->
          <view class="sub-bar">
            <text class="sub-text">共 {{ filteredList.length }} 门课程 · {{ sortLabel }}</text>
            <view class="sort-pill" hover-class="sort-press" :hover-stay-time="100" @click="showSortSheet">
              <text class="sort-text">排序</text>
              <view class="sort-arrow" />
            </view>
          </view>

          <!-- 课程卡片 -->
          <view
            v-for="(item, i) in filteredList"
            :key="item.id"
            class="card"
            :style="{ animationDelay: (i * 80) + 'ms' }"
            hover-class="card-press"
            :hover-stay-time="120"
            @click="goEnroll(item)"
          >
            <!-- 顶部横幅 16:9 -->
            <view class="banner">
              <image
                v-if="coverOf(item)"
                :src="coverOf(item)"
                class="banner-img"
                mode="aspectFill"
                lazy-load
                @load="onImgLoad(item.id)"
                :style="{ opacity: imgLoaded[item.id] ? 1 : 0 }"
              />
              <view v-else class="banner-fallback">
                <view class="cert-char">{{ certChar(item.cert_type) }}</view>
              </view>
              <view class="banner-mask" />

              <!-- 左上证书类型徽章 -->
              <view class="cert-badge" :class="'cb--' + certColorType(item.cert_type)">
                <text class="cert-badge-text">{{ certTypeLabel(item.cert_type) }}</text>
              </view>

              <!-- 右上状态徽章 -->
              <view class="status-badge" :class="'sb--' + statusBtn(item).type">
                <view v-if="statusBtn(item).type === 'enroll'" class="status-dot" />
                <text class="status-text">{{ statusBtn(item).text }}</text>
              </view>

              <!-- 左下评价信息：叠层头像 + 评价数 -->
              <view class="review-row">
                <view class="rev-avatars">
                  <view class="rev-avatar ra--1" />
                  <view class="rev-avatar ra--2" />
                  <view class="rev-avatar ra--3" />
                </view>
                <text class="rev-count">{{ reviewCount(item) }} 人评价</text>
              </view>
            </view>

            <!-- 卡片内容 -->
            <view class="card-info">
              <view class="title-row">
                <text class="org-name">{{ orgName(item) }}</text>
                <view class="rating-badge">
                  <text class="rating-star">★</text>
                  <text class="rating-num">{{ item.rating || '5.0' }}</text>
                </view>
              </view>
              <text class="card-subtitle">{{ item.title || item.name || '未知课程' }}</text>
              <view class="card-meta">
                <view class="meta-ico"><view class="loc-pin" /></view>
                <text class="meta-text">{{ shortRegion(item) }}</text>
              </view>
              <view class="card-tags">
                <text v-for="t in cardTags(item)" :key="t" class="pill" :class="pillCls(t)">{{ t }}</text>
              </view>
            </view>

            <!-- 卡片底栏 -->
            <view class="card-bottom">
              <view class="price-box">
                <text class="price-symbol">¥</text>
                <text class="price-num">{{ formatPrice(firstPrice(item)) }}</text>
                <text v-if="origPrice(item)" class="price-orig">¥{{ origPrice(item) }}</text>
                <text class="price-suffix">/人起</text>
              </view>
              <view
                class="cta-btn"
                :class="ctaCls(item)"
                :disabled="statusBtn(item).type === 'disabled'"
                hover-class="cta-press"
                :hover-stay-time="100"
                @click.stop="goEnroll(item)"
              >
                <text class="cta-text">{{ ctaText(item) }}</text>
              </view>
            </view>
          </view>

          <view v-if="filteredList.length > 0 && !hasMore" class="load-more-wrap">
            <text class="no-more">没有更多了</text>
          </view>
          <view style="height:48rpx" />
        </scroll-view>
      </StateView>
    </view>

    <!-- ⑤ 自定义 Toast -->
    <view v-if="toast.show" class="custom-toast" :class="{ 'custom-toast--out': toast.hide }">
      <view class="toast-icon"><view class="toast-check" /></view>
      <text class="toast-text">{{ toast.msg }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

/* ===== 状态 ===== */
const keyword = ref('')
const searchFocus = ref(false)
const selectedDistrict = ref('')   // 空 = 全部区县
const districtIndex = ref(0)
const activeCertType = ref('all')  // 证书类型筛选（all=全部）
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const allList = ref([])   // 全量（前端过滤）
const imgLoaded = ref({})
const hasMore = ref(false)

/* Toast */
const toast = ref({ show: false, hide: false, msg: '' })
let toastTimer = null
let toastOutTimer = null

/* ===== 证书类型 Tab ===== */
const certPills = [
  { label: '全部', value: 'all' },
  { label: 'CAAC', value: 'caac' },
  { label: 'UTC', value: 'utc_dji' },
  { label: '人社等级', value: 'gov_level' },
]

function certCount(value) {
  if (value === 'all') return allList.value.length || 0
  return allList.value.filter(function (c) { return (c.cert_type || '') === value }).length
}

/* ===== 前端过滤（Tab / 区县 / 关键词均在前端，不动后端） ===== */
const filteredList = computed(function () {
  var base = allList.value.slice()
  if (activeCertType.value !== 'all') {
    base = base.filter(function (c) { return (c.cert_type || '') === activeCertType.value })
  }
  if (selectedDistrict.value) {
    base = base.filter(function (c) { return c.district === selectedDistrict.value })
  }
  var kw = keyword.value.trim()
  if (kw) {
    base = base.filter(function (c) {
      return (c.title || '').indexOf(kw) !== -1 || (c.org_name || '').indexOf(kw) !== -1
    })
  }
  sortList(base)
  return base
})

/* ===== 排序（前端，ActionSheet 选择） ===== */
const sortKey = ref('hot')
const SORTERS = {
  hot: function (a, b) { return reviewNum(b) - reviewNum(a) },      // 热度（评价数）
  priceAsc: function (a, b) { return firstPrice(a) - firstPrice(b) }, // 价格低→高
  priceDesc: function (a, b) { return firstPrice(b) - firstPrice(a) }, // 价格高→低
  rating: function (a, b) { return (Number(b.rating) || 0) - (Number(a.rating) || 0) }, // 评分
}
const SORT_LABELS = {
  hot: '按热度推荐',
  priceAsc: '价格从低到高',
  priceDesc: '价格从高到低',
  rating: '评分最高',
}
const sortLabel = computed(function () { return SORT_LABELS[sortKey.value] || '按热度推荐' })

function sortList(arr) {
  var sorter = SORTERS[sortKey.value]
  if (sorter) arr.sort(sorter)
}
function reviewNum(item) {
  return Number(item.review_count || item.reviewCount || 0)
}

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

/* ===== 地区筛选 ===== */
function onDistrictChange(e) {
  const idx = Number(e.detail.value)
  districtIndex.value = idx
  selectedDistrict.value = chongqingDistricts[idx]
}

/* ===== 证书类型筛选 ===== */
function selectCertType(value) {
  if (activeCertType.value === value) return
  activeCertType.value = value
}

/* ===== 数据获取（一次拉全量，前端过滤） ===== */
async function fetchList(reset) {
  if (reset === undefined) reset = true
  if (reset) loading.value = true
  else loadingMore.value = true
  errorMsg.value = ''

  try {
    const params = { page: 1, page_size: 100 }
    if (activeCertType.value !== 'all') params.cert_type = activeCertType.value
    if (selectedDistrict.value) params.region = selectedDistrict.value
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/training-courses', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []

    if (items.length === 0) items = getMockCourses()
    allList.value = items
    hasMore.value = false
  } catch (e) {
    allList.value = getMockCourses()
    hasMore.value = false
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

/* ===== 本地 mock（API 空/失败时降级展示） ===== */
function getMockCourses() {
  return [
    {
      id: 'course-mock-1', title: 'CAAC民航局多旋翼无人机驾驶员执照班', cert_type: 'caac', status: 'recruiting',
      org_name: '重庆无人机飞行学院', district: '渝北区', price_fen: 980000, original_fee: 1280000,
      duration_days: 25, tags: ['CAAC', '多旋翼'], rating: '5.0', review_count: 128,
      image: '/static/home/hero-inspection.jpg',
    },
    {
      id: 'course-mock-2', title: '大疆UTC航拍工程师认证班', cert_type: 'utc_dji', status: 'urgent',
      org_name: '大疆慧飞重庆分校', district: '南岸区', price_fen: 39900, original_fee: 59900,
      duration_days: 7, tags: ['UTC', '航拍'], rating: '4.8', review_count: 89, remain: 3,
      image: '/static/home/demand-lift.jpg',
    },
    {
      id: 'course-mock-3', title: '人社职业技能等级证书·无人机装调检修', cert_type: 'gov_level', status: 'full',
      org_name: '重庆职业技能培训中心', district: '江北区', price_fen: 268000,
      duration_days: 15, tags: ['人社', '装调检修'], rating: '4.6', review_count: 67,
      image: '',
    },
    {
      id: 'course-mock-4', title: 'AOPA多旋翼机长执照班', cert_type: 'aopa', status: 'upcoming',
      org_name: '重庆空域无人机培训基地', district: '渝北区', price_fen: 1280000,
      duration_days: 30, tags: ['AOPA', '机长'], rating: '4.9', review_count: 45,
      image: '/static/home/demand-solar.jpg',
    },
    {
      id: 'course-mock-5', title: '无人机应急应用与植保作业实战班', cert_type: 'utc_dji', status: 'recruiting',
      org_name: '重庆农用无人机服务中心', district: '綦江区', price_fen: 45800,
      duration_days: 5, tags: ['UTC', '植保'], rating: '4.7', review_count: 158,
      image: '/static/home/home-bg.jpg',
    },
  ]
}

/* ===== 搜索 ===== */
var searchTimer = null
function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { /* computed 实时过滤 */ }, 0)
}
function onSearch() { clearTimeout(searchTimer) }
function clearKeyword() { keyword.value = '' }

/* ===== 数据映射 ===== */
function orgName(item) {
  return item.org_name || item.enterprise_name || item.name || item.title || '未知机构'
}
function certTypeLabel(certType) {
  const map = { caac: 'CAAC', utc_dji: 'UTC', gov_level: '人社等级', aopa: 'AOPA' }
  return map[certType] || '培训'
}
function certColorType(certType) {
  const map = { caac: 'blue', utc_dji: 'purple', gov_level: 'gold', aopa: 'teal' }
  return map[certType] || 'blue'
}
function coverOf(item) {
  const u = item.image || item.cover_image || item.image_url
  return u ? u : ''
}
function certChar(certType) {
  const map = { caac: 'CA', utc_dji: 'UT', aopa: 'AO', gov_level: '人社' }
  return map[certType] || '培'
}
function cardTags(item) {
  const tags = []
  if (item.cert_type) tags.push(certTypeLabel(item.cert_type))
  if (item.district) tags.push(item.district)
  else if (item.region) tags.push(item.region)
  if (tags.length === 0) tags.push('专业培训')
  return tags.slice(0, 3)
}
function pillCls(tag) {
  if (tag === 'CAAC') return 'pill--caac'
  if (tag === 'UTC') return 'pill--utc'
  if (tag === '人社' || tag === '人社等级') return 'pill--gov'
  return 'pill--gray'
}
function firstPrice(item) {
  if (item.price != null) return item.price
  if (item.price_fen != null) return item.price_fen / 100
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    const c = item.courses[0]
    return c.price != null ? c.price : (c.price_fen ? c.price_fen / 100 : 5800)
  }
  return 5800
}
function origPrice(item) {
  var o = item.original_fee
  if (o == null && item.original_price != null) o = item.original_price
  return o && o > firstPrice(item) ? o : null
}
function formatPrice(n) {
  return Number(n).toLocaleString()
}
function shortRegion(item) {
  if (item.district) return '重庆市 · ' + item.district
  if (item.province && item.city) return item.province + ' ' + item.city
  if (item.province || item.city) return item.province || item.city
  if (item.region) return item.region
  const loc = item.location || ''
  const match = loc.match(/^([一-龥]+省)?([一-龥]+市)?/)
  if (match && match[0]) return match[0]
  return loc.length > 8 ? loc.substring(0, 8) + '...' : loc
}
function reviewCount(item) {
  return item.review_count || item.reviewCount || 0
}
function remainCount(item) {
  if (item.remain != null) return item.remain
  if (item.max_students != null && item.enrolled_count != null) {
    return item.max_students - item.enrolled_count
  }
  return 0
}
function statusBtn(item) {
  var s = item.status
  if (s === 'full') return { type: 'disabled', text: '已满' }
  if (s === 'upcoming') return { type: 'disabled', text: '即将开课' }
  var remain = remainCount(item)
  if (s === 'urgent' || remain > 0) return { type: 'urgent', count: remain || 3, text: '仅剩 ' + (remain || 3) + ' 个' }
  return { type: 'enroll', text: '招生中' }
}
function ctaText(item) {
  var b = statusBtn(item)
  if (b.type === 'urgent') return '抢占名额'
  if (b.type === 'disabled') return '已满'
  return '立即报名'
}
function ctaCls(item) {
  var b = statusBtn(item)
  if (b.type === 'urgent') return 'cta--urgent'
  if (b.type === 'disabled') return 'cta--disabled'
  return 'cta--blue'
}

/* ===== 交互 ===== */
function goEnroll(item) {
  uni.navigateTo({ url: '/pkg-talent/pages/training/enroll?id=' + encodeURIComponent(item.id) })
}
function onImgLoad(id) { imgLoaded.value[id] = true }
function showSortSheet() {
  var labels = ['按热度推荐', '价格从低到高', '价格从高到低', '评分最高']
  uni.showActionSheet({
    itemList: labels,
    success: function (res) {
      var keys = ['hot', 'priceAsc', 'priceDesc', 'rating']
      sortKey.value = keys[res.tapIndex] || 'hot'
    },
  })
}
function goBack() { uni.navigateBack({ delta: 1 }) }

function showCustomToast(msg) {
  clearTimeout(toastTimer)
  clearTimeout(toastOutTimer)
  toast.value = { show: true, hide: false, msg: msg }
  toastTimer = setTimeout(function () {
    toast.value.hide = true
    toastOutTimer = setTimeout(function () {
      toast.value.show = false
    }, 200)
  }, 2000)
}

onLoad(function () { fetchList(true) })

onPullDownRefresh(function () {
  fetchList(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
  min-height: 100vh;
  background: #F4F6F8;
  display: flex;
  flex-direction: column;
  padding-left: constant(safe-area-inset-left);
  padding-left: env(safe-area-inset-left);
  padding-right: constant(safe-area-inset-right);
  padding-right: env(safe-area-inset-right);
  overflow-x: hidden;
}

/* ═══ ① 自定义导航栏 ═══ */
.nav-wrap {
  background: #ffffff;
  padding: var(--status-bar-height) 0 0;
  position: relative;
  z-index: 5;
  border-bottom: 1rpx solid #EEF1F4;
}
.nav-bar {
  display: flex;
  align-items: center;
  height: 88rpx;
  padding: 0 24rpx;
}
.nav-back {
  width: 64rpx;
  height: 64rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.nav-press { background: #F4F6F8; }
.nav-back-icon { font-size: 40rpx; color: #344054; font-weight: 300; line-height: 1; }
.nav-title { flex: 1; text-align: center; font-size: 34rpx; font-weight: 760; color: #17212B; }
.nav-capsule {
  width: 176rpx;
  height: 60rpx;
  border: 1rpx solid #E4E7EC;
  border-radius: 999rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  flex-shrink: 0;
}
.capsule-dot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #98A2B3; }
.capsule-divider { width: 1rpx; height: 28rpx; background: #E4E7EC; }
.capsule-arrow {
  width: 0; height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 8rpx solid #344054;
}
.nav-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24rpx 16rpx;
}
.meta-sync { display: flex; align-items: center; gap: 8rpx; }
.sync-dot {
  width: 10rpx; height: 10rpx;
  border-radius: 50%;
  background: #168A55;
  box-shadow: 0 0 0 0 rgba(22, 138, 85, 0.5);
  animation: syncPulse 1.8s ease-out infinite;
}
.sync-text { font-size: 20rpx; color: #667085; }
.meta-cert { display: flex; align-items: center; gap: 6rpx; }
.meta-cert-text { font-size: 22rpx; font-weight: 600; color: #344054; }
.meta-cert-arrow {
  width: 0; height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 7rpx solid #98A2B3;
}

/* ═══ ② 工具栏 ═══ */
.toolbar { padding: 16rpx 24rpx 0; }
.search-box {
  height: 80rpx;
  background: #F4F6F8;
  border: 1rpx solid #EDF0F5;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  padding: 0 20rpx;
  gap: 10rpx;
  transition: background 200ms var(--ease), border-color 200ms var(--ease), box-shadow 200ms var(--ease);
}
.search-box.focus {
  background: #ffffff;
  border-color: #0A66C2;
  box-shadow: 0 0 0 3px rgba(10, 102, 194, 0.12);
}
.search-ico { position: relative; width: 26rpx; height: 26rpx; flex-shrink: 0; }
.search-ring {
  width: 16rpx; height: 16rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50%;
  transition: border-color 200ms var(--ease);
}
.search-bar-ico {
  position: absolute;
  right: 0; bottom: 2rpx;
  width: 8rpx; height: 3rpx;
  background: #98A2B3;
  transform: rotate(45deg);
  transform-origin: right center;
  transition: background 200ms var(--ease);
}
.search-box.focus .search-ring { border-color: #0A66C2; }
.search-box.focus .search-bar-ico { background: #0A66C2; }
.search-input { flex: 1; font-size: 26rpx; color: #17212B; }
.search-ph { color: #98A2B3; }
.search-kbd {
  font-size: 18rpx;
  color: #98A2B3;
  border: 1rpx solid #E4E7EC;
  border-radius: 4rpx;
  padding: 2rpx 10rpx;
  transition: opacity 200ms var(--ease);
}
.search-box.focus .search-kbd { opacity: 0; }
.search-clear {
  width: 40rpx; height: 40rpx;
  border-radius: 50%;
  background: #EDF0F5;
  display: flex;
  align-items: center;
  justify-content: center;
}
.search-clear-x { font-size: 28rpx; color: #667085; line-height: 1; }

/* 城市卡 */
.city-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12rpx;
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 8rpx;
  padding: 20rpx;
}
.city-left { display: flex; align-items: center; gap: 10rpx; }
.city-ico {
  width: 44rpx; height: 44rpx;
  border-radius: 50%;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.loc-pin {
  width: 16rpx; height: 16rpx;
  border: 2rpx solid #0A66C2;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  box-sizing: border-box;
  position: relative;
}
.loc-pin::after {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  width: 4rpx; height: 4rpx;
  margin: -2rpx 0 0 -2rpx;
  border-radius: 50%;
  background: #0A66C2;
}
.city-copy { display: flex; align-items: center; gap: 8rpx; }
.city-name { font-size: 28rpx; font-weight: 720; color: #17212B; }
.city-sub { font-size: 22rpx; color: #98A2B3; }
.city-arrow {
  width: 0; height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 7rpx solid #98A2B3;
}
.city-count { font-size: 22rpx; color: #667085; }
.count-num { color: #0A66C2; font-weight: 720; }

/* ═══ ③ 证书类型 Tab ═══ */
.cert-tabs {
  white-space: nowrap;
  padding: 16rpx 0 0;
  position: relative;
}
.cert-tabs ::-webkit-scrollbar { display: none; width: 0; height: 0; }
.cert-tabs-inner {
  display: inline-flex;
  gap: 12rpx;
  padding: 0 24rpx;
}
.cert-tab {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  height: 60rpx;
  padding: 0 20rpx;
  border-radius: 6rpx;
  border: 1rpx solid #E4E7EC;
  background: #ffffff;
  color: #344054;
  flex-shrink: 0;
  transition: background 180ms var(--ease), color 180ms var(--ease), box-shadow 180ms var(--ease), border-color 180ms var(--ease);
}
.cert-tab.on {
  background: #0A66C2;
  color: #ffffff;
  border-color: #0A66C2;
  box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28);
}
.cert-tab-label { font-size: 24rpx; font-weight: 600; }
.cert-tab-count {
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  border-radius: 999rpx;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 18rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-tab-count--on { background: rgba(255, 255, 255, 0.28); color: #ffffff; }

/* Tab 线性图标（证件 / 无人机 / 圆加号） */
.cert-tab-ico { width: 26rpx; height: 26rpx; flex-shrink: 0; position: relative; }
.cti--all {
  border: 2rpx solid currentColor;
  border-radius: 4rpx;
  box-sizing: border-box;
}
.cti--caac {
  border: 2rpx solid currentColor;
  border-radius: 4rpx;
  box-sizing: border-box;
}
.cti--caac::before {
  content: '';
  position: absolute;
  left: 4rpx; right: 4rpx; top: 7rpx;
  height: 2rpx;
  background: currentColor;
}
.cti--utc_dji::before {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  width: 12rpx; height: 12rpx;
  margin: -6rpx 0 0 -6rpx;
  border: 2rpx solid currentColor;
  border-radius: 3rpx;
  box-sizing: border-box;
}
.cti--utc_dji::after {
  content: '';
  position: absolute;
  left: 0; top: 50%;
  width: 8rpx; height: 8rpx;
  margin-top: -4rpx;
  border: 2rpx solid currentColor;
  border-radius: 50%;
  box-sizing: border-box;
}
.cti--gov_level {
  border: 2rpx solid currentColor;
  border-radius: 50%;
  box-sizing: border-box;
}
.cti--gov_level::before,
.cti--gov_level::after {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  background: currentColor;
}
.cti--gov_level::before { width: 10rpx; height: 2rpx; margin: -1rpx 0 0 -5rpx; }
.cti--gov_level::after { width: 2rpx; height: 10rpx; margin: -5rpx 0 0 -1rpx; }

/* ═══ ④ 列表区 ═══ */
.content { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.state-fill { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.list-scroll { padding: 2rpx 24rpx 0; flex: 1; min-height: 0; box-sizing: border-box; }

.sub-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 4rpx 8rpx;
}
.sub-text { font-size: 22rpx; color: #667085; }
.sort-pill { display: flex; align-items: center; gap: 6rpx; padding: 8rpx 12rpx; border-radius: 6rpx; }
.sort-press { background: #F4F6F8; }
.sort-text { font-size: 22rpx; font-weight: 600; color: #344054; }
.sort-arrow {
  width: 0; height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 7rpx solid #98A2B3;
}

/* 课程卡片 */
.card {
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 10px;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  overflow: hidden;
  margin-bottom: 24rpx;
  animation: cardIn 460ms var(--ease) both;
  transition: transform 180ms var(--ease), box-shadow 180ms var(--ease);
}
.card-press {
  transform: scale(0.985);
  box-shadow: 0 6px 18px rgba(16, 24, 40, 0.08);
}

.banner { position: relative; width: 100%; height: 0; padding-bottom: 56.25%; overflow: hidden; }
.banner-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity 240ms ease-out;
}
.banner-fallback {
  position: absolute;
  inset: 0;
  background: linear-gradient(160deg, #0a5897 0%, #074D92 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-char {
  font-size: 64rpx;
  font-weight: 760;
  color: rgba(255, 255, 255, 0.9);
}
.banner-mask {
  position: absolute;
  left: 0; right: 0; bottom: 0;
  height: 40%;
  background: linear-gradient(180deg, rgba(7, 77, 146, 0) 30%, rgba(7, 77, 146, 0.55) 100%);
  pointer-events: none;
}

/* 证书类型徽章（左上） */
.cert-badge {
  position: absolute;
  top: 12rpx;
  left: 12rpx;
  padding: 6rpx 14rpx;
  border-radius: 6rpx;
  background: rgba(255, 255, 255, 0.92);
}
.cert-badge-text { font-size: 20rpx; font-weight: 700; }
.cb--blue .cert-badge-text { color: #0A66C2; }
.cb--purple .cert-badge-text { color: #7C3AED; }
.cb--gold .cert-badge-text { color: #B54708; }
.cb--teal .cert-badge-text { color: #0A66C2; }

/* 状态徽章（右上） */
.status-badge {
  position: absolute;
  top: 12rpx;
  right: 12rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 6rpx;
}
.status-text { font-size: 20rpx; font-weight: 600; color: #ffffff; }
.sb--enroll { background: #F97316; }
.sb--enroll .status-dot {
  width: 10rpx; height: 10rpx;
  border-radius: 50%;
  background: #ffffff;
  position: relative;
}
.sb--enroll .status-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: #ffffff;
  animation: badgeRing 1.4s ease-out infinite;
}
.sb--urgent { background: #D92D20; }
.sb--disabled { background: rgba(16, 24, 40, 0.65); }

/* 评价信息（左下） */
.review-row {
  position: absolute;
  left: 12rpx;
  bottom: 10rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
}
.rev-avatars { display: flex; }
.rev-avatar {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  box-sizing: border-box;
}
.rev-avatar + .rev-avatar { margin-left: -10rpx; }
.ra--1 { background: #0A66C2; }
.ra--2 { background: #F97316; }
.ra--3 { background: #168A55; }
.rev-count {
  font-size: 18rpx;
  font-weight: 500;
  color: #ffffff;
  padding: 2rpx 10rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.24);
  text-shadow: 0 1rpx 2rpx rgba(0, 0, 0, 0.3);
}

/* 卡片内容 */
.card-info { padding: 20rpx 28rpx 4rpx; }
.title-row { display: flex; align-items: center; justify-content: space-between; gap: 8rpx; }
.org-name {
  font-size: 30rpx;
  font-weight: 720;
  color: #17212B;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rating-badge {
  display: inline-flex;
  align-items: center;
  gap: 4rpx;
  padding: 4rpx 10rpx;
  border-radius: 6rpx;
  background: #FFF0E6;
  flex-shrink: 0;
}
.rating-star { font-size: 18rpx; color: #E96012; }
.rating-num { font-size: 20rpx; font-weight: 700; color: #E96012; }
.card-subtitle {
  display: block;
  font-size: 22rpx;
  color: #667085;
  margin-top: 8rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-meta {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-top: 10rpx;
}
.meta-ico { width: 22rpx; height: 22rpx; flex-shrink: 0; position: relative; }
.meta-ico .loc-pin { width: 12rpx; height: 12rpx; position: absolute; left: 50%; top: 2rpx; margin-left: -6rpx; }
.meta-text { font-size: 22rpx; color: #667085; }
.card-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 12rpx; }
.pill {
  padding: 4rpx 12rpx;
  border-radius: 4rpx;
  font-size: 20rpx;
  font-weight: 600;
}
.pill--caac { background: #EAF3FB; color: #0A66C2; }
.pill--utc { background: #F3F0FF; color: #7C3AED; }
.pill--gov { background: #FEF6E7; color: #B54708; }
.pill--gray { background: #F4F6F8; color: #667085; }

/* 卡片底栏 */
.card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 28rpx 20rpx;
}
.price-box { display: flex; align-items: baseline; }
.price-symbol { font-size: 22rpx; font-weight: 700; color: #E96012; }
.price-num {
  font-size: 44rpx;
  font-weight: 760;
  color: #E96012;
  line-height: 1;
  animation: priceIn 500ms var(--ease) both;
}
.price-orig {
  font-size: 20rpx;
  color: #98A2B3;
  margin-left: 8rpx;
  text-decoration: line-through;
}
.price-suffix { font-size: 20rpx; color: #98A2B3; margin-left: 4rpx; }

.cta-btn {
  height: 64rpx;
  padding: 0 28rpx;
  border-radius: 6rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 180ms var(--ease), background 180ms var(--ease);
}
.cta-press { transform: scale(0.96); }
.cta-text { font-size: 24rpx; font-weight: 700; color: #ffffff; }
.cta--blue { background: #0A66C2; box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.24); }
.cta--blue:active { background: #074D92; }
.cta--urgent { background: #F97316; box-shadow: 0 4rpx 10rpx rgba(249, 115, 22, 0.32); }
.cta--urgent:active { background: #E96012; }
.cta--disabled { background: #EEF1F4; opacity: 0.5; pointer-events: none; }
.cta--disabled .cta-text { color: #667085; }

.load-more-wrap { text-align: center; padding: 24rpx 0; }
.no-more { font-size: 22rpx; color: #98A2B3; }

/* ═══ ⑤ 自定义 Toast ═══ */
.custom-toast {
  position: fixed;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  z-index: 999;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 32rpx;
  background: rgba(16, 24, 40, 0.92);
  border-radius: 10rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.24);
  animation: toastIn 250ms var(--ease) both;
  max-width: 70vw;
}
.custom-toast--out { animation: toastOut 200ms ease both; }
.toast-icon {
  width: 32rpx; height: 32rpx;
  border-radius: 50%;
  background: rgba(91, 255, 176, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.toast-check {
  width: 16rpx; height: 9rpx;
  border-left: 3rpx solid #5BFFB0;
  border-bottom: 3rpx solid #5BFFB0;
  transform: rotate(-45deg) translate(1rpx, -1rpx);
}
.toast-text { font-size: 26rpx; color: #ffffff; font-weight: 500; line-height: 1.4; }

/* ═══ 动画 ═══ */
@keyframes cardIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes badgeRing {
  0% { transform: scale(1); opacity: 0.8; }
  80% { transform: scale(2.4); opacity: 0; }
  100% { transform: scale(2.4); opacity: 0; }
}
@keyframes syncPulse {
  0% { box-shadow: 0 0 0 0 rgba(22, 138, 85, 0.5); }
  70% { box-shadow: 0 0 0 12rpx rgba(22, 138, 85, 0); }
  100% { box-shadow: 0 0 0 0 rgba(22, 138, 85, 0); }
}
@keyframes priceIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes toastIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% - 20rpx)); }
  to { opacity: 1; transform: translate(-50%, -50%); }
}
@keyframes toastOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* ═══ 减少动态效果支持 ═══ */
@media (prefers-reduced-motion: reduce) {
  .card,
  .sync-dot,
  .sb--enroll .status-dot::after,
  .price-num,
  .custom-toast {
    animation: none !important;
    transition: none !important;
  }
}
</style>
