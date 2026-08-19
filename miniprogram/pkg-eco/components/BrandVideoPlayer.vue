<template>
  <view v-if="show" class="vplayer" @tap.stop>
    <!-- 关闭 -->
    <view class="vclose" @tap="$emit('close')"><text>✕</text></view>
    <text class="vtit">{{ title }}</text>

    <!-- ===== 接口替换点 =====
         真实播放：替换为 <video :src="videoUrl" :controls="true" autoplay />，
         videoUrl 来自详情接口 video_list[].url（后端返回视频地址）
    ===== -->
    <view class="vbig" :class="{ playing }" @tap="togglePlay">
      <text>{{ playing ? '❚❚' : (finished ? '↻' : '▶') }}</text>
    </view>

    <view class="vbar"><view class="vbar-i" :style="{ width: progress + '%' }"></view></view>
    <text class="vmeta">{{ curTime }} / {{ duration }}</text>
  </view>
</template>

<script setup>
// 品牌宣传视频 · 全屏播放浮层（原型交互模拟态）
import { ref, watch, onUnmounted } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  title: { type: String, default: '' },
  duration: { type: String, default: '03:00' }, // 展示用时长（真实数据以接口秒数换算）
})
const emit = defineEmits(['close'])

const playing = ref(false)
const finished = ref(false)
const progress = ref(0)
const curTime = ref('00:00')
let timer = null

const durationSec = 180 // 模拟总时长（秒）；接口替换点：video.duration

watch(
  () => props.show,
  (v) => {
    if (v) reset()
  }
)

function reset() {
  playing.value = false
  finished.value = false
  progress.value = 0
  curTime.value = '00:00'
  stopTimer()
  // 自动开始播放（模拟）
  setTimeout(() => togglePlay(), 350)
}

function togglePlay() {
  if (finished.value) { finished.value = false; progress.value = 0; curTime.value = '00:00' }
  playing.value = !playing.value
  if (playing.value) {
    timer = setInterval(() => {
      progress.value += 1.8
      if (progress.value >= 100) {
        progress.value = 100
        playing.value = false
        finished.value = true
        curTime.value = props.duration
        stopTimer()
        return
      }
      const t = Math.floor((progress.value / 100) * durationSec)
      curTime.value = pad(Math.floor(t / 60)) + ':' + pad(t % 60)
    }, 200)
  } else {
    stopTimer()
  }
}

function stopTimer() {
  if (timer) { clearInterval(timer); timer = null }
}

const pad = (n) => (n < 10 ? '0' + n : '' + n)

onUnmounted(stopTimer)
</script>

<style scoped>
.vplayer {
  position: fixed; inset: 0; z-index: 999;
  background: #000;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  animation: fadeIn .35s ease both;
}
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.vclose {
  position: absolute; top: calc(var(--status-bar, 20px) + 40px); right: 32rpx;
  width: 72rpx; height: 72rpx; border-radius: 50%;
  background: rgba(255,255,255,.15); color: #fff;
  display: flex; align-items: center; justify-content: center; font-size: 30rpx;
}
.vclose:active { background: rgba(255,255,255,.3); }
.vtit {
  position: absolute; top: calc(var(--status-bar, 20px) + 44px);
  left: 32rpx; right: 32rpx;
  color: #fff; font-size: 28rpx; font-weight: 600; text-align: center;
}
.vbig {
  width: 148rpx; height: 148rpx; border-radius: 50%;
  background: rgba(255,255,255,.16); border: 2px solid rgba(255,255,255,.7);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 52rpx;
  transition: transform .25s cubic-bezier(0.16, 1, 0.3, 1);
}
.vbig:active { transform: scale(.94); }
.vbig.playing { animation: spin 1.2s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.vbar {
  position: absolute; left: 40rpx; right: 40rpx; bottom: 180rpx;
  height: 6rpx; background: rgba(255,255,255,.25); border-radius: 3rpx; overflow: hidden;
}
.vbar-i { display: block; height: 100%; background: #fff; border-radius: 3rpx; }
.vmeta { position: absolute; bottom: 120rpx; color: rgba(255,255,255,.6); font-size: 20rpx; }
</style>
