<template>
  <Layout :current="1">
    <view class="hall-page">
      <!-- ═══════ 深蓝顶部 ═══════ -->
      <view class="topbar">
        <view class="topbar-row">
          <view class="city-btn" hover-class="tap-fade" @tap="openCity">
            <text class="city-label">{{ cityLabel }}</text>
            <text class="city-arrow">⌄</text>
          </view>
          <text class="top-title">供需大厅</text>
          <view class="icon-btn" hover-class="tap-fade" @tap="goMessages">
            <text class="icon-bell">◌</text>
          </view>
        </view>
        <view class="search-trigger" hover-class="tap-fade" @tap="goSearch">
          <view class="search-icon"></view>
          <text class="search-placeholder">搜索需求、服务或设备</text>
        </view>
      </view>

      <!-- ═══════ 一级 Tab：需求大厅 / 供给大厅 ═══════ -->
      <view class="tabs">
        <view
          class="tab"
          :class="{ active: primary === 'demand' }"
          @tap="switchPrimary('demand')"
        >需求大厅</view>
        <view
          class="tab"
          :class="{ active: primary === 'supply' }"
          @tap="switchPrimary('supply')"
        >供给大厅</view>
      </view>

      <!-- 供给分段：商品设备 / 服务能力 -->
      <view v-if="primary === 'supply'" class="subtabs">
        <view
          class="subtab"
          :class="{ active: supplyKind === 'product' }"
          @tap="switchSupplyKind('product')"
        >商品设备</view>
        <view
          class="subtab"
          :class="{ active: supplyKind === 'service' }"
          @tap="switchSupplyKind('service')"
        >服务能力</view>
      </view>

      <!-- ═══════ 分类横滑 + 筛选 ═══════ -->
      <view class="filter-row">
        <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
          <view class="filter-inner">
            <view
              v-for="c in categories"
              :key="c"
              class="filter-chip"
              :class="{ active: typeFilter === c }"
              @tap="switchTypeFilter(c)"
            >{{ c }}</view>
          </view>
        </scroll-view>
        <view v-if="!isProductMode" class="filter-chip filter" :class="{ on: hasActiveFilter }" @tap="openFilters">
          <text>筛选</text>
          <text class="filter-caret">▾</text>
        </view>
      </view>

      <!-- ═══════ 匹配条（商品模式为电商页，不展示匹配引导） ═══════ -->
      <view v-if="listState === 'ready' && visibleList.length > 0 && !isProductMode" class="match-strip">
        <view class="match-mark"><text class="match-mark-icon">⇄</text></view>
        <view class="match-copy">
          <text class="match-title">{{ primary === 'demand' ? '为需求方推荐承接能力' : '发现与你匹配的需求' }}</text>
          <text class="match-desc">依据业务类型、区域和服务能力进行匹配</text>
        </view>
        <view class="match-link" @tap="goMatches">查看匹配</view>
      </view>

      <!-- ═══════ 列表标题 ═══════ -->
      <view class="section-head">
        <text class="section-title">{{ sectionTitle }}</text>
        <text class="section-note">共 {{ visibleList.length }} 条</text>
      </view>

      <!-- ═══════ 列表四种状态 ═══════ -->
      <!-- 加载中 -->
      <view v-if="listState === 'loading'" class="skeleton-list">
        <view class="skeleton-card"></view>
        <view class="skeleton-card"></view>
        <view class="skeleton-card"></view>
      </view>

      <!-- 加载失败 -->
      <view v-else-if="listState === 'error'" class="state-panel">
        <view class="state-mark err">!</view>
        <text class="state-title">内容加载失败</text>
        <text class="state-desc">请检查网络后重新加载</text>
        <view class="state-btn" @tap="reload">重新加载</view>
      </view>

      <!-- 空数据 -->
      <view v-else-if="listState === 'empty' || visibleList.length === 0" class="state-panel">
        <view class="state-mark">⌁</view>
        <text class="state-title">没有找到匹配的{{ kindLabel }}</text>
        <text class="state-desc">换个筛选条件试试，或先发布一条信息</text>
        <view class="state-btn" @tap="resetFilters">清除筛选</view>
      </view>

      <!-- 正常列表 -->
      <!-- 商品模式：电商两列宫格（大图 + 价格 + 品牌型号 + 成色/浏览） -->
      <view v-if="isProductMode && visibleList.length > 0" class="ecom-grid">
        <view
          v-for="item in visibleList"
          :key="item.id"
          class="ecom-card"
          hover-class="tap-fade"
          @tap="goProductDetail(item)"
        >
          <view class="ecom-img-wrap">
            <image :src="item.image" mode="aspectFill" class="ecom-img" @error="onProductImgError(item)" />
            <text v-if="item.isUsed" class="ecom-used-tag">二手</text>
          </view>
          <view class="ecom-body">
            <text class="ecom-price">¥<text class="ecom-price-num">{{ item.price }}</text></text>
            <text class="ecom-title">{{ item.title }}</text>
            <text class="ecom-spec">{{ item.spec }}</text>
            <view class="ecom-foot">
              <text class="ecom-cat">{{ item.cat }}</text>
              <text class="ecom-views">{{ item.views ? '已浏览 ' + item.views + ' 次' : '平台商品' }}</text>
            </view>
          </view>
        </view>
      </view>
      <view v-else class="card-list">
        <view
          v-for="item in visibleList"
          :key="item.id"
          class="trade-card"
          hover-class="tap-fade"
          @tap="goDetail(item)"
        >
          <view class="trade-card-main">
            <image :src="item.image" mode="aspectFill" class="trade-visual" @error="onImageError(item)" />
            <view class="trade-body">
              <view class="tag-row">
                <text class="tag blue">{{ item.cat }}</text>
                <text class="tag" :class="statusTagClass(item)">{{ item.status }}</text>
              </view>
              <text class="trade-title">{{ item.title }}</text>
              <view class="trade-meta">
                <text class="meta-item">{{ item.region }}</text>
                <text class="meta-item">{{ item.time }}</text>
              </view>
            </view>
          </view>
          <view class="trade-footer">
            <view class="price-block">
              <text class="price">{{ item.price }}</text>
              <text class="price-unit"> {{ item.unit }}</text>
            </view>
            <view class="card-action">
              <text>查看详情</text>
              <text class="card-action-arrow">›</text>
            </view>
          </view>
        </view>
      </view>

    </view>

    <!-- ═══════ 筛选弹层 ═══════ -->
    <u-popup :show="showFilter" position="bottom" round @close="showFilter = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">筛选条件</text>
          <view class="sheet-close" @tap="showFilter = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body">
          <text class="field-label">所在地区</text>
          <view class="pick-row">
            <view
              v-for="r in regionOptions"
              :key="r"
              class="pick"
              :class="{ active: filterRegion === r }"
              @tap="filterRegion = r"
            >{{ r }}</view>
          </view>

          <text class="field-label">{{ primary === 'demand' ? '预算金额' : '价格范围' }}</text>
          <view class="pick-row">
            <view
              v-for="(r, i) in priceOptions"
              :key="i"
              class="pick"
              :class="{ active: filterPrice === r }"
              @tap="filterPrice = r"
            >{{ r }}</view>
          </view>

          <text class="field-label">排序方式</text>
          <view class="pick-row">
            <view
              v-for="s in sortOptions"
              :key="s.value"
              class="pick"
              :class="{ active: sortBy === s.value }"
              @tap="sortBy = s.value"
            >{{ s.label }}</view>
          </view>
        </view>
        <view class="sheet-footer">
          <view class="ghost-btn" @tap="resetFilters">重置</view>
          <view class="primary-btn" @tap="applyFilter">应用筛选</view>
        </view>
      </view>
    </u-popup>

    <!-- ═══════ 城市弹层 ═══════ -->
    <u-popup :show="showCity" position="bottom" round @close="showCity = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">选择城市</text>
          <view class="sheet-close" @tap="showCity = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body city-grid">
          <view
            v-for="r in cityOptions"
            :key="r.value"
            class="city-option"
            :class="{ active: city === r.value }"
            @tap="pickCity(r.value)"
          >
            <text>{{ r.label }}</text>
            <text v-if="city === r.value" class="city-check">✓</text>
          </view>
        </view>
      </view>
    </u-popup>

  </Layout>

  <!-- ═══════ 就地搜索覆盖层（点击搜索框在当前页展开，不跳搜索页） ═══════ -->
  <view v-if="showSearch" class="ov-overlay">
    <view class="ov-bar">
      <view class="ov-back" hover-class="tap-fade" hover-stay-time="120" @tap="closeSearch">
        <view class="ov-back-arrow"></view>
      </view>
      <view class="ov-search-box">
        <view class="ov-search-icon"></view>
        <input
          class="ov-search-input"
          v-model="keyword"
          placeholder="搜索需求、服务或设备"
          confirm-type="search"
          focus
          @confirm="onSearch"
        />
        <view v-if="keyword" class="ov-clear" @tap="keyword = ''"><text class="ov-clear-x">×</text></view>
      </view>
      <view class="ov-search-btn" hover-class="tap-fade" @tap="onSearch"><text>搜索</text></view>
    </view>

    <!-- 有结果 -->
    <view v-if="searched" class="ov-results">
      <view class="ov-result-head">找到 {{ searchResults.length }} 条相关内容</view>
      <view v-if="searchResults.length" class="ov-card-list">
        <view
          v-for="item in searchResults"
          :key="item.id"
          class="ov-trade-card"
          hover-class="tap-fade"
          @tap="goSearchResult(item)"
        >
          <view class="ov-trade-card-main">
            <image :src="item.image" mode="aspectFill" class="ov-trade-visual" @error="onImageError(item)" />
            <view class="ov-trade-body">
              <view class="ov-tag-row">
                <text class="ov-tag ov-tag-blue">{{ item.cat }}</text>
                <text class="ov-tag" :class="'ov-tag-' + statusTagClass(item)">{{ item.status }}</text>
              </view>
              <text class="ov-trade-title">{{ item.title }}</text>
              <view class="ov-trade-meta">
                <text>{{ item.region }}</text>
                <text>{{ item.time }}</text>
              </view>
            </view>
          </view>
          <view class="ov-trade-footer">
            <view class="ov-price-block">
              <text class="ov-price">{{ item.price }}</text>
              <text class="ov-price-unit"> {{ item.unit }}</text>
            </view>
            <view class="ov-card-action"><text>查看详情 ›</text></view>
          </view>
        </view>
      </view>
      <view v-else class="ov-state-panel">
        <view class="ov-state-mark">⌁</view>
        <text class="ov-state-title">没有找到相关内容</text>
        <text class="ov-state-desc">换个关键词试试</text>
      </view>
    </view>

    <!-- 推荐 / 最近搜索 -->
    <view v-else class="ov-suggest">
      <view class="ov-search-block">
        <text class="ov-block-title">推荐搜索</text>
        <view class="ov-keyword-row">
          <view v-for="w in hotWords" :key="w" class="ov-keyword" @tap="keyword = w; onSearch()">{{ w }}</view>
        </view>
      </view>
      <view class="ov-search-block">
        <text class="ov-block-title">最近搜索</text>
        <view class="ov-keyword-row">
          <view v-for="w in recentWords" :key="w" class="ov-keyword" @tap="keyword = w; onSearch()">{{ w }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request } from '../../utils/request'
import { safeNavigateTo } from '../../utils/nav'
import {
  HALL_CATEGORIES, getKindItems, kindTypeLabel, isEnded, normalizeDemand,
  IMG_SOLAR, IMG_LIFT, IMG_HERO,
  PRODUCT_CATEGORIES, normalizeProduct, normalizeService, getLocalLiveCards,
} from '../../utils/hallData'

const primary = ref('demand') // demand | supply
const supplyKind = ref('product') // product | service
const typeFilter = ref('全部')
const city = ref('全重庆')
const showCity = ref(false)

const cityLabel = computed(() => (city.value === '全重庆' ? '重庆' : city.value))

const ALL_REGIONS = ['不限', '渝北区', '江津区', '沙坪坝区', '南岸区', '涪陵区', '奉节县']
const regionOptions = ['不限', '渝北区', '江津区', '沙坪坝区', '南岸区', '涪陵区']
const priceOptions = ['不限', '1 万以下', '1-5 万', '5 万以上', '面议']
const sortOptions = [
  { label: '最新发布', value: 'newest' },
  { label: '匹配度优先', value: 'match' },
  { label: '价格优先', value: 'price' },
]
const cityOptions = [
  { value: '全重庆', label: '重庆' },
  ...regionOptions.filter((r) => r !== '不限').map((r) => ({ value: r, label: r })),
]

const showFilter = ref(false)
const filterRegion = ref('不限')
const filterPrice = ref('不限')
const sortBy = ref('newest')

const hasActiveFilter = computed(() => filterRegion.value !== '不限' || filterPrice.value !== '不限' || sortBy.value !== 'newest' || typeFilter.value !== '全部')

const categories = computed(() => {
  if (isProductMode.value) return PRODUCT_CATEGORIES
  const kind = primary.value === 'demand' ? 'demand' : supplyKind.value
  return HALL_CATEGORIES[kind]
})

const kindLabel = computed(() => kindTypeLabel(primary.value, supplyKind.value))
const sectionTitle = computed(() => {
  if (primary.value === 'demand') return '最新需求'
  return supplyKind.value === 'service' ? '可对接服务' : '优选商品设备'
})

// 商品设备模式：电商两列宫格展示（独立于需求/服务列表）
const isProductMode = computed(() => primary.value === 'supply' && supplyKind.value === 'product')

/* ================= 列表状态 ================= */
const listState = ref('loading') // loading | ready | empty | error
const list = ref([])

async function fetchList(showLoading = true) {
  if (showLoading) listState.value = 'loading'

  // 发布页已上架的本地内容并入列表（后端未接入期间的展示打通）
  const local = () => {
    if (primary.value === 'demand') return getLocalLiveCards('demand')
    return supplyKind.value === 'product' ? getLocalLiveCards('product') : getLocalLiveCards('service')
  }
  const merge = (remote) => {
    list.value = [...local(), ...remote]
    listState.value = list.value.length ? 'ready' : 'empty'
  }
  // 接口失败但本地有已上架发布时，降级展示本地内容（演示闭环）
  const fallback = () => {
    list.value = local()
    listState.value = list.value.length ? 'ready' : 'error'
  }

  if (primary.value === 'demand') {
    try {
      const res = await request({
        url: '/api/v1/demands',
        data: { page: 1, page_size: 20, sort: 'newest' },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      merge(items.map(normalizeDemand).filter(Boolean))
    } catch (e) {
      fallback()
    }
  } else if (supplyKind.value === 'product') {
    // 商品设备：电商模式走真实商品接口
    try {
      const res = await request({
        url: '/api/v1/products',
        data: { page: 1, page_size: 50 },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      merge(items.map(normalizeProduct).filter(Boolean))
    } catch (e) {
      fallback()
    }
  } else {
    // 服务能力：真实服务接口（/api/v1/service-listings 公开列表）
    try {
      const res = await request({
        url: '/api/v1/service-listings',
        data: { page: 1, page_size: 20 },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      merge(items.map(normalizeService).filter(Boolean))
    } catch (e) {
      fallback()
    }
  }
}

const reload = () => { fetchList(true) }

/* ================= 交互 ================= */
function switchPrimary(value) {
  if (primary.value === value) return
  primary.value = value
  typeFilter.value = '全部'
  fetchList(true)
}

function switchSupplyKind(value) {
  if (supplyKind.value === value) return
  supplyKind.value = value
  typeFilter.value = '全部'
  fetchList(true)
}

function switchTypeFilter(value) {
  typeFilter.value = value
}

const openFilters = () => { showFilter.value = true }
const applyFilter = () => {
  showFilter.value = false
  uni.showToast({ title: '筛选已应用', icon: 'none' })
}
const resetFilters = () => {
  typeFilter.value = '全部'
  filterRegion.value = '不限'
  filterPrice.value = '不限'
  sortBy.value = 'newest'
  showFilter.value = false
  fetchList(true)
}

const openCity = () => { showCity.value = true }
const pickCity = (value) => {
  city.value = value
  showCity.value = false
}

const goSearch = () => openSearch()

/* ================= 就地搜索（覆盖层内本地过滤，逻辑与 demands/search.vue 一致） ================= */
const showSearch = ref(false)
const keyword = ref('')
const searched = ref(false)
const searchResults = ref([])

const hotWords = ['巡检', '航拍', '测绘', '吊运', '设备租赁', '植保']
const recentWords = ['光伏巡检', 'M350']

const openSearch = () => {
  showSearch.value = true
}
const closeSearch = () => {
  showSearch.value = false
  keyword.value = ''
  searched.value = false
  searchResults.value = []
}

function searchAllItems() {
  return [...getKindItems('demand'), ...getKindItems('supply', 'service'), ...getKindItems('supply', 'product')]
}

function onSearch() {
  const kw = keyword.value.trim()
  searched.value = true
  if (!kw) {
    searchResults.value = []
    return
  }
  searchResults.value = searchAllItems().filter(
    (i) => i.title.includes(kw) || i.cat.includes(kw) || i.company.includes(kw)
  )
}

const goSearchResult = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
const goMessages = () => safeNavigateTo('/pages/messages/index')
const goMatches = () => safeNavigateTo('/pkg-demand/pages/demands/matches')
const goDetail = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
// 商品模式：跳电商商品详情页。
// 本地发布的商品已接后端（backendId 非空）→ 用后端 id 进真商品详情；仅旧版未接
// 后端的纯本地商品才走大厅详情页兜底展示。
const goProductDetail = (item) => {
  if (item.backendId) {
    safeNavigateTo('/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(item.backendId))
    return
  }
  if (String(item.id).indexOf('post-') === 0) {
    safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
    return
  }
  safeNavigateTo('/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(item.id))
}
const onProductImgError = (item) => {
  if (item.image !== IMG_HERO) item.image = IMG_HERO
}


/* ================= 展示辅助 ================= */
function statusTagClass(item) {
  if (isEnded(item)) return 'gray'
  if (item.type === '商品') return 'orange'
  return 'green'
}

function onImageError(item) {
  if (item.image !== IMG_LIFT && item.image !== IMG_SOLAR) {
    item.image = item.cat === '吊运' ? IMG_LIFT : IMG_SOLAR
  }
}

/* ================= 过滤展示 ================= */
const visibleList = computed(() => {
  let out = list.value
  if (typeFilter.value !== '全部') out = out.filter((i) => i.cat === typeFilter.value)
  if (filterRegion.value !== '不限' && !isProductMode.value) out = out.filter((i) => i.region.includes(filterRegion.value))
  return out
})

/* ================= 生命周期 ================= */
onLoad(() => {
  fetchList(true)
})

onPullDownRefresh(() => {
  fetchList(false).finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.hall-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 120rpx;
}

.tap-fade { opacity: 0.85; }

/* ═══════ 深蓝顶部 ═══════ */
.topbar {
  background: #074D92;
  color: #fff;
  padding: 16rpx 24rpx 28rpx;
  padding-top: calc(env(safe-area-inset-top) + 16rpx);
}
.topbar-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.city-btn {
  display: flex;
  align-items: center;
  gap: 4rpx;
  font-size: 26rpx;
  font-weight: 600;
  white-space: nowrap;
}
.city-arrow { font-size: 22rpx; }
.top-title {
  flex: 1;
  font-size: 38rpx;
  font-weight: 700;
  text-align: center;
}
.icon-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-bell { font-size: 40rpx; line-height: 1; }

/* 搜索框：居中、接近满宽、44px 高 */
.search-trigger {
  width: 100%;
  height: 44px;
  margin-top: 28rpx;
  border-radius: 7px;
  background: #fff;
  color: #98A2B3;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  padding: 0 24rpx;
  font-size: 26rpx;
  box-sizing: border-box;
}
.search-icon {
  width: 30rpx;
  height: 30rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.search-icon::after {
  content: '';
  position: absolute;
  right: -12rpx;
  bottom: -7rpx;
  width: 14rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #98A2B3;
  transform: rotate(45deg);
}

/* ═══════ 一级 Tab ═══════ */
.tabs {
  display: flex;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 0 32rpx;
}
.tab {
  flex: 1;
  height: 92rpx;
  line-height: 92rpx;
  text-align: center;
  position: relative;
  color: #667085;
  font-weight: 600;
  font-size: 28rpx;
}
.tab.active { color: #0A66C2; }
.tab.active::after {
  content: '';
  position: absolute;
  width: 56rpx;
  height: 6rpx;
  background: #0A66C2;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  border-radius: 3rpx;
}

/* 供给分段 */
.subtabs {
  display: flex;
  gap: 8rpx;
  margin: 20rpx 32rpx 16rpx;
  padding: 8rpx;
  background: #F2F5F8;
  border-radius: 16rpx;
}
.subtab {
  flex: 1;
  height: 68rpx;
  line-height: 68rpx;
  text-align: center;
  border-radius: 12rpx;
  color: #667085;
  font-size: 24rpx;
}
.subtab.active {
  color: #fff;
  background: #0A66C2;
  font-weight: 650;
}

/* ═══════ 分类横滑 ═══════ */
.filter-row {
  padding: 20rpx 24rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
}
.filter-scroll {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
}
.filter-inner {
  display: inline-flex;
  gap: 12rpx;
}
.filter-chip {
  display: inline-flex;
  align-items: center;
  height: 60rpx;
  padding: 0 20rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #344054;
  font-size: 24rpx;
  box-sizing: border-box;
  white-space: nowrap;
}
.filter-chip.active {
  color: #0A66C2;
  border-color: #B9D6EF;
  background: #EAF3FB;
}
.filter-chip.filter {
  flex-shrink: 0;
  color: #0A66C2;
  gap: 4rpx;
}
.filter-chip.filter.on {
  color: #fff;
  background: #0A66C2;
  border-color: #0A66C2;
}
.filter-caret { font-size: 20rpx; }

/* ═══════ 匹配条 ═══════ */
.match-strip {
  margin: 24rpx 32rpx 4rpx;
  padding: 24rpx;
  display: flex;
  align-items: center;
  gap: 16rpx;
  border: 1px solid #D4E5F5;
  background: #F7FBFF;
  border-radius: 16rpx;
}
.match-mark {
  width: 72rpx;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: #EAF3FB;
  color: #0A66C2;
  font-weight: 800;
}
.match-mark-icon { font-size: 40rpx; }
.match-copy { flex: 1; min-width: 0; }
.match-title { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; }
.match-desc { display: block; color: #667085; font-size: 22rpx; margin-top: 6rpx; }
.match-link { color: #0A66C2; font-size: 24rpx; font-weight: 600; white-space: nowrap; }

/* ═══════ 列表标题 ═══════ */
.section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.section-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.section-note { font-size: 24rpx; color: #667085; }

/* ═══════ 卡片 ═══════ */
.card-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 0 32rpx 32rpx;
}
.trade-card {
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  overflow: hidden;
}
.trade-card-main {
  display: flex;
  gap: 22rpx;
  padding: 24rpx;
}
.trade-visual {
  width: 164rpx;
  height: 164rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  background: #E8F2FC;
}
.trade-body { flex: 1; min-width: 0; }
.tag-row { display: flex; gap: 10rpx; align-items: center; margin-bottom: 12rpx; }
.tag {
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.gray { color: #667085; background: #F1F3F5; }
.trade-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #17212B;
  font-size: 28rpx;
  line-height: 1.42;
  font-weight: 700;
}
.trade-meta {
  display: flex;
  gap: 14rpx;
  align-items: center;
  margin-top: 14rpx;
  color: #667085;
  font-size: 22rpx;
  white-space: nowrap;
  overflow: hidden;
}
.meta-item { overflow: hidden; text-overflow: ellipsis; }
.trade-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 80rpx;
  padding: 0 24rpx;
  border-top: 1px solid #EEF1F4;
}
.price-block { display: flex; align-items: baseline; min-width: 0; }
.price { color: #DF620F; font-size: 30rpx; font-weight: 750; }
.price-unit { color: #667085; font-size: 20rpx; }
.card-action {
  color: #0A66C2;
  font-size: 24rpx;
  font-weight: 650;
  display: flex;
  align-items: center;
  gap: 4rpx;
  white-space: nowrap;
}
.card-action-arrow { font-size: 30rpx; line-height: 1; }

/* ═══════ 骨架屏 ═══════ */
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 0 32rpx;
}
.skeleton-card {
  height: 216rpx;
  border-radius: 16rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 状态面板 ═══════ */
.state-panel {
  min-height: 620rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
  color: #667085;
  font-size: 26rpx;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-mark.err { color: #D92D20; background: #FEF3F2; }
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}

/* ═══════ 弹层 ═══════ */
.sheet {
  padding-bottom: 20rpx;
}
.sheet-head {
  display: flex;
  align-items: center;
  padding: 28rpx 32rpx 20rpx;
}
.sheet-title { flex: 1; font-size: 32rpx; font-weight: 700; color: #17212B; }
.sheet-close { width: 56rpx; height: 56rpx; display: flex; align-items: center; justify-content: center; }
.sheet-x { font-size: 40rpx; color: #98A2B3; line-height: 1; }
.sheet-body { padding: 0 32rpx 24rpx; }
.sheet-footer {
  display: flex;
  gap: 20rpx;
  padding: 16rpx 32rpx calc(24rpx + env(safe-area-inset-bottom));
}
.sheet-footer > view {
  flex: 1;
  height: 84rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
}
.ghost-btn { border: 1px solid #E4E7EC; background: #fff; color: #344054; }
.primary-btn { background: #0A66C2; color: #fff; }

.field-label {
  display: block;
  color: #344054;
  font-size: 24rpx;
  font-weight: 650;
  margin: 24rpx 0 14rpx;
}
.pick-row { display: flex; flex-wrap: wrap; gap: 14rpx; }
.pick {
  padding: 14rpx 20rpx;
  border: 1px solid #E4E7EC;
  border-radius: 10rpx;
  background: #fff;
  color: #667085;
  font-size: 24rpx;
}
.pick.active {
  background: #EAF3FB;
  border-color: #B9D6EF;
  color: #0A66C2;
  font-weight: 650;
}

/* 城市网格 */
.city-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14rpx;
  padding-bottom: 32rpx;
}
.city-option {
  min-height: 76rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  color: #344054;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
}
.city-option.active {
  color: #074D92;
  border-color: #0A66C2;
  background: #F4F8FC;
  font-weight: 700;
}
.city-check { margin-left: 6rpx; }

/* 响应式：375px 微调 */
@media (max-width: 380px) {
  .trade-visual {
    width: 150rpx;
    height: 150rpx;
  }
}

/* ═══════ 商品设备：电商两列宫格 ═══════ */
.ecom-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  padding: 0 24rpx 20rpx;
}
.ecom-card {
  width: calc(50% - 10rpx);
  background: #fff;
  border-radius: 20rpx;
  overflow: hidden;
  margin-bottom: 20rpx;
  box-shadow: 0 4rpx 16rpx rgba(7, 77, 146, 0.06);
}
.ecom-img-wrap {
  position: relative;
  width: 100%;
  height: 320rpx;
  background: #F0F3F6;
}
.ecom-img {
  width: 100%;
  height: 100%;
}
.ecom-used-tag {
  position: absolute;
  left: 0;
  top: 16rpx;
  padding: 6rpx 14rpx;
  background: rgba(228, 100, 38, 0.92);
  color: #fff;
  font-size: 20rpx;
  border-radius: 0 16rpx 16rpx 0;
}
.ecom-body {
  padding: 18rpx 20rpx 20rpx;
}
.ecom-price {
  display: block;
  color: #E84C3D;
  font-size: 26rpx;
  font-weight: 700;
}
.ecom-price-num {
  font-size: 38rpx;
  font-weight: 800;
}
.ecom-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  margin-top: 8rpx;
  color: #17212B;
  font-size: 26rpx;
  font-weight: 600;
  line-height: 1.4;
  min-height: 73rpx;
}
.ecom-spec {
  display: block;
  margin-top: 6rpx;
  color: #98A2B3;
  font-size: 22rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ecom-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 14rpx;
}
.ecom-cat {
  padding: 4rpx 12rpx;
  background: #EAF3FD;
  color: #0A66C2;
  font-size: 20rpx;
  border-radius: 8rpx;
}
.ecom-views {
  color: #98A2B3;
  font-size: 20rpx;
}

/* ═══════ 就地搜索覆盖层（深蓝条 + 推荐/最近 + 结果卡片，.ov- 前缀避让页面同名类） ═══════ */
.ov-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9990;
  background: #F4F6F8;
  display: flex;
  flex-direction: column;
}
.ov-bar {
  background: #074D92;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 20rpx 16rpx;
  /* 原生导航页面：视口从导航条下方开始，无需状态栏避让；
     与页面 .topbar 同款写法，env(safe-area-inset-top) 在本页返回 0 */
  padding-top: calc(env(safe-area-inset-top) + 16rpx);
  flex-shrink: 0;
}
.ov-back {
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.ov-back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(45deg);
  margin-left: 10rpx;
}
.ov-search-box {
  flex: 1;
  height: 78rpx;
  border-radius: 12rpx;
  background: #fff;
  display: flex;
  align-items: center;
  gap: 14rpx;
  padding: 0 20rpx;
  min-width: 0;
}
.ov-search-icon {
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.ov-search-icon::after {
  content: '';
  position: absolute;
  right: -11rpx;
  bottom: -6rpx;
  width: 13rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #98A2B3;
  transform: rotate(45deg);
}
.ov-search-input { flex: 1; font-size: 26rpx; color: #17212B; }
.ov-clear { padding: 4rpx; }
.ov-clear-x { font-size: 30rpx; color: #98A2B3; line-height: 1; }
.ov-search-btn { color: #fff; font-size: 26rpx; font-weight: 600; white-space: nowrap; flex-shrink: 0; }

/* 结果区（可滚动） */
.ov-results {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.ov-result-head {
  margin: 32rpx 32rpx 8rpx;
  color: #667085;
  font-size: 24rpx;
}
.ov-card-list { padding: 0 32rpx 32rpx; display: flex; flex-direction: column; gap: 20rpx; }
.ov-trade-card {
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  overflow: hidden;
}
.ov-trade-card-main { display: flex; gap: 22rpx; padding: 24rpx; }
.ov-trade-visual {
  width: 164rpx;
  height: 164rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  background: #E8F2FC;
}
.ov-trade-body { flex: 1; min-width: 0; }
.ov-tag-row { display: flex; gap: 10rpx; align-items: center; margin-bottom: 12rpx; }
.ov-tag {
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.ov-tag-blue { color: #0A66C2; background: #EAF3FB; }
.ov-tag-orange { color: #DB5F0D; background: #FFF0E6; }
.ov-tag-green { color: #168A55; background: #E9F7F0; }
.ov-tag-gray { color: #667085; background: #F1F3F5; }
.ov-trade-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #17212B;
  font-size: 28rpx;
  line-height: 1.42;
  font-weight: 700;
}
.ov-trade-meta {
  display: flex;
  gap: 14rpx;
  margin-top: 14rpx;
  color: #667085;
  font-size: 22rpx;
  white-space: nowrap;
  overflow: hidden;
}
.ov-trade-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 80rpx;
  padding: 0 24rpx;
  border-top: 1px solid #EEF1F4;
}
.ov-price-block { display: flex; align-items: baseline; min-width: 0; }
.ov-price { color: #DF620F; font-size: 30rpx; font-weight: 750; }
.ov-price-unit { color: #667085; font-size: 20rpx; }
.ov-card-action { color: #0A66C2; font-size: 24rpx; font-weight: 650; white-space: nowrap; }

/* 空状态 */
.ov-state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.ov-state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.ov-state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.ov-state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }

/* 推荐 / 最近 */
.ov-suggest { padding-bottom: 40rpx; overflow-y: auto; }
.ov-search-block { padding: 36rpx 32rpx 0; }
.ov-block-title { display: block; font-size: 30rpx; font-weight: 700; color: #17212B; margin-bottom: 24rpx; }
.ov-keyword-row { display: flex; flex-wrap: wrap; gap: 16rpx; }
.ov-keyword {
  color: #344054;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10rpx;
  padding: 14rpx 20rpx;
  font-size: 24rpx;
}

@media (max-width: 380px) {
  .ov-trade-visual { width: 150rpx; height: 150rpx; }
}
</style>
