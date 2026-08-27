<template>
  <view class="demand-list-page" :class="{ 'no-motion': noMotion }">
    <!-- 搜索 -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索需求、项目"
        @search="onSearch"
      />
    </u-sticky>

    <!-- 筛选：业务类型一级分段（「全部」带 ▾ 独立开关；对齐科技成果库）+ ▾ 排序面板 -->
    <view class="stage-wrap">
      <view class="stages">
        <view
          v-for="(tab, index) in bizTypeTabs"
          :key="tab.value"
          class="stg"
          :class="{ on: activeBizType === tab.value }"
          @tap="pickStageTab(tab.value)"
        >
          <text>{{ tab.label }}</text>
          <!-- ▾ 独立面板开关：未停在「全部」时点「全部」先清类型；停在「全部」时再点开面板 -->
          <text v-if="tab.value === ''" class="stg-arr" :class="{ up: panel === 'all' }" @tap.stop="togglePanel">▾</text>
        </view>
      </view>
      <!-- 二级筛选面板：排序 chips（absolute 浮层，展开不挤动下方内容） -->
      <view v-if="panel === 'all'" class="field-panel" :class="{ closing }">
        <view class="p-group">排序方式</view>
        <view class="p-chips">
          <text v-for="a in sortActions" :key="a.value" class="p-chip" :class="{ act: currentSort === a.value }" @tap="pickSort(a)">{{ a.name }}</text>
        </view>
      </view>
    </view>
    <!-- 蒙层：从分段底部开始置灰，点外部退场收起 -->
    <view v-if="panel" class="panel-mask" :style="{ top: maskTop + 'px' }" @tap="startClosePanel" />

    <!-- 加载状态 -->
    <view v-if="loading && list.length === 0" class="state-wrap">
      <u-loading size="28rpx" />
      <text class="state-text">加载中...</text>
    </view>

    <!-- 筛选无结果 -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg && (activeBizType || searchText)" class="state-wrap">
      <u-empty description="没有符合条件的需求" />
      <view class="state-reset" @tap="resetFilter">清除筛选</view>
    </view>

    <!-- 空数据 -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="state-wrap">
      <u-empty description="暂无需求" />
    </view>

    <!-- 错误状态 -->
    <view v-else-if="errorMsg && list.length === 0" class="state-wrap">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)"><text>重新加载</text></view>
    </view>

    <!-- 列表 -->
    <view v-else class="list-body">
      <!-- 首条重点卡 -->
      <view class="featured-card" @tap="goDetail(list[0])">
        <image :src="featuredImage(list[0])" mode="aspectFill" class="featured-img" />
        <view class="featured-mask"></view>
        <view class="featured-copy">
          <view class="featured-tags">
            <text class="tag-blue">{{ bizTypeLabel(list[0].biz_type) }}</text>
            <text v-if="list[0].district" class="tag-white">{{ list[0].district }}</text>
          </view>
          <text class="featured-title">{{ list[0].title }}</text>
          <text class="featured-meta">{{ formatBudget(list[0].budget_fen) }} · {{ formatDate(list[0].created_at) }}</text>
        </view>
      </view>

      <!-- 紧凑卡 -->
      <view
        v-for="item in list.slice(1)"
        :key="item.id"
        class="compact-card"
        @tap="goDetail(item)"
      >
        <text class="compact-title">{{ item.title }}</text>
        <view class="compact-meta">
          <text class="tag-blue tag-mini">{{ bizTypeLabel(item.biz_type) }}</text>
          <text v-if="item.district" class="meta-text">{{ item.district }}</text>
          <text class="meta-text">{{ formatBudget(item.budget_fen) }}</text>
          <text class="meta-date">{{ formatDate(item.created_at) }}</text>
        </view>
      </view>

      <!-- 加载更多 -->
      <view class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="24rpx" />
          <text>加载更多...</text>
        </view>
        <text v-else-if="!hasMore" class="no-more">没有更多了</text>
      </view>
    </view>
  </view>
</template>

<script>
import { request, getStoredUser, BASE_URL } from '../../utils/request'
import { BIZ_TYPE_TABS, bizTypeLabel as bizTypeLabelOf } from '../../utils/enums'

// 筛选面板退场定时器（模块级，非响应式）
let panelCloseT = null
const PANEL_CLOSE_MS = 210 // 退场动画 .21s ease-in

export default {
  data() {
    return {
      searchText: '',
      activeBizType: '',
      currentSort: 'newest',
      noMotion: false, // 减弱动效（无障碍）：Options API 直存，避免 setup() 混合触发微信端 props 解析异常
      panel: '', // '' = 收起；'all' = 「全部」段的面板展开
      closing: false, // 面板退场中（先播退场动画再 v-if 移除）
      maskTop: 0, // 蒙层起点（面板打开时实测：tab 分段底部）
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      bizTypeTabs: BIZ_TYPE_TABS,
      sortActions: [
        { name: '最新发布', value: 'newest' },
        { name: '预算最高', value: 'budget_desc' },
        { name: '预算最低', value: 'budget_asc' },
      ],
    }
  },
  onLoad() {
    this.checkMotion() // 减弱动效检测（无障碍）
    this.fetchList(true)
  },
  onReady() {
    // 实测蒙层起点（tab 分段底部；面板打开时 togglePanel 会再实测一遍）
    this.measureMaskTop()
  },
  onPullDownRefresh() {
    this.fetchList(true).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  onReachBottom() {
    if (!this.loadingMore && this.hasMore) {
      this.loadMore()
    }
  },
  methods: {
    // 减弱动效检测（无障碍）：逻辑同 utils/motion.js，Options API 直实现（避免 setup() 混合）
    checkMotion() {
      try {
        const sys = uni.getSystemInfoSync()
        if (sys && sys.reduceMotion) this.noMotion = true
      } catch (e) { /* 忽略 */ }
      try {
        if (typeof uni.onAccessibilityInfoChange === 'function') {
          uni.onAccessibilityInfoChange((res) => { this.noMotion = !!(res && res.reduceMotion) })
        }
      } catch (e) { /* 旧基础库无此 API */ }
    },
    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''

      try {
        const params = {}
        if (this.activeBizType) params.biz_type = this.activeBizType
        if (this.currentSort) params.sort = this.currentSort
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        // 预算排序由前端本地完成（后端忽略 sort 参数）：放大每页至 100 条，
        // 保证排序窗口覆盖大部分数据（小分页下跨页排序不完整）。
        params.page_size = this.currentSort && this.currentSort.indexOf('budget') === 0 ? 100 : this.pageSize

        const res = await request({ url: '/api/v1/demands', data: params })
        const data = Array.isArray(res) ? res : (res && res.data) || res || {}
        const items = Array.isArray(data) ? data : (data && data.items) || []
        const total = (data && data.total) != null ? data.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
        // 后端忽略 sort 参数：预算排序在前端本地完成（缺失 budget_fen 视为 0）
        this.applySort()
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    applySort() {
      if (this.currentSort === 'budget_desc') {
        this.list.sort(function (a, b) {
          return (Number(b.budget_fen) || 0) - (Number(a.budget_fen) || 0)
        })
      } else if (this.currentSort === 'budget_asc') {
        this.list.sort(function (a, b) {
          return (Number(a.budget_fen) || 0) - (Number(b.budget_fen) || 0)
        })
      }
    },
    async loadMore() {
      this.page++
      await this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    // ---- 筛选交互（对齐科技成果库方案 A：tab 分段 + ▾ 面板 + 蒙层） ----
    startClosePanel() {
      if (this.closing) return // 已在退场中，防重复触发叠加定时器
      this.closing = true
      clearTimeout(panelCloseT)
      panelCloseT = setTimeout(() => { this.panel = ''; this.closing = false; panelCloseT = null }, PANEL_CLOSE_MS)
    },
    togglePanel() {
      if (this.panel === 'all') { this.startClosePanel(); return } // 再点「全部」→ 退场收起
      clearTimeout(panelCloseT); panelCloseT = null; this.closing = false
      this.panel = 'all'
      uni.nextTick(() => { this.measureMaskTop() })
    },
    measureMaskTop() {
      // 蒙层起点 = 分段容器底部（实测，头部不蒙）
      uni.createSelectorQuery().select('.stage-wrap').boundingClientRect((rect) => {
        if (rect && rect.bottom) this.maskTop = Math.round(rect.bottom)
      }).exec()
    },
    // 方案 A：非「全部」tab 再点取消回全部；「全部」未停先清筛、停下再点开面板；▾ 独立开关
    pickStageTab(value) {
      if (value !== '') {
        this.startClosePanel()
        this.activeBizType = this.activeBizType === value ? '' : value
        this.fetchList(true)
        return
      }
      if (this.activeBizType !== '') {
        this.startClosePanel()
        this.activeBizType = ''
        this.fetchList(true)
        return
      }
      this.togglePanel()
    },
    pickSort(action) {
      this.currentSort = action.value
      this.startClosePanel()
      this.fetchList(true)
    },
    resetFilter() {
      this.activeBizType = ''
      this.searchText = ''
      this.startClosePanel()
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(item.id) })
    },
    bizTypeLabel(type) {
      return bizTypeLabelOf(type)
    },
    featuredImage(item) {
      try {
        const arr = typeof item.images === 'string' ? JSON.parse(item.images) : item.images
        // 存库为相对路径 /uploads/xxx，预览必须补全域名，否则小程序按本地包内资源加载 → 白图
        if (Array.isArray(arr) && arr[0]) return this.resolveUrl(arr[0])
      } catch {}
      return '/static/home/hero-inspection.jpg'
    },
    resolveUrl(u) {
      if (!u) return ''
      if (u.indexOf('http') === 0) return u
      return BASE_URL + u
    },
    formatBudget(fen) {
      if (fen == null || fen === 0) return '面议'
      var yuan = (fen / 100).toFixed(2)
      return yuan.replace(/\.00$/, '') + ' 元'
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
  },
}
</script>

<style scoped>
.demand-list-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 一级筛选：下划线 tab 分段（对齐科技成果库）+ ▾ 浮层面板 + 蒙层 */
.stage-wrap { position: relative; z-index: 42; background: #fff; }
.stages { display: flex; gap: 40rpx; padding: 4rpx 28rpx 16rpx; white-space: nowrap; }
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 0 8rpx;
  font-size: 24rpx;
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; }
.stg.on::after {
  content: '';
  position: absolute;
  left: 8rpx;
  right: 8rpx;
  bottom: 16rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #074D92;
  animation: toc-in 0.22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.stg-arr {
  font-size: 24rpx;
  color: #667085;
  transition: transform 0.2s ease, color 0.2s ease;
  padding: 20rpx 16rpx;
  margin: -20rpx -16rpx;
}
.stg-arr.up { transform: rotate(180deg); color: #074D92; }
.field-panel {
  position: absolute;
  left: 0;
  right: 0;
  top: 100%;
  z-index: 43;
  background: #fff;
  border-radius: 0 0 12px 12px;
  box-shadow: 0 12px 24px rgba(16, 24, 40, 0.08);
  padding: 12px 14px 14px;
  max-height: 62vh;
  overflow-y: auto;
  animation: panelIn 0.3s cubic-bezier(0.32, 0.72, 0, 1);
}
.field-panel.closing { animation: panelOut 0.21s ease-in forwards; }
@keyframes panelOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-10px); }
}
@keyframes panelIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
.field-panel .p-group { font-size: 13px; font-weight: 700; color: #344054; margin: 12px 0 6px; }
.field-panel .p-group:first-child { margin-top: 0; }
.p-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.p-chip {
  min-height: 40px;
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
}
.p-chip.act { color: #fff; border-color: #074D92; background: #074D92; font-weight: 600; }
.p-chip { transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease, transform 0.3s cubic-bezier(0.34, 1.8, 0.64, 1); }
.p-chip:active { transform: scale(0.94); transition: transform 0.08s linear; }
.p-chip.act { animation: chipPop 0.3s cubic-bezier(0.34, 1.8, 0.64, 1); }
@keyframes chipPop { 0% { transform: scale(1); } 40% { transform: scale(0.94); } 100% { transform: scale(1); } }
/* 蒙层：从分段底部开始置灰（top 由 maskTop 实测）；u-sticky 吸顶 z-index 99，蒙层降到其下（同训练课程页先例） */
.panel-mask {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 15;
  background: rgba(16, 24, 40, 0.2);
  animation: maskIn 0.22s ease-out;
}
@keyframes maskIn { from { opacity: 0; } to { opacity: 1; } }
/* 减弱动效（无障碍）：装饰动画/位移缩放关闭，保留淡入与颜色反馈 */
.demand-list-page.no-motion .stg-arr { transition: none; }
.demand-list-page.no-motion .p-chip { transition: none; }
.demand-list-page.no-motion .p-chip.act { animation: none; }
.demand-list-page.no-motion .stg.on::after { animation: none; }
.demand-list-page.no-motion .field-panel { animation: panelIn 0.3s ease-out; }
.demand-list-page.no-motion .field-panel.closing { animation: panelOut 0.16s ease-in forwards; }
.demand-list-page.no-motion .panel-mask { animation: maskIn 0.22s ease-out; }
.demand-list-page.no-motion .p-chip:active { transform: none; }

/* 状态 */
.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 64px 0;
}

.state-text {
  font-size: 13px;
  color: #667085;
}

.state-reset {
  padding: 8px 24px;
  border-radius: 8px;
  border: 1px solid #0A66C2;
  color: #0A66C2;
  font-size: 13px;
}

.retry-btn {
  padding: 8px 24px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 13px;
}

/* 列表 */
.list-body {
  padding: 12px;
}

/* 首条重点卡 */
.featured-card {
  position: relative;
  height: 276rpx;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 8px;
}

.featured-img {
  width: 100%;
  height: 100%;
}

.featured-mask {
  position: absolute;
  inset: 0;
  background: rgba(16, 24, 40, 0.45);
}

.featured-copy {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.featured-tags {
  display: flex;
  gap: 6px;
}

.tag-blue {
  font-size: 11px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 2px 8px;
  border-radius: 4px;
}

.tag-white {
  font-size: 11px;
  color: #fff;
  background: rgba(255, 255, 255, 0.22);
  padding: 2px 8px;
  border-radius: 4px;
}

.featured-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  line-height: 1.35;
}

.featured-meta {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.85);
}

/* 紧凑卡 */
.compact-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.compact-title {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
  line-height: 1.4;
  display: block;
}

.compact-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.tag-mini {
  font-size: 10px;
}

.meta-text {
  font-size: 12px;
  color: #667085;
}

.meta-date {
  font-size: 11px;
  color: #98A2B3;
  margin-left: auto;
}

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 16px 0;
}

.loading-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #667085;
}

.no-more {
  color: #98A2B3;
  font-size: 12px;
}

/* 排序弹层已移除（排序维度移入 ▾ field-panel），此段样式随 u-popup 一并删除 */

@media (prefers-reduced-motion: reduce) {
  .stg, .stg-arr, .p-chip, .field-panel, .panel-mask {
    animation: none !important;
    transition: none !important;
  }
}
</style>
