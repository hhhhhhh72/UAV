<template>
  <view class="u-tabs">
    <scroll-view scroll-x :show-scrollbar="false" class="u-tabs-scroll">
      <view class="u-tabs-inner">
        <view
          v-for="(t, i) in titles"
          :key="i"
          class="u-tabs-item"
          :class="{ 'u-tabs-item--active': i === active }"
          @click="onSelect(i)"
        >
          <text class="u-tabs-text">{{ t }}</text>
        </view>
      </view>
    </scroll-view>
    <view class="u-tabs-line" :style="{ left: lineLeft + '%', width: lineWidth + '%' }" />
  </view>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  active: { type: Number, default: 0 },
  // 标题列表。uni-app 小程序端无法可靠地从 slots.default() 读取子组件
  // （<u-tab>）的 props，故标题一律由调用方通过 titles 传入；u-tab 仅作
  // 占位/内容分发容器，其 title 属性不再被 u-tabs 读取。
  titles: { type: Array, default: () => [] },
  // 预留的 van-tabs 兼容属性（当前未使用）
  type: { type: String, default: 'line' },
  swipeThreshold: { type: Number, default: 5 }
})
const emit = defineEmits(['update:active', 'change'])
const itemCount = computed(() => props.titles.length || 1)
const lineLeft = computed(() => (props.active / itemCount.value) * 100)
const lineWidth = computed(() => (100 / itemCount.value) * 80)
function onSelect(i) {
  emit('update:active', i)
  emit('change', i)
}
</script>

<style scoped>
.u-tabs { position: relative; background: #fff; }
.u-tabs-scroll { white-space: nowrap; }
.u-tabs-inner { display: inline-flex; }
.u-tabs-item { padding: 24rpx 32rpx; font-size: 30rpx; color: var(--ui-color-text-secondary, #969799); position: relative; }
.u-tabs-item--active { color: var(--color-primary, #0A66C2); font-weight: 600; }
.u-tabs-line { position: absolute; bottom: 0; height: 6rpx; border-radius: 3rpx; background: var(--color-primary, #0A66C2); transition: left 0.2s; }
</style>
