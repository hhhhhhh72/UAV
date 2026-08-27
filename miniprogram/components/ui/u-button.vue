<template>
  <button
    class="u-button"
    :class="[`u-button--${type}`, `u-button--${size}`, { 'u-button--block': block, 'u-button--round': round, 'u-button--disabled': disabled || loading }]"
    :disabled="disabled || loading"
    @click="onClick"
  >
    <u-loading v-if="loading" size="28rpx" color="#fff" />
    <text class="u-button-text" v-else><slot /></text>
  </button>
</template>

<script setup>
defineProps({
  type: { type: String, default: 'primary' },
  size: { type: String, default: 'normal' },
  block: { type: Boolean, default: false },
  round: { type: Boolean, default: true },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false }
})
const emit = defineEmits(['click'])
function onClick(e) {
  emit('click', e)
}
</script>

<style scoped>
.u-button { display: inline-flex; align-items: center; justify-content: center; gap: 8rpx; padding: 0 32rpx; height: 72rpx; font-size: 30rpx; border: none; border-radius: 16rpx; box-sizing: border-box; }
/* 重置微信 button 默认伪元素发丝边框 */
.u-button::after { border: none; }
.u-button--block { width: 100%; }
.u-button--round { border-radius: 50rpx; }
.u-button--large { height: 88rpx; font-size: 32rpx; }
.u-button--small { height: 56rpx; padding: 0 24rpx; font-size: 26rpx; }
.u-button--primary { background: var(--color-primary, #0A66C2); color: #fff; }
.u-button--default { background: var(--color-primary-light, #E8F2FC); color: var(--color-primary, #0A66C2); }
.u-button--danger { background: var(--color-danger, #ff3b30); color: #fff; }
.u-button--success { background: var(--color-success, #34c759); color: #fff; }
.u-button--disabled { opacity: 0.5; }
.u-button-text { line-height: 1; }
/* 微信 button[disabled] 默认灰底会覆盖类型色，此处强制保留禁用态样式 */
.u-button[disabled] { background: #c8c9cc; color: #fff; opacity: 1; }
</style>
