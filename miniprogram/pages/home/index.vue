<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 沉浸式背景 (固定定位) -->
      <view class="video-header">
        <image 
          class="bg-video"
          src="/static/home-bg.jpg"
          mode="aspectFill"
        />
        <view class="video-mask" />
      </view>

      <view class="video-placeholder" />

      <view class="header-section float-header">
        <view class="location-bar">
          <view class="location-text">
            重庆市
            <image class="arrow-down-icon" src="/static/icons/arrow-down.svg" mode="aspectFit" />
          </view>
        </view>
        <view class="search-bar" @tap="handleSearchClick">
          <view class="search-box">
            <image class="search-icon-img" src="/static/icons/search.svg" mode="aspectFit" />
            <swiper 
              class="search-swiper" 
              vertical 
              autoplay 
              circular 
              interval="3000" 
              @change="onSearchSwiperChange"
            >
              <swiper-item v-for="(word, index) in searchKeywords" :key="index">
                <text class="search-placeholder">{{ word }}</text>
              </swiper-item>
            </swiper>
          </view>
        </view>
      </view>

      <!-- 功能金刚区 -->
      <view class="main-functions overlay-card">
        <swiper 
          class="function-swipe" 
          :current="activeFunctionPage" 
          @change="onFunctionChange"
          :indicator-dots="false"
        >
          <swiper-item v-for="(page, pageIndex) in servicePages" :key="pageIndex">
            <view class="function-grid">
              <view 
                v-for="item in page" 
                :key="item.id"
                class="function-item"
                :style="{ visibility: item.isEmpty ? 'hidden' : 'visible' }"
                @tap="handleFunctionTap(item)"
              >
                <view class="function-icon-wrapper" :style="{ background: item.color }">
                  <image 
                    v-if="item.icon && !item.isEmpty" 
                    :src="item.icon" 
                    mode="aspectFit" 
                    class="function-icon-image" 
                  />
                </view>
                <text class="function-name">{{ item.name }}</text>
              </view>
            </view>
          </swiper-item>
        </swiper>
        <view class="swiper-dots">
          <view 
            v-for="(page, pageIndex) in servicePages" 
            :key="`dot-${pageIndex}`"
            class="dot"
            :class="{ active: pageIndex === activeFunctionPage }"
          />
        </view>
      </view>

      <!-- 消息通知栏 (修正显示问题) -->
      <view class="notice-bar-section">
        <view class="notice-inner">
          <image class="notice-icon-img" src="/static/icons/volume.svg" mode="aspectFit" />
          <swiper 
            class="notice-swipe" 
            vertical 
            autoplay 
            interval="3000" 
            circular
            :indicator-dots="false"
            :touchable="false"
          >
            <swiper-item v-for="(msg, index) in notices" :key="index" class="notice-item">
              <text class="notice-text">{{ msg }}</text>
            </swiper-item>
          </swiper>
        </view>
      </view>

      <!-- 核心内容区 -->
      <view class="content-area">
        <view class="banner-section">
          <swiper 
            class="banner-swipe" 
            autoplay 
            interval="5000" 
            circular
            :current="activeBanner"
            @change="onBannerChange"
          >
            <swiper-item v-for="(item, index) in banners" :key="index" @tap="handleBannerClick(item)">
              <image :src="item.image" class="banner-image" mode="aspectFill" />
            </swiper-item>
          </swiper>
          <view class="banner-dots">
            <view 
              v-for="(item, index) in banners" 
              :key="`banner-dot-${index}`"
              class="banner-dot"
              :class="{ active: index === activeBanner }"
            />
          </view>
        </view>

        <!-- 左右推荐卡片 (替换图标为 SVG) -->
        <view class="recommend-grid">
          <view class="recommend-card blue-card" @tap="navigateTo('/pages/cases/index')">
            <view>
              <view class="recommend-title">精选案例</view>
              <text class="recommend-subtitle">行业应用示范</text>
            </view>
            <image class="recommend-icon-img" src="/static/icons/drone-show-v2.svg" mode="aspectFit" />
          </view>
          <view class="recommend-card orange-card" @tap="navigateTo('/pages/services/index')">
            <view>
              <view class="recommend-title">服务大厅</view>
              <text class="recommend-subtitle">一站式办理</text>
            </view>
            <image class="recommend-icon-img" src="/static/icons/service.svg" mode="aspectFit" />
          </view>
        </view>

        <!-- 推荐列表 -->
        <view class="service-feed">
          <view class="section-title">为你推荐</view>
          <view 
            v-for="service in displayServices" 
            :key="service.id"
            class="feed-card"
            @tap="goToDetail(service.id)"
          >
            <view class="feed-icon" :style="{ background: service.color }">
              <image :src="service.icon" mode="aspectFit" class="feed-icon-image" />
            </view>
            <view class="feed-content">
              <view class="feed-title">{{ service.name }}</view>
              <view class="feed-desc">{{ service.description }}</view>
            </view>
            <text class="feed-arrow">›</text>
          </view>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'
import { request } from '../../utils/request'

const searchKeywords = ['搜索服务/案例', '产业供需对接', '培训认证', '无人机交易', '合同签约', '保险金融']
const activeSearchIndex = ref(0)

const onSearchSwiperChange = (e) => {
  activeSearchIndex.value = e.detail.current
}

const handleSearchClick = () => {
  const keyword = searchKeywords[activeSearchIndex.value]
  const p = keyword === '搜索服务/案例' ? '' : keyword
  safeSwitchTab(`/pages/services/index?keyword=${encodeURIComponent(p)}`)
}

// 获取胶囊按钮位置
const getCapsuleInfo = () => {
  let capsule = { top: 44, bottom: 76, height: 32, left: 281, right: 368, width: 87 }
  // #ifdef MP-WEIXIN
  capsule = uni.getMenuButtonBoundingClientRect()
  // #endif
  return capsule
}

const capsuleInfo = ref(getCapsuleInfo())
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 20)

const notices = ref(['重庆无人机产业平台正式上线', '无人机培训认证夏季班火热招生中', '产业供需对接大厅已开放入驻'])

const quickServices = ref([
  { id: 'supply-demand', name: '产业供需对接', icon: '/static/icons/flight.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 'training-cert', name: '培训认证', icon: '/static/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 'drone-trade', name: '无人机交易', icon: '/static/icons/shop.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 'contract-sign', name: '合同签约', icon: '/static/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' },
  { id: 'insurance-finance', name: '保险金融', icon: '/static/icons/finance.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 'emergency-resource', name: '应急资源协同', icon: '/static/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' },
  { id: 'contact', name: '联系客服', icon: '/static/icons/service.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' }
])

const servicePages = computed(() => {
  const pages = []
  const pageSize = 7
  const moreItem = { id: 'more', name: '更多服务', icon: '/static/icons/apps.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' }

  for (let i = 0; i < quickServices.value.length; i += pageSize) {
    const page = quickServices.value.slice(i, i + pageSize)
    while (page.length < 7) {
      page.push({ id: `empty-${pages.length}-${page.length}`, isEmpty: true, color: 'transparent' })
    }
    page.push(moreItem)
    pages.push(page)
  }
  return pages
})

const activeFunctionPage = ref(0)

const displayServices = ref([
  { id: 'supply-demand', name: '产业供需对接', description: '发布需求、浏览供应、在线竞标', icon: '/static/icons/flight.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 'training-cert', name: '培训认证', description: 'CAAC执照考证、UTC认证培训', icon: '/static/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 'drone-trade', name: '无人机交易', description: '整机购买、配件采购、维修服务', icon: '/static/icons/shop.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 'contract-sign', name: '合同签约', description: '标准化合同模板、在线电子签章', icon: '/static/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' },
  { id: 'insurance-finance', name: '保险金融', description: '无人机保单、金融贷款服务', icon: '/static/icons/finance.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 'emergency-resource', name: '应急资源协同', description: '应急救援案例、资源统筹调度', icon: '/static/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' }
])

const banners = ref([
  { image: 'https://images.unsplash.com/photo-1506947411487-a56738267384?auto=format&fit=crop&w=1000&q=80', link: '' },
  { image: 'https://www-cdn.djiits.com/dps/3e196dbfade1b1734dbbb335dde5de12.jpg?w=1184&h=592', link: '/pages/cases/detail?id=1' },
  { image: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?auto=format&fit=crop&w=1000&q=80', link: '/pages/cases/detail?id=2' }
])

const activeBanner = ref(0)

onPullDownRefresh(() => {
  setTimeout(() => {
    uni.showToast({ title: '刷新成功', icon: 'none' })
    uni.stopPullDownRefresh()
  }, 800)
})

const handleFunctionTap = (item) => {
  if (item.isEmpty) return
  if (item.id === 'contact') {
    uni.makePhoneCall({ phoneNumber: '02312345678' })
    return
  }
  if (item.id === 'more') {
    safeSwitchTab('/pages/services/index')
    return
  }
  uni.showToast({ title: '功能开发中', icon: 'none' })
}

const goToDetail = (id) => {
  safeNavigateTo(`/pages/services/detail?id=${id}`)
}

const handleBannerClick = (item) => {
  if (item.link) {
    safeNavigateTo(item.link)
  }
}

const navigateTo = (path) => {
  safeNavigateTo(path)
}

const openExternal = (url) => {
  safeNavigateTo(`/pages/webview/index?src=${encodeURIComponent(url)}`)
}

const onFunctionChange = (event) => {
  activeFunctionPage.value = event.detail.current
}

const onBannerChange = (event) => {
  activeBanner.value = event.detail.current
}

onMounted(async () => {
  // 加载后端配置的轮播消息
  try {
    const res = await request({ url: '/api/services/config' })
    const cfg = (res?.data || res)?._home || {}
    if (Array.isArray(cfg.notices) && cfg.notices.length > 0) {
      notices.value = cfg.notices.filter(m => m && typeof m === 'string' && m.trim())
    }
  } catch (e) {
    // ignore，使用默认值
  }

  // #ifdef MP-WEIXIN
  // 预加载详情页，提升秒开感
  uni.preloadPage({ url: '/pages/services/detail?id=1' })
  uni.preloadPage({ url: '/pages/services/detail?id=6' })
  // #endif
})

</script>

<style scoped>
.home-page {
  min-height: 100vh;
  position: relative;
  padding-bottom: 20px;
}

.video-header {
  height: 420px;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 0;
  overflow: hidden;
  background: #000;
}

.bg-video {
  width: 100%;
  height: 100%;
  display: block;
}

.video-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.3) 0%, rgba(0, 0, 0, 0) 50%, rgba(0, 0, 0, 0.1) 100%);
  pointer-events: none;
}

.video-placeholder {
  height: 350px;
  width: 100%;
}

.float-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  padding: 65px 16px 10px; /* 增加顶部填充，避开微信胶囊按钮 */
  color: #fff;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.4) 0%, transparent 100%);
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
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.arrow-down-icon {
  width: 12px;
  height: 12px;
  opacity: 0.9;
}

.search-bar {
  width: 100%;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.12);
  border-radius: 20px;
  padding: 6px 12px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  backdrop-filter: blur(10px);
}

.search-icon-img {
  width: 14px;
  height: 14px;
  opacity: 0.85;
}

.search-swiper {
  flex: 1;
  height: 24px;
}

.search-placeholder {
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  line-height: 24px;
}

.overlay-card {
  position: relative;
  z-index: 5;
  margin: -100px 12px 10px; /* 从 -60px 调整为 -100px，使其整体往上升 */
  background: rgba(255, 255, 255, 0.22);
  border-radius: 24px;
  padding: 20px 0 10px;
  backdrop-filter: blur(20px);
  border-top: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.05);
}

.function-swipe {
  padding-bottom: 0px;
  height: 180px;
}

.function-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  row-gap: 18px;
  padding: 0 20px;
}

.function-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.function-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.function-icon-image {
  width: 24px;
  height: 24px;
}

.function-name {
  font-size: 13px;
  color: #1a1a1a;
  font-weight: 500;
}

.swiper-dots {
  display: flex;
  justify-content: center;
  gap: 6px;
  padding-bottom: 16px;
}

.dot {
  width: 12px;
  height: 4px;
  border-radius: 2px;
  background: rgba(0, 0, 0, 0.1);
  transition: all 0.3s;
}

.dot.active {
  width: 20px;
  background: #1a1a1a;
}

.notice-bar-section {
  margin: 0 12px 20px;
}

.notice-inner {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 16px;
  padding: 6px 12px;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
}

.notice-icon-img {
  width: 18px;
  height: 18px;
  opacity: 0.9;
}

.notice-swipe {
  height: 40px;
  flex: 1;
}

.notice-item {
  display: flex;
  align-items: center;
}

.notice-text {
  font-size: 14px;
  color: #333;
}

.content-area {
  padding: 0 12px;
}

.banner-section {
  margin-bottom: 12px;
  border-radius: 16px;
  overflow: hidden;
  position: relative;
}

.banner-image {
  width: 100%;
  height: 160px;
  display: block;
}

.banner-dots {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 10px;
  display: flex;
  justify-content: center;
  gap: 6px;
  pointer-events: none;
}

.banner-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.55);
}

.banner-dot.active {
  background: rgba(255, 255, 255, 1);
}

.recommend-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}

.recommend-card {
  padding: 16px;
  border-radius: 16px;
  height: 100px;
  color: #fff;
  position: relative;
  overflow: hidden;
}

.blue-card {
  background: linear-gradient(rgba(0, 0, 0, 0.3), rgba(0, 0, 0, 0.3)),
    url('https://images.unsplash.com/photo-1473968512647-3e447244af8f?auto=format&fit=crop&w=600&q=80');
  background-size: cover;
}

.orange-card {
  background: linear-gradient(rgba(0, 0, 0, 0.3), rgba(0, 0, 0, 0.3)),
    url('https://www-cdn.djiits.com/dps/71685a7a83e4c70907f3c504f6806561.jpg');
  background-size: cover;
}

.recommend-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.recommend-subtitle {
  font-size: 12px;
}

.recommend-icon-img {
  position: absolute;
  right: 12px;
  bottom: 12px;
  width: 32px;
  height: 32px;
  opacity: 0.8;
  filter: brightness(0) invert(1);
}

.service-feed {
  background: rgba(255, 255, 255, 0.25);
  border-radius: 20px;
  padding: 24px;
  margin-bottom: 20px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.05);
  backdrop-filter: blur(20px);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #1a1a1a;
}

.feed-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
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
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.feed-icon-image {
  width: 24px;
  height: 24px;
}

.feed-content {
  flex: 1;
}

.feed-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 6px;
  color: #1a1a1a;
}

.feed-desc {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.5);
}

.feed-arrow {
  font-size: 16px;
  color: #c8c9cc;
}
</style>
