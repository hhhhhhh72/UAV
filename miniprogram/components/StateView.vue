<template>
  <view>
    <!-- 加载中 -->
    <view v-if="loading" class="state-loading flex-center flex-col" style="padding: 80rpx 0;">
      <u-loading size="48rpx" color="var(--color-primary)" />
      <text class="text-secondary" style="margin-top: 16rpx;">{{ loadingText }}</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="error" class="state-error flex-center flex-col" style="padding: 80rpx 0;">
      <u-empty :description="errorText" />
      <view v-if="showRetry" class="state-retry">
        <u-button type="primary" size="small" round @click="$emit('retry')">重新加载</u-button>
      </view>
    </view>

    <!-- 空数据 -->
    <view v-else-if="empty" class="state-empty flex-center flex-col" style="padding: 80rpx 0;">
      <u-empty :description="emptyText" />
      <view v-if="showCreate" class="state-action">
        <u-button type="primary" size="small" round @click="$emit('create')">
          {{ createText }}
        </u-button>
      </view>
    </view>

    <!-- 正常内容 -->
    <slot v-else />
  </view>
</template>

<script setup>
defineProps({
  loading: { type: Boolean, default: false },
  error: { type: Boolean, default: false },
  empty: { type: Boolean, default: false },
  showRetry: { type: Boolean, default: true },
  showCreate: { type: Boolean, default: false },
  loadingText: { type: String, default: '加载中...' },
  errorText: { type: String, default: '加载失败，请检查网络后重试' },
  emptyText: { type: String, default: '暂无数据' },
  createText: { type: String, default: '立即添加' }
})

defineEmits(['retry', 'create'])
</script>

<style scoped>
.state-loading,
.state-error,
.state-empty {
  min-height: 300rpx;
}
.state-retry,
.state-action {
  margin-top: 24rpx;
}
</style>
