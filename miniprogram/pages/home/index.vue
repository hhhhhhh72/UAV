<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 1. 蓝顶栏 + 搜索 -->
      <view class="top-bar" :style="{ paddingTop: (statusBarH || 24) + 'px' }">
        <text class="top-loc" @tap="showCityPicker=true">{{ city }} ▼</text>
        <view class="top-search" @tap="handleSearchClick">
          <text class="search-hint">大家都在搜：吊运项目</text>
          <text class="search-btn">搜索</text>
        </view>
      </view>

      <!-- 2. 轮播图 -->
      <view class="banner-box">
        <swiper v-if="banners.length" autoplay interval="5000" circular :current="activeBanner" @change="onBannerChange">
          <swiper-item v-for="(b, i) in banners" :key="i"><image :src="b.image_url || b.image" mode="aspectFill" class="banner-img" /></swiper-item>
        </swiper>
        <image v-else src="/static/home-bg.jpg" mode="aspectFill" class="banner-img" />
        <view class="banner-dots">
          <view v-for="(b, i) in banners" :key="i" class="bd" :class="{ on: i === activeBanner }" />
        </view>
      </view>

      <!-- 3. 数据条 -->
      <view class="stats-bar" v-if="stats.demands > 0 || stats.users > 0">
        <text><text class="stats-k">需求</text> {{ stats.demands || 0 }} 单 ｜ <text class="stats-k">商家</text> {{ stats.shops || 0 }} 家 ｜ <text class="stats-k">用户</text> {{ stats.users || 0 }} 人</text>
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
        <view class="vouch v-orange" @tap="navigateTo('/pages/challenges/list')">
          <text class="vt">研发难题</text><text class="vs">难题广场</text>
        </view>
      </view>

      <!-- 7. 入驻商家 -->
      <view class="shop-panel" v-if="shops.length">
        <view class="shop-head"><text class="shop-title">入驻商家</text><text class="shop-more" @tap="navigateTo('/pages/shops/index')">全部 ></text></view>
        <scroll-view scroll-x :show-scrollbar="false">
          <view class="shop-list">
            <view v-for="s in shops" :key="s.id" class="shop-item" @tap="navigateTo('/pages/services/detail?id=' + s.id)">
              <image :src="s.logo_url || '/static/home-bg.jpg'" mode="aspectFill" class="shop-img" />
              <text class="shop-name">{{ s.name || '商家' }}</text>
              <text class="shop-desc">{{ s.description || '' }}</text>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 飞手任务 -->
      <view class="tieba-panel">
        <view class="tieba-head">
          <text class="tieba-title">飞手任务</text>
          <text class="tieba-more" @tap="navigateTo('/pages/tasks/index')">更多 ›</text>
        </view>
        <scroll-view scroll-x :show-scrollbar="false" class="tieba-nav">
          <text v-for="c in taskCats" :key="c.id" class="tieba-cat" :class="{on:activeTaskCat===c.id}" @tap="activeTaskCat=c.id">{{ c.name }}</text>
        </scroll-view>
        <view v-if="!taskList.length" class="tieba-empty">暂无需求，去发布大厅看看</view>
        <view v-for="(t,i) in taskList" :key="i" class="tieba-post" @tap="navigateTo('/pages/tasks/detail?id=' + t.id)">
          <view class="post-top">
            <text class="post-title">{{ t.title || '无人机需求' }}</text>
            <text class="post-detail-btn">详情</text>
          </view>
          <view class="post-tags">
            <text class="post-tag"><text class="tag-key">货物类型</text>{{ t.biz_type || t.cargo_type || '树/木头' }}</text>
            <text class="post-tag"><text class="tag-key">项目总量</text>{{ t.quantity || t.budget || '300吨' }}</text>
            <text class="post-tag"><text class="tag-key">启动时间</text>{{ t.start_time || fmtTime(t.created_at) || '—' }}</text>
          </view>
          <text class="post-desc">{{ (t.description || '暂无详细描述').slice(0, 120) }}</text>
          <view class="post-imgs">
            <image v-if="t.images && t.images.length" v-for="(img,ix) in t.images.slice(0,2)" :key="ix" :src="img" mode="aspectFill" class="img-fill" />
            <image v-if="!t.images || !t.images.length" src="/static/home-bg.jpg" mode="aspectFill" class="img-fill" />
            <image v-if="!t.images || t.images.length<2" src="/static/home-bg.jpg" mode="aspectFill" class="img-fill" />
          </view>
          <view class="post-loc"><u-icon name="location" size="26rpx" color="#0A66C2" />{{ t.district || t.location || '重庆' }}</view>
          <view class="post-foot">
            <text>{{ t.views || 8575 }}浏览 · {{ fmtTime(t.created_at) || '3天前' }}</text>
            <text class="foot-dots">⋯</text>
          </view>
        </view>
      </view>
    </view>
  </Layout>

  <!-- 城市选择 -->
  <u-popup :show="showCityPicker" position="bottom" round @close="showCityPicker=false">
    <scroll-view scroll-y style="height:75vh;padding:24px 16px 36px;box-sizing:border-box">
      <text class="city-title">选择区域</text>
      <view class="city-grid">
        <view v-for="d in allDistricts" :key="d" class="city-cell" :class="{on:city===d}" @tap="pickCity(d)">{{ d }}</view>
      </view>
    </scroll-view>
  </u-popup>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Layout from '@/components/Layout.vue'
import { safeNavigateTo } from '../../utils/nav'
import { request } from '../../utils/request'

const statusBarH = ref(24)
const banners = ref([])
const activeBanner = ref(0)
const notices = ref([])
const shops = ref([])
const stats = ref({ demands: 0, shops: 0, users: 0, views: 0, devices: 0, pilots: 0, tasks: 0 })
const taskCats = ref([
  { id: '', name: '全部' }, { id: '吊运', name: '吊运' }, { id: '航拍', name: '航拍' },
  { id: '植保', name: '植保' }, { id: '巡检', name: '巡检' }, { id: '测绘', name: '测绘' },
  { id: '租赁', name: '租赁' }, { id: '培训', name: '培训' },
])
const activeTaskCat = ref('')
const taskList = ref([])

const fmtTime = (s) => {
  try { const d = new Date(s); return (d.getMonth()+1)+'-'+d.getDate() } catch { return '' }
}

const functions = ref([
  { name: '需求大厅', icon: '/static/icons/apps.svg', bg: 'linear-gradient(135deg,#e3f2fd,#90caf9)', path: '/pages/tasks/index' },
  { name: '买卖租赁', icon: '/static/icons/rent.svg', bg: 'linear-gradient(135deg,#fce4ec,#f48fb1)', path: '/pages/demands/list' },
  { name: '考证培训', icon: '/static/icons/training-v2.svg', bg: 'linear-gradient(135deg,#e8f5e9,#a5d6a7)', path: '/pages/training/courses' },
  { name: '课题攻关', icon: '/static/icons/flight.svg', bg: 'linear-gradient(135deg,#e3f2fd,#42a5f5)', path: '/pages/projects/list' },
  { name: '商家入驻', icon: '/static/icons/shop.svg', bg: 'linear-gradient(135deg,#fff8e1,#fff176)', path: '/pages/enterprise/register' },
  { name: '求职招聘', icon: '/static/icons/service.svg', bg: 'linear-gradient(135deg,#e0f2f1,#80cbc4)', path: '/pages/jobs/list' },
  { name: '成果转化', icon: '/static/icons/study-fpv.svg', bg: 'linear-gradient(135deg,#e8eaf6,#9fa8da)', path: '/pages/achievements/list' },
  { name: '认证飞手', icon: '/static/icons/fpv-racing.svg', bg: 'linear-gradient(135deg,#fbe9e7,#ff8a65)', path: '/pages/pilots/list' },
  { name: '政策法规', icon: '/static/icons/government.svg', bg: 'linear-gradient(135deg,#e0f7fa,#4dd0e1)', path: '/pages/policies/list' },
  { name: '更多功能', icon: '/static/icons/apps.svg', bg: 'linear-gradient(135deg,#f5f5f5,#bdbdbd)', path: '/pages/more/index' },
])

const products = ref([])
const formatPrice = (fen) => fen ? (fen / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : '0.00'
const productImage = (p) => {
  try { const arr = typeof p.images === 'string' ? JSON.parse(p.images) : p.images; if (Array.isArray(arr) && arr[0]) return arr[0] } catch {}
  return '/static/home-bg.jpg'
}
const loadHome = async () => {
  try {
    const params = city.value && city.value !== '全重庆' ? '?city=' + encodeURIComponent(city.value) : ''
    const res = await request({ url: '/api/v1/home' + params })
    const data = res.data || res
    if (data.banners?.length) banners.value = data.banners
    if (data.notices?.length) notices.value = data.notices.map(String)
    if (data.shops?.length) shops.value = data.shops
    if (data.products?.length) products.value = data.products
    if (data.stats) stats.value = data.stats
    if (data.latest_demands?.length) {
      taskList.value = data.latest_demands.slice(0, 10).map(d => ({
        id: d.id, publisher_name: d.publisher_name || d.publisher_id, biz_type: d.biz_type,
        title: d.title || '', description: d.description || '', district: d.district || '',
        replies: d.replies || 0, views: d.views || 8500, created_at: d.created_at,
        contact: d.contact || '',
        cargo_type: d.cargo_type || d.biz_type, quantity: d.quantity || '',
        start_time: d.start_time || '', location: d.district || '',
        images: []
      }))
    } else {
      taskList.value = [] // 无真实需求数据不展示假数据
    }
  } catch {
    taskList.value = []
  }
}
const formatViews = (v) => v >= 10000 ? (v / 10000).toFixed(1) + '万' : String(v)

onMounted(async () => {
  statusBarH.value = (uni.getSystemInfoSync().statusBarHeight || 24) + 6
  await loadHome()
})

const onBannerChange = (e) => { activeBanner.value = e.detail.current }
const handleSearchClick = () => safeNavigateTo('/pages/search/index')
const handleLocation = () => uni.showToast({ title: '城市选择开发中', icon: 'none' })

const showCityPicker = ref(false)
const city = ref('全重庆')

const allDistricts = ['全重庆','渝中区','江北区','南岸区','沙坪坝区','九龙坡区','大渡口区','北碚区','渝北区','巴南区','两江新区','高新区','涪陵区','长寿区','江津区','合川区','永川区','南川区','綦江区','大足区','璧山区','铜梁区','潼南区','荣昌区','开州区','梁平区','武隆区','万州区','黔江区','城口县','丰都县','垫江县','忠县','云阳县','奉节县','巫山县','巫溪县','石柱县','秀山县','酉阳县','彭水县']
const pickCity = (d) => { city.value = d; showCityPicker.value = false; loadHome() }

const handleFunc = (f) => safeNavigateTo(f.path)
const navigateTo = (p) => safeNavigateTo(p)
</script>

<style scoped>
.home-page { min-height: 100vh; background: #f2f5f7; padding-bottom: 20px; }

.top-bar { display: flex; align-items: center; gap: 10px; padding: 0 14px 10px; background: var(--color-primary); }
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
.stats-k { color: var(--color-primary); font-weight: 600; }
.stats-help { color: var(--color-primary); font-weight: 500; }

.func-panel { background: #fff; padding: 14px 10px; }
.func-grid { display: grid; grid-template-columns: repeat(5, 1fr); row-gap: 16px; text-align: center; }
.func-item { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.func-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.func-name { font-size: 11px; color: #333; }

.notice-bar { display: flex; align-items: center; gap: 8px; margin: 8px 12px; background: #fff; border-radius: 10px; padding: 8px 14px; }
.notice-tag { font-size: 13px; color: var(--color-warning); font-weight: 600; }
.notice-text { font-size: 12px; color: #666; line-height: 22px; }

.vouch-row { display: flex; gap: 8px; margin: 6px 12px; }
.vouch { flex: 1; padding: 10px; border-radius: 10px; display: flex; flex-direction: column; gap: 2px; }
.v-green { background: linear-gradient(135deg,#e8f8ee,#d4f2e2); }
.v-orange { background: linear-gradient(135deg,#fff3e8,#ffe8d4); }
.vt { font-size: 15px; font-weight: 600; color: #333; }
.vs { font-size: 11px; color: #999; }

.shop-panel { margin: 6px 12px; background: #fff; border-radius: 10px; padding: 8px 12px; }
.shop-head { display: flex; justify-content: space-between; margin-bottom: 6px; }
.shop-title { font-size: 14px; font-weight: 600; }
.shop-more { font-size: 12px; color: #999; }
.shop-list { display: flex; gap: 10px; white-space: nowrap; }
.shop-item { width: 72px; text-align: center; }
.shop-img { width: 56px; height: 56px; border-radius: 10px; margin: 0 auto 6px; background: var(--color-primary-light); display: block; }
.shop-name { font-size: 12px; font-weight: 500; color: #333; display: block; }
.shop-desc { font-size: 10px; color: #999; }

/* 贴吧风格任务 */
.tieba-panel { margin: 6px 0; background: #fff; border-radius: 0; padding: 10px 14px 0; }
.tieba-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.tieba-title { font-size: 16px; font-weight: 700; color: var(--color-text); }
.tieba-more { font-size: 12px; color: #999; }
.tieba-nav { display: flex; gap: 0; padding-bottom: 8px; border-bottom: 1px solid #f0f0f0; white-space: nowrap; }
.tieba-cat { font-size: 13px; color: #666; padding: 4px 12px; flex-shrink: 0; }
.tieba-cat.on { color: var(--color-primary); font-weight: 600; border-bottom: 2px solid var(--color-primary); }
.tieba-post { padding: 12px 0; border-bottom: 1px solid #f5f5f5; }
.tieba-post:last-child { border: none; padding-bottom: 4px; }
.post-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.post-title { font-size: 14px; font-weight: 600; color: var(--color-text); flex: 1; padding-right: 6px; }
.post-detail-btn { font-size: 11px; color: var(--color-primary); padding: 2px 8px; border: 1px solid var(--color-primary); border-radius: 3px; flex-shrink: 0; }
.post-tags { display: flex; flex-wrap: wrap; gap: 6px; margin: 6px 0; }
.post-tag { font-size: 11px; background: #fff8e1; color: var(--color-warning); padding: 2px 8px; border-radius: 3px; font-weight: 500; }
.tag-key { color: var(--color-text); font-weight: 600; margin-right: 2px; }
.post-desc { font-size: 13px; color: #666; line-height: 1.5; display: block; margin: 6px 0; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
.post-imgs { display: flex; gap: 6px; margin: 6px 0; }
.img-fill { width: 80px; height: 80px; border-radius: 4px; background: #f5f5f5; display: block; }
.post-loc { font-size: 12px; color: var(--color-primary); display: flex; align-items: center; gap: 3px; margin: 2px 0 4px; }
.post-foot { display: flex; justify-content: space-between; align-items: center; font-size: 11px; color: #999; }
.foot-dots { font-size: 18px; color: #ccc; letter-spacing: -2px; }

.city-title { font-size: 18px; font-weight: 700; text-align: center; margin-bottom: 20px; color: var(--color-text); }
.city-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.city-cell { padding: 8px 18px; background: #f5f5f7; border-radius: 20px; font-size: 13px; color: #333; text-align: center; }
.city-cell.on { background: var(--color-primary); color: #fff; font-weight: 600; }
</style>