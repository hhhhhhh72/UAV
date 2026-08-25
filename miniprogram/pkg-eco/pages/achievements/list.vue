<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- 白色刊头（navigationStyle: custom）：白底 + 衬线刊名（对齐生态服务页） -->
    <view class="mh" :style="{ paddingTop: sbh + 'px' }">
      <view class="mh-bar">
        <view class="mh-back" @tap="goBack"></view>
        <text class="mh-title">科技成果库</text>
        <view class="mh-side"></view>
      </view>
    </view>

    <!-- Search Bar -->
    <view class="sbar">
      <view class="sbox">
        <view class="sic"></view>
        <input class="sinp" v-model="q" placeholder="搜索成果名称、关键词" placeholder-class="ph" confirm-type="search" @input="onSearch" @confirm="applyFilter" />
        <text v-if="q" class="sclr" @tap="clearSearch">×</text>
        <view class="s-sep"></view>
        <text class="s-sbtn" @tap="onSearchTap">搜索</text>
      </view>

    </view>
    <view v-if="sV" class="mask" @tap="startCloseSort" />

    <!-- 家底叙事条：第一屏说清「这是什么、有多少」 -->
    <view class="asset-bar">
      <view class="asset-l">
        <text class="asset-t">协会 · 高校 · 企业技术家底</text>
        <text class="asset-d">覆盖整机 / 硬件 / 软件全链</text>
      </view>
      <view class="asset-nums">
        <view class="asset-num"><text class="an-b" :key="!loading && !err ? total : '-'">{{ !loading && !err ? total : '-' }}</text><text class="an-l">项成果</text></view>
        <view class="asset-num"><text class="an-b" :key="!loading && !err && allLoaded ? convCount : '-'">{{ !loading && !err && allLoaded ? convCount : '-' }}</text><text class="an-l">已转化</text></view>
        <view class="asset-num"><text class="an-b" :key="!loading && !err && allLoaded && srcCount > 0 ? srcCount : '-'">{{ !loading && !err && allLoaded && srcCount > 0 ? srcCount : '-' }}</text><text class="an-l">发布方</text></view>
      </view>
    </view>

    <!-- 阶段分段（一级筛选，TOC 注线式同场地预约）；吸顶 + 容器承载浮层面板 -->
    <!-- 筛选条用普通 view（同研发难题/课题攻关 .fbar，非 scroll-view）：5 个 tab 自然宽 ~660rpx < 750rpx 放得下；
         scroll-x 会接管横向手势 → 手指在 tab 上左右滑时页面被联动（参考页全部用普通 view 不晃），故移除 -->
    <view class="stage-wrap" :style="{ top: stickyTop + 'px' }">
      <view class="stages">
        <view
          v-for="s in STAGES"
          :key="s.k"
          class="stg"
          :class="{ on: stageKey === s.k }"
          @tap="pickStageTab(s.k)"
        >
          <text>{{ s.k === 'all' ? '全部' : s.l }}</text>
          <!-- ▾ 是独立的面板开关（方案 A）：未停在「全部」时点「全部」先清阶段，停在「全部」时再点才开面板；▾ 随时可开 -->
          <text v-if="s.k === 'all'" class="stg-arr" :class="{ up: panel === 'all' }" @tap.stop="togglePanel">▾</text>
        </view>
      </view>
      <!-- 领域面板：absolute 浮层（同研发难题面板），展开时不挤动内容 -->
      <view v-if="panel === 'all'" class="field-panel" :class="{ closing }">
        <view class="p-group">领域</view>
        <view class="p-chips">
          <text class="p-chip" :class="{ act: fField === '全部领域' }" @tap="pickField('全部领域')">全部领域</text>
          <text v-for="f in ALL_FIELDS" :key="f" class="p-chip" :class="{ act: fField === f }" @tap="pickField(f)">{{ f }}</text>
        </view>
        <view class="p-group">类型</view>
        <view class="p-chips">
          <text v-for="o in TYPE_OPTS" :key="o.v" class="p-chip" :class="{ act: isTypeAct(o) }" @tap="pickType(o.v)">{{ o.l }}</text>
        </view>
      </view>
    </view>
    <!-- 蒙层：从阶段分段底部开始置灰（同研发难题页），点击外部退场收起 -->
    <view v-if="panel" class="panel-mask" :style="{ top: maskTop + 'px' }" @tap="startClosePanel" />

    <!-- 信息行：计数 + 重置出口（筛选激活时）+ 排序弹层（右侧） -->
    <view class="ir">
      <view class="ir-left">
        <text v-if="!loading">已加载 <text class="irn">{{ shown }}</text> 项成果 · <text class="ir-conv">{{ convShown }}</text> 项已转化</text>
        <text v-if="hasActiveFilter" class="reset-chip" @tap="resetFilters">重置</text>
      </view>
      <view class="irs-wrap">
        <view class="irs" :class="{ on: sV }" @tap="toggleSort"><text class="irs-t">{{ sortLabel }}</text><view class="irs-arr"></view></view>
        <view v-if="sV" class="spop" :class="{ closing: sortClosing, up: sUp }" @tap.stop>
          <view
            v-for="o in sorts"
            :key="o.k"
            class="sp-opt"
            :class="{ active: sort === o.k }"
            @tap="pickSort(o.k)"
          >
            <text>{{ o.l }}</text>
            <text v-if="sort === o.k" class="sp-chk">✓</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Skeleton -->
    <view v-if="loading" class="cg">
      <view v-for="i in 6" :key="'sk' + i" class="card card-sk">
        <view class="sk-cv" />
        <view class="sk-bd">
          <view class="sk-l w90" />
          <view class="sk-l" />
          <view class="sk-l w60" />
        </view>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="err && !list.length" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchAll">重新加载</view>
      </u-empty>
    </view>

    <!-- Empty：区分搜索空 / 筛选空 / 真空（可诊断） -->
    <view v-else-if="!list.length" class="st">
      <u-empty :description="emptyTitle">
        <text class="sth">{{ emptyHint }}</text>
        <view class="stb" @tap="resetEverything">清除全部条件</view>
      </u-empty>
    </view>

    <!-- 双列网格卡：图位在上（图片终态，过渡期领域色），阶段/已转化信号叠图 -->
    <view v-else class="cg">
      <view v-for="x in list" :key="x.id" class="card" @tap="goDetail(x)">
        <view class="cover" :style="{ background: x.toneBg }">
          <image v-if="x.img && !imgFailed.has(x.id)" class="cov-img" :src="x.img" mode="aspectFill" @error="onImgError(x.id)" />
          <block v-else>
            <text class="cov-ic" :style="{ color: x.toneFg }">{{ x.ic }}</text>
            <text class="cov-name" :style="{ color: x.toneFg }">{{ x.f }}</text>
          </block>
          <text v-if="x.stage" class="stage-badge" :class="x.stageCls">{{ x.stageShort }}</text>
          <text v-if="x.transformed" class="conv-badge">✓ 已转化</text>
        </view>
        <view class="cbd">
          <text class="ct">{{ x.t }}</text>
          <view class="c-meta">
            <text v-if="x.src" class="c-tag c-tag-src">{{ x.src }}</text>
            <text class="c-tag c-tag-type">{{ x.tl }}</text>
          </view>
          <view class="cft">
            <text class="cf-date">{{ x.d }}</text>
            <text class="cf-stats">
              <text class="cf-star"></text>{{ x.s + (favSet.has(x.id) ? 1 : 0) }}
            </text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="filteredAll.length" class="lm">
      <text v-if="hasMoreRender">上拉加载更多</text>
      <text v-else-if="!allLoaded">已加载 {{ filteredAll.length }} / 共 {{ total }} 条</text>
      <text v-else>没有更多了</text>
    </view>
    <view v-if="mockMode" class="mock-note">当前为演示数据</view>

    <!-- 回顶按钮：长列表滚过一屏后出现 -->
    <view v-if="showBt && !panel" class="bt" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReady, onShow, onPageScroll, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { useReduceMotion } from '@/utils/motion'
import { request, BASE_URL } from '@/utils/request'
import { MOCK_ACHIEVEMENTS, ACH_TYPE_LABEL, ACH_STATUS_LABEL, STAGE_SHORT, STAGE_RANK, FIELD_TONE, TONE_DEFAULT, FIELD_ICON } from '@/utils/mockAchievements'

const PAGE_SIZE = 100
const MAX_PAGES = 10
const RENDER_STEP = 100 // DOM 渲染上限步长：数据全量入 fullList，渲染切片逐段揭示（同挑战页）

const q = ref('')
const sort = ref('latest')
const sV = ref(false)
const sUp = ref(false) // 排序弹层向上翻（触发点贴近视口底部时，防选项溢出屏幕）
const loading = ref(true)
const err = ref(false)
const list = ref([]) // 渲染切片：只渲染前 renderCap 项（防大列表卡顿）
const filteredAll = ref([]) // 过滤排序后的完整结果：「共 N 项」与筛选语义用完整集
const fullList = ref([])
const renderCap = ref(RENDER_STEP) // DOM 渲染上限：上拉触底 +100 逐段揭示（筛选仍作用于全量数据）
const total = ref(0)
const allLoaded = ref(true) // 服务端数据是否已全部拉取（false = 被 MAX_PAGES 截断，终态如实提示）
const mockMode = ref(false)
const sbh = ref(20) // 自定义导航：状态栏高度（真机按系统信息覆盖）
const maskTop = ref(200) // 蒙层起点（onReady 实测）：阶段分段底部，头部不蒙
const stickyTop = ref(108) // 阶段分段吸顶偏移：刊头（sbh + 88rpx）之下
const showBt = ref(false) // 回顶按钮：滚动超过 400px 出现
const favSet = ref(new Set()) // 收藏单键集合（fav_ach_set）
const imgFailed = ref(new Set()) // 图位图片加载失败集合（回退单字版式）
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
let nextPage = 1

// ---- 筛选状态 ----
const panel = ref('') // 'all' 时展开领域面板
const stageKey = ref('all') // 阶段分段：all/lab/pilot/industrialized/listed
const fType = ref('全部类型')
const fField = ref('全部领域')
const TYPE_OPTS = [
  { v: 'all', l: '全部类型' }, { v: 'patent', l: '发明专利' }, { v: 'utility', l: '实用新型' },
  { v: 'copyright', l: '软件著作' }, { v: 'paper', l: '论文成果' }, { v: 'standard', l: '技术标准' }, { v: 'design', l: '外观设计' },
]
const FIELD_HW = ['飞控系统', '动力系统', '载荷设备', '地面站', '遥感测绘']
const FIELD_SW = ['AI算法', '集群协同', '通信链路', '标准规范']

const sorts = [
  { k: 'latest', l: '最新发布' },
  { k: 'views', l: '最多浏览' },
  { k: 'favs', l: '最多收藏' }
]
const SORT_LABEL = { latest: '最新发布', views: '最多浏览', favs: '最多收藏' }

const ALL_FIELDS = [...FIELD_HW, ...FIELD_SW]
const isTypeAct = (o) => fType.value === o.l // fType 存 label（与 pickType 写入一致），'全部类型' 由 { v:'all', l:'全部类型' } 自然覆盖

// 阶段分段（TOC 注线式：文字 tab，无计数）
const STAGES = [
  { k: 'all', l: '全部' },
  { k: 'lab', l: '实验室' },
  { k: 'pilot', l: '中试' },
  { k: 'industrialized', l: '产业化' },
  { k: 'listed', l: '已上市' },
]
// 「全部」tab 恒为「全部」（阶段语义纯粹）：领域/类型选中回显交给面板 chip 高亮，不悬挂在阶段 tab 上

const shown = computed(() => list.value.length) // 真实可见数（渲染切片内的数量，诚实于所见）
const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新发布')
const hasMoreRender = computed(() => filteredAll.value.length > list.value.length) // 渲染侧：还有未渲染的项
// 家底条计数（全量已拉数据）
const convCount = computed(() => fullList.value.filter(x => x.transformed).length)
const srcCount = computed(() => new Set(fullList.value.map(x => x.src).filter(Boolean)).size)
const convShown = computed(() => filteredAll.value.filter(x => x.transformed).length)
// 空态三分：搜索空 / 筛选空 / 真空（可诊断，用户知道该改什么）
const hasFilters = computed(() => stageKey.value !== 'all' || fType.value !== '全部类型' || fField.value !== '全部领域')
// 非空态重置出口：任一筛选激活时显示（阶段/领域/类型）
const hasActiveFilter = computed(() => stageKey.value !== 'all' || fField.value !== '全部领域' || fType.value !== '全部类型')
// resetFilters：只清筛选（阶段/领域/类型），不动搜索词与排序——筛选面板场景的重置出口
const resetFilters = () => {
  stageKey.value = 'all'
  fType.value = '全部类型'
  fField.value = '全部领域'
  startClosePanel()
  applyFilter()
}
const emptyTitle = computed(() => {
  if ((q.value || '').trim()) return '没有找到相关成果'
  if (hasFilters.value) return '没有符合条件的成果'
  return '暂无成果'
})
const emptyHint = computed(() => {
  if ((q.value || '').trim()) return '换个关键词试试'
  if (hasFilters.value) return '试试调整筛选条件'
  return '试试调整筛选条件或搜索关键词'
})

const typeLabel = (t) => ACH_TYPE_LABEL[(t || '').toLowerCase()] || t || '其他'
const statusOf = (s) => {
  const key = (s || '').toLowerCase()
  const label = ACH_STATUS_LABEL[key]
  return label ? { label, cls: key } : null
}
const stageOf = (s) => {
  const key = (s || '').toLowerCase()
  const label = STAGE_SHORT[key]
  if (!label) return null
  return { label, short: label, key, rank: STAGE_RANK[key] || 0 }
}
const stageClsOf = (s) => {
  const key = (s || '').toLowerCase()
  if (key === 'lab' || key === 'laboratory') return 'cl-la' // 实验室：灰（未成熟）
  if (key === 'pilot') return 'cl-pi' // 中试：琥珀
  if (key === 'industrialized') return 'cl-in' // 产业化：绿
  return 'cl-li' // 已上市：深蓝
}

const imgSrc = (images) => {
  let arr = images
  if (typeof images === 'string') {
    try { arr = JSON.parse(images) } catch { return '' }
  }
  if (!Array.isArray(arr) || !arr.length) return ''
  const u = arr[0]
  return u ? (u.startsWith('http') ? u : BASE_URL + u) : ''
}

const mapItem = (it) => {
  const st = stageOf(it.stage)
  const tone = FIELD_TONE[it.field] || TONE_DEFAULT
  const sts = statusOf(it.status)
  return {
    id: it.id,
    t: it.title || '',
    f: it.field || '其他',
    ic: FIELD_ICON[it.field] || (it.field ? it.field.charAt(0) : '果'),
    toneBg: tone.bg, // 图位浅底（图片就绪后由图片替代）
    toneFg: tone.fg, // 单字 / 衬线刊名深字
    src: it.poster_name || '', // 发布主体（无则省略，不编造）
    tl: typeLabel(it.achieve_type),
    d: (it.created_at || '').slice(0, 10),
    v: it.views || 0,
    s: it.favs || 0,
    transformed: !!(sts && sts.cls === 'transformed'), // 已转化 = 被验证过（最高优先级信号）
    stage: st,
    stageShort: st ? st.short : '',
    stageCls: st ? stageClsOf(it.stage) : '',
    img: imgSrc(it.images),
  }
}
const onImgError = (id) => { imgFailed.value = new Set(imgFailed.value).add(id) } // 图位坏图 → 回退单字版式

const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter(x => (x.t + ' ' + x.f + ' ' + x.tl).toLowerCase().includes(kw))
  if (stageKey.value !== 'all') {
    const rank = STAGE_RANK[stageKey.value] || 0
    items = items.filter(x => x.stage && x.stage.rank === rank)
  }
  if (fType.value !== '全部类型') items = items.filter(x => x.tl === fType.value)
  if (fField.value !== '全部领域') items = items.filter(x => x.f === fField.value)
  if (sort.value === 'views') items.sort((a, b) => b.v - a.v)
  else if (sort.value === 'favs') items.sort((a, b) => (b.s + (favSet.value.has(b.id) ? 1 : 0)) - (a.s + (favSet.value.has(a.id) ? 1 : 0))) // 最多收藏排序随本地增量重排
  else items.sort((a, b) => ((b.d || '') < (a.d || '') ? 1 : (b.d || '') > (a.d || '') ? -1 : 0))
  filteredAll.value = items // 完整结果（计数/空态语义），DOM 只渲染前 renderCap 项防大列表卡顿
  list.value = items.slice(0, renderCap.value)
}

// ===== 数据获取 =====
// 接口替换点：GET /api/v1/achievements (field / page / page_size)；仅首次加载失败才回退 mock
const fetchAll = async (silent = false) => {
  loading.value = !silent // silent：下拉刷新保留当前列表，避免骨架屏顶替闪烁
  err.value = false
  nextPage = 1
  try {
    // 并行拉取全部页：冷启动墙 ≈ 1 个 RTT；单页失败不影响其余页
    const reqs = []
    for (let page = 1; page <= MAX_PAGES; page++) reqs.push(request({ url: '/api/v1/achievements', data: { page, page_size: PAGE_SIZE } }))
    const results = await Promise.allSettled(reqs)
    const acc = []
    let fetched = 0
    let anyOk = false
    for (const r of results) {
      if (r.status !== 'fulfilled') continue
      anyOk = true
      const res = r.value
      const items = Array.isArray(res) ? res : (res?.items || [])
      if (!items.length) continue
      acc.push(...items.map(mapItem))
      fetched = (res && res.total) || acc.length
      nextPage++
    }
    if (!anyOk) {
      // 从未成功加载过才回退演示数据；已有数据则保留——下拉刷新撞一次网络抖动不能顶替真实列表
      if (fullList.value.length === 0) useMock()
      else { uni.showToast({ title: '加载失败，已显示上次数据', icon: 'none' }); applyFilter() }
    } else {
      fullList.value = acc // 请求成功但为空 → 真实空态，不再被演示数据掩埋
      total.value = fetched
      mockMode.value = false
      allLoaded.value = acc.length >= fetched // 已拉到的 ≥ 服务器总数 = 数据齐了（否则被 MAX_PAGES 截断）
      applyFilter()
    }
  } catch {
    if (fullList.value.length === 0) useMock()
    else { uni.showToast({ title: '加载失败，已显示上次数据', icon: 'none' }); applyFilter() }
  } finally {
    loading.value = false
  }
}

const useMock = () => {
  // P2-6：生产环境绝不回退演示数据，失败即诚实错误态
  if (process.env.NODE_ENV === 'production') { err.value = true; applyFilter(); return }
  fullList.value = MOCK_ACHIEVEMENTS.map(mapItem)
  total.value = fullList.value.length
  mockMode.value = true
  allLoaded.value = true
  applyFilter()
}

const fetchMore = async () => {
  if (loading.value || nextPage > MAX_PAGES || mockMode.value) return
  try {
    const res = await request({ url: '/api/v1/achievements', data: { page: nextPage, page_size: PAGE_SIZE } })
    const items = Array.isArray(res) ? res : (res?.items || [])
    fullList.value = fullList.value.concat(items.map(mapItem))
    total.value = (res && res.total) || fullList.value.length
    nextPage++
    allLoaded.value = fullList.value.length >= total.value
    applyFilter()
  } catch {
    // 失败可见：不再静默表现为「滚动无响应」（触底重试仍会在下次触底时再次发起）
    uni.showToast({ title: '加载失败，请稍后重试', icon: 'none' })
  }
}

let searchTimer = null
const onSearch = () => { clearTimeout(searchTimer); searchTimer = setTimeout(applyFilter, 300) }
/* 点「搜索」：跳过防抖立即筛选（同研发难题 onSearchTap） */
const onSearchTap = () => { clearTimeout(searchTimer); applyFilter() }
const clearSearch = () => { q.value = ''; applyFilter() }

// ---- 浮层开合（退场必须存在：先加 closing 类播退场动画，定时器到点再 v-if 移除；可打断） ----
const closing = ref(false) // 筛选面板退场中
let closeT = null
const PANEL_CLOSE_MS = 210 // 退场动画 .21s ease-in（= 进场 ×0.7）
const startClosePanel = () => {
  if (closing.value) return // 已在退场中，防重复触发叠加定时器
  closing.value = true
  clearTimeout(closeT)
  closeT = setTimeout(() => { panel.value = ''; closing.value = false; closeT = null }, PANEL_CLOSE_MS)
}
const sortClosing = ref(false) // 排序弹层退场中
let sortT = null
const SORT_CLOSE_MS = 150 // 退场动画 .15s ease-in（= 进场 ×0.7）
const startCloseSort = () => {
  if (sortClosing.value) return
  sortClosing.value = true
  clearTimeout(sortT)
  sortT = setTimeout(() => { sV.value = false; sortClosing.value = false; sortT = null }, SORT_CLOSE_MS)
}
const toggleSort = () => {
  if (sV.value) { startCloseSort(); return }
  // 打开排序弹层：筛选面板直接收起（不播动画，避免同帧两个浮层同时动）
  clearTimeout(closeT); closeT = null; closing.value = false; panel.value = ''
  sV.value = true; sortClosing.value = false; sUp.value = false
  // 触底翻转：弹层只向下弹，贴近视口底部时改向上（短列表/小屏防溢出）
  uni.nextTick(() => {
    uni.createSelectorQuery().select('.spop').boundingClientRect((rect) => {
      if (!rect) return
      let sys = null
      try { sys = uni.getSystemInfoSync() } catch (e) { /* 取不到系统信息则保持向下 */ }
      const winH = (sys && sys.windowHeight) || 600
      const safeB = (sys && sys.safeAreaInsets && sys.safeAreaInsets.bottom) || 0
      if (rect.bottom > winH - safeB - 8) sUp.value = true
    }).exec()
  })
}
const pickSort = (k) => { sort.value = k; startCloseSort(); applyFilter() }

// ---- 筛选交互（阶段分段 + 「全部」领域面板） ----
const measureMaskTop = () => {
  // 蒙层起点 = 阶段分段容器底部（同研发难题页：头部不蒙）；面板打开时实测，内容自适应不错位
  uni.createSelectorQuery().select('.stage-wrap').boundingClientRect((rect) => {
    if (rect && rect.bottom) maskTop.value = Math.round(rect.bottom)
  }).exec()
}
const togglePanel = () => {
  if (panel.value === 'all') { startClosePanel(); return } // 再点「全部」→ 退场收起
  // 打开面板：取消未完成的退场；排序弹层直接收起，避免同帧两个浮层同时动
  clearTimeout(closeT); closeT = null; closing.value = false
  clearTimeout(sortT); sortT = null; sortClosing.value = false; sV.value = false
  panel.value = 'all'
  measureMaskTop() // 实测蒙层起点（asset-bar/stages 高度自适应）
}
// 方案 A（2026-08，用户确认分阶行为）：
// - 非全部阶段 tab：再点激活项取消（同面板 chip 规则）
// - 「全部」tab：未停在「全部」→ 先清阶段筛选；已停在「全部」→ 再点 = 展开/收起领域面板
// - ▾ 箭头始终可独立开面板（不影响已选阶段）
const pickStageTab = (k) => {
  if (k !== 'all') {
    startClosePanel()
    stageKey.value = stageKey.value === k ? 'all' : k
    applyFilter()
    return
  }
  if (stageKey.value !== 'all') {
    startClosePanel()
    stageKey.value = 'all'
    applyFilter()
    return
  }
  togglePanel()
}
// 面板 chip 点选即筛、不关面板：chip 蓝底高亮即时可见（选中反馈），再点一次取消；点外部/「全部」/▾ 收起
const pickType = (v) => {
  const o = TYPE_OPTS.find(x => x.v === v)
  if (!o) return
  fType.value = fType.value === o.l ? '全部类型' : o.l
  applyFilter()
}
const pickField = (f) => {
  fField.value = fField.value === f ? '全部领域' : f // 阶段×领域可并行，各自再点取消
  sV.value = false
  applyFilter()
}
// resetEverything：清全部（搜索词 + 筛选 + 排序）——空态出口；与 resetFilters（只清筛选）范围互补
const resetEverything = () => {
  q.value = ''
  stageKey.value = 'all'
  fType.value = '全部类型'
  fField.value = '全部领域'
  sort.value = 'latest'
  sV.value = false
  startClosePanel()
  applyFilter()
}

const goDetail = (x) => {
  uni.navigateTo({ url: '/pkg-eco/pages/achievements/detail?id=' + encodeURIComponent(x.id) })
}

const goBack = () => uni.navigateBack()

// ===== 收藏统一：本地单键增量过渡 =====
// 接口替换点：GET /api/v1/favorites/mine 落地后，favSet 以接口返回为准（《科技成果库-后端改动清单》§1）；
// 收藏时 POST /api/v1/favorites/{achievement_id}、取消 DELETE；列表/详情增量改由后端计数后删除本段
const loadFavs = () => {
  try {
    const raw = uni.getStorageSync('fav_ach_set')
    favSet.value = new Set(Array.isArray(raw) ? raw : [])
  } catch (e) { favSet.value = new Set() }
}

onLoad(() => {
  const sys = uni.getSystemInfoSync()
  sbh.value = sys.statusBarHeight || 20 // 自定义导航：状态栏高度（真机按系统信息覆盖）
  const rpx = sys.windowWidth / 750
  stickyTop.value = Math.ceil(sbh.value + 88 * rpx) // 阶段分段吸顶：刊头之下
  // 蒙层起点兜底估算（onReady 后实测修正）：刊头 + 搜索框 + 家底条 + 阶段分段
  maskTop.value = Math.ceil(sbh.value + (88 + 120) * rpx + 132 * rpx)
  checkMotion() // 减弱动效检测（无障碍）
  loadFavs()
  fetchAll()
})
onReady(() => { measureMaskTop() }) // 实测蒙层起点（asset-bar/stages 高度自适应，不用写死）
onPageScroll((e) => {
  showBt.value = e.scrollTop > 400 // 回顶按钮：滚过一屏后出现
  if (sV.value) startCloseSort() // 排序弹层挂滚动内容里，滚动即关（防隐形蒙层滞留成死屏）
})
const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })
onShow(() => {
  loadFavs() // 从详情页返回：收藏增量同步
  applyFilter() // 「最多收藏」排序随增量重排
})
onPullDownRefresh(async () => {
  await fetchAll(true) // silent：保留当前列表，避免骨架屏顶替闪烁（原生下拉动画本身已提供刷新反馈）
  uni.stopPullDownRefresh()
})
onReachBottom(() => {
  // 触底优先揭示已拉取数据（渲染切片 +100），渲染完再拉取下一页
  if (hasMoreRender.value) { renderCap.value += RENDER_STEP; applyFilter(); return }
  fetchMore()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff; /* 对齐研发难题页：纯白页面底 */
  padding-bottom: 80rpx;
  overflow-x: hidden; /* 保险：杜绝任何残余横向溢出传导为页面级左右滚动（参考页结构无此必要，本页有 grid 双列 + 历史 scroll-view） */
}

/* ===== 白色刊头（navigationStyle: custom）：白底 + 衬线刊名（对齐生态服务页） ===== */
.mh { background: #fff; }
.mh-bar { position: relative; display: flex; align-items: center; justify-content: center; height: 88rpx; }
.mh-back { position: absolute; left: 8rpx; top: 0; width: 88rpx; height: 88rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%2317212b' stroke-width='2.2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M15 4l-8 8 8 8'/%3E%3C/svg%3E"); background-size: 44rpx; background-repeat: no-repeat; background-position: center; transition: opacity .2s ease; }
.mh-back:active { opacity: .6; }
.mh-title { font-size: 32rpx; font-weight: 600; color: #17212B; } /* 对齐 u-nav-bar 标题（系统字体 32rpx/600） */
.mh-side { position: absolute; right: 0; width: 88rpx; height: 88rpx; }

/* ===== 搜索（白上白：纯白填充 + 灰描边 + 双层投影浮起；宽松热区 88rpx） ===== */
.sbar {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 24rpx 28rpx 16rpx;
}
.sbox {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 14rpx;
  height: 88rpx;
  background: #fff;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  padding: 0 22rpx;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05); /* 白上白浮起（同挑战页 .b-search） */
}
.sic { width: 30rpx; height: 30rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23667085' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='11' cy='11' r='7'/%3E%3Cpath d='M21 21l-4.35-4.35'/%3E%3C/svg%3E"); background-size: contain; background-repeat: no-repeat; flex: none; } /* 图标描边对齐组 Meta Grey（原 #969799 视觉沉底） */
.sinp { flex: 1; font-size: 26rpx; color: var(--color-text); }
.ph { color: #667085; } /* 对齐组 Meta Grey：原占位灰 #c8c9cc 白底 ≈1.9:1 不可见 */
.sclr { font-size: 30rpx; color: #667085; padding: 20rpx 24rpx; margin: -20rpx -24rpx; transition: transform .2s ease, opacity .2s ease; }
.sclr:active { transform: scale(.9); opacity: .7; }
/* 「搜索」按钮 + 左侧细竖杠分隔（同研发难题 .b-sep/.b-sbtn：竖线 2×30rpx，右距 18 左距 12） */
.s-sep { width: 2rpx; height: 30rpx; background: #DDE1E6; margin: 0 18rpx 0 12rpx; flex: none; }
.s-sbtn { flex: none; color: #344054; font-size: 26rpx; line-height: 1; padding: 12rpx 4rpx 12rpx 0; }
.s-sbtn:active { opacity: .5; }
.irs-wrap { position: relative; z-index: 60; }
.spop {
  position: absolute; top: calc(100% + 8rpx); right: 0; z-index: 50;
  background: #fff; border-radius: 16rpx;
  box-shadow: 0 8rpx 32rpx rgba(0,0,0,.08); padding: 12rpx; min-width: 320rpx;
  animation: popIn .2s cubic-bezier(.32, .72, 0, 1); /* ios-decel：下拉流体减速（同挑战页排序弹层） */
}
@keyframes popIn { from { opacity: 0; transform: translateY(-8rpx); } to { opacity: 1; transform: translateY(0); } }
.spop.closing { animation: popOut .15s ease-in forwards; } /* 退场 = 进场 ×0.7 */
@keyframes popOut { from { opacity: 1; transform: translateY(0); } to { opacity: 0; transform: translateY(-8rpx); } }
.spop.up { top: auto; bottom: calc(100% + 8rpx); animation-name: popInUp; } /* 触底翻转：向上弹 */
.spop.up.closing { animation: popOutUp .15s ease-in forwards; }
@keyframes popInUp { from { opacity: 0; transform: translateY(8rpx); } to { opacity: 1; transform: translateY(0); } }
@keyframes popOutUp { from { opacity: 1; transform: translateY(0); } to { opacity: 0; transform: translateY(8rpx); } }
.sp-opt {
  padding: 26rpx 36rpx; font-size: 26rpx; color: var(--color-text); /* 28→26rpx：同研发难题排序选项 13px */
  border-radius: 16rpx; display: flex; align-items: center; justify-content: space-between;
  gap: 16rpx; white-space: nowrap;
}
.sp-opt.active { color: #074D92; font-weight: 600; background: #EAF3FB; } /* 激活项浅底（同挑战页排序弹层）；文字用 AA 暗变体 */
.sp-opt:active { background: var(--color-bg); }
.sp-chk { font-size: 24rpx; }
.mask { position: fixed; inset: 0; z-index: 43; background: transparent; } /* 同 panel-mask(43)：覆盖吸顶 tab，点外即关 */

/* ===== 家底叙事条：深蓝渐变，第一屏说清「这是什么、有多少」 ===== */
.asset-bar {
  margin: 2rpx 24rpx 16rpx;
  background: linear-gradient(135deg, #0a3a6b, #074d92);
  border-radius: 24rpx;
  padding: 24rpx 28rpx;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.asset-l { display: flex; flex-direction: column; gap: 6rpx; }
.asset-t { font-size: 24rpx; font-weight: 600; letter-spacing: 2rpx; } /* 字阶阶梯内 */
.asset-d { font-size: 20rpx; color: rgba(255,255,255,.82); }
.asset-nums { display: flex; gap: 24rpx; }
.asset-num { text-align: center; }
.an-b { display: block; font-size: 36rpx; font-weight: 800; line-height: 1.2; }
.an-l { font-size: 20rpx; color: rgba(255,255,255,.75); }

/* ===== 阶段分段（TOC 注线式，同场地预约）：文字 tab + 底部注线，无框无计数；吸顶 ===== */
.stage-wrap { position: sticky; z-index: 42; background: #fff; } /* top 由 stickyTop 动态（刊头之下）；白底遮住滚动内容 */
/* 筛选条普通 view（同研发难题 .fbar，scroll-view 已移除）：padding 原在 .stages-scroll；
   5 个 tab 自然宽 ~660rpx < 750rpx 永不换行；flex 而非 inline-flex——子项 flex-shrink:0 保底 */
.stages { display: flex; gap: 40rpx; padding: 4rpx 28rpx 16rpx; white-space: nowrap; }
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 0 8rpx;
  font-size: 24rpx; /* 筛选器字体同研发难题 12px */
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; } /* 激活态用 AA 暗变体（#0A66C2 白底 ≈4.5:1 边缘，深档 ≈6.9:1） */
.stg.on::after { content: ''; position: absolute; left: 8rpx; right: 8rpx; bottom: 16rpx; height: 3rpx; border-radius: 2rpx; background: #074D92; animation: toc-in .22s ease-out; }
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.stg-arr { font-size: 24rpx; color: #667085; transition: transform .2s ease, color .2s ease; padding: 20rpx 16rpx; margin: -20rpx -16rpx; } /* 独立热区（方案 A 面板开关）；负 margin 抵消位移；箭头同筛选器 12px */
.stg-arr.up { transform: rotate(180deg); color: #074D92; } /* 面板展开：朝上 + AA 暗变体 */

/* ===== 领域面板：absolute 浮层（同研发难题面板），展开时不挤动下方内容 ===== */
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
  animation: panelIn .3s cubic-bezier(.32, .72, 0, 1); /* 浮层档：进场 ios-decel（同挑战页） */
}
.field-panel.closing { animation: panelOut .21s ease-in forwards; } /* 退场 = 进场 ×0.7，forwards 保持到 v-if 移除防闪跳 */
@keyframes panelOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-10px); }
}
@keyframes panelIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
/* 面板内部（与研发难题/课题攻关的筛选面板同款） */
.field-panel .p-group { font-size: 13px; font-weight: 700; color: #344054; margin: 12px 0 6px; }
.field-panel .p-group:first-child { margin-top: 0; }
.p-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.p-chip {
  min-height: 40px; /* 触控目标：36px→40px（移动端触达下限，同挑战页进阶） */
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
}
.p-chip.act { color: #fff; border-color: #074D92; background: #074D92; font-weight: 600; } /* 激活底用 AA 暗变体：白字对比度 4.45→6.9:1 */
.p-chip { transition: background .2s ease, border-color .2s ease, color .2s ease, transform .3s cubic-bezier(.34, 1.8, .64, 1); } /* ios-pop：松手弹簧回位 */
.p-chip:active { transform: scale(.94); transition: transform .08s linear; } /* 按下即时到位 */
.p-chip.act { animation: chipPop .3s cubic-bezier(.34, 1.8, .64, 1); } /* ios-pop：选中微弹带轻微过冲回位 */
@keyframes chipPop { 0% { transform: scale(1); } 40% { transform: scale(.94); } 100% { transform: scale(1); } }

/* ===== 蒙层：从阶段分段底部开始置灰（top 由 maskTop 实测），低于面板(43) ===== */
.panel-mask {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 41;
  background: rgba(16, 24, 40, 0.2); /* 真变暗：面板展开时置灰下方内容（同挑战页） */
  animation: maskIn .22s ease-out; /* 遮罩与面板同步淡入 */
}
@keyframes maskIn { from { opacity: 0; } to { opacity: 1; } }

/* ===== 信息行：计数 + 重置出口 + 排序（右侧） ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 4rpx 32rpx 16rpx; font-size: 24rpx; color: #667085; } /* 对齐组 Meta Grey（白底 AA）；面板展开时信息行被面板(43)自然盖住，重置/排序入口随面板收起后露出 */
.ir-left { display: flex; align-items: center; gap: 16rpx; min-width: 0; }
.irn { color: var(--color-primary); font-weight: 600; }
.ir-conv { color: #1F7A48; font-weight: 600; }
/* 重置出口：任一筛选激活时出现 */
.reset-chip {
  flex: none;
  color: #667085;
  font-size: 24rpx;
  padding: 12rpx 8rpx;
  animation: chipIn .22s cubic-bezier(.34, 1.8, .64, 1) backwards;
}
.reset-chip:active { opacity: .7; }
@keyframes chipIn { from { transform: scale(.85); } to { transform: scale(1); } }
.irs { color: var(--color-primary); font-weight: 500; display: flex; align-items: center; gap: 8rpx; padding: 16rpx 8rpx 16rpx 24rpx; margin: -16rpx -8rpx -16rpx -24rpx; transition: transform .2s ease, background .2s ease; } /* 热区扩大（负 margin 抵消视觉位移）：对齐挑战页 8px 纵向热区 */
.irs:active { transform: scale(.96); }
.irs-arr { width: 14rpx; height: 14rpx; border-right: 3rpx solid #0A66C2; border-bottom: 3rpx solid #0A66C2; transform: rotate(45deg); margin-bottom: 6rpx; transition: transform .2s ease; }
.irs.on .irs-arr { transform: rotate(225deg); } /* 展开朝上（同筛选 tab ▾→▴ 语义） */

/* ===== 双列网格卡：图位在上 + 情报区（转化看板形态） ===== */
.cg { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; padding: 0 28rpx 40rpx; }
.card { box-sizing: border-box; background: #fff; border-radius: 20rpx; overflow: hidden; border: 2rpx solid #E4E7EC; display: flex; flex-direction: column; box-shadow: 0 2rpx 6rpx rgba(10,30,60,.04), 0 12rpx 32rpx rgba(10,30,60,.05); transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease; } /* 白上白 + 蓝调双层阴影（contact + ambient，≤5%）；box-sizing: border-box——border 收进 grid 轨道（content-box 下 2rpx×2 溢出轨道 4rpx/卡 → 页面横向可滚动） */
.card:active { transform: scale(.97); opacity: .88; }
/* 图位：4/3，领域色浅底（图片终态铺满）；阶段徽章左上、已转化角标右上 */
.cover { position: relative; aspect-ratio: 4/3; min-height: 240rpx; display: flex; align-items: center; justify-content: center; } /* min-height：老 WebView 无 aspect-ratio 时的高度兜底（≈图位高），防图片加载后卡片高度跳变 */
.cov-img { position: absolute; left: 0; top: 0; width: 100%; height: 100%; display: block; } /* 图片脱离布局：异步加载完成不改变卡片高度（消除上下晃动） */
.cov-ic { font-size: 40rpx; font-weight: 800; } /* 图位单字：对齐 testsites 图块 40rpx 先例 */
/* 领域刊名（衬线微刻字 20rpx，DESIGN.md 封面 15-22rpx 例外） */
.cov-name { position: absolute; left: 12rpx; bottom: 10rpx; font-family: Georgia, "Songti SC", "STSong", SimSun, serif; font-size: 20rpx; letter-spacing: 2rpx; }
/* 阶段徽章（图位左上，第一视觉位）：AA 暗变体 */
.stage-badge { position: absolute; left: 12rpx; top: 12rpx; font-size: 24rpx; font-weight: 700; padding: 4rpx 14rpx; border-radius: 8rpx; white-space: nowrap; } /* 徽章 12px：同研发难题卡徽章基准 */
.stage-badge.cl-la { background: #EEF1F4; color: #5D6B82; } /* 实验室：灰（未成熟） */
.stage-badge.cl-pi { background: #FFF4E5; color: #B45309; } /* 中试：琥珀 */
.stage-badge.cl-in { background: #E9F7F0; color: #1F7A48; } /* 产业化：绿 */
.stage-badge.cl-li { background: #EAF3FB; color: #074D92; } /* 已上市：深蓝 */
/* 已转化角标（图位右上）：被验证过 = 最高优先级信号 */
.conv-badge { position: absolute; right: 12rpx; top: 12rpx; font-size: 24rpx; font-weight: 700; padding: 4rpx 14rpx; border-radius: 8rpx; background: #E9F7F0; color: #0B6B41; } /* 角标 12px：同徽章基准 */
.cbd { padding: 16rpx 20rpx 20rpx; display: flex; flex-direction: column; gap: 10rpx; }
.ct { font-size: 30rpx; font-weight: 700; color: #17212B; line-height: 1.45; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; } /* 卡片标题对齐挑战页 15px/700（DESIGN.md 对齐组 30rpx 例外） */
.c-meta { display: flex; align-items: center; gap: 10rpx; white-space: nowrap; overflow: hidden; }
.c-tag { flex: none; font-size: 24rpx; padding: 4rpx 12rpx; border-radius: 8rpx; font-weight: 700; } /* 字重 700：同研发难题徽章基准 */
.c-tag-src { color: #074D92; background: #EAF3FB; max-width: 50%; overflow: hidden; text-overflow: ellipsis; } /* 发布主体：可信度信号；长机构名限宽省略，类型 tag 恒可见 */
.c-tag-type { color: #344054; background: #EEF1F4; }
.cft { font-size: 24rpx; color: #667085; display: flex; justify-content: space-between; align-items: center; } /* 信息行 12px：同研发难题 c-meta */
.cf-date { color: #667085; }
.cf-stats { color: #667085; font-weight: 600; }
.cf-star { display: inline-block; width: 24rpx; height: 24rpx; background-repeat: no-repeat; background-size: contain; vertical-align: -2rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23667085'%3E%3Cpath d='M12 2.5l2.9 6.1 6.6.9-4.8 4.6 1.2 6.6L12 17.6l-5.9 3.1 1.2-6.6-4.8-4.6 6.6-.9z'/%3E%3C/svg%3E"); margin-right: 6rpx; }

/* ===== Skeleton（保留） ===== */
.card-sk .sk-cv { aspect-ratio: 4/3; min-height: 240rpx; background: #EDF0F3; animation: shimmer 1.5s ease; } /* min-height 同 .cover：骨架与真实卡同高，切换无跳动 */
.sk-bd { padding: 16rpx 20rpx; }
.sk-l { height: 24rpx; background: #EDF0F3; border-radius: 8rpx; margin-bottom: 12rpx; animation: shimmer 1.5s ease; }
.sk-l.w90 { width: 90%; }
.sk-l.w60 { width: 60%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== State（保留） ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 120rpx 40rpx; }
.sth { font-size: 24rpx; color: #667085; margin: 24rpx 0; display: block; }
.stb { display: inline-block; padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; box-shadow: 0 4rpx 14rpx rgba(10,102,194,.25), inset 0 2rpx 0 rgba(255,255,255,.2), inset 0 -4rpx 10rpx rgba(7,77,146,.18); transition: transform .2s ease, opacity .2s ease; } /* 品牌蓝晕 + 内高光/厚度 gloss（对齐组）；按钮 13px 同研发难题 .stb */
.stb:active { transform: scale(.95); opacity: .85; }

/* ===== Load More（保留） ===== */
.lm { text-align: center; padding: 24rpx; font-size: 24rpx; color: #667085; }

/* ===== 演示提示 ===== */
.mock-note { text-align: center; padding: 0 0 24rpx; font-size: 20rpx; color: #667085; }

/* ===== 回顶按钮（长列表断粮救济）：白底圆钮，右下 fixed ===== */
.bt {
  position: fixed;
  right: 32rpx;
  bottom: 120rpx;
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: #fff;
  border: 1rpx solid #E4E7EC;
  box-shadow: 0 4rpx 16rpx rgba(16,24,40,.12);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40rpx; /* 回顶箭头 20px：同研发难题 .bt */
  color: #0A66C2;
  z-index: 50;
  animation: btIn .2s ease-out;
}
.bt:active { transform: scale(.92); }
@keyframes btIn { from { opacity: 0; transform: translateY(12rpx); } to { opacity: 1; transform: translateY(0); } }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .cg .card { animation: none; } /* 卡片错峰入场全关 */
.page.no-motion .cg .card .stage-badge,
.page.no-motion .cg .card .conv-badge { animation: none; } /* 图位信号落位属位移，关闭 */
.page.no-motion .an-b { animation: none; } /* 家底条数字落定属位移，关闭 */
.page.no-motion .asset-bar { animation: none; }
.page.no-motion .field-panel { animation: panelFadeIn .22s ease-out; } /* 面板降级为纯淡入 */
.page.no-motion .field-panel.closing { animation: panelFadeOut .16s ease-in forwards; }
.page.no-motion .spop { animation: spopFadeIn .2s ease-out; }
.page.no-motion .spop.closing { animation: spopFadeOut .15s ease-in forwards; }
.page.no-motion .spop.up { animation: spopFadeIn .2s ease-out; }
.page.no-motion .spop.up.closing { animation: spopFadeOut .15s ease-in forwards; }
.page.no-motion .panel-mask { animation: maskIn .22s ease-out; }
.page.no-motion .p-chip.act { animation: none; } /* 选中微弹属缩放，关闭；选中色保留 */
.page.no-motion .reset-chip { animation: none; }
.page.no-motion .bt { animation: none; }
.page.no-motion .stg.on::after { animation: none; } /* 注线画出属位移，关闭 */
.page.no-motion .p-chip:active,
.page.no-motion .irs:active,
.page.no-motion .stb:active,
.page.no-motion .bt:active,
.page.no-motion .card:active,
.page.no-motion .stg:active { transform: none; } /* 按压微缩放关闭，保留颜色/透明度反馈 */
@keyframes panelFadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes panelFadeOut { from { opacity: 1; } to { opacity: 0; } }
@keyframes spopFadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes spopFadeOut { from { opacity: 1; } to { opacity: 0; } }

/* ===================== 入场动画：前 6 项卡片依次淡入 + 轻微上移（20ms 错开 ≤6 项、总时长 ≤400ms；backwards 不阻塞点击态）；切片揭示的新卡不重播 ===================== */
.cg .card { animation: none; }
.cg .card:nth-child(-n+6) { animation: uiCardIn .3s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.cg .card:nth-child(1) { animation-delay: 0ms; }
.cg .card:nth-child(2) { animation-delay: 20ms; }
.cg .card:nth-child(3) { animation-delay: 40ms; }
.cg .card:nth-child(4) { animation-delay: 60ms; }
.cg .card:nth-child(5) { animation-delay: 80ms; }
.cg .card:nth-child(6) { animation-delay: 100ms; }
@keyframes uiCardIn { from { opacity: 0; transform: translateY(12rpx); } to { opacity: 1; transform: translateY(0); } }

/* ===================== delight：图位信号落位（阶段徽章 + 已转化角标随卡轻轻落下，比卡晚一步「信号落定」） ===================== */
/* 卡片升起（+12rpx 上移浮现）、信号落下（-10rpx 落位），形成「卡升起、信号落定」的先后层次；一次、≤400ms、ios-decel */
.cg .card:nth-child(-n+6) .stage-badge,
.cg .card:nth-child(-n+6) .conv-badge { animation: badgeLand .24s cubic-bezier(.32, .72, 0, 1) backwards; }
.cg .card:nth-child(1) .stage-badge, .cg .card:nth-child(1) .conv-badge { animation-delay: 30ms; }
.cg .card:nth-child(2) .stage-badge, .cg .card:nth-child(2) .conv-badge { animation-delay: 50ms; }
.cg .card:nth-child(3) .stage-badge, .cg .card:nth-child(3) .conv-badge { animation-delay: 70ms; }
.cg .card:nth-child(4) .stage-badge, .cg .card:nth-child(4) .conv-badge { animation-delay: 90ms; }
.cg .card:nth-child(5) .stage-badge, .cg .card:nth-child(5) .conv-badge { animation-delay: 110ms; }
.cg .card:nth-child(6) .stage-badge, .cg .card:nth-child(6) .conv-badge { animation-delay: 130ms; }
@keyframes badgeLand { from { opacity: 0; transform: translateY(-10rpx); } to { opacity: 1; transform: translateY(0); } }

/* ===================== delight：家底条数字落定（'-' 变为真实数字时亮相，一次） ===================== */
.an-b { animation: anIn .3s ease-out backwards; }
@keyframes anIn { from { opacity: 0; transform: translateY(6rpx); } to { opacity: 1; transform: translateY(0); } }
</style>
