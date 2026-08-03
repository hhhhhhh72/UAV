<template>
  <view class="page" @scroll="onScroll">
    <!-- Skeleton -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-stats"><view class="sk-stat"></view><view class="sk-stat"></view><view class="sk-stat"></view></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- Error -->
    <view v-else-if="err" class="st"><text class="sti">!</text><text class="stt">加载失败，请检查网络</text><text class="sth">请确认网络后重试</text><view class="stb" @tap="fetchData">重新加载</view></view>

    <!-- Empty -->
    <view v-else-if="!d" class="st"><u-icon class="sti" name="search" size="96rpx" color="#d5d7db" /><text class="stt">该成果已下架或不存在</text><text class="sth">请返回列表浏览其他成果</text><view class="stb" @tap="goBack">返回列表</view></view>

    <!-- Content -->
    <template v-else>
      <view class="hero" :style="{background: heroBg}">
        <view class="hero-glow"></view>
        <text class="hero-icon">{{ heroIcon }}</text>
        <text class="hero-tag">{{ d.field }}</text>
        <view class="hero-wave"><view class="hero-wave-inner"></view></view>
      </view>

      <view class="ts" :class="{ stk: titleStuck }">
        <text class="mt">{{ d.title }}</text>
        <view class="badges">
          <text v-if="d.status==='hot'" class="b b-hot">热门</text>
          <text v-if="isTransformed" class="b b-tr">已转化</text>
          <text v-if="stageLabel" class="b b-st">{{ stageLabel }}</text>
        </view>
      </view>

      <view class="sec stats" :class="{ vis: vis }">
        <view class="si"><text class="sv">{{ (d.created_at||'').slice(0,10) }}</text><text class="sl">发布日期</text></view>
        <view class="si"><text class="sv">{{ typeLabel }}</text><text class="sl">成果类型</text></view>
        <view class="si"><text class="sv">{{ d.owner_id ? '已认证' : '未认证' }}</text><text class="sl">状态</text></view>
      </view>

      <view class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">成果描述</text></view>
        <text class="sb">{{ d.description || '暂无描述' }}</text>
      </view>

      <view class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">基本信息</text></view>
        <view class="it">
          <view v-if="d.field" class="ir2"><text class="ik">所属领域</text><text class="iv">{{ d.field }}</text></view>
          <view v-if="d.achieve_type" class="ir2"><text class="ik">成果类型</text><text class="iv">{{ typeLabel }}</text></view>
          <view v-if="d.stage" class="ir2"><text class="ik">成果阶段</text><text class="iv" :class="stageCls">{{ stageLabel }}</text></view>
          <view v-if="d.contact_info" class="ir2"><text class="ik">联系方式</text><text class="iv">{{ d.contact_info }}</text></view>
          <view v-if="d.owner_id" class="ir2"><text class="ik">归属用户</text><text class="iv">{{ d.owner_id }}</text></view>
          <view v-if="d.created_at" class="ir2"><text class="ik">创建时间</text><text class="iv">{{ (d.created_at||'').slice(0,10) }}</text></view>
          <view v-if="d.updated_at && d.updated_at !== d.created_at" class="ir2"><text class="ik">更新时间</text><text class="iv">{{ (d.updated_at||'').slice(0,10) }}</text></view>
        </view>
      </view>

      <view style="height:80px"></view>
    </template>

    <!-- Bottom Bar -->
    <view v-if="d && !loading" class="bb" :class="{ h: barHidden }">
      <view class="bi" :class="{ fv: isFav }" @tap="toggleFav"><text class="bit">{{ isFav ? '♥' : '♡' }}</text></view>
      <view class="bp" @tap="onContact">联系对接</view>
      <view class="bo" @tap="toggleFav">{{ isFav ? '已收藏' : '收藏' }}</view>
      <view class="bi" @tap="toggleShare"><text class="bit">↗</text></view>
    </view>

    <!-- Fav Pop -->
    <view v-if="favPop" class="fp" :class="{ hide: favHide }">♥</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'

const goBack = () => uni.navigateBack()

const ICONS = { '飞控系统':'飞','遥感测绘':'遥','动力系统':'动','AI算法':'算','载荷设备':'载','集群协同':'群','通信链路':'通','标准规范':'标','地面站':'地' }
const BGS = {
  '飞控系统':'linear-gradient(160deg,#0d47a1,#1565c0 30%,#1976d2 60%,#0d47a1)',
  '遥感测绘':'linear-gradient(160deg,#1b5e20,#2e7d32 30%,#388e3c 60%,#1b5e20)',
  '动力系统':'linear-gradient(160deg,#e65100,#ef6c00 50%,#f57c00)',
  'AI算法':'linear-gradient(160deg,#4a148c,#6a1b9a 50%,#7b1fa2)',
  '载荷设备':'linear-gradient(160deg,#b71c1c,#c62828 50%,#d32f2f)',
  '集群协同':'linear-gradient(160deg,#004d40,#00695c 50%,#00796b)',
  '通信链路':'linear-gradient(160deg,#1a237e,#283593 50%,#303f9f)',
  '标准规范':'linear-gradient(160deg,#37474f,#455a64 50%,#546e7a)',
  '地面站':'linear-gradient(160deg,#bf360c,#d84315 50%,#e64a19)'
}
const STAGE_MAP = { lab:'实验室阶段', pilot:'中试阶段', industrialized:'已产业化', listed:'已上市' }
const STAGE_CLS = { lab:'cl-la', pilot:'cl-pi', industrialized:'cl-in', hot:'cl-ho' }
const TYPE_MAP = { patent:'发明专利', utility:'实用新型', copyright:'软件著作', paper:'论文成果', standard:'技术标准', design:'外观设计' }


const id = ref(''), d = ref(null), loading = ref(true), err = ref(false), vis = ref(false)
const isFav = ref(false), favPop = ref(false), favHide = ref(false)
const titleStuck = ref(false), barHidden = ref(false), lastScroll = ref(0)

const heroIcon = computed(() => d.value ? (ICONS[d.value.field] || '果') : '果')
const heroBg = computed(() => d.value ? (BGS[d.value.field] || BGS['飞控系统']) : BGS['飞控系统'])
const isTransformed = computed(() => {
  const s = (d.value?.stage || d.value?.status || '')
  return ['transformed','industrialization','产业化','已转化','listed'].includes(s)
})
const stageLabel = computed(() => {
  const s = (d.value?.stage || '').toLowerCase()
  return STAGE_MAP[s] || d.value?.stage || ''
})
const stageCls = computed(() => {
  const s = (d.value?.stage || '').toLowerCase()
  return STAGE_CLS[s] || ''
})
const typeLabel = computed(() => {
  const t = (d.value?.achieve_type || '').toLowerCase()
  return TYPE_MAP[t] || d.value?.achieve_type || '未分类'
})

const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true; err.value = false
  try {
    const res = await request({ url: '/api/v1/achievements/' + encodeURIComponent(id.value) })
    const item = res?.data || res
    if (item) {
      d.value = {
        id: item.id,
        owner_id: item.owner_id || '',
        title: item.title || '',
        achieve_type: item.achieve_type || '',
        description: item.description || '',
        field: item.field || '',
        stage: item.stage || '',
        images: item.images || [],
        contact_info: item.contact_info || '',
        status: item.status || '',
        created_at: item.created_at || '',
        updated_at: item.updated_at || ''
      }
      if (d.value.title) uni.setNavigationBarTitle({ title: d.value.title })
    } else {
      err.value = true
    }
  } catch { err.value = true }
  finally { loading.value = false }
  setTimeout(() => { vis.value = true }, 200)
}

const onScroll = (e) => {
  const st = e?.detail?.scrollTop ?? 0
  titleStuck.value = st > 50
  if (st > lastScroll.value && st > 200) barHidden.value = true
  else if (st < lastScroll.value || st < 100) barHidden.value = false
  lastScroll.value = st
}

const toggleFav = () => {
  isFav.value = !isFav.value
  if (isFav.value) {
    favPop.value = true; favHide.value = false
    uni.showToast({ title: '已收藏', icon: 'success', duration: 1200 })
    setTimeout(() => favHide.value = true, 600)
    setTimeout(() => { favPop.value = false; favHide.value = false }, 1000)
  }
}
const toggleShare = () => uni.showToast({ title: '点击右上角分享', icon: 'none', duration: 1500 })
const onContact = () => uni.showToast({ title: '联系对接 (功能待开放)', icon: 'none', duration: 1500 })

onLoad((options) => {
  if (options?.id) { id.value = decodeURIComponent(options.id) }
  fetchData()
})
</script>

<style>
page { background:#f5f6f8 }
</style>
<style scoped>
.page { min-height:100vh; background:#f5f6f8; padding-bottom:env(safe-area-inset-bottom) }

/* Hero */
.hero { position:relative; width:100%; aspect-ratio:16/9; display:flex; flex-direction:column; align-items:center; justify-content:center; overflow:hidden; opacity:0; transform:translateY(-20px); animation:heroIn .5s cubic-bezier(.22,1,.36,1) forwards }
@keyframes heroIn { to { opacity:1; transform:translateY(0) } }
.hero-glow { position:absolute; inset:0; }
.hero-glow::after { content:''; position:absolute; top:-20%; right:-15%; width:200px; height:200px; border-radius:50%; background:radial-gradient(circle,rgba(255,255,255,.1) 0%,transparent 70%) }
.hero-icon { font-size:52px; position:relative; z-index:1; animation:heroFloat 3s ease-in-out infinite; animation-delay:.5s }
@keyframes heroFloat { 0%,100% { transform:translateY(0) } 50% { transform:translateY(-6px) } }
.hero-tag { position:relative; z-index:1; margin-top:10px; font-size:11px; color:rgba(255,255,255,.85); background:rgba(255,255,255,.15); padding:4px 14px; border-radius:12px; opacity:0; animation:fadeUp .4s ease .3s forwards }
@keyframes fadeUp { from { opacity:0; transform:translateY(8px) } to { opacity:1; transform:translateY(0) } }
.hero-wave { position:absolute; bottom:-1px; left:0; right:0; z-index:2; height:24px; overflow:hidden }
.hero-wave-inner { width:200%; height:24px; margin-left:-50%; background:#fff; border-radius:50% 50% 0 0 }

/* Title Sticky */
.ts { padding:20px 16px 12px; background:#fff; position:sticky; top:0; z-index:15; transition:box-shadow .25s }
.ts.stk { box-shadow:0 4px 16px rgba(0,0,0,.08) }
.mt { font-size:18px; font-weight:700; color:#1a1a1a; line-height:1.35; margin-bottom:10px; display:block }
.badges { display:flex; gap:8px; flex-wrap:wrap }
.b { font-size:11px; padding:3px 10px; border-radius:10px; font-weight:500 }
.b-hot { background:#ff3b30; color:#fff }.b-tr { background:#34c759; color:#fff }.b-st { background:#e8f0fe; color:#1967d2 }

/* Cards */
.sec { margin:0 16px 12px; padding:16px; background:#fff; border-radius:14px; box-shadow:0 2px 12px rgba(0,0,0,.03); opacity:0; transform:translateY(20px); transition:opacity .45s ease,transform .45s ease }
.sec.vis { opacity:1; transform:translateY(0) }
.sh { display:flex; align-items:center; gap:8px; margin-bottom:12px }
.sd { width:4px; height:18px; background:#0A66C2; border-radius:2px; flex-shrink:0 }
.sht { font-size:15px; font-weight:700; color:#1a1a1a }
.sb { font-size:14px; color:#666; line-height:1.75; white-space:pre-wrap }

/* Stats */
.stats { display:flex; padding:16px 0 }
.si { flex:1; text-align:center; position:relative }
.si+.si::before { content:''; position:absolute; left:0; top:8px; bottom:8px; width:.5px; background:#f0f0f0 }
.sv { font-size:17px; font-weight:700; color:#1a1a1a; display:block }
.sl { font-size:11px; color:#999; margin-top:2px; display:block }

/* Info Table */
.it { display:flex; flex-direction:column }
.ir2 { display:flex; padding:12px 0; border-bottom:.5px solid #f5f5f5 }
.ir2:last-child { border-bottom:none }
.ik { width:70px; flex-shrink:0; font-size:13px; color:#999 }
.iv { flex:1; font-size:14px; color:#333; word-break:break-all }
.iv.cl-la { color:#1967d2; font-weight:600 }.iv.cl-pi { color:#ff9f0a; font-weight:600 }.iv.cl-in { color:#34c759; font-weight:600 }.iv.cl-ho { color:#ff3b30; font-weight:600 }

/* Attachments */
.al { display:flex; flex-direction:column; gap:8px }
.ai { display:flex; align-items:center; gap:10px; padding:12px; background:#f5f6f8; border-radius:10px }
.ai:active { background:#eef1f5 }
.aic { font-size:18px; flex-shrink:0 }
.aif { flex:1; min-width:0 }
.ain { font-size:13px; color:#1a1a1a; font-weight:500; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; display:block }
.ais { font-size:11px; color:#999; margin-top:2px; display:block }
.aid { font-size:12px; color:#0A66C2; flex-shrink:0; font-weight:500 }

/* Bottom Bar */
.bb { position:sticky; bottom:0; background:#fff; border-top:.5px solid #f0f0f0; display:flex; align-items:center; padding:10px 16px; gap:10px; padding-bottom:calc(10px + env(safe-area-inset-bottom)); box-shadow:0 -2px 12px rgba(0,0,0,.04); transition:transform .3s cubic-bezier(.4,0,.2,1) }
.bb.h { transform:translateY(100%) }
.bi { width:40px; height:40px; border-radius:50%; background:#f5f6f8; display:flex; align-items:center; justify-content:center; flex-shrink:0 }
.bi:active { transform:scale(.88) }
.bi.fv { color:#ff3b30 }
.bit { font-size:20px }
.bp { flex:1; height:42px; border-radius:21px; background:linear-gradient(135deg,#1565c0,#1976d2); color:#fff; font-size:14px; font-weight:600; display:flex; align-items:center; justify-content:center; box-shadow:0 4px 14px rgba(25,118,210,.35); animation:btnPulse 2s ease-in-out infinite; animation-delay:1.5s }
@keyframes btnPulse { 0%,100% { box-shadow:0 4px 14px rgba(25,118,210,.35) } 50% { box-shadow:0 6px 20px rgba(25,118,210,.5) } }
.bp:active { transform:scale(.97) }
.bo { height:42px; border-radius:21px; border:1.5px solid #0A66C2; background:#fff; color:#0A66C2; font-size:14px; font-weight:600; padding:0 16px; display:flex; align-items:center; flex-shrink:0 }
.bo:active { background:#e8f0fe }

/* Fav Pop */
.fp { position:fixed; top:50%; left:50%; transform:translate(-50%,-50%) scale(0); font-size:48px; color:#ff3b30; z-index:100; pointer-events:none; transition:transform .4s cubic-bezier(.17,.89,.32,1.9),opacity .3s; opacity:0 }
.fp:not(.hide) { transform:translate(-50%,-50%) scale(1); opacity:1 }
.fp.hide { transform:translate(-50%,-50%) scale(1.2); opacity:0 }

/* Skeleton */
.sk-h { aspect-ratio:16/9; background:#f0f1f3; animation:shimmer 1.5s infinite }
.sk-stats { display:flex; margin:12px 16px; gap:8px }
.sk-stat { flex:1; height:60px; background:#f0f1f3; border-radius:14px; animation:shimmer 1.5s infinite }
.sk-sec { margin:0 16px 12px; padding:16px; background:#fff; border-radius:14px }
.sk-l { height:14px; background:#f0f1f3; border-radius:4px; margin-bottom:8px; animation:shimmer 1.5s infinite }
.sk-l.w80 { width:80% }.sk-l.w100 { width:100% }.sk-l.w60 { width:60% }.sk-l.w40 { width:40% }
@keyframes shimmer { 0%,100% { opacity:1 } 50% { opacity:.45 } }

/* State */
.st { display:flex; flex-direction:column; align-items:center; justify-content:center; padding:100px 20px; min-height:400px }
.sti { font-size:48px; margin-bottom:12px; opacity:.5 }
.stt { font-size:14px; color:#999; margin-bottom:4px; display:block }
.sth { font-size:12px; color:#bbb; margin-bottom:16px; display:block }
.stb { padding:8px 24px; border-radius:22px; background:#0A66C2; color:#fff; font-size:13px; font-weight:500 }
</style>
