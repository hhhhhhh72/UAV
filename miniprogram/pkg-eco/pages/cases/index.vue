<template>
  <view class="cases-page">
    <!-- ═══════ 渐变 Hero：返回 + 标题 + 实时统计 ═══════ -->
    <view class="hero" :style="{ paddingTop: (statusBarH + 4) + 'px' }">
      <view class="hero-glow hero-glow-a" />
      <view class="hero-glow hero-glow-b" />

      <view class="topbar-row">
        <view class="back-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
          <view class="back-arrow"></view>
        </view>
        <view class="topbar-center">
          <text class="top-title">企业案例</text>
          <text class="top-sub">协会认证 · 优质项目实践</text>
        </view>
        <view class="topbar-spacer"></view>
      </view>

      <!-- 统计条：案例总数 / 视频案例 / 覆盖分类（实时聚合） -->
      <view class="hero-stats">
        <view class="h-stat">
          <text class="h-stat-num">{{ totalCount }}</text>
          <text class="h-stat-label">全部案例</text>
        </view>
        <view class="h-stat-divider" />
        <view class="h-stat">
          <text class="h-stat-num">{{ videoCount }}</text>
          <text class="h-stat-label">视频案例</text>
        </view>
        <view class="h-stat-divider" />
        <view class="h-stat">
          <text class="h-stat-num">{{ catCount }}</text>
          <text class="h-stat-label">覆盖分类</text>
        </view>
      </view>
    </view>

    <!-- ═══════ 分类筛选（白底圆角胶囊，Hero 下叠压） ═══════ -->
    <view class="filter-panel">
      <scroll-view scroll-x class="chip-scroll" :show-scrollbar="false">
        <view class="chip-row">
          <view
            v-for="cat in categories"
            :key="cat.id"
            class="chip"
            :class="{ 'chip-active': activeCategory === cat.id }"
            @tap="onTabChange(cat.id)"
          >
            <text>{{ cat.name }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 加载中：骨架屏 -->
    <view v-if="loading" class="skeleton-list">
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
    </view>

    <!-- 空数据 -->
    <view v-else-if="cases.length === 0" class="state-panel">
      <view class="state-mark">
        <view class="state-mark-inner">
          <view class="state-search" />
        </view>
      </view>
      <text class="state-title">暂无相关案例</text>
      <text class="state-desc">优秀项目案例持续收录中，敬请期待</text>
    </view>

    <template v-else>
      <view class="cases-container">
        <view
          v-for="(caseItem, i) in cases"
          :key="caseItem.id"
          class="case-card"
          :style="{ animationDelay: (i * 60) + 'ms' }"
          hover-class="card-hover"
          :hover-stay-time="120"
          @tap="showCaseDetail(caseItem)"
        >
          <!-- 顶部品牌渐变条 -->
          <view class="card-strip" />

          <view class="case-cover">
            <!-- 视频封面：自动静音循环播放 + 播放指示 -->
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
            <!-- 无封面占位：品牌渐变 + CSS 装饰 -->
            <view v-else class="cover-fallback">
              <view class="fallback-ring fallback-ring-a" />
              <view class="fallback-ring fallback-ring-b" />
              <text class="fallback-text">企业案例</text>
            </view>

            <!-- 类型角标：视频/图片 -->
            <view v-if="coverTypeLabel(caseItem)" class="type-tag" :class="coverType(caseItem)">
              <view v-if="coverType(caseItem) === 'video'" class="tag-play" />
              <text>{{ coverTypeLabel(caseItem) }}</text>
            </view>
          </view>

          <view class="case-info">
            <view class="case-title">{{ caseItem.title || '未命名案例' }}</view>
            <view v-if="caseItem.description" class="case-desc">{{ caseItem.description }}</view>
            <view class="case-meta">
              <view class="meta-left">
                <text v-if="caseItem.category" class="meta-cat">{{ caseItem.category }}</text>
                <text v-if="caseItem.client_name" class="meta-client">{{ caseItem.client_name }}</text>
              </view>
              <text class="meta-date">{{ formatDate(caseItem.created_at) }}</text>
            </view>
          </view>
        </view>

        <!-- 加载更多 -->
        <view v-if="loadingMore" class="loading-more">
          <view class="loading-dot"></view>
          <text>加载中...</text>
        </view>
        <view v-else-if="finished && cases.length > 0" class="loading-more">
          <text>没有更多了</text>
        </view>
      </view>
    </template>

    <!-- ═══════ 案例详情弹窗 ═══════ -->
    <view class="detail-mask" v-if="showDetail" @tap="showDetail = false">
      <view class="detail-panel" @tap.stop v-if="currentCase">
        <view class="detail-header">
          <view class="detail-title-wrap">
            <view class="detail-bar" />
            <text class="detail-title">{{ currentCase.title || '案例详情' }}</text>
          </view>
          <view class="close-btn" hover-class="tap-fade" hover-stay-time="120" @tap="showDetail = false">
            <view class="close-x"></view>
          </view>
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
            <view class="section-head">
              <view class="head-bar" />
              <text class="section-label">案例介绍</text>
            </view>
            <text class="section-body">{{ currentCase.description || '暂无介绍' }}</text>
          </view>

          <!-- 项目成果 -->
          <view class="detail-section" v-if="currentCase.result">
            <view class="section-head">
              <view class="head-bar head-bar-teal" />
              <text class="section-label">项目成果</text>
            </view>
            <text class="section-body">{{ currentCase.result }}</text>
          </view>
        </scroll-view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const categories = [
  { id: 'all', name: '全部' },
  { id: '无人机物流', name: '无人机物流' },
  { id: '政务服务', name: '政务服务' },
  { id: '无人机吊运', name: '无人机吊运' },
  { id: '无人机表演', name: '无人机表演' },
  { id: '无人机赛事', name: '无人机赛事' }
]

const statusBarH = ref(20)
const activeCategory = ref('all')
const cases = ref([])
const loading = ref(false)
const loadingMore = ref(false)
const finished = ref(false)
const page = ref(1)
const pageSize = 10
const showDetail = ref(false)
const currentCase = ref(null)

const goBack = () => uni.navigateBack()

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

// Hero 统计：全部 / 视频 / 覆盖分类
const totalCount = computed(() => cases.value.length)
const videoCount = computed(() => cases.value.filter((c) => coverType(c) === 'video').length)
const catCount = computed(() => {
  const set = new Set()
  cases.value.forEach((c) => {
    if (c.category) set.add(c.category)
  })
  return set.size
})

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
  try {
    statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
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
.cases-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

.tap-fade { opacity: 0.85; }

/* ═══════ 渐变 Hero ═══════ */
.hero {
  position: relative;
  overflow: hidden;
  padding: 16rpx 24rpx 44rpx;
  background: linear-gradient(160deg, #074D92 0%, #0A66C2 62%, #0D7AE0 100%);
  color: #fff;
}
.hero-glow {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.hero-glow-a {
  top: -120rpx;
  right: -80rpx;
  width: 320rpx;
  height: 320rpx;
  background: rgba(255, 255, 255, 0.07);
}
.hero-glow-b {
  top: -30rpx;
  right: 10rpx;
  width: 200rpx;
  height: 200rpx;
  background: rgba(29, 212, 168, 0.12);
}
.topbar-row {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.back-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(45deg);
  margin-left: 10rpx;
}
.topbar-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
}
.top-title {
  font-size: 38rpx;
  font-weight: 700;
  letter-spacing: 2rpx;
}
.top-sub {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.78);
}
.topbar-spacer { width: 60rpx; }

/* 统计条 */
.hero-stats {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  margin-top: 26rpx;
  padding: 20rpx 8rpx 4rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.16);
}
.h-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rpx;
}
.h-stat-num {
  font-size: 36rpx;
  font-weight: 800;
  line-height: 1.1;
}
.h-stat-label {
  font-size: 20rpx;
  color: rgba(255, 255, 255, 0.72);
}
.h-stat-divider {
  width: 1rpx;
  height: 44rpx;
  background: rgba(255, 255, 255, 0.18);
}

/* ═══════ 分类筛选（白底圆角胶囊，叠压 Hero） ═══════ */
.filter-panel {
  position: relative;
  z-index: 3;
  margin: -24rpx 24rpx 0;
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.06);
  padding: 8rpx 0;
}
.chip-scroll {
  width: 100%;
  white-space: nowrap;
}
.chip-row {
  display: inline-flex;
  gap: 12rpx;
  padding: 8rpx 20rpx;
}
.chip {
  display: inline-flex;
  align-items: center;
  height: 56rpx;
  padding: 0 26rpx;
  border-radius: 999rpx;
  background: #F4F6F8;
  border: 1rpx solid #EEF1F4;
  font-size: 24rpx;
  color: #475467;
  transition: all 0.18s ease;
}
.chip-active {
  background: linear-gradient(135deg, #0A66C2, #0D7AE0);
  border-color: transparent;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 6rpx 16rpx rgba(10, 102, 194, 0.28);
}

/* ═══════ 骨架屏 ═══════ */
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 24rpx 32rpx;
}
.skeleton-card {
  height: 320rpx;
  border-radius: 16rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 空状态 ═══════ */
.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 132rpx;
  height: 132rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: linear-gradient(160deg, #EAF3FB, #F0FAF6);
  animation: floaty 3s ease-in-out infinite;
}
@keyframes floaty {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8rpx); }
}
.state-mark-inner {
  width: 92rpx;
  height: 92rpx;
  border-radius: 24rpx;
  background: #fff;
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}
/* CSS 放大镜（非 emoji） */
.state-search {
  width: 44rpx;
  height: 44rpx;
  border: 4rpx solid #0A66C2;
  border-radius: 50%;
  position: relative;
}
.state-search::after {
  content: '';
  position: absolute;
  right: -13rpx;
  bottom: -8rpx;
  width: 20rpx;
  height: 4rpx;
  border-radius: 2rpx;
  background: #0A66C2;
  transform: rotate(45deg);
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }

/* ═══════ 案例卡片 ═══════ */
.cases-container { padding: 24rpx 32rpx 0; }
.case-card {
  position: relative;
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
  margin-bottom: 24rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  border: 1rpx solid rgba(228, 231, 236, 0.7);
  animation: cardIn 0.42s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
}
@keyframes cardIn {
  from { opacity: 0; transform: translateY(24rpx); }
  to { opacity: 1; transform: translateY(0); }
}
.card-hover {
  transform: scale(0.98);
  box-shadow: 0 8px 20px rgba(16, 24, 40, 0.1);
}
.card-strip {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 5rpx;
  background: linear-gradient(90deg, #0A66C2, #1DD4A8);
  opacity: 0.85;
  z-index: 2;
}

.case-cover {
  position: relative;
  width: 100%;
  height: 340rpx;
  background: #E8EEF4;
}
.cover-img { width: 100%; height: 100%; }
.cover-video { width: 100%; height: 100%; }
/* 无封面占位：品牌渐变 + CSS 装饰 */
.cover-fallback {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #0A66C2 0%, #0D7AE0 55%, #1DD4A8 140%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}
.fallback-ring {
  position: absolute;
  border-radius: 50%;
  border: 2rpx solid rgba(255, 255, 255, 0.28);
}
.fallback-ring-a { width: 320rpx; height: 320rpx; top: -120rpx; right: -80rpx; }
.fallback-ring-b { width: 440rpx; height: 440rpx; bottom: -240rpx; left: 10%; }
.fallback-text {
  color: rgba(255, 255, 255, 0.92);
  font-size: 30rpx;
  font-weight: 600;
  letter-spacing: 4rpx;
}

/* 类型角标：视频/图片 */
.type-tag {
  position: absolute;
  top: 20rpx;
  right: 20rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 6rpx 16rpx;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 600;
  box-shadow: 0 4rpx 10rpx rgba(16, 24, 40, 0.12);
}
.type-tag.video { color: #0A66C2; }
.type-tag.image { color: #07c160; }
/* CSS 播放三角（非 emoji） */
.tag-play {
  width: 0;
  height: 0;
  border-left: 12rpx solid #0A66C2;
  border-top: 8rpx solid transparent;
  border-bottom: 8rpx solid transparent;
}

.case-info { padding: 24rpx 26rpx 26rpx; }
.case-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 10rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.case-desc {
  font-size: 24rpx;
  color: #667085;
  line-height: 1.6;
  margin-bottom: 18rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.case-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12rpx;
  padding-top: 18rpx;
  border-top: 1rpx dashed #E4E7EC;
  font-size: 22rpx;
  color: #98A2B3;
}
.meta-left {
  display: flex;
  align-items: center;
  gap: 10rpx;
  min-width: 0;
  overflow: hidden;
}
.meta-cat {
  flex-shrink: 0;
  color: #0A66C2;
  background: #EAF3FB;
  border-radius: 8rpx;
  padding: 4rpx 12rpx;
  font-size: 20rpx;
  font-weight: 500;
}
.meta-client {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #667085;
}
.meta-date { flex-shrink: 0; }

/* 加载更多 */
.loading-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 24rpx 0 12rpx;
  color: #98A2B3;
  font-size: 22rpx;
}
.loading-dot {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background: #0A66C2;
  animation: pulse 1s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 0.3; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.1); }
}

/* ═══════ 详情弹窗 ═══════ */
.detail-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(16, 24, 40, 0.55);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}
.detail-panel {
  background: #fff;
  width: 100%;
  height: 86vh;
  border-radius: 24rpx 24rpx 0 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: slideUp 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
}
@keyframes slideUp {
  from { transform: translateY(40rpx); opacity: 0.6; }
  to { transform: translateY(0); opacity: 1; }
}
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28rpx 32rpx 24rpx;
  border-bottom: 1rpx solid #F0F2F5;
}
.detail-title-wrap {
  display: flex;
  align-items: center;
  gap: 12rpx;
  min-width: 0;
  flex: 1;
}
.detail-bar {
  width: 8rpx;
  height: 30rpx;
  border-radius: 4rpx;
  background: linear-gradient(180deg, #0D7AE0, #0A66C2);
  flex-shrink: 0;
}
.detail-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.close-btn {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
/* CSS 关闭 ×（非 emoji） */
.close-x {
  width: 24rpx;
  height: 24rpx;
  position: relative;
}
.close-x::before,
.close-x::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 24rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #667085;
}
.close-x::before { transform: translate(-50%, -50%) rotate(45deg); }
.close-x::after { transform: translate(-50%, -50%) rotate(-45deg); }

.detail-scroll { flex: 1; overflow-y: auto; padding: 24rpx 32rpx 48rpx; }

/* 媒体区：竖排列表 */
.media-list { display: flex; flex-direction: column; gap: 16rpx; margin-bottom: 24rpx; }
.media-item { border-radius: 12rpx; overflow: hidden; background: #F2F3F5; }
.media-img { width: 100%; height: 380rpx; display: block; }
.media-video { width: 100%; height: 380rpx; }

.info-grid { display: flex; flex-wrap: wrap; gap: 16rpx; margin-bottom: 24rpx; }
.info-item {
  background: #F7F9FB;
  padding: 18rpx 22rpx;
  border-radius: 12rpx;
  min-width: 40%;
  flex: 1;
}
.info-label {
  font-size: 20rpx;
  color: #98A2B3;
  display: block;
  margin-bottom: 6rpx;
}
.info-value {
  font-size: 24rpx;
  color: #344054;
  font-weight: 500;
}

.detail-section { margin-bottom: 28rpx; }
.section-head {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 16rpx;
}
.head-bar {
  width: 8rpx;
  height: 28rpx;
  border-radius: 4rpx;
  background: linear-gradient(180deg, #0D7AE0, #0A66C2);
}
.head-bar-teal {
  background: linear-gradient(180deg, #2EE0B2, #1DD4A8);
}
.section-label {
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
}
.section-body {
  font-size: 24rpx;
  color: #475467;
  line-height: 1.8;
  display: block;
}
</style>
