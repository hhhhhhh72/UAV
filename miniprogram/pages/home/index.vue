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
        <view v-if="notice" class="notice-strip" hover-class="tap-fade" hover-stay-time="120" @tap="goNotices">
          <text class="notice-label">平台公告</text>
          <text class="notice-copy">{{ notice }}</text>
          <view class="icon icon-arrow"></view>
        </view>
      </view>

      <!-- 4. 企业展示：入驻平台的低空企业（横向滑动卡片，点击/全部 → 企业列表页） -->
      <view v-if="enterprises.length > 0" class="surface-section enterprise-section">
        <view class="section-head">
          <view class="section-title">
            <text class="section-title-main">企业展示</text>
            <text class="section-title-sub">入驻平台的低空服务企业</text>
          </view>
          <view class="more-button" hover-class="tap-fade" hover-stay-time="120" @tap="goEnterpriseList">
            <text>全部</text>
            <view class="icon icon-arrow"></view>
          </view>
        </view>
        <scroll-view scroll-x class="enterprise-scroll" :show-scrollbar="false">
          <view class="enterprise-row">
            <view
              v-for="e in enterprises"
              :key="e.id"
              class="enterprise-card"
              hover-class="tap-fade"
              hover-stay-time="120"
              @tap="goEnterpriseList"
            >
              <view class="ent-logo">
                <image
                  v-if="e.logo"
                  :src="resolveEntImg(e.logo)"
                  mode="aspectFill"
                  class="ent-logo-img"
                  @error="e.logo = ''"
                />
                <view v-else class="ent-logo-fallback">{{ e.name ? e.name.charAt(0) : '企' }}</view>
              </view>
              <view class="ent-info">
                <text class="ent-card-name">{{ e.name }}</text>
                <view class="ent-card-tags">
                  <text v-if="firstCategory(e)" class="ent-card-tag">{{ firstCategory(e) }}</text>
                  <text v-if="e.is_member" class="ent-card-tag ent-card-tag--member">会员</text>
                </view>
              </view>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 5. 供需项目：唯一内容板块，首条为重点卡 -->
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

        <!-- 正常列表：锚点无限循环窗口（首条重点卡动态展开/收缩，其余为左图右文紧凑卡） -->
        <DemandLoop
          v-else-if="filteredDemands.length"
          :items="filteredDemands"
          @select="goDemandDetail"
          @preview="previewFeaturedImg"
          @image-error="onDemandImageError"
        />
        <!-- 筛选无匹配 -->
        <view v-else class="filter-empty">当前类型暂无项目，试试其他筛选条件</view>
      </view>
    </view>
  </Layout>

  <!-- ═══════ 就地搜索覆盖层（点击搜索栏在当前页展开，不跳搜索页） ═══════ -->
  <view v-if="showSearch" class="ov-overlay">
    <view class="ov-topbar" :style="{ paddingTop: (statusBarH + 4) + 'px' }">
      <view class="ov-topbar-row">
        <view class="ov-back" hover-class="tap-fade" hover-stay-time="120" @tap="closeSearch">
          <view class="ov-back-arrow"></view>
        </view>
        <text class="ov-title">全局搜索</text>
        <view class="ov-spacer"></view>
      </view>
      <view class="ov-search-box">
        <view class="ov-search-icon"></view>
        <input
          class="ov-search-input"
          v-model="searchText"
          placeholder="搜索需求、企业..."
          placeholder-class="ov-search-ph"
          confirm-type="search"
          focus
          @input="onSearchInput($event.detail.value)"
          @confirm="onSearch"
        />
      </view>
    </view>

    <view class="ov-tabs">
      <view
        class="ov-tab"
        :class="{ active: activeTab === 'demand' }"
        @tap="onTabChange(0)"
      >搜需求</view>
      <view
        class="ov-tab"
        :class="{ active: activeTab === 'enterprise' }"
        @tap="onTabChange(1)"
      >搜企业</view>
    </view>

    <!-- 无输入：搜索历史 -->
    <view v-if="!searchText" class="ov-history">
      <view v-if="searchHistory.length > 0" class="ov-history-header">
        <text class="ov-history-title">搜索历史</text>
        <view class="ov-history-clear" @tap="clearSearchHistory"><text>清除</text></view>
      </view>
      <view v-if="searchHistory.length > 0" class="ov-history-tags">
        <text
          v-for="(tag, index) in searchHistory"
          :key="index"
          class="ov-history-tag"
          @tap="fillSearch(tag)"
        >{{ tag }}</text>
      </view>
      <view v-else class="ov-empty">
        <u-empty description="暂无搜索历史" />
      </view>
    </view>

    <!-- 有输入：搜索结果 -->
    <view v-else class="ov-results">
      <view v-if="searchLoading" class="ov-loading">
        <view class="ov-loading-inline">
          <u-loading size="28rpx" />
          <text>搜索中...</text>
        </view>
      </view>
      <view v-else-if="searchError && searchResults.length === 0" class="ov-error">
        <u-empty description="搜索失败" />
        <view class="ov-retry" @tap="doSearch"><text>重新加载</text></view>
      </view>
      <view v-else-if="!searchLoading && searchResults.length === 0 && searched" class="ov-empty">
        <u-empty description="未找到相关内容" />
      </view>
      <view v-else-if="searchResults.length > 0" class="ov-result-list">
        <template v-if="activeTab === 'demand'">
          <view
            v-for="item in searchResults"
            :key="item.id"
            class="ov-result-card"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="goSearchDemand(item)"
          >
            <text class="ov-result-title">{{ item.title }}</text>
            <view class="ov-result-meta">
              <text class="ov-meta-tag" :class="'ov-meta-tag--' + bizTypeTagType(item.biz_type)">{{ bizTypeLabel(item.biz_type) }}</text>
              <text v-if="item.district" class="ov-meta-text">{{ item.district }}</text>
              <text class="ov-meta-text">{{ formatBudget(item.budget_fen) }}</text>
              <text class="ov-meta-date">{{ formatDate(item.created_at) }}</text>
            </view>
          </view>
        </template>
        <template v-else>
          <view
            v-for="item in searchResults"
            :key="item.id"
            class="ov-result-card ov-ent-card"
            hover-class="tap-fade"
            hover-stay-time="120"
            @tap="goSearchEnterprise(item)"
          >
            <view class="ov-ent-icon">
              <text class="ov-ent-icon-text">企</text>
            </view>
            <view class="ov-ent-title-row">
              <text class="ov-ent-name">{{ item.name || item.enterprise_name }}</text>
              <text v-if="item.description || item.business_scope" class="ov-ent-desc">{{ item.description || item.business_scope || '' }}</text>
            </view>
          </view>
        </template>
      </view>
    </view>
  </view>

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
import DemandLoop from './components/DemandLoop.vue'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'
import { request, BASE_URL } from '../../utils/request'

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
  '巴南区', '两江新区', '涪陵区', '长寿区', '江津区', '合川区', '永川区',
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
  findProject: '/static/home/icons/briefcase-blue.svg', // 浅色底用彩色描边
  publishDemand: '/static/home/icons/edit.svg',
  // 六项核心服务
  services: {
    demand: '/static/home/icons/demand.svg', // 需求大厅
    trade: '/static/home/icons/trade.svg', // 场地预约
    enterprise: '/static/home/icons/briefcase-blue.svg', // 企业入驻
    shop: '/static/home/icons/ecoservice.svg', // 生态服务
    training: '/static/home/icons/training.svg', // 培训认证
    pilot: '/static/home/icons/pilot.svg', // 认证飞手
    policy: '/static/home/icons/policy.svg', // 政策资讯
  },
}
const services = ref([
  { name: '需求大厅', icon: HOME_ICONS.services.demand, path: '/pages/demands/index', tab: true },
  { name: '培训认证', icon: HOME_ICONS.services.training, path: '/pkg-talent/pages/training/courses', tab: false },
  { name: '企业入驻', icon: HOME_ICONS.services.enterprise, path: '/pkg-eco/pages/enterprise/register', tab: false },
  { name: '场地预约', icon: HOME_ICONS.services.trade, path: '/pkg-service/pages/testsites/list', tab: false },
  { name: '认证飞手', icon: HOME_ICONS.services.pilot, path: '/pkg-talent/pages/pilots/list', tab: false },
  { name: '生态服务', icon: HOME_ICONS.services.shop, path: '/pages/services/index', tab: true },
])
const goService = (s) => {
  if (s.tab) safeSwitchTab(s.path)
  else safeNavigateTo(s.path)
}
const goAllServices = () => safeSwitchTab('/pages/services/index')
// 平台公告 = 政策资讯：公告条点击直接进入合规资讯列表
const goNotices = () => safeNavigateTo('/pkg-service/pages/compliance/news')

/* ================= 供需项目 ================= */
const demandState = ref('loading')
const demands = ref([])
const enterprises = ref([]) // 企业展示区（/api/v1/enterprises/public 已认证企业）
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
  // /uploads/ 相对路径必须补域名（否则小程序按本地包资源加载 → 白图 → 降级静态图，
  // 首页所有卡片都变成同一张图）；/static/ 是本地包资源保持原样；http 开头原样
  const resolveImgUrl = (u) => (u.indexOf('http') === 0 || u.indexOf('/static/') === 0 ? u : BASE_URL + u)
  const imgs = Array.isArray(d.images) ? d.images.filter((u) => typeof u === 'string' && u.trim()).map(resolveImgUrl) : []
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
    description: String(d.description || '').trim(),
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

// 超高图定向裁剪：旧上传图未走 16:9 裁剪（竖图在 112px 列中可撑到 200px+），
// 把卡片撑高后右侧数据与其他小图卡不对齐；needsCrop=true 时图区裁到 126px
// （aspectFill 填满无变形），16:9 等矮图不受影响。
// 不能用 @load 事件标记（同 URL 图片从缓存加载时 load 不触发，刷新后标记丢失），
// 改用 uni.getImageInfo 在数据渲染前获取原始比例 + URL 级缓存，刷新秒回。
const imgRatioCache = {}
const getImageRatio = (url) =>
  new Promise((resolve) => {
    if (imgRatioCache[url]) return resolve(imgRatioCache[url])
    uni.getImageInfo({
      src: url,
      success: (info) => {
        if (info.width > 0 && info.height > 0) {
          imgRatioCache[url] = { w: info.width, h: info.height }
        }
        resolve(imgRatioCache[url] || null)
      },
      fail: () => resolve(null),
    })
  })

const markTallImages = async (items) => {
  if (!items || !items.length) return
  await Promise.all(
    items.map(async (it) => {
      const ratio = await getImageRatio(it.image)
      if (ratio && ratio.h / ratio.w > 1.2) it.needsCrop = true
    })
  )
}

// 重点卡图片点击 → 预览原图（发布端已统一 16:9，此处兼容旧数据查看原图）
const previewFeaturedImg = (d) => {
  if (d.image) uni.previewImage({ urls: [d.image] })
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
    const [homeRes, demandsRes, entRes] = await Promise.allSettled([
      request({ url: '/api/v1/home' + cityQuery }),
      request({ url: '/api/v1/demands', data: demandsParams }),
      request({ url: '/api/v1/enterprises/public' }),
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
      // 渲染前标记超高图（getImageInfo 原始比例，URL 级缓存），
      // 避免竖图在 112px 列中把卡片撑高、右侧数据与其他小图卡错位
      await markTallImages(items)
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

    // 企业展示独立处理：失败/空数据 → 隐藏板块
    if (entRes.status === 'fulfilled') {
      const ents = Array.isArray(entRes.value) ? entRes.value : []
      enterprises.value = ents.slice(0, 10)
    } else {
      enterprises.value = []
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
const goSearch = () => openSearch()
const goMessages = () => safeNavigateTo('/pages/messages/index')
const goFindProject = () => safeNavigateTo('/pages/demands/list')
const goPublishDemand = () => safeNavigateTo('/pkg-demand/pages/demands/publish')
const goDemandsList = () => safeNavigateTo('/pages/demands/list')
const goDemandDetail = (d) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(d.id))
// 企业无公开详情页（status 是"我的企业"），展示区卡片与"全部"统一进企业列表页
const goEnterpriseList = () => safeNavigateTo('/pkg-eco/pages/enterprise/list')

/* 企业展示辅助 */
const splitTags = (str) => {
  if (!str) return []
  return String(str).split(',').map((t) => t.trim()).filter(Boolean)
}
// 相对路径（存库格式）→ 完整 URL（预览格式），完整 URL 原样返回
const resolveEntImg = (u) => (u.indexOf('http') === 0 ? u : BASE_URL + u)
const firstCategory = (e) => {
  const cats = splitTags(e.industry_category)
  return cats.length ? cats[0] : ''
}

/* ================= 就地搜索（覆盖层内搜索，逻辑与 search/index.vue 一致） ================= */
const showSearch = ref(false)
const searchText = ref('')
const activeTab = ref('demand') // demand | enterprise
const searchHistory = ref([])
const searchResults = ref([])
const searchLoading = ref(false)
const searchError = ref('')
const searched = ref(false)

const SEARCH_HISTORY_KEY = 'searchHistory'
const SEARCH_HISTORY_MAX = 10

const openSearch = () => {
  showSearch.value = true
  activeTab.value = 'demand'
  loadSearchHistory()
}
const closeSearch = () => {
  showSearch.value = false
  searchText.value = ''
  searchResults.value = []
  searched.value = false
  searchError.value = ''
}

/* --- 历史 --- */
function loadSearchHistory() {
  try {
    const raw = uni.getStorageSync(SEARCH_HISTORY_KEY)
    const arr = raw ? JSON.parse(raw) : []
    searchHistory.value = Array.isArray(arr) ? arr : []
  } catch (e) {
    searchHistory.value = []
  }
}
function saveSearchHistory() {
  try {
    uni.setStorageSync(SEARCH_HISTORY_KEY, JSON.stringify(searchHistory.value))
  } catch (e) {
    // ignore
  }
}
function addToSearchHistory(keyword) {
  if (!keyword || !keyword.trim()) return
  const kw = keyword.trim()
  const list = searchHistory.value.filter((w) => w !== kw)
  list.unshift(kw)
  searchHistory.value = list.slice(0, SEARCH_HISTORY_MAX)
  saveSearchHistory()
}
const clearSearchHistory = () => {
  searchHistory.value = []
  uni.removeStorageSync(SEARCH_HISTORY_KEY)
}
const fillSearch = (tag) => {
  searchText.value = tag
  doSearch()
}

/* --- 搜索 --- */
const onSearchInput = (val) => {
  searchText.value = val
  if (!searchText.value) {
    searchResults.value = []
    searched.value = false
    searchError.value = ''
  }
}
const onSearch = () => {
  if (searchText.value && searchText.value.trim()) doSearch()
}
async function doSearch() {
  const kw = searchText.value && searchText.value.trim()
  if (!kw) {
    searchResults.value = []
    searched.value = false
    return
  }
  searchLoading.value = true
  searchError.value = ''
  searchResults.value = []
  try {
    const res = await request({
      url: '/api/v1/search',
      data: { q: kw, type: activeTab.value },
    })
    const data = Array.isArray(res) ? res : (res && res.items) || []
    searchResults.value = data
    searched.value = true
    addToSearchHistory(kw)
  } catch (e) {
    searchError.value = '网络异常，请稍后重试'
  } finally {
    searchLoading.value = false
  }
}
const onTabChange = (index) => {
  activeTab.value = index === 1 ? 'enterprise' : 'demand'
  searchResults.value = []
  searched.value = false
  searchError.value = ''
  if (searchText.value && searchText.value.trim()) doSearch()
}

/* --- 结果跳转 --- */
const goSearchDemand = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
const goSearchEnterprise = (item) => safeNavigateTo('/pkg-eco/pages/enterprise/status?id=' + encodeURIComponent(item.id))

/* --- 展示辅助 --- */
function bizTypeLabel(type) {
  const map = {
    cable_inspection: '巡检',
    plant_transport: '植保',
    spray_pesticide: '农药',
    trade_lease: '租赁',
    clean_paint: '清洗',
    other: '其他',
  }
  return map[type] || type || '其他'
}
function bizTypeTagType(type) {
  const map = {
    cable_inspection: 'primary',
    plant_transport: 'success',
    spray_pesticide: 'warning',
    trade_lease: 'danger',
    clean_paint: 'primary',
    other: 'default',
  }
  return map[type] || 'default'
}
function formatBudget(fen) {
  if (fen == null || fen === 0) return '面议'
  const yuan = (fen / 100).toFixed(2)
  return yuan.replace(/\.00$/, '') + ' 元'
}
function formatDate(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  const m = d.getMonth() + 1
  const day = d.getDate()
  return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
}

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
  gap: 4px;
  font-size: 14px;
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
  min-height: 58px;
  border-radius: 9px;
  padding: 0 10px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 9px;
  font-size: 12px;
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
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.78);
}
.service-icon-img {
  width: 24px;
  height: 24px;
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

/* ================= 企业展示：横向滑动卡片 ================= */
.enterprise-section {
  padding-top: 8px;
  padding-bottom: 8px;
}
.enterprise-scroll {
  margin: 0 -12px;
  padding: 0 12px;
  box-sizing: content-box;
  white-space: nowrap;
}
.enterprise-scroll::-webkit-scrollbar {
  display: none;
}
.enterprise-row {
  display: inline-flex;
  gap: 8px;
}
.enterprise-card {
  width: 180px;
  height: 64px;
  flex-shrink: 0;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  background: #fff;
  padding: 0 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  box-sizing: border-box;
}
.ent-logo {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  background: #F4F8FC;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ent-logo-img {
  width: 100%;
  height: 100%;
}
.ent-logo-fallback {
  font-size: 18px;
  font-weight: 700;
  color: #0A66C2;
}
.ent-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.ent-card-name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 700;
  color: #17212B;
}
.ent-card-tags {
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.ent-card-tag {
  font-size: 10px;
  color: #667085;
  background: #F1F3F5;
  padding: 2px 7px;
  border-radius: 4px;
  line-height: 1.5;
}
.ent-card-tag--member {
  color: #0A66C2;
  background: #EAF3FB;
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
  min-height: 36px;
  padding: 0 16px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  color: #667085;
  font-size: 12px;
  box-sizing: border-box;
}
.category-chip + .category-chip {
  margin-left: 8px;
}
.category-chip.active {
  color: #fff;
  border-color: #0A66C2;
  background: #0A66C2;
  font-weight: 700;
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
  height: 100px;
  border-radius: 8px;
  background: #EDF0F3;
  position: relative;
  overflow: hidden;
}
.skeleton-card.featured-skeleton {
  height: 234px;
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

/* ================= 就地搜索覆盖层（对齐需求大厅：深蓝条 + 白底下划线 tabs） ================= */
.ov-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9990;
  background: #F4F6F8;
  display: flex;
  flex-direction: column;
}
.ov-topbar {
  background: #074D92;
  color: #fff;
  padding: 16rpx 24rpx 28rpx;
  /* 顶部避让由 :style 动态 paddingTop 控制 */
}
.ov-topbar-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.ov-back {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ov-back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(45deg);
  margin-left: 10rpx;
}
.ov-title {
  flex: 1;
  font-size: 38rpx;
  font-weight: 700;
  text-align: center;
}
.ov-spacer { width: 60rpx; }
.ov-search-box {
  width: 100%;
  height: 44px;
  margin-top: 24rpx;
  border-radius: 7px;
  background: #fff;
  display: flex;
  align-items: center;
  gap: 14rpx;
  padding: 0 24rpx;
  box-sizing: border-box;
}
.ov-search-icon {
  width: 30rpx;
  height: 30rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.ov-search-icon::after {
  content: '';
  position: absolute;
  right: -12rpx;
  bottom: -7rpx;
  width: 14rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #98A2B3;
  transform: rotate(45deg);
}
.ov-search-input { flex: 1; font-size: 28rpx; color: #17212B; }
.ov-search-ph { color: #98A2B3; }

/* 白底下划线 Tabs */
.ov-tabs {
  display: flex;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 0 32rpx;
  flex-shrink: 0;
}
.ov-tab {
  flex: 1;
  height: 92rpx;
  line-height: 92rpx;
  text-align: center;
  position: relative;
  color: #667085;
  font-weight: 600;
  font-size: 28rpx;
}
.ov-tab.active { color: #0A66C2; }
.ov-tab.active::after {
  content: '';
  position: absolute;
  width: 56rpx;
  height: 6rpx;
  background: #0A66C2;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  border-radius: 3rpx;
}

/* 历史 */
.ov-history { padding: 24rpx; }
.ov-history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}
.ov-history-title { font-size: 28rpx; font-weight: 600; color: #17212B; }
.ov-history-clear { font-size: 24rpx; color: #667085; }
.ov-history-tags { display: flex; flex-wrap: wrap; gap: 20rpx; }
.ov-history-tag {
  padding: 10rpx 28rpx;
  background: #fff;
  border: 1rpx solid #E4E7EC;
  border-radius: 8rpx;
  font-size: 26rpx;
  color: #344054;
}

/* 结果 */
.ov-results {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-top: 8px;
}
.ov-loading {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
.ov-loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: #667085;
}
.ov-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}
.ov-retry {
  margin-top: 12px;
  padding: 8px 24px;
  background: #0A66C2;
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
}
.ov-empty { padding-top: 60px; }
.ov-result-list { padding: 20rpx 24rpx 24rpx; }
.ov-result-card {
  background: #fff;
  border: 1rpx solid #E4E7EC;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.ov-result-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
}
.ov-result-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-top: 14rpx;
}
.ov-meta-tag {
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
  font-size: 20rpx;
  font-weight: 600;
}
.ov-meta-tag--primary { background: #EAF3FB; color: #0A66C2; }
.ov-meta-tag--success { background: #E8F7EF; color: #16A34A; }
.ov-meta-tag--warning { background: #FDF1E7; color: #E46426; }
.ov-meta-tag--danger { background: #FDECEC; color: #E84C3D; }
.ov-meta-tag--default { background: #F0F3F6; color: #667085; }
.ov-meta-text { font-size: 24rpx; color: #667085; }
.ov-meta-date { font-size: 24rpx; color: #98A2B3; }

/* 企业结果 */
.ov-ent-card { display: flex; align-items: center; gap: 20rpx; }
.ov-ent-icon {
  width: 72rpx;
  height: 72rpx;
  flex-shrink: 0;
  border-radius: 50%;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ov-ent-icon-text { font-size: 30rpx; font-weight: 600; color: #0A66C2; }
.ov-ent-title-row {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}
.ov-ent-name { font-size: 28rpx; font-weight: 700; color: #17212B; }
.ov-ent-desc {
  font-size: 24rpx;
  color: #667085;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* ================= 按压反馈 ================= */
.tap-fade {
  opacity: 0.85;
}
.tap-scale {
  transform: scale(0.985);
  opacity: 0.9;
}
</style>
