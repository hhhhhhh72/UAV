<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- ① 台账头卡（品牌渐变 + 实时统计） -->
    <view class="hero-card">
      <view class="hero-icon"><text class="hero-icon-text">资</text></view>
      <text class="hero-kicker">INDUSTRY ASSETS</text>
      <text class="hero-title">产业资源台账</text>
      <text class="hero-desc">集中查看无人机、机场、试飞场地与测试基地</text>
      <view class="hero-stats">
        <view>
          <text class="stat-value">{{ list.length }}</text>
          <text class="stat-label">已加载资源</text>
        </view>
        <view>
          <text class="stat-value">{{ activeType ? '1' : '4' }}</text>
          <text class="stat-label">资源类别</text>
        </view>
      </view>
    </view>

    <!-- ② 搜索框（白上白：双层投影浮起） -->
    <view class="sbar">
      <view class="b-search">
        <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
        <input
          class="b-sinp"
          v-model="searchText"
          placeholder="搜索资源名称、型号或位置"
          placeholder-class="b-ph"
          confirm-type="search"
          @confirm="onSearch"
        />
        <text v-if="searchText" class="b-sclr" @tap="clearSearch">×</text>
        <view class="b-sep" />
        <text class="b-sbtn" @tap="onSearch">搜索</text>
      </view>
    </view>

    <!-- ③ 筛选条：一级下划线 tab 分段（对齐科技成果库；资源类型单维度 → tab 即维度，无二级面板） -->
    <view class="stage-wrap">
      <view class="stages">
        <view
          v-for="item in typeTabs"
          :key="item.value"
          class="stg"
          :class="{ on: activeType === item.value }"
          @tap="pickStageTab(item.value)"
        >
          <text>{{ item.label }}</text>
        </view>
      </view>
    </view>

    <!-- ④ 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ displayList.length }}</text> 项资源</text>
      <text class="ir-hint">{{ activeType ? typeLabel(activeType) : '全部类别' }}</text>
    </view>

    <!-- ⑤ 骨架屏：首次加载 -->
    <view v-if="loading && !list.length" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w70"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w50"></view>
        </view>
      </view>
    </view>

    <!-- ⑥ 空 / 错误 -->
    <view v-else-if="errorMsg && !list.length" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="fetchList(true)">重新加载</view>
      </u-empty>
    </view>
    <view v-else-if="!displayList.length" class="st">
      <u-empty description="暂无匹配的产业资源" />
    </view>

    <!-- ⑦ 资源列表 -->
    <view v-else class="cl">
      <view
        v-for="item in displayList"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="goDetail(item)"
      >
        <view class="resource-icon">
          <text class="resource-icon-text">{{ resourceIcon(item.res_type) }}</text>
        </view>
        <view class="resource-main">
          <view class="title-row">
            <text class="resource-name text-ellipsis">{{ item.name || '未命名资源' }}</text>
            <text class="type-tag">{{ typeLabel(item.res_type) }}</text>
            <text v-if="item.status && item.status !== 'available'" class="status-tag" :class="'status-' + item.status">{{ statusLabel(item.status) }}</text>
          </view>
          <text class="resource-model text-ellipsis">{{ item.model || item.specs || '型号信息暂未填写' }}</text>
          <view class="meta-row">
            <text class="price">{{ formatPrice(item.price_fen) }}</text>
            <text class="location text-ellipsis">{{ item.location || '位置待确认' }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="20rpx" color="#0A66C2" />
          <text>加载更多</text>
        </view>
        <text v-else-if="!hasMore" class="no-more">已加载全部资源</text>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      noMotion: false,
      searchText: '',
      activeType: '',
      loading: true,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      typeTabs: [
        { label: '全部', value: '' },
        { label: '无人机', value: 'drone' },
        { label: '无人机机场', value: 'airport' },
        { label: '试飞场地', value: 'test_site' },
        { label: '测试基地', value: 'test_base' },
      ],
    }
  },
  computed: {
    displayList() {
      var keyword = this.searchText.trim().toLowerCase()
      if (!keyword) return this.list
      return this.list.filter(function (item) {
        return [item.name, item.model, item.specs, item.location]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
          .includes(keyword)
      })
    },
  },
  onLoad() {
    this.checkMotion()
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).finally(function () {
      uni.stopPullDownRefresh()
    })
  },
  onReachBottom() {
    if (!this.loading && !this.loadingMore && this.hasMore) {
      this.page += 1
      this.fetchList(false)
    }
  },
  methods: {
    // 减弱动效（无障碍）
    checkMotion() {
      try {
        const sys = uni.getSystemInfoSync()
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
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''
      try {
        var data = { page: this.page, page_size: this.pageSize }
        if (this.activeType) data.res_type = this.activeType
        var res = await request({ url: '/api/v1/industry-resources', data: data })
        var items = Array.isArray(res) ? res : (res && res.data) || []
        if (!Array.isArray(items)) items = []
        this.list = reset ? items : this.list.concat(items)
        this.hasMore = items.length === this.pageSize
      } catch (error) {
        if (!reset) this.page = Math.max(1, this.page - 1)
        if (reset) this.list = []
        this.errorMsg = '网络异常，请稍后重试'
        if (!reset) uni.showToast({ title: '加载更多失败', icon: 'none' })
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    onSearch() {
      // 搜索为前端过滤（displayList），无需请求；确认后收起输入
      uni.hideKeyboard()
    },
    clearSearch() {
      this.searchText = ''
    },
    // 方案 A（单维度，无面板）：非「全部」tab 再点取消 → 回「全部」；「全部」清为全部
    pickStageTab(value) {
      if (value === '') {
        if (this.activeType !== '') { this.activeType = ''; this.fetchList(true) }
        return
      }
      this.activeType = this.activeType === value ? '' : value
      this.fetchList(true)
    },
    goDetail(item) {
      uni.setStorageSync('resource_detail_' + item.id, item)
      uni.navigateTo({ url: '/pkg-service/pages/resources/detail?id=' + encodeURIComponent(item.id) })
    },
    typeLabel(type) {
      var item = this.typeTabs.find(function (tab) { return tab.value === type })
      return item ? item.label : '产业资源'
    },
    statusLabel(status) {
      var map = { available: '可用', in_use: '使用中', maintenance: '维护中' }
      return map[status] || status || ''
    },
    resourceIcon(type) {
      var map = { drone: '机', airport: '场', test_site: '地', test_base: '基' }
      return map[type] || '源'
    },
    formatPrice(value) {
      var amount = Number(value || 0) / 100
      return amount ? '¥' + amount.toFixed(2) : '免费 / 面议'
    },
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  box-sizing: border-box;
  padding: 12px 12px 30px;
  background: #fff;
}

/* ===== 台账头卡（品牌渐变 + 统计，有内容支撑） ===== */
.hero-card {
  position: relative;
  overflow: hidden;
  padding: 19px 18px 15px;
  color: #ffffff;
  background: linear-gradient(135deg, #0A66C2 0%, #074D92 100%);
  border-radius: 10px;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}

.hero-icon {
  position: absolute;
  top: 17px;
  right: 17px;
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.34);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
}

.hero-icon-text {
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

.hero-kicker,
.hero-title,
.hero-desc,
.stat-value,
.stat-label {
  display: block;
}

.hero-kicker { color: rgba(255, 255, 255, 0.75); font-size: 11px; letter-spacing: 0.5px; }
.hero-title { margin-top: 4px; font-size: 20px; font-weight: 700; }
.hero-desc { width: 76%; margin-top: 5px; color: rgba(255, 255, 255, 0.85); font-size: 12px; line-height: 1.5; }

.hero-stats {
  display: flex;
  gap: 45px;
  margin-top: 15px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.18);
}

.stat-value { font-size: 18px; font-weight: 700; }
.stat-label { margin-top: 2px; color: rgba(255, 255, 255, 0.7); font-size: 11px; }

/* ===== 搜索框：白上白 ===== */
.sbar { padding: 12px 0 8px; background: #fff; }
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

/* ===== 筛选条：一级下划线 tab 分段（对齐科技成果库） ===== */
.stage-wrap { position: relative; z-index: 42; }
.stages { display: flex; gap: 24rpx; padding: 4rpx 28rpx 16rpx; white-space: nowrap; } /* gap 40→24rpx：资源类型含「无人机机场」5 字 tab，收窄让 5 项单行放得下（成果库 gap 40rpx 仅够 2-3 字 tab） */
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 0 8rpx;
  font-size: 24rpx; /* 筛选器字体同研发难题 12px */
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; } /* 激活态用 AA 暗变体 */
.stg.on::after { content: ''; position: absolute; left: 8rpx; right: 8rpx; bottom: 16rpx; height: 3rpx; border-radius: 2rpx; background: #074D92; animation: toc-in .22s ease-out; }
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2px 4px;
  font-size: 12px;
  color: #667085;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { font-size: 12px; color: #98A2B3; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; }
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
.sk-tag { width: 32px; height: 32px; border-radius: 8px; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w50 { width: 50%; }
.sk-l.w70 { width: 70%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 资源卡片 ===== */
.cl { display: flex; flex-direction: column; gap: 8px; }
.card {
  display: flex;
  align-items: flex-start;
  gap: 10px;
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

.resource-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.resource-icon-text { font-size: 14px; font-weight: 700; color: #0A66C2; }

.resource-main { flex: 1; min-width: 0; }
.title-row { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.resource-name { font-size: 15px; font-weight: 600; color: #17212B; }

.type-tag {
  padding: 1px 8px;
  border-radius: 4px;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}
.status-tag {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}
.status-in_use { background: #FFF4EC; color: #E96012; }
.status-maintenance { background: #EEF1F4; color: #5D6B82; }

.resource-model {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #667085;
}
.meta-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-top: 6px; }
.price { font-size: 13px; font-weight: 700; color: #C2410C; }
.location { font-size: 12px; color: #667085; }

.text-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 加载更多 */
.load-more { text-align: center; padding: 12px 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 12px; color: #667085; }
.no-more { color: #98A2B3; font-size: 12px; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .stg.on::after { animation: none; } /* 注线画出属位移，关闭 */

/* prefers-reduced-motion（无障碍）：装饰动画/位移缩放全关，保留颜色反馈 */
@media (prefers-reduced-motion: reduce) {
  .stg { animation: none !important; transition: none !important; }
  .stg.on::after { animation: none !important; }
}
</style>
