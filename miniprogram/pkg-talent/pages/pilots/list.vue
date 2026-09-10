<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="认证飞手" show-back :fixed="true" @back="goBack" />

    <!-- 右下角浮动申请按钮（避开微信胶囊，单次弹性入场） -->
    <view class="apply-fab" hover-class="apply-fab-hover" :hover-stay-time="80" @tap="applyPilot">
      <view class="fab-icon"><image class="fab-img" src="/static/mine-icons/drone.svg" mode="aspectFit" /></view>
      <text class="fab-text">{{ applyText }}</text>
    </view>

    <!-- ① 名录概览（仅当整份名录已加载时展示，避免"只统计首页"的失实口径） -->
    <view v-if="showOverview" class="overview">
      <view class="ov-cell">
        <text class="ov-num">{{ ovNums.rosters }}</text>
        <text class="ov-label">飞手</text>
      </view>
      <view class="ov-divider" />
      <view class="ov-cell">
        <text class="ov-num">{{ ovNums.certs }}</text>
        <text class="ov-label">认证证书</text>
      </view>
      <view class="ov-divider" />
      <view class="ov-cell">
        <text class="ov-num">{{ ovNums.hours }}</text>
        <text class="ov-label">累计飞行小时</text>
      </view>
    </view>

    <!-- ② 搜索框 -->
    <view class="sbar">
      <view class="b-search">
        <u-icon name="search" size="34rpx" color="#667085" />
        <input
          class="b-sinp"
          v-model="searchText"
          placeholder="搜索认证飞手姓名"
          placeholder-class="b-ph"
          confirm-type="search"
          @confirm="onSearch"
        />
        <view v-if="searchText" class="b-sclr" hover-class="b-sclr-hover" :hover-stay-time="80" @tap="clearSearch">
          <u-icon name="close" size="28rpx" color="#667085" />
        </view>
        <view class="b-sep" />
        <text class="b-sbtn" hover-class="b-sbtn-hover" :hover-stay-time="80" @tap="onSearch">搜索</text>
      </view>
    </view>

    <!-- ③ 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ total }}</text> 位飞手</text>
      <text class="ir-hint">{{ keyword ? '搜索结果' : '协会认证' }}</text>
    </view>

    <!-- ④ 骨架屏：与真实卡片 1:1（头像圆 + 姓名/副行 + 两格数据 + 标签行） -->
    <view v-if="loading && !list.length" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-head">
          <view class="sk-avatar" />
          <view class="sk-head-main">
            <view class="sk-l w50" />
            <view class="sk-l w30 sk-l--sm" />
          </view>
        </view>
        <view class="sk-grid">
          <view v-for="j in 2" :key="'skg' + j" class="sk-cell">
            <view class="sk-dot" />
            <view class="sk-l w40 sk-l--sm" />
          </view>
        </view>
        <view class="sk-tags">
          <view class="sk-tag" />
          <view class="sk-tag sk-tag--w" />
        </view>
      </view>
    </view>

    <!-- ⑤ 错误态 -->
    <view v-else-if="!loading && errorMsg" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" hover-class="stb-hover" :hover-stay-time="80" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑥ 空态 -->
    <view v-else-if="!loading && !list.length" class="st">
      <u-empty :description="keyword ? '没有匹配的认证飞手' : '暂无认证飞手'">
        <text class="sth">{{ keyword ? '换个姓名关键词试试' : '成为协会认证飞手，即可展示在此名录' }}</text>
        <view v-if="!keyword" class="stb" hover-class="stb-hover" :hover-stay-time="80" @tap="applyPilot">申请认证</view>
        <view v-else class="stb" hover-class="stb-hover" :hover-stay-time="80" @tap="clearSearch">清除搜索</view>
      </u-empty>
    </view>

    <!-- ⑦ 飞手卡片列表 -->
    <view v-else class="cl">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="goDetail(item)"
      >
        <!-- 卡片头部：头像 + 认证徽章 + 姓名/区域 -->
        <view class="card-head">
          <view class="avatar-wrap">
            <image
              v-if="item.avatar"
              :src="item.avatar"
              mode="aspectFill"
              class="avatar"
              lazy-load
            />
            <view v-else class="avatar avatar-fallback" :style="{ background: avatarBg(item.real_name) }">
              <text class="avatar-char">{{ firstChar(item.real_name) }}</text>
            </view>
            <view class="cert-badge" />
          </view>
          <view class="head-main">
            <text class="name">{{ item.real_name || '认证飞手' }}</text>
            <view class="sub-row">
              <text class="pilot-sub">{{ item.region || '地区未填写' }}</text>
            </view>
          </view>
        </view>

        <!-- 数据行（两格：只放后端真实字段，无数据来源的指标不占位） -->
        <view class="data-grid">
          <view class="data-item">
            <view class="data-icon data-icon-blue"><image class="data-img" src="/static/mine-icons/drone.svg" mode="aspectFit" /></view>
            <view class="data-body">
              <text class="data-label">飞行</text>
              <text class="data-value">{{ hoursText(item) }}</text>
            </view>
          </view>
          <view class="data-item">
            <view class="data-icon data-icon-green"><image class="data-img" src="/static/mine-icons/certification-green.svg" mode="aspectFit" /></view>
            <view class="data-body">
              <text class="data-label">证书</text>
              <text class="data-value">{{ certsText(item) }}</text>
            </view>
          </view>
        </view>

        <!-- 作业类型标签（沿用全站四色板：蓝/橙/绿/紫 + 中性） -->
        <view v-if="jobTags(item).length > 0" class="tag-row">
          <text
            v-for="(t, ti) in shownTags(item)"
            :key="ti"
            class="job-tag"
            :class="tagClass(t)"
          >{{ t }}</text>
          <text v-if="jobTags(item).length > 3" class="more-tag">+{{ jobTags(item).length - 3 }}</text>
        </view>

        <!-- 底部：查看飞手档案（整行可点击） -->
        <view class="card-footer" hover-class="footer-hover" :hover-stay-time="100" @tap.stop="goDetail(item)">
          <text class="card-hint">查看飞手档案</text>
          <u-icon name="arrow" size="22rpx" color="#0A66C2" />
        </view>
      </view>

      <!-- 列表底部：仅在确实取完整份名录时出现 -->
      <view v-if="!hasMore" class="list-footer">
        <text class="footer-line" />
        <text class="footer-text">{{ keyword ? '已显示全部匹配结果' : '已显示全部认证飞手' }}</text>
        <text class="footer-line" />
      </view>
      <view v-else class="list-more">
        <!-- 槽位常驻：loader 出现/消失不推动文字，避免底部跳动 -->
        <view class="list-more-slot">
          <u-loading v-if="loadingMore" size="28rpx" color="#0A66C2" />
        </view>
        <text class="list-more-text">{{ loadingMore ? '加载中…' : '上拉加载更多' }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { onLoad, onReachBottom } from '@dcloudio/uni-app'
import { request, getErrorMessage } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'

const PAGE_SIZE = 20

const searchText = ref('')
const keyword = ref('') // 已提交的关键词（与输入框解耦：输入中不改变列表口径文案）
const list = ref([])
const total = ref(0)
const page = ref(1)
const hasMore = ref(false)
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const statusBarHeight = ref(20)
const { noMotion, checkMotion } = useReduceMotion()
const goBack = () => uni.navigateBack()

// 右上按钮文案：随我的认证状态变化，入口永远存在
const applyText = ref('申请认证')
const refreshMineLabel = async () => {
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = res && res.data ? res.data : res
    if (mine && mine.id) {
      applyText.value = { pending: '审核中', approved: '我的档案', rejected: '重新申请' }[mine.status] || '申请认证'
    } else {
      applyText.value = '申请认证'
    }
  } catch (e) { applyText.value = '申请认证' }
}

// ── 名录概览：只统计"已加载"的条目；仅当整份名录一次取完时展示，避免失实口径 ──
const stats = computed(() => {
  const totalCerts = list.value.reduce((s, p) => s + (p.cert_ids || []).length, 0)
  const totalHours = list.value.reduce((s, p) => s + (p.flight_hours || 0), 0)
  // 不展示"平均评分"：certified_pilots.rating 全流程无写入点（无 API 产出），展示即为死数据
  return { totalCerts, totalHours }
})
const showOverview = computed(() => list.value.length > 0 && !hasMore.value && total.value === list.value.length)

// ── 概览数字滚动：名录"落位"的署名动效，只在整份名录取完时演一次 ──
// 滚动值只做呈现，终值永远等于真实统计（reduced-motion 直接落终值，中间态不承担信息）
const ovNums = ref({ rosters: 0, certs: 0, hours: 0 })
let ovTimer = null
const runOvCountUp = () => {
  const target = { rosters: list.value.length, certs: stats.value.totalCerts, hours: stats.value.totalHours }
  if (ovTimer) { clearTimeout(ovTimer); ovTimer = null }
  if (noMotion.value) { ovNums.value = target; return }
  const from = { ...ovNums.value }
  const start = Date.now()
  const duration = 520
  const tick = () => {
    if (noMotion.value) { ovNums.value = target; ovTimer = null; return }
    const p = Math.min(1, (Date.now() - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3) // 指数减速收尾，不弹跳
    ovNums.value = {
      rosters: Math.round(from.rosters + (target.rosters - from.rosters) * ease),
      certs: Math.round(from.certs + (target.certs - from.certs) * ease),
      hours: Math.round(from.hours + (target.hours - from.hours) * ease),
    }
    if (p < 1) ovTimer = setTimeout(tick, 16)
    else ovTimer = null
  }
  tick()
}
watch(showOverview, (on) => { if (on) runOvCountUp() })

// ── 头像兜底：姓名首字 + 姓名哈希在"品牌蓝同色阶"内取一档（不再自造彩色渐变）──
const AVATAR_TINTS = ['#0A66C2', '#0B5AA8', '#0C4F94', '#0D4480', '#0E3A6C']
const firstChar = (name) => String(name || '飞').charAt(0)
const avatarBg = (name) => {
  const n = String(name || '')
  if (!n) return AVATAR_TINTS[0]
  let h = 0
  for (let i = 0; i < n.length; i++) h = (h * 31 + n.charCodeAt(i)) >>> 0
  return AVATAR_TINTS[h % AVATAR_TINTS.length]
}

// ── 卡片字段映射（缺什么显示什么，不编造、不补零）──────────────
const hoursText = (item) => (item.flight_hours > 0 ? item.flight_hours + ' 小时' : '—')
const certsText = (item) => {
  const n = (item.cert_ids || []).length
  return n > 0 ? String(n) : '—'
}
const bioList = (bio) => String(bio || '').split(/[/，,、\s]+/).filter(Boolean)
const jobTags = (item) => bioList(item.bio)
const shownTags = (item) => jobTags(item).slice(0, 3)

// 作业类型标签直接用全站四色板（pub-style.css 的 demand/service/product/course），
// 术语含义由文字承担，颜色只做分组，不再自建 7 色映射表。
const TAG_TONES = [
  { key: ['电力巡检', '巡检', '测绘', '航拍', '拍摄'], cls: 'tag-blue' },
  { key: ['应急', '救援', '侦察', '吊运', '吊装', '实操'], cls: 'tag-orange' },
  { key: ['植保', '喷洒', '物流', '运输', '投送'], cls: 'tag-green' },
  { key: ['培训', '教学', '宣讲'], cls: 'tag-purple' },
]
const tagClass = (tag) => {
  for (const t of TAG_TONES) {
    if (t.key.some((k) => String(tag).includes(k))) return t.cls
  }
  return 'tag-neutral'
}

// ── 数据加载（真分页：下拉刷新 + 上拉加载，footer 只在取完后宣告结束）──
const goDetail = (item) => {
  uni.setStorageSync('pilot_detail', item)
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(item.id) })
}
const onSearch = () => {
  keyword.value = searchText.value.trim()
  fetchData(true)
}
const clearSearch = () => {
  searchText.value = ''
  keyword.value = ''
  fetchData(true)
}

const fetchData = async (reset = true) => {
  if (reset) {
    loading.value = true
    page.value = 1
  } else {
    if (loadingMore.value || !hasMore.value) return
    loadingMore.value = true
  }
  errorMsg.value = ''
  try {
    const res = await request({
      url: '/api/v1/certified-pilots',
      data: { page: page.value, page_size: PAGE_SIZE, keyword: keyword.value },
    })
    const rows = Array.isArray(res) ? res : (res.data || [])
    const t = typeof res.total === 'number' ? res.total : rows.length
    list.value = reset ? rows : list.value.concat(rows)
    total.value = reset ? t : total.value
    hasMore.value = list.value.length < total.value && rows.length > 0
  } catch (e) {
    if (reset) {
      list.value = []
      total.value = 0
      hasMore.value = false
    }
    errorMsg.value = getErrorMessage(e) || '加载失败，请检查网络后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

onReachBottom(() => fetchData(false))

// ---- 申请认证 / 我的状态 ----
const applyPilot = async () => {
  if (!requireLogin()) return
  let mine = null
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    mine = res && res.data ? res.data : res
  } catch (e) {}
  if (mine && mine.id) {
    const label = { pending: '待审核', approved: '已认证', rejected: '未通过' }[mine.status] || mine.status
    if (mine.status === 'approved') {
      uni.showModal({
        title: '我的飞手认证',
        content: `当前状态：${label}，已展示在认证飞手名录中`,
        confirmText: '查看档案',
        success: (r) => {
          if (r.confirm) {
            uni.removeStorageSync('pilot_detail')
            uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(mine.id) })
          }
        },
      })
      return
    }
    if (mine.status === 'rejected') {
      uni.showModal({
        title: '我的飞手认证',
        content: `当前状态：${label}，可修改资料后重新提交\n${mine.real_name || ''}`,
        confirmText: '重新提交',
        success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' }) },
      })
      return
    }
    uni.showModal({ title: '我的飞手认证', content: `当前状态：${label}\n${mine.real_name || ''}`, showCancel: false, confirmText: '知道了' })
    return
  }
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    if (sys && sys.statusBarHeight) statusBarHeight.value = sys.statusBarHeight
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  fetchData(true)
  refreshMineLabel()
})
</script>

<style scoped>
/* 设计口径：与 App.vue 令牌阶梯 / pub-style.css / courses.vue 同一体系
   —— 单位一律 rpx（1px = 2rpx），正文最小 24rpx，可点元素最小 88rpx，
      中性色只用 #17212B / #667085 两级（#98A2B3 对比度 2.58:1 已弃用）。 */
.page {
  min-height: 100vh;
  background: #F5F6F8;
  padding-bottom: 160rpx; /* 给右下角浮动按钮留空间 */
}

/* ═══ 右下角浮动申请按钮 ═══ */
.apply-fab {
  position: fixed;
  right: 32rpx;
  bottom: calc(48rpx + env(safe-area-inset-bottom));
  z-index: 60;
  display: flex;
  align-items: center;
  gap: 12rpx;
  min-height: 88rpx;
  padding: 12rpx 32rpx 12rpx 12rpx;
  background: #0A66C2;
  border-radius: 999rpx;
  box-shadow: 0 12rpx 36rpx rgba(10, 102, 194, 0.28);
  animation: fab-in 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.apply-fab-hover {
  transform: scale(0.94);
  opacity: 0.92;
}
.fab-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}
.fab-img {
  width: 32rpx;
  height: 32rpx;
}
.fab-text {
  font-size: 28rpx;
  font-weight: 600;
  color: #fff;
}
@keyframes fab-in {
  from { opacity: 0; transform: translateY(60rpx) scale(0.6); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

/* ═══ ① 名录概览 ═══ */
.overview {
  display: flex;
  align-items: center;
  margin: 20rpx 24rpx 4rpx;
  padding: 24rpx 16rpx;
  background: #FFFFFF;
  border: 2rpx solid #EEF1F4;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.06);
  /* 落位：面板先于数字成形，数字滚动在同一拍里跟上 */
  animation: ovIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}
@keyframes ovIn { from { opacity: 0; transform: translateY(12rpx); } to { opacity: 1; transform: translateY(0); } }
.ov-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
}
.ov-num {
  font-size: 40rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}
.ov-label {
  font-size: 24rpx;
  color: #667085;
}
.ov-divider {
  width: 2rpx;
  height: 64rpx;
  background: #EEF1F4;
}

/* ═══ ② 搜索框 ═══ */
.sbar { padding: 16rpx 24rpx 12rpx; }
.b-search {
  min-height: 88rpx;
  padding: 0 12rpx 0 22rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  background: #fff;
  box-shadow: 0 2rpx 4rpx rgba(16, 24, 40, 0.06), 0 8rpx 24rpx rgba(16, 24, 40, 0.05);
  display: flex;
  align-items: center;
  gap: 14rpx;
  box-sizing: border-box;
}
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 28rpx; color: #17212B; }
.b-ph { color: #667085; }
/* 清除按钮：视觉 24rpx，触控面 88rpx（负 margin 抵消占位） */
.b-sclr {
  flex: none;
  width: 88rpx;
  height: 88rpx;
  margin: -22rpx -12rpx -22rpx 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.b-sclr-hover { opacity: 0.55; }
.b-sep { width: 2rpx; height: 30rpx; background: #EEF1F4; flex: none; }
.b-sbtn {
  flex: none;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  padding: 0 12rpx 0 4rpx;
  color: #0A66C2;
  font-size: 28rpx;
  font-weight: 600;
}
.b-sbtn-hover { opacity: 0.6; }

/* ═══ ③ 信息行 ═══ */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 28rpx 8rpx;
  font-size: 24rpx;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; font-variant-numeric: tabular-nums; }
.ir-hint { font-size: 24rpx; color: #667085; }

/* ═══ ④ 骨架屏（1:1 复刻真实卡片）═══ */
.skl { display: flex; flex-direction: column; gap: 16rpx; padding: 8rpx 24rpx 24rpx; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  background: #fff;
  border: 2rpx solid #EEF1F4;
  border-radius: 24rpx;
}
.sk-head { display: flex; align-items: center; gap: 20rpx; }
.sk-avatar { width: 88rpx; height: 88rpx; border-radius: 50%; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-head-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 12rpx; }
.sk-grid {
  display: flex;
  gap: 16rpx;
  padding: 20rpx;
  background: #F5F6F8;
  border-radius: 16rpx;
}
.sk-cell { flex: 1; display: flex; align-items: center; gap: 12rpx; }
.sk-dot { width: 36rpx; height: 36rpx; border-radius: 50%; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-tags { display: flex; gap: 12rpx; }
.sk-tag { width: 120rpx; height: 36rpx; border-radius: 8rpx; background: #EDF0F3; animation: skPulse 1.4s linear infinite; }
.sk-tag--w { width: 160rpx; }
.sk-l { height: 24rpx; background: #EDF0F3; border-radius: 8rpx; animation: skPulse 1.4s linear infinite; }
.sk-l--sm { height: 20rpx; }
.sk-l.w30 { width: 30%; }
.sk-l.w40 { width: 40%; }
.sk-l.w50 { width: 50%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.55; } }

/* ═══ ⑤⑥ 空 / 错误 ═══ */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.sth { font-size: 24rpx; color: #667085; display: block; margin-bottom: 32rpx; }
.stb {
  min-height: 88rpx;
  display: flex;
  align-items: center;
  padding: 0 48rpx;
  border-radius: 50rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
}
.stb-hover { opacity: 0.85; transform: scale(0.985); }

/* ═══ ⑦ 飞手卡片 ═══ */
.cl { display: flex; flex-direction: column; gap: 16rpx; padding: 0 24rpx 24rpx; }
.card {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  position: relative;
  background: #fff;
  border: 2rpx solid #EEF1F4;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.06);
  transition: transform 0.35s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.15s ease;
  /* 入场对所有卡片生效（含上拉加载追加的那几屏，此前只有前 6 张会动、其余硬弹出）；
     错峰只给前 5 张，第 6 张起封顶 88ms——总延迟不随列表长度增长 */
  animation: cardIn 0.24s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}
.card:nth-child(2) { animation-delay: 22ms; }
.card:nth-child(3) { animation-delay: 44ms; }
.card:nth-child(4) { animation-delay: 66ms; }
.card:nth-child(5) { animation-delay: 88ms; }
.card:nth-child(n+6) { animation-delay: 88ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(16rpx); } to { opacity: 1; transform: translateY(0); } }
.tap-scale { transform: scale(0.97); opacity: 0.9; }

/* 7.1 卡片头部 */
.card-head { display: flex; align-items: center; gap: 20rpx; }
.avatar-wrap { position: relative; flex-shrink: 0; }
.avatar { width: 88rpx; height: 88rpx; border-radius: 50%; }
.avatar-fallback { display: flex; align-items: center; justify-content: center; }
.avatar-char { font-size: 36rpx; font-weight: 700; color: #ffffff; }
/* 头像右下角认证徽章：36rpx 绿圆 + 白描边；对勾为 in-flow ::after，靠 flex 居中，
   不用绝对定位 + 手调 left/top（旋转后仍落在圆心，不会被白圈或圆边切掉） */
.cert-badge {
  position: absolute;
  right: 2rpx;
  bottom: 2rpx;
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  background: #25915A;
  border: 3rpx solid #ffffff;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-badge::after {
  content: '';
  width: 9rpx;
  height: 15rpx;
  border: solid #ffffff;
  border-width: 0 3rpx 3rpx 0;
  /* 对勾旋转后墨迹重心偏下（+3rpx），用 margin 上提归中：墨迹中心 = 圆心 */
  margin-bottom: 6rpx;
  transform: rotate(45deg);
}
.head-main { flex: 1; min-width: 0; }
.name {
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  display: block;
}
.sub-row { display: flex; align-items: center; gap: 8rpx; margin-top: 8rpx; }
.pilot-sub {
  font-size: 24rpx;
  color: #667085;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* 7.2 数据行（两格，数值列 tabular） */
.data-grid {
  display: flex;
  background: #F5F6F8;
  border-radius: 16rpx;
  padding: 20rpx 12rpx;
}
.data-item {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.data-icon {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.data-icon-blue { background: #EAF3FB; }
.data-icon-green { background: #E9F7F0; }
.data-img {
  width: 28rpx;
  height: 28rpx;
}
.data-body { display: flex; flex-direction: column; min-width: 0; }
.data-label { font-size: 24rpx; color: #667085; }
.data-value {
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
  font-variant-numeric: tabular-nums;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}


/* 图标体系：界面控件用 u-icon 组件，领域标识用 static/mine-icons 下的 SVG（禁 emoji/Unicode 字符） */

/* 7.3 作业标签（沿用全站四色板，颜色只做分组，含义由文字承担） */
.tag-row { display: flex; flex-wrap: wrap; gap: 12rpx; }
.job-tag,
.more-tag {
  font-size: 24rpx;
  font-weight: 600;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  max-width: 280rpx;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.tag-blue { background: #EAF3FB; color: #0A66C2; }
.tag-orange { background: #FFF0E6; color: #E96012; }
.tag-green { background: #E9F7F0; color: #25915A; }
.tag-purple { background: #F0EDFF; color: #7056D6; }
.tag-neutral,
.more-tag { background: #EEF1F4; color: #667085; }

/* 7.4 卡片底部：整行可点击 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8rpx;
  min-height: 88rpx;
  margin-top: -4rpx;
  border-top: 2rpx solid #F0F1F3;
}
.footer-hover { opacity: 0.7; }
.card-hint { font-size: 28rpx; color: #0A66C2; font-weight: 600; }

/* ═══ 列表底部 / 加载更多 ═══ */
.list-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20rpx;
  padding: 40rpx 0 16rpx;
}
.footer-line { width: 96rpx; height: 2rpx; background: #EBEDF0; }
.footer-text { font-size: 24rpx; color: #667085; }
.list-more { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 32rpx 0 16rpx; }
.list-more-slot { width: 28rpx; height: 28rpx; display: flex; align-items: center; justify-content: center; }
.list-more-text { font-size: 24rpx; color: #667085; }

/* ═══ 减弱动效（无障碍）：系统偏好 + 页内开关 双保险 ═══ */
.page.no-motion .card,
.page.no-motion .overview,
.page.no-motion .apply-fab { animation: none; }
.page.no-motion .sk-avatar,
.page.no-motion .sk-dot,
.page.no-motion .sk-l,
.page.no-motion .sk-tag { animation: none; }
.page.no-motion .card { transition: opacity 0.15s ease; }
.page.no-motion .tap-scale { transform: none; }
@media (prefers-reduced-motion: reduce) {
  .card,
  .card:nth-child(n+1),
  .overview,
  .apply-fab { animation: none !important; transition: none !important; }
  .sk-avatar,
  .sk-dot,
  .sk-l,
  .sk-tag { animation: none !important; }
  .tap-scale { transform: none !important; }
}
</style>
