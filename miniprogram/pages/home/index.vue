<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 1. 蓝顶栏 + 搜索 -->
      <view class="top-bar" :style="{ paddingTop: (statusBarH || 24) + 'px' }">
        <text class="top-loc" @tap="handleLocation">全国 ▼</text>
        <view class="top-search" @tap="handleSearchClick">
          <text class="search-hint">大家都在搜：吊运项目</text>
          <text class="search-btn">搜索</text>
        </view>
      </view>

      <!-- 2. 轮播图 -->
      <view class="banner-box">
        <swiper autoplay interval="5000" circular :current="activeBanner" @change="onBannerChange">
          <swiper-item v-for="(b, i) in banners" :key="i"><image :src="b.image" mode="aspectFill" class="banner-img" /></swiper-item>
        </swiper>
        <view class="banner-dots">
          <view v-for="(b, i) in banners" :key="i" class="bd" :class="{ on: i === activeBanner }" />
        </view>
      </view>

      <!-- 3. 数据条 -->
      <view class="stats-bar">
        <text>📢 浏览：669万 ｜ 发布：848 ｜ 商家：105</text>
        <text class="stats-help">帮助</text>
      </view>

      <!-- 4. 金刚区 2×4 -->
      <view class="func-panel">
        <view class="func-grid">
          <view v-for="(f, i) in functions" :key="i" class="func-item" @tap="handleFunc(f)">
            <view class="func-icon" :style="{ background: f.bg }">
              <image :src="f.icon" mode="aspectFit" style="width:24px;height:24px" />
            </view>
            <text class="func-name">{{ f.name }}</text>
          </view>
        </view>
      </view>

      <!-- 5. 公告 -->
      <view class="notice-bar" v-if="notices.length">
        <text class="notice-tag">公告</text>
        <swiper vertical autoplay circular interval="3000" style="flex:1;height:22px">
          <swiper-item v-for="(n, i) in notices" :key="i"><text class="notice-text">{{ n }}</text></swiper-item>
        </swiper>
      </view>

      <!-- 6. 会员/合伙人 -->
      <view class="vouch-row">
        <view class="vouch v-green" @tap="navigateTo('/pages/mine/index')">
          <text class="vt">加入会员</text><text class="vs">更优惠</text>
        </view>
        <view class="vouch v-orange" @tap="navigateTo('/pages/enterprise/register')">
          <text class="vt">同城合伙人</text><text class="vs">加入合伙人</text>
        </view>
      </view>

      <!-- 7. 本地商家 -->
      <view class="shop-panel">
        <view class="shop-head"><text class="shop-title">本地商家</text><text class="shop-more" @tap="navigateTo('/pages/shops/index')">全部 ></text></view>
        <scroll-view scroll-x :show-scrollbar="false">
          <view class="shop-list">
            <view v-for="s in shops" :key="s.id" class="shop-item" @tap="navigateTo('/pages/services/detail?id=' + s.id)">
              <image :src="s.logo || '/static/home-bg.jpg'" mode="aspectFill" class="shop-img" />
              <text class="shop-name">{{ s.name }}</text>
              <text class="shop-desc">{{ s.desc }}</text>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 8. 需求信息流 -->
      <view class="demand-panel">
        <scroll-view scroll-x :show-scrollbar="false" class="demand-tabs">
          <text v-for="t in demandCats" :key="t.id" class="dt" :class="{ on: activeCat === t.id }" @tap="switchCat(t.id)">{{ t.name }}</text>
        </scroll-view>
        <view v-for="(d, i) in demandList" :key="i" class="d-card" @tap="goDemand(d.id)">
          <view class="d-head">
            <view class="d-ava">{{ d.userName?.[0] || '?' }}</view>
            <view class="d-user"><text class="d-name">{{ d.userName }}</text><text class="d-tag" v-if="d.tag">{{ d.tag }}</text></view>
            <view class="d-call" @tap.stop="callTo(d.phone)">📞</view>
          </view>
          <text class="d-title">{{ d.title }}</text>
          <text class="d-loc">📍 {{ d.location }}</text>
          <text class="d-desc">{{ d.desc }}</text>
          <view class="d-meta"><text>{{ d.views }}浏览</text><text>{{ d.time }}</text><text>♥ {{ d.likes }}</text></view>
        </view>
        <view v-if="!demandList.length" class="d-empty">暂无需求</view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import Layout from '@/components/Layout.vue'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'
import { request } from '../../utils/request'

const statusBarH = ref(24)
const banners = ref([{ image: '/static/home-bg.jpg' }])
const activeBanner = ref(0)
const notices = ref(['飞行须知：保持安全高度', '无人机登记政策已更新', '欢迎加入同城合伙人'])

const functions = ref([
  { name: '吊运服务', icon: '/static/icons/lifting.svg', bg: 'linear-gradient(135deg,#e3f2fd,#bbdefb)', path: '/pages/demands/list' },
  { name: '设备租赁', icon: '/static/icons/rent.svg', bg: 'linear-gradient(135deg,#fce4ec,#f8bbd0)', path: '/pages/demands/list' },
  { name: '培训考证', icon: '/static/icons/training-v2.svg', bg: 'linear-gradient(135deg,#e8f5e9,#c8e6c9)', path: '/pages/training/courses' },
  { name: '植保飞防', icon: '/static/icons/flight.svg', bg: 'linear-gradient(135deg,#fff3e0,#ffe0b2)', path: '/pages/demands/list' },
  { name: '赛事活动', icon: '/static/icons/competition.svg', bg: 'linear-gradient(135deg,#ede7f6,#d1c4e9)', path: '/pages/competitions/list' },
  { name: '维修保养', icon: '/static/icons/wrench.svg', bg: 'linear-gradient(135deg,#e0f2f1,#b2dfdb)', path: '/pages/demands/list' },
  { name: '商家入驻', icon: '/static/icons/shop.svg', bg: 'linear-gradient(135deg,#fff8e1,#fff9c4)', path: '/pages/enterprise/register' },
  { name: '金融服务', icon: '/static/icons/finance.svg', bg: 'linear-gradient(135deg,#f3e5f5,#e1bee7)', path: '/pages/demands/list' },
])

const shops = ref([
  { id: '1', name: '大疆授权店', logo: '', desc: '无人机销售维修' },
  { id: '2', name: '飞手之家', logo: '', desc: '航拍培训基地' },
  { id: '3', name: '天行植保', logo: '', desc: '农业植保服务' },
  { id: '4', name: '极飞科技', logo: '', desc: '智能农业方案' },
])

const demandCats = ref([
  { id: '', name: '最新信息' }, { id: 'lift', name: '吊运独家' },
  { id: 'trade', name: '买卖租赁' }, { id: 'training', name: '考证培训' }, { id: 'plant', name: '植保运输' },
])
const activeCat = ref('')
const demandList = ref([])
const loadDemands = async () => {
  try {
    const res = await request({ url: '/api/v1/demands', data: { biz_type: activeCat.value, page: 1, page_size: 10 } })
    const data = Array.isArray(res) ? res : (res.data || [])
    demandList.value = data.slice(0, 10).map(d => ({
      id: d.id, userName: d.publisher_name || '匿名', tag: d.biz_type || '', title: d.title || '',
      location: d.district || '', desc: (d.description || '').slice(0, 80),
      views: 0, likes: 0, time: d.created_at ? new Date(d.created_at).toLocaleDateString() : '', phone: d.contact || ''
    }))
  } catch { demandList.value = [] }
  if (!demandList.value.length) {
    demandList.value = [
      { id: '1', userName: '张飞行', tag: '吊运', title: '需要大疆T50运输化肥200亩', location: '山东省青岛市', desc: 'FC100型号3台，T10型号7台，共10台设备需运输到指定地点', views: 1592, likes: 12, time: '07-09 12:22', phone: '' },
      { id: '2', userName: '李航拍', tag: '航拍', title: '婚庆航拍需要飞手', location: '广东省广州市', desc: '下周六婚礼现场航拍，熟练飞手一名，设备自带Mavic3', views: 834, likes: 8, time: '07-08 15:30', phone: '' },
    ]
  }
}
const switchCat = (id) => { activeCat.value = id; loadDemands() }
const callTo = (p) => p && uni.makePhoneCall({ phoneNumber: p })
const goDemand = (id) => id && safeNavigateTo('/pages/demands/detail?id=' + id)

onMounted(async () => {
  statusBarH.value = (uni.getSystemInfoSync().statusBarHeight || 24) + 6
  try {
    const cfg = (await request({ url: '/api/services/config' }))._home || {}
    if (cfg.banners?.length) banners.value = cfg.banners.filter(b => b.image)
    if (cfg.notices?.length) notices.value = cfg.notices.filter(Boolean)
  } catch {}
  loadDemands()
})

const onBannerChange = (e) => { activeBanner.value = e.detail.current }
const handleSearchClick = () => safeSwitchTab('/pages/services/index')
const handleLocation = () => uni.showToast({ title: '城市选择开发中', icon: 'none' })
const handleFunc = (f) => safeNavigateTo(f.path)
const navigateTo = (p) => safeNavigateTo(p)
</script>

<style scoped>
.home-page { min-height: 100vh; background: #f2f5f7; padding-bottom: 20px; }

.top-bar { display: flex; align-items: center; gap: 10px; padding: 0 14px 10px; background: #1989fa; }
.top-loc { color: #fff; font-size: 15px; font-weight: 600; white-space: nowrap; }
.top-search { flex: 1; display: flex; align-items: center; justify-content: space-between; background: rgba(255,255,255,0.2); border-radius: 20px; padding: 8px 14px; }
.search-hint { font-size: 13px; color: rgba(255,255,255,0.6); }
.search-btn { font-size: 13px; color: #fff; font-weight: 500; }

.banner-box { position: relative; height: 150px; overflow: hidden; }
.banner-box swiper { width: 100%; height: 100%; }
.banner-img { width: 100%; height: 100%; display: block; }
.banner-dots { position: absolute; bottom: 8px; left: 0; right: 0; display: flex; justify-content: center; gap: 6px; }
.bd { width: 6px; height: 6px; border-radius: 50%; background: rgba(255,255,255,0.4); }
.bd.on { background: #fff; }

.stats-bar { display: flex; justify-content: space-between; padding: 10px 14px; background: #fff; font-size: 12px; color: #666; border-bottom: 1px solid #eee; }
.stats-help { color: #1989fa; font-weight: 500; }

.func-panel { background: #fff; padding: 16px 14px; }
.func-grid { display: grid; grid-template-columns: repeat(4, 1fr); row-gap: 18px; text-align: center; }
.func-item { display: flex; flex-direction: column; align-items: center; gap: 8px; }
.func-icon { width: 44px; height: 44px; border-radius: 14px; display: flex; align-items: center; justify-content: center; }
.func-name { font-size: 12px; color: #333; }

.notice-bar { display: flex; align-items: center; gap: 8px; margin: 8px 12px; background: #fff; border-radius: 10px; padding: 8px 14px; }
.notice-tag { font-size: 13px; color: #ff6b35; font-weight: 600; }
.notice-text { font-size: 12px; color: #666; line-height: 22px; }

.vouch-row { display: flex; gap: 10px; margin: 8px 12px; }
.vouch { flex: 1; padding: 14px; border-radius: 12px; display: flex; flex-direction: column; gap: 4px; }
.v-green { background: linear-gradient(135deg,#e8f8ee,#d4f2e2); }
.v-orange { background: linear-gradient(135deg,#fff3e8,#ffe8d4); }
.vt { font-size: 15px; font-weight: 600; color: #333; }
.vs { font-size: 11px; color: #999; }

.shop-panel { margin: 8px 12px; background: #fff; border-radius: 12px; padding: 16px; }
.shop-head { display: flex; justify-content: space-between; margin-bottom: 14px; }
.shop-title { font-size: 16px; font-weight: 600; }
.shop-more { font-size: 13px; color: #999; }
.shop-list { display: flex; gap: 12px; white-space: nowrap; }
.shop-item { width: 100px; text-align: center; }
.shop-img { width: 80px; height: 80px; border-radius: 12px; margin: 0 auto 8px; background: #e8f2fc; display: block; }
.shop-name { font-size: 13px; font-weight: 500; color: #333; display: block; }
.shop-desc { font-size: 11px; color: #999; }

.demand-panel { margin: 8px 12px; background: #fff; border-radius: 12px; padding: 14px; }
.demand-tabs { display: flex; gap: 16px; padding-bottom: 10px; border-bottom: 1px solid #eee; margin-bottom: 8px; white-space: nowrap; }
.dt { font-size: 14px; color: #666; flex-shrink: 0; }
.dt.on { color: #1989fa; font-weight: 600; }
.d-card { padding: 14px 0; border-bottom: 1px solid #f0f0f0; }
.d-card:last-child { border-bottom: none; padding-bottom: 4px; }
.d-head { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.d-ava { width: 36px; height: 36px; border-radius: 50%; background: #e8f2fc; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 600; color: #1989fa; }
.d-user { flex: 1; }
.d-name { font-size: 13px; font-weight: 500; }
.d-tag { font-size: 11px; color: #ff6b35; margin-left: 8px; }
.d-call { width: 48px; height: 28px; border-radius: 14px; background: #ff6b35; color: #fff; font-size: 12px; display: flex; align-items: center; justify-content: center; }
.d-title { font-size: 16px; font-weight: 600; color: #1a1a1a; display: block; margin-bottom: 6px; }
.d-loc { font-size: 12px; color: #1989fa; display: block; margin-bottom: 6px; }
.d-desc { font-size: 13px; color: #666; line-height: 1.5; display: block; margin-bottom: 10px; }
.d-meta { display: flex; gap: 12px; font-size: 11px; color: #999; }
.d-empty { text-align: center; padding: 30px 0; color: #999; font-size: 14px; }
</style>
