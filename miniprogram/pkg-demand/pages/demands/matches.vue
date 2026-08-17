<template>
  <view class="matches-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">智能匹配</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 匹配说明 -->
    <view class="match-hero">
      <view class="tag-row">
        <text class="tag blue">为你找需求</text>
      </view>
      <text class="match-title">输入关键词，匹配与你能力契合的需求</text>
      <text class="match-desc">匹配依据：业务领域、服务区域、作业能力</text>

      <!-- 搜索输入 -->
      <view class="search-box">
        <view class="search-icon" />
        <input
          class="search-input"
          v-model="keyword"
          placeholder="如：巡检、吊运、植保…"
          placeholder-class="search-ph"
          confirm-type="search"
          @confirm="doSearch"
        />
        <view v-if="keyword" class="search-clear" hover-class="tap-fade" hover-stay-time="120" @tap="keyword = ''" />
      </view>

      <!-- 快捷分类 -->
      <scroll-view scroll-x class="chip-scroll" :show-scrollbar="false">
        <view class="chip-row">
          <view
            v-for="c in quickCats"
            :key="c"
            class="chip"
            :class="{ 'chip-active': activeCat === c }"
            @tap="pickCat(c)"
          >
            <text>{{ c }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="state-panel">
      <view class="loading-ring"></view>
      <text class="state-desc">正在为你匹配…</text>
    </view>

    <!-- 初始引导 -->
    <view v-else-if="!searched" class="state-panel">
      <view class="state-mark">
        <view class="state-mark-inner">
          <view class="state-search" />
        </view>
      </view>
      <text class="state-title">输入关键词开始匹配</text>
      <text class="state-desc">例如「巡检」「吊运」「植保」，或直接选择上方分类</text>
    </view>

    <!-- 无结果 -->
    <view v-else-if="results.length === 0" class="state-panel">
      <view class="state-mark">
        <view class="state-mark-inner">
          <view class="state-search" />
        </view>
      </view>
      <text class="state-title">暂未找到匹配需求</text>
      <text class="state-desc">换个关键词试试，或去需求大厅浏览全部</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goHall">
        <text>去需求大厅</text>
      </view>
    </view>

    <!-- 推荐结果 -->
    <template v-else>
      <view class="recommend-head">
        <text class="recommend-title">匹配结果</text>
        <text class="recommend-note">共 {{ results.length }} 条，匹配度由高到低</text>
      </view>

      <view class="card-list">
        <view
          v-for="(item, index) in results"
          :key="item.demand.id"
          class="trade-card"
          :style="{ animationDelay: (index * 60) + 'ms' }"
          hover-class="card-hover"
          :hover-stay-time="120"
          @tap="goDetail(item.demand)"
        >
          <view class="trade-card-main">
            <view class="trade-body">
              <view class="tag-row">
                <text class="tag green">匹配度 {{ matchPercent(item.score) }}%</text>
                <text class="tag blue">{{ bizLabel(item.demand.biz_type) }}</text>
              </view>
              <text class="trade-title">{{ item.demand.title }}</text>
              <text v-if="item.demand.description" class="trade-desc">{{ item.demand.description }}</text>
              <view class="trade-meta">
                <text v-if="item.demand.district">{{ item.demand.district }}</text>
                <text v-if="item.demand.created_at">发布 {{ fmtDate(item.demand.created_at) }}</text>
              </view>
              <view v-if="item.reasons && item.reasons.length" class="match-tags">
                <text v-for="(r, i) in item.reasons.slice(0, 3)" :key="i" class="reason-tag">{{ r }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { safeNavigateTo } from '../../../utils/nav'
import { request, BASE_URL } from '../../../utils/request'

const quickCats = ['巡检', '吊运', '植保', '测绘', '航拍']

const keyword = ref('')
const activeCat = ref('')
const results = ref([])
const loading = ref(false)
const searched = ref(false)

// 业务类型 → 中文标签（与首页/需求大厅一致）
const BIZ_LABELS = {
  cable_inspection: '工业巡检',
  plant_transport: '植保运输',
  spray_pesticide: '农药喷洒',
  clean_paint: '清洗保洁',
  trade_lease: '租赁服务',
  other: '其他服务',
}
const bizLabel = (b) => BIZ_LABELS[b] || b || '综合服务'

// 匹配度：0~1 → 百分比（保留整数）
const matchPercent = (score) => Math.round((Number(score) || 0) * 100)

const fmtDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return ''
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

const doSearch = async () => {
  const q = keyword.value.trim()
  if (!q && !activeCat.value) {
    uni.showToast({ title: '请输入关键词或选择分类', icon: 'none' })
    return
  }
  loading.value = true
  try {
    const params = { limit: 20 }
    if (q) params.q = q
    if (activeCat.value) params.q = activeCat.value
    const res = await request({ url: '/api/v1/match', data: params })
    results.value = Array.isArray(res) ? res : res?.results || []
    searched.value = true
  } catch (e) {
    results.value = []
    searched.value = true
    uni.showToast({ title: '匹配失败，请稍后重试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const pickCat = (c) => {
  activeCat.value = c
  keyword.value = c
  doSearch()
}

const goDetail = (d) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(d.id))
const goHall = () => safeNavigateTo('/pages/demands/index')
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
.tag-row { display: flex; gap: 10rpx; align-items: center; flex-wrap: wrap; }
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

/* 搜索框（对齐输入框规范：radius 24rpx） */
.search-box {
  display: flex;
  align-items: center;
  gap: 14rpx;
  height: 76rpx;
  margin-top: 24rpx;
  padding: 0 24rpx;
  border-radius: 24rpx;
  background: #F4F6F8;
  border: 1rpx solid #E4E7EC;
}
.search-icon {
  width: 24rpx;
  height: 24rpx;
  border: 3rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}
.search-icon::after {
  content: '';
  position: absolute;
  right: -9rpx;
  bottom: -6rpx;
  width: 12rpx;
  height: 3rpx;
  background: #98A2B3;
  border-radius: 2rpx;
  transform: rotate(45deg);
}
.search-input { flex: 1; height: 76rpx; font-size: 26rpx; color: #17212B; }
.search-ph { color: #98A2B3; }
.search-clear {
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: #D9DEE4;
  position: relative;
  flex-shrink: 0;
}
.search-clear::before,
.search-clear::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 14rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #fff;
}
.search-clear::before { transform: translate(-50%, -50%) rotate(45deg); }
.search-clear::after { transform: translate(-50%, -50%) rotate(-45deg); }

/* 快捷分类 chips */
.chip-scroll { width: 100%; white-space: nowrap; margin-top: 20rpx; }
.chip-row { display: inline-flex; gap: 14rpx; }
.chip {
  display: inline-flex;
  align-items: center;
  height: 56rpx;
  padding: 0 26rpx;
  border-radius: 999rpx;
  background: #F4F6F8;
  border: 1rpx solid #EEF1F4;
  font-size: 24rpx;
  color: #475467;
  transition: all 0.18s ease;
}
.chip-active {
  background: linear-gradient(135deg, #0A66C2, #0D7AE0);
  border-color: transparent;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 6rpx 16rpx rgba(10, 102, 194, 0.28);
}

/* 状态面板 */
.state-panel {
  min-height: 480rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 132rpx;
  height: 132rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: linear-gradient(160deg, #EAF3FB, #F0FAF6);
  animation: floaty 3s ease-in-out infinite;
}
@keyframes floaty {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8rpx); }
}
.state-mark-inner {
  width: 92rpx;
  height: 92rpx;
  border-radius: 24rpx;
  background: #fff;
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}
.state-search {
  width: 44rpx;
  height: 44rpx;
  border: 4rpx solid #0A66C2;
  border-radius: 50%;
  position: relative;
}
.state-search::after {
  content: '';
  position: absolute;
  right: -13rpx;
  bottom: -8rpx;
  width: 20rpx;
  height: 4rpx;
  border-radius: 2rpx;
  background: #0A66C2;
  transform: rotate(45deg);
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  margin-top: 36rpx;
  padding: 16rpx 64rpx;
  border-radius: 50rpx;
  background: linear-gradient(135deg, #0A66C2, #0D7AE0);
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.28);
  font-size: 26rpx;
  font-weight: 600;
  color: #fff;
}
.loading-ring {
  width: 64rpx;
  height: 64rpx;
  border: 6rpx solid #E4E7EC;
  border-top-color: #0A66C2;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
  margin-bottom: 24rpx;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* 结果列表 */
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
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  border: 1rpx solid rgba(228, 231, 236, 0.7);
  overflow: hidden;
  animation: cardIn 0.42s cubic-bezier(0.25, 0.46, 0.45, 0.94) both;
}
@keyframes cardIn {
  from { opacity: 0; transform: translateY(22rpx); }
  to { opacity: 1; transform: translateY(0); }
}
.card-hover {
  transform: scale(0.98);
  box-shadow: 0 8px 20px rgba(16, 24, 40, 0.1);
}
.trade-card-main { padding: 24rpx 26rpx; }
.trade-body { min-width: 0; }
.trade-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #17212B;
  font-size: 30rpx;
  line-height: 1.42;
  font-weight: 700;
  margin-top: 14rpx;
}
.trade-desc {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: #667085;
  font-size: 24rpx;
  line-height: 1.55;
  margin-top: 10rpx;
}
.trade-meta {
  display: flex;
  gap: 18rpx;
  margin-top: 14rpx;
  color: #98A2B3;
  font-size: 22rpx;
}
.match-tags { margin-top: 14rpx; display: flex; gap: 10rpx; flex-wrap: wrap; }
.reason-tag {
  color: #0A66C2;
  background: #EAF3FB;
  border-radius: 8rpx;
  padding: 5rpx 12rpx;
  font-size: 20rpx;
}
</style>
