<template>
  <view class="page">
    <!-- ① 学术蓝 Banner -->
    <view class="banner">
      <view class="banner-nav">
        <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
        <text class="banner-nav-title">院校展示</text>
      </view>
      <text class="banner-title">无人机专业院校</text>
      <text class="banner-subtitle">产学研融合 · 人才培养摇篮</text>
    </view>

    <!-- ② Tab + 搜索 -->
    <view class="main-card">
      <view class="tabs-container">
        <view class="tab-item" :class="{ active: currentTab === 'all' }" @click="switchTab('all')">全部院校</view>
        <view class="tab-item" :class="{ active: currentTab === 'undergraduate' }" @click="switchTab('undergraduate')">本科院校</view>
        <view class="tab-item" :class="{ active: currentTab === 'vocational' }" @click="switchTab('vocational')">专科院校</view>
      </view>

      <view class="search-bar">
        <u-icon name="search" size="28rpx" color="#969799" />
        <input class="search-input" v-model="keyword" placeholder="搜索院校名称" @input="onSearch" />
      </view>

      <!-- ③ 院校卡片 -->
      <StateView
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && list.length === 0"
        empty-text="暂无院校"
        @retry="loadData"
      >
        <scroll-view class="list-scroll" scroll-y @scrolltolower="loadMore">
          <view v-for="item in list" :key="item.id" class="college-card" @click="goDetail(item)">
            <!-- 封面 -->
            <view class="card-cover">
              <image v-if="item.cover || item.cover_image || item.image" :src="item.cover || item.cover_image || item.image" class="cover-img" mode="aspectFill" />
              <view v-else class="cover-placeholder"><text class="cover-emoji">校</text></view>
            </view>

            <view class="card-body">
              <!-- 头像 + 名称 -->
              <view class="card-header">
                <view class="college-avatar">{{ initShort(item) }}</view>
                <view class="header-info">
                  <text class="college-name">{{ item.name || item.title || '未知院校' }}</text>
                  <text class="college-location">{{ item.city || '未知城市' }} · {{ (item.tags || ['无人机专业']).join(' · ') }}</text>
                </view>
              </view>

              <!-- 数据统计条 -->
              <view class="stats-bar">
                <view class="stat-item">
                  <text class="stat-value">{{ item.majorCount || item.major_count || '6' }}</text>
                  <text class="stat-label">无人机专业</text>
                </view>
                <view class="stat-divider" />
                <view class="stat-item">
                  <text class="stat-value">{{ item.partnerCount || item.partner_count || '28' }}</text>
                  <text class="stat-label">合作企业</text>
                </view>
                <view class="stat-divider" />
                <view class="stat-item">
                  <text class="stat-value">{{ item.studentCount || item.student_count || '320' }}+</text>
                  <text class="stat-label">在读学生</text>
                </view>
              </view>

              <!-- 简介 -->
              <text class="college-intro">{{ item.intro || item.description || '暂无简介' }}</text>

              <!-- 专业标签 -->
              <view v-if="specTags(item).length > 0" class="tag-row">
                <text v-for="tag in specTags(item)" :key="tag" class="spec-tag"
                  :style="{ background: tagBgColor(tag), color: tagTc(tag) }">{{ tag }}</text>
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

function tagBgColor(tag) {
  if (['博士点', '硕士点', '双一流'].indexOf(tag) >= 0) return '#fff4e6'
  return 'var(--color-primary-light)'
}

function tagTc(tag) {
  if (['博士点', '硕士点', '双一流'].indexOf(tag) >= 0) return 'var(--color-warning)'
  return 'var(--color-primary)'
}

function initShort(item) {
  if (item.shortName) return item.shortName
  var name = item.name || ''
  return name.charAt(0) || '院'
}

function specTags(item) {
  if (Array.isArray(item.specialties) && item.specialties.length > 0) return item.specialties
  if (Array.isArray(item.majors)) return item.majors
  if (item.tags) return item.tags
  return ['飞行器设计', '无人机工程']
}

function switchTab(tab) {
  if (currentTab.value === tab) return
  currentTab.value = tab
  page.value = 1
  loadData(true)
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
    if (currentTab.value !== 'all') params.type = currentTab.value
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/colleges', data: params })
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
    { id: 'college-1', name: '北京航空航天大学', city: '北京市', tags: ['双一流', '985'], cover: '', majorCount: 6, partnerCount: 28, studentCount: '320', intro: '航空科学与工程学院是国内顶尖的航空航天教育基地，拥有无人机系统设计与控制工程等6个本科专业方向。', specialties: ['飞行器设计', '无人机工程', '博士点'] },
    { id: 'college-2', name: '南京航空航天大学', city: '南京市', tags: ['双一流', '211'], cover: '', majorCount: 5, partnerCount: 22, studentCount: '280', intro: '民航学院是首批设立无人机应用技术专业的高校之一，与多家无人机企业共建产学研基地。', specialties: ['无人机应用', '适航技术', '硕士点'] },
    { id: 'college-3', name: '西北工业大学', city: '西安市', tags: ['双一流', '985'], cover: '', majorCount: 7, partnerCount: 35, studentCount: '450', intro: '无人机特种技术重点实验室依托单位，在军用和民用无人机领域均有深厚的研究积累。', specialties: ['飞行控制', '无人机系统', '博士点'] },
    { id: 'college-4', name: '成都航空职业技术学院', city: '成都市', tags: ['专科', '示范校'], cover: '', majorCount: 3, partnerCount: 15, studentCount: '180', intro: '西南地区最早开设无人机应用技术专业的高职院校，与成飞、成发等企业深度合作。', specialties: ['无人机装调', '航拍测绘'] },
    { id: 'college-5', name: '长沙航空职业技术学院', city: '长沙市', tags: ['专科', '示范校'], cover: '', majorCount: 4, partnerCount: 18, studentCount: '210', intro: '与中航工业、中国航发等企业共建实训基地，注重实操能力培养。', specialties: ['无人机装调', '农业植保', '巡检'] },
    { id: 'college-6', name: '中国民航大学', city: '天津市', tags: ['双一流'], cover: '', majorCount: 5, partnerCount: 25, studentCount: '350', intro: '民航系统唯一博士学位授予单位，设有无人机适航与运行管理专业方向。', specialties: ['适航管理', '无人机运行', '硕士点'] },
  ]
}

function loadMore() {
  if (!loadingMore.value && hasMore.value) { page.value++; loadData(false) }
}

function goDetail(item) {
  uni.navigateTo({ url: '/pages/colleges/detail?id=' + encodeURIComponent(item.id) })
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
.banner { background: linear-gradient(135deg, var(--color-primary), #1565c0); padding: 80rpx 32rpx 72rpx; }
.banner-nav { display: flex; align-items: center; gap: 12rpx; margin-bottom: 24rpx; }

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.15);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
}

.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }
.banner-nav-title { color: rgba(255,255,255,0.9); font-size: 28rpx; font-weight: 500; }

.banner-title { color: #ffffff; font-size: 56rpx; font-weight: 700; line-height: 1.2; margin-bottom: 12rpx; }
.banner-subtitle { color: rgba(255,255,255,0.7); font-size: 26rpx; font-weight: 400; }

/* ② Tab + 搜索 */
.main-card { background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -32rpx; position: relative; z-index: 2; }

.tabs-container { display: flex; justify-content: center; gap: 16rpx; padding: 0 24rpx; margin-top: -36rpx; }

.tab-item {
  width: 200rpx; height: 72rpx; line-height: 72rpx; text-align: center;
  border-radius: 40rpx; font-size: 26rpx; font-weight: 400;
  color: #666666; background: #ffffff; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06);
}

.tab-item.active {
  background: var(--color-primary); color: #ffffff; font-weight: 600;
  box-shadow: 0 4rpx 16rpx rgba(10,102,194,0.35);
}

.search-bar {
  margin: 24rpx 24rpx 0; background: var(--color-bg); border-radius: 40rpx;
  padding: 16rpx 24rpx; display: flex; align-items: center; gap: 12rpx;
}

.search-input { flex: 1; font-size: 28rpx; color: var(--color-text); }

/* ③ 卡片 */
.list-scroll { padding: 24rpx 24rpx 0; height: calc(100vh - 480rpx); }

.college-card {
  background: #ffffff; border-radius: 20rpx; overflow: hidden;
  margin-bottom: 24rpx; box-shadow: 0 2rpx 16rpx rgba(0,0,0,0.04);
}

.card-cover { height: 200rpx; position: relative; overflow: hidden; }
.cover-img { width: 100%; height: 100%; }

.cover-placeholder {
  width: 100%; height: 100%; background: linear-gradient(135deg, var(--color-primary), #1976d2);
  display: flex; align-items: center; justify-content: center;
}

.cover-emoji { font-size: 80rpx; opacity: 0.12; }
.card-body { padding: 24rpx; }

.card-header { display: flex; align-items: center; gap: 16rpx; margin-bottom: 20rpx; }

.college-avatar {
  width: 88rpx; height: 88rpx; background: var(--color-primary); border-radius: 20rpx;
  display: flex; align-items: center; justify-content: center;
  color: #ffffff; font-size: 40rpx; font-weight: 600; flex-shrink: 0;
}

.header-info { flex: 1; min-width: 0; }

.college-name { font-size: 32rpx; font-weight: 600; color: var(--color-text); display: block; line-height: 1.3; }
.college-location { font-size: 24rpx; color: var(--color-primary); font-weight: 500; margin-top: 4rpx; display: block; }

/* 数据条 */
.stats-bar {
  display: flex; align-items: center; padding: 20rpx 16rpx;
  background: #f8fafc; border-radius: 16rpx; margin-bottom: 18rpx;
}

.stat-item { flex: 1; text-align: center; }
.stat-value { font-size: 36rpx; font-weight: 700; color: var(--color-primary); display: block; }
.stat-label { font-size: 22rpx; color: var(--color-text-secondary); display: block; margin-top: 4rpx; }
.stat-divider { width: 2rpx; height: 40rpx; background: #e0e4e8; }

.college-intro {
  font-size: 26rpx; color: #4a4a4a; line-height: 1.6;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
  overflow: hidden; margin-bottom: 16rpx;
}

.tag-row { display: flex; flex-wrap: wrap; gap: 10rpx; }
.spec-tag { padding: 6rpx 18rpx; border-radius: 12rpx; font-size: 22rpx; font-weight: 500; }

.load-more-wrap { text-align: center; padding: 20rpx 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8rpx; font-size: 24rpx; color: var(--color-text-secondary); }
.no-more { font-size: 24rpx; color: var(--color-text-secondary); }
</style>
