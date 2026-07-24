<template>
  <div class="fruitbox">
    <div class="top-card" aria-label="status card">
      <div class="card-hd">
        <div class="card-title">
          <span class="fruit">🍍</span>
          <span>水果消消乐</span>
        </div>
        <div class="icons">
          <button class="icon" aria-label="设置" @click="settingsOpen = true">
            <van-icon name="setting-o" size="18" />
          </button>
        </div>
      </div>

      <!-- user requested to remove level name -->
    </div>

    <van-popup
      v-model:show="settingsOpen"
      class="pause-popup"
      position="center"
      round
      :overlay-style="{ background: 'rgba(120,120,120,0.28)' }"
      :style="{ background: 'transparent' }"
    >
      <div class="pause-panel" role="dialog" aria-label="暂停菜单">
        <div class="pause-actions">
          <button class="gbtn play" @click="continueGame" aria-label="继续">
            <van-icon name="play" size="26" />
          </button>
          <button class="gbtn restart" @click="requestRestart" aria-label="重新开始">
            <van-icon name="replay" size="24" />
          </button>
        </div>
      </div>
    </van-popup>

    <main class="main">
      <div class="board-shell">
        <div
          ref="boardEl"
          class="board"
          :style="boardStyle"
          @pointerdown.prevent="onPointerDown"
          @pointermove.prevent="onPointerMove"
          @pointerup.prevent="onPointerUp"
          @pointercancel.prevent="onPointerCancel"
        >
          <div
            v-for="(cell, idx) in board"
            :key="idx"
            class="cell"
            :class="{
              selected: selectedSet.has(idx),
              empty: cell.value == null,
              wrong: flashWrong && selectedSet.has(idx)
            }"
          >
            <FruitToken v-if="cell.value != null" :value="cell.value" />
          </div>

          <div v-if="dragging" class="selection" :style="selectionStyle" />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, watchEffect } from 'vue'
import { showSuccessToast } from 'vant'
import FruitToken from '@/fames/fruit_box/FruitToken.vue'

const props = defineProps({
  playerKey: { type: String, default: '' },
  seed: { type: [Number, String], default: null }
})

const emit = defineEmits(['restart'])

const ROWS = 8
const COLS = 8
const VALUE_MIN = 1
const VALUE_MAX = 10
const BEST_KEY = 'fruit_box_best_score'

const boardEl = ref(null)
const boardPx = ref({ width: 320, height: 320, cell: 40, gap: 0, pad: 0 })

function hashToSeed(str) {
  // FNV-1a 32-bit
  let h = 2166136261
  const s = String(str || '')
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  return h >>> 0
}

function makeRng(seed) {
  // Mulberry32
  let a = (typeof seed === 'number' ? seed : hashToSeed(seed)) >>> 0
  return function rng() {
    a = (a + 0x6d2b79f5) >>> 0
    let t = a
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

let rng = Math.random
function setRng(seed) {
  rng = seed == null || seed === '' ? Math.random : makeRng(Number(seed) >>> 0)
}

function randInt(min, max) {
  return Math.floor(rng() * (max - min + 1)) + min
}

function shuffleInPlace(arr) {
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(rng() * (i + 1))
    const t = arr[i]
    arr[i] = arr[j]
    arr[j] = t
  }
}

function generateSum10Composition(k) {
  // For 1x1 rectangles, we allow a single "10" tile (value=10) so 10 exists and is solvable (select it alone).
  if (k === 1) return [10]
  let remaining = 10
  const out = []
  for (let i = 0; i < k - 1; i++) {
    const minV = 1
    const maxV = Math.min(9, remaining - (k - 1 - i))
    const v = randInt(minV, maxV)
    out.push(v)
    remaining -= v
  }
  out.push(remaining)
  return out
}

function passesRectDifficulty(nums) {
  const ones = nums.filter((n) => n === 1).length
  const max = Math.max(...nums)
  if (nums.length === 5 && ones > 2) return false
  if (nums.length === 4 && ones > 1) return false
  if (nums.length === 3 && ones > 1) return false
  if (max < 5) return false
  return true
}

function tileRects(rows, cols) {
  const used = new Array(rows * cols).fill(false)
  const rects = []

  const MAX_SINGLE_TILES = 6

  const shapes = [
    // allow a few single tiles (become value=10), but cap them via MAX_SINGLE_TILES
    { h: 1, w: 1, a: 1, wgt: 1 },
    { h: 1, w: 3, a: 3, wgt: 3 },
    { h: 3, w: 1, a: 3, wgt: 3 },
    { h: 2, w: 2, a: 4, wgt: 6 },
    { h: 1, w: 4, a: 4, wgt: 1 },
    { h: 4, w: 1, a: 4, wgt: 1 },
    { h: 1, w: 5, a: 5, wgt: 1 },
    { h: 5, w: 1, a: 5, wgt: 1 }
  ]

  function pickFirstEmpty() {
    for (let i = 0; i < used.length; i++) if (!used[i]) return i
    return -1
  }

  function canPlace(r0, c0, h, w) {
    if (r0 + h > rows || c0 + w > cols) return false
    for (let r = r0; r < r0 + h; r++) {
      for (let c = c0; c < c0 + w; c++) {
        if (used[r * cols + c]) return false
      }
    }
    return true
  }

  function setPlace(r0, c0, h, w, val) {
    for (let r = r0; r < r0 + h; r++) {
      for (let c = c0; c < c0 + w; c++) {
        used[r * cols + c] = val
      }
    }
  }

  function expandWeighted(shps) {
    const out = []
    for (const s of shps) for (let i = 0; i < s.wgt; i++) out.push(s)
    shuffleInPlace(out)
    return out
  }

  const candidates = expandWeighted(shapes)

  function dfs(singleCount = 0) {
    const idx = pickFirstEmpty()
    if (idx === -1) return true
    const r0 = Math.floor(idx / cols)
    const c0 = idx % cols

    for (const s of candidates) {
      if (s.a === 1 && singleCount >= MAX_SINGLE_TILES) continue
      if (!canPlace(r0, c0, s.h, s.w)) continue
      setPlace(r0, c0, s.h, s.w, true)
      rects.push({ r0, c0, h: s.h, w: s.w, area: s.a })
      if (dfs(singleCount + (s.a === 1 ? 1 : 0))) return true
      rects.pop()
      setPlace(r0, c0, s.h, s.w, false)
    }
    return false
  }

  return dfs() ? rects : null
}

function generateSolvableBoard(seed) {
  setRng(seed)
  const rects = tileRects(ROWS, COLS)
  if (!rects) return null

  const out = new Array(ROWS * COLS).fill(null)
  for (const rect of rects) {
    const k = rect.area
    let nums = null
    for (let t = 0; t < 20; t++) {
      const cand = generateSum10Composition(k)
      if (passesRectDifficulty(cand)) {
        nums = cand
        break
      }
    }
    if (!nums) nums = generateSum10Composition(k)
    shuffleInPlace(nums)

    let p = 0
    for (let r = rect.r0; r < rect.r0 + rect.h; r++) {
      for (let c = rect.c0; c < rect.c0 + rect.w; c++) {
        const idx = r * COLS + c
        const v = nums[p++]
        out[idx] = {
          id: `c_${Date.now()}_${idx}_${Math.random()}`,
          value: v
        }
      }
    }
  }
  return out
}

function makeCell(id) {
  return {
    id,
    value: randInt(VALUE_MIN, VALUE_MAX)
  }
}

function makeBoard(seed) {
  setRng(seed)
  for (let attempt = 0; attempt < 15; attempt++) {
    const b = generateSolvableBoard(((Number(seed) >>> 0) + attempt * 2654435761) >>> 0)
    if (b) return b
  }
  const out = []
  for (let i = 0; i < ROWS * COLS; i++) out.push(makeCell(`c_${Date.now()}_${i}_${Math.random()}`))
  return out
}

const bestKey = computed(() => (props.playerKey ? `${BEST_KEY}:${props.playerKey}` : BEST_KEY))

const board = ref(makeBoard(props.seed))
const score = ref(0)
const bestScore = ref(0)
const settingsOpen = ref(false)

watchEffect(() => {
  bestScore.value = Number(localStorage.getItem(bestKey.value) || 0)
})

watch(score, (v) => {
  if (v > bestScore.value) {
    bestScore.value = v
    localStorage.setItem(bestKey.value, String(v))
  }
})

// seed 变化时也强制重开（防止“重开不刷新”）
watch(
  () => props.seed,
  () => {
    newGame()
  }
)

// NOTE: top card is intentionally minimal (no steps/tool/progress UI)

const dragging = ref(false)
const drag = ref({ startX: 0, startY: 0, x: 0, y: 0 })
const selectedSet = ref(new Set())
const flashWrong = ref(false)
let wrongTimer = null

let lastSelKey = ''
let rafId = 0
let pendingPoint = null

function clamp(n, min, max) {
  return Math.max(min, Math.min(max, n))
}

function localPointFromEvent(e) {
  const el = boardEl.value
  if (!el) return { x: 0, y: 0 }
  const r = el.getBoundingClientRect()
  const x = clamp(e.clientX - r.left, 0, r.width)
  const y = clamp(e.clientY - r.top, 0, r.height)
  return { x, y }
}

function rectFromDrag() {
  const x1 = drag.value.startX
  const y1 = drag.value.startY
  const x2 = drag.value.x
  const y2 = drag.value.y
  const left = Math.min(x1, x2)
  const top = Math.min(y1, y2)
  const right = Math.max(x1, x2)
  const bottom = Math.max(y1, y2)
  return { left, top, right, bottom, width: right - left, height: bottom - top }
}

function computeSelection(rect) {
  const { cell, gap, pad } = boardPx.value
  const indices = []
  let sum = 0
  for (let idx = 0; idx < ROWS * COLS; idx++) {
    const row = Math.floor(idx / COLS)
    const col = idx % COLS
    // IMPORTANT: board has padding + gap; selection math must match visual grid
    const cx = pad + col * (cell + gap) + cell / 2
    const cy = pad + row * (cell + gap) + cell / 2
    const inside = rect.left <= cx && rect.top <= cy && rect.right >= cx && rect.bottom >= cy
    const v = board.value[idx].value
    if (inside && v != null) {
      indices.push(idx)
      sum += v
    }
  }
  const key = indices.join('|')
  if (key === lastSelKey) return { count: selectionCount.value, sum: selectionSumValue.value }
  lastSelKey = key
  selectedSet.value = new Set(indices)
  return { count: indices.length, sum }
}

const selectionCount = ref(0)
const selectionSumValue = ref(0)

function clearWrongFlash() {
  flashWrong.value = false
  if (wrongTimer) {
    clearTimeout(wrongTimer)
    wrongTimer = null
  }
}

function triggerWrongFlash() {
  clearWrongFlash()
  flashWrong.value = true
  wrongTimer = setTimeout(() => {
    flashWrong.value = false
    wrongTimer = null
  }, 220)
}

function commitSelectionIfValid() {
  const count = selectionCount.value
  if (!count) return
  const sum = selectionSumValue.value
  if (sum !== 10) {
    triggerWrongFlash()
    return
  }
  const indices = Array.from(selectedSet.value)
  score.value += count
  for (const idx of indices) {
    board.value[idx] = { ...board.value[idx], value: null }
  }
  selectedSet.value = new Set()
  selectionCount.value = 0
  selectionSumValue.value = 0
  lastSelKey = ''
  if (board.value.every((c) => c.value == null)) showSuccessToast('已全部消除！')
}

function onPointerDown(e) {
  if (e.button != null && e.button !== 0) return
  clearWrongFlash()
  const el = boardEl.value
  if (!el) return
  el.setPointerCapture?.(e.pointerId)
  const p = localPointFromEvent(e)
  dragging.value = true
  drag.value = { startX: p.x, startY: p.y, x: p.x, y: p.y }
  const r = computeSelection(rectFromDrag())
  selectionCount.value = r?.count || 0
  selectionSumValue.value = r?.sum || 0
  window.addEventListener('pointerup', onGlobalPointerUp, { once: true })
  window.addEventListener('pointercancel', onGlobalPointerCancel, { once: true })
  window.addEventListener('blur', onGlobalBlur, { once: true })
}

function onPointerMove(e) {
  if (!dragging.value) return
  pendingPoint = localPointFromEvent(e)
  if (rafId) return
  rafId = requestAnimationFrame(() => {
    rafId = 0
    if (!dragging.value || !pendingPoint) return
    const p = pendingPoint
    pendingPoint = null
    drag.value = { ...drag.value, x: p.x, y: p.y }
    const r = computeSelection(rectFromDrag())
    selectionCount.value = r?.count || 0
    selectionSumValue.value = r?.sum || 0
  })
}

function finishDrag() {
  if (!dragging.value) return
  dragging.value = false
  commitSelectionIfValid()
}

function onPointerUp() {
  finishDrag()
}

function onPointerCancel() {
  dragging.value = false
  selectedSet.value = new Set()
  selectionCount.value = 0
  selectionSumValue.value = 0
  lastSelKey = ''
}

function onGlobalPointerUp() {
  finishDrag()
}
function onGlobalPointerCancel() {
  onPointerCancel()
}
function onGlobalBlur() {
  onPointerCancel()
}

function newGame() {
  clearWrongFlash()
  score.value = 0
  selectedSet.value = new Set()
  selectionCount.value = 0
  selectionSumValue.value = 0
  lastSelKey = ''
  board.value = makeBoard(props.seed)
  nextTick(updateBoardMetrics)
}

function continueGame() {
  settingsOpen.value = false
}

function requestRestart() {
  settingsOpen.value = false
  // Immediately refresh board locally (guaranteed), and also notify route-level restart
  newGame()
  emit('restart')
}

const selectionStyle = computed(() => {
  const r = rectFromDrag()
  return {
    transform: `translate(${r.left}px, ${r.top}px)`,
    width: `${r.width}px`,
    height: `${r.height}px`
  }
})

const boardStyle = computed(() => ({
  gridTemplateColumns: `repeat(${COLS}, 1fr)`
}))

function updateBoardMetrics() {
  const el = boardEl.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const size = Math.min(r.width, r.height)
  const cs = window.getComputedStyle(el)
  const pad = parseFloat(cs.paddingLeft || '0') || 0
  // Vite/Vue: CSS gap shows up in columnGap/rowGap; fallback to gap
  const gap = parseFloat(cs.columnGap || cs.gap || '0') || 0
  const inner = Math.max(0, size - pad * 2 - gap * (COLS - 1))
  const cell = inner / COLS
  boardPx.value = { width: r.width, height: r.height, cell, gap, pad }
}

let ro = null
onMounted(() => {
  nextTick(updateBoardMetrics)
  if (window.ResizeObserver && boardEl.value) {
    ro = new ResizeObserver(() => updateBoardMetrics())
    ro.observe(boardEl.value)
  }
})

onBeforeUnmount(() => {
  clearWrongFlash()
  if (rafId) cancelAnimationFrame(rafId)
  if (ro) ro.disconnect()
})
</script>

<style scoped>
.fruitbox {
  --text: rgba(55, 52, 55, 0.92);
  --muted: rgba(55, 52, 55, 0.58);
  min-height: 100vh;
  width: 100%;
  background: radial-gradient(1200px 900px at 20% 10%, #fff2f6 0%, #f8f4ff 35%, #f6fffb 70%, #ffffff 100%);
  color: var(--text);
  padding: 18px 14px 24px;
  touch-action: manipulation;
  position: relative;
  overflow: hidden;
  overscroll-behavior: contain;
}

.top-card {
  width: 100%;
  max-width: 560px;
  margin: 0 auto 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(17, 24, 39, 0.06);
  box-shadow: 0 18px 50px rgba(17, 24, 39, 0.10);
  padding: 14px;
}
.card-hd {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.card-title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-weight: 950;
  font-size: 18px;
  color: var(--text);
}
.fruit {
  font-size: 22px;
}
.icons .icon {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  border: 1px solid rgba(17, 24, 39, 0.06);
  background: rgba(255, 255, 255, 0.88);
  display: grid;
  place-items: center;
  color: rgba(55, 52, 55, 0.85);
}
.card-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: rgba(55, 52, 55, 0.62);
}
/* Level removed */
.card-subtitle {
  display: none;
}

/* Pause popup (glass) */
:deep(.pause-popup.van-popup) {
  background: transparent !important;
  box-shadow: none !important;
}
.pause-panel {
  width: min(360px, calc(100vw - 36px));
  padding: 14px 14px 18px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.45);
  border: 1px solid rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  box-shadow: 0 28px 70px rgba(17, 24, 39, 0.16);
}
.pause-actions {
  display: flex;
  justify-content: center;
  gap: 14px;
}
.gbtn {
  width: 64px;
  height: 64px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.65);
  background: rgba(255, 255, 255, 0.35);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: 0 16px 40px rgba(17, 24, 39, 0.12);
  color: rgba(55, 52, 55, 0.92);
  display: grid;
  place-items: center;
  grid-template-rows: 1fr;
  transition: transform 0.12s ease, background 0.12s ease;
}
.gbtn:active {
  transform: scale(0.96);
}
.gbtn.play {
  background: rgba(245, 138, 163, 0.22);
}
.gbtn.restart {
  background: rgba(185, 167, 200, 0.18);
}
/* icon-only: no label */
/* top-card UI is minimal now: no progress/steps/tool blocks */

.main {
  width: 100%;
  max-width: 560px;
  margin: 0 auto;
}

.board-shell {
  border-radius: 16px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(17, 24, 39, 0.06);
  box-shadow: 0 18px 50px rgba(17, 24, 39, 0.10);
}

.board {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  display: grid;
  user-select: none;
  touch-action: none;
  background: rgba(255, 255, 255, 0.0);
  border-radius: 14px;
  overflow: hidden;
  overscroll-behavior: contain;
  gap: 6px; /* 更小间距：更饱满 */
  padding: 6px;
}

.cell {
  position: relative;
  display: grid;
  place-items: center;
}

.cell.selected :deep(.token) {
  outline: 3px solid rgba(255, 255, 255, 0.95);
  outline-offset: 2px;
}
.cell.wrong :deep(.token) {
  outline: 3px solid rgba(244, 63, 94, 0.85);
  outline-offset: 2px;
}

.selection {
  position: absolute;
  border: 2px solid rgba(185, 167, 200, 0.65);
  background: rgba(245, 138, 163, 0.10);
  border-radius: 10px;
  pointer-events: none;
  box-shadow: 0 0 0 1px rgba(17, 24, 39, 0.06) inset;
  will-change: transform, width, height;
}

@media (max-width: 420px) {
  .board-shell {
    padding: 8px;
  }
  .board {
    gap: 5px;
    padding: 5px;
  }
}
</style>


