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

      <!-- 供给分段：服务能力 / 商品设备 -->
      <view v-if="primary === 'supply'" class="subtabs">
        <view
          class="subtab"
          :class="{ active: supplyKind === 'service' }"
          @tap="switchSupplyKind('service')"
        >服务能力</view>
        <view
          class="subtab"
          :class="{ active: supplyKind === 'product' }"
          @tap="switchSupplyKind('product')"
        >商品设备</view>
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
        <view class="filter-chip filter" :class="{ on: hasActiveFilter }" @tap="openFilters">
          <text>筛选</text>
          <text class="filter-caret">▾</text>
        </view>
      </view>

      <!-- ═══════ 匹配条 ═══════ -->
      <view v-if="listState === 'ready' && visibleList.length > 0" class="match-strip">
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

      <!-- 发布浮钮 -->
      <view class="floating-publish" hover-class="tap-fade" @tap="openPublishSheet">
        <view class="publish-plus"><text class="publish-plus-sym">＋</text></view>
        <text class="publish-label">发布</text>
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

    <!-- ═══════ 发布选择弹层 ═══════ -->
    <u-popup :show="showPublish" position="bottom" round @close="showPublish = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">选择发布类型</text>
          <view class="sheet-close" @tap="showPublish = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body">
          <view class="choice-list">
            <view class="choice" hover-class="tap-fade" @tap="choosePublish('demand')">
              <view class="choice-icon demand"><text class="choice-sym">需</text></view>
              <view class="choice-copy">
                <text class="choice-name">发布需求</text>
                <text class="choice-desc">发布作业、采购、技术或场景需求</text>
              </view>
              <text class="choice-arrow">›</text>
            </view>
            <view class="choice" hover-class="tap-fade" @tap="choosePublish('service')">
              <view class="choice-icon service"><text class="choice-sym">服</text></view>
              <view class="choice-copy">
                <text class="choice-name">发布服务能力</text>
                <text class="choice-desc">展示巡检、测绘、航拍等可承接能力</text>
              </view>
              <text class="choice-arrow">›</text>
            </view>
            <view class="choice" hover-class="tap-fade" @tap="choosePublish('product')">
              <view class="choice-icon product"><text class="choice-sym">商</text></view>
              <view class="choice-copy">
                <text class="choice-name">发布商品设备</text>
                <text class="choice-desc">展示设备租赁、整机、零部件或载荷</text>
              </view>
              <text class="choice-arrow">›</text>
            </view>
          </view>
        </view>
      </view>
    </u-popup>
  </Layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request } from '../../utils/request'
import { safeNavigateTo } from '../../utils/nav'
import {
  HALL_CATEGORIES, getKindItems, kindTypeLabel, isEnded, normalizeDemand,
  IMG_SOLAR, IMG_LIFT,
} from '../../utils/hallData'

const primary = ref('demand') // demand | supply
const supplyKind = ref('service') // service | product
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
  const kind = primary.value === 'demand' ? 'demand' : supplyKind.value
  return HALL_CATEGORIES[kind]
})

const kindLabel = computed(() => kindTypeLabel(primary.value, supplyKind.value))
const sectionTitle = computed(() => {
  if (primary.value === 'demand') return '最新需求'
  return supplyKind.value === 'service' ? '可对接服务' : '优选商品设备'
})

/* ================= 列表状态 ================= */
const listState = ref('loading') // loading | ready | empty | error
const list = ref([])

async function fetchList(showLoading = true) {
  if (showLoading) listState.value = 'loading'

  if (primary.value === 'demand') {
    try {
      const res = await request({
        url: '/api/v1/demands',
        data: { page: 1, page_size: 20, sort: 'newest' },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      const normalized = items.map(normalizeDemand).filter(Boolean)
      list.value = normalized
      listState.value = normalized.length ? 'ready' : 'empty'
    } catch (e) {
      // 后端不可用：降级到模拟数据，保证页面可交互
      list.value = getKindItems('demand').slice()
      listState.value = list.value.length ? 'ready' : 'empty'
    }
  } else {
    // 供给本期为模拟数据
    await new Promise((resolve) => setTimeout(resolve, 260))
    list.value = getKindItems(primary.value, supplyKind.value).slice()
    listState.value = list.value.length ? 'ready' : 'empty'
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

const goSearch = () => safeNavigateTo('/pages/demands/search')
const goMessages = () => safeNavigateTo('/pages/messages/index')
const goMatches = () => safeNavigateTo('/pages/demands/matches')
const goDetail = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))

const showPublish = ref(false)
const openPublishSheet = () => { showPublish.value = true }
const choosePublish = (type) => {
  showPublish.value = false
  safeNavigateTo('/pages/demands/publish?type=' + type)
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
  if (filterRegion.value !== '不限') out = out.filter((i) => i.region.includes(filterRegion.value))
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

/* ═══════ 发布浮钮 ═══════ */
.floating-publish {
  position: fixed;
  right: 32rpx;
  bottom: 176rpx;
  height: 92rpx;
  border-radius: 46rpx;
  padding: 0 34rpx 0 26rpx;
  display: flex;
  align-items: center;
  gap: 14rpx;
  color: #fff;
  background: #F97316;
  box-shadow: 0 8px 22px rgba(233, 96, 18, 0.32);
  font-weight: 700;
  font-size: 26rpx;
  z-index: 20;
}
.publish-plus {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.19);
}
.publish-plus-sym { font-size: 40rpx; line-height: 1; }

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

/* 发布选择 */
.choice-list { display: flex; flex-direction: column; gap: 20rpx; }
.choice {
  display: flex;
  align-items: center;
  gap: 24rpx;
  width: 100%;
  text-align: left;
  padding: 26rpx;
  border: 1px solid #E4E7EC;
  border-radius: 16rpx;
  background: #fff;
  box-sizing: border-box;
}
.choice-icon {
  width: 86rpx;
  height: 86rpx;
  flex-shrink: 0;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0A66C2;
  background: #EAF3FB;
  font-size: 40rpx;
  font-weight: 700;
}
.choice-icon.service { color: #D15A10; background: #FFF0E6; }
.choice-icon.product { color: #168A55; background: #E9F7F0; }
.choice-copy { flex: 1; min-width: 0; }
.choice-name { display: block; color: #17212B; font-size: 28rpx; font-weight: 700; }
.choice-desc { display: block; color: #667085; font-size: 22rpx; margin-top: 8rpx; line-height: 1.4; }
.choice-arrow { margin-left: auto; color: #98A2B3; font-size: 40rpx; }

/* 响应式：375px 微调 */
@media (max-width: 380px) {
  .trade-visual {
    width: 150rpx;
    height: 150rpx;
  }
}
</style>
