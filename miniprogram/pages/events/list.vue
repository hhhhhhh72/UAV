<template>
  <view class="page">
    <!-- ① 海军蓝 Banner -->
    <view class="banner">
      <view class="banner-nav">
        <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
        <text class="banner-nav-title">赛事列表</text>
      </view>
      <text class="banner-title">无人机竞技与技能大赛</text>
      <text class="banner-subtitle">权威赛事 · 技能比拼 · 荣誉认证</text>
    </view>

    <!-- ② Tab + 搜索 -->
    <view class="main-card">
      <view class="tabs-container">
        <view
          class="tab-item"
          :class="{ active: currentTab === 'all' }"
          @click="switchTab('all')"
        >全部赛事</view>
        <view
          class="tab-item"
          :class="{ active: currentTab === 'enrolling' }"
          @click="switchTab('enrolling')"
        >报名中</view>
      </view>

      <view class="search-bar">
        <u-icon name="search" size="28rpx" color="#969799" />
        <input
          class="search-input"
          v-model="keyword"
          placeholder="搜索赛事名称"
          @input="onSearch"
        />
      </view>

      <!-- ③ 赛事卡片列表 -->
      <StateView
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && list.length === 0"
        empty-text="暂无赛事"
        @retry="loadData"
      >
        <scroll-view class="list-scroll" scroll-y @scrolltolower="loadMore">
          <view
            v-for="item in list"
            :key="item.id"
            class="comp-card"
            @click="goDetail(item)"
          >
            <!-- 封面 -->
            <view class="card-cover">
              <image
                v-if="item.cover || item.cover_image || item.image"
                :src="item.cover || item.cover_image || item.image"
                class="cover-img"
                mode="aspectFill"
              />
              <view v-else class="cover-placeholder">
                <text class="cover-emoji">{{ statusEmoji[item.status] || '赛' }}</text>
              </view>
              <view
                class="status-badge"
                :style="{ background: statusColor[item.status] || '#969799' }"
              >{{ statusText[item.status] || '未知' }}</view>
            </view>

            <!-- 信息 -->
            <view class="card-body">
              <view class="card-title">{{ item.title || item.name || '未知赛事' }}</view>

              <view class="card-info">
                <view class="info-row">
                  <text class="info-label">时间</text>
                  <text class="info-value">{{ item.start_date || '2026.09.15' }} - {{ item.end_date || '2026.09.18' }}</text>
                </view>
                <view class="info-row">
                  <text class="info-label">地点</text>
                  <text class="info-value ellipsis">{{ item.location || '深圳宝安区国际会展中心' }}</text>
                </view>
                <view class="info-row">
                  <text class="info-label">主办方</text>
                  <text class="info-value ellipsis">{{ item.organizer || '中国航空器拥有者及驾驶员协会' }}</text>
                </view>
              </view>

              <!-- 标签 -->
              <view v-if="compTags(item).length > 0" class="card-tags">
                <text
                  v-for="tag in compTags(item)"
                  :key="tag"
                  class="card-tag"
                  :style="{ background: tagBgColor(tag), color: tagTc(tag) }"
                >{{ tag }}</text>
              </view>

              <!-- 底部 -->
              <view class="card-footer">
                <template v-if="item.status === 'enrolling' || item.registration_status === 'open'">
                  <view class="price-box">
                    <text class="price-label">报名费</text>
                    <text class="price-value">¥{{ compFee(item).toLocaleString() }}</text>
                    <text class="price-unit">/人</text>
                  </view>
                  <view class="enroll-btn" @click.stop="goRegister(item)">立即报名</view>
                </template>
                <template v-else-if="item.status === 'ongoing'">
                  <text class="participants">{{ item.participants || item.registration_count || 0 }}人已参赛</text>
                </template>
                <template v-else>
                  <text class="closed-text">报名已截止</text>
                </template>
              </view>
            </view>
          </view>

          <view v-if="list.length > 0" class="load-more-wrap">
            <view v-if="loadingMore" class="loading-inline">
              <u-loading size="24rpx" />
              <text>加载更多...</text>
            </view>
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

const currentTab = ref('all')
const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

const statusColor = {
  enrolling: 'var(--color-warning)', open: 'var(--color-warning)',
  ongoing: 'var(--color-primary)',
  closed: 'var(--color-text-secondary)', full: 'var(--color-text-secondary)',
}

const statusText = {
  enrolling: '报名中', open: '报名中',
  ongoing: '进行中',
  closed: '已结束', full: '已满额',
}

const statusEmoji = {
  enrolling: '赛', open: '赛',
  ongoing: '竞',
  closed: '奖', full: '奖',
}

/* 数据映射 */

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  var tags = []
  if (item.category) tags.push(item.category)
  if (tags.length === 0) tags = ['多旋翼', '国家级']
  return tags
}

function compFee(item) {
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  if (item.price != null) return item.price
  return 380
}

function tagBgColor(tag) {
  if (['多旋翼', '固定翼', '竞速FPV', '航拍', '无人机竞技', '创新创业', '技能大赛'].indexOf(tag) >= 0) return 'var(--color-primary-light)'
  if (['国家级', '国际赛'].indexOf(tag) >= 0) return '#fff4e6'
  return '#f5f6f8'
}

function tagTc(tag) {
  if (['多旋翼', '固定翼', '竞速FPV', '航拍', '无人机竞技', '创新创业', '技能大赛'].indexOf(tag) >= 0) return 'var(--color-primary)'
  if (['国家级', '国际赛'].indexOf(tag) >= 0) return 'var(--color-warning)'
  return '#666666'
}

/* 数据获取 */

async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''

  try {
    var params = { page: page.value, page_size: pageSize }
    if (currentTab.value === 'enrolling') params.status = 'enrolling'
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/competitions', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []
    var total = (data && data.total) != null ? data.total : items.length

    if (reset) { list.value = items } else { list.value = list.value.concat(items) }
    hasMore.value = list.value.length < total
    // API 返回空数据，降级到本地 mock
    if (list.value.length === 0) {
      list.value = getMockList()
      hasMore.value = false
    }
  } catch (e) {
    // API 不可用，降级到本地 mock
    if (reset) {
      list.value = getMockList()
      hasMore.value = false
    }
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function getMockList() {
  return [
    {
      id: 'comp-1', title: '2026全国无人机职业技能大赛', status: 'enrolling',
      location: '深圳宝安区国际会展中心', organizer: '中国航空器拥有者及驾驶员协会',
      start_date: '2026.09.15', end_date: '2026.09.18',
      tags: ['多旋翼', '固定翼', '国家级'], fee: 380, registration_count: 128,
    },
    {
      id: 'comp-2', title: '首届西南无人机FPV竞速挑战赛', status: 'enrolling',
      location: '成都天府新区无人机竞速基地', organizer: '四川省航空运动协会',
      start_date: '2026.10.01', end_date: '2026.10.03',
      tags: ['竞速FPV', '多旋翼'], fee: 280, registration_count: 56,
    },
    {
      id: 'comp-3', title: '2026无人机创新应用大赛', status: 'ongoing',
      location: '北京亦庄经济技术开发区', organizer: '工信部人才交流中心',
      start_date: '2026.08.01', end_date: '2026.08.15',
      tags: ['航拍', '固定翼', '国家级'], fee: 0, registration_count: 256,
    },
    {
      id: 'comp-4', title: '青少年无人机编程挑战赛', status: 'enrolling',
      location: '上海市浦东新区青少年活动中心', organizer: '上海市教委',
      start_date: '2026.11.01', end_date: '2026.11.02',
      tags: ['多旋翼', '航拍'], fee: 120, registration_count: 340,
    },
    {
      id: 'comp-5', title: '国际无人机系统博览会竞技赛', status: 'enrolling',
      location: '广州琶洲国际会展中心', organizer: '广州市低空经济产业协会',
      start_date: '2026.12.05', end_date: '2026.12.07',
      tags: ['多旋翼', '固定翼', '国际赛'], fee: 580, registration_count: 89,
    },
    {
      id: 'comp-6', title: '2026贵州无人机应急救援演练赛', status: 'closed',
      location: '贵阳市观山湖区应急指挥中心', organizer: '贵州省应急管理厅',
      start_date: '2026.06.10', end_date: '2026.06.12',
      tags: ['多旋翼', '航拍'], fee: 0, registration_count: 72,
    },
  ]
}

function loadMore() {
  if (!loadingMore.value && hasMore.value) { page.value++; loadData(false) }
}

/* 筛选 & 搜索 */

function switchTab(tab) {
  if (currentTab.value === tab) return
  currentTab.value = tab
  loadData(true)
}

var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { loadData(true) }, 300)
}

/* 交互 */

function goDetail(item) {
  uni.navigateTo({ url: '/pages/events/detail?id=' + encodeURIComponent(item.id) })
}

function goRegister(item) {
  uni.navigateTo({ url: '/pages/events/register?id=' + encodeURIComponent(item.id) })
}

function goBack() { uni.navigateBack({ delta: 1 }) }

onLoad(function () { loadData(true) })

onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ① Banner */
.banner {
  background: linear-gradient(135deg, #d4a017, #b8860b);
  padding: 80rpx 32rpx 72rpx;
}

.banner-nav { display: flex; align-items: center; gap: 12rpx; margin-bottom: 24rpx; }

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.15);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
}

.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }
.banner-nav-title { color: rgba(255,255,255,0.9); font-size: 28rpx; font-weight: 500; }

.banner-title {
  color: #ffffff; font-size: 56rpx; font-weight: 700; line-height: 1.2; margin-bottom: 12rpx;
}

.banner-subtitle { color: rgba(255,255,255,0.7); font-size: 26rpx; font-weight: 400; }

/* ② Tab + 搜索 */
.main-card {
  background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -32rpx;
  position: relative; z-index: 2;
}

.tabs-container {
  display: flex; justify-content: center; gap: 24rpx;
  padding: 0 24rpx; margin-top: -36rpx;
}

.tab-item {
  width: 320rpx; height: 72rpx; line-height: 72rpx; text-align: center;
  border-radius: 40rpx; font-size: 28rpx; font-weight: 400;
  color: #666666; background: #ffffff;
  box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06);
}

.tab-item.active {
  background: var(--color-warning); color: #ffffff; font-weight: 600;
  box-shadow: 0 4rpx 16rpx rgba(255, 159, 10, 0.35);
}

.search-bar {
  margin: 24rpx 24rpx 0; background: var(--color-bg); border-radius: 40rpx;
  padding: 16rpx 24rpx; display: flex; align-items: center; gap: 12rpx;
}

.search-input { flex: 1; font-size: 28rpx; color: var(--color-text); }

/* ③ 卡片 */
.list-scroll { padding: 24rpx 24rpx 0; height: calc(100vh - 500rpx); box-sizing: border-box; }

.comp-card {
  background: #ffffff; border-radius: 16rpx; overflow: hidden;
  margin-bottom: 24rpx; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
}

.card-cover { height: 200rpx; position: relative; overflow: hidden; }
.cover-img { width: 100%; height: 100%; }
.cover-placeholder {
  width: 100%; height: 100%;
  background: linear-gradient(135deg, #1a365d, #2a4a7f);
  display: flex; align-items: center; justify-content: center;
}
.cover-emoji { font-size: 80rpx; opacity: 0.12; }

.status-badge {
  position: absolute; top: 16rpx; right: 16rpx;
  padding: 6rpx 18rpx; border-radius: 20rpx;
  color: #ffffff; font-size: 22rpx; font-weight: 600;
}

.card-body { padding: 24rpx 24rpx 44rpx; }

.card-title {
  font-size: 34rpx; font-weight: 600; color: var(--color-text);
  line-height: 1.4; margin-bottom: 16rpx;
}

.card-info { margin-bottom: 14rpx; }

.info-row { display: flex; align-items: center; gap: 12rpx; margin-bottom: 8rpx; }
.info-label { font-size: 24rpx; color: var(--color-text-secondary); width: 72rpx; flex-shrink: 0; }
.info-value { font-size: 26rpx; color: var(--color-text); }
.info-value.ellipsis { overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }

.card-tags { display: flex; flex-wrap: wrap; gap: 10rpx; margin-bottom: 20rpx; }
.card-tag { padding: 4rpx 14rpx; border-radius: 12rpx; font-size: 22rpx; font-weight: 500; }

.card-footer {
  display: flex; justify-content: space-between; align-items: center;
  border-top: 1rpx solid #f0f0f0; padding-top: 20rpx;
}

.price-label { font-size: 24rpx; color: var(--color-text-secondary); }
.price-value { font-size: 40rpx; font-weight: 700; color: var(--color-warning); margin: 0 8rpx; }
.price-unit { font-size: 24rpx; color: var(--color-text-secondary); }

.enroll-btn {
  padding: 12rpx 32rpx; background: var(--color-primary); color: #ffffff;
  border-radius: 24rpx; font-size: 28rpx; font-weight: 600;
}

.participants { font-size: 26rpx; color: var(--color-primary); font-weight: 500; margin-left: auto; }
.closed-text { font-size: 26rpx; color: var(--color-text-placeholder); margin-left: auto; }

.load-more-wrap { text-align: center; padding: 20rpx 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8rpx; font-size: 24rpx; color: var(--color-text-secondary); }
.no-more { font-size: 24rpx; color: var(--color-text-secondary); }
</style>
