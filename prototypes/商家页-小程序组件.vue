<!--
  商家页面 - 无人机同城好店
  替换路径: miniprogram/pages/cases/index.vue 或新增 pages/shops/index.vue
  设计对齐：城市+店铺搜索 / Hero 双卡片 / 5宫格 / 商家头条 / 3Tab 筛选 / 商家卡片
-->
<template>
  <view class="shops-page">
    <!-- 顶部导航 -->
    <view class="top-nav">
      <view class="home-btn" @tap="onHome">⌂</view>
      <view class="nav-title">无人机同城好店 - 小飞虾</view>
      <view class="nav-actions">
        <text class="nav-act" @tap="onMore">⋯</text>
        <text class="nav-act" @tap="onMinimize">⊖</text>
      </view>
    </view>

    <!-- 城市 + 搜索 -->
    <view class="search-row">
      <view class="city-pill" @tap="onCityTap">
        <text>{{ currentCity }}</text>
        <text class="arrow">▼</text>
      </view>
      <view class="search-bar" @tap="onSearch">
        <text class="shop-select">店铺</text>
        <text class="sep">|</text>
        <text class="ph">搜店铺</text>
        <text class="ico">🔍</text>
      </view>
    </view>

    <!-- Hero Banner (双卡片) -->
    <view class="hero-banner-row">
      <view class="hero-banner" @tap="onBannerTap('main')">
        <text class="hero-drone">🚁</text>
        <view class="hero-text">
          <text class="ht-h">无人机同城好店</text>
          <text class="ht-p">找好店 · 选好货</text>
        </view>
      </view>
      <view class="hero-slide2" @tap="onBannerTap('caac')">
        <view>
          <text class="hs-h">CAAC 培训</text>
          <text class="hs-p">考证一站式服务</text>
        </view>
      </view>
    </view>

    <!-- 5图标分类 -->
    <view class="cat-row">
      <view v-for="cat in categoryIcons" :key="cat.id" class="cat-item" @tap="onCategoryTap(cat)">
        <view class="cat-icon" :style="{ background: cat.color }">{{ cat.emoji }}</view>
        <text class="cat-label">{{ cat.label }}</text>
      </view>
    </view>

    <!-- 商家头条 -->
    <view class="news-strip" @tap="onNewsTap">
      <text class="news-label">📰 商家头条</text>
      <text class="news-text">四川翱翔智控技术有限责任公司 <text style="color:#969799">...</text></text>
      <text class="news-cta">立即入驻</text>
    </view>

    <!-- Tab 筛选 -->
    <view class="filter-tabs">
      <view v-for="(t, idx) in filterTabs" :key="t.key" :class="['filter-tab', { active: activeFilter === t.key }]" @tap="switchFilter(t.key)">
        {{ t.label }}<text class="sub">{{ t.sub }}</text>
      </view>
    </view>

    <!-- 商家列表 -->
    <view class="shop-list" v-if="loading && shopList.length === 0">
      <view class="loading-state"><text>加载中...</text></view>
    </view>

    <view v-else-if="!loading && shopList.length === 0" class="empty-state">
      <text>暂无商家</text>
    </view>

    <view v-else class="shop-list">
      <view v-for="shop in shopList" :key="shop.id" class="shop-card" @tap="onShopTap(shop)">
        <view class="shop-logo" :style="{ background: shop.logo_bg }">
          <text v-if="shop.logo_text">{{ shop.logo_text }}</text>
          <text v-else>{{ shop.name.charAt(0) }}</text>
        </view>

        <view class="shop-body">
          <!-- 名称 + VIP/认证 -->
          <view class="shop-name-row">
            <view class="shop-name">
              {{ shop.name }}
              <text v-if="shop.liked" class="heart">💛</text>
              <text v-else-if="shop.medal === 'silver'" class="heart">🤍</text>
            </view>
            <text v-if="shop.vip" class="shop-tag-vip">{{ shop.vip === 'gold' ? 'VIP' : shop.vip }}</text>
          </view>

          <!-- 营业时间 + 浏览数 -->
          <view class="shop-meta-row">
            <text class="shop-hours">{{ shop.hours }}</text>
            <text class="views"><text class="v-num">{{ formatNum(shop.views) }}</text>浏览</text>
          </view>

          <!-- 标签 -->
          <view v-if="shop.tags && shop.tags.length" class="shop-tags">
            <text v-for="tag in shop.tags" :key="tag.text" :class="['shop-tag', tag.cls]">{{ tag.text }}</text>
          </view>

          <!-- 地址 -->
          <view v-if="shop.address" class="shop-loc">📍 {{ shop.address }}</view>

          <!-- 额外信息 -->
          <view v-if="shop.extra" class="shop-extra">
            <text class="ico">{{ shop.extra.icon || '🔋' }}</text>
            <text>{{ shop.extra.text }}</text>
          </view>
        </view>

        <view class="shop-call" @tap.stop="onCallTap(shop)">📞</view>
      </view>

      <view v-if="hasMore" class="load-more" @tap="loadMore"><text>加载更多 ▼</text></view>
      <view v-else class="no-more"><text>— 已加载全部 —</text></view>
    </view>

    <view style="height: 40rpx;" />
  </view>

  <!-- 浮动按钮 -->
  <view class="float-fab" @tap="onFloatFab">
    <text class="fab-i">📋</text>
    <text class="fab-l">入驻</text>
  </view>
  <view class="float-share" @tap="onFloatShare">
    <text class="fs-i">↗️</text>
    <text class="fs-l">分享</text>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { onReachBottom } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'

// 顶部
const currentCity = ref('全国')
const onHome = () => safeSwitchTab('/pages/home/index')
const onCityTap = () => { /* 城市选择 action sheet */ uni.showActionSheet({ itemList: ['全国','重庆市','北京市','成都市'] }) }
const onSearch = () => safeNavigateTo('/pages/search/index?type=shop')
const onMore = () => uni.showActionSheet({ itemList: ['邀请好友','意见反馈','客服中心'] })
const onMinimize = () => uni.navigateBack()

// Hero
const onBannerTap = (type) => {
  if (type === 'main') safeNavigateTo('/pages/community/index')
  if (type === 'caac') safeNavigateTo('/pages/training/courses')
}

// 5 类目
const categoryIcons = ref([
  { id: 'training', label: '培训机构', emoji: '🎓', color: '#dbeafe', link: '/pages/training/courses' },
  { id: 'sales', label: '无人机销售', emoji: '🚁', color: '#fef3c7', link: '/pages/products/index' },
  { id: 'app', label: '无人机应用', emoji: '🌾', color: '#d1fae5', link: '/pages/demands/list' },
  { id: 'parts', label: '无人机配件', emoji: '🔋', color: '#ede9fe', link: '/pages/products/index?cat=parts' },
  { id: 'repair', label: '无人机维修', emoji: '🔧', color: '#fee2e2', link: '/pages/services/index?cat=repair' },
])
const onCategoryTap = (c) => { if (c.link) safeNavigateTo(c.link) }

// 头条
const onNewsTap = () => safeNavigateTo('/pages/community/shop-news')

// 筛选
const filterTabs = ref([
  { key: 'recommend', label: '推荐', sub: '为你推荐' },
  { key: 'newest', label: '新入', sub: '最新加入' },
  { key: 'nearby', label: '附近', sub: '附近店铺' },
])
const activeFilter = ref('recommend')
const switchFilter = (k) => { activeFilter.value = k; fetchList(true) }

// 数据
const shopList = ref([])
const loading = ref(false)
const hasMore = ref(true)
const pageNum = ref(1)

// MOCK 数据 (5 家店, 模拟原型)
const MOCK_SHOPS = [
  {
    id: 's1', name: '小飞虾无人机厂家（官方自营店）', liked: true,
    logo_bg: 'linear-gradient(135deg,#fbbf24,#f59e0b)', logo_text: '🦐',
    vip: 'gold', hours: '9:00-20:00', views: 6165,
    tags: [
      { text: '无人机生产' }, { text: '无人机销售' }, { text: '无人机吊运' },
    ],
    address: '郑州市中原区亿达科技新城三期',
    extra: { icon: '🔋', text: '充电20分，续航40分' },
  },
  {
    id: 's2', name: '小飞虾无人机（重庆丰都店）', liked: false,
    logo_bg: 'linear-gradient(135deg,#fbbf24,#f59e0b)', logo_text: '🦐',
    hours: '9:00-18:00', views: 5155,
    tags: [
      { text: '无人机生产' }, { text: '无人机销售' }, { text: '无人机吊运' },
    ],
    address: '重庆市丰都县连新路42号',
  },
  {
    id: 's3', name: '中科未来飞行科技(承德)有限公司', liked: true,
    logo_bg: 'linear-gradient(135deg,#1e40af,#3b82f6)', logo_text: '🚀',
    hours: '9:00', views: 849,
    tags: [
      { text: '培训机构' }, { text: 'CAAC执照', cls: 'tag-blue' }, { text: '无人机驾照' }, { text: '无人机巡检' },
    ],
    address: '承德市双桥区冯营子镇大刘线',
  },
  {
    id: 's4', name: '四川翱翔智控技术有限责任公司', medal: 'silver',
    logo_bg: 'linear-gradient(135deg,#1e3a8a,#1e40af)', logo_text: '🛩️',
    hours: '0:00-24:00', views: 750,
    tags: [
      { text: '无人机应用' }, { text: '吊装运输' },
    ],
    address: '成都市高新区天府软件园',
  },
  {
    id: 's5', name: '鹰眼航空无人机服务', liked: true,
    logo_bg: 'linear-gradient(135deg,#8b5cf6,#7c3aed)', logo_text: '鹰',
    vip: '认证', hours: '8:00-22:00', views: 3210,
    tags: [
      { text: '航拍服务' }, { text: '测绘建模' },
    ],
    address: '重庆市江北区五里店',
  },
]

const fetchList = async (reset) => {
  if (reset) { pageNum.value = 1; hasMore.value = true; shopList.value = [] }
  loading.value = true
  try {
    const res = await request({
      url: '/api/v1/shops',
      data: { tab: activeFilter.value, page: pageNum.value, page_size: 10, city: currentCity.value },
    })
    const data = Array.isArray(res) ? res : (res?.data || res || {})
    const items = Array.isArray(data) ? data : (data.items || [])
    const total = (data.total != null) ? data.total : items.length

    const newItems = items.length > 0 ? items.map(adaptShop) : (pageNum.value === 1 ? MOCK_SHOPS.map(s => ({ ...s })) : [])

    if (reset) shopList.value = newItems
    else shopList.value = shopList.value.concat(newItems)

    hasMore.value = shopList.value.length < total || (items.length === 10)
  } catch (e) {
    if (reset) shopList.value = MOCK_SHOPS.map(s => ({ ...s }))
  } finally { loading.value = false }
}
const adaptShop = (raw) => ({
  id: raw.id, name: raw.name, liked: raw.liked || false, vip: raw.vip,
  logo_bg: raw.logo_bg || 'linear-gradient(135deg,#1da1f2,#0ea5e9)',
  logo_text: raw.logo_text || raw.name?.charAt(0),
  hours: raw.hours || '9:00-18:00',
  views: raw.views || 0,
  tags: raw.tags || [],
  address: raw.address,
  extra: raw.extra,
})

// 工具：格式化数字 (12345 → "1.2万")
const formatNum = (n) => {
  if (!n) return '0'
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return n.toString()
}

onMounted(() => fetchList(true))
onReachBottom(() => { if (!loading.value && hasMore.value) { pageNum.value++; fetchList(false) } })

// 交互
const onShopTap = (shop) => safeNavigateTo(`/pages/services/detail?id=${shop.id}&type=shop`)
const onCallTap = (shop) => uni.showModal({ title: '联系商家', content: '即将调用电话功能', confirmText: '拨打' })
const onFloatFab = () => safeNavigateTo('/pages/enterprise/register?type=shop')
const onFloatShare = () => uni.showActionSheet({ itemList: ['分享店铺页','生成分享海报','复制链接'] })
</script>

<style scoped>
.shops-page{min-height:100vh;background:#f5f6f8;padding-bottom:env(safe-area-inset-bottom)}

/* 顶部导航 */
.top-nav{
  display:flex;align-items:center;padding:16rpx 24rpx;background:#fff;gap:20rpx
}
.home-btn{font-size:36rpx;padding:8rpx 16rpx;color:#1a1a1a}
.nav-title{
  flex:1;text-align:center;font-size:30rpx;font-weight:600;
  overflow:hidden;text-overflow:ellipsis;white-space:nowrap
}
.nav-actions{display:flex;gap:8rpx}
.nav-act{padding:8rpx 16rpx;font-size:34rpx;color:#1a1a1a}

/* 城市 + 搜索 */
.search-row{
  display:flex;align-items:center;gap:16rpx;padding:16rpx 28rpx;background:#fff
}
.city-pill{
  display:flex;align-items:center;gap:8rpx;font-size:28rpx;font-weight:600;
  padding:10rpx 0;flex-shrink:0
}
.city-pill .arrow{font-size:18rpx;color:#969799}
.search-bar{
  flex:1;display:flex;align-items:center;
  background:#fff;border:2rpx solid #ebedf0;border-radius:999rpx;
  padding:12rpx 24rpx;font-size:24rpx;gap:12rpx
}
.shop-select{
  display:flex;align-items:center;gap:4rpx;font-size:24rpx;
  color:#646566;font-weight:600;flex-shrink:0
}
.shop-select::after{
  content:"";display:inline-block;width:0;height:0;border:8rpx solid transparent;
  border-top-color:#969799;transform:translateY(4rpx);margin-left:8rpx
}
.sep{color:#ebedf0;flex-shrink:0;font-size:28rpx}
.ph{flex:1;color:#b4b6b8}
.ico{font-size:28rpx;opacity:.5}

/* Hero Banner */
.hero-banner-row{
  display:flex;gap:12rpx;padding:0 24rpx
}
.hero-banner{
  flex:1.4;margin-top:20rpx;border-radius:28rpx;overflow:hidden;
  height:260rpx;background:linear-gradient(135deg,#dbeafe 0%,#e0f2fe 100%);
  position:relative
}
.hero-drone{
  position:absolute;left:40rpx;top:50%;transform:translateY(-50%);
  font-size:120rpx;z-index:2
}
.hero-text{position:absolute;right:32rpx;top:50%;transform:translateY(-50%);text-align:right}
.ht-h{font-size:30rpx;font-weight:700;color:#1da1f2;display:block}
.ht-p{font-size:20rpx;color:#1989fa;opacity:.8;display:block;margin-top:4rpx}

.hero-slide2{
  flex:1;margin-top:20rpx;height:260rpx;
  background:linear-gradient(135deg,#1da1f2,#3b82f6);
  border-radius:28rpx;color:#fff;display:flex;align-items:center;padding:0 28rpx
}
.hs-h{font-size:26rpx;font-weight:700;display:block}
.hs-p{font-size:18rpx;opacity:.85;margin-top:4rpx;display:block}

/* 5宫格 */
.cat-row{
  display:grid;grid-template-columns:repeat(5,1fr);gap:8rpx;
  margin:20rpx 24rpx 0;padding:28rpx 16rpx;background:#fff;border-radius:28rpx
}
.cat-item{display:flex;flex-direction:column;align-items:center;padding:12rpx 0}
.cat-icon{
  width:80rpx;height:80rpx;border-radius:24rpx;
  display:flex;align-items:center;justify-content:center;font-size:44rpx;margin-bottom:8rpx
}
.cat-label{font-size:22rpx;color:#1a1a1a;font-weight:500}

/* 商家头条 */
.news-strip{
  display:flex;align-items:center;gap:16rpx;margin:20rpx 24rpx 0;
  padding:20rpx 28rpx;background:#fff;border-radius:24rpx
}
.news-label{
  display:flex;align-items:center;gap:6rpx;
  font-size:28rpx;font-weight:700;color:#1a1a1a;flex-shrink:0
}
.news-text{
  flex:1;font-size:22rpx;color:#646566;
  overflow:hidden;text-overflow:ellipsis;white-space:nowrap
}
.news-cta{
  background:linear-gradient(135deg,#4a96f0,#1572cc);color:#fff;
  font-size:22rpx;font-weight:700;padding:10rpx 24rpx;border-radius:999rpx;
  box-shadow:0 2rpx 12rpx rgba(25,137,250,.25);flex-shrink:0
}

/* 筛选Tab */
.filter-tabs{
  display:grid;grid-template-columns:repeat(3,1fr);background:#fff;
  border-bottom:1px solid #ebedf0
}
.filter-tab{
  text-align:center;padding:28rpx 0;font-size:26rpx;color:#646566;
  font-weight:500;position:relative
}
.filter-tab.active{color:#1989fa;font-weight:700}
.filter-tab.active::after{
  content:"";position:absolute;bottom:0;left:50%;transform:translateX(-50%);
  width:48rpx;height:6rpx;border-radius:4rpx;background:#1989fa
}
.filter-tab .sub{display:block;font-size:20rpx;color:#c8c9cc;font-weight:400;margin-top:2rpx}
.filter-tab.active .sub{color:#1989fa}

/* 商家列表 */
.shop-list{padding:16rpx 0}
.shop-card{
  margin:16rpx 24rpx;background:#fff;border-radius:28rpx;
  padding:24rpx;display:flex;gap:24rpx;
  box-shadow:0 2rpx 8rpx rgba(0,0,0,.04)
}
.shop-logo{
  width:120rpx;height:120rpx;border-radius:24rpx;flex-shrink:0;
  display:flex;align-items:center;justify-content:center;
  font-weight:800;color:#fff;font-size:30rpx
}
.shop-body{flex:1;min-width:0}
.shop-name-row{display:flex;align-items:center;gap:12rpx;margin-bottom:12rpx}
.shop-name{
  font-size:28rpx;font-weight:700;color:#1a1a1a;
  overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1;min-width:0
}
.shop-name .heart{font-size:26rpx}
.shop-tag-vip{
  background:#f59e0b;color:#fff;font-size:18rpx;font-weight:700;
  padding:4rpx 12rpx;border-radius:8rpx;flex-shrink:0
}

.shop-meta-row{
  display:flex;align-items:center;gap:12rpx;margin-bottom:12rpx;font-size:22rpx
}
.shop-hours{
  display:inline-flex;align-items:center;gap:6rpx;background:#fff5e6;
  color:#f97316;font-size:22rpx;font-weight:600;padding:6rpx 14rpx;border-radius:10rpx
}
.shop-hours::before{content:"🕐";font-size:18rpx}
.shop-meta-row .views{margin-left:auto;color:#969799;font-size:22rpx}
.shop-meta-row .views .v-num{color:#1a1a1a;font-weight:700;font-size:24rpx}

.shop-tags{display:flex;flex-wrap:wrap;gap:8rpx;margin-bottom:12rpx}
.shop-tag{
  font-size:20rpx;padding:4rpx 14rpx;border-radius:8rpx;font-weight:500;
  background:#f5f6f8;color:#646566
}
.shop-tag.tag-blue{background:#dbeafe;color:#1e40af}

.shop-loc{
  font-size:22rpx;color:#969799;display:flex;align-items:center;gap:6rpx;
  overflow:hidden;text-overflow:ellipsis;white-space:nowrap;margin-bottom:8rpx
}
.shop-extra{
  display:inline-flex;align-items:center;gap:6rpx;
  font-size:22rpx;color:#646566;background:#fff5e6;
  padding:8rpx 16rpx;border-radius:12rpx;margin-top:8rpx
}
.shop-extra .ico{font-size:24rpx}

.shop-call{
  width:76rpx;height:76rpx;border-radius:50%;
  background:#fff;border:2rpx solid #1989fa;color:#1989fa;
  display:flex;align-items:center;justify-content:center;font-size:32rpx;
  flex-shrink:0;align-self:flex-start
}
.shop-call:active{background:#1989fa;color:#fff}

/* 状态 */
.loading-state,.empty-state{
  text-align:center;padding:80rpx 0;color:#969799;font-size:24rpx
}
.load-more,.no-more{text-align:center;padding:24rpx 0;color:#c8c9cc;font-size:22rpx}
.load-more:active{color:#1989fa}

/* 浮动按钮 */
.float-fab{
  position:fixed;bottom:200rpx;right:50%;margin-right:-680rpx;
  width:108rpx;height:108rpx;border-radius:50%;
  background:linear-gradient(135deg,#4a96f0,#1572cc);color:#fff;
  display:flex;align-items:center;justify-content:center;flex-direction:column;
  box-shadow:0 8rpx 24rpx rgba(25,137,250,.4);z-index:10;line-height:1.1
}
.fab-i{font-size:30rpx;display:block}
.fab-l{font-size:18rpx;margin-top:4rpx;font-weight:600}

.float-share{
  position:fixed;bottom:260rpx;right:50%;margin-right:-560rpx;
  width:92rpx;height:92rpx;border-radius:50%;
  background:#f97316;color:#fff;
  display:flex;align-items:center;justify-content:center;flex-direction:column;
  box-shadow:0 4rpx 16rpx rgba(249,115,22,.35);z-index:10;line-height:1
}
.fs-i{font-size:32rpx;display:block}
.fs-l{font-size:18rpx;margin-top:2rpx;font-weight:700}
</style>