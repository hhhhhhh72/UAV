<template>
  <view class="page">
    <!-- Search Wrap (sticky) -->
    <view class="sw" :class="{ stk: stuck }">
      <view class="sbar">
        <view class="sbox">
          <u-icon name="search" size="28rpx" color="#969799" />
          <input class="sinp" v-model="q" placeholder="搜索课题、关键词" placeholder-style="color:#bbb" @input="onSearch" />
          <text v-if="q" class="sclr" @tap="clearSearch">✕</text>
        </view>
        <view class="sbtn" @tap="toggleSort">
          <view class="sort-ic">
            <view class="sort-line"></view>
            <view class="sort-line"></view>
            <view class="sort-line"></view>
          </view>
          <view v-if="sV" class="spop">
            <view v-for="o in sorts" :key="o.k" class="sp-opt" :class="{ active: sort === o.k }" @tap.stop="pickSort(o.k)">
              <text class="sp-chk">{{ sort === o.k ? '✓' : '' }}</text>{{ o.l }}
            </view>
          </view>
        </view>
      </view>
      <!-- Tabs with fade mask -->
      <view class="tmw">
        <scroll-view class="tm" scroll-x enhanced :show-scrollbar="false">
          <text v-for="t in tabs" :key="t.k" class="tab" :class="{ active: activeTab === t.k }" @tap="onTab(t.k)">{{ t.l }}</text>
        </scroll-view>
      </view>
    </view>
    <view v-if="sV" class="mask" @tap="sV = false"></view>

    <!-- Banner -->
    <view class="banner">
      <text class="banner-icon">攻</text>
      <view class="banner-info">
        <text class="banner-title">联合攻关 · 攻克核心技术难题</text>
        <text class="banner-sub">高校院所与企业协同 · 共研关键核心技术</text>
      </view>
    </view>

    <!-- Info Row -->
    <view class="ir">
      <text>共 <text class="irn">{{ shown }}</text> 项课题</text>
      <text class="irs" @tap="toggleSort">{{ sortLabel }} ▼</text>
    </view>

    <!-- Skeleton -->
    <view v-if="loading" class="skl">
      <view v-for="i in 4" :key="'sk' + i" class="skc">
        <view class="sk-l w30"></view>
        <view class="sk-l w90"></view>
        <view class="sk-l w70"></view>
        <view class="sk-l w50"></view>
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
      <u-empty description="暂无相关课题">
        <text class="sth">试试调整筛选条件或搜索关键词</text>
        <view class="stb" @tap="resetAll">清除筛选</view>
      </u-empty>
    </view>

    <!-- Card List -->
    <view v-else class="cl">
      <view v-for="x in list" :key="x.id" class="card" @tap="goDetail(x)">
        <view class="ctop">
          <view class="ctags">
            <text class="cf" :style="{ background: fc(x.f) }">{{ x.f }}</text>
            <text class="cph" :class="x.phCls">{{ phLabel(x.ph) }}</text>
          </view>
        </view>
        <text class="ctit">{{ x.t }}</text>
        <view v-if="x.lead || x.orgsText" class="corg">
          <text v-if="x.lead" class="lead">牵头</text>
          <text>{{ x.lead }}{{ x.lead && x.orgsText ? ' · ' : '' }}{{ x.orgsText }}</text>
        </view>
        <text v-if="x.d" class="cd">{{ x.d }}</text>
        <view class="cft">
          <view class="cstats">
            <text class="cpn">{{ x.pn > 0 ? x.pn + ' 家参与单位' : '参与单位待定' }}</text>
            <text class="cbg">经费预算</text>
          </view>
          <view class="cr">
            <text class="cbu">{{ fmtMoney(x.budget) }}</text>
            <text class="cdl" :class="{ hot: x.urgent }">{{ x.dl }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="list.length" class="lm">{{ hasMore ? '— 上拉加载更多 —' : '— 没有更多了 —' }}</view>

    <!-- Back to top -->
    <view class="bt" :class="{ show: showBt }" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom, onPageScroll } from '@dcloudio/uni-app'
import { request } from '@/utils/request'

const PAGE_SIZE = 100
const MAX_PAGES = 10

const q = ref('')
const sort = ref('latest')
const sV = ref(false)
const activeTab = ref('')
const loading = ref(true)
const err = ref(false)
const list = ref([])
const fullList = ref([])
const total = ref(0)
const stuck = ref(false)
const showBt = ref(false)
let nextPage = 1

const sorts = [
  { k: 'latest', l: '最新发布' },
  { k: 'budget', l: '经费最高' },
  { k: 'deadline', l: '即将截止' }
]
const SORT_LABEL = { latest: '最新发布', budget: '经费最高', deadline: '即将截止' }
const tabs = [
  { k: '', l: '全部' },
  { k: '飞控', l: '飞控系统' },
  { k: '电池', l: '动力电池' },
  { k: 'AI', l: 'AI算法' },
  { k: '通信', l: '通信链路' },
  { k: '材料', l: '新型材料' },
  { k: '载荷', l: '载荷设备' },
  { k: '标准', l: '技术标准' },
  { k: '集群', l: '集群协同' }
]
const FC = { '飞控': '#0d47a1', '电池': '#e65100', 'AI': '#4a148c', '通信': '#1a237e', '材料': '#004d40', '载荷': '#b71c1c', '标准': '#37474f', '集群': '#bf360c' }
const PH = { planning: '规划中', recruiting: '招募中', progress: '进行中', completed: '已完成' }
const PH_CLS = { planning: 'planning', recruiting: 'recruiting', progress: 'progress', completed: 'completed' }

const shown = computed(() => list.value.length)
const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新发布')
const hasMore = computed(() => fullList.value.length < total.value)

const fc = (f) => FC[f] || '#666'
const phLabel = (p) => PH[p] || p || '规划中'
const phCls = (p) => PH_CLS[p] || 'planning'
const fmtMoney = (n) => (n > 0 ? (n >= 10000 ? '¥' + (n / 10000).toFixed(0) + '万' : '¥' + n.toLocaleString()) : '面议')

const daysLeft = (d) => {
  if (!d) return null
  const diff = new Date(d) - new Date()
  return Number.isFinite(diff) ? Math.max(0, Math.ceil(diff / 86400000)) : null
}
const dlOf = (d) => {
  const dl = daysLeft(d)
  if (dl == null) return { text: '截止日期待定', urgent: false }
  return { text: dl + ' 天后截止', urgent: dl <= 30 }
}

// 后端字段映射为卡片字段（缺失字段优雅降级）
const mapItem = (it) => {
  const dl = dlOf(it.end_date)
  const members = Array.isArray(it.members) ? it.members : []
  return {
    id: it.id,
    t: it.title || '未命名课题',
    f: it.field || '其他',
    d: it.description || '',
    lead: it.lead_org || '',
    orgsText: members.slice(0, 3).join(' · '),
    pn: members.length,
    budget: it.budget_fen != null ? it.budget_fen / 100 : 0,
    ph: it.status || 'planning',
    phCls: phCls(it.status),
    dl: dl.text,
    urgent: dl.urgent,
    created: it.created_at || '',
    ddl: it.end_date || ''
  }
}

// 客户端搜索/筛选/排序（与原型交互一致，作用于完整数据集）
const applyFilter = () => {
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter((x) => (x.t + ' ' + x.f + ' ' + x.d).toLowerCase().includes(kw))
  if (activeTab.value) items = items.filter((x) => x.f === activeTab.value)
  if (sort.value === 'budget') items.sort((a, b) => (b.budget || 0) - (a.budget || 0))
  else if (sort.value === 'deadline') items.sort((a, b) => String(a.ddl || '').localeCompare(String(b.ddl || '')))
  else items.sort((a, b) => String(b.created || '').localeCompare(String(a.created || '')))
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
      const res = await request({ url: '/api/v1/research-projects', data: { page, page_size: PAGE_SIZE } })
      const items = Array.isArray(res) ? res : (res?.items || [])
      acc.push(...items)
      fetched = (Array.isArray(res) && res.total) || (res && res.total) || acc.length
      nextPage = page
      if (items.length < PAGE_SIZE || acc.length >= fetched) break
    }
    fullList.value = acc.map(mapItem)
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
  if (loading.value || nextPage >= MAX_PAGES || !hasMore.value) return
  try {
    const res = await request({ url: '/api/v1/research-projects', data: { page: nextPage + 1, page_size: PAGE_SIZE } })
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
const onTab = (k) => { activeTab.value = k; applyFilter() }
const resetAll = () => { q.value = ''; sort.value = 'latest'; activeTab.value = ''; sV.value = false; applyFilter() }

const goDetail = (x) => { uni.showToast({ title: '详情: ' + x.t, icon: 'none' }) }

onPageScroll((e) => {
  const st = e?.scrollTop ?? 0
  stuck.value = st > 5
  showBt.value = st > 400
})
const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })

onLoad(fetchAll)
onPullDownRefresh(async () => {
  await fetchAll()
  uni.stopPullDownRefresh()
})
onReachBottom(fetchMore)
</script>

<style>
page { background: var(--color-bg); }
</style>
<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ===== Search Wrap ===== */
.sw { background: #fff; position: sticky; top: 0; z-index: 29; transition: box-shadow .25s; }
.sw.stk { box-shadow: 0 2px 12px rgba(0, 0, 0, .06) }
.sbar { display: flex; align-items: center; gap: 10px; padding: 8px 14px 10px; background: #fff; }
.sbox { flex: 1; display: flex; align-items: center; background: #f0f1f3; border-radius: 22px; padding: 10px 14px; gap: 8px; }
.sinp { flex: 1; border: none; outline: none; background: transparent; font-size: 14px; color: var(--color-text); min-width: 0; height: 20px; line-height: 20px; }
.sclr { color: #bbb; font-size: 15px; padding: 2px; flex-shrink: 0; }

/* ===== Sort Button + Popup ===== */
.sbtn { width: 38px; height: 38px; border-radius: 50%; background: #f0f1f3; display: flex; align-items: center; justify-content: center; position: relative; flex-shrink: 0; }
.sbtn:active { transform: scale(.93) }
.sort-ic { display: flex; flex-direction: column; gap: 4px; }
.sort-line { width: 14px; height: 2px; border-radius: 1px; background: #666; }
.spop { position: absolute; top: 44px; right: -4px; z-index: 50; background: #fff; border-radius: 12px; box-shadow: 0 4px 24px rgba(0, 0, 0, .12); padding: 6px 0; min-width: 130px; animation: dropIn .18s ease; }
@keyframes dropIn { from { opacity: 0; transform: translateY(-4px) } to { opacity: 1; transform: translateY(0) } }
.sp-opt { padding: 10px 16px; font-size: 13px; color: var(--color-text); display: flex; align-items: center; gap: 8px; white-space: nowrap; }
.sp-opt.active { color: var(--color-primary); font-weight: 600 }
.sp-opt:active { background: #f5f7fa }
.sp-chk { font-size: 12px; width: 14px; }
.mask { position: fixed; inset: 0; z-index: 40; background: transparent }

/* ===== Tabs (capsule + fade mask) ===== */
.tmw { position: relative; background: #fff; padding: 4px 0; }
.tmw::before, .tmw::after { content: ''; position: absolute; top: 0; bottom: 0; width: 28px; z-index: 2; pointer-events: none; }
.tmw::before { left: 0; background: linear-gradient(to right, #fff 0%, transparent 100%) }
.tmw::after { right: 0; background: linear-gradient(to left, #fff 0%, transparent 100%) }
.tm { display: flex; gap: 8px; padding: 6px 14px 10px; white-space: nowrap; }
.tab { flex-shrink: 0; padding: 8px 16px; font-size: 13px; color: #666; border-radius: 18px; }
.tab.active { background: #e8f0fe; color: var(--color-primary); font-weight: 600 }
.tab:active { background: #f0f1f3 }

/* ===== Banner ===== */
.banner { margin: 12px 14px; padding: 16px; border-radius: 14px; background: linear-gradient(135deg, #004d40, #00695c 30%, #00796b); display: flex; align-items: center; gap: 12px; color: #fff; position: relative; overflow: hidden; opacity: 0; animation: cardIn .4s ease .1s forwards; }
@keyframes cardIn { to { opacity: 1; transform: translateY(0) } from { opacity: 0; transform: translateY(12px) } }
.banner::after { content: ''; position: absolute; top: -30%; right: -20%; width: 160px; height: 160px; border-radius: 50%; background: radial-gradient(circle, rgba(255, 255, 255, .08) 0%, transparent 70%) }
.banner-icon { width: 40px; height: 40px; border-radius: 50%; background: rgba(255, 255, 255, .18); font-size: 18px; font-weight: 600; display: flex; align-items: center; justify-content: center; flex-shrink: 0; position: relative; z-index: 1; }
.banner-info { flex: 1; min-width: 0; position: relative; z-index: 1; }
.banner-title { font-size: 14px; font-weight: 600; margin-bottom: 4px; display: block; line-height: 1.3; }
.banner-sub { font-size: 11px; color: rgba(255, 255, 255, .7) }

/* ===== Info Row ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 8px 16px 4px; font-size: 12px; color: #999; }
.irn { color: var(--color-primary); font-weight: 600 }
.irs { color: var(--color-primary); font-weight: 500; padding: 4px 0 4px 12px; }

/* ===== Card List ===== */
.cl { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 100px; }
.card { background: #fff; border-radius: 10px; padding: 14px; border: .5px solid #eee; opacity: 0; transform: translateY(16px); animation: cardIn2 .4s ease forwards; }
@keyframes cardIn2 { to { opacity: 1; transform: translateY(0) } }
.card:active { transform: scale(.98) }
.ctop { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.ctags { display: flex; gap: 6px; align-items: center; }
.cf { font-size: 11px; padding: 2px 8px; border-radius: 8px; font-weight: 500; color: #fff; }
.cph { font-size: 10px; padding: 2px 8px; border-radius: 8px; font-weight: 500; }
.cph.planning { background: #ede7f6; color: #5e35b1 }
.cph.recruiting { background: #e8f5e9; color: #2e7d32 }
.cph.progress { background: #e3f2fd; color: #1565c0 }
.cph.completed { background: #f5f5f5; color: #999 }
.ctit { font-size: 15px; font-weight: 600; color: var(--color-text); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 8px; display: block; }
.corg { font-size: 12px; color: #666; display: flex; align-items: center; gap: 6px; margin-bottom: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.lead { font-size: 11px; background: #e8f0fe; color: #1967d2; padding: 1px 6px; border-radius: 4px; flex-shrink: 0; }
.cd { font-size: 12px; color: #999; line-height: 1.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 12px; }
.cft { display: flex; align-items: flex-end; justify-content: space-between; }
.cstats { display: flex; flex-direction: column; gap: 4px; }
.cpn { font-size: 11px; color: var(--color-primary); font-weight: 500 }
.cbg { font-size: 11px; color: #999 }
.cr { text-align: right; display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.cbu { font-size: 16px; font-weight: 700; color: #f57c00; }
.cdl { font-size: 11px; color: #999; }
.cdl.hot { color: #e65100; font-weight: 500 }

/* ===== Skeleton ===== */
.skl { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 20px; }
.skc { background: #fff; border-radius: 10px; padding: 14px; }
.sk-l { height: 14px; background: #f0f1f3; border-radius: 4px; margin-bottom: 8px; animation: shimmer 1.5s infinite; }
.sk-l.w30 { width: 30% }
.sk-l.w90 { width: 90% }
.sk-l.w70 { width: 70% }
.sk-l.w50 { width: 50% }
@keyframes shimmer { 0%, 100% { opacity: 1 } 50% { opacity: .45 } }

/* ===== State ===== */
.st { text-align: center; padding: 40px 20px; }
.sth { font-size: 12px; color: #bbb; display: block; margin-bottom: 16px; }
.stb { display: inline-block; padding: 8px 24px; border-radius: 22px; background: var(--color-primary); color: #fff; font-size: 13px; font-weight: 500; }
.stb:active { opacity: .8 }

/* ===== Load More / Back Top ===== */
.lm { text-align: center; padding: 12px; font-size: 12px; color: #bbb; }
.bt { position: fixed; bottom: 90px; right: 16px; width: 44px; height: 44px; border-radius: 50%; background: #fff; box-shadow: 0 4px 16px rgba(0, 0, 0, .12); display: flex; align-items: center; justify-content: center; z-index: 60; opacity: 0; transform: scale(.5); pointer-events: none; transition: opacity .3s, transform .3s cubic-bezier(.17, .89, .32, 1.9); font-size: 20px; color: #666; }
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto }
.bt:active { transform: scale(.88) }
</style>
