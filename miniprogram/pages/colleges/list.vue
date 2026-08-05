<template>
  <view class="colleges-page">
    <!-- 搜索（原生导航栏为白色，标题「院校展示」） -->
    <u-sticky>
      <u-search
        v-model="keyword"
        placeholder="搜索院校名称"
        @change="onKeywordInput"
        @search="onSearch"
      />
    </u-sticky>

    <!-- 院校层次筛选 -->
    <view class="filter-bar">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-tabs">
          <view
            v-for="t in tabs"
            :key="t.value"
            class="filter-tab"
            :class="{ active: currentTab === t.value }"
            @tap="switchTab(t.value)"
          >{{ t.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 院校列表 -->
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && list.length === 0"
      empty-text="暂无院校"
      @retry="loadData"
    >
      <view class="list-body">
        <view
          v-for="item in list"
          :key="item.id"
          class="college-card"
          hover-class="tap-fade"
          @tap="goDetail(item)"
        >
          <view class="card-cover">
            <image
              v-if="coverOf(item)"
              :src="coverOf(item)"
              mode="aspectFill"
              class="cover-img"
            />
            <view v-else class="cover-placeholder">
              <text class="cover-initial">{{ initShort(item) }}</text>
            </view>
            <text class="cover-tag" :class="'cover-tag--' + collegeLevel(item)">{{ levelLabel(item) }}</text>
          </view>

          <view class="card-body">
            <text class="college-name">{{ item.name || item.title || '未知院校' }}</text>
            <text class="college-location">{{ locationText(item) }}</text>

            <!-- 关键数据 -->
            <view class="stats-bar">
              <view class="stat-item">
                <text class="stat-value">{{ (item.majors && item.majors.length) || item.majorCount || 0 }}</text>
                <text class="stat-label">无人机专业</text>
              </view>
              <view class="stat-divider" />
              <view class="stat-item">
                <text class="stat-value">{{ item.partnerCount || item.partner_count || 0 }}</text>
                <text class="stat-label">合作企业</text>
              </view>
              <view class="stat-divider" />
              <view class="stat-item">
                <text class="stat-value">{{ item.studentCount || item.student_count || 0 }}+</text>
                <text class="stat-label">在读学生</text>
              </view>
            </view>

            <text class="college-intro">{{ item.intro || item.description || '暂无简介' }}</text>

            <view v-if="specTags(item).length > 0" class="tag-row">
              <text
                v-for="tag in specTags(item)"
                :key="tag"
                class="spec-tag"
                :class="levelTagClass(tag)"
              >{{ tag }}</text>
            </view>
          </view>
        </view>

        <view v-if="list.length > 0" class="load-more">
          <view v-if="loadingMore" class="loading-inline">
            <u-loading size="24rpx" />
            <text>加载更多...</text>
          </view>
          <text v-else-if="!hasMore" class="no-more">没有更多了</text>
        </view>
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const tabs = [
  { label: '全部院校', value: 'all' },
  { label: '本科院校', value: 'undergraduate' },
  { label: '专科院校', value: 'vocational' },
]

const currentTab = ref('all')
const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

/* 院校层级：985/211=顶尖、本科、专科 */
function collegeLevel(item) {
  var tags = item.tags || []
  if (tags.indexOf('985') >= 0 || tags.indexOf('211') >= 0) return 'top'
  if (tags.indexOf('专科') >= 0 || tags.indexOf('高职') >= 0) return 'vocational'
  return 'undergraduate'
}

function levelLabel(item) {
  return { top: '985/211', undergraduate: '本科', vocational: '专科' }[collegeLevel(item)] || '本科'
}

function coverOf(item) {
  return item.cover || item.cover_image || item.image || ''
}

function initShort(item) {
  if (item.shortName) return item.shortName
  var name = item.name || ''
  return name.charAt(0) || '院'
}

function locationText(item) {
  return [item.city || '', (item.tags || []).join(' · ')].filter(Boolean).join(' · ') || '暂无位置信息'
}

function specTags(item) {
  if (Array.isArray(item.specialties) && item.specialties.length > 0) return item.specialties
  if (Array.isArray(item.majors)) return item.majors
  if (Array.isArray(item.tags)) return item.tags
  return []
}

/* 标签配色：重点资质用橙，特色专业用蓝 */
function levelTagClass(tag) {
  if (['博士点', '博士', '硕士点', '硕士', '双一流', '985', '211'].indexOf(tag) >= 0) return 'tag--hot'
  return 'tag--feature'
}

function switchTab(tab) {
  if (currentTab.value === tab) return
  currentTab.value = tab
  page.value = 1
  loadData(true)
}

var searchTimer = null
function onKeywordInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { page.value = 1; loadData(true) }, 300)
}

function onSearch() {
  clearTimeout(searchTimer)
  page.value = 1
  loadData(true)
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
  } catch (e) {
    if (reset) errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  page.value++
  loadData(false)
}

function goDetail(item) {
  uni.navigateTo({ url: '/pages/colleges/detail?id=' + encodeURIComponent(item.id) })
}

onLoad(function () { loadData(true) })

onPullDownRefresh(function () {
  loadData(true).finally(function () { uni.stopPullDownRefresh() })
})

onReachBottom(loadMore)
</script>

<style scoped>
.colleges-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 筛选 */
.filter-bar {
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 8px 12px;
}

.filter-scroll {
  white-space: nowrap;
}

.filter-tabs {
  display: inline-flex;
  gap: 8px;
}

.filter-tab {
  flex-shrink: 0;
  padding: 6px 16px;
  border-radius: 8px;
  font-size: 13px;
  color: #344054;
  background: #F4F6F8;
  border: 1px solid #EEF1F4;
}

.filter-tab.active {
  color: #fff;
  background: #0A66C2;
  border-color: #0A66C2;
  font-weight: 500;
}

/* 列表 */
.list-body {
  padding: 12px;
}

.college-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 8px;
}

.card-cover {
  width: 100%;
  height: 110px;
  position: relative;
  background: #F4F6F8;
}

.cover-img {
  width: 100%;
  height: 100%;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-initial {
  font-size: 40px;
  font-weight: 700;
  color: #0A66C2;
}

.cover-tag {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  background: #fff;
  border: 1px solid #EEF1F4;
}

.cover-tag--top { color: #E96012; }
.cover-tag--undergraduate { color: #0A66C2; }
.cover-tag--vocational { color: #168A55; }

.card-body {
  padding: 12px;
}

.college-name {
  font-size: 16px;
  font-weight: 600;
  color: #17212B;
  line-height: 1.4;
  display: block;
}

.college-location {
  font-size: 11px;
  color: #667085;
  display: block;
  margin-top: 4px;
}

/* 关键数据 */
.stats-bar {
  display: flex;
  align-items: center;
  background: #F4F6F8;
  border-radius: 8px;
  padding: 10px 0;
  margin: 10px 0;
}

.stat-item {
  flex: 1;
  text-align: center;
}

.stat-value {
  font-size: 16px;
  font-weight: 700;
  color: #17212B;
  display: block;
}

.stat-label {
  font-size: 10px;
  color: #667085;
  display: block;
  margin-top: 2px;
}

.stat-divider {
  width: 1px;
  height: 24px;
  background: #EEF1F4;
}

.college-intro {
  font-size: 12px;
  color: #344054;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 8px;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.spec-tag {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.tag--feature { background: #EAF3FB; color: #0A66C2; }
.tag--hot { background: #FFF0E6; color: #E96012; }

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.loading-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #667085;
}

.no-more {
  color: #98A2B3;
  font-size: 12px;
}

.tap-fade {
  opacity: 0.7;
}
</style>
