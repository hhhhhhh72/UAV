<template>
  <div class="study-detail-page" v-if="pkg">
    <van-nav-bar
      :title="pkg.name"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
      :border="false"
    />
    <HomeFloatButton />

    <div class="pkg-header" :style="headerStyle">
      <div class="header-mask" />
      <div class="header-inner">
        <span class="pkg-tag-top">{{ pkg.tag }}</span>
        <h1 class="pkg-title">{{ pkg.name }}</h1>
        <div class="pkg-price-row">
          <span class="currency">¥</span>
          <span class="price-num">{{ pkg.price }}</span>
          <span class="price-unit">/人</span>
        </div>
      </div>
    </div>

    <div class="detail-body" v-if="contentReady">
      <div class="section-card">
        <h2 class="section-title">研学内容</h2>
        <p class="section-text">{{ pkg.intro }}</p>
      </div>

      <!-- 研学目标 -->
      <div class="section-card" v-if="pkg.studyGoals && pkg.studyGoals.length > 0">
        <h2 class="section-title">研学目标</h2>
        <div class="goals-list">
          <div v-for="(goal, i) in pkg.studyGoals" :key="i" class="goal-item">
            <div class="goal-label">{{ goal.label }}</div>
            <div class="goal-content">{{ goal.content }}</div>
          </div>
        </div>
      </div>

      <!-- 服务项目（课程包独立） -->
      <div class="section-card" v-if="pkg.projects && pkg.projects.length > 0">
        <h2 class="section-title">服务项目</h2>
        <div class="project-grid">
          <div v-for="(p, i) in pkg.projects" :key="i" class="project-item">
            <template v-if="p.icon && p.icon.startsWith('http')">
              <img :src="p.icon" style="width:24px;height:24px;object-fit:contain;" />
            </template>
            <template v-else>
              <van-icon :name="p.icon || 'star-o'" size="24" color="#424245" />
            </template>
            <span>{{ p.name }}</span>
          </div>
        </div>
      </div>

      <!-- 往期活动展示（课程包独立） -->
      <div class="section-card" v-if="pkg.showcase && pkg.showcase.length > 0">
        <h2 class="section-title">精彩回顾</h2>
        <div class="showcase-grid">
          <div
            v-for="(item, idx) in pkg.showcase"
            :key="idx"
            class="showcase-item"
            @click="previewShowcase(item)"
          >
            <van-image
              :src="item.image ? getCompressedImage(item.image, 600) : ''"
              fit="cover"
              width="100%"
              height="140"
              radius="12"
              lazy-load
            >
              <template #loading>
                <div class="image-loading">
                  <van-loading type="spinner" size="20" />
                </div>
              </template>
              <template #error>
                <div class="image-error">
                  <van-icon name="photo-fail" size="24" color="#c8c8c8" />
                  <span>加载失败</span>
                </div>
              </template>
            </van-image>
            <div class="showcase-info">
              <div class="showcase-title">{{ item.title }}</div>
              <div class="showcase-desc">{{ item.desc }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="section-card">
        <h2 class="section-title">课程安排</h2>
        <div class="session-toggle">
          <div
            class="toggle-btn"
            :class="{ active: activeSession === 'am' }"
            @click="activeSession = 'am'"
          >上午场</div>
          <div
            class="toggle-btn"
            :class="{ active: activeSession === 'pm' }"
            @click="activeSession = 'pm'"
          >下午场</div>
        </div>
        <div class="schedule-list">
          <div v-for="(item, index) in pkg.schedule" :key="index" class="schedule-item">
            <div class="schedule-time">{{ activeSession === 'am' ? item.amTime : item.pmTime }}</div>
            <div class="schedule-marker">
              <div class="schedule-dot" />
            </div>
            <div class="schedule-content">
              <div class="schedule-name">{{ item.name }}</div>
              <div class="schedule-location" v-if="item.location">
                <van-icon name="location-o" size="12" color="#0071e3" />
                <span>{{ item.location }}</span>
              </div>
              <div class="schedule-desc">{{ item.desc }}</div>
              <div class="schedule-purpose" v-if="item.purpose">
                <span class="purpose-label">课程目的</span>
                <span>{{ item.purpose }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="section-card">
        <h2 class="section-title">适合人群</h2>
        <div class="audience-list">
          <div v-for="(a, i) in pkg.audience" :key="i" class="audience-item">
            <van-icon name="friends-o" size="16" color="#0071e3" />
            <span>{{ a }}</span>
          </div>
        </div>
      </div>

      <div class="section-card">
        <h2 class="section-title">费用说明</h2>
        <div class="fee-info">
          <div v-for="(f, i) in pkg.feeInfo" :key="i" class="fee-item">
            <span class="fee-label">{{ f.label }}</span>
            <span class="fee-value">{{ f.value }}</span>
          </div>
        </div>
      </div>

      <!-- 服务优势（课程包独立） -->
      <div class="section-card" v-if="pkg.advantages && pkg.advantages.length > 0">
        <h2 class="section-title">服务优势</h2>
        <div class="advantage-list">
          <div v-for="(adv, i) in pkg.advantages" :key="i" class="advantage-item">
            <van-icon name="success" size="16" color="#0071e3" />
            <span>{{ adv }}</span>
          </div>
        </div>
      </div>

      <!-- 安全宣讲 -->
      <div class="section-card" v-if="pkg.safetyBriefing && pkg.safetyBriefing.length > 0">
        <h2 class="section-title">安全宣讲</h2>
        <div class="safety-list">
          <div v-for="(item, i) in pkg.safetyBriefing" :key="i" class="safety-item">
            <van-icon name="shield-o" size="16" color="#0071e3" />
            <span>{{ item }}</span>
          </div>
        </div>
      </div>

      <!-- 研学总结 -->
      <div class="section-card" v-if="pkg.studySummary">
        <h2 class="section-title">研学总结</h2>
        <p class="section-text">{{ pkg.studySummary }}</p>
      </div>

      <div class="section-card">
        <h2 class="section-title">温馨提示</h2>
        <div class="tips-list">
          <div v-for="(tip, i) in pkg.tips" :key="i" class="tip-item">
            <span class="tip-index">{{ i + 1 }}</span>
            <span class="tip-text">{{ tip }}</span>
          </div>
        </div>
      </div>

      <div class="section-card contact-section">
        <h2 class="section-title">联系客服</h2>
        <div class="contact-info">
          <p class="contact-row">如有疑问，请咨询客服热线：</p>
          <a class="phone-link" href="tel:0577-55550500">0577-55550500</a>
          <p class="work-time">工作时间：工作日 9:00-17:30</p>
        </div>
      </div>
    </div>

    <div v-else class="loading-wrap">
      <van-loading vertical>加载中...</van-loading>
    </div>

    <div class="action-bar">
      <van-button type="primary" block round color="#0071e3" @click="onApply">
        立即报名
      </van-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showImagePreview } from 'vant'
import HomeFloatButton from '@/components/HomeFloatButton.vue'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const contentReady = ref(false)
const pkg = ref(null)
const activeSession = ref('am')

const normalizeUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http') || url.startsWith('data:') || url.startsWith('/')) return url
  if (url.includes('.') || url.includes('/')) return `/${url}`
  return url
}

// 获取压缩后的图片URL
const getCompressedImage = (url, width = 800) => {
  if (!url) return ''
  const normalizedUrl = normalizeUrl(url)
  // 对于空图片返回空字符串
  if (!normalizedUrl) return ''
  // 使用后端压缩接口
  return `/api/image?url=${encodeURIComponent(normalizedUrl)}&width=${width}&quality=75`
}

const headerStyle = computed(() => {
  const bg = pkg.value?.headerBg
  if (!bg) {
    return { background: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' }
  }
  // 如果是图片URL（http或相对路径），使用压缩接口
  if (bg.startsWith('http') || bg.startsWith('/') || bg.startsWith('uploads/')) {
    const imageUrl = bg.startsWith('uploads/') ? `/${bg}` : bg
    return {
      backgroundImage: `url(${getCompressedImage(imageUrl, 800)})`,
      backgroundSize: 'cover',
      backgroundPosition: 'center',
      backgroundRepeat: 'no-repeat'
    }
  }
  // 否则使用渐变色
  return { background: bg }
})

const previewShowcase = (item) => {
  if (!item?.image) return
  showImagePreview([getCompressedImage(item.image, 1200)])
}

onMounted(async () => {
  const id = route.params.id || 'study-halfday'

  try {
    const configRes = await axios.get('/api/services/config')
    const allConfigs = configRes?.data?.data || {}
    const config = allConfigs['9'] || {}

    // 直接使用后台课程包数据
    if (config.packages && config.packages[id]) {
      const remotePkg = config.packages[id]
      pkg.value = {
        id,
        name: remotePkg.name || '',
        tag: remotePkg.tag || '',
        price: remotePkg.price || 0,
        headerBg: remotePkg.headerBg || '',
        intro: remotePkg.intro || '',
        studyGoals: remotePkg.studyGoals || [],
        schedule: remotePkg.schedule || [],
        audience: remotePkg.audience || [],
        feeInfo: remotePkg.feeInfo || [],
        tips: remotePkg.tips || [],
        projects: remotePkg.projects || [],
        advantages: remotePkg.advantages || [],
        showcase: remotePkg.showcase || [],
        safetyBriefing: remotePkg.safetyBriefing || [],
        studySummary: remotePkg.studySummary || '',
      }
    }
  } catch (e) {
    console.warn('加载配置失败:', e)
  }

  contentReady.value = true
})

const onApply = () => {
  router.push(`/service-apply/9?package=${pkg.value.id}`)
}
</script>

<style scoped>
.study-detail-page {
  min-height: 100vh;
  background: #fbfbfd;
  padding-bottom: 100px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  max-width: 520px;
  margin: 0 auto;
}

:deep(.van-nav-bar) {
  background-color: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px);
  border: none;
}

:deep(.van-nav-bar__title) {
  font-weight: 600;
  color: #1d1d1f;
}

.pkg-header {
  position: relative;
  padding: 40px 20px 54px;
  overflow: hidden;
}

.header-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, rgba(0,0,0,0.15) 0%, rgba(0,0,0,0.35) 100%), radial-gradient(circle at 80% 20%, rgba(255,255,255,0.18) 0%, transparent 55%);
}

.header-inner {
  position: relative;
  z-index: 1;
  animation: fadeUp 0.6s cubic-bezier(0.25, 1, 0.5, 1);
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
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
  font-size: 24px;
  font-weight: 700;
  color: #fff;
  margin: 0 0 16px;
  letter-spacing: -0.5px;
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

.detail-body {
  animation: fadeUp 0.6s cubic-bezier(0.25, 1, 0.5, 1);
}

.section-card {
  background: #fff;
  margin: 12px 16px;
  padding: 18px;
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
}

.section-card:first-child {
  margin-top: -24px;
  position: relative;
  z-index: 2;
}

.section-title {
  font-size: 16px;
  font-weight: 800;
  color: #1d1d1f;
  margin: 0 0 16px;
  padding-left: 12px;
  border-left: 4px solid #0071e3;
}

.section-text {
  font-size: 15px;
  color: #424245;
  line-height: 1.7;
  margin: 0;
}

/* 课程安排时间线 */
/* 精彩回顾 / 往期活动展示 */
.showcase-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.showcase-item {
  background: #f5f5f7;
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s;
}

.showcase-item:active {
  transform: scale(0.98);
}

.showcase-info {
  padding: 14px 16px 16px;
}

/* 图片加载状态 */
.image-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 140px;
  background: #f5f5f7;
}

.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 140px;
  background: #f5f5f7;
  gap: 8px;
}

.image-error span {
  font-size: 12px;
  color: #c8c8c8;
}

.showcase-title {
  font-size: 15px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 4px;
}

.showcase-desc {
  font-size: 13px;
  color: #86868b;
  line-height: 1.5;
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
  color: #86868b;
  cursor: pointer;
  transition: all 0.25s;
  user-select: none;
}

.toggle-btn.active {
  background: #fff;
  color: #0071e3;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.schedule-list {
  padding-left: 4px;
}

.schedule-item {
  display: grid;
  grid-template-columns: 48px 20px 1fr;
  gap: 12px;
  padding-bottom: 20px;
  position: relative;
}

.schedule-item:last-child {
  padding-bottom: 0;
}

.schedule-time {
  font-size: 13px;
  font-weight: 600;
  color: #1d1d1f;
  padding-top: 2px;
  text-align: right;
}

.schedule-marker {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  padding-top: 5px;
}

.schedule-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #0071e3;
  box-shadow: 0 0 0 4px rgba(0, 113, 227, 0.1);
  position: relative;
  z-index: 1;
}

.schedule-item:not(:last-child) .schedule-marker::after {
  content: "";
  position: absolute;
  left: 50%;
  top: 18px;
  bottom: -20px;
  transform: translateX(-0.5px);
  width: 1px;
  border-left: 1px dashed #d2d2d7;
}

.schedule-content {
  flex: 1;
}

.schedule-name {
  font-size: 15px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 4px;
}

.schedule-desc {
  font-size: 13px;
  color: #86868b;
  line-height: 1.5;
}

.schedule-location {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #0071e3;
  margin-bottom: 4px;
}

.schedule-purpose {
  font-size: 12px;
  color: #636366;
  line-height: 1.5;
  margin-top: 6px;
  padding: 8px 10px;
  background: #f5f5f7;
  border-radius: 8px;
}

.purpose-label {
  display: inline-block;
  font-size: 11px;
  font-weight: 600;
  color: #0071e3;
  background: rgba(0, 113, 227, 0.08);
  padding: 1px 6px;
  border-radius: 4px;
  margin-right: 6px;
}

/* 研学目标 */
.goals-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.goal-item {
  padding: 14px;
  background: #f5f5f7;
  border-radius: 12px;
}

.goal-label {
  font-size: 13px;
  font-weight: 700;
  color: #0071e3;
  margin-bottom: 6px;
}

.goal-content {
  font-size: 14px;
  color: #424245;
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
  gap: 10px;
  font-size: 14px;
  color: #424245;
  line-height: 1.6;
}

.safety-item .van-icon {
  flex-shrink: 0;
  margin-top: 3px;
}

/* 服务项目 */
.project-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.project-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 20px 12px;
  background: #f5f5f7;
  border-radius: 18px;
  font-size: 14px;
  font-weight: 500;
  color: #1d1d1f;
  overflow: visible;
}

.project-item .van-icon {
  flex-shrink: 0;
  min-width: 24px;
  min-height: 24px;
}

/* 服务优势 */
.advantage-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.advantage-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #424245;
}

/* 适合人群 */
.audience-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.audience-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #424245;
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
  color: #86868b;
}

.fee-value {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
}

/* 温馨提示 */
.tips-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tip-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.tip-index {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #f0f2ff;
  color: #0071e3;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.tip-text {
  font-size: 13px;
  color: #424245;
  line-height: 1.6;
  flex: 1;
}

/* 联系客服 */
.contact-section {
  text-align: center;
}

.contact-row {
  font-size: 14px;
  color: #86868b;
  margin: 0 0 8px;
}

.phone-link {
  font-size: 24px;
  font-weight: 700;
  color: #0071e3;
  text-decoration: none;
  display: block;
  margin: 12px 0;
}

.work-time {
  font-size: 12px;
  color: #86868b;
  margin: 8px 0 0;
}

/* 底部操作栏 */
.action-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: min(520px, calc(100% - 40px));
  z-index: 100;
}

.action-bar :deep(.van-button) {
  height: 54px;
  font-size: 17px;
  font-weight: 600;
  box-shadow: 0 10px 20px rgba(0, 113, 227, 0.2);
}

.loading-wrap {
  padding: 100px 0;
  text-align: center;
}
</style>
