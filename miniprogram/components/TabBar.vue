<template>
  <view class="tabbar-wrap">
    <!-- 顶部发光边缘 -->
    <view class="tabbar-glow"></view>

    <view class="tabbar-body">
      <view
        v-for="(tab, idx) in tabs"
        :key="idx"
        class="tabbar-item"
        :class="{ 'tabbar-item--active': current === idx }"
        @tap="switchTab(tab.path, idx)"
      >
        <!-- 激活态底部滑动指示条 -->
        <view v-if="current === idx" class="tab-indicator"></view>

        <!-- icon -->
        <view class="tab-icon-box" :class="{ 'tab-icon-box--active': current === idx }">
          <image
            class="tab-icon-img"
            :src="current === idx ? tab.activeSrc : tab.iconSrc"
            mode="aspectFit"
          />

          <!-- 未读角标 -->
          <view v-if="tab.badge > 0" class="tab-badge">
            <text>{{ tab.badge > 99 ? '99+' : tab.badge }}</text>
          </view>
        </view>

        <!-- 标签 -->
        <text
          class="tab-label"
          :class="{ 'tab-label--active': current === idx }"
        >{{ tab.label }}</text>

        <!-- 涟漪 -->
        <view class="tab-ripple"></view>
      </view>
    </view>

    <!-- 安全区 -->
    <view class="tabbar-safe"></view>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps(['current'])

// icons 目录下的 SVG 文件路径（相对于项目根）
const iconRoot = '/static/tabbar/'

const tabs = ref([
  { label: '首页',  iconSrc: iconRoot + 'home.svg',         activeSrc: iconRoot + 'home-active.svg',   path: '/pages/home/index',    badge: 0 },
  { label: '供给',  iconSrc: iconRoot + 'mall.svg',         activeSrc: iconRoot + 'mall-active.svg',   path: '/pages/mall/index',    badge: 0 },
  { label: '发布',  iconSrc: iconRoot + 'publish.svg',      activeSrc: iconRoot + 'publish.svg',        path: '/pages/publish/index', badge: 0 },
  { label: '商家',  iconSrc: iconRoot + 'shop.svg',         activeSrc: iconRoot + 'shop-active.svg',   path: '/pages/shops/index',   badge: 0 },
  { label: '我的',  iconSrc: iconRoot + 'mine.svg',         activeSrc: iconRoot + 'mine-active.svg',   path: '/pages/mine/index',    badge: 0 },
])

const switchTab = (path, idx) => {
  if (idx === props.current) return
  uni.switchTab({ url: path })
}
</script>

<style scoped>
.tabbar-wrap {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 999;
  pointer-events: none;
}

.tabbar-glow {
  height: 2rpx;
  background: linear-gradient(90deg,
    transparent 0%,
    rgba(10, 102, 194, 0.12) 20%,
    rgba(10, 102, 194, 0.20) 50%,
    rgba(10, 102, 194, 0.12) 80%,
    transparent 100%
  );
  pointer-events: none;
}

.tabbar-body {
  display: flex;
  justify-content: space-around;
  align-items: stretch;
  height: 100rpx;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  box-shadow: 0 -4rpx 24rpx rgba(0, 0, 0, 0.04);
  pointer-events: auto;
}

.tabbar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  position: relative;
  padding-top: 6rpx;
  overflow: hidden;
  -webkit-tap-highlight-color: transparent;
}

.tab-indicator {
  position: absolute;
  bottom: 0;
  width: 48rpx;
  height: 6rpx;
  border-radius: 3rpx;
  background: linear-gradient(135deg, #ff6b35, #ff5a1f);
  animation: indicatorSlideIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes indicatorSlideIn {
  from { transform: scaleX(0); opacity: 0; }
  to   { transform: scaleX(1); opacity: 1; }
}

/* icon 容器 */
.tab-icon-box {
  position: relative;
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.tab-icon-box--active {
  transform: scale(1.15);
}

/* icon 图片 */
.tab-icon-img {
  width: 44rpx;
  height: 44rpx;
}

/* 未读角标 */
.tab-badge {
  position: absolute;
  top: -6rpx;
  right: -10rpx;
  min-width: 28rpx;
  height: 28rpx;
  line-height: 28rpx;
  padding: 0 8rpx;
  font-size: 18rpx;
  font-weight: 700;
  color: #fff;
  background: #ff3b30;
  border-radius: 14rpx;
  text-align: center;
  border: 2rpx solid #fff;
  box-shadow: 0 2rpx 8rpx rgba(255, 59, 48, 0.3);
  animation: badgePopIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes badgePopIn {
  from { transform: scale(0); }
  to   { transform: scale(1); }
}

/* 标签文字 */
.tab-label {
  font-size: 20rpx;
  font-weight: 500;
  color: #969799;
  transition: color 0.2s ease;
}

.tab-label--active {
  color: #1a1a1a;
  font-weight: 600;
}

/* 涟漪 */
.tab-ripple {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.tabbar-item:active .tab-ripple {
  background: rgba(10, 102, 194, 0.06);
  animation: rippleFade 0.4s ease-out forwards;
}

@keyframes rippleFade {
  0%   { opacity: 1; transform: scale(0.6); }
  100% { opacity: 0; transform: scale(1); }
}

/* 安全区 */
.tabbar-safe {
  height: env(safe-area-inset-bottom);
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
</style>
