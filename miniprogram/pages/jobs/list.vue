<template>
  <view class="jobs-page">
    <!-- Nav -->
    <u-nav-bar
      title="招聘求职"
      show-back
      @back="goBack"
    />

    <!-- Search -->
    <u-sticky>
      <u-search
        v-model="searchText"
        placeholder="搜索职位"
        @search="onSearch"
      />
    </u-sticky>

    <!-- 双入口条：我的招聘 / 我的投递 -->
    <view class="entry-bar">
      <view class="entry-item" @tap="goMyJobs">
        <view class="entry-ico entry-ico-jobs">聘</view>
        <view class="entry-text">
          <text class="entry-title">我的招聘</text>
          <text class="entry-sub">企业发布与管理职位</text>
        </view>
        <text class="entry-arrow">›</text>
      </view>
      <view class="entry-item" @tap="goApplications">
        <view class="entry-ico entry-ico-apps">投</view>
        <view class="entry-text">
          <text class="entry-title">我的投递</text>
          <text class="entry-sub">跟踪投递进展</text>
        </view>
        <text class="entry-arrow">›</text>
      </view>
    </view>

    <!-- Tabs -->
    <u-tabs
      :active="activeTabIndex"
      :titles="typeTitles"
      @change="onTabChange"
    />

    <!-- Loading state -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && list.length === 0 && !errorMsg" class="empty-state-wrapper">
      <u-empty description="暂无职位" />
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg && list.length === 0" class="error-state">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchList(true)">
        <text>重新加载</text>
      </view>
    </view>

    <!-- 职位卡片列表 -->
    <view v-else class="job-list">
      <view v-for="item in list" :key="item.id" class="job-card" @tap="goDetail(item)">
        <view class="job-top">
          <text class="job-title">{{ item.title }}</text>
          <text class="job-salary">{{ salaryText(item) }}</text>
        </view>
        <view class="job-tags">
          <text v-if="item.job_type" class="job-type">{{ item.job_type }}</text>
          <text v-if="item.location" class="job-loc">{{ item.location }}</text>
        </view>
        <view class="job-foot">
          <text class="job-date">{{ formatDate(item.created_at) }} 发布</text>
          <view
            class="apply-btn"
            :class="{ applied: appliedIds.includes(item.id) }"
            @tap.stop="applyJob(item)"
          >{{ appliedIds.includes(item.id) ? '已投递' : '投递' }}</view>
        </view>
      </view>

      <!-- Load more -->
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
import { request, getStoredUser } from '../../utils/request'

export default {
  data() {
    return {
      searchText: '',
      activeType: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      appliedIds: [],
      typeTabs: [
        { label: '全部', value: '' },
        { label: '全职', value: '全职' },
        { label: '兼职', value: '兼职' },
        { label: '实习', value: '实习' },
        { label: '项目制', value: '项目制' },
      ],
    }
  },
  computed: {
    // u-tabs 只接受字符串标题数组 + 数字 active 索引，映射自 typeTabs
    typeTitles() {
      return this.typeTabs.map(function (t) { return t.label })
    },
    activeTabIndex() {
      for (var i = 0; i < this.typeTabs.length; i++) {
        if (this.typeTabs[i].value === this.activeType) return i
      }
      return 0
    },
  },
  onLoad() {
    this.fetchList(true)
    this.loadApplied()
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
        if (this.activeType) params.type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/jobs', data: params })
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
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    loadMore() {
      this.page++
      this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    onTabChange(index) {
      this.activeType = this.typeTabs[index].value
      this.fetchList(true)
    },

    // ---- 投递闭环 ----
    goMyJobs() {
      uni.navigateTo({ url: '/pages/jobs/mine' })
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/jobs/detail?id=' + encodeURIComponent(item.id) })
    },
    goApplications() {
      uni.navigateTo({ url: '/pages/jobs/applications' })
    },

    // 薪资展示：分 → 元（橙色高亮；未填显示面议）
    salaryText(item) {
      if (item && item.salary_fen) {
        return '¥' + (item.salary_fen / 100).toLocaleString('zh-CN') + '/月'
      }
      return '面议'
    },

    // 已投递标记：进入页面时拉取我的投递 ID 集合
    async loadApplied() {
      const user = getStoredUser()
      if (!user) return
      try {
        const res = await request({ url: '/api/v1/applications' })
        const list = Array.isArray(res) ? res : ((res && res.data) || [])
        this.appliedIds = list.map((a) => a.job_id).filter(Boolean)
      } catch (e) {}
    },

    async applyJob(item) {
      if (this.appliedIds.includes(item.id)) return
      const user = getStoredUser()
      if (!user) {
        uni.showToast({ title: '请先登录', icon: 'none' })
        return
      }
      try {
        const resumes = await request({ url: '/api/v1/resumes/mine' })
        const rlist = Array.isArray(resumes) ? resumes : ((resumes && resumes.data) || [])
        if (!rlist.length) {
          uni.showModal({
            title: '需要简历',
            content: '投递职位需要一份简历，是否现在去创建？',
            success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pages/jobs/resume' }) },
          })
          return
        }
        await request({
          url: '/api/v1/applications',
          method: 'POST',
          data: { job_id: item.id, resume_id: rlist[0].id },
        })
        this.appliedIds.push(item.id)
        uni.showToast({ title: '投递成功', icon: 'success' })
      } catch (e) {
        uni.showToast({ title: (e && e.message) || '投递失败', icon: 'none' })
      }
    },
    goBack() {
      uni.navigateBack()
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
.jobs-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* 双入口条 */
.entry-bar {
  display: flex;
  gap: 16rpx;
  padding: 16rpx 24rpx;
}
.entry-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 20rpx;
  box-shadow: 0 1px 3px rgba(0,0,0,.04);
}
.entry-ico {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
  flex-shrink: 0;
}
.entry-ico-jobs { background: var(--color-primary-light); color: var(--color-primary); }
.entry-ico-apps { background: #FFF0E6; color: #E96012; }
.entry-text { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.entry-title { font-size: 26rpx; font-weight: 700; color: var(--color-text); }
.entry-sub { font-size: 20rpx; color: var(--color-text-placeholder); margin-top: 2rpx; }
.entry-arrow { font-size: 28rpx; color: var(--color-text-placeholder); }

/* 职位卡片 */
.job-list { padding: 8px 12px; }
.job-card {
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 3px 12px rgba(16,24,40,.05);
}
.job-top { display: flex; align-items: flex-start; justify-content: space-between; gap: 16rpx; }
.job-title { font-size: 30rpx; font-weight: 700; color: var(--color-text); line-height: 1.4; flex: 1; min-width: 0; }
.job-salary { font-size: 28rpx; font-weight: 700; color: var(--color-accent-deep); flex-shrink: 0; }
.job-tags { display: flex; align-items: center; gap: 16rpx; margin-top: 14rpx; }
.job-type {
  font-size: 20rpx;
  padding: 2rpx 14rpx;
  border-radius: 4px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: 600;
}
.job-loc { font-size: 22rpx; color: var(--color-text-secondary); }
.job-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 18rpx;
  padding-top: 18rpx;
  border-top: 1rpx solid var(--color-divider);
}
.job-date { font-size: 20rpx; color: var(--color-text-placeholder); }
.apply-btn {
  padding: 8rpx 28rpx;
  border-radius: 6px;
  background: var(--color-primary);
  color: #fff;
  font-size: 24rpx;
  font-weight: 600;
}
.apply-btn.applied { background: var(--color-divider); color: var(--color-text-secondary); }

/* 状态 */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
.loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
}
.empty-state-wrapper { padding-top: 60px; }
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}
.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}
.load-more { text-align: center; padding: 16px 0; }
.no-more { color: var(--color-text-placeholder); font-size: 13px; }
</style>
