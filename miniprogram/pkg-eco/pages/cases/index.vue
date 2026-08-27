<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarH + 44) + 'px' }">
    <u-nav-bar title="企业案例" show-back :fixed="true" @back="goBack" />

    <!-- 吸顶分类筛选（对齐成果库：下划线 tab 分段 + ▾ 分类浮层面板） -->
    <view class="sticky-head" :style="{ top: (statusBarH + 44) + 'px' }">
      <view class="stage-wrap">
        <scroll-view scroll-x class="chip-scroll" :show-scrollbar="false">
          <view class="stages">
            <view
              v-for="cat in categories"
              :key="cat.id"
              class="stg"
              :class="{ on: activeCategory === cat.id }"
              @tap="pickStageTab(cat.id)"
            >
              <text>{{ cat.name }}</text>
              <!-- ▾ 独立面板开关（方案 A）：未停在「全部」时点「全部」先清分类；已停时再点开面板 -->
              <text v-if="cat.id === 'all'" class="stg-arr" :class="{ up: panel === 'all' }" @tap.stop="togglePanel">▾</text>
            </view>
          </view>
        </scroll-view>
        <!-- 分类面板：absolute 浮层（同成果库），展开时不挤动下方内容 -->
        <view v-if="panel === 'all'" class="field-panel" :class="{ closing }">
          <view class="p-group">分类</view>
          <view class="p-chips">
            <text
              v-for="cat in categories"
              :key="cat.id"
              class="p-chip"
              :class="{ act: activeCategory === cat.id }"
              @tap="pickCategory(cat.id)"
            >{{ cat.name }}</text>
          </view>
        </view>
      </view>
    </view>
    <!-- 蒙层：从 tab 分段底部开始置灰（top 由 maskTop 实测），点击外部退场收起 -->
    <view v-if="panel" class="panel-mask" :style="{ top: maskTop + 'px' }" @tap="startClosePanel" />

    <!-- Banner 渐变卡：实时统计 -->
    <view class="banner">
      <view class="banner-top">
        <view class="banner-icon">案</view>
        <view class="banner-info">
          <text class="banner-title">行业标杆，实战见证</text>
          <text class="banner-sub">协会认证 · 优质项目实践</text>
        </view>
      </view>
      <view class="banner-stats">
        <view class="bs"><text class="bs-num">{{ totalCount }}</text><text class="bs-lb">全部案例</text></view>
        <view class="bs-div" />
        <view class="bs"><text class="bs-num">{{ videoCount }}</text><text class="bs-lb">视频案例</text></view>
        <view class="bs-div" />
        <view class="bs"><text class="bs-num">{{ catCount }}</text><text class="bs-lb">覆盖分类</text></view>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 + 当前状态 -->
      <view class="ir">
        <text>共 <text class="irn">{{ totalCount }}</text> 项案例</text>
        <text class="ir-hint">{{ irHint }}</text>
      </view>

      <!-- 骨架 -->
      <view v-if="loading" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-cover"></view>
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w70"></view>
          </view>
        </view>
      </view>

      <!-- 错误 -->
      <view v-else-if="err && !cases.length" class="st">
        <u-empty description="加载失败，请检查网络">
          <view class="stb" @tap="fetchCases(true)">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="!cases.length" class="st">
        <u-empty description="暂无相关案例">
          <text class="sth">优秀项目案例持续收录中，敬请期待</text>
          <view v-if="activeCategory !== 'all'" class="stb" @tap="pickCategory('all')">查看全部</view>
        </u-empty>
      </view>

      <!-- 列表：卡片（封面 + 徽章 + 标题 + 描述 + 元信息，无左缘色条） -->
      <view v-else class="cl">
        <view
          v-for="caseItem in cases"
          :key="caseItem.id"
          class="card"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="showCaseDetail(caseItem)"
        >

          <!-- 封面：视频自动播放 / 图片 / 渐变占位 -->
          <view class="case-cover">
            <video
              v-if="coverUrl(caseItem) && isVideoUrl(coverUrl(caseItem))"
              :src="coverUrl(caseItem)"
              autoplay
              muted
              loop
              object-fit="cover"
              class="cover-video"
            ></video>
            <image
              v-else-if="coverUrl(caseItem)"
              :src="coverUrl(caseItem)"
              mode="aspectFill"
              class="cover-img"
              lazy-load
              @error="markCoverBroken(caseItem)"
            />
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
            <view v-if="caseItem.category" class="c-badges">
              <text class="c-tag" :style="{ color: catColor(caseItem.category).tagC, background: catColor(caseItem.category).tagBg }">{{ caseItem.category }}</text>
            </view>
            <text class="ct">{{ caseItem.title || '未命名案例' }}</text>
            <text v-if="caseItem.description" class="c-desc">{{ caseItem.description }}</text>
            <view class="c-meta">
              <text v-if="caseItem.client_name" class="c-client">{{ caseItem.client_name }}</text>
              <text v-if="caseItem.client_name" class="c-dot">·</text>
              <text>{{ formatDate(caseItem.created_at) }}</text>
            </view>
          </view>
        </view>

        <!-- 加载更多 -->
        <view v-if="loadingMore" class="lm">— 加载中 —</view>
        <view v-else-if="finished && cases.length" class="lm">— 没有更多了 —</view>
      </view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>

    <!-- ═══════ 案例详情弹窗（功能保留） ═══════ -->
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
          <!-- 媒体区：竖排（图片可预览、视频可播放） -->
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
import { onPullDownRefresh, onReachBottom, onPageScroll } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'
import { useReduceMotion } from '../../../utils/motion'

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
// 筛选面板（对齐成果库：tab 分段 + ▾ 浮层面板 + 蒙层）
const panel = ref('')       // '' = 收起；'all' = 分类面板展开
const closing = ref(false)  // 面板退场中（先播退场动画再 v-if 移除）
const maskTop = ref(0)      // 蒙层起点（面板打开时实测：tab 分段底部）
let panelCloseT = null
const PANEL_CLOSE_MS = 210 // 退场动画 .21s ease-in（= 进场 ×0.7）
const loading = ref(false)
const loadingMore = ref(false)
const finished = ref(false)
const err = ref(false)
const page = ref(1)
const pageSize = 10
const showDetail = ref(false)
const currentCase = ref(null)
const showBt = ref(false)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）

/* 分类强调色：卡片左缘色条 + 分类 tag（对比度 ≥4.5:1） */
const CAT_COLOR = {
  '无人机物流': { tagC: '#0d47a1', tagBg: '#E3EDF9' },
  '政务服务': { tagC: '#1a237e', tagBg: '#E7E9F4' },
  '无人机吊运': { tagC: '#B54708', tagBg: '#FDEEE4' },
  '无人机表演': { tagC: '#4a148c', tagBg: '#F0E9F7' },
  '无人机赛事': { tagC: '#b71c1c', tagBg: '#FBE9E9' },
}
const CAT_COLOR_DEFAULT = { tagC: '#344054', tagBg: '#EEF1F4' }
const catColor = (c) => CAT_COLOR[c] || CAT_COLOR_DEFAULT

const goBack = () => uni.navigateBack()
const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })

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

// Banner 统计：全部 / 视频 / 覆盖分类
const totalCount = computed(() => cases.value.length)
const videoCount = computed(() => cases.value.filter((c) => coverType(c) === 'video').length)
const catCount = computed(() => {
  const set = new Set()
  cases.value.forEach((c) => {
    if (c.category) set.add(c.category)
  })
  return set.size
})

// 信息行右侧：当前分类 / 加载状态提示
const irHint = computed(() => {
  if (activeCategory.value !== 'all') return activeCategory.value
  return finished.value ? '已加载全部' : '上拉加载更多'
})

const fetchCases = async (reset = false, silent = false) => {
  if (reset) {
    page.value = 1
    finished.value = false
    cases.value = []
    err.value = false
    // silent（下拉刷新）：保留当前列表，避免骨架屏顶替闪烁
    if (!silent) loading.value = true
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
    // 加载失败：不注入假数据，展示错误/空态
    err.value = true
    cases.value = []
    finished.value = true
  } finally {
    loadingMore.value = false
    loading.value = false
  }
}

// ---- 筛选交互（tab 分段 + 「全部」分类面板，方案 A，同成果库） ----
const measureMaskTop = () => {
  // 蒙层起点 = 分段容器底部（吸顶头部内），面板打开时实测，内容自适应不错位
  uni.createSelectorQuery().select('.stage-wrap').boundingClientRect((rect) => {
    if (rect && rect.bottom) maskTop.value = Math.round(rect.bottom)
  }).exec()
}
const startClosePanel = () => {
  if (closing.value) return // 已在退场中，防重复触发叠加定时器
  closing.value = true
  clearTimeout(panelCloseT)
  panelCloseT = setTimeout(() => { panel.value = ''; closing.value = false; panelCloseT = null }, PANEL_CLOSE_MS)
}
const togglePanel = () => {
  if (panel.value === 'all') { startClosePanel(); return } // 再点「全部」→ 退场收起
  clearTimeout(panelCloseT); panelCloseT = null; closing.value = false
  panel.value = 'all'
  measureMaskTop()
}
// 方案 A（同成果库）：非全部 tab 再点取消；「全部」未停时先清筛、已停时开面板；▾ 独立开关
const pickStageTab = (k) => {
  if (k !== 'all') {
    startClosePanel()
    activeCategory.value = activeCategory.value === k ? 'all' : k
    fetchCases(true)
    return
  }
  if (activeCategory.value !== 'all') {
    startClosePanel()
    activeCategory.value = 'all'
    fetchCases(true)
    return
  }
  togglePanel()
}
// 面板 chip 点选即筛（选中即时高亮），再点一次取消；「全部」chip 恒为全部
const pickCategory = (c) => {
  activeCategory.value = activeCategory.value === c ? 'all' : c
  startClosePanel()
  fetchCases(true)
}

onMounted(() => {
  try {
    statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
  checkMotion()
  fetchCases(true)
})

onPullDownRefresh(() => {
  fetchCases(true, true).then(() => {
    uni.stopPullDownRefresh()
  })
})

onReachBottom(() => {
  if (!finished.value && !loadingMore.value) {
    fetchCases()
  }
})

onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})

const showCaseDetail = (item) => {
  currentCase.value = item
  showDetail.value = true
}
</script>

<style>
page {
  background: #fff;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 40px;
}

.tap-fade { opacity: 0.85; }

/* ===== 吸顶分类筛选（对齐成果库：下划线 tab 分段 + ▾ 面板 + 蒙层） ===== */
.sticky-head {
  position: sticky;
  z-index: 40;
  background: #fff;
}
/* 分类多（6 项）超一行：横向滑，全部 tab 可达（原页面即横滑交互；参考页 5 项不换行故用普通 view） */
.chip-scroll { width: 100%; white-space: nowrap; }
.stage-wrap {
  position: relative;
  z-index: 42;
  background: #fff;
}
.stages {
  display: flex;
  gap: 40rpx;
  padding: 4rpx 28rpx 16rpx;
  white-space: nowrap;
}
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 0 8rpx;
  font-size: 24rpx;
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; }
.stg.on::after {
  content: '';
  position: absolute;
  left: 8rpx;
  right: 8rpx;
  bottom: 16rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #074D92;
  animation: toc-in .22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.stg-arr {
  font-size: 24rpx;
  color: #667085;
  transition: transform .2s ease, color .2s ease;
  padding: 20rpx 16rpx;
  margin: -20rpx -16rpx;
}
.stg-arr.up { transform: rotate(180deg); color: #074D92; }

/* 分类面板：absolute 浮层（同成果库），展开时不挤动下方内容 */
.field-panel {
  position: absolute;
  left: 0;
  right: 0;
  top: 100%;
  z-index: 43;
  background: #fff;
  border-radius: 0 0 12px 12px;
  box-shadow: 0 12px 24px rgba(16, 24, 40, 0.08);
  padding: 12px 14px 14px;
  max-height: 62vh;
  overflow-y: auto;
  animation: panelIn .3s cubic-bezier(.32, .72, 0, 1);
}
.field-panel.closing { animation: panelOut .21s ease-in forwards; }
@keyframes panelOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-10px); }
}
@keyframes panelIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
.field-panel .p-group { font-size: 13px; font-weight: 700; color: #344054; margin: 12px 0 6px; }
.field-panel .p-group:first-child { margin-top: 0; }
.p-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.p-chip {
  min-height: 40px;
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
}
.p-chip.act { color: #fff; border-color: #074D92; background: #074D92; font-weight: 600; }
.p-chip { transition: background .2s ease, border-color .2s ease, color .2s ease, transform .3s cubic-bezier(.34, 1.8, .64, 1); }
.p-chip:active { transform: scale(.94); transition: transform .08s linear; }
.p-chip.act { animation: chipPop .3s cubic-bezier(.34, 1.8, .64, 1); }
@keyframes chipPop { 0% { transform: scale(1); } 40% { transform: scale(.94); } 100% { transform: scale(1); } }

/* 蒙层：从 tab 分段底部开始置灰（top 由 maskTop 实测）；低于吸顶容器(40)，面板在容器内不被遮（同成果库） */
.panel-mask {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 35;
  background: rgba(16, 24, 40, 0.2);
  animation: maskIn .22s ease-out;
}
@keyframes maskIn { from { opacity: 0; } to { opacity: 1; } }

/* ===== Banner 渐变卡 ===== */
.banner {
  margin: 12px 14px;
  padding: 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #0A66C2 0%, #074D92 100%);
  color: #fff;
  position: relative;
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}
.banner::after {
  content: '';
  position: absolute;
  top: -30%;
  right: -20%;
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.08) 0%, transparent 70%);
}
.banner-top {
  display: flex;
  align-items: center;
  gap: 12px;
}
.banner-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}
.banner-info { flex: 1; min-width: 0; position: relative; z-index: 1; }
.banner-title { font-size: 14px; font-weight: 600; margin-bottom: 4px; display: block; line-height: 1.3; color: #fff; }
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: block; }
/* 实时统计行 */
.banner-stats {
  display: flex;
  align-items: center;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.16);
  position: relative;
  z-index: 1;
}
.bs {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.bs-num { font-size: 20px; font-weight: 800; line-height: 1.15; color: #fff; }
.bs-lb { font-size: 11px; color: rgba(255, 255, 255, 0.8); }
.bs-div { width: 1px; height: 26px; background: rgba(255, 255, 255, 0.18); }

/* ===== 白色板块（信息行 + 列表） ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { color: #667085; font-weight: 500; padding: 8px 4px 8px 12px; }

/* ===== 列表卡片（左缘分类色条 + 徽章 + 标题 + 描述 + 元信息） ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  overflow: hidden;
}

/* 封面 */
.case-cover {
  position: relative;
  width: 100%;
  height: 170px;
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
  border: 1px solid rgba(255, 255, 255, 0.28);
}
.fallback-ring-a { width: 160px; height: 160px; top: -60px; right: -40px; }
.fallback-ring-b { width: 220px; height: 220px; bottom: -120px; left: 10%; }
.fallback-text {
  color: rgba(255, 255, 255, 0.92);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 4px;
}

/* 类型角标：视频/图片 */
.type-tag {
  position: absolute;
  top: 10px;
  right: 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(16, 24, 40, 0.12);
}
.type-tag.video { color: #0A66C2; }
.type-tag.image { color: #07c160; }
/* CSS 播放三角（非 emoji；clip-path 绘制，避免厚色单边描边反模式） */
.tag-play {
  width: 7px;
  height: 9px;
  background: #0A66C2;
  clip-path: polygon(0 0, 100% 50%, 0 100%);
}

/* 信息区 */
.case-info {
  display: flex;
  flex-direction: column;
  gap: 7px;
  padding: 12px 14px 14px 18px;
}
.c-badges { display: flex; gap: 6px; }
.c-tag {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.ct {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-desc {
  font-size: 12.5px;
  color: #667085;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #667085;
}
.c-client { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.c-dot { color: #DDE1E6; }

/* ===== 骨架 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-cover { height: 150px; border-radius: 8px; background: #EDF0F3; }
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w70 { width: 70%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 加载更多 ===== */
.lm { text-align: center; padding: 12px; font-size: 12px; color: #667085; }

/* ===== 回到顶部 ===== */
.bt {
  position: fixed;
  bottom: 90px;
  right: 16px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 35;
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(0.16, 1, 0.3, 1);
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; }

/* ===================== 动效规范（对齐全局动画规范） =====================
   白名单：仅 transform / opacity（小尺寸 color/background 过渡允许）
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1) 松手柔顺减速 + ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* Banner 内部微编排：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms → 统计 180ms，总 ≤400ms */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner-stats { animation: fadeUp .2s ease-out 180ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; }
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }
/* Banner 单次扫光（非循环装饰，100ms 起播 280ms 线性，380ms 内收完） */
.banner::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  background: linear-gradient(100deg, transparent 0%, rgba(255, 255, 255, 0.22) 50%, transparent 100%);
  transform: translateX(-150%) skewX(-20deg);
  animation: shineOnce .28s linear 100ms backwards;
  pointer-events: none;
}
@keyframes shineOnce {
  from { transform: translateX(-150%) skewX(-20deg); }
  to { transform: translateX(320%) skewX(-20deg); }
}

/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 列表入场：前 6 项每 20ms 依次淡入上移（首屏可见范围） */
.card { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；一页仅此 1 处循环） */
.sk-tag, .sk-l, .sk-cover { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 交互反馈：可点元素按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位） */
.card { transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.tap-scale { transform: scale(.97); opacity: .92; transition-duration: .1s; transition-timing-function: linear; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }

/* ===== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用 ===== */
.page.no-motion .card,
.page.no-motion .banner,
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner-stats,
.page.no-motion .banner::before,
.page.no-motion .banner::after,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l, .page.no-motion .sk-cover { animation: none; }
.page.no-motion .tap-scale { transform: none !important; }
.page.no-motion .stb:active,
.page.no-motion .bt:active { transform: none; }
/* 筛选分段/面板（对齐成果库 no-motion 适配） */
.page.no-motion .stg-arr { transition: none; }
.page.no-motion .p-chip { transition: none; }
.page.no-motion .p-chip.act { animation: none; }
.page.no-motion .stg.on::after { animation: none; }
.page.no-motion .field-panel { animation: panelIn .3s ease-out; }
.page.no-motion .field-panel.closing { animation: panelOut .16s ease-in forwards; }
.page.no-motion .panel-mask { animation: maskIn .22s ease-out; }
.page.no-motion .p-chip:active,
.page.no-motion .stg:active { transform: none; }

/* ═══════ 案例详情弹窗 ═══════ */
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

/* 小程序 scroll-view 必须显式高度才能滚动：面板 86vh - 头部约 54px */
.detail-scroll {
  flex: 1;
  min-height: 0;
  height: calc(86vh - 54px);
  box-sizing: border-box;
  overflow-y: auto;
  padding: 24rpx 32rpx 48rpx;
}

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
  word-break: break-word;
  overflow-wrap: break-word;
}

/* ===== prefers-reduced-motion（系统级减弱动效）：筛选分段/面板动画与过渡全关 ===== */
@media (prefers-reduced-motion: reduce) {
  .stg, .stg-arr, .p-chip, .field-panel, .panel-mask {
    animation: none !important;
    transition: none !important;
  }
}
</style>
