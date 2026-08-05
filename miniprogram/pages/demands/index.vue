<template>
  <Layout :current="1">
    <view class="demand-hall-page">
      <!-- 需求大厅 / 供给大厅 / 匹配推荐 三段切换 -->
      <u-tabs :titles="hallTabs" v-model:active="activeTab" @change="onTabChange" />

      <!-- ═══════════ 需求大厅段 ═══════════ -->
      <view v-if="activeTab === 0" class="hall-body">
        <view class="search-row">
          <u-search
            v-model="searchText"
            placeholder="搜索需求、项目"
            @search="onSearch"
          />
        </view>

        <!-- 业务分类筛选 -->
        <view class="filter-tabs">
          <view
            v-for="(tab, index) in bizTypeTabs"
            :key="index"
            class="filter-tab"
            :class="{ active: activeBizType === tab.value }"
            @tap="switchBizType(tab.value)"
          >
            {{ tab.label }}
          </view>
        </view>

        <!-- 加载状态 -->
        <view v-if="loading && list.length === 0" class="state-wrap">
          <u-loading size="28rpx" />
          <text class="state-text">加载中...</text>
        </view>
        <!-- 空状态 -->
        <view v-else-if="!loading && list.length === 0 && !errorMsg" class="state-wrap">
          <u-empty description="暂无需求" />
        </view>
        <!-- 失败状态 -->
        <view v-else-if="errorMsg && list.length === 0" class="state-wrap">
          <u-empty description="加载失败" />
          <view class="retry-btn" @tap="fetchList(true)"><text>重新加载</text></view>
        </view>
        <!-- 列表 -->
        <view v-else class="list-body">
          <view
            v-for="item in list"
            :key="item.id"
            class="demand-card"
            @tap="goDetail(item)"
          >
            <text class="card-title">{{ item.title }}</text>
            <view class="card-meta">
              <u-tag :type="bizTypeTagType(item.biz_type)" size="mini">
                {{ bizTypeLabel(item.biz_type) }}
              </u-tag>
              <text v-if="item.district" class="meta-text">{{ item.district }}</text>
              <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
              <text class="meta-date">{{ formatDate(item.created_at) }}</text>
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
      </view>

      <!-- ═══════════ 供给大厅段 ═══════════ -->
      <view v-else-if="activeTab === 1" class="hall-body">
        <view class="supply-grid">
          <view class="supply-item" @tap="go('/pages/mall/index')">
            <text class="supply-name">设备商城</text>
            <text class="supply-desc">整机/零部件/配件</text>
          </view>
          <view class="supply-item" @tap="go('/pages/services/detail?id=trade')">
            <text class="supply-name">服务供给</text>
            <text class="supply-desc">航拍/维修/试飞/检测</text>
          </view>
          <view class="supply-item" @tap="go('/pages/shops/index')">
            <text class="supply-name">商家店铺</text>
            <text class="supply-desc">会员企业店铺</text>
          </view>
          <view class="supply-item" @tap="go('/pages/resources/list')">
            <text class="supply-name">资源池</text>
            <text class="supply-desc">场地/试飞/中试</text>
          </view>
        </view>
        <view class="supply-tip">
          <text class="tip-label">对接方式</text>
          <text class="tip-copy">供需双方在线下联系洽谈，成交后在管理端登记金额</text>
        </view>
      </view>

      <!-- ═══════════ 匹配推荐段 ═══════════ -->
      <view v-else class="hall-body">
        <view class="match-panel">
          <text class="match-title">智能匹配</text>
          <text class="match-desc">根据你的能力标签与作业区域，为你推荐合适需求</text>
          <view class="match-btn" @tap="go('/pages/match/recommend')">
            <text>查看我的匹配</text>
          </view>
        </view>

        <view class="match-list">
          <text class="match-list-title">为你推荐</text>
          <view v-if="recLoading" class="state-wrap">
            <u-loading size="28rpx" />
            <text class="state-text">加载中...</text>
          </view>
          <view v-else-if="recList.length === 0" class="state-wrap">
            <u-empty description="暂无可推荐需求" />
          </view>
          <view
            v-for="item in recList"
            :key="item.id"
            class="demand-card"
            @tap="goDetail(item)"
          >
            <text class="card-title">{{ item.title }}</text>
            <view class="card-meta">
              <u-tag :type="bizTypeTagType(item.biz_type)" size="mini">
                {{ bizTypeLabel(item.biz_type) }}
              </u-tag>
              <text v-if="item.district" class="meta-text">{{ item.district }}</text>
              <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request } from '../../utils/request'
import { BIZ_TYPE_TABS, bizTypeLabel } from '../../utils/enums'

const hallTabs = ['需求大厅', '供给大厅', '匹配推荐']
const activeTab = ref(0)

// ---- 需求大厅 ----
const searchText = ref('')
const activeBizType = ref('')
const bizTypeTabs = BIZ_TYPE_TABS
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

async function fetchList(reset) {
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
    if (activeBizType.value) params.biz_type = activeBizType.value
    if (searchText.value) params.q = searchText.value

    const res = await request({ url: '/api/v1/demands', data: params })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    const total = (data && data.total) != null ? data.total : items.length

    list.value = reset ? items : list.value.concat(items)
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

function onSearch() { fetchList(true) }
function switchBizType(value) {
  activeBizType.value = value
  fetchList(true)
}
function onTabChange() {
  if (activeTab.value === 0 && list.value.length === 0) fetchList(true)
  if (activeTab.value === 2 && recList.value.length === 0) fetchRec()
}

// ---- 匹配推荐 ----
const recLoading = ref(false)
const recList = ref([])

async function fetchRec() {
  recLoading.value = true
  try {
    const res = await request({ url: '/api/v1/recommendations' })
    const data = Array.isArray(res) ? res : (res && res.data) || []
    recList.value = Array.isArray(data) ? data : []
  } catch (e) {
    recList.value = []
  } finally {
    recLoading.value = false
  }
}

// ---- 通用 ----
function go(url) {
  uni.navigateTo({ url })
}
function goDetail(item) {
  uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(item.id) })
}
function bizTypeTagType(type) {
  const map = {
    cable_inspection: 'primary',
    plant_transport: 'success',
    spray_pesticide: 'warning',
    trade_lease: 'danger',
    clean_paint: 'primary',
    other: 'default',
  }
  return map[type] || 'default'
}
function formatBudget(fen) {
  if (fen == null || fen === 0) return '面议'
  const yuan = (fen / 100).toFixed(2)
  return yuan.replace(/\.00$/, '') + ' 元'
}
function formatDate(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  const m = d.getMonth() + 1
  const day = d.getDate()
  return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
}

onLoad(() => fetchList(true))
onPullDownRefresh(() => {
  if (activeTab.value === 0) fetchList(true)
  if (activeTab.value === 2) fetchRec()
  uni.stopPullDownRefresh()
})
onReachBottom(() => {
  if (activeTab.value === 0) loadMore()
})
</script>

<style scoped>
.demand-hall-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

.hall-body {
  padding-bottom: 16px;
}

/* 搜索行 */
.search-row {
  padding: 8px 12px 0;
  background: #fff;
}

/* 分类筛选 */
.filter-tabs {
  display: flex;
  padding: 8px 12px 10px;
  gap: 8px;
  background: #fff;
  overflow-x: auto;
  white-space: nowrap;
  -webkit-overflow-scrolling: touch;
}

.filter-tabs::-webkit-scrollbar {
  display: none;
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

/* 需求卡片 */
.list-body {
  padding: 0 12px;
}

.demand-card {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #EEF1F4;
  padding: 12px;
  margin-top: 8px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  display: block;
  line-height: 1.4;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.meta-text {
  font-size: 12px;
  color: #667085;
}

.meta-date {
  font-size: 11px;
  color: #98A2B3;
  margin-left: auto;
}

/* 状态 */
.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 0;
}

.state-text {
  font-size: 13px;
  color: #667085;
}

.retry-btn {
  padding: 8px 24px;
  border-radius: 8px;
  border: 1px solid #0A66C2;
  color: #0A66C2;
  font-size: 13px;
}

/* 加载更多 */
.load-more {
  padding: 14px 0;
  text-align: center;
}

.loading-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #667085;
}

.no-more {
  font-size: 12px;
  color: #98A2B3;
}

/* 供给大厅 */
.supply-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  padding: 12px;
}

.supply-item {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.supply-name {
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
}

.supply-desc {
  font-size: 11px;
  color: #98A2B3;
}

.supply-tip {
  margin: 0 12px;
  padding: 10px 12px;
  background: #F4F8FC;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tip-label {
  font-size: 11px;
  color: #0A66C2;
  font-weight: 600;
}

.tip-copy {
  font-size: 12px;
  color: #344054;
}

/* 匹配推荐 */
.match-panel {
  margin: 12px;
  padding: 16px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.match-title {
  font-size: 17px;
  font-weight: 700;
  color: #17212B;
}

.match-desc {
  font-size: 12px;
  color: #667085;
}

.match-btn {
  margin-top: 8px;
  align-self: flex-start;
  padding: 8px 20px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
}

.match-list {
  padding: 0 12px;
}

.match-list-title {
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  display: block;
  padding: 4px 0 8px;
}
</style>
