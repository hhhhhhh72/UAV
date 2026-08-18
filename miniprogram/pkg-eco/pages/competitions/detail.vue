<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="赛事不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- ═══ ① 全屏 Hero（250px，有图用图，无图渐变兜底）═══ -->
        <view class="hero">
          <!-- 赛事实景图（有则显示） -->
          <image v-if="heroPoster(detail)" :src="heroPoster(detail)" mode="aspectFill" class="hero-img" lazy-load />
          <!-- 无图兜底：深色渐变 + 线性无人机轮廓 -->
          <view v-else class="hero-fallback">
            <view class="drone-svg">
              <view class="drone-prop p1" /><view class="drone-prop p2" /><view class="drone-prop p3" /><view class="drone-prop p4" />
              <view class="drone-arm a1" /><view class="drone-arm a2" />
              <view class="drone-body" />
              <view class="drone-gimbal" />
            </view>
          </view>

          <!-- 双层遮罩：渐变 + 径向高光 -->
          <view class="hero-mask" />
          <view class="hero-highlight" />

          <!-- 顶部导航（毛玻璃半透明） -->
          <view class="hero-nav" :style="{ top: statusBarHeight + 'px' }">
            <view class="nav-back" hover-class="nav-press" :hover-stay-time="100" @click="goBack">
              <text class="nav-back-icon">‹</text>
            </view>
            <view class="nav-capsule">
              <view class="capsule-dot" />
              <view class="capsule-divider" />
              <view class="capsule-arrow" />
            </view>
          </view>

          <!-- 左上状态徽章 -->
          <view class="status-badge" :class="'badge--' + normStatus(detail)">
            <view v-if="normStatus(detail) === 'ongoing'" class="badge-dot" />
            <text class="badge-text">{{ statusText(detail) }}</text>
          </view>

          <!-- Hero 底部信息区 -->
          <view class="hero-bottom">
            <view class="hero-tags">
              <view v-for="t in compTags(detail)" :key="t" class="hero-tag">{{ t }}</view>
            </view>
            <text class="hero-title">{{ detail.title || detail.name || '未知赛事' }}</text>
            <view class="hero-meta-row">
              <view class="meta-ico meta-ico--cal"><view class="cal-top" /><view class="cal-body"><view class="cal-line l1" /><view class="cal-line l2" /><view class="cal-line l3" /></view></view>
              <text class="hero-meta-text">{{ dateRange(detail) }}</text>
            </view>
            <view class="hero-meta-row">
              <view class="meta-ico meta-ico--loc"><view class="loc-pin" /></view>
              <text class="hero-meta-text">{{ detail.location || '待定' }}</text>
            </view>
          </view>
        </view>

        <!-- ═══ ② 白色内容区（向上圆角 14px 覆盖 Hero 底部）═══ -->
        <view class="content">
          <!-- 关键信息卡：左侧时间线 -->
          <view class="section">
            <view class="tl-card">
              <view class="tl-item">
                <view class="tl-dot tl-dot--primary" />
                <view class="tl-body">
                  <text class="tl-key">比赛时间</text>
                  <text class="tl-value">{{ dateRange(detail) }}</text>
                </view>
              </view>
              <view class="tl-item">
                <view class="tl-dot tl-dot--primary" />
                <view class="tl-body">
                  <text class="tl-key">比赛地点</text>
                  <text class="tl-value">{{ detail.location || '待定' }}</text>
                </view>
              </view>
              <view class="tl-item">
                <view class="tl-dot tl-dot--danger" />
                <view class="tl-body">
                  <text class="tl-key">报名截止</text>
                  <text class="tl-value tl-value--danger">{{ deadlineText(detail) }}</text>
                  <text v-if="countdownText(detail)" class="tl-countdown">{{ countdownText(detail) }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 报名条件卡 -->
          <view class="section">
            <text class="section-title">报名条件</text>
            <view class="req-card">
              <view
                v-for="(req, i) in requirements(detail)"
                :key="req.name"
                class="req-item"
                :class="{ 'req-item--last': i === requirements(detail).length - 1 }"
              >
                <view class="req-icon" :class="'req-icon--' + (i % 5)">
                  <text class="req-icon-text">{{ req.icon }}</text>
                </view>
                <view class="req-body">
                  <text class="req-name">{{ req.name }}</text>
                  <text class="req-desc">{{ req.desc }}</text>
                </view>
                <view class="req-badge" :class="req.level === '必满足' ? 'req-badge--must' : 'req-badge--advise'">{{ req.level }}</view>
              </view>
            </view>
          </view>

          <!-- 参赛项目卡 -->
          <view class="section">
            <view class="section-head">
              <text class="section-title">参赛项目</text>
              <text class="section-sub">{{ eventList(detail).length }} 个项目</text>
            </view>
            <view class="event-card">
              <view
                v-for="(ev, i) in eventList(detail)"
                :key="ev.name"
                class="event-item"
                hover-class="event-press"
                :hover-stay-time="120"
                @click="handleEventTap(ev)"
              >
                <view class="event-info">
                  <view class="event-name-row">
                    <text class="event-name">{{ ev.name }}</text>
                    <view v-if="i === 0" class="hot-badge">热门</view>
                  </view>
                  <text class="event-meta">{{ ev.type }} · {{ ev.format }}</text>
                </view>
                <view class="event-price">
                  <template v-if="ev.fee === 0 || ev.fee == null">
                    <text class="event-fee-free">免费</text>
                  </template>
                  <template v-else>
                    <text class="event-symbol">¥</text>
                    <text class="event-num">{{ ev.fee.toLocaleString() }}</text>
                    <text class="event-unit">{{ ev.type === '团体赛' ? '/队' : '/人' }}</text>
                  </template>
                </view>
              </view>
            </view>
          </view>

          <!-- 奖项设置卡 -->
          <view v-if="prizes(detail).length > 0" class="section">
            <view class="section-head">
              <text class="section-title">奖项设置</text>
              <text class="section-sub">总奖金池 ¥{{ totalPrize(detail) }}</text>
            </view>
            <view class="prize-grid">
              <view
                v-for="p in prizes(detail)"
                :key="p.level"
                class="prize-card"
                :class="'prize-card--' + p.metal"
              >
                <view class="prize-glare" />
                <view class="prize-medal">{{ p.medal }}</view>
                <text class="prize-level">{{ p.level }}</text>
                <view class="prize-amount-row">
                  <text class="prize-symbol">¥</text>
                  <text class="prize-amount">{{ p.amount.toLocaleString() }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 主办单位卡 -->
          <view class="section section--last">
            <text class="section-title">主办单位</text>
            <view class="org-card">
              <view class="org-avatar">{{ orgInitial(detail) }}</view>
              <view class="org-info">
                <text class="org-name">{{ detail.organizer || detail.sponsor || '待定' }}</text>
                <text class="org-sub">{{ detail.organizer_sub || '主办单位' }}</text>
              </view>
              <view class="org-arrow">›</view>
            </view>
          </view>

          <!-- 底部留白，防被固定底栏遮挡 -->
          <view class="bottom-space" />
        </view>
      </template>
    </StateView>

    <!-- ═══ ③ 底部固定操作栏（仅详情存在时）═══ -->
    <view v-if="detail" class="bottom-bar">
      <view class="bottom-left">
        <text class="fee-label">报名费</text>
        <view class="fee-price">
          <text class="fee-symbol">¥</text>
          <text class="fee-value">{{ compMinFee(detail) }}</text>
          <text class="fee-unit">起/人</text>
        </view>
      </view>
      <view class="bottom-actions">
        <view class="btn-outline" hover-class="btn-outline-press" :hover-stay-time="100" @click="handleConsult">
          <view class="btn-phone-ico" />
          <text class="btn-outline-text">咨询</text>
        </view>
        <view
          class="btn-primary"
          :class="isClosed(detail) ? 'btn-primary--disabled' : 'btn-primary--active'"
          :disabled="isClosed(detail)"
          hover-class="btn-primary-press"
          :hover-stay-time="100"
          @click="goRegister"
        >
          <text class="btn-primary-text">{{ isClosed(detail) ? '已截止' : '立即报名' }}</text>
        </view>
      </view>
    </view>

    <!-- ═══ ④ 自定义 Toast ═══ -->
    <view v-if="toast.show" class="custom-toast" :class="{ 'custom-toast--out': toast.hide }">
      <view class="toast-icon"><view class="toast-check" /></view>
      <text class="toast-text">{{ toast.msg }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

// 状态栏高度：微信端 CSS 变量 --status-bar-height 不生效，须 JS 读取动态设置
const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 20

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

/* Toast */
const toast = ref({ show: false, hide: false, msg: '' })
let toastTimer = null
let toastOutTimer = null

/* ===== 状态（与列表页同源） ===== */
function normStatus(item) {
  var s = item.status
  if (s === 'open' || s === 'enrolling' || s === 'registration_open') return 'enrolling'
  if (s === 'upcoming' || s === 'not_started') return 'upcoming'
  if (s === 'ongoing' || s === 'live') return 'ongoing'
  if (s === 'full' || s === 'deadline') return 'closed'
  return s || 'enrolling'
}
function statusText(item) {
  var map = { enrolling: '报名中', upcoming: '即将开始', ongoing: '进行中', closed: '已结束' }
  return map[normStatus(item)] || '报名中'
}
function isClosed(item) {
  return normStatus(item) === 'closed'
}

/* ===== Hero ===== */
function heroPoster(item) {
  var u = item.poster || item.cover || item.image || item.banner
  return u ? u : ''
}

/* ===== 数据映射 ===== */
function dateRange(item) {
  var s = fmtDate(item.start_date)
  var e = fmtDate(item.end_date)
  if (s === '待定' && e === '待定') return '日期待定'
  return e && e !== s ? s + ' - ' + e : s
}
function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}
function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags.slice(0, 3)
  var tags = []
  if (item.category) tags.push(item.category)
  if (tags.length === 0) tags = ['多旋翼', '国家级']
  return tags
}
function deadlineText(item) {
  var d = item.deadline || item.enroll_deadline
  if (!d) return '—'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}
function countdownText(item) {
  var d = item.deadline || item.enroll_deadline
  if (!d) return ''
  var days = deadlineDays(d)
  if (days == null) return ''
  if (days <= 0) return '已截止'
  return '距截止 ' + days + ' 天'
}
function deadlineDays(d) {
  var m = String(d).match(/(\d{4})[年.\-\/](\d{1,2})[月.\-\/](\d{1,2})/)
  if (!m) return null
  var target = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  var today = new Date()
  return Math.ceil((target - today) / 86400000)
}

/* ===== 报名条件 / 参赛项目 / 奖项（有数据用数据，无则兜底） ===== */
function requirements(item) {
  if (Array.isArray(item.requirements) && item.requirements.length > 0) return item.requirements
  return [
    { icon: '证', name: '持证要求', desc: '须持有CAAC/AOPA/UTC任一类无人机执照', level: '必满足' },
    { icon: '龄', name: '年龄限制', desc: '年满16周岁，未满18周岁需监护人签字同意', level: '必满足' },
    { icon: '时', name: '飞行时长', desc: '累计飞行时长不低于20小时', level: '建议满足' },
    { icon: '康', name: '健康要求', desc: '身体健康，无色盲色弱', level: '必满足' },
    { icon: '险', name: '保险要求', desc: '须自行购买比赛期间的第三方责任险', level: '建议满足' },
  ]
}
function eventList(item) {
  if (Array.isArray(item.events) && item.events.length > 0) return item.events
  return [
    { name: '多旋翼竞速赛', type: '个人赛', format: '计时排名', fee: 380 },
    { name: '固定翼编队赛', type: '团体赛', format: '3人一队', fee: 680 },
    { name: '航拍创作赛', type: '个人赛', format: '主题创作', fee: 280 },
  ]
}
function prizes(item) {
  if (Array.isArray(item.prizes) && item.prizes.length > 0) return item.prizes
  return [
    { level: '一等奖', amount: 10000, metal: 'gold', medal: '金' },
    { level: '二等奖', amount: 5000, metal: 'silver', medal: '银' },
    { level: '三等奖', amount: 2000, metal: 'bronze', medal: '铜' },
  ]
}
function totalPrize(item) {
  var arr = prizes(item)
  var sum = 0
  for (var i = 0; i < arr.length; i++) sum += arr[i].amount || 0
  return sum.toLocaleString()
}
function compMinFee(item) {
  if (item.minFee != null) return item.minFee
  if (item.fee != null) return item.fee
  if (item.price != null) return item.price
  if (item.price_fen != null) return item.price_fen / 100
  var evts = eventList(item)
  if (evts.length > 0) {
    var fees = evts.map(function (e) { return e.fee })
    return Math.min.apply(null, fees)
  }
  return 280
}
function orgInitial(item) {
  var name = item.organizer || item.sponsor || '中'
  return name.charAt(0)
}

/* ===== 加载（优先 storage，其次后端列表匹配，最后 mock） ===== */
async function loadDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    // 优先读取列表页经 storage 传入的完整数据（后端无公开单查接口）
    var cached = uni.getStorageSync('competition_detail')
    if (cached && String(cached.id) === String(id.value)) {
      detail.value = cached
      loading.value = false
      return
    }
    // 兜底：拉全量列表按 id 匹配
    var res = await request({ url: '/api/v1/competitions' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    var found = null
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { found = items[i]; break }
    }
    detail.value = found
    if (!found) detail.value = getMockDetail()
  } catch (e) {
    detail.value = getMockDetail()
  } finally {
    loading.value = false
  }
}

function getMockDetail() {
  var mockMap = {
    'comp-1': {
      id: 'comp-1',
      title: '2026 全国无人机职业技能大赛',
      status: 'enrolling',
      tags: ['多旋翼', '固定翼', '国家级'],
      start_date: '2026-09-15', end_date: '2026-09-18',
      location: '深圳 · 宝安国际会展中心',
      deadline: '2026-09-01',
      organizer: '中国航空器拥有者及驾驶员协会',
      organizer_sub: '简称中国AOPA · 国家级行业协会',
      fee: 380,
      poster: '/static/home/hero-inspection.jpg',
    },
    'comp-2': {
      id: 'comp-2',
      title: '首届西南无人机 FPV 竞速挑战赛',
      status: 'enrolling',
      tags: ['竞速FPV', '多旋翼', '省级'],
      start_date: '2026-10-01', end_date: '2026-10-03',
      location: '成都 · 天府新区无人机竞速基地',
      deadline: '2026-09-20',
      organizer: '四川省航空运动协会',
      organizer_sub: '省级行业协会 · 专业竞速赛事',
      fee: 280,
      poster: '/static/home/demand-lift.jpg',
    },
    'comp-3': {
      id: 'comp-3',
      title: '低空物流配送实战演练赛',
      status: 'ongoing',
      tags: ['物流配送', '多旋翼', '城市级'],
      start_date: '2026-08-01', end_date: '2026-08-15',
      location: '杭州 · 余杭未来科技城',
      deadline: '2026-07-20',
      organizer: '杭州市低空经济产业协会',
      organizer_sub: '市级低空经济行业组织',
      fee: 0,
      poster: '/static/home/demand-lift.jpg',
    },
    'comp-4': {
      id: 'comp-4',
      title: '第十届植保无人机飞防作业大赛',
      status: 'upcoming',
      tags: ['植保飞防', '多旋翼', '国家级'],
      start_date: '2026-11-01', end_date: '2026-11-02',
      location: '郑州 · 黄河农场',
      deadline: '2026-10-25',
      organizer: '河南省植保技术推广站',
      organizer_sub: '省级植保技术推广机构',
      fee: 580,
      poster: '',
    },
    'comp-5': {
      id: 'comp-5',
      title: '2025 长三角无人机巡检技能赛',
      status: 'closed',
      tags: ['电力巡检', '固定翼', '省级'],
      start_date: '2025-12-05', end_date: '2025-12-07',
      location: '苏州 · 工业园区',
      deadline: '2025-11-20',
      organizer: '长三角低空经济发展联盟',
      organizer_sub: '区域协同发展组织',
      fee: 220,
      poster: '/static/home/hero-inspection.jpg',
    },
  }
  return mockMap[id.value] || mockMap['comp-1']
}

/* ===== 交互 ===== */
function goBack() { uni.navigateBack({ delta: 1 }) }

function goRegister() {
  if (detail.value && isClosed(detail.value)) return
  uni.navigateTo({ url: '/pkg-eco/pages/competitions/register?id=' + encodeURIComponent(id.value) })
}

function handleConsult() { showCustomToast('已连接赛事客服') }

function handleEventTap(ev) { showCustomToast(ev.name + ' 详情') }

function showCustomToast(msg) {
  clearTimeout(toastTimer)
  clearTimeout(toastOutTimer)
  toast.value = { show: true, hide: false, msg: msg }
  toastTimer = setTimeout(function () {
    toast.value.hide = true
    toastOutTimer = setTimeout(function () {
      toast.value.show = false
    }, 200)
  }, 2000)
}

onLoad(function (options) {
  id.value = options.id || ''
  loadDetail()
})
</script>

<style scoped>
.page {
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
  padding-left: constant(safe-area-inset-left);
  padding-left: env(safe-area-inset-left);
  padding-right: constant(safe-area-inset-right);
  padding-right: env(safe-area-inset-right);
  overflow-x: hidden;
}

/* ═══ ① 全屏 Hero（250px） ═══ */
.hero {
  position: relative;
  width: 100%;
  height: 500rpx;
  overflow: hidden;
}
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.hero-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #0a5897 0%, #074D92 100%);
}
.drone-svg { position: relative; width: 160rpx; height: 120rpx; opacity: 0.9; }
.drone-prop {
  position: absolute;
  width: 44rpx; height: 44rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.65);
  border-radius: 50%;
  box-sizing: border-box;
}
.drone-prop.p1 { left: 0; top: 0; }
.drone-prop.p2 { right: 0; top: 0; }
.drone-prop.p3 { left: 0; bottom: 0; }
.drone-prop.p4 { right: 0; bottom: 0; }
.drone-arm {
  position: absolute;
  left: 50%; top: 50%;
  width: 116rpx; height: 3rpx;
  background: rgba(255, 255, 255, 0.4);
}
.drone-arm.a1 { transform: translate(-50%, -50%) rotate(-45deg); }
.drone-arm.a2 { transform: translate(-50%, -50%) rotate(45deg); }
.drone-body {
  position: absolute;
  left: 50%; top: 50%;
  width: 56rpx; height: 36rpx;
  margin: -18rpx 0 0 -28rpx;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 10rpx;
}
.drone-gimbal {
  position: absolute;
  left: 50%; top: 50%;
  width: 24rpx; height: 24rpx;
  margin: 20rpx 0 0 -12rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.65);
  border-radius: 50%;
  box-sizing: border-box;
}

/* 双层遮罩 */
.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(7, 77, 146, 0.35) 0%, rgba(7, 77, 146, 0.10) 30%, rgba(7, 77, 146, 0.65) 100%);
}
.hero-highlight {
  position: absolute;
  left: 80%;
  top: 10%;
  width: 900rpx;
  height: 400rpx;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.18) 0%, transparent 70%);
  transform: translate(-50%, -50%);
  pointer-events: none;
}

/* 顶部导航（毛玻璃） */
.hero-nav {
  position: absolute;
  top: var(--status-bar-height);
  left: 0;
  right: 0;
  padding: 8rpx 24rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 5;
}
.nav-back {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.18);
  border: 1rpx solid rgba(255, 255, 255, 0.24);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 180ms var(--ease);
}
.nav-press { background: rgba(255, 255, 255, 0.32); }
.nav-back-icon { font-size: 40rpx; color: #ffffff; font-weight: 300; line-height: 1; }
.nav-capsule {
  width: 176rpx;
  height: 60rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.24);
  border-radius: 999rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  background: rgba(255, 255, 255, 0.16);
  transition: background 180ms var(--ease);
}
.nav-capsule:active { background: rgba(255, 255, 255, 0.28); }
.capsule-dot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #ffffff; }
.capsule-divider { width: 1rpx; height: 28rpx; background: rgba(255, 255, 255, 0.4); }
.capsule-arrow {
  width: 0; height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 8rpx solid #ffffff;
}

/* 状态徽章（左上） */
.status-badge {
  position: absolute;
  top: calc(var(--status-bar-height) + 80rpx);
  left: 32rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 16rpx;
  border-radius: 6rpx;
  z-index: 4;
}
.badge-text { font-size: 20rpx; font-weight: 600; color: #ffffff; }
.badge--enrolling { background: #F97316; box-shadow: 0 4rpx 10rpx rgba(249, 115, 22, 0.32); }
.badge--upcoming { background: #0A66C2; box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28); }
.badge--ongoing { background: #168A55; }
.badge--ongoing .badge-dot {
  width: 10rpx; height: 10rpx;
  border-radius: 50%;
  background: #ffffff;
  position: relative;
}
.badge--ongoing .badge-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: #ffffff;
  animation: badgeRing 1.4s ease-out infinite;
}
.badge--closed { background: rgba(16, 24, 40, 0.62); }

/* Hero 底部信息区 */
.hero-bottom {
  position: absolute;
  left: 32rpx;
  right: 32rpx;
  bottom: 24rpx;
  z-index: 3;
}
.hero-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-bottom: 12rpx; }
.hero-tag {
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
  border: 1rpx solid rgba(255, 255, 255, 0.24);
  color: #ffffff;
  font-size: 20rpx;
  font-weight: 500;
}
.hero-title {
  display: block;
  font-size: 40rpx;
  font-weight: 760;
  color: #ffffff;
  line-height: 1.35;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.32);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero-meta-row { display: flex; align-items: center; gap: 8rpx; margin-top: 12rpx; }
.hero-meta-text { font-size: 24rpx; color: rgba(255, 255, 255, 0.85); }

/* 线性图标（Hero 副信息） */
.meta-ico { width: 28rpx; height: 28rpx; flex-shrink: 0; position: relative; }
.meta-ico--cal {
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  border-radius: 4rpx;
  box-sizing: border-box;
}
.meta-ico--cal::before,
.meta-ico--cal::after {
  content: '';
  position: absolute;
  top: 4rpx;
  width: 2rpx; height: 5rpx;
  background: rgba(255, 255, 255, 0.9);
}
.meta-ico--cal::before { left: 6rpx; }
.meta-ico--cal::after { right: 6rpx; }
.meta-ico--cal .cal-top {
  position: absolute;
  left: 3rpx; right: 3rpx; top: 8rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.9);
}
.meta-ico--cal .cal-line {
  position: absolute;
  left: 5rpx; right: 5rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.9);
  opacity: 0.6;
}
.meta-ico--cal .cal-line.l1 { top: 14rpx; }
.meta-ico--cal .cal-line.l2 { top: 19rpx; }
.meta-ico--cal .cal-line.l3 { top: 24rpx; }
.meta-ico--loc { width: 22rpx; }
.meta-ico--loc .loc-pin {
  position: absolute;
  left: 50%; top: 4rpx;
  width: 14rpx; height: 14rpx;
  margin-left: -7rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  box-sizing: border-box;
}
.meta-ico--loc .loc-pin::after {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  width: 4rpx; height: 4rpx;
  margin: -2rpx 0 0 -2rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
}

/* ═══ ② 白色内容区 ═══ */
.content {
  position: relative;
  background: #F4F6F8;
  border-radius: 28rpx 28rpx 0 0;
  margin-top: -28rpx;
  padding: 28rpx 24rpx 0;
  box-shadow: 0 -16rpx 48rpx rgba(7, 77, 146, 0.12);
  animation: contentIn 400ms var(--ease) 100ms both;
}
.section { margin-bottom: 28rpx; }
.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 760;
  color: #17212B;
  margin-bottom: 14rpx;
  padding-left: 16rpx;
  border-left: 6rpx solid #0A66C2;
}
.section-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14rpx; padding-left: 16rpx; border-left: 6rpx solid #0A66C2; }
.section-head .section-title { margin-bottom: 0; padding-left: 0; border-left: none; }
.section-sub { font-size: 20rpx; color: #98A2B3; }

/* 关键信息卡：左侧时间线 */
.tl-card {
  position: relative;
  background: #ffffff;
  border-radius: 24rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  padding: 28rpx 0 28rpx 56rpx;
}
.tl-card::before {
  content: '';
  position: absolute;
  left: 36rpx;
  top: 56rpx;
  bottom: 56rpx;
  width: 2rpx;
  background: #EEF1F4;
}
.tl-item { position: relative; padding-bottom: 24rpx; }
.tl-item:last-child { padding-bottom: 0; }
.tl-dot {
  position: absolute;
  left: -44rpx;
  top: 4rpx;
  width: 24rpx;
  height: 24rpx;
  border-radius: 50%;
  background: #ffffff;
  border: 4rpx solid #0A66C2;
  box-sizing: border-box;
  box-shadow: 0 0 0 4rpx #ffffff;
}
.tl-dot--danger { border-color: #D92D20; }
.tl-dot--ok { border-color: #168A55; }
.tl-key { display: block; font-size: 22rpx; color: #667085; }
.tl-value {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
  margin-top: 4rpx;
}
.tl-value--danger { color: #D92D20; }
.tl-countdown { display: block; font-size: 22rpx; color: #98A2B3; margin-top: 4rpx; }

/* 报名条件卡 */
.req-card {
  background: linear-gradient(135deg, #FFF0E6 0%, #fff7ee 100%);
  border-radius: 20rpx;
  padding: 8rpx 24rpx;
}
.req-item {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx dashed rgba(233, 96, 18, 0.16);
}
.req-item--last { border-bottom: none; }
.req-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.req-icon-text { font-size: 28rpx; font-weight: 700; }
.req-icon--0 { background: #EAF3FB; color: #0A66C2; }
.req-icon--1 { background: #F3F0FF; color: #7C3AED; }
.req-icon--2 { background: #E9F7F0; color: #168A55; }
.req-icon--3 { background: #FFF0E6; color: #E96012; }
.req-icon--4 { background: #FEF3F2; color: #D92D20; }
.req-body { flex: 1; min-width: 0; }
.req-name { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; }
.req-desc { display: block; font-size: 22rpx; color: #667085; margin-top: 4rpx; line-height: 1.5; }
.req-badge {
  flex-shrink: 0;
  padding: 4rpx 14rpx;
  border-radius: 6rpx;
  font-size: 20rpx;
  font-weight: 600;
}
.req-badge--must { background: #D92D20; color: #ffffff; }
.req-badge--advise { background: transparent; border: 1rpx solid #0A66C2; color: #0A66C2; }

/* 参赛项目卡 */
.event-card {
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 20rpx;
  padding: 8rpx 24rpx;
}
.event-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 22rpx 0;
  border-bottom: 1rpx solid #EEF1F4;
  transition: transform 180ms var(--ease), box-shadow 180ms var(--ease);
}
.event-item:last-child { border-bottom: none; }
.event-press { transform: scale(0.985); box-shadow: 0 6px 18px rgba(16, 24, 40, 0.08); }
.event-info { flex: 1; min-width: 0; }
.event-name-row { display: flex; align-items: center; gap: 8rpx; }
.event-name { font-size: 28rpx; font-weight: 700; color: #17212B; }
.hot-badge {
  padding: 2rpx 10rpx;
  border-radius: 4rpx;
  background: #F97316;
  color: #ffffff;
  font-size: 18rpx;
  font-weight: 600;
}
.event-meta { display: block; font-size: 22rpx; color: #667085; margin-top: 4rpx; }
.event-price { display: flex; align-items: baseline; flex-shrink: 0; }
.event-symbol { font-size: 22rpx; font-weight: 700; color: #E96012; }
.event-num { font-size: 36rpx; font-weight: 760; color: #E96012; line-height: 1; }
.event-unit { font-size: 20rpx; color: #98A2B3; margin-left: 4rpx; }
.event-fee-free { font-size: 36rpx; font-weight: 760; color: #168A55; }

/* 奖项设置卡（金银铜） */
.prize-grid {
  display: flex;
  gap: 16rpx;
}
.prize-card {
  flex: 1;
  position: relative;
  border-radius: 20rpx;
  padding: 20rpx 0 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  overflow: hidden;
}
.prize-card--gold { background: linear-gradient(160deg, #FFD56B, #F5B301); }
.prize-card--silver { background: linear-gradient(160deg, #e9edf2, #c2c8d1); }
.prize-card--bronze { background: linear-gradient(160deg, #ffd2a8, #f2945a); }
.prize-glare {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 50%;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0) 0%, rgba(255, 255, 255, 0.4) 100%);
  pointer-events: none;
}
.prize-medal {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 700;
}
.prize-card--gold .prize-medal { color: #5c3b00; }
.prize-card--silver .prize-medal { color: #4a4a4a; }
.prize-card--bronze .prize-medal { color: #6e3300; }
.prize-level { font-size: 24rpx; font-weight: 760; position: relative; }
.prize-card--gold .prize-level { color: #5c3b00; }
.prize-card--silver .prize-level { color: #4a4a4a; }
.prize-card--bronze .prize-level { color: #6e3300; }
.prize-amount-row { display: flex; align-items: baseline; position: relative; }
.prize-symbol { font-size: 18rpx; font-weight: 700; }
.prize-amount { font-size: 26rpx; font-weight: 760; }
.prize-card--gold .prize-symbol, .prize-card--gold .prize-amount { color: #5c3b00; }
.prize-card--silver .prize-symbol, .prize-card--silver .prize-amount { color: #4a4a4a; }
.prize-card--bronze .prize-symbol, .prize-card--bronze .prize-amount { color: #6e3300; }

/* 主办单位卡 */
.org-card {
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 20rpx;
  padding: 20rpx 24rpx;
}
.org-avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 16rpx;
  background: #0A66C2;
  color: #ffffff;
  font-size: 32rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.org-info { flex: 1; min-width: 0; }
.org-name { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; }
.org-sub { display: block; font-size: 20rpx; color: #667085; margin-top: 4rpx; }
.org-arrow { font-size: 32rpx; color: #98A2B3; }
.bottom-space { height: 20rpx; }

/* ═══ ③ 底部固定操作栏 ═══ */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid #EEF1F4;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  z-index: 50;
}
.bottom-left { display: flex; align-items: baseline; gap: 8rpx; flex-shrink: 0; }
.fee-label { font-size: 20rpx; color: #98A2B3; }
.fee-price { display: flex; align-items: baseline; }
.fee-symbol { font-size: 22rpx; font-weight: 700; color: #E96012; }
.fee-value {
  font-size: 44rpx;
  font-weight: 760;
  color: #E96012;
  line-height: 1;
  animation: priceIn 500ms var(--ease) both;
}
.fee-unit { font-size: 20rpx; color: #98A2B3; margin-left: 4rpx; }
.bottom-actions { display: flex; gap: 12rpx; flex: 1; justify-content: flex-end; }
.btn-outline {
  display: flex;
  align-items: center;
  gap: 8rpx;
  height: 76rpx;
  padding: 0 24rpx;
  border: 1rpx solid #0A66C2;
  border-radius: 10rpx;
  background: transparent;
  transition: transform 180ms var(--ease), background 180ms var(--ease);
}
.btn-outline-press { background: #EAF3FB; }
.btn-phone-ico {
  width: 20rpx;
  height: 20rpx;
  border: 2rpx solid #0A66C2;
  border-radius: 4rpx;
  position: relative;
}
.btn-phone-ico::before {
  content: '';
  position: absolute;
  left: 50%; top: -3rpx;
  width: 6rpx; height: 3rpx;
  background: #0A66C2;
  margin-left: -3rpx;
}
.btn-outline-text { font-size: 24rpx; font-weight: 700; color: #0A66C2; }
.btn-primary {
  height: 76rpx;
  padding: 0 32rpx;
  border-radius: 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 180ms var(--ease), background 180ms var(--ease), box-shadow 180ms var(--ease);
}
.btn-primary--active { background: #F97316; box-shadow: 0 4rpx 10rpx rgba(249, 115, 22, 0.32); }
.btn-primary--disabled { background: #EEF1F4; }
.btn-primary-press { transform: scale(0.97); }
.btn-primary-text { font-size: 28rpx; font-weight: 700; color: #ffffff; }
.btn-primary--disabled .btn-primary-text { color: #667085; }

/* ═══ ④ 自定义 Toast ═══ */
.custom-toast {
  position: fixed;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  z-index: 999;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 32rpx;
  background: rgba(16, 24, 40, 0.92);
  border-radius: 10rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.24);
  animation: toastIn 250ms var(--ease) both;
  max-width: 70vw;
}
.custom-toast--out { animation: toastOut 200ms ease both; }
.toast-icon {
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: rgba(91, 255, 176, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.toast-check {
  width: 16rpx;
  height: 9rpx;
  border-left: 3rpx solid #5BFFB0;
  border-bottom: 3rpx solid #5BFFB0;
  transform: rotate(-45deg) translate(1rpx, -1rpx);
}
.toast-text { font-size: 26rpx; color: #ffffff; font-weight: 500; line-height: 1.4; }

/* ═══ 动画 ═══ */
@keyframes contentIn {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes badgeRing {
  0% { transform: scale(1); opacity: 0.8; }
  80% { transform: scale(2.4); opacity: 0; }
  100% { transform: scale(2.4); opacity: 0; }
}
@keyframes priceIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes toastIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% - 20rpx)); }
  to { opacity: 1; transform: translate(-50%, -50%); }
}
@keyframes toastOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* ═══ 减少动态效果支持 ═══ */
@media (prefers-reduced-motion: reduce) {
  .content,
  .badge--ongoing .badge-dot::after,
  .fee-value,
  .custom-toast {
    animation: none !important;
    transition: none !important;
  }
}
</style>
