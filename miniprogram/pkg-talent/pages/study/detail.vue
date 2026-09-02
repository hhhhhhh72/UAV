<template>
  <view class="study-detail-page" v-if="tour">
    <!-- ═══ 顶部导航（对齐课程详情：独立导航条，白圆钮返回 + 居中标题） ═══ -->
    <view class="detail-nav" :style="{ paddingTop: statusBarH + 'px' }">
      <view class="detail-nav-back" hover-class="detail-nav-back--press" :hover-stay-time="120" aria-role="button" aria-label="返回" @click="goBack">
        <text>‹</text>
      </view>
      <text class="detail-nav-title">研学详情</text>
      <view class="detail-nav-balance" />
    </view>

    <!-- ═══ 一、Hero（对齐课程详情：内嵌圆角卡片，封面 + 蒙层 + 状态徽章 + 信息贴底） ═══ -->
    <view class="hero">
      <image v-if="tour.cover_image" :src="tour.cover_image" mode="aspectFill" class="hero-img" lazy-load />
      <view v-else class="hero-fallback">
        <image src="/static/images/study/cover-1.jpg" mode="aspectFill" class="hero-fallback-img" />
      </view>

      <view class="hero-mask" />

      <!-- 状态徽章（对齐课程：白底胶囊 + 彩色文字） -->
      <view class="status-pill" :style="statusPillStyle">{{ statusLabel[tour.status] || '招募中' }}</view>

      <!-- Hero 底部信息（对齐课程：标题 + 主题 + 周期 meta 行） -->
      <view class="hero-bottom">
        <text class="hero-title">{{ tour.title || '研学活动' }}</text>
        <text class="hero-org">{{ themeInfo.label }}</text>
        <view class="hero-meta-row">
          <view class="meta-ico meta-ico--cal"><view class="cal-top" /><view class="cal-body"><view class="cal-line l1" /><view class="cal-line l2" /><view class="cal-line l3" /></view></view>
          <text class="hero-meta-text">{{ dateRange }}</text>
        </view>
      </view>
    </view>

    <view class="content" v-if="contentReady">
      <!-- ═══ 二、活动信息卡（四列 + 报名状态提示）═══ -->
      <view class="section-card card-float">
        <view class="info-head">
          <text class="section-title">活动信息</text>
          <text class="info-status-pill" :style="statusPillStyle">{{ statusLabel[tour.status] || '招募中' }}</text>
        </view>
        <view class="info-grid">
          <view class="info-cell">
            <view class="info-tile info-tile--purple"><view class="info-ico info-ico--clock" /></view>
            <text class="info-num">{{ tour.duration || '-' }}</text>
            <text class="info-label">研学时长</text>
          </view>
          <view class="info-divider" />
          <view class="info-cell">
            <view class="info-tile info-tile--green"><view class="info-ico info-ico--users" /></view>
            <text class="info-num info-num-green">{{ capacityText }}</text>
            <text class="info-label">招募名额</text>
          </view>
          <view class="info-divider" />
          <view class="info-cell">
            <view class="info-tile info-tile--orange"><view class="info-ico info-ico--loc" /></view>
            <text class="info-num">{{ shortLoc }}</text>
            <text class="info-label">研学地点</text>
          </view>
          <view class="info-divider" />
          <view class="info-cell">
            <view class="info-tile info-tile--gold"><view class="info-ico info-ico--rmb" /></view>
            <text class="info-num info-num-price">{{ priceNum }}</text>
            <text class="info-label">{{ priceNum !== '面议' ? '/人' : '' }}</text>
          </view>
        </view>
        <!-- 报名情况提示（名额/截止，随活动信息卡；CSS 时钟图标非 emoji） -->
        <view class="deadline-pill" v-if="deadlineText">
          <view class="deadline-clock"><view class="clock-hand" /></view>
          <text class="deadline-text">{{ deadlineText }}</text>
        </view>
      </view>

      <!-- ═══ 三、研学介绍卡 ═══ -->
      <view class="section-card" v-if="tour.description">
        <text class="section-title">研学介绍</text>
        <text class="section-text">{{ tour.description }}</text>
      </view>

      <!-- ═══ 四、行程安排卡 ═══ -->
      <view class="section-card" v-if="schedule.length > 0">
        <text class="section-title">行程安排</text>
        <view class="timeline">
          <view class="tl-item" v-for="(d, i) in schedule" :key="i">
            <view class="tl-node" :style="{ background: nodeColors[i % nodeColors.length] }" />
            <view class="tl-line" v-if="i < schedule.length - 1" />
            <view class="tl-content">
              <text class="tl-day">{{ d.title }}</text>
              <text v-for="(item, j) in d.items" :key="j" class="tl-text">{{ item }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- ═══ 五、活动时间卡 ═══ -->
      <view class="section-card">
        <text class="section-title">活动时间</text>
        <view class="time-block">
          <view class="time-item">
            <view class="time-row-label">
              <view class="time-ico time-ico--start" />
              <text class="time-label">开始</text>
            </view>
            <text class="time-value">{{ fullDate(tour.start_date) }}</text>
          </view>
          <view class="time-item">
            <view class="time-row-label">
              <view class="time-ico time-ico--end" />
              <text class="time-label">结束</text>
            </view>
            <text class="time-value">{{ fullDate(tour.end_date) }}</text>
          </view>
        </view>
      </view>

      <!-- ═══ 六、费用说明卡 ═══ -->
      <view class="section-card">
        <text class="section-title">费用说明</text>
        <view class="fee-section">
          <text class="fee-subtitle">费用包含</text>
          <view class="fee-row" v-for="(f, i) in feeInclude" :key="'in'+i">
            <view class="fee-mark fee-mark-ok" />
            <text class="fee-text">{{ f }}</text>
          </view>
        </view>
        <view class="fee-section fee-exclude" v-if="feeExclude.length > 0">
          <text class="fee-subtitle">费用不含</text>
          <view class="fee-row" v-for="(f, i) in feeExclude" :key="'ex'+i">
            <view class="fee-mark fee-mark-no" />
            <text class="fee-text fee-text-muted">{{ f }}</text>
          </view>
        </view>
      </view>

      <!-- ═══ 七、温馨提示卡 ═══ -->
      <view class="section-card">
        <text class="section-title">温馨提示</text>
        <view class="tips-box">
          <view class="tip-row" v-for="(t, i) in tips" :key="i">
            <text class="tip-index">{{ i + 1 }}</text>
            <text class="tip-text">{{ t }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-else class="skeleton-wrap">
      <view class="skeleton-block" />
      <view class="skeleton-block" />
    </view>

    <!-- ═══ 八、底部 CTA（分享 + 立即报名，平分） ═══ -->
    <view class="action-bar">
      <button class="share-btn" open-type="share" hover-class="share-btn--hover" :hover-stay-time="120">
        <view class="share-ico" />
        <text class="share-btn-text">分享</text>
      </button>
      <button class="apply-btn" :disabled="!recruiting" @tap="onApply">{{ recruiting ? '立即报名' : (statusLabel[tour.status] || '未开放报名') }}</button>
    </view>
  </view>
</template>

<script setup>
import { safeBack } from '../../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad, onReady, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'

const contentReady = ref(false)
const tour = ref(null)
const goBack = () => safeBack()

// 自定义导航：返回按钮下沉到状态栏下方（JS 方式）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }

const statusLabel = { active: '招募中', draft: '即将开始', closed: '已结束' }
/* 状态徽章（对齐课程详情：浅底胶囊 + 彩色文字） */
const statusPillStyle = computed(() => {
  const s = (tour.value && tour.value.status) || 'active'
  if (s === 'closed') return { background: '#EEF1F4', color: '#5D6B82' }
  if (s === 'draft') return { background: '#FFF3E4', color: '#E96012' }
  return { background: '#E9F7F0', color: '#0B6B41' }
})

// ── 主题推断（与列表页一致）──
const THEMES = [
  { key: ['职业', '院校', '学院', '专业', '开放日'], label: '职业研学', color: '#FF8E3C', bg: 'rgba(255,142,60,.1)' },
  { key: ['实践', '实训', '训练营', '穿越机', '巡检', '应急救援', '测绘', '实战', '试飞'], label: '实践研学', color: '#0EA5E9', bg: 'rgba(14,165,233,.1)' },
  { key: ['科普', '科技', '科学', '航模', '体验', '组装'], label: '科普研学', color: '#1E5EFF', bg: 'rgba(30,94,255,.1)' },
  { key: ['产业', '低空经济', '企业', '龙头'], label: '产业研学', color: '#8B5CF6', bg: 'rgba(139,92,246,.1)' },
]
const themeInfo = computed(() => {
  const t = tour.value
  if (!t) return { label: '研学', color: '#1E5EFF', bg: 'rgba(30,94,255,.1)' }
  const title = t.title || ''
  const desc = t.description || ''
  for (const x of THEMES) {
    if (x.key.some((k) => title.includes(k))) return x
  }
  for (const x of THEMES) {
    if (x.key.some((k) => desc.includes(k))) return x
  }
  return { label: '研学', color: '#1E5EFF', bg: 'rgba(30,94,255,.1)' }
})

// ── 招募 / 名额 ──
const recruiting = computed(() => (tour.value ? tour.value.status === 'active' : false))
const capacityText = computed(() => {
  const c = tour.value && tour.value.capacity
  return c != null && c > 0 ? `${c}` : '不限'
})

// ── 时间 ──
const fmtDate = (v, withYear) => {
  if (!v) return ''
  const d = new Date(v)
  if (isNaN(d.getTime())) return ''
  if (d.getFullYear() <= 1) return ''
  const p = (n) => String(n).padStart(2, '0')
  if (withYear) return `${d.getFullYear()}年${p(d.getMonth() + 1)}月${p(d.getDate())}日`
  return `${p(d.getMonth() + 1)}月${p(d.getDate())}日`
}
const dateRange = computed(() => {
  const t = tour.value
  if (!t) return '时间待定'
  const ss = fmtDate(t.start_date, true)
  const ee = fmtDate(t.end_date, false)
  if (ss && ee) return `${ss}-${ee}`
  return ss || '时间待定'
})
const fullDate = (v) => {
  if (!v) return '待定'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '待定'
  if (d.getFullYear() <= 1) return '待定'
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}年${p(d.getMonth() + 1)}月${p(d.getDate())}日 ${p(d.getHours())}:${p(d.getMinutes())}`
}

// ── 地点 ──
const locationText = computed(() => {
  const t = tour.value
  if (!t) return '地点待定'
  return t.location || t.destination || '地点待定'
})
const shortLoc = computed(() => {
  const t = tour.value
  if (!t) return '-'
  const loc = t.location || t.destination || ''
  if (!loc) return '-'
  const m = loc.match(/[·区县](\S{2,3}[区县])/)
  return m ? m[1] : loc.slice(0, 4)
})

// ── 价格（优先后端 price_fen（分）字段，缺失显示"面议"）──
const priceNum = computed(() => {
  const t = tour.value
  if (!t) return '面议'
  const fen = t.price_fen
  if (fen == null || fen <= 0) return '面议'
  const yuan = Number(fen) / 100
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
})

// ── 行程安排（后端 schedule 字段，数组或 JSON 字符串都容错；缺失则隐藏行程区块）──
const nodeColors = ['#0A66C2', '#7C3AED', '#0B6B41', '#E96012']
const schedule = computed(() => {
  const t = tour.value
  if (!t || t.schedule == null) return []
  let arr = t.schedule
  if (typeof arr === 'string') {
    try {
      arr = JSON.parse(arr)
    } catch (e) {
      return []
    }
  }
  if (!Array.isArray(arr) || arr.length === 0) return []
  return arr
    .map((d, i) => {
      if (typeof d === 'string') return { title: 'Day ' + (i + 1), items: [d] }
      if (d && typeof d === 'object') {
        const items = Array.isArray(d.items) ? d.items.slice() : []
        const single = d.text || d.content
        if (single) items.push(single)
        return { title: d.title || d.day || 'Day ' + (i + 1), items: items }
      }
      return null
    })
    .filter((d) => d && Array.isArray(d.items) && d.items.length > 0)
})

// ── 报名截止（后端无独立截止字段时仅展示通用提示，不编造日期）──
const deadlineText = computed(() => {
  const t = tour.value
  if (!t) return ''
  const dl = t.deadline || t.register_deadline || t.enroll_deadline
  if (dl && new Date(dl).getFullYear() > 1) {
    return `报名截止 ${fmtDate(new Date(dl).getTime(), true)}`
  }
  return '名额有限 · 报满即止'
})

// ── 费用说明（优先后端 fee_include/fee_exclude；缺失显示"费用面议"，不编造）──
const feeInclude = computed(() => {
  const t = tour.value
  if (t && Array.isArray(t.fee_include) && t.fee_include.length > 0) return t.fee_include
  return ['费用面议']
})
const feeExclude = computed(() => {
  const t = tour.value
  if (t && Array.isArray(t.fee_exclude) && t.fee_exclude.length > 0) return t.fee_exclude
  return []
})

// ── 温馨提示 ──
const tips = [
  '研学活动名额有限，报名前请确认行程安排',
  '建议年龄 8-16 岁，需家长签署知情同意书',
  '如遇恶劣天气，活动将顺延并提前通知',
  '报名后 24 小时内可申请全额退款',
]

// ── 交互：跳研学报名页（专用报名页，提交 POST /api/v1/study/tours/{id}/enroll）──
const onApply = () => {
  if (!recruiting.value) return
  const t = tour.value
  if (!t || !t.id) {
    uni.showToast({ title: '活动信息缺失，请返回列表重新进入', icon: 'none' })
    return
  }
  uni.navigateTo({ url: '/pkg-talent/pages/study/enroll?id=' + encodeURIComponent(t.id) })
}

onLoad((options) => {
  const id = options && options.id ? decodeURIComponent(options.id) : ''
  // 冷启动/分享直达：storage 可能为空或为上次浏览的其它活动——
  // 先按 id 走公开详情接口自取（此前无接口，直达即误判"不存在"并返回）。
  if (id) {
    request({ url: '/api/v1/study/tours/' + encodeURIComponent(id) })
      .then((res) => {
        const d = (res && res.data) || res
        if (d && d.id) {
          tour.value = d
        } else {
          showMissing()
        }
      })
      .catch(() => {
        // 接口失败：回退缓存（仅当缓存 id 匹配），否则按不存在处理
        const cached = uni.getStorageSync('study_tour_detail')
        if (cached && cached.id === id) {
          tour.value = cached
        } else {
          showMissing()
        }
      })
      .finally(() => {
        uni.removeStorageSync('study_tour_detail')
        uni.setNavigationBarTitle({ title: (tour.value && tour.value.title) || '研学详情' })
      })
    return
  }
  // 兼容旧入口（无 id 参数时仅用缓存）
  const cached = uni.getStorageSync('study_tour_detail')
  if (cached && cached.id) {
    tour.value = cached
    uni.removeStorageSync('study_tour_detail')
    uni.setNavigationBarTitle({ title: cached.title || '研学详情' })
  } else {
    showMissing()
  }
})

const showMissing = () => {
  uni.showToast({ title: '研学活动不存在或已下架', icon: 'none' })
  setTimeout(() => uni.navigateBack(), 1200)
}

onReady(() => {
  setTimeout(() => { contentReady.value = true }, 150)
})

/* 分享（底部"分享"按钮 / 右上角菜单） */
onShareAppMessage(() => {
  const t = tour.value || {}
  return {
    title: '研学活动：' + (t.title || '低空研学'),
    path: '/pkg-talent/pages/study/detail?id=' + encodeURIComponent(t.id || ''),
  }
})
</script>

<style scoped>
.study-detail-page {
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: calc(180rpx + env(safe-area-inset-bottom));
}

/* ═══ 顶部导航（对齐课程详情：白圆钮返回 + 居中标题） ═══ */
.detail-nav {
  position: relative;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-left: 24rpx;
  padding-right: 24rpx;
  box-sizing: content-box;
  background: #F5F8FC;
}
.detail-nav-back {
  width: 60rpx;
  height: 60rpx;
  flex: 0 0 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #ffffff;
  box-shadow: 0 6rpx 16rpx rgba(31, 89, 169, 0.13);
}
.detail-nav-back--press { transform: scale(0.94); opacity: 0.86; }
.detail-nav-back text { margin-top: -4rpx; color: #1A3353; font-size: 42rpx; line-height: 1; }
.detail-nav-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  max-width: 56%;
  display: block;
  color: #17212B;
  font-size: 34rpx;
  font-weight: 700;
  text-align: center;
  line-height: 88rpx;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.detail-nav-balance { width: 60rpx; height: 60rpx; flex: 0 0 60rpx; }

/* ═══ 一、Hero（对齐课程详情：内嵌圆角卡片） ═══ */
.hero {
  position: relative;
  width: auto;
  height: 348rpx;
  margin: 0 24rpx;
  border-radius: 24rpx;
  overflow: hidden;
  background: linear-gradient(145deg, #163C66 0%, #0A66C2 100%);
  box-shadow: 0 14rpx 34rpx rgba(31, 89, 169, 0.2);
}
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.hero-fallback { position: absolute; inset: 0; }
.hero-fallback-img { width: 100%; height: 100%; }
.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(4, 30, 68, 0.08) 0%, rgba(4, 30, 68, 0.05) 34%, rgba(4, 30, 68, 0.8) 100%);
  pointer-events: none;
}

/* 状态徽章（白底胶囊 + 彩色文字，颜色由 statusPillStyle 内联控制） */
.status-pill {
  position: absolute;
  top: 18rpx;
  left: 18rpx;
  padding: 7rpx 16rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: none;
  z-index: 4;
  font-size: 20rpx;
  font-weight: 650;
  color: #0A66C2;
}

/* Hero 底部信息（标题 + 主题 + 周期 meta 行） */
.hero-bottom {
  position: absolute;
  left: 24rpx;
  right: 24rpx;
  bottom: 24rpx;
  z-index: 3;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.hero-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.35;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.32);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.hero-org {
  display: block;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.78);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.hero-meta-row { display: flex; align-items: center; gap: 8rpx; }
.hero-meta-text { font-size: 22rpx; color: rgba(255, 255, 255, 0.85); }
.meta-ico { width: 26rpx; height: 26rpx; flex-shrink: 0; position: relative; }
.meta-ico--cal { border: 2rpx solid rgba(255, 255, 255, 0.9); border-radius: 4rpx; box-sizing: border-box; }
.meta-ico--cal::before, .meta-ico--cal::after { content: ''; position: absolute; top: 4rpx; width: 2rpx; height: 5rpx; background: rgba(255, 255, 255, 0.9); }
.meta-ico--cal::before { left: 6rpx; }
.meta-ico--cal::after { right: 6rpx; }
.meta-ico--cal .cal-top { position: absolute; left: 3rpx; right: 3rpx; top: 8rpx; height: 2rpx; background: rgba(255, 255, 255, 0.9); }
.meta-ico--cal .cal-line { position: absolute; left: 5rpx; right: 5rpx; height: 2rpx; background: rgba(255, 255, 255, 0.9); opacity: 0.6; }
.meta-ico--cal .cal-line.l1 { top: 14rpx; }
.meta-ico--cal .cal-line.l2 { top: 18rpx; }
.meta-ico--cal .cal-line.l3 { top: 22rpx; }

/* ═══ 内容区（直接在页面底色排布，卡片语言对齐课程） ═══ */
.content {
  position: relative;
  margin-top: 20rpx;
  z-index: 2;
}
.section-card {
  background: #ffffff;
  margin: 0 24rpx 20rpx;
  padding: 24rpx;
  border-radius: 16px;
  border: 1rpx solid #E8EDF3;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.06);
  animation: cardIn 0.4s ease both;
}
/* 卡片依次入场（stagger） */
.section-card:nth-child(1) { animation-delay: 0.08s; }
.section-card:nth-child(2) { animation-delay: 0.14s; }
.section-card:nth-child(3) { animation-delay: 0.2s; }
.section-card:nth-child(4) { animation-delay: 0.26s; }
.section-card:nth-child(5) { animation-delay: 0.32s; }
.section-card:nth-child(6) { animation-delay: 0.38s; }
.card-float { margin-top: 0; }

/* 卡片头部：标题 + 状态 pill（对齐课程信息卡） */
.info-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14rpx;
  margin-bottom: 18rpx;
}
.info-head .section-title { margin-bottom: 0; }
.info-status-pill {
  flex-shrink: 0;
  padding: 5rpx 14rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 650;
  background: #EEF1F4;
  color: #5D6B82;
}

.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.3;
  margin-bottom: 16rpx;
}

/* ═══ 二、活动信息 ═══ */
.info-grid {
  display: flex;
  align-items: center;
  text-align: center;
  padding: 8rpx 0;
}
.info-cell { flex: 1; }
.info-tile {
  width: 56rpx;
  height: 56rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 10rpx;
}
.info-tile--purple { background: #F3F0FF; }
.info-tile--green { background: #E9F7F0; }
.info-tile--orange { background: #FFF0E6; }
.info-tile--gold { background: #FEF6E7; }
.info-ico { width: 30rpx; height: 30rpx; background-size: contain; background-repeat: no-repeat; background-position: center; }
.info-ico--clock { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%237C3AED' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='12' cy='12' r='9'/%3E%3Cpath d='M12 7v5l3.5 2'/%3E%3C/svg%3E"); }
.info-ico--users { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230B6B41' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='9' cy='8' r='3.5'/%3E%3Cpath d='M3 20a6 6 0 0 1 12 0'/%3E%3Ccircle cx='17' cy='9' r='2.5'/%3E%3Cpath d='M14.5 20a5.5 5.5 0 0 1 6-5.4'/%3E%3C/svg%3E"); }
.info-ico--loc { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23E96012' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M12 21s-7-6.5-7-11a7 7 0 0 1 14 0c0 4.5-7 11-7 11z'/%3E%3Ccircle cx='12' cy='10' r='2.5'/%3E%3C/svg%3E"); }
.info-ico--rmb { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23D97706' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M12 4v16M7 8l5 4 5-4M7 13h10M7 16h10'/%3E%3C/svg%3E"); }
.info-num {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  display: block;
  animation: numPop 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.info-num-green { color: #0B6B41; }
.info-num-price { color: #E96012; }
.info-label { font-size: 20rpx; color: #7A8798; margin-top: 4rpx; display: block; }
.info-divider {
  width: 1rpx;
  height: 56rpx;
  background: linear-gradient(180deg, rgba(10, 102, 194, 0), rgba(10, 102, 194, 0.12) 50%, rgba(10, 102, 194, 0));
}

/* ═══ 三、研学介绍 ═══ */
.section-text {
  font-size: 26rpx;
  color: #2C3E50;
  line-height: 1.7;
  display: block;
}

/* ═══ 四、行程时间轴 ═══ */
.timeline {
  padding-left: 8rpx;
}
.tl-item {
  position: relative;
  display: flex;
  gap: 20rpx;
  padding-bottom: 24rpx;
  animation: tlItemIn 0.45s ease both;
}
/* 时间轴节点依次点亮（stagger 120ms） */
.tl-item:nth-child(1) { animation-delay: 0.2s; }
.tl-item:nth-child(2) { animation-delay: 0.32s; }
.tl-item:nth-child(3) { animation-delay: 0.44s; }
.tl-item:nth-child(4) { animation-delay: 0.56s; }
.tl-node {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  margin-top: 10rpx;
  flex-shrink: 0;
  z-index: 1;
  animation: tlNodePop 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.tl-item:nth-child(1) .tl-node { animation-delay: 0.3s; }
.tl-item:nth-child(2) .tl-node { animation-delay: 0.42s; }
.tl-item:nth-child(3) .tl-node { animation-delay: 0.54s; }
.tl-item:nth-child(4) .tl-node { animation-delay: 0.66s; }
.tl-line {
  position: absolute;
  left: 7rpx;
  top: 26rpx;
  bottom: 0;
  width: 2rpx;
  background: linear-gradient(180deg, #0A66C2, #00E0FF);
  opacity: 0.3;
}
.tl-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.tl-day {
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
}
.tl-text {
  font-size: 24rpx;
  color: #667085;
  line-height: 1.6;
}

/* ═══ 五、活动时间 ═══ */
.time-block { display: flex; flex-direction: column; gap: 16rpx; }
.time-item { display: flex; align-items: center; justify-content: space-between; }
.time-label { font-size: 26rpx; color: #7A8798; }
.time-value { font-size: 26rpx; font-weight: 600; color: #17212B; }
.deadline-pill {
  margin-top: 16rpx;
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  background: #FDECEC;
  font-size: 22rpx;
  color: #D92D20;
  font-weight: 600;
}
.deadline-clock {
  width: 24rpx;
  height: 24rpx;
  border: 2rpx solid #EF4444;
  border-radius: 50%;
  position: relative;
  box-sizing: border-box;
}
.deadline-clock::before {
  content: '';
  position: absolute;
  top: -4rpx;
  left: 4rpx;
  right: 4rpx;
  height: 2rpx;
  background: #EF4444;
  border-radius: 2rpx;
}
.clock-hand {
  position: absolute;
  left: 50%;
  top: 4rpx;
  bottom: 4rpx;
  width: 2rpx;
  background: #EF4444;
  transform-origin: bottom;
  transform: rotate(45deg);
}

/* ═══ 六、费用说明 ═══ */
.fee-section { padding-bottom: 8rpx; }
.fee-exclude {
  margin-top: 8rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #E8EDF3;
}
.fee-subtitle {
  font-size: 26rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 14rpx;
  display: block;
}
.fee-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 12rpx;
}
.fee-mark {
  width: 28rpx;
  height: 28rpx;
  border-radius: 50%;
  flex-shrink: 0;
}
.fee-mark-ok {
  background: #0B6B41;
  position: relative;
}
.fee-mark-ok::after {
  content: '';
  position: absolute;
  left: 8rpx;
  top: 5rpx;
  width: 10rpx;
  height: 14rpx;
  border: solid #fff;
  border-width: 0 3rpx 3rpx 0;
  transform: rotate(45deg);
}
.fee-mark-no {
  background: #E5E7EB;
  position: relative;
}
.fee-mark-no::before,
.fee-mark-no::after {
  content: '';
  position: absolute;
  left: 6rpx;
  top: 13rpx;
  width: 16rpx;
  height: 3rpx;
  background: #9CA3AF;
}
.fee-mark-no::before { transform: rotate(45deg); }
.fee-mark-no::after { transform: rotate(-45deg); }
.fee-text { font-size: 24rpx; color: #344054; }
.fee-text-muted { color: #98A2B3; }

/* ═══ 七、温馨提示 ═══ */
.tips-box {
  background: linear-gradient(135deg, #FFF9EC, #FFF3DD);
  border-radius: 12px;
  padding: 20rpx;
}
.tip-row {
  display: flex;
  gap: 12rpx;
  align-items: flex-start;
  margin-bottom: 12rpx;
}
.tip-row:last-child { margin-bottom: 0; }
.tip-index {
  width: 28rpx;
  height: 28rpx;
  border-radius: 50%;
  background: #F59E0B;
  color: #fff;
  font-size: 20rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.tip-text {
  font-size: 24rpx;
  color: #92400E;
  line-height: 1.6;
  flex: 1;
}

/* ═══ 八、底部 CTA（对齐课程详情：毛白底 + 上投影 + 橙价 + 蓝按钮） ═══ */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid #E8EDF3;
  box-shadow: 0 -6rpx 18rpx rgba(16, 24, 40, 0.06);
}
.share-btn {
  flex: 1;
  height: 76rpx;
  border-radius: 12px;
  border: 2rpx solid #A6C9EE;
  background: #fff;
  color: #0A66C2;
  font-size: 28rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  padding: 0;
  box-sizing: border-box;
}
.share-btn::after { border: none; }
.share-btn--hover { background: #EAF3FB; }
.share-ico {
  width: 30rpx;
  height: 30rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230A66C2' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7'/%3E%3Cpath d='M12 15V3M8 7l4-4 4 4'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}
.share-btn-text { color: #0A66C2; }
.apply-btn {
  flex: 1;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: 12px;
  font-weight: 700;
  font-size: 28rpx;
  color: #fff;
  background: #0A66C2;
  border: none;
  padding: 0;
  box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28);
}
.apply-btn:active { background: #0759AA; }
.apply-btn[disabled] {
  background: #C9CDD4 !important;
  box-shadow: none;
}

/* ═══ 骨架屏 ═══ */
.skeleton-wrap { padding: 20px; }
.skeleton-block {
  height: 120px;
  background: #EDF0F3;
  border-radius: 8px;
  margin-bottom: 16px;
  animation: blink 1.5s infinite;
}

/* ═══ 时间行图标 ═══ */
.time-row-label { display: flex; align-items: center; gap: 10rpx; }
.time-ico { width: 26rpx; height: 26rpx; flex-shrink: 0; background-size: contain; background-repeat: no-repeat; background-position: center; }
.time-ico--start { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230A66C2' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Crect x='3' y='5' width='18' height='16' rx='2'/%3E%3Cpath d='M8 3v4M16 3v4M3 10h18'/%3E%3C/svg%3E"); }
.time-ico--end { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230A66C2' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Crect x='3' y='5' width='18' height='16' rx='2'/%3E%3Cpath d='M8 3v4M16 3v4M3 10h18'/%3E%3Cpath d='M9 16l2 2 4-4'/%3E%3C/svg%3E"); }

/* ═══ 微动效 ═══ */
@keyframes cardIn {
  from {
    transform: translateY(20rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
@keyframes numPop {
  from {
    transform: scale(0.8);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
/* 行程时间轴 */
@keyframes tlItemIn {
  from {
    transform: translateY(16rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
@keyframes tlNodePop {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  60% {
    transform: scale(1.3);
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}
@keyframes blink {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}
</style>
