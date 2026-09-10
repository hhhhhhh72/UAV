<template>
  <view class="pilot-detail-page" v-if="pilot">
    <!-- ═══ 一、Hero 区（浅色名片风）═══ -->
    <view class="hero">
      <!-- 右上角淡蓝同心圆装饰 -->
      <view class="hero-ring hero-ring-outer" />
      <view class="hero-ring hero-ring-inner" />

      <!-- 返回按钮（浅色背景用深灰） -->
      <view class="back-btn" :style="{ top: (statusBarH + 10) + 'px' }" hover-class="back-btn-hover" :hover-stay-time="120" @tap="goBack">
        <u-icon name="back" size="36rpx" color="#17212B" />
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
          <view v-if="certPill" class="status-pill" :class="'status-pill-' + certPill.tone">
            <text class="status-dot" />
            <text>{{ certPill.text }}</text>
          </view>
        </view>
        <text class="hero-id">{{ idText }}</text>
      </view>

      <!-- 数据横排（2 项：只放后端真实字段） -->
      <view class="hero-stats">
        <view class="hero-stat">
          <image class="hero-stat-img" src="/static/mine-icons/drone.svg" mode="aspectFit" />
          <text class="hero-stat-num">{{ displayNums.hours }}</text>
          <text class="hero-stat-label">飞行小时</text>
        </view>
        <view class="hero-stat">
          <image class="hero-stat-img" src="/static/mine-icons/certification-green.svg" mode="aspectFit" />
          <text class="hero-stat-num">{{ displayNums.certs }}</text>
          <text class="hero-stat-label">证书</text>
        </view>
      </view>
    </view>

    <view class="content" v-if="contentReady">
      <!-- ═══ 二、飞行数据卡 ═══ -->
      <view class="section-card card-float">
        <view class="section-title"><view class="title-bar" />飞行数据</view>
        <view class="data-grid">
          <view class="data-cell">
            <view class="data-icon data-icon-blue"><image class="data-img" src="/static/mine-icons/drone.svg" mode="aspectFit" /></view>
            <text class="data-num">{{ displayNums.hours }}</text>
            <text class="data-label">飞行小时</text>
          </view>
          <view class="data-divider" />
          <view class="data-cell">
            <view class="data-icon data-icon-green"><image class="data-img" src="/static/mine-icons/certification-green.svg" mode="aspectFit" /></view>
            <text class="data-num">{{ displayNums.certs }}</text>
            <text class="data-label">证书认证</text>
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
            <view class="cert-ico"><image class="cert-img" src="/static/mine-icons/certification-green.svg" mode="aspectFit" /></view>
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
  'linear-gradient(135deg,#17212B,#0A66C2)',
  'linear-gradient(135deg,#6D28D9,#DB2777)',
  'linear-gradient(135deg,#0EA5E9,#06B6D4)',
  'linear-gradient(135deg,#FF8E3C,#F97316)',
  'linear-gradient(135deg,#25915A,#34c759)',
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
// 认证状态标签：证据取自后端 certificates —— service 侧只返回"审核通过且在有效期内"的证书，
// 因此 certificates.length > 0 是可验证事实；无有效证书的历史记录不冒充"可接单"，退回中性表述。
const validCertCount = computed(() => ((pilot.value && pilot.value.certificates) || []).length)
const certPill = computed(() => {
  const p = pilot.value
  if (!p) return null
  if (validCertCount.value > 0) return { text: '证书有效', tone: 'green' }
  if (p.status === 'approved') return { text: '协会已核验', tone: 'blue' }
  return null
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
  { key: ['电力巡检', '巡检'], color: '#0A66C2', bg: 'rgba(10,102,194,.08)' },
  { key: ['测绘', '航拍', '拍摄'], color: '#7056D6', bg: 'rgba(139,92,246,.08)' },
  { key: ['植保', '喷洒'], color: '#25915A', bg: 'rgba(37,145,90,.08)' },
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
  return { color: '#667085', bg: 'rgba(102,112,133,.08)' }
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
    // 只陈述有数据来源的事实：飞行小时为飞手申请时填报值，证书为协会审核通过且在有效期内的记录
    const parts = []
    if (hours > 0) parts.push(`累计飞行 ${hours} 小时`)
    if (skills.length) parts.push(`擅长${skills.join('、')}等作业方向`)
    if (validCertCount.value > 0) parts.push('持有协会核验且在有效期内的证书')
    return parts.length ? parts.join('，') + '。' : '该飞手暂未填写简介'
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
const displayNums = ref({ hours: 0, certs: 0 })
const countUpNums = (target, duration = 800) => {
  const from = { hours: 0, certs: 0 }
  const start = Date.now()
  // 微信小程序无 requestAnimationFrame，用 setTimeout 模拟帧（16ms）
  const tick = () => {
    const p = Math.min(1, (Date.now() - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3)
    displayNums.value = {
      hours: Math.round(from.hours + (target.hours - from.hours) * ease),
      certs: Math.round(from.certs + (target.certs - from.certs) * ease),
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
  const wantID = options && options.id ? decodeURIComponent(options.id) : ''
  const cached = uni.getStorageSync('pilot_detail')
  // 入参 id 优先：缓存只是"同一个人"时的加速，不能盖掉 id。
  // （此前无论入参是什么都优先用缓存，从消息通知/分享进入详情会打开上一次看过的人）
  if (wantID && cached && cached.id === wantID) {
    pilot.value = cached
  } else if (wantID) {
    try {
      const res = await request({ url: '/api/v1/certified-pilots/' + encodeURIComponent(wantID) })
      pilot.value = res && res.data ? res.data : res
    } catch (e) { pilot.value = null }
    if (!pilot.value || !pilot.value.id) {
      uni.showToast({ title: '飞手信息不存在', icon: 'none' })
      setTimeout(() => uni.navigateBack(), 1200)
      return
    }
  } else if (cached && cached.id) {
    pilot.value = cached
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
      })
    }
  }, 150)
})
</script>

<style scoped>
.pilot-detail-page {
  min-height: 100vh;
  background: #F5F6F8;
  padding-bottom: calc(170rpx + env(safe-area-inset-bottom));
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
  background: rgba(10,102,194,0.05);
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
  background: linear-gradient(135deg, #0A66C2, #1DD4A8);
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
  background: #EAF3FB;
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
/* 头像认证徽章：36rpx 绿圆 + 白描边；对勾为 in-flow ::after，靠 flex 居中，
   不用绝对定位 + 手调 left/top（旋转后仍落在圆心） */
.cert-badge {
  position: absolute;
  right: 2rpx;
  bottom: 2rpx;
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  background: #25915A;
  border: 3rpx solid #ffffff;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-badge::after {
  content: '';
  width: 9rpx;
  height: 15rpx;
  border: solid #ffffff;
  border-width: 0 3rpx 3rpx 0;
  /* 对勾旋转后墨迹重心偏下（+3rpx），用 margin 上提归中：墨迹中心 = 圆心 */
  margin-bottom: 6rpx;
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
  color: #17212B;
}
.status-pill {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  font-weight: 600;
}
/* 有在有效期内的已核验证书 → 绿；无有效证书的历史名录记录 → 品牌蓝中性表述 */
.status-pill-green {
  background: #E9F7F0;
  color: #25915A;
}
.status-pill-blue {
  background: #EAF3FB;
  color: #0A66C2;
}
/* 静态圆点：该标签描述的是核验事实，不是实时在线状态，不做脉动 */
.status-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #25915A;
}
.status-pill-blue .status-dot {
  background: #0A66C2;
}
.hero-id {
  font-size: 24rpx;
  color: #667085;
}

/* 数据横排（2 项） */
.hero-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  margin-top: 24rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid rgba(10,102,194,0.08);
  z-index: 3;
}
.hero-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rpx;
}
.hero-stat-num {
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.2;
}
.hero-stat-label {
  font-size: 24rpx;
  color: #667085;
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
  color: #17212B;
  margin-bottom: 20rpx;
}
.title-bar {
  width: 6rpx;
  height: 28rpx;
  border-radius: 3rpx;
  background: linear-gradient(180deg, #0A66C2, #0A66C2);
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
.data-icon-blue { background: #EAF3FB; }
.data-icon-green { background: #E9F7F0; }
.data-num {
  font-size: 36rpx;
  font-weight: 700;
  color: #17212B;
  display: block;
  animation: numPop 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.data-label { font-size: 24rpx; color: #667085; margin-top: 4rpx; display: block; }
.data-divider {
  width: 1rpx;
  height: 56rpx;
  background: linear-gradient(180deg, rgba(10,102,194,0), rgba(10,102,194,0.12) 50%, rgba(10,102,194,0));
}

/* ═══ 三、擅长领域 ═══ */
.bio-tags { display: flex; flex-wrap: wrap; gap: 12rpx; }
.bio-tag {
  font-size: 24rpx;
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
  background: #E9F7F0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.cert-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.cert-name { font-size: 28rpx; font-weight: 600; color: #17212B; }
.cert-desc { font-size: 24rpx; color: #667085; }
.cert-badge-tag {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  background: rgba(37,145,90,0.1);
  border: 1rpx solid rgba(37,145,90,0.3);
  font-size: 24rpx;
  color: #00B87F;
  flex-shrink: 0;
}
.cert-badge-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #25915A;
}
.cert-empty { font-size: 24rpx; color: #667085; }

/* ═══ 五、个人信息 ═══ */
.profile-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #E8EEF7;
}
.profile-row:last-child { border-bottom: none; }
.profile-label { font-size: 26rpx; color: #667085; }
.profile-value { font-size: 28rpx; font-weight: 600; color: #17212B; }
.profile-id {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4rpx;
}
.profile-id-hint { font-size: 24rpx; color: #667085; }

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
  border: 2rpx solid #0A66C2;
  color: #0A66C2;
  font-size: 28rpx;
  font-weight: 600;
  padding: 0;
}
.cta-invite {
  flex: 2;
  height: 80rpx;
  line-height: 80rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #0A66C2, #0A66C2);
  border: none;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 700;
  padding: 0;
  box-shadow: 0 8rpx 24rpx rgba(10,102,194,0.35);
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



/* 图标体系：控件用 u-icon，领域标识用 static/mine-icons 下 SVG（禁 emoji/Unicode 字符） */
.hero-stat-img,
.data-img { width: 28rpx; height: 28rpx; }
.cert-img { width: 32rpx; height: 32rpx; }

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
  0%, 100% { box-shadow: 0 8rpx 24rpx rgba(10,102,194,0.35); }
  50% { box-shadow: 0 8rpx 32rpx rgba(10,102,194,0.55); }
}
@keyframes blink {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}
</style>
