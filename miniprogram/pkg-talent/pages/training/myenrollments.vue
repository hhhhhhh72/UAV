<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="我的报名" show-back :fixed="true" @back="goBack" />

    <!-- Banner：点击逛逛新课 -->
    <view class="banner" @tap="goCourses">
      <view class="banner-icon">报</view>
      <view class="banner-info">
        <text class="banner-title">我的报名</text>
        <text class="banner-sub">培训课程报名记录 · 报名进度一目了然</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 -->
      <view class="ir">
        <text>共 <text class="irn">{{ enrollments.length }}</text> 项报名</text>
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
      <view v-else-if="errorMsg && !enrollments.length" class="st">
        <u-empty :description="errorMsg">
          <view class="stb" @tap="fetchList">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="!enrollments.length" class="st">
        <u-empty description="还没有报名记录">
          <text class="sth">完成培训课程报名后，报名记录将展示在这里</text>
          <view class="stb" @tap="goCourses">去逛逛培训课程</view>
        </u-empty>
      </view>

      <!-- 列表：状态徽章 + 标题 + 元信息 -->
      <view v-else class="cl">
        <view v-for="e in enrollments" :key="e.id" class="card">
          <view class="c-badges">
            <text class="c-st" :class="statusCls(e.status)">{{ statusLabel(e.status) }}</text>
          </view>
          <text class="ct">{{ e.course_title || '培训课程' }}</text>
          <view class="c-meta">
            <text>报名时间 {{ dateText(e.created_at) }}</text>
          </view>
          <view class="c-meta">
            <text>报名人 {{ e.name || '—' }}</text>
            <text v-if="e.phone" class="c-dot">·</text>
            <text v-if="e.phone">{{ e.phone }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShow, onPullDownRefresh, onPageScroll } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { safeBack, requireLogin } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'

// 与后端 validEnrollmentStatus / 管理后台状态语义对齐（用户视角文案）
const STATUS_MAP = {
  pending: '待审核',
  approved: '已通过',
  paid: '已缴费',
  enrolled: '已报名',
  rejected: '已驳回',
}
/* 状态 → 徽章色（对齐挑战广场语义：待处理=蓝、已通过=绿、已驳回=灰） */
const STATUS_CLS = {
  pending: 'st-pending',
  paid: 'st-pending',
  approved: 'st-open',
  enrolled: 'st-open',
  rejected: 'st-closed',
}

const loading = ref(false)
const errorMsg = ref('')
const enrollments = ref([])
const statusBarHeight = ref(20)
const showBt = ref(false)
const { noMotion, checkMotion } = useReduceMotion()

function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function statusCls(s) { return STATUS_CLS[s] || 'st-closed' }
function dateText(iso) { return iso ? String(iso).slice(0, 10) : '—' }

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/enrollments/mine' })
    const list = Array.isArray(res) ? res : (res && res.data) || []
    enrollments.value = Array.isArray(list) ? list : []
  } catch (e) {
    // 401 未登录/登录过期：由 request 自动跳登录，文案明确提示登录而非网络异常
    if (e && e.statusCode === 401) {
      errorMsg.value = '登录已过期，请重新登录'
      return
    }
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goCourses() {
  uni.navigateTo({ url: '/pkg-talent/pages/training/courses' })
}

function goBack() {
  safeBack()
}

function scrollToTop() { uni.pageScrollTo({ scrollTop: 0, duration: 300 }) }

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
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

/* ===== Banner（对齐研发难题广场：蓝渐变 + 圆角图标 + 标题/副标题） ===== */
.banner {
  margin: 12px 14px;
  padding: 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #0A66C2 0%, #074D92 100%);
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
  position: relative;
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}
.banner::after {
  content: '';
  position: absolute;
  top: -30%;
  right: -20%;
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.08) 0%, transparent 70%);
}
.banner-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}
.banner-info { flex: 1; min-width: 0; position: relative; z-index: 1; }
.banner-title { font-size: 14px; font-weight: 600; margin-bottom: 4px; display: block; line-height: 1.3; color: #fff; }
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: block; }

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

/* Banner 内部微编排：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms，总 340ms ≤ 400ms */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; }
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }

/* Banner 单次扫光（非循环装饰：100ms 起播 280ms 线性，380ms 内收完） */
.banner::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  background: linear-gradient(100deg, transparent 0%, rgba(255, 255, 255, 0.22) 50%, transparent 100%);
  transform: translateX(-150%) skewX(-20deg);
  animation: shineOnce .28s linear 100ms backwards;
  pointer-events: none;
}
@keyframes shineOnce {
  from { transform: translateX(-150%) skewX(-20deg); }
  to { transform: translateX(320%) skewX(-20deg); }
}

/* 2) 交互反馈：Banner 可点击（按压反馈） */
.banner { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.banner:active { transform: scale(.985); opacity: .95; transition: transform .08s linear; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }

/* 骨架呼吸（加载中环境光；循环动画 1.4s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */
.page.no-motion .card,
.page.no-motion .banner,
.page.no-motion .ir { animation: none; }
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner::before,
.page.no-motion .banner::after { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .stb:active,
.page.no-motion .bt:active,
.page.no-motion .banner:active { transform: none; }
</style>
