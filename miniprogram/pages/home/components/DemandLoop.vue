<template>
  <!-- 供需项目锚点无限循环窗口（420px 高，接管窗口内纵向手势）
       顶部锚点用 2px 蓝色短线提示，不添加解释性文案 -->
  <view
    class="dl-window"
    @touchstart="onTouchStart"
    @touchmove.stop.prevent="onTouchMove"
    @touchend="onTouchEnd"
    @touchcancel="onTouchEnd"
  >
    <view class="dl-anchor-mark"></view>

    <view
      v-for="slot in slots"
      :key="slot.poolIndex"
      class="dl-card"
      :class="{ 'is-anchor': slot.progress > 0.5 }"
      :style="slot.style"
      @tap="onCardTap(slot)"
    >
      <!-- 固定全宽 150px 图片图层 + 裁剪窗口：图片不随尺寸重新裁剪，
           裁剪窗口连续揭示更多左上角区域 -->
      <view class="dl-media-wrap" :style="slot.mediaStyle">
        <image
          class="dl-media-img"
          mode="aspectFill"
          :src="slot.item.image"
          :style="slot.mediaImgStyle"
          @error="onImageError(slot.item)"
        />
      </view>
      <!-- 仅锚点卡大图区域可预览原图（与首页 featured-photo 行为一致）；
           紧凑卡图片不拦截点击，交由卡片进入详情 -->
      <view v-if="slot.progress > 0.5" class="dl-media-hit" @tap.stop="onMediaTap(slot)"></view>

      <!-- 两套固定尺寸信息层错峰切换：紧凑层（左图右文）/ 展开层（图下正文），
           变形期间只改 opacity/transform，不重排文字 -->
      <view class="dl-copy dl-compact" :style="slot.compactStyle">
        <view class="dl-badges">
          <text class="dl-type">{{ slot.item.type }}</text>
          <text v-if="slot.item.statusLabel" class="dl-status">{{ slot.item.statusLabel }}</text>
        </view>
        <text class="dl-title">{{ slot.item.title }}</text>
        <view v-if="slot.item.district || slot.item.publishedAt" class="dl-meta">
          <text v-if="slot.item.district" class="dl-meta-item">{{ slot.item.district }}</text>
          <text v-if="slot.item.publishedAt" class="dl-meta-item">{{ slot.item.publishedAt }}</text>
        </view>
        <view class="dl-foot">
          <view class="dl-budget">
            <text class="dl-budget-label">预算</text>
            <text class="dl-budget-value">{{ slot.item.budgetText }}</text>
          </view>
          <text v-if="slot.item.bidCount != null" class="dl-bid">已有 {{ slot.item.bidCount }} 家报价</text>
        </view>
      </view>

      <view class="dl-copy dl-expanded" :style="slot.expandedStyle">
        <view class="dl-badges">
          <text class="dl-type">{{ slot.item.type }}</text>
          <text v-if="slot.item.statusLabel" class="dl-status">{{ slot.item.statusLabel }}</text>
        </view>
        <text class="dl-title">{{ slot.item.title }}</text>
        <view v-if="slot.item.district || slot.item.publishedAt" class="dl-meta">
          <text v-if="slot.item.district" class="dl-meta-item">{{ slot.item.district }}</text>
          <text v-if="slot.item.publishedAt" class="dl-meta-item">{{ slot.item.publishedAt }}</text>
        </view>
        <text v-if="slot.item.description" class="dl-extra">{{ slot.item.description }}</text>
        <view class="dl-foot">
          <view class="dl-budget">
            <text class="dl-budget-label">预算</text>
            <text class="dl-budget-value">{{ slot.item.budgetText }}</text>
          </view>
          <text v-if="slot.item.bidCount != null" class="dl-bid">已有 {{ slot.item.bidCount }} 家报价</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'

const props = defineProps({
  items: { type: Array, default: () => [] },
})

const emit = defineEmits(['select', 'preview', 'image-error'])
const instance = getCurrentInstance()

/* ================= 几何参数（按批准原型，禁止自行改） ================= */
const COMPACT_HEIGHT = 112
const EXPANDED_HEIGHT = 290
const COMPACT_PITCH = 121
const COMPACT_IMAGE_WIDTH = 96
const EXPANDED_IMAGE_HEIGHT = 150
const CARD_GAP = 9
const SNAP_MIN_MS = 240
const SNAP_MAX_MS = 270
const SNAP_OMEGA = 0.026
const POOL_SIZE = 8
// 手势灵敏度：纵向拖动 230px 对应一个卡片间距
const DRAG_RATIO = 230
// 超过 7px 视为拖动，不触发详情
const DRAG_TAP_LIMIT = 7

/* ================= 工具函数 ================= */
const modulo = (value, length) => ((value % length) + length) % length
const clamp = (value, min, max) => Math.max(min, Math.min(max, value))
const lerp = (from, to, progress) => from + (to - from) * progress
const smoothstep = (from, to, value) => {
  const p = clamp((value - from) / (to - from), 0, 1)
  return p * p * (3 - 2 * p)
}
const cardHeightFor = (progress) =>
  lerp(COMPACT_HEIGHT, EXPANDED_HEIGHT, smoothstep(0.42, 0.88, progress))
// 统一的动画时钟：rAF 时间戳与 performance.now 同为页面时间原点，避免 Date.now 混用
const perfNow = () =>
  typeof performance !== 'undefined' && typeof performance.now === 'function'
    ? performance.now()
    : Date.now()

/* ================= 动画帧：全局优先，失败回退 16ms 定时器 ================= */
let raf =
  typeof requestAnimationFrame === 'function'
    ? requestAnimationFrame
    : (cb) => setTimeout(() => cb(perfNow()), 16)
let caf =
  typeof cancelAnimationFrame === 'function'
    ? cancelAnimationFrame
    : (handle) => clearTimeout(handle)

/* ================= 循环位置状态（position 为渲染源，ref 驱动 slots computed） ================= */
const position = ref(0)
const targetPosition = ref(0)
const windowWidth = ref(375)
// 响应式数据源：items 变化驱动 activeItems 替换，进而触发 slots computed 重算
const activeItems = ref([])
const itemCount = computed(() => activeItems.value.length)

// 非响应式运行态（每帧访问，避免额外响应式开销）
let dragging = false
let snapping = false
let snapFrom = 0
let snapTo = 0
let snapStartedAt = 0
let snapDuration = SNAP_MIN_MS
let snapDisplacement = 0
let snapVelocity = 0
let frameHandle = 0
let frameActive = false

// 手势
let dragStartY = 0
let dragStartPosition = 0
let dragDistance = 0
let inputVelocity = 0
let lastInputPosition = 0
let lastInputAt = 0

/* ================= 视图池 ================= */
const poolIndexToVirtual = (poolIndex) => {
  const base = Math.floor(position.value)
  return base - 2 + poolIndex
}

const geometryFor = (virtualIndex, base, fraction) => {
  const currentProgress = 1 - fraction
  const currentY = -COMPACT_PITCH * fraction
  const currentHeight = cardHeightFor(currentProgress)
  const nextY = currentY + currentHeight + CARD_GAP
  if (virtualIndex === base) return { progress: currentProgress, y: currentY }
  if (virtualIndex === base + 1) return { progress: fraction, y: nextY }
  if (virtualIndex > base + 1) {
    const nextHeight = cardHeightFor(fraction)
    return { progress: 0, y: nextY + nextHeight + CARD_GAP + (virtualIndex - base - 2) * COMPACT_PITCH }
  }
  return { progress: 0, y: -COMPACT_PITCH * (base - virtualIndex + fraction) }
}

const buildSlots = () => {
  const list = activeItems.value
  if (list.length <= 1) {
    const item = list[0]
    if (!item) return []
    // 唯一卡片静态展开，不制造重复卡片
    return [
      {
        poolIndex: 0,
        virtualIndex: 0,
        item,
        progress: 1,
        style: 'transform:translate3d(0,0,0);height:290px;visibility:visible',
        compactStyle: 'opacity:0;transform:translate3d(-5px,0,0)',
        expandedStyle: 'opacity:1;transform:translate3d(0,0,0)',
        mediaStyle: `width:${windowWidth.value}px;height:150px;left:0;top:0;overflow:hidden`,
        mediaImgStyle: `width:${windowWidth.value}px;height:150px;transform:translate3d(0,0,0);max-width:none`,
      },
    ]
  }

  const pos = position.value
  const base = Math.floor(pos)
  const fraction = pos - base
  const count = itemCount.value
  const slots = []
  for (let poolIndex = 0; poolIndex < POOL_SIZE; poolIndex++) {
    const virtualIndex = poolIndexToVirtual(poolIndex)
    const itemIndex = modulo(virtualIndex, count)
    const item = list[itemIndex]
    if (!item) continue

    const geometry = geometryFor(virtualIndex, base, fraction)
    const p = clamp(geometry.progress, 0, 1)

    const mediaWidth = lerp(COMPACT_IMAGE_WIDTH, windowWidth.value, p)
    const mediaHeight = lerp(COMPACT_HEIGHT, EXPANDED_IMAGE_HEIGHT, p)
    const cardHeight = cardHeightFor(p)
    const mediaInsetX = Math.max(0, (windowWidth.value - mediaWidth) / 2)
    const compactOpacity = 1 - smoothstep(0.15, 0.45, p)
    const expandedOpacity = smoothstep(0.55, 0.85, p)

    // 裁剪窗口（等价于原型 clip-path 的可见区域）：
    // 窗口固定在卡片左上角 (left:0, top:0)，宽高随 progress 从 96×112 连续
    // 展开到全宽×150；图片固定全宽 150 高，translateX(-insetX) 使可见区域
    // 始终取图片水平中心带 —— 紧凑态正好是 96×112 左图，展开态是完整大图。
    const mediaStyle = [
      `width:${mediaWidth.toFixed(2)}px`,
      `height:${mediaHeight.toFixed(2)}px`,
      'left:0',
      'top:0',
      'overflow:hidden',
    ].join(';')
    const mediaImgStyle = [
      `width:${windowWidth.value}px`,
      `height:${EXPANDED_IMAGE_HEIGHT}px`,
      `transform:translate3d(${(-mediaInsetX).toFixed(2)}px,0,0)`,
      'max-width:none',
    ].join(';')

    // 卡片仅允许 translate3d 移动 + 高度变化；越界槽位隐藏
    const hidden = geometry.y < -EXPANDED_HEIGHT || geometry.y > 470
    const cardStyle = [
      `transform:translate3d(0,${geometry.y.toFixed(2)}px,0)`,
      `height:${cardHeight.toFixed(2)}px`,
      `visibility:${hidden ? 'hidden' : 'visible'}`,
    ].join(';')

    const compactStyle = `opacity:${compactOpacity.toFixed(3)};transform:translate3d(${lerp(0, -5, p).toFixed(2)}px,0,0)`
    const expandedStyle = `opacity:${expandedOpacity.toFixed(3)};transform:translate3d(0,${lerp(6, 0, expandedOpacity).toFixed(2)}px,0)`

    slots.push({
      poolIndex,
      virtualIndex,
      item,
      progress: p,
      style: cardStyle,
      compactStyle,
      expandedStyle,
      mediaStyle,
      mediaImgStyle,
    })
  }
  return slots
}

const slots = computed(() => buildSlots())

/* ================= 数据同步 ================= */
const applyItems = () => {
  const next = (props.items || []).filter(Boolean)
  const previousItems = activeItems.value
  const previousCount = previousItems.length
  // 数据集合变化后优先保留当前业务 id；不存在时重置到第 0 项
  const anchorId =
    previousCount > 0
      ? previousItems[modulo(Math.round(position.value), previousCount)]?.id
      : null
  activeItems.value = next
  if (itemCount.value === 0) {
    stopMotion()
    return
  }
  if (itemCount.value === 1) {
    // 唯一卡片静态展开
    stopMotion()
    position.value = 0
    targetPosition.value = 0
    return
  }
  const anchorIndex = anchorId ? next.findIndex((it) => it.id === anchorId) : -1
  if (previousCount > 0 && anchorIndex >= 0) {
    // 保留当前锚点对应业务 id：把 position 移到该 id 的新下标，视觉锚点不动
    const oldAnchorIndex = modulo(Math.round(position.value), previousCount)
    if (anchorIndex !== oldAnchorIndex) {
      const shift = (anchorIndex - oldAnchorIndex) % itemCount.value
      const nextPos = Math.round(position.value) + shift
      position.value = nextPos
      targetPosition.value = nextPos
    }
  } else {
    stopMotion()
    position.value = 0
    targetPosition.value = 0
  }
}

watch(
  () => props.items,
  () => {
    applyItems()
  },
  { deep: false }
)

/* ================= 动画帧循环 ================= */
function animate(now) {
  const t = typeof now === 'number' ? now : perfNow()
  if (dragging) {
    position.value = targetPosition.value
    frameActive = false
    frameHandle = 0
    return
  }
  if (!snapping) {
    position.value = targetPosition.value
    frameActive = false
    frameHandle = 0
    return
  }
  const elapsed = Math.max(0, t - snapStartedAt)
  const progress = clamp(elapsed / snapDuration, 0, 1)
  const decay = Math.exp(-SNAP_OMEGA * elapsed)
  const offset = (snapDisplacement + (snapVelocity + SNAP_OMEGA * snapDisplacement) * elapsed) * decay
  position.value = snapTo + offset
  targetPosition.value = snapTo + offset
  if (progress < 1) {
    frameHandle = raf(animate)
    return
  }
  // 最后一帧直接落定目标整数，停止帧循环
  position.value = snapTo
  targetPosition.value = snapTo
  snapping = false
  frameActive = false
  frameHandle = 0
  normalizePosition()
}

function requestRender() {
  if (!frameActive) {
    frameActive = true
    frameHandle = raf(animate)
  }
}

function stopMotion() {
  snapping = false
  if (frameActive) {
    if (caf) caf(frameHandle)
    frameActive = false
    frameHandle = 0
  }
  targetPosition.value = position.value
}

function normalizePosition() {
  if (itemCount.value < 2) return
  const abs = Math.abs(position.value)
  if (abs > 10000) {
    const cycles = Math.trunc(position.value / itemCount.value) * itemCount.value
    position.value -= cycles
    targetPosition.value -= cycles
  }
}

/* ================= 松手吸附（临界阻尼，无回弹） ================= */
function sampleInput(nextPosition, now) {
  if (lastInputAt > 0) {
    const elapsed = Math.max(1, now - lastInputAt)
    const instantVelocity = (nextPosition - lastInputPosition) / elapsed
    inputVelocity = inputVelocity * 0.55 + instantVelocity * 0.45
  }
  lastInputPosition = nextPosition
  lastInputAt = now
}

function startSnap(nextPosition, releaseVelocity) {
  const velocity = releaseVelocity === undefined ? inputVelocity : releaseVelocity
  const distance = Math.abs(nextPosition - position.value)
  if (distance < 0.0008) {
    stopMotion()
    position.value = nextPosition
    targetPosition.value = nextPosition
    return
  }
  snapping = true
  snapFrom = position.value
  snapTo = nextPosition
  targetPosition.value = nextPosition
  snapStartedAt = perfNow()
  snapDuration = clamp(SNAP_MIN_MS + distance * 30, SNAP_MIN_MS, SNAP_MAX_MS)
  snapDisplacement = snapFrom - snapTo
  // 限制松手速度不会改变目标方向（单调逼近，不越过锚点）
  const monotonicLimit = -SNAP_OMEGA * snapDisplacement
  snapVelocity = clamp(velocity, Math.min(0, monotonicLimit), Math.max(0, monotonicLimit))
  inputVelocity = 0
  lastInputAt = 0
  requestRender()
}

/* ================= 手势 ================= */
function onTouchStart(e) {
  if (!e.touches || !e.touches.length) return
  const touch = e.touches[0]
  stopMotion()
  dragging = true
  dragStartY = touch.clientY || touch.pageY || 0
  dragStartPosition = position.value
  dragDistance = 0
  inputVelocity = 0
  lastInputPosition = position.value
  lastInputAt = perfNow()
}

function onTouchMove(e) {
  if (!dragging || !e.touches || !e.touches.length) return
  const touch = e.touches[0]
  const clientY = touch.clientY || touch.pageY || dragStartY
  const deltaY = clientY - dragStartY
  dragDistance = deltaY
  const next = dragStartPosition - deltaY / DRAG_RATIO
  targetPosition.value = next
  sampleInput(next, perfNow())
  requestRender()
}

function onTouchEnd() {
  if (!dragging) return
  const velocity = inputVelocity
  dragging = false
  startSnap(Math.round(targetPosition.value), velocity)
}

/* ================= 点击详情 / 图片预览 ================= */
function onCardTap(slot) {
  // 拖动超过 7px 不触发详情
  if (Math.abs(dragDistance) > DRAG_TAP_LIMIT) return
  emit('select', slot.item)
}

function onMediaTap(slot) {
  // 仅锚点大图可预览原图（与首页 featured-photo 行为一致）；拖动超过 7px 不触发
  if (Math.abs(dragDistance) > DRAG_TAP_LIMIT) return
  if (slot.progress > 0.5) {
    emit('preview', slot.item)
  }
}

function onImageError(item) {
  emit('image-error', item)
}

/* ================= 尺寸 ================= */
// 循环窗口宽度：必须以卡片实际容器宽度为准（首页 .surface-section 有 12px 边距），
// 不能用整屏 windowWidth，否则展开图宽度会溢出卡片。
function measureWidth() {
  // 先同步兜底（整屏宽 - 页面左右边距 24px），随后用实际元素宽度校准
  try {
    const sys = uni.getSystemInfoSync()
    const w = (sys && sys.windowWidth) || 375
    windowWidth.value = w - 24
  } catch (e) {
    windowWidth.value = 375 - 24
  }
  try {
    const query = uni.createSelectorQuery()
    if (instance && instance.proxy) query.in(instance.proxy)
    query
      .select('.dl-window')
      .boundingClientRect((rect) => {
        if (rect && rect.width > 0) {
          windowWidth.value = rect.width
        }
      })
      .exec()
  } catch (e) {
    // 保持同步兜底值
  }
}

/* ================= 生命周期 ================= */
onMounted(() => {
  measureWidth()
  applyItems()
})

onBeforeUnmount(() => {
  stopMotion()
  dragging = false
})
</script>

<style scoped>
/* ================= 循环窗口 ================= */
.dl-window {
  position: relative;
  height: 420px;
  overflow: hidden;
  border-top: 2px solid #dbeaf7;
  border-bottom: 1px solid #EEF1F4;
  touch-action: none;
  overscroll-behavior: contain;
  user-select: none;
  -webkit-user-select: none;
}

.dl-anchor-mark {
  position: absolute;
  z-index: 9;
  top: -2px;
  left: 0;
  width: 42px;
  height: 2px;
  background: #0A66C2;
  pointer-events: none;
}

/* ================= 卡片 ================= */
.dl-card {
  position: absolute;
  z-index: 1;
  left: 0;
  width: 100%;
  overflow: hidden;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  will-change: transform, height;
  box-sizing: border-box;
}
.dl-card.is-anchor {
  z-index: 4;
  border-color: #dbe8f4;
  box-shadow: 0 8px 24px rgba(16, 57, 91, 0.09);
}

/* 图片裁剪窗口：绝对定位遮罩宽高，图片全宽 150px 不重排，向左位移露出中心区域 */
.dl-media-wrap {
  position: absolute;
  z-index: 1;
  left: 0;
  top: 0;
  overflow: hidden;
  background: #d9e2e8;
}
.dl-media-img {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 150px;
  display: block;
  pointer-events: none;
}
/* 锚点卡大图点击热区：仅展开卡存在，转发预览 */
.dl-media-hit {
  position: absolute;
  z-index: 1;
  left: 0;
  top: 0;
  width: 100%;
  height: 150px;
  background: transparent;
}

/* ================= 两套固定文字层 ================= */
.dl-copy {
  position: absolute;
  z-index: 2;
  overflow: hidden;
  padding: 9px 10px;
  background: #fff;
  pointer-events: none;
  will-change: transform, opacity;
  box-sizing: border-box;
}

/* 紧凑层：左图右文，标题一行，不渲染描述 */
.dl-compact {
  top: 0;
  left: 96px;
  width: calc(100% - 96px);
  height: 112px;
}

/* 展开层：图下正文，标题最多两行，描述一行 */
.dl-expanded {
  top: 150px;
  left: 0;
  width: 100%;
  height: 140px;
}

.dl-badges {
  display: flex;
  align-items: center;
  gap: 5px;
}
.dl-type,
.dl-status {
  display: inline-flex;
  align-items: center;
  min-height: 18px;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: 700;
}
.dl-type {
  color: #074D92;
  background: #EAF3FB;
}
.dl-status {
  color: #168A55;
  background: #E9F7F0;
}

.dl-title {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  color: #17212B;
  font-weight: 700;
  line-height: 1.38;
}
.dl-compact .dl-title {
  margin-top: 5px;
  -webkit-line-clamp: 1;
  font-size: 13px;
}
.dl-expanded .dl-title {
  margin-top: 5px;
  -webkit-line-clamp: 2;
  font-size: 14px;
}

.dl-meta {
  margin-top: 5px;
  color: #667085;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 9px;
}
.dl-expanded .dl-meta {
  margin-top: 3px;
}

.dl-extra {
  margin-top: 3px;
  display: block;
  overflow: hidden;
  color: #667085;
  font-size: 9px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dl-foot {
  position: absolute;
  right: 10px;
  bottom: 8px;
  left: 10px;
  padding-top: 7px;
  border-top: 1px solid #EEF1F4;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 6px;
}
.dl-budget {
  display: flex;
  align-items: baseline;
  gap: 3px;
  color: #E96012;
}
.dl-budget-label {
  font-size: 9px;
  font-weight: 500;
}
.dl-budget-value {
  font-size: 15px;
  font-weight: 700;
  white-space: nowrap;
}
.dl-expanded .dl-budget-value {
  font-size: 16px;
}
.dl-bid {
  color: #98A2B3;
  font-size: 8px;
  white-space: nowrap;
}
</style>
