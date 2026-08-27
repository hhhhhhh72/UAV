<template>
  <view v-if="show" class="u-popup" @click="onOverlayClick">
    <view class="u-popup-mask" />
    <view class="u-popup-panel" :class="[`u-popup--${position}`, { 'u-popup--round': round }]" @click.stop>
      <view v-if="showClose" class="u-popup-close" @click="onClose"><u-icon name="close" size="28rpx" color="#969799" /></view>
      <slot />
    </view>
  </view>
</template>

<script setup>
const props = defineProps({
  show: { type: Boolean, default: false },
  position: { type: String, default: 'bottom' },
  round: { type: Boolean, default: false },
  closeOnClickOverlay: { type: Boolean, default: true },
  showClose: { type: Boolean, default: false }
})
const emit = defineEmits(['close', 'update:show'])
function onClose() {
  emit('close')
  emit('update:show', false)
}
function onOverlayClick() {
  if (props.closeOnClickOverlay) onClose()
}
</script>

<style scoped>
.u-popup { position: fixed; top: 0; right: 0; bottom: 0; left: 0; z-index: 1000; }
.u-popup-mask { position: absolute; top: 0; right: 0; bottom: 0; left: 0; background: rgba(0,0,0,0.5); }
.u-popup-panel { position: absolute; background: #fff; }
.u-popup--bottom { left: 0; right: 0; bottom: 0; padding-bottom: env(safe-area-inset-bottom); }
.u-popup--top { left: 0; right: 0; top: 0; }
.u-popup--center { left: 50%; top: 50%; transform: translate(-50%, -50%); border-radius: 24rpx; min-width: 600rpx; }
.u-popup--round { border-radius: 24rpx 24rpx 0 0; }
.u-popup-close { position: absolute; top: 20rpx; right: 24rpx; z-index: 2; }
</style>
