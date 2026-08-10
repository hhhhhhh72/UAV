<template>
  <view class="search-page">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <view class="search-box">
        <view class="search-icon"></view>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索需求、服务或设备"
          confirm-type="search"
          focus
          @confirm="onSearch"
        />
        <view v-if="keyword" class="clear-btn" @tap="keyword = ''"><text class="clear-x">×</text></view>
      </view>
      <view class="search-btn" @tap="onSearch"><text>搜索</text></view>
      <view class="cancel-btn" @tap="goBack"><text>取消</text></view>
    </view>

    <!-- 有结果 -->
    <template v-if="searched">
      <view class="result-head">找到 {{ results.length }} 条相关内容</view>
      <view v-if="results.length" class="card-list">
        <view
          v-for="item in results"
          :key="item.id"
          class="trade-card"
          hover-class="tap-fade"
          @tap="goDetail(item)"
        >
          <view class="trade-card-main">
            <image :src="item.image" mode="aspectFill" class="trade-visual" />
            <view class="trade-body">
              <view class="tag-row">
                <text class="tag blue">{{ item.cat }}</text>
                <text class="tag" :class="statusTagClass(item)">{{ item.status }}</text>
              </view>
              <text class="trade-title">{{ item.title }}</text>
              <view class="trade-meta">
                <text>{{ item.region }}</text>
                <text>{{ item.time }}</text>
              </view>
            </view>
          </view>
          <view class="trade-footer">
            <view class="price-block">
              <text class="price">{{ item.price }}</text>
              <text class="price-unit"> {{ item.unit }}</text>
            </view>
            <view class="card-action"><text>查看详情 ›</text></view>
          </view>
        </view>
      </view>
      <view v-else class="state-panel">
        <view class="state-mark">⌁</view>
        <text class="state-title">没有找到相关内容</text>
        <text class="state-desc">换个关键词试试</text>
      </view>
    </template>

    <!-- 推荐 / 最近搜索 -->
    <template v-else>
      <view class="search-block">
        <text class="block-title">推荐搜索</text>
        <view class="keyword-row">
          <view v-for="w in hotWords" :key="w" class="keyword" @tap="keyword = w; onSearch()">{{ w }}</view>
        </view>
      </view>
      <view class="search-block">
        <text class="block-title">最近搜索</text>
        <view class="keyword-row">
          <view v-for="w in recentWords" :key="w" class="keyword" @tap="keyword = w; onSearch()">{{ w }}</view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { safeNavigateTo } from '../../../utils/nav'
import { getKindItems, isEnded } from '../../../utils/hallData'

const keyword = ref('')
const searched = ref(false)
const results = ref([])

const hotWords = ['巡检', '航拍', '测绘', '吊运', '设备租赁', '植保']
const recentWords = ['光伏巡检', 'M350']

function allItems() {
  return [...getKindItems('demand'), ...getKindItems('supply', 'service'), ...getKindItems('supply', 'product')]
}

function onSearch() {
  const kw = keyword.value.trim()
  searched.value = true
  if (!kw) {
    results.value = []
    return
  }
  results.value = allItems().filter(
    (i) => i.title.includes(kw) || i.cat.includes(kw) || i.company.includes(kw)
  )
}

function statusTagClass(item) {
  if (isEnded(item)) return 'gray'
  if (item.type === '商品') return 'orange'
  return 'green'
}

const goDetail = (item) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(item.id))
const goBack = () => uni.navigateBack()
</script>

<style scoped>
.search-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 40rpx;
}

.tap-fade { opacity: 0.85; }

.search-bar {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 20rpx 28rpx;
  background: #074D92;
  padding-top: calc(env(safe-area-inset-top) + 20rpx);
}
.search-box {
  flex: 1;
  height: 78rpx;
  border-radius: 12rpx;
  background: #fff;
  display: flex;
  align-items: center;
  gap: 14rpx;
  padding: 0 20rpx;
  min-width: 0;
}
.search-icon {
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.search-icon::after {
  content: '';
  position: absolute;
  right: -11rpx;
  bottom: -6rpx;
  width: 13rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #98A2B3;
  transform: rotate(45deg);
}
.search-input { flex: 1; font-size: 26rpx; color: #17212B; }
.clear-btn { padding: 4rpx; }
.clear-x { font-size: 30rpx; color: #98A2B3; line-height: 1; }
.search-btn { color: #fff; font-size: 26rpx; font-weight: 600; white-space: nowrap; }
.cancel-btn { color: rgba(255, 255, 255, 0.9); font-size: 26rpx; white-space: nowrap; }

.result-head {
  margin: 32rpx 32rpx 8rpx;
  color: #667085;
  font-size: 24rpx;
}

/* 搜索推荐 */
.search-block { padding: 36rpx 32rpx 0; }
.block-title { display: block; font-size: 30rpx; font-weight: 700; color: #17212B; margin-bottom: 24rpx; }
.keyword-row { display: flex; flex-wrap: wrap; gap: 16rpx; }
.keyword {
  color: #344054;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10rpx;
  padding: 14rpx 20rpx;
  font-size: 24rpx;
}

/* 结果卡片 */
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
.tag-row { display: flex; gap: 10rpx; align-items: center; margin-bottom: 12rpx; }
.tag {
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.gray { color: #667085; background: #F1F3F5; }
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
.trade-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 80rpx;
  padding: 0 24rpx;
  border-top: 1px solid #EEF1F4;
}
.price-block { display: flex; align-items: baseline; min-width: 0; }
.price { color: #DF620F; font-size: 30rpx; font-weight: 750; }
.price-unit { color: #667085; font-size: 20rpx; }
.card-action { color: #0A66C2; font-size: 24rpx; font-weight: 650; white-space: nowrap; }

/* 空状态 */
.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }

@media (max-width: 380px) {
  .trade-visual { width: 150rpx; height: 150rpx; }
}
</style>
