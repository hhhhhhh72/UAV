<template>
  <view class="page">
    <!-- Skeleton -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-stats"><view class="sk-stat"></view><view class="sk-stat"></view><view class="sk-stat"></view></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- Error -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- Empty -->
    <view v-else-if="!d" class="st">
      <u-empty description="该成果已下架或不存在">
        <view class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <!-- Content -->
    <template v-else>
      <!-- Hero -->
      <view class="hero" :style="heroStyle">
        <view class="hero-glow"></view>
        <text class="hero-ic">{{ heroIcon }}</text>
        <view class="hero-tag" v-if="d.field"><text>{{ d.field }}</text></view>
        <view class="hero-tag hero-tag-st" v-if="stageLabel" :class="stageCls"><text>{{ stageLabel }}</text></view>
        <view class="hero-wave"><view class="hero-wave-inner"></view></view>
      </view>

      <!-- Title + Badges -->
      <view class="ts" :class="{ stk: titleStuck }">
        <text class="mt">{{ d.title }}</text>
        <view class="badges">
          <text v-if="isHot" class="b b-hot">热门</text>
          <text v-if="isTransformed" class="b b-tr">已转化</text>
          <text v-if="stageLabel" class="b b-st">{{ stageLabel }}</text>
        </view>
      </view>

      <!-- Stats Row -->
      <view class="sec stats" :class="{ vis: vis }">
        <view class="si"><text class="sv">{{ d.date }}</text><text class="sl">发布日期</text></view>
        <view class="si"><text class="sv">{{ typeLabel }}</text><text class="sl">成果类型</text></view>
        <view class="si"><text class="sv">{{ d.owner_id ? '已认证' : '未认证' }}</text><text class="sl">发布方</text></view>
      </view>

      <!-- Description -->
      <view class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">成果描述</text></view>
        <text class="sb">{{ d.description || '暂无描述' }}</text>
      </view>

      <!-- Basic Info -->
      <view class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">基本信息</text></view>
        <view class="it">
          <view v-if="d.field" class="ir2"><text class="ik">所属领域</text><text class="iv">{{ d.field }}</text></view>
          <view v-if="d.achieve_type" class="ir2"><text class="ik">成果类型</text><text class="iv">{{ typeLabel }}</text></view>
          <view v-if="d.stage" class="ir2"><text class="ik">成果阶段</text><text class="iv" :class="stageCls">{{ stageLabel }}</text></view>
          <view v-if="d.contact_info" class="ir2"><text class="ik">联系方式</text><text class="iv">{{ d.contact_info }}</text></view>
          <view v-if="d.owner_id" class="ir2"><text class="ik">发布方</text><text class="iv">平台认证成员</text></view>
          <view v-if="d.created_at" class="ir2"><text class="ik">创建时间</text><text class="iv">{{ d.date }}</text></view>
          <view v-if="d.updated_at && d.updated_at !== d.created_at" class="ir2"><text class="ik">更新时间</text><text class="iv">{{ (d.updated_at || '').slice(0, 10) }}</text></view>
        </view>
      </view>

      <!-- 转化进展（管理后台录入 → 小程序展示；点击进入成果转化页） -->
      <view v-if="transforms.length" class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">转化进展</text></view>
        <view v-for="(t, i) in transforms" :key="i" class="tr-card">
          <text class="tr-title">{{ t.title }}</text>
          <!-- 流程：虚线轨道 + 进度动画（实验室→中试→产业化→上市） -->
          <view class="tr-flow">
            <view class="tr-track">
              <view class="tr-base"></view>
              <view class="tr-prog" :style="{ width: flowReady ? flowPctOf(t.stage) : '0%' }"></view>
              <view class="tr-stages">
                <view
                  v-for="(st, si) in stages"
                  :key="si"
                  class="tr-stage"
                  :class="{ done: stageRank(t.stage) >= si + 1, cur: stageRank(t.stage) === si + 1 }"
                >
                  <view class="tr-dot"></view>
                  <text class="tr-stage-name">{{ st }}</text>
                </view>
              </view>
            </view>
            <view class="tr-meta">
              <text>已推进 {{ stageRank(t.stage) || 0 }} / 4 阶段</text>
              <text class="tr-cur">当前：{{ stageLabelOf(t.stage) }}</text>
            </view>
          </view>
          <view v-if="t.progress" class="tr-progress"><text class="tr-k">当前进度</text><text class="tr-v">{{ t.progress }}</text></view>
          <view v-if="t.partner_id" class="tr-progress"><text class="tr-k">合作单位</text><text class="tr-v">{{ t.partner_id }}</text></view>
        </view>
        <view class="tr-go" @tap="goTrack"><text>查看成果转化详情 ›</text></view>
      </view>

      <!-- 附件资料 -->
      <view v-if="d.attachments && d.attachments.length" class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">附件资料</text></view>
        <view v-for="(at, i) in d.attachments" :key="i" class="at-row" @tap="downloadAt(at)">
          <view class="at-ic"><text>附</text></view>
          <view class="at-info">
            <text class="at-name">{{ at.name || '附件' }}</text>
            <text class="at-size">{{ at.size || '' }}</text>
          </view>
          <text class="at-btn">下载</text>
        </view>
      </view>

      <view style="height: 160rpx"></view>

      <!-- 底部操作栏 -->
      <view class="bb" :class="{ h: barHidden }">
        <view class="bi" :class="{ fv: isFav }" @tap="toggleFav"><text class="bit">{{ isFav ? '★' : '☆' }}</text></view>
        <view class="bo" @tap="goTrack">成果转化</view>
        <view class="bp" @tap="onContact">联系对接</view>
      </view>
      <view class="fp" :class="favPop ? (favHide ? 'hide' : '') : 'hide'">★</view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPageScroll } from '@dcloudio/uni-app'
import { request, BASE_URL } from '@/utils/request'
import { MOCK_ACHIEVEMENTS, MOCK_TRANSFORMS_BY_ACH, ACH_TYPE_LABEL, STAGE_LABEL, STAGE_RANK } from '@/utils/mockAchievements'

const goBack = () => uni.navigateBack()

const FIELD_ICON = { '飞控系统': '飞', '遥感测绘': '遥', '动力系统': '动', 'AI算法': '算', '载荷设备': '载', '集群协同': '群', '通信链路': '通', '标准规范': '标', '地面站': '地' }
const FIELD_BG = {
  '飞控系统': '#0d47a1', '遥感测绘': '#1b5e20', '动力系统': '#e65100', 'AI算法': '#4a148c',
  '载荷设备': '#b71c1c', '集群协同': '#004d40', '通信链路': '#1a237e', '标准规范': '#37474f', '地面站': '#bf360c'
}
const STAGE_CLS = { lab: 'cl-la', laboratory: 'cl-la', pilot: 'cl-pi', industrialization: 'cl-in', industrialized: 'cl-in', listed: 'cl-in', hot: 'cl-ho' }

const id = ref('')
const d = ref(null)
const loading = ref(true)
const err = ref(false)
const vis = ref(false)
const isFav = ref(false)
const favPop = ref(false)
const favHide = ref(false)
const titleStuck = ref(false)
const barHidden = ref(false)
let lastScroll = 0

const heroIcon = computed(() => {
  if (!d.value) return '果'
  return FIELD_ICON[d.value.field] || (d.value.field ? d.value.field.charAt(0) : '果')
})
const heroStyle = computed(() => {
  if (!d.value) return { background: FIELD_BG['飞控系统'] }
  if (d.value.img) {
    return { backgroundImage: 'url(' + d.value.img + ')', backgroundSize: 'cover', backgroundPosition: 'center' }
  }
  return { background: FIELD_BG[d.value.field] || FIELD_BG['飞控系统'] }
})
const isHot = computed(() => (d.value?.status || '').toLowerCase() === 'hot')
const isTransformed = computed(() => {
  const s = (d.value?.stage || d.value?.status || '').toLowerCase()
  return ['transformed', 'industrialization', 'industrialized', 'listed', '已转化', '产业化'].includes(s)
})
const stageLabel = computed(() => {
  const s = (d.value?.stage || '').toLowerCase()
  return STAGE_LABEL[s] || d.value?.stage || ''
})
const stageCls = computed(() => {
  const s = (d.value?.stage || '').toLowerCase()
  return STAGE_CLS[s] || ''
})
const transforms = ref([])
const stages = ['实验室', '中试', '产业化', '上市']
const stageRank = (s) => STAGE_RANK[(s || '').toLowerCase()] || 0
const flowReady = ref(false)
const flowPctOf = (s) => {
  const r = STAGE_RANK[(s || '').toLowerCase()] || 0
  return Math.max(0, Math.min(75, (r - 1) * 25)) + '%'
}
const stageLabelOf = (s) => STAGE_LABEL[(s || '').toLowerCase()] || s || ''
const applyTransforms = (list) => {
  transforms.value = list
  flowReady.value = false
  setTimeout(() => { flowReady.value = true }, 300)
}
const fetchTransformations = async (aid) => {
  try {
    const res = await request({ url: '/api/v1/transformations', data: { achievement_id: aid } })
    applyTransforms(Array.isArray(res) ? res : (res?.data || []))
  } catch (e) { applyTransforms([]) }
}
const typeLabel = computed(() => {
  const t = (d.value?.achieve_type || '').toLowerCase()
  return ACH_TYPE_LABEL[t] || d.value?.achieve_type || '未分类'
})

const imgSrc = (images) => {
  let arr = images
  if (typeof images === 'string') {
    try { arr = JSON.parse(images) } catch { return '' }
  }
  if (!Array.isArray(arr) || !arr.length) return ''
  const u = arr[0]
  return u ? (u.startsWith('http') ? u : BASE_URL + u) : ''
}

// ===== 数据获取 =====
// 接口替换点：GET /api/v1/achievements/{id} + GET /api/v1/transformations?achievement_id=
const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    const res = await request({ url: '/api/v1/achievements/' + encodeURIComponent(id.value) })
    const item = res?.data || res
    if (item) {
      applyDetail(item)
      fetchTransformations(item.id)
    } else {
      err.value = true
    }
  } catch {
    useMock()
  } finally {
    loading.value = false
  }
  setTimeout(() => { vis.value = true }, 200)
}

const applyDetail = (item) => {
  d.value = {
    id: item.id,
    owner_id: item.owner_id || '',
    title: item.title || '',
    achieve_type: item.achieve_type || '',
    description: item.description || '',
    field: item.field || '',
    stage: item.stage || '',
    images: item.images || [],
    img: imgSrc(item.images),
    attachments: item.attachments || [],
    contact_info: item.contact_info || '',
    status: item.status || '',
    created_at: item.created_at || '',
    updated_at: item.updated_at || '',
    date: (item.created_at || '').slice(0, 10)
  }
  if (d.value.title) uni.setNavigationBarTitle({ title: d.value.title })
}

// 演示数据回退（仅 demo- 前缀 id）
const useMock = () => {
  const mock = MOCK_ACHIEVEMENTS.find((x) => x.id === id.value)
  if (mock) {
    applyDetail(mock)
    applyTransforms(MOCK_TRANSFORMS_BY_ACH[id.value] || [])
  } else {
    err.value = true
  }
}

onPageScroll((e) => {
  const st = e.scrollTop || 0
  titleStuck.value = st > 60
  if (st > lastScroll && st > 200) barHidden.value = true
  else if (st < lastScroll || st < 100) barHidden.value = false
  lastScroll = st
})

const toggleFav = () => {
  isFav.value = !isFav.value
  if (isFav.value) {
    favPop.value = true
    favHide.value = false
    uni.showToast({ title: '已收藏', icon: 'success', duration: 1200 })
    setTimeout(() => { favHide.value = true }, 600)
    setTimeout(() => { favPop.value = false; favHide.value = false }, 1000)
  }
}
const onContact = () => uni.showToast({ title: '联系对接功能待开放', icon: 'none', duration: 1500 })

// 进入成果转化页（详情 → 转化）
const goTrack = () => {
  if (!transforms.value.length) {
    uni.showToast({ title: '暂无转化进展', icon: 'none' })
    return
  }
  const t = transforms.value[0]
  uni.navigateTo({ url: '/pkg-eco/pages/transformations/track?achievement_id=' + encodeURIComponent(id.value) + '&id=' + encodeURIComponent(t.id || '') })
}

const downloadAt = (at) => {
  if (!at?.url) { uni.showToast({ title: '附件地址缺失', icon: 'none' }); return }
  uni.downloadFile({
    url: at.url.startsWith('http') ? at.url : BASE_URL + at.url,
    success: (res) => {
      if (res.statusCode === 200) {
        uni.openDocument({ filePath: res.tempFilePath, showMenu: true, fail: () => uni.showToast({ title: '无法预览，请用浏览器打开', icon: 'none' }) })
      } else {
        uni.showToast({ title: '下载失败', icon: 'none' })
      }
    },
    fail: () => uni.showToast({ title: '下载失败', icon: 'none' })
  })
}

onLoad((options) => {
  if (options?.id) id.value = decodeURIComponent(options.id)
  fetchData()
})
</script>

<style>
page { background: var(--color-bg); }
</style>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ===== Hero（保留原动画：heroIn / heroFloat / fadeUp） ===== */
.hero { position: relative; width: 100%; aspect-ratio: 16/9; display: flex; flex-direction: column; align-items: center; justify-content: center; overflow: hidden; opacity: 0; transform: translateY(-20px); animation: heroIn .5s cubic-bezier(.22,1,.36,1) forwards; }
@keyframes heroIn { to { opacity: 1; transform: translateY(0); } }
.hero-glow { position: absolute; inset: 0; }
.hero-glow::after { content: ''; position: absolute; top: -20%; right: -15%; width: 200px; height: 200px; border-radius: 50%; background: radial-gradient(circle, rgba(255,255,255,.1) 0%, transparent 70%); }
.hero-ic { font-size: 52px; position: relative; z-index: 1; color: #fff; text-shadow: 0 4px 12px rgba(0,0,0,.25); animation: heroFloat 3s ease-in-out infinite; animation-delay: .5s; }
@keyframes heroFloat { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-6px); } }
.hero-tag { position: relative; z-index: 1; margin-top: 10px; font-size: 11px; color: rgba(255,255,255,.85); background: rgba(255,255,255,.15); padding: 4px 14px; border-radius: 8px; letter-spacing: .5px; opacity: 0; animation: fadeUp .4s ease .3s forwards; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
/* 转化阶段徽章（新增，沿用原标签风格） */
.hero-tag.hero-tag-st { margin-top: 6px; color: #1967d2; background: rgba(255,255,255,.92); font-weight: 600; }
.hero-tag.hero-tag-st.cl-pi { color: #e65100; }
.hero-tag.hero-tag-st.cl-in { color: #168a55; }
.hero-wave { position: absolute; bottom: -1px; left: 0; right: 0; z-index: 2; height: 24px; overflow: hidden; }
.hero-wave-inner { width: 200%; height: 24px; margin-left: -50%; background: #fff; border-radius: 50% 50% 0 0; }

/* ===== Title Sticky（保留） ===== */
.ts { padding: 40rpx 32rpx 24rpx; background: #fff; position: sticky; top: 0; z-index: 15; transition: box-shadow .25s; }
.ts.stk { box-shadow: 0 4px 16px rgba(0,0,0,.08); }
.mt { font-size: 36rpx; font-weight: 700; color: var(--color-text); line-height: 1.35; margin-bottom: 20rpx; display: block; }
.badges { display: flex; gap: 16rpx; flex-wrap: wrap; }
.b { font-size: 22rpx; padding: 6rpx 20rpx; border-radius: 16rpx; font-weight: 500; }
.b-hot { background: var(--color-danger); color: #fff; }
.b-tr { background: var(--color-success); color: #fff; }
.b-st { background: #e8f0fe; color: #1967d2; }

/* ===== Sections（保留 .vis 上浮动画） ===== */
.sec { margin: 0 32rpx 24rpx; padding: 32rpx; background: #fff; border-radius: 16rpx; box-shadow: 0 2px 12px rgba(0,0,0,.03); opacity: 0; transform: translateY(20px); transition: opacity .45s ease, transform .45s ease; }
.sec.vis { opacity: 1; transform: translateY(0); }
.sh { display: flex; align-items: center; gap: 16rpx; margin-bottom: 24rpx; }
.sd { width: 8rpx; height: 36rpx; background: var(--color-primary); border-radius: 4rpx; flex-shrink: 0; }
.sht { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.sb { font-size: 28rpx; color: var(--color-text-secondary); line-height: 1.75; white-space: pre-wrap; }

/* ===== Stats（保留） ===== */
.stats { display: flex; padding: 32rpx 0; }
.si { flex: 1; text-align: center; position: relative; }
.si + .si::before { content: ''; position: absolute; left: 0; top: 8px; bottom: 8px; width: .5px; background: #f0f0f0; }
.sv { font-size: 34rpx; font-weight: 700; color: var(--color-text); display: block; }
.sl { font-size: 22rpx; color: var(--color-text-placeholder); margin-top: 4rpx; display: block; }

/* ===== Info Table（保留） ===== */
.it { display: flex; flex-direction: column; }
.ir2 { display: flex; padding: 24rpx 0; border-bottom: .5px solid #f5f5f5; }
.ir2:last-child { border-bottom: none; }
.ik { width: 140rpx; flex-shrink: 0; font-size: 26rpx; color: var(--color-text-placeholder); }
.iv { flex: 1; font-size: 28rpx; color: var(--color-text); word-break: break-all; }
.iv.cl-la { color: #1967d2; font-weight: 600; }
.iv.cl-pi { color: var(--color-warning); font-weight: 600; }
.iv.cl-in { color: var(--color-success); font-weight: 600; }
.iv.cl-ho { color: var(--color-danger); font-weight: 600; }

/* ===== 附件资料（保留） ===== */
.at-row { display: flex; align-items: center; gap: 16rpx; padding: 20rpx 0; border-bottom: .5px solid #f5f5f5; }
.at-row:last-child { border-bottom: none; }
.at-ic { width: 56rpx; height: 56rpx; border-radius: 12rpx; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.at-ic text { font-size: 24rpx; color: var(--color-primary); font-weight: 600; }
.at-info { flex: 1; min-width: 0; }
.at-name { display: block; font-size: 26rpx; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.at-size { display: block; font-size: 20rpx; color: var(--color-text-placeholder); margin-top: 4rpx; }
.at-btn { flex-shrink: 0; padding: 8rpx 20rpx; border-radius: 8rpx; background: var(--color-primary); color: #fff; font-size: 22rpx; }

/* ===== 转化进展（保留原时间线；新增查看入口） ===== */
.tr-card { padding: 8rpx 0 4rpx; display: flex; flex-direction: column; gap: 16rpx; }
.tr-title { display: block; font-size: 28rpx; font-weight: 600; color: var(--color-text); margin: 0; }
/* 流程：虚线轨道 + 已完成进度动画；间距用 flex/gap 自适应，不写死 */
.tr-flow { position: relative; display: flex; flex-direction: column; gap: 16rpx; }
.tr-track { position: relative; padding: 6rpx 0 10rpx; }
.tr-base { position: absolute; left: 12.5%; right: 12.5%; top: 10rpx; border-top: 2rpx dashed #C7D2DE; z-index: 0; }
.tr-prog { position: absolute; left: 12.5%; top: 9rpx; height: 4rpx; background: linear-gradient(90deg, var(--color-primary), #42a5f5); border-radius: 2rpx; z-index: 1; transition: width 1.1s cubic-bezier(.2,.9,.3,1); }
.tr-stages { display: flex; justify-content: space-between; position: relative; z-index: 2; }
.tr-stage { display: flex; flex-direction: column; align-items: center; gap: 8rpx; width: 25%; }
.tr-dot { width: 20rpx; height: 20rpx; border-radius: 50%; background: #fff; border: 4rpx solid var(--color-divider); box-sizing: border-box; transition: all .3s; }
.tr-stage.done .tr-dot { background: var(--color-primary); border-color: var(--color-primary); }
.tr-stage.cur .tr-dot { background: var(--color-primary); border-color: var(--color-primary); animation: dotPulse 1.8s ease-out infinite; }
@keyframes dotPulse { 0% { box-shadow: 0 0 0 0 rgba(10,102,194,.4); } 70% { box-shadow: 0 0 0 14rpx rgba(10,102,194,0); } 100% { box-shadow: 0 0 0 0 rgba(10,102,194,0); } }
.tr-stage-name { font-size: 20rpx; color: var(--color-text-placeholder); }
.tr-stage.done .tr-stage-name, .tr-stage.cur .tr-stage-name { color: var(--color-primary); font-weight: 600; }
.tr-meta { display: flex; justify-content: space-between; align-items: center; font-size: 20rpx; color: var(--color-text-placeholder); }
.tr-cur { color: var(--color-primary); font-weight: 600; }
.tr-progress { display: flex; gap: 12rpx; padding: 10rpx 0; border-top: .5px solid #f5f5f5; }
.tr-k { width: 130rpx; flex-shrink: 0; font-size: 24rpx; color: var(--color-text-placeholder); }
.tr-v { flex: 1; font-size: 24rpx; color: var(--color-text); }
.tr-go { margin-top: 20rpx; padding: 20rpx 0 4rpx; border-top: 1rpx dashed var(--color-border); text-align: center; font-size: 26rpx; color: var(--color-primary); font-weight: 600; }
.tr-go:active { opacity: .7; }

/* ===== 底部操作栏（保留原有按钮风格与隐藏动画） ===== */
.bb { position: sticky; bottom: 0; z-index: 50; background: #fff; border-top: .5px solid #f0f0f0; display: flex; align-items: center; padding: 20rpx 32rpx; gap: 20rpx; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); box-shadow: 0 -2px 12px rgba(0,0,0,.04); transition: transform .3s cubic-bezier(.4,0,.2,1); }
.bb.h { transform: translateY(100%); }
.bi { width: 80rpx; height: 80rpx; border-radius: 50%; background: var(--color-bg); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.bi:active { transform: scale(.88); }
.bi.fv { color: #F5A623; }
.bit { font-size: 40rpx; line-height: 1; }
.bp { flex: 1; height: 84rpx; border-radius: 16rpx; background: #1565c0; color: #fff; font-size: 28rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; box-shadow: 0 4px 14px rgba(25,118,210,.35); animation: btnPulse 2s ease-in-out infinite; animation-delay: 1.5s; }
@keyframes btnPulse { 0%, 100% { box-shadow: 0 4px 14px rgba(25,118,210,.35); } 50% { box-shadow: 0 6px 20px rgba(25,118,210,.5); } }
.bp:active { transform: scale(.97); }
.bo { height: 84rpx; border-radius: 16rpx; border: 1.5px solid var(--color-primary); background: #fff; color: var(--color-primary); font-size: 28rpx; font-weight: 600; padding: 0 32rpx; display: flex; align-items: center; flex-shrink: 0; }
.bo:active { background: #e8f0fe; }

/* ===== 收藏弹心（保留） ===== */
.fp { position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%) scale(0); font-size: 48px; color: #F5A623; z-index: 100; pointer-events: none; transition: transform .4s cubic-bezier(.17,.89,.32,1.9), opacity .3s; opacity: 0; }
.fp:not(.hide) { transform: translate(-50%, -50%) scale(1); opacity: 1; }
.fp.hide { transform: translate(-50%, -50%) scale(1.2); opacity: 0; }

/* ===== Skeleton（保留） ===== */
.sk-h { aspect-ratio: 16/9; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.sk-stats { display: flex; margin: 24rpx 32rpx; gap: 16rpx; }
.sk-stat { flex: 1; height: 120rpx; background: #f0f1f3; border-radius: 16rpx; animation: shimmer 1.5s infinite; }
.sk-sec { margin: 0 32rpx 24rpx; padding: 32rpx; background: #fff; border-radius: 16rpx; }
.sk-l { height: 28rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== State（保留） ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 40rpx; min-height: 800rpx; }
.stb { padding: 16rpx 48rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 26rpx; font-weight: 500; }
.stb:active { opacity: .8; }

/* ===================== UI/UX 体验优化（仅新增/修改 wxss，不动模板/数据/逻辑） ===================== */
/* 动画统一 200-400ms；优先 transform/opacity；生产级轻量克制 */

/* 1) 入场动画：附件行、进度行依次淡入（60ms 错开；backwards） */
.sec .at-row { animation: uiRowIn .3s ease backwards; }
.sec .at-row:nth-child(2) { animation-delay: 0ms; }
.sec .at-row:nth-child(3) { animation-delay: 60ms; }
.tr-card .tr-progress { animation: uiRowIn .3s ease backwards; }
.tr-card .tr-progress:nth-child(3) { animation-delay: 60ms; }
.tr-card .tr-progress:nth-child(4) { animation-delay: 120ms; }
@keyframes uiRowIn { from { opacity: 0; transform: translateX(10rpx); } to { opacity: 1; transform: translateX(0); } }

/* 2) 交互反馈：列表行/按钮按压轻微缩放 + 透明度（200ms） */
.at-row { transition: transform .2s ease, opacity .2s ease; }
.at-row:active { transform: scale(.99); opacity: .8; }
.tr-go { transition: transform .2s ease, opacity .2s ease; }
.tr-go:active { transform: scale(.98); opacity: .75; }
.bi { transition: transform .2s ease, color .2s ease, background .2s ease; }
.bo { transition: transform .2s ease, background .2s ease, color .2s ease; }
.bo:active { transform: scale(.97); opacity: .9; }
.bp { transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease; }
.bp:active { transform: scale(.97); opacity: .92; }
.stb { transition: transform .2s ease, opacity .2s ease; }
.stb:active { transform: scale(.95); opacity: .85; }

/* 3) 状态过渡：阶段圆点/徽章切换 300ms 平滑 */
.tr-dot { transition: background .3s ease, border-color .3s ease, box-shadow .3s ease; }
.b { transition: transform .2s ease, opacity .2s ease; }

/* 4) 层级加固：底部操作栏置顶于内容之上；收藏弹心最高层 */
.bb { z-index: 60; }
.fp { z-index: 100; }

/* ===== 【首页风格】同步 pages/home 样式（仅颜色/圆角/阴影/字重；如需回退删除本块即可） ===== */
/* 区块卡片：对齐首页 demand-card */
.sec { border-radius: 16rpx; box-shadow: 0 4rpx 16rpx rgba(16,24,40,.035); }
/* 徽章：对齐首页 badge */
.b { border-radius: 8rpx; font-weight: 700; }
.b-hot { background: #F04438; }
.b-tr { background: #168A55; }
.b-st { background: #EAF3FB; color: #074D92; }
/* 主/次按钮：对齐首页按钮（主色 #0A66C2 / 圆角 6px=12rpx） */
.bp { background: #0A66C2; border-radius: 12rpx; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.3); }
.bo { border-radius: 12rpx; }
.stb { background: #0A66C2; border-radius: 12rpx; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.3); }
.at-btn { background: #0A66C2; border-radius: 12rpx; }
</style>
