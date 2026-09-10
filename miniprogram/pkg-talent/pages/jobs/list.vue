<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="招聘求职" show-back :fixed="true" @back="goBack" />

    <!-- ① 白底头部：搜索 + 双入口 + 筛选 -->
    <view class="head-zone">
      <!-- 搜索框（全站 b-search 同款：白上白双层投影） -->
      <view class="sbar">
        <view class="b-search">
          <u-icon name="search" size="30rpx" color="#667085" />
          <input
            class="b-sinp"
            v-model="searchText"
            placeholder="搜索职位名称 / 工作地点"
            placeholder-class="b-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <view v-if="searchText" class="b-sclr" @tap="clearSearch">
            <u-icon name="close" size="24rpx" color="#667085" />
          </view>
          <view class="b-sep" />
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- 双入口条：我的招聘 / 我的投递 -->
      <view class="entry-bar">
        <view class="entry-item" hover-class="tap-scale" :hover-stay-time="100" @tap="goMyJobs">
          <view class="entry-ico entry-ico-jobs">聘</view>
          <view class="entry-text">
            <text class="entry-title">我的招聘</text>
            <text class="entry-sub">发布与管理职位</text>
          </view>
          <u-icon name="arrow" size="28rpx" color="#667085" />
        </view>
        <view class="entry-item" hover-class="tap-scale" :hover-stay-time="100" @tap="goApplications">
          <view class="entry-ico entry-ico-apps">投</view>
          <view class="entry-text">
            <text class="entry-title">我的投递</text>
            <text class="entry-sub">跟踪投递进展</text>
          </view>
          <u-icon name="arrow" size="28rpx" color="#667085" />
        </view>
      </view>

      <!-- 类型一级筛选：下划线 tab 分段（对齐科技成果库；单维度 → tab 即维度） -->
      <view class="stage-wrap">
        <view class="stages">
          <view
            v-for="t in typeTabs"
            :key="t.value"
            class="stg"
            :class="{ on: activeType === t.value }"
            @tap="pickStageTab(t.value)"
          >
            <text>{{ t.label }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- ② 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 个职位</text>
      <text class="ir-hint">{{ activeType || '全部类型' }}</text>
    </view>

    <!-- ③ 骨架屏：首次加载 -->
    <view v-if="loading && list.length === 0" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w60"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w80"></view>
        </view>
      </view>
    </view>

    <!-- ④ 空 / 错误 -->
    <view v-else-if="!loading && list.length === 0" class="st">
      <u-empty :description="errorMsg || '暂无职位'">
        <view v-if="errorMsg" class="stb" @tap="fetchList(true)">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑤ 职位卡片列表 -->
    <view v-else class="cl">
      <view
        v-for="item in list"
        :key="item.id"
        class="job-card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="goDetail(item)"
      >
        <view class="job-top">
          <text class="job-title">{{ item.title }}</text>
          <text class="job-salary">{{ salaryText(item) }}</text>
        </view>
        <view class="job-tags">
          <text v-if="item.job_type" class="tag tag-blue">{{ item.job_type }}</text>
          <view v-if="item.location" class="job-loc">
            <u-icon name="location" size="24rpx" color="#667085" />
            <text>{{ item.location }}</text>
          </view>
        </view>
        <view class="job-foot">
          <text class="job-date">{{ formatDate(item.created_at) }} 发布</text>
          <view
            class="apply-btn"
            :class="{ applied: appliedIds.includes(item.id) }"
            hover-class="apply-press"
            :hover-stay-time="100"
            @tap.stop="applyJob(item)"
          >{{ appliedIds.includes(item.id) ? '已投递' : '投递' }}</view>
        </view>
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
import { request, getStoredUser, getErrorMessage } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'

export default {
  data() {
    return {
      noMotion: false,
      statusBarHeight: 20,
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
      submittingJobId: '',
      typeTabs: [
        { label: '全部', value: '' },
        { label: '全职', value: '全职' },
        { label: '兼职', value: '兼职' },
        { label: '实习', value: '实习' },
        { label: '项目制', value: '项目制' },
      ],
    }
  },
  onLoad() {
    this.checkMotion()
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
        this.errorMsg = getErrorMessage(e) || '网络异常，请稍后重试'
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
    clearSearch() {
      this.searchText = ''
      this.fetchList(true)
    },
    // 方案 A（对齐成果库）：非「全部」tab 再点取消；「全部」tab 清筛；单维度无 ▾ 面板，已在全部再点不动作
    pickStageTab(value) {
      if (value === '') {
        if (this.activeType === '') return // 已在全部，单维度无面板可开
        this.activeType = ''
      } else {
        this.activeType = this.activeType === value ? '' : value // 再点取消
      }
      this.fetchList(true)
    },

    // ---- 投递闭环 ----
    goMyJobs() {
      uni.navigateTo({ url: '/pkg-talent/pages/jobs/mine' })
    },
    goDetail(item) {
      uni.navigateTo({ url: '/pkg-talent/pages/jobs/detail?id=' + encodeURIComponent(item.id) })
    },
    goApplications() {
      uni.navigateTo({ url: '/pkg-talent/pages/jobs/applications' })
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
      // submittingJobId 防重复点击：同一职位提交中直接拦截
      if (this.submittingJobId || this.appliedIds.includes(item.id)) return
      if (!requireLogin()) return
      this.submittingJobId = item.id
      try {
        const resumes = await request({ url: '/api/v1/resumes/mine' })
        const rlist = Array.isArray(resumes) ? resumes : ((resumes && resumes.data) || [])
        if (!rlist.length) {
          uni.showModal({
            title: '需要简历',
            content: '投递职位需要一份简历，是否现在去创建？',
            success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pkg-talent/pages/jobs/resume' }) },
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
      } finally {
        this.submittingJobId = ''
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
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}

/* ===== 白底头部 ===== */
.head-zone { background: #fff; }

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 双层投影（全站 b-search 同款） ===== */
.sbar { padding: 20rpx 28rpx 12rpx; background: #fff; }
.b-search {
  height: 88rpx;
  padding: 0 24rpx;
  border: 1px solid #E4E7EC;
  border-radius: 16rpx;
  background: #fff;
  box-shadow: 0 2rpx 4rpx rgba(16, 24, 40, 0.06), 0 8rpx 24rpx rgba(16, 24, 40, 0.05);
  display: flex;
  align-items: center;
  gap: 16rpx;
  box-sizing: border-box;
}
.b-sinp { flex: 1; min-width: 0; height: 88rpx; background: transparent; font-size: 28rpx; color: #17212B; }
.b-ph { color: #667085; font-size: 28rpx; }
.b-sclr { padding: 16rpx; margin: -16rpx; display: flex; align-items: center; }
.b-sep { width: 1px; height: 30rpx; background: #E4E7EC; margin: 0 20rpx 0 8rpx; flex: none; }
.b-sbtn { flex: none; color: #0A66C2; font-size: 28rpx; font-weight: 600; line-height: 1; padding: 24rpx 0; }

/* ===== 双入口条：白卡片 ===== */
.entry-bar { display: flex; gap: 20rpx; padding: 4rpx 28rpx 0; }
.entry-item {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 20rpx;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
}
.entry-ico {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 700;
  flex: none;
}
.entry-ico-jobs { background: #EAF3FB; color: #0A66C2; }
.entry-ico-apps { background: #FFF0E6; color: #E96012; }
.entry-text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.entry-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.entry-sub { font-size: 22rpx; color: #667085; }

/* ===== 类型一级筛选：下划线 tab 分段（对齐科技成果库） ===== */
.stage-wrap { position: relative; z-index: 42; background: #fff; border-bottom: 1px solid #EEF1F4; }
.stages { display: flex; gap: 40rpx; padding: 4rpx 28rpx 16rpx; white-space: nowrap; } /* 5 个类型 tab 自然宽 < 750rpx，单行放下 */
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
  padding: 24rpx 32rpx 8rpx;
  font-size: 24rpx;
  color: #667085;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 700; }
.ir-hint { font-size: 20rpx; line-height: 1; padding: 6rpx 12rpx; border-radius: 8rpx; background: #EAF3FB; color: #0A66C2; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 20rpx; padding: 8rpx 32rpx 0; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 26rpx;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
}
.sk-row { display: flex; align-items: center; gap: 16rpx; }
.sk-tag { width: 112rpx; height: 36rpx; border-radius: 8rpx; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 16rpx; }
.sk-l { height: 24rpx; background: #EDF0F3; border-radius: 8rpx; animation: skPulse 1.4s linear infinite; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.stb { height: 72rpx; padding: 0 40rpx; border-radius: 12rpx; background: #0A66C2; color: #fff; font-size: 24rpx; line-height: 72rpx; }

/* ===== 职位卡片（对齐「我的发布」卡片规范：16rpx 圆角 + 细描边 + 轻投影） ===== */
.cl { display: flex; flex-direction: column; gap: 20rpx; padding: 8rpx 32rpx 32rpx; }
.job-card {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 26rpx;
  position: relative;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.job-card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.job-card:nth-child(1) { animation-delay: 80ms; }
.job-card:nth-child(2) { animation-delay: 100ms; }
.job-card:nth-child(3) { animation-delay: 120ms; }
.job-card:nth-child(4) { animation-delay: 140ms; }
.job-card:nth-child(5) { animation-delay: 160ms; }
.job-card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.tap-scale { transform: scale(0.97); opacity: 0.9; }

.job-top { display: flex; align-items: flex-start; justify-content: space-between; gap: 20rpx; }
.job-title { font-size: 30rpx; font-weight: 700; color: #17212B; line-height: 1.4; flex: 1; min-width: 0; }
.job-salary { font-size: 28rpx; font-weight: 700; color: #E96012; flex: none; }
.job-tags { display: flex; align-items: center; flex-wrap: wrap; gap: 12rpx 20rpx; }
.tag { border-radius: 8rpx; padding: 6rpx 12rpx; font-size: 20rpx; line-height: 1; }
.tag-blue { color: #0A66C2; background: #EAF3FB; }
.job-loc { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: #667085; }
.job-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin-top: 4rpx;
  padding-top: 20rpx;
  border-top: 1px solid #EEF1F4;
}
.job-date { font-size: 22rpx; color: #667085; }
.apply-btn {
  min-height: 72rpx;
  padding: 0 32rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.apply-btn.applied { background: #F1F3F5; color: #667085; }
.apply-press { transform: scale(0.95); opacity: 0.85; }

/* ===== 加载更多 ===== */
.load-more { text-align: center; padding: 24rpx 0 8rpx; }
.loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  font-size: 24rpx;
  color: #667085;
}
.no-more { color: #667085; font-size: 22rpx; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .job-card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .stg.on::after { animation: none; } /* 注线画出属位移，关闭 */

/* prefers-reduced-motion：装饰动画/过渡全关（对齐科技成果库） */
@media (prefers-reduced-motion: reduce) {
  .stg, .job-card { animation: none !important; transition: none !important; }
}
</style>
