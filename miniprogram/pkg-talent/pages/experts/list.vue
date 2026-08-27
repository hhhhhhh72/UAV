<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="专家智库" show-back :fixed="true" @back="goBack" />

    <!-- ① 白底头部：搜索 + 筛选 -->
    <view class="head-zone">
      <!-- 搜索框（白上白：双层投影浮起） -->
      <view class="sbar">
        <view class="b-search">
          <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
          <input
            class="b-sinp"
            v-model="searchText"
            placeholder="搜索专家姓名或领域"
            placeholder-class="b-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <text v-if="searchText" class="b-sclr" @tap="clearSearch">×</text>
          <view class="b-sep" />
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- 领域一级筛选：下划线 tab 分段（对齐科技成果库；7 类超宽 → 单行横向滚动） -->
      <scroll-view scroll-x :show-scrollbar="false" class="stages-scroll">
        <view class="stage-wrap">
          <view class="stages">
            <view
              v-for="tab in fieldTabs"
              :key="tab.value"
              class="stg"
              :class="{ on: activeField === tab.value }"
              @tap="pickStageTab(tab.value)"
            >
              <text>{{ tab.label }}</text>
            </view>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- ② 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 位专家</text>
      <text class="ir-hint">{{ activeField || '全部领域' }}</text>
    </view>

    <!-- ③ 骨架屏：首次加载 -->
    <view v-if="loading && list.length === 0" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w60"></view></view>
        <view class="sk-bd">
          <view class="sk-l w80"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- ④ 空 / 错误 -->
    <view v-else-if="!loading && list.length === 0" class="st">
      <u-empty :description="errorMsg || '暂无专家信息'">
        <view v-if="errorMsg" class="stb" @tap="fetchList(true)">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑤ 专家列表 -->
    <view v-else class="cl">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="goDetail(item)"
      >
        <view class="cell-name">{{ item.name }}</view>
        <view class="cell-meta">
          <text v-for="(f, fi) in parseFields(item.field)" :key="fi" class="field-tag">{{ f }}</text>
          <text v-if="item.org || item.organization" class="meta-text">{{ item.org || item.organization }}</text>
        </view>
        <view v-if="item.title" class="cell-title">{{ item.title }}</view>
      </view>

      <!-- 加载更多 -->
      <view v-if="list.length > 0" class="load-more">
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
import { request, getErrorMessage } from '../../../utils/request'

export default {
  data() {
    return {
      noMotion: false,
      statusBarHeight: 20,
      searchText: '',
      activeField: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      fieldTabs: [
        { label: '全部', value: '' },
        { label: '低空管控', value: '低空管控' },
        { label: '适航认证', value: '适航认证' },
        { label: '无人机研发', value: '无人机研发' },
        { label: '应急救援', value: '应急救援' },
        { label: '政策法规', value: '政策法规' },
        { label: '投融资', value: '投融资' },
      ],
    }
  },
  onLoad() {
    this.checkMotion()
    this.fetchList(true)
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
    // 减弱动效（无障碍）+ 状态栏高度
    checkMotion() {
      try {
        const sys = uni.getSystemInfoSync()
        if (sys && sys.statusBarHeight) this.statusBarHeight = sys.statusBarHeight
        if (sys && sys.reduceMotion) this.noMotion = true
      } catch (e) {}
      try {
        if (typeof uni.onAccessibilityInfoChange === 'function') {
          uni.onAccessibilityInfoChange((res) => { this.noMotion = !!(res && res.reduceMotion) })
        }
      } catch (e) {}
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
        var params = {}
        if (this.activeField) params.field = this.activeField
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/experts', data: params })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var total = (data && data.total) != null ? data.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
      } catch (e) {
        this.errorMsg = getErrorMessage(e) || '网络异常，请稍后重试'
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    async loadMore() {
      this.page++
      await this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    clearSearch() {
      this.searchText = ''
      this.fetchList(true)
    },
    // 方案 A（对齐成果库）：非「全部」tab 再点取消；「全部」tab 清筛；单维度无 ▾ 面板，已在全部再点不动作
    pickStageTab(value) {
      if (value === '') {
        if (this.activeField === '') return // 已在全部，单维度无面板可开
        this.activeField = ''
      } else {
        this.activeField = this.activeField === value ? '' : value // 再点取消
      }
      this.fetchList(true)
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pkg-talent/pages/experts/detail?id=' + encodeURIComponent(item.id) })
    },
    goBack() {
      uni.navigateBack()
    },
    parseFields(field) {
      if (!field) return []
      if (typeof field === 'string') {
        return field.split(/[,，]/).filter(Boolean)
      }
      if (Array.isArray(field)) return field
      return []
    },
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== 白底头部 ===== */
.head-zone { background: #fff; }

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 双层投影 ===== */
.sbar { padding: 12px 12px 8px; background: #fff; }
.b-search {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05);
  display: flex;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.b-search-ic { position: relative; width: 15px; height: 15px; flex: none; }
.ic-ring {
  width: 9px; height: 9px;
  border: 1.5px solid #98A2B3;
  border-radius: 50%;
  position: absolute; top: 0; left: 0;
}
.ic-bar {
  position: absolute; right: 0; bottom: 1px;
  width: 5px; height: 1.5px;
  background: #98A2B3;
  transform: rotate(45deg);
}
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; }
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== 领域一级筛选：下划线 tab 分段（对齐科技成果库；7 类超宽单行横滑，不换行） ===== */
.stages-scroll { white-space: nowrap; width: 100%; }
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
  animation: toc-in .22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { font-size: 12px; color: #98A2B3; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 专家卡片 ===== */
.cl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px 12px; }
.card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.tap-scale { transform: scale(0.97); opacity: 0.9; }

.cell-name { font-size: 15px; font-weight: 600; color: #17212B; }
.cell-meta { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.field-tag {
  padding: 1px 8px;
  border-radius: 4px;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 11px;
  font-weight: 600;
}
.meta-text { font-size: 12px; color: #667085; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cell-title { font-size: 12px; color: #98A2B3; }

/* 加载更多 */
.load-more { text-align: center; padding: 12px 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 12px; color: #667085; }
.no-more { color: #98A2B3; font-size: 12px; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .stg.on::after { animation: none; } /* 注线画出属位移，关闭 */

/* prefers-reduced-motion：装饰动画/过渡全关（对齐科技成果库） */
@media (prefers-reduced-motion: reduce) {
  .stg, .stg-arr, .p-chip, .field-panel, .panel-mask {
    animation: none !important;
    transition: none !important;
  }
}
</style>
