<template>
  <!-- 供需项目锚点无限循环窗口（420px 高，接管窗口内纵向手势）
       顶部锚点用 2px 蓝色短线提示，不添加解释性文案 -->
  <view
    class="dl-window"
    data-eventsync="true"
    :data-loop-width="windowWidth"
    :data-loop-disabled="itemCount < 2"
    :loopVersion="renderVersion"
    :change:loopVersion="loopGesture.onVersionChange"
    @touchstart="loopGesture.touchstart"
    @touchmove.stop="loopGesture.touchmove"
    @touchend="loopGesture.touchend"
    @touchcancel="loopGesture.touchcancel"
  >
    <view class="dl-anchor-mark"></view>

    <view
      v-for="slot in slots"
      :key="slot.poolIndex"
      class="dl-card"
      :class="[`dl-card-${slot.poolIndex}`, { 'is-anchor': slot.virtualOffset === 0 }]"
      :data-pool-index="slot.poolIndex"
      :style="slot.style"
      @tap="loopGesture.cardTap"
      @transitionend="loopGesture.transitionend"
    >
      <!-- 固定全宽 150px 图片图层 + 裁剪窗口：图片不随尺寸重新裁剪，
           裁剪窗口连续揭示更多左上角区域 -->
      <view
        class="dl-media-wrap"
        :class="`dl-media-${slot.poolIndex}`"
        :style="slot.mediaStyle"
      >
        <image
          class="dl-media-img"
          :class="`dl-img-${slot.poolIndex}`"
          mode="aspectFill"
          :src="slot.item.image"
          :style="slot.mediaImgStyle"
          @error="onImageError(slot.item)"
        />
      </view>
      <!-- 仅锚点卡大图区域可预览原图（与首页 featured-photo 行为一致）；
           紧凑卡图片不拦截点击，交由卡片进入详情 -->
      <view
        v-if="slot.virtualOffset === 0"
        class="dl-media-hit"
        :data-pool-index="slot.poolIndex"
        @tap.stop="loopGesture.mediaTap"
      ></view>

      <!-- 两套固定尺寸信息层错峰切换：紧凑层（左图右文）/ 展开层（图下正文），
           变形期间只改 opacity/transform，不重排文字 -->
      <view
        class="dl-copy dl-compact"
        :class="`dl-compact-${slot.poolIndex}`"
        :style="slot.compactStyle"
      >
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

      <view
        class="dl-copy dl-expanded"
        :class="`dl-expanded-${slot.poolIndex}`"
        :style="slot.expandedStyle"
      >
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

<script>
const COMPACT_HEIGHT = 112
const EXPANDED_HEIGHT = 290
const COMPACT_PITCH = 121
const COMPACT_IMAGE_WIDTH = 96
const EXPANDED_IMAGE_HEIGHT = 150
const CARD_GAP = 9
const POOL_SIZE = 8

const modulo = (value, length) => ((value % length) + length) % length

export default {
  name: 'DemandLoop',
  props: {
    items: { type: Array, default: () => [] },
  },
  emits: ['select', 'preview', 'image-error'],
  data() {
    return {
      activeItems: [],
      anchorIndex: 0,
      renderVersion: 0,
      windowWidth: 351,
    }
  },
  computed: {
    itemCount() {
      return this.activeItems.length
    },
    slots() {
      if (this.itemCount === 0) return []
      if (this.itemCount === 1) return [this.buildStaticSlot(0, 0, this.activeItems[0], true)]

      const slots = []
      for (let poolIndex = 0; poolIndex < POOL_SIZE; poolIndex += 1) {
        const virtualOffset = poolIndex - 2
        const itemIndex = modulo(this.anchorIndex + virtualOffset, this.itemCount)
        slots.push(this.buildStaticSlot(poolIndex, virtualOffset, this.activeItems[itemIndex], false))
      }
      return slots
    },
  },
  watch: {
    items: {
      handler() {
        this.applyItems()
      },
      immediate: true,
    },
  },
  mounted() {
    this.measureWidth()
  },
  methods: {
    buildStaticSlot(poolIndex, virtualOffset, item, single) {
      let y = 0
      let progress = 0
      if (single || virtualOffset === 0) {
        progress = 1
      } else if (virtualOffset === 1) {
        y = EXPANDED_HEIGHT + CARD_GAP
      } else if (virtualOffset > 1) {
        y = EXPANDED_HEIGHT + CARD_GAP + (virtualOffset - 1) * COMPACT_PITCH
      } else {
        y = virtualOffset * COMPACT_PITCH
      }

      const expanded = progress === 1
      const mediaWidth = expanded ? this.windowWidth : COMPACT_IMAGE_WIDTH
      const mediaHeight = expanded ? EXPANDED_IMAGE_HEIGHT : COMPACT_HEIGHT
      const mediaInsetX = Math.max(0, (this.windowWidth - mediaWidth) / 2)
      return {
        poolIndex,
        virtualOffset,
        item,
        style: `transform:translate3d(0,${y}px,0);height:${expanded ? EXPANDED_HEIGHT : COMPACT_HEIGHT}px;visibility:${y < -EXPANDED_HEIGHT || y > 470 ? 'hidden' : 'visible'}`,
        compactStyle: `opacity:${expanded ? 0 : 1};transform:translate3d(${expanded ? -5 : 0}px,0,0)`,
        expandedStyle: `opacity:${expanded ? 1 : 0};transform:translate3d(0,${expanded ? 0 : 6}px,0)`,
        mediaStyle: `width:${mediaWidth}px;height:${mediaHeight}px;left:0;top:0;overflow:hidden`,
        mediaImgStyle: `width:${this.windowWidth}px;height:${EXPANDED_IMAGE_HEIGHT}px;transform:translate3d(${-mediaInsetX}px,0,0);max-width:none`,
      }
    },
    applyItems() {
      const next = (this.items || []).filter(Boolean)
      const current = this.activeItems[this.anchorIndex]
      const anchorId = current && current.id
      this.activeItems = next
      const retainedIndex = anchorId ? next.findIndex((item) => item && item.id === anchorId) : -1
      this.anchorIndex = retainedIndex >= 0 ? retainedIndex : 0
      this.renderVersion += 1
    },
    commitLoopStep(payload) {
      if (this.itemCount < 2) return
      const rawStep = payload && typeof payload === 'object' ? payload.step : payload
      const step = Number(rawStep)
      if (step !== 1 && step !== -1) return
      this.anchorIndex = modulo(this.anchorIndex + step, this.itemCount)
      // 锚点轮换和 WXS 复位在同一次 setData 中提交，避免一帧内容错位。
      this.renderVersion += 1
    },
    selectPoolItem(payload) {
      const poolIndex = Number(payload && payload.poolIndex)
      const slot = this.slots.find((candidate) => candidate.poolIndex === poolIndex)
      if (slot) this.$emit('select', slot.item)
    },
    previewPoolItem(payload) {
      const poolIndex = Number(payload && payload.poolIndex)
      const slot = this.slots.find((candidate) => candidate.poolIndex === poolIndex)
      if (slot && slot.virtualOffset === 0) this.$emit('preview', slot.item)
    },
    onImageError(item) {
      this.$emit('image-error', item)
    },
    measureWidth() {
      try {
        const system = uni.getSystemInfoSync()
        this.windowWidth = ((system && system.windowWidth) || 375) - 24
      } catch (error) {
        this.windowWidth = 351
      }
      this.$nextTick(() => {
        try {
          uni
            .createSelectorQuery()
            .in(this)
            .select('.dl-window')
            .boundingClientRect((rect) => {
              if (rect && rect.width > 0 && rect.width !== this.windowWidth) {
                this.windowWidth = rect.width
                this.renderVersion += 1
              }
            })
            .exec()
        } catch (error) {
          // 保留同步取得的兜底宽度。
        }
      })
    },
  },
}
</script>

<!-- 高频手势和吸附只在小程序视图层运行，避免每帧穿过逻辑层通信桥。 -->
<script module="loopGesture" lang="wxs">
var COMPACT_HEIGHT = 112
var EXPANDED_HEIGHT = 290
var COMPACT_PITCH = 121
var COMPACT_IMAGE_WIDTH = 96
var EXPANDED_IMAGE_HEIGHT = 150
var CARD_GAP = 9
var POOL_SIZE = 8
var DRAG_RATIO = 230
var TAP_LIMIT = 7

var owner = null
var width = 351
var dragging = false
var locked = false
var moved = false
var startY = 0
var position = 0
var velocity = 0
var lastPosition = 0
var lastAt = 0
var settleTarget = 0

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value))
}

function lerp(from, to, progress) {
  return from + (to - from) * progress
}

function smoothstep(from, to, value) {
  var progress = clamp((value - from) / (to - from), 0, 1)
  return progress * progress * (3 - 2 * progress)
}

function cardHeight(progress) {
  return lerp(COMPACT_HEIGHT, EXPANDED_HEIGHT, smoothstep(0.42, 0.88, progress))
}

function geometry(offset, base, fraction) {
  var currentProgress = 1 - fraction
  var currentY = -COMPACT_PITCH * fraction
  var currentHeight = cardHeight(currentProgress)
  var nextY = currentY + currentHeight + CARD_GAP
  if (offset === base) return { progress: currentProgress, y: currentY }
  if (offset === base + 1) return { progress: fraction, y: nextY }
  if (offset > base + 1) {
    var nextHeight = cardHeight(fraction)
    return {
      progress: 0,
      y: nextY + nextHeight + CARD_GAP + (offset - base - 2) * COMPACT_PITCH,
    }
  }
  return { progress: 0, y: -COMPACT_PITCH * (base - offset + fraction) }
}

function select(ins, selector) {
  return ins && ins.selectComponent ? ins.selectComponent(selector) : null
}

function setStyle(node, style) {
  if (node && node.setStyle) node.setStyle(style)
}

function transitionFor(duration, properties) {
  if (!duration) return 'none'
  var names = properties.split(',')
  var transitions = []
  for (var index = 0; index < names.length; index += 1) {
    transitions.push(names[index] + ' ' + duration + 'ms cubic-bezier(0.22,0.8,0.32,1)')
  }
  return transitions.join(',')
}

function renderAt(nextPosition, ins, duration) {
  if (!ins) return
  position = nextPosition
  var base = Math.floor(nextPosition)
  var fraction = nextPosition - base
  var cardTransition = transitionFor(duration, 'transform,height')
  var mediaTransition = transitionFor(duration, 'width,height')
  var imageTransition = transitionFor(duration, 'transform')
  var copyTransition = transitionFor(duration, 'opacity,transform')

  for (var poolIndex = 0; poolIndex < POOL_SIZE; poolIndex += 1) {
    var offset = poolIndex - 2
    var result = geometry(offset, base, fraction)
    var progress = clamp(result.progress, 0, 1)
    var height = cardHeight(progress)
    var mediaWidth = lerp(COMPACT_IMAGE_WIDTH, width, progress)
    var mediaHeight = lerp(COMPACT_HEIGHT, EXPANDED_IMAGE_HEIGHT, progress)
    var insetX = Math.max(0, (width - mediaWidth) / 2)
    var compactOpacity = 1 - smoothstep(0.15, 0.45, progress)
    var expandedOpacity = smoothstep(0.55, 0.85, progress)
    var hidden = result.y < -EXPANDED_HEIGHT || result.y > 470

    setStyle(select(ins, '.dl-card-' + poolIndex), {
      transform: 'translate3d(0,' + result.y.toFixed(2) + 'px,0)',
      height: height.toFixed(2) + 'px',
      visibility: hidden ? 'hidden' : 'visible',
      zIndex: progress > 0.5 ? 4 : 1,
      borderColor: progress > 0.5 ? '#dbe8f4' : '#EEF1F4',
      transition: cardTransition,
    })
    setStyle(select(ins, '.dl-media-' + poolIndex), {
      width: mediaWidth.toFixed(2) + 'px',
      height: mediaHeight.toFixed(2) + 'px',
      transition: mediaTransition,
    })
    setStyle(select(ins, '.dl-img-' + poolIndex), {
      width: width + 'px',
      transform: 'translate3d(' + (-insetX).toFixed(2) + 'px,0,0)',
      transition: imageTransition,
    })
    setStyle(select(ins, '.dl-compact-' + poolIndex), {
      opacity: compactOpacity.toFixed(3),
      transform: 'translate3d(' + lerp(0, -5, progress).toFixed(2) + 'px,0,0)',
      transition: copyTransition,
    })
    setStyle(select(ins, '.dl-expanded-' + poolIndex), {
      opacity: expandedOpacity.toFixed(3),
      transform: 'translate3d(0,' + lerp(6, 0, expandedOpacity).toFixed(2) + 'px,0)',
      transition: copyTransition,
    })
  }
}

function readY(event) {
  var touches = event && event.touches
  var touch = touches && touches[0]
  if (!touch) return null
  return touch.clientY || touch.pageY || 0
}

function touchstart(event, ins) {
  if (locked) return false
  var y = readY(event)
  if (y === null) return false
  owner = ins
  var dataset = event.currentTarget && event.currentTarget.dataset
  if (dataset && dataset.loopWidth) width = dataset.loopWidth
  if (dataset && dataset.loopDisabled) return false
  dragging = true
  moved = false
  startY = y
  position = 0
  velocity = 0
  lastPosition = 0
  lastAt = Date.now()
  renderAt(0, ins, 0)
  return false
}

function touchmove(event, ins) {
  if (!dragging || locked) return false
  var y = readY(event)
  if (y === null) return false
  var deltaY = y - startY
  moved = moved || Math.abs(deltaY) > TAP_LIMIT
  var raw = -deltaY / DRAG_RATIO
  if (Math.abs(raw) > 1) {
    raw = (raw < 0 ? -1 : 1) * (1 + Math.min(0.12, (Math.abs(raw) - 1) * 0.15))
  }
  var now = Date.now()
  var elapsed = Math.max(1, now - lastAt)
  var instant = (raw - lastPosition) / elapsed
  velocity = velocity * 0.55 + instant * 0.45
  lastPosition = raw
  lastAt = now
  renderAt(raw, ins, 0)
  return false
}

function settle(ins) {
  if (!dragging || locked) return false
  dragging = false
  var projected = position + velocity * 45
  var target = 0
  if (Math.abs(position) >= 0.18 || Math.abs(projected) >= 0.28) {
    target = projected <= 0 ? -1 : 1
  }
  var remaining = Math.abs(target - position)
  if (remaining < 0.001) {
    renderAt(0, ins, 0)
    locked = false
    return false
  }
  var duration = Math.round(clamp(176 + remaining * 38, 184, 224))
  locked = true
  settleTarget = target
  renderAt(target, ins, duration)
  return false
}

function transitionend(event, ins) {
  if (!locked) return false
  var target = settleTarget
  settleTarget = 0
  locked = false
  dragging = false
  position = target
  velocity = 0
  if (target === 0) {
    renderAt(0, ins, 0)
  } else if (ins && ins.callMethod) {
    ins.callMethod('commitLoopStep', { step: target })
  } else {
    renderAt(0, ins, 0)
  }
  return false
}

function touchend(event, ins) {
  return settle(ins)
}

function touchcancel(event, ins) {
  velocity = 0
  return settle(ins)
}

function onVersionChange(newValue, oldValue, ownerInstance, instance) {
  var ins = ownerInstance || instance || owner
  owner = ins
  renderAt(0, ins, 0)
  position = 0
  velocity = 0
  settleTarget = 0
  dragging = false
  locked = false
}

function cardTap(event, ins) {
  if (moved || locked) return false
  var dataset = event.currentTarget && event.currentTarget.dataset
  if (ins && ins.callMethod) {
    ins.callMethod('selectPoolItem', { poolIndex: dataset && dataset.poolIndex })
  }
  return false
}

function mediaTap(event, ins) {
  if (moved || locked) return false
  var dataset = event.currentTarget && event.currentTarget.dataset
  if (ins && ins.callMethod) {
    ins.callMethod('previewPoolItem', { poolIndex: dataset && dataset.poolIndex })
  }
  return false
}

module.exports = {
  touchstart: touchstart,
  touchmove: touchmove,
  touchend: touchend,
  touchcancel: touchcancel,
  transitionend: transitionend,
  onVersionChange: onVersionChange,
  cardTap: cardTap,
  mediaTap: mediaTap,
}
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
