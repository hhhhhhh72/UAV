<template>
  <view class="u-steps">
    <view
      v-for="(s, i) in steps"
      :key="s.key || i"
      class="u-step"
      :class="{ 'u-step--done': i < activeIndex, 'u-step--on': i === activeIndex }"
    >
      <view class="u-step-track">
        <view v-if="i > 0" class="u-step-line" :class="{ 'u-step-line--on': i <= activeIndex }" />
        <view class="u-step-dot">
          <text class="u-step-mark">{{ i < activeIndex ? '✓' : (i + 1) }}</text>
        </view>
        <view v-if="i < steps.length - 1" class="u-step-line" :class="{ 'u-step-line--on': i < activeIndex }" />
      </view>
      <text class="u-step-label">{{ s.label }}</text>
    </view>
  </view>
</template>

<script setup>
// 横向步骤条。easycom 规则 ^u-(.*) 已把它注册为 <u-steps>，页面无需 import。
// activeIndex 为当前停在第几步（0 基）；i < activeIndex 视为已完成。
defineProps({
  steps: { type: Array, default: () => [] },
  activeIndex: { type: Number, default: 0 },
})
</script>

<style scoped>
.u-steps {
  display: flex;
  align-items: flex-start;
  width: 100%;
}
.u-step {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}
/* 每步内部左右各半条连线，圆点自然落在该步正中 */
.u-step-track {
  width: 100%;
  display: flex;
  align-items: center;
}
.u-step-line {
  flex: 1;
  height: 2rpx;
  background: #E4EAF2;
}
.u-step-line--on { background: #0A66C2; }
.u-step-dot {
  width: 40rpx;
  height: 40rpx;
  flex: 0 0 40rpx;
  border-radius: 50%;
  background: #E8EDF3;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.u-step--done .u-step-dot { background: #0A66C2; }
/* 当前步加一圈光晕，与"已完成"的实心圆区分开 */
.u-step--on .u-step-dot {
  background: #0A66C2;
  box-shadow: 0 0 0 6rpx rgba(10, 102, 194, 0.14);
}
.u-step-mark {
  font-size: 22rpx;
  font-weight: 700;
  color: #98A2B3;
  line-height: 1;
}
.u-step--done .u-step-mark,
.u-step--on .u-step-mark { color: #fff; }
.u-step-label {
  margin-top: 10rpx;
  font-size: 20rpx;
  color: #98A2B3;
  line-height: 1.2;
  text-align: center;
}
.u-step--done .u-step-label { color: #0A66C2; }
.u-step--on .u-step-label { color: #0A66C2; font-weight: 700; }
</style>
