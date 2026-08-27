<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="我的报名" show-back :fixed="true" @back="goBack" />

    <!-- 白色板块：类型切换 + 信息行 + 列表 -->
    <view class="section">
      <!-- 分类 tab：培训 / 赛事 / 活动 / 研学（胶囊分段控件，对齐挑战广场筛选 pill） -->
      <view class="tab-shell">
        <view
          v-for="t in TABS"
          :key="t.key"
          class="tab"
          :class="{ on: tab === t.key }"
          @tap="switchTab(t.key)"
        >{{ t.label }}</view>
      </view>

      <!-- 信息行：共 N 项 -->
      <view class="ir">
        <text>{{ activeLabel }}共 <text class="irn">{{ cardItems.length }}</text> 项</text>
      </view>

      <!-- 骨架 -->
      <view v-if="loading" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w80"></view>
            <view class="sk-l w60"></view>
          </view>
        </view>
      </view>

      <!-- 错误 -->
      <view v-else-if="errorMsg && !cardItems.length" class="st">
        <u-empty :description="errorMsg">
          <view class="stb" @tap="fetchList">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="!cardItems.length" class="st">
        <u-empty :description="emptyTitle">
          <text class="sth">{{ emptyHint }}</text>
          <view class="stb" @tap="goRelevant">{{ emptyAction }}</view>
        </u-empty>
      </view>

      <!-- 列表：状态徽章 + 标题 + 元信息（可点跳对应详情） -->
      <view v-else class="cl">
        <view
          v-for="(c, i) in cardItems"
          :key="c.id"
          class="card"
          :style="{ animationDelay: Math.min(i * 40, 240) + 'ms' }"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="openItem(c)"
        >
          <view class="c-badges">
            <text class="c-tag" :class="c.tagCls">{{ c.tag }}</text>
            <text class="c-st" :class="statusCls(c.status)">{{ statusLabel(c.status) }}</text>
          </view>
          <text class="ct">{{ c.title }}</text>
          <view class="c-meta">
            <text v-if="c.meta1">{{ c.meta1 }}</text>
          </view>
          <view class="c-meta">
            <text v-if="c.meta2">{{ c.meta2 }}</text>
          </view>
          <view class="c-arrow"></view>
        </view>
      </view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShow, onPullDownRefresh, onPageScroll } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { safeBack, requireLogin } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'

// ===== 分类 tab =====
const TABS = [
  { key: 'training', label: '培训' },
  { key: 'competition', label: '赛事' },
  { key: 'activity', label: '活动' },
  { key: 'study', label: '研学' },
]
const tab = ref('training')

const enrollments = ref([]) // 培训报名
const competitions = ref([]) // 赛事报名
const activities = ref([]) // 协会活动报名
const studies = ref([]) // 研学报名

const activeList = computed(() => {
  if (tab.value === 'competition') return competitions.value
  if (tab.value === 'activity') return activities.value
  if (tab.value === 'study') return studies.value
  return enrollments.value
})
const activeLabel = computed(() => (TABS.find((t) => t.key === tab.value) || TABS[0]).label)
const emptyTitle = computed(() => ({
  training: '还没有培训报名',
  competition: '还没有赛事报名',
  activity: '还没有活动报名',
  study: '还没有研学报名',
}[tab.value] || '还没有报名记录'))
const emptyHint = computed(() => ({
  training: '完成培训课程报名后，记录将展示在这里',
  competition: '报名赛事后，记录将展示在这里',
  activity: '报名协会活动后，记录将展示在这里',
  study: '报名低空研学后，记录将展示在这里',
}[tab.value] || ''))
const emptyAction = computed(() => ({
  training: '去逛逛培训课程',
  competition: '去看看赛事',
  activity: '去看看活动',
  study: '去看看研学',
}[tab.value] || '去逛逛'))
function goRelevant() {
  if (tab.value === 'competition') { uni.navigateTo({ url: '/pkg-eco/pages/competitions/list' }); return }
  if (tab.value === 'activity') { uni.navigateTo({ url: '/pkg-eco/pages/activities/list' }); return }
  if (tab.value === 'study') { uni.navigateTo({ url: '/pkg-talent/pages/study/index' }); return }
  uni.navigateTo({ url: '/pkg-talent/pages/training/courses' })
}
function switchTab(k) { tab.value = k }

// 卡片点击：跳对应详情（培训/赛事/活动/研学均按 id 直达；缺 id 时兜底对应列表）
function openItem(c) {
  const src = activeList.value || []
  const it = src.find((x) => x.id === c.id)
  if (!it) return
  if (tab.value === 'competition' && it.competition_id) {
    uni.navigateTo({ url: '/pkg-eco/pages/competitions/detail?id=' + encodeURIComponent(it.competition_id) })
    return
  }
  if (tab.value === 'activity' && it.event_id) {
    uni.navigateTo({ url: '/pkg-eco/pages/activities/detail?id=' + encodeURIComponent(it.event_id) })
    return
  }
  if (tab.value === 'study' && it.tour_id) {
    uni.navigateTo({ url: '/pkg-talent/pages/study/detail?id=' + encodeURIComponent(it.tour_id) })
    return
  }
  if (tab.value === 'training' && it.course_id) {
    uni.navigateTo({ url: '/pkg-talent/pages/training/enroll?id=' + encodeURIComponent(it.course_id) })
    return
  }
  uni.navigateTo({ url: '/pkg-talent/pages/training/courses' })
}

// 统一卡片视图（四类报名 → 状态/标题/元信息）
const cardItems = computed(() => {
  const src = activeList.value || []
  return src.map((it) => {
    if (tab.value === 'competition') {
      return {
        id: it.id,
        tag: '赛事报名',
        tagCls: 'tag-comp',
        title: it.title || '赛事报名',
        status: it.status,
        meta1: '报名时间 ' + dateText(it.created_at),
        meta2: (it.team_name ? it.team_name + ' · ' : '') + (it.name || '') + (it.member_count ? ' · ' + it.member_count + ' 人' : ''),
      }
    }
    if (tab.value === 'activity') {
      return {
        id: it.id,
        tag: '活动报名',
        tagCls: 'tag-act',
        title: it.title || '协会活动',
        status: it.status,
        meta1: (it.start_time ? dateText(it.start_time) + ' · ' : '') + (it.location || '地址待定'),
        meta2: '报名时间 ' + dateText(it.created_at),
      }
    }
    if (tab.value === 'study') {
      return {
        id: it.id,
        tag: '研学报名',
        tagCls: 'tag-study',
        title: it.tour_title || '低空研学',
        status: it.status,
        meta1: (it.start_date ? '出发 ' + dateText(it.start_date) : ''),
        meta2: '成 ' + (it.adult_count || 1) + ' · 儿 ' + (it.child_count || 0) + ' · 报名 ' + dateText(it.created_at),
      }
    }
    return {
      id: it.id,
      tag: '培训报名',
      tagCls: 'tag-train',
      title: it.course_title || '培训课程',
      status: it.status,
      meta1: '报名时间 ' + dateText(it.created_at),
      meta2: '报名人 ' + (it.name || '—') + (it.phone ? ' · ' + it.phone : ''),
    }
  })
})

// 与后端 validEnrollmentStatus / 管理后台状态语义对齐（用户视角文案）
const STATUS_MAP = {
  pending: '待审核',
  approved: '已通过',
  paid: '已缴费',
  enrolled: '已报名',
  rejected: '已驳回',
  completed: '已完成',
}
/* 状态 → 徽章色（对齐挑战广场语义：待处理=蓝、已通过=绿、已驳回=灰） */
const STATUS_CLS = {
  pending: 'st-pending',
  paid: 'st-pending',
  approved: 'st-open',
  enrolled: 'st-open',
  completed: 'st-open',
  rejected: 'st-closed',
}

const loading = ref(false)
const errorMsg = ref('')
const statusBarHeight = ref(20)
const showBt = ref(false)
const { noMotion, checkMotion } = useReduceMotion()

function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function statusCls(s) { return STATUS_CLS[s] || 'st-closed' }
function dateText(iso) { return iso ? String(iso).slice(0, 10) : '—' }

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  const targets = [
    { url: '/api/v1/enrollments/mine', key: 'enrollments' },
    { url: '/api/v1/competitions/registrations/mine', key: 'competitions' },
    { url: '/api/v1/events/registrations/mine', key: 'activities' },
    { url: '/api/v1/study-tours/enrollments/mine', key: 'studies' },
  ]
  try {
    // 三类并行拉取：任一类失败不阻塞其他（空数组如实展示）
    const results = await Promise.allSettled(targets.map((t) => request({ url: t.url })))
    results.forEach((r, i) => {
      const list = r.status === 'fulfilled' ? (Array.isArray(r.value) ? r.value : (r.value && r.value.data) || []) : []
      if (targets[i].key === 'enrollments') enrollments.value = list
      else if (targets[i].key === 'competitions') competitions.value = list
      else if (targets[i].key === 'activities') activities.value = list
      else studies.value = list
    })
  } catch (e) {
    // 竞态兜底：全部失败时给出提示
    if (e && e.statusCode === 401) {
      errorMsg.value = '登录已过期，请重新登录'
    } else {
      errorMsg.value = '网络异常，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

function goBack() {
  safeBack()
}

function scrollToTop() { uni.pageScrollTo({ scrollTop: 0, duration: 300 }) }

onLoad((options) => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  // 深链定位：活动/赛事报名成功页跳转时可指定分类（tab=activity|competition|training）
  if (options && options.tab && TABS.some((t) => t.key === options.tab)) {
    tab.value = options.tab
  }
})

// onShow 而非 onLoad：报名提交返回后立即看到最新记录
// 登录守卫前置：未登录直接引导登录，避免 401 误报"网络异常"
onShow(() => {
  if (!requireLogin()) return
  fetchList()
})
onPullDownRefresh(() => {
  fetchList().finally(() => uni.stopPullDownRefresh())
})
onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})
</script>

<style>
page {
  background: #fff;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 40px;
}

/* ===== 白色板块 ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }

/* ===== 分类 tab（胶囊分段控件，对齐挑战广场筛选 pill） ===== */
.tab-shell {
  display: flex;
  gap: 10rpx;
  margin: 8rpx 28rpx 20rpx;
  padding: 8rpx;
  background: #F2F4F7;
  border-radius: 18rpx;
}
.tab {
  flex: 1;
  text-align: center;
  font-size: 26rpx;
  color: #667085;
  padding: 14rpx 0;
  border-radius: 12rpx;
  transition: background .2s ease, color .2s ease, box-shadow .2s ease;
}
.tab.on {
  color: #fff;
  font-weight: 700;
  background: #0A66C2;
  box-shadow: 0 4rpx 12rpx rgba(10, 102, 194, 0.28);
}

/* ===== 列表卡片（白上白：灰描边 + 极淡灰投影浮起；无左缘色条） ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
  display: flex;
  flex-direction: column;
  gap: 7px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}
.c-badges { display: flex; gap: 6px; }
.c-arrow {
  position: absolute;
  right: 20rpx;
  top: 50%;
  width: 32rpx;
  height: 32rpx;
  transform: translateY(-50%);
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23C6CFDA' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M9 6l6 6-6 6'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
  pointer-events: none;
}
.card {
  animation: fadeUp .28s cubic-bezier(.16, 1, .3, 1) backwards;
}
.tap-scale { transform: scale(.985); opacity: .92; }
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.page.no-motion .card { animation: none; }
.page.no-motion .tab { transition: none; }
.page.no-motion .tap-scale { transform: none !important; opacity: 1; }
.c-st {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.c-st.st-open { color: #0B6B41; background: #E9F7F0; }
.c-st.st-pending { color: #0A66C2; background: #EAF3FB; }
.c-st.st-closed { color: #5D6B82; background: #EEF1F4; }
.c-tag {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.tag-train { color: #0A66C2; background: #EAF3FB; }
.tag-comp { color: #7A3E9D; background: #F4EBF9; }
.tag-act { color: #B54708; background: #FDEEE4; }
.tag-study { color: #1F7A48; background: #E9F7F0; }
.ct {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #667085;
  flex-wrap: wrap;
}
.c-dot { color: #DDE1E6; }

/* ===== 骨架 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; text-align: center; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 回到顶部 ===== */
.bt {
  position: fixed;
  bottom: 90px;
  right: 16px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 35;
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(0.16, 1, 0.3, 1);
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; }

/* ===================== 动效规范（对齐研发难题广场） =====================
   白名单：仅 transform / opacity（小尺寸颜色过渡允许）
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1) + ios-decel cubic-bezier(.32,.72,0,1)
   数量：列表入场仅错峰首屏 6 项，其余静置
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 列表入场：前 6 项每 20ms 依次淡入上移（backwards 填充 → 延迟期不闪跳） */
.card { animation: none; }
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 2) 交互反馈：卡片按压（快进慢出） */
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }

/* 骨架呼吸（加载中环境光；循环动画 1.4s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .stb:active,
.page.no-motion .bt:active { transform: none; }
</style>
