<template>
  <div class="cases-page">
    <van-nav-bar
      title="案例展示"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
    />

    <div class="page-content">
      <!-- 案例分类tabs -->
      <van-tabs
        :active="activeCategory"
        @update:active="(v) => (activeCategory = v)"
        sticky
        offset-top="46"
        color="#667eea"
        @change="onTabChange"
      >
        <van-tab
          v-for="category in categories"
          :key="category.id"
          :name="category.id"
          :title="category.name"
        >
          <div class="cases-container">
            <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
            <van-list
              :loading="loadingMore"
              @update:loading="(v) => (loadingMore = v)"
              :finished="finished"
              finished-text="没有更多了"
              @load="onLoad"
            >
            <!-- 案例列表 -->
            <div 
              v-for="caseItem in cases" 
              :key="caseItem.id"
              class="case-card"
              @click="showCaseDetail(caseItem)"
            >
              <!-- 案例封面 -->
              <div class="case-cover">
                <img 
                  v-if="caseItem.coverType === 'image'" 
                  :src="caseItem.cover" 
                  :alt="caseItem.title"
                />
                <div v-else class="video-cover">
                  <video 
                    :src="caseItem.cover" 
                    muted 
                    loop 
                    playsinline
                    webkit-playsinline
                    x5-playsinline
                    preload="metadata"
                    style="width: 100%; height: 100%; object-fit: cover;"
                  ></video>
                  <div class="play-icon">
                    <van-icon name="play" size="24" />
                  </div>
                </div>
                
                <!-- 类型标签 -->
                <div class="type-tag">
                  <van-tag :type="caseItem.coverType === 'video' ? 'primary' : 'success'" size="medium">
                    {{ caseItem.coverType === 'video' ? '视频' : '图片' }}
                  </van-tag>
                </div>
              </div>

              <!-- 案例信息 -->
              <div class="case-info">
                <h3 class="case-title">
                  {{ caseItem.title }}
                  <van-tag v-if="caseItem.subTag" type="warning" size="medium" style="margin-left:6px;">{{ caseItem.subTag }}</van-tag>
                </h3>
                <p class="case-desc">{{ caseItem.description }}</p>
                <div class="case-meta">
                  <span class="meta-item">
                    <van-icon name="clock-o" />
                    {{ caseItem.date }}
                  </span>
                  <span class="meta-item">
                    <van-icon name="eye-o" />
                    {{ caseItem.views }}
                  </span>
                </div>
              </div>
            </div>
            </van-list>
            </van-pull-refresh>

            <!-- 空状态 -->
            <van-empty
              v-if="!loadingMore && !refreshing && cases.length === 0"
              description="暂无案例"
              image="search"
            />
          </div>
        </van-tab>
      </van-tabs>
    </div>

    <!-- 案例详情弹窗 -->
    <van-popup
      :show="showDetail"
      @update:show="(v) => (showDetail = v)"
      position="bottom"
      :style="{ height: '90%' }"
      round
      closeable
      close-icon="close"
    >
      <div class="detail-content" v-if="currentCase">
        <h2 class="detail-title">
          {{ currentCase.title }}
          <van-tag v-if="currentCase.subTag" type="warning" size="medium" style="margin-left:8px; vertical-align:middle;">{{ currentCase.subTag }}</van-tag>
        </h2>
        
        <!-- 媒体轮播 -->
        <van-swipe
          ref="swipeRef"
          :autoplay="autoplayDuration"
          :loop="true"
          indicator-color="#667eea"
          class="media-swiper"
        >
          <van-swipe-item v-for="(media, index) in currentCase.media" :key="index">
            <!-- 图片 -->
            <div v-if="media.type === 'image'" class="media-item">
              <img :src="media.url" :alt="`案例图片${index + 1}`" />
            </div>
            
            <!-- 视频 -->
            <div v-else class="media-item video-item">
              <video
                :src="media.url"
                controls
                preload="metadata"
                playsinline
                webkit-playsinline
                x5-playsinline
                @play="onPlay"
                @pause="onPause"
                @ended="onEnded"
              >
                您的浏览器不支持视频播放
              </video>
            </div>
          </van-swipe-item>
        </van-swipe>

        <!-- 案例详情 -->
        <div class="detail-info">
          <div class="info-row">
            <van-icon name="bookmark-o" color="#667eea" />
            <span class="info-label">服务类型：</span>
            <span class="info-value">{{ currentCase.service }}</span>
          </div>
          <div class="info-row">
            <van-icon name="location-o" color="#667eea" />
            <span class="info-label">项目地点：</span>
            <span class="info-value">{{ currentCase.location }}</span>
          </div>
          <div class="info-row">
            <van-icon name="clock-o" color="#667eea" />
            <span class="info-label">完成时间：</span>
            <span class="info-value">{{ currentCase.date }}</span>
          </div>
        </div>

        <!-- 案例描述 -->
        <div class="detail-description">
          <h3 class="section-title">案例介绍</h3>
          <p class="description-text">{{ currentCase.fullDescription || currentCase.description }}</p>
        </div>

        <!-- 项目亮点 -->
        <div class="detail-highlights" v-if="currentCase.highlights">
          <h3 class="section-title">项目亮点</h3>
          <div class="highlight-item" v-for="(highlight, idx) in currentCase.highlights" :key="idx">
            <van-icon name="passed" color="#667eea" />
            <span>{{ highlight }}</span>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { showFailToast } from 'vant'

// 将接口返回的媒体地址归一化为“同域可访问”的地址
// - 兼容后端返回：http://127.0.0.1:8090/uploads/xxx、http://172.17.0.1:8090/uploads/xxx
// - 兼容缺少前导斜杠：uploads/xxx
// - 保留外链：https://xxx/yyy
const normalizeMediaUrl = (raw) => {
  if (!raw || typeof raw !== 'string') return raw
  const url = raw.trim()
  if (!url) return url
  if (url.startsWith('data:') || url.startsWith('blob:')) return url

  // 相对路径 uploads/xxx → /uploads/xxx（避免被当成 /cases/uploads/xxx）
  if (url.startsWith('uploads/')) return `/${url}`

  // 绝对路径直接返回
  if (url.startsWith('/')) return url

  // 处理后端返回的本机/容器内网地址：提取 pathname 走同域 nginx 反代
  if (url.startsWith('http://') || url.startsWith('https://')) {
    try {
      const u = new URL(url)
      const host = u.hostname
      const port = u.port
      const isLocalish =
        host === 'localhost' ||
        host === '127.0.0.1' ||
        host === '0.0.0.0' ||
        host === '172.17.0.1' ||
        port === '8090'

      if (isLocalish) {
        // 保留查询参数，避免签名/缓存参数丢失
        return `${u.pathname}${u.search}${u.hash}`
      }
    } catch {
      // ignore
    }
    return url
  }

  // 兜底：保持原样
  return url
}

const normalizeCaseItem = (item) => {
  if (!item || typeof item !== 'object') return item
  const normalized = { ...item }
  normalized.cover = normalizeMediaUrl(normalized.cover)
  if (Array.isArray(normalized.media)) {
    normalized.media = normalized.media.map((m) => ({
      ...m,
      url: normalizeMediaUrl(m?.url),
    }))
  }
  return normalized
}

// 分类（从后端 /api/case-categories 动态加载，首项固定为"全部案例"）
const categories = ref([{ id: 0, name: '全部案例' }])

const fetchCategories = async () => {
  try {
    const res = await axios.get('/api/case-categories')
    const list = Array.isArray(res.data) ? res.data : []
    categories.value = [
      { id: 0, name: '全部案例' },
      ...list.map(c => ({ id: Number(c.id), name: c.name }))
    ]
  } catch (e) {
    console.error('获取分类失败', e)
  }
}

// activeCategory 直接存 categoryId（0/1/4/5），避免“index/id 混用”导致某些 tab 拉取异常
const activeCategory = ref(0)
const showDetail = ref(false)
const currentCase = ref(null)

// 案例数据
const cases = ref([])
const page = ref(1)
const loadingMore = ref(false)
const finished = ref(false)
const refreshing = ref(false)

const fetchCases = async () => {
  try {
    const categoryId = activeCategory.value

    const params = {
      page: page.value,
      limit: 10,
    }
    // “全部案例”不传 categoryId（更通用，兼容后端按是否存在该参数来判断过滤）
    if (categoryId && String(categoryId) !== '0') {
      params.categoryId = categoryId
    }

    const res = await axios.get('/api/cases', { params })
    
    // 兼容处理：如果返回的是数组，直接用；如果是分页对象，取 .data
    let data = Array.isArray(res.data) ? res.data : (res.data.data || [])
    // 兼容媒体 URL（避免线上 https 下图片/视频加载失败）
    data = data.map(normalizeCaseItem)
    console.log('Cases API response:', res.data, 'Parsed data:', data)

    if (refreshing.value) {
        // 刷新时
        cases.value = data
        refreshing.value = false
    } else {
        // 加载更多时
        if (page.value === 1) {
             cases.value = data
        } else {
             cases.value = [...cases.value, ...data]
        }
    }

    loadingMore.value = false

    // 判断是否结束：使用解析后的 data 长度判断
    if (data.length < 10) {
        finished.value = true
    } else {
        page.value++
    }
  } catch (error) {
    console.error('Failed to fetch cases:', error)
    loadingMore.value = false
    refreshing.value = false
    finished.value = true
    showFailToast('获取案例失败')
  }
}

const onLoad = () => {
    if (refreshing.value) return
    fetchCases()
}

const onRefresh = () => {
    finished.value = false
    loadingMore.value = true 
    page.value = 1
    onLoad() 
}

const onTabChange = () => {
    cases.value = []
    page.value = 1
    finished.value = false
    loadingMore.value = true
    onLoad()
}

onMounted(async () => {
  // 优先加载分类，再拉取案例
  await fetchCategories()
  loadingMore.value = true
  onLoad()
})

// 显示案例详情
const showCaseDetail = (caseItem) => {
  console.log('Clicked case item:', caseItem);
  currentCase.value = caseItem
  showDetail.value = true
}

// 视频播放控制
const swipeRef = ref(null)
const autoplayDuration = ref(3000)

const onPlay = () => {
  autoplayDuration.value = 0 // 停止轮播
}

const onPause = () => {
  autoplayDuration.value = 3000 // 恢复轮播
}

const onEnded = () => {
  autoplayDuration.value = 3000 // 恢复轮播
  swipeRef.value?.next() // 播放结束立即切换下一张
}
</script>

<style scoped>
.cases-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 20px;
}

.page-content {
  min-height: calc(100vh - 46px);
}

/* 案例容器 */
.cases-container {
  padding: 16px;
  min-height: 400px;
}

/* 案例卡片 */
.case-card {
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  margin-bottom: 20px;
  box-shadow: 0 4px 20px rgba(100, 101, 102, 0.08);
  transition: all 0.3s ease;
  border: 1px solid rgba(255, 255, 255, 0.5);
}

.case-card:active {
  transform: scale(0.98);
  box-shadow: 0 2px 10px rgba(100, 101, 102, 0.05);
}

/* 封面 */
.case-cover {
  position: relative;
  width: 100%;
  height: 210px;
  overflow: hidden;
}

.case-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s;
}

/* 图片轻微放大效果 */
.case-card:active .case-cover img {
  transform: scale(1.05);
}

.video-cover {
  position: relative;
  width: 100%;
  height: 100%;
}

.play-icon {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(4px);
  border-radius: 50%;
  padding: 8px;
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.type-tag {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 2;
}

.type-tag :deep(.van-tag) {
  background: rgba(255, 255, 255, 0.9) !important;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.8);
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.type-tag :deep(.van-tag--primary) {
  color: #1989fa !important;
}

.type-tag :deep(.van-tag--success) {
  color: #07c160 !important;
}

/* 案例信息 */
.case-info {
  padding: 16px 20px;
}

.case-title {
  font-size: 17px;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}

.case-desc {
  font-size: 14px;
  color: #646566;
  line-height: 1.6;
  margin-bottom: 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  height: 44px; /* 固定高度保持对齐 */
}

.case-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: #969799;
  padding-top: 12px;
  border-top: 1px solid #f5f6f7;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 详情弹窗 */
.detail-content {
  padding: 24px;
  padding-bottom: 40px;
  height: 100%;
  overflow-y: auto;
  background: #fff;
}

.detail-title {
  font-size: 22px;
  font-weight: 800;
  color: #1a1a1a;
  margin-bottom: 20px;
  line-height: 1.4;
  letter-spacing: 0.5px;
}

/* 媒体轮播 */
.media-swiper {
  width: 100%;
  height: 260px;
  border-radius: 16px;
  overflow: hidden;
  margin-bottom: 24px;
  background: #f7f8fa;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.05);
}

.media-item {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.media-item img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.video-item video {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

/* 详情信息 */
.detail-info {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  border: 1px solid #edf0f3;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  font-size: 14px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  color: #646566;
  font-weight: 500;
  min-width: 70px;
}

.info-value {
  color: #323233;
  font-weight: 500;
  flex: 1;
}

/* 区块标题 */
.section-title {
  font-size: 18px;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid #667eea;
  line-height: 1;
}

/* 描述 */
.detail-description {
  margin-bottom: 28px;
}

.description-text {
  font-size: 15px;
  color: #4b4c4d;
  line-height: 1.8;
  text-align: justify;
}

/* 亮点 */
.detail-highlights .highlight-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 15px;
  color: #323233;
  line-height: 1.6;
  background: #f2f4ff;
  padding: 12px;
  border-radius: 8px;
}

.detail-highlights .highlight-item:last-child {
  margin-bottom: 0;
}
</style>

