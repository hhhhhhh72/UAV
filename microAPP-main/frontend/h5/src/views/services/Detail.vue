<template>
  <div class="service-detail-page" :class="{ 'study-detail-theme': ['9', '6'].includes(serviceId) }">
    <van-nav-bar
      title="服务详情"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
      :border="!['9', '6'].includes(serviceId)"
    />
    <HomeFloatButton />

    <div class="detail-content" v-if="loading">
      <div style="padding: 100px 0; text-align: center;">
        <van-loading vertical>加载中...</van-loading>
      </div>
    </div>

    <div class="detail-content" v-else>
      <!-- [研学 ID: 9 & 培训 ID: 6] 专属 Banner 头部 -->
      <template v-if="['9', '6'].includes(serviceId)">
        <div class="study-banner">
          <img 
            :src="getCompressedImage(serviceConfig.headerImage || (serviceId === '9' ? '/images/study/study-banner.jpg' : '/images/training/training-banner.jpg'), 800)" 
            class="banner-img"
            :style="{ objectPosition: serviceConfig.headerImagePosition || 'center' }"
          />
          <div class="banner-overlay">
            <h1 class="study-hero-title">{{ serviceConfig.name }}</h1>
            <p class="study-hero-slogan">{{ serviceConfig.slogan }}</p>
          </div>
        </div>
      </template>

      <!-- [普通] 服务图标头部 -->
      <template v-else>
        <div class="service-header">
          <div class="service-icon-big" :style="{ background: serviceConfig.mainColor || '#0071e3' }">
            <van-icon :name="normalizeUrl(serviceConfig.icon)" size="48" color="#ffffff" />
          </div>
          <h1 class="service-name">{{ serviceConfig.name }}</h1>
          <p class="service-slogan">{{ serviceConfig.slogan }}</p>
        </div>
      </template>

      <!-- [研学 ID: 9 & 培训 ID: 6] 核心卖点卡片 -->
      <div class="study-highlights-row" v-if="['9', '6'].includes(serviceId) && serviceConfig.highlights && serviceConfig.highlights.length > 0">
        <div v-for="(hl, idx) in serviceConfig.highlights" :key="idx" class="hl-card">
          <div class="hl-icon-wrap">
            <van-icon :name="hl.icon || 'star'" class="hl-icon" />
          </div>
          <div class="hl-text">
            <div class="hl-title">{{ hl.title }}</div>
            <div class="hl-desc">{{ hl.desc }}</div>
          </div>
        </div>
      </div>

      <!-- 服务介绍 -->
      <div class="section-card" v-if="(serviceId !== '6' || !serviceConfig.trainingInfo) && serviceConfig.intro && serviceId !== '9'">
        <h2 class="section-title">服务介绍</h2>
        <p class="section-text">{{ serviceConfig.intro }}</p>
      </div>

      <!-- 研学课程包列表 -->
      <template v-if="serviceId === '9' && studyPackagesList.length > 0">
        <div class="section-card">
          <h2 class="section-title">选择课程</h2>
          <p class="section-text" style="margin-bottom: 16px;">我们为您精心准备了以下研学课程，点击卡片查看详情：</p>
          <div class="study-package-list">
            <div
              v-for="pkg in studyPackagesList"
              :key="pkg.id"
              class="study-package-card"
              :class="{ recommended: pkg.recommended }"
              @click="goToStudyDetail(pkg.id)"
            >
              <div v-if="pkg.recommended" class="recommend-badge">推荐</div>
              <div class="pkg-card-header">
                <div class="pkg-card-name">{{ pkg.name }}</div>
                <div class="pkg-card-tag">{{ pkg.tag }}</div>
              </div>
              <div class="pkg-card-price">
                <span class="currency">¥</span>
                <span class="amount">{{ pkg.price }}</span>
                <span class="unit">/人</span>
              </div>
              <div class="pkg-card-desc">{{ pkg.desc }}</div>
              <div class="pkg-card-highlights" v-if="pkg.highlights && pkg.highlights.length > 0">
                <div v-for="(h, i) in pkg.highlights.slice(0, 4)" :key="i" class="pkg-highlight-item">
                  <van-icon name="success" size="12" color="#06b6d4" />
                  <span>{{ h }}</span>
                </div>
              </div>
              <div class="pkg-card-action">
                <van-button type="primary" size="small" round color="#0071e3">查看详情</van-button>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 服务项目 -->
      <div class="section-card" v-if="(serviceId !== '6' || !serviceConfig.trainingInfo) && serviceConfig.projects && serviceConfig.projects.length > 0">
        <h2 class="section-title">服务项目</h2>
        <div class="project-grid">
          <div 
            v-for="(item, index) in serviceConfig.projects" 
            :key="index" 
            class="project-item"
            :class="{ 'clickable-project': serviceId === '13' || isClickableProject(item) }"
            @click="onProjectClick(item)"
          >
            <van-icon :name="normalizeUrl(item.icon)" size="24" color="#424245" />
            <span>{{ item.name }}</span>
          </div>
        </div>
      </div>

      <!-- 课程安排 (针对研学) -->
      <div class="section-card" v-if="serviceId === '9' && serviceConfig.courseSchedule">
        <h2 class="section-title">课程安排</h2>
        <div class="schedule-container">
          <div v-for="(step, idx) in serviceConfig.courseSchedule" :key="idx" class="schedule-item">
            <div class="schedule-time">{{ step.time }}</div>
            <div class="schedule-marker">
              <div class="schedule-dot"></div>
            </div>
            <div class="schedule-content">
              <div class="schedule-text">{{ step.content }}</div>
              <div class="schedule-remark" v-if="step.remark">{{ step.remark }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 研学费用 (针对研学) -->
      <div class="section-card" v-if="serviceId === '9' && serviceConfig.studyPrice">
        <h2 class="section-title">研学费用</h2>
        <div class="price-list">
          <div class="price-item">
            <span class="label">半日营 - 单纯体验课堂</span>
            <span class="price">{{ serviceConfig.studyPrice }}</span>
          </div>
        </div>
      </div>

      <!-- 服务优势 -->
      <div class="section-card" v-if="(serviceId !== '6' || !serviceConfig.trainingInfo) && serviceConfig.advantages && serviceConfig.advantages.length > 0">
        <h2 class="section-title">服务优势</h2>
        <div class="advantage-list">
          <div v-for="(adv, index) in serviceConfig.advantages" :key="index" class="advantage-item">
            <van-icon name="success" size="16" :color="BRAND_COLOR" />
            <span>{{ adv }}</span>
          </div>
        </div>
      </div>

      <!-- 飞手培训专属内容 -->
      <template v-if="serviceId === '6' && serviceConfig.trainingInfo">
        <div class="section-card">
          <h2 class="section-title" :style="{ borderLeftColor: serviceConfig.mainColor }">报名条件</h2>
          <div class="training-list">
            <div v-for="(cond, idx) in serviceConfig.trainingInfo.conditions" :key="idx" class="training-item">
              {{ cond }}
            </div>
          </div>
        </div>

        <div class="section-card">
          <h2 class="section-title" :style="{ borderLeftColor: serviceConfig.mainColor }">培训费用</h2>
          <div class="price-list">
            <div v-for="(p, idx) in serviceConfig.trainingInfo.prices" :key="idx" class="price-item">
              <span class="label">{{ p.label }}</span>
              <span class="price">{{ p.price }}</span>
            </div>
          </div>
        </div>

        <div class="section-card">
          <h2 class="section-title">教学特色</h2>
          <div class="feature-list">
            <div v-for="(f, idx) in serviceConfig.trainingInfo.features" :key="idx" class="feature-item">
              <div class="feature-title">{{ f.title }}</div>
              <div class="feature-desc">{{ f.desc }}</div>
            </div>
          </div>
        </div>

        <!-- [恢复] 飞手培训图文展示 (已迁移至通用展示) -->
        <div class="section-card" v-if="serviceId === '6' && studyShowcase && studyShowcase.length > 0">
          <h2 class="section-title">图文展示</h2>
          <div class="training-showcase-list">
            <div
              v-for="(item, idx) in studyShowcase"
              :key="idx"
              class="training-showcase-item"
              @click="previewStudy(item)"
            >
              <div class="training-showcase-image">
                <van-image :src="getCompressedImage(item.image, 600)" fit="cover" width="100%" height="100%" />
              </div>
              <div class="training-showcase-info">
                <div class="training-showcase-title">{{ item.title }}</div>
                <div class="training-showcase-desc">{{ item.desc }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- [恢复] 公司简介 -->
        <div class="section-card" v-if="serviceConfig.trainingInfo.companyIntro">
          <h2 class="section-title">公司简介</h2>
          <div class="company-intro">
            <div class="intro-title">{{ serviceConfig.trainingInfo.companyIntro.title }}</div>
            <p class="section-text">{{ serviceConfig.trainingInfo.companyIntro.content }}</p>
          </div>
        </div>

        <!-- [恢复] 执照功能 -->
        <div class="section-card" v-if="serviceConfig.trainingInfo.licenseFunction">
          <h2 class="section-title">执照功能</h2>
          <div class="license-intro">
            <p class="section-text">{{ serviceConfig.trainingInfo.licenseFunction.content }}</p>
            <div class="law-quote">{{ serviceConfig.trainingInfo.licenseFunction.quote }}</div>
          </div>
        </div>
      </template>

        <!-- [恢复] 联系方式 (复刻原版) -->
        <div class="section-card contact-section" v-if="serviceConfig.contactPhone || serviceConfig.address">
          <h2 class="section-title">联系方式</h2>
          <div class="contact-list">
            <!-- 电话 1 -->
            <div class="contact-item" v-if="serviceConfig.contactPhone">
              <div class="contact-icon-wrap">
                <van-icon name="phone-o" />
              </div>
              <div class="contact-content">
                <div class="contact-label">联系电话</div>
                <div class="contact-value">{{ serviceConfig.contactPhone }}</div>
              </div>
              <van-button
                type="primary"
                size="small"
                round
                class="call-btn"
                :url="'tel:' + serviceConfig.contactPhone"
              >
                拨打
              </van-button>
            </div>

            <!-- 电话 2 (针对飞手培训、研学等) -->
            <div class="contact-item" v-if="serviceConfig.contactPhone2">
              <div class="contact-icon-wrap">
                <van-icon name="phone-o" />
              </div>
              <div class="contact-content">
                <div class="contact-label">咨询热线</div>
                <div class="contact-value">{{ serviceConfig.contactPhone2 }}</div>
              </div>
              <van-button
                type="primary"
                size="small"
                round
                class="call-btn"
                :url="'tel:' + serviceConfig.contactPhone2"
              >
                拨打
              </van-button>
            </div>

            <!-- 联系地址 -->
            <div class="contact-item" v-if="serviceConfig.address">
              <div class="contact-icon-wrap">
                <van-icon name="location-o" />
              </div>
              <div class="contact-content">
                <div class="contact-label">联系地址</div>
                <div class="contact-value">{{ serviceConfig.address }}</div>
              </div>
            </div>
          </div>
        </div>
    </div>

    <!-- 底部操作栏 -->
    <div class="action-bar">
      <van-button type="primary" block @click="onApply" :color="BRAND_COLOR">
        {{ actionButtonText }}
      </van-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast, showImagePreview } from 'vant'
import HomeFloatButton from '@/components/HomeFloatButton.vue'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const serviceId = String(route.params.id)

const loading = ref(true)
const serviceConfig = ref({})
const studyShowcase = ref([])

// 研学课程包列表
const studyPackagesList = computed(() => {
  if (serviceId !== '9') return []
  const packages = serviceConfig.value?.packages || {}
  return Object.entries(packages)
    .filter(([_, pkg]) => pkg)
    .map(([id, pkg]) => ({
      id,
      name: pkg.name || '',
      tag: pkg.tag || '',
      price: pkg.price || 0,
      recommended: pkg.recommended || false,
      desc: pkg.desc || pkg.intro || '',
      highlights: pkg.cardHighlights || []
    }))
    .sort((a, b) => (b.recommended ? 1 : 0) - (a.recommended ? 1 : 0))
})

// 跳转到研学课程包详情
const goToStudyDetail = (pkgId) => {
  router.push(`/study/${pkgId}`)
}

const normalizeUrl = (url) => {
  if (!url) return ''
  // 如果是远程地址、base64 或已经以 / 开头，直接返回
  if (url.startsWith('http') || url.startsWith('data:') || url.startsWith('/')) return url
  // 如果包含点（如 .svg, .png）或者是存放在特定目录下的，补齐开头的 /
  if (url.includes('.') || url.includes('/')) return `/${url}`
  // 否则认为是 Vant 内置图标名，不加斜杠（加了斜杠 Vant 会当成图片加载导致 404）
  return url
}

// 获取压缩后的图片URL
const getCompressedImage = (url, width = 800) => {
  if (!url) return ''
  const normalizedUrl = normalizeUrl(url)
  if (!normalizedUrl) return ''
  return `/api/image?url=${encodeURIComponent(normalizedUrl)}&width=${width}&quality=75`
}

// 苹果风品牌色常量
const BRAND_COLOR = '#0071e3'

const fetchServiceData = async () => {
  loading.value = true
  try {
    // 1. 获取基础配置
    const configRes = await axios.get('/api/services/config')
    const allConfigs = configRes?.data?.data || {}
    const config = allConfigs[serviceId] || {}
    serviceConfig.value = config

    // 2. 如果是研学，优先使用 config 里的 studyShowcase，没有则尝试拉取旧接口（兜底）
    if (serviceId === '9') {
      if (config.studyShowcase && config.studyShowcase.length > 0) {
        studyShowcase.value = config.studyShowcase
      } else {
        const showcaseRes = await axios.get('/api/study/showcase')
        studyShowcase.value = showcaseRes?.data?.data || []
      }
    }
  } catch (e) {
    showToast('加载配置失败')
    console.error(e)
  } finally {
    loading.value = false
  }
}

const previewStudy = (item) => {
  if (!item?.image) return
  showImagePreview([getCompressedImage(item.image, 1200)])
}

const actionButtonText = computed(() => {
  if (['1', '4', '8'].includes(serviceId)) return '立即下单'
  if (serviceId === '13') return '立即报名'
  if (['6', '9'].includes(serviceId)) return '立即报名'
  return '立即办理'
})

const onApply = () => {
  if (serviceId === '8') {
    window.location.href = 'https://app.wzsjy.com:8446/h5/#/pages/diy/diy?pageId=130&title=%E6%97%A0%E4%BA%BA%E6%9C%BA%E5%A4%96%E5%8D%96%E9%85%8D%E9%80%81&jyauthcode='
  } else if (serviceId === '9') {
    // 研学服务：如果有课程包，跳转到第一个推荐的或第一个课程包
    const packages = studyPackagesList.value
    if (packages.length > 0) {
      const recommended = packages.find(p => p.recommended)
      const targetPkg = recommended || packages[0]
      router.push(`/study/${targetPkg.id}`)
    } else {
      router.push(`/service-apply/${serviceId}`)
    }
  } else {
    router.push(`/service-apply/${serviceId}`)
  }
}

// 判断服务项目是否可点击跳转
const isClickableProject = (item) => {
  if (serviceId === '1' && item.name === '医疗运输') return true
  return false
}

const onProjectClick = (item) => {
  if (serviceId === '13') {
    const roleMap = {
      '裁判员': 'referee', '教练员': 'coach', '运动员': 'athlete', '俱乐部': 'club'
    }
    const role = roleMap[item.name]
    if (role) {
      router.push({ path: `/service-apply/${serviceId}`, query: { role } })
    }
  } else if (serviceId === '1' && item.name === '医疗运输') {
    router.push('/medical/certification/status')
  }
}

onMounted(() => {
  // 医疗配送服务直接跳转到专属下单页
  if (serviceId === '14') {
    router.replace('/medical/order/create')
    return
  }
  fetchServiceData()
})
</script>

<style scoped>
/* 苹果风全局变量与基础 */
.service-detail-page {
  min-height: 100vh;
  background: #fbfbfd; /* 苹果官网常用底色 */
  padding-bottom: 100px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
}

/* 沉浸式导航栏 */
:deep(.van-nav-bar) {
  background-color: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px); /* 毛玻璃效果 */
  border: none;
}
:deep(.van-nav-bar__title) {
  font-weight: 600;
  color: #1d1d1f;
}

.detail-content {
  animation: appleFadeIn 0.6s cubic-bezier(0.25, 1, 0.5, 1);
  width: 100%;
  max-width: 520px; /* 主要适配移动端：PC 上也保持手机宽度观感 */
  margin: 0 auto;
}

@keyframes appleFadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 研学专属 Banner */
.study-banner {
  position: relative;
  width: 100%;
  aspect-ratio: 5 / 3; /* 顶部背景图统一 5:3 */
  height: auto;
  margin: 0;
  overflow: hidden;
}

.banner-img {
  width: 100%;
  height: 100%;
  object-fit: cover; /* 核心：确保填充且不拉伸 */
  object-position: center; /* 默认居中对齐 */
  transition: transform 0.8s ease;
}

.banner-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 60px 24px 30px;
  background: linear-gradient(to top, rgba(0,0,0,0.6) 0%, transparent 100%);
  color: #fff;
}

.study-hero-title {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 4px;
  letter-spacing: -0.5px;
}

.study-hero-slogan {
  font-size: 15px;
  font-weight: 400;
  opacity: 0.85;
  margin: 0;
}

/* 核心亮点卡片（改为横向滑动，避免垂直堆叠导致的重叠） */
.study-highlights-row {
  margin: 16px 0 24px; /* 不再覆盖 Banner，避免遮挡口号标语 */
  display: flex;
  overflow-x: auto;
  padding: 0 16px;
  gap: 14px;
  -webkit-overflow-scrolling: touch;
}
.study-highlights-row::-webkit-scrollbar {
  display: none; /* 隐藏滚动条 */
}

.hl-card {
  flex: 0 0 280px; /* 固定宽度，支持横向滚动 */
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  padding: 20px;
  border-radius: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 12px 30px rgba(0,0,0,0.06);
  border: 1px solid rgba(255,255,255,0.8);
}

.hl-icon-wrap {
  background: #f5f5f7;
  width: 52px;
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  flex-shrink: 0;
}

.hl-icon {
  font-size: 28px;
  color: #0071e3; /* Apple Blue */
}

.hl-title {
  font-size: 17px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 2px;
}

.hl-desc {
  font-size: 13px;
  color: #86868b;
  line-height: 1.4;
}

/* 普通头部 */
.service-header {
  background: #fff;
  padding: 48px 24px;
  text-align: center;
  border-bottom: 1px solid #f1f5f9;
}

.service-icon-big {
  width: 90px;
  height: 90px;
  border-radius: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  box-shadow: 0 12px 24px rgba(0,0,0,0.12);
}

.service-name {
  font-size: 24px;
  font-weight: 700;
  color: #1d1d1f;
  margin-bottom: 8px;
}

/* 通用卡片设计 */
.section-card {
  background: #fff;
  margin: 12px 16px;
  padding: 18px;
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
  position: relative;
  z-index: 1;
}

/* 针对研学服务，确保第一个板块有足够空间 */
.study-detail-theme .section-card:first-of-type {
  margin-top: 0; /* 使用上方的 margin-bottom 32px 统一控制间距 */
}

.section-title {
  font-size: 16px;
  font-weight: 800;
  color: #1d1d1f;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid #0071e3; /* 恢复品牌特征蓝色边条 */
}

.section-text {
  font-size: 15px;
  color: #424245;
  line-height: 1.6;
}

/* 服务项目网格 */
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
  background: #fff;
  border: 1px solid #f1f5f9;
  border-radius: 18px;
  font-size: 14px;
  font-weight: 500;
  color: #1d1d1f;
  transition: all 0.2s ease;
}

.project-item:active {
  background: #f5f5f7;
  transform: scale(0.96);
}

.clickable-project {
  cursor: pointer;
  border-color: #0071e3;
  background: #f0f7ff;
  position: relative;
}
.clickable-project::after {
  content: '';
  position: absolute;
  top: 8px;
  right: 8px;
  width: 6px;
  height: 6px;
  border-top: 2px solid #0071e3;
  border-right: 2px solid #0071e3;
  transform: rotate(45deg);
}

/* 优势列表 */
.advantage-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  color: #424245;
  margin-bottom: 14px;
}

/* 底部胶囊按钮 */
.action-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: min(520px, calc(100% - 40px)); /* 跟随页面最大宽度 */
  padding: 0;
  background: transparent;
  border-top: none;
  z-index: 100;
}

.action-bar :deep(.van-button) {
  height: 54px;
  border-radius: 27px; /* 胶囊形状 */
  font-size: 17px;
  font-weight: 600;
  box-shadow: 0 10px 20px rgba(0, 113, 227, 0.2);
}

.study-showcase {
  margin-top: 24px;
  border-top: 1px solid #f1f5f9;
  padding-top: 24px;
}

.study-subtitle {
  font-size: 17px;
  font-weight: 700;
  color: #1d1d1f;
  margin-bottom: 16px;
  letter-spacing: -0.3px;
}

.study-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

/* 图文展示 */
.study-item {
  border: none;
  background: #f5f5f7;
  border-radius: 20px;
  overflow: hidden;
  transition: transform 0.3s ease;
  margin-bottom: 0; /* 使用 grid 的 gap 控制 */
}

.study-info {
  padding: 16px;
}

.study-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 6px;
  letter-spacing: -0.2px;
}

.study-desc {
  font-size: 13px;
  color: #86868b;
  line-height: 1.5;
  margin: 0;
}

.section-card {
  background: #fff;
  margin: 12px 16px;
  padding: 18px;
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
  position: relative;
  z-index: 1;
}

.section-title {
  font-size: 16px;
  font-weight: 800;
  color: #1d1d1f;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid #0071e3; /* 恢复品牌特征蓝色边条 */
}

/* 价格与列表样式美化（恢复原版虚线） */
.price-list {
  margin-top: 4px;
}

.price-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px dashed #ebedf0; /* 恢复原版虚线 */
}

.price-item:last-child {
  border-bottom: none;
}

.price-item .label {
  font-size: 14px;
  color: #323233;
}

.price-item .price {
  font-size: 16px;
  font-weight: 700;
  color: #ee0a24; /* 恢复原版警告红 */
}

.training-list {
  padding: 4px 0;
}

.training-item {
  font-size: 14px;
  color: #424245;
  margin-bottom: 12px;
  line-height: 1.7;
  padding-left: 12px;
  position: relative;
}

.training-item::before {
  content: "";
  position: absolute;
  left: 0;
  top: 9px;
  width: 4px;
  height: 4px;
  background: #0071e3; /* 改为品牌蓝 */
  border-radius: 50%;
}

.feature-item {
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f5f5f7;
}

.feature-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.feature-title {
  font-size: 17px;
  font-weight: 700;
  color: #1d1d1f;
  margin-bottom: 8px;
  letter-spacing: -0.3px;
}

.feature-desc {
  font-size: 14px;
  color: #86868b;
  line-height: 1.6;
}

/* 恢复飞手培训图文展示样式（左图右文） */
.training-showcase-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.training-showcase-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  background: #f7f8fa;
}

.training-showcase-image {
  flex: 0 0 110px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
}

.training-showcase-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.training-showcase-title {
  font-size: 15px;
  font-weight: 700;
  color: #323233;
  margin-bottom: 4px;
}

.training-showcase-desc {
  font-size: 12px;
  color: #646566;
  line-height: 1.5;
  display: -webkit-box;
  line-clamp: 2;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 研学课程表样式 */
.schedule-container {
  padding: 10px 0;
}

.schedule-item {
  display: grid;
  grid-template-columns: 96px 22px 1fr;
  column-gap: 14px;
  padding-bottom: 24px;
  position: relative;
}

.schedule-item:last-child {
  padding-bottom: 0;
}

.schedule-time {
  font-size: 13px;
  color: #86868b;
  padding-top: 2px;
  text-align: right;
  white-space: pre-line; /* 支持 “早班/午班” 两行时间 */
  line-height: 1.15;
}

.schedule-marker {
  position: relative;
  width: 22px;
  display: flex;
  justify-content: center;
  padding-top: 6px; /* 让点与第一行文本视觉对齐 */
}

.schedule-item:not(:last-child) .schedule-marker::after {
  content: "";
  position: absolute;
  left: 50%;
  top: 14px; /* dot center */
  bottom: -24px; /* 延伸到下一项 */
  transform: translateX(-0.5px);
  width: 1px;
  border-left: 1px dashed #d2d2d7;
}

.schedule-dot {
  width: 8px;
  height: 8px;
  background: #0071e3;
  border-radius: 50%;
  position: relative;
  z-index: 1;
  box-shadow: 0 0 0 4px rgba(0, 113, 227, 0.1);
}

.schedule-content {
  flex: 1;
}

.schedule-text {
  font-size: 15px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 4px;
}

.schedule-remark {
  font-size: 13px;
  color: #86868b;
}

/* 飞手培训专项样式恢复 */
.intro-title {
  font-size: 17px;
  font-weight: 700;
  color: #1d1d1f;
  margin-bottom: 12px;
}

.law-quote {
  margin-top: 16px;
  padding: 16px;
  background: #f5f5f7;
  border-radius: 16px;
  font-size: 13px;
  color: #86868b;
  line-height: 1.6;
  border-left: 4px solid #0071e3;
}

/* 联系方式复刻原版样式 */
.contact-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.contact-icon-wrap {
  width: 32px;
  height: 32px;
  background: #f0f7ff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0071e3;
  font-size: 18px;
}

.contact-content {
  flex: 1;
}

.contact-label {
  font-size: 12px;
  color: #969799;
  margin-bottom: 2px;
}

.contact-value {
  font-size: 15px;
  font-weight: 500;
  color: #323233;
}

.call-btn {
  height: 28px !important;
  padding: 0 12px !important;
  font-size: 12px !important;
}

/* 联系客服 (废弃) */
.contact-card {
  display: none;
}

.service-icon-big :deep(.van-icon__image) {
  filter: brightness(0) invert(1);
}

/* 研学课程包列表样式 */
.study-package-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.study-package-card {
  background: #fff;
  border-radius: 20px;
  padding: 20px;
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  border: 2px solid transparent;
  transition: all 0.3s;
  cursor: pointer;
}

.study-package-card:active {
  transform: scale(0.98);
}

.study-package-card.recommended {
  border-color: #0071e3;
}

.recommend-badge {
  position: absolute;
  top: -1px;
  right: 16px;
  background: linear-gradient(135deg, #0071e3 0%, #06b6d4 100%);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 0 0 10px 10px;
}

.pkg-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.pkg-card-name {
  font-size: 16px;
  font-weight: 700;
  color: #1d1d1f;
  flex: 1;
}

.pkg-card-tag {
  font-size: 11px;
  color: #0071e3;
  background: rgba(0, 113, 227, 0.08);
  padding: 3px 10px;
  border-radius: 20px;
  font-weight: 500;
}

.pkg-card-price {
  display: flex;
  align-items: baseline;
  margin-bottom: 10px;
}

.pkg-card-price .currency {
  font-size: 14px;
  font-weight: 700;
  color: #ee0a24;
}

.pkg-card-price .amount {
  font-size: 28px;
  font-weight: 800;
  color: #ee0a24;
  line-height: 1;
  margin: 0 2px;
}

.pkg-card-price .unit {
  font-size: 12px;
  color: #86868b;
}

.pkg-card-desc {
  font-size: 13px;
  color: #86868b;
  line-height: 1.6;
  margin-bottom: 12px;
}

.pkg-card-highlights {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.pkg-highlight-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #424245;
}

.pkg-card-action {
  display: flex;
  justify-content: flex-end;
}
</style>
