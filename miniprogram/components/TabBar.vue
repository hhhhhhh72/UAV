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
        @tap="handleTap(tab, idx)"
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

    <!-- 发布选择弹层：需求 / 服务能力 / 商品设备 -->
    <view v-if="showPublishSheet" class="pub-mask" @tap="closePublishSheet">
      <view class="pub-sheet" @tap.stop>
        <view class="pub-sheet-grip"></view>
        <view class="pub-sheet-head">
          <text class="pub-sheet-title">选择发布类型</text>
          <view class="pub-sheet-close" @tap="closePublishSheet">
            <text class="pub-sheet-x">×</text>
          </view>
        </view>
        <view class="pub-choice-list">
          <view class="pub-choice" hover-class="tap-fade" @tap="choosePublish('demand')">
            <view class="pub-choice-icon demand">需</view>
            <view class="pub-choice-copy">
              <text class="pub-choice-name">发布需求</text>
              <text class="pub-choice-desc">发布作业、采购、技术或场景需求</text>
            </view>
            <text class="pub-choice-arrow">›</text>
          </view>
          <view class="pub-choice" hover-class="tap-fade" @tap="choosePublish('service')">
            <view class="pub-choice-icon service">服</view>
            <view class="pub-choice-copy">
              <text class="pub-choice-name">发布服务能力</text>
              <text class="pub-choice-desc">展示巡检、测绘、航拍等可承接能力</text>
            </view>
            <text class="pub-choice-arrow">›</text>
          </view>
          <view class="pub-choice" hover-class="tap-fade" @tap="choosePublish('product')">
            <view class="pub-choice-icon product">商</view>
            <view class="pub-choice-copy">
              <text class="pub-choice-name">发布商品设备</text>
              <text class="pub-choice-desc">展示设备租赁、整机、零部件或载荷</text>
            </view>
            <text class="pub-choice-arrow">›</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps(['current'])

// icons 目录下的 SVG 文件路径（相对于项目根）
const iconRoot = '/static/tabbar/'

const tabs = ref([
  { label: '首页',    iconSrc: iconRoot + 'home.svg',         activeSrc: iconRoot + 'home-active.svg', path: '/pages/home/index',     badge: 0, kind: 'tab' },
  { label: '供需大厅', iconSrc: iconRoot + 'demands.svg',      activeSrc: iconRoot + 'demands-active.svg', path: '/pages/demands/index',  badge: 0, kind: 'tab' },
  { label: '发布',    iconSrc: iconRoot + 'publish.svg',      activeSrc: iconRoot + 'publish.svg',       path: '/pages/publish/index',  badge: 0, kind: 'sheet' },
  { label: '生态服务', iconSrc: iconRoot + 'services.svg',     activeSrc: iconRoot + 'services-active.svg', path: '/pages/services/index', badge: 0, kind: 'tab' },
  { label: '我的',    iconSrc: iconRoot + 'mine.svg',         activeSrc: iconRoot + 'mine-active.svg',   path: '/pages/mine/index',     badge: 0, kind: 'none' },
])

const showPublishSheet = ref(false)

const handleTap = (tab, idx) => {
  // 已选中 tab 直接忽略
  if (idx === props.current && tab.kind !== 'sheet') return

  if (tab.kind === 'sheet') {
    // 发布 → 打开选择入口
    showPublishSheet.value = true
    return
  }
  if (tab.kind === 'none') {
    // 我的：保留图标与文字，移除点击路径与跳转
    uni.showToast({ title: '「我的」功能将陆续开放', icon: 'none' })
    return
  }
  uni.switchTab({ url: tab.path })
}

const closePublishSheet = () => { showPublishSheet.value = false }

const choosePublish = (type) => {
  showPublishSheet.value = false
  uni.navigateTo({ url: '/pages/demands/publish?type=' + type })
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
  background: #0A66C2;
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
  color: #0A66C2;
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

/* ===== 发布选择弹层 ===== */
.pub-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 2000;
  background: rgba(16, 24, 40, 0.45);
  display: flex;
  align-items: flex-end;
  pointer-events: auto;
}

.pub-sheet {
  width: 100%;
  background: #fff;
  border-radius: 20rpx 20rpx 0 0;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}

.pub-sheet-grip {
  width: 68rpx;
  height: 8rpx;
  background: #d8dde3;
  margin: 16rpx auto;
  border-radius: 4rpx;
}

.pub-sheet-head {
  display: flex;
  align-items: center;
  padding: 4rpx 28rpx 16rpx;
}

.pub-sheet-title {
  flex: 1;
  font-size: 34rpx;
  font-weight: 700;
  color: #17212B;
}

.pub-sheet-close {
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pub-sheet-x {
  font-size: 40rpx;
  color: #98A2B3;
  line-height: 1;
}

.pub-choice-list {
  padding: 0 24rpx 8rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.pub-choice {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  border: 1px solid #E4E7EC;
  border-radius: 16rpx;
  background: #fff;
  text-align: left;
}

.pub-choice-icon {
  width: 86rpx;
  height: 86rpx;
  flex-shrink: 0;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40rpx;
  font-weight: 700;
}

.pub-choice-icon.demand { color: #0A66C2; background: #EAF3FB; }
.pub-choice-icon.service { color: #D15A10; background: #FFF0E6; }
.pub-choice-icon.product { color: #168A55; background: #E9F7F0; }

.pub-choice-copy {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.pub-choice-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
}

.pub-choice-desc {
  font-size: 22rpx;
  color: #667085;
  line-height: 1.45;
}

.pub-choice-arrow {
  font-size: 40rpx;
  color: #98A2B3;
}

.tap-fade {
  opacity: 0.85;
}
</style>
