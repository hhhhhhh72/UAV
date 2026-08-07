<template>
  <view class="mine-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">我的发布</text>
      <view class="head-action" @tap="goIntents">
        <text class="head-action-text">收到的意向</text>
      </view>
    </view>

    <!-- 类型筛选 -->
    <view class="mine-filters">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="t in typeOptions"
            :key="t"
            class="filter-chip"
            :class="{ active: mineType === t }"
            @tap="mineType = t"
          >{{ t }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 状态筛选 -->
    <view class="mine-filters status">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="s in statusOptions"
            :key="s"
            class="filter-chip"
            :class="{ active: mineStatus === s }"
            @tap="mineStatus = s"
          >{{ s }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">发布记录</text>
      <text class="list-count">共 {{ filteredPosts.length }} 条</text>
    </view>

    <!-- 空状态 -->
    <view v-if="filteredPosts.length === 0" class="state-panel">
      <view class="state-mark">⌁</view>
      <text class="state-title">暂无符合条件的发布</text>
      <text class="state-desc">换个筛选条件试试，或先发布一条信息</text>
      <view class="state-btn" @tap="resetMineFilter">清除筛选</view>
    </view>

    <!-- 发布记录 -->
    <view v-else class="post-list">
      <view v-for="post in filteredPosts" :key="post.id" class="mine-card">
        <view class="tag-row">
          <text class="tag" :class="typeTagClass(post.type)">{{ post.type }}</text>
          <text class="tag" :class="statusTagClass(post.status)">{{ post.status }}</text>
        </view>
        <text class="post-title">{{ post.title }}</text>
        <text class="post-meta">{{ post.date }}{{ post.reason ? ' · ' + post.reason : '' }}</text>
        <view class="mine-action-row">
          <template v-if="post.status === '已驳回'">
            <view class="action-link" @tap="republish(post)">编辑重提</view>
          </template>
          <template v-else-if="post.status === '已上架'">
            <view class="action-link" @tap="goIntents">查看意向</view>
            <view class="action-link" @tap="completePost(post)">标记完成</view>
            <view class="action-link danger" @tap="closePost(post)">下架</view>
          </template>
          <template v-else>
            <view class="action-link" @tap="toastPending">查看审核进度</view>
          </template>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { safeNavigateTo } from '../../utils/nav'
import { getPosts, savePosts } from '../../utils/hallData'

const mineType = ref('全部')
const mineStatus = ref('全部')

const typeOptions = ['全部', '需求', '服务', '商品']
const statusOptions = ['全部', '待审核', '已上架', '已驳回', '已结束']

const posts = ref(getPosts())

const filteredPosts = computed(() => {
  return posts.value.filter(
    (p) =>
      (mineType.value === '全部' || p.type === mineType.value) &&
      (mineStatus.value === '全部' || p.status === mineStatus.value)
  )
})

function typeTagClass(t) {
  return t === '需求' ? 'blue' : t === '服务' ? 'green' : 'orange'
}
function statusTagClass(s) {
  return s === '已驳回' ? 'red' : s === '待审核' ? 'orange' : s === '已结束' ? 'gray' : 'green'
}

const resetMineFilter = () => {
  mineType.value = '全部'
  mineStatus.value = '全部'
}

const goIntents = () => safeNavigateTo('/pages/demands/intents')
const goBack = () => uni.navigateBack()

function republish(post) {
  safeNavigateTo('/pages/demands/publish?type=' + postTypeValue(post.type))
}
function postTypeValue(t) {
  return t === '需求' ? 'demand' : t === '服务' ? 'service' : 'product'
}

function completePost(post) {
  post.status = '已结束'
  savePosts(posts.value)
  uni.showToast({ title: '已标记完成', icon: 'success' })
}

function closePost(post) {
  post.status = '已结束'
  savePosts(posts.value)
  uni.showToast({ title: '已下架', icon: 'none' })
}

const toastPending = () => {
  uni.showToast({ title: '审核进度：等待协会审核', icon: 'none' })
}
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* 头部 */
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
.head-action { padding: 14rpx; }
.head-action-text { color: #0A66C2; font-size: 26rpx; font-weight: 600; }

/* 筛选 */
.mine-filters {
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 20rpx 24rpx;
}
.mine-filters.status { border-bottom: 0; padding-top: 4rpx; }
.filter-scroll { white-space: nowrap; }
.filter-inner { display: inline-flex; gap: 12rpx; }
.filter-chip {
  display: inline-flex;
  align-items: center;
  height: 56rpx;
  padding: 0 20rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #344054;
  font-size: 24rpx;
  box-sizing: border-box;
}
.filter-chip.active {
  color: #0A66C2;
  border-color: #B9D6EF;
  background: #EAF3FB;
  font-weight: 650;
}

/* 列表 */
.list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

.post-list { padding: 0 32rpx 32rpx; }
.mine-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 26rpx;
  border: 1px solid #EEF1F4;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.mine-card + .mine-card { margin-top: 20rpx; }

.tag-row { display: flex; gap: 10rpx; }
.tag {
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.red { color: #D92D20; background: #FEF3F2; }
.tag.gray { color: #667085; background: #F1F3F5; }

.post-title {
  display: block;
  font-size: 28rpx;
  line-height: 1.45;
  color: #17212B;
  font-weight: 700;
  margin: 16rpx 0 8rpx;
}
.post-meta { display: block; font-size: 22rpx; color: #667085; }

.mine-action-row {
  display: flex;
  gap: 32rpx;
  border-top: 1px solid #EEF1F4;
  margin-top: 22rpx;
  padding-top: 20rpx;
}
.action-link { color: #0A66C2; font-size: 24rpx; font-weight: 600; }
.action-link.danger { color: #D92D20; }

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
.state-desc { margin: 12rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}
</style>
