<template>
  <view class="page">
    <!-- ① 深空蓝 Hero -->
    <view class="hero">
      <view class="hero-deco">
        <view class="deco-grid" />
        <view class="deco-radar" />
        <view class="deco-star s1" />
        <view class="deco-star s2" />
      </view>
      <view class="hero-nav">
        <view class="back-btn" hover-class="press-feedback" :hover-stay-time="120" @click="goBack">
          <text class="back-icon">‹</text>
        </view>
      </view>
      <view class="hero-content">
        <text class="hero-label">应急协同</text>
        <text class="hero-title">调度记录</text>
        <text class="hero-subtitle">事件响应 · 资源调度 · 全程追踪</text>
      </view>
    </view>

    <!-- ② 状态筛选 pills -->
    <view class="filter-row">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-inner">
          <view
            v-for="(tab, i) in tabTitles"
            :key="i"
            class="filter-pill"
            :class="{ on: activeTabIndex === i }"
            @click="onTabChange(i)"
          >{{ tab }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- ③ 列表 -->
    <StateView
      :loading="loading && list.length === 0"
      :error="!!errorMsg && list.length === 0"
      :empty="!loading && !errorMsg && list.length === 0"
      empty-text="暂无调度记录"
      @retry="fetchList(true)"
    >
      <scroll-view class="list-scroll" scroll-y>
        <view class="list">
          <view
            v-for="item in list"
            :key="item.id"
            class="dispatch-card"
            hover-class="press-feedback"
            :hover-stay-time="120"
          >
            <view class="dc-left">
              <view class="dc-dot" :class="'dc-dot--' + statusKey(item.status)" />
            </view>
            <view class="dc-body">
              <view class="dc-top">
                <text class="dc-title">{{ item.event_desc || item.title || '未命名事件' }}</text>
                <text class="dc-status" :class="'dc-status--' + statusKey(item.status)">{{ statusLabel(item.status) }}</text>
              </view>
              <text class="dc-loc">{{ item.location || '位置待定' }}</text>
              <view class="dc-meta">
                <text class="dc-time">{{ formatDateTime(item.start_time || item.created_at) }}</text>
                <text v-if="item.commander" class="dc-commander">负责人 · {{ item.commander }}</text>
              </view>
              <text v-if="item.result" class="dc-result">{{ item.result }}</text>
            </view>
          </view>
        </view>
      </scroll-view>
    </StateView>
  </view>
</template>

<script>
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

export default {
  components: { StateView },
  data() {
    return {
      activeTabIndex: 0,
      loading: false,
      errorMsg: '',
      list: [],
      statusMap: ['', 'pending', 'dispatched', 'completed'],
      tabTitles: ['全部', '待响应', '已调度', '已完成'],
    }
  },
  onLoad() {
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    async fetchList(reset) {
      if (reset) {
        this.loading = true
      }
      this.errorMsg = ''

      try {
        var params = {}
        var statusVal = this.statusMap[this.activeTabIndex]
        if (statusVal) params.status = statusVal

        var res = await request({
          url: '/api/v1/emergency-dispatches',
          data: params,
        })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []

        this.list = items
        // API 返回空数据，降级到本地 mock
        if (this.list.length === 0) {
          this.list = this.getMockDispatches(statusVal)
        }
      } catch (e) {
        // API 不可用，降级到本地 mock
        this.list = this.getMockDispatches(this.statusMap[this.activeTabIndex])
        this.errorMsg = ''
      } finally {
        this.loading = false
      }
    },
    onTabChange(index) {
      this.activeTabIndex = index
      this.fetchList(true)
    },
    /* 本地 mock（API 空/失败时降级展示） */
    getMockDispatches(status) {
      var all = [
        { id: 'dsp-1', event_desc: '南山森林火情应急侦察', status: 'ongoing', location: '重庆市南岸区南山', start_time: '2026-08-03 09:20', commander: '张队', result: '侦察火线分布，已回传指挥部' },
        { id: 'dsp-2', event_desc: '嘉陵江洪水区域巡查', status: 'ongoing', location: '重庆市北碚区嘉陵江段', start_time: '2026-08-02 15:40', commander: '李指挥', result: '巡查12公里堤岸，发现2处隐患' },
        { id: 'dsp-3', event_desc: '渝中区高层建筑外墙隐患检测', status: 'done', location: '重庆市渝中区解放碑', start_time: '2026-08-01 10:00', commander: '王工', result: '完成检测，已出具报告' },
        { id: 'dsp-4', event_desc: '綦江暴雨受困群众搜索', status: 'done', location: '重庆市綦江区东溪镇', start_time: '2026-07-28 08:15', commander: '赵队', result: '定位3名受困群众，配合救援' },
        { id: 'dsp-5', event_desc: '大型活动低空安保监控', status: 'cancelled', location: '重庆市江北区国博中心', start_time: '2026-07-25 14:00', commander: '孙队', result: '活动取消，调度终止' },
      ]
      if (status) {
        return all.filter(function (d) { return d.status === status })
      }
      return all
    },
    goBack() {
      uni.navigateBack()
    },
    /* 状态归一：pending/dispatched/completed/ongoing/done/cancelled */
    statusKey(status) {
      if (status === 'completed' || status === 'done') return 'completed'
      if (status === 'pending') return 'pending'
      if (status === 'ongoing' || status === 'dispatched') return 'ongoing'
      if (status === 'cancelled') return 'cancelled'
      return 'pending'
    },
    statusLabel(status) {
      var map = {
        pending: '待响应',
        dispatched: '已调度',
        completed: '已完成',
        ongoing: '进行中',
        done: '已完成',
        cancelled: '已取消',
      }
      return map[status] || status || '未知'
    },
    formatDateTime(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      var h = d.getHours()
      var min = d.getMinutes()
      return (
        d.getFullYear() +
        '-' +
        (m < 10 ? '0' : '') +
        m +
        '-' +
        (day < 10 ? '0' : '') +
        day +
        ' ' +
        (h < 10 ? '0' : '') +
        h +
        ':' +
        (min < 10 ? '0' : '') +
        min
      )
    },
  },
}
</script>

<style scoped>
.page {
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.25, 0.46, 0.45, 0.94);
  min-height: 100vh;
  background: linear-gradient(180deg, #f5f6f8 0%, #E8F2FC 100%);
  padding-bottom: env(safe-area-inset-bottom);
}

/* ================================================================= */
/* ① Hero                                                             */
/* ================================================================= */
.hero {
  background: linear-gradient(135deg, #074D92 0%, #0A66C2 100%);
  padding: 88rpx 32rpx 56rpx;
  position: relative;
  overflow: hidden;
}

.hero-deco { position: absolute; inset: 0; pointer-events: none; }

.deco-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.12) 2rpx, transparent 2rpx);
  background-size: 40rpx 40rpx;
  opacity: 0.6;
}

.deco-radar {
  position: absolute;
  right: -80rpx;
  top: -80rpx;
  width: 300rpx;
  height: 300rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.15);
  border-radius: 50%;
}

.deco-radar::before,
.deco-radar::after {
  content: '';
  position: absolute;
  inset: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

.deco-radar::after { inset: 90rpx; border: 2rpx solid rgba(255, 255, 255, 0.08); }

.deco-star {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  animation: twinkle 2.5s ease-in-out infinite;
}
.s1 { left: 60rpx; top: 80rpx; width: 6rpx; height: 6rpx; animation-delay: 0s; }
.s2 { left: 200rpx; top: 120rpx; width: 8rpx; height: 8rpx; animation-delay: 0.6s; }

.hero-nav {
  margin-bottom: 24rpx;
  position: relative;
  z-index: 2;
}

.back-btn {
  width: 88rpx;
  height: 88rpx;
  background: rgba(255, 255, 255, 0.15);
  border: 1rpx solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-icon { color: #ffffff; font-size: 44rpx; font-weight: 300; }

.hero-content { position: relative; z-index: 2; animation: pageIn var(--anim-slow) var(--ease-out) both; }

.hero-label {
  color: rgba(255, 255, 255, 0.75);
  font-size: 24rpx;
  display: block;
  margin-bottom: 8rpx;
}

.hero-title {
  color: #ffffff;
  font-size: 48rpx;
  font-weight: 700;
  line-height: 1.2;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.3);
  display: block;
  margin-bottom: 12rpx;
}

.hero-subtitle {
  color: rgba(255, 255, 255, 0.75);
  font-size: 24rpx;
  line-height: 1.5;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

/* ================================================================= */
/* ② 筛选 pills                                                       */
/* ================================================================= */
.filter-row {
  padding: 20rpx 24rpx 8rpx;
}

.filter-scroll {
  white-space: nowrap;
  width: 100%;
}

.filter-inner {
  display: inline-flex;
  gap: 12rpx;
}

.filter-pill {
  padding: 10rpx 28rpx;
  border-radius: 999rpx;
  border: 1rpx solid #ebedf0;
  background: #ffffff;
  color: #969799;
  font-size: 26rpx;
  flex-shrink: 0;
  transition: background-color var(--anim-fast) ease, color var(--anim-fast) ease, border-color var(--anim-fast) ease;
}

.filter-pill.on {
  border-color: #0A66C2;
  background: #0A66C2;
  color: #ffffff;
  font-weight: 500;
}

/* ================================================================= */
/* ③ 列表（时间线卡片）                                               */
/* ================================================================= */
.list-scroll { height: calc(100vh - 400rpx); }

.list { padding: 8rpx 24rpx 24rpx; }

.dispatch-card {
  display: flex;
  gap: 16rpx;
  background: #ffffff;
  border: 1rpx solid #f0f1f3;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
  animation: cardIn var(--anim-base) var(--ease-out) both;
  transition: transform var(--anim-fast) ease;
}

.dc-left {
  width: 24rpx;
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  padding-top: 8rpx;
}

.dc-dot {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
}

.dc-dot--pending { background: #F97316; }
.dc-dot--ongoing { background: #0A66C2; }
.dc-dot--completed { background: #34c759; }
.dc-dot--cancelled { background: #98A2B3; }

.dc-body { flex: 1; min-width: 0; }

.dc-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8rpx;
  margin-bottom: 8rpx;
}

.dc-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #17212B;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dc-status {
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.dc-status--pending { background: #FFF4E6; color: #E96012; }
.dc-status--ongoing { background: #E8F2FC; color: #0A66C2; }
.dc-status--completed { background: #E8F5E9; color: #34c759; }
.dc-status--cancelled { background: #F5F5F5; color: #98A2B3; }

.dc-loc {
  font-size: 26rpx;
  color: #969799;
  display: block;
  margin-bottom: 8rpx;
}

.dc-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-bottom: 8rpx;
}

.dc-time { font-size: 22rpx; color: #98A2B3; }
.dc-commander { font-size: 22rpx; color: #98A2B3; }

.dc-result {
  font-size: 24rpx;
  color: #17212B;
  line-height: 1.5;
  background: #fafafa;
  border-radius: 8rpx;
  padding: 12rpx 16rpx;
}

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes cardIn {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.2; }
  50%      { opacity: 0.8; }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .hero-content, .dispatch-card {
    animation: none !important;
    transition: none !important;
  }
}
</style>
