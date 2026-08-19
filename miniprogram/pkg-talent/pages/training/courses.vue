<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- ① 搜索栏（吸顶，原生导航栏之下） -->
    <view class="sticky-head">
      <view class="search-container">
        <view class="b-search">
          <image class="b-search-ic" src="/static/home/icons/search.svg" mode="aspectFit" />
          <input
            class="search-input"
            v-model="keyword"
            placeholder="搜索机构或课程名称"
            placeholder-class="b-ph"
            confirm-type="search"
            @confirm="onSearch"
            @input="onSearchInput"
          />
          <view v-if="keyword" class="b-sclr" @click="clearKeyword"><text>×</text></view>
          <view class="b-sep"></view>
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- ② 位置筛选 + 证书类型（胶囊条一体） -->
      <view class="fbar">
        <picker
          mode="selector"
          :range="chongqingDistricts"
          :value="districtIndex"
          @change="onDistrictChange"
          class="district-picker"
        >
          <view class="fpill">
            <text class="fpv">{{ selectedDistrict || '全部区县' }}</text><text class="farr">▾</text>
          </view>
        </picker>
        <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
          <view class="filter-inner">
            <view
              v-for="pill in certPills"
              :key="pill.value"
              class="fpill"
              :class="{ on: activeCertType === pill.value }"
              @click="selectCertType(pill.value)"
            >
              <text class="fpv">{{ pill.label }}</text>
            </view>
          </view>
        </scroll-view>
      </view>
    </view>

    <!-- Banner 渐变卡 -->
    <view class="banner">
      <view class="banner-icon">训</view>
      <view class="banner-info">
        <text class="banner-title">技能提升，认证护航</text>
        <text class="banner-sub">CAAC / UTC 执照培训 · 权威机构授课</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <view class="ir">
        <text>共 <text class="irn">{{ list.length }}</text> 家机构</text>
        <text class="ir-hint">{{ activeCertType ? certTypeLabel(activeCertType) : '全部类型' }}</text>
      </view>

    <!-- ④ 机构列表（水平卡片） -->
    <!-- 骨架屏：首次加载时显示 3 条 shimmer 卡片 -->
    <view v-if="initialLoading" class="skeleton-list">
      <view class="skeleton-card" v-for="n in 3" :key="n">
        <view class="sk-cover"></view>
        <view class="sk-body">
          <view class="sk-line w60"></view>
          <view class="sk-line w40"></view>
          <view class="sk-line w80"></view>
        </view>
      </view>
    </view>

    <StateView
      v-else
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && list.length === 0"
      empty-text="暂无机构"
      @retry="fetchList"
    >
      <scroll-view
        class="content-list"
        scroll-y
        :show-scrollbar="false"
        @scrolltolower="loadMore"
      >
        <view class="list">
          <view
            v-for="item in list"
            :key="item.id"
            class="card"
            hover-class="press-feedback"
            :hover-stay-time="120"
            @click="goEnroll(item)"
          >
            <!-- 封面图区 -->
            <view class="card-cover" :class="'cover--' + certColorType(item.cert_type)">
              <!-- 真实图片 / 类型渐变兜底 -->
              <image
                v-if="coverOf(item)"
                :src="coverOf(item)"
                class="cover-img"
                mode="aspectFill"
                lazy-load
                @load="onImgLoad(item.id)"
                :style="{ opacity: imgLoaded[item.id] ? 1 : 0 }"
              />
              <view v-else class="cover-placeholder">
                <text class="cover-icon">{{ certChar(item.cert_type) }}</text>
              </view>

              <!-- 图片底部渐变蒙层 + 类型简称 -->
              <view class="cover-mask">
                <text class="cover-tag">{{ certTypeLabel(item.cert_type) }}</text>
              </view>

              <!-- 左上角类型胶囊 -->
              <view class="type-pill" :class="'pill--' + certColorType(item.cert_type)">
                <text>{{ certTypeLabel(item.cert_type) }}</text>
              </view>

              <!-- 右上角状态徽章（悬浮） -->
              <view class="status-badge" :class="statusClass[item.status]">
                <text v-if="statusBtn(item).type === 'urgent'">仅剩 {{ statusBtn(item).count }} 个</text>
                <text v-else>{{ statusText[item.status] }}</text>
              </view>
            </view>

            <!-- 信息区 -->
            <view class="card-info">
              <view class="info-top">
                <text class="card-title">{{ orgName(item) }}</text>
                <view class="rating-box">
                  <text class="rating-star">★</text>
                  <text class="rating-num">{{ item.rating || '5.0' }}</text>
                </view>
              </view>

              <text class="card-subtitle">{{ item.title || item.name || '未知课程' }}</text>

              <view class="card-meta">
                <text class="meta-text">{{ shortRegion(item) }}</text>
                <text v-if="item.review_count" class="meta-reviews">{{ item.review_count }} 人评价</text>
              </view>

              <view class="card-tags">
                <text class="tag" v-for="t in cardTags(item)" :key="t">{{ t }}</text>
              </view>

              <view class="card-bottom">
                <view class="price-box">
                  <text class="price-symbol">¥</text>
                  <text class="price-num">{{ formatPrice(firstPrice(item)) }}</text>
                  <text class="price-suffix">/人起</text>
                </view>
                <view v-if="statusBtn(item).type === 'enroll'" class="btn-primary" @click.stop="goEnroll(item)">立即报名</view>
                <view v-else-if="statusBtn(item).type === 'urgent'" class="btn-urgent" @click.stop="goEnroll(item)">仅剩 {{ statusBtn(item).count }} 个名额</view>
                <view v-else class="btn-disabled">{{ statusBtn(item).text }}</view>
              </view>
            </view>
          </view>

          <view v-if="list.length > 0" class="load-more-wrap">
            <view v-if="loadingMore" class="loading-inline">
              <u-loading size="12px" />
              <text>加载中...</text>
            </view>
            <text v-else-if="!hasMore" class="no-more">— 没有更多了 —</text>
          </view>

          <view style="height:24px" />
        </view>
      </scroll-view>
    </StateView>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPageScroll, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { useReduceMotion } from '../../../utils/motion'

// 减弱动效（无障碍）+ 回到顶部
const { noMotion, checkMotion } = useReduceMotion()
const showBt = ref(false)
const scrollToTop = () => {
  uni.pageScrollTo({ scrollTop: 0, duration: 200 })
}
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

// 页面滚动：回到顶部按钮浮现
onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})

/* ===== 状态 ===== */
const keyword = ref('')
const selectedDistrict = ref('')   // 空 = 全部区县
const districtIndex = ref(0)
const activeCertType = ref('all')  // 证书类型筛选（all=全部）
const loading = ref(false)
const initialLoading = ref(true)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const imgLoaded = ref({})
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

/* ===== 证书类型筛选 Pills ===== */
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

/* ===== 状态标签 ===== */
const statusText = { recruiting: '招生中', full: '已满', upcoming: '即将开课', urgent: '名额紧张' }
const statusClass = { recruiting: 'recruiting', full: 'full', upcoming: 'upcoming', urgent: 'urgent' }

/* ===== 地区筛选 ===== */
function onDistrictChange(e) {
  const idx = Number(e.detail.value)
  districtIndex.value = idx
  selectedDistrict.value = chongqingDistricts[idx]
  fetchList(true)
}

/* 证书类型筛选 */
function selectCertType(value) {
  if (activeCertType.value === value) return
  activeCertType.value = value
  fetchList(true)
}

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
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
    initialLoading.value = false
  }
}

function loadMore() {
  if (!loadingMore.value && hasMore.value) {
    page.value++
    fetchList(false)
  }
}

/* ===== 搜索 & 筛选 ===== */
var searchTimer = null
function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { fetchList(true) }, 400)
}

function onSearch() {
  fetchList(true)
}

function clearKeyword() {
  keyword.value = ''
  fetchList(true)
}

/* 图片加载完成后淡入 */
function onImgLoad(id) {
  imgLoaded.value[id] = true
}

/* ===== 数据映射 ===== */

/** 机构名（卡片主标题） */
function orgName(item) {
  return item.org_name || item.enterprise_name || item.name || item.title || '未知机构'
}

/** 证书类型 → 展示名 */
function certTypeLabel(certType) {
  const map = { caac: 'CAAC', utc_dji: 'UTC', gov_level: '人社等级', aopa: 'AOPA' }
  return map[certType] || '培训'
}

/** 证书类型 → 色块类型（CAAC=蓝、UTC=紫、人社=金） */
function certColorType(certType) {
  const map = { caac: 'blue', utc_dji: 'purple', gov_level: 'gold', aopa: 'teal' }
  return map[certType] || 'blue'
}

/** 封面图 URL：兼容 image / cover_image / image_url，空串视为无图走渐变兜底 */
function coverOf(item) {
  const u = item.image || item.cover_image || item.image_url
  return u ? u : ''
}

/** 卡片标签：cert_type 映射 + 地区，最多 2-3 个 */
function cardTags(item) {
  const tags = []
  if (item.cert_type) tags.push(certTypeLabel(item.cert_type))
  if (item.district) tags.push(item.district)
  else if (item.region) tags.push(item.region)
  if (tags.length === 0) tags.push('专业培训')
  return tags.slice(0, 3)
}

/** 缩略图图标字符（无 emoji） */
function certChar(certType) {
  const map = { caac: 'CA', utc_dji: 'UT', aopa: 'AO', gov_level: '人社' }
  return map[certType] || '培'
}

/** 首个价格（元，兼容 price_fen） */
function firstPrice(item) {
  if (item.price != null) return item.price
  if (item.price_fen != null) return item.price_fen / 100
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    const c = item.courses[0]
    return c.price != null ? c.price : (c.price_fen ? c.price_fen / 100 : 5800)
  }
  return 5800
}

function formatPrice(n) {
  return Number(n).toLocaleString()
}

/** 地区简称 */
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

/** 剩余名额：优先用后端 remain，否则用 max_students - enrolled_count 计算 */
function remainCount(item) {
  if (item.remain != null) return item.remain
  if (item.max_students != null && item.enrolled_count != null) {
    return item.max_students - item.enrolled_count
  }
  return 0
}

/** 状态按钮：enroll=立即报名 / urgent=仅剩N个 / disabled=已满/即将开课 */
function statusBtn(item) {
  var s = item.status
  if (s === 'full') return { type: 'disabled', text: '已报满' }
  if (s === 'upcoming') return { type: 'disabled', text: '即将开课' }
  var remain = remainCount(item)
  if (s === 'urgent' || remain > 0) return { type: 'urgent', count: remain || 3 }
  return { type: 'enroll' }
}

/* ===== 交互 ===== */
function goEnroll(item) {
  uni.navigateTo({ url: '/pkg-talent/pages/training/enroll?id=' + encodeURIComponent(item.id) })
}

/* ===== 生命周期 ===== */
onLoad(() => {
  checkMotion()
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
.page {
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.16, 1, 0.3, 1);
  min-height: 100vh;
  background: #fff;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ================================================================= */
/* ① 吸顶头部：搜索 + 筛选胶囊                                        */
/* ================================================================= */
.sticky-head {
  position: sticky;
  top: 0;
  z-index: 20;
  background: #fff;
  padding-bottom: 4px;
  box-shadow: 0 1px 0 rgba(16, 24, 40, 0.04);
}

.search-container {
  padding: 12px 12px 0;
}

.b-search {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #ffffff;
  border: 1px solid #EDF0F4;
  border-radius: 7px;
  padding: 9px 12px;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05);
}

.b-search-ic { width: 15px; height: 15px; flex-shrink: 0; }
.b-ph { color: #98A2B3; }

.search-input {
  flex: 1;
  font-size: 13px;
  color: #17212B;
  height: 18px;
}

.b-sclr { padding: 2px 4px; color: #98A2B3; font-size: 15px; }
.b-sep { width: 1px; height: 14px; background: #E5E8EC; }
.b-sbtn { font-size: 13px; color: #0A66C2; font-weight: 600; }

/* 筛选胶囊条 */
.fbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px 6px;
}
.district-picker { display: block; flex-shrink: 0; }
.filter-scroll { flex: 1; min-width: 0; }
.filter-inner { display: flex; gap: 8px; padding-right: 4px; }

.fpill {
  display: flex;
  align-items: center;
  gap: 3px;
  height: 30px;
  padding: 0 14px;
  border-radius: 999rpx;
  background: #fff;
  border: 1px solid #E5E8EC;
  font-size: 13px;
  color: #344054;
  white-space: nowrap;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.15s ease;
}
.fpill.on {
  border-color: #0A66C2;
  color: #0A66C2;
  font-weight: 600;
  background: #F4F8FC;
}
.fpill:active { transform: scale(0.95); opacity: 0.85; }
.fpv { line-height: 1; }
.farr { font-size: 10px; color: #98A2B3; }

/* Banner 渐变卡 */
.banner {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 8px 12px 0;
  padding: 14px 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #0A66C2, #074D92);
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
  overflow: hidden;
  animation: bannerIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) 60ms backwards;
}
.banner::after {
  content: '';
  position: absolute;
  top: -30px;
  right: -20px;
  width: 90px;
  height: 90px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
}
.banner-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 20px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  animation: fadeUp 0.25s ease-out backwards;
}
.banner-info { display: flex; flex-direction: column; gap: 3px; position: relative; z-index: 1; }
.banner-title { font-size: 16px; font-weight: 700; color: #fff; animation: fadeUp 0.25s ease-out 80ms backwards; }
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.85); animation: fadeUp 0.25s ease-out 140ms backwards; }

/* 白色板块：信息行 */
.section { padding: 12px 12px 0; }
.ir {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-radius: 8px;
  padding: 10px 14px;
  margin-bottom: 10px;
  font-size: 13px;
  color: #344054;
  box-shadow: 0 8px 28px rgba(16, 24, 40, 0.12);
  animation: fadeUp 0.25s ease-out backwards;
}
.irn { color: #0A66C2; font-weight: 700; }
.ir-hint { font-size: 12px; color: #98A2B3; }

/* 回到顶部 */
.bt {
  position: fixed;
  right: 16px;
  bottom: 60px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #344054;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.25s ease, transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 30;
}
.bt.show { opacity: 1; pointer-events: auto; }
.bt:active { transform: scale(0.92); }

/* 动效 */
@keyframes bannerIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
@keyframes fadeUp { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* no-motion：装饰动画关闭 */
.page.no-motion .banner,
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .ir,
.page.no-motion .card { animation: none; transition: opacity 0.15s ease; }
.page.no-motion .fpill { transition: none; }

/* ================================================================= */
/* ④ 机构列表（水平卡片）                                             */
/* ================================================================= */
.content-list {
  height: calc(100vh - 150px);
  box-sizing: border-box;
}

/* 骨架屏（封面图 + 信息区） */
.skeleton-list { padding: 0 12px; }

.skeleton-card {
  background: #ffffff;
  border: 1px solid #ebedf0;
  border-radius: 16rpx;
  overflow: hidden;
  margin-bottom: 16rpx;
}

.sk-cover {
  width: 100%;
  height: 240rpx;
  background: #f0f1f3;
  animation: shimmer 1.4s ease-in-out infinite;
  background-image: linear-gradient(90deg, #f0f1f3 25%, #f8fafc 50%, #f0f1f3 75%);
  background-size: 200% 100%;
}

.sk-body {
  padding: 14px 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sk-line {
  height: 12px;
  border-radius: 4px;
  background: #f0f1f3;
  animation: shimmer 1.4s ease-in-out infinite;
  background-image: linear-gradient(90deg, #f0f1f3 25%, #f8fafc 50%, #f0f1f3 75%);
  background-size: 200% 100%;
}

.sk-line.w60 { width: 60%; }
.sk-line.w40 { width: 40%; }
.sk-line.w80 { width: 80%; }

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.list { padding: 0 12px 20px; box-sizing: border-box; }

/* 卡片：纵向封面图 + 信息区（白底圆角 + 柔和阴影 + 错峰入场） */
.card {
  background: #ffffff;
  border: 1px solid #EDF0F4;
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 14px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  box-sizing: border-box;
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.15s ease;
  animation: cardIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}

.card:nth-child(1) { animation-delay: 60ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 140ms; }
.card:nth-child(4) { animation-delay: 180ms; }
.card:nth-child(5) { animation-delay: 220ms; }
.card:nth-child(6) { animation-delay: 260ms; }

/* ===== 封面图区 ===== */
.card-cover {
  position: relative;
  width: 100%;
  height: 240rpx;
  overflow: hidden;
}

/* 类型渐变兜底底色（图片缺失/加载失败时可见） */
.cover--blue { background: linear-gradient(135deg, #074D92, #0A66C2); }
.cover--purple { background: linear-gradient(135deg, #6D28D9, #DB2777); }
.cover--gold { background: linear-gradient(135deg, #D97706, #FB923C); }
.cover--teal { background: linear-gradient(135deg, #065F46, #06B6D4); }

.cover-img {
  width: 100%;
  height: 100%;
  display: block;
  transition: opacity 0.2s ease;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-icon { font-size: 48px; font-weight: 700; color: rgba(255, 255, 255, 0.9); }

/* 底部渐变蒙层 + 类型简称 */
.cover-mask {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 72rpx;
  padding: 0 16rpx;
  display: flex;
  align-items: flex-end;
  justify-content: flex-start;
  background: linear-gradient(180deg, rgba(10, 31, 68, 0) 0%, rgba(10, 31, 68, 0.55) 100%);
  pointer-events: none;
}

.cover-tag {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
  padding-bottom: 8rpx;
}

/* 左上角类型胶囊 */
.type-pill {
  position: absolute;
  top: 10rpx;
  left: 10rpx;
  padding: 2px 10px;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.92);
  font-size: 11px;
  font-weight: 600;
}

.pill--blue { color: #0A66C2; }
.pill--purple { color: #6D28D9; }
.pill--gold { color: #D97706; }
.pill--teal { color: #065F46; }

/* 右上角状态徽章（悬浮） */
.status-badge {
  position: absolute;
  top: 10rpx;
  right: 10rpx;
  padding: 3px 8px;
  border-radius: 999rpx;
  font-size: 10px;
  font-weight: 600;
  line-height: 1.2;
  max-width: 140rpx;
  box-sizing: border-box;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #ffffff;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.2);
}

.status-badge.recruiting { background: #F97316; }
.status-badge.urgent { background: #EF4444; }
.status-badge.full { background: #98A2B3; }
.status-badge.upcoming { background: #0A66C2; }

/* ===== 信息区 ===== */
.card-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 14px 12px;
}

.info-top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.rating-box { display: flex; align-items: center; gap: 2px; flex-shrink: 0; }
.rating-star { font-size: 13px; color: #FFB300; }
.rating-num { font-size: 12px; font-weight: 600; color: #17212B; }

.card-subtitle {
  font-size: 12px;
  font-weight: 500;
  color: #969799;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.card-meta { display: flex; align-items: center; gap: 8px; }
.meta-text { font-size: 11px; color: #969799; }
.meta-reviews { font-size: 10px; color: #c8c9cc; }

.card-tags { display: flex; gap: 4px; flex-wrap: nowrap; overflow: hidden; }

.tag {
  font-size: 9px;
  color: #0A66C2;
  background: rgba(10, 102, 194, 0.08);
  border: 1rpx solid rgba(10, 102, 194, 0.2);
  padding: 2px 8px;
  border-radius: 999rpx;
  white-space: nowrap;
}

.card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 2px;
}

.price-box { display: flex; align-items: baseline; gap: 1px; }
.price-symbol { font-size: 10px; color: #E96012; font-weight: 600; }
.price-num { font-size: 16px; color: #E96012; font-weight: 800; line-height: 1; }
.price-suffix { font-size: 9px; color: #c8c9cc; margin-left: 2px; }

.btn-primary {
  padding: 5px 14px;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  color: #ffffff;
  font-size: 11px;
  font-weight: 500;
  border-radius: 50rpx;
  box-shadow: 0 2rpx 8rpx rgba(10, 102, 194, 0.25);
  transition: transform var(--anim-fast) ease;
  animation: badgePulse 2s ease-in-out infinite;
}

.btn-urgent {
  padding: 5px 14px;
  background: linear-gradient(135deg, #F97316, #E96012);
  color: #ffffff;
  font-size: 11px;
  font-weight: 600;
  border-radius: 50rpx;
}

.btn-disabled {
  padding: 5px 14px;
  background: #E8F2FC;
  color: #98A2B3;
  font-size: 11px;
  border-radius: 999rpx;
}

/* 加载更多 */
.load-more-wrap { text-align: center; padding: 10px 0; }
.no-more { font-size: 11px; color: #c8c9cc; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 6px; font-size: 11px; color: #c8c9cc; }

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes cardIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes badgePulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(10, 102, 194, 0.3); }
  50% { box-shadow: 0 0 0 6rpx rgba(10, 102, 194, 0); }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .card, .btn-primary {
    animation: none !important;
    transition: none !important;
  }
}
</style>
