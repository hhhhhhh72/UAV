<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- ① 白底导航（对齐应急资源页：返回 + 标题 + 胶囊 + 同步栏） -->
    <view class="nav-wrap" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="nav-bar">
        <view class="nav-back" hover-class="nav-press" :hover-stay-time="100" @click="goBack">
          <text class="nav-back-icon">‹</text>
        </view>
        <view class="nav-title-area">
          <text class="nav-title">调度记录</text>
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

    <!-- ② Tab 筛选条 -->
    <view class="filter-area">
      <scroll-view scroll-x :show-scrollbar="false" class="filter-scroll">
        <view class="filter-inner">
          <view
            v-for="(tab, i) in tabTitles"
            :key="i"
            class="filter-pill"
            :class="{ on: activeTabIndex === i }"
            @click="onTabChange(i)"
          >
            <text class="pill-text">{{ tab }}</text>
            <view v-if="i === 0" class="pill-count">{{ allList.length }}</view>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- ③ 时间轴卡片列表 -->
    <view v-if="loading && list.length === 0" class="loading-state">
      <view class="loading-inline"><view class="spinner" /><text>加载中...</text></view>
    </view>
    <view v-else-if="errorMsg && list.length === 0" class="state-view">
      <view class="state-ico">!</view>
      <text class="state-text">加载失败</text>
      <view class="retry-btn" @tap="fetchList(true)"><text>重新加载</text></view>
    </view>
    <view v-else-if="!loading && allList.length > 0 && list.length === 0" class="state-view">
      <view class="state-ico">◌</view>
      <text class="state-text">该状态下暂无记录</text>
    </view>
    <view v-else-if="!loading && list.length === 0" class="state-view">
      <view class="state-ico">◌</view>
      <text class="state-text">暂无调度记录</text>
    </view>

    <view v-else class="timeline-list" :key="'tl' + listFadeKey">
      <view
        v-for="(item, i) in list"
        :key="item.id"
        class="timeline-item"
        :style="{ animationDelay: (i * 80) + 'ms' }"
        @click="openDetail(item)"
      >
        <view class="timeline-dot" :class="'dot--' + statusKey(item.status)">
          <view class="dot-core" />
        </view>
        <view class="timeline-card" hover-class="card-press" :hover-stay-time="150">
          <view class="tc-top">
            <text class="tc-title">{{ item.event_desc || item.title || '未命名事件' }}</text>
            <view class="status-badge" :class="'badge--' + statusKey(item.status)">
              <view class="badge-dot" />
              <text class="badge-text">{{ statusLabel(item.status) }}</text>
            </view>
          </view>
          <view class="tc-row">
            <view class="tc-ico"><view class="tc-pin" /></view>
            <text class="tc-row-text">{{ item.location || '位置待定' }}</text>
          </view>
          <view class="tc-row">
            <view class="tc-ico"><view class="tc-clock"><view class="clock-hand hand-h" /><view class="clock-hand hand-m" /></view></view>
            <text class="tc-row-text">{{ formatDateTime(item.start_time || item.created_at) }}</text>
          </view>
          <view v-if="item.commander" class="tc-row">
            <view class="tc-ico"><view class="tc-person"><view class="person-head" /><view class="person-shoulder" /></view></view>
            <text class="tc-row-text">负责人：{{ item.commander }}</text>
          </view>
          <view v-if="item.result" class="tc-desc-wrap">
            <view class="tc-desc-bar" />
            <text class="tc-desc">{{ item.result }}</text>
          </view>
        </view>
      </view>
      <view class="timeline-end" />
    </view>

    <!-- ④ 详情抽屉 -->
    <view v-if="showDetail && activeDetail" class="mask" :class="{ 'mask--close': sheetClosing }" catchtouchmove="noop" @click="closeDetail">
      <view class="sheet" :class="{ 'sheet--close': sheetClosing }" @click.stop>
        <view class="sheet-handle" />
        <view class="sheet-head">
          <text class="sheet-title">事件详情</text>
          <view class="sheet-close" @click="closeDetail"><text class="sheet-close-x">×</text></view>
        </view>
        <scroll-view scroll-y class="sheet-body">
          <view class="detail-head">
            <view class="detail-icon" :style="eventIconStyle(activeDetail)">
              <text class="detail-icon-char">{{ eventIcon(activeDetail) }}</text>
            </view>
            <view class="detail-info">
              <text class="detail-title">{{ activeDetail.event_desc || activeDetail.title || '未命名事件' }}</text>
              <text class="detail-sub">{{ activeDetail.location || '位置待定' }} · {{ formatDateTime(activeDetail.start_time || activeDetail.created_at) }}</text>
              <view class="detail-badges">
                <view class="status-badge" :class="'badge--' + statusKey(activeDetail.status)">
                  <view class="badge-dot" />
                  <text class="badge-text">{{ statusLabel(activeDetail.status) }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 处置进度（恒显，含入场动画） -->
          <view class="progress-sec">
            <view class="progress-head">
              <text class="progress-label">处置进度</text>
              <text class="progress-val" :class="'pv--' + statusKey(activeDetail.status)">{{ progressFor(activeDetail) }}%</text>
            </view>
            <view class="progress-track">
              <view class="progress-fill" :class="'fill--' + statusKey(activeDetail.status)" :style="{ width: (progReady ? progressFor(activeDetail) : 0) + '%' }" />
            </view>
          </view>

          <!-- 信息栅格（单列堆叠） -->
          <view class="kv-list">
            <view class="kv-row"><text class="kv-k">事件编号</text><text class="kv-v">{{ shortId(activeDetail) }}</text></view>
            <view class="kv-row">
              <text class="kv-k">响应等级</text>
              <view class="level-badge" :class="levelBadge(activeDetail).cls"><text class="level-text">{{ levelBadge(activeDetail).text }}</text></view>
            </view>
            <view class="kv-row"><text class="kv-k">责任单位</text><text class="kv-v">{{ activeDetail.unit || '待定' }}</text></view>
            <view class="kv-row"><text class="kv-k">负责人</text><text class="kv-v">{{ activeDetail.commander || '暂无' }}</text></view>
            <view class="kv-row"><text class="kv-k">联系电话</text><text class="kv-v" :class="{ muted: !phoneFor(activeDetail) }">{{ phoneFor(activeDetail) || '暂无' }}</text></view>
            <view class="kv-row"><text class="kv-k">到场时间</text><text class="kv-v">{{ formatDateTime(activeDetail.start_time || activeDetail.created_at) }}</text></view>
          </view>

          <!-- 处置记录时间线 -->
          <view class="detail-sec">
            <text class="detail-sec-title">处置记录</text>
            <view class="stage-timeline">
              <view class="stage-item">
                <view class="stage-line">
                  <view class="stage-dot" :class="{ filled: stageIndex(activeDetail) >= 1 }" />
                  <view class="stage-bar" :class="{ filled: stageIndex(activeDetail) >= 2 }" />
                </view>
                <view class="stage-content">
                  <text class="stage-title">事件发生</text>
                  <text class="stage-time">{{ formatDateTime(activeDetail.created_at) }}</text>
                </view>
              </view>
              <view class="stage-item">
                <view class="stage-line">
                  <view class="stage-dot" :class="{ filled: stageIndex(activeDetail) >= 2 }" />
                  <view class="stage-bar" :class="{ filled: stageIndex(activeDetail) >= 3 }" />
                </view>
                <view class="stage-content">
                  <text class="stage-title">现场响应</text>
                  <text class="stage-time">{{ formatDateTime(activeDetail.start_time) }}</text>
                </view>
              </view>
              <view class="stage-item">
                <view class="stage-line">
                  <view class="stage-dot" :class="{ filled: stageIndex(activeDetail) >= 3 }" />
                </view>
                <view class="stage-content">
                  <text class="stage-title">完成处置</text>
                  <text class="stage-time">{{ stageEndLabel(activeDetail) }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 关联资源（横向缩略卡片） -->
          <view class="detail-sec">
            <text class="detail-sec-title">关联资源</text>
            <scroll-view scroll-x :show-scrollbar="false" class="res-scroll">
              <view class="res-cards">
                <view
                  v-for="r in relatedResources"
                  :key="r.id"
                  class="res-card"
                  hover-class="res-card-press"
                  :hover-stay-time="120"
                  @click="goResource(r)"
                >
                  <view class="drone-thumb">
                    <view class="drone-svg">
                      <view class="drone-prop p1" /><view class="drone-prop p2" /><view class="drone-prop p3" /><view class="drone-prop p4" />
                      <view class="drone-prop-conn c1" /><view class="drone-prop-conn c2" />
                      <view class="drone-body" />
                      <view class="drone-gimbal" />
                    </view>
                  </view>
                  <view class="res-type-row">
                    <view class="res-thumb-icon" :style="resIconStyle(r)"><text class="res-thumb-char">{{ resIconChar(r) }}</text></view>
                  </view>
                  <text class="res-card-name">{{ r.name }}</text>
                  <text class="res-card-qty">{{ resQty(r) }} 台 · {{ statusLabel(r.status) }}</text>
                </view>
                <view v-if="relatedResources.length === 0" class="res-empty">暂无关联资源</view>
              </view>
            </scroll-view>
          </view>
        </scroll-view>
        <view class="sheet-foot">
          <view class="btn btn--ghost" hover-class="btn-press" :hover-stay-time="120" @click="callContact(activeDetail.commander)">
            <text class="btn-text">联系负责人</text>
          </view>
          <view
            class="btn btn--primary"
            :class="{ 'btn--urgent': statusKey(activeDetail.status) === 'ongoing' }"
            hover-class="btn-press"
            :hover-stay-time="120"
            @click="viewScene(activeDetail)"
          >
            <text class="btn-text">查看现场</text>
          </view>
        </view>
      </view>
    </view>

    <!-- ⑤ 自定义 Toast -->
    <view v-if="toast.show" class="custom-toast" :class="{ 'custom-toast--out': toast.hide }">
      <text class="toast-text">{{ toast.msg }}</text>
    </view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'
import { safeBack } from '../../../utils/nav'

export default {
  data() {
    return {
      activeTabIndex: 0,
      noMotion: false,
      // 顶部状态栏高度：自定义导航需自行下移，避免与状态栏重叠
      statusBarHeight: 24,
      loading: false,
      errorMsg: '',
      allList: [],
      allResources: [],
      statusMap: ['', 'pending', 'ongoing', 'completed'],
      tabTitles: ['全部', '待响应', '进行中', '已完成'],
      listFadeKey: 0,
      showDetail: false,
      activeDetail: null,
      progReady: false,
      sheetClosing: false,
      toast: { show: false, hide: false, msg: '' },
      toastTimer: null,
      toastOutTimer: null,
    }
  },
  computed: {
    list() {
      var k = this.statusMap[this.activeTabIndex]
      if (!k) return this.allList
      var self = this
      return this.allList.filter(function (d) {
        return self.statusKey(d.status) === k
      })
    },
    relatedResources() {
      var d = this.activeDetail
      if (!d) return []
      // 服务端内嵌关联资源优先（后端 related 是单对象 EmergencyResourceBrief，兼容数组/对象两种形态）
      if (d.related && !Array.isArray(d.related)) return [d.related]
      if (d.related && d.related.length) return d.related
      if (!d.resource_id) return []
      var rid = String(d.resource_id)
      return this.allResources.filter(function (r) {
        return String(r.id) === rid
      })
    },
  },
  onLoad() {
    this.checkMotion()
    this.statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 24
    this.fetchList(true)
    this.fetchResources()
  },
  onPullDownRefresh() {
    var p = this.fetchList(true)
    p.then(function () {
      uni.stopPullDownRefresh()
    })
  },
  methods: {
    // 减弱动效（无障碍）：系统开启时装饰动画/位移全关
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
        this.loading = true
      }
      this.errorMsg = ''
      try {
        // 一次拉全量，Tab 由 computed list 客户端过滤（避免 dispatched 状态被精确筛选漏掉）
        var res = await request({ url: '/api/v1/emergency-dispatches' })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        var items = Array.isArray(data) ? data : (data && data.items) || []
        this.allList = items
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    async fetchResources() {
      try {
        var res = await request({ url: '/api/v1/emergency-resources', data: { page: 1, page_size: 100 } })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
        this.allResources = Array.isArray(data) ? data : (data && data.items) || []
      } catch (e) {
        this.allResources = []
      }
    },
    onTabChange(index) {
      this.activeTabIndex = index
      this.listFadeKey++
    },
    openDetail(item) {
      this.sheetClosing = false
      this.activeDetail = item
      this.showDetail = true
      // 进度条入场：宽度 0 → 目标值（触发 transition），150ms 延迟
      this.progReady = false
      var self = this
      setTimeout(function () {
        self.progReady = true
      }, 150)
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
    viewScene(item) {
      this.showCustomToast('现场画面对接中')
    },
    callContact(commander) {
      var phone = this.phoneFor(this.activeDetail)
      if (phone) {
        uni.makePhoneCall({ phoneNumber: phone })
      } else {
        this.showCustomToast('已联系 ' + (commander || '负责人'))
      }
    },
    goResource(r) {
      var self = this
      uni.navigateTo({
        url: '/pkg-emergency/pages/emergency/resources',
        fail: function () {
          self.showCustomToast('资源页未就绪')
        },
      })
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
    noop() {},
    showCityToast() {
      uni.showToast({ title: '当前仅支持重庆市', icon: 'none' })
    },
    /* 状态 */
    statusKey(status) {
      if (status === 'completed' || status === 'done') return 'completed'
      if (status === 'pending') return 'pending'
      if (status === 'ongoing' || status === 'dispatched') return 'ongoing'
      if (status === 'cancelled') return 'cancelled'
      return 'pending'
    },
    statusLabel(status) {
      var map = {
        pending: '待响应',
        dispatched: '进行中',
        completed: '已完成',
        ongoing: '进行中',
        done: '已完成',
        cancelled: '已取消',
        available: '可用',
        in_use: '调度中',
        maintenance: '离线',
      }
      return map[status] || status || '未知'
    },
    /* 事件图标（类型分色） */
    eventIcon(item) {
      var desc = item.event_desc || ''
      if (desc.indexOf('火') >= 0) return '火'
      if (desc.indexOf('照明') >= 0) return '光'
      if (desc.indexOf('通讯') >= 0) return '讯'
      if (desc.indexOf('搜索') >= 0 || desc.indexOf('搜救') >= 0 || desc.indexOf('救援') >= 0) return '救'
      if (desc.indexOf('洪水') >= 0 || desc.indexOf('暴雨') >= 0) return '水'
      if (desc.indexOf('巡检') >= 0 || desc.indexOf('检测') >= 0 || desc.indexOf('巡查') >= 0) return '巡'
      return '调'
    },
    eventIconStyle(item) {
      var desc = item.event_desc || ''
      if (desc.indexOf('火') >= 0) return { background: '#FEF3F2', color: '#D92D20' }
      if (desc.indexOf('照明') >= 0) return { background: '#FFF0E6', color: '#E96012' }
      if (desc.indexOf('通讯') >= 0) return { background: '#F4F8FC', color: '#0A66C2' }
      if (desc.indexOf('搜索') >= 0 || desc.indexOf('搜救') >= 0 || desc.indexOf('救援') >= 0) return { background: '#E9F7F0', color: '#168A55' }
      if (desc.indexOf('洪水') >= 0 || desc.indexOf('暴雨') >= 0) return { background: '#EAF3FB', color: '#0A66C2' }
      return { background: '#EAF3FB', color: '#0A66C2' }
    },
    /* 进度 */
    progressFor(item) {
      var k = this.statusKey(item.status)
      if (k === 'completed') return 100
      if (k === 'cancelled') return 0
      if (k === 'ongoing') {
        var p = Number(item.progress)
        if (!isNaN(p) && p > 0 && p < 100) return p
        return 50
      }
      return 10
    },
    stageIndex(item) {
      var k = this.statusKey(item.status)
      if (k === 'completed') return 3
      if (k === 'ongoing') return 2
      if (k === 'cancelled') return 0
      return 1
    },
    stageEndLabel(item) {
      var k = this.statusKey(item.status)
      if (k === 'completed') return '任务已闭环'
      if (k === 'cancelled') return '调度已终止'
      return '待推进'
    },
    /* 响应等级色块：特红 / 一橙 / 二蓝 / 三灰 */
    levelBadge(item) {
      var desc = item.event_desc || ''
      if (desc.indexOf('火') >= 0) return { text: '特级', cls: 'level--urgent' }
      if (desc.indexOf('洪水') >= 0 || desc.indexOf('暴雨') >= 0) return { text: '一级', cls: 'level--one' }
      if (desc.indexOf('搜索') >= 0 || desc.indexOf('搜救') >= 0) return { text: '一级', cls: 'level--one' }
      if (desc.indexOf('安保') >= 0) return { text: '三级', cls: 'level--three' }
      return { text: '二级', cls: 'level--two' }
    },
    phoneFor(item) {
      return item.phone || item.contact || ''
    },
    shortId(item) {
      var id = item.id || ''
      return id.length > 12 ? id.slice(-8).toUpperCase() : (id || '未编号')
    },
    /* 关联资源图标 */
    resIconChar(r) {
      var t = r.res_type || r.resource_type || 'drone'
      var map = { drone: '机', comm: '信', vehicle: '车', medical: '医', rescue: '救' }
      return map[t] || '他'
    },
    resIconStyle(r) {
      var t = r.res_type || r.resource_type || 'drone'
      var map = {
        drone: { background: '#EAF3FB', color: '#0A66C2' },
        comm: { background: '#F4F8FC', color: '#0A66C2' },
        vehicle: { background: '#FFF0E6', color: '#E96012' },
        medical: { background: '#E9F7F0', color: '#168A55' },
        rescue: { background: '#FEF6E7', color: '#B54708' },
      }
      return map[t] || { background: '#F4F8FC', color: '#0A66C2' }
    },
    resQty(r) {
      return r.quantity != null ? r.quantity : 1
    },
    /* 时间（兼容 iOS 不解析 'YYYY-MM-DD HH:mm' 的问题） */
    parseDate(s) {
      if (s instanceof Date) return s
      var str = String(s)
      var m = str.match(/^(\d{4})-(\d{1,2})-(\d{1,2})(?:[ T](\d{1,2}):(\d{1,2})(?::(\d{1,2}))?)?$/)
      if (m) {
        return new Date(+m[1], +m[2] - 1, +m[3], +(m[4] || 0), +(m[5] || 0), +(m[6] || 0))
      }
      var d = new Date(str)
      return isNaN(d.getTime()) ? null : d
    },
    formatDateTime(iso) {
      if (!iso) return ''
      var d = this.parseDate(iso)
      if (!d) return iso
      var p = function (n) { return (n < 10 ? '0' : '') + n }
      return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) + ' ' + p(d.getHours()) + ':' + p(d.getMinutes())
    },
    goBack() {
      safeBack()
    },
  },
}
</script>

<style scoped>
.page {
  --ease: cubic-bezier(0.16, 1, 0.3, 1);
  min-height: 100vh;
  background: #F4F6F8;
  padding-left: constant(safe-area-inset-left);
  padding-left: env(safe-area-inset-left);
  padding-right: constant(safe-area-inset-right);
  padding-right: env(safe-area-inset-right);
  padding-bottom: calc(constant(safe-area-inset-bottom) + 80rpx);
  padding-bottom: calc(env(safe-area-inset-bottom) + 80rpx);
  overflow-x: hidden;
}

/* ═══ ① 白底导航（对齐应急资源页）═══ */
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
.nav-title-area {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.nav-title { font-size: 34rpx; font-weight: 700; color: #17212B; }
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

/* ═══ ② Tab 筛选 ═══ */
.filter-area {
  background: #ffffff;
  padding: 16rpx 24rpx 20rpx;
}
.filter-scroll { width: 100%; white-space: nowrap; }
.filter-inner { display: inline-flex; gap: 12rpx; }
.filter-pill {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  min-height: 40px;
  padding: 0 14px;
  border-radius: 8px;
  background: #ffffff;
  border: 1px solid #E4E7EC;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.filter-pill.on {
  border-color: #0A66C2;
  color: #0A66C2;
  font-weight: 600;
  background: #F4F8FC;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
}
.pill-text { font-size: 12px; font-weight: 600; }
.pill-count {
  display: inline-flex;
  align-items: center;
  min-width: 18px;
  height: 16px;
  padding: 0 4px;
  border-radius: 999px;
  background: #F4F8FC;
  color: #0A66C2;
  font-size: 10px;
  font-weight: 600;
}
.filter-pill.on .pill-count { background: rgba(10, 102, 194, 0.12); color: #0A66C2; }

/* ═══ 状态视图 ═══ */
.loading-state { display: flex; justify-content: center; padding: 100rpx 0; }
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

/* ═══ ③ 时间轴列表 ═══ */
.timeline-list { padding: 8rpx 24rpx 40rpx; }
.timeline-item {
  position: relative;
  padding-left: 40rpx;
  animation: cardIn 420ms var(--ease) both;
}
.timeline-item::before {
  content: '';
  position: absolute;
  left: 11rpx;
  top: 30rpx;
  bottom: -10rpx;
  width: 2rpx;
  background: #E4E7EC;
}
.timeline-item:last-child::before { display: none; }

.timeline-dot {
  position: absolute;
  left: 0;
  top: 26rpx;
  width: 24rpx;
  height: 24rpx;
  border-radius: 50%;
  background: #ffffff;
  z-index: 2;
  animation: dotIn 240ms cubic-bezier(0.16, 1, 0.3, 1) both;
}
.dot-core {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 12rpx;
  height: 12rpx;
  margin-left: -6rpx;
  margin-top: -6rpx;
  border-radius: 50%;
}
.dot--pending .dot-core { background: #B54708; }
.dot--ongoing .dot-core { background: #168A55; }
.dot--completed .dot-core { background: #0A66C2; }
.dot--cancelled .dot-core { background: #D92D20; }

/* 卡片 */
.timeline-card {
  background: #ffffff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  padding: 14px;
  margin-bottom: 10px;
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.card-press { transform: scale(0.985); box-shadow: 0 2px 8px rgba(16, 24, 40, 0.08); }
.tc-top { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; margin-bottom: 12rpx; }
.tc-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tc-row { display: flex; align-items: center; gap: 10rpx; margin-bottom: 8rpx; }
.tc-row-text { font-size: 24rpx; color: #667085; }
.tc-desc-wrap {
  display: flex;
  align-items: stretch;
  gap: 8rpx;
  margin-top: 10rpx;
}
.tc-desc-bar {
  flex-shrink: 0;
  width: 4rpx;
  border-radius: 2rpx;
  background: #0A66C2;
  opacity: 0.85;
}
.tc-desc {
  flex: 1;
  min-width: 0;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  font-size: 22rpx;
  color: #667085;
  line-height: 1.5;
}

/* 图标（pin / clock） */
.tc-ico { flex-shrink: 0; width: 56rpx; display: flex; align-items: center; }
.tc-pin {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50% 50% 50% 0;
  background: #98A2B3;
  transform: rotate(-45deg);
  margin: 2rpx 4rpx 2rpx 2rpx;
}
.tc-clock {
  width: 22rpx;
  height: 22rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50%;
  position: relative;
  box-sizing: border-box;
}
.tc-clock .clock-hand {
  position: absolute;
  background: #98A2B3;
  border-radius: 1rpx;
  left: 50%;
  top: 50%;
  transform-origin: 50% 100%;
}
.tc-clock .hand-h {
  width: 2rpx;
  height: 6rpx;
  transform: translate(-50%, -100%) rotate(-45deg);
}
.tc-clock .hand-m {
  width: 2rpx;
  height: 8rpx;
  transform: translate(-50%, -100%) rotate(35deg);
}

/* 负责人人物图标（圆形头 + 肩线） */
.tc-person {
  width: 22rpx;
  height: 22rpx;
  position: relative;
  box-sizing: border-box;
}
.tc-person .person-head {
  position: absolute;
  left: 50%;
  top: 1rpx;
  width: 8rpx;
  height: 8rpx;
  margin-left: -4rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50%;
  box-sizing: border-box;
}
.tc-person .person-shoulder {
  position: absolute;
  left: 50%;
  top: 11rpx;
  width: 14rpx;
  height: 10rpx;
  margin-left: -7rpx;
  border: 2rpx solid #98A2B3;
  border-top: none;
  border-radius: 0 0 10rpx 10rpx;
  box-sizing: border-box;
}

/* 状态徽章（overflow hidden 裁剪扩散环） */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  flex-shrink: 0;
  overflow: hidden;
}
.badge-dot { width: 10rpx; height: 10rpx; border-radius: 50%; background: currentColor; position: relative; }
.badge--completed { background: #E8F2FC; color: #0A66C2; }
.badge--ongoing { background: #E9F7F0; color: #168A55; }
.badge--pending { background: #FEF6E7; color: #B54708; }
.badge--cancelled { background: #FEF3F2; color: #D92D20; }
.badge-text { font-size: 20rpx; font-weight: 600; }
.timeline-end { height: 20rpx; }

/* ═══ ④ 抽屉 ═══ */
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
  animation: sheetUp 280ms var(--ease) both;
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

/* 详情头 */
.detail-head { display: flex; gap: 16rpx; align-items: flex-start; }
.detail-icon {
  width: 96rpx;
  height: 96rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.detail-icon-char { font-size: 40rpx; font-weight: 700; }
.detail-info { flex: 1; min-width: 0; }
.detail-title { display: block; font-size: 32rpx; font-weight: 700; color: #17212B; line-height: 1.3; }
.detail-sub { display: block; font-size: 22rpx; color: #667085; margin-top: 6rpx; }
.detail-badges { margin-top: 12rpx; }

/* 进度条（恒显） */
.progress-sec { margin-top: 24rpx; }
.progress-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12rpx; }
.progress-label { font-size: 22rpx; color: #667085; font-weight: 600; }
.progress-val { font-size: 22rpx; font-weight: 700; }
.pv--completed { color: #0A66C2; }
.pv--ongoing { color: #168A55; }
.pv--pending { color: #B54708; }
.pv--cancelled { color: #98A2B3; }
.progress-track {
  height: 12rpx;
  border-radius: 999rpx;
  background: #EEF1F4;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  border-radius: 999rpx;
}
.fill--completed { background: #0A66C2; }
.fill--ongoing { background: #168A55; }
.fill--pending { background: #B54708; }
.fill--cancelled { background: #CBD2DA; }

/* 信息栅格（单列堆叠） */
.kv-list {
  margin-top: 12px;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  overflow: hidden;
}
.kv-row {
  display: flex;
  align-items: center;
  min-height: 72rpx;
  padding: 0 20rpx;
  box-sizing: border-box;
}
.kv-row + .kv-row { border-top: 1rpx solid #EEF1F4; }
.kv-k {
  width: 120rpx;
  flex-shrink: 0;
  font-size: 24rpx;
  font-weight: 500;
  color: #667085;
  white-space: nowrap;
}
.kv-v {
  flex: 1;
  min-width: 0;
  text-align: right;
  font-size: 26rpx;
  font-weight: 700;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kv-v.muted { color: #98A2B3; font-weight: 500; }

/* 响应等级色块：特红 / 一橙 / 二蓝 / 三灰 */
.level-badge {
  margin-left: auto;
  padding: 4rpx 14rpx;
  border-radius: 6rpx;
}
.level-text { font-size: 22rpx; font-weight: 700; }
.level--urgent { background: #FEF3F2; }
.level--urgent .level-text { color: #D92D20; }
.level--one { background: #FFF0E6; }
.level--one .level-text { color: #E96012; }
.level--two { background: #EAF3FB; }
.level--two .level-text { color: #0A66C2; }
.level--three { background: #F3F4F6; }
.level--three .level-text { color: #667085; }

/* 处置阶段时间线 */
.detail-sec { margin-top: 24rpx; }
.detail-sec-title {
  display: block;
  font-size: 22rpx;
  color: #667085;
  font-weight: 600;
  margin-bottom: 16rpx;
  padding-left: 16rpx;
  border-left: 4rpx solid #0A66C2;
}
.stage-item { display: flex; gap: 16rpx; }
.stage-line { display: flex; flex-direction: column; align-items: center; flex-shrink: 0; }
.stage-dot {
  width: 18rpx;
  height: 18rpx;
  border-radius: 50%;
  border: 2rpx solid #CBD2DA;
  background: #ffffff;
  margin-top: 2rpx;
  animation: dotIn 240ms cubic-bezier(0.16, 1, 0.3, 1) both;
}
.stage-item:nth-child(2) .stage-dot { animation-delay: 200ms; }
.stage-item:nth-child(3) .stage-dot { animation-delay: 400ms; }
.stage-dot.filled { background: #0A66C2; border-color: #0A66C2; box-shadow: 0 0 0 4rpx rgba(10, 102, 194, 0.18); }
@keyframes dotIn {
  from { transform: scale(0); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
.stage-bar {
  width: 2rpx;
  min-height: 40rpx;
  background: #E4E7EC;
}
.stage-bar.filled { background: #0A66C2; }
.stage-content { flex: 1; padding-bottom: 20rpx; }
.stage-title { display: block; font-size: 26rpx; font-weight: 600; color: #17212B; }
.stage-time { display: block; font-size: 20rpx; color: #98A2B3; margin-top: 4rpx; }

/* 关联资源（横向缩略卡片） */
.res-scroll { width: 100%; white-space: nowrap; }
.res-cards { display: inline-flex; gap: 12rpx; }
.res-card {
  display: inline-flex;
  flex-direction: column;
  width: 96px;
  background: #ffffff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  padding: 0 6px 6px;
  box-sizing: border-box;
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
  animation: resIn 500ms cubic-bezier(0.16, 1, 0.3, 1) 400ms backwards;
}
@keyframes resIn {
  from { opacity: 0; transform: translateY(6rpx); }
  to { opacity: 1; transform: translateY(0); }
}
.res-card-press { transform: scale(0.97); }
/* 4:3 缩略图：应急无人机线性 SVG（机身 + 四旋翼 + 云台） */
.drone-thumb {
  width: 100%;
  height: 128rpx;
  background: #EAF3FB;
  border-radius: 8rpx 8rpx 0 0;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}
.drone-svg {
  position: relative;
  width: 72rpx;
  height: 56rpx;
}
.drone-prop {
  position: absolute;
  width: 20rpx;
  height: 20rpx;
  border: 2rpx solid #0A66C2;
  border-radius: 50%;
  box-sizing: border-box;
  opacity: 0.9;
}
.drone-prop.p1 { left: 0; top: 0; }
.drone-prop.p2 { right: 0; top: 0; }
.drone-prop.p3 { left: 0; bottom: 0; }
.drone-prop.p4 { right: 0; bottom: 0; }
.drone-prop-conn {
  position: absolute;
  background: #0A66C2;
  height: 2rpx;
  opacity: 0.55;
}
.drone-prop-conn.c1 { left: 10rpx; top: 50%; width: 52rpx; transform: rotate(-45deg); }
.drone-prop-conn.c2 { left: 10rpx; top: 50%; width: 52rpx; transform: rotate(45deg); }
.drone-body {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 26rpx;
  height: 18rpx;
  margin: -9rpx 0 0 -13rpx;
  background: #0A66C2;
  border-radius: 5rpx;
}
.drone-gimbal {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 12rpx;
  height: 12rpx;
  margin: 8rpx 0 0 -6rpx;
  border: 2rpx solid #0A66C2;
  border-radius: 50%;
  box-sizing: border-box;
  opacity: 0.85;
}

/* 类型图标行（缩略图下方 32×32） */
.res-type-row {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 10rpx;
}
.res-thumb-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.res-thumb-char { font-size: 28rpx; font-weight: 700; }
.res-card-name {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  font-weight: 700;
  color: #17212B;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.res-card-qty {
  display: block;
  margin-top: 4rpx;
  font-size: 20rpx;
  color: #667085;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.res-empty { font-size: 24rpx; color: #98A2B3; text-align: center; padding: 20rpx 0; }

/* 详情抽屉内容 stagger 入场 */
.detail-head { animation: detailIn 380ms cubic-bezier(0.16, 1, 0.3, 1) both; }
.progress-sec { animation: detailIn 380ms cubic-bezier(0.16, 1, 0.3, 1) 70ms both; }
.kv-list { animation: detailIn 380ms cubic-bezier(0.16, 1, 0.3, 1) 140ms both; }
.detail-sec { animation: detailIn 380ms cubic-bezier(0.16, 1, 0.3, 1) 210ms both; }
.detail-sec + .detail-sec { animation-delay: 280ms; }
@keyframes detailIn {
  from { opacity: 0; transform: translateY(16rpx); }
  to { opacity: 1; transform: translateY(0); }
}

/* 抽屉底栏 */
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
.btn--urgent { background: #0A66C2; box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28); }

/* ═══ 动画 ═══ */
@keyframes cardIn {
  from { opacity: 0; transform: translateY(8rpx); }
  to { opacity: 1; transform: translateY(0); }
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
.mask--close { animation: fadeOut 250ms ease both; }
.sheet--close { animation: sheetDown 250ms var(--ease) both; }
@keyframes fadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ═══ ⑤ 自定义 Toast（对齐 pub-toast：底部黑底白字，无图标） ═══ */
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
  animation: toastIn 250ms cubic-bezier(0.16, 1, 0.3, 1) both;
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

/* ═══ 减弱动效（无障碍）：no-motion 时装饰动画全关 ═══ */
.page.no-motion .timeline-item,
.page.no-motion .dot--ongoing,
.page.no-motion .badge--ongoing .badge-dot,
.page.no-motion .mask,
.page.no-motion .sheet,
.page.no-motion .mask--close,
.page.no-motion .sheet--close,
.page.no-motion .progress-fill,
.page.no-motion .stage-dot,
.page.no-motion .res-card,
.page.no-motion .custom-toast {
  animation: none !important;
  transition: none !important;
}
</style>
