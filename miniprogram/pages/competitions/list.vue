<template>
  <view class="page">
    <u-nav-bar title="赛事列表" show-back @back="goBack" />

    <!-- 筛选 Tab + 搜索 -->
    <view class="tabs-wrap">
      <u-tabs :active="tabIndex" :titles="['全部赛事', '报名中']" @change="onTabChange" />
    </view>
    <u-search v-model="keyword" placeholder="搜索赛事名称" @change="onSearch" />

    <!-- 赛事卡片列表 -->
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && list.length === 0"
      empty-text="暂无赛事"
      @retry="loadData(true)"
    >
      <view class="list">
        <view
          v-for="item in list"
          :key="item.id"
          class="card"
          hover-class="card--press"
          :hover-stay-time="120"
          @click="goDetail(item)"
        >
          <!-- 左侧：类型图标块 -->
          <view class="thumb" :class="'thumb--' + thumbType(item)">
            <text class="thumb-char">{{ thumbChar(item) }}</text>
          </view>

          <!-- 右侧：信息 -->
          <view class="info">
            <view class="info-top">
              <text class="title">{{ item.title || item.name || '未知赛事' }}</text>
              <text class="chip" :class="statusClass(item.status)">{{ statusText(item.status) }}</text>
            </view>

            <view class="meta">
              <text class="meta-line">{{ fmtDate(item.start_date) }} - {{ fmtDate(item.end_date) }}</text>
              <text class="meta-line">{{ item.location || '地点待定' }}</text>
              <text v-if="item.organizer || item.sponsor" class="meta-line meta-line--sub">{{ item.organizer || item.sponsor }}</text>
            </view>

            <view v-if="compTags(item).length > 0" class="tags">
              <text v-for="t in compTags(item)" :key="t" class="chip chip--tag" :class="tagClass(t)">{{ t }}</text>
            </view>

            <view class="bottom">
              <view v-if="compFee(item) > 0" class="fee">
                <text class="fee-symbol">¥</text>
                <text class="fee-num">{{ compFee(item).toLocaleString() }}</text>
                <text class="fee-unit">/人</text>
              </view>
              <text v-else class="fee-free">免费</text>

              <view v-if="!isClosed(item)" class="enroll-btn" @click.stop="goRegister(item)">立即报名</view>
              <text v-else class="closed-text">已截止</text>
            </view>
          </view>
        </view>

        <view v-if="list.length > 0" class="load-more">
          <u-loading v-if="loadingMore" size="24rpx" color="#667085" />
          <text class="load-text">{{ loadingMore ? '加载更多...' : (hasMore ? '' : '没有更多了') }}</text>
        </view>
        <view class="list-bottom-space" />
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const tabIndex = ref(0)
const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

/* ===== 状态 ===== */

function isClosed(item) {
  return item.status === 'closed' || item.status === 'full'
}

function statusText(item) {
  var map = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已截止', full: '已满员' }
  return map[item.status] || '报名中'
}

function statusClass(status) {
  if (status === 'ongoing') return 'chip--ongoing'
  if (status === 'closed' || status === 'full') return 'chip--closed'
  return 'chip--enrolling'
}

/* ===== 数据映射 ===== */

function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags.slice(0, 3)
  if (item.category) return [item.category]
  return []
}

function compFee(item) {
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  if (item.price != null) return item.price
  return 0
}

/* 分类类型：竞速=紫、创新=绿、技能=蓝、其余=橙 */
function thumbType(item) {
  var t = item.title || ''
  if (t.indexOf('FPV') >= 0 || t.indexOf('竞速') >= 0) return 'purple'
  if (t.indexOf('创新') >= 0 || t.indexOf('应用') >= 0) return 'green'
  if (t.indexOf('技能') >= 0 || t.indexOf('全国') >= 0 || t.indexOf('职业') >= 0) return 'blue'
  return 'orange'
}

function thumbChar(item) {
  var t = item.title || ''
  if (t.indexOf('全国') >= 0) return '国'
  if (t.indexOf('西南') >= 0) return '西'
  if (t.indexOf('国际') >= 0) return '国'
  if (t.indexOf('青少年') >= 0) return '青'
  if (t.indexOf('贵州') >= 0) return '贵'
  if (t.indexOf('创新') >= 0) return '创'
  return '赛'
}

function tagClass(tag) {
  if (['国家级', '国际赛', '国际', '航空级'].indexOf(tag) >= 0) return 'chip--level'
  return 'chip--model'
}

/* ===== 数据获取 ===== */

async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''

  try {
    var params = { page: page.value, page_size: pageSize }
    if (tabIndex.value === 1) params.status = 'enrolling'
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/competitions', data: params })
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

function loadMore() {
  if (!loadingMore.value && hasMore.value) { page.value++; loadData(false) }
}

/* ===== 筛选 & 搜索 ===== */

function onTabChange(index) {
  if (tabIndex.value === index) return
  tabIndex.value = index
  loadData(true)
}

var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { loadData(true) }, 300)
}

/* ===== 交互 ===== */

function goDetail(item) {
  uni.navigateTo({ url: '/pages/competitions/detail?id=' + encodeURIComponent(item.id) })
}

function goRegister(item) {
  uni.navigateTo({ url: '/pages/competitions/register?id=' + encodeURIComponent(item.id) })
}

function goBack() { uni.navigateBack({ delta: 1 }) }

onLoad(function () { loadData(true) })

onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})

onReachBottom(function () { loadMore() })
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ① Tab + 搜索 */
.tabs-wrap { background: #ffffff; }
.tabs-wrap :deep(.u-tabs) { background: #ffffff; }

/* ② 卡片列表 */
.list { padding: 20rpx 24rpx 0; }

.card {
  display: flex;
  gap: 20rpx;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-sizing: border-box;
  transition: opacity 160ms ease;
}

.card--press { opacity: 0.85; }

/* 左侧类型图标块（低饱和色块） */
.thumb {
  width: 88rpx;
  height: 88rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.thumb--orange { background: #FFF0E6; }
.thumb--blue { background: #EAF3FB; }
.thumb--green { background: #E9F7F0; }
.thumb--purple { background: #F6F4FF; }

.thumb--orange .thumb-char { color: #E96012; }
.thumb--blue .thumb-char { color: #0A66C2; }
.thumb--green .thumb-char { color: #168A55; }
.thumb--purple .thumb-char { color: #6E56CF; }

.thumb-char { font-size: 34rpx; font-weight: 700; }

/* 右侧信息 */
.info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 10rpx; }

.info-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 12rpx; }

.title {
  font-size: 32rpx;
  font-weight: 600;
  color: #17212B;
  flex: 1;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* 状态徽章（4px 圆角，非胶囊） */
.chip {
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 500;
  flex-shrink: 0;
  line-height: 1.6;
}

.chip--enrolling { background: #FFF0E6; color: #E96012; }
.chip--ongoing { background: #EAF3FB; color: #0A66C2; }
.chip--closed { background: #F4F6F8; color: #667085; }

.meta { display: flex; flex-direction: column; gap: 2rpx; }

.meta-line {
  font-size: 24rpx;
  color: #667085;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.5;
}

.meta-line--sub { color: #98A2B3; }

.tags { display: flex; gap: 10rpx; flex-wrap: wrap; }

.chip--tag { font-weight: 400; }
.chip--model { background: #EAF3FB; color: #0A66C2; }
.chip--level { background: #FFF0E6; color: #E96012; }

/* 底部：价格 + 按钮 */
.bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4rpx;
  padding-top: 16rpx;
  border-top: 1px solid #EEF1F4;
}

.fee { display: flex; align-items: baseline; }

.fee-symbol { font-size: 24rpx; font-weight: 700; color: #E96012; }
.fee-num { font-size: 34rpx; font-weight: 700; color: #E96012; line-height: 1; }
.fee-unit { font-size: 22rpx; color: #98A2B3; margin-left: 4rpx; }

.fee-free {
  padding: 6rpx 14rpx;
  background: #E9F7F0;
  color: #168A55;
  font-size: 24rpx;
  font-weight: 600;
  border-radius: 8rpx;
}

.enroll-btn {
  padding: 10rpx 28rpx;
  background: #0A66C2;
  color: #ffffff;
  font-size: 26rpx;
  font-weight: 500;
  border-radius: 12rpx;
  transition: opacity 160ms ease;
}

.enroll-btn:active { opacity: 0.85; }

.closed-text { font-size: 26rpx; color: #98A2B3; font-weight: 400; }

/* 加载更多 */
.load-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  padding: 20rpx 0 8rpx;
}

.load-text { font-size: 24rpx; color: #98A2B3; }
.list-bottom-space { height: 40rpx; }
</style>
