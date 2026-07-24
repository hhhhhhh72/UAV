<template>
  <div class="home-page">
    <van-pull-refresh v-model="loading" @refresh="onRefresh">
      <!-- 沉浸式背景 (固定定位) -->
      <div class="video-header">
        <img 
          class="bg-video"
          :src="homeHeaderImage"
          :style="{ objectPosition: homeHeaderImagePosition }"
          alt="background"
        />
        <div class="video-mask"></div>
      </div>

      <!-- 视频占位块 (撑开文档流) -->
      <div class="video-placeholder"></div>

      <!-- 悬浮顶部区域：定位+搜索 -->
      <div class="header-section float-header">
        <div class="location-bar">
          <span class="location-text">温州市 <van-icon name="arrow-down" size="12" /></span>
          <div class="weather-info">20°C 晴</div>
        </div>
        <div class="search-bar">
          <van-search :placeholder="searchPlaceholder" shape="round" background="transparent" />
        </div>
      </div>

      <!-- 功能金刚区 - 悬浮卡片风格 -->
      <div class="main-functions overlay-card">
        <van-swipe class="function-swipe" :loop="false" indicator-color="#1677ff">
          <van-swipe-item v-for="(page, pageIndex) in servicePages" :key="pageIndex">
            <div class="function-grid">
              <div 
                v-for="item in page" 
                :key="item.id"
                class="function-item"
                :style="{ visibility: item.isEmpty ? 'hidden' : 'visible' }"
                @click="!item.isEmpty && goToService(item.id)"
              >
                <!-- AI 蓝紫色渐变风格 -->
                <div class="function-icon-wrapper" :style="{ background: item.color }">
                  <van-icon :name="item.icon" size="24" color="#ffffff" />
                </div>
                <span class="function-name">{{ item.name }}</span>
              </div>
            </div>
          </van-swipe-item>
        </van-swipe>
      </div>

      <!-- 消息通知栏 -->
      <div class="notice-bar-section">
        <van-notice-bar
          left-icon="volume-o"
          :scrollable="false"
          background="transparent"
          color="#333"
        >
          <van-swipe
            vertical
            class="notice-swipe"
            :autoplay="3000"
            :touchable="false"
            :show-indicators="false"
          >
            <van-swipe-item v-for="(msg, idx) in notices" :key="idx">{{ msg }}</van-swipe-item>
          </van-swipe>
        </van-notice-bar>
      </div>

      <!-- 核心服务卡片区 -->
      <div class="content-area">
        <!-- 轮播 Banner 区 -->
        <div class="banner-section">
          <van-swipe class="my-swipe" :autoplay="5000" indicator-color="white">
            <van-swipe-item v-for="(item, index) in banners" :key="index" @click="handleBannerClick(item)">
              <img :src="item.image" class="banner-image" alt="banner" />
            </van-swipe-item>
          </van-swipe>
        </div>

        <!-- 左右分栏推荐 -->
        <div class="recommend-grid">
          <div class="recommend-card blue-card" @click="$router.push('/cases')">
            <h4>精选案例</h4>
            <p>行业应用示范</p>
            <van-icon name="video-o" size="32" class="card-icon" />
          </div>
          <div class="recommend-card orange-card" @click="$router.push('/services')">
            <h4>服务大厅</h4>
            <p>一站式办理</p>
            <van-icon name="service-o" size="32" class="card-icon" />
          </div>
        </div>

        <!-- 为你推荐 (服务列表) -->
        <div class="service-feed">
          <div class="section-title">为你推荐</div>
          <div 
            v-for="service in displayServices" 
            :key="service.id"
            class="feed-card"
            @click="goToDetail(service.id)"
          >
            <div class="feed-icon" :style="{ background: service.color }">
              <van-icon :name="service.icon" size="24" color="#ffffff" />
            </div>
            <div class="feed-content">
              <div class="feed-title">{{ service.name }}</div>
              <div class="feed-desc">{{ service.description }}</div>
            </div>
            <!-- iOS 风格箭头 -->
            <van-icon name="arrow" color="#969799" size="16" />
          </div>
        </div>
      </div>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import axios from 'axios'

const router = useRouter()

const normalizeUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http') || url.startsWith('data:') || url.startsWith('/')) return url
  if (url.includes('.') || url.includes('/')) return `/${url}`
  return url
}

const homeHeaderImage = ref('https://www-cdn.djiits.com/cms/uploads/4d6128a30991074b6bad20e7e13a0c62.png')
const homeHeaderImagePosition = ref('center')
const notices = ref(['交享点无人机外卖配送正式上线', '新开通江心屿无人机外卖配送'])

// 下拉刷新
const loading = ref(false)
const onRefresh = () => {
  setTimeout(() => {
    showToast('刷新成功')
    loading.value = false
  }, 1000)
}

// 搜索栏 Placeholder 轮播
const searchPlaceholder = ref('搜索服务/案例')
const searchKeywords = [
  '搜索服务/案例',
  '无人机外卖',
  '行业应用示范',
  '飞行服务',
  '低空研学',
  '无人机吊运'
]
let searchTimer = null

onMounted(() => {
  let index = 0
  searchTimer = setInterval(() => {
    index = (index + 1) % searchKeywords.length
    searchPlaceholder.value = searchKeywords[index]
  }, 3000)
})

onMounted(async () => {
  try {
    const res = await axios.get('/api/services/config')
    const allConfigs = res?.data?.data || {}
    const cfg = allConfigs?._home || {}
    homeHeaderImage.value = normalizeUrl(cfg.headerImage || homeHeaderImage.value)
    homeHeaderImagePosition.value =
      cfg.headerImagePosition === 'top'
        ? 'top'
        : cfg.headerImagePosition === 'bottom'
          ? 'bottom'
          : 'center'
    // 加载后端配置的轮播消息
    if (Array.isArray(cfg.notices) && cfg.notices.length > 0) {
      notices.value = cfg.notices.filter(m => m && typeof m === 'string' && m.trim())
    }
    // 加载后端配置的轮播 Banner
    if (Array.isArray(cfg.banners) && cfg.banners.length > 0) {
      banners.value = cfg.banners.filter(b => b && b.image).map(b => ({
        image: normalizeUrl(b.image),
        link: b.link || ''
      }))
    }
  } catch (e) {
    // ignore，使用默认值
  }
})

onUnmounted(() => {
  if (searchTimer) {
    clearInterval(searchTimer)
  }
})

// 快捷服务 - 模仿支付宝首页图标
const quickServices = ref([
  { id: 'flight', name: '飞行服务', icon: '/icons/flight.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' }, // 蓝紫色
  { id: 8, name: '无人机外卖', icon: '/icons/delivery.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' }, // 淡蓝色
  { id: 1, name: '无人机物流', icon: '/icons/logistics-drone.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' }, // 海蓝色
  { id: 10, name: '无人机销售', icon: '/icons/shop.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' }, // 橙色
  { id: 'reviews', name: '服务评价', icon: 'comment-o', color: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }, // 粉红色
  { id: 6, name: '飞手培训', icon: '/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' }, // 橙色
  { id: 9, name: '低空研学', icon: '/icons/study.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' }, // 海蓝色
  { id: 13, name: '无人机赛事', icon: '/icons/competition.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' }, // 红色渐变
  { id: 14, name: '医疗配送', icon: 'shield-o', color: 'linear-gradient(135deg, #34d399 0%, #059669 100%)' }, // 绿色渐变
  { id: 2, name: '政务服务', icon: 'eye-o', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' }, // 蓝紫色
  { id: 3, name: '无人机托管', icon: '/icons/maintenance.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' }, // 归类为海蓝色
  { id: 5, name: '无人机表演', icon: '/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' }, // 归类为蓝紫色
  { id: 7, name: '无人机租赁', icon: 'coupon-o', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' }, // 归类为淡蓝色
  { id: 4, name: '无人机吊运', icon: '/icons/lifting.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' }, // 归类为橙色
  { id: 11, name: '金融服务', icon: '/icons/finance.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' }, // 归类为蓝紫色
  { id: 12, name: '维修服务', icon: 'setting-o', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' } // 归类为淡蓝色
])

// 将服务分组，每页 7 个 + 1 个更多
const servicePages = computed(() => {
  const pages = []
  const pageSize = 7
  const moreItem = { id: 'more', name: '更多服务', icon: 'apps-o', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' }

  for (let i = 0; i < quickServices.value.length; i += pageSize) {
    const page = quickServices.value.slice(i, i + pageSize)
    // 填充空位
    while (page.length < 7) {
      page.push({ id: `empty-${pages.length}-${page.length}`, isEmpty: true })
    }
    // 添加更多服务按钮
    page.push(moreItem)
    pages.push(page)
  }
  return pages
})

// 推荐服务列表
const displayServices = ref([
  { id: 1, name: '无人机物流', description: '城市极速配送，解决最后一公里难题', icon: '/icons/logistics-drone.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 2, name: '政务服务', description: '高效环保监测、交通疏导、安全巡查', icon: 'eye-o', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 3, name: '无人机托管', description: '专业机库托管，定期维护保养', icon: '/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' }, // 绿色渐变
  { id: 4, name: '无人机吊运', description: '建筑材料、基站设备高空精准吊运', icon: '/icons/lifting.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' }
])

const goToService = (id) => {
  if (id === 'flight') {
    window.location.href = 'https://wx.zndkfx.com'
  } else if (id === 8) {
    goToDelivery()
  } else if (id === 9) {
    router.push('/study')
  } else if (id === 14) {
    router.push('/medical/order/create')
  } else if (id === 'reviews') {
    router.push('/reviews')
  } else if (id === 'more') {
    // 切换到服务列表 Tab (假设 Tabbar中有这个路径，或者跳转到列表页)
    router.push('/services')
  } else {
    router.push(`/service-detail/${id}`)
  }
}

const goToDetail = (id) => {
  router.push(`/service-detail/${id}`)
}

// Banner 数据（从后端 _home.banners 配置加载，以下为默认兜底）
const banners = ref([
  {
    image: 'https://wenzhoumall-prod.oss-cn-shanghai.aliyuncs.com/test/shop/20250930/0fa02eb2dc8b4a6382784fedc0b44dc0.jpg?Expires=3337231191&OSSAccessKeyId=LTAI5tSbLByCMG16D3eoErCU&Signature=Zk8QXbZAJhw08908Er3iuy9dKg0%3D',
    link: 'delivery'
  },
  {
    image: 'https://www-cdn.djiits.com/dps/3e196dbfade1b1734dbbb335dde5de12.jpg?w=1184&h=592',
    link: '/cases/1'
  },
  {
    image: 'https://images.unsplash.com/photo-1506947411487-a56738267384?ixlib=rb-4.0.3&auto=format&fit=crop&w=1000&q=80',
    link: '/cases/2'
  }
])

const handleBannerClick = (item) => {
  if (item.link === 'delivery') {
    goToDelivery()
  } else if (item.link) {
    // router.push(item.link) // 暂时注释，等待案例页面开发
    console.log('Navigate to:', item.link)
  }
}

const goToDelivery = () => {
  window.location.href = 'https://app.wzsjy.com:8446/h5/#/pages/diy/diy?pageId=130&title=%E6%97%A0%E4%BA%BA%E6%9C%BA%E5%A4%96%E5%8D%96%E9%85%8D%E9%80%81&jyauthcode='
}
</script>

<style scoped>
.home-page {
  background-color: transparent;
  min-height: 100vh;
  /* 与 Tabbar 实际占位高度保持一致，确保底部卡片完整露出且“贴合” */
  padding-bottom: var(--tabbar-height);
  width: 100%;
  max-width: var(--page-max-width);
  margin: 0 auto;

  /* === Home 首屏可调参数 ===
   * - 背景高度固定，不跟“调参”变化（避免背景被压缩裁切）
   * - 只调金刚区覆盖高度，下面整体会跟着动
   */
  --home-bg-height: 300px;
  --home-overlay-overlap: 40px;
}

/* 沉浸式视频背景 */
.video-header {
  height: var(--home-bg-height);
  position: fixed; /* 固定定位 */
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: var(--page-max-width);
  z-index: 0;
  overflow: hidden;
  background: #000;
}

.video-placeholder {
  height: var(--home-bg-height);
  width: 100%;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

.bg-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  display: block;
}

.video-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(180deg, rgba(0,0,0,0.3) 0%, rgba(0,0,0,0) 50%, rgba(0,0,0,0.1) 100%);
  pointer-events: none;
}

/* 悬浮顶部区域 */
.float-header {
  position: fixed; /* 固定定位 */
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: var(--page-max-width);
  z-index: 100;
  padding: calc(12px + env(safe-area-inset-top)) 16px 20px;
  color: #fff;
  background: linear-gradient(180deg, rgba(0,0,0,0.4) 0%, transparent 100%);
}

.location-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
}

.location-text {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0,0,0,0.5);
}

.weather-info {
  color: #fff;
  text-shadow: 0 1px 2px rgba(0,0,0,0.5);
}

.search-bar :deep(.van-search) {
  padding: 0;
  background: transparent;
}

.search-bar :deep(.van-search__content) {
  background: rgba(255, 255, 255, 0.15); /* 毛玻璃背景 */
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 20px; /* iOS 风格大圆角 */
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.search-bar :deep(.van-field__control) {
  color: #fff;
}

.search-bar :deep(.van-icon) {
  color: rgba(255, 255, 255, 0.8);
}

.search-bar :deep(input::placeholder) {
  color: rgba(255, 255, 255, 0.6);
}

/* 功能金刚区 - 悬浮卡片 (磨玻璃) */
.overlay-card {
  position: relative;
  z-index: 10;
  margin: calc(-1 * var(--home-overlay-overlap)) 12px 10px; /* 向上重叠，左右留空 */
  background: rgba(255, 255, 255, 0.25); /* 极简透明度 */
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-top: 1px solid rgba(255, 255, 255, 0.3); /* 细腻顶部边框 */
  border-radius: 24px 24px 24px 24px; /* iOS 风格大圆角 */
  padding: 24px 0 0; /* 增加顶部留白 */
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.05); /* 柔和投影 */
}

.function-swipe {
  padding-bottom: 28px;
}

.function-swipe :deep(.van-swipe__indicator) {
  width: 12px;
  height: 4px;
  border-radius: 2px;
  background-color: rgba(0, 0, 0, 0.1);
  opacity: 1;
  transition: all 0.3s;
}

.function-swipe :deep(.van-swipe__indicator--active) {
  width: 20px;
  background-color: #1a1a1a; /* 深色指示器 */
}

.function-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  row-gap: 20px; /* 增加行间距 */
}

.function-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px; /* 增加图标与文字间距 */
  cursor: pointer;
}

.function-icon-wrapper {
  width: 44px; /* 稍微增大图标容器 */
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 14px; /* iOS 风格圆角 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08); /* 悬浮投影 */
}

/* 强制 SVG 图片图标变白 */
.function-icon-wrapper :deep(.van-icon__image) {
  filter: brightness(0) invert(1);
}

.function-name {
  font-size: 13px; /* 微调字号 */
  color: #1a1a1a; /* 主体深色文本 */
  font-weight: 500;
}

/* 消息通知 */
.notice-bar-section {
  margin: 0 12px 20px; /* 增加底部间距 */
  background: rgba(255, 255, 255, 0.25); /* 极简透明度 */
  border-radius: 16px; /* iOS 风格圆角 */
  overflow: hidden;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3); /* 细腻边框 */
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05); /* 柔和投影 */
}

.notice-swipe {
  height: 40px;
  line-height: 40px;
}

/* 内容区域 */
.content-area {
  padding: 0 12px;
}

/* 特色卡片 */
.banner-section {
  margin-bottom: 12px;
  border-radius: 16px; /* iOS 风格圆角 */
  overflow: hidden;
}

.my-swipe .van-swipe-item {
  color: #fff;
  font-size: 20px;
  text-align: center;
}

.banner-image {
  width: 100%;
  height: 160px;
  object-fit: cover;
  display: block;
}

.recommend-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 20px;
}

.recommend-card {
  padding: 16px;
  border-radius: 16px; /* iOS 风格圆角 */
  position: relative;
  height: 100px;
}

.blue-card {
  background: linear-gradient(rgba(0,0,0,0.3), rgba(0,0,0,0.3)), url('https://images.unsplash.com/photo-1473968512647-3e447244af8f?ixlib=rb-4.0.3&auto=format&fit=crop&w=600&q=80');
  background-size: cover;
  background-position: center;
  color: #fff;
}

.orange-card {
  background: linear-gradient(rgba(0,0,0,0.3), rgba(0,0,0,0.3)), url('https://www-cdn.djiits.com/dps/71685a7a83e4c70907f3c504f6806561.jpg');
  background-size: cover;
  background-position: center;
  color: #fff;
}

.recommend-card h4 {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 4px;
  text-shadow: 0 1px 2px rgba(0,0,0,0.5);
}

.recommend-card p {
  font-size: 12px;
  opacity: 0.9;
  text-shadow: 0 1px 2px rgba(0,0,0,0.5);
}

.card-icon {
  position: absolute;
  bottom: 12px;
  right: 12px;
  opacity: 0.8;
  color: #fff;
}

/* 推荐列表 */
.service-feed {
  background: rgba(255, 255, 255, 0.25); /* 极度克制的透明度 0.1 ~ 0.25 */
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 20px; /* 圆润圆角 16~20px */
  padding: 24px; /* 内部大留白 */
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.05); /* 轻微投影 */
  border: 1px solid rgba(255, 255, 255, 0.3); /* 细腻边框 */
  margin-bottom: 20px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #1a1a1a; /* 主体深色文本 */
  letter-spacing: 0.5px;
}

.feed-card {
  display: flex;
  align-items: center;
  gap: 16px; /* 增加间距 */
  padding: 16px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.03); /* 极淡分割线 */
}

.feed-card:first-child {
  padding-top: 0;
}

.feed-card:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.feed-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px; /* iOS 风格圆角 */
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08); /* 图标悬浮感 */
}

/* 强制 SVG 图片图标变白 (适配推荐列表中的图片图标) */
.feed-icon :deep(.van-icon__image) {
  filter: brightness(0) invert(1);
}

.feed-content {
  flex: 1;
}

.feed-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a; /* 主体深色 */
  margin-bottom: 6px;
}

.feed-desc {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.5); /* 次要信息克制淡灰 */
  line-height: 1.4;
}
</style>
