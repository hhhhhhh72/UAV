<template>
  <view class="study-detail-page" v-if="pkg">
    <view class="detail-content">
      <view class="pkg-header" :style="{ background: pkg.headerBg }">
        <view class="header-mask" />
        <view class="header-inner">
          <view class="pkg-tag-top">{{ pkg.tag }}</view>
          <view class="pkg-title">{{ pkg.name }}</view>
          <view class="pkg-price-row">
            <text class="currency">¥</text>
            <text class="price-num">{{ pkg.price }}</text>
            <text class="price-unit">/人</text>
          </view>
        </view>
      </view>

      <view v-if="contentReady">
        <view class="section-card">
          <view class="section-title">研学内容</view>
          <text class="section-text">{{ pkg.intro }}</text>
        </view>

        <!-- 研学目标 -->
        <view class="section-card" v-if="pkg.studyGoals && pkg.studyGoals.length > 0">
          <view class="section-title">研学目标</view>
          <view class="goals-list">
            <view v-for="(goal, i) in pkg.studyGoals" :key="i" class="goal-item">
              <view class="goal-label">{{ goal.label }}</view>
              <view class="goal-content">{{ goal.content }}</view>
            </view>
          </view>
        </view>

        <!-- 往期活动展示（支持后台配置） -->
        <view class="section-card" v-if="studyShowcase.length > 0">
          <view class="section-title">精彩回顾</view>
          <view class="showcase-grid">
            <view
              v-for="(item, idx) in studyShowcase"
              :key="idx"
              class="showcase-item"
              @tap="previewShowcase(idx)"
            >
              <image :src="item.image" mode="aspectFill" class="showcase-img" />
              <view class="showcase-info">
                <view class="showcase-title">{{ item.title }}</view>
                <view class="showcase-desc">{{ item.desc }}</view>
              </view>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">课程安排</view>
          <view class="session-toggle">
            <view
              class="toggle-btn"
              :class="{ active: activeSession === 'am' }"
              @tap="activeSession = 'am'"
            >
              <text>上午场</text>
            </view>
            <view
              class="toggle-btn"
              :class="{ active: activeSession === 'pm' }"
              @tap="activeSession = 'pm'"
            >
              <text>下午场</text>
            </view>
          </view>
          <view class="schedule-list">
            <view v-for="(item, index) in pkg.schedule" :key="index" class="schedule-item">
              <view class="schedule-time">{{ activeSession === 'am' ? item.amTime : item.pmTime }}</view>
              <view class="schedule-line">
                <view class="schedule-dot" />
                <view class="schedule-bar" v-if="index < pkg.schedule.length - 1" />
              </view>
              <view class="schedule-content">
                <view class="schedule-name">{{ item.name }}</view>
                <view class="schedule-location" v-if="item.location">
                  <text class="location-icon">📍</text>
                  <text>{{ item.location }}</text>
                </view>
                <view class="schedule-desc">{{ item.desc }}</view>
                <view class="schedule-purpose" v-if="item.purpose">
                  <text class="purpose-label">课程目的</text>
                  <text>{{ item.purpose }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">课程亮点</view>
          <view class="highlight-grid">
            <view v-for="(h, i) in pkg.highlights" :key="i" class="highlight-card">
              <text class="highlight-emoji">{{ h.emoji }}</text>
              <view class="highlight-name">{{ h.name }}</view>
              <view class="highlight-desc">{{ h.desc }}</view>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">适合人群</view>
          <view class="audience-list">
            <view v-for="(a, i) in pkg.audience" :key="i" class="audience-item">
              <text class="audience-icon">👤</text>
              <text>{{ a }}</text>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">费用说明</view>
          <view class="fee-info">
            <view v-for="(f, i) in pkg.feeInfo" :key="i" class="fee-item">
              <text class="fee-label">{{ f.label }}</text>
              <text class="fee-value">{{ f.value }}</text>
            </view>
          </view>
        </view>

        <!-- 安全宣讲 -->
        <view class="section-card" v-if="pkg.safetyBriefing && pkg.safetyBriefing.length > 0">
          <view class="section-title">安全宣讲</view>
          <view class="safety-list">
            <view v-for="(item, i) in pkg.safetyBriefing" :key="i" class="safety-item">
              <text class="safety-icon">🛡️</text>
              <text>{{ item }}</text>
            </view>
          </view>
        </view>

        <!-- 研学总结 -->
        <view class="section-card" v-if="pkg.studySummary">
          <view class="section-title">研学总结</view>
          <text class="section-text">{{ pkg.studySummary }}</text>
        </view>

        <view class="section-card">
          <view class="section-title">温馨提示</view>
          <view class="tips-list">
            <view v-for="(tip, i) in pkg.tips" :key="i" class="tip-item">
              <text class="tip-index">{{ i + 1 }}</text>
              <text class="tip-text">{{ tip }}</text>
            </view>
          </view>
        </view>

        <view class="section-card contact-section">
          <view class="section-title">联系客服</view>
          <view class="contact-info">
            <view class="contact-row">如有疑问，请咨询客服热线：</view>
            <view class="phone-link" @tap="makeCall('0577-55550500')">0577-55550500</view>
            <view class="work-time">工作时间：工作日 8:30-17:30</view>
          </view>
        </view>
      </view>

      <view v-else class="skeleton-wrap">
        <view class="skeleton-block" />
        <view class="skeleton-block" />
      </view>
    </view>

    <view class="action-bar">
      <button class="apply-btn" @tap="onApply">立即报名</button>
    </view>

    <HomeFloatButton />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReady } from '@dcloudio/uni-app'
import HomeFloatButton from '@/components/HomeFloatButton.vue'
import { request } from '../../utils/request'

const contentReady = ref(false)
const pkg = ref(null)
const activeSession = ref('am')

const studyShowcase = ref([])

onLoad(async (options) => {
  const id = options.package || 'study-halfday'

  try {
    const res = await request({ url: '/api/services/config' })
    const allConfigs = res?.data || res || {}
    const config = allConfigs['9'] || {}

    // 从API获取课程包数据
    if (config.packages && config.packages[id]) {
      const remotePkg = config.packages[id]
      pkg.value = {
        id,
        name: remotePkg.name || '',
        tag: remotePkg.tag || '',
        price: remotePkg.price || 0,
        headerBg: remotePkg.headerBg || 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)',
        intro: remotePkg.intro || '',
        studyGoals: remotePkg.studyGoals || [],
        schedule: remotePkg.schedule || [],
        highlights: remotePkg.highlights || [],
        audience: remotePkg.audience || [],
        feeInfo: remotePkg.feeInfo || [],
        tips: remotePkg.tips || [],
        safetyBriefing: remotePkg.safetyBriefing || [],
        studySummary: remotePkg.studySummary || '',
      }
      studyShowcase.value = remotePkg.showcase || []
    }

    // 全局精彩回顾作为备选
    if (studyShowcase.value.length === 0 && config.studyShowcase) {
      studyShowcase.value = config.studyShowcase
    }
  } catch (e) {
    console.warn('加载配置失败:', e)
  }

  if (pkg.value) {
    uni.setNavigationBarTitle({ title: pkg.value.name })
  }
})

const previewShowcase = (index) => {
  const urls = studyShowcase.value.map(item => item.image).filter(Boolean)
  if (urls.length > 0) {
    uni.previewImage({ urls, current: urls[index] || urls[0] })
  }
}

onReady(() => {
  setTimeout(() => { contentReady.value = true }, 150)
})

const onApply = () => {
  uni.navigateTo({ url: `/pages/services/apply?id=9&package=${pkg.value.id}` })
}

const makeCall = (phone) => {
  uni.makePhoneCall({ phoneNumber: phone })
}
</script>

<style scoped>
.study-detail-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 100px;
}

.pkg-header {
  position: relative;
  padding: 40px 20px 50px;
  overflow: hidden;
}

.header-mask {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 80% 20%, rgba(255,255,255,0.18) 0%, transparent 55%);
}

.header-inner {
  position: relative;
  z-index: 1;
}

.pkg-tag-top {
  display: inline-block;
  font-size: 11px;
  color: rgba(255,255,255,0.9);
  background: rgba(255,255,255,0.2);
  padding: 3px 12px;
  border-radius: 20px;
  margin-bottom: 12px;
}

.pkg-title {
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  margin-bottom: 16px;
}

.pkg-price-row {
  display: flex;
  align-items: baseline;
}

.pkg-price-row .currency {
  font-size: 18px;
  font-weight: 700;
  color: #fff;
}

.pkg-price-row .price-num {
  font-size: 40px;
  font-weight: 800;
  color: #fff;
  line-height: 1;
  margin: 0 2px;
}

.pkg-price-row .price-unit {
  font-size: 14px;
  color: rgba(255,255,255,0.8);
}

.section-card {
  background: #fff;
  margin: 12px 16px;
  padding: 16px;
  border-radius: 12px;
}

.section-card:first-child {
  margin-top: -24px;
  position: relative;
  z-index: 2;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #323233;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid #2563eb;
}

.section-text {
  font-size: 14px;
  color: #646566;
  line-height: 1.8;
  display: block;
}

/* 课程安排时间线 */
/* 精彩回顾 / 往期活动展示 */
.showcase-grid {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.showcase-item {
  background: #f7f8fa;
  border-radius: 16rpx;
  overflow: hidden;
}

.showcase-img {
  width: 100%;
  height: 280rpx;
}

.showcase-info {
  padding: 16rpx 20rpx 20rpx;
}

.showcase-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.showcase-desc {
  font-size: 24rpx;
  color: #646566;
  line-height: 1.6;
}

/* 上午/下午切换 */
.session-toggle {
  display: flex;
  background: #f5f5f7;
  border-radius: 10px;
  padding: 3px;
  margin-bottom: 16px;
}

.toggle-btn {
  flex: 1;
  text-align: center;
  padding: 8px 0;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #646566;
  transition: all 0.25s;
}

.toggle-btn.active {
  background: #fff;
  color: #2563eb;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.schedule-list {
  padding-left: 4px;
}

.schedule-item {
  display: flex;
  gap: 12px;
  min-height: 60px;
}

.schedule-time {
  width: 42px;
  font-size: 13px;
  font-weight: 600;
  color: #323233;
  padding-top: 2px;
  flex-shrink: 0;
}

.schedule-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 16px;
  flex-shrink: 0;
}

.schedule-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #2563eb;
  border: 2px solid #dbeafe;
  flex-shrink: 0;
}

.schedule-bar {
  width: 2px;
  flex: 1;
  background: #e5e7eb;
  margin-top: 4px;
}

.schedule-content {
  flex: 1;
  padding-bottom: 16px;
}

.schedule-name {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  margin-bottom: 4px;
}

.schedule-desc {
  font-size: 12px;
  color: #969799;
  line-height: 1.5;
}

.schedule-location {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #2563eb;
  margin-bottom: 4px;
}

.location-icon {
  font-size: 12px;
}

.schedule-purpose {
  font-size: 12px;
  color: #636366;
  line-height: 1.5;
  margin-top: 6px;
  padding: 8px 10px;
  background: #f7f8fa;
  border-radius: 8px;
}

.purpose-label {
  display: inline-block;
  font-size: 11px;
  font-weight: 600;
  color: #2563eb;
  background: rgba(37, 99, 235, 0.08);
  padding: 1px 6px;
  border-radius: 4px;
  margin-right: 6px;
}

/* 研学目标 */
.goals-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.goal-item {
  padding: 14px;
  background: #f7f8fa;
  border-radius: 10px;
}

.goal-label {
  font-size: 13px;
  font-weight: 700;
  color: #2563eb;
  margin-bottom: 6px;
}

.goal-content {
  font-size: 14px;
  color: #646566;
  line-height: 1.7;
}

/* 安全宣讲 */
.safety-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.safety-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 14px;
  color: #646566;
  line-height: 1.6;
}

.safety-icon {
  font-size: 14px;
  flex-shrink: 0;
  margin-top: 2px;
}

/* 课程亮点 */
.highlight-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.highlight-card {
  background: #f7f8fa;
  border-radius: 10px;
  padding: 14px 12px;
  text-align: center;
}

.highlight-emoji {
  font-size: 28px;
  display: block;
  margin-bottom: 6px;
}

.highlight-name {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  margin-bottom: 4px;
}

.highlight-desc {
  font-size: 11px;
  color: #969799;
}

/* 适合人群 */
.audience-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.audience-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #646566;
}

/* 费用说明 */
.fee-info {
  display: flex;
  flex-direction: column;
}

.fee-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px dashed #ebedf0;
}

.fee-item:last-child {
  border-bottom: none;
}

.fee-label {
  font-size: 14px;
  color: #646566;
}

.fee-value {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
}

/* 温馨提示 */
.tips-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tip-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.tip-index {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #f0f2ff;
  color: #2563eb;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.tip-text {
  font-size: 13px;
  color: #646566;
  line-height: 1.6;
  flex: 1;
}

/* 联系客服 */
.contact-section {
  text-align: center;
}

.contact-row {
  font-size: 14px;
  color: #646566;
  margin-bottom: 8px;
}

.phone-link {
  font-size: 24px;
  font-weight: 700;
  color: #2563eb;
  margin: 12px 0;
}

.work-time {
  font-size: 12px;
  color: #969799;
  margin-top: 8px;
}

/* 底部操作栏 */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: #fff;
  border-top: 1px solid #eee;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}

.apply-btn {
  width: 100%;
  height: 48px;
  line-height: 48px;
  border-radius: 999px;
  font-weight: 700;
  font-size: 16px;
  color: #fff;
  background: linear-gradient(135deg, #06b6d4 0%, #2563eb 100%) !important;
  border: none;
}

/* 骨架屏 */
.skeleton-wrap {
  padding: 20px;
}

.skeleton-block {
  height: 120px;
  background: #eee;
  border-radius: 12px;
  margin-bottom: 16px;
  animation: blink 1.5s infinite;
}

@keyframes blink {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}
</style>
