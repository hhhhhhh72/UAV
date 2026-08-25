<template>
  <!--
  IMPECCABLE-DIRECTION v1 · seed=user-pinned · 行业报告（蓝皮书刊物）
  THESIS: 打开的是一份正式协会刊物，不是一条数据记录；拒绝卡片流+下划线筛选的通用列表。
  OWN-WORLD: 蓝皮书刊物世界：深蓝刊头版（渐变 #0A3A6B→#074D92）、CSS 排版封面（白皮书深蓝/调研青绿/分析橙/其他墨灰）、宋体刊感标题、细线书脊、总目式筛选；品牌令牌 #0A66C2/#1DD4A8/#F97316。
  STORY: 读者按类型寻刊 → 打开刊物 → 读完正文 → 下载 PDF。
  FIRST VIEWPORT: 深蓝刊头（状态栏+返回+协会刊物+搜索）→ 总目筛选行 → 两列封面网格。
  FORM: 用户指定方向 B（蓝皮书刊物），真机对比拍板；封面纯 CSS 排版生成，不虚构图片字段；数据契约收敛到后端真实枚举 category/created_at。
  FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
  -->
  <view class="page">
    <!-- ===== 深蓝刊头 ===== -->
    <view class="head" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="nav">
        <image :src="ICON_BACK" class="nav-back" hover-class="icon-press" mode="aspectFit" aria-role="button" aria-label="返回" @tap="goBack" />
        <text class="nav-title">行业报告</text>
      </view>
      <view class="brand-row">
        <text class="brand-big">协会刊物</text>
        <text class="brand-sub">重庆市无人机产业协会 · 权威发布</text>
      </view>
      <view class="head-rule"></view>
      <view class="search">
        <image :src="ICON_SEARCH_WHITE" class="search-ico" mode="aspectFit" />
        <input
          v-model="searchText"
          class="search-input"
          placeholder="搜索刊物"
          placeholder-style="color:rgba(255,255,255,0.75)"
          confirm-type="search"
          maxlength="50"
          @confirm="onSearch"
        />
        <image
          v-if="searchText"
          :src="ICON_X_WHITE"
          class="search-clear"
          hover-class="icon-press"
          mode="aspectFit"
          aria-role="button"
          aria-label="清除搜索"
          @tap="clearSearch"
        />
      </view>
    </view>

    <!-- ===== 总目式筛选（刊物目录语汇，替代下划线 Tab） ===== -->
    <scroll-view scroll-x class="toc" :show-scrollbar="false">
      <view class="toc-row">
        <block v-for="(t, i) in REPORT_TYPES" :key="t.value">
          <view v-if="i > 0" class="toc-sep"></view>
          <view
            class="toc-item"
            :class="{ on: activeType === t.value }"
            hover-class="toc-press"
            @tap="onTypeChange(t.value)"
          >{{ t.label }}</view>
        </block>
      </view>
    </scroll-view>

    <!-- ===== 加载中：骨架屏，形状与封面网格一致 ===== -->
    <view v-if="loading && list.length === 0" class="grid">
      <view v-for="i in 4" :key="'sk' + i" class="cell">
        <view class="sk-cover"></view>
        <view class="sk-line sk-line-1"></view>
        <view class="sk-line sk-line-2"></view>
      </view>
    </view>

    <!-- ===== 加载失败 ===== -->
    <view v-else-if="errorMsg && list.length === 0" class="state">
      <text class="state-title">加载失败</text>
      <text class="state-desc">{{ errorMsg }}</text>
      <view class="retry-btn" hover-class="btn-press" aria-role="button" aria-label="重新加载" @tap="fetchList(true)">重新加载</view>
    </view>

    <!-- ===== 空态：空白刊物（按筛选/搜索上下文区分文案与动作） ===== -->
    <view v-else-if="!loading && list.length === 0" class="state">
      <view class="blank-cover">
        <text class="blank-cover-text">刊</text>
      </view>
      <template v-if="activeType || searchText.trim()">
        <text class="state-title">未找到相关刊物</text>
        <text class="state-desc">{{ emptyDesc }}</text>
        <view class="retry-btn" hover-class="btn-press" aria-role="button" :aria-label="clearBtnText" @tap="clearFilters">{{ clearBtnText }}</view>
      </template>
      <template v-else>
        <text class="state-title">暂无报告</text>
        <text class="state-desc">协会刊物正在编撰中，敬请期待</text>
      </template>
    </view>

    <!-- ===== 封面网格 ===== -->
    <view v-else class="grid">
      <!-- 筛选/搜索重载中：半透明遮罩，防止旧数据被误点 -->
      <view v-if="loading" class="grid-mask">
        <text class="grid-mask-text">正在整理刊架…</text>
      </view>
      <view
        v-for="(item, i) in list"
        :key="item.id"
        class="cell"
        :class="{ 'cell-enter': !gridAnimated }"
        :style="gridAnimated ? {} : { animationDelay: cellDelay(i) + 'ms' }"
        hover-class="cell-press"
        aria-role="button"
        :aria-label="'打开报告《' + (item.title || '') + '》'"
        @tap="openDetail(item)"
      >
        <view class="cover" :class="typeOf(item.category).cover">
          <text class="c-brand">重庆市无人机产业协会</text>
          <view class="c-mid">
            <view class="c-line"></view>
            <text class="c-type">{{ typeOf(item.category).label }}</text>
            <view class="c-line"></view>
            <text v-if="typeOf(item.category).en" class="c-en">{{ typeOf(item.category).en }}</text>
          </view>
          <view class="c-foot">
            <text class="c-title">{{ item.title || '-' }}</text>
            <text v-if="item.period" class="c-period">报告期 {{ item.period }}</text>
          </view>
        </view>
        <view class="cap">
          <text class="cap-d">刊发于 {{ formatDate(item.created_at) }}</text>
        </view>
      </view>
    </view>

    <!-- ===== 加载更多 ===== -->
    <view v-if="list.length > 0" class="load-more">
      <view v-if="loadingMore" class="loading-inline">
        <u-loading size="24rpx" color="#0A66C2" />
        <text>加载更多…</text>
      </view>
      <view v-else-if="!hasMore" class="no-more">
        <view class="no-more-line"></view>
        <text>已展示全部刊物</text>
        <view class="no-more-line"></view>
      </view>
      <view v-else class="load-more-btn" hover-class="btn-press" aria-role="button" aria-label="加载更多" @tap="loadMore">点击加载更多</view>
    </view>

    <!-- ===== 演示数据横幅（仅开发环境；开关见 utils/config.js FORCE_MOCK_REPORTS） ===== -->
    <view v-if="mockMode && isDev" class="mock-note">当前为演示数据</view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'
import { FORCE_MOCK_REPORTS } from '../../../utils/config'
import { MOCK_REPORTS } from '../../../utils/mockReports'
import { REPORT_TYPES, typeOf, formatDate, setReportCache, ICON_BACK, ICON_SEARCH_WHITE, svgUri } from './report-common.js'

const ICON_X_WHITE = svgUri('M18 6 6 18 M6 6l12 12', 'B9CBE8')

export default {
  data() {
    return {
      ICON_BACK,
      ICON_SEARCH_WHITE,
      ICON_X_WHITE,
      REPORT_TYPES,
      statusBarHeight: 24,
      searchText: '',
      activeType: '',
      mockMode: process.env.NODE_ENV === 'development' && FORCE_MOCK_REPORTS,
      isDev: process.env.NODE_ENV === 'development',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      // 封面入场动画仅首屏播一次：筛选/搜索/下拉刷新不重播（见 markGridAnimated）
      gridAnimated: false,
    }
  },
  computed: {
    // 空态文案如实说明当前激活的筛选/搜索条件（不静默吞掉用户语境）
    emptyDesc() {
      var kw = this.searchText.trim()
      var t = this.typeOf(this.activeType).label
      if (this.activeType && kw) return '当前「' + t + '」下搜索「' + kw + '」暂无匹配'
      if (this.activeType) return '当前「' + t + '」分类下暂无匹配的刊物'
      return '搜索「' + kw + '」暂无匹配的刊物'
    },
    clearBtnText() {
      var hasType = !!this.activeType
      var hasKw = !!this.searchText.trim()
      if (hasType && hasKw) return '清除筛选与搜索'
      if (hasType) return '清除筛选'
      return '清除搜索'
    },
  },
  onLoad() {
    // 真机状态栏高（编译期 --status-bar-height 恒为 25px，高状态栏设备会压住导航标题）
    this.statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 24
    this.fetchList(true)
  },
  onShareAppMessage() {
    return {
      title: '行业报告 · 重庆市无人机产业协会',
      path: '/pkg-service/pages/reports/list',
    }
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
    typeOf,
    formatDate,
    async fetchList(reset) {
      // 请求序号：过滤/搜索快速连点时丢弃过期响应，防旧结果覆盖新筛选（后端模式竞态守卫）
      var seq = (this._reqSeq = (this._reqSeq || 0) + 1)
      // 演示开关（utils/config.js FORCE_MOCK_REPORTS）：不看后端，直接本地过滤展示 mock
      if (this.mockMode) {
        this.loading = false
        this.loadingMore = false
        this.applyMockFilter()
        return
      }
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''

      try {
        var params = {
          page: this.page,
          page_size: this.pageSize,
        }
        // 后端契约：category 枚举筛选 + keyword 搜索（旧代码传 type/q，筛选与搜索均不生效）
        if (this.activeType) params.category = this.activeType
        if (this.searchText.trim()) params.keyword = this.searchText.trim()

        var res = await request({ url: '/api/v1/industry-reports', data: params })
        if (seq !== this._reqSeq) return // 过期响应：已被更新的筛选/搜索取代
        var items = Array.isArray(res) ? res : []
        var total = res && res.total != null ? res.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
        this.markGridAnimated()
      } catch (e) {
        if (seq !== this._reqSeq) return // 过期请求的失败同样忽略
        this.errorMsg = '网络异常，请稍后重试'
        if (this.list.length > 0) {
          if (this.mockMode) {
            // 当前是演示数据：后端不可用，本地按 category/keyword 过滤
            this.applyMockFilter()
          } else {
            // 已有真实数据时刷新失败：保留旧数据，但 toast 明示失败，不吞错误
            uni.showToast({ title: '刷新失败，请稍后重试', icon: 'none' })
          }
        } else if (this.isDev && MOCK_REPORTS && MOCK_REPORTS.length) {
          // 从未加载成功：仅开发环境回退演示数据（生产走下方「加载失败」空态，不展示 mock）
          this.mockMode = true
          this.applyMockFilter()
        }
        // else：非开发环境或无可回退数据 → errorMsg 生效，走「加载失败」空态
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    async loadMore() {
      this.page++
      await this.fetchList(false)
    },
    // mock 回退：按后端语义本地过滤（category 精确 + keyword 标题包含），15 条 < pageSize 20，分页退化为全部展示
    applyMockFilter() {
      var kw = this.searchText.trim()
      var items = MOCK_REPORTS.filter((r) => {
        if (this.activeType && r.category !== this.activeType) return false
        if (kw && r.title.indexOf(kw) < 0) return false
        return true
      })
      this.list = items
      this.hasMore = false
      this.errorMsg = ''
      this.markGridAnimated()
    },
    // 封面入场仅首屏播一次：延迟翻转开关，使筛/搜/翻页的后续渲染不再带 cell-enter
    markGridAnimated() {
      if (this.gridAnimated) return
      var self = this
      setTimeout(function () { self.gridAnimated = true }, 500)
    },
    // 封面错峰：前 8 个按 20ms 递增，其后不再递增（首屏全可见 8 卡，避免深链分页无限延迟）
    cellDelay(i) {
      return Math.min(i, 7) * 20
    },
    onSearch() {
      this.fetchList(true)
    },
    clearSearch() {
      this.searchText = ''
      this.fetchList(true)
    },
    clearFilters() {
      this.activeType = ''
      this.searchText = ''
      this.fetchList(true)
    },
    onTypeChange(value) {
      if (this.activeType === value) return
      this.activeType = value
      this.fetchList(true)
    },
    openDetail(item) {
      // 决策②：无公开按 ID 详情接口，导航前缓存完整对象（含正文），详情页按 id 读取
      setReportCache(item)
      uni.navigateTo({ url: '/pkg-service/pages/reports/detail?id=' + item.id })
    },
    goBack() {
      // 深链/分享直达时可能无页面栈：兜底回首页，避免 navigateBack 静默失败
      if (getCurrentPages().length > 1) {
        uni.navigateBack()
      } else {
        uni.reLaunch({ url: '/pages/home/index' })
      }
    },
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: var(--color-bg);
  /* 末端收口 40rpx：32/48 均越 4rpx 容差，保留字面量 */
  padding-bottom: 40rpx;
  /* 次级文字对比度校准：全局 --color-text-secondary(#969799≈2.9:1) 不达 AA，
     本页局部覆盖为 #6b6e73(≈4.8:1)；全局令牌统一修复需团队确认后改 App.vue */
  --color-text-secondary: #6b6e73;
}

/* ===== 深蓝刊头 ===== */
.head {
  background: linear-gradient(160deg, #0a3a6b 0%, #074d92 100%);
  /* 顶部安全区由 JS 按真机状态栏高经 :style paddingTop 注入；CSS 顶部恒 0，勿回写 var(--status-bar-height) */
  padding: 0 var(--space-lg) var(--space-lg);
  position: relative;
  overflow: hidden;
}
.nav {
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}
.nav-back {
  position: absolute;
  left: 0;
  width: 44rpx;
  height: 44rpx;
  /* 触达热区：视觉 44rpx，盒模型含内边距 88rpx */
  padding: 22rpx;
  margin-left: -22rpx;
}
.nav-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #ffffff;
}
.brand-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-sm);
  padding: var(--space-xs) 0 var(--space-md);
}
.brand-big {
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 44rpx;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 4rpx;
}
.brand-sub {
  font-size: 21rpx;
  color: rgba(255, 255, 255, 0.78);
  letter-spacing: 2rpx;
}
.head-rule {
  height: 2rpx;
  background: rgba(255, 255, 255, 0.28);
  margin-bottom: var(--space-md);
  position: relative;
}
.head-rule::after {
  content: '';
  position: absolute;
  left: 0;
  top: -5rpx;
  width: 10rpx;
  height: 10rpx;
  background: #ffffff;
  border-radius: 50%;
}
.search {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  height: 68rpx;
  padding: 0 var(--space-md);
  background: rgba(255, 255, 255, 0.12);
  border: 1rpx solid rgba(255, 255, 255, 0.25);
  border-radius: 12rpx;
}
.search-ico {
  width: 26rpx;
  height: 26rpx;
  flex-shrink: 0;
}
.search-input {
  flex: 1;
  min-width: 0;
  /* 撑满 68rpx 可见行高：原生 input 自身热区不足，拉伸后整行可聚焦 */
  height: 68rpx;
  font-size: 25rpx;
  color: #ffffff;
}
.search-clear {
  width: 28rpx;
  height: 28rpx;
  flex-shrink: 0;
  /* 触达热区：视觉 28rpx，盒模型含内边距 88rpx */
  padding: 30rpx;
  margin: -30rpx;
}

/* ===== 总目式筛选 ===== */
.toc {
  background: var(--color-bg);
  white-space: nowrap;
}
.toc-row {
  display: inline-flex;
  align-items: center;
  padding: 0 var(--space-lg) var(--space-lg);
}
.toc-sep {
  width: 1rpx;
  height: 24rpx;
  background: #dde0e4;
  margin: 0 var(--space-md);
}
.toc-item {
  font-size: 26rpx;
  color: #666;
  /* 总目项自带垂直热区：字高约 39rpx + 上下 24rpx ≈ 87rpx 触达盒（微信 88rpx 规范），hover 反馈随项盒 */
  padding: var(--space-md) 0;
  position: relative;
}
.toc-item.on {
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-weight: 700;
  color: #074d92;
  letter-spacing: 2rpx;
}
/* 激活标注：总目细线自中心展开（微动效一次 0.22s），替代下划线 Tab 的廉价底边线 */
.toc-item.on::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 0;
  width: 44rpx;
  height: 3rpx;
  margin-left: -22rpx;
  border-radius: 2rpx;
  background: #074d92;
  transform: scaleX(0);
  animation: rpt-toc-rule 0.22s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}
@keyframes rpt-toc-rule {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}

/* ===== 状态 ===== */
.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  /* 120/60/80 为视口居中构图值，非节奏间距，保留字面量 */
  padding: 120rpx 60rpx 80rpx;
}
.state-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--color-text);
  margin-top: var(--space-md);
}
.state-desc {
  margin-top: var(--space-sm);
  font-size: 24rpx;
  color: var(--color-text-secondary);
  text-align: center;
}
/* 按钮圆角标尺（模块规则）：统一 38rpx，介于卡片 20rpx 与全局 pill(50rpx) 之间，适配刊物方正美学；
   按全局约定补齐 box-shadow */
.retry-btn {
  margin-top: var(--space-lg);
  height: 64rpx;
  line-height: 64rpx;
  padding: 0 56rpx;
  border-radius: 38rpx;
  background: #074d92;
  color: #ffffff;
  font-size: 26rpx;
  box-shadow: 0 6rpx 16rpx rgba(7, 77, 146, 0.25);
}
.blank-cover {
  width: 132rpx;
  height: 180rpx;
  border: 2rpx dashed #c3cad6;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
}
.blank-cover-text {
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 48rpx;
  color: #c3cad6;
}

/* ===== 封面网格 ===== */
.grid {
  display: flex;
  flex-wrap: wrap;
  /* 槽距 20rpx 为锁定值：.cell 宽度 calc((100% - 20rpx)/2) 依赖它（阶梯无 20 档），不得改为令牌 */
  gap: 20rpx;
  /* 顶部不设内边距：网格与总目的间隙唯一来源是 .toc-row 底部 --space-lg（勿复加，防 16+4 复合间距回潮） */
  padding: 0 var(--space-lg) var(--space-md);
  position: relative;
}
.grid-mask {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  background: rgba(245, 246, 248, 0.65);
  z-index: 2;
}
.grid-mask-text {
  position: absolute;
  top: var(--space-xl);
  left: 50%;
  transform: translateX(-50%);
  font-size: 22rpx;
  color: var(--color-text-secondary);
  white-space: nowrap;
}
.cell {
  width: calc((100% - 20rpx) / 2);
  transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1);
}
/* 首屏入场：封面如刊物逐一落上书架（轻抬+淡入）；仅首次加载带 .cell-enter，
   延迟 500ms 后由 markGridAnimated 移除开关，筛/搜/翻页不再重播 */
.cell-enter {
  animation: rpt-cover-in 0.26s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}
@keyframes rpt-cover-in {
  from { opacity: 0; transform: translateY(14rpx); }
  to { opacity: 1; transform: translateY(0); }
}
.cover {
  height: 404rpx;
  border-radius: 12rpx;
  color: #ffffff;
  padding: var(--space-md) var(--space-md);
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  box-shadow: 0 10rpx 24rpx rgba(7, 77, 146, 0.18);
}
/* 封面顶部烫线 + 左侧书脊压暗（纯 CSS 排版，无图片素材） */
.cover::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.45);
}
.cover::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 8rpx;
  background: rgba(0, 0, 0, 0.22);
}
/* 封面渐变深端按小字号白字 AA(≥4.5:1) 校准：青绿/分析橙加深，文字全白不加透明度 */
.cover.navy { background: linear-gradient(155deg, #0c4477 0%, #08335e 100%); }
.cover.teal { background: linear-gradient(155deg, #0b8569 0%, #075c49 100%); }
.cover.orange { background: linear-gradient(155deg, #bd5a15 0%, #8f3d0c 100%); }
.cover.slate { background: linear-gradient(155deg, #455a78 0%, #31415a 100%); }

.c-brand {
  font-size: 18rpx;
  letter-spacing: 5rpx;
  text-shadow: 0 1rpx 4rpx rgba(0, 0, 0, 0.28);
  padding-left: var(--space-xs);
  white-space: nowrap;
  overflow: hidden;
}
.c-mid {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
}
.c-line {
  width: 68rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.55);
}
.c-type {
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 52rpx;
  font-weight: 700;
  letter-spacing: 16rpx;
  padding-left: 16rpx;
}
.c-en {
  font-size: 15rpx;
  letter-spacing: 4rpx;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
}
.c-foot {
  padding-left: var(--space-xs);
}
.c-title {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  /* 封面标题 24rpx（--font-sm）：22rpx 低于全局字阶且步行中难读，已入阶梯 */
  font-size: var(--font-sm);
  font-weight: 600;
  line-height: 1.5;
}
.c-period {
  margin-top: var(--space-xs);
  font-size: 19rpx;
  letter-spacing: 3rpx;
  display: block;
}
/* 图注只留日期（封面已是可读标题，杂志目录惯例），居中于封面下方 */
.cap {
  padding: var(--space-sm) var(--space-xs) 0;
  text-align: center;
}
.cap-d {
  display: block;
  font-size: 21rpx;
  color: var(--color-text-secondary);
  letter-spacing: 2rpx;
}

/* ===== 加载更多 ===== */
.load-more {
  text-align: center;
  padding: var(--space-sm) 0 var(--space-xs);
}
.loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  font-size: 24rpx;
  color: var(--color-text-secondary);
}
.no-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  color: var(--color-text-secondary);
  font-size: 24rpx;
}
.no-more-line {
  width: 48rpx;
  height: 1rpx;
  background: #dde0e4;
}
.load-more-btn {
  display: inline-block;
  height: 60rpx;
  line-height: 60rpx;
  padding: 0 44rpx;
  border-radius: 38rpx;
  border: 1rpx solid #cfe3f8;
  background: #ffffff;
  color: var(--color-primary);
  font-size: 24rpx;
  box-shadow: var(--shadow-sm);
}

/* 演示数据横幅（照 projects/list.vue 惯例；单位用 rpx 入令牌，勿回写 px） */
.mock-note {
  text-align: center;
  padding: 0 0 var(--space-lg);
  font-size: var(--font-xs);
  color: #667085;
}

/* ===== 交互反馈（M3 静态档：仅按压反馈，无自动动效） ===== */
/* 按压反馈：小程序无 hover，用 hover-class 提供 tap 触觉 */
/* 整卡缩放：封面与图注一体按下，避免接缝撕裂 */
.cell-press { transform: scale(0.985); }
.toc-press { opacity: 0.55; }
.icon-press { opacity: 0.6; }
.btn-press { transform: scale(0.96); opacity: 0.92; }
.toc-item { transition: opacity 0.15s; }
.nav-back, .search-clear { transition: opacity 0.15s; }
.retry-btn, .load-more-btn { transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1); }

/* 骨架屏：形状对齐封面网格（M3 静态档：不闪烁，靠形状差异传达加载态） */
.sk-cover {
  height: 404rpx;
  border-radius: 12rpx;
  background: #e9edf3;
}
.sk-line {
  height: 26rpx;
  border-radius: 8rpx;
  background: #e9edf3;
  /* 跟随 .cap 顶部间距（14→16）：骨架线与真实图注保持对齐 */
  margin-top: var(--space-sm);
}
.sk-line-1 { width: 78%; }
.sk-line-2 { width: 46%; height: 20rpx; margin-top: var(--space-xs); }
</style>
