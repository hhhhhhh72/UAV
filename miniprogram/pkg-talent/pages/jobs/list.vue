<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="招聘求职" show-back :fixed="true" @back="goBack" />

    <!-- ① 白底头部：搜索 + 双入口 + 筛选 -->
    <view class="head-zone">
      <!-- 搜索框（白上白：双层投影浮起） -->
      <view class="sbar">
        <view class="b-search">
          <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
          <input
            class="b-sinp"
            v-model="searchText"
            placeholder="搜索职位名称 / 公司"
            placeholder-class="b-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <text v-if="searchText" class="b-sclr" @tap="clearSearch">×</text>
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
            <text class="entry-sub">企业发布与管理职位</text>
          </view>
          <text class="entry-arrow">›</text>
        </view>
        <view class="entry-item" hover-class="tap-scale" :hover-stay-time="100" @tap="goApplications">
          <view class="entry-ico entry-ico-apps">投</view>
          <view class="entry-text">
            <text class="entry-title">我的投递</text>
            <text class="entry-sub">跟踪投递进展</text>
          </view>
          <text class="entry-arrow">›</text>
        </view>
      </view>

      <!-- 筛选胶囊：全部 / 全职 / 兼职 / 实习 / 项目制 -->
      <view class="fbar">
        <view
          v-for="t in typeTabs"
          :key="t.value"
          class="fpill"
          :class="{ on: activeType === t.value }"
          hover-class="fpill-press"
          :hover-stay-time="100"
          @tap="onTabChange(t)"
        >
          <text class="fpv">{{ t.label }}</text>
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
        class="card job-card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="goDetail(item)"
      >
        <view class="job-top">
          <text class="job-title">{{ item.title }}</text>
          <text class="job-salary">{{ salaryText(item) }}</text>
        </view>
        <view class="job-tags">
          <text v-if="item.job_type" class="job-type">{{ item.job_type }}</text>
          <view v-if="item.location" class="job-loc"><view class="loc-pin" /><text>{{ item.location }}</text></view>
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
    onTabChange(t) {
      if (this.activeType === t.value) return
      this.activeType = t.value
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
      if (this.appliedIds.includes(item.id)) return
      if (!requireLogin()) return
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

/* ===== 双入口条：白卡片 ===== */
.entry-bar { display: flex; gap: 10px; padding: 4px 12px 0; }
.entry-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  padding: 12px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}
.entry-ico {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 700;
  flex-shrink: 0;
}
.entry-ico-jobs { background: #EAF3FB; color: #0A66C2; }
.entry-ico-apps { background: #FFF4EC; color: #E96012; }
.entry-text { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.entry-title { font-size: 14px; font-weight: 700; color: #17212B; }
.entry-sub { font-size: 11px; color: #98A2B3; margin-top: 2px; }
.entry-arrow { font-size: 15px; color: #98A2B3; flex: none; }

/* ===== 筛选胶囊 ===== */
.fbar { display: flex; gap: 8px; padding: 10px 12px 4px; background: #fff; }
.fpill {
  flex: 1;
  min-width: 0;
  min-height: 40px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  overflow: hidden;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpill-press { transform: scale(0.95); opacity: 0.85; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

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
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 职位卡片 ===== */
.cl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px 12px; }
.card {
  display: flex;
  flex-direction: column;
  gap: 7px;
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

.job-top { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; }
.job-title { font-size: 15px; font-weight: 700; color: #17212B; line-height: 1.4; flex: 1; min-width: 0; }
.job-salary { font-size: 14px; font-weight: 700; color: #C2410C; flex-shrink: 0; }
.job-tags { display: flex; align-items: center; gap: 8px; margin-top: 2px; }
.job-type {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 4px;
  background: #EAF3FB;
  color: #0A66C2;
  font-weight: 600;
}
.job-loc { display: flex; align-items: center; gap: 4px; font-size: 12px; color: #667085; }
.loc-pin {
  width: 7px;
  height: 7px;
  border: 1.5px solid #98A2B3;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  flex: none;
}
.job-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
  padding-top: 10px;
  border-top: 1px solid #F0F1F3;
}
.job-date { font-size: 11px; color: #98A2B3; }
.apply-btn {
  padding: 5px 18px;
  border-radius: 6px;
  background: #0A66C2;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}
.apply-btn.applied { background: #EEF1F4; color: #667085; }
.apply-press { transform: scale(0.95); opacity: 0.85; }

/* ===== 加载更多 ===== */
.load-more { text-align: center; padding: 12px 0; }
.loading-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: #667085;
}
.no-more { color: #98A2B3; font-size: 12px; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
</style>
