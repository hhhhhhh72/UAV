<template>
  <view>
    <!-- 加载中 -->
    <view v-if="loading" class="state-loading flex-center flex-col" style="padding: 80rpx 0;">
      <van-loading size="48rpx" color="var(--color-primary)" vertical>
        <text class="text-secondary" style="margin-top: 16rpx;">{{ loadingText }}</text>
      </van-loading>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="error" class="state-error flex-center flex-col" style="padding: 80rpx 0;">
      <van-empty image="error" :description="errorText" />
      <view v-if="showRetry" class="state-retry">
        <van-button type="primary" size="small" round @click="$emit('retry')">重新加载</van-button>
      </view>
    </view>

    <!-- 空数据 -->
    <view v-else-if="empty" class="state-empty flex-center flex-col" style="padding: 80rpx 0;">
      <van-empty
        :image="emptyImage"
        :description="emptyText"
      />
      <view v-if="showCreate" class="state-action">
        <van-button type="primary" size="small" round @click="$emit('create')">
          {{ createText }}
        </van-button>
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
  emptyImage: { type: String, default: 'search' },
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
