<template>
  <div class="games-page">
    <div class="bg-fruit" aria-hidden="true">
      <div class="blob b1"></div>
      <div class="blob b2"></div>
      <div class="blob b3"></div>
      <div class="float f1">🍓</div>
      <div class="float f2">🍋</div>
      <div class="float f3">🍇</div>
      <div class="float f4">🍑</div>
    </div>
    <div class="shell">
      <div class="hero">
        <div class="brand">
          <div class="logo" aria-hidden="true">
            <svg viewBox="0 0 64 64" width="42" height="42">
              <defs>
                <linearGradient id="g1" x1="0" y1="0" x2="1" y2="1">
                  <stop offset="0" stop-color="#ff4d6d" />
                  <stop offset="1" stop-color="#ffb703" />
                </linearGradient>
                <linearGradient id="g2" x1="0" y1="0" x2="1" y2="1">
                  <stop offset="0" stop-color="#2dd4bf" />
                  <stop offset="1" stop-color="#60a5fa" />
                </linearGradient>
              </defs>
              <circle cx="32" cy="32" r="28" fill="url(#g2)" opacity="0.18" />
              <path
                d="M20 30c0-9 5-16 12-16s12 7 12 16c0 11-7 22-12 22S20 41 20 30Z"
                fill="url(#g1)"
                stroke="rgba(0,0,0,0.18)"
                stroke-width="1.2"
              />
              <path
                d="M32 10c6 0 10 3 12 7-4-1-8 0-12 3-4-3-8-4-12-3 2-4 6-7 12-7Z"
                fill="#22c55e"
                opacity="0.9"
              />
              <circle cx="26" cy="28" r="4" fill="rgba(255,255,255,0.55)" />
            </svg>
          </div>
          <div class="titles">
            <div class="t1">Fruit Box</div>
            <div class="t2">框选水果，凑满 10 分消除，清空全盘通关</div>
          </div>
        </div>

        <div class="info">
          <div class="pill">
            <span class="k">Player</span>
            <span class="v">{{ loading ? 'Detecting…' : playerKey }}</span>
          </div>
          <div class="pill">
            <span class="k">IP</span>
            <span class="v">{{ loading ? '...' : (ip || 'unknown') }}</span>
          </div>
          <div class="pill">
            <span class="k">Seed</span>
            <span class="v">{{ effectiveSeed ?? '-' }}</span>
          </div>
          <div class="pill">
            <span class="k">Bucket</span>
            <span class="v">{{ bucket ?? '-' }}</span>
          </div>
        </div>

        <div class="cta">
          <van-button class="btn primary" size="large" type="primary" :disabled="loading" @click="start">
            开始游戏
          </van-button>
          <van-button class="btn ghost" size="large" type="default" :disabled="loading" @click="restart">
            重新开始
          </van-button>
        </div>

        <div class="how">
          <div class="how-item"><b>操作</b>：拖拽框选（松手结算）</div>
          <div class="how-item"><b>规则</b>：选中数字之和必须 = 10</div>
          <div class="how-item"><b>目标</b>：全部消除（不掉落、不补位）</div>
        </div>
      </div>

      <div v-if="started" class="game-wrap">
        <FruitBoxGame
          :key="gameKey"
          :player-key="playerKey"
          :seed="effectiveSeed"
        />
      </div>
      <div v-else class="empty">
        <div class="empty-inner">
          <div class="h">Ready?</div>
          <div class="p">点击上方「开始游戏」进入水果盒挑战。</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import FruitBoxGame from '@/fames/fruit_box/FruitBoxGame.vue'
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'

const loading = ref(true)
const started = ref(false)

const ip = ref('')
const playerKey = ref('unknown')
const baseSeed = ref(null)
const bucket = ref(null)

const session = ref(0)
const gameKey = computed(() => `${playerKey.value}:${session.value}`)
const effectiveSeed = computed(() => {
  if (baseSeed.value == null) return null
  // derive a new deterministic seed per restart session (still "based on IP")
  return (Number(baseSeed.value) + session.value * 1013904223) >>> 0
})

async function loadAssign() {
  loading.value = true
  try {
    const res = await axios.get('/api/games/assign')
    const data = res?.data
    if (data?.success) {
      ip.value = data.ip || ''
      playerKey.value = data.playerKey || data.ip || 'unknown'
      baseSeed.value = typeof data.seed === 'number' ? data.seed : Number(data.seed)
      bucket.value = data.bucket ?? null
      return
    }
    throw new Error('assign failed')
  } catch (e) {
    // Fallback: no backend/proxy available, still allow play
    ip.value = ''
    playerKey.value = 'local'
    baseSeed.value = null
    bucket.value = null
  } finally {
    loading.value = false
  }
}

function start() {
  started.value = true
}

function restart() {
  session.value += 1
  started.value = true
}

onMounted(async () => {
  await loadAssign()
})
</script>

<style scoped>
.games-page {
  width: 100%;
  min-height: 100vh;
  background: radial-gradient(1200px 900px at 20% 10%, #fff0f6 0%, #f3f7ff 35%, #f7fff9 70%, #ffffff 100%);
  overflow: hidden;
  position: relative;
}

.bg-fruit {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}
.blob {
  position: absolute;
  filter: blur(26px);
  opacity: 0.55;
  border-radius: 999px;
}
.b1 {
  width: 320px;
  height: 320px;
  left: -90px;
  top: -70px;
  background: radial-gradient(circle at 30% 30%, rgba(255, 77, 109, 0.9), rgba(255, 183, 3, 0.1));
}
.b2 {
  width: 380px;
  height: 380px;
  right: -120px;
  top: 30px;
  background: radial-gradient(circle at 30% 30%, rgba(59, 130, 246, 0.55), rgba(45, 212, 191, 0.1));
}
.b3 {
  width: 420px;
  height: 420px;
  left: 20%;
  bottom: -180px;
  background: radial-gradient(circle at 30% 30%, rgba(34, 197, 94, 0.35), rgba(255, 183, 3, 0.12));
}
.float {
  position: absolute;
  font-size: 28px;
  opacity: 0.55;
  transform: translateZ(0);
  animation: floaty 7s ease-in-out infinite;
}
.f1 {
  left: 16px;
  top: 130px;
}
.f2 {
  right: 18px;
  top: 170px;
  animation-delay: -1.2s;
}
.f3 {
  left: 14%;
  bottom: 110px;
  animation-delay: -2.4s;
}
.f4 {
  right: 10%;
  bottom: 140px;
  animation-delay: -3.1s;
}
@keyframes floaty {
  0%,
  100% {
    transform: translateY(0) rotate(-2deg);
  }
  50% {
    transform: translateY(-14px) rotate(2deg);
  }
}

.shell {
  width: 100%;
  max-width: 560px;
  margin: 0 auto;
  padding: 14px 12px 18px;
  position: relative;
  z-index: 1;
}

.hero {
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 18px 50px rgba(17, 24, 39, 0.12);
  padding: 14px;
  backdrop-filter: blur(10px);
}

.brand {
  display: flex;
  gap: 12px;
  align-items: center;
}
.logo {
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(0, 0, 0, 0.06);
}
.titles .t1 {
  font-size: 18px;
  font-weight: 950;
  color: #111827;
  letter-spacing: 0.2px;
}
.titles .t2 {
  margin-top: 4px;
  font-size: 12px;
  color: #6b7280;
  line-height: 1.35;
}

.info {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.pill {
  display: inline-flex;
  gap: 8px;
  align-items: baseline;
  padding: 7px 10px;
  border-radius: 999px;
  background: rgba(17, 24, 39, 0.04);
  border: 1px solid rgba(17, 24, 39, 0.06);
}
.pill .k {
  font-size: 11px;
  color: #6b7280;
  font-weight: 800;
}
.pill .v {
  font-size: 12px;
  color: #111827;
  font-weight: 900;
}

.cta {
  margin-top: 12px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.btn :deep(.van-button__text) {
  font-weight: 900;
}
.btn.primary {
  background: linear-gradient(135deg, #ff4d6d 0%, #ffb703 100%);
  border: none;
  box-shadow: 0 14px 30px rgba(255, 77, 109, 0.25);
}
.btn.ghost {
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(17, 24, 39, 0.10);
}

.how {
  margin-top: 12px;
  display: grid;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(0, 0, 0, 0.05);
}
.how-item {
  font-size: 12px;
  color: #374151;
}
.how-item b {
  color: #111827;
}

.game-wrap {
  margin-top: 12px;
}

.empty {
  margin-top: 12px;
  border-radius: 14px;
  border: 1px dashed rgba(0, 0, 0, 0.15);
  background: rgba(255, 255, 255, 0.75);
  padding: 18px;
}
.empty-inner .h {
  font-size: 16px;
  font-weight: 900;
  color: #111827;
}
.empty-inner .p {
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

@media (max-width: 420px) {
  .cta {
    grid-template-columns: 1fr;
  }
}
</style>


