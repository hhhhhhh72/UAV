<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="难题详情" show-back :fixed="true" @back="goBack" />

    <!-- 骨架 -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- 错误 -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- 空 -->
    <view v-else-if="!d" class="st">
      <u-empty description="该难题已下架或不存在">
        <view class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <template v-else>
      <!-- 领域色 Hero：左缘领域色竖条 + 白底领域胶囊（深字浅底对，随领域换色）——领域身份锚定，与 list 页同源配色 -->
      <view class="hero">
        <view class="hero-bar" :style="{ background: d.fieldC }"></view>
        <view class="h-tags">
          <view class="field-pill">
            <view class="fp-dot" :style="{ background: d.fieldBg, color: d.fieldC }">{{ d.fieldChar }}</view>
            <text class="fp-txt" :style="{ color: d.fieldC }">{{ d.f }}</text>
          </view>
          <text class="h-tag" :class="'st-' + d.stCls">{{ d.stLabel }}</text>
        </view>
        <text class="h-title">{{ d.t }}</text>
        <view class="h-budget">
          <text class="h-vl" :class="{ face: d.isFace }">{{ d.budgetText }}</text>
          <text class="h-lb">悬赏金额 · 择优揭榜</text>
        </view>
      </view>

      <!-- 信息卡 -->
      <view class="info">
        <view class="stat">
          <view class="si"><text class="sv">{{ d.budgetText }}</text><text class="sl">悬赏金额</text></view>
          <view class="si"><text class="sv">{{ d.daysLeft }}</text><text class="sl">剩余天数</text></view>
          <view class="si"><text class="sv">{{ d.dateShort }}</text><text class="sl">发布时间</text></view>
          <!-- 揭榜人数：数据未到前显示 -（数字诚实铁律） -->
          <view class="si"><text class="sv">{{ claimCountText }}</text><text class="sl">揭榜人数</text></view>
        </view>
        <view class="row">
          <view class="ic">主</view>
          <view class="it"><text class="il">发布单位</text><text class="iv">{{ d.organizer }}</text></view>
        </view>
        <view class="row">
          <view class="ic orange">域</view>
          <view class="it"><text class="il">所属领域</text><text class="iv">{{ d.f }}</text></view>
        </view>
        <view class="row">
          <view class="ic green">止</view>
          <view class="it"><text class="il">截止日期</text><text class="iv">{{ d.deadlineText }}</text></view>
        </view>
      </view>

      <!-- 难题描述 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">难题描述</text></view>
        <text class="p">{{ d.desc || '暂无描述' }}</text>
      </view>

      <!-- 攻关要求 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">攻关要求</text></view>
        <view v-if="d.reqs.length" class="ul">
          <view v-for="(r, i) in d.reqs" :key="i" class="li">
            <view class="dot"></view>
            <text class="li-t">{{ r }}</text>
          </view>
        </view>
        <text v-else class="p dim">暂无具体指标，详情以发布单位沟通为准</text>
      </view>

      <!-- 揭榜流程 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">揭榜流程</text></view>
        <view class="steps">
          <view class="step"><view class="no">1</view><text class="txt">提交<br/>揭榜意向</text></view>
          <view class="step"><view class="no">2</view><text class="txt">协会<br/>审核对接</text></view>
          <view class="step"><view class="no">3</view><text class="txt">达成合作<br/>攻关结题</text></view>
        </view>
      </view>

      <!-- 揭榜动态：聚合数 + 脱敏动态（展示名一律脱敏，公信力铁律） -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">揭榜动态</text></view>
        <view class="cl-sum">
          <text v-if="claimStatus === 'loading'" class="cl-dim">正在加载揭榜动态…</text>
          <text v-else-if="claimStatus === 'error'" class="cl-dim">揭榜动态加载失败</text>
          <text v-else-if="claimCount > 0"><text class="cl-num">{{ claimCount }}</text> 份揭榜意向已提交 · 协会审核对接中</text>
          <text v-else class="cl-dim">暂无揭榜记录，等你来揭榜</text>
        </view>
        <view v-if="claimStatus === 'ok' && claimItems.length" class="cl-list">
          <view v-for="c in claimItems" :key="c.id" class="cl-it">
            <view class="cl-av">{{ c.avatar }}</view>
            <view class="cl-mid">
              <text class="cl-name">{{ c.claimer }}</text>
              <text class="cl-tag" :class="'ct-' + c.statusCls">{{ c.statusLabel }}</text>
            </view>
            <text class="cl-date">{{ c.created_at }}</text>
          </view>
        </view>
      </view>
      <view style="height: 100px"></view>

      <!-- 底部操作栏（布局对齐 achievements/detail：底部吸附 + safe-area + 44px 按钮） -->
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" @tap="toggleFav"><view class="bit"></view></view>
        <button class="bo" open-type="share" hover-class="bo-hover" hover-start-time="0" hover-stay-time="300">转发</button>
        <view class="bp" :class="{ 'bp-off': d.stCls === 'closed' || claimed || submitting }" @tap="submitClaim">{{ submitLabel }}</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request, authStorage, getErrorMessage } from '@/utils/request'
import { MOCK_CHALLENGES } from '@/utils/mockChallenges'
import { useReduceMotion } from '@/utils/motion'

// 生产环境禁止演示数据回退：接口异常时如实呈现失败态（数字诚实铁律）
const isProduction = typeof process !== 'undefined' && process.env.NODE_ENV === 'production'

const loading = ref(true)
const err = ref(false)
const d = ref(null)
const isFav = ref(false)
const statusBarHeight = ref(20)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
let id = ''
/* 收藏持久化：本地存储兜底（后端收藏接口就绪前的纯前端实现），按难题 id 去重 */
const FAV_KEY = 'challenge_favs'
const loadFavs = () => {
  try {
    const v = uni.getStorageSync(FAV_KEY)
    return Array.isArray(v) ? new Set(v) : new Set()
  } catch (e) { return new Set() }
}
const favs = loadFavs()
const saveFavs = () => {
  try { uni.setStorageSync(FAV_KEY, [...favs]) } catch (e) { /* 忽略 */ }
}

const pad = (n) => (n < 10 ? '0' + n : '' + n)
const daysLeft = (dt) => {
  if (!dt) return null
  const diff = new Date(dt) - new Date()
  return Number.isFinite(diff) ? Math.max(0, Math.ceil(diff / 86400000)) : null
}
const statusOf = (it) => {
  const s = String(it.status || '').toLowerCase()
  const dl = daysLeft(it.deadline)
  if (s === 'closed' || s === '已截止') return { label: '已截止', cls: 'closed' }
  if (s === 'urgent' || s === '紧急' || (dl != null && dl <= 7)) return { label: '紧急', cls: 'urgent' }
  return { label: '进行中', cls: 'open' }
}
const fmtMoney = (wan) => {
  if (wan == null || wan <= 0) return '面议'
  if (wan >= 1) return '¥' + (wan % 1 === 0 ? wan : wan.toFixed(1)) + '万'
  return '¥' + Math.round(wan * 10000)
}
// 攻关要求：仅展示后端真实提供的 requirements；缺省时走模板空态文案，
// 不伪造指标（协会公信力铁律——绝不把平台编造的技术规格当作发布方要求展示）

// 领域色身份：与 challenges/list.vue 同源映射（列表→详情身份连续；深字浅底对，对比度 ≥4.5:1）
const FIELD_ALIAS = {
  '飞控': '飞控系统', '飞控系统': '飞控系统',
  '电池': '动力电池', '动力电池': '动力电池',
  'AI': 'AI算法', 'AI算法': 'AI算法',
  '通信': '通信链路', '通信链路': '通信链路',
  '材料': '新型材料', '新型材料': '新型材料',
  '载荷': '载荷设备', '载荷设备': '载荷设备',
  '集群': '集群协同', '集群协同': '集群协同',
}
const FIELD_TAG = {
  '飞控系统': { tagC: '#0d47a1', tagBg: '#E3EDF9' },
  '动力电池': { tagC: '#B54708', tagBg: '#FDEEE4' },
  'AI算法': { tagC: '#4a148c', tagBg: '#F0E9F7' },
  '通信链路': { tagC: '#1a237e', tagBg: '#E7E9F4' },
  '新型材料': { tagC: '#004d40', tagBg: '#E4F2EF' },
  '载荷设备': { tagC: '#b71c1c', tagBg: '#FBE9E9' },
  '集群协同': { tagC: '#0e7490', tagBg: '#E5F3F8' },
}
const FIELD_TAG_DEFAULT = { tagC: '#344054', tagBg: '#EEF1F4' }
const normField = (f) => FIELD_ALIAS[f] || f || '其他'

const mapItem = (it) => {
  const dl = daysLeft(it.deadline)
  const st = statusOf(it)
  const wan = it.budget_fen != null ? it.budget_fen / 100 / 10000 : 0
  const field = normField(it.field)
  const ft = FIELD_TAG[field] || FIELD_TAG_DEFAULT
  let reqs = []
  if (Array.isArray(it.requirements) && it.requirements.length) reqs = it.requirements
  else if (typeof it.requirements === 'string' && it.requirements.trim()) {
    reqs = it.requirements.split(/[\n;；]/).map((s) => s.trim()).filter(Boolean)
  }
  return {
    id: it.id,
    t: it.title || '未命名难题',
    f: field,
    fieldC: ft.tagC,
    fieldBg: ft.tagBg,
    fieldChar: field.charAt(0),
    desc: it.description || '',
    budgetText: fmtMoney(wan),
    isFace: wan <= 0, // 面议：无金额可展示，hero 降重（不伪造金额）
    stLabel: st.label,
    stCls: st.cls,
    daysLeft: dl == null ? '不限' : dl + ' 天',
    dateShort: (it.created_at || '').slice(5, 10),
    deadlineText: (it.deadline || '').slice(0, 10) + (st.cls === 'closed' ? ' · 已截止' : ''),
    organizer: it.poster_name || '协会会员企业',
    reqs,
  }
}

// ===== 揭榜动态（聚合数 + 脱敏条目） =====
// claimStatus: loading（数字未到前统计列显示 -）/ ok / error（如实显示加载失败，不伪造数据）
const claimStatus = ref('loading')
const claimCount = ref(0)
const claimItems = ref([])
const claimed = ref(false)
const submitting = ref(false)
// 状态语义：submitted 待审核 / reviewing 审核中 / matched 已对接（未知状态一律按待审核兜底展示）
const CLAIM_STATES = {
  submitted: { label: '待审核', cls: 'a' },
  reviewing: { label: '审核中', cls: 'b' },
  matched: { label: '已对接', cls: 'c' },
}
const mapClaim = (c) => {
  const st = CLAIM_STATES[c.status] || CLAIM_STATES.submitted
  const name = c.claimer || '匿名会员'
  return {
    id: c.id,
    claimer: name,
    avatar: name.charAt(0),
    statusLabel: st.label,
    statusCls: st.cls,
    created_at: c.created_at || '',
  }
}
const applyClaims = (res) => {
  claimCount.value = Number(res?.total) || 0
  claimItems.value = Array.isArray(res?.items) ? res.items.slice(0, 3).map(mapClaim) : []
  claimed.value = !!res?.claimed
  claimStatus.value = 'ok'
}
const claimCountText = computed(() => (claimStatus.value === 'ok' ? String(claimCount.value) : '-'))
// 底部主按钮文案：截止 / 已揭榜 / 提交中 三态
const submitLabel = computed(() => {
  if (submitting.value) return '提交中…'
  if (d.value?.stCls === 'closed') return '已截止揭榜'
  if (claimed.value) return '已提交 · 待审核'
  return '提交揭榜意向'
})

const fetchData = async () => {
  loading.value = true
  err.value = false
  claimStatus.value = 'loading'
  // 详情与揭榜动态并行拉取（Promise.allSettled：一方失败不拖累另一方）
  const [detailRes, claimsRes] = await Promise.allSettled([
    request({ url: '/api/v1/challenges/' + encodeURIComponent(id) }),
    request({ url: '/api/v1/challenges/' + encodeURIComponent(id) + '/claims' }),
  ])
  if (detailRes.status === 'fulfilled') {
    const it = (detailRes.value && detailRes.value.data) || detailRes.value
    if (it && it.id) d.value = mapItem(it)
  }
  if (!d.value) {
    // 演示数据仅限开发环境；生产环境绝不回退（数字诚实铁律）
    if (!isProduction) {
      const mock = (MOCK_CHALLENGES || []).find((x) => x.id === id)
      d.value = mock ? mapItem(mock) : null
    }
    if (!d.value) err.value = true
  }
  if (claimsRes.status === 'fulfilled') {
    applyClaims(claimsRes.value)
  } else {
    claimStatus.value = 'error'
    claimCount.value = 0
    claimItems.value = []
  }
  loading.value = false
}

const toggleFav = () => {
  isFav.value = !isFav.value
  if (isFav.value) {
    favs.add(id)
  } else {
    favs.delete(id)
  }
  saveFavs()
  uni.showToast({ title: isFav.value ? '已收藏' : '已取消收藏', icon: 'none' })
}
onShareAppMessage(() => ({
  title: d.value ? '研发难题：' + d.value.t : '低空经济生态服务平台 · 研发难题广场',
  path: '/pkg-eco/pages/challenges/detail?id=' + encodeURIComponent(id),
}))
// 揭榜意向提交：截止/已揭榜守卫 → 登录门 → 二次确认 → POST
const submitClaim = () => {
  if (!d.value) return
  if (d.value.stCls === 'closed') {
    uni.showToast({ title: '该难题已截止揭榜', icon: 'none' })
    return
  }
  if (claimed.value) {
    uni.showToast({ title: '已提交过揭榜意向，请等待协会审核', icon: 'none' })
    return
  }
  if (!authStorage.getAccessToken()) {
    uni.showToast({ title: '请先登录后再揭榜', icon: 'none' })
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  uni.showModal({
    title: '提交揭榜意向',
    content: '请确认单位具备攻关能力，提交后由协会审核对接。',
    confirmText: '确认提交',
    cancelText: '再想想',
    success: (res) => {
      if (res.confirm) doClaim()
    },
  })
}
const doClaim = async () => {
  submitting.value = true
  try {
    await request({
      url: '/api/v1/challenges/' + encodeURIComponent(id) + '/claims',
      method: 'POST',
      data: {},
    })
    claimed.value = true
    uni.showToast({ title: '揭榜意向已提交，协会将尽快与您对接', icon: 'none' })
    fetchClaims()
  } catch (e) {
    // 409/404 场景后端返回中文提示（已揭榜/已截止/已下架）；网络失败等走兜底文案
    uni.showToast({ title: getErrorMessage(e) || '提交失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
const fetchClaims = async () => {
  claimStatus.value = 'loading'
  try {
    applyClaims(await request({ url: '/api/v1/challenges/' + encodeURIComponent(id) + '/claims' }))
  } catch {
    claimStatus.value = 'error'
    claimCount.value = 0
    claimItems.value = []
  }
}
const goBack = () => uni.navigateBack()

onLoad((options) => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  // decode：列表页传入的 id 经 encodeURIComponent 编码，此处须解码一次再用于请求
  id = options?.id ? decodeURIComponent(options.id) : ''
  isFav.value = favs.has(id)
  fetchData()
})
</script>

<style>
page {
  background: #f5f6f8;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #f5f6f8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== 领域色 Hero ===== */
.hero {
  position: relative;
  min-height: 180px;
  padding: 18px 16px 56px;
  color: #fff;
  /* 授权配方：160deg 重色叙事条（#0a3a6b → #074d92） */
  background: linear-gradient(160deg, #0a3a6b 0%, #074d92 100%);
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}
/* 领域色竖条：左缘 6rpx 全高，领域身份锚点（与 list 页色条同宽同色源） */
.hero-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 6rpx;
  transform-origin: center top;
}
.h-tags { display: flex; justify-content: space-between; align-items: flex-start; gap: 8px; margin-bottom: 10px; }
/* 白底领域胶囊：深字浅底对（随领域换色，与 list 页 tag 同源）；dot 为领域首字徽标 */
.field-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #fff;
  padding: 3px 12px 3px 3px;
  border-radius: 999rpx;
}
.fp-dot {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
}
.fp-txt { font-size: 12px; font-weight: 700; }
/* 状态角标：对齐组三件套（tint 底 + 深色字 + 8rpx 圆角），与 list 页同套语义色 */
.h-tag {
  font-size: 12px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 8rpx;
}
.h-tag.st-open { background: #E9F7F0; color: #0B6B41; }
.h-tag.st-urgent { background: #FDECEC; color: #B42318; }
.h-tag.st-closed { background: #EEF1F4; color: #5D6B82; }
.h-title { display: block; font-size: 20px; font-weight: 800; line-height: 1.35; }
.h-budget { display: flex; align-items: baseline; gap: 6px; margin-top: 10px; }
/* 悬赏金额：Bounty 橙 #C2410C（list 页同款语义色，原型图决策版；深蓝底对比约 2.2:1，真机看效果）
   20px 落字阶表；面议态降重为半透明白 15px/600（无金额不伪造、不喧宾） */
.h-vl { font-size: 20px; font-weight: 800; color: #C2410C; }
.h-vl.face { font-size: 15px; font-weight: 600; color: rgba(255, 255, 255, 0.85); }
/* 悬赏金额数字：hero 的心脏——落位时 ios-pop 弹簧弹出一次（80ms 起播 .25s，总 330ms ≤ 400ms） */
.h-vl { animation: popIn .25s cubic-bezier(.34, 1.8, .64, 1) 80ms backwards; }
.h-lb { font-size: 12px; color: rgba(255, 255, 255, 0.92); }

/* ===== 信息卡 ===== */
.info {
  position: relative;
  z-index: 5;
  margin: -32px 12px 0;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
  padding: 12px;
}
.stat { display: flex; padding-bottom: 10px; border-bottom: 1px solid #F2F3F5; margin-bottom: 4px; }
.si { flex: 1; text-align: center; position: relative; }
.si + .si::before { content: ''; position: absolute; left: 0; top: 6px; bottom: 6px; width: 0.5px; background: #F0F0F0; }
/* 统计数字：16px 阶梯值（原 15px）——数字是信息卡主读点，参考 list 金额 18px 的提级逻辑 */
.sv { font-size: 16px; font-weight: 800; color: #17212B; display: block; }
.sl { font-size: 12px; color: #667085; margin-top: 2px; display: block; }
.row { display: flex; align-items: flex-start; gap: 9px; padding: 8px 0; border-bottom: 0.5px solid #F5F5F5; }
.row:last-child { border-bottom: none; }
.ic {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #EAF3FB;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
  font-size: 12px;
  font-weight: 700;
}
.ic.orange { background: #FFF0E6; color: #E96012; }
.ic.green { background: #E9F7F0; color: #168A55; }
.it { flex: 1; min-width: 0; }
/* 标签对比度：原 #98A2B3 白底 2.6:1 违 AA，升 #667085（4.9:1）+ 字号 11px→12px */
.il { font-size: 12px; color: #667085; display: block; margin-bottom: 1px; }
/* 信息值：15px（例外值，与 list 卡片标题同尺寸，原 14px）——详情页信息行是主读内容，不得低于列表标题级 */
.iv { font-size: 15px; color: #17212B; font-weight: 500; line-height: 1.4; display: block; }

/* ===== 区块 ===== */
.sec {
  margin: 10px 12px 0;
  padding: 13px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
}
.sh { display: flex; align-items: center; gap: 7px; margin-bottom: 10px; }
.sd { width: 3px; height: 14px; background: #0A66C2; border-radius: 2px; }
.sht { font-size: 16px; font-weight: 700; color: #17212B; }
/* 正文：15px（例外值，原 14px）——阅读页主体，参考 list 页标题 15px 的层级，明显高于列表描述 12px */
.p { font-size: 15px; color: #667085; line-height: 1.7; display: block; }
.p.dim { color: #5D6B82; } /* 空态灰：#5D6B82 白底 5.4:1（原 #98A2B3 2.6:1 违 AA），比正文 #667085 更稳但不失弱化感 */
.ul { display: flex; flex-direction: column; }
.li { display: flex; align-items: flex-start; gap: 8px; padding: 0 0 9px; }
.li:last-child { padding-bottom: 0; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: #0A66C2; margin-top: 6px; flex: none; }
/* 攻关要求条目：13px（例外值，原 12px）——内容条目，与 list 排序选项同级，不再低于正文太多 */
.li-t { font-size: 13px; color: #667085; line-height: 1.6; flex: 1; }

/* ===== 揭榜流程 ===== */
.steps { display: flex; align-items: flex-start; }
.step { flex: 1; text-align: center; position: relative; }
.step .no {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 6px;
}
.step .txt { font-size: 13px; color: #667085; line-height: 1.4; } /* 流程文字：13px（例外值），引导可读性同正文降一级 */
.step::after {
  content: '';
  position: absolute;
  top: 13px;
  left: calc(50% + 16px);
  right: calc(-50% + 16px);
  height: 2px;
  background: #EAF3FB;
}
.step:last-child::after { display: none; }

/* ===== 揭榜动态 ===== */
.cl-sum { font-size: 13px; color: #5D6B82; line-height: 1.6; } /* 揭榜动态摘要：13px（例外值），内容句同攻关要求条目 */
.cl-num { color: #0A66C2; font-weight: 800; } /* 聚合数：品牌色加重，数字诚实铁律下只有真实数据才显示 */
.cl-list { border-top: 0.5px solid #F5F5F5; margin-top: 6px; }
.cl-it { display: flex; align-items: center; gap: 8px; padding: 8px 0; }
.cl-it + .cl-it { border-top: 0.5px solid #F5F5F5; }
.cl-av {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
}
.cl-mid { flex: 1; min-width: 0; display: flex; align-items: center; gap: 6px; }
/* 揭榜人展示名：15px（例外值，原 14px）——列表内主文本，与信息值同步，参考 list 标题级 */
.cl-name { font-size: 15px; color: #17212B; font-weight: 500; }
.cl-tag { font-size: 12px; padding: 1px 7px; border-radius: 8rpx; }
.cl-tag.ct-a { background: #EAF3FB; color: #0A66C2; } /* 待审核 */
.cl-tag.ct-b { background: #FFF0E6; color: #E96012; } /* 审核中 */
.cl-tag.ct-c { background: #E9F7F0; color: #168A55; } /* 已对接 */
.cl-date { font-size: 12px; color: #667085; flex: none; }

/* ===== 底部操作栏（参考 achievements/detail 布局） ===== */
.bb {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 50;
  background: #fff;
  border-top: 0.5px solid #F0F0F0;
  display: flex;
  align-items: center;
  padding: 10px 16px;
  gap: 10px;
  padding-bottom: calc(10px + env(safe-area-inset-bottom));
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.04);
}
.bi {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: #667085; /* 未收藏心形：原 #98A2B3 2.6:1 低于非文本控件 3:1，升 #667085（4.9:1） */
}
.bi.fv { color: #ff3b30; }
/* 心形：SVG data-URI（No-Emoji 规范——♥/♡ 字符属 emoji 区，绘制渲染一致且可随状态换色） */
.bit {
  width: 20px;
  height: 20px;
  background-image: url("data:image/svg+xml;base64,PHN2ZyB4bWxucz0naHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmcnIHZpZXdCb3g9JzAgMCAyNCAyNCcgZmlsbD0nbm9uZScgc3Ryb2tlPScjNjY3MDg1JyBzdHJva2Utd2lkdGg9JzInIHN0cm9rZS1saW5lam9pbj0ncm91bmQnPjxwYXRoIGQ9J00yMC44NCA0LjYxYTUuNSA1LjUgMCAwIDAtNy43OCAwTDEyIDUuNjdsLTEuMDYtMS4wNmE1LjUgNS41IDAgMCAwLTcuNzggNy43OGwxLjA2IDEuMDZMMTIgMjEuMjNsNy43OC03Ljc4IDEuMDYtMS4wNmE1LjUgNS41IDAgMCAwIDAtNy43OHonLz48L3N2Zz4=");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}
.bi.fv .bit {
  background-image: url("data:image/svg+xml;base64,PHN2ZyB4bWxucz0naHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmcnIHZpZXdCb3g9JzAgMCAyNCAyNCcgZmlsbD0nI2ZmM2IzMCc+PHBhdGggZD0nTTIwLjg0IDQuNjFhNS41IDUuNSAwIDAgMC03Ljc4IDBMMTIgNS42N2wtMS4wNi0xLjA2YTUuNSA1LjUgMCAwIDAtNy43OCA3Ljc4bDEuMDYgMS4wNkwxMiAyMS4yM2w3Ljc4LTcuNzggMS4wNi0xLjA2YTUuNSA1LjUgMCAwIDAgMC03Ljc4eicvPjwvc3ZnPg==");
}
.bo {
  height: 44px;
  border-radius: 8px;
  border: 1px solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  font-size: 13px;
  font-weight: 600;
  padding: 0 16px;
  margin: 0;
  line-height: 1;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
/* open-type=share 按钮：清除小程序 button 默认伪元素边框与 hover 底色，按压反馈走 hover-class（:active 对 button 不生效） */
.bo::after { border: none; }
.bo-hover { transform: scale(.95); background: #F4F8FC; }
.bp {
  flex: 1;
  height: 44px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.3);
}
/* 禁用态（已截止/已揭榜/提交中）：品牌色降透明度，不新增色值、无按压反馈 */
.bp-off, .bp-off:active { background: #0A66C2; opacity: .35; box-shadow: none; transform: none; }

/* ===== 骨架 ===== */
.skw { padding-top: 10px; }
.sk-h { height: 200px; margin: 12px; border-radius: 12px; background: #EDF0F3; }
.sk-sec { margin: 12px; padding: 16px; background: #fff; border-radius: 10px; }
.sk-l { height: 14px; background: #EDF0F3; border-radius: 4px; margin-bottom: 10px; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; margin-top: 16px; }

/* ===================== 动效规范（对齐全局动画规范） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许——仅重绘不重排）
   禁参与动画：top/left/width/height/margin（触发重排）、box-shadow/filter（低端安卓掉帧）
   时长：微反馈 150-200ms（按压按下 .08s 即时到位）/ 松手弹簧回位 .3s（ios-pop）/ 浮层 200-300ms / 页面级 ≤400ms；
        退场 = 进场 ×0.7 且必须存在
   曲线：两枚固定曲线——ios-pop cubic-bezier(.34,1.8,.64,1) 松手弹簧回弹（仅按压/弹出类 transform）+
        ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速（底部操作栏进场）；
        其余进场 ease-out / 退场 ease-in / 循环 linear；除这两枚外禁手写 cubic-bezier
   数量：入场错峰首屏可见项，整页编排 ≤400ms；循环动画任何时刻全页 ≤1 处
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 入场编排（hero 0ms → 信息卡 50ms → 四个区块 60/100/140/180ms，总 400ms ≤ 400ms）
   backwards 填充 → 延迟期保持隐藏不闪跳 */
.hero { animation: fadeUp .25s ease-out backwards; }
.info { animation: fadeUp .25s ease-out .05s backwards; }
.sec { animation: fadeUp .22s ease-out 60ms backwards; }
/* 四个 .sec 相邻兄弟，用 + 选择器级联（不依赖 nth-child 数节点，nav 组件混入也安全） */
.sec + .sec { animation-delay: 100ms; }
.sec + .sec + .sec { animation-delay: 140ms; }
.sec + .sec + .sec + .sec { animation-delay: 180ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }

/* 领域身份注入：色条自顶部 scaleY 抽出 + 领域胶囊淡入上落（60ms 错峰，与 list 页色条同拍；总 360ms ≤ 400ms） */
.hero-bar { animation: barGrow .3s ease-out 60ms backwards; }
@keyframes barGrow { from { transform: scaleY(0); } to { transform: scaleY(1); } }
.field-pill { animation: pillIn .2s ease-out 60ms backwards; }
@keyframes pillIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 攻关要求条目：从左侧淡入，前 6 条 20ms 错峰（总 300ms ≤ 400ms；超出立即入场） */
.li { animation: liIn .2s ease-out backwards; }
.li:nth-child(1) { animation-delay: 0ms; }
.li:nth-child(2) { animation-delay: 20ms; }
.li:nth-child(3) { animation-delay: 40ms; }
.li:nth-child(4) { animation-delay: 60ms; }
.li:nth-child(5) { animation-delay: 80ms; }
.li:nth-child(6) { animation-delay: 100ms; }
@keyframes liIn { from { opacity: 0; transform: translateX(-10px); } to { opacity: 1; transform: translateX(0); } }

/* 2) 揭榜流程 3 步：依次淡入上移（40ms 错峰）+ 连线自左向右 scaleX 生长 + 步骤圈 ios-pop 弹簧弹出
   （总 380ms ≤ 400ms；scaleX/scale 均 transform，零重排；弹簧曲线仅作用于 transform） */
.step { animation: stIn .2s ease-out backwards; }
.step:nth-child(1) { animation-delay: 0ms; }
.step:nth-child(2) { animation-delay: 40ms; }
.step:nth-child(3) { animation-delay: 80ms; }
@keyframes stIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
.step::after { animation: lineGrow .25s ease-out backwards; transform-origin: left center; }
.step:nth-child(2)::after { animation-delay: 40ms; }
@keyframes lineGrow { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.step .no { animation: popIn .3s cubic-bezier(.34, 1.8, .64, 1) backwards; } /* ios-pop：步骤圈弹簧弹出 */
.step:nth-child(2) .no { animation-delay: 40ms; }
.step:nth-child(3) .no { animation-delay: 80ms; }
@keyframes popIn { from { transform: scale(.6); opacity: 0; } to { transform: scale(1); opacity: 1; } }

/* 统计数字：信息卡落位后弹簧弹出（40ms 错峰，四列总 360ms ≤ 400ms；ios-pop 数字过冲回位） */
.sv { animation: popIn .2s cubic-bezier(.34, 1.8, .64, 1) 40ms backwards; }
.si:nth-child(2) .sv { animation-delay: 80ms; }
.si:nth-child(3) .sv { animation-delay: 120ms; }
.si:nth-child(4) .sv { animation-delay: 160ms; }

/* 章节标题条：自顶部下压弹出（scaleY，transform 零重排） */
.sd { transform-origin: top; animation: barDrop .25s ease-out 60ms backwards; }
@keyframes barDrop { from { transform: scaleY(0); } to { transform: scaleY(1); } }

/* 底部操作栏：自下而上滑入（fixed 定位，transform 不影响布局；ios-decel 流体减速——sheet 落地感） */
.bb { animation: bbUp .3s cubic-bezier(.32, .72, 0, 1) backwards; }
@keyframes bbUp { from { opacity: 0; transform: translateY(24px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；循环 1.2-1.6s linear） */
.sk-h, .sk-sec, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 交互反馈：按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位——与 list 同套手感） */
.stb { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop */
.stb:active { transform: scale(.94); opacity: .85; transition: transform .08s linear; }
.bi { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; } /* ios-pop */
.bi:active { transform: scale(.9); background: #EAF3FB; transition: transform .08s linear; }
.bi.fv:active { background: #FDECEC; } /* 已收藏时按压给红色系反馈 */
.bo { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; } /* ios-pop */
.bo:active { transform: scale(.95); background: #F4F8FC; transition: transform .08s linear; }
.bp { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop */
.bp:active { transform: scale(.95); opacity: .92; transition: transform .08s linear; }

/* 4) 状态过渡：收藏点亮时心形 ios-pop 弹簧弹出（scale .8→1 自然过冲回位，iOS 收藏手感）；取消收藏无反向动画 */
.bi.fv .bit { animation: heartPop .3s cubic-bezier(.34, 1.8, .64, 1); }
@keyframes heartPop { from { transform: scale(.8); } to { transform: scale(1); } }

/* 5) 视觉层级：信息卡与 Hero 重叠处补柔和投影，强化"卡片浮在 Hero 上"的层次（纯视觉，无布局影响） */
.info { box-shadow: 0 4px 12px rgba(16, 24, 40, 0.05); }

/* 6) "紧急"状态标签呼吸——全页唯一循环动画（语义反馈：紧急在呼吸）
   与骨架呼吸（skPulse）按 loading 状态互斥出现，任何时刻页面循环动画 ≤1 处 */
.h-tag.st-urgent { animation: urgPulse 1.6s linear infinite; }
@keyframes urgPulse { 0%, 100% { opacity: 1; } 50% { opacity: .72; } }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .hero,
.page.no-motion .hero-bar,
.page.no-motion .field-pill,
.page.no-motion .info,
.page.no-motion .sec,
.page.no-motion .li,
.page.no-motion .step,
.page.no-motion .bb,
.page.no-motion .sv,
.page.no-motion .h-vl,
.page.no-motion .sd,
.page.no-motion .step .no,
.page.no-motion .step::after,
.page.no-motion .bi.fv .bit { animation: none; } /* 装饰入场/状态弹出全关 */
.page.no-motion .sk-h, .page.no-motion .sk-sec, .page.no-motion .sk-l { animation: none; } /* 骨架呼吸关 */
.page.no-motion .h-tag.st-urgent { animation: none; } /* 紧急呼吸关（语义靠颜色传达，静态可见） */
.page.no-motion .stb:active,
.page.no-motion .bi:active,
.page.no-motion .bo:active,
.page.no-motion .bp:active { transform: none; } /* 按压微缩放关，保留颜色/透明度反馈 */
</style>
