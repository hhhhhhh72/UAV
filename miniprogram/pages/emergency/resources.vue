<template>
  <view class="res-page">
    <u-nav-bar
      title="应急资源"
      show-back
      @back="goBack"
    />

    <!-- 主双 Tab：应急资源 / 部门对接 -->
    <u-sticky>
      <u-tabs
        v-model:active="mainTabIndex"
        :titles="mainTitles"
        @change="onMainTabChange"
      />
    </u-sticky>

    <!-- Tab 1：应急资源 -->
    <template v-if="mainTabIndex === 0">
      <view class="filter-area">
        <u-tabs
          v-model:active="typeIndex"
          :titles="typeTitles"
          @change="onTypeChange"
        />
        <u-search
          v-model="keyword"
          placeholder="搜索资源名称"
          @search="onSearch"
        />
      </view>

      <!-- Loading -->
      <view v-if="loading && list.length === 0" class="loading-state">
        <view class="loading-inline">
          <u-loading size="24rpx" color="#667085" />
          <text>加载中...</text>
        </view>
      </view>

      <!-- Error -->
      <view v-else-if="errorMsg && list.length === 0" class="state-view">
        <u-empty description="加载失败" />
        <view class="retry-btn" @tap="fetchList(true)">
          <text>重新加载</text>
        </view>
      </view>

      <!-- Empty -->
      <view v-else-if="!loading && list.length === 0" class="state-view">
        <u-empty description="暂无应急资源" />
      </view>

      <!-- List -->
      <view v-else class="list-body">
        <view class="res-list">
          <view v-for="item in list" :key="item.id" class="res-card">
            <view class="res-top">
              <view class="res-icon" :style="resIconStyle(item)"><text>{{ resIcon(item) }}</text></view>
              <view class="res-info">
                <text class="res-name">{{ item.name || '未命名资源' }}</text>
                <text class="res-spec">{{ item.specs || item.model || '暂无规格' }}</text>
              </view>
              <u-tag :type="statusTagType(item.status)" size="mini" :round="false" plain>{{ statusLabel(item.status) }}</u-tag>
            </view>
            <view class="res-meta">
              <view class="meta-row"><text class="meta-label">数量</text><text class="meta-value">{{ item.quantity || 0 }}</text></view>
              <view class="meta-row"><text class="meta-label">位置</text><text class="meta-value meta-ellipsis">{{ item.location || '未知' }}</text></view>
              <view class="meta-row"><text class="meta-label">联系人</text><text class="meta-value">{{ item.contact || item.contact_info || '暂无' }}</text></view>
            </view>
          </view>

          <!-- Load more -->
          <view v-if="list.length > 0" class="load-more">
            <view v-if="loadingMore" class="loading-inline">
              <u-loading size="24rpx" color="#667085" />
              <text>加载更多...</text>
            </view>
            <text v-else-if="!hasMore" class="no-more">没有更多了</text>
          </view>
        </view>
        <view class="bottom-space" />
      </view>
    </template>

    <!-- Tab 2：部门对接 -->
    <template v-else>
      <!-- Loading -->
      <view v-if="deptLoading && deptList.length === 0" class="loading-state">
        <view class="loading-inline">
          <u-loading size="24rpx" color="#667085" />
          <text>加载中...</text>
        </view>
      </view>

      <!-- Error -->
      <view v-else-if="deptError && deptList.length === 0" class="state-view">
        <u-empty description="加载失败" />
        <view class="retry-btn" @tap="loadDepts">
          <text>重新加载</text>
        </view>
      </view>

      <!-- Empty -->
      <view v-else-if="!deptLoading && deptList.length === 0" class="state-view">
        <u-empty description="暂无部门信息" />
      </view>

      <!-- List -->
      <view v-else class="list-body">
        <view class="res-list">
          <view v-for="d in deptList" :key="d.id" class="dept-card">
            <view class="dept-header">
              <view class="dept-icon" :style="deptIconStyle(d)"><text>{{ deptIcon(d) }}</text></view>
              <view class="dept-info">
                <text class="dept-name">{{ d.name || '未知部门' }}</text>
                <text class="dept-sub">{{ d.type || d.dept_type || '未知类型' }} · {{ d.area || d.region || '未知区域' }}</text>
              </view>
            </view>
            <view class="dept-contact">
              <text class="dept-c-label">联系人</text>
              <text class="dept-c-value">{{ d.contact || d.contact_name || '暂无' }}</text>
              <text class="dept-c-label">电话</text>
              <text class="dept-c-value">{{ d.phone || d.contact_phone || '暂无' }}</text>
            </view>
          </view>
        </view>
        <view class="bottom-space" />
      </view>
    </template>
  </view>
</template>

<script>
import { request } from '../../utils/request'

export default {
  data() {
    return {
      mainTabIndex: 0,
      mainTitles: ['应急资源', '部门对接'],
      typeIndex: 0,
      typeTitles: ['全部', '无人机', '通讯', '车辆', '医疗'],
      typeMap: ['', 'drone', 'comm', 'vehicle', 'medical'],
      keyword: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      deptLoading: false,
      deptError: '',
      deptList: [],
      searchTimer: null,
    }
  },
  watch: {
    keyword() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(function () { this.fetchList(true) }.bind(this), 300)
    },
  },
  onLoad() {
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
        var typeVal = this.typeMap[this.typeIndex]
        if (typeVal) params.res_type = typeVal
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
      this.mainTabIndex = index
      if (index === 1 && this.deptList.length === 0) {
        this.loadDepts()
      }
    },
    onTypeChange(index) {
      this.typeIndex = index
      this.fetchList(true)
    },
    onSearch() {
      clearTimeout(this.searchTimer)
      this.fetchList(true)
    },
    /* 状态归一：可用/使用中/维护中（兼容英文 key） */
    statusLabel(status) {
      var map = {
        '可用': '可用',
        'available': '可用',
        'standby': '可用',
        '使用中': '使用中',
        'in_use': '使用中',
        '维护中': '维护中',
        'maintenance': '维护中',
      }
      return map[status] || status || '未知'
    },
    statusTagType(status) {
      var map = {
        '可用': 'success',
        'available': 'success',
        'standby': 'success',
        '使用中': 'warning',
        'in_use': 'warning',
        '维护中': 'default',
        'maintenance': 'default',
      }
      return map[status] || 'default'
    },
    /* 资源类型字符图标（低饱和色块） */
    resIcon(item) {
      var t = item.resource_type || item.res_type || 'drone'
      var map = { drone: '机', comm: '信', vehicle: '车', medical: '医' }
      return map[t] || '他'
    },
    resIconStyle(item) {
      var t = item.resource_type || item.res_type || 'drone'
      var map = {
        drone: { background: '#EAF3FB', color: '#0A66C2' },
        comm: { background: '#F6F4FF', color: '#667085' },
        vehicle: { background: '#FFF0E6', color: '#E96012' },
        medical: { background: '#E9F7F0', color: '#168A55' },
      }
      return map[t] || { background: '#F4F6F8', color: '#667085' }
    },
    /* 部门类型字符图标（低饱和色块） */
    deptIcon(d) {
      var t = d.type || d.dept_type || ''
      if (t.indexOf('消防') >= 0 || t === 'fire') return '防'
      if (t.indexOf('公安') >= 0 || t === 'police') return '警'
      if (t.indexOf('应急') >= 0 || t === 'emergency_bureau') return '应'
      if (t.indexOf('医疗') >= 0 || t === 'civil_affairs') return '医'
      return '部'
    },
    deptIconStyle(d) {
      var t = d.type || d.dept_type || ''
      if (t.indexOf('消防') >= 0 || t === 'fire') return { background: '#FFF0E6', color: '#E96012' }
      if (t.indexOf('公安') >= 0 || t === 'police') return { background: '#EAF3FB', color: '#0A66C2' }
      if (t.indexOf('医疗') >= 0 || t === 'civil_affairs') return { background: '#E9F7F0', color: '#168A55' }
      if (t.indexOf('应急') >= 0 || t === 'emergency_bureau') return { background: '#EAF3FB', color: '#0A66C2' }
      return { background: '#F4F6F8', color: '#667085' }
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.res-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 筛选区 */
.filter-area {
  background: #ffffff;
  padding-bottom: 8rpx;
}

/* State views */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #667085;
}

.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 80px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #0A66C2;
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
}

/* List */
.list-body {
  padding-top: 12px;
}

.res-list { padding: 0 24rpx; }

/* 资源卡片 */
.res-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.res-top {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
  margin-bottom: 16rpx;
}

.res-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.res-info {
  flex: 1;
  min-width: 0;
}

.res-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #17212B;
  display: block;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.res-spec {
  font-size: 24rpx;
  color: #667085;
  display: block;
  margin-top: 6rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.res-meta {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  font-size: 24rpx;
}

.meta-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}

.meta-label {
  color: #98A2B3;
  width: 72rpx;
  flex-shrink: 0;
}

.meta-value {
  color: #344054;
  font-weight: 500;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 部门卡片 */
.dept-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.dept-header {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 14rpx;
}

.dept-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
  flex-shrink: 0;
}

.dept-info {
  flex: 1;
  min-width: 0;
}

.dept-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
  display: block;
}

.dept-sub {
  font-size: 23rpx;
  color: #667085;
  display: block;
  margin-top: 4rpx;
}

.dept-contact {
  display: flex;
  align-items: center;
  gap: 16rpx;
  font-size: 24rpx;
  flex-wrap: wrap;
}

.dept-c-label {
  color: #98A2B3;
}

.dept-c-value {
  color: #344054;
  font-weight: 500;
}

/* Load more */
.load-more {
  text-align: center;
  padding: 20rpx 0;
}

.no-more {
  font-size: 24rpx;
  color: #98A2B3;
}

.bottom-space { height: 24rpx; }
</style>
