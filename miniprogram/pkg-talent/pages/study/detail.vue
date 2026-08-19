<template>
  <view class="study-detail-page" v-if="tour">
    <!-- ═══ 一、Hero 区：有真实图用图，无图兜底蓝色渐变 ═══ -->
    <view class="hero">
      <!-- 真实封面图（有则显示，作为背景） -->
      <image v-if="tour.cover_image" :src="tour.cover_image" mode="aspectFill" class="hero-bg" />
      <!-- 无图兜底：本地封面图 -->
      <image v-else src="/static/images/study/cover-1.jpg" mode="aspectFill" class="hero-bg" />
      <view class="hero-mask-top" />
      <view class="hero-mask-bottom" />

      <!-- 返回按钮 -->
      <view class="back-btn" :style="backStyle" hover-class="back-btn-hover" :hover-stay-time="120" @tap="goBack">
        <text class="back-icon">‹</text>
      </view>

      <!-- 状态徽章（左上角下方） -->
      <view class="status-pill" :style="statusPillStyle">{{ statusLabel[tour.status] || '招募中' }}</view>

      <!-- 主标题 -->
      <view class="hero-title-wrap">
        <text class="hero-title">{{ tour.title || '研学活动' }}</text>
        <!-- 主题胶囊（左下角白底） -->
        <view class="theme-pill" :style="{ color: themeInfo.color, background: themeInfo.bg }">{{ themeInfo.label }}</view>
        <!-- 信息行 -->
        <view class="hero-meta">
          <view class="meta-item">
            <text class="meta-ico">日</text>
            <text class="meta-text">{{ dateRange }}</text>
          </view>
          <view class="meta-item">
            <text class="meta-ico">点</text>
            <text class="meta-text">{{ locationText }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="content" v-if="contentReady">
      <!-- ═══ 二、活动信息卡（四列）═══ -->
      <view class="section-card card-float">
        <view class="section-title"><view class="title-bar" />活动信息</view>
        <view class="info-grid">
          <view class="info-cell">
            <view class="info-icon info-icon-purple"><view class="icon-clock" /></view>
            <text class="info-num">{{ tour.duration || '-' }}</text>
            <text class="info-label">研学时长</text>
          </view>
          <view class="info-divider" />
          <view class="info-cell">
            <view class="info-icon info-icon-green"><view class="icon-users" /></view>
            <text class="info-num info-num-green">{{ capacityText }}</text>
            <text class="info-label">招募名额</text>
          </view>
          <view class="info-divider" />
          <view class="info-cell">
            <view class="info-icon info-icon-orange"><view class="icon-loc" /></view>
            <text class="info-num">{{ shortLoc }}</text>
            <text class="info-label">研学地点</text>
          </view>
          <view class="info-divider" />
          <view class="info-cell">
            <view class="info-icon info-icon-gold"><text class="icon-rmb">¥</text></view>
            <text class="info-num info-num-price">{{ priceNum }}</text>
            <text class="info-label">/人</text>
          </view>
        </view>
      </view>

      <!-- ═══ 三、研学介绍卡 ═══ -->
      <view class="section-card" v-if="tour.description">
        <view class="section-title"><view class="title-bar" />研学介绍</view>
        <text class="section-text">{{ tour.description }}</text>
      </view>

      <!-- ═══ 四、行程安排卡 ═══ -->
      <view class="section-card" v-if="schedule.length > 0">
        <view class="section-title"><view class="title-bar" />行程安排</view>
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
        <view class="section-title"><view class="title-bar" />活动时间</view>
        <view class="time-block">
          <view class="time-item">
            <text class="time-label">开始</text>
            <text class="time-value">{{ fullDate(tour.start_date) }}</text>
          </view>
          <view class="time-item">
            <text class="time-label">结束</text>
            <text class="time-value">{{ fullDate(tour.end_date) }}</text>
          </view>
        </view>
        <view class="deadline-pill" v-if="deadlineText">
          <text class="deadline-clock">⏰</text>
          <text class="deadline-text">{{ deadlineText }}</text>
        </view>
      </view>

      <!-- ═══ 六、费用说明卡 ═══ -->
      <view class="section-card">
        <view class="section-title"><view class="title-bar" />费用说明</view>
        <view class="fee-section">
          <text class="fee-subtitle">费用包含</text>
          <view class="fee-row" v-for="(f, i) in feeInclude" :key="'in'+i">
            <view class="fee-mark fee-mark-ok" />
            <text class="fee-text">{{ f }}</text>
          </view>
        </view>
        <view class="fee-section fee-exclude">
          <text class="fee-subtitle">费用不含</text>
          <view class="fee-row" v-for="(f, i) in feeExclude" :key="'ex'+i">
            <view class="fee-mark fee-mark-no" />
            <text class="fee-text fee-text-muted">{{ f }}</text>
          </view>
        </view>
      </view>

      <!-- ═══ 七、温馨提示卡 ═══ -->
      <view class="section-card">
        <view class="section-title"><view class="title-bar" />温馨提示</view>
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

    <!-- ═══ 八、底部 CTA 栏（蓝色）═══ -->
    <view class="action-bar">
      <view class="price-area">
        <text class="price-label">研学费用</text>
        <view class="price-row">
          <text class="price-symbol">¥</text>
          <text class="price-num">{{ priceNum }}</text>
          <text class="price-unit">/人</text>
        </view>
      </view>
      <button class="apply-btn" :disabled="!recruiting" @tap="onApply">{{ recruiting ? '立即报名' : (statusLabel[tour.status] || '未开放报名') }}</button>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReady } from '@dcloudio/uni-app'

const contentReady = ref(false)
const tour = ref(null)
const goBack = () => uni.navigateBack()

// 自定义导航：返回按钮下沉到状态栏下方（JS 方式）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const backStyle = computed(() => ({ top: (38 + statusBarH.value) + 'px' })) // 原 76rpx = 38px

const statusLabel = { active: '招募中', draft: '即将开始', closed: '已结束' }
const statusPillStyle = computed(() => {
  const s = (tour.value && tour.value.status) || 'active'
  if (s === 'closed') return { background: 'rgba(255,255,255,0.2)', color: '#fff' }
  if (s === 'draft') return { background: 'rgba(239,68,68,0.9)', color: '#fff' }
  return { background: '#FF8E3C', color: '#fff' }
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

// ── 价格（后端暂无 price 字段，前端按时长兜底；补字段后替换）──
const priceNum = computed(() => {
  const d = (tour.value && tour.value.duration) || ''
  if (d.includes('3天') || d.includes('3 天')) return '1280'
  if (d.includes('2天') || d.includes('2 天')) return '880'
  if (d.includes('1天') || d.includes('1 天')) return '480'
  return '980'
})

// ── 行程安排（后端无行程字段，按天数组装通用模板；补字段后替换）──
const nodeColors = ['#1E5EFF', '#8B5CF6', '#00C896', '#FF8E3C']
const schedule = computed(() => {
  const t = tour.value
  if (!t) return []
  const d = (t.duration || '').match(/(\d+)\s*天/)
  const days = d ? parseInt(d[1]) : 2
  const base = t.title || '研学'
  const sd = t.start_date
  const startDate = sd && new Date(sd).getFullYear() > 1 ? new Date(sd) : null
  const list = []
  for (let i = 1; i <= days; i++) {
    const dateStr = startDate ? fmtDate(startDate.getTime() + (i - 1) * 86400000, false) : ''
    list.push({
      title: `Day ${i}${dateStr ? ' · ' + dateStr : ''}`,
      items: i === 1
        ? [`开营仪式 + ${base}主题讲解`, '团队破冰 + 分组', '参观无人机展示区']
        : i === days
          ? ['户外实操 / 成果展示', '研学总结 + 结营仪式', '颁发研学证书']
          : ['专业课程教学', '分组实操练习', '晚间交流分享'],
    })
  }
  return list
})

// ── 倒计时（报名截止）──
const deadlineText = computed(() => {
  const t = tour.value
  if (!t) return ''
  const sd = t.start_date
  if (sd && new Date(sd).getFullYear() > 1) {
    const cutoff = new Date(new Date(sd).getTime() - 3 * 86400000)
    return `报名截止 ${fmtDate(cutoff.getTime(), true)}`
  }
  return '名额有限 · 报满即止'
})

// ── 费用说明 ──
const feeInclude = computed(() => {
  const days = parseInt(((tour.value && tour.value.duration) || '2天').match(/(\d+)/)[1])
  const night = Math.max(days - 1, 1)
  return [
    `${days}天${night}晚住宿（标间）`,
    '全程餐饮（正餐 + 加餐）',
    '无人机课程材料包',
    '研学结业证书',
  ]
})
const feeExclude = ['往返交通费用', '个人消费及保险']

// ── 温馨提示 ──
const tips = [
  '研学活动名额有限，报名前请确认行程安排',
  '建议年龄 8-16 岁，需家长签署知情同意书',
  '如遇恶劣天气，活动将顺延并提前通知',
  '报名后 24 小时内可申请全额退款',
]

// ── 交互 ──
const onApply = () => {
  if (!recruiting.value) return
  uni.navigateTo({ url: '/pkg-service/pages/services/apply?id=9' })
}

onLoad((options) => {
  const cached = uni.getStorageSync('study_tour_detail')
  if (cached && cached.id) {
    tour.value = cached
  } else {
    uni.showToast({ title: '研学活动不存在或已下架', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1200)
    return
  }
  uni.setNavigationBarTitle({ title: (tour.value && tour.value.title) || '研学详情' })
})

onReady(() => {
  setTimeout(() => { contentReady.value = true }, 150)
})
</script>

<style scoped>
.study-detail-page {
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: 140rpx;
}

/* ═══ 一、Hero（蓝色系）═══ */
.hero {
  position: relative;
  height: 440rpx;
  overflow: hidden;
  background: linear-gradient(135deg, #0A1F44 0%, #1E5EFF 100%);
}
.hero-bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.hero-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(255,255,255,0.08) 2rpx, transparent 2rpx);
  background-size: 40rpx 40rpx;
}
.hero-radar {
  position: absolute;
  top: 60rpx;
  right: -60rpx;
  width: 320rpx;
  height: 320rpx;
  border-radius: 50%;
  border: 1rpx solid rgba(255,255,255,0.12);
}
.hero-radar::after {
  content: '';
  position: absolute;
  top: 18rpx;
  left: 18rpx;
  right: 18rpx;
  bottom: 18rpx;
  border-radius: 50%;
  border: 1rpx solid rgba(255,255,255,0.08);
}
.hero-mask-top {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 80rpx;
  background: linear-gradient(180deg, rgba(10,31,68,0.55), transparent);
}
.hero-mask-bottom {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 180rpx;
  background: linear-gradient(180deg, transparent, rgba(10,31,68,0.85));
}

/* 返回按钮 */
.back-btn {
  position: absolute;
  left: 24rpx;
  z-index: 5;
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: rgba(255,255,255,0.15);
  border: 1rpx solid rgba(255,255,255,0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}
.back-btn-hover { background: rgba(255,255,255,0.3); }
.back-icon {
  font-size: 44rpx;
  color: #fff;
  font-weight: 300;
  line-height: 1;
}

/* 状态徽章 */
.status-pill {
  position: absolute;
  top: 92rpx;
  left: 104rpx;
  z-index: 5;
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 700;
}

/* 标题区 */
.hero-title-wrap {
  position: absolute;
  left: 32rpx;
  right: 32rpx;
  bottom: 48rpx;
  z-index: 4;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}
.hero-title {
  font-size: 44rpx;
  font-weight: 700;
  color: #ffffff;
  text-shadow: 0 2rpx 8rpx rgba(0,0,0,0.25);
  line-height: 1.3;
}
.theme-pill {
  display: inline-block;
  align-self: flex-start;
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 600;
}
.hero-meta {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-top: 4rpx;
}
.meta-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
}
.meta-ico {
  width: 28rpx;
  height: 28rpx;
  border-radius: 50%;
  background: rgba(255,255,255,0.18);
  font-size: 20rpx;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.meta-text {
  font-size: 24rpx;
  color: rgba(255,255,255,0.85);
}

/* ═══ 内容区 ═══ */
.content {
  position: relative;
  margin-top: -40rpx;
  z-index: 2;
}
.section-card {
  background: #ffffff;
  margin: 16rpx 20rpx;
  padding: 24rpx;
  border-radius: 16rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
  animation: cardIn 0.4s ease both;
}
/* 卡片依次入场（stagger 60ms） */
.section-card:nth-child(1) { animation-delay: 0.1s; }
.section-card:nth-child(2) { animation-delay: 0.16s; }
.section-card:nth-child(3) { animation-delay: 0.22s; }
.section-card:nth-child(4) { animation-delay: 0.28s; }
.section-card:nth-child(5) { animation-delay: 0.34s; }
.section-card:nth-child(6) { animation-delay: 0.4s; }
.section-card:nth-child(7) { animation-delay: 0.46s; }
.card-float {
  margin-top: 0;
  border-radius: 16rpx 16rpx 8rpx 8rpx;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 12rpx;
  font-size: 30rpx;
  font-weight: 700;
  color: #0A1F44;
  margin-bottom: 20rpx;
}
.title-bar {
  width: 6rpx;
  height: 28rpx;
  border-radius: 3rpx;
  background: linear-gradient(180deg, #1E5EFF, #0A66C2);
}

/* ═══ 二、活动信息 ═══ */
.info-grid {
  display: flex;
  align-items: center;
  text-align: center;
  padding: 8rpx 0;
}
.info-cell { flex: 1; }
.info-icon {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 10rpx;
}
.info-icon-purple { background: #F3E8FF; }
.info-icon-green { background: #D1FAE5; }
.info-icon-orange { background: #FED7AA; }
.info-icon-gold { background: #FEF3C7; }
.info-num {
  font-size: 30rpx;
  font-weight: 700;
  color: #0A1F44;
  display: block;
  animation: numPop 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.info-num-green { color: #00C896; }
.info-num-price { color: #FF8E3C; }
.info-label { font-size: 20rpx; color: #6B7B95; margin-top: 4rpx; display: block; }
.info-divider {
  width: 1rpx;
  height: 56rpx;
  background: linear-gradient(180deg, rgba(30,94,255,0), rgba(30,94,255,0.12) 50%, rgba(30,94,255,0));
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
  background: linear-gradient(180deg, #1E5EFF, #00E5FF);
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
  color: #0A1F44;
}
.tl-text {
  font-size: 24rpx;
  color: #6B7B95;
  line-height: 1.6;
}

/* ═══ 五、活动时间 ═══ */
.time-block { display: flex; flex-direction: column; gap: 16rpx; }
.time-item { display: flex; align-items: center; justify-content: space-between; }
.time-label { font-size: 26rpx; color: #6B7B95; }
.time-value { font-size: 26rpx; font-weight: 600; color: #0A1F44; }
.deadline-pill {
  margin-top: 16rpx;
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  background: #FEE2E2;
  font-size: 22rpx;
  color: #EF4444;
  font-weight: 600;
}
.deadline-clock { font-size: 24rpx; }

/* ═══ 六、费用说明 ═══ */
.fee-section { padding-bottom: 8rpx; }
.fee-exclude {
  margin-top: 8rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #E8EEF7;
}
.fee-subtitle {
  font-size: 26rpx;
  font-weight: 700;
  color: #0A1F44;
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
  background: #00C896;
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
.fee-text { font-size: 24rpx; color: #2C3E50; }
.fee-text-muted { color: #ADB8C7; }

/* ═══ 七、温馨提示 ═══ */
.tips-box {
  background: linear-gradient(135deg, #FFFBEB, #FEF3C7);
  border-radius: 12rpx;
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

/* ═══ 八、底部 CTA ═══ */
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
  background: #ffffff;
  border-top: 1rpx solid #E8EEF7;
  box-shadow: 0 -2rpx 8rpx rgba(0,0,0,0.04);
}
.price-area { flex-shrink: 0; }
.price-label {
  font-size: 20rpx;
  color: #6B7B95;
  display: block;
}
.price-row {
  display: flex;
  align-items: baseline;
  gap: 4rpx;
}
.price-symbol { font-size: 24rpx; color: #FF8E3C; font-weight: 700; }
.price-num { font-size: 44rpx; color: #FF8E3C; font-weight: 800; line-height: 1; }
.price-unit { font-size: 22rpx; color: #ADB8C7; }
.apply-btn {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 999rpx;
  font-weight: 700;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, #1E5EFF, #0A66C2);
  border: none;
  padding: 0;
  box-shadow: 0 8rpx 24rpx rgba(30,94,255,0.35);
  animation: ctaGlow 2.5s ease-in-out infinite;
}
.apply-btn[disabled] {
  background: #C8C9CC !important;
  box-shadow: none;
  animation: none;
}

/* ═══ 骨架屏 ═══ */
.skeleton-wrap { padding: 20px; }
.skeleton-block {
  height: 120px;
  background: #eee;
  border-radius: 8px;
  margin-bottom: 16px;
  animation: blink 1.5s infinite;
}

/* ═══ CSS 图标 ═══ */
.icon-clock {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  border: 4rpx solid #8B5CF6;
  position: relative;
}
.icon-clock::before {
  content: '';
  position: absolute;
  left: 7rpx;
  top: -2rpx;
  width: 4rpx;
  height: 10rpx;
  background: #8B5CF6;
}
.icon-clock::after {
  content: '';
  position: absolute;
  left: 2rpx;
  top: 7rpx;
  width: 10rpx;
  height: 4rpx;
  background: #8B5CF6;
}
.icon-users {
  width: 24rpx;
  height: 16rpx;
  border: 4rpx solid #00C896;
  border-radius: 50% 50% 4rpx 4rpx;
  margin-top: -6rpx;
}
.icon-loc {
  width: 14rpx;
  height: 20rpx;
  border: 4rpx solid #FF8E3C;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  margin-top: 4rpx;
}
.icon-rmb {
  font-size: 24rpx;
  font-weight: 700;
  color: #F59E0B;
}

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
@keyframes ctaGlow {
  0%, 100% { box-shadow: 0 8rpx 24rpx rgba(30,94,255,0.35); }
  50% { box-shadow: 0 8rpx 32rpx rgba(30,94,255,0.55); }
}
@keyframes blink {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}
</style>
