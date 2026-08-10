<template>
  <view class="page">
    <!-- Search Bar -->
    <view class="sbar">
      <view class="sbox">
        <u-icon name="search" size="28rpx" color="#969799" />
        <input class="sinp" v-model="q" placeholder="搜索成果名称、关键词" placeholder-class="ph" @input="onSearch" />
        <text v-if="q" class="sclr" @tap="clearSearch">×</text>
      </view>
      <view class="sbtn" @tap="toggleSort">
        <view class="sort-ic">
          <view class="sort-line" />
          <view class="sort-line" />
          <view class="sort-line" />
        </view>
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
    <view v-if="sV" class="mask" @tap="sV = false" />

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

    <!-- Func Nav -->
    <view class="fn">
      <view v-for="n in navs" :key="n.k" class="fi" @tap="onNav(n.k)">
        <view class="fii" :style="{ background: n.bg }"><text class="fie">{{ n.ic }}</text></view>
        <text class="fl">{{ n.label }}</text>
      </view>
    </view>

    <!-- Info Row -->
    <view class="ir">
      <text>共 <text class="irn">{{ shown }}</text> 项成果</text>
      <text class="irs" @tap="toggleSort">{{ sortLabel }} ▼</text>
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
    <view v-else-if="err" class="st">
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
          <text v-if="x.st" class="ccs" :class="x.st.cls">{{ x.st.label }}</text>
        </view>
        <view class="cbd">
          <text class="ct">{{ x.t }}</text>
          <text class="co">{{ x.tl }}</text>
          <view class="cft">
            <text>{{ x.d }}</text>
            <text v-if="x.v > 0" class="cfv">浏览 {{ fmt(x.v) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="hasMore" class="lm">— 上拉加载更多 —</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request, BASE_URL } from '@/utils/request'

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
let nextPage = 1

const sorts = [
  { k: 'latest', l: '最新发布' },
  { k: 'views', l: '最多浏览' },
  { k: 'favs', l: '最多收藏' }
]
const SORT_LABEL = { latest: '最新发布', views: '最多浏览', favs: '最多收藏' }

const TYPE_MAP = { patent: '发明专利', utility: '实用新型', copyright: '软件著作', paper: '论文成果', standard: '技术标准', design: '外观设计' }
const STATUS_MAP = { hot: '热门', transformed: '已转化', 'new': '新成果' }
const FIELD_ICON = { '飞控系统': '飞', '遥感测绘': '遥', '动力系统': '动', 'AI算法': '算', '载荷设备': '载', '集群协同': '群', '通信链路': '通', '标准规范': '标', '地面站': '地' }
const FIELD_BG = { '飞控系统': '#e3f2fd', '遥感测绘': '#e8f5e9', '动力系统': '#fff3e0', 'AI算法': '#f3e5f5', '载荷设备': '#fce4ec', '集群协同': '#e0f2f1', '通信链路': '#e8eaf6', '标准规范': '#f5f5f5', '地面站': '#fff8e1' }

// 轮播内容为通用宣传文案，不虚构统计数字
const slides = [
  { ic: '智', title: 'AI 赋能飞控新突破', sub: '汇聚前沿科技成果，加速产学研对接', bg: '#0d47a1' },
  { ic: '创', title: '产学研协同创新', sub: '高校院所与企业共建创新生态', bg: '#1b5e20' },
  { ic: '标', title: '标准引领行业发展', sub: '技术标准与规范助力产业升级', bg: '#4a148c' }
]

const navs = [
  { k: 'patent', ic: '发', label: '发明专利', bg: '#e3f2fd' },
  { k: 'utility', ic: '实', label: '实用新型', bg: '#fff3e0' },
  { k: 'copyright', ic: '软', label: '软件著作', bg: '#e8f5e9' },
  { k: 'paper', ic: '论', label: '论文成果', bg: '#f3e5f5' },
  { k: 'standard', ic: '标', label: '技术标准', bg: '#fce4ec' },
  { k: 'design', ic: '设', label: '外观设计', bg: '#e0f2f1' },
  { k: 'transformed', ic: '成', label: '已转化', bg: '#fff8e1' },
  { k: 'all', ic: '全', label: '全部成果', bg: '#e8eaf6' }
]

const sIdx = ref(0)
const onSlide = (e) => { sIdx.value = e.detail.current }

const shown = computed(() => list.value.length)
const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新发布')
const hasMore = computed(() => list.value.length < total.value)

const icOf = (f) => FIELD_ICON[f] || (f ? f.charAt(0) : '果')
const bgOf = (f) => FIELD_BG[f] || '#e8f2fc'
const fmt = (n) => (n >= 1e4 ? (n / 1e4).toFixed(1) + 'w' : n >= 1e3 ? (n / 1e3).toFixed(1) + 'k' : String(n))

const typeLabel = (t) => TYPE_MAP[(t || '').toLowerCase()] || t || '其他'
const statusOf = (s) => {
  const key = (s || '').toLowerCase()
  const label = STATUS_MAP[key]
  return label ? { label, cls: key } : null
}

// images 可能为字符串 JSON 或数组；相对路径拼接后端地址
const imgSrc = (images) => {
  let arr = images
  if (typeof images === 'string') {
    try { arr = JSON.parse(images) } catch { return '' }
  }
  if (!Array.isArray(arr) || !arr.length) return ''
  const u = arr[0]
  return u ? (u.startsWith('http') ? u : BASE_URL + u) : ''
}

// 后端返回字段映射为卡片模板字段（缺省字段优雅降级）
const mapItem = (it) => ({
  id: it.id,
  t: it.title || '',
  f: it.field || '其他',
  tl: typeLabel(it.achieve_type),
  d: (it.created_at || '').slice(0, 10),
  v: it.views || 0,
  s: it.favs || 0,
  st: statusOf(it.status),
  img: imgSrc(it.images)
})

// 客户端搜索/排序（与原型交互一致，作用于完整数据集）
const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) {
    items = items.filter(x => (x.t + ' ' + x.f + ' ' + x.tl).toLowerCase().includes(kw))
  }
  if (sort.value === 'views') items.sort((a, b) => b.v - a.v)
  else if (sort.value === 'favs') items.sort((a, b) => b.s - a.s)
  else items.sort((a, b) => ((b.d || '') < (a.d || '') ? 1 : (b.d || '') > (a.d || '') ? -1 : 0))
  list.value = items
}

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
    fullList.value = acc
    total.value = fetched
    applyFilter()
  } catch {
    err.value = true
  } finally {
    loading.value = false
  }
}

// 触底加载更多（仅当总数超过首轮拉取上限时出现）
const fetchMore = async () => {
  if (loading.value || nextPage >= MAX_PAGES) return
  try {
    const res = await request({ url: '/api/v1/achievements', data: { page: nextPage + 1, page_size: PAGE_SIZE } })
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
const onNav = (k) => {
  if (k === 'all') { resetAll(); return }
  const map = { patent: '发明专利', utility: '实用新型', copyright: '软件著作', paper: '论文成果', standard: '技术标准', design: '外观设计', transformed: '已转化' }
  q.value = map[k] || ''
  applyFilter()
}
const resetAll = () => { q.value = ''; sort.value = 'latest'; sV.value = false; applyFilter() }

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
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ===== Search Bar ===== */
.sbar { display: flex; align-items: center; gap: 20rpx; padding: 24rpx 28rpx; background: #fff; }
.sbox { flex: 1; display: flex; align-items: center; background: #f0f1f3; border-radius: 16rpx; padding: 20rpx 28rpx; gap: 16rpx; }
.sinp { flex: 1; border: none; outline: none; background: transparent; font-size: 28rpx; color: var(--color-text); min-width: 0; height: 40rpx; line-height: 40rpx; }
.ph { color: #bbb; }
.sclr { color: #bbb; font-size: 32rpx; padding: 4rpx; flex-shrink: 0; }

/* ===== Sort Button + Popup ===== */
.sbtn { width: 76rpx; height: 76rpx; border-radius: 50%; background: #f0f1f3; display: flex; align-items: center; justify-content: center; position: relative; flex-shrink: 0; }
.sbtn:active { transform: scale(.93); }
.sort-ic { display: flex; flex-direction: column; gap: 6rpx; }
.sort-line { width: 26rpx; height: 3rpx; border-radius: 2rpx; background: #666; }
.spop { position: absolute; top: 88rpx; right: -8rpx; z-index: 50; background: #fff; border-radius: 16rpx; box-shadow: 0 4px 24px rgba(0,0,0,.12); padding: 12rpx 0; min-width: 264rpx; animation: popIn .18s ease; }
@keyframes popIn { from { opacity: 0; transform: translateY(-8rpx); } to { opacity: 1; transform: translateY(0); } }
.sp-opt { padding: 20rpx 32rpx; font-size: 26rpx; color: var(--color-text); display: flex; align-items: center; gap: 16rpx; white-space: nowrap; }
.sp-opt.active { color: var(--color-primary); font-weight: 600; }
.sp-opt:active { background: #f5f7fa; }
.sp-chk { font-size: 24rpx; }
.mask { position: fixed; inset: 0; z-index: 40; background: transparent; }

/* ===== Banner Carousel ===== */
.carousel-wrap { position: relative; margin: 24rpx 28rpx; }
.carousel { height: 328rpx; border-radius: 16rpx; overflow: hidden; }
.cslide { width: 100%; height: 100%; display: flex; align-items: center; padding: 0 40rpx; gap: 28rpx; }
.cs-ic { width: 88rpx; height: 88rpx; border-radius: 50%; background: rgba(255,255,255,.18); color: #fff; font-size: 44rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.cs-info { flex: 1; min-width: 0; }
.cs-title { font-size: 30rpx; font-weight: 600; color: #fff; margin-bottom: 8rpx; display: block; line-height: 1.3; }
.cs-sub { font-size: 24rpx; color: rgba(255,255,255,.72); display: block; }
.cdots { position: absolute; bottom: 20rpx; left: 50%; transform: translateX(-50%); display: flex; gap: 12rpx; }
.cdot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #fff; opacity: .35; transition: opacity .2s; }
.cdot.on { opacity: 1; }

/* ===== Func Nav ===== */
.fn { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8rpx; padding: 0 28rpx 24rpx; }
.fi { text-align: center; padding: 16rpx 8rpx; }
.fi:active { transform: scale(.93); }
.fii { width: 88rpx; height: 88rpx; border-radius: 16rpx; display: flex; align-items: center; justify-content: center; margin: 0 auto 12rpx; }
.fie { font-size: 40rpx; font-weight: 600; color: var(--color-text); }
.fl { font-size: 22rpx; color: var(--color-text-secondary); }

/* ===== Info Row ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 4rpx 32rpx 16rpx; font-size: 24rpx; color: var(--color-text-secondary); }
.irn { color: var(--color-primary); font-weight: 600; }
.irs { color: var(--color-primary); font-weight: 500; }

/* ===== Card Grid ===== */
.cg { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; padding: 0 28rpx 40rpx; }
.card { background: #fff; border-radius: 16rpx; overflow: hidden; border: .5px solid var(--color-divider); }
.card:active { transform: scale(.97); }
.cc { position: relative; aspect-ratio: 4/3; display: flex; align-items: center; justify-content: center; }
.cc-img { width: 100%; height: 100%; display: block; }
.cci { font-size: 60rpx; font-weight: 600; color: var(--color-text); opacity: .8; }
.cct { position: absolute; top: 12rpx; left: 12rpx; background: rgba(0,0,0,.45); color: #fff; font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 16rpx; font-weight: 500; }
.ccs { position: absolute; top: 12rpx; right: 12rpx; color: #fff; font-size: 20rpx; padding: 4rpx 14rpx; border-radius: 12rpx; font-weight: 500; }
.ccs.hot { background: var(--color-danger); }
.ccs.transformed { background: var(--color-success); }
.ccs.new { background: var(--color-primary); }
.cbd { padding: 16rpx 20rpx 20rpx; }
.ct { font-size: 26rpx; font-weight: 600; color: var(--color-text); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 8rpx; }
.co { font-size: 22rpx; color: var(--color-text-secondary); margin-bottom: 8rpx; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cft { font-size: 20rpx; color: #bbb; display: flex; justify-content: space-between; align-items: center; }
.cfv { color: var(--color-text-secondary); }

/* ===== Skeleton ===== */
.card-sk .sk-cv { aspect-ratio: 4/3; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.sk-bd { padding: 16rpx 20rpx; }
.sk-l { height: 24rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 12rpx; animation: shimmer 1.5s infinite; }
.sk-l.w90 { width: 90%; }
.sk-l.w60 { width: 60%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== State ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 120rpx 40rpx; }
.sth { font-size: 24rpx; color: #ccc; margin: 24rpx 0; display: block; }
.stb { display: inline-block; padding: 16rpx 48rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 26rpx; font-weight: 500; }
.stb:active { opacity: .8; }

/* ===== Load More ===== */
.lm { text-align: center; padding: 24rpx; font-size: 24rpx; color: #ccc; }
</style>
