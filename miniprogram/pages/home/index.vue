<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 沉浸式背景 (固定定位) -->
      <view class="video-header">
        <image 
          class="bg-video"
          :src="headerImage || '/static/home-bg.jpg'"
          mode="aspectFill"
        />
        <view class="video-mask" />
      </view>

      <!-- 占位：高度刚好到搜索栏底部 -->
      <view class="video-placeholder" />

      <!-- 头部：搜索栏（固定定位） -->
      <view class="header-section float-header">
        <view class="location-bar">
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

      <!-- 轮播图 - 紧挨搜索栏下方 -->
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

      <!-- 核心内容区 -->
      <view class="content-area">
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

        <!-- 消息通知栏 -->
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

        <!-- 贴吧式分类Tab -->
        <view class="category-tabs overlay-card-sm">
          <scroll-view scroll-x class="tabs-scroll" :show-scrollbar="false">
            <view 
              v-for="cat in categories" 
              :key="cat.id"
              class="tab-chip"
              :class="{ active: activeCategory === cat.id }"
              @tap="switchCategory(cat.id)"
            >
              <image v-if="cat.icon" :src="cat.icon" mode="aspectFit" class="tab-icon-img" />
              <text>{{ cat.name }}</text>
            </view>
          </scroll-view>
        </view>

        <!-- 贴吧式需求流 -->
        <view class="post-feed overlay-card">
          <view class="feed-header">
            <text class="section-title">任务大厅</text>
            <view class="feed-sort" @tap="handlePublish">
              <image src="/static/icons/plus.svg" mode="aspectFit" class="sort-icon" />
              <text class="sort-text">发布</text>
            </view>
          </view>

          <view 
            v-for="post in feedList" 
            :key="post.id"
            class="post-card"
            @tap="goToDemand(post.id)"
          >
            <!-- 发布者信息 -->
            <view class="post-user-row">
              <view class="post-avatar" :style="{ background: post.avatarColor || '#e8f2fc' }">
                <image v-if="post.avatar" :src="post.avatar" mode="aspectFill" class="avatar-img" />
                <text v-else class="avatar-letter">{{ post.userName?.charAt(0) || '?' }}</text>
              </view>
              <view class="post-user-info">
                <text class="post-username">{{ post.userName }}</text>
                <text class="post-time">{{ post.timeAgo }}</text>
              </view>
              <view class="post-badge" :class="'badge-' + (post.urgency || 'normal')">
                <text>{{ urgencyLabel(post.urgency) }}</text>
              </view>
            </view>

            <!-- 需求标题 -->
            <view class="post-title">{{ post.title }}</view>

            <!-- 需求标签 -->
            <view class="post-tags" v-if="post.tags && post.tags.length">
              <text v-for="tag in post.tags" :key="tag" class="post-tag">{{ tag }}</text>
            </view>

            <!-- 需求简述 + 预算 -->
            <view class="post-desc" v-if="post.description">{{ post.description }}</view>
            <view class="post-meta">
              <view class="post-budget" v-if="post.budget">
                <image src="/static/icons/wallet.svg" mode="aspectFit" class="meta-icon" />
                <text>￥{{ post.budget }}</text>
              </view>
              <view class="post-location" v-if="post.location">
                <image src="/static/icons/location.svg" mode="aspectFit" class="meta-icon" />
                <text>{{ post.location }}</text>
              </view>
              <view class="post-deadline" v-if="post.deadline">
                <image src="/static/icons/clock.svg" mode="aspectFit" class="meta-icon" />
                <text>{{ post.deadline }}</text>
              </view>
            </view>

            <!-- 互动栏 -->
            <view class="post-actions">
              <view class="action-item" @tap.stop="likePost(post.id)">
                <image src="/static/icons/heart.svg" mode="aspectFit" class="action-icon" />
                <text>{{ post.likeCount || 0 }}</text>
              </view>
              <view class="action-item" @tap.stop="goToComments(post.id)">
                <image src="/static/icons/comment.svg" mode="aspectFit" class="action-icon" />
                <text>{{ post.commentCount || 0 }}</text>
              </view>
              <view class="action-item" @tap.stop="sharePost(post.id)">
                <image src="/static/icons/share.svg" mode="aspectFit" class="action-icon" />
                <text>分享</text>
              </view>
              <view class="action-item bid-btn" @tap.stop="goToBid(post.id)">
                <text>我要竞标</text>
              </view>
            </view>
          </view>

          <!-- 空状态 -->
          <view class="feed-empty" v-if="feedList.length === 0 && !feedLoading">
            <image src="/static/icons/empty-feed.svg" mode="aspectFit" class="empty-img" />
            <text class="empty-text">暂无需求，去发布第一个吧</text>
          </view>

          <!-- 加载更多 -->
          <view class="feed-more" v-if="hasMore" @tap="loadMore">
            <text>{{ feedLoading ? '加载中...' : '查看更多' }}</text>
          </view>
        </view>

        <!-- 推荐卡片（保留原有两卡） -->
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
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'
import { request, BASE_URL } from '../../utils/request'

const searchKeywords = ref(['搜索服务/案例', '无人机外卖'])
const activeSearchIndex = ref(0)

const onSearchSwiperChange = (e) => {
  activeSearchIndex.value = e.detail.current
}

const handleSearchClick = () => {
  const keyword = searchKeywords.value[activeSearchIndex.value]
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

const notices = ref([])
const headerImage = ref('')
const contactPhone = ref('023-55550500')

const quickServices = ref([])

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

const displayServices = ref([])

// ---- 贴吧式需求流数据 ----
const feedList = ref([])
const categories = ref([
  { id: 'all', name: '全部', icon: '' },
  { id: 'inspection', name: '巡检', icon: '/static/icons/drone.svg' },
  { id: 'plant', name: '植保', icon: '/static/icons/plant.svg' },
  { id: 'logistics', name: '物流', icon: '/static/icons/package.svg' },
  { id: 'mapping', name: '测绘', icon: '/static/icons/map.svg' },
  { id: 'emergency', name: '应急', icon: '/static/icons/shield.svg' },
  { id: 'other', name: '其他', icon: '' }
])
const activeCategory = ref('all')
const feedLoading = ref(false)
const hasMore = ref(true)
let feedPage = 1

const urgencyLabel = (u) => ({ urgent: '紧急', normal: '普通', low: '低' }[u] || '普通')

const switchCategory = (catId) => {
  activeCategory.value = catId
  feedPage = 1
  loadFeed()
}

const loadFeed = async () => {
  feedLoading.value = true
  try {
    const res = await request({
      url: '/api/v1/demands',
      data: { category: activeCategory.value === 'all' ? '' : activeCategory.value, page: feedPage, page_size: 10 }
    })
    const list = (res.data || []).map(d => ({
      id: d.id,
      title: d.title,
      description: d.description?.substring(0, 80),
      userName: d.publisher?.name || '匿名用户',
      avatar: d.publisher?.avatar || '',
      avatarColor: d.publisher?.avatarColor || '#e8f2fc',
      timeAgo: formatTimeAgo(d.created_at),
      tags: d.tags || d.type ? [d.type] : [],
      budget: d.budget || d.price_range,
      location: d.location,
      deadline: d.deadline ? formatDeadline(d.deadline) : '',
      urgency: d.urgency || 'normal',
      likeCount: d.like_count || 0,
      commentCount: d.comment_count || 0
    }))
    if (feedPage === 1) feedList.value = list
    else feedList.value = [...feedList.value, ...list]
    hasMore.value = list.length >= 10
  } catch { /* keep current */ } finally { feedLoading.value = false }
}

const formatTimeAgo = (t) => {
  if (!t) return ''
  const diff = (Date.now() - new Date(t).getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff/60) + '分钟前'
  if (diff < 86400) return Math.floor(diff/3600) + '小时前'
  return Math.floor(diff/86400) + '天前'
}

const formatDeadline = (d) => {
  if (!d) return ''
  const date = new Date(d)
  return `${date.getMonth()+1}/${date.getDate()}截止`
}

const loadMore = () => { feedPage++; loadFeed() }
const goToDemand = (id) => safeNavigateTo(`/pages/demands/detail?id=${id}`)
const goToBid = (id) => safeNavigateTo(`/pages/demands/bid?id=${id}`)
const goToComments = (id) => safeNavigateTo(`/pages/demands/detail?id=${id}#comments`)
const likePost = (id) => uni.showToast({ title: '已收藏', icon: 'none' })
const sharePost = (id) => uni.showToast({ title: '已复制链接', icon: 'none' })
const handlePublish = () => safeNavigateTo('/pages/demands/publish')

const banners = ref([
  { image: '/static/home-bg.jpg', link: '' }
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
  if (item.id === 'flight') {
    openExternal('https://wx.zndkfx.com')
    return
  }
  if (item.id === 8) {
    goToDelivery()
    return
  }
  if (item.id === 9) {
    safeNavigateTo('/pages/study/index')
    return
  }
  if (item.id === 'contact') {
    uni.makePhoneCall({ phoneNumber: contactPhone.value })
    return
  }
  if (item.id === 'more') {
    safeSwitchTab('/pages/services/index')
    return
  }
  if (item.id === 'news' || item.id === 'policy') {
    uni.showToast({ title: '功能开发中', icon: 'none' })
    return
  }
  goToDetail(item.id)
}

const goToDetail = (id) => {
  safeNavigateTo(`/pages/services/detail?id=${id}`)
}

const goToDelivery = () => {
  openExternal('https://app.wzsjy.com:8446/h5/#/pages/diy/diy?pageId=130')
}

const handleBannerClick = (item) => {
  if (item.link === 'delivery') {
    goToDelivery()
  } else if (item.link) {
    navigateTo(item.link)
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
  try {
    const res = await request({ url: '/api/services/config' })
    const cfg = (res?._home) || {}
    console.log('[Home] config loaded:', JSON.stringify({ banners: cfg.banners?.length, headerImage: cfg.headerImage?.substring(0,30) }))
    
    // 轮播消息
    if (Array.isArray(cfg.notices) && cfg.notices.length > 0) {
      notices.value = cfg.notices.filter(m => m && typeof m === 'string' && m.trim())
    }
    // Banners: use API data directly (backend already returns full URLs)
    if (Array.isArray(cfg.banners) && cfg.banners.length > 0) {
      banners.value = cfg.banners.filter(b => b.image)
    }
    // 背景图
    if (cfg.headerImage) {
      headerImage.value = cfg.headerImage
    }
    // 快捷入口
    if (Array.isArray(cfg.quickServices) && cfg.quickServices.length > 0) {
      quickServices.value = cfg.quickServices
    }
    // 推荐服务
    if (Array.isArray(cfg.displayServices) && cfg.displayServices.length > 0) {
      displayServices.value = cfg.displayServices
    }
    // 搜索关键词
    if (Array.isArray(cfg.searchKeywords) && cfg.searchKeywords.length > 0) {
      searchKeywords.value = cfg.searchKeywords
    }
    // 联系电话
    if (cfg.contactPhone) contactPhone.value = cfg.contactPhone
  } catch (e) {
    // ignore, use defaults
  }

  // #ifdef MP-WEIXIN
  // preloadPage 在微信小程序中不可用，忽略
  // #endif

  loadFeed()
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

/* 占位高度 = 状态栏(~45px) + 搜索栏(~60px) + 轮播图与搜索栏的间距 */
.video-placeholder {
  height: 105px;
  width: 100%;
}

.float-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  padding: 45px 16px 10px;
  color: #fff;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.4) 0%, transparent 100%);
}

.location-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
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

/* 轮播图样式 - 紧挨搜索栏 */
.banner-section {
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.banner-swipe {
  width: 100%;
  height: 160px;
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

.content-area {
  padding: 0 12px;
}

.overlay-card {
  position: relative;
  z-index: 5;
  margin: 0 0 10px;
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
  margin: 0 0 20px;
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
  color: #1a1a1a;
}

/* 贴吧式分类Tab */
.overlay-card-sm {
  position: relative;
  z-index: 5;
  margin: 0 0 8px;
  padding: 12px 0;
}

.category-tabs {
  background: rgba(255, 255, 255, 0.22);
  border-radius: 16px;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.tabs-scroll {
  white-space: nowrap;
  padding: 0 12px;
}

.tab-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  margin-right: 8px;
  border-radius: 20px;
  font-size: 13px;
  color: #555;
  background: transparent;
  transition: all 0.2s;
}

.tab-chip.active {
  background: rgba(25, 137, 250, 0.12);
  color: #1989fa;
  font-weight: 600;
}

.tab-icon-img {
  width: 16px;
  height: 16px;
}

/* 贴吧式需求流 */
.post-feed {
  margin-bottom: 12px;
}

.post-feed.overlay-card {
  padding: 24px;
}

.feed-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.feed-sort {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  background: rgba(25, 137, 250, 0.1);
  border-radius: 16px;
}

.sort-icon {
  width: 14px;
  height: 14px;
  opacity: 0.7;
}

.sort-text {
  font-size: 13px;
  color: #1989fa;
  font-weight: 500;
}

.post-card {
  padding: 16px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.post-card:first-of-type {
  padding-top: 0;
}

.post-card:last-of-type {
  border-bottom: none;
  padding-bottom: 0;
}

.post-user-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.post-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
}

.avatar-letter {
  font-size: 14px;
  font-weight: 600;
  color: #1989fa;
}

.post-user-info {
  flex: 1;
}

.post-username {
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  display: block;
}

.post-time {
  font-size: 11px;
  color: rgba(0, 0, 0, 0.4);
}

.post-badge {
  padding: 3px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.badge-urgent {
  background: rgba(255, 59, 48, 0.1);
  color: #ff3b30;
}

.badge-normal {
  background: rgba(0, 0, 0, 0.05);
  color: #666;
}

.badge-low {
  background: rgba(52, 199, 89, 0.1);
  color: #34c759;
}

.post-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8px;
  line-height: 1.4;
}

.post-tags {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.post-tag {
  padding: 3px 10px;
  background: rgba(25, 137, 250, 0.08);
  color: #1989fa;
  font-size: 12px;
  border-radius: 6px;
}

.post-desc {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.5);
  margin-bottom: 8px;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.post-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.post-budget, .post-location, .post-deadline {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.5);
}

.post-budget {
  color: #ff6b35;
  font-weight: 500;
}

.meta-icon {
  width: 14px;
  height: 14px;
  opacity: 0.6;
}

.post-actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.4);
}

.action-icon {
  width: 16px;
  height: 16px;
  opacity: 0.5;
}

.bid-btn {
  margin-left: auto;
  padding: 6px 16px;
  background: linear-gradient(135deg, #1989fa 0%, #3b82f6 100%);
  color: #fff;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(25, 137, 250, 0.3);
}

.feed-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 0;
  gap: 12px;
}

.empty-img {
  width: 80px;
  height: 80px;
  opacity: 0.3;
}

.empty-text {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.4);
}

.feed-more {
  text-align: center;
  padding: 16px 0 4px;
  font-size: 13px;
  color: #1989fa;
}

/* keep existing recommend grid */
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