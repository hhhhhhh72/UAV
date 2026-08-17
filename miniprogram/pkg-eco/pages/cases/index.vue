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
          <!-- 视频封面：自动静音循环播放 -->
          <video
            v-if="coverUrl(caseItem) && isVideoUrl(coverUrl(caseItem))"
            :src="coverUrl(caseItem)"
            autoplay
            muted
            loop
            object-fit="cover"
            class="cover-video"
          ></video>
          <!-- 图片封面 -->
          <image
            v-else-if="coverUrl(caseItem)"
            :src="coverUrl(caseItem)"
            mode="aspectFill"
            class="cover-img"
            lazy-load
            @error="markCoverBroken(caseItem)"
          />
          <!-- 无封面占位 -->
          <view v-else class="cover-fallback">
            <view class="fallback-icon fallback-icon-a" />
            <view class="fallback-icon fallback-icon-b" />
            <text class="fallback-text">企业案例</text>
          </view>

          <view v-if="coverTypeLabel(caseItem)" class="type-tag" :class="coverType(caseItem)">
            {{ coverTypeLabel(caseItem) }}
          </view>
        </view>

        <view class="case-info">
          <view class="case-title">{{ caseItem.title || '未命名案例' }}</view>
          <view v-if="caseItem.description" class="case-desc">{{ caseItem.description }}</view>
          <view class="case-meta">
            <view class="meta-item" v-if="caseItem.category">
              <text class="meta-label">{{ caseItem.category }}</text>
            </view>
            <view class="meta-item" v-if="caseItem.client_name">
              <text class="meta-label meta-client">{{ caseItem.client_name }}</text>
            </view>
            <view class="meta-item meta-date">{{ formatDate(caseItem.created_at) }}</view>
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
        <view class="empty-icon">
          <view class="empty-search" />
        </view>
        <view class="empty-text">暂无相关案例</view>
      </view>
    </view>

    <!-- 案例详情弹窗 -->
    <view class="detail-popup" v-if="showDetail" @tap="showDetail = false">
      <view class="detail-content" @tap.stop v-if="currentCase">
        <view class="detail-header">
          <text class="detail-title">{{ currentCase.title || '案例详情' }}</text>
          <text class="close-btn" @tap="showDetail = false">✕</text>
        </view>

        <scroll-view scroll-y class="detail-scroll">
          <!-- 媒体区：竖排（图片可预览、视频可播放；不用 swiper 避免滑动与播放冲突） -->
          <view class="media-list" v-if="mediaList(currentCase).length">
            <view v-for="(m, idx) in mediaList(currentCase)" :key="idx" class="media-item">
              <image
                v-if="m.type === 'image'"
                :src="m.url"
                mode="aspectFill"
                class="media-img"
                @tap="previewMedia(m.url)"
              />
              <video v-else :src="m.url" controls class="media-video" object-fit="contain" />
            </view>
          </view>

          <!-- 项目信息 -->
          <view class="info-grid" v-if="currentCase.category || currentCase.client_name || currentCase.created_at">
            <view class="info-item" v-if="currentCase.category">
              <text class="info-label">所属分类</text>
              <text class="info-value">{{ currentCase.category }}</text>
            </view>
            <view class="info-item" v-if="currentCase.client_name">
              <text class="info-label">服务对象</text>
              <text class="info-value">{{ currentCase.client_name }}</text>
            </view>
            <view class="info-item" v-if="currentCase.created_at">
              <text class="info-label">发布时间</text>
              <text class="info-value">{{ formatDate(currentCase.created_at) }}</text>
            </view>
          </view>

          <!-- 案例介绍 -->
          <view class="detail-section">
            <text class="section-label">案例介绍</text>
            <text class="section-body">{{ currentCase.description || '暂无介绍' }}</text>
          </view>

          <!-- 项目成果 -->
          <view class="detail-section" v-if="currentCase.result">
            <text class="section-label">项目成果</text>
            <text class="section-body">{{ currentCase.result }}</text>
          </view>
        </scroll-view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

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

// 相对路径（存库格式）→ 完整 URL（video/image 均需，缺省会直接无法加载）
const resolveUrl = (u) => {
  if (!u) return ''
  if (u.indexOf('http') === 0) return u
  return BASE_URL + u
}

// 视频识别：按扩展名（mp4/m3u8/webm），避免仅凭路径猜测
const isVideoUrl = (u) => {
  if (!u) return false
  return /\.(mp4|m3u8|webm)([?#].*)?$/i.test(u)
}

const coverUrl = (c) => {
  const first = (c.images && c.images[0]) || ''
  return resolveUrl(first)
}
const coverType = (c) => (coverUrl(c) ? (isVideoUrl(coverUrl(c)) ? 'video' : 'image') : 'none')
const coverTypeLabel = (c) => {
  const t = coverType(c)
  if (t === 'video') return '视频'
  if (t === 'image') return '图片'
  return ''
}

// 单张图片加载失败：降级为该案例无封面（显示占位）
const markCoverBroken = (c) => {
  if (c.images && c.images.length) c.images = []
}

// 详情媒体列表：全量图片/视频，统一补全 URL
const mediaList = (c) =>
  (c.images || []).map((u) => {
    const url = resolveUrl(u)
    return { url, type: isVideoUrl(url) ? 'video' : 'image' }
  })

const previewMedia = (url) => {
  const urls = mediaList(currentCase.value)
    .filter((m) => m.type === 'image')
    .map((m) => m.url)
  uni.previewImage({ current: url, urls: urls.length ? urls : [url] })
}

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return ''
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

const fetchCases = async (reset = false) => {
  if (reset) {
    page.value = 1
    finished.value = false
    cases.value = []
  }

  if (finished.value || loadingMore.value) return
  loadingMore.value = true

  try {
    const params = { page: page.value, page_size: pageSize }
    if (activeCategory.value !== 'all') {
      params.category = activeCategory.value
    }
    const res = await request({ url: '/api/v1/cases', data: params })
    // 后端分页契约：{ data, total, page, page_size }
    const list = Array.isArray(res) ? res : res?.data || res?.list || []
    const total = res?.total

    if (Array.isArray(list) && (list.length < pageSize || (typeof total === 'number' && cases.value.length + list.length >= total))) {
      finished.value = true
    }
    cases.value = reset ? list : [...cases.value, ...list]
    page.value++
  } catch (e) {
    // 加载失败：不注入假数据，展示空态
    cases.value = []
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
</script>

<style scoped>
.cases-page { min-height: 100vh; background: #f7f8fa; }

.page-header { position: sticky; top: 0; z-index: 10; background: #fff; border-bottom: 1px solid #f2f3f5; }
.tabs-scroll { white-space: nowrap; }
.tabs { display: flex; padding: 0 16px; }
.tab-item { padding: 14px 0; margin-right: 24px; font-size: 15px; color: #646566; position: relative; flex-shrink: 0; }
.tab-item.active { color: #0A66C2; font-weight: bold; }
.tab-item.active::after { content: ''; position: absolute; bottom: 0; left: 0; right: 0; height: 3px; background: #0A66C2; border-radius: 2px; }

.cases-container { padding: 16px; }

.case-card { background: #fff; border-radius: 8px; overflow: hidden; margin-bottom: 20px; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }
.case-cover { position: relative; width: 100%; height: 180px; background: #e8eef4; }
.cover-img { width: 100%; height: 100%; }
.cover-video { width: 100%; height: 100%; }
/* 无封面占位：品牌渐变 + CSS 山形装饰 */
.cover-fallback {
  width: 100%; height: 100%;
  background: linear-gradient(135deg, #0A66C2 0%, #0D7AE0 55%, #1DD4A8 140%);
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
}
.fallback-icon {
  position: absolute; border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.28);
}
.fallback-icon-a { width: 160px; height: 160px; top: -60px; right: -40px; }
.fallback-icon-b { width: 220px; height: 220px; bottom: -120px; left: 10%; }
.fallback-text { color: rgba(255, 255, 255, 0.92); font-size: 14px; font-weight: 600; letter-spacing: 2px; }

.play-overlay { position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.2); }

.type-tag { position: absolute; top: 12px; right: 12px; padding: 2px 8px; background: rgba(255, 255, 255, 0.9); border-radius: 4px; font-size: 11px; font-weight: bold; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.type-tag.video { color: #0A66C2; }
.type-tag.image { color: #07c160; }

.case-info { padding: 16px; }
.case-title { font-size: 17px; font-weight: bold; color: #1a1a1a; margin-bottom: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.case-desc { font-size: 14px; color: #646566; line-height: 1.5; margin-bottom: 12px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.case-meta { display: flex; justify-content: space-between; align-items: center; padding-top: 12px; border-top: 1px solid #f5f6f7; font-size: 12px; color: #969799; gap: 8px; }
.meta-item { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.meta-label { color: #0A66C2; }
.meta-client { color: #646566; }
.meta-date { flex-shrink: 0; }

.loading-more { text-align: center; padding: 20px 0; color: #969799; font-size: 13px; }

.empty-state { padding-top: 100px; text-align: center; }
.empty-icon { display: flex; justify-content: center; margin-bottom: 16px; }
/* CSS 放大镜（非 emoji） */
.empty-search {
  width: 40px; height: 40px;
  border: 3px solid #d5d7db;
  border-radius: 50%;
  position: relative;
}
.empty-search::after {
  content: '';
  position: absolute;
  right: -12px; bottom: -8px;
  width: 18px; height: 3px;
  border-radius: 2px;
  background: #d5d7db;
  transform: rotate(45deg);
}
.empty-text { color: #969799; font-size: 14px; }

/* 详情弹窗 */
.detail-popup { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); z-index: 1000; display: flex; align-items: flex-end; }
.detail-content { background: #fff; width: 100%; height: 85vh; border-radius: 8px 16px 0 0; overflow: hidden; display: flex; flex-direction: column; }
.detail-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid #f2f3f5; }
.detail-title { font-size: 17px; font-weight: bold; color: #323233; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.close-btn { font-size: 18px; color: #969799; padding: 4px 8px; }
.detail-scroll { flex: 1; overflow-y: auto; padding: 16px 20px 40px; }

/* 媒体区：竖排列表 */
.media-list { display: flex; flex-direction: column; gap: 14px; margin-bottom: 16px; }
.media-item { border-radius: 8px; overflow: hidden; background: #f2f3f5; }
.media-img { width: 100%; height: 200px; display: block; }
.media-video { width: 100%; height: 200px; }

.info-grid { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 16px; }
.info-item { background: #f7f8fa; padding: 12px; border-radius: 8px; min-width: 45%; flex: 1; }
.info-label { font-size: 12px; color: #969799; display: block; margin-bottom: 4px; }
.info-value { font-size: 14px; color: #323233; font-weight: 500; }

.detail-section { margin-bottom: 16px; }
.section-label { font-size: 15px; font-weight: bold; color: #323233; margin-bottom: 8px; display: block; }
.section-body { font-size: 14px; color: #646566; line-height: 1.8; display: block; }
</style>
