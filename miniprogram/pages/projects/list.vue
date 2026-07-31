<template>
  <view class="page-container">
    <van-nav-bar title="课题攻关" left-arrow @click-left="goBack" />

    <!-- Search + Sort (sticky) -->
    <view class="search-wrap">
      <view class="search-bar">
        <view class="search-box">
          <van-icon name="search" size="15px" color="#bbb" />
          <input class="search-input" v-model="searchText" placeholder="搜索课题、关键词"
            placeholder-style="color:#bbb" @confirm="onSearchConfirm" />
          <view v-if="searchText" class="search-clear" @tap.stop="clearSearch">
            <van-icon name="clear" size="16px" color="#bbb" />
          </view>
        </view>
        <view class="sort-btn" @tap.stop="toggleSort">
          <van-icon name="bars" size="16px" color="#666" />
          <view class="sort-drop" v-if="sortVisible">
            <view v-for="opt in sortOptions" :key="opt.key" class="sort-item"
              :class="{ active: currentSort === opt.key }" @tap="pickSort(opt.key)">
              <text>{{ opt.label }}</text>
              <van-icon v-if="currentSort === opt.key" name="success" size="14px" color="#1989fa" />
            </view>
          </view>
        </view>
      </view>
      <scroll-view class="tabs" scroll-x enhanced show-scrollbar="false">
        <text v-for="tab in fieldTabs" :key="tab.value" class="tab"
          :class="{ active: activeField === tab.value }" @tap="onTabChange(tab.value)">{{ tab.label }}</text>
      </scroll-view>
    </view>
    <view class="sort-mask" v-if="sortVisible" @tap="sortVisible = false" @touchmove.stop></view>

    <!-- Banner -->
    <view class="banner">
      <text class="banner-icon">&#129309;</text>
      <view class="banner-info">
        <text class="banner-title">联合攻关 · 攻克核心技术难题</text>
        <text class="banner-sub">在研课题 {{ bannerStats.active }} 项 · 累计经费 ¥{{ bannerStats.totalBudget }}万 · 参与单位 {{ bannerStats.units }} 家</text>
      </view>
    </view>

    <view class="info-row">
      <text>共 <text class="info-num">{{ totalCount }}</text> 项课题</text>
      <text class="info-sort" @tap="toggleSort">{{ sortLabel }} <van-icon name="arrow-down" size="10px" /></text>
    </view>

    <!-- Skeleton -->
    <view v-if="loading && list.length === 0" class="skel-list">
      <view v-for="i in 4" :key="'sk'+i" class="skel-card">
        <view class="skel-line w30"></view><view class="skel-line w90"></view>
        <view class="skel-line w70"></view><view class="skel-line w50"></view>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <text class="state-icon">&#9888;</text>
      <text class="state-text">加载失败，请检查网络</text>
      <text class="state-hint">请确认网络连接后重试</text>
      <view class="state-btn" @tap="retry">重新加载</view>
    </view>

    <!-- Empty -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="state-view">
      <text class="state-icon">&#128269;</text>
      <text class="state-text">暂无相关课题</text>
      <text class="state-hint">试试调整筛选条件或搜索关键词</text>
      <view class="state-btn" @tap="resetAll">清除筛选</view>
    </view>

    <!-- Card List -->
    <view v-else class="card-list">
      <view v-for="item in list" :key="item.id" class="card" @tap="goDetail(item)">
        <view class="card-top">
          <view class="card-tags">
            <text class="card-field" :style="{ background: fieldColor(item.field) }">{{ item.field || '其他' }}</text>
            <text class="card-phase" :class="item.phase || 'recruiting'">{{ phaseLabel(item.phase) }}</text>
          </view>
        </view>
        <text class="card-title">{{ item.title }}</text>
        <view class="card-org">
          <text class="lead-tag">牵头</text>
          <text>{{ item.lead || item.lead_org || '重庆市无人机产业协会' }}</text>
          <text v-if="item.orgs && item.orgs.length"> · {{ (item.orgs || []).slice(0, 3).join(' · ') }}</text>
          <text v-else-if="item.org_name"> · {{ item.org_name }}</text>
        </view>
        <text class="card-desc">{{ item.desc || item.description || '' }}</text>
        <view class="card-footer">
          <view class="card-stats">
            <text class="card-participants">&#128101; {{ item.participants || item.participant_count || 0 }} 家参与单位</text>
            <text class="card-budget-text">&#128188; 经费预算</text>
          </view>
          <view class="card-right">
            <text class="card-budget">{{ fmtMoney(item.budget || item.budget_amount) }}</text>
            <text class="card-deadline">&#9200; {{ deadlineText(item.deadline) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="list.length > 0" class="load-more">
      <van-loading v-if="loadingMore" size="16px" color="#c8c9cc">加载更多...</van-loading>
      <text v-else-if="!hasMore">— 没有更多了 —</text>
      <text v-else>— 上拉加载更多 —</text>
    </view>

    <view class="back-top" :class="{ show: showBackTop }" @tap="scrollToTop">
      <van-icon name="arrow-up" size="20px" color="#666" />
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

var FIELD_COLORS = {
  '飞控': '#0d47a1', '电池': '#e65100', 'AI': '#4a148c',
  '通信': '#1a237e', '材料': '#004d40', '载荷': '#b71c1c',
  '标准': '#37474f', '集群': '#bf360c', '无人机': '#0d47a1',
}
var PHASE_MAP = { recruiting: '招募中', progress: '进行中', completed: '已完成' }

var DEMO_DATA = [
  { id: 1, field: '飞控', title: '新一代无人机智能飞控系统联合研发', lead: '重庆市无人机产业协会', orgs: ['北航无人机研究所', '大疆创新科技', '成都纵横'], budget: 1800000, deadline: '2026-12-31', phase: 'recruiting', participants: 5, desc: '针对复杂环境下无人机自主飞行控制、多传感器融合、在线故障诊断等关键问题，联合攻关新一代智能飞控系统...' },
  { id: 2, field: '电池', title: '高能量密度固态电池工程化应用研究', lead: '重庆市无人机产业协会', orgs: ['宁德时代', '清华大学', '重庆大学'], budget: 2500000, deadline: '2026-10-15', phase: 'progress', participants: 8, desc: '面向工业无人机长航时需求，攻关固态电池量产工艺、低温性能优化、安全防护等核心技术...' },
  { id: 3, field: 'AI', title: '低空无人机交通管理AI决策系统', lead: '重庆市无人机产业协会', orgs: ['华为', '中国民航大学', '中科院自动化所'], budget: 1200000, deadline: '2027-03-01', phase: 'recruiting', participants: 3, desc: '构建面向城市低空密集飞行的AI决策引擎，实现实时冲突检测、路径重规划、优先级调度...' },
  { id: 4, field: '通信', title: '无人机5G超视距通信组网技术攻关', lead: '重庆市无人机产业协会', orgs: ['中国移动', '华为', '航天科工二院'], budget: 2000000, deadline: '2026-08-30', phase: 'progress', participants: 6, desc: '针对无人机超视距控制的通信瓶颈，攻关5G网络切片、低延迟视频传输、抗干扰编码等...' },
  { id: 5, field: '材料', title: '轻量化高强度复合材料机身结构研究', lead: '重庆市无人机产业协会', orgs: ['成都纵横', '重庆大学', '西南铝业'], budget: 1500000, deadline: '2026-11-20', phase: 'recruiting', participants: 4, desc: '联合攻关碳纤维/铝合金混合结构设计、低成本成型工艺、疲劳寿命评估等关键技术...' },
  { id: 6, field: '载荷', title: '多源传感器融合成像载荷研制', lead: '重庆市无人机产业协会', orgs: ['海康威视', '中科院光电所', '武汉大学'], budget: 2800000, deadline: '2027-06-30', phase: 'recruiting', participants: 7, desc: '集成多光谱、热红外、激光雷达的轻量化成像载荷，面向农业监测与应急救援场景...' },
  { id: 7, field: '标准', title: '民用无人机适航审定技术标准联合制定', lead: '重庆市无人机产业协会', orgs: ['民航科研院', '中国商飞', '北航'], budget: 800000, deadline: '2026-09-15', phase: 'completed', participants: 12, desc: '联合制定重庆市无人机适航审定技术标准和测试规范，覆盖整机、飞控、通信等子系统...' },
  { id: 8, field: '集群', title: '无人机集群协同作战关键技术研究', lead: '重庆市无人机产业协会', orgs: ['国防科技大学', '中国电科', '哈工大'], budget: 3500000, deadline: '2027-12-31', phase: 'progress', participants: 9, desc: '面向大规模无人机集群的协同感知、任务分配、编队控制、抗干扰通信等核心技术攻关...' },
]

export default {
  data() {
    return {
      searchText: '', searchKeyword: '', currentSort: 'latest', sortVisible: false,
      sortOptions: [
        { key: 'latest', label: '最新发布' },
        { key: 'budget', label: '经费最高' },
        { key: 'deadline', label: '即将截止' },
      ],
      activeField: '', fieldTabs: [
        { label: '全部', value: '' }, { label: '飞控系统', value: '飞控' },
        { label: '动力电池', value: '电池' }, { label: 'AI算法', value: 'AI' },
        { label: '通信链路', value: '通信' }, { label: '新型材料', value: '材料' },
        { label: '载荷设备', value: '载荷' }, { label: '技术标准', value: '标准' },
        { label: '集群协同', value: '集群' },
      ],
      list: [], totalCount: 0, page: 1, pageSize: 20, hasMore: true,
      loading: false, loadingMore: false, errorMsg: '', showBackTop: false,
      bannerStats: { active: 8, totalBudget: 650, units: 36 },
    }
  },
  computed: {
    sortLabel() {
      var map = { latest: '最新发布', budget: '经费最高', deadline: '即将截止' }
      return (map[this.currentSort] || '最新发布') + ' ▼'
    },
  },
  onLoad() { this.fetchList(true) },
  onPullDownRefresh() { this.fetchList(true).finally(function () { uni.stopPullDownRefresh() }) },
  onReachBottom() { if (!this.loadingMore && this.hasMore) this.loadMore() },
  onPageScroll(e) { this.showBackTop = e.scrollTop > 400 },

  methods: {
    async fetchList(reset) {
      if (reset) { this.page = 1; this.hasMore = true; this.loading = true }
      else { this.loadingMore = true }
      this.errorMsg = ''
      try {
        var params = { page: this.page, page_size: this.pageSize }
        if (this.searchKeyword) params.q = this.searchKeyword
        if (this.activeField) params.field = this.activeField
        if (this.currentSort === 'budget') params.sort = 'budget'
        else if (this.currentSort === 'deadline') params.sort = 'deadline'
        else params.sort = 'latest'
        var res = await request({ url: '/api/v1/research-projects', data: params })
        var items = Array.isArray(res) ? res : (res && res.items) || []
        var total = (res && res.total) != null ? res.total : items.length
        if (reset) { this.list = items } else { this.list = this.list.concat(items) }
        this.totalCount = total; this.hasMore = this.list.length < total
      } catch (e) {
        if (reset && DEMO_DATA.length) { this.list = DEMO_DATA; this.totalCount = DEMO_DATA.length; this.hasMore = false; this.errorMsg = '' }
        else if (reset) { this.errorMsg = '网络异常，请稍后重试' }
      } finally { this.loading = false; this.loadingMore = false }
    },
    async loadMore() { this.page++; await this.fetchList(false) },
    retry() { this.fetchList(true) },
    onSearchConfirm() { uni.hideKeyboard(); this.searchKeyword = this.searchText.trim(); this.fetchList(true) },
    clearSearch() { this.searchText = ''; this.searchKeyword = ''; this.fetchList(true) },
    toggleSort() { this.sortVisible = !this.sortVisible },
    pickSort(key) { if (this.currentSort !== key) { this.currentSort = key; this.fetchList(true) } this.sortVisible = false },
    onTabChange(val) { this.activeField = val; this.currentSort = 'latest'; this.fetchList(true) },
    goDetail(item) { uni.showToast({ title: '详情即将上线', icon: 'none' }) },
    goBack() { uni.navigateBack() },
    scrollToTop() { uni.pageScrollTo({ scrollTop: 0, duration: 300 }) },
    resetAll() { this.searchText = ''; this.searchKeyword = ''; this.currentSort = 'latest'; this.activeField = ''; this.fetchList(true) },

    fieldColor(f) { return FIELD_COLORS[f] || '#666' },
    phaseLabel(p) { return PHASE_MAP[p] || '招募中' },
    deadlineText(d) { if (!d) return ''; var t = new Date(d); if (isNaN(t.getTime())) return ''; var days = Math.max(0, Math.ceil((t - new Date()) / (1000 * 60 * 60 * 24))); return days + '天后截止' },
    fmtMoney(n) { if (!n) return '面议'; if (n >= 10000) return '¥' + (n / 10000).toFixed(0) + '万'; return '¥' + (n || 0).toLocaleString() },
  },
}
</script>

<style scoped>
page { background: #f5f6f8; }
.page-container { min-height: 100vh; background: #f5f6f8; padding-bottom: env(safe-area-inset-bottom); }

.search-wrap { background: #fff; position: sticky; top: 0; z-index: 20; }
.search-bar { display: flex; align-items: center; gap: 10px; padding: 8px 14px 10px; background: #fff; }
.search-box { flex: 1; display: flex; align-items: center; background: #f0f1f3; border-radius: 22px; padding: 10px 14px; gap: 8px; }
.search-input { flex: 1; border: none; outline: none; background: transparent; font-size: 14px; color: #1a1a1a; min-width: 0; height: 20px; line-height: 20px; }
.search-clear { flex-shrink: 0; padding: 2px; }
.sort-btn { width: 38px; height: 38px; border-radius: 50%; background: #f0f1f3; display: flex; align-items: center; justify-content: center; flex-shrink: 0; position: relative; }
.sort-btn:active { transform: scale(.93); }
.sort-drop { position: absolute; top: 44px; right: -4px; z-index: 50; background: #fff; border-radius: 12px; box-shadow: 0 4px 24px rgba(0,0,0,.12); padding: 6px 0; min-width: 140px; }
.sort-item { display: flex; align-items: center; justify-content: space-between; padding: 10px 16px; font-size: 13px; color: #333; }
.sort-item:active { background: #f5f7fa; }
.sort-item.active { color: #1989fa; font-weight: 600; }
.sort-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 40; background: transparent; }

.tabs { display: flex; padding: 6px 14px 8px; white-space: nowrap; background: #fff; }
.tab { flex-shrink: 0; padding: 8px 16px; font-size: 13px; color: #666; display: inline-block; position: relative; }
.tab.active { color: #1989fa; font-weight: 600; }
.tab.active::after { content: ''; position: absolute; bottom: -8px; left: 50%; transform: translateX(-50%); width: 20px; height: 3px; background: #1989fa; border-radius: 2px; }

.banner { margin: 12px 14px; padding: 16px; border-radius: 14px; background: linear-gradient(135deg,#004d40,#00695c 30%,#00796b); display: flex; align-items: center; gap: 12px; position: relative; overflow: hidden; }
.banner-icon { font-size: 32px; flex-shrink: 0; }
.banner-info { flex: 1; min-width: 0; }
.banner-title { font-size: 14px; font-weight: 600; color: #fff; display: block; margin-bottom: 4px; }
.banner-sub { font-size: 11px; color: rgba(255,255,255,.7); }

.info-row { display: flex; justify-content: space-between; align-items: center; padding: 8px 16px 4px; font-size: 12px; color: #999; }
.info-num { color: #1989fa; font-weight: 600; }
.info-sort { color: #1989fa; font-weight: 500; }

.card-list { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 20px; }
.card { background: #fff; border-radius: 10px; padding: 14px; border: .5px solid #eee; }
.card:active { transform: scale(.98); }
.card-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.card-tags { display: flex; gap: 6px; align-items: center; }
.card-field { font-size: 11px; padding: 2px 8px; border-radius: 8px; font-weight: 500; color: #fff; }
.card-phase { font-size: 10px; padding: 2px 8px; border-radius: 8px; font-weight: 500; }
.card-phase.recruiting { background: #e8f5e9; color: #2e7d32; }
.card-phase.progress { background: #e3f2fd; color: #1565c0; }
.card-phase.completed { background: #f5f5f5; color: #999; }
.card-title { font-size: 15px; font-weight: 600; color: #1a1a1a; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 8px; }
.card-org { font-size: 12px; color: #666; display: flex; align-items: center; gap: 4px; margin-bottom: 6px; overflow: hidden; white-space: nowrap; }
.lead-tag { font-size: 10px; background: #e8f0fe; color: #1967d2; padding: 1px 6px; border-radius: 4px; flex-shrink: 0; }
.card-desc { font-size: 12px; color: #999; line-height: 1.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 12px; }
.card-footer { display: flex; align-items: flex-end; justify-content: space-between; }
.card-stats { display: flex; flex-direction: column; gap: 4px; }
.card-participants { font-size: 11px; color: #1989fa; font-weight: 500; }
.card-budget-text { font-size: 11px; color: #999; }
.card-right { text-align: right; display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.card-budget { font-size: 16px; font-weight: 700; color: #f57c00; }
.card-deadline { font-size: 11px; color: #999; }

.skel-list { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 20px; }
.skel-card { background: #fff; border-radius: 10px; padding: 14px; }
.skel-line { height: 14px; background: #f0f1f3; border-radius: 4px; margin-bottom: 8px; animation: shimmer 1.5s infinite; }
.skel-line.w90 { width: 90%; } .skel-line.w70 { width: 70%; } .skel-line.w50 { width: 50%; } .skel-line.w30 { width: 30%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

.state-view { text-align: center; padding: 60px 20px; }
.state-icon { font-size: 48px; margin-bottom: 12px; opacity: .5; display: block; }
.state-text { font-size: 14px; color: #999; display: block; margin-bottom: 4px; }
.state-hint { font-size: 12px; color: #bbb; display: block; margin-bottom: 16px; }
.state-btn { display: inline-block; padding: 8px 24px; border-radius: 22px; background: #1989fa; color: #fff; font-size: 13px; font-weight: 500; }
.state-btn:active { opacity: .8; }

.load-more { text-align: center; padding: 12px 0 24px; font-size: 12px; color: #ccc; }

.back-top { position: fixed; bottom: 80px; right: 16px; width: 44px; height: 44px; border-radius: 50%; background: #fff; box-shadow: 0 4px 16px rgba(0,0,0,.12); display: flex; align-items: center; justify-content: center; z-index: 60; opacity: 0; transform: scale(.5); pointer-events: none; transition: opacity .3s, transform .3s cubic-bezier(.17,.89,.32,1.9); }
.back-top.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.back-top:active { transform: scale(.88); background: #f5f6f8; }
</style>
