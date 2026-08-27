<template>
  <Layout :current="1">
    <view class="hall-page" :class="{ 'no-motion': noMotion }">
      <!-- ═══════ 深蓝顶部 ═══════ -->
      <view class="topbar" :style="{ paddingTop: statusBarHeight + 'px' }">
        <view class="topbar-row">
          <text class="top-title">供需大厅</text>
        </view>
        <view class="search-trigger" hover-class="tap-fade" @tap="goSearch">
          <view class="search-icon"></view>
          <text class="search-placeholder">搜索需求、服务或设备</text>
        </view>
      </view>

      <!-- ═══════ 一级 Tab：需求大厅 / 供给大厅 ═══════ -->
      <view class="tabs">
        <view
          class="tab"
          :class="{ active: primary === 'demand' }"
          @tap="switchPrimary('demand')"
        >需求大厅</view>
        <view
          class="tab"
          :class="{ active: primary === 'supply' }"
          @tap="switchPrimary('supply')"
        >供给大厅</view>
      </view>

      <!-- 供给分段：商品设备 / 服务能力 -->
      <view v-if="primary === 'supply'" class="subtabs">
        <view
          class="subtab"
          :class="{ active: supplyKind === 'product' }"
          @tap="switchSupplyKind('product')"
        >商品设备</view>
        <view
          class="subtab"
          :class="{ active: supplyKind === 'service' }"
          @tap="switchSupplyKind('service')"
        >服务能力</view>
      </view>

      <!-- ═══════ 一级筛选：分类下划线 tab 分段（「全部」带 ▾ 独立开关；对齐科技成果库） ═══════ -->
      <view class="stage-wrap">
        <view class="stages">
          <view
            v-for="t in stageTabs"
            :key="t.value"
            class="stg"
            :class="{ on: activeStage === t.value }"
            @tap="pickStageTab(t.value)"
          >
            <text>{{ t.label }}</text>
            <!-- ▾ 独立面板开关：未停在「全部」时点「全部」先清分类；停在「全部」时再点开面板 -->
            <text v-if="t.value === 'all' && !isProductMode" class="stg-arr" :class="{ up: panel === 'all' }" @tap.stop="togglePanel">▾</text>
          </view>
        </view>
        <!-- 二级筛选面板：地区 / 预算 / 排序 chips（absolute 浮层，展开不挤动下方内容） -->
        <view v-if="panel === 'all' && !isProductMode" class="field-panel" :class="{ closing }">
          <view class="p-group">所在地区</view>
          <view class="p-chips">
            <text class="p-chip" :class="{ act: filterRegion === '不限' }" @tap="pickRegion('不限')">全部地区</text>
            <text v-for="r in regionOptions.slice(1)" :key="r" class="p-chip" :class="{ act: filterRegion === r }" @tap="pickRegion(r)">{{ r }}</text>
          </view>
          <view class="p-group">{{ primary === 'demand' ? '预算金额' : '价格范围' }}</view>
          <view class="p-chips">
            <text class="p-chip" :class="{ act: filterPrice === '不限' }" @tap="pickPrice('不限')">全部</text>
            <text v-for="p in priceOptions.slice(1)" :key="p" class="p-chip" :class="{ act: filterPrice === p }" @tap="pickPrice(p)">{{ p }}</text>
          </view>
          <view class="p-group">排序方式</view>
          <view class="p-chips">
            <text v-for="s in sortOptions" :key="s.value" class="p-chip" :class="{ act: sortBy === s.value }" @tap="pickSort(s.value)">{{ s.label }}</text>
          </view>
        </view>
      </view>
      <!-- 蒙层：从分段底部开始置灰，点外部退场收起 -->
      <view v-if="panel && !isProductMode" class="panel-mask" :style="{ top: maskTop + 'px' }" @tap="startClosePanel" />

      <!-- ═══════ 匹配条（商品模式为电商页，不展示匹配引导） ═══════ -->
      <view v-if="listState === 'ready' && visibleList.length > 0 && !isProductMode" class="match-strip">
        <view class="match-mark"><text class="match-mark-icon">⇄</text></view>
        <view class="match-copy">
          <text class="match-title">{{ primary === 'demand' ? '为需求方推荐承接能力' : '发现与你匹配的需求' }}</text>
          <text class="match-desc">依据业务类型、区域和服务能力进行匹配</text>
        </view>
        <view class="match-link" @tap="goMatches">查看匹配</view>
      </view>

      <!-- ═══════ 列表标题 ═══════ -->
      <view class="section-head">
        <text class="section-title">{{ sectionTitle }}</text>
        <text class="section-note">共 {{ visibleList.length }} 条</text>
      </view>

      <!-- ═══════ 列表四种状态 ═══════ -->
      <!-- 加载中 -->
      <view v-if="listState === 'loading'" class="skeleton-list">
        <view class="skeleton-card"></view>
        <view class="skeleton-card"></view>
        <view class="skeleton-card"></view>
      </view>

      <!-- 加载失败 -->
      <view v-else-if="listState === 'error'" class="state-panel">
        <view class="state-mark err">!</view>
        <text class="state-title">内容加载失败</text>
        <text class="state-desc">请检查网络后重新加载</text>
        <view class="state-btn" @tap="reload">重新加载</view>
      </view>

      <!-- 空数据 -->
      <view v-else-if="listState === 'empty' || visibleList.length === 0" class="state-panel">
        <view class="state-mark">⌁</view>
        <text class="state-title">没有找到匹配的{{ kindLabel }}</text>
        <text class="state-desc">换个筛选条件试试，或先发布一条信息</text>
        <view class="state-btn" @tap="resetFilters">清除筛选</view>
      </view>

      <!-- 正常列表 -->
      <!-- 商品模式：电商两列宫格（大图 + 价格 + 品牌型号 + 成色/浏览） -->
      <view v-if="isProductMode && visibleList.length > 0" class="ecom-grid">
        <view
          v-for="item in visibleList"
          :key="item.id"
          class="ecom-card"
          hover-class="tap-fade"
          @tap="goProductDetail(item)"
        >
          <view class="ecom-img-wrap">
            <image :src="item.image" mode="aspectFill" class="ecom-img" @error="onProductImgError(item)" />
            <text v-if="item.isUsed" class="ecom-used-tag">二手</text>
          </view>
          <view class="ecom-body">
            <text class="ecom-price">¥<text class="ecom-price-num">{{ item.price }}</text></text>
            <text class="ecom-title">{{ item.title }}</text>
            <text class="ecom-spec">{{ item.spec }}</text>
            <view class="ecom-foot">
              <text class="ecom-cat">{{ item.cat }}</text>
              <text class="ecom-views">{{ item.views ? '已浏览 ' + item.views + ' 次' : '平台商品' }}</text>
            </view>
          </view>
        </view>
      </view>
      <view v-else class="card-list">
        <view
          v-for="item in visibleList"
          :key="item.id"
          class="trade-card"
          hover-class="tap-fade"
          @tap="goDetail(item)"
        >
          <view class="trade-card-main">
            <image :src="item.image" mode="aspectFill" class="trade-visual" @error="onImageError(item)" />
            <view class="trade-body">
              <view class="tag-row">
                <text class="tag blue">{{ item.cat }}</text>
                <text class="tag" :class="statusTagClass(item)">{{ item.status }}</text>
              </view>
              <text class="trade-title">{{ item.title }}</text>
              <view class="trade-meta">
                <text class="meta-item">{{ item.region }}</text>
                <text class="meta-item">{{ item.time }}</text>
              </view>
            </view>
          </view>
          <view class="trade-footer">
            <view class="price-block">
              <text class="price">{{ item.price }}</text>
              <text class="price-unit"> {{ item.unit }}</text>
            </view>
            <view class="card-action">
              <text>查看详情</text>
              <text class="card-action-arrow">›</text>
            </view>
          </view>
        </view>
      </view>

    </view>

  </Layout>

  <!-- ═══════ 就地搜索覆盖层（点击搜索框在当前页展开，不跳搜索页） ═══════ -->
  <view v-if="showSearch" class="ov-overlay">
    <view class="ov-bar" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="ov-back" hover-class="tap-fade" hover-stay-time="120" @tap="closeSearch">
        <view class="ov-back-arrow"></view>
      </view>
      <view class="ov-search-box">
        <view class="ov-search-icon"></view>
        <input
          class="ov-search-input"
          v-model="keyword"
          placeholder="搜索需求、服务或设备"
          confirm-type="search"
          focus
          @confirm="onSearch"
        />
        <view v-if="keyword" class="ov-clear" @tap="keyword = ''"><text class="ov-clear-x">×</text></view>
      </view>
      <view class="ov-search-btn" hover-class="tap-fade" @tap="onSearch"><text>搜索</text></view>
    </view>

    <!-- 有结果 -->
    <view v-if="searched" class="ov-results">
      <view class="ov-result-head">找到 {{ searchResults.length }} 条相关内容</view>
      <view v-if="searchResults.length" class="ov-card-list">
        <view
          v-for="item in searchResults"
          :key="item.id"
          class="ov-trade-card"
          hover-class="tap-fade"
          @tap="goSearchResult(item)"
        >
          <view class="ov-trade-card-main">
            <image :src="item.image" mode="aspectFill" class="ov-trade-visual" @error="onImageError(item)" />
            <view class="ov-trade-body">
              <view class="ov-tag-row">
                <text class="ov-tag ov-tag-blue">{{ item.cat }}</text>
                <text class="ov-tag" :class="'ov-tag-' + statusTagClass(item)">{{ item.status }}</text>
              </view>
              <text class="ov-trade-title">{{ item.title }}</text>
              <view class="ov-trade-meta">
                <text>{{ item.region }}</text>
                <text>{{ item.time }}</text>
              </view>
            </view>
          </view>
          <view class="ov-trade-footer">
            <view class="ov-price-block">
              <text class="ov-price">{{ item.price }}</text>
              <text class="ov-price-unit"> {{ item.unit }}</text>
            </view>
            <view class="ov-card-action"><text>查看详情 ›</text></view>
          </view>
        </view>
      </view>
      <view v-else class="ov-state-panel">
        <view class="ov-state-mark">⌁</view>
        <text class="ov-state-title">没有找到相关内容</text>
        <text class="ov-state-desc">换个关键词试试</text>
      </view>
    </view>

    <!-- 推荐 / 最近搜索 -->
    <view v-else class="ov-suggest">
      <view class="ov-search-block">
        <text class="ov-block-title">推荐搜索</text>
        <view class="ov-keyword-row">
          <view v-for="w in hotWords" :key="w" class="ov-keyword" @tap="keyword = w; onSearch()">{{ w }}</view>
        </view>
      </view>
      <view class="ov-search-block">
        <text class="ov-block-title">最近搜索</text>
        <view class="ov-keyword-row">
          <view v-for="w in recentWords" :key="w" class="ov-keyword" @tap="keyword = w; onSearch()">{{ w }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request } from '../../utils/request'
import { safeNavigateTo } from '../../utils/nav'
import { useReduceMotion } from '@/utils/motion'
import {
  HALL_CATEGORIES, kindTypeLabel, isEnded, normalizeDemand,
  IMG_SOLAR, IMG_LIFT, IMG_HERO,
  PRODUCT_CATEGORIES, normalizeProduct, normalizeService, getLocalLiveCards,
} from '../../utils/hallData'

// 状态栏高度：custom 导航下 topbar 需 JS 接管（env(safe-area-inset-top) 在微信端返回 0）
const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 20

const primary = ref('demand') // demand | supply
const supplyKind = ref('product') // product | service
const activeStage = ref('all') // 分类分段：all=全部（对齐成果库阶段分段）
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）

// 重庆市 38 个区县（筛选地区选项完整）
const regionOptions = [
  '不限', '渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区', '南岸区', '北碚区',
  '渝北区', '巴南区', '两江新区', '长寿区', '江津区', '合川区', '永川区', '南川区',
  '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区',
  '武隆区', '万州区', '黔江区', '涪陵区', '奉节县', '云阳县', '忠县', '垫江县', '丰都县',
  '城口县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县',
]
const priceOptions = ['不限', '1 万以下', '1-5 万', '5 万以上', '面议']
const sortOptions = [
  { label: '最新发布', value: 'newest' },
  { label: '匹配度优先', value: 'match' },
  { label: '价格优先', value: 'price' },
]

const filterRegion = ref('不限')
const filterPrice = ref('不限')
const sortBy = ref('newest')

// ---- 筛选面板（对齐科技成果库：tab 分段 + ▾ 浮层面板 + 蒙层） ----
const panel = ref('') // '' = 收起；'all' = 「全部」段的面板展开
const closing = ref(false) // 面板退场中（先播退场动画再 v-if 移除）
const maskTop = ref(0) // 蒙层起点（面板打开时实测：tab 分段底部）
let panelCloseT = null
const PANEL_CLOSE_MS = 210 // 退场动画 .21s ease-in

const categories = computed(() => {
  if (isProductMode.value) return PRODUCT_CATEGORIES
  const kind = primary.value === 'demand' ? 'demand' : supplyKind.value
  return HALL_CATEGORIES[kind]
})

// 分类 → 一级分段 tab（首项「全部」映射为 all，带 ▾）
const stageTabs = computed(() => categories.value.map(c => c === '全部' ? { label: c, value: 'all' } : { label: c, value: c }))

const kindLabel = computed(() => kindTypeLabel(primary.value, supplyKind.value))
const sectionTitle = computed(() => {
  if (primary.value === 'demand') return '最新需求'
  return supplyKind.value === 'service' ? '可对接服务' : '优选商品设备'
})

// 商品设备模式：电商两列宫格展示（独立于需求/服务列表）
const isProductMode = computed(() => primary.value === 'supply' && supplyKind.value === 'product')

/* ================= 列表状态 ================= */
const listState = ref('loading') // loading | ready | empty | error
const list = ref([])

async function fetchList(showLoading = true) {
  if (showLoading) listState.value = 'loading'

  // 发布页已上架的本地内容并入列表（仅 DEV：后端未接入期间的展示打通；
  // 生产构建不混入本地 storage，避免与后端真实数据重复/失真）
  const local = () => {
    if (!import.meta.env.DEV) return []
    if (primary.value === 'demand') return getLocalLiveCards('demand')
    return supplyKind.value === 'product' ? getLocalLiveCards('product') : getLocalLiveCards('service')
  }
  const merge = (remote) => {
    list.value = [...local(), ...remote]
    listState.value = list.value.length ? 'ready' : 'empty'
  }
  // 接口失败但本地有已上架发布时，降级展示本地内容（演示闭环）
  const fallback = () => {
    list.value = local()
    listState.value = list.value.length ? 'ready' : 'error'
  }

  if (primary.value === 'demand') {
    try {
      const res = await request({
        url: '/api/v1/demands',
        data: { page: 1, page_size: 20, sort: 'newest' },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      merge(items.map(normalizeDemand).filter(Boolean))
    } catch (e) {
      fallback()
    }
  } else if (supplyKind.value === 'product') {
    // 商品设备：电商模式走真实商品接口
    try {
      const res = await request({
        url: '/api/v1/products',
        data: { page: 1, page_size: 50 },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      merge(items.map(normalizeProduct).filter(Boolean))
    } catch (e) {
      fallback()
    }
  } else {
    // 服务能力：真实服务接口（/api/v1/service-listings 公开列表）
    try {
      const res = await request({
        url: '/api/v1/service-listings',
        data: { page: 1, page_size: 20 },
      })
      const data = Array.isArray(res) ? res : (res && res.data) || res || {}
      const items = Array.isArray(data) ? data : (data && data.items) || []
      merge(items.map(normalizeService).filter(Boolean))
    } catch (e) {
      fallback()
    }
  }
}

const reload = () => { fetchList(true) }

/* ================= 交互 ================= */
function switchPrimary(value) {
  if (primary.value === value) return
  clearTimeout(panelCloseT); panelCloseT = null; closing.value = false; panel.value = '' // 切量即收起面板（防蒙层滞留）
  primary.value = value
  activeStage.value = 'all'
  fetchList(true)
}

function switchSupplyKind(value) {
  if (supplyKind.value === value) return
  clearTimeout(panelCloseT); panelCloseT = null; closing.value = false; panel.value = '' // 切量即收起面板（防蒙层滞留）
  supplyKind.value = value
  activeStage.value = 'all'
  fetchList(true)
}

// ---- 筛选交互（对齐科技成果库方案 A：tab 分段 + ▾ 面板 + 蒙层） ----
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
  // 蒙层起点 = 分段容器底部（实测，头部不蒙）
  uni.nextTick(() => {
    uni.createSelectorQuery().select('.stage-wrap').boundingClientRect((rect) => {
      if (rect && rect.bottom) maskTop.value = Math.round(rect.bottom)
    }).exec()
  })
}
// 方案 A：非「全部」tab 再点取消回全部；「全部」未停先清筛、停下再点开面板；▾ 独立开关
const pickStageTab = (k) => {
  if (k !== 'all') {
    startClosePanel()
    activeStage.value = activeStage.value === k ? 'all' : k
    return
  }
  if (activeStage.value !== 'all') {
    startClosePanel()
    activeStage.value = 'all'
    return
  }
  togglePanel()
}
// 面板 chips 点选即筛（client 端可见，多组并行），再点取消；不关面板（可继续调其他维度）
const pickRegion = (r) => { filterRegion.value = filterRegion.value === r ? '不限' : r }
const pickPrice = (p) => { filterPrice.value = filterPrice.value === p ? '不限' : p }
const pickSort = (v) => { sortBy.value = v }
const resetFilters = () => {
  activeStage.value = 'all'
  filterRegion.value = '不限'
  filterPrice.value = '不限'
  sortBy.value = 'newest'
  startClosePanel()
}

const goSearch = () => openSearch()

/* ================= 就地搜索（覆盖层内本地过滤） ================= */
const showSearch = ref(false)
const keyword = ref('')
const searched = ref(false)
const searchResults = ref([])

const hotWords = ['巡检', '航拍', '测绘', '吊运', '设备租赁', '植保']
const recentWords = ['光伏巡检', 'M350']

const openSearch = () => {
  showSearch.value = true
}
const closeSearch = () => {
  showSearch.value = false
  keyword.value = ''
  searched.value = false
  searchResults.value = []
}

// 就地搜索走真实接口 /api/v1/search（与 home/index.vue 同款模式，返回 { demands, enterprises }）。
// 需求结果用 normalizeDemand 映射为现有卡片结构；不再检索本地 mock
// （生产构建 getKindItems 返回 []，本地检索正式包永远无结果）
async function onSearch() {
  const kw = keyword.value.trim()
  searched.value = true
  if (!kw) {
    searchResults.value = []
    return
  }
  try {
    const res = await request({ url: '/api/v1/search', data: { q: kw } })
    const demands = (res && Array.isArray(res.demands)) ? res.demands : []
    searchResults.value = demands.map(normalizeDemand).filter(Boolean)
  } catch (e) {
    searchResults.value = []
    uni.showToast({ title: '搜索失败，请稍后重试', icon: 'none' })
  }
}

const goSearchResult = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
const goMatches = () => safeNavigateTo('/pkg-demand/pages/demands/matches')
const goDetail = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
// 商品模式：跳电商商品详情页。
// 本地发布的商品已接后端（backendId 非空）→ 用后端 id 进真商品详情；仅旧版未接
// 后端的纯本地商品才走大厅详情页兜底展示。
const goProductDetail = (item) => {
  if (item.backendId) {
    safeNavigateTo('/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(item.backendId))
    return
  }
  if (String(item.id).indexOf('post-') === 0) {
    safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
    return
  }
  safeNavigateTo('/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(item.id))
}
const onProductImgError = (item) => {
  if (item.image !== IMG_HERO) item.image = IMG_HERO
}


/* ================= 展示辅助 ================= */
function statusTagClass(item) {
  if (isEnded(item)) return 'gray'
  if (item.type === '商品') return 'orange'
  return 'green'
}

function onImageError(item) {
  if (item.image !== IMG_LIFT && item.image !== IMG_SOLAR) {
    item.image = item.cat === '吊运' ? IMG_LIFT : IMG_SOLAR
  }
}

/* ================= 过滤展示 ================= */
const visibleList = computed(() => {
  let out = list.value
  if (activeStage.value !== 'all') out = out.filter((i) => i.cat === activeStage.value)
  if (filterRegion.value !== '不限' && !isProductMode.value) out = out.filter((i) => i.region.includes(filterRegion.value))
  // 预算区间（需求 budget_fen / 服务 price_fen，单位分——后端是 snake_case）
  if (filterPrice.value !== '不限' && !isProductMode.value) {
    const fen = (i) => Number(i.budget_fen || i.price_fen || 0)
    const f = filterPrice.value
    if (f === '1 万以下') out = out.filter((i) => fen(i) > 0 && fen(i) < 1000000)
    else if (f === '1-5 万') out = out.filter((i) => fen(i) >= 1000000 && fen(i) <= 5000000)
    else if (f === '5 万以上') out = out.filter((i) => fen(i) > 5000000)
    else if (f === '面议') out = out.filter((i) => fen(i) === 0)
  }
  // 排序：价格优先 = 预算升序（低价优先）；最新发布/匹配度保持列表原序
  if (sortBy.value === 'price' && !isProductMode.value) {
    out = [...out].sort((a, b) => (Number(a.budget_fen || a.price_fen || 0)) - (Number(b.budget_fen || b.price_fen || 0)))
  }
  return out
})

/* ================= 生命周期 ================= */
onLoad(() => {
  checkMotion() // 减弱动效检测（无障碍）
  fetchList(true)
})

onPullDownRefresh(() => {
  fetchList(false).finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.hall-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 120rpx;
}

.tap-fade { opacity: 0.85; }

/* ═══════ 深蓝顶部 ═══════ */
.topbar {
  background: #074D92;
  color: #fff;
  padding: 16rpx 24rpx 28rpx;
  padding-top: calc(env(safe-area-inset-top) + 16rpx);
}
.topbar-row {
  display: flex;
  align-items: center;
  justify-content: center;
}
.top-title {
  font-size: 38rpx;
  font-weight: 700;
  text-align: center;
}

/* 搜索框：居中、接近满宽、44px 高 */
.search-trigger {
  width: 100%;
  height: 44px;
  margin-top: 28rpx;
  border-radius: 7px;
  background: #fff;
  color: #98A2B3;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  padding: 0 24rpx;
  font-size: 26rpx;
  box-sizing: border-box;
}
.search-icon {
  width: 30rpx;
  height: 30rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.search-icon::after {
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

/* ═══════ 一级 Tab ═══════ */
.tabs {
  display: flex;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 0 32rpx;
}
.tab {
  flex: 1;
  height: 92rpx;
  line-height: 92rpx;
  text-align: center;
  position: relative;
  color: #667085;
  font-weight: 600;
  font-size: 28rpx;
}
.tab.active { color: #0A66C2; }
.tab.active::after {
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

/* 供给分段 */
.subtabs {
  display: flex;
  gap: 8rpx;
  margin: 20rpx 32rpx 16rpx;
  padding: 8rpx;
  background: #F2F5F8;
  border-radius: 16rpx;
}
.subtab {
  flex: 1;
  height: 68rpx;
  line-height: 68rpx;
  text-align: center;
  border-radius: 12rpx;
  color: #667085;
  font-size: 24rpx;
}
.subtab.active {
  color: #fff;
  background: #0A66C2;
  font-weight: 650;
}

/* ═══════ 一级筛选：下划线 tab 分段（对齐科技成果库）+ ▾ 浮层面板 + 蒙层 ═══════ */
.stage-wrap { position: relative; z-index: 42; background: #fff; border-bottom: 1px solid #EEF1F4; }
.stages { display: flex; gap: 40rpx; padding: 4rpx 28rpx 16rpx; white-space: nowrap; }
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
  animation: toc-in 0.22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.stg-arr {
  font-size: 24rpx;
  color: #667085;
  transition: transform 0.2s ease, color 0.2s ease;
  padding: 20rpx 16rpx;
  margin: -20rpx -16rpx;
}
.stg-arr.up { transform: rotate(180deg); color: #074D92; }
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
  animation: panelIn 0.3s cubic-bezier(0.32, 0.72, 0, 1);
}
.field-panel.closing { animation: panelOut 0.21s ease-in forwards; }
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
.p-chip { transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease, transform 0.3s cubic-bezier(0.34, 1.8, 0.64, 1); }
.p-chip:active { transform: scale(0.94); transition: transform 0.08s linear; }
.p-chip.act { animation: chipPop 0.3s cubic-bezier(0.34, 1.8, 0.64, 1); }
@keyframes chipPop { 0% { transform: scale(1); } 40% { transform: scale(0.94); } 100% { transform: scale(1); } }
.panel-mask {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 41;
  background: rgba(16, 24, 40, 0.2);
  animation: maskIn 0.22s ease-out;
}
@keyframes maskIn { from { opacity: 0; } to { opacity: 1; } }
/* 减弱动效（无障碍）：装饰动画/位移缩放关闭，保留淡入与颜色反馈 */
.hall-page.no-motion .stg-arr { transition: none; }
.hall-page.no-motion .p-chip { transition: none; }
.hall-page.no-motion .p-chip.act { animation: none; }
.hall-page.no-motion .stg.on::after { animation: none; }
.hall-page.no-motion .field-panel { animation: panelIn 0.3s ease-out; }
.hall-page.no-motion .field-panel.closing { animation: panelOut 0.16s ease-in forwards; }
.hall-page.no-motion .panel-mask { animation: maskIn 0.22s ease-out; }
.hall-page.no-motion .p-chip:active { transform: none; }

/* ═══════ 匹配条 ═══════ */
.match-strip {
  margin: 24rpx 32rpx 4rpx;
  padding: 24rpx;
  display: flex;
  align-items: center;
  gap: 16rpx;
  border: 1px solid #D4E5F5;
  background: #F7FBFF;
  border-radius: 16rpx;
}
.match-mark {
  width: 72rpx;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: #EAF3FB;
  color: #0A66C2;
  font-weight: 800;
}
.match-mark-icon { font-size: 40rpx; }
.match-copy { flex: 1; min-width: 0; }
.match-title { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; }
.match-desc { display: block; color: #667085; font-size: 22rpx; margin-top: 6rpx; }
.match-link { color: #0A66C2; font-size: 24rpx; font-weight: 600; white-space: nowrap; }

/* ═══════ 列表标题 ═══════ */
.section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.section-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.section-note { font-size: 24rpx; color: #667085; }

/* ═══════ 卡片 ═══════ */
.card-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 0 32rpx 32rpx;
}
.trade-card {
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  overflow: hidden;
}
.trade-card-main {
  display: flex;
  gap: 22rpx;
  padding: 24rpx;
}
.trade-visual {
  width: 164rpx;
  height: 164rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  background: #E8F2FC;
}
.trade-body { flex: 1; min-width: 0; }
.tag-row { display: flex; gap: 10rpx; align-items: center; margin-bottom: 12rpx; }
.tag {
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.gray { color: #667085; background: #F1F3F5; }
.trade-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #17212B;
  font-size: 28rpx;
  line-height: 1.42;
  font-weight: 700;
}
.trade-meta {
  display: flex;
  gap: 14rpx;
  align-items: center;
  margin-top: 14rpx;
  color: #667085;
  font-size: 22rpx;
  white-space: nowrap;
  overflow: hidden;
}
.meta-item { overflow: hidden; text-overflow: ellipsis; }
.trade-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 80rpx;
  padding: 0 24rpx;
  border-top: 1px solid #EEF1F4;
}
.price-block { display: flex; align-items: baseline; min-width: 0; }
.price { color: #DF620F; font-size: 30rpx; font-weight: 750; }
.price-unit { color: #667085; font-size: 20rpx; }
.card-action {
  color: #0A66C2;
  font-size: 24rpx;
  font-weight: 650;
  display: flex;
  align-items: center;
  gap: 4rpx;
  white-space: nowrap;
}
.card-action-arrow { font-size: 30rpx; line-height: 1; }

/* ═══════ 骨架屏 ═══════ */
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 0 32rpx;
}
.skeleton-card {
  height: 216rpx;
  border-radius: 16rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 状态面板 ═══════ */
.state-panel {
  min-height: 620rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
  color: #667085;
  font-size: 26rpx;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-mark.err { color: #D92D20; background: #FEF3F2; }
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}

/* 响应式：375px 微调 */
@media (max-width: 380px) {
  .trade-visual {
    width: 150rpx;
    height: 150rpx;
  }
}

/* ═══════ 商品设备：电商两列宫格 ═══════ */
.ecom-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  padding: 0 24rpx 20rpx;
}
.ecom-card {
  width: calc(50% - 10rpx);
  background: #fff;
  border-radius: 20rpx;
  overflow: hidden;
  margin-bottom: 20rpx;
  box-shadow: 0 4rpx 16rpx rgba(7, 77, 146, 0.06);
}
.ecom-img-wrap {
  position: relative;
  width: 100%;
  height: 320rpx;
  background: #F0F3F6;
}
.ecom-img {
  width: 100%;
  height: 100%;
}
.ecom-used-tag {
  position: absolute;
  left: 0;
  top: 16rpx;
  padding: 6rpx 14rpx;
  background: rgba(228, 100, 38, 0.92);
  color: #fff;
  font-size: 20rpx;
  border-radius: 0 16rpx 16rpx 0;
}
.ecom-body {
  padding: 18rpx 20rpx 20rpx;
}
.ecom-price {
  display: block;
  color: #E84C3D;
  font-size: 26rpx;
  font-weight: 700;
}
.ecom-price-num {
  font-size: 38rpx;
  font-weight: 800;
}
.ecom-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  margin-top: 8rpx;
  color: #17212B;
  font-size: 26rpx;
  font-weight: 600;
  line-height: 1.4;
  min-height: 73rpx;
}
.ecom-spec {
  display: block;
  margin-top: 6rpx;
  color: #98A2B3;
  font-size: 22rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ecom-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 14rpx;
}
.ecom-cat {
  padding: 4rpx 12rpx;
  background: #EAF3FD;
  color: #0A66C2;
  font-size: 20rpx;
  border-radius: 8rpx;
}
.ecom-views {
  color: #98A2B3;
  font-size: 20rpx;
}

/* ═══════ 就地搜索覆盖层（深蓝条 + 推荐/最近 + 结果卡片，.ov- 前缀避让页面同名类） ═══════ */
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
.ov-bar {
  background: #074D92;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 20rpx;
  /* paddingTop 由模板 JS 注入（与页面 .topbar 一致，env(safe-area-inset-top) 微信端返回 0） */
  flex-shrink: 0;
}
.ov-back {
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.ov-back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(45deg);
  margin-left: 10rpx;
}
.ov-search-box {
  flex: 1;
  height: 78rpx;
  border-radius: 12rpx;
  background: #fff;
  display: flex;
  align-items: center;
  gap: 14rpx;
  padding: 0 20rpx;
  min-width: 0;
}
.ov-search-icon {
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.ov-search-icon::after {
  content: '';
  position: absolute;
  right: -11rpx;
  bottom: -6rpx;
  width: 13rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #98A2B3;
  transform: rotate(45deg);
}
.ov-search-input { flex: 1; font-size: 26rpx; color: #17212B; }
.ov-clear { padding: 4rpx; }
.ov-clear-x { font-size: 30rpx; color: #98A2B3; line-height: 1; }
.ov-search-btn { color: #fff; font-size: 26rpx; font-weight: 600; white-space: nowrap; flex-shrink: 0; }

/* 结果区（可滚动） */
.ov-results {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.ov-result-head {
  margin: 32rpx 32rpx 8rpx;
  color: #667085;
  font-size: 24rpx;
}
.ov-card-list { padding: 0 32rpx 32rpx; display: flex; flex-direction: column; gap: 20rpx; }
.ov-trade-card {
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  overflow: hidden;
}
.ov-trade-card-main { display: flex; gap: 22rpx; padding: 24rpx; }
.ov-trade-visual {
  width: 164rpx;
  height: 164rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  background: #E8F2FC;
}
.ov-trade-body { flex: 1; min-width: 0; }
.ov-tag-row { display: flex; gap: 10rpx; align-items: center; margin-bottom: 12rpx; }
.ov-tag {
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.ov-tag-blue { color: #0A66C2; background: #EAF3FB; }
.ov-tag-orange { color: #DB5F0D; background: #FFF0E6; }
.ov-tag-green { color: #168A55; background: #E9F7F0; }
.ov-tag-gray { color: #667085; background: #F1F3F5; }
.ov-trade-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #17212B;
  font-size: 28rpx;
  line-height: 1.42;
  font-weight: 700;
}
.ov-trade-meta {
  display: flex;
  gap: 14rpx;
  margin-top: 14rpx;
  color: #667085;
  font-size: 22rpx;
  white-space: nowrap;
  overflow: hidden;
}
.ov-trade-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 80rpx;
  padding: 0 24rpx;
  border-top: 1px solid #EEF1F4;
}
.ov-price-block { display: flex; align-items: baseline; min-width: 0; }
.ov-price { color: #DF620F; font-size: 30rpx; font-weight: 750; }
.ov-price-unit { color: #667085; font-size: 20rpx; }
.ov-card-action { color: #0A66C2; font-size: 24rpx; font-weight: 650; white-space: nowrap; }

/* 空状态 */
.ov-state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.ov-state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.ov-state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.ov-state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }

/* 推荐 / 最近 */
.ov-suggest { padding-bottom: 40rpx; overflow-y: auto; }
.ov-search-block { padding: 36rpx 32rpx 0; }
.ov-block-title { display: block; font-size: 30rpx; font-weight: 700; color: #17212B; margin-bottom: 24rpx; }
.ov-keyword-row { display: flex; flex-wrap: wrap; gap: 16rpx; }
.ov-keyword {
  color: #344054;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10rpx;
  padding: 14rpx 20rpx;
  font-size: 24rpx;
}

@media (max-width: 380px) {
  .ov-trade-visual { width: 150rpx; height: 150rpx; }
}

@media (prefers-reduced-motion: reduce) {
  .stg, .stg-arr, .p-chip, .field-panel, .panel-mask {
    animation: none !important;
    transition: none !important;
  }
}
</style>
