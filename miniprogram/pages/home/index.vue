<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 1. 图片化顶部（masthead）：横幅图片贯穿安全区/导航/文案，头部与工具行叠放其上 -->
      <view class="masthead">
        <view class="hero-carousel">
          <swiper class="hero-swiper" autoplay circular :interval="5000" @change="onBannerChange">
            <swiper-item v-for="(b, i) in banners" :key="i">
              <view class="hero-slide" @tap="onBannerTap(b)">
                <image :src="b.image" mode="aspectFill" class="hero-img" @error="onBannerError(i)" />
                <view class="hero-mask"></view>
                <view class="hero-copy">
                  <text class="hero-kicker">{{ b.kicker }}</text>
                  <text class="hero-title">{{ b.title }}</text>
                  <text class="hero-desc">{{ b.desc }}</text>
                </view>
              </view>
            </swiper-item>
          </swiper>
          <view class="hero-dots">
            <view v-for="(b, i) in banners" :key="i" class="hero-dot" :class="{ active: i === activeBanner }"></view>
          </view>
        </view>

        <view class="mini-header" :style="headerStyle">
          <view class="nav-row" :style="navRowStyle">
            <text class="brand">低空综合服务平台</text>
          </view>
          <view class="tool-row">
            <view class="city-button" hover-class="tap-fade" hover-stay-time="120" @tap="openCityPicker">
              <text class="city-label">{{ cityLabel }}</text>
              <view class="icon icon-chevron"></view>
            </view>
            <view class="search-button" hover-class="tap-fade" hover-stay-time="120" @tap="goSearch">
              <image class="head-icon" :src="HOME_ICONS.search" mode="aspectFit" />
              <text class="search-placeholder">搜索项目、飞手、服务</text>
            </view>
            <view class="message-button" hover-class="tap-fade" hover-stay-time="120" @tap="goMessages">
              <image class="head-icon head-icon-msg" :src="HOME_ICONS.message" mode="aspectFit" />
              <view v-if="unreadCount > 0" class="msg-dot"></view>
            </view>
          </view>
        </view>
      </view>

      <!-- 2. 找项目 / 发需求 双行动入口（白底紧凑卡） -->
      <view class="action-panel">
        <view class="action-card primary" hover-class="tap-scale" hover-stay-time="120" @tap="goFindProject">
          <view class="action-icon"><image class="action-icon-img" :src="HOME_ICONS.findProject" mode="aspectFit" /></view>
          <view class="action-copy">
            <text class="action-title">找项目</text>
            <text class="action-sub">进入需求大厅找机会</text>
          </view>
        </view>
        <view class="action-card accent" hover-class="tap-scale" hover-stay-time="120" @tap="goPublishDemand">
          <view class="action-icon"><image class="action-icon-img" :src="HOME_ICONS.publishDemand" mode="aspectFit" /></view>
          <view class="action-copy">
            <text class="action-title">发需求</text>
            <text class="action-sub">发布需求快速获报价</text>
          </view>
        </view>
      </view>

      <!-- 3. 核心服务：3 x 2 横向图文功能砖 + 底部单行公告 -->
      <view class="surface-section services-section">
        <view class="section-head">
          <view class="section-title">
            <text class="section-title-main">核心服务</text>
            <text class="section-title-sub">低空业务一站式服务入口</text>
          </view>
          <view class="more-button" hover-class="tap-fade" hover-stay-time="120" @tap="goAllServices">
            <text>全部</text>
            <view class="icon icon-arrow"></view>
          </view>
        </view>
        <view class="service-grid">
          <view
            v-for="s in services"
            :key="s.name"
            class="service-item"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="goService(s)"
          >
            <view class="service-icon">
              <image class="service-icon-img" :src="s.icon" mode="aspectFit" />
            </view>
            <text class="service-name">{{ s.name }}</text>
          </view>
        </view>
        <view v-if="notice" class="notice-strip">
          <text class="notice-label">平台公告</text>
          <text class="notice-copy">{{ notice }}</text>
          <view class="icon icon-arrow"></view>
        </view>
      </view>

      <!-- 4. 供需项目：唯一内容板块，首条为重点卡 -->
      <view class="surface-section demand-section">
        <view class="section-head">
          <view class="section-title">
            <text class="section-title-main">供需项目</text>
            <text class="section-title-sub">最新发布的低空作业需求</text>
          </view>
          <view class="more-button" hover-class="tap-fade" hover-stay-time="120" @tap="goDemandsList">
            <text>进入大厅</text>
            <view class="icon icon-arrow"></view>
          </view>
        </view>

        <scroll-view scroll-x class="category-scroll" :show-scrollbar="false">
          <view
            v-for="f in filters"
            :key="f.key"
            class="category-chip"
            :class="{ active: activeFilter === f.key }"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="setFilter(f.key)"
          >{{ f.label }}</view>
        </scroll-view>

        <!-- 加载中：第一个骨架为重点卡尺寸，第二个为紧凑卡 -->
        <view v-if="demandState === 'loading'" class="skeleton-list">
          <view class="skeleton-card featured-skeleton"></view>
          <view class="skeleton-card"></view>
        </view>

        <!-- 加载失败 -->
        <view v-else-if="demandState === 'error'" class="state-panel">
          <view class="state-icon"><view class="icon icon-refresh"></view></view>
          <text class="state-title">供需项目加载失败</text>
          <text class="state-desc">其他首页服务仍可使用，请检查网络后重试。</text>
          <view class="retry-button" hover-class="tap-fade" hover-stay-time="120" @tap="reloadDemands">重新加载</view>
        </view>

        <!-- 空数据 -->
        <view v-else-if="demandState === 'empty'" class="state-panel">
          <view class="state-icon"><view class="icon icon-inbox"></view></view>
          <text class="state-title">暂无符合条件的供需项目</text>
          <text class="state-desc">新的需求发布后会第一时间展示在这里。</text>
        </view>

        <!-- 正常列表：首条可见数据为重点卡，其余为紧凑横卡 -->
        <view v-else class="demand-list">
          <view
            v-for="(d, index) in filteredDemands"
            :key="d.id"
            class="demand-card"
            :class="{ featured: index === 0 }"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="goDemandDetail(d)"
          >
            <image :src="d.image" mode="aspectFill" class="demand-photo" @error="onDemandImageError(d)" />
            <view class="demand-body">
              <view class="card-badges">
                <text class="type-badge">{{ d.type }}</text>
                <text v-if="d.statusLabel" class="status-badge" :class="d.statusClass">{{ d.statusLabel }}</text>
              </view>
              <text class="demand-title">{{ d.title }}</text>
              <view v-if="d.district || d.publishedAt" class="demand-meta">
                <text v-if="d.district" class="meta-item">{{ d.district }}</text>
                <text v-if="d.publishedAt" class="meta-item">{{ d.publishedAt }}</text>
              </view>
              <view class="demand-foot">
                <view class="budget">
                  <text class="budget-label">预算</text>
                  <text class="budget-value">{{ d.budgetText }}</text>
                </view>
                <text v-if="d.bidCount != null" class="bid-count">已有 {{ d.bidCount }} 家报价</text>
              </view>
            </view>
          </view>
          <!-- 筛选无匹配 -->
          <view v-if="filteredDemands.length === 0" class="filter-empty">当前类型暂无项目，试试其他筛选条件</view>
        </view>
      </view>
    </view>
  </Layout>

  <!-- 城市选择底部弹层 -->
  <u-popup :show="showCityPicker" position="bottom" round @close="closeCityPicker">
    <view class="city-sheet">
      <view class="city-sheet-head">
        <text class="city-sheet-title">选择服务区域</text>
        <view class="city-sheet-close" hover-class="tap-fade" hover-stay-time="120" @tap="closeCityPicker">
          <text class="icon icon-x">×</text>
        </view>
      </view>
      <scroll-view scroll-y class="city-sheet-body">
        <view class="city-grid">
          <view
            v-for="opt in cityOptions"
            :key="opt.value"
            class="city-option"
            :class="{ active: city === opt.value }"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="pickCity(opt)"
          >{{ opt.label }}</view>
        </view>
      </scroll-view>
    </view>
  </u-popup>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'
import { request } from '../../utils/request'

/* ================= 路由安全 ================= */
const ROUTE_MAP = {
  '/pages/demand/list': '/pages/demands/list',
}
const TAB_PAGES = new Set([
  '/pages/home/index',
  '/pages/demands/index',
  '/pages/publish/index',
  '/pages/services/index',
  '/pages/mine/index',
])
const ALLOWED_ROUTES = new Set([
  '/pages/home/index', '/pages/demands/index', '/pages/publish/index',
  '/pages/services/index', '/pages/mine/index',
  '/pages/demands/list', '/pages/demands/detail', '/pkg-demand/pages/demands/publish',
  '/pkg-demand/pages/demands/mine', '/pages/intents/mine',
  '/pages/search/index', '/pages/messages/index',
  '/pkg-talent/pages/pilots/list', '/pkg-talent/pages/training/courses',
  '/pkg-talent/pages/experts/list', '/pkg-emergency/pages/emergency/resources', '/pkg-service/pages/compliance/news',
  '/pkg-eco/pages/mall/index', '/pkg-eco/pages/shops/index', '/pkg-service/pages/more/index',
])

const resolveBannerLink = (raw) => {
  if (!raw) return ''
  const s = String(raw).trim()
  if (!s || /^https?:\/\//i.test(s)) return ''
  const mapped = ROUTE_MAP[s] || s
  return ALLOWED_ROUTES.has(mapped) ? mapped : ''
}

/* ================= 头部几何 ================= */
const statusBarH = ref(20)
const navBarH = ref(32)
const headerPadTop = ref(24)
const capsuleGap = ref(87)

const headerStyle = computed(() => ({ paddingTop: headerPadTop.value + 'px' }))
const navRowStyle = computed(() => ({
  height: navBarH.value + 'px',
  paddingRight: capsuleGap.value + 'px',
}))

const initHeader = () => {
  try {
    const info = uni.getSystemInfoSync()
    statusBarH.value = info.statusBarHeight || 20
    let mr = null
    if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
      mr = uni.getMenuButtonBoundingClientRect()
    }
    if (mr) {
      navBarH.value = mr.height || 32
      capsuleGap.value = Math.max(info.windowWidth - mr.left + 8, 60)
      headerPadTop.value = mr.top || statusBarH.value + 4
    } else {
      navBarH.value = 32
      capsuleGap.value = 87
      headerPadTop.value = statusBarH.value + 4
    }
  } catch (e) {
    statusBarH.value = 20
    navBarH.value = 32
    capsuleGap.value = 87
    headerPadTop.value = 24
  }
}
initHeader()

/* ================= 城市 ================= */
const ALL_CITY = '全重庆'
const city = ref(ALL_CITY)
const showCityPicker = ref(false)
const cityLabel = computed(() => (city.value === ALL_CITY ? '重庆' : city.value))

const DISTRICT_LIST = [
  '渝中区', '江北区', '南岸区', '沙坪坝区', '九龙坡区', '大渡口区', '北碚区', '渝北区',
  '巴南区', '两江新区', '高新区', '涪陵区', '长寿区', '江津区', '合川区', '永川区',
  '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区',
  '梁平区', '武隆区', '万州区', '黔江区', '城口县', '丰都县', '垫江县', '忠县',
  '云阳县', '奉节县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县',
]
const cityOptions = [
  { value: ALL_CITY, label: '重庆' },
  ...DISTRICT_LIST.map((d) => ({ value: d, label: d })),
]

const openCityPicker = () => { showCityPicker.value = true }
const closeCityPicker = () => { showCityPicker.value = false }
const pickCity = (opt) => {
  if (city.value !== opt.value) {
    city.value = opt.value
    loadAll()
  }
  showCityPicker.value = false
}

/* ================= 横幅 ================= */
const DEFAULT_BANNERS = [
  {
    image: '/static/home/hero-inspection.jpg',
    kicker: '低空作业供需直连',
    title: '专业资源，快速匹配',
    desc: '覆盖巡检、吊运、航拍、植保与测绘服务',
    link: '',
  },
  {
    image: '/static/home/home-bg.jpg',
    kicker: '平台服务保障',
    title: '合规飞行，安心作业',
    desc: '连接项目、飞手、服务商与培训机构',
    link: '',
  },
]
const banners = ref(DEFAULT_BANNERS.map((b) => ({ ...b })))
const activeBanner = ref(0)
const notice = ref('')

const buildBanners = (apiList) => {
  const valid = (Array.isArray(apiList) ? apiList : []).filter((b) => {
    if (!b || typeof b.image_url !== 'string' || !b.image_url.trim()) return false
    const st = b.status
    if (st && st !== 'active' && st !== 'enabled' && st !== '') return false
    return true
  })
  if (!valid.length) return DEFAULT_BANNERS.map((b) => ({ ...b }))
  return DEFAULT_BANNERS.map((def, i) => {
    const api = valid[i]
    if (!api) return { ...def }
    return { ...def, image: api.image_url, link: api.link_url }
  })
}
const onBannerChange = (e) => { activeBanner.value = e.detail.current }
const onBannerError = (i) => {
  const def = DEFAULT_BANNERS[i]
  if (def && banners.value[i] && banners.value[i].image !== def.image) {
    banners.value[i] = { ...banners.value[i], image: def.image }
  }
}
const onBannerTap = (b) => {
  const url = resolveBannerLink(b.link)
  if (!url) return
  if (TAB_PAGES.has(url)) safeSwitchTab(url)
  else safeNavigateTo(url)
}

/* ================= 首页入口图标统一配置 =================
 * 全部指向 Codex 交付的首页图标资源 miniprogram/static/home/icons/，
 * 不在此处内联绘制图标。替换文件即可调整视觉，无需改动结构。
 */
const HOME_ICONS = {
  // 顶部工具行
  search: '/static/home/icons/search-white.svg',
  message: '/static/home/icons/message.svg',
  // 双行动入口
  findProject: '/static/home/icons/briefcase.svg',
  publishDemand: '/static/home/icons/edit.svg',
  // 六项核心服务
  services: {
    demand: '/static/home/icons/demand.svg', // 需求大厅
    trade: '/static/home/icons/trade.svg', // 场地预约
    pilot: '/static/home/icons/pilot.svg', // 专家智库
    shop: '/static/home/icons/ecoservice.svg', // 生态服务
    training: '/static/home/icons/training.svg', // 培训认证
    policy: '/static/home/icons/policy.svg', // 政策资讯
  },
}
const services = ref([
  { name: '需求大厅', icon: HOME_ICONS.services.demand, path: '/pages/demands/index', tab: true },
  { name: '培训认证', icon: HOME_ICONS.services.training, path: '/pkg-talent/pages/training/courses', tab: false },
  { name: '专家智库', icon: HOME_ICONS.services.pilot, path: '/pkg-talent/pages/experts/list', tab: false },
  { name: '场地预约', icon: HOME_ICONS.services.trade, path: '/pkg-service/pages/testsites/list', tab: false },
  { name: '政策资讯', icon: HOME_ICONS.services.policy, path: '/pkg-service/pages/compliance/news', tab: false },
  { name: '生态服务', icon: HOME_ICONS.services.shop, path: '/pages/services/index', tab: true },
])
const goService = (s) => {
  if (s.tab) safeSwitchTab(s.path)
  else safeNavigateTo(s.path)
}
const goAllServices = () => safeSwitchTab('/pages/services/index')

/* ================= 供需项目 ================= */
const demandState = ref('loading')
const demands = ref([])
const activeFilter = ref('all')
const filters = [
  { key: 'all', label: '全部' },
  { key: '吊运', label: '吊运' },
  { key: '航拍', label: '航拍' },
  { key: '植保', label: '植保' },
  { key: '巡检', label: '巡检' },
  { key: '测绘', label: '测绘' },
]
const filteredDemands = computed(() => {
  if (activeFilter.value === 'all') return demands.value
  return demands.value.filter((d) => d.filterKey === activeFilter.value)
})
const setFilter = (key) => { activeFilter.value = key }

const LOCAL_DEMAND_IMG = '/static/home/demand-solar.jpg'
const LOCAL_LIFT_IMG = '/static/home/demand-lift.jpg'
const TYPE_LABELS = {
  cable_inspection: '工业巡检',
  plant_transport: '植保运输',
  spray_pesticide: '农药喷洒',
  clean_paint: '清洗保洁',
  trade_lease: '租赁服务',
  other: '其他服务',
}
const CATEGORY_BADGE = {
  巡检: '工业巡检',
  吊运: '物资吊运',
  航拍: '航拍服务',
  植保: '植保作业',
  测绘: '测绘服务',
}
const STATUS_MAP = {
  published: { label: '招募中', cls: 'ok' },
  matched: { label: '已匹配', cls: 'neutral' },
  completed: { label: '已完成', cls: 'neutral' },
  pending: { label: '待审核', cls: 'neutral' },
}

const classifyDemand = (d) => {
  const biz = String(d.biz_type || '').toLowerCase()
  const text = `${d.biz_type || ''} ${d.title || ''} ${d.description || ''}`
  if (biz === 'cable_inspection' || /巡检/.test(text)) return '巡检'
  if (biz === 'plant_transport' || biz === 'spray_pesticide' || /植保|农药|喷洒/.test(text)) return '植保'
  if (/吊运|运输|物资/.test(text)) return '吊运'
  if (/航拍|摄影|影像/.test(text)) return '航拍'
  if (/测绘|建模|测量/.test(text)) return '测绘'
  return ''
}

const demandTypeLabel = (d) => {
  const biz = String(d.biz_type || '')
  if (biz && TYPE_LABELS[biz]) return TYPE_LABELS[biz]
  const cat = classifyDemand(d)
  return CATEGORY_BADGE[cat] || '低空需求'
}

const fmtRelative = (iso) => {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  if (Number.isNaN(t)) return ''
  const diff = Date.now() - t
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const d = new Date(t)
  return `${d.getMonth() + 1}-${d.getDate()}`
}

const fmtBudget = (fen) => {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const whole = Math.floor(yuan)
  const cents = Math.round((yuan - whole) * 100)
  const w = whole.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return cents > 0 ? `¥${w}.${cents < 10 ? '0' : ''}${cents}` : `¥${w}`
}

const normalizeDemand = (d) => {
  if (!d) return null
  const title = String(d.title || '').trim()
  if (!title) return null
  const cat = classifyDemand(d)
  const imgs = Array.isArray(d.images) ? d.images.filter((u) => typeof u === 'string' && u.trim()) : []
  const image = imgs[0] || (cat === '吊运' ? LOCAL_LIFT_IMG : LOCAL_DEMAND_IMG)
  const budgetText = fmtBudget(d.budget_fen)
  const st = STATUS_MAP[d.status] || null
  let bidCount = null
  if (typeof d.bid_count === 'number' && d.bid_count >= 0) bidCount = d.bid_count
  else if (typeof d.bidCount === 'number' && d.bidCount >= 0) bidCount = d.bidCount
  return {
    id: String(d.id || ''),
    title,
    type: demandTypeLabel(d),
    filterKey: cat,
    statusLabel: st ? st.label : '',
    statusClass: st ? st.cls : '',
    image,
    district: String(d.district || ''),
    publishedAt: fmtRelative(d.created_at),
    budgetText,
    bidCount,
  }
}

const normalizeDemandList = (res) => {
  let list = []
  if (Array.isArray(res)) list = res
  else if (res && Array.isArray(res.data)) list = res.data
  else if (res && Array.isArray(res.items)) list = res.items
  const out = []
  for (const d of list) {
    const item = normalizeDemand(d)
    if (item) out.push(item)
  }
  return out
}

const onDemandImageError = (d) => {
  if (d.image !== LOCAL_LIFT_IMG && d.image !== LOCAL_DEMAND_IMG) {
    d.image = d.filterKey === '吊运' ? LOCAL_LIFT_IMG : LOCAL_DEMAND_IMG
  }
}

/* ================= 数据加载 ================= */
let loadSeq = 0

const loadAll = async (opts = {}) => {
  const seq = ++loadSeq
  const isRefresh = !!opts.refresh
  if (!isRefresh || (demandState.value !== 'ready' && demandState.value !== 'empty')) {
    demandState.value = 'loading'
  }

  const isAllCity = city.value === ALL_CITY
  const cityQuery = isAllCity ? '' : '?city=' + encodeURIComponent(city.value)
  const demandsParams = { page: 1, page_size: 20, sort: 'newest' }
  if (!isAllCity) demandsParams.district = city.value

  try {
    const [homeRes, demandsRes] = await Promise.allSettled([
      request({ url: '/api/v1/home' + cityQuery }),
      request({ url: '/api/v1/demands', data: demandsParams }),
    ])
    if (seq !== loadSeq) return

    // home 成功：应用真实横幅与公告；失败：保留本地横幅与无公告状态
    if (homeRes.status === 'fulfilled') {
      const data = homeRes.value || {}
      if (Array.isArray(data.banners)) banners.value = buildBanners(data.banners)
      if (Array.isArray(data.notices) && data.notices.length) {
        notice.value = String(data.notices[0] || '').trim()
      } else {
        notice.value = ''
      }
    }

    // demands 独立处理，与 home 互不连坐
    if (demandsRes.status === 'fulfilled') {
      const items = normalizeDemandList(demandsRes.value)
      if (!items.length) {
        demands.value = []
        demandState.value = 'empty'
      } else {
        demands.value = items
        demandState.value = 'ready'
      }
    } else {
      demands.value = []
      demandState.value = 'error'
    }
  } catch (e) {
    if (seq !== loadSeq) return
    demands.value = []
    demandState.value = 'error'
  }
}

const reloadDemands = () => { loadAll() }

/* ================= 消息红点（仅真实未读数驱动） ================= */
const unreadCount = ref(0)
const loadUnreadCount = async () => {
  try {
    if (!uni.getStorageSync('accessToken')) {
      unreadCount.value = 0
      return
    }
    const res = await request({ url: '/api/v1/messages/unread-count' })
    const n = Number(res && res.count != null ? res.count : 0)
    unreadCount.value = n > 0 ? n : 0
  } catch (e) {
    unreadCount.value = 0
  }
}

/* ================= 导航 ================= */
const goSearch = () => safeNavigateTo('/pages/search/index')
const goMessages = () => safeNavigateTo('/pages/messages/index')
const goFindProject = () => safeNavigateTo('/pages/demands/list')
const goPublishDemand = () => safeNavigateTo('/pkg-demand/pages/demands/publish')
const goDemandsList = () => safeNavigateTo('/pages/demands/list')
const goDemandDetail = (d) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(d.id))

/* ================= 生命周期 ================= */
onLoad(() => {
  initHeader()
  loadAll()
  loadUnreadCount()
})

onPullDownRefresh(() => {
  loadAll({ refresh: true }).finally(() => {
    uni.stopPullDownRefresh()
  })
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #F4F6F8;
}

/* ================= 通用图标（CSS 线性绘制，仅用于小装饰） ================= */
.icon { position: relative; display: block; }

.icon-chevron {
  width: 9px;
  height: 9px;
  border-right: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
  transform: rotate(45deg);
  margin-top: -3px;
}

.icon-arrow {
  width: 7px;
  height: 7px;
  border-top: 1.5px solid currentColor;
  border-right: 1.5px solid currentColor;
  transform: rotate(45deg);
}

.icon-x {
  position: relative;
  width: 14px;
  height: 14px;
}
.icon-x::before,
.icon-x::after {
  content: '';
  position: absolute;
  left: 0;
  top: 6px;
  width: 14px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}
.icon-x::before { transform: rotate(45deg); }
.icon-x::after { transform: rotate(-45deg); }

.icon-refresh {
  position: relative;
  width: 17px;
  height: 17px;
  border: 2px solid currentColor;
  border-left-color: transparent;
  border-radius: 50%;
}
.icon-refresh::after {
  content: '';
  position: absolute;
  top: -4px;
  right: -1px;
  width: 6px;
  height: 6px;
  border-top: 2px solid currentColor;
  border-right: 2px solid currentColor;
  transform: rotate(45deg);
}

.icon-inbox {
  position: relative;
  width: 17px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 3px;
}
.icon-inbox::before {
  content: '';
  position: absolute;
  top: 3px;
  left: 1px;
  right: 1px;
  height: 2px;
  background: currentColor;
}

/* ================= 图片化顶部（masthead） ================= */
.masthead {
  position: relative;
  height: 236px;
  overflow: hidden;
  color: #fff;
  background: #17314A;
}
.hero-carousel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 1;
  overflow: hidden;
  background: #1E3448;
}
.hero-swiper {
  width: 100%;
  height: 100%;
}
.hero-slide {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}
.hero-img {
  width: 100%;
  height: 100%;
  display: block;
}
.hero-mask {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  background: rgba(7, 34, 59, 0.56);
}
.hero-copy {
  position: absolute;
  z-index: 2;
  left: 15px;
  top: 142px;
  width: 270px;
  color: #fff;
  display: flex;
  flex-direction: column;
}
.hero-kicker {
  display: block;
  margin-bottom: 6px;
  font-size: 10px;
  font-weight: 700;
  color: #D3E9FB;
}
.hero-title {
  display: block;
  font-size: 20px;
  line-height: 1.25;
  font-weight: 700;
}
.hero-desc {
  display: block;
  margin-top: 6px;
  color: rgba(255, 255, 255, 0.82);
  font-size: 10px;
  line-height: 1.5;
}
.hero-dots {
  position: absolute;
  z-index: 4;
  right: 12px;
  bottom: 12px;
  display: flex;
  gap: 5px;
}
.hero-dot {
  width: 6px;
  height: 6px;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.44);
  transition: width 0.2s;
}
.hero-dot.active {
  width: 18px;
  background: #fff;
}

/* 叠放在图片上的头部 */
.mini-header {
  position: absolute;
  z-index: 30;
  top: 0;
  left: 0;
  right: 0;
  color: #fff;
  background: transparent;
  padding-left: 12px;
  padding-right: 12px;
  padding-bottom: 8px;
}
.nav-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.brand {
  font-size: 17px;
  font-weight: 700;
  line-height: 1;
  color: #fff;
}
.tool-row {
  margin-top: 6px;
  display: grid;
  grid-template-columns: 64px 1fr 38px;
  gap: 7px;
  align-items: center;
}
.city-button,
.search-button,
.message-button {
  height: 40px;
  border-radius: 7px;
}
.city-button {
  padding: 0 5px;
  color: #fff;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1px;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
}
.city-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 50px;
}
.search-button {
  padding: 0 11px;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.28);
  color: #fff;
  display: flex;
  align-items: center;
  gap: 7px;
  text-align: left;
}
.search-placeholder {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.9);
}
.message-button {
  position: relative;
  color: #fff;
  background: rgba(7, 35, 61, 0.48);
  border: 1px solid rgba(255, 255, 255, 0.22);
  display: flex;
  align-items: center;
  justify-content: center;
}
.head-icon {
  width: 15px;
  height: 15px;
  flex: 0 0 auto;
  display: block;
}
.head-icon-msg {
  width: 17px;
  height: 17px;
}
.msg-dot {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #F97316;
  border: 1px solid #074D92;
}

/* ================= 双行动入口（白底紧凑卡） ================= */
.action-panel {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 8px 12px 0;
  background: #F4F6F8;
}
.action-card {
  min-height: 60px;
  border-radius: 8px;
  border: 1px solid #E4E7EC;
  padding: 7px 9px;
  display: grid;
  grid-template-columns: 34px 1fr;
  align-items: center;
  gap: 8px;
  text-align: left;
  background: #fff;
  box-shadow: 0 2px 8px rgba(16, 24, 40, 0.035);
  box-sizing: border-box;
}
.action-card.primary {
  color: #074D92;
  border-color: #CFE2F4;
}
.action-card.accent {
  color: #E96012;
  border-color: #FFD7BD;
}
.action-icon {
  width: 34px;
  height: 34px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0A66C2;
  background: #EAF3FB;
}
.action-card.accent .action-icon {
  color: #E96012;
  background: #FFF0E6;
}
.action-icon-img {
  width: 21px;
  height: 21px;
  display: block;
}
.action-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.action-title {
  display: block;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.2;
}
.action-sub {
  display: block;
  margin-top: 3px;
  color: #667085;
  font-size: 8px;
  line-height: 1.3;
}

/* ================= 通用白色板块 ================= */
.surface-section {
  margin-top: 8px;
  padding: 12px;
  background: #fff;
}
.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 6px;
}
.section-title {
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.section-title-main {
  font-size: 17px;
  line-height: 1.25;
  font-weight: 700;
  color: #17212B;
}
.section-title-sub {
  margin-top: 3px;
  color: #667085;
  font-size: 10px;
  line-height: 1.45;
}
.more-button {
  min-height: 28px;
  display: flex;
  align-items: center;
  gap: 1px;
  color: #667085;
  font-size: 11px;
  white-space: nowrap;
}

/* ================= 核心服务：横向图文功能砖 ================= */
.services-section {
  padding-top: 8px;
  padding-bottom: 8px;
}
.services-section .section-title-sub {
  display: none;
}
.service-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
}
.service-item {
  min-width: 0;
  min-height: 42px;
  border-radius: 7px;
  padding: 0 7px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 6px;
  font-size: 10px;
  color: #344054;
  background: #F4F8FC;
  white-space: nowrap;
}
.service-item:nth-child(2),
.service-item:nth-child(5) {
  background: #FFF7F1;
}
.service-item:nth-child(3) {
  background: #F0F8F4;
}
.service-item:nth-child(4) {
  background: #F6F4FF;
}
.service-icon {
  flex: 0 0 auto;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.78);
}
.service-icon-img {
  width: 17px;
  height: 17px;
  display: block;
}
.service-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ================= 公告（并入服务板块底部） ================= */
.notice-strip {
  width: 100%;
  min-height: 32px;
  margin: 6px 0 0;
  padding: 0 9px;
  border: 1px solid #DCEAF7;
  border-radius: 7px;
  background: #F4F8FC;
  color: #344054;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.notice-label {
  padding-right: 8px;
  border-right: 1px solid #CBDDEC;
  color: #074D92;
  font-size: 11px;
  font-weight: 700;
}
.notice-copy {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 10px;
}
.notice-strip .icon-arrow {
  color: #98A2B3;
}

/* ================= 供需项目 ================= */
.category-scroll {
  margin: -1px -12px 8px;
  padding: 0 12px;
  box-sizing: content-box;
  white-space: nowrap;
}
.category-scroll::-webkit-scrollbar {
  display: none;
}
.category-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  min-height: 30px;
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 11px;
  box-sizing: border-box;
}
.category-chip + .category-chip {
  margin-left: 7px;
}
.category-chip.active {
  color: #fff;
  border-color: #0A66C2;
  background: #0A66C2;
  font-weight: 700;
}

.demand-list {
  display: grid;
  gap: 9px;
}
.demand-card {
  min-height: 126px;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
}
.demand-photo {
  width: 112px;
  min-height: 126px;
  display: block;
}
.demand-body {
  min-width: 0;
  padding: 10px;
  display: flex;
  flex-direction: column;
}
.card-badges {
  display: flex;
  align-items: center;
  gap: 5px;
}
.type-badge,
.status-badge {
  display: inline-flex;
  align-items: center;
  min-height: 20px;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: 700;
}
.type-badge {
  color: #074D92;
  background: #EAF3FB;
}
.status-badge.ok {
  color: #168A55;
  background: #E9F7F0;
}
.status-badge.neutral {
  color: #667085;
  background: #EEF1F4;
}
.demand-title {
  margin-top: 7px;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  font-size: 13px;
  line-height: 1.4;
  font-weight: 700;
  color: #17212B;
}
.demand-meta {
  margin-top: 7px;
  color: #667085;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 9px;
}
.demand-foot {
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid #EEF1F4;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 7px;
}
.budget {
  display: flex;
  align-items: baseline;
  gap: 3px;
  color: #E96012;
}
.budget-label {
  font-size: 9px;
  font-weight: 500;
}
.budget-value {
  font-size: 15px;
  font-weight: 700;
  white-space: nowrap;
}
.bid-count {
  color: #98A2B3;
  font-size: 9px;
  white-space: nowrap;
}

/* 首条重点卡：全宽大图 + 紧凑正文 */
.demand-card.featured {
  min-height: 0;
  display: block;
}
.demand-card.featured .demand-photo {
  width: 100%;
  height: 138px;
  min-height: 138px;
}
.demand-card.featured .demand-body {
  min-height: 112px;
  padding: 9px 11px 10px;
}
.demand-card.featured .demand-title {
  margin-top: 6px;
  font-size: 14px;
  -webkit-line-clamp: 1;
}
.demand-card.featured .demand-meta {
  margin-top: 5px;
}
.demand-card.featured .demand-foot {
  padding-top: 7px;
}

.filter-empty {
  display: block;
  padding: 28px 12px 24px;
  text-align: center;
  color: #667085;
  font-size: 11px;
}

/* ================= 骨架屏 ================= */
.skeleton-list {
  display: grid;
  gap: 9px;
}
.skeleton-card {
  height: 126px;
  border-radius: 8px;
  background: #EDF0F3;
  position: relative;
  overflow: hidden;
}
.skeleton-card.featured-skeleton {
  height: 250px;
}
.skeleton-card::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.72), transparent);
  transform: translateX(-100%);
  animation: shimmer 1.15s infinite;
}
@keyframes shimmer {
  to { transform: translateX(100%); }
}

/* ================= 状态面板 ================= */
.state-panel {
  min-height: 172px;
  padding: 28px 22px;
  border: 1px dashed #D6DCE3;
  border-radius: 8px;
  background: #FAFBFC;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.state-icon {
  width: 38px;
  height: 38px;
  margin-bottom: 10px;
  border-radius: 8px;
  color: #0A66C2;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
}
.state-title {
  font-size: 13px;
  font-weight: 700;
  color: #17212B;
}
.state-desc {
  margin: 6px auto 11px;
  max-width: 248px;
  color: #667085;
  font-size: 10px;
  line-height: 1.55;
}
.retry-button {
  min-height: 32px;
  padding: 0 14px;
  border-radius: 6px;
  color: #fff;
  background: #0A66C2;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
}

/* ================= 城市选择弹层 ================= */
.city-sheet {
  height: 75vh;
  display: flex;
  flex-direction: column;
  padding: 8px 16px calc(18px + env(safe-area-inset-bottom));
  box-sizing: border-box;
}
.city-sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 0 12px;
}
.city-sheet-title {
  font-size: 17px;
  font-weight: 700;
  color: #17212B;
}
.city-sheet-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667085;
}
.city-sheet-body {
  flex: 1;
  min-height: 0;
}
.city-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 7px;
}
.city-option {
  min-height: 40px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  color: #344054;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  box-sizing: border-box;
}
.city-option.active {
  color: #074D92;
  border-color: #0A66C2;
  background: #F4F8FC;
  font-weight: 700;
}

/* ================= 按压反馈 ================= */
.tap-fade {
  opacity: 0.85;
}
.tap-scale {
  transform: scale(0.985);
  opacity: 0.9;
}

/* ================= 响应式：375px 宽度 ================= */
@media (max-width: 380px) {
  .demand-card:not(.featured) {
    grid-template-columns: 104px minmax(0, 1fr);
  }
  .demand-card:not(.featured) .demand-photo {
    width: 104px;
  }
}
</style>
