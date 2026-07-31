<template>
  <view class="page">
    <!-- ① 绿色 Banner -->
    <view class="banner">
      <view class="status-placeholder" :style="{ height: statusBarHeight + 'px' }" />
      <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
      <text class="banner-pretitle">研学活动</text>
      <text class="banner-heading">无人机研学实践</text>
      <text class="banner-sub">走进企业 · 体验飞行 · 拓展视野</text>
    </view>

    <!-- ② 搜索 + 列表 -->
    <view class="main-card">
      <view class="search-bar">
        <text class="search-icon">🔍</text>
        <input class="search-input" v-model="keyword" placeholder="搜索研学活动" @input="onSearch" />
      </view>

      <StateView
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && list.length === 0"
        empty-text="暂无研学活动"
        @retry="loadData"
      >
        <scroll-view class="list-scroll" scroll-y @scrolltolower="loadMore">
          <view v-for="item in list" :key="item.id" class="study-card" @click="openDetail(item)">
            <!-- 封面 -->
            <view class="card-cover">
              <image v-if="item.cover || item.cover_image || item.image" :src="item.cover || item.cover_image || item.image" class="cover-img" mode="aspectFill" />
              <view v-else class="cover-placeholder"><text class="cover-emoji">{{ item.icon || '🚁' }}</text></view>
              <view class="status-badge" :style="{ background: statusColor[item.status] || '#07c160' }">{{ item.status || '报名中' }}</view>
            </view>

            <view class="card-body">
              <text class="card-name">{{ item.name || item.title || '未知活动' }}</text>

              <view class="card-info">
                <view class="info-row"><text class="info-label">时间</text><text class="info-value">{{ item.date || (item.start_date || '2026.08.15') + ' - ' + (item.end_date || '08.16') }}</text></view>
                <view class="info-row"><text class="info-label">地点</text><text class="info-value ellipsis">{{ item.location || '待定' }}</text></view>
                <view class="info-row"><text class="info-label">对象</text><text class="info-value">{{ item.target || '不限' }}</text></view>
              </view>

              <view v-if="studyTags(item).length > 0" class="tag-row">
                <text v-for="t in studyTags(item)" :key="t" class="study-tag" :style="tagStyle(t)">{{ t }}</text>
              </view>

              <view class="card-footer">
                <view>
                  <text class="fee-label">费用</text>
                  <text class="fee-value">¥{{ studyFee(item).toLocaleString() }}</text>
                  <text class="fee-unit">/人</text>
                </view>
                <text class="detail-hint">点击查看详情 →</text>
              </view>
            </view>
          </view>

          <view v-if="list.length > 0" class="load-more-wrap">
            <van-loading v-if="loadingMore" size="20">加载更多...</van-loading>
            <text v-else-if="!hasMore" class="no-more">没有更多了</text>
          </view>
          <view style="height:40rpx" />
        </scroll-view>
      </StateView>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const statusBarHeight = ref(44)
const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

const statusColor = { '报名中': '#07c160', '即将开始': '#ff6b35', '已结束': '#969799' }

function tagStyle(t) {
  if (['企业参访', '工厂参观'].indexOf(t) >= 0) return { background: '#e8f5e9', color: '#07c160' }
  if (['动手实操', '航模制作'].indexOf(t) >= 0) return { background: '#e8f0ff', color: '#2b5ea7' }
  return { background: '#fff4e6', color: '#ff6b35' }
}

function studyTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (item.category) return [item.category]
  return ['企业参访', '动手实操']
}

function studyFee(item) {
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  if (item.price != null) return item.price
  return 0
}

function openDetail(item) {
  var url = (item && item.url) || 'https://example.com/study'
  uni.navigateTo({ url: '/pages/webview/index?url=' + encodeURIComponent(url) })
}

var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { page.value = 1; loadData(true) }, 300)
}

async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''

  try {
    var params = { page: page.value, page_size: pageSize }
    if (keyword.value) params.keyword = keyword.value
    var res = await request({ url: '/api/v1/cooperation-programs', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []
    var total = (data && data.total) != null ? data.total : items.length
    if (reset) { list.value = items } else { list.value = list.value.concat(items) }
    hasMore.value = list.value.length < total
    if (list.value.length === 0) { list.value = getMockList(); hasMore.value = false }
  } catch (e) {
    if (reset) { list.value = getMockList(); hasMore.value = false }
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function getMockList() {
  return [
    { id: 'study-1', name: '大疆创新总部研学之旅', status: '报名中', date: '2026.08.15 - 08.16', location: '深圳市南山区大疆创新总部', target: '12岁以上青少年', tags: ['企业参访', '动手实操', '证书'], fee: 580, icon: '🏢', url: 'https://www.dji.com/cn/robomaster' },
    { id: 'study-2', name: '成都航空产业基地一日研学', status: '报名中', date: '2026.08.20', location: '成都市高新区无人机产业基地', target: '8-16岁青少年', tags: ['工厂参观', '航模制作', '企业参访'], fee: 380, icon: '✈️', url: 'https://example.com/study/chengdu' },
    { id: 'study-3', name: '贵州无人机应急救援体验营', status: '报名中', date: '2026.09.01 - 09.02', location: '贵阳市观山湖区应急指挥中心', target: '16岁以上', tags: ['动手实操', '证书'], fee: 680, icon: '🛟', url: 'https://example.com/study/guizhou' },
    { id: 'study-4', name: '无人机航拍创作夏令营', status: '即将开始', date: '2026.08.25 - 08.28', location: '云南省大理市洱海实训基地', target: '14岁以上', tags: ['动手实操', '证书'], fee: 1280, icon: '📸', url: 'https://example.com/study/aerial' },
    { id: 'study-5', name: '北航无人机科技研学周', status: '报名中', date: '2026.10.01 - 10.05', location: '北京航空航天大学', target: '高中生', tags: ['企业参访', '动手实操', '航模制作', '证书'], fee: 1980, icon: '🎓', url: 'https://www.buaa.edu.cn' },
    { id: 'study-6', name: '青少年FPV穿越机体验日', status: '已结束', date: '2026.07.10', location: '上海市浦东新区竞速基地', target: '10-18岁青少年', tags: ['动手实操'], fee: 280, icon: '🥽', url: 'https://example.com/study/fpv' },
  ]
}

function loadMore() {
  if (!loadingMore.value && hasMore.value) { page.value++; loadData(false) }
}

function goBack() { uni.navigateBack({ delta: 1 }) }

onLoad(function () {
  try { statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 44 } catch (e) {}
  loadData(true)
})

onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f6f8; padding-bottom: env(safe-area-inset-bottom); }

/* ① Banner */
.banner { background: linear-gradient(135deg, #07c160, #05a854); padding: 0 32rpx 72rpx; }

.status-placeholder { width: 100%; }

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.15);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  margin-bottom: 24rpx;
}

.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }

.banner-pretitle { color: rgba(255,255,255,0.9); font-size: 26rpx; font-weight: 500; display: block; margin-bottom: 12rpx; }

.banner-heading { color: #ffffff; font-size: 56rpx; font-weight: 700; line-height: 1.2; display: block; margin-bottom: 16rpx; }

.banner-sub { color: rgba(255,255,255,0.7); font-size: 26rpx; font-weight: 400; display: block; }

/* ② 搜索 + 卡片 */
.main-card { background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -32rpx; padding: 24rpx 0 0; position: relative; z-index: 2; }

.search-bar { margin: 0 24rpx 24rpx; background: #f5f6f8; border-radius: 40rpx; padding: 16rpx 24rpx; display: flex; align-items: center; gap: 12rpx; }
.search-icon { font-size: 28rpx; opacity: 0.4; }
.search-input { flex: 1; font-size: 28rpx; color: #1a1a1a; }

.list-scroll { padding: 0 24rpx; height: calc(100vh - 440rpx); }

.study-card { background: #ffffff; border-radius: 20rpx; overflow: hidden; margin-bottom: 24rpx; box-shadow: 0 2rpx 16rpx rgba(0,0,0,0.04); }

.card-cover { height: 200rpx; position: relative; overflow: hidden; }
.cover-img { width: 100%; height: 100%; }
.cover-placeholder { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #1a237e, #283593); }
.cover-emoji { font-size: 80rpx; opacity: 0.12; }

.status-badge { position: absolute; top: 16rpx; right: 16rpx; padding: 6rpx 18rpx; border-radius: 20rpx; color: #ffffff; font-size: 22rpx; font-weight: 600; }

.card-body { padding: 24rpx; }
.card-name { font-size: 32rpx; font-weight: 600; color: #1a1a1a; display: block; margin-bottom: 16rpx; }

.card-info { margin-bottom: 14rpx; }
.info-row { display: flex; align-items: center; gap: 12rpx; margin-bottom: 8rpx; }
.info-label { font-size: 24rpx; color: #969799; width: 56rpx; flex-shrink: 0; }
.info-value { font-size: 26rpx; color: #4a4a4a; }
.info-value.ellipsis { overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }

.tag-row { display: flex; flex-wrap: wrap; gap: 10rpx; margin-bottom: 20rpx; }
.study-tag { padding: 4rpx 14rpx; border-radius: 12rpx; font-size: 22rpx; font-weight: 500; }

.card-footer { display: flex; justify-content: space-between; align-items: center; border-top: 1rpx solid #f0f0f0; padding-top: 16rpx; }
.fee-label { font-size: 24rpx; color: #969799; }
.fee-value { font-size: 36rpx; font-weight: 700; color: #ff6b35; margin: 0 6rpx; }
.fee-unit { font-size: 22rpx; color: #969799; }
.detail-hint { font-size: 24rpx; color: #07c160; font-weight: 500; }

.load-more-wrap { text-align: center; padding: 20rpx 0; }
.no-more { font-size: 24rpx; color: #969799; }
</style>
