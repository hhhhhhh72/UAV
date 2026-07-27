<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 头部：搜索栏 -->
      <view class="header-section">
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

        <!-- 贴吧式任务大厅 -->
        <view class="service-feed">
          <view class="section-title">任务大厅</view>
          
          <!-- 分类Tab -->
          <scroll-view scroll-x class="cat-scroll" :show-scrollbar="false">
            <view class="cat-scroll-inner">
              <view 
                v-for="cat in categories" :key="cat.id"
                class="cat-chip" :class="{ on: activeCategory === cat.id }"
                @tap="switchCategory(cat.id)"
              >{{ cat.name }}</view>
            </view>
          </scroll-view>

          <!-- 需求卡片 -->
          <view v-if="feedList.length" class="card-list">
            <view 
              v-for="post in feedList" :key="post.id"
              class="demand-card" @tap="goToDemand(post.id)"
            >
              <view class="card-head">
                <view class="card-avatar" :style="{ background: post.avatarColor || '#e8f2fc' }">
                  <text class="card-avatar-text">{{ (post.userName || '?')[0] }}</text>
                </view>
                <view class="card-user">
                  <text class="card-username">{{ post.userName }}</text>
                  <text class="card-time">{{ post.timeAgo }}</text>
                </view>
                <text v-if="post.urgency === 'urgent'" class="card-urgent">紧急</text>
              </view>
              <view class="card-title">{{ post.title }}</view>
              <view class="card-row" v-if="post.tags && post.tags.length">
                <text v-for="t in post.tags" :key="t" class="card-tag">{{ t }}</text>
              </view>
              <view class="card-row card-meta">
                <text v-if="post.budget" class="card-price">￥{{ post.budget }}</text>
                <text v-if="post.location" class="card-loc">{{ post.location }}</text>
                <text v-if="post.deadline" class="card-date">{{ post.deadline }}</text>
              </view>
              <view class="card-foot">
                <text class="card-stat">❤️ {{ post.likeCount || 0 }}</text>
                <text class="card-stat">💬 {{ post.commentCount || 0 }}</text>
                <view class="card-spacer" />
                <view class="card-bid-btn" @tap.stop="goToBid(post.id)">竞标</view>
              </view>
            </view>
          </view>
          <view v-else class="card-empty">暂无需求</view>
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
const feedList = ref([])
const categories = ref([
  { id: '', name: '全部' }, { id: 'inspection', name: '巡检' }, { id: 'plant', name: '植保' },
  { id: 'logistics', name: '物流' }, { id: 'mapping', name: '测绘' }, { id: 'emergency', name: '应急' }
])
const activeCategory = ref('')

const loadFeed = async () => {
  try {
    const res = await request({ url: '/api/v1/demands', data: { category: activeCategory.value, page: 1, page_size: 10 } })
    feedList.value = (res.data || []).map(d => ({
      id: d.id, title: d.title, userName: d.publisher?.name || '匿名', avatarColor: d.publisher?.avatarColor || '#e8f2fc',
      timeAgo: fmtAgo(d.created_at), tags: d.tags || [], budget: d.budget || d.price_range,
      location: d.location, deadline: d.deadline ? fmtDead(d.deadline) : '',
      urgency: d.urgency || 'normal', likeCount: d.like_count || 0, commentCount: d.comment_count || 0
    }))
  } catch { feedList.value = [] }
}
const fmtAgo = (t) => { if(!t)return''; const d=(Date.now()-new Date(t).getTime())/1000; return d<60?'刚刚':d<3600?Math.floor(d/60)+'分钟前':d<86400?Math.floor(d/3600)+'小时前':Math.floor(d/86400)+'天前' }
const fmtDead = (d) => { if(!d)return''; const dd=new Date(d); return (dd.getMonth()+1)+'/'+dd.getDate()+'截止' }
const switchCategory = (id) => { activeCategory.value = id; loadFeed() }
const goToDemand = (id) => safeNavigateTo('/pages/demands/detail?id='+id)
const goToBid = (id) => safeNavigateTo('/pages/demands/bid?id='+id)

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
.home-page { min-height: 100vh; background: #eff3f8; padding-bottom: 20px; }

.header-section { padding: 12px 16px 8px; background: #1989fa; }

.search-box {
  display: flex; align-items: center; gap: 8px;
  background: rgba(255,255,255,0.2); border-radius: 20px; padding: 8px 14px;
}
.search-icon-img { width: 16px; height: 16px; opacity: 0.7; filter: brightness(0) invert(1); }
.search-swiper { flex: 1; height: 24px; }
.search-placeholder { color: rgba(255,255,255,0.85); font-size: 14px; line-height: 24px; }

.banner-section { margin: 10px 12px 12px; border-radius: 16px; overflow: hidden; }
.banner-swipe { width: 100%; height: 160px; }
.banner-image { width: 100%; height: 160px; display: block; }
.banner-dots { display: flex; justify-content: center; gap: 6px; margin-top: -20px; position: relative; z-index: 2; }
.banner-dot { width: 6px; height: 6px; border-radius: 50%; background: rgba(0,0,0,0.2); }
.banner-dot.active { background: #1989fa; }

.content-area { padding: 0 12px; }

.overlay-card {
  margin: 0 0 10px; background: #fff; border-radius: 20px;
  padding: 20px 0 10px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);
}
.function-swipe { height: 180px; }
.function-grid { display: grid; grid-template-columns: repeat(4,1fr); row-gap: 18px; padding: 0 20px; }
.function-item { display: flex; flex-direction: column; align-items: center; gap: 10px; }
.function-icon-wrapper { width:44px;height:44px;border-radius:14px;display:flex;align-items:center;justify-content:center;box-shadow:0 4px 12px rgba(0,0,0,0.08); }
.function-icon-image { width:24px; height:24px; }
.function-name { font-size:13px; color:#1a1a1a; font-weight:500; }

.swiper-dots { display:flex; justify-content:center; gap:6px; padding-bottom:16px; }
.dot { width:12px; height:4px; border-radius:2px; background:rgba(0,0,0,0.1); transition:all 0.3s; }
.dot.active { width:20px; background:#1a1a1a; }

.notice-bar-section { margin:0 0 12px; }
.notice-inner {
  display:flex; align-items:center; gap:10px;
  background:#fff; border-radius:14px; padding:8px 14px;
  box-shadow:0 2px 8px rgba(0,0,0,0.04);
}
.notice-icon-img { width:18px; height:18px; opacity:0.9; }
.notice-swipe { height:36px; flex:1; }
.notice-text { font-size:13px; color:#333; }

.section-title { font-size:18px; font-weight:600; margin-bottom:20px; color:#1a1a1a; }

.service-feed {
  background:#fff; border-radius:20px; padding:24px;
  margin-bottom: 20px; box-shadow:0 2px 12px rgba(0,0,0,0.06);
}

.cat-scroll { margin-bottom: 20px; }
.cat-scroll-inner { display:flex; gap:8px; }
.cat-chip {
  padding:8px 18px; border-radius:20px; font-size:13px;
  color:#666; background:#f0f2f5; white-space:nowrap;
}
.cat-chip.on { background:rgba(25,137,250,0.12); color:#1989fa; font-weight:600; }

.card-list { display:flex; flex-direction:column; }
.demand-card { padding:18px 0; border-bottom:1px solid #f0f0f0; }
.demand-card:last-child { border-bottom:none; padding-bottom:0; }

.card-head { display:flex; align-items:center; gap:10px; margin-bottom:10px; }
.card-avatar { width:32px; height:32px; border-radius:50%; display:flex; align-items:center; justify-content:center; }
.card-avatar-text { font-size:13px; font-weight:600; color:#1989fa; }
.card-user { flex:1; }
.card-username { font-size:14px; font-weight:500; color:#1a1a1a; display:block; }
.card-time { font-size:11px; color:#999; }
.card-urgent { padding:2px 10px; border-radius:10px; background:rgba(255,59,48,0.1); color:#ff3b30; font-size:11px; font-weight:500; }

.card-title { font-size:16px; font-weight:600; color:#1a1a1a; margin-bottom:8px; line-height:1.4; }

.card-row { display:flex; gap:6px; margin-bottom:8px; flex-wrap:wrap; align-items:center; }
.card-tag { padding:3px 10px; background:rgba(25,137,250,0.08); color:#1989fa; font-size:12px; border-radius:6px; }
.card-meta { gap:14px; }
.card-price { font-size:13px; color:#ff6b35; font-weight:500; }
.card-loc,.card-date { font-size:12px; color:#999; }

.card-foot { display:flex; align-items:center; gap:16px; }
.card-stat { font-size:12px; color:#999; }
.card-spacer { flex:1; }
.card-bid-btn {
  padding:5px 16px; background:#1989fa; color:#fff;
  border-radius:14px; font-size:12px; font-weight:500;
}

.card-empty { text-align:center; padding:30px 0; font-size:14px; color:#999; }
</style>
