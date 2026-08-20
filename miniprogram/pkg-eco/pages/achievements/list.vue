<template>
  <view class="page">
    <!-- Search Bar -->
    <view class="sbar">
      <view class="sbox">
        <u-icon name="search" size="28rpx" color="#969799" />
        <input class="sinp" v-model="q" placeholder="搜索成果名称、关键词" placeholder-class="ph" @input="onSearch" />
        <text v-if="q" class="sclr" @tap="clearSearch">×</text>
      </view>

    </view>
    <view v-if="sV" class="mask" @tap="sV = false" />

    <!-- 筛选器：类型 / 领域 / 阶段（chip 选项 + 领域二级分组） -->
    <view class="filter-wrap">
      <view class="fbar">
      <view class="fpill" :class="{ on: panel === 'type' }" @tap="togglePanel('type')">
        <text class="fpv">{{ fType }}</text><text class="farr">▾</text>
      </view>
      <view class="fpill" :class="{ on: panel === 'field' }" @tap="togglePanel('field')">
        <text class="fpv">{{ fField }}</text><text class="farr">▾</text>
      </view>
      <view class="fpill" :class="{ on: panel === 'stage' }" @tap="togglePanel('stage')">
        <text class="fpv">{{ fStage }}</text><text class="farr">▾</text>
      </view>
      <view class="freset" @tap="resetFilters">重置</view>
      </view>
      <view v-if="panel" class="panel">
      <view class="p-head">{{ panelTitle }}</view>
      <view v-if="panel === 'field'" class="p-group">整机与硬件</view>
      <view v-if="panel === 'field'" class="p-chips">
        <text v-for="f in FIELD_HW" :key="f" class="p-chip" :class="{ act: fField === f }" @tap="pickField(f)">{{ f }}</text>
      </view>
      <view v-if="panel === 'field'" class="p-group">软件与服务</view>
      <view v-if="panel === 'field'" class="p-chips">
        <text v-for="f in FIELD_SW" :key="f" class="p-chip" :class="{ act: fField === f }" @tap="pickField(f)">{{ f }}</text>
      </view>
      <view v-else class="p-chips">
        <text v-for="o in chipOptions" :key="o.v" class="p-chip" :class="{ act: isAct(o.v) }" @tap="pickChip(o.v)">{{ o.l }}</text>
      </view>
      </view>
    </view>
    <view v-if="panel" class="panel-mask" @tap="panel = ''" />

    <!-- Banner Carousel -->
    <view class="carousel-wrap">
      <swiper
        class="carousel"
        circular
        autoplay
        :interval="3500"
        :duration="400"
        :indicator-dots="false"
        @change="onSlide"
      >
        <swiper-item v-for="(s, i) in slides" :key="i">
          <view class="cslide" :style="{ background: s.bg }">
            <view class="cs-ic">{{ s.ic }}</view>
            <view class="cs-info">
              <text class="cs-title">{{ s.title }}</text>
              <text class="cs-sub">{{ s.sub }}</text>
            </view>
          </view>
        </swiper-item>
      </swiper>
      <view class="cdots">
        <view v-for="(_, i) in slides" :key="i" class="cdot" :class="{ on: sIdx === i }" />
      </view>
    </view>

    <!-- Info Row -->
    <view class="ir">
      <text>共 <text class="irn">{{ shown }}</text> 项成果</text>
      <view class="irs-wrap">
        <text class="irs" @tap="toggleSort">{{ sortLabel }} ▼</text>
        <view v-if="sV" class="spop" @tap.stop>
          <view
            v-for="o in sorts"
            :key="o.k"
            class="sp-opt"
            :class="{ active: sort === o.k }"
            @tap="pickSort(o.k)"
          >
            <text v-if="sort === o.k" class="sp-chk">✓</text>
            <text>{{ o.l }}</text>
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

    <!-- Empty -->
    <view v-else-if="!list.length" class="st">
      <u-empty description="暂无成果">
        <text class="sth">试试调整筛选条件或搜索关键词</text>
        <view class="stb" @tap="resetAll">清除筛选</view>
      </u-empty>
    </view>

    <!-- Card Grid -->
    <view v-else class="cg">
      <view v-for="x in list" :key="x.id" class="card" @tap="goDetail(x)">
        <view class="cc" :style="{ background: bgOf(x.f) }">
          <image v-if="x.img" class="cc-img" :src="x.img" mode="aspectFill" />
          <text v-else class="cci">{{ icOf(x.f) }}</text>
          <text class="cct">{{ x.f }}</text>
          <text v-if="x.stage" class="ccs" :class="x.stageCls">{{ x.stageShort }}</text>
        </view>
        <view class="cbd">
          <text class="ct">{{ x.t }}</text>
          <view class="c-meta">
            <text class="c-tag">{{ x.tl }}</text>
            <text v-if="x.st" class="c-tag" :class="'tag-' + x.st.cls">{{ x.st.label }}</text>
          </view>
          <view class="cft">
            <text class="cf-date">{{ x.d }}</text>
            <text v-if="x.v > 0 || x.s > 0" class="cf-stats">
              <text v-if="x.v > 0" class="cf-eye">👁</text>
              <text v-if="x.v > 0">{{ fmt(x.v) }}</text>
              <text v-if="x.s > 0" class="cf-star">★</text>
              <text v-if="x.s > 0">{{ x.s }}</text>
            </text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="hasMore" class="lm">— 上拉加载更多 —</view>
    <view v-if="mockMode" class="mock-note">当前为演示数据 · 接口就绪后自动切换</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request, BASE_URL } from '@/utils/request'
import { MOCK_ACHIEVEMENTS, ACH_TYPE_LABEL, ACH_STATUS_LABEL, STAGE_LABEL, STAGE_SHORT, STAGE_RANK, FIELD_BG } from '@/utils/mockAchievements'

const PAGE_SIZE = 100
const MAX_PAGES = 10

const q = ref('')
const sort = ref('latest')
const sV = ref(false)
const loading = ref(true)
const err = ref(false)
const list = ref([])
const fullList = ref([])
const total = ref(0)
const mockMode = ref(false)
let nextPage = 1

// ---- 筛选状态 ----
const panel = ref('')
const fType = ref('全部类型')
const fField = ref('全部领域')
const fStage = ref('全部阶段')
const TYPE_OPTS = [
  { v: 'all', l: '全部类型' }, { v: 'patent', l: '发明专利' }, { v: 'utility', l: '实用新型' },
  { v: 'copyright', l: '软件著作' }, { v: 'paper', l: '论文成果' }, { v: 'standard', l: '技术标准' }, { v: 'design', l: '外观设计' },
]
const STAGE_OPTS = [
  { v: 'all', l: '全部阶段' }, { v: 'lab', l: '实验室' }, { v: 'pilot', l: '中试' },
  { v: 'industrialized', l: '产业化' }, { v: 'listed', l: '已上市' },
]
const FIELD_HW = ['飞控系统', '动力系统', '载荷设备', '地面站', '遥感测绘']
const FIELD_SW = ['AI算法', '集群协同', '通信链路', '标准规范']

const sorts = [
  { k: 'latest', l: '最新发布' },
  { k: 'views', l: '最多浏览' },
  { k: 'favs', l: '最多收藏' }
]
const SORT_LABEL = { latest: '最新发布', views: '最多浏览', favs: '最多收藏' }

const STATUS_MAP = { hot: '热门', transformed: '已转化', new: '新成果' }
const FIELD_ICON = { '飞控系统': '飞', '遥感测绘': '遥', '动力系统': '动', 'AI算法': '算', '载荷设备': '载', '集群协同': '群', '通信链路': '通', '标准规范': '标', '地面站': '地' }

const slides = [
  { ic: '智', title: 'AI 赋能飞控新突破', sub: '汇聚前沿科技成果，加速产学研对接', bg: '#0d47a1' },
  { ic: '创', title: '产学研协同创新', sub: '高校院所与企业共建创新生态', bg: '#1b5e20' },
  { ic: '标', title: '标准引领行业发展', sub: '技术标准与规范助力产业升级', bg: '#4a148c' }
]

const sIdx = ref(0)
const onSlide = (e) => { sIdx.value = e.detail.current }

const panelTitle = computed(() => ({ type: '请选择成果类型', field: '请选择研究领域', stage: '请选择转化阶段' }[panel.value] || ''))
const chipOptions = computed(() => (panel.value === 'type' ? TYPE_OPTS : STAGE_OPTS))
const isAct = (v) => (panel.value === 'type' ? (v === 'all' ? fType.value === '全部类型' : fType.value === v) : (v === 'all' ? fStage.value === '全部阶段' : fStage.value === v))

const shown = computed(() => list.value.length)
const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新发布')
const hasMore = computed(() => list.value.length < total.value)

const icOf = (f) => FIELD_ICON[f] || (f ? f.charAt(0) : '果')
const bgOf = (f) => FIELD_BG[f] || '#e8f2fc'
const fmt = (n) => (n >= 1e4 ? (n / 1e4).toFixed(1) + 'w' : n >= 1e3 ? (n / 1e3).toFixed(1) + 'k' : String(n))

const typeLabel = (t) => ACH_TYPE_LABEL[(t || '').toLowerCase()] || t || '其他'
const statusOf = (s) => {
  const key = (s || '').toLowerCase()
  const label = ACH_STATUS_LABEL[key]
  return label ? { label, cls: key } : null
}
const stageOf = (s) => {
  const key = (s || '').toLowerCase()
  const label = STAGE_LABEL[key]
  if (!label) return null
  return { label, short: STAGE_SHORT[key] || label.replace('阶段', ''), key, rank: STAGE_RANK[key] || 0 }
}
const stageClsOf = (s) => {
  const key = (s || '').toLowerCase()
  if (key === 'lab' || key === 'laboratory') return 'cl-la'
  if (key === 'pilot') return 'cl-pi'
  return 'cl-in'
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
  return {
    id: it.id,
    t: it.title || '',
    f: it.field || '其他',
    tl: typeLabel(it.achieve_type),
    d: (it.created_at || '').slice(0, 10),
    v: it.views || 0,
    s: it.favs || 0,
    st: statusOf(it.status),
    stage: st,
    stageShort: st ? st.short : '',
    stageCls: st ? stageClsOf(it.stage) : '',
    img: imgSrc(it.images),
  }
}

const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter(x => (x.t + ' ' + x.f + ' ' + x.tl).toLowerCase().includes(kw))
  if (fType.value !== '全部类型') items = items.filter(x => x.tl === fType.value)
  if (fField.value !== '全部领域') items = items.filter(x => x.f === fField.value)
  if (fStage.value !== '全部阶段') items = items.filter(x => x.stage && x.stage.short === fStage.value)
  if (sort.value === 'views') items.sort((a, b) => b.v - a.v)
  else if (sort.value === 'favs') items.sort((a, b) => b.s - a.s)
  else items.sort((a, b) => ((b.d || '') < (a.d || '') ? 1 : (b.d || '') > (a.d || '') ? -1 : 0))
  list.value = items
}

// ===== 数据获取 =====
// 接口替换点：GET /api/v1/achievements (field / page / page_size)；失败/为空时回退 mock
const fetchAll = async () => {
  loading.value = true
  err.value = false
  nextPage = 1
  try {
    const acc = []
    let fetched = 0
    for (let page = 1; page <= MAX_PAGES; page++) {
      const res = await request({ url: '/api/v1/achievements', data: { page, page_size: PAGE_SIZE } })
      const items = Array.isArray(res) ? res : (res?.items || [])
      acc.push(...items.map(mapItem))
      fetched = (Array.isArray(res) && res.total) || (res && res.total) || acc.length
      if (items.length < PAGE_SIZE || acc.length >= fetched) break
      nextPage++
    }
    if (!acc.length) { if (import.meta.env.DEV) { useMock() } } else {
      fullList.value = acc
      total.value = fetched
      mockMode.value = false
    }
    applyFilter()
  } catch {
    if (import.meta.env.DEV) { useMock() }
  } finally {
    loading.value = false
  }
}

const useMock = () => {
  fullList.value = MOCK_ACHIEVEMENTS.map(mapItem)
  total.value = fullList.value.length
  mockMode.value = true
  applyFilter()
}

const fetchMore = async () => {
  if (loading.value || nextPage >= MAX_PAGES || mockMode.value) return
  try {
    const res = await request({ url: '/api/v1/achievements', data: { page: nextPage, page_size: PAGE_SIZE } })
    const items = Array.isArray(res) ? res : (res?.items || [])
    fullList.value = fullList.value.concat(items.map(mapItem))
    total.value = (Array.isArray(res) && res.total) || (res && res.total) || fullList.value.length
    nextPage++
    applyFilter()
  } catch { /* 静默：下次触底重试 */ }
}

const onSearch = () => { applyFilter() }
const clearSearch = () => { q.value = ''; applyFilter() }
const toggleSort = () => { sV.value = !sV.value }
const pickSort = (k) => { sort.value = k; sV.value = false; applyFilter() }

// ---- 筛选交互 ----
const togglePanel = (k) => { panel.value = panel.value === k ? '' : k; sV.value = false }
const pickChip = (v) => {
  const opts = panel.value === 'type' ? TYPE_OPTS : STAGE_OPTS
  const o = opts.find(x => x.v === v)
  if (!o) return
  if (panel.value === 'type') fType.value = o.l
  else fStage.value = o.l
  panel.value = ''
  applyFilter()
}
const pickField = (f) => { fField.value = f; panel.value = ''; applyFilter() }
// 三个筛选器统一重置（fbar 右侧「重置」）
const resetFilters = () => {
  fType.value = '全部类型'
  fField.value = '全部领域'
  fStage.value = '全部阶段'
  panel.value = ''
  applyFilter()
}
const resetAll = () => {
  q.value = ''
  fType.value = '全部类型'
  fField.value = '全部领域'
  fStage.value = '全部阶段'
  sort.value = 'latest'
  sV.value = false
  applyFilter()
}

const goDetail = (x) => {
  uni.navigateTo({ url: '/pkg-eco/pages/achievements/detail?id=' + encodeURIComponent(x.id) })
}

onLoad(fetchAll)
onPullDownRefresh(async () => {
  await fetchAll()
  uni.stopPullDownRefresh()
})
onReachBottom(fetchMore)
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 40px;
}

/* ===== 搜索（保留现有风格） ===== */
.sbar {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 20rpx 28rpx 8rpx;
}
.sbox {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8rpx;
  height: 72rpx;
  background: var(--color-bg-card);
  border: 1rpx solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 0 20rpx;
}
.sinp { flex: 1; font-size: 26rpx; color: var(--color-text); }
.ph { color: var(--color-text-placeholder); }
.sclr { font-size: 30rpx; color: var(--color-text-placeholder); padding: 4rpx; }
.irs-wrap { position: relative; z-index: 60; }
.spop {
  position: absolute; top: calc(100% + 8rpx); right: 0; z-index: 50;
  background: var(--color-bg-card); border-radius: var(--radius-md);
  box-shadow: 0 4px 24px rgba(0,0,0,.12); padding: 12rpx 0; min-width: 264rpx;
  animation: popIn .18s ease;
}
@keyframes popIn { from { opacity: 0; transform: translateY(-8rpx); } to { opacity: 1; transform: translateY(0); } }
.sp-opt {
  padding: 20rpx 32rpx; font-size: 26rpx; color: var(--color-text);
  display: flex; align-items: center; gap: 16rpx; white-space: nowrap;
}
.sp-opt.active { color: var(--color-primary); font-weight: 600; }
.sp-opt:active { background: #f5f7fa; }
.sp-chk { font-size: 24rpx; }
.mask { position: fixed; inset: 0; z-index: 40; background: transparent; }

/* ===== 筛选器（新布局，沿用现有令牌） ===== */
.filter-wrap { position: relative; z-index: 45; }
.fbar { display: flex; gap: 16rpx; padding: 8rpx 28rpx 4rpx; }
.freset { flex: none; height: 60rpx; display: flex; align-items: center; padding: 0 12rpx; font-size: 24rpx; color: var(--color-primary); font-weight: 500; transition: transform .2s ease, opacity .2s ease; }
.freset:active { transform: scale(.95); opacity: .7; }
.fpill {
  flex: 1; height: 60rpx; display: flex; align-items: center; justify-content: center; gap: 6rpx;
  background: var(--color-bg-card); border: 1rpx solid var(--color-border);
  border-radius: var(--radius-md); font-size: 24rpx; color: var(--color-text-secondary);
}
.fpill.on { border-color: var(--color-primary); color: var(--color-primary); background: var(--color-primary-light); }
.farr { font-size: 18rpx; color: var(--color-text-placeholder); }
.panel-mask { position: fixed; inset: 0; z-index: 44; background: transparent; }
.panel {
  position: absolute; left: 28rpx; right: 28rpx; top: calc(100% + 8rpx); z-index: 45;
  background: var(--color-bg-card); border-radius: var(--radius-lg);
  box-shadow: 0 12rpx 36rpx rgba(0,0,0,.14); padding: 16rpx 20rpx 12rpx;
  animation: popIn .18s ease;
}
.p-head { font-size: 22rpx; color: var(--color-text-placeholder); padding: 4rpx 4rpx 12rpx; }
.p-group { font-size: 22rpx; color: var(--color-text-secondary); font-weight: 600; padding: 8rpx 4rpx 8rpx; }
.p-chips { display: flex; flex-wrap: wrap; gap: 14rpx; padding: 0 4rpx 10rpx; }
.p-chip {
  flex: none; padding: 10rpx 24rpx; border-radius: var(--radius-round);
  font-size: 24rpx; color: var(--color-text); background: var(--color-bg);
  border: 1rpx solid var(--color-border);
}
.p-chip.act { color: #fff; background: var(--color-primary); border-color: var(--color-primary); font-weight: 600; }

/* ===== Banner（保留现有风格） ===== */
.carousel-wrap { position: relative; margin: 16rpx 28rpx; }
.carousel { height: 328rpx; border-radius: 16rpx; overflow: hidden; }
.cslide { width: 100%; height: 100%; display: flex; align-items: center; padding: 0 40rpx; gap: 28rpx; }
.cs-ic { width: 88rpx; height: 88rpx; border-radius: 50%; background: rgba(255,255,255,.18); color: #fff; font-size: 44rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.cs-info { flex: 1; min-width: 0; }
.cs-title { font-size: 30rpx; font-weight: 600; color: #fff; margin-bottom: 8rpx; display: block; line-height: 1.3; }
.cs-sub { font-size: 24rpx; color: rgba(255,255,255,.72); display: block; }
.cdots { position: absolute; bottom: 20rpx; left: 50%; transform: translateX(-50%); display: flex; gap: 12rpx; }
.cdot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #fff; opacity: .35; transition: opacity .2s; }
.cdot.on { opacity: 1; }

/* ===== Info Row（保留） ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 4rpx 32rpx 16rpx; font-size: 24rpx; color: var(--color-text-secondary); }
.irn { color: var(--color-primary); font-weight: 600; }
.irs { color: var(--color-primary); font-weight: 500; }

/* ===== Card Grid（保留卡片框风格，仅调整内部结构） ===== */
.cg { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; padding: 0 28rpx 40rpx; }
.card { background: var(--color-bg-card); border-radius: 16rpx; overflow: hidden; border: .5px solid var(--color-divider); }
.card:active { transform: scale(.97); }
.cc { position: relative; aspect-ratio: 4/3; display: flex; align-items: center; justify-content: center; }
.cc-img { width: 100%; height: 100%; display: block; }
.cci { font-size: 60rpx; font-weight: 600; color: var(--color-text); opacity: .8; }
.cct { position: absolute; top: 12rpx; left: 12rpx; background: rgba(0,0,0,.45); color: #fff; font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 16rpx; font-weight: 500; }
/* 转化阶段徽章（右上角，沿用 .ccs 风格） */
.ccs { position: absolute; top: 12rpx; right: 12rpx; color: #fff; font-size: 20rpx; padding: 4rpx 14rpx; border-radius: 12rpx; font-weight: 500; }
.ccs.cl-la { background: #1967d2; }
.ccs.cl-pi { background: var(--color-warning); }
.ccs.cl-in { background: var(--color-success); }
.cbd { padding: 16rpx 20rpx 20rpx; }
.ct { font-size: 26rpx; font-weight: 600; color: var(--color-text); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 10rpx; }
.c-meta { display: flex; align-items: center; gap: 10rpx; margin-bottom: 10rpx; white-space: nowrap; overflow: hidden; }
.c-tag { flex: none; font-size: 19rpx; padding: 4rpx 14rpx; border-radius: 8rpx; font-weight: 500; color: var(--color-primary); background: var(--color-primary-light); }
.c-tag.tag-hot { color: var(--color-danger); background: rgba(255,59,48,.1); }
.c-tag.tag-transformed { color: var(--color-success); background: rgba(52,199,89,.12); }
.c-tag.tag-new { color: var(--color-primary); background: var(--color-primary-light); }
.cft { font-size: 20rpx; color: var(--color-text-placeholder); display: flex; justify-content: space-between; align-items: center; }
.cf-date { color: var(--color-text-placeholder); }
.cf-stats { color: var(--color-text-secondary); font-weight: 600; }
.cf-eye { font-size: 18rpx; margin-right: 4rpx; }
.cf-star { font-size: 18rpx; margin: 0 4rpx 0 10rpx; color: var(--color-text-placeholder); }

/* ===== Skeleton（保留） ===== */
.card-sk .sk-cv { aspect-ratio: 4/3; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.sk-bd { padding: 16rpx 20rpx; }
.sk-l { height: 24rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 12rpx; animation: shimmer 1.5s infinite; }
.sk-l.w90 { width: 90%; }
.sk-l.w60 { width: 60%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== State（保留） ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 120rpx 40rpx; }
.sth { font-size: 24rpx; color: var(--color-text-placeholder); margin: 24rpx 0; display: block; }
.stb { display: inline-block; padding: 16rpx 48rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 26rpx; font-weight: 500; }
.stb:active { opacity: .8; }

/* ===== Load More（保留） ===== */
.lm { text-align: center; padding: 24rpx; font-size: 24rpx; color: var(--color-text-placeholder); }

/* ===== 演示提示 ===== */
.mock-note { text-align: center; padding: 0 0 24rpx; font-size: 20rpx; color: var(--color-text-placeholder); }

/* ===================== UI/UX 体验优化（仅新增/修改 wxss，不动模板/数据/逻辑） ===================== */
/* 动画统一 200-400ms；优先 transform/opacity，避免重排；生产级轻量克制 */

/* 1) 入场动画：列表卡片依次淡入 + 轻微上移（40ms 错开；backwards 不阻塞点击态） */
.cg .card { animation: uiCardIn .3s ease backwards; }
.cg .card:nth-child(1) { animation-delay: 0ms; }
.cg .card:nth-child(2) { animation-delay: 40ms; }
.cg .card:nth-child(3) { animation-delay: 80ms; }
.cg .card:nth-child(4) { animation-delay: 120ms; }
.cg .card:nth-child(5) { animation-delay: 160ms; }
.cg .card:nth-child(6) { animation-delay: 200ms; }
.cg .card:nth-child(7) { animation-delay: 240ms; }
.cg .card:nth-child(8) { animation-delay: 280ms; }
.cg .card:nth-child(9) { animation-delay: 320ms; }
.cg .card:nth-child(10) { animation-delay: 360ms; }
.cg .card:nth-child(11) { animation-delay: 400ms; }
.cg .card:nth-child(12) { animation-delay: 440ms; }
@keyframes uiCardIn { from { opacity: 0; transform: translateY(12rpx); } to { opacity: 1; transform: translateY(0); } }

/* 2) 交互反馈：卡片/按钮按下轻微缩放 + 透明度（200ms；transform/opacity 不触发重排） */
.card { transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease; }
.card:active { transform: scale(.97); opacity: .88; }
.stb { transition: transform .2s ease, opacity .2s ease; }
.stb:active { transform: scale(.95); opacity: .85; }
.p-chip { transition: transform .2s ease, background .2s ease, color .2s ease, opacity .2s ease; }
.p-chip:active { transform: scale(.95); opacity: .85; }
.fpill { transition: transform .2s ease, background .2s ease, border-color .2s ease, color .2s ease; }
.fpill:active { transform: scale(.97); }
.irs { transition: transform .2s ease, background .2s ease; }
.irs:active { transform: scale(.96); }
.sclr { transition: transform .2s ease, opacity .2s ease; }
.sclr:active { transform: scale(.9); opacity: .7; }

/* 3) 渲染视觉细节：卡片轻阴影提升层次（原卡片已 overflow:hidden 防溢出，保持不变） */
.card { box-shadow: 0 2rpx 10rpx rgba(0,0,0,.04); }

/* 4) 层级加固：浮层/弹层 z-index 显式化，避免内容穿透 */
.panel { z-index: 45; }
.spop { z-index: 50; }
.fbar { z-index: 30; }

/* ===== 【首页风格】同步 pages/home 样式（仅颜色/圆角/阴影/字重；如需回退删除本块即可） ===== */
/* 搜索框：白底 + #E4E7EC 细边框 + 圆角 14rpx（对齐首页搜索按钮） */
.sbox { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 14rpx; box-shadow: none; }
/* 筛选 pill：对齐首页 city-option */
.fpill { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 12rpx; color: #344054; font-weight: 500; }
.fpill.on { color: #074D92; border-color: #0A66C2; background: #F4F8FC; font-weight: 700; }
/* 筛选 chip：对齐首页 badge */
.p-chip { border-radius: 8rpx; color: #344054; background: #F4F8FC; border: 1rpx solid #E4E7EC; font-weight: 500; }
.p-chip.act { color: #fff; background: #0A66C2; border-color: #0A66C2; font-weight: 600; }
/* 卡片：对齐首页 demand-card（圆角 8px=16rpx + 轻阴影） */
.card { border-radius: 16rpx; box-shadow: 0 4rpx 16rpx rgba(16,24,40,.035); }
/* 面板 / 排序弹层：对齐首页弹层阴影 */
.panel { border-radius: 16rpx; box-shadow: 0 6rpx 24rpx rgba(16,24,40,.08); }
.spop { box-shadow: 0 6rpx 24rpx rgba(16,24,40,.08); }
/* 标签徽章：对齐首页 type-badge / status-badge */
.c-tag { border-radius: 8rpx; font-weight: 700; color: #074D92; background: #EAF3FB; }
.c-tag.tag-hot { color: #fff; background: #F04438; }
.c-tag.tag-transformed { color: #168A55; background: #E9F7F0; }
.c-tag.tag-new { color: #074D92; background: #EAF3FB; }
.ccs { border-radius: 8rpx; font-weight: 700; }
/* 主按钮：对齐首页 retry-button */
.stb { background: #0A66C2; border-radius: 12rpx; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.3); }

/* ===== 【搜索/筛选尺寸优化】搜索框、筛选按钮、面板按钮加大、加宽松（样式级；可整体删除回退） ===== */
/* 搜索框：对齐首页搜索按钮比例（高 40px=80rpx / 浅面底 / 圆角 7px=14rpx / 内边距宽松） */
.sbar { padding: 24rpx 28rpx 16rpx; }
.sbox { height: 80rpx; padding: 0 28rpx; border-radius: 14rpx; background: #F4F8FC; border: 1rpx solid #E4E7EC; gap: 12rpx; }
.sinp { font-size: 28rpx; }
/* 筛选栏：更宽松 */
.fbar { gap: 20rpx; padding: 10rpx 28rpx 16rpx; }
.fpill { height: 72rpx; border-radius: 14rpx; font-size: 26rpx; gap: 8rpx; }
.freset { height: 72rpx; font-size: 26rpx; padding: 0 16rpx; }
/* 筛选面板：更大内边距、更宽松的选项 */
.panel { padding: 28rpx 28rpx 20rpx; border-radius: 20rpx; }
.p-head { font-size: 24rpx; padding: 4rpx 4rpx 16rpx; }
.p-group { font-size: 24rpx; padding: 12rpx 4rpx 10rpx; }
.p-chips { gap: 20rpx; padding: 4rpx 4rpx 14rpx; }
.p-chip { padding: 14rpx 32rpx; font-size: 26rpx; border-radius: 12rpx; }
/* 排序弹层：选项更宽松 */
.spop { min-width: 320rpx; padding: 16rpx 0; }
.sp-opt { padding: 26rpx 36rpx; font-size: 28rpx; }
</style>
