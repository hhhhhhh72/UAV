<template>
  <!-- 弹窗打开时锁定底层 page 滚动，防止穿透 -->
  <page-meta :page-style="overlayStyle" />
  <view class="page">
    <!-- ① 白底导航（返回 + 双 Tab + 胶囊占位） -->
    <view class="nav-wrap" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="nav-bar">
        <view class="nav-back" hover-class="nav-press" :hover-stay-time="100" @click="goBack">
          <text class="nav-back-icon">‹</text>
        </view>
        <view class="nav-tabs">
          <view
            v-for="(t, i) in mainTitles"
            :key="t"
            class="nav-tab"
            :class="{ on: mainTabIndex === i }"
            @click="onMainTabChange(i)"
          >
            <text class="nav-tab-text">{{ t }}</text>
            <view class="nav-tab-line" />
          </view>
        </view>
        <view class="nav-capsule" />
      </view>
      <view class="nav-meta">
        <view class="meta-sync">
          <view class="sync-dot" />
          <text class="sync-text">已同步 · 重庆市应急指挥调度平台</text>
        </view>
        <view class="meta-city" @click="showCityToast">
          <text class="city-text">重庆市</text>
          <text class="city-arrow">▾</text>
        </view>
      </view>
    </view>

    <!-- Tab 1：应急资源 -->
    <template v-if="mainTabIndex === 0">
      <!-- ② 搜索区（跟随滚动） -->
      <view class="search-area">
        <view class="search-box" :class="{ focus: searchFocus }">
          <view class="search-ico"><view class="search-ico-ring" /><view class="search-ico-bar" /></view>
          <input
            class="search-input"
            v-model="keyword"
            placeholder="搜索资源名称 / 编号 / 联系人"
            placeholder-class="search-ph"
            confirm-type="search"
            @focus="searchFocus = true"
            @blur="searchFocus = false"
            @input="onInput"
            @confirm="onSearch"
          />
          <view v-if="keyword" class="search-clear" @click="clearKeyword"><text class="search-clear-x">×</text></view>
        </view>
        <view class="filter-btn" hover-class="filter-press" :hover-stay-time="100" @click="openFilter">
          <view class="filter-ico"><view class="filter-ico-l" /><view class="filter-ico-m" /><view class="filter-ico-r" /></view>
          <text class="filter-btn-text">筛选</text>
        </view>
      </view>

      <view class="content">
        <!-- ③ 统计卡（真实聚合） -->
        <view class="overview">
          <view class="ov-item" v-for="(s, i) in statCols" :key="s.label">
            <text class="ov-num" :class="{ muted: s.muted }">{{ statNums[i] }}<text class="ov-unit">项</text></text>
            <text class="ov-label">{{ s.label }}</text>
            <view class="ov-trend">
              <view class="chip-dot" :class="s.chipCls" />
              <text class="ov-trend-text">{{ s.chip }}</text>
            </view>
          </view>
        </view>

        <!-- ⑤ 列表 -->
        <view v-if="loading && list.length === 0" class="loading-state">
          <view class="loading-inline"><view class="spinner" /><text>加载中...</text></view>
        </view>
        <view v-else-if="errorMsg && list.length === 0" class="state-view">
          <view class="state-ico">!</view>
          <text class="state-text">加载失败</text>
          <view class="retry-btn" @tap="fetchList(true)"><text>重新加载</text></view>
        </view>
        <view v-else-if="!loading && list.length === 0" class="state-view">
          <view class="state-ico">◌</view>
          <text class="state-text">暂无应急资源</text>
        </view>
        <view v-else-if="!loading && filteredList.length === 0" class="state-view">
          <view class="state-ico">◌</view>
          <text class="state-text">无匹配结果</text>
          <view class="retry-btn" @tap="resetFilter"><text>清除筛选</text></view>
        </view>
        <template v-else>
          <view class="sub-bar">
            <text class="sub-bar-text">共 <text class="sub-strong">{{ filteredList.length }}</text> 项 · 按可调度优先级排序</text>
          </view>

          <view class="resource-list" :key="'l' + listFadeKey">
            <view
              v-for="(item, i) in filteredList"
              :key="item.id"
              class="resource"
              hover-class="card-press"
              :hover-stay-time="150"
              :style="{ animationDelay: (i * 80) + 'ms' }"
              @click="openDetail(item)"
            >
              <view v-if="isNew(item)" class="res-new" />

              <view class="resource-head">
                <view class="res-icon" :class="resType(item)">
                  <text class="res-icon-char">{{ resIcon(item) }}</text>
                </view>
                <view class="res-body">
                  <view class="res-title-row">
                    <text class="res-title">{{ item.name || '未命名资源' }}</text>
                    <view class="status-badge" :class="statusKey(item.status)">
                      <view class="badge-dot" />
                      <text class="badge-text">{{ statusLabel(item.status) }}</text>
                    </view>
                  </view>
                  <text class="res-desc">{{ item.specs || '暂无规格描述' }}</text>
                  <view class="res-tags-row">
                    <view v-if="itemTags(item).length" class="res-tags">
                      <text v-for="tg in itemTags(item)" :key="tg.t" class="pill" :class="tg.cls">{{ tg.t }}</text>
                    </view>
                  </view>
                </view>
              </view>

              <view class="res-meta">
                <view class="meta-row">
                  <text class="meta-label">数量</text>
                  <text class="meta-val qty">{{ item.quantity || 0 }} 台</text>
                </view>
                <view class="meta-row">
                  <text class="meta-label">位置</text>
                  <text class="meta-val ellipsis">{{ item.location || '待定' }}</text>
                </view>
                <view class="meta-row meta-row--full">
                  <text class="meta-label">联系人</text>
                  <text class="meta-val">{{ item.contact_info || '暂无' }}</text>
                </view>
              </view>

              <view class="res-foot">
                <view class="res-link" hover-class="link-press" :hover-stay-time="100" @click.stop="openDetail(item)">
                  <text class="link-text">查看详情</text>
                  <text class="link-arrow">›</text>
                </view>
                <view class="res-cta" :class="ctaFor(item).cls" hover-class="cta-press" :hover-stay-time="120" @click.stop="onAction(item)">
                  <text class="cta-text">{{ ctaFor(item).text }}</text>
                </view>
              </view>
            </view>

            <view v-if="filteredList.length > 0" class="load-more">
              <view v-if="loadingMore" class="loading-inline"><view class="spinner" /><text>加载更多...</text></view>
              <view v-else-if="!hasMore" class="no-more-wrap"><text class="no-more-line" /><text class="no-more">没有更多了</text><text class="no-more-line" /></view>
            </view>
            <view class="bottom-space" />
          </view>
        </template>
      </view>
    </template>

    <!-- Tab 2：部门对接 -->
    <template v-else>
      <view class="content">
        <view v-if="deptLoading && deptList.length === 0" class="loading-state">
          <view class="loading-inline"><view class="spinner" /><text>加载中...</text></view>
        </view>
        <view v-else-if="deptError && deptList.length === 0" class="state-view">
          <view class="state-ico">!</view>
          <text class="state-text">加载失败</text>
          <view class="retry-btn" @tap="loadDepts"><text>重新加载</text></view>
        </view>
        <view v-else-if="!deptLoading && deptList.length === 0" class="state-view">
          <view class="state-ico">◌</view>
          <text class="state-text">暂无部门信息</text>
        </view>
        <view v-else class="resource-list">
          <view
            v-for="(d, i) in deptList"
            :key="d.id"
            class="resource dept-card"
            hover-class="card-press"
            :hover-stay-time="150"
            :style="{ animationDelay: (i * 60) + 'ms' }"
          >
            <view class="resource-head">
              <view class="res-icon" :style="deptIconStyle(d)">
                <text class="res-icon-char">{{ deptIcon(d) }}</text>
              </view>
              <view class="res-body">
                <view class="res-title-row">
                  <text class="res-title">{{ d.name || '未知部门' }}</text>
                  <view class="status-badge ok"><view class="badge-dot" /><text class="badge-text">在联</text></view>
                </view>
                <text class="res-desc">{{ deptLabel(d) }} · {{ d.region || '未知区域' }}</text>
              </view>
            </view>
            <view class="res-meta">
              <view class="meta-row meta-row--full">
                <text class="meta-label">联系人</text>
                <text class="meta-val">{{ contactName(d) }}</text>
              </view>
              <view class="meta-row meta-row--full">
                <text class="meta-label">电话</text>
                <view class="meta-phone" hover-class="phone-press" :hover-stay-time="120" @click.stop="callPhone(contactPhone(d))">
                  <text class="meta-val phone-val">{{ contactPhone(d) || '暂无' }}</text>
                </view>
              </view>
            </view>
          </view>
          <view class="bottom-space" />
        </view>
      </view>
    </template>

    <!-- ⑥ 详情抽屉 -->
    <view v-if="showDetail && activeDetail" class="mask" :class="{ 'mask--close': sheetClosing }" catchtouchmove="noop" @click="closeDetail">
      <view class="sheet" :class="{ 'sheet--close': sheetClosing }" @click.stop>
        <view class="sheet-handle" />
        <view class="sheet-head">
          <text class="sheet-title">资源详情</text>
          <view class="sheet-close" @click="closeDetail"><text class="sheet-close-x">×</text></view>
        </view>
        <scroll-view scroll-y class="sheet-body">
          <view class="detail-head">
            <view class="res-icon detail-icon" :class="resType(activeDetail)">
              <text class="res-icon-char">{{ resIcon(activeDetail) }}</text>
            </view>
            <view class="detail-info">
              <text class="detail-title">{{ activeDetail.name }}</text>
              <text class="detail-sub">{{ activeDetail.specs }}</text>
              <view class="detail-tags">
                <view class="status-badge" :class="statusKey(activeDetail.status)">
                  <view class="badge-dot" />
                  <text class="badge-text">{{ statusLabel(activeDetail.status) }}</text>
                </view>
                <text v-for="tg in itemTags(activeDetail)" :key="tg.t" class="pill" :class="tg.cls">{{ tg.t }}</text>
              </view>
            </view>
          </view>

          <view class="kv-grid">
            <view class="kv"><text class="kv-k">数量</text><text class="kv-v">{{ activeDetail.quantity || 0 }} 台</text></view>
            <view class="kv"><text class="kv-k">状态</text><text class="kv-v" :class="'kv-status--' + statusKey(activeDetail.status)">{{ statusLabel(activeDetail.status) }}</text></view>
            <view class="kv"><text class="kv-k">型号</text><text class="kv-v kv-ellipsis">{{ specShort(activeDetail) }}</text></view>
            <view class="kv"><text class="kv-k">联系人</text><text class="kv-v">{{ activeDetail.contact_info || '暂无' }}</text></view>
            <view class="kv kv--wide"><text class="kv-k">位置</text><text class="kv-v">{{ activeDetail.location || '待定' }}</text></view>
          </view>

          <view class="detail-sec">
            <text class="detail-sec-title">能力描述</text>
            <text class="detail-desc">{{ activeDetail.specs || activeDetail.description || '暂无描述' }}</text>
          </view>

          <!-- 调度记录时间线（无记录时隐藏） -->
          <view v-if="detailDispatches.length" class="detail-sec">
            <text class="detail-sec-title">调度记录</text>
            <view class="timeline">
              <view
                v-for="(d, idx) in detailDispatches"
                :key="d.id"
                class="tl-item"
                :class="{ muted: idx !== 0 }"
              >
                <view class="tl-line">
                  <view class="tl-dot" :class="{ active: idx === 0 }" />
                  <view v-if="idx < detailDispatches.length - 1" class="tl-bar" />
                </view>
                <view class="tl-content">
                  <text class="tl-time">{{ formatDispatchTime(d.start_time || d.created_at) }}</text>
                  <text class="tl-title">{{ d.event_desc || '调度事件' }}</text>
                  <text v-if="d.result" class="tl-result">{{ d.result }}</text>
                </view>
              </view>
            </view>
          </view>
        </scroll-view>
        <view class="sheet-foot">
          <view class="btn btn--ghost" hover-class="btn-press" :hover-stay-time="120" @click="callContact(activeDetail.contact_info)">
            <text class="btn-text">联系</text>
          </view>
          <view
            class="btn btn--primary"
            :class="ctaFor(activeDetail).cls"
            :disabled="ctaFor(activeDetail).disabled"
            hover-class="btn-press"
            :hover-stay-time="120"
            @click="onAction(activeDetail)"
          >
            <text class="btn-text">{{ ctaFor(activeDetail).text }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- ⑦ 筛选抽屉 -->
    <view v-if="showFilter" class="mask" :class="{ 'mask--close': sheetClosing }" catchtouchmove="noop" @click="closeFilter">
      <view class="sheet" :class="{ 'sheet--close': sheetClosing }" @click.stop>
        <view class="sheet-handle" />
        <view class="sheet-head">
          <text class="sheet-title">高级筛选</text>
          <view class="sheet-close" @click="closeFilter"><text class="sheet-close-x">×</text></view>
        </view>
        <scroll-view scroll-y class="sheet-body">
          <view class="filter-group">
            <text class="filter-group-title">资源类型</text>
            <view class="filter-opts">
              <view
                v-for="p in typePills"
                :key="p.value"
                class="filter-opt"
                :class="{ on: filterType === p.value }"
                @click="filterType = p.value"
              >{{ p.label }}</view>
            </view>
          </view>
          <view class="filter-group">
            <text class="filter-group-title">可调度状态</text>
            <view class="filter-opts">
              <view
                v-for="s in statusOptions"
                :key="s.value"
                class="filter-opt"
                :class="{ on: filterStatus === s.value }"
                @click="filterStatus = s.value"
              >{{ s.label }}</view>
            </view>
          </view>
          <view class="filter-group">
            <text class="filter-group-title">所在区域</text>
            <view class="filter-opts">
              <view
                v-for="r in regionOptions"
                :key="r"
                class="filter-opt"
                :class="{ on: filterRegion === r }"
                @click="filterRegion = r"
              >{{ r }}</view>
            </view>
          </view>
        </scroll-view>
        <view class="sheet-foot">
          <view class="btn btn--ghost" hover-class="btn-press" :hover-stay-time="120" @click="resetFilter"><text class="btn-text">重置</text></view>
          <view class="btn btn--primary" hover-class="btn-press" :hover-stay-time="120" @click="applyFilter"><text class="btn-text">应用筛选 · {{ filterCount }} 项</text></view>
        </view>
      </view>
    </view>

    <!-- ⑧ 自定义 Toast -->
    <view v-if="toast.show" class="custom-toast" :class="{ 'custom-toast--out': toast.hide }">
      <text class="toast-text">{{ toast.msg }}</text>
    </view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      mainTabIndex: 0,
      mainTitles: ['应急资源', '部门对接'],
      // 顶部状态栏高度：自定义导航需自行下移，避免与状态栏重叠
      statusBarHeight: 24,
      typePills: [
        { label: '全部', value: '', icon: '◈' },
        { label: '无人机', value: 'drone', icon: '机' },
        { label: '通讯', value: 'comm', icon: '信' },
        { label: '车辆', value: 'vehicle', icon: '车' },
        { label: '医疗', value: 'medical', icon: '医' },
        { label: '救援', value: 'rescue', icon: '救' },
      ],
      statusOptions: [
        { label: '全部状态', value: '' },
        { label: '可用', value: 'available' },
        { label: '调度中', value: 'in_use' },
        { label: '离线', value: 'maintenance' },
      ],
      keyword: '',
      searchFocus: false,
      listFadeKey: 0,
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 50,
      hasMore: true,
      deptLoading: false,
      deptError: '',
      deptList: [],
      searchTimer: null,
      showDetail: false,
      activeDetail: null,
      detailDispatches: [],
      showFilter: false,
      filterType: '',
      filterStatus: '',
      filterRegion: '',
      statNums: [0, 0, 0],
      statTimer: null,
      sheetClosing: false,
      toast: { show: false, hide: false, msg: '' },
      toastTimer: null,
      toastOutTimer: null,
    }
  },
  computed: {
    overlayStyle() {
      return (this.showDetail || this.showFilter) ? 'overflow: hidden;' : ''
    },
    statCols() {
      var self = this
      var base = this.filteredList
      var total = base.length
      var avail = base.filter(function (it) { return self.statusKey(it.status) === 'available' }).length
      var busy = base.filter(function (it) { return self.statusKey(it.status) === 'in_use' }).length
      return [
        { label: '在册资源', value: total, chip: '实时', chipCls: 'chip-dot--ok' },
        { label: '当前可用', value: avail, chip: '随时调用', chipCls: 'chip-dot--blue' },
        { label: '在调度中', value: busy, chip: '跟踪任务', chipCls: 'chip-dot--green' },
      ]
    },
    regionOptions() {
      var seen = {}
      var list = this.list.filter(function (it) {
        var m = String(it.location || '').match(/[一-龥]+区|[一-龥]+县|[一-龥]+新区/)
        if (!m) return false
        var r = m[0]
        if (seen[r]) return false
        seen[r] = true
        return true
      }).map(function (it) {
        return String(it.location).match(/[一-龥]+区|[一-龥]+县|[一-龥]+新区/)[0]
      })
      return ['全部区域'].concat(list)
    },
    filterCount() {
      return this.filteredList.length
    },
    filteredList() {
      var self = this
      return this.list.filter(function (it) {
        if (self.filterType && self.resType(it) !== self.filterType) return false
        if (self.filterStatus) {
          var k = self.statusKey(it.status)
          if (k !== self.filterStatus) return false
        }
        if (self.filterRegion && self.filterRegion !== '全部区域') {
          var m = String(it.location || '').match(/[一-龥]+区|[一-龥]+县|[一-龥]+新区/)
          if (!m || m[0] !== self.filterRegion) return false
        }
        return true
      })
    },
  },
  onLoad() {
    this.statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 24
    this.fetchList(true)
  },
  onPullDownRefresh() {
    var p = this.mainTabIndex === 0 ? this.fetchList(true) : this.loadDepts()
    p.then(function () {
      uni.stopPullDownRefresh()
    })
  },
  onReachBottom() {
    if (this.mainTabIndex !== 0) return
    if (!this.loadingMore && this.hasMore) {
      this.loadMore()
    }
  },
  methods: {
    resType(item) {
      return item.res_type || item.resource_type || 'drone'
    },
    async fetchList(reset) {
      if (reset === undefined) reset = true
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
        if (this.keyword) params.q = this.keyword
        var res = await request({ url: '/api/v1/emergency-resources', data: params })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var total = (data && data.total) != null ? data.total : items.length
        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
        if (reset) this.animateStats()
      } catch (e) {
        if (reset) this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    loadMore() {
      this.page++
      this.fetchList(false)
    },
    async loadDepts() {
      this.deptLoading = true
      this.deptError = ''
      try {
        var res = await request({ url: '/api/v1/emergency-depts' })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        this.deptList = Array.isArray(data) ? data : (data && data.items) || data || []
      } catch (e) {
        this.deptError = '网络异常，请稍后重试'
      } finally {
        this.deptLoading = false
      }
    },
    onMainTabChange(index) {
      if (this.mainTabIndex === index) return
      this.mainTabIndex = index
      this.listFadeKey++
      if (index === 1 && this.deptList.length === 0) {
        this.loadDepts()
      }
    },
    onInput() {
      clearTimeout(this.searchTimer)
      var self = this
      this.searchTimer = setTimeout(function () {
        self.page = 1
        self.fetchList(true)
      }, 300)
    },
    onSearch() {
      clearTimeout(this.searchTimer)
      this.page = 1
      this.fetchList(true)
    },
    clearKeyword() {
      this.keyword = ''
      this.onSearch()
    },
    openDetail(item) {
      this.sheetClosing = false
      this.showFilter = false // 抽屉互斥：先关筛选
      this.activeDetail = item
      this.showDetail = true
      this.detailDispatches = []
      this.loadDispatches(item)
    },
    closeDetail() {
      if (!this.showDetail || this.sheetClosing) return
      // 关闭动画：遮罩淡出 + 抽屉下滑 250ms 后再移除
      this.sheetClosing = true
      var self = this
      setTimeout(function () {
        self.showDetail = false
        self.sheetClosing = false
      }, 260)
    },
    openFilter() {
      this.sheetClosing = false
      this.showDetail = false // 抽屉互斥：先关详情
      this.showFilter = true
    },
    closeFilter() {
      if (!this.showFilter || this.sheetClosing) return
      this.sheetClosing = true
      var self = this
      setTimeout(function () {
        self.showFilter = false
        self.sheetClosing = false
      }, 260)
    },
    /* 拉取该资源的调度记录（前端按 resource_id 过滤） */
    async loadDispatches(item) {
      if (!item || !item.id) return
      try {
        var res = await request({ url: '/api/v1/emergency-dispatches', data: { page: 1, page_size: 50 } })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var self = this
        var rid = String(item.id)
        var filtered = items.filter(function (d) {
          return d.resource_id && String(d.resource_id) === rid
        })
        // 按时间倒序（最新在前）
        filtered.sort(function (a, b) {
          var ta = new Date(a.start_time || a.created_at || 0).getTime()
          var tb = new Date(b.start_time || b.created_at || 0).getTime()
          return tb - ta
        })
        this.detailDispatches = filtered.slice(0, 5)
      } catch (e) {
        this.detailDispatches = []
      }
    },
    resetFilter() {
      this.filterType = ''
      this.filterStatus = ''
      this.filterRegion = ''
      this.showFilter = false
      this.listFadeKey++
      this.animateStats()
      this.showCustomToast('已重置筛选')
    },
    applyFilter() {
      this.showFilter = false
      this.listFadeKey++
      this.animateStats()
      this.showCustomToast('筛选已应用 · ' + this.filterCount + ' 项')
    },
    ctaFor(item) {
      var k = this.statusKey(item.status)
      if (k === 'available') {
        // 可用 + 高优（紧急 tag）→ 橙色 + 闪烁呼吸
        var urgent = this.itemTags(item).some(function (tg) { return tg.cls === 'pill--warn' })
        return { text: '立即调度', cls: urgent ? 'cta--urgent cta--flash' : 'cta--urgent', disabled: false }
      }
      if (k === 'in_use') {
        return { text: '加入排队', cls: 'cta--blue', disabled: false }
      }
      return { text: '离线', cls: 'cta--disabled', disabled: true }
    },
    onAction(item) {
      var k = this.statusKey(item.status)
      if (k === 'available') {
        this.showCustomToast('调度单已下发至物资库')
        return
      }
      if (k === 'in_use') {
        this.showCustomToast('该资源调度中，请稍后')
        return
      }
      this.showCustomToast('该资源离线')
    },
    /* 自定义 Toast：入场淡入 + 2000ms 停留 + 淡出（连续触发重置计时） */
    showCustomToast(msg) {
      var self = this
      clearTimeout(this.toastTimer)
      clearTimeout(this.toastOutTimer)
      this.toast = { show: true, hide: false, msg: msg }
      this.toastTimer = setTimeout(function () {
        self.toast.hide = true
        self.toastOutTimer = setTimeout(function () {
          self.toast.show = false
        }, 200)
      }, 2000)
    },
    isNew(item) {
      if (!item.created_at) return false
      var d = new Date(item.created_at)
      if (isNaN(d.getTime())) return false
      return (Date.now() - d.getTime()) < 24 * 3600 * 1000
    },
    itemTags(item) {
      var t = this.resType(item)
      var map = {
        drone: { t: '适航', cls: 'pill--blue' },
        comm: { t: '通讯', cls: 'pill--blue' },
        vehicle: { t: '高优', cls: 'pill--warn' },
        medical: { t: '生命', cls: 'pill--ok' },
        rescue: { t: '救援', cls: 'pill--warn' },
      }
      var tags = []
      if (map[t]) tags.push(map[t])
      if (this.statusKey(item.status) === 'available') tags.push({ t: '可调度', cls: 'pill--ok' })
      return tags.slice(0, 2)
    },
    callContact(info) {
      var phone = this.extractPhone(info)
      if (!phone) {
        this.showCustomToast('暂无联系电话')
        return
      }
      uni.makePhoneCall({ phoneNumber: phone })
    },
    callPhone(phone) {
      if (!phone) {
        this.showCustomToast('暂无联系电话')
        return
      }
      uni.makePhoneCall({ phoneNumber: phone })    },
    extractPhone(str) {
      if (!str) return ''
      var m = String(str).match(/1[3-9]\d{9}/)
      return m ? m[0] : ''
    },
    // 列表卡片联系人：后端只有 contact_info 单字段（含姓名/电话文本），兼容旧 contact_name
    contactName(d) {
      if (!d) return '暂无'
      return d.contact_name || d.contact_info || '暂无'
    },
    contactPhone(d) {
      if (!d) return ''
      return this.extractPhone(d.contact_info || d.contact_name)
    },
    statusKey(status) {
      var s = status || 'available'
      if (s === 'available' || s === 'standby' || s === '可用') return 'available'
      if (s === 'in_use' || s === '使用中' || s === 'dispatched') return 'in_use'
      if (s === 'maintenance' || s === '维护中') return 'maintenance'
      return 'available'
    },
    statusLabel(status) {
      var map = {
        available: '可用',
        standby: '可用',
        可用: '可用',
        in_use: '调度中',
        使用中: '调度中',
        dispatched: '调度中',
        maintenance: '离线',
        维护中: '离线',
      }
      return map[status] || status || '可用'
    },
    resIcon(item) {
      var map = { drone: '机', comm: '信', vehicle: '车', medical: '医', rescue: '救' }
      return map[this.resType(item)] || '他'
    },
    /* 型号：specs 前半段（详情抽屉属性用） */
    specShort(item) {
      var s = item.specs || item.description || ''
      if (!s) return '未标注'
      // 取型号 + 能力（如 "大疆 M300RTK + 热成像云台" → "大疆 M300RTK"）
      var parts = s.split(/[＋+]/)
      return parts[0] ? parts[0].trim() : s
    },
    formatDispatchTime(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      if (isNaN(d.getTime())) return iso
      var p = function (n) { return (n < 10 ? '0' : '') + n }
      return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) + ' ' + p(d.getHours()) + ':' + p(d.getMinutes())
    },
    deptLabel(d) {
      var t = d.dept_type || d.type || ''
      var map = { fire: '消防', police: '公安', civil_affairs: '医疗', emergency_bureau: '应急局' }
      return map[t] || (d.dept_type || '对接单位')
    },
    deptIcon(d) {
      var t = d.dept_type || d.type || ''
      if (t === 'fire') return '防'
      if (t === 'police') return '警'
      if (t === 'civil_affairs') return '医'
      if (t === 'emergency_bureau') return '应'
      var name = d.name || ''
      if (name.indexOf('消防') >= 0) return '防'
      if (name.indexOf('公安') >= 0) return '警'
      if (name.indexOf('应急') >= 0) return '应'
      if (name.indexOf('医疗') >= 0 || name.indexOf('卫生') >= 0) return '医'
      return '部'
    },
    deptIconStyle(d) {
      var t = d.dept_type || d.type || ''
      var name = d.name || ''
      if (t === 'fire' || name.indexOf('消防') >= 0) return { background: '#FFF0E6', color: '#E96012' }
      if (t === 'police' || name.indexOf('公安') >= 0) return { background: '#EAF3FB', color: '#0A66C2' }
      if (t === 'civil_affairs' || name.indexOf('医疗') >= 0 || name.indexOf('卫生') >= 0) return { background: '#E9F7F0', color: '#168A55' }
      if (t === 'emergency_bureau' || name.indexOf('应急') >= 0) return { background: '#EAF3FB', color: '#0A66C2' }
      return { background: '#F4F8FC', color: '#0A66C2' }
    },
    /* 统计卡数字滚动：0 → 目标值（easeOutCubic） */
    animateStats() {
      var self = this
      var targets = this.statCols.map(function (s) { return s.value })
      var duration = 600
      var steps = 24
      var step = 0
      clearInterval(this.statTimer)
      this.statTimer = setInterval(function () {
        step++
        var p = step / steps
        var e = 1 - Math.pow(1 - p, 3)
        self.statNums = targets.map(function (t) { return Math.round(t * e) })
        if (step >= steps) {
          clearInterval(self.statTimer)
          self.statNums = targets
        }
      }, Math.round(duration / steps))
    },
    showCityToast() {
      uni.showToast({ title: '当前仅支持重庆市', icon: 'none' })
    },
    goBack() {
      uni.navigateBack()
    },
    noop() {},
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  /* 安全区：iOS11 constant 前缀 + env 兜底，底部留手势条余量 */
  padding-left: constant(safe-area-inset-left);
  padding-left: env(safe-area-inset-left);
  padding-right: constant(safe-area-inset-right);
  padding-right: env(safe-area-inset-right);
  padding-bottom: calc(constant(safe-area-inset-bottom) + 80rpx);
  padding-bottom: calc(env(safe-area-inset-bottom) + 80rpx);
  overflow-x: hidden;
}

/* ═══ ① 导航（白底，对齐其他页面）═══ */
.nav-wrap {
  background: #ffffff;
  /* 顶部内边距由 JS 读取的真实状态栏高度接管（模板 :style），此处归零 */
  padding: 0;
  position: relative;
  z-index: 5;
  border-bottom: 1rpx solid #EEF1F4;
}
.nav-bar {
  display: flex;
  align-items: center;
  height: 88rpx;
  padding: 0 24rpx;
}
.nav-back {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #F5F8FC;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.nav-press { transform: scale(0.92); background: #EAF3FB; }
.nav-back-icon { color: #17212B; font-size: 40rpx; font-weight: 300; line-height: 1; }
.nav-tabs { flex: 1; display: flex; gap: 48rpx; justify-content: center; }
.nav-tab { position: relative; padding: 10rpx 4rpx; }
.nav-tab-text { font-size: 34rpx; color: #6B7B95; transition: color 200ms ease; }
.nav-tab.on .nav-tab-text { color: #0A66C2; font-weight: 700; }
.nav-tab-line {
  position: absolute;
  bottom: 0;
  left: 50%;
  width: 48rpx;
  height: 6rpx;
  margin-left: -24rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  transform: scaleX(0);
  transition: transform 350ms cubic-bezier(0.2, 0.8, 0.2, 1);
}
.nav-tab.on .nav-tab-line { transform: scaleX(1); }
.nav-capsule {
  width: 88rpx;
  height: 60rpx;
  border: 1rpx solid #E4E7EC;
  border-radius: 999rpx;
  flex-shrink: 0;
}
.nav-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24rpx 20rpx;
}
.meta-sync { display: flex; align-items: center; gap: 8rpx; }
.sync-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #168A55;
}
.sync-text { font-size: 20rpx; color: #6B7B95; }
.meta-city { display: flex; align-items: center; gap: 4rpx; }
.city-text { font-size: 20rpx; color: #17212B; font-weight: 600; }
.city-arrow { font-size: 18rpx; color: #98A2B3; }

/* ═══ ② 搜索区（白底浅灰框，对齐其他页面）═══ */
.search-area {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx 24rpx;
  background: #ffffff;
}
.search-box {
  flex: 1;
  height: 76rpx;
  background: #F5F8FC;
  border: 1rpx solid #EDF0F5;
  border-radius: 999rpx;
  display: flex;
  align-items: center;
  padding: 0 20rpx;
  gap: 10rpx;
  transition: background 200ms ease, border-color 200ms ease, box-shadow 200ms ease;
}
.search-box.focus {
  background: #ffffff;
  border-color: #0A66C2;
  box-shadow: 0 0 0 4rpx rgba(10, 102, 194, 0.12);
}
.search-ico { position: relative; width: 26rpx; height: 26rpx; flex-shrink: 0; }
.search-ico-ring {
  width: 16rpx;
  height: 16rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50%;
  transition: border-color 200ms ease;
}
.search-box.focus .search-ico-ring { border-color: #0A66C2; }
.search-ico-bar {
  position: absolute;
  right: 0;
  bottom: 2rpx;
  width: 8rpx;
  height: 3rpx;
  background: #98A2B3;
  transform: rotate(45deg);
  transform-origin: right center;
  transition: background 200ms ease;
}
.search-box.focus .search-ico-bar { background: #0A66C2; }
.search-input { flex: 1; font-size: 24rpx; color: #17212B; }
.search-ph { color: #ADB8C7; }
.search-clear {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #EDF0F5;
  display: flex;
  align-items: center;
  justify-content: center;
}
.search-clear-x { font-size: 28rpx; color: #6B7B95; line-height: 1; }
.filter-btn {
  height: 76rpx;
  padding: 0 22rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #ffffff;
  border: 1rpx solid #E4E7EC;
  border-radius: 999rpx;
  flex-shrink: 0;
}
.filter-press { transform: scale(0.96); background: #F5F8FC; }
.filter-ico { display: flex; flex-direction: column; gap: 4rpx; }
.filter-ico-l { width: 20rpx; height: 2rpx; background: #344054; }
.filter-ico-m { width: 28rpx; height: 2rpx; background: #344054; }
.filter-ico-r { width: 14rpx; height: 2rpx; background: #344054; align-self: flex-end; }
.filter-btn-text { font-size: 24rpx; color: #344054; font-weight: 600; }

/* ═══ ③ 统计卡 ═══ */
.content { padding-top: 4rpx; }
.overview {
  display: flex;
  box-sizing: border-box;
  margin: 14rpx 24rpx 0;
  padding: 22rpx 8rpx;
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 12rpx;
}
.ov-item { flex: 1; text-align: center; }
.ov-item + .ov-item { border-left: 1rpx solid #EEF1F4; }
.ov-num {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
  line-height: 1.1;
  color: #0A66C2;
  animation: numIn 300ms cubic-bezier(0.16, 1, 0.3, 1) both;
}
.ov-unit { font-size: 20rpx; font-weight: 500; color: #98A2B3; margin-left: 2rpx; }
.ov-label { display: block; font-size: 20rpx; color: #667085; margin-top: 4rpx; }
.ov-trend {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  margin-top: 8rpx;
  padding: 4rpx 12rpx;
  background: #F4F6F8;
  border-radius: 6rpx;
}
.ov-trend-text { font-size: 18rpx; font-weight: 600; color: #667085; }
/* chip 语义色点 */
.chip-dot {
  width: 8rpx;
  height: 8rpx;
  border-radius: 50%;
  flex-shrink: 0;
}
.chip-dot--ok { background: #168A55; }
.chip-dot--blue { background: #0A66C2; }
.chip-dot--green { background: #168A55; }

/* ═══ 状态视图 ═══ */
.loading-state { display: flex; justify-content: center; padding: 160rpx 0; }
.loading-inline { display: flex; align-items: center; gap: 16rpx; font-size: 26rpx; color: #667085; }
.spinner {
  width: 28rpx;
  height: 28rpx;
  border: 3rpx solid #EAF3FB;
  border-top-color: #0A66C2;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
.state-view { display: flex; flex-direction: column; align-items: center; padding-top: 80rpx; }
.state-ico {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #E8F2FC;
  color: #0A66C2;
  font-size: 36rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16rpx;
}
.state-text { font-size: 28rpx; color: #17212B; }
.retry-btn {
  display: inline-block;
  margin-top: 24rpx;
  padding: 14rpx 48rpx;
  background: #074D92;
  color: #ffffff;
  border-radius: 999rpx;
  font-size: 26rpx;
}

/* ═══ ⑤ 列表 ═══ */
.sub-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  padding: 12rpx 24rpx 4rpx;
}
.sub-bar-text { font-size: 20rpx; color: #667085; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sub-strong { color: #17212B; font-weight: 700; }
.resource-list { padding: 2rpx 0 14rpx; }
.resource {
  position: relative;
  box-sizing: border-box;
  width: auto;
  margin: 0 24rpx 12rpx;
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 8rpx;
  padding: 24rpx 28rpx 20rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  animation: cardIn 420ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
  transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease;
}
.card-press { transform: scale(0.985); box-shadow: 0 2rpx 8rpx rgba(16, 24, 40, 0.08); border-color: #EAF3FB; }
.res-new {
  position: absolute;
  top: 16rpx;
  right: 16rpx;
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  background: #0A66C2;
}

/* 头部 */
.resource-head { display: flex; gap: 16rpx; align-items: flex-start; }
.res-icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.res-icon-char { font-size: 32rpx; font-weight: 700; }
.res-icon.drone { background: #EAF3FB; color: #0A66C2; }
.res-icon.comm { background: #F4F8FC; color: #0A66C2; }
.res-icon.vehicle { background: #FFF0E6; color: #E96012; }
.res-icon.medical { background: #E9F7F0; color: #168A55; }
.res-icon.rescue { background: #FEF6E7; color: #B54708; }
.res-body { flex: 1; min-width: 0; }
.res-title-row { display: flex; align-items: center; justify-content: space-between; gap: 8rpx; }
.res-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.res-desc {
  display: block;
  font-size: 22rpx;
  color: #667085;
  margin-top: 6rpx;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.res-tags-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8rpx;
  margin-top: 10rpx;
}
.res-tags { display: flex; flex-wrap: wrap; gap: 8rpx; }

/* 状态徽章（overflow hidden 裁剪扩散环，防溢出闪烁） */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  flex-shrink: 0;
  overflow: hidden;
}
.status-badge.available { background: #E8F2FC; }
.status-badge.in_use { background: #E9F7F0; }
.status-badge.maintenance { background: #F3F4F6; }
.badge-dot { width: 10rpx; height: 10rpx; border-radius: 50%; background: currentColor; position: relative; }
.status-badge.available { color: #0A66C2; }
.status-badge.in_use { color: #168A55; }
.status-badge.maintenance { color: #98A2B3; }
.status-badge.available .badge-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: currentColor;
  animation: ringPulse 1.8s ease-out infinite;
}
/* 调度中：琥珀 dot 1.4s 横向漂移光点 */
.status-badge.in_use .badge-dot {
  background: transparent;
  border: 2rpx solid currentColor;
  box-sizing: border-box;
  overflow: hidden;
}
.status-badge.in_use .badge-dot::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 6rpx;
  background: currentColor;
  border-radius: 999rpx;
  animation: driftDot 1.4s ease-in-out infinite;
}
@keyframes driftDot {
  0% { transform: translateX(-8rpx); opacity: 0.9; }
  100% { transform: translateX(14rpx); opacity: 0.1; }
}
.badge-text { font-size: 20rpx; font-weight: 600; }
.status-badge.ok { color: #0A66C2; background: #E8F2FC; }
.status-badge.ok .badge-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: currentColor;
  animation: ringPulse 1.8s ease-out infinite;
}

/* 标签 pill */
.pill {
  font-size: 18rpx;
  font-weight: 600;
  padding: 4rpx 12rpx;
  border-radius: 6rpx;
}
.pill--blue { background: #F4F8FC; color: #0A66C2; }
.pill--warn { background: #FFF0E6; color: #E96012; }
.pill--ok { background: #E9F7F0; color: #168A55; }

/* 元信息 2 列 */
.res-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx 0;
  margin-top: 16rpx;
  padding: 14rpx 0;
  border-top: 1rpx solid #EEF1F4;
  border-bottom: 1rpx solid #EEF1F4;
}
.meta-row { width: 50%; display: flex; align-items: center; gap: 8rpx; min-width: 0; }
.meta-row--full { width: 100%; }
.meta-label { font-size: 18rpx; color: #98A2B3; width: 56rpx; flex-shrink: 0; }
.meta-val { font-size: 22rpx; color: #344054; font-weight: 600; white-space: nowrap; }
.meta-val.ellipsis {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.meta-val.qty { color: #0A66C2; font-weight: 700; }
.meta-phone { display: inline-flex; align-items: center; }
.phone-val { color: #0A66C2; }
.phone-press { transform: scale(0.96); opacity: 0.8; }

/* 底栏 */
.res-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #EEF1F4;
}
.res-link { display: flex; align-items: center; gap: 4rpx; }
.link-press { opacity: 0.7; }
.link-text { font-size: 24rpx; color: #0A66C2; font-weight: 600; }
.link-arrow { font-size: 30rpx; color: #0A66C2; font-weight: 300; line-height: 1; }
.res-cta {
  height: 60rpx;
  padding: 0 26rpx;
  border-radius: 999rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: transform 180ms ease, background 180ms ease;
}
.cta--urgent { background: #0A66C2; box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28); }
/* 调度中：主蓝浅底文字版 */
.cta--blue { background: #EAF3FB; }
.cta--blue:active { background: #D8E9FB; }
.cta--disabled { background: #EEF1F4; opacity: 0.5; pointer-events: none; }
.cta-press { transform: scale(0.97); }
.cta-text { font-size: 24rpx; font-weight: 700; color: #ffffff; }
.cta--blue .cta-text { color: #0A66C2; }
.cta--disabled .cta-text { color: #667085; }

/* 部门卡片 */
.dept-card .res-icon { background: #EAF3FB; color: #0A66C2; }

/* 底部 */
.load-more { text-align: center; padding: 24rpx 0; }
.no-more-wrap { display: flex; align-items: center; justify-content: center; gap: 16rpx; }
.no-more { font-size: 22rpx; color: #98A2B3; }
.no-more-line {
  width: 80rpx;
  height: 1rpx;
  background: linear-gradient(90deg, rgba(152, 162, 179, 0), rgba(152, 162, 179, 0.5), rgba(152, 162, 179, 0));
}
.bottom-space { height: 40rpx; }

/* ═══ 抽屉 ═══ */
.mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(16, 24, 40, 0.45);
  z-index: 100;
  display: flex;
  align-items: flex-end;
  animation: fadeIn 250ms ease both;
}
.sheet {
  width: 100%;
  height: 76vh;
  background: #ffffff;
  border-radius: 14rpx 14rpx 0 0;
  box-shadow: 0 -12rpx 32rpx rgba(16, 24, 40, 0.16);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: sheetUp 280ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
}
.sheet-handle {
  width: 72rpx;
  height: 8rpx;
  margin: 16rpx auto 8rpx;
  border-radius: 999rpx;
  background: #CBD2DA;
  flex-shrink: 0;
}
.sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 24rpx 16rpx;
  border-bottom: 1rpx solid #EEF1F4;
  flex-shrink: 0;
}
.sheet-title { font-size: 32rpx; font-weight: 700; color: #17212B; }
.sheet-close {
  width: 56rpx;
  height: 56rpx;
  border-radius: 999rpx;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
}
.sheet-close-x { font-size: 32rpx; color: #667085; line-height: 1; }
.sheet-body {
  height: calc(76vh - 230rpx);
  padding: 24rpx;
  box-sizing: border-box;
}

/* 详情抽屉 */
.detail-head { display: flex; gap: 16rpx; align-items: flex-start; }
.detail-icon { width: 88rpx; height: 88rpx; }
.detail-info { flex: 1; min-width: 0; }
.detail-title { display: block; font-size: 32rpx; font-weight: 700; color: #17212B; }
.detail-sub { display: block; font-size: 22rpx; color: #667085; margin-top: 4rpx; }
.detail-tags { display: flex; gap: 8rpx; margin-top: 12rpx; flex-wrap: wrap; }
.kv-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 24rpx;
}
.kv {
  width: calc(50% - 6rpx);
  padding: 16rpx 20rpx;
  background: #F4F8FC;
  border-radius: 6rpx;
  box-sizing: border-box;
}
.kv--wide { width: 100%; }
.kv-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kv-k { display: block; font-size: 18rpx; color: #667085; }
.kv-v { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; margin-top: 6rpx; }
.kv-status--available { color: #0A66C2; }
.kv-status--in_use { color: #168A55; }
.kv-status--maintenance { color: #98A2B3; }

/* 调度记录时间线 */
.timeline { padding-left: 4rpx; }
.tl-item { display: flex; gap: 16rpx; align-items: flex-start; }
.tl-line { display: flex; flex-direction: column; align-items: center; flex-shrink: 0; }
.tl-dot {
  width: 18rpx;
  height: 18rpx;
  border-radius: 50%;
  background: #CBD2DA;
  border: 2rpx solid #ffffff;
  box-shadow: 0 0 0 1rpx #EEF1F4;
  margin-top: 4rpx;
  flex-shrink: 0;
}
.tl-dot.active {
  background: #0A66C2;
  box-shadow: 0 0 0 2rpx #EAF3FB;
}
.tl-bar {
  width: 2rpx;
  flex: 1;
  min-height: 24rpx;
  background: #EEF1F4;
  margin-top: 4rpx;
}
.tl-content { flex: 1; min-width: 0; padding-bottom: 20rpx; }
.tl-time { display: block; font-size: 18rpx; color: #667085; }
.tl-title { display: block; font-size: 24rpx; font-weight: 600; color: #17212B; margin-top: 4rpx; line-height: 1.4; }
.tl-result {
  display: block;
  font-size: 20rpx;
  color: #667085;
  margin-top: 4rpx;
  line-height: 1.5;
}
.tl-item.muted .tl-title { color: #667085; font-weight: 500; }
.tl-item.muted .tl-result { color: #98A2B3; }
.detail-sec { margin-top: 24rpx; }
.detail-sec-title {
  display: block;
  font-size: 22rpx;
  color: #667085;
  font-weight: 600;
  margin-bottom: 12rpx;
  padding-left: 16rpx;
  border-left: 4rpx solid #0A66C2;
}
.detail-desc {
  display: block;
  font-size: 24rpx;
  color: #344054;
  line-height: 1.7;
  background: #F4F6F8;
  border-radius: 6rpx;
  padding: 20rpx;
}

/* 详情抽屉内容 stagger 入场 */
.detail-head { animation: detailIn 400ms cubic-bezier(0.2, 0.8, 0.2, 1) both; }
.kv-grid { animation: detailIn 400ms cubic-bezier(0.2, 0.8, 0.2, 1) 150ms both; }
.detail-sec { animation: detailIn 400ms cubic-bezier(0.2, 0.8, 0.2, 1) 300ms both; }
.detail-sec + .detail-sec { animation-delay: 380ms; }
@keyframes detailIn {
  from { opacity: 0; transform: translateY(6rpx); }
  to { opacity: 1; transform: translateY(0); }
}

/* 筛选抽屉 filter-group 错峰淡入 */
.sheet .filter-group { animation: detailIn 300ms cubic-bezier(0.2, 0.8, 0.2, 1) both; }
.sheet .filter-group:nth-child(2) { animation-delay: 100ms; }
.sheet .filter-group:nth-child(3) { animation-delay: 200ms; }
.sheet .filter-group:nth-child(4) { animation-delay: 300ms; }

/* 筛选抽屉 */
.filter-group { margin-bottom: 24rpx; }
.filter-group-title {
  display: block;
  font-size: 22rpx;
  color: #667085;
  font-weight: 600;
  margin-bottom: 14rpx;
}
.filter-opts { display: flex; flex-wrap: wrap; gap: 12rpx; }
.filter-opt {
  padding: 14rpx 28rpx;
  border-radius: 999rpx;
  border: 1rpx solid #E4E7EC;
  background: #ffffff;
  color: #344054;
  font-size: 24rpx;
  font-weight: 600;
  transition: all 180ms ease;
}
.filter-opt.on {
  background: #EAF3FB;
  border-color: #EAF3FB;
  color: #0A66C2;
}

/* 抽屉底部按钮 */
.sheet-foot {
  display: flex;
  gap: 12rpx;
  padding: 16rpx 28rpx calc(24rpx + constant(safe-area-inset-bottom));
  padding: 16rpx 28rpx calc(24rpx + env(safe-area-inset-bottom));
  border-top: 1rpx solid #EEF1F4;
  flex-shrink: 0;
}
.btn {
  flex: 1;
  height: 80rpx;
  border-radius: 999rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 180ms ease, background 180ms ease;
}
.btn-press { transform: scale(0.98); }
.btn-text { font-size: 28rpx; font-weight: 700; color: #ffffff; }
.btn--ghost { background: #F4F8FC; }
.btn--ghost .btn-text { color: #0A66C2; }
.btn--primary { background: #0A66C2; }

/* ═══ 动画 ═══ */
@keyframes cardIn {
  from { opacity: 0; transform: translateY(8rpx); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes ringPulse {
  0% { transform: scale(1); opacity: 0.8; }
  80% { transform: scale(2.4); opacity: 0; }
  100% { transform: scale(2.4); opacity: 0; }
}
@keyframes numIn {
  from { opacity: 0; transform: scale(0.7); }
  to { opacity: 1; transform: scale(1); }
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes sheetUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}
@keyframes sheetDown {
  from { transform: translateY(0); }
  to { transform: translateY(100%); }
}
@keyframes fadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}
.mask--close { animation: fadeOut 250ms ease both; }
.sheet--close { animation: sheetDown 250ms cubic-bezier(0.2, 0.8, 0.2, 1) both; }
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ═══ ⑧ 自定义 Toast（对齐 pub-toast：底部黑底白字，无图标） ═══ */
.custom-toast {
  position: fixed;
  left: 50%;
  bottom: 168rpx;
  transform: translateX(-50%);
  z-index: 999;
  padding: 20rpx 28rpx;
  background: rgba(23, 33, 43, 0.92);
  border-radius: 18rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.24);
  animation: toastIn 250ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
  max-width: 70vw;
}
.custom-toast--out { animation: toastOut 200ms ease both; }
.toast-text { font-size: 26rpx; color: #ffffff; font-weight: 500; line-height: 1.4; }
@keyframes toastIn {
  from { opacity: 0; transform: translate(-50%, 16rpx); }
  to { opacity: 1; transform: translate(-50%, 0); }
}
@keyframes toastOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* ═══ 减少动态效果支持（prefers-reduced-motion）═══ */
@media (prefers-reduced-motion: reduce) {
  .resource,
  .status-badge.available .badge-dot::after,
  .status-badge.in_use .badge-dot::before,
  .mask,
  .sheet,
  .mask--close,
  .sheet--close,
  .detail-head,
  .kv-grid,
  .detail-sec,
  .filter-group,
  .custom-toast {
    animation: none !important;
    transition: none !important;
  }
}
</style>
