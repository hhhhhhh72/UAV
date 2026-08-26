<template>
  <view class="tsd-page" :class="{ 'no-motion': noMotion }">
    <u-nav-bar title="场地详情" show-back @back="goBack" />

    <!-- 骨架（对齐组：轮播块 + 分区灰条，呼吸动画） -->
    <view v-if="loading">
      <view class="sk-carousel"></view>
      <view class="sk-section">
        <view class="sk-l w50"></view>
        <view class="sk-l w80"></view>
        <view class="sk-l w90"></view>
      </view>
      <view class="sk-section">
        <view class="sk-l w40"></view>
        <view class="sk-l w60"></view>
        <view class="sk-l w70"></view>
      </view>
    </view>

    <!-- 错误：渲染真实原因（评审 P0：不再硬编码；缺 id 也进错误态） -->
    <view v-else-if="errorMsg" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="onRetry">{{ retryLabel }}</view>
      </u-empty>
    </view>

    <template v-else-if="site">
      <!-- 轮播：有真实主图（image_url）显示图片；无图保留微构图占位（天空/地面/地平线/跑道/块体） -->
      <swiper
        class="carousel"
        indicator-dots
        circular
        indicator-color="rgba(23, 33, 43, 0.45)"
        indicator-active-color="#0A66C2"
        aria-label="场地图片"
      >
        <swiper-item v-for="(s, i) in sceneSlots" :key="i">
          <image v-if="s === '__IMG__'" class="scene-img" :src="site.image_url" mode="aspectFill" />
          <view v-else class="scene" :class="'scene--' + (i + 1)">
            <view class="scene-sky"></view>
            <view class="scene-ground"></view>
            <view class="scene-horizon"></view>
            <view v-if="i <= 1" class="scene-runway" :class="{ narrow: i === 1 }"></view>
            <view v-else class="scene-blocks">
              <view v-for="j in (i === 2 ? 3 : 2)" :key="j" class="scene-block" :class="{ tall: j === 2 && i === 2 }"></view>
            </view>
            <text class="scene-label">{{ s }}</text>
          </view>
        </swiper-item>
      </swiper>

      <!-- 区 1：摘要 -->
      <view class="detail-section">
        <text class="detail-name">{{ site.name }}</text>
        <view class="detail-tags">
          <text class="v-tag" :class="'vt--' + typeKey(site.site_type)">{{ typeLabel(site.site_type) }}</text>
          <text class="v-tag" :class="'st--' + site.status">{{ statusLabel(site.status) }}</text>
        </view>
        <view class="detail-addr">
          <u-icon name="location" size="24rpx" color="#667085" />
          <text class="addr-text">{{ site.location || '位置待定' }}</text>
        </view>
      </view>

      <!-- 区 2：核心设备（设施代码 → 语义卡；空清单显示待补充，不编造参数） -->
      <view class="detail-section">
        <text class="section-title">核心设备</text>
        <view v-if="equipRows(site.facilities).length > 0" class="equip-tags">
          <text
            v-for="(e, i) in equipRows(site.facilities)"
            :key="i"
            class="equip-tag"
            :class="{ alt: isSense(e.code) }"
          >
            <text class="et-name">{{ e.label }}</text>
            <text v-if="e.role" class="et-role">{{ e.role }}</text>
          </text>
        </view>
        <text v-else class="equip-empty">设备清单待补充，以场地方实际为准</text>
      </view>

      <!-- 区 3：场地参数（有真值的行才渲染；整卡全空不渲染——空承诺比删卡更伤信任） -->
      <view v-if="realParams().length > 0" class="detail-section">
        <text class="section-title">场地参数</text>
        <view v-for="(p, i) in realParams()" :key="i" class="param-row">
          <text class="param-label">{{ p[0] }}</text>
          <text class="param-value">{{ p[1] }}</text>
        </view>
      </view>

      <!-- 区 4：收费说明 -->
      <view class="detail-section">
        <text class="section-title">收费说明</text>
        <view class="fee-item">
          <text class="fee-label">参考单价</text>
          <text class="fee-value" :class="{ face: isFace(site.price_fen) }">{{ formatPrice(site.price_fen) }}</text>
        </view>
        <view class="fee-item">
          <text class="fee-label">计费方式</text>
          <text class="fee-value sm">以场地方实际计费（按次 / 按时段）为准</text>
        </view>
        <view class="fee-item">
          <text class="fee-label">支付方式</text>
          <text class="fee-value sm">线下向场地方支付，平台不参与资金流转</text>
        </view>
      </view>

      <!-- 区 5：预约安排（近 7 天可约条——点选日期带该日期直达预约表单；规则 / 确认流程；占用以场地方确认为准） -->
      <view class="detail-section">
        <text class="section-title">预约安排</text>
        <view class="day-block">
          <view class="day-title-row">
            <text class="day-title">意向日期（近 7 天）</text>
            <text class="day-more" @tap="goBooking()">去预约 ›</text>
          </view>
          <view class="day-strip">
            <view v-for="(d, i) in nextDays()" :key="i" class="day-cell" :class="{ today: d.today }" hover-class="day-press" @tap="goBooking(d.dateKey)">
              <text class="day-text">{{ d.label }}</text>
            </view>
          </view>
          <text class="day-hint">点选日期带该日期去预约；当天是否可约以场地方确认为准</text>
        </view>
        <view class="param-row">
          <text class="param-label">预约规则</text>
          <text class="param-value">{{ site.booking_rule || '以场地方实际安排为准' }}</text>
        </view>
        <view class="param-row">
          <text class="param-label">确认方式</text>
          <text class="param-value">预计 24 小时内（工作日）场地方电话与您联系确认</text>
        </view>
      </view>
    </template>

    <!-- 底部栏：两行估价 + 胶囊 CTA（safe-area 补齐，评审 P1） -->
    <view v-if="site" class="detail-bottom-bar">
      <view class="estimate">
        <text class="est-lb">参考价</text>
        <text class="est-price" :class="{ face: isFace(site.price_fen) }">{{ formatPrice(site.price_fen) }}</text>
      </view>
      <view
        class="btn-book"
        :class="{ disabled: !canBook }"
        :hover-class="canBook ? 'btn-book-hover' : 'none'"
        @tap="goBooking"
      >{{ bookLabel }}</view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShow, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { safeNavigateTo, safeBack } from '@/utils/nav'
import { useReduceMotion } from '@/utils/motion'

const SITE_TYPE_MAP = {
  flying_field: '飞行场地',
  lab: '实验室',
  anechoic_chamber: '暗室',
  wind_tunnel: '风洞',
}
const TYPE_KEYS = { flying_field: 'fly', lab: 'lab', anechoic_chamber: 'chamber', wind_tunnel: 'tunnel' }
const STATUS_MAP = { available: '可预约', maintenance: '维护中', reserved: '已约满' }
const FACILITY_MAP = { '5G': '5G', RTK: 'RTK', radar: '雷达', spectrum_analyzer: '频谱分析' }
// 设施语义分类：定位/感知类（RTK/雷达）用青雾标签，其余（5G/频谱=通信类）蓝雾——按语义着色，不机械交替
const SENSE_FACILITIES = { RTK: true, radar: true }
// 设备职能：通用分类知识（通信/定位感知），非场地承诺值
const EQUIP_ROLES = {
  '5G': '通信链路 · 图传与数据回传',
  RTK: '差分定位 · 高精度基准',
  radar: '雷达监测 · 空域目标感知',
  spectrum_analyzer: '频谱监测 · 电磁环境分析',
}
// 场地参数行定义：[中文标签, 后端字段名]；字段无值时不渲染，整卡全空折叠为一行（与核心设备空态同构）。
// 当前后端 TestSite 未提供参数字段 → 恒为折叠态；字段就绪后按实际字段名映射，行自动长出。
const SITE_PARAM_FIELDS = [
  ['空域范围', 'airspace_range'],
  ['最大起飞重量', 'max_takeoff_weight'],
  ['跑道长度', 'runway_length'],
  ['最大飞行高度', 'max_flight_height'],
  ['适用机型', 'compatible_models'],
]
// 轮播数据：有真实主图（image_url）→ 单张图（标记 __IMG__）；无图 → 4 个微构图占位
const sceneSlots = computed(() =>
  site.value && site.value.image_url ? ['__IMG__'] : ['飞行区全景', '跑道实景', '设备区', '配套设施']
)

const loading = ref(false)
const errorMsg = ref('')
const site = ref(null)
const { noMotion, checkMotion } = useReduceMotion()

let siteId = ''
let fetchSeq = 0

const bookLabel = computed(() => {
  if (!site.value) return ''
  if (site.value.status === 'maintenance') return '维护中暂不可约'
  if (site.value.status === 'reserved') return '该时段已约满，请选择其他场地'
  return '立即预约'
})

// 是否可预约：驱动 CTA 禁用态与按压反馈（disabled 时 hover 置 none，禁用按钮不再缩放）
const canBook = computed(() => !!site.value && site.value.status === 'available')

// 错误恢复：已下架/缺参 → 返回列表（重试是死路）；网络错误 → 重新加载
const retryLabel = computed(() => {
  if (errorMsg.value === '场地不存在或已下架' || errorMsg.value === '场地参数缺失，请返回后重试') return '返回列表'
  return '重新加载'
})

function typeLabel(t) { return SITE_TYPE_MAP[t] || t || '测试场地' }
function typeKey(t) { return TYPE_KEYS[t] || 'fly' }
function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function equipRows(list) {
  return (list || []).map((f) => ({ code: f, label: FACILITY_MAP[f] || f, role: EQUIP_ROLES[f] || '' }))
}
// 有真值的参数行：值存在且非空才渲染（当前后端无参数字段 → 空数组 → 折叠态一行）
function realParams() {
  const s = site.value || {}
  return SITE_PARAM_FIELDS.filter(([, key]) => s[key] != null && s[key] !== '').map(([label, key]) => [label, String(s[key])])
}
// 近 7 天可预约条：与 booking 页 picker 的 minDate=today 一致（今起 7 天均可提交，占用以场地方确认为准；点选日期带 dateKey 直达预约表单）
function nextDays() {
  const out = []
  const now = new Date()
  const pad = (n) => (n < 10 ? '0' + n : '' + n)
  for (let i = 0; i < 7; i++) {
    const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() + i)
    out.push({
      label: i === 0 ? '今天' : i === 1 ? '明天' : d.getMonth() + 1 + '/' + d.getDate(),
      today: i === 0,
      dateKey: d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()),
    })
  }
  return out
}
function isSense(f) { return !!SENSE_FACILITIES[f] }
// 面议判定：供模板 :class 使用（与 list 同款；重构时曾遗漏，模板调用致白屏）
function isFace(fen) { return fen == null || fen <= 0 }
function formatPrice(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const text = Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
  return '¥' + text
}

async function fetchDetail(initial) {
  if (!siteId) {
    // 评审 P0：缺 id 不再白屏，直接进错误态
    errorMsg.value = '场地参数缺失，请返回后重试'
    return
  }
  const seq = ++fetchSeq
  // initial=false 为 onShow 静默重拉：不清骨架不闪空（评审 P2 状态中途过期）
  if (initial !== false) {
    loading.value = true
    errorMsg.value = ''
    site.value = null
  }
  try {
    const res = await request({ url: '/api/v1/test-sites/' + encodeURIComponent(siteId) })
    if (seq !== fetchSeq) return
    const d = (res && res.data) || res
    if (d && d.id) {
      site.value = d
    } else {
      errorMsg.value = '场地不存在或已下架'
    }
  } catch (e) {
    if (seq !== fetchSeq) return
    // 评审 P2：404（陈旧分享链）与网络错误区分，不再统一伪装成网络异常
    const code = e && (e.statusCode || e.status)
    errorMsg.value = code === 404 ? '场地不存在或已下架' : '网络异常，请稍后重试'
  } finally {
    if (seq === fetchSeq) loading.value = false
  }
}

let lastBookTap = 0
function goBooking(dateStr) {
  if (!canBook.value) {
    // 决策③：reserved/维护中禁用（视觉 disabled + 原因 toast 双保险）
    if (site.value) uni.showToast({ title: bookLabel.value, icon: 'none' })
    return
  }
  const now = Date.now()
  if (now - lastBookTap < 600) return // 双击防护（评审 P2）：连续 tap 只放行一次导航
  lastBookTap = now
  let url = '/pkg-service/pages/testsites/booking?id=' + encodeURIComponent(siteId)
  if (typeof dateStr === 'string' && dateStr) url += '&date=' + dateStr // 日期格点选：带该日期直达表单
  safeNavigateTo(url)
}

function onRetry() {
  if (retryLabel.value === '返回列表') { safeBack(); return }
  fetchDetail()
}

function goBack() {
  safeBack()
}

onShareAppMessage(() => {
  const s = site.value || {}
  return {
    title: (s.name || '测试场地') + ' · 无人机测试场地预约',
    path: '/pkg-service/pages/testsites/detail?id=' + encodeURIComponent(siteId),
  }
})

onLoad((options) => {
  checkMotion()
  siteId = (options && options.id) || ''
  fetchDetail()
})

onShow(() => {
  checkMotion()
  // 评审 P2：预约往返后静默重拉（已有数据不清骨架；首次 onShow 早于 fetchDetail 返回时跳过）
  if (siteId && site.value) fetchDetail(false)
})
</script>

<style scoped>
.tsd-page {
  min-height: 100vh;
  background: #fff; /* 白上白：白底页面 + 描边软角卡片（对齐组） */
  padding-bottom: 240rpx; /* 固定底栏（约 136rpx + safe-area）之上留呼吸，滚动到底不被遮 */
}

/* ===== 骨架（对齐组） ===== */
.sk-carousel {
  height: 400rpx;
  background: #EEF1F4;
}
.sk-section {
  background: #fff;
  margin: 16rpx 24rpx 0;
  padding: 32rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
}
.sk-l {
  height: 24rpx;
  background: #EDF0F3;
  border-radius: 8rpx;
  margin-bottom: 20rpx;
  animation: sk-pulse 1.4s linear infinite;
}
.sk-l.w40 { width: 40%; }
.sk-l.w50 { width: 50%; }
.sk-l.w60 { width: 60%; }
.sk-l.w70 { width: 70%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
@keyframes sk-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.55; } }

/* ===== 错误态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.stb {
  margin-top: 32rpx;
  min-height: 88rpx;
  padding: 0 48rpx;
  border-radius: 16rpx; /* 对齐组按钮：16rpx，非全圆（list 同款） */
  background: #0A66C2;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
}
.stb:active { opacity: 0.85; }

/* ===== 轮播：微构图占位（后端 image_url 就绪 → 换 <image>，结构不变） ===== */
.carousel {
  height: 400rpx;
  background: #EEF1F4;
  animation: fade-in 0.22s ease-out backwards;
}
.scene { position: relative; width: 100%; height: 100%; overflow: hidden; }
.scene-img { width: 100%; height: 100%; background: #EEF1F4; }
.scene-sky { position: absolute; top: 0; left: 0; right: 0; height: 58%; }
.scene-ground { position: absolute; bottom: 0; left: 0; right: 0; height: 42%; }
.scene-horizon { position: absolute; top: 58%; left: 0; right: 0; height: 4rpx; background: rgba(255,255,255,.55); }
.scene-runway {
  position: absolute;
  bottom: -18%;
  left: 50%;
  transform: translateX(-50%);
  width: 46%;
  height: 74%;
  background: rgba(255,255,255,.32);
  clip-path: polygon(38% 0, 62% 0, 90% 100%, 10% 100%);
}
.scene-runway::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  width: 4rpx;
  height: 100%;
  background: repeating-linear-gradient(180deg, rgba(255,255,255,.75) 0 18rpx, transparent 18rpx 32rpx);
  transform: translateX(-50%);
}
.scene-runway.narrow { width: 30%; height: 78%; }
.scene-blocks { position: absolute; bottom: 8%; display: flex; gap: 16rpx; }
.scene-block { border-radius: 8rpx; background: rgba(255,255,255,.38); }
.scene--1 .scene-sky { background: linear-gradient(180deg, #8fbbe8 0%, #cfe4f7 100%); }
.scene--1 .scene-ground { background: linear-gradient(180deg, #d8e0e8 0%, #b9c6d2 100%); }
.scene--2 .scene-sky { background: linear-gradient(180deg, #6f9bd0 0%, #a9c2e0 100%); }
.scene--2 .scene-ground { background: linear-gradient(180deg, #aeb9c6 0%, #8f9cac 100%); }
.scene--3 .scene-sky { background: linear-gradient(180deg, #c9d2dc 0%, #e4e9ee 100%); }
.scene--3 .scene-ground { background: linear-gradient(180deg, #c6cdd5 0%, #aeb6c0 100%); }
.scene--3 .scene-blocks { left: 12%; }
.scene--3 .scene-block:nth-child(1) { width: 80rpx; height: 52rpx; }
.scene--3 .scene-block:nth-child(2) { width: 60rpx; height: 72rpx; }
.scene--3 .scene-block:nth-child(3) { width: 72rpx; height: 44rpx; }
.scene--4 .scene-sky { background: linear-gradient(180deg, #a9d6d2 0%, #d6ecea 100%); }
.scene--4 .scene-ground { background: linear-gradient(180deg, #c4dcd9 0%, #a4c8c3 100%); }
.scene--4 .scene-blocks { right: 10%; }
.scene--4 .scene-block:nth-child(1) { width: 88rpx; height: 80rpx; }
.scene--4 .scene-block:nth-child(2) { width: 60rpx; height: 48rpx; }
.scene-label {
  position: absolute;
  left: 24rpx;
  bottom: 20rpx;
  font-size: 24rpx;
  font-weight: 600;
  color: #074D92; /* 夜航蓝 on 白玻璃 ≈7:1（原白字 on rgba(0,0,0,.18) ≈2:1 不达标） */
  background: rgba(255, 255, 255, 0.9);
  padding: 6rpx 20rpx;
  border-radius: 999rpx;
  box-shadow: 0 2rpx 8rpx rgba(10, 30, 60, 0.06); /* 轻影浮起，白玻璃在浅色场景上不失位 */
}

/* ===== 三区：白上白卡片（2rpx 描边 + 20rpx 软角 + 蓝调双层阴影，与 list 同构） ===== */
.detail-section {
  background: #fff;
  margin: 16rpx 24rpx 0;
  padding: 32rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.04),
    0 12rpx 32rpx rgba(10, 30, 60, 0.05);
  animation: fade-in 0.22s ease-out backwards;
}
.detail-section + .detail-section { animation-delay: 20ms; }
.detail-section + .detail-section + .detail-section { animation-delay: 40ms; }
.detail-section + .detail-section + .detail-section + .detail-section { animation-delay: 60ms; }
@keyframes fade-in {
  from { opacity: 0; transform: translateY(16rpx); }
  to { opacity: 1; transform: translateY(0); }
}

.detail-name {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
  margin-bottom: 16rpx;
  word-break: break-all; /* 硬化：超长场地名（英文/数字串）在卡片内断行，不撑破 */
}
.detail-tags {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
  margin-bottom: 16rpx;
}
/* 类型/状态角标：对齐组 8rpx 小徽章 + 类型 4 色相 / 状态暗变体（与 list 同款） */
.v-tag {
  display: inline-flex;
  align-items: center;
  min-height: 44rpx;
  padding: 0 14rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  font-weight: 700;
}
.vt--fly { color: #0A66C2; background: #EAF3FB; }
.vt--lab { color: #0B6E5F; background: #E6F5F1; }
.vt--chamber { color: #4A5AC8; background: #EEF0FB; }
.vt--tunnel { color: #074D92; background: #E7EEF6; }
.st--available { color: #0B6B41; background: #E9F7F0; }
.st--reserved { color: #B45309; background: #FFF4E5; }
.st--maintenance { color: #5D6B82; background: #EEF1F4; }
.detail-addr {
  display: flex;
  align-items: center;
  gap: 8rpx;
  font-size: 26rpx;
  color: #667085;
}
.addr-text { flex: 1; min-width: 0; }

.section-title {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #344054;
  margin-bottom: 24rpx;
}

/* 核心设备：胶囊标签（名称+职能；通信=蓝雾 / 定位感知=青雾，DESIGN.md 双令牌） */
.equip-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.equip-tag {
  display: inline-flex;
  align-items: baseline;
  gap: 8rpx;
  padding: 10rpx 20rpx;
  border-radius: 8rpx; /* 对齐组 chip（polish：999rpx 全圆轮廓读作按钮，系统语言全圆=可交互） */
  background: #EAF3FB;
}
.equip-tag.alt { background: #E6F5F1; }
.et-name {
  font-size: 26rpx;
  font-weight: 700;
  color: #0A66C2;
}
.equip-tag.alt .et-name { color: #0B6E5F; }
.et-role {
  font-size: 24rpx; /* 字阶底（polish：22rpx 阶梯外） */
  color: #344054; /* 深灰蓝 on 浅雾 ≈9:1，小字可读 */
}
.equip-empty {
  display: block;
  font-size: 24rpx;
  color: #667085;
}

.param-row {
  display: flex;
  justify-content: space-between;
  padding: 18rpx 0;
  font-size: 26rpx;
  border-bottom: 2rpx solid #f0f1f3;
}
.param-row:last-child { border-bottom: none; }
.param-label { color: #667085; flex-shrink: 0; margin-right: 24rpx; }
.param-value { color: #17212B; font-weight: 500; text-align: right; }
/* 预约安排：近 7 天可预约条（静态预览，今天格=塔台蓝描边+蓝雾底；占用情况以场地方确认为准） */
.day-block {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding: 18rpx 0;
  border-bottom: 2rpx solid #f0f1f3;
}
.day-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.day-title {
  font-size: 26rpx;
  color: #667085;
}
.day-more {
  font-size: 24rpx; /* 字阶底（polish：22rpx 阶梯外） */
  font-weight: 500;
  color: #0A66C2;
  padding: 8rpx 0 8rpx 16rpx; /* 扩大热区，右缘对齐 */
}
.day-strip {
  display: flex;
  gap: 12rpx;
}
.day-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6rpx;
  padding: 16rpx 0 14rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  transition: opacity 0.15s ease-out; /* 轻压反馈：小格不缩放，透明加深与全站轻压语言一致 */
}
.day-press {
  opacity: 0.7;
}
.day-cell.today {
  border-color: #0A66C2;
  background: #EAF3FB;
}
.day-text {
  font-size: 26rpx;
  font-weight: 600;
  color: #344054;
}
.day-cell.today .day-text { color: #0A66C2; }
.day-hint {
  display: block;
  font-size: 24rpx; /* 字阶底（polish：22rpx 阶梯外） */
  color: #667085;
}

/* 收费 */
.fee-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18rpx 0;
  font-size: 26rpx;
  border-bottom: 2rpx solid #f0f1f3;
}
.fee-item:last-child { border-bottom: none; }
.fee-label { color: #667085; flex-shrink: 0; margin-right: 24rpx; }
.fee-value {
  font-size: 30rpx; /* 中等级：金额强调 */
  font-weight: 600;
  color: #17212B;
  text-align: right;
}
.fee-value.sm { font-size: 26rpx; font-weight: 400; color: #5D6B82; }
.fee-value.face { font-size: 24rpx; font-weight: 600; color: #667085; } /* 面议降级：不是数字，不用金额重量渲染（list 同款） */

/* 底部栏：两行估价 + 胶囊 CTA */
.detail-bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  border-top: 2rpx solid #f0f1f3;
  box-shadow: 0 -4rpx 20rpx rgba(10, 30, 60, 0.05);
  z-index: 50;
  padding: 24rpx 32rpx calc(24rpx + env(safe-area-inset-bottom));
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
}
.estimate {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
  flex: 1;
  min-width: 0;
}
.est-lb {
  font-size: 24rpx;
  color: #667085;
}
.est-price {
  font-size: 40rpx; /* 大文本线（375px 下 20px），AA 3:1 达标 */
  font-weight: 800;
  color: #C2410C; /* 价格深色令牌：烬橙 #E96012 白底 ≈3:1 → 提深过 AA（评估 P3 对比度） */
  line-height: 1.2;
}
.est-price.face {
  font-size: 24rpx; /* 面议降级：不是数字，不用金额重量渲染（list 同款） */
  font-weight: 600;
  color: #667085;
}
.btn-book {
  height: 88rpx;
  padding: 0 56rpx;
  border-radius: 16rpx; /* 对齐组主行动按钮：16rpx（list CTA 同款），非全圆 */
  background: #0A66C2;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.35s cubic-bezier(0.34, 1.8, 0.64, 1), opacity 0.15s ease;
}
.btn-book-hover {
  transform: scale(0.96);
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.14), inset 0 -2rpx 6rpx rgba(7,77,146,.22), 0 2rpx 8rpx rgba(10,102,194,.18);
  transition-duration: 0.1s;
  transition-timing-function: linear;
}
.btn-book.disabled {
  background: #EEF1F4;
  color: #5D6B82; /* ≈4.9:1，禁用原因可读（评审 P2） */
  box-shadow: none;
}

/* 减弱动效（无障碍）：装饰动画全关 */
.no-motion .sk-l,
.no-motion .carousel,
.no-motion .detail-section { animation: none; }
.no-motion .btn-book-hover { transform: none !important; }
</style>
