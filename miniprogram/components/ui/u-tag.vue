<template>
  <text class="u-tag" :class="[`u-tag--${type}`, { 'u-tag--plain': plain, 'u-tag--mini': size === 'mini', 'u-tag--round': round }]">
    <slot />
  </text>
</template>

<script setup>
defineProps({
  type: { type: String, default: 'default' },
  plain: { type: Boolean, default: false },
  size: { type: String, default: 'normal' },
  round: { type: Boolean, default: true }
})
</script>

<style scoped>
.u-tag { display: inline-flex; align-items: center; padding: 4rpx 16rpx; font-size: var(--ui-font-size-sm, 26rpx); border-radius: 8rpx; line-height: 1.6; }
.u-tag--round { border-radius: 999rpx; }
.u-tag--mini { padding: 2rpx 12rpx; font-size: 22rpx; }
.u-tag--primary { background: var(--color-primary, #0A66C2); color: #fff; }
.u-tag--primary.u-tag--plain { background: var(--color-primary-light, #E8F2FC); color: var(--color-primary, #0A66C2); }
.u-tag--success { background: var(--color-success, #34c759); color: #fff; }
.u-tag--success.u-tag--plain { background: #E6FAF5; color: var(--color-success, #34c759); }
.u-tag--danger { background: var(--color-danger, #ff3b30); color: #fff; }
.u-tag--danger.u-tag--plain { background: #FFECEB; color: var(--color-danger, #ff3b30); }
.u-tag--warning { background: var(--color-warning, #ff9f0a); color: #fff; }
.u-tag--warning.u-tag--plain { background: #FFF4E6; color: var(--color-warning, #ff9f0a); }
.u-tag--default { background: var(--color-primary-light, #E8F2FC); color: var(--color-primary, #0A66C2); }
</style>
