<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }" @tap="closeAll">
    <u-nav-bar title="研发难题广场" show-back :fixed="true" @back="goBack" />

    <!-- 固定头部：搜索 + 筛选器 + 展开面板（一体吸顶） -->
    <view class="sticky-head" :style="{ top: (statusBarHeight + 44) + 'px' }" @tap.stop>

      <!-- 搜索框（保留上一版样式） -->
      <view class="sbar">
        <view class="b-search">
          <image class="b-search-ic" src="/static/home/icons/search.svg" mode="aspectFit" />
          <input class="b-sinp" v-model="q" placeholder="搜索难题、关键词" placeholder-class="b-ph" confirm-type="search" @input="onSearch" />
          <text v-if="q" class="b-sclr" aria-role="button" aria-label="清除搜索" @tap="clearSearch">×</text>
          <view class="b-sep"></view>
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- 筛选器（搜索框下、banner 上） -->
      <view class="fbar">
        <view class="fpill" :class="{ on: panel === 'comp' }" @tap="togglePanel('comp')">
          <text class="fpv">{{ compLabel }}</text><text class="farr">▾</text>
        </view>
        <view class="fpill" :class="{ on: panel === 'time' }" @tap="togglePanel('time')">
          <text class="fpv">{{ timeLabel }}</text><text class="farr">▾</text>
        </view>
        <view class="fpill" :class="{ on: panel === 'money' }" @tap="togglePanel('money')">
          <text class="fpv">{{ moneyLabel }}</text><text class="farr">▾</text>
        </view>
        <view v-if="hasActiveFilters" class="freset" @tap="resetAll">重置</view>
      </view>

      <!-- 展开面板：从筛选器下方平滑展开、全宽、与搜索/筛选一体 -->
      <view v-if="panel" class="panel-wrap" :class="{ closing }">
        <!-- 领域状态 -->
        <view v-if="panel === 'comp'" class="panel">
          <view class="p-head">领域状态 · 领域 / 状态</view>
          <view class="p-group">领域（一级）</view>
          <view class="p-chips">
            <text v-for="f in FIELD_OPTS" :key="f.v" class="p-chip" :class="{ act: selField === f.v }" @tap="toggleComp('field', f.v)">{{ f.l }}</text>
          </view>
          <view class="p-group">状态（二级）</view>
          <view class="p-chips">
            <text v-for="s in STATUS_OPTS" :key="s.v" class="p-chip" :class="{ act: selStatus === s.v }" @tap="toggleComp('status', s.v)">{{ s.l }}</text>
          </view>
        </view>

        <!-- 时间 -->
        <view v-else-if="panel === 'time'" class="panel">
          <view class="p-head">请选择发布时间</view>
          <view class="p-chips">
            <text v-for="o in QUICK_OPTS" :key="o.v" class="p-chip" :class="{ act: quick === o.v }" @tap="pickQuick(o.v)">{{ o.l }}</text>
          </view>
          <view class="cal" @touchstart="calTouchStart" @touchend="calTouchEnd">
            <view class="cal-head">
              <view class="cal-nav" aria-role="button" aria-label="上一个月" @tap="calShift(-1)">‹</view>
              <text class="cal-title">{{ calYear }}年{{ calMonth }}月</text>
              <view class="cal-nav" aria-role="button" aria-label="下一个月" @tap="calShift(1)">›</view>
            </view>
            <view class="cal-week">
              <text v-for="w in WEEK" :key="w">{{ w }}</text>
            </view>
            <view class="cal-cells">
              <view v-for="(c, i) in calCells" :key="i" class="cal-cell" :class="c.cls" @tap="onCalClick(c.key)">{{ c.day }}</view>
            </view>
            <view class="cal-foot">
              <text class="cal-tip">{{ calTip }}</text>
              <view class="cal-reset" @tap="resetTime">清除</view>
            </view>
          </view>
        </view>

        <!-- 金钱 -->
        <view v-else-if="panel === 'money'" class="panel">
          <view class="p-head">请选择悬赏金额（万元）</view>
          <view class="p-chips">
            <text v-for="o in MONEY_OPTS" :key="o.v" class="p-chip" :class="{ act: mPreset === o.v }" @tap="pickPreset(o.v)">{{ o.l }}</text>
            <text class="p-chip" :class="{ act: mFace }" @tap="toggleFace">面议</text>
          </view>
          <view class="slider" :class="{ disabled: mFace }">
            <view class="slider-track"></view>
            <view class="slider-range" :style="rangeStyle"></view>
            <view class="slider-bar" @touchstart.stop.prevent="onTouchStart" @touchmove.stop.prevent="onTouchMove" @touchend.stop.prevent="onTouchEnd" @touchcancel.stop.prevent="onTouchEnd">
              <view class="thumb thumb-min" :style="{ left: mMinPx + 'px' }"></view>
              <view class="thumb thumb-max" :style="{ left: mMaxPx + 'px' }"></view>
            </view>
          </view>
          <view class="slider-labels">
            <text>{{ Math.round(mMin) }}万</text>
            <text>{{ Math.round((mMin + mMax) / 2) }}万</text>
            <text>{{ Math.round(mMax) }}万</text>
          </view>
          <view class="money-inputs" :class="{ disabled: mFace }">
            <input class="m-input" type="number" :value="Math.round(mMin)" placeholder="最低" @input="onMinInput" />
            <text class="m-unit">万</text>
            <text class="m-dash">—</text>
            <input class="m-input" type="number" :value="Math.round(mMax)" placeholder="最高" @input="onMaxInput" />
            <text class="m-unit">万</text>
          </view>
          <view class="cal-foot">
            <text class="cal-reset" @tap="clearMoney">清除</text>
            <view class="money-done" @tap="startClosePanel">收起</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 蒙层：面板展开时置灰下方内容，点击外部收起 -->
    <view v-if="panel" class="panel-catcher" :class="{ closing }" :style="{ top: maskTop }" @tap="startClosePanel"></view>

    <!-- Banner（可点击 → 发布难题；内部微编排 + 单次扫光） -->
    <view class="banner" @tap="goPublish">
      <view class="banner-icon">战</view>
      <view class="banner-info">
        <text class="banner-title">技术攻关，等你来战</text>
        <text class="banner-sub">汇聚产业技术难题 · 诚邀揭榜攻关</text>
      </view>
    </view>
    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
    <!-- 信息行：共 N 项 + 排序 -->
    <view class="ir">
      <text>共 <text class="irn">{{ filteredAll.length }}</text> 项难题</text>
      <view class="irs-wrap">
        <text class="irs" @tap.stop="toggleSort">{{ sortLabel }} ▾</text>
        <view v-if="showSort" class="spop" :class="{ closing: sortClosing }" @tap.stop>
          <view v-for="s in SORTS" :key="s.v" class="sp-opt" :class="{ act: sort === s.v }" @tap="pickSort(s.v)">
            <text>{{ s.l }}</text><text v-if="sort === s.v" class="chk">✓</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 骨架 -->
    <view v-if="loading" class="skl">
      <view v-for="i in 4" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w80"></view>
          <view class="sk-l w60"></view>
        </view>
      </view>
    </view>

    <!-- 错误 -->
    <view v-else-if="err && !list.length" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchAll">重新加载</view>
      </u-empty>
    </view>

    <!-- 空 -->
    <view v-else-if="!list.length" class="st">
      <u-empty description="暂无匹配难题">
        <text class="sth">试试调整筛选条件或搜索关键词</text>
        <view class="stb" @tap="resetAll">清除筛选</view>
      </u-empty>
    </view>

    <!-- 列表：纯文字卡片（后端不支持图片；左缘领域色条 + tag 为视觉锚点） -->
    <view v-else class="cl" :class="{ replay }">
      <view v-for="x in list" :key="x.id" class="card" :data-id="x.id" :class="{ seen: seen.has(x.id) }" hover-class="tap-scale" hover-start-time="0" hover-stay-time="120" @tap="goDetail(x)">
        <view class="c-bar" :style="{ background: x.tagC }" />
        <view class="c-top">
          <view class="c-badges">
            <text class="c-tag" :style="{ color: x.tagC, background: x.tagBg }">{{ x.f }}</text>
            <text class="c-st" :class="'st-' + x.stCls">{{ x.stLabel }}</text>
          </view>
          <view class="c-budget"><text class="lb">悬赏</text><text class="vl" :class="{ face: x.budgetText === '面议' }">{{ x.budgetText }}</text></view>
        </view>
        <text class="ct">{{ x.t }}</text>
        <text v-if="x.d" class="c-desc">{{ x.d }}</text>
        <view class="c-meta">
          <text>发布于 {{ x.dt }}</text>
          <text class="c-dot">·</text>
          <text class="c-dl" :class="{ hot: x.urgent }">{{ x.dl }}</text>
        </view>
      </view>
    </view>

    <view v-if="list.length" class="lm">{{ loadMoreText }}</view>
    <view v-if="mockMode && isDev" class="mock-note">当前为演示数据 · 接口就绪后自动切换</view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>

</template>

<script setup>
import { ref, computed, nextTick, getCurrentInstance, reactive } from 'vue'
import { onLoad, onShow, onReady, onPullDownRefresh, onReachBottom, onPageScroll, onUnload } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { MOCK_CHALLENGES } from '@/utils/mockChallenges'
import { useReduceMotion } from '@/utils/motion'

const PAGE_SIZE = 100
const MAX_PAGES = 10
const RENDER_STEP = 100 // DOM 渲染上限步长：数据全量入 fullList，渲染切片逐段揭示
const SEARCH_DEBOUNCE_MS = 250 // 搜索防抖：击键即筛改为停顿 250ms 后筛（防每键整表重渲染）
const PANEL_CLOSE_MS = 230 // 面板退场 210ms + 缓冲：动画播完再 v-if 移除
const SORT_CLOSE_MS = 170  // 排序弹层退场 150ms + 缓冲
const isDev = process.env.NODE_ENV === 'development' // 演示数据横幅仅开发环境展示

/* ===== 静态配置 ===== */
const FIELD_OPTS = [
  { v: '', l: '全部' },
  { v: '飞控系统', l: '飞控系统' },
  { v: '动力电池', l: '动力电池' },
  { v: 'AI算法', l: 'AI算法' },
  { v: '通信链路', l: '通信链路' },
  { v: '新型材料', l: '新型材料' },
  { v: '载荷设备', l: '载荷设备' },
]
const STATUS_OPTS = [
  { v: '', l: '全部' },
  { v: '进行中', l: '进行中' },
  { v: '紧急', l: '紧急' },
  { v: '已截止', l: '已截止' },
]
const QUICK_OPTS = [
  { v: 'all', l: '全部' },
  { v: '7d', l: '一周内' },
  { v: '30d', l: '一个月内' },
]
const MONEY_OPTS = [
  { v: 'lt10', l: '10万以下' },
  { v: '10-50', l: '10-50万' },
  { v: 'gt50', l: '50万以上' },
]
const SORTS = [
  { v: 'latest', l: '最新发布' },
  { v: 'reward', l: '悬赏最高' },
  { v: 'deadline', l: '即将截止' },
]
const SORT_LABEL = { latest: '最新发布', reward: '悬赏最高', deadline: '即将截止' }
const WEEK = ['日', '一', '二', '三', '四', '五', '六']
const FIELD_ALIAS = {
  '飞控': '飞控系统', '飞控系统': '飞控系统',
  '电池': '动力电池', '动力电池': '动力电池',
  'AI': 'AI算法', 'AI算法': 'AI算法',
  '通信': '通信链路', '通信链路': '通信链路',
  '材料': '新型材料', '新型材料': '新型材料',
  '载荷': '载荷设备', '载荷设备': '载荷设备',
  '集群': '集群协同', '集群协同': '集群协同',
}
/* 领域标签配色：文字卡片无图片时的视觉锚点（按领域着色，保持身份辨识；对比度 ≥4.5:1） */
const FIELD_TAG = {
  '飞控系统': { tagC: '#0d47a1', tagBg: '#E3EDF9' },
  '动力电池': { tagC: '#B54708', tagBg: '#FDEEE4' },
  'AI算法': { tagC: '#4a148c', tagBg: '#F0E9F7' },
  '通信链路': { tagC: '#1a237e', tagBg: '#E7E9F4' },
  '新型材料': { tagC: '#004d40', tagBg: '#E4F2EF' },
  '载荷设备': { tagC: '#b71c1c', tagBg: '#FBE9E9' },
}
const FIELD_TAG_DEFAULT = { tagC: '#344054', tagBg: '#EEF1F4' }

/* ===== 状态 ===== */
const q = ref('')
const panel = ref('')
const selField = ref('')
const selStatus = ref('')
const quick = ref('all')
const rangeStart = ref('')
const rangeEnd = ref('')
const mMin = ref(0)
const mMax = ref(100)
const mPreset = ref('')
const mFace = ref(false) // 面议独立档位（与金额区间互斥）
const moneyActive = ref(false)
const sort = ref('latest')
const showSort = ref(false)
const closing = ref(false) // 筛选面板退场动画中（v-if 延迟移除）
const sortClosing = ref(false) // 排序弹层退场动画中
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
const replay = ref(false) // 列表轻淡入重播开关：离散筛选/排序后开启
const revealList = () => { // 无重排轻淡入：仅离散操作调用（滑块拖动/搜索打字每帧变化，不播）
  if (noMotion.value) return
  replay.value = false
  nextTick(() => { replay.value = true })
}
const loading = ref(true)
const err = ref(false)
const mockMode = ref(false)
const list = ref([])
const filteredAll = ref([]) // 过滤排序后的完整结果："共 N 项"与筛选语义用完整集，DOM 渲染用切片
const renderCap = ref(RENDER_STEP) // DOM 渲染上限：上拉触底 +100 逐段揭示（筛选仍作用于全量数据）
const fullList = ref([])
const total = ref(0)
const statusBarHeight = ref(20)
const headH = ref(102) // sticky-head 实测高度（兜底 ≈ 搜索框+筛选器）
const showBt = ref(false)
const sliderW = ref(0)
const sliderLeft = ref(0)
const allLoaded = ref(true) // 服务端数据是否已全部拉取（false 且 hasMore → 被 MAX_PAGES 截断）
let nextPage = 1
let dragging = null // 'min' | 'max' | null
const moneyDragging = ref(false) // 拖动中：抑制重置按钮出现（防拖动中筛选行挤压），松手后恢复显示
let closeT = null // 面板退场定时器（onUnload 清理）
let sortT = null // 排序弹层退场定时器（onUnload 清理）
let searchT = null // 搜索防抖定时器（onUnload 清理）
let fetching = false
let fetchingMore = false

const inst = getCurrentInstance()

/* ===== 滚动触达显示 ===== */
const seen = reactive(new Set()) // 已浮现卡片 id：首屏外卡片进入视口才浮现（边滚边看的丝滑感）
let obs = null // 视口交叉观察器（每次列表变化后重建，保证新增节点也被观察）
const forceSeen = () => { list.value.forEach((x) => seen.add(x.id)) } // 观察器不可用时全显示兜底
const reObserve = () => {
  if (noMotion.value) return
  if (obs) { try { obs.disconnect() } catch (e) { /* 忽略 */ } }
  try {
    obs = uni.createIntersectionObserver(inst.proxy, { thresholds: [0], initialRatio: 0 })
    obs.relativeToViewport({ bottom: 24 }).observe('.cl .card', (res) => {
      if (res && res.intersectionRatio > 0) {
        const id = res.dataset && res.dataset.id
        if (id) seen.add(id)
      }
    })
  } catch (e) { forceSeen() }
}

const pad = (n) => (n < 10 ? '0' + n : '' + n)
const fmt = (key) => {
  if (!key) return ''
  const p = key.split('-')
  return p[1] + '-' + p[2]
}
const today = () => {
  const now = new Date()
  return now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate())
}
const addDays = (n) => {
  const d = new Date()
  d.setDate(d.getDate() + n)
  return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
}

/* ===== 派生 ===== */
const compLabel = computed(() => {
  if (selField.value && selStatus.value) return selField.value + ' · ' + selStatus.value
  return selField.value || selStatus.value || '领域状态'
})
const timeLabel = computed(() => {
  if (rangeStart.value && rangeEnd.value) return fmt(rangeStart.value) + ' ~ ' + fmt(rangeEnd.value)
  if (rangeStart.value) return fmt(rangeStart.value) + ' 起'
  if (quick.value === '7d') return '一周内'
  if (quick.value === '30d') return '一个月内'
  return '发布时间'
})
const moneyLabel = computed(() => {
  if (!moneyActive.value) return '悬赏金额'
  if (mFace.value) return '面议'
  if (mPreset.value) return MONEY_OPTS.find((o) => o.v === mPreset.value)?.l || '金钱'
  if (mMin.value === 0) return '≤' + Math.round(mMax.value) + '万'
  if (mMax.value === 100) return '≥' + Math.round(mMin.value) + '万'
  return Math.round(mMin.value) + '-' + Math.round(mMax.value) + '万'
})
/* 是否有激活筛选：控制重置按钮显隐（金钱拖动中不显示，松手后再出现） */
const hasActiveFilters = computed(() => !!(
  q.value || selField.value || selStatus.value ||
  quick.value !== 'all' || rangeStart.value ||
  (moneyActive.value && !moneyDragging.value)
))
const maskTop = computed(() => (statusBarHeight.value + 44 + headH.value) + 'px')
const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新发布')
const hasMore = computed(() => filteredAll.value.length < total.value)
const hasMoreRender = computed(() => filteredAll.value.length > list.value.length)
const loadMoreText = computed(() => {
  if (hasMoreRender.value) return '— 上拉加载更多 —'
  if (!allLoaded.value && hasMore.value) return '共 ' + total.value + ' 条，仅显示前 ' + fullList.value.length + ' 条'
  return hasMore.value ? '— 上拉加载更多 —' : '— 没有更多了 —'
})

const mMinPx = computed(() => (sliderW.value * mMin.value) / 100)
const mMaxPx = computed(() => (sliderW.value * mMax.value) / 100)
const rangeStyle = computed(() => ({
  left: mMinPx.value + 'px',
  width: Math.max(0, mMaxPx.value - mMinPx.value) + 'px',
}))

const calTip = computed(() => {
  if (rangeStart.value && rangeEnd.value) return '已选择 ' + fmt(rangeStart.value) + ' ~ ' + fmt(rangeEnd.value)
  if (rangeStart.value) return '已选开始 ' + fmt(rangeStart.value) + '，请选择结束日期'
  return '点击选择开始日期，再点结束日期'
})

/* ===== 日历 ===== */
const nowD = new Date()
const calYear = ref(nowD.getFullYear())
const calMonth = ref(nowD.getMonth() + 1)
const calCells = computed(() => {
  const cells = []
  const first = new Date(calYear.value, calMonth.value - 1, 1)
  const startDow = first.getDay() // 周日=0
  const days = new Date(calYear.value, calMonth.value, 0).getDate()
  const todayKey = today()
  // quick 档位（一周内/一个月内）起点：给日历范围高亮用（quick 与区间互斥，不会叠加）
  const quickFrom = quick.value === '7d' ? addDays(-6) : quick.value === '30d' ? addDays(-29) : ''
  for (let i = 0; i < startDow; i++) cells.push({ empty: true, cls: ['empty'], key: '', day: '' })
  for (let d = 1; d <= days; d++) {
    const key = calYear.value + '-' + pad(calMonth.value) + '-' + pad(d)
    const cls = []
    if (key === todayKey) cls.push('today')
    if (rangeStart.value && key === rangeStart.value) cls.push('sel')
    if (rangeEnd.value && key === rangeEnd.value) cls.push('sel')
    if (rangeStart.value && rangeEnd.value && key > rangeStart.value && key < rangeEnd.value) cls.push('range')
    // quick 档位：起点至今天（含今天，today 描边 + 范围底色叠加）弱高亮，跨月时只高亮当前月内部分
    else if (quickFrom && !rangeStart.value && key >= quickFrom && key <= todayKey) cls.push('range')
    cells.push({ empty: false, cls, key, day: d })
  }
  return cells
})

/* ===== 数据映射 ===== */
const normField = (f) => FIELD_ALIAS[f] || f || '其他'
const daysLeft = (d) => {
  if (!d) return null
  const diff = new Date(d) - new Date()
  return Number.isFinite(diff) ? Math.max(0, Math.ceil(diff / 86400000)) : null
}
const statusOf = (it) => {
  const s = String(it.status || '').toLowerCase()
  const d = daysLeft(it.deadline)
  if (s === 'closed' || s === '已截止') return { label: '已截止', cls: 'closed' }
  // 已过截止时间（含不足一天的过期，daysLeft 取整到 0）优先于"紧急"展示
  if (d != null && new Date(it.deadline).getTime() < Date.now()) return { label: '已截止', cls: 'closed' }
  if (s === 'urgent' || s === '紧急' || (d != null && d <= 7)) return { label: '紧急', cls: 'urgent' }
  return { label: '进行中', cls: 'open' }
}
const fmtMoney = (wan) => {
  if (wan == null || wan <= 0) return '面议'
  if (wan >= 1) return '¥' + (wan % 1 === 0 ? wan : wan.toFixed(1)) + '万'
  return '¥' + Math.round(wan * 10000)
}
const mapItem = (it) => {
  const dl = daysLeft(it.deadline)
  const st = statusOf(it)
  const wan = (it.budget_fen != null ? it.budget_fen / 100 / 10000 : 0)
  const f = normField(it.field)
  return {
    id: it.id,
    t: it.title || '未命名难题',
    f,
    d: it.description || '',
    dt: (it.created_at || '').slice(0, 10),
    budget: wan,
    budgetText: fmtMoney(wan),
    stLabel: st.label,
    stCls: st.cls,
    urgent: st.cls === 'urgent',
    dl: dl == null ? '截止待定' : st.cls === 'closed' ? '截止 ' + (it.deadline || '').slice(0, 10) : dl + ' 天后截止',
    ddl: it.deadline || '',
    created: it.created_at || '',
    tagC: (FIELD_TAG[f] || FIELD_TAG_DEFAULT).tagC,
    tagBg: (FIELD_TAG[f] || FIELD_TAG_DEFAULT).tagBg,
  }
}

/* ===== 过滤与排序 ===== */
const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter((x) => (x.t + ' ' + x.f + ' ' + x.d).toLowerCase().includes(kw))
  if (selField.value) items = items.filter((x) => x.f === selField.value)
  if (selStatus.value) items = items.filter((x) => x.stLabel === selStatus.value)
  if (quick.value === '7d') {
    const from = addDays(-6)
    items = items.filter((x) => x.dt >= from)
  } else if (quick.value === '30d') {
    const from = addDays(-29)
    items = items.filter((x) => x.dt >= from)
  } else if (rangeStart.value && rangeEnd.value) {
    items = items.filter((x) => x.dt >= rangeStart.value && x.dt <= rangeEnd.value)
  } else if (rangeStart.value) {
    items = items.filter((x) => x.dt === rangeStart.value)
  }
  if (moneyActive.value) {
    if (mFace.value) items = items.filter((x) => x.budget === 0) // 面议档：只看面议
    else items = items.filter((x) => x.budget > 0 && x.budget >= mMin.value && x.budget <= mMax.value)
  }
  if (sort.value === 'reward') items.sort((a, b) => b.budget - a.budget)
  else if (sort.value === 'deadline') {
    // 已截止排最后，其余按截止时间升序（越近越靠前）
    items.sort((a, b) => {
      const ac = a.stCls === 'closed' ? 1 : 0
      const bc = b.stCls === 'closed' ? 1 : 0
      if (ac !== bc) return ac - bc
      return String(a.ddl || '9999').localeCompare(String(b.ddl || '9999'))
    })
  }
  else items.sort((a, b) => String(b.created).localeCompare(String(a.created)))
  filteredAll.value = items // 完整结果（计数/空态语义），DOM 只渲染前 renderCap 项防大列表卡顿
  list.value = items.slice(0, renderCap.value)
  list.value.slice(0, 6).forEach((x) => seen.add(x.id)) // 首屏 6 项同步标记 seen：走入场错峰不走滚动浮现
  nextTick(reObserve) // 列表变化后重建观察，新增节点也能触达浮现
}

const fetchAll = async (silent = false) => {
  if (fetching) return
  fetching = true
  loading.value = !silent // silent：保留当前列表，避免骨架屏顶替闪烁
  err.value = false
  nextPage = 1
  try {
    const acc = []
    let fetched = 0
    for (let page = 1; page <= MAX_PAGES; page++) {
      const res = await request({ url: '/api/v1/rd-challenges', data: { page, page_size: PAGE_SIZE } })
      const items = Array.isArray(res) ? res : (res?.items || [])
      acc.push(...items.map(mapItem))
      fetched = res?.total ?? acc.length
      nextPage = page
      if (items.length < PAGE_SIZE || acc.length >= fetched) break
    }
    allLoaded.value = nextPage < MAX_PAGES || acc.length >= fetched // 拉满10页且未齐 = 截断
    fullList.value = acc
    total.value = fetched
    mockMode.value = false
    applyFilter()
  } catch {
    // 接口失败：从未成功加载过才回退演示数据；已有数据则保留，
    // 避免下拉刷新时一次网络抖动就用演示数据顶替真实列表
    if (fullList.value.length === 0) {
      if (MOCK_CHALLENGES && MOCK_CHALLENGES.length) {
        fullList.value = (MOCK_CHALLENGES || []).map(mapItem)
        total.value = fullList.value.length
        mockMode.value = true
      } else {
        err.value = true // 无任何可显示数据 → 加载失败空态
      }
    } else {
      uni.showToast({ title: '加载失败，已显示上次数据', icon: 'none' })
    }
    applyFilter()
  } finally {
    loading.value = false
    fetching = false
  }
}

const fetchMore = async () => {
  // fetching 守卫：onShow 静默刷新（loading 保持 false）期间触底，防止与 fetchAll 并发拉同一页造成重复条目
  if (loading.value || fetching || fetchingMore || nextPage >= MAX_PAGES || !hasMore.value) return
  fetchingMore = true
  try {
    const page = nextPage + 1 // await 前捕获，防并发读到同一页
    const res = await request({ url: '/api/v1/rd-challenges', data: { page, page_size: PAGE_SIZE } })
    const items = Array.isArray(res) ? res : (res?.items || [])
    fullList.value = fullList.value.concat(items.map(mapItem))
    total.value = res?.total ?? fullList.value.length
    nextPage = page
    allLoaded.value = items.length < PAGE_SIZE || fullList.value.length >= total.value
    applyFilter()
  } catch { /* 静默：下次触底重试 */ } finally { fetchingMore = false }
}

/* ===== 筛选器交互（点选即筛、再点取消） ===== */
/* 开合规范：退场必须存在且可打断——先加 closing 类播退场动画，定时器到点再 v-if 移除；
   退场期间再次操作直接取消旧退场（动画可打断，不排队叠加） */
const startClosePanel = () => {
  if (closing.value) return // 已在退场中，防重复触发叠加定时器
  closing.value = true
  clearTimeout(closeT)
  closeT = setTimeout(() => {
    panel.value = ''
    closing.value = false
    closeT = null
  }, PANEL_CLOSE_MS)
}
const openPanel = (p) => {
  // 打开新面板：取消未完成的退场（可打断）；排序弹层直接收起不播动画，避免同帧两个浮层同时动
  clearTimeout(closeT); closeT = null
  closing.value = false
  clearTimeout(sortT); sortT = null
  sortClosing.value = false
  showSort.value = false
  panel.value = p
  if (p === 'money') nextTick(measureSlider)
}
const startCloseSort = () => {
  if (sortClosing.value) return
  sortClosing.value = true
  clearTimeout(sortT)
  sortT = setTimeout(() => {
    showSort.value = false
    sortClosing.value = false
    sortT = null
  }, SORT_CLOSE_MS)
}
const closeAll = () => {
  // 页面空白处点击：打开着的浮层走退场动画收起（mask 之外区域也能关面板）
  if (panel.value) startClosePanel()
  if (showSort.value) startCloseSort()
}
const togglePanel = (p) => {
  if (panel.value === p) { startClosePanel(); return } // 再点当前 pill → 退场收起
  openPanel(p)
}
const toggleComp = (key, v) => {
  if (key === 'field') selField.value = selField.value === v ? '' : v
  else selStatus.value = selStatus.value === v ? '' : v
  applyFilter() // 点选即筛；不关面板，支持"领域+状态"一次组合
  revealList()
  // 收起方式：点蒙层 / 再点当前 pill / 点其他 pill
}
const pickQuick = (v) => {
  if (v !== 'all' && quick.value === v) quick.value = 'all'
  else quick.value = v
  rangeStart.value = ''
  rangeEnd.value = ''
  applyFilter()
  revealList()
}
const calShift = (d) => {
  calMonth.value += d
  if (calMonth.value < 1) { calMonth.value = 12; calYear.value-- }
  if (calMonth.value > 12) { calMonth.value = 1; calYear.value++ }
}
/* 日历横滑切月：touchstart 记起点，touchend 算位移，横向 >50px 且明显偏横向才切月（左滑下一月，右滑上一月）；纵向滑动不拦截，面板自身滚动不受影响 */
let calSwipeX = 0
let calSwipeY = 0
const calTouchStart = (e) => {
  const t = e.touches && e.touches[0]
  if (!t) return
  calSwipeX = t.clientX
  calSwipeY = t.clientY
}
const calTouchEnd = (e) => {
  const t = e.changedTouches && e.changedTouches[0]
  if (!t || !calSwipeX) return
  const dx = t.clientX - calSwipeX
  const dy = t.clientY - calSwipeY
  calSwipeX = 0
  if (Math.abs(dx) > 50 && Math.abs(dx) > Math.abs(dy) * 2) calShift(dx < 0 ? 1 : -1)
}
const onCalClick = (key) => {
  if (!key) return
  if (!rangeStart.value) {
    // 第一击：设开始
    rangeStart.value = key
    rangeEnd.value = ''
  } else if (!rangeEnd.value) {
    // 第二击：无论先点哪天，都直接形成 [早, 晚] 区间（顺序无关）
    if (key < rangeStart.value) {
      rangeEnd.value = rangeStart.value
      rangeStart.value = key
    } else {
      rangeEnd.value = key
    }
  } else {
    // 已有完整区间 → 重新开始
    rangeStart.value = key
    rangeEnd.value = ''
  }
  quick.value = 'all'
  // 每次点击都刷新列表，保证筛选结果与日历高亮/提示一致
  applyFilter()
  revealList()
}
const resetTime = () => {
  rangeStart.value = ''
  rangeEnd.value = ''
  quick.value = 'all'
  applyFilter()
  revealList()
}

/* ===== 金钱：滑动 + 填写 + 区间按钮 ===== */
const measureSlider = () => {
  uni.createSelectorQuery().in(inst?.proxy || inst).select('.slider-bar').boundingClientRect((rect) => {
    if (rect) {
      sliderW.value = rect.width
      sliderLeft.value = rect.left
    }
  }).exec()
}
const measureHead = () => {
  uni.createSelectorQuery().in(inst?.proxy || inst).select('.sticky-head').boundingClientRect((rect) => {
    if (rect && rect.height) headH.value = rect.height
  }).exec()
}
const setPos = (x) => {
  if (!sliderW.value || mFace.value) return
  // 连续浮点追踪不取整：拇指 1:1 跟手零跳格；金额只在显示层四舍五入
  const v = Math.max(0, Math.min(100, (x / sliderW.value) * 100))
  if (dragging === 'min') {
    mMin.value = Math.min(v, mMax.value)
  } else if (dragging === 'max') {
    mMax.value = Math.max(v, mMin.value)
  }
  // 手动滑动后预设区间不再适用，标签回退为实际数值
  mPreset.value = ''
  // 点选即筛：拖动即激活区间（全范围 = 不筛选），与预设档位交互一致
  moneyActive.value = !(mMin.value === 0 && mMax.value === 100)
  // 拖动中不重筛列表：applyFilter 整表重渲染 + 重建交叉观察器，每帧执行会让 touchmove 排队丢帧，
  // 拇指滞后且手指停下后仍在"追赶"。此处只更新拇指/标签（轻量重渲染），列表在松手时一次性筛选
}
const onTouchStart = (e) => {
  if (!sliderW.value || mFace.value) return
  const x = (e.touches && e.touches[0] ? e.touches[0].clientX : 0) - sliderLeft.value
  const dMin = Math.abs(x - mMinPx.value)
  const dMax = Math.abs(x - mMaxPx.value)
  // 两拇指重叠/贴近（≤10px）时按触点方位选：触点在中点右侧 → 拖 max，否则 min——避免抓到重叠拇指后拖动被卡死
  if (Math.abs(mMinPx.value - mMaxPx.value) <= 10) {
    dragging = x >= (mMinPx.value + mMaxPx.value) / 2 ? 'max' : 'min'
  } else {
    dragging = dMin <= dMax ? 'min' : 'max'
  }
  moneyDragging.value = true
  setPos(x)
}
const onTouchMove = (e) => {
  if (!dragging) return
  const x = (e.touches && e.touches[0] ? e.touches[0].clientX : 0) - sliderLeft.value
  setPos(x)
}
const onTouchEnd = () => {
  if (!moneyDragging.value) return // 未真正开始拖动（如测量失败/面议禁用），不处理
  dragging = null
  moneyDragging.value = false
  // 松手吸附到整数档位：标签/筛选语义干净（浮点仅用于拖动中的跟手追踪）
  mMin.value = Math.round(mMin.value)
  mMax.value = Math.round(mMax.value)
  moneyActive.value = !(mMin.value === 0 && mMax.value === 100)
  applyFilter() // 松手才筛列表：拖动中零整表重渲染，拇指 1:1 跟手、松手即停
}
const onMinInput = (e) => {
  const v = e.detail.value === '' ? 0 : parseInt(e.detail.value, 10) // 清空 = 无下限
  if (!isNaN(v)) {
    mMin.value = Math.max(0, Math.min(100, Math.min(v, mMax.value)))
    mPreset.value = '' // 手动输入后预设区间不再适用
    moneyActive.value = !(mMin.value === 0 && mMax.value === 100)
    applyFilter()
  }
}
const onMaxInput = (e) => {
  const v = e.detail.value === '' ? 100 : parseInt(e.detail.value, 10) // 清空 = 无上限
  if (!isNaN(v)) {
    mMax.value = Math.max(0, Math.min(100, Math.max(v, mMin.value)))
    mPreset.value = ''
    moneyActive.value = !(mMin.value === 0 && mMax.value === 100)
    applyFilter()
  }
}
const toggleFace = () => {
  mFace.value = !mFace.value
  if (mFace.value) {
    // 选中面议：与预设/区间互斥，只看面议
    mPreset.value = ''
    moneyActive.value = true
  } else {
    // 取消面议：恢复区间语义（全范围则取消激活）
    moneyActive.value = !(mMin.value === 0 && mMax.value === 100)
  }
  applyFilter()
  revealList()
}
const pickPreset = (v) => {
  if (mPreset.value === v) {
    mPreset.value = ''
    moneyActive.value = false
    mMin.value = 0
    mMax.value = 100
    applyFilter()
    revealList()
    return
  }
  mPreset.value = v
  mFace.value = false // 预设与面议互斥
  moneyActive.value = true
  if (v === 'lt10') { mMin.value = 0; mMax.value = 10 }
  else if (v === '10-50') { mMin.value = 10; mMax.value = 50 }
  else if (v === 'gt50') { mMin.value = 50; mMax.value = 100 }
  applyFilter()
  revealList()
}
const clearMoney = () => {
  mPreset.value = ''
  mFace.value = false
  moneyActive.value = false
  mMin.value = 0
  mMax.value = 100
  applyFilter()
  revealList()
}

/* ===== 排序 / 其他 ===== */
const toggleSort = () => {
  if (showSort.value) { startCloseSort(); return }
  // 打开排序：面板若有展开则直接收起（不播动画，避免面板+弹层同帧动画）
  clearTimeout(closeT); closeT = null
  closing.value = false
  panel.value = ''
  clearTimeout(sortT); sortT = null
  sortClosing.value = false
  showSort.value = true
}
const pickSort = (v) => { sort.value = v; startCloseSort(); applyFilter(); revealList() }
/* 搜索防抖：击键/点搜索均走 250ms 停顿后筛选，防每键整表重渲染（点 × 清除不防抖，跟手优先） */
const onSearch = () => {
  clearTimeout(searchT)
  searchT = setTimeout(applyFilter, SEARCH_DEBOUNCE_MS)
}
const clearSearch = () => { clearTimeout(searchT); q.value = ''; applyFilter(); revealList() } // 搜索框 ×：即时（跟手优先）
const resetAll = () => {
  q.value = ''
  selField.value = ''
  selStatus.value = ''
  quick.value = 'all'
  rangeStart.value = ''
  rangeEnd.value = ''
  mPreset.value = ''
  mFace.value = false
  moneyActive.value = false
  mMin.value = 0
  mMax.value = 100
  sort.value = 'latest'
  // 重置是整体操作：直接移除浮层（不播退场），并清理退场定时器，避免残留 closing 态
  clearTimeout(closeT); closeT = null
  closing.value = false
  clearTimeout(sortT); sortT = null
  sortClosing.value = false
  clearTimeout(searchT); searchT = null
  showSort.value = false
  panel.value = ''
  applyFilter()
  revealList()
}
const goDetail = (x) => {
  uni.navigateTo({ url: '/pkg-eco/pages/challenges/detail?id=' + encodeURIComponent(x.id) })
}
const goPublish = () => {
  uni.navigateTo({ url: '/pkg-eco/pages/challenges/publish' })
}
const goBack = () => uni.navigateBack({ fail: () => uni.redirectTo({ url: '/pkg-eco/pages/challenges/list' }) })
const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })

onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  fetchAll()
})
let shownOnce = false
onShow(() => {
  // onLoad 先于首个 onShow 触发，用 flag 去重首次；从 detail/publish 返回时静默刷新
  if (shownOnce) fetchAll(true)
  shownOnce = true
})

onUnload(() => {
  // 规范强制：页面卸载清除所有定时器（退场延迟），防回调泄漏
  clearTimeout(closeT)
  clearTimeout(sortT)
  clearTimeout(searchT)
})
onReady(() => { measureHead(); reObserve() })
onPullDownRefresh(async () => {
  // silent：原生下拉动画本身已提供刷新反馈，静默刷新避免列表被骨架屏顶替闪烁
  await fetchAll(true)
  uni.stopPullDownRefresh()
})
onReachBottom(() => {
  // 先揭示已拉取的渲染切片（DOM 渐次展开，防千条卡片同屏），渲染殆尽后再向服务端翻页
  if (hasMoreRender.value) {
    renderCap.value += RENDER_STEP
    applyFilter()
    return
  }
  fetchMore()
})
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

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 极淡灰投影，从白底上"浮"起 ===== */
.sbar {
  padding: 12px 12px 8px;
  background: #fff;
}
.b-search {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05); /* 双层投影：接触阴影贴地 + 环境阴影弥散浮起 */
  display: flex;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.b-search-ic { width: 15px; height: 15px; flex: none; }
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; } /* 热区扩大：视觉 × 外扩 10px，点击不脱靶 */
/* 小红书风格搜索按钮：无底色文字 + 左侧细竖杠分隔 */
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== Banner（projects 风格） ===== */
.banner {
  margin: 12px 14px;
  padding: 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #0A66C2 0%, #074D92 100%);
  display: flex;
  align-items: center;
  gap: 12px;
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
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: block; } /* 白 95%：蓝底上 ≥4.5:1 达标（原 90% 约 4.8:1，字号 11→12 更稳） */
/* ===== 固定头部：Banner + 筛选器 ===== */
.panel-catcher {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 30; /* 低于 sticky-head(40)：否则蒙层盖住筛选面板，面板上点击/滑动全被拦截 */
  background: rgba(16, 24, 40, 0.2); /* 真变暗：面板展开时置灰下方内容（蓝灰黑 20%） */
  animation: maskIn .22s ease-out; /* 遮罩与面板同步淡入 */
}
.panel-catcher.closing {
  animation: maskOut .16s ease-in forwards; /* 同步淡出 */
}
@keyframes maskIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes maskOut { from { opacity: 1; } to { opacity: 0; } }
.sticky-head {
  position: sticky;
  z-index: 40;
  background: #fff;
}

/* ===== 白色板块（信息行 + 列表）：与页面同底，融入不分块 ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 三个筛选器 ===== */
.fbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px 12px;
  background: #fff;
}
.fpill {
  flex: 1;
  min-width: 0; /* 优化：允许收缩，配合 .fpv 省略号防长标签溢出 */
  min-height: 40px; /* 触控目标：34px→40px（接近微信 44px 建议值，筛选高频操作） */
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff; /* 白上白：纯白填充 + 灰描边 + 极淡灰投影 */
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  overflow: hidden; /* 优化：溢出裁剪，长组合标签（如"飞控系统 · 紧急"）不再撑破 pill */
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease; /* 优化：transition:all → 仅过渡实际变化的属性 */
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; } /* 优化：新增——pill 长标签省略号 */
.farr { font-size: 11px; color: #667085; flex: none; }
.freset {
  flex: none;
  min-height: 40px; /* 触控目标：34px→40px */
  padding: 0 10px;
  border-radius: 8px;
  color: #667085;
  font-size: 12px;
  display: flex;
  align-items: center;
}
/* 重置按钮：随激活筛选状态弹出（ios-pop 弹簧） */
.freset { animation: chipIn .22s cubic-bezier(.34, 1.8, .64, 1) backwards; }
@keyframes chipIn { from { transform: scale(.85); } to { transform: scale(1); } }

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
.irs-wrap { position: relative; }
.irs { color: #0A66C2; font-weight: 500; padding: 8px 4px 8px 12px; } /* 热区扩大：6px→8px 纵向 */
.spop {
  position: absolute;
  top: 32px;
  right: 0;
  z-index: 90;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 8px 28px rgba(16, 24, 40, 0.12);
  padding: 6px;
  min-width: 140px;
  animation: spopIn .22s cubic-bezier(.32, .72, 0, 1); /* ios-decel：下拉流体减速，越到终点越柔 */
}
.spop.closing {
  animation: spopOut .15s ease-in forwards;
}
@keyframes spopIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes spopOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-4px); }
}
.sp-opt {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  color: #17212B;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sp-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }
.chk { color: #0A66C2; font-size: 12px; }

/* ===== 列表项：纯文字卡片（白上白：灰描边 + 极淡灰投影浮起，窄缝 8px；左缘领域色条 + tag 为视觉锚点） ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
  display: flex;
  flex-direction: column;
  gap: 7px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC; /* 描边提级：低端设备投影失效时仍与白底可分辨 */
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06); /* 卡片浮层：大偏移 + 宽模糊 + 低透明，柔和环境阴影 */
}
.tap-scale { transform: scale(0.95); opacity: 0.9; }
/* 领域色条：卡片左缘 3px 全高色条，无图片时的领域身份锚点 */
.c-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  border-radius: 10px 0 0 10px; /* 跟随卡片左圆角 */
  transform-origin: center top; /* 入场 scaleY 从顶部抽出 */
}
.c-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.c-badges { display: flex; gap: 6px; }
.c-tag, .c-st {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.c-tag { color: #074D92; background: #EAF3FB; } /* 兜底色；实际按领域色由 mapItem 传入 */
.c-st.st-open { color: #0B6B41; background: #E9F7F0; }
.c-st.st-urgent { color: #B42318; background: #FDECEC; }
.c-st.st-closed { color: #5D6B82; background: #EEF1F4; }
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
  font-size: 12px; /* 11.5px→12px：正文下限 */
  color: #667085;
}
.c-dot { color: #DDE1E6; }
.c-dl { color: #667085; font-weight: 500; }
.c-dl.hot { color: #D92D20; font-weight: 700; }
.c-budget { display: flex; align-items: baseline; gap: 3px; color: #C2410C; }
.c-budget .lb { font-size: 12px; font-weight: 500; }
.c-budget .vl { font-size: 18px; font-weight: 800; }
.c-budget .vl.face { font-size: 13px; font-weight: 600; color: #667085; } /* 面议非数字：不用金额的重量渲染，避免误读为最高悬赏 */

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
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 面板（通用） ===== */
.panel-wrap {
  position: relative;
  background: #fff;
  /* 浮层档 200-300ms：进场 ios-decel .3s（iOS sheet 流体减速）；退场 .21s ease-in（= 进场 ×0.7，必须存在） */
  animation: panelIn .3s cubic-bezier(.32, .72, 0, 1);
}
.panel-wrap.closing {
  animation: panelOut .21s ease-in forwards; /* forwards：退场结束态保持到 v-if 移除，防闪跳 */
}
.panel {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  z-index: 41;
  background: #fff;
  border-radius: 0 0 12px 12px;
  box-shadow: 0 12px 24px rgba(16, 24, 40, 0.1);
  padding: 12px 14px;
  max-height: 62vh;
  overflow-y: auto;
}
@keyframes panelIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes panelOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-10px); }
}
.p-head { font-size: 12px; color: #667085; margin-bottom: 8px; }
.p-group { font-size: 13px; font-weight: 700; color: #344054; margin: 12px 0 6px; }
.p-group:first-of-type { margin-top: 0; }
.p-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.p-chip {
  min-height: 36px; /* 触控目标：30px→36px（与 publish 页 pill 同步） */
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
}
.p-chip.act { color: #fff; border-color: #0A66C2; background: #0A66C2; font-weight: 600; }

/* ===== 日历 ===== */
.cal { margin-top: 12px; border-top: 1px solid #F0F1F3; padding-top: 10px; }
.cal-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.cal-title { font-size: 13px; font-weight: 600; color: #17212B; }
.cal-nav {
  width: 36px; /* 触控目标：26px→36px（月历翻页高频拇指操作） */
  height: 36px;
  border-radius: 50%;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667085;
  font-size: 14px;
}
.cal-week, .cal-cells { display: grid; grid-template-columns: repeat(7, 1fr); text-align: center; }
.cal-week { font-size: 12px; color: #667085; margin-bottom: 4px; }
.cal-week text { padding: 4px 0; }
.cal-cells { gap: 2px; }
.cal-cell {
  height: 40px; /* 触控目标：34px→40px */
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #17212B;
  border-radius: 8px;
}
.cal-cell.empty { pointer-events: none; }
.cal-cell.today { box-shadow: inset 0 0 0 1px #0A66C2; color: #0A66C2; font-weight: 600; }
.cal-cell.sel { background: #0A66C2; color: #fff; font-weight: 600; }
.cal-cell.range { background: #EAF3FB; color: #0A66C2; }
.cal-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #F0F1F3;
}
.cal-tip { font-size: 12px; color: #667085; }
.cal-reset { font-size: 13px; color: #0A66C2; font-weight: 500; padding: 8px 10px; } /* 热区扩大：4px→8px 纵向 */

/* ===== 金钱：滑动 + 填写 ===== */
.slider { position: relative; height: 44px; margin: 14px 8px 0; }
.slider-track {
  position: absolute;
  top: 20px;
  left: 0;
  right: 0;
  height: 4px;
  border-radius: 2px;
  background: #E4E7EC;
}
.slider-range {
  position: absolute;
  top: 20px;
  height: 4px;
  border-radius: 2px;
  background: #0A66C2;
  transition: background .2s ease; /* 优化：面议档位切换时颜色平滑过渡；拖动由内联 left/width 驱动，不加位置过渡防拖拽延迟——拖动跟手 1:1 无缓动（规范铁律四），松手即定格 */
}
.slider-bar { position: absolute; inset: 0; }
.thumb {
  position: absolute;
  top: 10px;
  width: 20px;
  height: 20px;
  margin-left: -10px;
  border-radius: 50%;
  background: #fff;
  border: 2px solid #0A66C2;
  box-shadow: 0 2px 6px rgba(16, 24, 40, 0.12);
  transition: border-color .2s ease; /* 优化：面议禁用时描边色平滑过渡 */
}
.slider-labels {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #667085;
  margin: 0 2px;
}
.money-inputs { display: flex; align-items: center; gap: 6px; margin-top: 12px; }
.m-input {
  flex: 1;
  min-height: 36px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  padding: 0 10px;
  font-size: 13px;
  color: #17212B;
  box-sizing: border-box;
}
.m-unit { font-size: 12px; color: #667085; flex: none; }
.m-dash { color: #667085; font-size: 12px; flex: none; }
.money-done {
  font-size: 13px;
  color: #fff;
  background: #0A66C2;
  padding: 6px 18px;
  border-radius: 6px;
  font-weight: 600;
}
/* 面议档位选中时：滑块/输入/确定置灰禁用 */
.slider.disabled .slider-bar { pointer-events: none; }
.slider.disabled .slider-range { background: #D9DEE3; }
.slider.disabled .thumb { border-color: #D9DEE3; }
.money-inputs.disabled { opacity: 0.5; pointer-events: none; transition: opacity .2s ease; } /* 优化：面议禁用时透明度平滑过渡 */

/* ===== 遮罩 / 加载更多 / 返回顶部 / mock 提示 ===== */
.lm { text-align: center; padding: 12px; font-size: 12px; color: #667085; }
.mock-note { text-align: center; padding: 0 0 16px; font-size: 10px; color: #667085; }
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
  z-index: 35; /* 优化：50→35，低于 sticky-head(40)——修复筛选面板展开时返回顶部按钮浮层穿透抢点击；仍高于蒙层(30)与页面内容 */
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(.34, 1.8, .64, 1); /* ios-pop：出现/隐藏弹簧收尾，返回顶部"弹"出来 */
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; } /* 按压即时到位 */

/* ===================== 动效规范（对齐全局动画规范） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许——仅重绘不重排）
   禁参与动画：top/left/width/height/margin（触发重排）、box-shadow/filter（低端安卓掉帧）
   时长：微反馈 150-200ms（按压按下 .08s 即时到位）/ 松手弹簧回位 .3s（ios-pop）/ 浮层 200-300ms / 页面级 ≤400ms；
        退场 = 进场 ×0.7 且必须存在
   曲线：两枚固定曲线——ios-pop cubic-bezier(.34,1.8,.64,1) 松手弹簧回弹（仅按压/弹出类 transform）+
        ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速（sheet/下拉进场）；
        其余进场 ease-out / 退场 ease-in / 循环 linear；除这两枚外禁手写 cubic-bezier
   数量：列表入场仅错峰首屏 6 项，其余卡片滚动触达浮现（observer 标记）；离散筛选/排序后仅前 4 项 180ms 轻淡入（禁大批量并发动画）
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 0) 滚动触达显示：首屏外卡片默认隐藏，进入视口才浮现（observer 标记 .seen → cardReveal）
   applyFilter 已同步标记前 6 项 seen → 首帧走入场错峰不受影响；
   无 fill：动画结束回落可见基态，不干扰按压弹簧；置于入场块之前 → 前 6 项由入场错峰优先接管 */
.card:not(.seen) { opacity: 0; transform: translateY(10px); }
.card.seen { animation: cardReveal .28s cubic-bezier(.32, .72, 0, 1); } /* ios-decel：浮现流体减速 */
@keyframes cardReveal { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* 1) 列表入场：前 6 项每 20ms 依次淡入上移（首屏可见范围；80ms 起 + 100ms 错峰 + 220ms 动画 = 400ms ≤ 400ms）
   backwards 填充 → 延迟期保持隐藏不闪跳；其余卡片走滚动触达显示 */
.card { animation: none; }
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
/* 领域色条与卡片同拍"点亮"：身份注入是入场的主旋律（scaleY 顶部抽出，与 cardIn 同错峰，380ms 内收完） */
.card:nth-child(-n+6) .c-bar { animation: barIn .2s ease-out backwards; }
.card:nth-child(1) .c-bar { animation-delay: 80ms; }
.card:nth-child(2) .c-bar { animation-delay: 100ms; }
.card:nth-child(3) .c-bar { animation-delay: 120ms; }
.card:nth-child(4) .c-bar { animation-delay: 140ms; }
.card:nth-child(5) .c-bar { animation-delay: 160ms; }
.card:nth-child(6) .c-bar { animation-delay: 180ms; }
@keyframes barIn { from { opacity: 0; transform: scaleY(.3); } to { opacity: 1; transform: scaleY(1); } }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
/* 列表刷新（离散筛选/排序后）：前 4 项轻淡入——2px 上移 + 半透明起，180ms + 30ms 错峰 ≤270ms；
   说明"这组内容因操作而更新"，无大位移不抢戏；nth-child 优先级高于入场错峰（同源覆盖） */
.cl.replay .card:nth-child(-n+4) { animation: listFade .18s ease-out backwards; }
.cl.replay .card:nth-child(1) { animation-delay: 0ms; }
.cl.replay .card:nth-child(2) { animation-delay: 30ms; }
.cl.replay .card:nth-child(3) { animation-delay: 60ms; }
.cl.replay .card:nth-child(4) { animation-delay: 90ms; }
@keyframes listFade { from { opacity: .3; transform: translateY(2px); } to { opacity: 1; transform: translateY(0); } }
/* 卡片按压（快进慢出）：hover-start-time=0 按下立即触发；按下 .1s linear 直接到位，松手 .35s ios-pop 弹簧回位——iOS"即按即应、松手弹回"手感 */
.card { transition: transform .35s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop */
.card.tap-scale { transition-duration: .1s; transition-timing-function: linear; }

/* Banner 内部微编排（替代整块 fadeUp）：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms，
   总 340ms ≤ 400ms；同帧并发 ≤2（与信息行/卡片首项错开）；全部单次动画非循环，合规 */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; } /* 装饰圆收缩落位（scale 1.1→1 单次） */
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }
/* Banner 单次扫光（非循环装饰：100ms 起播 280ms 线性，380ms 内收完；
   ::before 默认态 translateX(-150%) 隐藏，动画结束复位 → 不驻留不循环） */
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
/* Banner 可点击：按压反馈（按下 .08s 即时到位，松手 .3s 弹簧回位；transform/opacity） */
.banner { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop */
.banner:active { transform: scale(.985); opacity: .95; transition: transform .08s linear; }
/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；循环动画 1.2-1.6s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 2) 交互反馈：可点元素按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位；opacity/background 150-200ms） */
.freset:active { opacity: .7; }
.irs { transition: opacity .2s ease, transform .3s cubic-bezier(.34, 1.8, .64, 1); } /* ios-pop */
.irs:active { opacity: .7; transform: scale(.95); transition: transform .08s linear; }
.sp-opt { transition: background .2s ease, color .2s ease; }
.sp-opt:active { background: #F4F8FC; }
.b-sclr:active { opacity: .6; }
.b-sbtn { transition: opacity .2s ease; }
.b-sbtn:active { opacity: .5; }
.stb { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop：松手弹簧回位 */
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }
.cal-nav { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease; } /* ios-pop */
.cal-nav:active { transform: scale(.9); background: #E9EFF7; transition: transform .08s linear; }
.cal-reset { transition: opacity .2s ease; }
.cal-reset:active { opacity: .6; }
.money-done { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, opacity .15s ease; } /* ios-pop */
.money-done:active { transform: scale(.95); opacity: .9; transition: transform .08s linear; }

/* 3) 状态过渡：chip 选中 200ms 平滑 + ios-pop 微弹过冲回位；日历格选中/范围高亮 200ms 平滑；日历格按压轻微缩放（弹簧回位） */
.p-chip { transition: background .2s ease, border-color .2s ease, color .2s ease, transform .3s cubic-bezier(.34, 1.8, .64, 1); } /* ios-pop：松手弹簧回位 */
.p-chip:active { transform: scale(.94); transition: transform .08s linear; } /* 按下即时到位，其余按压变化同步走即时 */
.p-chip.act { animation: chipPop .3s cubic-bezier(.34, 1.8, .64, 1); } /* ios-pop：选中微弹带轻微过冲回位 */
@keyframes chipPop { from { transform: scale(.9); } to { transform: scale(1); } }
.cal-cell { transition: background .2s ease, color .2s ease, box-shadow .2s ease, transform .3s cubic-bezier(.34, 1.8, .64, 1); } /* ios-pop */
.cal-cell:active { transform: scale(.9); transition: transform .08s linear; }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .card,
.page.no-motion .c-bar,
.page.no-motion .banner,
.page.no-motion .ir { animation: none; } /* 装饰入场全关（色条抽出属缩放，一并关闭） */
.page.no-motion .card:not(.seen) { opacity: 1; transform: none; } /* 滚动浮现关闭：全部直接可见 */
.page.no-motion .cl.replay .card { animation: none; } /* 列表刷新轻淡入关闭（覆盖高优先级重播规则） */
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner::before,
.page.no-motion .banner::after { animation: none; } /* banner 内部微编排/扫光/装饰圆全关 */
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; } /* 循环呼吸关 */
.page.no-motion .panel-wrap { animation: panelFadeIn .22s ease-out; } /* 面板降级为纯淡入 */
.page.no-motion .panel-wrap.closing { animation: panelFadeOut .16s ease-in forwards; }
.page.no-motion .spop { animation: spopFadeIn .2s ease-out; }
.page.no-motion .spop.closing { animation: spopFadeOut .15s ease-in forwards; }
.page.no-motion .panel-catcher { animation: maskIn .22s ease-out; }
.page.no-motion .panel-catcher.closing { animation: maskOut .16s ease-in forwards; }
.page.no-motion .p-chip.act { animation: none; } /* 选中微弹属缩放，关闭；选中色保留 */
.page.no-motion .freset { animation: none; } /* 重置弹出属缩放，关闭 */
.page.no-motion .tap-scale { transform: none !important; } /* 按压缩放关闭，保留 opacity 反馈 */
.page.no-motion .p-chip:active,
.page.no-motion .irs:active,
.page.no-motion .stb:active,
.page.no-motion .money-done:active,
.page.no-motion .cal-nav:active,
.page.no-motion .cal-cell:active,
.page.no-motion .bt:active,
.page.no-motion .banner:active { transform: none; } /* 按压微缩放关闭，保留颜色/透明度反馈 */
@keyframes panelFadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes panelFadeOut { from { opacity: 1; } to { opacity: 0; } }
@keyframes spopFadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes spopFadeOut { from { opacity: 1; } to { opacity: 0; } }
</style>
