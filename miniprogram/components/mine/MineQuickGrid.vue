<template>
  <view class="mqg-grid">
    <view
      v-for="(item, i) in items"
      :key="i"
      class="mqg-cell"
      hover-class="mqg-fade"
      @tap="onItem(i)"
    >
      <view class="mqg-icon" :class="'mqg-icon--' + item.tone">
        <image class="mqg-icon-img" :src="item.icon" mode="aspectFit" />
      </view>
      <text class="mqg-label">{{ item.label }}</text>
    </view>
  </view>
</template>

<script setup>
// 3×2 业务宫格展示组件：icon 为统一线性 SVG，tone 决定底色色序。
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
.mqg-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
}
.mqg-cell {
  min-height: 164rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 16rpx 4rpx;
}
.mqg-icon {
  width: 68rpx;
  height: 68rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.mqg-icon-img {
  width: 36rpx;
  height: 36rpx;
}
/* 色序：发布蓝、意向紫、预约橙、报名绿、收藏金、订单灰蓝 */
.mqg-icon--publish { background: #E8F2FC; }
.mqg-icon--intent { background: #F0EDFF; }
.mqg-icon--appointment { background: #FFF0E6; }
.mqg-icon--enroll { background: #E9F7F0; }
.mqg-icon--favorite { background: #FFF4DF; }
.mqg-icon--order { background: #EFF2F5; }
.mqg-label {
  font-size: 22rpx;
  color: #344054;
}
.mqg-fade { opacity: .8; }
</style>
