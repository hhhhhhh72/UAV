<template>
  <view class="pilot-detail-page" v-if="pilot">
    <!-- ═══ 一、Hero 区（浅色名片风）═══ -->
    <view class="hero">
      <!-- 右上角淡蓝同心圆装饰 -->
      <view class="hero-ring hero-ring-outer" />
      <view class="hero-ring hero-ring-inner" />

      <!-- 返回按钮（浅色背景用深灰） -->
      <view class="back-btn" :style="{ top: (statusBarH + 10) + 'px' }" hover-class="back-btn-hover" :hover-stay-time="120" @tap="goBack">
        <text class="back-icon">‹</text>
      </view>

      <!-- 头像区（居中） -->
      <view class="hero-avatar-zone">
        <view class="avatar-halo">
          <view v-if="pilot.avatar" class="avatar-holder">
            <image :src="pilot.avatar" mode="aspectFill" class="avatar" />
          </view>
          <view v-else class="avatar avatar-fallback" :style="{ background: avatarBg(pilot.real_name) }">
            <text class="avatar-char">{{ firstChar(pilot.real_name) }}</text>
          </view>
          <view class="cert-badge" />
        </view>
      </view>

      <!-- 信息区（居中） -->
      <view class="hero-info">
        <view class="hero-name-row">
          <text class="hero-name">{{ pilot.real_name || '认证飞手' }}</text>
          <view class="status-pill"><text class="status-dot" />可接单</view>
        </view>
        <text class="hero-id">{{ idText }}</text>
      </view>

      <!-- 数据横排（4 项） -->
      <view class="hero-stats">
        <view class="hero-stat">
          <text class="hero-stat-icon hero-stat-icon-orange">★</text>
          <text class="hero-stat-num">{{ ratingText }}</text>
          <text class="hero-stat-label">评分</text>
        </view>
        <view class="hero-stat">
          <text class="hero-stat-icon hero-stat-icon-blue">飞</text>
          <text class="hero-stat-num">{{ displayNums.hours }}</text>
          <text class="hero-stat-label">飞行小时</text>
        </view>
        <view class="hero-stat">
          <text class="hero-stat-icon hero-stat-icon-purple">证</text>
          <text class="hero-stat-num">{{ displayNums.certs }}</text>
          <text class="hero-stat-label">证书</text>
        </view>
        <view class="hero-stat">
          <text class="hero-stat-icon hero-stat-icon-green">✓</text>
          <text class="hero-stat-num">{{ displayNums.jobs }}</text>
          <text class="hero-stat-label">作业</text>
        </view>
      </view>
    </view>

    <view class="content" v-if="contentReady">
      <!-- ═══ 二、飞行数据卡 ═══ -->
      <view class="section-card card-float">
        <view class="section-title"><view class="title-bar" />飞行数据</view>
        <view class="data-grid">
          <view class="data-cell">
            <view class="data-icon data-icon-blue"><view class="icon-plane" /></view>
            <text class="data-num">{{ displayNums.hours }}</text>
            <text class="data-label">飞行小时</text>
          </view>
          <view class="data-divider" />
          <view class="data-cell">
            <view class="data-icon data-icon-purple"><view class="icon-cert" /></view>
            <text class="data-num">{{ displayNums.certs }}</text>
            <text class="data-label">证书认证</text>
          </view>
          <view class="data-divider" />
          <view class="data-cell">
            <view class="data-icon data-icon-green"><view class="icon-check" /></view>
            <text class="data-num">{{ displayNums.jobs }}</text>
            <text class="data-label">完成作业</text>
          </view>
        </view>
      </view>

      <!-- ═══ 三、擅长领域卡 ═══ -->
      <view class="section-card" v-if="bioTags.length > 0">
        <view class="section-title"><view class="title-bar" />擅长领域</view>
        <view class="bio-tags">
          <text v-for="(b, i) in bioTags" :key="i" class="bio-tag" :style="tagStyle(b)">{{ b }}</text>
        </view>
      </view>

      <!-- ═══ 四、认证证书卡 ═══ -->
      <view class="section-card">
        <view class="section-title"><view class="title-bar" />认证证书</view>
        <view v-if="certCount > 0" class="cert-list">
          <view class="cert-item" v-for="(c, i) in certItems" :key="i">
            <view class="cert-ico"><view class="icon-cert" /></view>
            <view class="cert-info">
              <text class="cert-name">{{ c.name }}</text>
              <text class="cert-desc">{{ c.desc }}</text>
            </view>
            <view class="cert-badge-tag"><text class="cert-badge-dot" />已核验</view>
          </view>
        </view>
        <view v-else class="cert-empty">
          <text>暂无证书信息</text>
        </view>
      </view>

      <!-- ═══ 五、个人信息卡 ═══ -->
      <view class="section-card">
        <view class="section-title"><view class="title-bar" />个人信息</view>
        <view class="profile-row">
          <text class="profile-label">证件号</text>
          <view class="profile-id">
            <text class="profile-value">{{ maskIdCard(pilot.id_card) }}</text>
            <text class="profile-id-hint">已脱敏，仅协会审核可见</text>
          </view>
        </view>
        <view class="profile-row">
          <text class="profile-label">认证时间</text>
          <text class="profile-value">{{ fullDate(pilot.updated_at) }}</text>
        </view>
        <view class="profile-row">
          <text class="profile-label">所在地区</text>
          <text class="profile-value">{{ regionText }}</text>
        </view>
      </view>

      <!-- ═══ 六、飞手简介卡 ═══ -->
      <view class="section-card">
        <view class="section-title"><view class="title-bar" />飞手简介</view>
        <text class="section-text">{{ bioText }}</text>
      </view>
    </view>

    <view v-else class="skeleton-wrap">
      <view class="skeleton-block" />
      <view class="skeleton-block" />
    </view>

    <!-- ═══ 七、底部 CTA 栏 ═══ -->
    <view class="action-bar">
      <button class="cta-phone" @tap="contactPilot">联系飞手</button>
      <button class="cta-invite" @tap="invitePilot">邀请作业</button>
    </view>
  </view>
</template>

<script setup>
import { safeBack } from '../../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad, onReady } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'

const contentReady = ref(false)
const pilot = ref(null)
// 状态栏高度（Hero 返回钮避让）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const goBack = () => safeBack()

// ── 头像：姓名首字 + 姓名哈希渐变（每人不同）──
const AVATAR_GRADIENTS = [
  'linear-gradient(135deg,#0A1F44,#1E5EFF)',
  'linear-gradient(135deg,#6D28D9,#DB2777)',
  'linear-gradient(135deg,#0EA5E9,#06B6D4)',
  'linear-gradient(135deg,#FF8E3C,#F97316)',
  'linear-gradient(135deg,#00C896,#34c759)',
]
const firstChar = (name) => String(name || '飞').charAt(0)
const avatarBg = (name) => {
  const n = String(name || '')
  if (!n) return AVATAR_GRADIENTS[0]
  let h = 0
  for (let i = 0; i < n.length; i++) h = (h * 31 + n.charCodeAt(i)) >>> 0
  return AVATAR_GRADIENTS[h % AVATAR_GRADIENTS.length]
}

// ── 核心信息 ──────────────────────────
const certCount = computed(() => (pilot.value && (pilot.value.cert_ids || []).length) || 0)
const ratingText = computed(() => (pilot.value && pilot.value.rating > 0 ? pilot.value.rating.toFixed(1) : '—'))
// 评价数弱化：≥10 显示"XX 人评价"，否则"暂无评价"
const ratingSub = computed(() => {
  const n = (pilot.value && pilot.value.completed_jobs) || 0
  return n >= 10 ? `${n} 人评价` : '暂无评价'
})
// 编号：后端无编号，用"协会认证 · N 项证书"
const idText = computed(() => {
  const n = certCount.value
  return `协会认证 · ${n} 项证书`
})
// 地区：有 region 显示真实区县，未填写兜底"重庆·协会名录"
const regionText = computed(() => pilot.value.region || '重庆·协会名录')

// ── 擅长领域标签（7 类分色）─────────────
const JOB_TAG_MAP = [
  { key: ['电力巡检', '巡检'], color: '#1E5EFF', bg: 'rgba(30,94,255,.08)' },
  { key: ['测绘', '航拍', '拍摄'], color: '#8B5CF6', bg: 'rgba(139,92,246,.08)' },
  { key: ['植保', '喷洒'], color: '#00C896', bg: 'rgba(0,200,150,.08)' },
  { key: ['应急', '救援', '侦察'], color: '#EF4444', bg: 'rgba(239,68,68,.08)' },
  { key: ['吊运', '吊装', '实操'], color: '#FF8E3C', bg: 'rgba(255,142,60,.08)' },
  { key: ['物流', '运输', '投送'], color: '#06B6D4', bg: 'rgba(6,182,212,.08)' },
  { key: ['航拍', '宣传'], color: '#EC4899', bg: 'rgba(236,72,153,.08)' },
]
const bioList = (bio) => String(bio || '').split(/[/，,、\s]+/).filter(Boolean)
const bioTags = computed(() => bioList(pilot.value && pilot.value.bio).slice(0, 5))
const matchTag = (tag) => {
  for (const m of JOB_TAG_MAP) {
    if (m.key.some((k) => tag.includes(k))) return m
  }
  return { color: '#6B7B95', bg: 'rgba(107,123,149,.08)' }
}
const tagStyle = (t) => {
  const m = matchTag(t)
  return { color: m.color, background: m.bg, borderColor: m.color }
}

// ── 证书展示：优先用后端 certificates 明细（cert_name/issuer_org），无则退回 cert_ids 占位 ──
const CERT_TYPE_NAME = { caac: 'CAAC无人机驾驶员执照', utc_dji: 'DJI UTC 植保无人机驾驶证', gov_level: '政府职业技能等级证书' }
const certItems = computed(() => {
  const briefs = (pilot.value && pilot.value.certificates) || []
  if (briefs.length) {
    return briefs.map((c) => {
      const issuer = c.issuer_org || ''
      const lv = c.level || ''
      return {
        name: c.cert_name || CERT_TYPE_NAME[c.cert_type] || '无人机驾驶证照',
        desc: [issuer, lv].filter(Boolean).join(' · ') || '协会已核验',
      }
    })
  }
  const ids = (pilot.value && pilot.value.cert_ids) || []
  return ids.map((id, i) => ({
    name: `无人机驾驶证照 #${i + 1}`,
    desc: '协会已核验 · 证书编号 ' + String(id).slice(0, 8),
  }))
})

// ── 飞手简介：不重复标签，生成段落 ──────
const bioText = computed(() => {
  const p = pilot.value
  if (!p) return '该飞手暂未填写简介'
  if (p.bio) {
    const skills = bioList(p.bio)
    const hours = p.flight_hours || 0
    const jobs = p.completed_jobs || 0
    // 生成一段真实可读的简介（避免与标签重复感）
    return `从事无人机飞行多年，累计飞行 ${hours} 小时，擅长${skills.join('、')}等作业方向。持有协会认证资质，累计完成 ${jobs} 项作业，作业质量稳定可靠。`
  }
  return '该飞手暂未填写简介'
})

// ── CTA（后端暂无联系/邀请接口，仅提示即将上线）──────────
const contactPilot = () => {
  uni.showToast({ title: '功能即将上线', icon: 'none' })
}
const invitePilot = () => {
  uni.showToast({ title: '功能即将上线', icon: 'none' })
}

// ── 数据滚动计数 ───────────────────────
const displayNums = ref({ hours: 0, certs: 0, jobs: 0 })
const countUpNums = (target, duration = 800) => {
  const from = { hours: 0, certs: 0, jobs: 0 }
  const start = Date.now()
  // 微信小程序无 requestAnimationFrame，用 setTimeout 模拟帧（16ms）
  const tick = () => {
    const p = Math.min(1, (Date.now() - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3)
    displayNums.value = {
      hours: Math.round(from.hours + (target.hours - from.hours) * ease),
      certs: Math.round(from.certs + (target.certs - from.certs) * ease),
      jobs: Math.round(from.jobs + (target.jobs - from.jobs) * ease),
    }
    if (p < 1) setTimeout(tick, 16)
  }
  tick()
}

// ── 证件号脱敏：保留前 3 位 + 11 个 * + 后 4 位（与后端 MaskIDCard 口径一致）──
const maskIdCard = (v) => {
  const s = String(v || '')
  if (s.length < 8) return '已脱敏'
  return `${s.slice(0, 3)}***********${s.slice(-4)}`
}

// ── 时间格式化 ────────────────────────
function fullDate(v) {
  if (!v) return '-'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '-'
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

onLoad(async (options) => {
  // 优先读列表页缓存（含完整字段），无缓存且有 id 时回源单查接口
  const cached = uni.getStorageSync('pilot_detail')
  if (cached && cached.id) {
    pilot.value = cached
  } else if (options && options.id) {
    try {
      const res = await request({ url: '/api/v1/certified-pilots/' + encodeURIComponent(options.id) })
      pilot.value = res && res.data ? res.data : res
    } catch (e) { pilot.value = null }
    if (!pilot.value || !pilot.value.id) {
      uni.showToast({ title: '飞手信息不存在', icon: 'none' })
      setTimeout(() => uni.navigateBack(), 1200)
      return
    }
  } else {
    uni.showToast({ title: '飞手信息不存在', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1200)
    return
  }
  uni.setNavigationBarTitle({ title: (pilot.value && pilot.value.real_name) || '飞手档案' })
})

onReady(() => {
  setTimeout(() => {
    contentReady.value = true
    // 数据滚动计数动画
    if (pilot.value) {
      countUpNums({
        hours: pilot.value.flight_hours || 0,
        certs: (pilot.value.cert_ids || []).length,
        jobs: pilot.value.completed_jobs || 0,
      })
    }
  }, 150)
})
</script>

<style scoped>
.pilot-detail-page {
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: 130rpx;
}

/* ═══ 一、Hero（浅色名片风）═══ */
.hero {
  position: relative;
  min-height: 320rpx;
  padding: 0 32rpx 32rpx;
  background: linear-gradient(180deg, #F8FAFE 0%, #EEF3FA 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
}
/* 右上角淡蓝同心圆 */
.hero-ring {
  position: absolute;
  border-radius: 50%;
  background: rgba(30,94,255,0.05);
  pointer-events: none;
}
.hero-ring-outer {
  top: -100rpx;
  right: -100rpx;
  width: 300rpx;
  height: 300rpx;
}
.hero-ring-inner {
  top: -30rpx;
  right: -30rpx;
  width: 200rpx;
  height: 200rpx;
}

/* 返回按钮（浅色背景用深灰） */
.back-btn {
  position: absolute;
  left: 24rpx;
  z-index: 5;
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: rgba(255,255,255,0.9);
  border: 1rpx solid #E8EEF7;
  box-shadow: 0 2rpx 8rpx rgba(10,31,68,0.06);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}
.back-btn-hover {
  background: #E8EEF7;
}
.back-icon {
  font-size: 44rpx;
  color: #6B7B95;
  font-weight: 300;
  line-height: 1;
}

/* 头像区（居中） */
.hero-avatar-zone {
  margin-top: 96rpx;
  z-index: 3;
}
/* 光环：渐变描边 */
.avatar-halo {
  position: relative;
  width: 128rpx;
  height: 128rpx;
  border-radius: 50%;
  padding: 4rpx;
  background: linear-gradient(135deg, #1E5EFF, #00E5FF);
}
.avatar-holder {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  overflow: hidden;
}
.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: #DCE9FF;
  animation: avatarIn 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}
.avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-char {
  font-size: 48rpx;
  font-weight: 700;
  color: #ffffff;
}
.cert-badge {
  position: absolute;
  right: 2rpx;
  bottom: 2rpx;
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  background: #00C896;
  border: 3rpx solid #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-badge::after {
  content: '';
  position: absolute;
  left: 9rpx;
  top: 8rpx;
  width: 12rpx;
  height: 16rpx;
  border: solid #fff;
  border-width: 0 3rpx 3rpx 0;
  transform: rotate(45deg);
}

/* 信息区（居中） */
.hero-info {
  margin-top: 16rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  z-index: 3;
}
.hero-name-row {
  display: flex;
  align-items: center;
  gap: 14rpx;
}
.hero-name {
  font-size: 44rpx;
  font-weight: 700;
  color: #0A1F44;
}
.status-pill {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  background: #D1FAE5;
  font-size: 20rpx;
  color: #00C896;
  font-weight: 600;
}
.status-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #00C896;
  animation: pulse 2s infinite;
}
.hero-id {
  font-size: 24rpx;
  color: #6B7B95;
}

/* 数据横排（4 项） */
.hero-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  margin-top: 24rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid rgba(30,94,255,0.08);
  z-index: 3;
}
.hero-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rpx;
}
.hero-stat-icon {
  font-size: 26rpx;
  font-weight: 700;
  line-height: 1;
}
.hero-stat-icon-orange { color: #F59E0B; }
.hero-stat-icon-blue { color: #1E5EFF; }
.hero-stat-icon-purple { color: #8B5CF6; }
.hero-stat-icon-green { color: #00C896; }
.hero-stat-num {
  font-size: 32rpx;
  font-weight: 700;
  color: #0A1F44;
  line-height: 1.2;
}
.hero-stat-label {
  font-size: 20rpx;
  color: #6B7B95;
}

/* ═══ 内容区 ═══ */
.content {
  position: relative;
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
/* 首张"飞行数据"卡向上重叠 16rpx，与 Hero 自然衔接 */
.card-float {
  margin-top: -16rpx;
  border-radius: 16rpx;
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

/* ═══ 二、飞行数据 ═══ */
.data-grid {
  display: flex;
  align-items: center;
  text-align: center;
  padding: 8rpx 0;
}
.data-cell { flex: 1; }
.data-icon {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 10rpx;
}
.data-icon-blue { background: #DCE9FF; }
.data-icon-purple { background: #F3E8FF; }
.data-icon-green { background: #D1FAE5; }
.data-num {
  font-size: 36rpx;
  font-weight: 700;
  color: #0A1F44;
  display: block;
  animation: numPop 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.data-label { font-size: 22rpx; color: #6B7B95; margin-top: 4rpx; display: block; }
.data-divider {
  width: 1rpx;
  height: 56rpx;
  background: linear-gradient(180deg, rgba(30,94,255,0), rgba(30,94,255,0.12) 50%, rgba(30,94,255,0));
}

/* ═══ 三、擅长领域 ═══ */
.bio-tags { display: flex; flex-wrap: wrap; gap: 12rpx; }
.bio-tag {
  font-size: 22rpx;
  padding: 8rpx 24rpx;
  border-radius: 8rpx;
  border: 1rpx solid;
  font-weight: 600;
}

/* ═══ 四、认证证书 ═══ */
.cert-list { display: flex; flex-direction: column; gap: 20rpx; }
.cert-item { display: flex; align-items: center; gap: 16rpx; }
.cert-ico {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  background: #DCE9FF;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.cert-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.cert-name { font-size: 28rpx; font-weight: 600; color: #0A1F44; }
.cert-desc { font-size: 22rpx; color: #6B7B95; }
.cert-badge-tag {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  background: rgba(0,200,150,0.1);
  border: 1rpx solid rgba(0,200,150,0.3);
  font-size: 20rpx;
  color: #00B87F;
  flex-shrink: 0;
}
.cert-badge-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #00C896;
}
.cert-empty { font-size: 24rpx; color: #ADB8C7; }

/* ═══ 五、个人信息 ═══ */
.profile-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #E8EEF7;
}
.profile-row:last-child { border-bottom: none; }
.profile-label { font-size: 26rpx; color: #6B7B95; }
.profile-value { font-size: 28rpx; font-weight: 600; color: #0A1F44; }
.profile-id {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4rpx;
}
.profile-id-hint { font-size: 20rpx; color: #ADB8C7; }

/* ═══ 六、飞手简介 ═══ */
.section-text {
  font-size: 26rpx;
  color: #2C3E50;
  line-height: 1.7;
  display: block;
}

/* ═══ 七、底部 CTA ═══ */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  gap: 16rpx;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  border-top: 1rpx solid #E8EEF7;
  box-shadow: 0 -2rpx 8rpx rgba(0,0,0,0.04);
}
.cta-phone {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  border-radius: 999rpx;
  background: #ffffff;
  border: 2rpx solid #1E5EFF;
  color: #1E5EFF;
  font-size: 28rpx;
  font-weight: 600;
  padding: 0;
}
.cta-invite {
  flex: 2;
  height: 80rpx;
  line-height: 80rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #1E5EFF, #0A66C2);
  border: none;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 700;
  padding: 0;
  box-shadow: 0 8rpx 24rpx rgba(30,94,255,0.35);
  animation: ctaGlow 2.5s ease-in-out infinite;
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
.icon-plane {
  width: 0;
  height: 0;
  border-top: 6rpx solid transparent;
  border-bottom: 6rpx solid transparent;
  border-right: 20rpx solid #8B5CF6;
  position: relative;
  transform: rotate(-8deg);
}
.icon-plane::after {
  content: '';
  position: absolute;
  right: -18rpx;
  top: -5rpx;
  width: 0;
  height: 0;
  border-left: 4rpx solid transparent;
  border-right: 4rpx solid transparent;
  border-bottom: 10rpx solid #8B5CF6;
}
.icon-cert {
  width: 18rpx;
  height: 22rpx;
  border-radius: 4rpx;
  background: #1E5EFF;
  position: relative;
}
.icon-cert::after {
  content: '';
  position: absolute;
  left: 4rpx;
  right: 4rpx;
  top: 9rpx;
  height: 2rpx;
  background: #DCE9FF;
  box-shadow: 0 4rpx 0 #DCE9FF;
}
.icon-check {
  width: 14rpx;
  height: 24rpx;
  border: solid #00C896;
  border-width: 0 4rpx 4rpx 0;
  transform: rotate(45deg);
  margin-top: -6rpx;
}

/* ═══ 微动效 ═══ */
@keyframes avatarIn {
  from {
    transform: scale(0.9);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
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
@keyframes ctaGlow {
  0%, 100% { box-shadow: 0 8rpx 24rpx rgba(30,94,255,0.35); }
  50% { box-shadow: 0 8rpx 32rpx rgba(30,94,255,0.55); }
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
@keyframes blink {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}
</style>
