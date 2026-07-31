<template>
  <view class="page-container">
    <!-- Navbar -->
    <van-nav-bar
      title="研发难题广场"
      left-arrow
      @click-left="goBack"
    />

    <!-- Search + Sort (sticky wrapper) -->
    <view class="search-wrap">
      <view class="search-bar">
        <view class="search-box">
          <van-icon name="search" size="15px" color="#bbb" />
          <input
            class="search-input"
            v-model="searchText"
            placeholder="搜索难题、关键词"
            placeholder-style="color:#bbb"
            @confirm="onSearchConfirm"
          />
          <view v-if="searchText" class="search-clear" @tap.stop="clearSearch">
            <van-icon name="clear" size="16px" color="#bbb" />
          </view>
        </view>
        <view class="sort-btn" @tap.stop="toggleSort">
          <van-icon name="bars" size="16px" color="#666" />
          <view class="sort-drop" v-if="sortVisible">
            <view
              v-for="opt in sortOptions"
              :key="opt.key"
              class="sort-item"
              :class="{ active: currentSort === opt.key }"
              @tap="pickSort(opt.key)"
            >
              <text>{{ opt.label }}</text>
              <van-icon v-if="currentSort === opt.key" name="success" size="14px" color="#1989fa" />
            </view>
          </view>
        </view>
      </view>

      <!-- Tabs -->
      <scroll-view class="tabs" scroll-x enhanced show-scrollbar="false">
        <text
          v-for="tab in fieldTabs"
          :key="tab.value"
          class="tab"
          :class="{ active: activeField === tab.value }"
          @tap="onTabChange(tab.value)"
        >{{ tab.label }}</text>
      </scroll-view>
    </view>

    <!-- Sort mask -->
    <view class="sort-mask" v-if="sortVisible" @tap="sortVisible = false" @touchmove.stop></view>

    <!-- Banner -->
    <view class="banner">
      <text class="banner-icon">&#127942;</text>
      <view class="banner-info">
        <text class="banner-title">揭榜挂帅 · 技术攻关等你来战</text>
        <text class="banner-sub">累计悬赏 ¥{{ bannerStats.totalReward }}万 · 已攻克 {{ bannerStats.solved }} 项 · 正在揭榜 {{ bannerStats.inProgress }} 项</text>
      </view>
      <view class="banner-btn" @tap="onPublish">发布难题</view>
    </view>

    <!-- Info Row -->
    <view class="info-row">
      <text>共 <text class="info-num">{{ totalCount }}</text> 项难题</text>
      <text class="info-sort" @tap="toggleSort">{{ sortLabel }} <van-icon name="arrow-down" size="10px" /></text>
    </view>

    <!-- ===== SKELETON ===== -->
    <view v-if="loading && list.length === 0" class="skel-list">
      <view v-for="i in 4" :key="'sk'+i" class="skel-card">
        <view class="skel-line w30"></view>
        <view class="skel-line w90"></view>
        <view class="skel-line w70"></view>
        <view class="skel-line w50"></view>
      </view>
    </view>

    <!-- ===== ERROR ===== -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <text class="state-icon">&#9888;</text>
      <text class="state-text">加载失败，请检查网络</text>
      <text class="state-hint">请确认网络连接后重试</text>
      <view class="state-btn" @tap="retry">重新加载</view>
    </view>

    <!-- ===== EMPTY ===== -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="state-view">
      <text class="state-icon">&#128269;</text>
      <text class="state-text">暂无相关难题</text>
      <text class="state-hint">试试调整筛选条件或搜索关键词</text>
      <view class="state-btn" @tap="resetAll">清除筛选</view>
    </view>

    <!-- ===== CARD LIST ===== -->
    <view v-else class="card-list">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        @tap="goDetail(item)"
      >
        <view class="card-header">
          <view class="card-tags">
            <text class="card-field" :style="{ background: fieldColor(item.field) }">{{ item.field || '其他' }}</text>
            <text class="card-status" :class="statusClass(item)">{{ statusLabel(item) }}</text>
          </view>
        </view>
        <text class="card-title">{{ item.title }}</text>
        <text class="card-desc">{{ item.desc || item.description || '' }}</text>
        <view class="card-footer">
          <view class="card-meta">
            <text class="card-org">{{ item.org_name || item.org || '未知' }}</text>
            <text v-if="item.bids || item.bid_count" class="card-bids">&#9997; {{ item.bids || item.bid_count }} 家已揭榜</text>
          </view>
          <view class="card-right">
            <text class="card-reward">{{ fmtMoney(item.reward || item.reward_amount) }}</text>
            <text class="card-deadline">&#9200; {{ deadlineText(item.deadline) }}</text>
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

    <!-- Back to top -->
    <view class="back-top" :class="{ show: showBackTop }" @tap="scrollToTop">
      <van-icon name="arrow-up" size="20px" color="#666" />
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

/* ===== 字段颜色 ===== */
var FIELD_COLORS = {
  '飞控': '#0d47a1', '电池': '#e65100', 'AI': '#4a148c',
  '通信': '#1a237e', '材料': '#004d40', '载荷': '#b71c1c',
  '无人机': '#0d47a1', '飞控系统': '#0d47a1', 'AI算法': '#4a148c',
}

/* ===== DEMO DATA ===== */
var DEMO_DATA = [
  { id: 1, field: '飞控', title: '多旋翼无人机抗6级阵风飞控算法研发', org: '大疆创新科技', org_name: '大疆创新科技', reward: 150000, reward_amount: 150000, deadline: '2026-09-30', status: 'open', desc: '需要开发一套能在6级阵风条件下保持稳定的自适应飞控算法，适配多旋翼平台...', bids: 12 },
  { id: 2, field: '电池', title: '高能量密度固态电池无人机适配方案', org: '匿名企业', org_name: '匿名企业', reward: 300000, reward_amount: 300000, deadline: '2026-08-15', status: 'urgent', desc: '急需适配现有工业无人机平台的固态电池方案，要求能量密度>400Wh/kg...', bids: 28 },
  { id: 3, field: 'AI', title: '基于边缘计算的实时目标识别与追踪系统', org: '华为', org_name: '华为', reward: 200000, reward_amount: 200000, deadline: '2026-10-20', status: 'open', desc: '在无人机端侧实现实时目标检测，延迟<50ms，准确率>95%...', bids: 18 },
  { id: 4, field: '通信', title: '无人机超视距5G通信中继系统', org: '中国移动', org_name: '中国移动', reward: 180000, reward_amount: 180000, deadline: '2026-11-01', status: 'open', desc: '设计一套基于5G网络切片的无人机超视距控制与数据传输方案...', bids: 9 },
  { id: 5, field: '材料', title: '轻量化高强度碳纤维机身结构优化', org: '成都纵横', org_name: '成都纵横', reward: 120000, reward_amount: 120000, deadline: '2026-08-30', status: 'open', desc: '在现有碳纤维机身基础上减重15%，同时保证结构强度不降低...', bids: 6 },
  { id: 6, field: '载荷', title: '多光谱+热成像一体化轻量载荷装置', org: '海康威视', org_name: '海康威视', reward: 250000, reward_amount: 250000, deadline: '2026-07-25', status: 'urgent', desc: '需将多光谱和热成像传感器集成到<500g的载荷中...', bids: 22 },
  { id: 7, field: '飞控', title: '基于MPC的倾转旋翼过渡段控制优化', org: '匿名企业', org_name: '匿名企业', reward: 160000, reward_amount: 160000, deadline: '2026-12-15', status: 'open', desc: '优化倾转旋翼无人机从悬停到前飞的过渡段控制策略...', bids: 5 },
  { id: 8, field: 'AI', title: '无人机群协同任务分配与路径规划', org: '国防科技大学', org_name: '国防科技大学', reward: 280000, reward_amount: 280000, deadline: '2026-09-10', status: 'open', desc: '面向大规模无人机集群的分布式任务分配与动态路径规划...', bids: 15 },
  { id: 9, field: '材料', title: '无人机隐身涂层材料研发', org: '航天科工', org_name: '航天科工', reward: 500000, reward_amount: 500000, deadline: '2026-10-01', status: 'open', desc: '开发适用于中小型无人机的雷达波吸收涂层材料...', bids: 31 },
  { id: 10, field: '通信', title: '无人机自组网Mesh通信协议栈', org: '中国电科', org_name: '中国电科', reward: 220000, reward_amount: 220000, deadline: '2026-11-20', status: 'open', desc: '设计低延迟高可靠的自组网通信协议，支持50+节点动态拓扑...', bids: 11 },
  { id: 11, field: '飞控', title: '城市峡谷GPS拒止环境下的视觉导航', org: '顺丰科技', org_name: '顺丰科技', reward: 180000, reward_amount: 180000, deadline: '2026-08-01', status: 'urgent', desc: '在GPS信号弱的城市环境中实现基于视觉SLAM的精准定位导航...', bids: 20 },
  { id: 12, field: '载荷', title: '无人机载小型SAR雷达成像系统', org: '中科院电子所', org_name: '中科院电子所', reward: 350000, reward_amount: 350000, deadline: '2026-12-31', status: 'open', desc: '开发重量<3kg的无人机载合成孔径雷达成像系统...', bids: 8 },
  { id: 13, field: 'AI', title: '低空无人机交通管理AI决策引擎', org: '匿名企业', org_name: '匿名企业', reward: 420000, reward_amount: 420000, deadline: '2026-11-15', status: 'open', desc: '面向城市低空密集飞行场景的实时冲突检测与自主避让决策引擎...', bids: 17 },
  { id: 14, field: '通信', title: '抗干扰卫星导航接收机小型化设计', org: '北斗星通', org_name: '北斗星通', reward: 190000, reward_amount: 190000, deadline: '2026-10-10', status: 'open', desc: '在30g以内实现抗干扰卫星导航接收，抗干信比≥60dB...', bids: 13 },
  { id: 15, field: '电池', title: '无人机无线充电起降平台', org: '中兴通讯', org_name: '中兴通讯', reward: 260000, reward_amount: 260000, deadline: '2026-09-20', status: 'open', desc: '设计适用于工业无人机的自动无线充电起降平台，充电效率>90%...', bids: 19 },
  { id: 16, field: '飞控', title: '基于数字孪生的飞控仿真验证平台', org: '中国商飞', org_name: '中国商飞', reward: 380000, reward_amount: 380000, deadline: '2026-12-01', status: 'open', desc: '构建高保真数字孪生模型，实现飞控系统全生命周期仿真验证...', bids: 7 },
]

export default {
  data() {
    return {
      searchText: '',
      searchKeyword: '',
      currentSort: 'latest',
      sortVisible: false,
      sortOptions: [
        { key: 'latest', label: '最新发布' },
        { key: 'reward', label: '悬赏最高' },
        { key: 'deadline', label: '即将截止' },
      ],
      activeField: '',
      fieldTabs: [
        { label: '全部', value: '' },
        { label: '飞控系统', value: '飞控' },
        { label: '动力电池', value: '电池' },
        { label: 'AI算法', value: 'AI' },
        { label: '通信链路', value: '通信' },
        { label: '新型材料', value: '材料' },
        { label: '载荷设备', value: '载荷' },
      ],
      list: [],
      totalCount: 0,
      page: 1,
      pageSize: 20,
      hasMore: true,
      loading: false,
      loadingMore: false,
      errorMsg: '',
      showBackTop: false,
      bannerStats: { totalReward: 128, solved: 47, inProgress: 12 },
    }
  },

  computed: {
    sortLabel() {
      var map = { latest: '最新发布', reward: '悬赏最高', deadline: '即将截止' }
      return (map[this.currentSort] || '最新发布') + ' ▼'
    },
  },

  onLoad() {
    this.fetchList(true)
  },

  onPullDownRefresh() {
    this.fetchList(true).finally(function () { uni.stopPullDownRefresh() })
  },

  onReachBottom() {
    if (!this.loadingMore && this.hasMore) this.loadMore()
  },

  onPageScroll(e) {
    this.showBackTop = e.scrollTop > 400
  },

  methods: {
    async fetchList(reset) {
      if (reset) {
        this.page = 1; this.hasMore = true; this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''

      try {
        var params = { page: this.page, page_size: this.pageSize }
        if (this.searchKeyword) params.q = this.searchKeyword
        if (this.activeField) params.field = this.activeField
        if (this.currentSort === 'reward') params.sort = 'reward'
        else if (this.currentSort === 'deadline') params.sort = 'deadline'
        else params.sort = 'latest'

        var res = await request({ url: '/api/v1/rd-challenges', data: params })
        var items = Array.isArray(res) ? res : (res && res.items) || []
        var total = (res && res.total) != null ? res.total : items.length

        if (reset) { this.list = items } else { this.list = this.list.concat(items) }
        this.totalCount = total
        this.hasMore = this.list.length < total
      } catch (e) {
        if (reset && DEMO_DATA.length) {
          this.list = DEMO_DATA
          this.totalCount = DEMO_DATA.length
          this.hasMore = false
          this.errorMsg = ''
        } else if (reset) {
          this.errorMsg = '网络异常，请稍后重试'
        }
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },

    async loadMore() { this.page++; await this.fetchList(false) },
    retry() { this.fetchList(true) },

    onSearchConfirm() { uni.hideKeyboard(); this.searchKeyword = this.searchText.trim(); this.fetchList(true) },
    clearSearch() { this.searchText = ''; this.searchKeyword = ''; this.fetchList(true) },

    toggleSort() { this.sortVisible = !this.sortVisible },
    pickSort(key) {
      if (this.currentSort !== key) { this.currentSort = key; this.fetchList(true) }
      this.sortVisible = false
    },

    onTabChange(val) { this.activeField = val; this.currentSort = 'latest'; this.fetchList(true) },

    goDetail(item) { uni.showToast({ title: '即将上线', icon: 'none' }) },
    goBack() { uni.navigateBack() },
    scrollToTop() { uni.pageScrollTo({ scrollTop: 0, duration: 300 }) },

    resetAll() {
      this.searchText = ''; this.searchKeyword = ''; this.currentSort = 'latest'; this.activeField = ''
      this.fetchList(true)
    },

    onPublish() { uni.showToast({ title: '发布难题（仅会员）', icon: 'none' }) },

    /* ===== UTILS ===== */
    fieldColor(f) { return FIELD_COLORS[f] || '#666' },
    statusClass(item) {
      var s = item.status || item.stage || ''
      return s === 'urgent' ? 'urgent' : s === 'closed' ? 'closed' : 'open'
    },
    statusLabel(item) {
      var s = item.status || item.stage || ''
      return s === 'urgent' ? '⚠ 紧急' : s === 'closed' ? '已截止' : '进行中'
    },
    deadlineText(d) {
      if (!d) return ''
      var t = new Date(d)
      if (isNaN(t.getTime())) return ''
      var diff = t - new Date()
      var days = Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
      return days + '天后截止'
    },
    fmtMoney(n) {
      if (!n) return '面议'
      if (n >= 10000) return '¥' + (n / 10000).toFixed(0) + '万'
      return '¥' + (n || 0).toLocaleString()
    },
  },
}
</script>

<style scoped>
page { background: #f5f6f8; }
.page-container { min-height: 100vh; background: #f5f6f8; padding-bottom: env(safe-area-inset-bottom); }

/* ===== Search ===== */
.search-wrap { background: #fff; position: sticky; top: 0; z-index: 20; }
.search-bar { display: flex; align-items: center; gap: 10px; padding: 8px 14px 10px; background: #fff; }
.search-box {
  flex: 1; display: flex; align-items: center; background: #f0f1f3;
  border-radius: 22px; padding: 10px 14px; gap: 8px;
}
.search-input { flex: 1; border: none; outline: none; background: transparent; font-size: 14px; color: #1a1a1a; min-width: 0; height: 20px; line-height: 20px; }
.search-clear { flex-shrink: 0; padding: 2px; }
.sort-btn {
  width: 38px; height: 38px; border-radius: 50%; background: #f0f1f3;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0; position: relative;
}
.sort-btn:active { transform: scale(.93); }
.sort-drop {
  position: absolute; top: 44px; right: -4px; z-index: 50;
  background: #fff; border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0,0,0,.12); padding: 6px 0; min-width: 130px;
}
.sort-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 16px; font-size: 13px; color: #333;
}
.sort-item:active { background: #f5f7fa; }
.sort-item.active { color: #1989fa; font-weight: 600; }
.sort-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 40; background: transparent; }

/* ===== Tabs ===== */
.tabs { display: flex; padding: 6px 14px 8px; white-space: nowrap; background: #fff; }
.tab {
  flex-shrink: 0; padding: 8px 16px; font-size: 13px; color: #666;
  display: inline-block; position: relative;
}
.tab.active { color: #1989fa; font-weight: 600; }
.tab.active::after {
  content: ''; position: absolute; bottom: -8px; left: 50%; transform: translateX(-50%);
  width: 20px; height: 3px; background: #1989fa; border-radius: 2px;
}

/* ===== Banner ===== */
.banner {
  margin: 12px 14px; padding: 16px; border-radius: 14px;
  background: linear-gradient(135deg,#1a237e,#283593 30%,#3949ab);
  display: flex; align-items: center; gap: 12px; position: relative; overflow: hidden;
}
.banner-icon { font-size: 32px; flex-shrink: 0; }
.banner-info { flex: 1; min-width: 0; }
.banner-title { font-size: 14px; font-weight: 600; color: #fff; display: block; margin-bottom: 4px; }
.banner-sub { font-size: 11px; color: rgba(255,255,255,.7); }
.banner-btn {
  flex-shrink: 0; padding: 7px 16px; border-radius: 22px;
  background: rgba(255,255,255,.2); color: #fff; font-size: 12px; font-weight: 500; white-space: nowrap;
}
.banner-btn:active { background: rgba(255,255,255,.35); }

/* ===== Info ===== */
.info-row { display: flex; justify-content: space-between; align-items: center; padding: 8px 16px 4px; font-size: 12px; color: #999; }
.info-num { color: #1989fa; font-weight: 600; }
.info-sort { color: #1989fa; font-weight: 500; }

/* ===== Cards ===== */
.card-list { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 20px; }
.card {
  background: #fff; border-radius: 10px; padding: 14px;
  border: .5px solid #eee;
}
.card:active { transform: scale(.98); }
.card-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.card-tags { display: flex; gap: 6px; align-items: center; }
.card-field { font-size: 11px; padding: 2px 8px; border-radius: 8px; font-weight: 500; color: #fff; }
.card-status { font-size: 10px; padding: 2px 8px; border-radius: 8px; font-weight: 500; }
.card-status.open { background: #e8f5e9; color: #2e7d32; }
.card-status.urgent { background: #fce4ec; color: #c62828; }
.card-status.closed { background: #f5f5f5; color: #999; }
.card-title {
  font-size: 15px; font-weight: 600; color: #1a1a1a; line-height: 1.4;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
  overflow: hidden; margin-bottom: 8px;
}
.card-desc {
  font-size: 12px; color: #999; line-height: 1.5;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 12px;
}
.card-footer { display: flex; align-items: flex-end; justify-content: space-between; }
.card-meta { display: flex; flex-direction: column; gap: 4px; }
.card-org { font-size: 11px; color: #666; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 140px; }
.card-bids { font-size: 11px; color: #1989fa; font-weight: 500; }
.card-right { text-align: right; display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.card-reward { font-size: 16px; font-weight: 700; color: #ff3b30; }
.card-deadline { font-size: 11px; color: #999; }

/* ===== Skeleton ===== */
.skel-list { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 20px; }
.skel-card { background: #fff; border-radius: 10px; padding: 14px; }
.skel-line { height: 14px; background: #f0f1f3; border-radius: 4px; margin-bottom: 8px; animation: shimmer 1.5s infinite; }
.skel-line.w90 { width: 90%; } .skel-line.w70 { width: 70%; } .skel-line.w50 { width: 50%; } .skel-line.w30 { width: 30%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== States ===== */
.state-view { text-align: center; padding: 60px 20px; }
.state-icon { font-size: 48px; margin-bottom: 12px; opacity: .5; display: block; }
.state-text { font-size: 14px; color: #999; display: block; margin-bottom: 4px; }
.state-hint { font-size: 12px; color: #bbb; display: block; margin-bottom: 16px; }
.state-btn { display: inline-block; padding: 8px 24px; border-radius: 22px; background: #1989fa; color: #fff; font-size: 13px; font-weight: 500; }
.state-btn:active { opacity: .8; }

/* ===== Load More ===== */
.load-more { text-align: center; padding: 12px 0 24px; font-size: 12px; color: #ccc; }

/* ===== Back Top ===== */
.back-top {
  position: fixed; bottom: 80px; right: 16px;
  width: 44px; height: 44px; border-radius: 50%;
  background: #fff; box-shadow: 0 4px 16px rgba(0,0,0,.12);
  display: flex; align-items: center; justify-content: center; z-index: 60;
  opacity: 0; transform: scale(.5); pointer-events: none;
  transition: opacity .3s, transform .3s cubic-bezier(.17,.89,.32,1.9);
}
.back-top.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.back-top:active { transform: scale(.88); background: #f5f6f8; }
</style>
