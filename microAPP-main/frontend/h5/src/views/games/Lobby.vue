<template>
  <div class="lobby">
    <div class="bg" aria-hidden="true">
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
            <div class="t2">框选水果，凑满 10 消除，清空全盘通关</div>
          </div>
        </div>

        <div class="how">
          <div class="how-item"><b>操作</b>：拖拽框选（松手结算）</div>
          <div class="how-item"><b>规则</b>：选中数字之和必须 = 10</div>
          <div class="how-item"><b>目标</b>：全部消除（不掉落、不补位）</div>
        </div>
      </div>

      <div class="dock" role="group" aria-label="Game controls">
        <button class="icon-btn play" :disabled="loading" @click="start" aria-label="开始游戏">
          <van-icon name="play" size="26" />
        </button>
        <button class="icon-btn restart" :disabled="loading" @click="restart" aria-label="重新开始">
          <van-icon name="replay" size="24" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const loading = ref(true)
const session = ref(0)

// We still assign by IP (seed), but we don't show it in UI.
const baseSeed = ref(null)

async function loadAssign() {
  loading.value = true
  try {
    const res = await axios.get('/api/games/assign')
    const data = res?.data
    if (data?.success) {
      baseSeed.value = typeof data.seed === 'number' ? data.seed : Number(data.seed)
    }
  } catch (e) {
    baseSeed.value = null
  } finally {
    loading.value = false
  }
}

function ensureSeed() {
  // If backend assign isn't available, still generate a seed so restart/start can roll a new board.
  const n = Number(baseSeed.value)
  if (Number.isFinite(n)) return n >>> 0
  const fallback = (Date.now() >>> 0) ^ ((Math.random() * 0xffffffff) >>> 0)
  baseSeed.value = fallback
  return fallback >>> 0
}

function start() {
  // always bump session so "start" also refreshes if user returns to lobby
  session.value += 1
  router.push({
  path: '/games/play',
  query: { s: session.value, seed: ensureSeed(), roll: Date.now(), t: Date.now() }
})
}

function restart() {
  // always create a new session so it definitely refreshes
  session.value += 1
  router.push({
  path: '/games/play',
  query: { s: session.value, seed: ensureSeed(), roll: Date.now(), t: Date.now() }
})
}

onMounted(loadAssign)
</script>

<style scoped>
.lobby {
  width: 100%;
  min-height: 100vh;
  background:
    radial-gradient(900px 700px at 18% 12%, rgba(215, 154, 162, 0.26) 0%, rgba(215, 154, 162, 0.06) 55%, rgba(0, 0, 0, 0) 100%),
    radial-gradient(900px 700px at 82% 18%, rgba(185, 167, 200, 0.22) 0%, rgba(185, 167, 200, 0.06) 55%, rgba(0, 0, 0, 0) 100%),
    radial-gradient(900px 700px at 40% 92%, rgba(169, 196, 178, 0.18) 0%, rgba(169, 196, 178, 0.05) 55%, rgba(0, 0, 0, 0) 100%),
    linear-gradient(180deg, #fff7fb 0%, #f7f2ff 45%, #f2fbf7 100%);
  overflow: hidden;
  position: relative;
}

.bg {
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
  opacity: 0; /* no floating emojis in Morandi style */
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
  min-height: 100vh;
  display: grid;
  place-items: center;
  gap: 18px;
}

.hero {
  width: 100%;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.76);
  border: 1px solid rgba(17, 24, 39, 0.06);
  box-shadow: 0 18px 50px rgba(17, 24, 39, 0.10);
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
  font-size: 20px;
  font-weight: 950;
  color: rgba(60, 55, 60, 0.92);
  letter-spacing: 0.2px;
}
.titles .t2 {
  margin-top: 4px;
  font-size: 12px;
  color: rgba(60, 55, 60, 0.62);
  line-height: 1.35;
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
  color: rgba(60, 55, 60, 0.72);
}
.how-item b {
  color: rgba(60, 55, 60, 0.92);
}

.dock {
  position: fixed;
  left: 50%;
  bottom: 18px;
  transform: translateX(-50%);
  width: min(560px, calc(100% - 24px));
  display: flex;
  justify-content: center;
  gap: 14px;
  padding: 10px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.76);
  border: 1px solid rgba(17, 24, 39, 0.06);
  box-shadow: 0 18px 50px rgba(17, 24, 39, 0.10);
  backdrop-filter: blur(10px);
}

.icon-btn {
  width: 52px;
  height: 52px;
  border-radius: 999px;
  border: none;
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: transform 0.12s ease, box-shadow 0.12s ease, opacity 0.12s ease;
}
.icon-btn:active {
  transform: scale(0.96);
}
.icon-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.icon-btn.play {
  background: linear-gradient(135deg, #d79aa2 0%, #e6b2a8 100%);
  color: #fff;
  box-shadow: 0 14px 30px rgba(215, 154, 162, 0.18);
}
.icon-btn.restart {
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(17, 24, 39, 0.08);
  color: rgba(60, 55, 60, 0.92);
}
</style>


