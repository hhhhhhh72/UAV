<template>
  <view class="page">
    <!-- Navbar -->
    <view class="nb">
      <text class="nb-back" @tap="goBack">←</text>
      <text class="nb-title">成果库</text>
      <text class="nb-spacer"></text>
    </view>

    <!-- Search Bar -->
    <view class="sbar">
      <view class="sbox">
        <text class="sic">🔍</text>
        <input class="sinp" v-model="q" placeholder="搜索成果名称、关键词" placeholder-style="color:#bbb" @input="onSearch" />
        <text v-if="q" class="sclr" @tap="clearSearch">✕</text>
      </view>
      <view class="sbtn" @tap="toggleSort">
        <text class="sic">☰</text>
        <view v-if="sV" class="spop">
          <view v-for="o in sorts" :key="o.k" class="sp-opt" :class="{active:sort===o.k}" @tap.stop="pickSort(o.k)">
            <text v-if="sort===o.k" class="sp-chk">✓</text>
            <text>{{ o.l }}</text>
          </view>
        </view>
      </view>
    </view>
    <view v-if="sV" class="mask" @tap="sV=false"></view>

    <!-- Banner Carousel -->
    <swiper class="cs" :indicator-dots="true" :autoplay="true" :interval="3500" :duration="400" :circular="true"
      indicator-color="rgba(255,255,255,.35)" indicator-active-color="#fff">
      <swiper-item v-for="(s,i) in banners" :key="i">
        <view class="cslide" :style="{background:s.bg}">
          <text class="csi">{{ s.icon }}</text>
          <view class="csinfo"><text class="cst">{{ s.title }}</text><text class="css">{{ s.sub }}</text></view>
        </view>
      </swiper-item>
    </swiper>

    <!-- Func Nav -->
    <view class="fn">
      <view v-for="n in navs" :key="n.k" class="fi" @tap="onNav(n.k)">
        <view class="fii" :style="{background:n.bg}"><text class="fie">{{ n.icon }}</text></view>
        <text class="fl">{{ n.label }}</text>
      </view>
    </view>

    <!-- Info Row -->
    <view class="ir"><text>共 <text class="irn">{{ total }}</text> 项成果</text></view>

    <!-- Skeleton -->
    <view v-if="loading" class="cg">
      <view v-for="i in 6" :key="'sk'+i" class="card card-sk">
        <view class="sk-cv"></view><view class="sk-bd"><view class="sk-l w90"></view><view class="sk-l"></view><view class="sk-l w60"></view></view>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="err" class="st">
      <text class="sti">⚠</text><text class="stt">加载失败，请检查网络</text>
      <text class="sth">请确认网络连接后重试</text><view class="stb" @tap="fetchData">重新加载</view>
    </view>

    <!-- Empty -->
    <view v-else-if="!list.length" class="st">
      <text class="sti">🔍</text><text class="stt">暂无相关成果</text>
      <text class="sth">试试调整筛选条件或搜索关键词</text><view class="stb" @tap="resetAll">清除筛选</view>
    </view>

    <!-- Card Grid -->
    <view v-else class="cg">
      <view v-for="x in list" :key="x.id" class="card" @tap="goDetail(x)">
        <view class="cc" :style="{background:bgOf(x.f)}">
          <text class="cci">{{ icOf(x.f) }}</text>
          <text class="cct">{{ x.f }}</text>
          <text v-if="stOf(x.st)" class="ccs" :class="x.st">{{ stOf(x.st) }}</text>
        </view>
        <view class="cbd">
          <text class="ct">{{ x.t }}</text>
          <text class="co">{{ x.o }}</text>
          <view class="cft"><text>{{ x.d }}</text><text>👁 {{ fmt(x.v) }}</text></view>
        </view>
      </view>
    </view>

    <view v-if="list.length" class="lm">— 上拉加载更多 —</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request } from '@/utils/request'

const goBack = () => uni.navigateBack()
const goDetail = (x) => uni.navigateTo({ url: '/pages/achievements/detail?id=' + encodeURIComponent(x.id) })

const ICONS = { '飞控系统':'✈','遥感测绘':'🌍','动力系统':'⚙','AI算法':'🧠','载荷设备':'📷','集群协同':'📡','通信链路':'📦','标准规范':'📋','地面站':'💻' }
const BGS = { '飞控系统':'#e3f2fd','遥感测绘':'#e8f5e9','动力系统':'#fff3e0','AI算法':'#f3e5f5','载荷设备':'#fce4ec','集群协同':'#e0f2f1','通信链路':'#e8eaf6','标准规范':'#f5f5f5','地面站':'#fff8e1' }
const STL = { hot:'热门', transformed:'已转化', 'new':'新成果' }
const icOf = (f) => ICONS[f] || '🚀'
const bgOf = (f) => BGS[f] || '#f0f1f3'
const stOf = (s) => STL[s] || ''
const fmt = (n) => n ? (n >= 1e4 ? (n/1e4).toFixed(1)+'w' : n >= 1e3 ? (n/1e3).toFixed(1)+'k' : String(n)) : '0'

const q = ref(''), sort = ref('latest'), sV = ref(false), loading = ref(true), err = ref(false)
const list = ref([]), total = ref(0)
const sorts = [{ k:'latest', l:'最新发布' },{ k:'views', l:'最多浏览' },{ k:'favs', l:'最多收藏' }]
const navs = [
  { k:'patent', icon:'📄', label:'发明专利', bg:'#e3f2fd' },{ k:'utility', icon:'⚙', label:'实用新型', bg:'#fff3e0' },
  { k:'copyright', icon:'💻', label:'软件著作', bg:'#e8f5e9' },{ k:'paper', icon:'📚', label:'论文成果', bg:'#f3e5f5' },
  { k:'standard', icon:'📶', label:'技术标准', bg:'#fce4ec' },{ k:'design', icon:'🎨', label:'外观设计', bg:'#e0f2f1' },
  { k:'transformed', icon:'🚀', label:'已转化', bg:'#fff8e1' },{ k:'all', icon:'🔍', label:'全部成果', bg:'#e8eaf6' }
]
const banners = [
  { icon:'✈', title:'AI 赋能飞控新突破', sub:'本月新增 42 项前沿成果', bg:'linear-gradient(135deg,#0d47a1,#1976d2)' },
  { icon:'📋', title:'产学研对接加速', sub:'326 项成果已实现转化', bg:'linear-gradient(135deg,#1b5e20,#2e7d32)' },
  { icon:'🚀', title:'标准引领行业', sub:'最新无人机适航标准发布', bg:'linear-gradient(135deg,#4a148c,#7b1fa2)' }
]

const DEMO = [
  { id:1, f:'飞控系统', t:'无人机智能自适应飞控系统 V3.0', o:'北航无人机研究所', d:'2026-07-15', v:2380, s:186, st:'hot' }
]

const applyFilter = () => {
  let items = DEMO.slice()
  if (q.value) {
    const kw = q.value.toLowerCase()
    items = items.filter(x => x.t.toLowerCase().includes(kw) || x.o.toLowerCase().includes(kw) || x.f.toLowerCase().includes(kw))
  }
  if (sort.value === 'views') items.sort((a,b) => b.v - a.v)
  else if (sort.value === 'favs') items.sort((a,b) => b.s - a.s)
  else items.sort((a,b) => new Date(b.d) - new Date(a.d))
  total.value = items.length
  return items
}

// 将后端返回的数据统一映射为模板用的字段
const mapItem = (item) => ({
  id: item.id || item.ID,
  f: item.field || item.category || item.f || '其他',
  t: item.title || item.t || '',
  o: item.org_name || item.org || item.o || '未知机构',
  d: (item.created_at || item.date || item.d || '').slice(0, 10),
  v: item.views || item.view_count || item.v || 0,
  s: item.favs || item.fav_count || item.s || 0,
  st: item.status || item.st || ''
})

const fetchData = async () => {
  loading.value = true; err.value = false
  try {
    const res = await request({ url: '/api/v1/achievements', data: { page: 1, page_size: 20 } })
    const data = Array.isArray(res) ? res : (res?.data?.items || res?.data || res?.items || [])
    const items = Array.isArray(data) ? data : (data?.items || data?.data || [])
    if (items.length) {
      list.value = items.map(mapItem)
      total.value = (res?.data?.total || res?.total || items.length)
    } else {
      list.value = applyFilter()
    }
  } catch { list.value = applyFilter() }
  finally { loading.value = false }
}

const onSearch = () => { list.value = applyFilter() }
const clearSearch = () => { q.value = ''; list.value = applyFilter() }
const toggleSort = () => { sV.value = !sV.value }
const pickSort = (k) => { sort.value = k; sV.value = false; list.value = applyFilter() }
const onNav = (k) => {
  if (k === 'all') { resetAll(); return }
  const map = { patent:'发明专利', utility:'实用新型', copyright:'软件著作', paper:'论文成果', standard:'技术标准', design:'外观设计', transformed:'已转化' }
  q.value = map[k] || ''; list.value = applyFilter()
}
const resetAll = () => { q.value = ''; sort.value = 'latest'; list.value = applyFilter(); sV.value = false }

fetchData()
</script>

<style>
page { --cp:#0A66C2; --cbg:#f5f6f8; --cwh:#fff; --ct:#1a1a1a; --ct2:#666; --ct3:#999; --ct4:#bbb; --cb:#eee; --cibg:#f0f1f3; background:#e8eaed }
</style>
<style scoped>
.page { min-height:100vh; background:var(--cbg); padding-bottom:env(safe-area-inset-bottom) }
.nb { display:flex; align-items:center; justify-content:space-between; padding:12px 14px; background:var(--cwh); position:sticky; top:0; z-index:20; border-bottom:.5px solid var(--cb) }
.nb-back { font-size:20px; color:var(--ct); padding:4px; width:36px }
.nb-title { font-size:17px; font-weight:600; color:var(--ct) }
.nb-spacer { width:36px }
.sbar { display:flex; align-items:center; gap:10px; padding:12px 14px; background:var(--cwh) }
.sbox { flex:1; display:flex; align-items:center; background:var(--cibg); border-radius:22px; padding:10px 14px; gap:8px }
.sic { font-size:15px; opacity:.5; flex-shrink:0 }
.sinp { flex:1; border:none; outline:none; background:transparent; font-size:14px; color:var(--ct); min-width:0; height:20px; line-height:20px }
.sclr { color:var(--ct4); font-size:16px; padding:2px; flex-shrink:0 }
.sbtn { width:38px; height:38px; border-radius:50%; background:var(--cibg); display:flex; align-items:center; justify-content:center; position:relative; flex-shrink:0 }
.sbtn:active { transform:scale(.93) }
.spop { position:absolute; top:44px; right:-4px; z-index:50; background:var(--cwh); border-radius:12px; box-shadow:0 4px 24px rgba(0,0,0,.12); padding:6px 0; min-width:132px; animation:popIn .18s ease }
@keyframes popIn { from { opacity:0; transform:translateY(-4px) } to { opacity:1; transform:translateY(0) } }
.sp-opt { padding:10px 16px; font-size:13px; color:var(--ct); display:flex; align-items:center; gap:8px; white-space:nowrap }
.sp-opt.active { color:var(--cp); font-weight:600 }
.sp-opt:active { background:#f5f7fa }
.sp-chk { font-size:12px }
.mask { position:fixed; inset:0; z-index:40; background:transparent }
.cs { margin:12px 14px; border-radius:14px; overflow:hidden; height:130px }
.cslide { width:100%; height:100%; display:flex; align-items:center; padding:0 20px; gap:14px }
.csi { font-size:42px; flex-shrink:0 }
.csinfo { flex:1; min-width:0 }
.cst { font-size:15px; font-weight:600; color:#fff; margin-bottom:4px; display:block; line-height:1.3 }
.css { font-size:12px; color:rgba(255,255,255,.72) }
.fn { display:grid; grid-template-columns:repeat(4,1fr); gap:4px; padding:0 14px 12px }
.fi { text-align:center; padding:8px 4px }
.fi:active { transform:scale(.93) }
.fii { width:44px; height:44px; border-radius:12px; display:flex; align-items:center; justify-content:center; margin:0 auto 6px }
.fie { font-size:20px }
.fl { font-size:11px; color:var(--ct2) }
.ir { display:flex; justify-content:space-between; align-items:center; padding:2px 16px 8px; font-size:12px; color:var(--ct3) }
.irn { color:var(--cp); font-weight:600 }
.cg { display:grid; grid-template-columns:1fr 1fr; gap:8px; padding:0 14px 20px }
.card { background:var(--cwh); border-radius:10px; overflow:hidden; border:.5px solid var(--cb) }
.card:active { transform:scale(.97) }
.cc { position:relative; aspect-ratio:4/3; display:flex; align-items:center; justify-content:center }
.cci { font-size:30px }
.cct { position:absolute; top:6px; left:6px; background:rgba(0,0,0,.45); color:#fff; font-size:10px; padding:2px 8px; border-radius:8px; font-weight:500 }
.ccs { position:absolute; top:6px; right:6px; color:#fff; font-size:10px; padding:2px 7px; border-radius:6px; font-weight:500 }
.ccs.hot { background:#ff3b30 }.ccs.transformed { background:#34c759 }.ccs.new { background:var(--cp) }
.cbd { padding:8px 10px 10px }
.ct { font-size:13px; font-weight:600; color:var(--ct); line-height:1.4; display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; overflow:hidden; margin-bottom:4px }
.co { font-size:11px; color:var(--ct2); margin-bottom:4px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap }
.cft { font-size:10px; color:var(--ct4); display:flex; justify-content:space-between; align-items:center }
.card-sk .sk-cv { aspect-ratio:4/3; background:var(--cibg); animation:shimmer 1.5s infinite }
.sk-bd { padding:8px 10px }
.sk-l { height:12px; background:var(--cibg); border-radius:4px; margin-bottom:6px; animation:shimmer 1.5s infinite }
.sk-l.w90 { width:90% }.sk-l.w60 { width:60% }
@keyframes shimmer { 0%,100% { opacity:1 } 50% { opacity:.45 } }
.st { grid-column:1/-1; text-align:center; padding:60px 20px }
.sti { font-size:48px; margin-bottom:12px; opacity:.5; display:block }
.stt { font-size:14px; color:var(--ct3); margin-bottom:4px; display:block }
.sth { font-size:12px; color:#ccc; margin-bottom:16px; display:block }
.stb { display:inline-block; padding:8px 24px; border-radius:22px; background:var(--cp); color:#fff; font-size:13px; font-weight:500 }
.lm { text-align:center; padding:12px; font-size:12px; color:#ccc }
</style>
