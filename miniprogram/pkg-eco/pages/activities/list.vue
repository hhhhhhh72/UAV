<template>
  <view class="act-page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <!-- 顶部导航 -->
    <u-nav-bar title="协会活动" show-back :fixed="true" @back="goBack" />

    <!-- 搜索 -->
    <view class="sbar">
      <view class="sbox">
        <u-icon name="search" size="28rpx" color="#98A2B3" />
        <input
          class="sinp"
          v-model="q"
          placeholder="搜索活动名称、关键词"
          placeholder-class="ph"
          confirm-type="search"
          @input="onSearch"
        />
        <view v-if="q" class="sclr" @tap="clearSearch">×</view>
      </view>
    </view>

    <!-- 三个筛选器 -->
    <view class="filter-bar">
      <view class="fpill" :class="{ on: showType }" @tap="openType">
        <text class="fval">{{ typeLabel }}</text><text class="farr">▾</text>
      </view>
      <view class="fpill" :class="{ on: showStatus }" @tap="openStatus">
        <text class="fval">{{ statusLabel }}</text><text class="farr">▾</text>
      </view>
      <view class="fpill" :class="{ on: showTime }" @tap="openTime">
        <text class="fval">{{ timeLabel }}</text><text class="farr">▾</text>
      </view>
    </view>

    <!-- 信息行：共 N 项 + 排序 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 项活动</text>
      <view class="irs-wrap">
        <text class="irs" @tap="toggleSort">{{ sortLabel }} ▾</text>
        <!-- 排序面板 -->
        <view v-if="showSort" class="spop" @tap.stop>
          <view
            v-for="s in SORTS"
            :key="s.v"
            class="sp-opt"
            :class="{ act: sort === s.v }"
            @tap="pickSort(s.v)"
          >
            <text>{{ s.l }}</text><text v-if="sort === s.v" class="chk">✓</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 列表 -->
    <StateView
      :loading="loading"
      :error="!!err"
      :empty="!loading && !err && !list.length"
      empty-text="暂无活动"
      @retry="fetchAll"
    >
      <view class="list-body">
        <view
          v-for="x in list"
          :key="x.id"
          class="site-card"
          hover-class="tap-fade"
          @tap="goDetail(x)"
        >
          <view class="card-thumb" :class="'thumb--' + x.catKey">
            <text class="thumb-char">{{ x.char }}</text>
          </view>
          <view class="card-body">
            <view class="card-top">
              <text class="card-name">{{ x.t }}</text>
              <text class="card-status" :class="'status--' + x.status">{{ x.stLabel }}</text>
            </view>
            <text class="meta-line">{{ x.timeText }}</text>
            <text class="meta-line">{{ x.loc }}</text>
            <view class="card-bottom">
              <view class="quota-row">
                <text class="quota-num">{{ x.reg }}</text>
                <text class="quota-unit">{{ x.total ? '/' + x.total + ' 已报名' : '人已报名' }}</text>
              </view>
              <text class="card-type">{{ x.catLabel }}</text>
            </view>
          </view>
        </view>
      </view>
    </StateView>

    <view v-if="hasMore" class="lm">— 上拉加载更多 —</view>

    <!-- ═══ 类型：二级表格面板 ═══ -->
    <view v-if="showType" class="sub2" :style="{ top: panelTop }" @tap.stop>
      <scroll-view scroll-y class="s2-col s2-left">
        <view
          v-for="g in GROUPS"
          :key="g.key"
          class="s2-lopt"
          :class="{ act: activeGroup === g.key }"
          @tap="pickGroup(g.key)"
        >{{ g.label }}</view>
      </scroll-view>
      <scroll-view scroll-y class="s2-col s2-right">
        <view class="s2-rhead">请选择二级分类</view>
        <view
          v-for="s in subItems"
          :key="s"
          class="s2-ropt"
          :class="{ act: subSel === s }"
          @tap="pickSub(s)"
        >
          <text>{{ s }}</text><text v-if="subSel === s" class="chk">✓</text>
        </view>
      </scroll-view>
    </view>

    <!-- ═══ 状态：横排按钮 ═══ -->
    <view v-if="showStatus" class="stpanel" :style="{ top: panelTop }" @tap.stop>
      <view class="strow">
        <view
          v-for="s in STATUS_OPTS"
          :key="s.v"
          class="stb"
          :class="{ act: activeStatus === s.v }"
          @tap="pickStatus(s.v)"
        >{{ s.l }}</view>
      </view>
    </view>

    <!-- ═══ 时间：快捷按钮 + 区间日历 ═══ -->
    <view v-if="showTime" class="calpanel" :style="{ top: panelTop }" @tap.stop>
      <view class="cp-quick">
        <view
          v-for="qk in QUICK_OPTS"
          :key="qk.v"
          class="cq"
          :class="{ act: quick === qk.v }"
          @tap="pickQuick(qk.v)"
        >{{ qk.l }}</view>
      </view>
      <view class="cal-head">
        <view class="cal-nav" @tap="calShift(-1)">‹</view>
        <text class="cal-title">{{ calYear }}年{{ calMonth }}月</text>
        <view class="cal-nav" @tap="calShift(1)">›</view>
      </view>
      <view class="cal-week">
        <text v-for="w in WEEK" :key="w">{{ w }}</text>
      </view>
      <view class="cal-cells">
        <view
          v-for="(c, i) in calCells"
          :key="i"
          class="cal-cell"
          :class="c.cls"
          @tap="c.empty ? null : onCalClick(c.key)"
        >{{ c.empty ? '' : c.day }}</view>
      </view>
      <view class="cal-foot">
        <text class="cal-tip">{{ calTip }}</text>
        <text class="cal-reset" @tap="resetTime">重置</text>
      </view>
    </view>

    <!-- 遮罩 -->
    <view v-if="panelOpen" class="mask" @tap="closeAll" />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import StateView from '@/components/StateView.vue'
import { dateOf, timeOf } from '@/utils/eventTime'

/* ===== 静态配置 ===== */
const PAGE_SIZE = 100
const MAX_PAGES = 10

// 一级分组（活动类型）
const GROUPS = [
  { key: 'all', label: '全部活动' },
  { key: 'forum', label: '行业论坛' },
  { key: 'salon', label: '主题沙龙' },
  { key: 'visit', label: '企业考察' },
  { key: 'expo', label: '行业展会' },
]
// 二级子项（运营可配置）
const SUB_DATA = {
  all: ['全部活动'],
  forum: ['产业论坛', '技术论坛', '政策宣讲', '供需对接会'],
  salon: ['闭门沙龙', '开放沙龙'],
  visit: ['渝北站', '两江新区站', '市外考察'],
  expo: ['博览会', '组团参展'],
}
const TYPE_CHAR = { forum: '论', salon: '沙', visit: '察', expo: '展', all: '活' }
const TYPE_LABEL = { forum: '行业论坛', salon: '主题沙龙', visit: '企业考察', expo: '行业展会' }
const STATUS_LABEL = { open: '报名中', soon: '即将开始', end: '已结束' }
const STATUS_OPTS = [
  { v: 'all', l: '全部状态' },
  { v: 'open', l: '报名中' },
  { v: 'soon', l: '即将开始' },
  { v: 'end', l: '已结束' },
]
const SORTS = [
  { v: 'latest', l: '最新发布' },
  { v: 'soon', l: '即将开始' },
  { v: 'regs', l: '报名最多' },
]
const QUICK_OPTS = [
  { v: 'week', l: '一周内' },
  { v: 'month', l: '30天内' },
  { v: 'all', l: '全部' },
]
const WEEK = ['一', '二', '三', '四', '五', '六', '日']

/* ===== 状态 ===== */
const q = ref('')
const activeGroup = ref('all')
const subSel = ref('全部活动')
const activeStatus = ref('all')
const quick = ref('all')
const sort = ref('latest')
const showType = ref(false)
const showStatus = ref(false)
const showTime = ref(false)
const showSort = ref(false)
const loading = ref(true)
const err = ref(false)
const list = ref([])
const fullList = ref([])
const total = ref(0)
const statusBarHeight = ref(20)

// 日历状态
const calYear = ref(2026)
const calMonth = ref(8)
const rangeStart = ref('')
const rangeEnd = ref('')

let nextPage = 1

const panelOpen = computed(() => showType.value || showStatus.value || showTime.value)
const panelTop = computed(() => (statusBarHeight.value + 152) + 'px')
const subItems = computed(() => SUB_DATA[activeGroup.value] || ['全部活动'])
const typeLabel = computed(() => (subSel.value === '全部活动' ? '全部类型' : subSel.value))
const statusLabel = computed(() => {
  const o = STATUS_OPTS.find((s) => s.v === activeStatus.value)
  return o ? o.l : '全部状态'
})
const timeLabel = computed(() => {
  if (rangeStart.value && rangeEnd.value) return fmt(rangeStart.value) + ' ~ ' + fmt(rangeEnd.value)
  if (rangeStart.value) return fmt(rangeStart.value) + ' 起'
  const o = QUICK_OPTS.find((s) => s.v === quick.value)
  return o ? (o.v === 'all' ? '全部时间' : o.l) : '全部时间'
})
const sortLabel = computed(() => {
  const o = SORTS.find((s) => s.v === sort.value)
  return o ? o.l : '最新发布'
})
const calTip = computed(() => {
  if (rangeStart.value && rangeEnd.value) return '已选择 ' + fmt(rangeStart.value) + ' ~ ' + fmt(rangeEnd.value)
  if (rangeStart.value) return '已选开始 ' + fmt(rangeStart.value) + '，请选择结束日期'
  return '点击选择开始日期，再点结束日期'
})
const hasMore = computed(() => list.value.length < total.value)

/* ===== 日历生成 ===== */
const calCells = computed(() => {
  const cells = []
  const first = new Date(calYear.value, calMonth.value - 1, 1)
  const startDow = (first.getDay() + 6) % 7 // 周一=0
  const days = new Date(calYear.value, calMonth.value, 0).getDate()
  const now = new Date()
  const todayKey = now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate())
  for (let i = 0; i < startDow; i++) cells.push({ empty: true, cls: ['empty'], key: '', day: '' })
  for (let d = 1; d <= days; d++) {
    const key = calYear.value + '-' + pad(calMonth.value) + '-' + pad(d)
    const cls = []
    if (key === todayKey) cls.push('today')
    if (rangeStart.value && key === rangeStart.value) cls.push('sel')
    if (rangeEnd.value && key === rangeEnd.value) cls.push('sel')
    if (rangeStart.value && rangeEnd.value && key > rangeStart.value && key < rangeEnd.value) cls.push('range')
    cells.push({ empty: false, cls, key, day: d })
  }
  return cells
})

/* ===== 数据映射（后端字段优雅降级） ===== */
const keyOf = (c) => {
  const s = String(c || '').toLowerCase()
  if (TYPE_LABEL[s]) return s
  // 兼容中文值
  if (s.includes('论坛')) return 'forum'
  if (s.includes('沙龙')) return 'salon'
  if (s.includes('考察')) return 'visit'
  if (s.includes('展')) return 'expo'
  return 'forum'
}
const statusOf = (it) => {
  const s = String(it.status || '').toLowerCase()
  if (s === 'open' || s === '报名中' || s === 'published') return 'open'
  if (s === 'soon' || s === '即将开始' || s === 'upcoming') return 'soon'
  if (s === 'end' || s === '已结束' || s === 'finished' || s === 'closed') return 'end'
  // 兜底：按日期推导
  const d = dateOf(it.start_date || it.event_date || it.start_time || '')
  if (d && d < today()) return 'end'
  return 'open'
}
const mapItem = (it) => {
  const catKey = keyOf(it.category || it.activity_type || it.field || it.event_type)
  const rawTime = it.start_date || it.event_date || it.start_time || ''
  const date = dateOf(rawTime)
  const timeTxt = it.time_text || it.time || timeOf(rawTime)
  return {
    id: it.id,
    t: it.title || '',
    catKey,
    char: TYPE_CHAR[catKey] || '活',
    catLabel: TYPE_LABEL[catKey] || it.category || '协会活动',
    status: statusOf(it),
    stLabel: STATUS_LABEL[statusOf(it)] || '报名中',
    date,
    timeText: (date ? fmt(date) + (timeTxt ? ' ' + timeTxt : '') : timeTxt) || '时间待定',
    loc: it.location || it.address || '地点待定',
    reg: it.reg_count ?? it.registered ?? 0,
    total: it.quota || it.capacity || it.max_attendees || 0,
    raw: it, // 原始后端对象，透传给详情页（详情接口未部署时兜底）
  }
}

/* ===== 工具 ===== */
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
/* ===== 过滤与排序（全量数据 + 前端过滤，后端未就绪亦可演示） ===== */
const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter((x) => (x.t + ' ' + x.catLabel).toLowerCase().includes(kw))
  if (activeGroup.value !== 'all') items = items.filter((x) => x.catKey === activeGroup.value)
  // 二级分类过滤：后端无独立子类字段，对原始数据中的分类/主题/标题做包含匹配
  const sub = subSel.value
  if (sub && sub !== '全部活动') {
    items = items.filter((x) => {
      const raw = x.raw || {}
      const blob = [
        x.t, x.catLabel,
        raw.category, raw.event_type, raw.activity_type,
        raw.theme, raw.topic, raw.sub_category, raw.field, raw.location,
      ].filter(Boolean).join(' ')
      return blob.indexOf(sub) >= 0
    })
  }
  if (activeStatus.value !== 'all') items = items.filter((x) => x.status === activeStatus.value)
  // 时间过滤
  if (quick.value === 'week') {
    const from = today(), to = addDays(7)
    items = items.filter((x) => x.date >= from && x.date <= to)
  } else if (quick.value === 'month') {
    const from = today(), to = addDays(30)
    items = items.filter((x) => x.date >= from && x.date <= to)
  } else if (rangeStart.value && rangeEnd.value) {
    items = items.filter((x) => x.date >= rangeStart.value && x.date <= rangeEnd.value)
  }
  // 排序
  if (sort.value === 'regs') items.sort((a, b) => b.reg - a.reg)
  else if (sort.value === 'soon') {
    items.sort((a, b) => (a.status === 'end' ? 1 : 0) - (b.status === 'end' ? 1 : 0) || (a.date > b.date ? 1 : -1))
  } else {
    items.sort((a, b) => (b.date > a.date ? 1 : -1))
  }
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
      const res = await request({ url: '/api/v1/events', data: { page, page_size: PAGE_SIZE } })
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

const fetchMore = async () => {
  if (loading.value || nextPage >= MAX_PAGES) return
  try {
    const res = await request({ url: '/api/v1/events', data: { page: nextPage + 1, page_size: PAGE_SIZE } })
    const items = Array.isArray(res) ? res : (res?.items || [])
    fullList.value = fullList.value.concat(items.map(mapItem))
    total.value = (Array.isArray(res) && res.total) || (res && res.total) || fullList.value.length
    nextPage++
    applyFilter()
  } catch { /* 静默 */ }
}

/* ===== 筛选器交互 ===== */
const closeAll = () => {
  showType.value = false
  showStatus.value = false
  showTime.value = false
  showSort.value = false
}
const openType = () => { closeAll(); showType.value = true }
const openStatus = () => { closeAll(); showStatus.value = true }
const openTime = () => { closeAll(); showTime.value = true }
const toggleSort = () => { closeAll(); showSort.value = !showSort.value }

const pickGroup = (k) => {
  activeGroup.value = k
  subSel.value = SUB_DATA[k][0] || '全部活动'
}
const pickSub = (s) => {
  subSel.value = s
  closeAll()
  applyFilter()
}
const pickStatus = (v) => { activeStatus.value = v; closeAll(); applyFilter() }
const pickSort = (v) => { sort.value = v; closeAll(); applyFilter() }

const pickQuick = (v) => {
  quick.value = v
  rangeStart.value = ''
  rangeEnd.value = ''
  closeAll()
  applyFilter()
}
const calShift = (d) => {
  calMonth.value += d
  if (calMonth.value < 1) { calMonth.value = 12; calYear.value-- }
  if (calMonth.value > 12) { calMonth.value = 1; calYear.value++ }
}
const onCalClick = (key) => {
  if (!rangeStart.value) {
    rangeStart.value = key
    rangeEnd.value = ''
  } else if (!rangeEnd.value) {
    if (key < rangeStart.value) rangeStart.value = key
    else rangeEnd.value = key
  } else {
    rangeStart.value = key
    rangeEnd.value = ''
  }
  if (rangeStart.value && rangeEnd.value) {
    closeAll()
    applyFilter()
  }
}
const resetTime = () => {
  rangeStart.value = ''
  rangeEnd.value = ''
  quick.value = 'all'
  applyFilter()
}

let searchTimer = null
const onSearch = () => {
  // 250ms 防抖：连续输入只做一次全量重筛
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { applyFilter() }, 250)
}
const clearSearch = () => { q.value = ''; applyFilter() }
const goDetail = (x) => {
  try { uni.setStorageSync('act_detail_' + x.id, x.raw || x) } catch (e) { /* 存储失败不影响跳转 */ }
  uni.navigateTo({ url: '/pkg-eco/pages/activities/detail?id=' + encodeURIComponent(x.id) })
}
const goBack = () => uni.navigateBack()

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  fetchAll()
})
onPullDownRefresh(async () => {
  await fetchAll()
  uni.stopPullDownRefresh()
})
onReachBottom(fetchMore)
</script>

<style scoped>
.act-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== 搜索 ===== */
.sbar { display: flex; align-items: center; padding: 12px 12px 10px; background: #fff; }
.sbox { flex: 1; display: flex; align-items: center; background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 8px; padding: 10px 12px; gap: 8px; }
.sinp { flex: 1; border: none; outline: none; background: transparent; font-size: 14px; color: #17212B; min-width: 0; height: 20px; line-height: 20px; }
.ph { color: #bbb; }
.sclr { color: #98A2B3; font-size: 16px; padding: 2px 4px; flex-shrink: 0; }

/* ===== 筛选器 ===== */
.filter-bar { background: #fff; padding: 0 12px 10px; border-bottom: 1px solid #EEF1F4; display: flex; gap: 8px; }
.fpill {
  flex: 1; height: 34px; display: flex; align-items: center; justify-content: center; gap: 5px;
  background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 8px;
  font-size: 12.5px; color: #344054; position: relative; min-width: 0;
}
.fpill:active { transform: scale(.97); }
.fpill.on { border-color: #0A66C2; color: #0A66C2; background: #EAF3FB; }
.fval { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.farr { font-size: 9px; color: #98A2B3; flex: none; }
.fpill.on .farr { color: #0A66C2; }

/* ===== 信息行 ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 10px 16px 8px; font-size: 12px; color: #667085; position: relative; z-index: 20; }
.irn { color: #0A66C2; font-weight: 600; }
.irs-wrap { position: relative; }
.irs { color: #0A66C2; font-weight: 500; padding: 4px 8px; border-radius: 6px; }
.irs:active { background: #EAF3FB; }
.spop {
  position: absolute; top: 32px; right: 0; z-index: 90;
  background: #fff; border-radius: 10px; box-shadow: 0 8px 28px rgba(0,0,0,.14);
  padding: 6px; min-width: 132px; animation: popIn .16s ease;
}
@keyframes popIn { from { opacity: 0; transform: translateY(-6px); } to { opacity: 1; transform: translateY(0); } }
.sp-opt { padding: 11px 12px; border-radius: 7px; font-size: 13px; color: #17212B; display: flex; align-items: center; justify-content: space-between; }
.sp-opt:active { background: #F4F6F8; }
.sp-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }
.chk { color: #0A66C2; font-size: 12px; }

/* ===== 列表卡片 ===== */
.list-body { padding: 4px 12px 12px; }
.site-card { display: flex; gap: 12px; background: #fff; border: 1px solid #EEF1F4; border-radius: 8px; padding: 12px; margin-bottom: 8px; }
.tap-fade { opacity: .7; }
.card-thumb { width: 84px; height: 84px; border-radius: 8px; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.thumb-char { font-size: 30px; font-weight: 700; }
.thumb--forum { background: #EAF3FB; } .thumb--forum .thumb-char { color: #0A66C2; }
.thumb--visit { background: #E9F7F0; } .thumb--visit .thumb-char { color: #168A55; }
.thumb--salon { background: #FFF0E6; } .thumb--salon .thumb-char { color: #E96012; }
.thumb--expo { background: #F6F4FF; } .thumb--expo .thumb-char { color: #7A5AF8; }
.card-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6px; }
.card-top { display: flex; align-items: center; gap: 6px; }
.card-name { flex: 1; min-width: 0; font-size: 15px; font-weight: 600; color: #17212B; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-status { flex-shrink: 0; font-size: 10px; padding: 2px 8px; border-radius: 4px; font-weight: 500; }
.status--open { color: #0A66C2; background: #EAF3FB; }
.status--soon { color: #E96012; background: #FFF0E6; }
.status--end { color: #667085; background: #F2F4F7; }
.meta-line { font-size: 12px; color: #667085; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: block; }
.card-bottom { display: flex; align-items: baseline; justify-content: space-between; margin-top: auto; }
.quota-row { display: flex; align-items: baseline; gap: 3px; }
.quota-num { font-size: 15px; font-weight: 700; color: #0A66C2; }
.quota-unit { font-size: 10px; color: #98A2B3; }
.card-type { font-size: 11px; color: #98A2B3; }
.lm { text-align: center; padding: 12px; font-size: 12px; color: #ccc; }

/* ===== 二级表格面板 ===== */
.sub2 {
  position: fixed; left: 12px; right: 12px; z-index: 80;
  background: #fff; border-radius: 12px; overflow: hidden;
  box-shadow: 0 12px 36px rgba(0,0,0,.16);
  display: flex; height: 236px; animation: panelIn .24s cubic-bezier(.2,.9,.3,1);
}
@keyframes panelIn { from { opacity: 0; transform: translateY(-10px) scale(.985); } to { opacity: 1; transform: translateY(0) scale(1); } }
.s2-col { height: 100%; }
.s2-left { width: 118px; background: #F7F9FB; border-right: 1px solid #EEF1F4; flex: none; }
.s2-lopt { padding: 13px 14px; font-size: 13px; color: #667085; position: relative; }
.s2-lopt.act { color: #0A66C2; font-weight: 600; background: #fff; }
.s2-lopt.act::before { content: ''; position: absolute; left: 0; top: 50%; transform: translateY(-50%); width: 3px; height: 16px; background: #0A66C2; border-radius: 0 2px 2px 0; }
.s2-right { flex: 1; background: #fff; padding: 6px; }
.s2-rhead { padding: 10px 14px 4px; font-size: 11px; color: #98A2B3; }
.s2-ropt { padding: 11px 14px; border-radius: 7px; font-size: 13px; color: #17212B; display: flex; align-items: center; justify-content: space-between; }
.s2-ropt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }

/* ===== 状态横排面板 ===== */
.stpanel {
  position: fixed; left: 12px; right: 12px; z-index: 80;
  background: #fff; border-radius: 12px; box-shadow: 0 12px 36px rgba(0,0,0,.16);
  padding: 12px 14px; animation: panelIn .24s cubic-bezier(.2,.9,.3,1);
}
.strow { display: flex; gap: 8px; }
.stb { flex: 1; height: 36px; display: flex; align-items: center; justify-content: center; background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 8px; font-size: 12.5px; color: #344054; }
.stb.act { background: #0A66C2; border-color: #0A66C2; color: #fff; font-weight: 600; }

/* ===== 时间面板 ===== */
.calpanel {
  position: fixed; left: 12px; right: 12px; z-index: 80;
  background: #fff; border-radius: 12px; overflow: hidden;
  box-shadow: 0 12px 36px rgba(0,0,0,.16);
  padding: 14px 14px 18px; animation: panelIn .24s cubic-bezier(.2,.9,.3,1);
}
.cp-quick { display: flex; gap: 8px; margin-bottom: 14px; }
.cq { flex: 1; height: 34px; display: flex; align-items: center; justify-content: center; background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 8px; font-size: 12.5px; color: #344054; }
.cq.act { background: #0A66C2; border-color: #0A66C2; color: #fff; font-weight: 600; }
.cal-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.cal-title { font-size: 14px; font-weight: 600; color: #17212B; }
.cal-nav { width: 30px; height: 30px; border-radius: 50%; display: flex; align-items: center; justify-content: center; background: #F4F6F8; color: #667085; font-size: 15px; }
.cal-week, .cal-cells { display: grid; grid-template-columns: repeat(7, 1fr); text-align: center; }
.cal-week { font-size: 11px; color: #98A2B3; margin-bottom: 6px; }
.cal-week text { padding: 6px 0; }
.cal-cells { gap: 2px; }
.cal-cell { height: 36px; display: flex; align-items: center; justify-content: center; font-size: 13px; color: #17212B; border-radius: 8px; }
.cal-cell.empty { pointer-events: none; }
.cal-cell.today { box-shadow: inset 0 0 0 1px #0A66C2; color: #0A66C2; font-weight: 600; }
.cal-cell.sel { background: #0A66C2; color: #fff; font-weight: 600; box-shadow: 0 2px 8px rgba(10,102,194,.3); }
.cal-cell.range { background: #EAF3FB; color: #0A66C2; }
.cal-foot { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; padding-top: 12px; border-top: 1px solid #EBEDF0; }
.cal-tip { font-size: 11px; color: #98A2B3; }
.cal-reset { font-size: 12px; color: #0A66C2; font-weight: 500; padding: 4px 8px; }

/* ===== 遮罩 ===== */
.mask { position: fixed; inset: 0; z-index: 60; background: transparent; }
</style>
