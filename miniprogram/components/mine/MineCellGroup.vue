<template>
  <view class="mcg-group">
    <view
      v-for="(item, i) in items"
      :key="i"
      class="mcg-item"
      hover-class="mcg-fade"
      @tap="onItem(i)"
    >
      <view class="mcg-icon" :class="'mcg-icon--' + (item.tone || 'primary')">
        <image class="mcg-icon-img" :src="item.icon" mode="aspectFit" />
      </view>
      <view class="mcg-copy">
        <text class="mcg-name">{{ item.label }}</text>
        <text v-if="item.desc" class="mcg-desc">{{ item.desc }}</text>
      </view>
      <text v-if="item.tail" class="mcg-tail" :class="item.tailClass">{{ item.tail }}</text>
      <text class="mcg-chev">›</text>
    </view>
  </view>
</template>

<script setup>
// 紧凑行列表：容器无内容 padding，由每行承担 28rpx 水平 padding。
const props = defineProps({
  items: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['select'])

const onItem = (i) => emit('select', i)
</script>

<style scoped>
.mcg-group { background: #fff; }
.mcg-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  min-height: 104rpx;
  padding: 0 28rpx;
  border-top: 1rpx solid #EEF1F4;
  box-sizing: border-box;
}
.mcg-item:first-child { border-top: none; }
.mcg-icon {
  width: 56rpx;
  height: 56rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
}
.mcg-icon--primary { background: #E8F2FC; }
.mcg-icon--gray { background: #EEF2F6; }
.mcg-icon--orange { background: #FFF0E6; }
.mcg-icon--violet { background: #F0EDFF; }
.mcg-icon-img {
  width: 32rpx;
  height: 32rpx;
}
.mcg-copy { flex: 1; min-width: 0; }
.mcg-name {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: #17212B;
}
.mcg-desc {
  display: block;
  margin-top: 4rpx;
  font-size: 20rpx;
  color: #98A2B3;
}
.mcg-tail {
  flex-shrink: 0;
  font-size: 22rpx;
  color: #98A2B3;
}
.mcg-tail.ok { color: #168A55; }
.mcg-tail.wait { color: #B54708; }
.mcg-tail.danger { color: #D92D20; }
.mcg-tail.dim { color: #C0C8D2; }
.mcg-chev {
  color: #98A2B3;
  font-size: 30rpx;
  font-weight: 300;
  margin-left: 4rpx;
}
.mcg-fade { opacity: .8; }
</style>
