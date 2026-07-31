<template>
  <view class="page" @scroll="onScroll">
    <!-- Navbar -->
    <view class="nb" :class="{ stk: stuck }">
      <text class="nb-back" @tap="goBack">←</text>
      <text class="nb-title">研发难题广场</text>
      <text class="nb-spacer"></text>
    </view>

    <!-- Search Wrap -->
    <view class="sw" :class="{ stk: stuck }">
      <view class="sbar">
        <view class="sbox">
          <text class="sic">🔍</text>
          <input class="sinp" v-model="q" placeholder="搜索难题、关键词" placeholder-style="color:#bbb" @input="onSearch" />
          <text v-if="q" class="sclr" @tap="clearSearch">✕</text>
        </view>
        <view class="sbtn" @tap="toggleSort">
          <text class="sic">☰</text>
          <view v-if="sV" class="spop">
            <view v-for="o in sorts" :key="o.k" class="sp-opt" :class="{active:sort===o.k}" @tap.stop="pickSort(o.k)">{{ o.l }}</view>
          </view>
        </view>
      </view>
      <scroll-view class="tabs" scroll-x enhanced :show-scrollbar="false">
        <text v-for="t in tabs" :key="t.k" class="tab" :class="{active:activeTab===t.k}" @tap="onTab(t.k)">{{ t.l }}</text>
      </scroll-view>
    </view>
    <view v-if="sV" class="mask" @tap="sV=false"></view>

    <!-- Banner -->
    <view class="banner">
      <text class="banner-icon">🏆</text>
      <view class="banner-info">
        <text class="banner-title">揭榜挂帅 · 技术攻关等你来战</text>
        <text class="banner-sub">累计悬赏 ¥128万 · 已攻克 47 项 · 正在揭榜 12 项</text>
      </view>
      <view class="banner-btn" @tap="onPublish">发布难题</view>
    </view>

    <view class="ir"><text>共 <text class="irn">{{ total }}</text> 项难题</text></view>

    <!-- Skeleton -->
    <view v-if="loading" class="skl">
      <view v-for="i in 4" :key="'sk'+i" class="skc"><view class="sk-l w30"></view><view class="sk-l w90"></view><view class="sk-l w70"></view><view class="sk-l w50"></view></view>
    </view>

    <!-- Error -->
    <view v-else-if="err" class="st"><text class="sti">⚠</text><text class="stt">加载失败</text><view class="stb" @tap="fetchData">重新加载</view></view>

    <!-- Empty -->
    <view v-else-if="!list.length" class="st"><text class="sti">🔍</text><text class="stt">暂无相关难题</text><view class="stb" @tap="resetAll">清除筛选</view></view>

    <!-- Card List -->
    <view v-else class="cl">
      <view v-for="x in list" :key="x.id" class="card" @tap="goDetail(x)">
        <view class="chd">
          <view class="ctags"><text class="cf" :style="{background:fc(x.field)}">{{ x.field }}</text><text class="cs2" :class="x.status">{{ stLabel(x.status) }}</text></view>
        </view>
        <text class="ctit">{{ x.title }}</text>
        <text class="cd">{{ x.desc }}</text>
        <view class="cft">
          <view class="cm"><text class="co">{{ x.org }}</text><text class="cbids">✏️ {{ x.bids }} 家已揭榜</text></view>
          <view class="cr"><text class="crw">{{ fmtMoney(x.reward) }}</text><text class="cdl">⏰ {{ dlText(x.deadline) }}</text></view>
        </view>
      </view>
    </view>

    <view v-if="list.length" class="lm">— 上拉加载更多 —</view>
    <view class="bt" :class="{show:showBt}" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request } from '@/utils/request'

const goBack = () => uni.navigateBack()
const goDetail = (x) => uni.showToast({ title:'详情: '+x.title, icon:'none' })
const onPublish = () => uni.showToast({ title:'发布难题 (仅会员)', icon:'none' })

const FC = { '飞控':'#0d47a1','电池':'#e65100','AI':'#4a148c','通信':'#1a237e','材料':'#004d40','载荷':'#b71c1c' }
const fc = (f) => FC[f]||'#666'
const stLabel = (s) => s === 'urgent' ? '⚠ 紧急' : s === 'closed' ? '已截止' : '进行中'
const fmtMoney = (n) => n ? (n >= 10000 ? '¥'+(n/10000).toFixed(0)+'万' : '¥'+n.toLocaleString()) : '面议'
const daysLeft = (d) => Math.max(0, Math.ceil((new Date(d) - new Date()) / 86400000))
const dlText = (d) => { const dl = daysLeft(d); return dl <= 7 ? dl+'天后截止' : dl+'天后截止' }

const q = ref(''), sort = ref('latest'), sV = ref(false), activeTab = ref('')
const loading = ref(true), err = ref(false), list = ref([]), total = ref(0)
const stuck = ref(false), showBt = ref(false)

const sorts = [{ k:'latest', l:'最新发布' },{ k:'reward', l:'悬赏最高' },{ k:'deadline', l:'即将截止' }]
const tabs = [{ k:'', l:'全部' },{ k:'飞控', l:'飞控系统' },{ k:'电池', l:'动力电池' },{ k:'AI', l:'AI算法' },{ k:'通信', l:'通信链路' },{ k:'材料', l:'新型材料' },{ k:'载荷', l:'载荷设备' }]

const DEMO = [
  { id:1, field:'飞控', title:'多旋翼无人机抗6级阵风飞控算法研发', org:'大疆创新科技', reward:150000, deadline:'2026-09-30', status:'open', desc:'需要开发一套能在6级阵风条件下保持稳定的自适应飞控算法...', bids:12 },
  { id:2, field:'电池', title:'高能量密度固态电池无人机适配方案', org:'匿名企业', reward:300000, deadline:'2026-08-15', status:'urgent', desc:'急需适配现有工业无人机平台的固态电池方案...', bids:28 },
  { id:3, field:'AI', title:'基于边缘计算的实时目标识别与追踪系统', org:'华为', reward:200000, deadline:'2026-10-20', status:'open', desc:'在无人机端侧实现实时目标检测，延迟<50ms...', bids:18 },
  { id:4, field:'通信', title:'无人机超视距5G通信中继系统', org:'中国移动', reward:180000, deadline:'2026-11-01', status:'open', desc:'设计一套基于5G网络切片的无人机超视距控制与数据传输方案...', bids:9 },
]

const applyFilter = () => {
  let items = DEMO.slice()
  if (q.value) { const kw = q.value.toLowerCase(); items = items.filter(x => x.title.toLowerCase().includes(kw) || x.org.toLowerCase().includes(kw) || x.field.toLowerCase().includes(kw)) }
  if (activeTab.value) items = items.filter(x => x.field === activeTab.value)
  if (sort.value === 'reward') items.sort((a,b) => b.reward - a.reward)
  else if (sort.value === 'deadline') items.sort((a,b) => new Date(a.deadline) - new Date(b.deadline))
  else items.sort((a,b) => b.id - a.id)
  list.value = items; total.value = items.length
}

const fetchData = async () => {
  loading.value = true; err.value = false
  try {
    const res = await request({ url: '/api/v1/challenges', data: { page:1, page_size:20 } })
    const data = res?.data || res
    if (Array.isArray(data)) { list.value = data; total.value = data.length }
    else if (data?.items) { list.value = data.items; total.value = data.total||data.items.length }
    else applyFilter()
  } catch { applyFilter() }
  finally { loading.value = false }
}

const onSearch = () => applyFilter()
const clearSearch = () => { q.value = ''; applyFilter() }
const toggleSort = () => { sV.value = !sV.value }
const pickSort = (k) => { sort.value = k; sV.value = false; applyFilter() }
const onTab = (k) => { activeTab.value = k; applyFilter() }
const resetAll = () => { q.value = ''; sort.value = 'latest'; activeTab.value = ''; sV.value = false; applyFilter() }
const onScroll = (e) => { const st = e?.detail?.scrollTop ?? 0; stuck.value = st > 5; showBt.value = st > 400 }
const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })

fetchData()
</script>

<style>
page { background:#e8eaed; --cp:#0A66C2; --cbg:#f5f6f8; --cwh:#fff; --ct:#1a1a1a; --ct2:#666; --ct3:#999; --cibg:#f0f1f3 }
</style>
<style scoped>
.page { min-height:100vh; background:var(--cbg); padding-bottom:env(safe-area-inset-bottom) }
.nb { display:flex; align-items:center; justify-content:space-between; padding:12px 14px; background:var(--cwh); position:sticky; top:0; z-index:30; border-bottom:.5px solid transparent; transition:border-color .25s,box-shadow .25s }
.nb.stk { border-bottom-color:#eee; box-shadow:0 2px 8px rgba(0,0,0,.04) }
.nb-back { font-size:20px; color:var(--ct); padding:4px; width:36px }
.nb-title { font-size:17px; font-weight:600; color:var(--ct) }
.nb-spacer { width:36px }
.sw { background:var(--cwh); position:sticky; top:50px; z-index:29; transition:box-shadow .25s }
.sw.stk { box-shadow:0 2px 12px rgba(0,0,0,.06) }
.sbar { display:flex; align-items:center; gap:10px; padding:8px 14px 12px; background:var(--cwh) }
.sbox { flex:1; display:flex; align-items:center; background:var(--cibg); border-radius:22px; padding:10px 14px; gap:8px }
.sic { font-size:15px; opacity:.5; flex-shrink:0 }
.sinp { flex:1; border:none; outline:none; background:transparent; font-size:14px; color:var(--ct); min-width:0; height:20px; line-height:20px }
.sclr { color:#bbb; font-size:15px; padding:2px; flex-shrink:0 }
.sbtn { width:38px; height:38px; border-radius:50%; background:var(--cibg); display:flex; align-items:center; justify-content:center; position:relative; flex-shrink:0 }
.sbtn:active { transform:scale(.93) }
.spop { position:absolute; top:44px; right:-4px; z-index:50; background:var(--cwh); border-radius:12px; box-shadow:0 4px 24px rgba(0,0,0,.12); padding:6px 0; min-width:130px; animation:dropIn .18s ease }
@keyframes dropIn { from { opacity:0; transform:translateY(-4px) } to { opacity:1; transform:translateY(0) } }
.sp-opt { padding:10px 16px; font-size:13px; color:var(--ct); white-space:nowrap }
.sp-opt.active { color:var(--cp); font-weight:600 }
.sp-opt:active { background:#f5f7fa }
.mask { position:fixed; inset:0; z-index:40; background:transparent }
.tabs { display:flex; gap:0; padding:6px 14px 8px; background:var(--cwh); white-space:nowrap }
.tab { flex-shrink:0; padding:8px 16px; font-size:13px; color:var(--ct2); position:relative }
.tab.active { color:var(--cp); font-weight:600 }
.tab.active::after { content:''; position:absolute; bottom:-8px; left:50%; transform:translateX(-50%); width:20px; height:3px; background:var(--cp); border-radius:2px }
.banner { margin:12px 14px; padding:16px; border-radius:14px; background:linear-gradient(135deg,#1a237e,#283593 30%,#3949ab); display:flex; align-items:center; gap:12px; color:#fff; position:relative; overflow:hidden; opacity:0; animation:cardIn .4s ease .1s forwards }
@keyframes cardIn { to { opacity:1; transform:translateY(0) } from { opacity:0; transform:translateY(12px) } }
.banner::after { content:''; position:absolute; top:-30%; right:-20%; width:160px; height:160px; border-radius:50%; background:radial-gradient(circle,rgba(255,255,255,.08) 0%,transparent 70%) }
.banner-icon { font-size:32px; flex-shrink:0; position:relative; z-index:1 }
.banner-info { flex:1; min-width:0; position:relative; z-index:1 }
.banner-title { font-size:14px; font-weight:600; margin-bottom:4px; display:block; line-height:1.3 }
.banner-sub { font-size:11px; color:rgba(255,255,255,.7) }
.banner-btn { flex-shrink:0; padding:7px 16px; border-radius:22px; background:rgba(255,255,255,.2); color:#fff; font-size:12px; font-weight:500; position:relative; z-index:1 }
.banner-btn:active { background:rgba(255,255,255,.35) }
.ir { display:flex; justify-content:space-between; align-items:center; padding:8px 16px 4px; font-size:12px; color:var(--ct3) }
.irn { color:var(--cp); font-weight:600 }
.cl { display:flex; flex-direction:column; gap:10px; padding:0 14px 100px }
.card { background:var(--cwh); border-radius:10px; padding:14px; border:.5px solid #eee; opacity:0; transform:translateY(16px); animation:cardIn2 .4s ease forwards }
@keyframes cardIn2 { to { opacity:1; transform:translateY(0) } }
.card:active { transform:scale(.98) }
.chd { display:flex; align-items:center; justify-content:space-between; margin-bottom:10px }
.ctags { display:flex; gap:6px; align-items:center }
.cf { font-size:11px; padding:2px 8px; border-radius:8px; font-weight:500; color:#fff }
.cs2 { font-size:10px; padding:2px 8px; border-radius:8px; font-weight:500 }
.cs2.open { background:#e8f5e9; color:#2e7d32 }.cs2.urgent { background:#fce4ec; color:#c62828 }.cs2.closed { background:#f5f5f5; color:#999 }
.ctit { font-size:15px; font-weight:600; color:var(--ct); line-height:1.4; display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin-bottom:8px;display:block }
.cd { font-size:12px; color:var(--ct3); line-height:1.5; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; margin-bottom:12px }
.cft { display:flex; align-items:flex-end; justify-content:space-between }
.cm { display:flex; flex-direction:column; gap:4px }
.co { font-size:11px; color:var(--ct2); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; max-width:140px }
.cbids { font-size:11px; color:var(--cp); font-weight:500 }
.cr { text-align:right; display:flex; flex-direction:column; align-items:flex-end; gap:3px }
.crw { font-size:16px; font-weight:700; color:#ff3b30 }
.cdl { font-size:11px; color:var(--ct3) }
.skl { display:flex; flex-direction:column; gap:10px; padding:0 14px 20px }
.skc { background:var(--cwh); border-radius:10px; padding:14px }
.sk-l { height:14px; background:var(--cibg); border-radius:4px; margin-bottom:8px; animation:shimmer 1.5s infinite }
.sk-l.w30 { width:30% }.sk-l.w90 { width:90% }.sk-l.w70 { width:70% }.sk-l.w50 { width:50% }
@keyframes shimmer { 0%,100% { opacity:1 } 50% { opacity:.45 } }
.st { text-align:center; padding:60px 20px }
.sti { font-size:48px; margin-bottom:12px; opacity:.5; display:block }
.stt { font-size:14px; color:var(--ct3); margin-bottom:16px; display:block }
.stb { display:inline-block; padding:8px 24px; border-radius:22px; background:var(--cp); color:#fff; font-size:13px; font-weight:500 }
.lm { text-align:center; padding:12px; font-size:12px; color:#bbb }
.bt { position:fixed; bottom:90px; right:16px; width:44px; height:44px; border-radius:50%; background:var(--cwh); box-shadow:0 4px 16px rgba(0,0,0,.12); display:flex; align-items:center; justify-content:center; z-index:60; opacity:0; transform:scale(.5); pointer-events:none; transition:opacity .3s,transform .3s cubic-bezier(.17,.89,.32,1.9); font-size:20px; color:#666 }
.bt.show { opacity:1; transform:scale(1); pointer-events:auto }
.bt:active { transform:scale(.88) }
</style>
