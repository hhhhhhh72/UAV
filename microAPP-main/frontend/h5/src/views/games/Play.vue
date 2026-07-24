<template>
  <div class="play">
    <FruitBoxGame
      :key="gameKey"
      :player-key="playerKey"
      :seed="effectiveSeed"
      @restart="restart"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import FruitBoxGame from '@/fames/fruit_box/FruitBoxGame.vue'

const route = useRoute()
const router = useRouter()

const playerKey = ref('local')
const baseSeed = ref(null)

const session = ref(Number(route.query.s || 0))

watch(
  () => route.query.s,
  (v) => {
    session.value = Number(v || 0)
  }
)

const gameKey = computed(() => `${playerKey.value}:${session.value}:${route.query.t || ''}`)

const roll = computed(() => {
  const n = Number(route.query.roll || 0)
  return Number.isFinite(n) ? (n >>> 0) : 0
})

const effectiveSeed = computed(() => {
  const b = Number(baseSeed.value ?? route.query.seed)
  if (!Number.isFinite(b)) return null
  return (b + session.value * 1013904223 + roll.value) >>> 0
})

async function loadAssign() {
  try {
    const res = await axios.get('/api/games/assign')
    const data = res?.data
    if (data?.success) {
      playerKey.value = data.playerKey || data.ip || 'unknown'
      baseSeed.value = typeof data.seed === 'number' ? data.seed : Number(data.seed)
      // persist seed into URL so refresh/restart works even if assign later fails
      if (!route.query.seed || route.query.seed === '') {
        router.replace({ path: '/games/play', query: { ...route.query, seed: baseSeed.value, t: route.query.t || Date.now() } })
      }
      return
    }
    throw new Error('assign failed')
  } catch (e) {
    playerKey.value = 'local'
    const q = Number(route.query.seed)
    baseSeed.value = Number.isFinite(q) ? (q >>> 0) : ((Date.now() >>> 0) ^ ((Math.random() * 0xffffffff) >>> 0))
    if (!route.query.seed || route.query.seed === '') {
      router.replace({ path: '/games/play', query: { ...route.query, seed: baseSeed.value, roll: route.query.roll || Date.now(), t: route.query.t || Date.now() } })
    }
  }
}

function restart() {
  const next = session.value + 1
  router.replace({ path: '/games/play', query: { ...route.query, s: next, roll: Date.now(), t: Date.now() } })
}

function back() {
  router.push('/games')
}

onMounted(loadAssign)
</script>

<style scoped>
.play {
  width: 100%;
  min-height: 100vh;
  background: transparent;
}
</style>


