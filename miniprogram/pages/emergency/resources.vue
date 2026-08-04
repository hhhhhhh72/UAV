<template>
  <view class="page">
    <!-- ① 应急橙 Banner -->
    <view class="banner">
      <view class="status-placeholder" :style="{ height: statusBarHeight + 'px' }" />
      <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
      <text class="banner-label">应急协同</text>
      <text class="banner-title">应急资源调度</text>
      <text class="banner-sub">低空应急 · 协同救援 · 快速响应</text>
    </view>

    <view class="main-card">
      <view class="tab-main">
        <view class="tab-main-item" :class="{ active: mainTab === 'resources' }" @click="switchMainTab('resources')">应急资源</view>
        <view class="tab-main-item" :class="{ active: mainTab === 'depts' }" @click="switchMainTab('depts')">部门对接</view>
      </view>

      <template v-if="mainTab === 'resources'">
        <view class="type-pills-wrap">
          <view class="type-pills">
            <view class="pill" :class="{ active: filterType === 'all' }" @click="filterType = 'all'">全部</view>
            <view class="pill" :class="{ active: filterType === 'drone' }" @click="filterType = 'drone'">无人机</view>
            <view class="pill" :class="{ active: filterType === 'comm' }" @click="filterType = 'comm'">通讯</view>
            <view class="pill" :class="{ active: filterType === 'vehicle' }" @click="filterType = 'vehicle'">车辆</view>
            <view class="pill" :class="{ active: filterType === 'medical' }" @click="filterType = 'medical'">医疗</view>
          </view>
        </view>
        <view class="search-bar"><u-icon name="search" size="28rpx" color="#969799" /><input class="search-input" v-model="keyword" placeholder="搜索资源名称" @input="onSearch" /></view>

        <StateView :loading="loading" :error="!!errorMsg" :empty="!loading && !errorMsg && list.length === 0" empty-text="暂无应急资源" @retry="loadResources">
          <scroll-view class="list-scroll" scroll-y @scrolltolower="loadMore">
            <view v-for="item in list" :key="item.id" class="resource-card" :style="{ borderLeftColor: statusColor[item.status] || 'var(--color-success)' }">
              <view class="card-top">
                <view class="card-icon" :style="{ background: statusBg[item.status] || 'var(--ui-color-accent-light)' }"><text>{{ resIcon(item) }}</text></view>
                <view class="card-info"><text class="card-name">{{ item.name || '未命名资源' }}</text><text class="card-spec">{{ item.specs || item.model || '暂无规格' }}</text></view>
                <view class="status-tag" :style="{ background: statusBg[item.status] || 'var(--ui-color-accent-light)', color: statusColor[item.status] || 'var(--color-text-secondary)' }">{{ item.status || '未知' }}</view>
              </view>
              <view class="card-meta">
                <view class="meta-row"><text class="meta-label">数量</text><text class="meta-value">{{ item.quantity || 0 }}</text></view>
                <view class="meta-row"><text class="meta-label">位置</text><text class="meta-value ellipsis">{{ item.location || '未知' }}</text></view>
                <view class="meta-row"><text class="meta-label">联系人</text><text class="meta-value">{{ item.contact || item.contact_info || '暂无' }}</text></view>
              </view>
            </view>
            <view v-if="list.length > 0" class="load-more-wrap"><view v-if="loadingMore" class="loading-inline"><u-loading size="24rpx" /><text>加载更多...</text></view><text v-else-if="!hasMore" class="no-more">没有更多了</text></view>
            <view style="height:40rpx" />
          </scroll-view>
        </StateView>
      </template>

      <template v-else>
        <StateView :loading="deptLoading" :error="!!deptError" :empty="!deptLoading && !deptError && deptList.length === 0" empty-text="暂无部门信息" @retry="loadDepts">
          <scroll-view class="list-scroll" scroll-y>
            <view v-for="d in deptList" :key="d.id" class="dept-card">
              <view class="dept-header"><view class="dept-icon"><text><text>{{ deptIcon(d) }}</text></text></view><view><text class="dept-name">{{ d.name || '未知部门' }}</text><text class="dept-type-area">{{ d.type || d.dept_type || '未知类型' }} · {{ d.area || d.region || '未知区域' }}</text></view></view>
              <view class="dept-contact">联系人: <text class="bold">{{ d.contact || d.contact_name || '暂无' }}</text><text style="margin-left:24rpx;">电话: <text class="bold">{{ d.phone || d.contact_phone || '暂无' }}</text></text></view>
            </view>
            <view style="height:40rpx" />
          </scroll-view>
        </StateView>
      </template>
    </view>
  </view>
</template>

<script setup>
import { ref, watch } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const statusBarHeight = ref(44)
const mainTab = ref('resources')
const filterType = ref('all')
const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = ref(20)
const hasMore = ref(true)
const deptLoading = ref(false)
const deptError = ref('')
const deptList = ref([])

const resourceTypes = [
  { emoji: '', label: '全部', value: 'all' },
  { emoji: '机', label: '无人机', value: 'drone' },
  { emoji: '信', label: '通讯', value: 'comm' },
  { emoji: '车', label: '车辆', value: 'vehicle' },
  { emoji: '医', label: '医疗', value: 'medical' },
  { emoji: '他', label: '其他', value: 'other' },
]

const statusColor = { '可用': 'var(--color-success)', '使用中': 'var(--color-warning)', '维护中': 'var(--color-text-secondary)', 'available': 'var(--color-success)', 'standby': 'var(--color-success)', 'in_use': 'var(--color-warning)', 'maintenance': 'var(--color-text-secondary)' }
const statusBg = { '可用': '#e8f5e9', '使用中': '#fff3e0', '维护中': '#f5f5f5', 'available': '#e8f5e9', 'standby': '#e8f5e9', 'in_use': '#fff3e0', 'maintenance': '#f5f5f5' }
const typeEmoji = { 'drone': '机', 'comm': '信', 'vehicle': '车', 'medical': '医', 'other': '他' }
const deptEmoji = { '消防': '防', 'fire': '防', '公安': '警', 'police': '警', '应急局': '应', 'emergency_bureau': '应', '医疗': '医', 'civil_affairs': '医', '交通': '交' }

var searchTimer = null
function resIcon(item) {
  var t = item.resource_type || item.res_type || 'drone'
  if (t === 'drone') return '机'
  if (t === 'comm') return '信'
  if (t === 'vehicle') return '车'
  if (t === 'medical') return '医'
  return '他'
}
function deptIcon(d) {
  var t = d.type || d.dept_type || ''
  if (t.indexOf('消防') >= 0 || t === 'fire') return '防'
  if (t.indexOf('公安') >= 0 || t === 'police') return '警'
  if (t.indexOf('应急') >= 0 || t === 'emergency_bureau') return '应'
  if (t.indexOf('医疗') >= 0 || t === 'civil_affairs') return '医'
  if (t.indexOf('交通') >= 0) return '交'
  return '部'
}
function onSearch() { clearTimeout(searchTimer); searchTimer = setTimeout(function () { page.value = 1; loadResources(true) }, 300) }
function switchMainTab(tab) { if (mainTab.value === tab) return; mainTab.value = tab; if (tab === 'depts' && deptList.value.length === 0) loadDepts() }
watch(filterType, function () { page.value = 1; loadResources(true) })

async function loadResources(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true } else { loadingMore.value = true }
  errorMsg.value = ''
  try {
    var params = { page: page.value, page_size: pageSize.value }
    if (filterType.value !== 'all') params.res_type = filterType.value
    if (keyword.value) params.q = keyword.value
    var res = await request({ url: '/api/v1/emergency-resources', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []
    var total = (data && data.total) != null ? data.total : items.length
    if (reset) { list.value = items } else { list.value = list.value.concat(items) }
    hasMore.value = list.value.length < total
  } catch (e) {
    if (reset) errorMsg.value = '网络异常，请稍后重试'
  }
  finally { loading.value = false; loadingMore.value = false }
}

function loadMore() { if (!loadingMore.value && hasMore.value) { page.value++; loadResources(false) } }

async function loadDepts() {
  deptLoading.value = true; deptError.value = ''
  try {
    var res = await request({ url: '/api/v1/emergency-depts' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    deptList.value = Array.isArray(data) ? data : (data && data.items) || data || []
  } catch (e) { deptError.value = '网络异常，请稍后重试' }
  finally { deptLoading.value = false }
}

function goBack() { uni.navigateBack({ delta: 1 }) }

onLoad(function () {
  try { statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 44 } catch (e) {}
  loadResources(true)
})

onPullDownRefresh(function () {
  var p = mainTab.value === 'resources' ? loadResources(true) : loadDepts()
  p.then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }
.banner { background: linear-gradient(135deg, var(--color-warning), #d84315); padding: 0 32rpx 72rpx; }
.status-placeholder { width: 100%; }
.back-btn { width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.2); border-radius: 50%; display: flex; align-items: center; justify-content: center; margin-bottom: 24rpx; }
.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }
.banner-label { color: rgba(255,255,255,0.85); font-size: 28rpx; font-weight: 500; display: block; }
.banner-title { color: #ffffff; font-size: 56rpx; font-weight: 700; display: block; margin: 8rpx 0 12rpx; }
.banner-sub { color: rgba(255,255,255,0.7); font-size: 26rpx; }
.main-card { background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -32rpx; position: relative; z-index: 2; }
.tab-main { display: flex; justify-content: center; gap: 24rpx; padding: 0 24rpx; margin-top: -36rpx; }
.tab-main-item { width: 320rpx; height: 72rpx; line-height: 72rpx; text-align: center; border-radius: 40rpx; font-size: 28rpx; font-weight: 400; color: #666666; background: #ffffff; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06); }
.tab-main-item.active { background: var(--color-warning); color: #ffffff; font-weight: 600; box-shadow: 0 4rpx 16rpx rgba(255,159,10,0.35); }
.type-pills-wrap { width: 100%; overflow-x: auto; }
.type-pills { display: flex; flex-wrap: nowrap; overflow-x: auto; gap: 16rpx; padding: 24rpx 24rpx 16rpx; }
.type-pills::-webkit-scrollbar { display: none; }
.pill { flex: 0 0 auto; white-space: nowrap; padding: 14rpx 36rpx; border-radius: 28rpx; font-size: 28rpx; background: var(--color-bg); color: var(--color-text-secondary); }
.pill.active { background: var(--color-warning); color: #ffffff; font-weight: 500; }
.search-bar { margin: 0 24rpx 16rpx; background: var(--color-bg); border-radius: 40rpx; padding: 14rpx 24rpx; display: flex; align-items: center; gap: 12rpx; }
.search-input { flex: 1; font-size: 28rpx; color: var(--color-text); }
.list-scroll { height: calc(100vh - 560rpx); }
.resource-card { background: #ffffff; border-radius: 14rpx; padding: 20rpx 20rpx 20rpx 16rpx; margin: 0 24rpx 14rpx; border-left: 6rpx solid; box-shadow: 0 2rpx 10rpx rgba(0,0,0,0.03); }
.card-top { display: flex; align-items: flex-start; gap: 12rpx; margin-bottom: 12rpx; }
.card-icon { width: 72rpx; height: 72rpx; border-radius: 16rpx; display: flex; align-items: center; justify-content: center; font-size: 36rpx; flex-shrink: 0; }
.card-info { flex: 1; }
.card-name { font-size: 30rpx; font-weight: 500; color: var(--color-text); display: block; }
.card-spec { font-size: 24rpx; color: var(--color-text-secondary); display: block; margin-top: 4rpx; }
.status-tag { padding: 6rpx 16rpx; border-radius: 12rpx; font-size: 22rpx; font-weight: 500; flex-shrink: 0; }
.card-meta { display: flex; flex-direction: column; gap: 6rpx; font-size: 24rpx; }
.meta-row { display: flex; gap: 12rpx; align-items: baseline; }
.meta-label { color: var(--color-text-secondary); width: 80rpx; flex-shrink: 0; }
.meta-value { color: var(--color-text); font-weight: 500; flex: 1; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.meta-value.ellipsis { overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.card-meta .bold { color: var(--color-text); font-weight: 500; }
.dept-card { background: #ffffff; border-radius: 14rpx; padding: 20rpx; margin: 0 24rpx 14rpx; box-shadow: 0 2rpx 10rpx rgba(0,0,0,0.03); }
.dept-header { display: flex; align-items: center; gap: 14rpx; margin-bottom: 10rpx; }
.dept-icon { width: 60rpx; height: 60rpx; background: #fff3e0; border-radius: 14rpx; display: flex; align-items: center; justify-content: center; font-size: 28rpx; flex-shrink: 0; }
.dept-name { font-size: 28rpx; font-weight: 500; color: var(--color-text); display: block; }
.dept-type-area { font-size: 24rpx; color: var(--color-text-secondary); display: block; margin-top: 2rpx; }
.dept-contact { font-size: 24rpx; color: var(--color-text-secondary); }
.dept-contact .bold { color: var(--color-text); font-weight: 500; }
.load-more-wrap { text-align: center; padding: 20rpx 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8rpx; font-size: 24rpx; color: var(--color-text-secondary); }
.no-more { font-size: 24rpx; color: var(--color-text-secondary); }
</style>
