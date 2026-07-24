<template>
  <view class="cases-page">
    <view class="page-header">
      <scroll-view scroll-x class="tabs-scroll">
        <view class="tabs">
          <view 
            v-for="cat in categories" 
            :key="cat.id" 
            class="tab-item"
            :class="{ active: activeCategory === cat.id }"
            @tap="onTabChange(cat.id)"
          >
            {{ cat.name }}
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="cases-container">
      <view 
        v-for="caseItem in cases" 
        :key="caseItem.id"
        class="case-card"
        @tap="showCaseDetail(caseItem)"
      >
        <view class="case-cover">
          <image 
            v-if="caseItem.coverType === 'image'" 
            :src="caseItem.cover" 
            mode="aspectFill" 
            class="cover-img" 
            lazy-load 
          />
          <view v-else class="video-cover">
            <video 
              :src="caseItem.cover" 
              muted 
              loop 
              style="width: 100%; height: 100%; object-fit: cover;"
            ></video>
            <view class="play-overlay">
              <text class="play-icon-text">▶</text>
            </view>
          </view>
          <view class="type-tag" :class="caseItem.coverType">
            {{ caseItem.coverType === 'video' ? '视频' : '图片' }}
          </view>
        </view>

        <view class="case-info">
          <view class="case-title">{{ caseItem.title }}</view>
          <view class="case-desc">{{ caseItem.description }}</view>
          <view class="case-meta">
            <view class="meta-item">
              <text class="meta-icon">🕒</text>
              <text>{{ caseItem.date }}</text>
            </view>
            <view class="meta-item">
              <text class="meta-icon">👁️</text>
              <text>{{ caseItem.views }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="loadingMore" class="loading-more">
        <text>加载中...</text>
      </view>
      <view v-else-if="finished && cases.length > 0" class="loading-more">
        <text>没有更多了</text>
      </view>

      <view v-if="!loadingMore && cases.length === 0" class="empty-state">
        <view class="empty-icon">🔍</view>
        <view class="empty-text">暂无相关案例</view>
      </view>
    </view>

    <!-- 案例详情弹窗 -->
    <view class="detail-popup" v-if="showDetail" @tap="showDetail = false">
      <view class="detail-content" @tap.stop v-if="currentCase">
        <view class="detail-header">
          <text class="detail-title">{{ currentCase.title }}</text>
          <text class="close-btn" @tap="showDetail = false">✕</text>
        </view>

        <scroll-view scroll-y class="detail-scroll">
          <!-- 媒体区 -->
          <view class="media-section" v-if="currentCase.media && currentCase.media.length">
            <swiper class="media-swipe" autoplay circular indicator-dots>
              <swiper-item v-for="(m, idx) in currentCase.media" :key="idx">
                <image v-if="m.type === 'image'" :src="m.url" mode="aspectFill" class="media-img" @tap="previewMedia(m.url)" />
                <video v-else :src="m.url" controls class="media-video"></video>
              </swiper-item>
            </swiper>
          </view>

          <!-- 项目信息 -->
          <view class="info-grid">
            <view class="info-item" v-if="currentCase.service">
              <text class="info-label">所属服务</text>
              <text class="info-value">{{ currentCase.service }}</text>
            </view>
            <view class="info-item" v-if="currentCase.location">
              <text class="info-label">实施地点</text>
              <text class="info-value">{{ currentCase.location }}</text>
            </view>
            <view class="info-item" v-if="currentCase.date">
              <text class="info-label">实施日期</text>
              <text class="info-value">{{ currentCase.date }}</text>
            </view>
          </view>

          <!-- 案例介绍 -->
          <view class="detail-section">
            <text class="section-label">案例介绍</text>
            <text class="section-body">{{ currentCase.description }}</text>
          </view>

          <!-- 亮点 -->
          <view class="detail-section" v-if="currentCase.highlights && currentCase.highlights.length">
            <text class="section-label">项目亮点</text>
            <view class="highlight-list">
              <view v-for="(h, idx) in currentCase.highlights" :key="idx" class="highlight-item">
                <text class="highlight-icon">✓</text>
                <text>{{ h }}</text>
              </view>
            </view>
          </view>
        </scroll-view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import { caseList as localCaseList } from '../../utils/cases'

const categories = [
  { id: 'all', name: '全部案例' },
  { id: '无人机物流', name: '无人机物流' },
  { id: '政务服务', name: '政务服务' },
  { id: '无人机吊运', name: '无人机吊运' },
  { id: '无人机表演', name: '无人机表演' },
  { id: '无人机赛事', name: '无人机赛事' }
]

const activeCategory = ref('all')
const cases = ref([])
const loadingMore = ref(false)
const finished = ref(false)
const page = ref(1)
const pageSize = 10
const showDetail = ref(false)
const currentCase = ref(null)

const fetchCases = async (reset = false) => {
  if (reset) {
    page.value = 1
    finished.value = false
    cases.value = []
  }

  if (finished.value || loadingMore.value) return
  loadingMore.value = true

  try {
    const params = { page: page.value, pageSize }
    if (activeCategory.value !== 'all') {
      params.category = activeCategory.value
    }
    const res = await request({ url: '/api/cases', data: params })
    const list = res?.data || res?.list || (Array.isArray(res) ? res : [])

    if (list.length < pageSize) {
      finished.value = true
    }
    cases.value = reset ? list : [...cases.value, ...list]
    page.value++
  } catch (e) {
    let filtered = localCaseList
    if (activeCategory.value !== 'all') {
      filtered = localCaseList.filter(c => c.service === activeCategory.value)
    }
    cases.value = filtered
    finished.value = true
  } finally {
    loadingMore.value = false
  }
}

const onTabChange = (id) => {
  activeCategory.value = id
  fetchCases(true)
}

onMounted(() => {
  fetchCases(true)
})

onPullDownRefresh(() => {
  fetchCases(true).then(() => {
    uni.stopPullDownRefresh()
  })
})

onReachBottom(() => {
  if (!finished.value && !loadingMore.value) {
    fetchCases()
  }
})

const showCaseDetail = (item) => {
  currentCase.value = item
  showDetail.value = true
}

const previewMedia = (url) => {
  const urls = (currentCase.value?.media || [])
    .filter(m => m.type === 'image')
    .map(m => m.url)
  uni.previewImage({ current: url, urls: urls.length ? urls : [url] })
}
</script>

<style scoped>
.cases-page { min-height: 100vh; background: #f7f8fa; }

.page-header { position: sticky; top: 0; z-index: 10; background: #fff; border-bottom: 1px solid #f2f3f5; }
.tabs-scroll { white-space: nowrap; }
.tabs { display: flex; padding: 0 16px; }
.tab-item { padding: 14px 0; margin-right: 24px; font-size: 15px; color: #646566; position: relative; flex-shrink: 0; }
.tab-item.active { color: #667eea; font-weight: bold; }
.tab-item.active::after { content: ''; position: absolute; bottom: 0; left: 0; right: 0; height: 3px; background: #667eea; border-radius: 2px; }

.cases-container { padding: 16px; }

.case-card { background: #fff; border-radius: 16px; overflow: hidden; margin-bottom: 20px; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }
.case-cover { position: relative; width: 100%; height: 180px; }
.cover-img { width: 100%; height: 100%; }
.video-cover { width: 100%; height: 100%; position: relative; }
.play-overlay { position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.2); }
.play-icon-text { font-size: 32px; color: #fff; }

.type-tag { position: absolute; top: 12px; right: 12px; padding: 2px 8px; background: rgba(255, 255, 255, 0.9); border-radius: 4px; font-size: 11px; font-weight: bold; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.type-tag.video { color: #1989fa; }
.type-tag.image { color: #07c160; }

.case-info { padding: 16px; }
.case-title { font-size: 17px; font-weight: bold; color: #1a1a1a; margin-bottom: 8px; }
.case-desc { font-size: 14px; color: #646566; line-height: 1.5; margin-bottom: 12px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; height: 42px; }
.case-meta { display: flex; justify-content: space-between; padding-top: 12px; border-top: 1px solid #f5f6f7; font-size: 12px; color: #969799; }
.meta-item { display: flex; align-items: center; gap: 4px; }

.loading-more { text-align: center; padding: 20px 0; color: #969799; font-size: 13px; }

.empty-state { padding-top: 100px; text-align: center; }
.empty-icon { font-size: 48px; margin-bottom: 16px; }
.empty-text { color: #969799; font-size: 14px; }

/* 详情弹窗 */
.detail-popup { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); z-index: 1000; display: flex; align-items: flex-end; }
.detail-content { background: #fff; width: 100%; height: 85vh; border-radius: 16px 16px 0 0; overflow: hidden; display: flex; flex-direction: column; }
.detail-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid #f2f3f5; }
.detail-title { font-size: 17px; font-weight: bold; color: #323233; flex: 1; }
.close-btn { font-size: 18px; color: #969799; padding: 4px 8px; }
.detail-scroll { flex: 1; overflow-y: auto; padding: 16px 20px 40px; }

.media-swipe { height: 220px; border-radius: 12px; overflow: hidden; margin-bottom: 16px; }
.media-img { width: 100%; height: 100%; }
.media-video { width: 100%; height: 100%; }

.info-grid { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 16px; }
.info-item { background: #f7f8fa; padding: 12px; border-radius: 8px; min-width: 45%; flex: 1; }
.info-label { font-size: 12px; color: #969799; display: block; margin-bottom: 4px; }
.info-value { font-size: 14px; color: #323233; font-weight: 500; }

.detail-section { margin-bottom: 16px; }
.section-label { font-size: 15px; font-weight: bold; color: #323233; margin-bottom: 8px; display: block; }
.section-body { font-size: 14px; color: #646566; line-height: 1.8; display: block; }

.highlight-list { display: flex; flex-direction: column; gap: 8px; }
.highlight-item { display: flex; align-items: center; gap: 8px; font-size: 14px; color: #646566; }
.highlight-icon { color: #07c160; font-weight: bold; }
</style>
