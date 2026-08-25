<template>
  <view class="ts-page" :class="{ 'no-motion': noMotion }" @tap="closeSort">
    <u-nav-bar title="测试场地" show-back @back="goBack" />

    <!-- 搜索（与 challenges/projects 同款：白上白描边 + 双层投影 + 搜索按钮） -->
    <view class="sbar">
      <view class="b-search">
        <image class="b-search-ic" src="/static/home/icons/search.svg" mode="aspectFit" />
        <input class="b-sinp" v-model="keyword" placeholder="搜索场地名称或位置" placeholder-class="b-ph" confirm-type="search" aria-label="搜索场地名称或位置" />
        <text v-if="keyword" class="b-sclr" aria-role="button" aria-label="清除搜索" @tap="clearSearch">×</text>
        <view class="b-sep"></view>
        <text class="b-sbtn" aria-role="button" aria-label="搜索" @tap="dismissSearch">搜索</text>
      </view>
    </view>

    <!-- 类型筛选：TOC 注线签名（本页差异点：轻量类型 tab，非下拉 pill） -->
    <view class="fbar">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="toc">
          <view
            v-for="p in typePills"
            :key="p.value"
            class="toc-item"
            :class="{ active: activeType === p.value }"
            aria-role="button"
            @tap.stop="selectType(p.value)"
          >{{ p.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 仅看可预约：独立开关行（本页差异点：可约性提为筛选维度） -->
    <view class="orow">
      <text class="olab">仅看可预约</text>
      <view class="oflex"></view>
      <switch :checked="onlyAvailable" color="#0A66C2" @change="onToggleOnly" aria-label="仅看可预约" />
    </view>

    <!-- 信息行：共 N 项 + 排序（与 challenges/projects 同构；骨架/错误/空态时不显示，防"共 0"闪现） -->
    <view v-if="!loading && !errorMsg && filteredSites.length > 0" class="ir">
      <text>共 <text class="irn">{{ filteredSites.length }}</text> 个场地 · <text class="irn">{{ availableCount }}</text> 个可预约</text>
      <text v-if="hasFilter" class="reset-chip" @tap="resetFilters">重置</text>
      <view class="irs-wrap" @tap.stop>
        <text class="irs" @tap="toggleSort">{{ sortLabel }} ▾</text>
        <view v-if="showSort" class="spop" :class="{ closing: sortClosing }" aria-role="menu" aria-label="排序选项" @tap.stop>
          <view v-for="s in SORTS" :key="s.v" class="sp-opt" :class="{ act: sort === s.v }" @tap="pickSort(s.v)">
            <text>{{ s.l }}</text><text v-if="sort === s.v" class="chk">✓</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 骨架（与 challenges/projects 同款：白卡描边 + 呼吸） -->
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
    <view v-else-if="errorMsg" class="st">
      <u-empty :description="errorMsg || '加载失败，请检查网络'">
        <view class="stb" @tap="fetchList">重新加载</view>
      </u-empty>
    </view>

    <!-- 空（三分：真空 / 筛选空 / 仅看可约空） -->
    <view v-else-if="filteredSites.length === 0" class="st">
      <u-empty :description="emptyText">
        <text v-if="emptySub" class="sth">{{ emptySub }}</text>
        <view v-if="createText" class="stb" @tap="onCreateAction">{{ createText }}</view>
      </u-empty>
    </view>

    <!-- 列表：白上白卡片，左图块（类型身份）+ 右侧信息，与 challenges/projects 同构 -->
    <view v-else class="cl">
      <view
        v-for="site in sortedSites"
        :key="site.id"
        class="card"
        hover-class="tap-scale"
        hover-start-time="0"
        hover-stay-time="120"
        @tap="goDetail(site)"
      >
        <!-- 图块：类型字符（"图片"锚点，替代全文字卡片的左缘色条） -->
        <view class="c-thumb" :class="'th--' + typeKey(site.site_type)">
          <text class="c-ch">{{ typeChar(site.site_type) }}</text>
        </view>

        <view class="c-body">
          <view class="c-top">
            <view class="c-badges">
              <text class="c-tag" :class="'tg--' + typeKey(site.site_type)">{{ typeLabel(site.site_type) }}</text>
              <text class="c-st" :class="'st--' + site.status">{{ statusLabel(site.status) }}</text>
            </view>
            <view class="c-price">
              <text class="vl" :class="{ face: isFace(site.price_fen) }">{{ formatPrice(site.price_fen) }}</text>
              <text class="lb">参考价</text>
            </view>
          </view>
          <text class="ct">{{ site.name }}</text>
          <text class="c-loc">{{ site.location || '位置待定' }}</text>
          <view class="c-bottom">
            <view v-if="facilityTags(site.facilities).length > 0" class="tags-row">
              <text v-for="(f, i) in facilityTags(site.facilities)" :key="i" class="fac-tag">{{ f }}</text>
            </view>
            <view class="bflex"></view>
            <!-- 可预约 → 直达预约；已约满/维护中 → 原因标签 -->
            <view
              v-if="site.status === 'available'"
              class="cta"
              hover-class="cta-hover"
              @tap.stop="goBooking(site)"
            >去预约</view>
            <text v-else class="locked">{{ lockedLabel(site) }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShow, onUnload, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { safeNavigateTo, safeBack } from '@/utils/nav'
import { useReduceMotion } from '@/utils/motion'

const SITE_TYPE_MAP = {
  flying_field: '飞行场地',
  lab: '实验室',
  anechoic_chamber: '暗室',
  wind_tunnel: '风洞',
}
const STATUS_MAP = { available: '可预约', maintenance: '维护中', reserved: '已约满' }
const FACILITY_MAP = { '5G': '5G', RTK: 'RTK', radar: '雷达', spectrum_analyzer: '频谱分析' }
const TYPE_KEYS = { flying_field: 'fly', lab: 'lab', anechoic_chamber: 'chamber', wind_tunnel: 'tunnel' }
const TYPE_CHARS = { flying_field: '飞', lab: '实', anechoic_chamber: '暗', wind_tunnel: '风' }

const SORTS = [
  { v: 'default', l: '默认排序' },
  { v: 'priceAsc', l: '价格从低到高' },
  { v: 'priceDesc', l: '价格从高到低' },
]
const SORT_LABEL = { default: '默认排序', priceAsc: '价格从低到高', priceDesc: '价格从高到低' }

const keyword = ref('')
const activeType = ref('all')
const onlyAvailable = ref(false)
const loading = ref(false)
const errorMsg = ref('')
const sites = ref([])
const sort = ref('default')
const showSort = ref(false)
const sortClosing = ref(false) // 排序弹层退场动画中（定时器到点再 v-if 移除，规范：退场必须存在）
let sortT = null
const { noMotion, checkMotion } = useReduceMotion()

// 请求序号：防竞态，旧响应晚到即丢弃
let fetchSeq = 0

const typePills = [
  { label: '全部', value: 'all' },
  { label: '飞行场地', value: 'flying_field' },
  { label: '实验室', value: 'lab' },
  { label: '暗室', value: 'anechoic_chamber' },
  { label: '风洞', value: 'wind_tunnel' },
]

// 可预约计数统一基于过滤后列表（与信息行同口径，不再同屏双数字）
const availableCount = computed(() => filteredSites.value.filter((s) => s.status === 'available').length)

const sortLabel = computed(() => SORT_LABEL[sort.value] || '默认排序')

// 全部过滤在本地（后端返回全量数组），零网络往返
const filteredSites = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return sites.value.filter((s) => {
    if (onlyAvailable.value && s.status !== 'available') return false
    if (activeType.value !== 'all' && s.site_type !== activeType.value) return false
    if (!q) return true
    // 搜索范围含类型中文名，避免"风洞"搜不到
    const hay = [
      s.name,
      s.location,
      typeLabel(s.site_type),
      ...(s.facilities || []).map((f) => FACILITY_MAP[f] || f),
    ].join(' ').toLowerCase()
    return hay.indexOf(q) >= 0
  })
})

// 排序：面议永远排最后（priceDesc 下 Infinity 不再误排最前）
const sortedSites = computed(() => {
  const face = []
  const priced = []
  for (const s of filteredSites.value) {
    if (s.price_fen == null || s.price_fen <= 0) face.push(s)
    else priced.push(s)
  }
  if (sort.value === 'priceAsc') priced.sort((a, b) => a.price_fen - b.price_fen)
  else if (sort.value === 'priceDesc') priced.sort((a, b) => b.price_fen - a.price_fen)
  return priced.concat(face)
})

// 空态三分：真空 / 筛选空 / 仅看可约空，动作各不相同（诊断顺序：关键词/类型优先于开关——真实原因优先展示）
const emptyText = computed(() => {
  if (keyword.value.trim() || activeType.value !== 'all') return '未找到匹配的场地'
  if (onlyAvailable.value) return '没有可预约的场地'
  return '暂无测试场地'
})
const emptySub = computed(() => {
  if (keyword.value.trim() || activeType.value !== 'all') return '试试更换关键词或筛选条件'
  if (onlyAvailable.value) return '当前没有可预约的测试场地'
  return ''
})
const createText = computed(() => {
  if (keyword.value.trim() || activeType.value !== 'all') return '清除筛选'
  if (onlyAvailable.value) return '查看全部场地'
  return ''
})

function typeLabel(t) { return SITE_TYPE_MAP[t] || t || '测试场地' }
function typeKey(t) { return TYPE_KEYS[t] || 'fly' }
function typeChar(t) { return TYPE_CHARS[t] || '测' }
function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function facilityTags(list) {
  return (list || []).map((f) => FACILITY_MAP[f] || f)
}
function formatPrice(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const text = Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
  return '¥' + text
}
function isFace(fen) { return fen == null || fen <= 0 }

async function fetchList(opts = {}) {
  const silent = !!opts.silent
  const seq = ++fetchSeq
  if (!silent) {
    loading.value = true
    errorMsg.value = ''
  }
  try {
    const res = await request({ url: '/api/v1/test-sites' })
    if (seq !== fetchSeq) return
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    sites.value = Array.isArray(data) ? data : (data && data.items) || []
    errorMsg.value = ''
  } catch (e) {
    if (seq !== fetchSeq) return
    // 已有内容时不整页报错（stale-while-revalidate），只提示
    if (sites.value.length === 0) {
      errorMsg.value = '网络异常，请稍后重试'
    } else {
      uni.showToast({ title: '刷新失败，请稍后重试', icon: 'none' })
    }
  } finally {
    // 当前序号请求完成即清 loading（含 silent 刷新超越初始请求的边缘：防骨架屏卡死）
    if (seq === fetchSeq && loading.value) loading.value = false
  }
}

function selectType(value) {
  if (activeType.value === value) return
  activeType.value = value
  closeSortImmediately() // 切类型时收起排序弹层，防弹层悬留
}

function onToggleOnly(e) {
  onlyAvailable.value = !!(e && e.detail && e.detail.value)
}

function onCreateAction() {
  // 与空态诊断顺序对齐：先清筛选类（关键词/类型），再清开关——一次点击清到能出结果
  if (keyword.value.trim() || activeType.value !== 'all') {
    keyword.value = ''
    activeType.value = 'all'
    return
  }
  if (onlyAvailable.value) onlyAvailable.value = false
}

function clearSearch() {
  keyword.value = ''
}

// 「搜索」按钮：实时过滤已即时生效，按钮承担收起键盘手势（避免与清空行为反向）
function dismissSearch() {
  uni.hideKeyboard()
}

function closeSortImmediately() {
  clearTimeout(sortT); sortT = null
  sortClosing.value = false
  showSort.value = false
}

function startCloseSort() {
  if (sortClosing.value) return // 已在退场中，防重复触发叠加定时器
  sortClosing.value = true
  sortT = setTimeout(() => {
    showSort.value = false
    sortClosing.value = false
    sortT = null
  }, 150)
}

function toggleSort() {
  if (showSort.value) { startCloseSort(); return }
  closeSortImmediately() // 清残留退场态，防同帧双动画
  showSort.value = true
}

function closeSort() {
  if (showSort.value) startCloseSort()
}

function pickSort(v) {
  sort.value = v
  startCloseSort()
}

// 任一筛选激活（关键词/类型/仅看可约）且结果非空 → 信息行提供「重置」出口（评审 P2）
const hasFilter = computed(() => !!(keyword.value.trim() || activeType.value !== 'all' || onlyAvailable.value))

function resetFilters() {
  keyword.value = ''
  activeType.value = 'all'
  onlyAvailable.value = false
}

// 已约满与维护中分开呈现（评审 P3：原同一灰字压平两个状态，保留色+文本双通道）
function lockedLabel(site) {
  return site.status === 'reserved' ? '已约满，暂不可预约' : '维护中暂不可预约'
}

function goDetail(site) {
  safeNavigateTo('/pkg-service/pages/testsites/detail?id=' + encodeURIComponent(site.id))
}

let lastBookTap = 0
function goBooking(site) {
  // 双击防护（评审 P1：整卡与内嵌 CTA 双触达目标，补 detail 同款 600ms 守卫）
  const now = Date.now()
  if (now - lastBookTap < 600) return
  lastBookTap = now
  safeNavigateTo('/pkg-service/pages/testsites/booking?id=' + encodeURIComponent(site.id))
}

function goBack() {
  safeBack()
}

let shownOnce = false
onLoad(() => {
  checkMotion()
  fetchList()
})
onUnload(() => {
  clearTimeout(sortT) // 规范：页面卸载清除退场定时器，防回调泄漏
})
// onShow 静默刷新：预约提交返回后立即看到最新状态（首次 onShow 与 onLoad 重复，用 flag 去重）
onShow(() => {
  if (shownOnce) fetchList({ silent: true })
  shownOnce = true
})
onPullDownRefresh(() => {
  fetchList({ silent: true }).finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.ts-page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 极淡灰投影（与 challenges/projects 同款） ===== */
.sbar {
  padding: 12rpx 24rpx 8rpx;
  background: #fff;
}
.b-search {
  height: 88rpx;
  padding: 0 22rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  background: #fff;
  /* 小红书风白上白：接触阴影贴地 + 环境阴影弥散浮起，蓝调低透明（各层 ≤8%），无重影 */
  box-shadow:
    0 2rpx 4rpx rgba(10, 30, 60, 0.04),
    0 10rpx 24rpx rgba(10, 30, 60, 0.05);
  display: flex;
  align-items: center;
  gap: 14rpx;
  box-sizing: border-box;
}
.b-search-ic { width: 30rpx; height: 30rpx; flex: none; }
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 26rpx; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 30rpx; padding: 28rpx; margin: -28rpx; } /* 热区扩大至 ≈44px：视觉 × 外扩，点击不脱靶 */
.b-sep { width: 2rpx; height: 30rpx; background: #DDE1E6; margin: 0 18rpx 0 12rpx; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 26rpx; line-height: 1; padding: 12rpx 4rpx 12rpx 0; }
.b-sbtn:active { opacity: 0.5; }
.b-sclr:active { opacity: 0.6; }

/* ===== 类型筛选：TOC 注线签名（本页差异点） ===== */
.fbar {
  background: #fff;
  padding: 8rpx 24rpx 0;
}
.filter-scroll {
  white-space: nowrap;
}
.toc {
  display: inline-flex;
  gap: 40rpx;
}
.toc-item {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  padding: 0 8rpx;
  font-size: 24rpx; /* 筛选器字体同研发难题 12px（注线样式保留） */
  color: #667085;
}
.toc-item.active {
  color: #0A66C2;
  font-weight: 600;
}
.toc-item.active::after {
  content: '';
  position: absolute;
  left: 8rpx;
  right: 8rpx;
  bottom: 16rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #0A66C2;
  animation: toc-in 0.22s ease-out; /* 规范：非浮层动画用 ease-out（第三枚手写曲线已撤） */
}
@keyframes toc-in {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}

/* ===== 仅看可预约开关行（本页差异点） ===== */
.orow {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 0 24rpx 16rpx;
}
.olab {
  font-size: 24rpx; /* 同研发难题元信息 12px */
  color: #17212B;
}
.oflex { flex: 1; }

/* ===== 信息行：共 N 项 + 排序（与 challenges/projects 同构） ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4rpx 28rpx 12rpx;
  font-size: 24rpx;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }
/* 筛选激活时的一键重置（评审 P2：结果越稀缺恢复成本越高，反向纠正） */
.reset-chip {
  color: #0A66C2;
  font-weight: 500;
  margin-left: 16rpx;
  min-height: 88rpx; /* 触达标准 */
  display: flex;
  align-items: center;
  flex-shrink: 0; /* polish：信息行窄屏三元素不挤压（共 N · 重置 · 排序） */
}
.irs-wrap { position: relative; }
.irs {
  color: #0A66C2;
  font-weight: 500;
  min-height: 88rpx; /* 触达目标：28px → 44px */
  display: flex;
  align-items: center;
  padding: 0 8rpx 0 24rpx;
}
.irs:active { opacity: 0.7; }
.spop {
  position: absolute;
  top: 64rpx;
  right: 0;
  z-index: 90;
  background: #fff;
  border-radius: 20rpx;
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.06),
    0 16rpx 48rpx rgba(10, 30, 60, 0.08);
  padding: 12rpx;
  min-width: 280rpx;
  animation: spop-in 0.22s cubic-bezier(0.32, 0.72, 0, 1);
}
.spop.closing {
  animation: spop-out 0.15s ease-in forwards; /* 退场 = 进场 ×0.7，forwards 保持结束态到 v-if 移除 */
}
@keyframes spop-in {
  from { opacity: 0; transform: translateY(-8rpx); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes spop-out {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-8rpx); }
}
.sp-opt {
  min-height: 88rpx; /* 触达目标：33px → 44px */
  padding: 20rpx 28rpx;
  border-radius: 16rpx;
  font-size: 26rpx;
  color: #17212B;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sp-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }
.chk { color: #0A66C2; font-size: 24rpx; }

/* ===== 骨架（与 challenges/projects 同款） ===== */
.skl { display: flex; flex-direction: column; gap: 16rpx; padding: 0 24rpx; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  background: #fff;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
}
.sk-row { display: flex; align-items: center; gap: 16rpx; }
.sk-tag { width: 112rpx; height: 36rpx; border-radius: 8rpx; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 16rpx; }
.sk-l { height: 24rpx; background: #EDF0F3; border-radius: 8rpx; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态（与 challenges/projects 同构） ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.sth { font-size: 24rpx; color: #667085; display: block; margin-bottom: 32rpx; }
.stb {
  padding: 16rpx 48rpx;
  border-radius: 16rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 26rpx;
  font-weight: 500;
  /* 同 CTA：顶部内高光 + 品牌蓝晕 */
  box-shadow:
    inset 0 2rpx 0 rgba(255, 255, 255, 0.2),
    0 4rpx 14rpx rgba(10, 102, 194, 0.25);
}
.stb:active { opacity: 0.85; }

/* ===== 列表卡片：白上白（描边 + 极淡投影浮起），左图块 + 右信息 ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 0 24rpx;
}
.card {
  display: flex;
  align-items: center; /* 图块垂直居中，消除底部空洞 */
  gap: 20rpx;
  padding: 24rpx;
  background: #fff;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
  /* 双层柔阴影：接触层贴地 + 弥散层浮起，蓝调低透明（≤8%），去"飘"感 */
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.04),
    0 12rpx 32rpx rgba(10, 30, 60, 0.05);
  /* 按压弹簧：按下 .1s linear 即时到位，松手 .35s ios-pop 回位（与 challenges 同手感） */
  transition: transform 0.35s cubic-bezier(0.34, 1.8, 0.64, 1), opacity 0.15s ease;
}
.tap-scale {
  transform: scale(0.95);
  opacity: 0.9;
  /* 按压物理感：阴影随卡片收拢 */
  box-shadow:
    0 1rpx 3rpx rgba(10, 30, 60, 0.03),
    0 6rpx 16rpx rgba(10, 30, 60, 0.04);
  transition-duration: 0.1s;
  transition-timing-function: linear;
}

/* 图块：类型字符（"图片"锚点，蓝族 4 色相低饱和） */
.c-thumb {
  width: 128rpx;
  height: 128rpx;
  border-radius: 20rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.th--fly { background: #EAF3FB; }
.th--lab { background: #E6F5F1; }
.th--chamber { background: #EEF0FB; }
.th--tunnel { background: #E7EEF6; }
.c-ch {
  font-size: 40rpx;
  font-weight: 700;
  color: #0A66C2;
}
.th--lab .c-ch { color: #0E8F7E; }
.th--chamber .c-ch { color: #4A5AC8; }
.th--tunnel .c-ch { color: #074D92; }

.c-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}
.c-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}
.c-badges { display: flex; gap: 12rpx; }
.c-tag, .c-st {
  display: inline-flex;
  align-items: center;
  min-height: 44rpx;
  padding: 0 14rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  font-weight: 700;
}
.tg--fly { color: #0A66C2; background: #EAF3FB; }
/* lab 小字加深至 AA 4.5+（#0B6E5F on #E6F5F1 ≈5.4:1）；图块大字保留 #0E8F7E（40rpx 粗体达大文本线 3:1） */
.tg--lab { color: #0B6E5F; background: #E6F5F1; }
.tg--chamber { color: #4A5AC8; background: #EEF0FB; }
.tg--tunnel { color: #074D92; background: #E7EEF6; }
.st--available { color: #0B6B41; background: #E9F7F0; }
.st--reserved { color: #B45309; background: #FFF4E5; }
.st--maintenance { color: #5D6B82; background: #EEF1F4; }

/* 价格右上：烬橙大数字（决策②锁定），面议降级灰（与 challenges/projects 同构） */
.c-price {
  display: flex;
  align-items: baseline;
  gap: 6rpx;
  flex-shrink: 0;
}
/* 价格数值 18px/800 烬橙：同研发难题/课题攻关 .c-budget .vl 金额基准（需求方指示对齐）；
   无 price_unit 期间「参考价」标签兜底语义，单位字段合入后恢复全量展示 */
.c-price .vl {
  font-size: 36rpx; /* 18px：同研发难题金额基准 */
  font-weight: 800;
  color: #C2410C; /* 烬橙：同研发难题/课题攻关金额色 */
  line-height: 1.2;
}
.c-price .vl.face {
  font-size: 26rpx; /* 13px：同研发难题 .vl.face 面议档 */
  font-weight: 600;
  color: #667085;
}
.c-price .lb {
  font-size: 24rpx; /* 与正文下限一致（参照页 12px 下限，不跌破） */
  color: #667085;
}

.ct {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-loc {
  font-size: 24rpx;
  color: #667085;
  line-height: 1.5;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.c-bottom {
  display: flex;
  align-items: center;
  margin-top: 6rpx;
}
.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
}
.fac-tag {
  font-size: 24rpx;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
}
.bflex { flex: 1; }
.cta {
  height: 88rpx;
  padding: 0 40rpx;
  border-radius: 16rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 26rpx; /* 按钮 13px：同研发难题 .stb */
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  /* 精致细节：顶部内高光受光 + 底部内阴影厚度 + 品牌蓝晕投影 */
  box-shadow:
    inset 0 2rpx 0 rgba(255, 255, 255, 0.22),
    inset 0 -4rpx 10rpx rgba(7, 77, 146, 0.18),
    0 4rpx 14rpx rgba(10, 102, 194, 0.25);
  transition: transform 0.35s cubic-bezier(0.34, 1.8, 0.64, 1), opacity 0.15s ease;
}
.cta-hover {
  transform: scale(0.96);
  /* 按压：高光收紧、蓝晕收拢；按下即时，松手弹簧回位 */
  box-shadow:
    inset 0 2rpx 0 rgba(255, 255, 255, 0.14),
    inset 0 -2rpx 6rpx rgba(7, 77, 146, 0.22),
    0 2rpx 8rpx rgba(10, 102, 194, 0.18);
  transition-duration: 0.1s;
  transition-timing-function: linear;
}
.locked {
  font-size: 24rpx;
  color: #667085;
  flex-shrink: 0;
}

/* ===== 动效（对齐 challenges/projects 规范：白名单 transform/opacity，错峰 ≤6 项） ===== */
/* 卡片入场：前 6 项 20ms 错峰淡入上移（backwards 填充防闪跳） */
.card { animation: none; }
.card:nth-child(-n+6) { animation: card-in 0.22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes card-in {
  from { opacity: 0; transform: translateY(20rpx); }
  to { opacity: 1; transform: translateY(0); }
}
/* 图块与卡片同拍"点亮"（与 challenges 色条抽出同构） */
.card:nth-child(-n+6) .c-thumb { animation: thumb-in 0.2s ease-out backwards; }
.card:nth-child(1) .c-thumb { animation-delay: 80ms; }
.card:nth-child(2) .c-thumb { animation-delay: 100ms; }
.card:nth-child(3) .c-thumb { animation-delay: 120ms; }
.card:nth-child(4) .c-thumb { animation-delay: 140ms; }
.card:nth-child(5) .c-thumb { animation-delay: 160ms; }
.card:nth-child(6) .c-thumb { animation-delay: 180ms; }
@keyframes thumb-in {
  from { opacity: 0; transform: scale(0.92); }
  to { opacity: 1; transform: scale(1); }
}

/* 骨架呼吸（加载中环境光；一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: sk-pulse 1.4s linear infinite; }
@keyframes sk-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.55; } }

/* 减弱动效（无障碍）：装饰动画/位移缩放全关，保留淡入与颜色反馈 */
.no-motion .card,
.no-motion .c-thumb,
.no-motion .sk-tag,
.no-motion .sk-l,
.no-motion .toc-item.active::after,
.no-motion .spop { animation: none; }
.no-motion .tap-scale,
.no-motion .cta-hover { transform: none !important; }
</style>
