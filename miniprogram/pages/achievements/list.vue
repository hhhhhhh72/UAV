<template>
  <view class="page-container">
    <!-- Navbar -->
    <van-nav-bar
      title="成果库"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search + Sort Bar -->
    <view class="search-bar">
      <view class="search-box">
        <van-icon name="search" size="15px" color="#bbb" />
        <input
          class="search-input"
          v-model="searchText"
          placeholder="搜索成果名称、关键词"
          placeholder-style="color:#bbb"
          @input="onSearchInput"
          @confirm="onSearchConfirm"
        />
        <view
          v-if="searchText"
          class="search-clear"
          @tap.stop="clearSearch"
        >
          <van-icon name="clear" size="16px" color="#bbb" />
        </view>
      </view>
      <view class="sort-btn" @tap.stop="toggleSort">
        <van-icon name="bars" size="16px" color="#666" />
        <!-- 侧边下拉菜单 -->
        <view class="sort-drop" v-if="sortVisible" @tap.stop>
          <view
            v-for="opt in sortOptions"
            :key="opt.key"
            class="sort-drop-item"
            :class="{ active: currentSort === opt.key }"
            @tap="pickSort(opt.key)"
          >
            <text>{{ opt.label }}</text>
            <van-icon
              v-if="currentSort === opt.key"
              name="success"
              size="14px"
              color="#1989fa"
            />
          </view>
        </view>
      </view>
    </view>
    <!-- 排序遮罩层 -->
    <view class="sort-mask" v-if="sortVisible" @tap="sortVisible = false" @touchmove.stop></view>

    <!-- Banner Carousel -->
    <swiper
      class="carousel"
      :indicator-dots="true"
      :autoplay="true"
      :interval="3500"
      :duration="400"
      :circular="true"
      indicator-color="rgba(255,255,255,0.35)"
      indicator-active-color="#fff"
    >
      <swiper-item v-for="(slide, i) in banners" :key="i">
        <view class="cslide" :style="{ background: slide.bg }">
          <text class="cs-icon">{{ slide.icon }}</text>
          <view class="cs-info">
            <text class="cs-title">{{ slide.title }}</text>
            <text class="cs-sub">{{ slide.sub }}</text>
          </view>
        </view>
      </swiper-item>
    </swiper>

    <!-- Function Nav -->
    <view class="func-nav">
      <view
        v-for="item in funcNavs"
        :key="item.key"
        class="func-item"
        @tap="onFuncNav(item)"
      >
        <view class="func-icon" :style="{ background: item.bg }">
          <text class="func-emoji">{{ item.icon }}</text>
        </view>
        <text class="func-label">{{ item.label }}</text>
      </view>
    </view>

    <!-- Info Row -->
    <view class="info-row">
      <text>共 <text class="info-num">{{ totalCount }}</text> 项成果</text>
      <text class="info-sort" @tap="toggleSort">{{ sortLabel }} <van-icon name="arrow-down" size="10px" /></text>
    </view>

    <!-- ===== SKELETON LOADING ===== -->
    <view v-if="loading && list.length === 0" class="card-grid">
      <view v-for="i in 6" :key="'sk'+i" class="card card-skeleton">
        <view class="sk-cover"></view>
        <view class="sk-body">
          <view class="sk-line w90"></view>
          <view class="sk-line"></view>
          <view class="sk-line w60"></view>
        </view>
      </view>
    </view>

    <!-- ===== ERROR STATE ===== -->
    <view v-else-if="errorMsg && list.length === 0" class="card-grid">
      <view class="state-view">
        <view class="state-icon">&#9888;</view>
        <text class="state-text">加载失败，请检查网络</text>
        <text class="state-hint">请确认网络连接后重试</text>
        <view class="state-btn" @tap="retry">重新加载</view>
      </view>
    </view>

    <!-- ===== EMPTY STATE ===== -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="card-grid">
      <view class="state-view">
        <view class="state-icon">&#128269;</view>
        <text class="state-text">暂无相关成果</text>
        <text class="state-hint">试试调整筛选条件或搜索关键词</text>
        <view class="state-btn" @tap="resetAll">清除筛选</view>
      </view>
    </view>

    <!-- ===== CARD GRID ===== -->
    <view v-else class="card-grid">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        @tap="goDetail(item)"
      >
        <view
          class="card-cover"
          :style="{ background: fieldBg(item.field || item.category) }"
        >
          <text class="card-cover-emoji">{{ fieldIcon(item.field || item.category) }}</text>
          <text class="card-cover-tag">{{ item.field || item.category || '科技成果' }}</text>
          <text
            v-if="itemStatus(item)"
            class="card-cover-status"
            :class="itemStatus(item).cls"
          >{{ itemStatus(item).label }}</text>
        </view>
        <view class="card-body">
          <text class="card-title">{{ item.title }}</text>
          <text class="card-org">{{ item.org_name || item.org || '未知机构' }}</text>
          <view class="card-foot">
            <text>{{ formatDate(item.created_at || item.date) }}</text>
            <text>&#128065; {{ fmtNum(item.views || item.view_count || 0) }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Load More -->
    <view v-if="list.length > 0" class="load-more">
      <van-loading v-if="loadingMore" size="16px" color="#c8c9cc">加载更多...</van-loading>
      <text v-else-if="!hasMore">— 没有更多了 —</text>
      <text v-else>— 上拉加载更多 —</text>
    </view>

  </view>
</template>

<script>
import { request } from '../../utils/request'

/* ===== 字段映射 ===== */
var FIELD_ICONS = {
  '飞控系统': '\u2708',
  '遥感测绘': '\uD83C\uDF0D',
  '动力系统': '\u2699',
  'AI算法': '\uD83E\uDDE0',
  '载荷设备': '\uD83D\uDCF7',
  '集群协同': '\uD83D\uDCE1',
  '通信链路': '\uD83D\uDCE6',
  '标准规范': '\uD83D\uDCCB',
  '地面站': '\uD83D\uDCBB',
  '无人机': '\uD83D\uDEE9',
  '飞控': '\u2708',
  '载荷': '\uD83D\uDCF7',
  '软件': '\uD83D\uDCBB',
  '材料': '\u2699',
}

var FIELD_BG = {
  '飞控系统': '#e3f2fd',
  '遥感测绘': '#e8f5e9',
  '动力系统': '#fff3e0',
  'AI算法': '#f3e5f5',
  '载荷设备': '#fce4ec',
  '集群协同': '#e0f2f1',
  '通信链路': '#e8eaf6',
  '标准规范': '#f5f5f5',
  '地面站': '#fff8e1',
  '无人机': '#e3f2fd',
  '飞控': '#e8f5e9',
  '载荷': '#fce4ec',
  '软件': '#f3e5f5',
  '材料': '#fff3e0',
}

/* ===== DEMO DATA (后端不可用时兜底) ===== */
var DEMO_DATA = [
  { id: 1, field: '飞控系统', title: '无人机智能自适应飞控系统 V3.0', org_name: '北航无人机研究所', date: '2026-07-15', views: 2380, favs: 186, status: 'hot', stage: 'industrialization', description: '本项目针对复杂气象环境下无人机自主飞行控制的核心难题，提出了一种基于深度强化学习的自适应飞控架构。\n\n该系统集成了多传感器融合、在线参数优化与故障容错三大核心技术，可在强风、低能见度等极端条件下保持飞行稳定性。\n\n经实测验证，在6级阵风条件下姿态控制精度提升42%，系统响应延迟降低至8ms以内。', inventors: '张明远、李晓峰、王建国、陈思雨', patent_number: 'CN202610012345.6', application_area: '工业巡检 | 应急救援 | 农业植保 | 物流配送' },
  { id: 2, field: '遥感测绘', title: '高精度无人机航测三维建模技术研究', org_name: '中科院遥感所', date: '2026-06-28', views: 1820, favs: 142, status: '', stage: 'pilot' },
  { id: 3, field: '动力系统', title: '工业级氢燃料电池动力系统', org_name: '清华大学', date: '2026-07-02', views: 3100, favs: 256, status: 'transformed', stage: 'industrialization' },
  { id: 4, field: 'AI算法', title: '基于视觉Transformer的自主避障算法', org_name: '浙江大学', date: '2026-05-18', views: 1650, favs: 98, status: '', stage: 'laboratory' },
  { id: 5, field: '载荷设备', title: '轻量化多光谱成像载荷装置', org_name: '武汉大学', date: '2026-07-10', views: 1420, favs: 112, status: '', stage: 'pilot' },
  { id: 6, field: '集群协同', title: '无人机群协同搜索与救援调度系统', org_name: '国防科技大学', date: '2026-04-22', views: 2750, favs: 203, status: '', stage: 'laboratory' },
  { id: 7, field: '通信链路', title: '无人机超视距5G通信中继系统', org_name: '华为', date: '2026-07-20', views: 4200, favs: 315, status: 'hot', stage: 'industrialization' },
  { id: 8, field: '标准规范', title: '民用无人机适航审定技术标准', org_name: '民航科研院', date: '2026-03-15', views: 980, favs: 67, status: '', stage: 'listed' },
  { id: 9, field: '地面站', title: '便携式无人机地面控制站GCS-200', org_name: '成都纵横', date: '2026-06-10', views: 1560, favs: 89, status: '', stage: 'pilot' },
  { id: 10, field: '飞控系统', title: '基于MPC的倾转旋翼过渡段控制', org_name: '南京航空航天大学', date: '2026-07-08', views: 1230, favs: 74, status: '', stage: 'laboratory' },
  { id: 11, field: '载荷设备', title: '无人机载SAR雷达小型化成像系统', org_name: '中国电子科技集团', date: '2026-05-30', views: 2100, favs: 178, status: 'transformed', stage: 'listed' },
  { id: 12, field: 'AI算法', title: '无人机图像实时目标检测与跟踪平台', org_name: '大疆创新科技', date: '2026-07-01', views: 3800, favs: 290, status: 'hot', stage: 'industrialization' },
]

export default {
  data() {
    return {
      /* 搜索 & 排序 */
      searchText: '',
      searchKeyword: '',
      currentSort: 'latest',
      sortVisible: false,
      sortOptions: [
        { key: 'latest', label: '最新发布' },
        { key: 'views', label: '最多浏览' },
        { key: 'favs', label: '最多收藏' },
      ],

      /* 列表数据 */
      list: [],
      totalCount: 0,
      page: 1,
      pageSize: 20,
      hasMore: true,

      /* 状态 */
      loading: false,
      loadingMore: false,
      errorMsg: '',

      /* Banner */
      banners: [
        { icon: '\u2708', title: 'AI 赋能飞控新突破', sub: '本月新增 42 项前沿成果', bg: 'linear-gradient(135deg,#0d47a1,#1976d2)' },
        { icon: '\uD83D\uDCCB', title: '产学研对接加速', sub: '326 项成果已实现转化', bg: 'linear-gradient(135deg,#1b5e20,#2e7d32)' },
        { icon: '\uD83D\uDE80', title: '标准引领行业', sub: '最新无人机适航标准发布', bg: 'linear-gradient(135deg,#4a148c,#7b1fa2)' },
      ],

      /* 功能导航 */
      funcNavs: [
        { key: 'patent', icon: '\uD83D\uDCC4', label: '发明专利', bg: '#e3f2fd' },
        { key: 'utility', icon: '\u2699', label: '实用新型', bg: '#fff3e0' },
        { key: 'copyright', icon: '\uD83D\uDCBB', label: '软件著作', bg: '#e8f5e9' },
        { key: 'paper', icon: '\uD83D\uDCDA', label: '论文成果', bg: '#f3e5f5' },
        { key: 'standard', icon: '\uD83D\uDCF6', label: '技术标准', bg: '#fce4ec' },
        { key: 'design', icon: '\uD83C\uDFA8', label: '外观设计', bg: '#e0f2f1' },
        { key: 'transformed', icon: '\uD83D\uDE80', label: '已转化', bg: '#fff8e1' },
        { key: 'all', icon: '\uD83D\uDD0D', label: '全部成果', bg: '#e8eaf6' },
      ],
    }
  },

  computed: {
    sortLabel() {
      var map = { latest: '最新发布', views: '最多浏览', favs: '最多收藏' }
      return (map[this.currentSort] || '最新发布') + ' ▼'
    },
  },

  onLoad() {
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
    /* ===== API ===== */
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
        var params = { page: this.page, page_size: this.pageSize }
        if (this.searchKeyword) params.q = this.searchKeyword
        if (this.currentSort === 'views') params.sort = 'views'
        else if (this.currentSort === 'favs') params.sort = 'favs'
        else params.sort = 'latest'

        var res = await request({ url: '/api/v1/achievements', data: params })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var total = (data && data.total) != null ? data.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.totalCount = total
        this.hasMore = this.list.length < total
      } catch (e) {
        /* 后端不可用时使用演示数据 */
        if (reset && DEMO_DATA.length) {
          var sorted = DEMO_DATA.slice()
          if (this.currentSort === 'views') {
            sorted.sort(function(a, b) { return (b.views || 0) - (a.views || 0) })
          } else if (this.currentSort === 'favs') {
            sorted.sort(function(a, b) { return (b.favs || 0) - (a.favs || 0) })
          } else {
            sorted.sort(function(a, b) { return (b.date || '').localeCompare(a.date || '') })
          }
          this.list = sorted
          this.totalCount = sorted.length
          this.hasMore = false
        } else if (reset) {
          this.errorMsg = '网络异常，请稍后重试'
        }
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },

    async loadMore() {
      this.page++
      await this.fetchList(false)
    },

    retry() {
      this.fetchList(true)
    },

    /* ===== 搜索 ===== */
    onSearchInput(e) {
      this.searchText = e.detail ? e.detail.value : (e.target ? e.target.value : '')
    },

    onSearchConfirm() {
      this.searchKeyword = this.searchText.trim()
      this.fetchList(true)
    },

    clearSearch() {
      this.searchText = ''
      this.searchKeyword = ''
      this.fetchList(true)
    },

    /* ===== 排序 ===== */
    toggleSort() {
      this.sortVisible = !this.sortVisible
    },

    pickSort(key) {
      if (this.currentSort !== key) {
        this.currentSort = key
        this.fetchList(true)
      }
      this.sortVisible = false
    },

    /* ===== 功能导航 ===== */
    onFuncNav(item) {
      if (item.key === 'all') {
        this.resetAll()
        return
      }
      if (item.key === 'transformed') {
        this.searchKeyword = '已转化'
      } else {
        var labels = { patent: '发明专利', utility: '实用新型', copyright: '软件著作', paper: '论文成果', standard: '技术标准', design: '外观设计' }
        this.searchKeyword = labels[item.key] || ''
      }
      this.searchText = this.searchKeyword
      this.fetchList(true)
    },

    /* ===== 导航 ===== */
    goDetail(item) {
      uni.navigateTo({
        url: '/pages/achievements/detail?id=' + encodeURIComponent(item.id),
      })
    },

    goBack() {
      uni.navigateBack()
    },

    resetAll() {
      this.searchText = ''
      this.searchKeyword = ''
      this.currentSort = 'latest'
      this.fetchList(true)
    },

    /* ===== 工具函数 ===== */
    fieldIcon(field) {
      return FIELD_ICONS[field] || '\uD83D\uDE80'
    },

    fieldBg(field) {
      return FIELD_BG[field] || '#f0f1f3'
    },

    itemStatus(item) {
      var stage = item.stage || item.status || ''
      if (stage === 'hot') return { label: '热门', cls: 'hot' }
      if (stage === 'transformed' || stage === 'industrialization' || stage === '产业化' || stage === '已转化') return { label: '已转化', cls: 'transformed' }
      if (stage === 'new') return { label: '新成果', cls: 'new' }
      return null
    },

    fmtNum(n) {
      if (!n) return '0'
      if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
      if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
      return String(n)
    },

    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      if (isNaN(d.getTime())) return String(iso).slice(0, 10)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
  },
}
</script>

<style scoped>
/* ===== 基础 ===== */
page {
  --c-primary: #1989fa;
  --c-bg: #f5f6f8;
  --c-white: #fff;
  --c-text: #1a1a1a;
  --c-text-2: #666;
  --c-text-3: #999;
  --c-text-4: #bbb;
  --c-text-5: #ccc;
  --c-border: #eee;
  --c-input-bg: #f0f1f3;
  background: #e8eaed;
}

.page-container {
  min-height: 100vh;
  background: var(--c-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== Search Bar ===== */
.search-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: var(--c-white);
}

.search-box {
  flex: 1;
  display: flex;
  align-items: center;
  background: var(--c-input-bg);
  border-radius: 22px;
  padding: 10px 14px;
  gap: 8px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: var(--c-text);
  min-width: 0;
  height: 20px;
  line-height: 20px;
}

.search-clear {
  flex-shrink: 0;
  padding: 2px;
}

.sort-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--c-input-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
}

.sort-btn:active {
  transform: scale(0.93);
}

/* ===== Carousel ===== */
.carousel {
  margin: 12px 14px;
  border-radius: 14px;
  overflow: hidden;
  height: 130px;
}

.cslide {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 14px;
}

.cs-icon {
  font-size: 38px;
  flex-shrink: 0;
}

.cs-info {
  flex: 1;
  min-width: 0;
}

.cs-title {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  display: block;
  margin-bottom: 4px;
  line-height: 1.3;
}

.cs-sub {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.72);
}

/* ===== Func Nav ===== */
.func-nav {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;
  padding: 0 14px 12px;
}

.func-item {
  text-align: center;
  padding: 8px 4px;
}

.func-item:active {
  transform: scale(0.93);
}

.func-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 6px;
}

.func-emoji {
  font-size: 20px;
}

.func-label {
  font-size: 11px;
  color: var(--c-text-2);
}

/* ===== Info Row ===== */
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2px 16px 8px;
  font-size: 12px;
  color: var(--c-text-3);
}

.info-num {
  color: var(--c-primary);
  font-weight: 600;
}

.info-sort {
  color: var(--c-primary);
  font-weight: 500;
}

/* ===== Card Grid ===== */
.card-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 0 14px 20px;
}

.card {
  background: var(--c-white);
  border-radius: 10px;
  overflow: hidden;
  border: 0.5px solid var(--c-border);
}

.card:active {
  transform: scale(0.97);
}

.card-cover {
  position: relative;
  aspect-ratio: 4 / 3;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px 10px 0 0;
}

.card-cover-emoji {
  font-size: 28px;
}

.card-cover-tag {
  position: absolute;
  top: 6px;
  left: 6px;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 8px;
  font-weight: 500;
}

.card-cover-status {
  position: absolute;
  top: 6px;
  right: 6px;
  color: #fff;
  font-size: 10px;
  padding: 2px 7px;
  border-radius: 6px;
  font-weight: 500;
}

.card-cover-status.hot {
  background: #ff3b30;
}

.card-cover-status.transformed {
  background: #34c759;
}

.card-cover-status.new {
  background: var(--c-primary);
}

.card-body {
  padding: 8px 10px 10px;
}

.card-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--c-text);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 4px;
}

.card-org {
  font-size: 11px;
  color: var(--c-text-2);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-foot {
  font-size: 10px;
  color: var(--c-text-4);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* ===== Skeleton ===== */
.card-skeleton .sk-cover {
  aspect-ratio: 4 / 3;
  background: var(--c-input-bg);
  animation: shimmer 1.5s infinite;
}

.card-skeleton .sk-body {
  padding: 8px 10px;
}

.card-skeleton .sk-line {
  height: 12px;
  background: var(--c-input-bg);
  border-radius: 4px;
  margin-bottom: 6px;
  animation: shimmer 1.5s infinite;
}

.card-skeleton .sk-line.w90 {
  width: 90%;
}

.card-skeleton .sk-line.w60 {
  width: 60%;
}

@keyframes shimmer {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.45;
  }
}

/* ===== State Views ===== */
.state-view {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
}

.state-icon {
  font-size: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
  display: block;
}

.state-text {
  font-size: 14px;
  color: var(--c-text-3);
  display: block;
  margin-bottom: 4px;
}

.state-hint {
  font-size: 12px;
  color: var(--c-text-5);
  display: block;
  margin-bottom: 16px;
}

.state-btn {
  display: inline-block;
  padding: 8px 24px;
  border-radius: 22px;
  background: var(--c-primary);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
}

.state-btn:active {
  opacity: 0.8;
}

/* ===== Load More ===== */
.load-more {
  text-align: center;
  padding: 12px 0 24px;
  font-size: 12px;
  color: var(--c-text-5);
}

/* ===== Sort Dropdown ===== */
.sort-drop {
  position: absolute;
  top: 44px;
  right: -4px;
  z-index: 50;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12);
  padding: 6px 0;
  min-width: 130px;
  animation: dropIn 0.18s ease;
}

@keyframes dropIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.sort-drop-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  font-size: 13px;
  color: #333;
  white-space: nowrap;
}

.sort-drop-item:active {
  background: #f5f7fa;
}

.sort-drop-item.active {
  color: #1989fa;
  font-weight: 600;
}

.sort-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  background: transparent;
}
</style>
