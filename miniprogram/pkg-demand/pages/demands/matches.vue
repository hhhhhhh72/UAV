<template>
  <view class="matches-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">智能匹配</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 匹配说明 -->
    <view class="match-hero">
      <view class="tag-row">
        <text class="tag blue">{{ isSupplyMode ? '为我找需求' : '为我找供给' }}</text>
      </view>
      <text class="match-title">{{ isSupplyMode ? '与你的服务能力匹配的需求' : '可承接当前需求的服务与设备' }}</text>
      <text class="match-desc">匹配依据：业务领域、服务区域、作业能力</text>
    </view>

    <!-- 推荐标题 -->
    <view class="recommend-head">
      <text class="recommend-title">推荐结果</text>
      <text class="recommend-note">匹配度由高到低</text>
    </view>

    <!-- 推荐列表 -->
    <view class="card-list">
      <view
        v-for="(item, index) in shownItems"
        :key="item.id"
        class="trade-card"
        hover-class="tap-fade"
        @tap="goDetail(item)"
      >
        <view class="trade-card-main">
          <image :src="item.image" mode="aspectFill" class="trade-visual" />
          <view class="trade-body">
            <view class="tag-row">
              <text class="tag green">匹配度 {{ 94 - index * 7 }}%</text>
            </view>
            <text class="trade-title">{{ item.title }}</text>
            <view class="trade-meta"><text>{{ item.region }}</text></view>
            <view class="tag-row match-tags">
              <text class="tag blue">{{ index === 0 ? '业务类型匹配' : '服务区域匹配' }}</text>
              <text class="tag blue">{{ index === 0 ? '区域覆盖' : '预算合理' }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import { safeNavigateTo } from '../../../utils/nav'
import { getKindItems } from '../../../utils/hallData'

const isSupplyMode = false // 本期按需求方视角展示

const shownItems = computed(() => {
  const pool = isSupplyMode ? getKindItems('demand') : getKindItems('supply', 'service')
  return pool.slice(0, 2)
})

const goDetail = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
const goBack = () => uni.navigateBack()
</script>

<style scoped>
.matches-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 40rpx;
}

.tap-fade { opacity: 0.85; }

.page-header {
  height: 56px;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-spacer { width: 72rpx; }

.match-hero {
  padding: 36rpx 32rpx 28rpx;
  background: #fff;
  border-bottom: 16rpx solid #F4F6F8;
}
.tag-row { display: flex; gap: 10rpx; align-items: center; }
.tag {
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.green { color: #168A55; background: #E9F7F0; }
.match-title {
  display: block;
  font-size: 40rpx;
  font-weight: 750;
  color: #17212B;
  line-height: 1.35;
  margin-top: 20rpx;
}
.match-desc {
  display: block;
  color: #667085;
  font-size: 24rpx;
  margin-top: 12rpx;
}

.recommend-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.recommend-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.recommend-note { font-size: 22rpx; color: #667085; }

.card-list { padding: 0 32rpx 32rpx; display: flex; flex-direction: column; gap: 20rpx; }
.trade-card {
  background: #fff;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  border: 1px solid rgba(228, 231, 236, 0.7);
  overflow: hidden;
}
.trade-card-main { display: flex; gap: 22rpx; padding: 24rpx; }
.trade-visual {
  width: 164rpx;
  height: 164rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  background: #E8F2FC;
}
.trade-body { flex: 1; min-width: 0; }
.trade-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #17212B;
  font-size: 28rpx;
  line-height: 1.42;
  font-weight: 700;
}
.trade-meta {
  display: flex;
  gap: 14rpx;
  margin-top: 14rpx;
  color: #667085;
  font-size: 22rpx;
  white-space: nowrap;
  overflow: hidden;
}
.match-tags { margin-top: 14rpx; }
</style>
