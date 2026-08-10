<template>
  <view class="mine-header" :style="{ paddingTop: (statusBarH + 8) + 'px' }">
    <!-- 用户信息行：左侧资料，右侧消息（唯一入口）+ 设置，避开微信胶囊 -->
    <view class="mh-head">
      <view class="mh-profile" hover-class="mh-fade" @tap="onProfile">
        <view class="mh-avatar-wrap">
          <image v-if="vm.avatar" class="mh-avatar" :src="vm.avatar" mode="aspectFill" />
          <view v-else class="mh-avatar mh-avatar-ph">
            <text class="mh-avatar-text">{{ vm.initial }}</text>
          </view>
        </view>
        <view class="mh-copy">
          <view class="mh-name-line">
            <text class="mh-name">{{ vm.name }}</text>
            <text v-if="vm.badge" class="mh-badge" :class="vm.badgeClass">{{ vm.badge }}</text>
          </view>
          <text class="mh-note">{{ vm.note }}</text>
        </view>
        <text class="mh-chev">›</text>
      </view>

      <view class="mh-top-group" :style="{ paddingRight: capsuleGap + 'px' }">
        <view class="mh-top-btn" hover-class="mh-fade" @tap="onMessages">
          <image class="mh-top-icon" :src="icons.message" mode="aspectFit" />
          <view v-if="unreadCount > 0" class="mh-unread-dot"></view>
        </view>
        <view class="mh-top-btn" hover-class="mh-fade" @tap="onSettings">
          <image class="mh-top-icon" :src="icons.settings" mode="aspectFit" />
        </view>
      </view>
    </view>

    <!-- 认证/会员状态横条（仅登录） -->
    <view v-if="vm.showCertBar" class="mh-cert" hover-class="mh-fade" @tap="onCertTap">
      <view class="mh-cert-icon">
        <image :src="vm.certIcon" mode="aspectFit" class="mh-cert-icon-img" />
      </view>
      <text class="mh-cert-main">{{ vm.certMain }}</text>
      <text class="mh-cert-state" :class="vm.certStateClass">{{ vm.certState }}</text>
      <text class="mh-cert-arrow">›</text>
    </view>
  </view>
</template>

<script setup>
// 深蓝身份区展示组件：只接收 view model，不请求接口。
const props = defineProps({
  statusBarH: { type: Number, default: 20 },
  capsuleGap: { type: Number, default: 0 },
  unreadCount: { type: Number, default: 0 },
  vm: {
    type: Object,
    default: () => ({
      name: '点击登录',
      initial: '?',
      avatar: '',
      badge: '',
      badgeClass: '',
      note: '登录后查看需求、对接意向与商城订单',
      showCertBar: false,
      certIcon: '',
      certMain: '',
      certState: '',
      certStateClass: '',
    }),
  },
})

const emit = defineEmits(['messages', 'settings', 'profile', 'cert'])

const icons = {
  message: '/static/mine-icons/message.svg',
  settings: '/static/mine-icons/settings.svg',
}

const onMessages = () => emit('messages')
const onSettings = () => emit('settings')
const onProfile = () => emit('profile')
const onCertTap = () => emit('cert')
</script>

<style scoped>
.mine-header {
  background: #074D92;
  padding: 0 32rpx 32rpx;
  position: relative;
  overflow: hidden;
}
.mine-header::before,
.mine-header::after {
  content: '';
  position: absolute;
  border: 1rpx solid rgba(255,255,255,.13);
  border-radius: 50%;
  pointer-events: none;
}
.mine-header::before {
  width: 270px;
  height: 270px;
  right: -127px;
  top: -107px;
}
.mine-header::after {
  width: 164px;
  height: 164px;
  right: 20px;
  top: 74px;
  border-color: rgba(123,198,255,.16);
}

.mh-head {
  display: flex;
  align-items: center;
  gap: 16rpx;
  position: relative;
  z-index: 1;
}
.mh-top-group {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.mh-top-btn {
  position: relative;
  width: 64rpx;
  height: 64rpx;
  border-radius: 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255,255,255,.10);
  box-sizing: border-box;
}
.mh-top-icon {
  width: 34rpx;
  height: 34rpx;
}
.mh-unread-dot {
  position: absolute;
  top: 12rpx;
  right: 14rpx;
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  background: #F97316;
  border: 3rpx solid #074D92;
}

.mh-profile {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 24rpx;
}
.mh-avatar-wrap { flex-shrink: 0; }
.mh-avatar {
  width: 116rpx;
  height: 116rpx;
  border-radius: 50%;
  display: block;
  box-sizing: border-box;
}
.mh-avatar-ph {
  background: linear-gradient(145deg, #3A8BDD, #0B579F);
  border: 3rpx solid rgba(255,255,255,.46);
  display: flex;
  align-items: center;
  justify-content: center;
}
.mh-avatar-text {
  font-size: 40rpx;
  font-weight: 700;
  color: #DDECFF;
}
.mh-copy { flex: 1; min-width: 0; }
.mh-name-line {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.mh-name {
  font-size: 40rpx;
  font-weight: 700;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 380rpx;
}
.mh-badge {
  flex-shrink: 0;
  min-height: 40rpx;
  padding: 0 16rpx;
  display: inline-flex;
  align-items: center;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 600;
  box-sizing: border-box;
}
.mh-badge.ok {
  color: #D7F5E5;
  background: rgba(16,138,85,.42);
  border: 1rpx solid rgba(204,246,224,.22);
}
.mh-badge.wait {
  color: #FFE9D1;
  background: rgba(185,87,8,.4);
  border: 1rpx solid rgba(255,227,194,.22);
}
.mh-badge.plain {
  color: #E6F3FF;
  background: rgba(255,255,255,.12);
  border: 1rpx solid rgba(255,255,255,.15);
}
.mh-note {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: rgba(255,255,255,.74);
}
.mh-chev {
  color: rgba(255,255,255,.72);
  font-size: 44rpx;
  font-weight: 300;
  flex-shrink: 0;
}

/* 认证横条 */
.mh-cert {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-height: 84rpx;
  margin-top: 28rpx;
  padding: 14rpx 22rpx;
  border-radius: 16rpx;
  background: rgba(0,0,0,.16);
  border: 1rpx solid rgba(255,255,255,.12);
  position: relative;
  z-index: 1;
  box-sizing: border-box;
}
.mh-cert-icon {
  width: 56rpx;
  height: 56rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: rgba(255,255,255,.12);
}
.mh-cert-icon-img {
  width: 32rpx;
  height: 32rpx;
}
.mh-cert-main {
  flex: 1;
  min-width: 0;
  font-size: 24rpx;
  font-weight: 600;
  color: #E7F3FF;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.mh-cert-state {
  flex-shrink: 0;
  padding: 6rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  white-space: nowrap;
}
.mh-cert-state.ok {
  color: #D7F5E5;
  background: rgba(22,138,85,.34);
}
.mh-cert-state.wait {
  color: #FFE9D1;
  background: rgba(185,87,8,.4);
}
.mh-cert-arrow {
  color: rgba(255,255,255,.6);
  font-size: 32rpx;
  font-weight: 300;
}

.mh-fade { opacity: .8; }
</style>
