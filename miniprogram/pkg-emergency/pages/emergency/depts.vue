<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="部门对接" show-back :fixed="true" @back="goBack" />

    <!-- ① 骨架屏：首次加载 -->
    <view v-if="loading" class="skl">
      <view v-for="i in 2" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w60"></view></view>
        <view class="sk-bd">
          <view class="sk-l w80"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- ② 空 / 错误 -->
    <view v-else-if="errorMsg && depts.length === 0 && drills.length === 0" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>
    <view v-else-if="!loading && depts.length === 0 && drills.length === 0" class="st">
      <u-empty description="暂无数据" />
    </view>

    <!-- ③ 正常内容 -->
    <template v-else>
      <!-- 对接部门 -->
      <view v-if="depts.length > 0" class="section">
        <view class="section-header">
          <text class="section-title">对接部门</text>
          <text class="section-count">{{ depts.length }} 个</text>
        </view>
        <view class="dp-list">
          <view
            v-for="item in depts"
            :key="item.id"
            class="card dp-card"
            hover-class="tap-scale"
            :hover-stay-time="100"
          >
            <view class="dp-icon" :style="deptIconStyle(item.type || item.name)"><text>{{ deptIcon(item.type || item.name) }}</text></view>
            <view class="dp-info">
              <view class="dp-name-row">
                <text class="dp-name">{{ item.name }}</text>
                <text class="agree-tag" :class="item.agreement_status === '已签署' ? 'agree-tag--yes' : 'agree-tag--no'">{{ item.agreement_status === '已签署' ? '已签署' : '未签署' }}</text>
              </view>
              <view class="dp-meta">
                <text v-if="item.contact_name" class="dp-meta-item">{{ item.contact_name }}</text>
                <text v-if="item.contact_phone" class="dp-meta-item">{{ item.contact_phone }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 演练记录 -->
      <view v-if="drills.length > 0" class="section">
        <view class="section-header">
          <text class="section-title">演练记录</text>
          <text class="section-count">{{ drills.length }} 条</text>
        </view>
        <view class="card tl-card">
          <view
            v-for="(item, idx) in drills"
            :key="item.id || idx"
            class="tl-item"
          >
            <view class="tl-line">
              <view class="tl-dot" :class="{ active: idx === 0 }" />
              <view v-if="idx < drills.length - 1" class="tl-bar" />
            </view>
            <view class="tl-content">
              <text class="tl-date">{{ formatDate(item.date || item.created_at) }}</text>
              <text class="tl-event">{{ item.event_name || item.title }}</text>
              <text v-if="item.description" class="tl-desc">{{ item.description }}</text>
            </view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      noMotion: false,
      statusBarHeight: 20,
      loading: false,
      errorMsg: '',
      depts: [],
      drills: [],
    }
  },
  onLoad() {
    this.checkMotion()
    this.fetchData()
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

    async fetchData() {
      this.loading = true
      this.errorMsg = ''

      try {
        var results = await Promise.all([
          request({ url: '/api/v1/emergency-depts' }),
          request({ url: '/api/v1/emergency-drills' }),
        ])

        var deptRes = results[0]
        var drillRes = results[1]

        var deptData = Array.isArray(deptRes) ? deptRes : (deptRes && deptRes.data) || deptRes || {}
        var drillData = Array.isArray(drillRes) ? drillRes : (drillRes && drillRes.data) || drillRes || {}

        var deptItems = Array.isArray(deptData) ? deptData : (deptData && deptData.items) || (deptData && deptData.list) || []
        var drillItems = Array.isArray(drillData) ? drillData : (drillData && drillData.items) || (drillData && drillData.list) || []

        this.depts = deptItems
        this.drills = drillItems
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    /* 部门类型字符图标（低饱和色块，非 emoji） */
    deptIcon(name) {
      var nameStr = (name || '').toLowerCase()
      if (nameStr.indexOf('消防') !== -1) return '防'
      if (nameStr.indexOf('公安') !== -1) return '警'
      if (nameStr.indexOf('医疗') !== -1 || nameStr.indexOf('卫生') !== -1) return '医'
      if (nameStr.indexOf('应急') !== -1) return '应'
      return '部'
    },
    deptIconStyle(name) {
      var nameStr = (name || '').toLowerCase()
      if (nameStr.indexOf('消防') !== -1) return { background: '#FFF0E6', color: '#E96012' }
      if (nameStr.indexOf('公安') !== -1) return { background: '#EAF3FB', color: '#0A66C2' }
      if (nameStr.indexOf('医疗') !== -1 || nameStr.indexOf('卫生') !== -1) return { background: '#E9F7F0', color: '#168A55' }
      if (nameStr.indexOf('应急') !== -1) return { background: '#EAF3FB', color: '#0A66C2' }
      return { background: '#F4F6F8', color: '#667085' }
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: calc(env(safe-area-inset-bottom) + 24rpx);
}

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 12px; }
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
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== Section ===== */
.section { margin-bottom: 16px; }

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
}
.section-title { font-size: 14px; font-weight: 700; color: #17212B; }
.section-count { font-size: 12px; color: #98A2B3; }

/* ===== 部门卡片 ===== */
.dp-list { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.dp-card { animation: cardIn .22s ease-out backwards; }
.tap-scale { transform: scale(0.97); opacity: 0.9; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.dp-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.dp-info { flex: 1; min-width: 0; }

.dp-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.dp-name { font-size: 14px; font-weight: 600; color: #17212B; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.agree-tag {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}
.agree-tag--yes { background: #E9F7F0; color: #0B6B41; }
.agree-tag--no { background: #FDECEC; color: #B42318; }

.dp-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}
.dp-meta-item { font-size: 12px; color: #667085; }

/* ===== 演练记录时间线 ===== */
.tl-card {
  margin: 0 12px;
  padding: 4px 14px;
  display: block;
  animation: cardIn .22s ease-out backwards;
  animation-delay: 80ms;
}

.tl-item {
  display: flex;
  align-items: flex-start;
  padding-bottom: 10px;
}

.tl-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-right: 10px;
  flex-shrink: 0;
}

.tl-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #D0D5DD;
  flex-shrink: 0;
  margin-top: 6px;
}

.tl-dot.active {
  background: #0A66C2;
  box-shadow: 0 0 0 3px rgba(10, 102, 194, 0.15);
}

.tl-bar {
  width: 2px;
  flex: 1;
  background: #EEF1F4;
  margin-top: 6px;
  min-height: 100%;
}

.tl-content {
  flex: 1;
  padding-bottom: 2px;
}

.tl-date {
  font-size: 11px;
  color: #98A2B3;
  display: block;
  margin-bottom: 2px;
}

.tl-event {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
  display: block;
  line-height: 1.4;
}

.tl-desc {
  font-size: 12px;
  color: #667085;
  display: block;
  margin-top: 2px;
  line-height: 1.5;
}

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .dp-card,
.page.no-motion .tl-card { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
</style>
