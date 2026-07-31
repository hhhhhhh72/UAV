<template>
<Layout :current="1">
  <view class="page">
    <!-- 1. 顶部轮播图 -->
    <swiper class="banner" autoplay circular interval="3500" indicator-dots indicator-color="rgba(255,255,255,.4)" indicator-active-color="#fff">
      <swiper-item v-for="(b,i) in banners" :key="i">
        <image :src="b.image_url || b.image" mode="aspectFill" class="banner-img" />
      </swiper-item>
    </swiper>
    <image v-if="!banners.length" src="/static/home-bg.jpg" mode="aspectFill" class="banner-img" />

    <!-- 2. 优惠活动 -->
    <view class="activity-row">
      <view v-for="a in activities" :key="a.id" class="act-item" @tap="onActivity(a)">
        <view class="act-icon" :style="{ background: a.bg }">{{ a.emoji }}</view>
        <text class="act-name">{{ a.name }}</text>
      </view>
    </view>

    <!-- 3. 店铺推荐 -->
    <view class="section">
      <view class="sec-head">
        <view class="sec-title">
          <text class="sec-zh">店铺推荐</text>
          <text class="sec-en">GOOD SHOP</text>
        </view>
        <text class="sec-more" @tap="goMoreShops">全部 ›</text>
      </view>
      <scroll-view scroll-x :show-scrollbar="false" class="shop-row">
        <view v-for="s in shopRecs" :key="s.id" class="shop-card" @tap="goShop(s.id)">
          <image :src="s.logo || '/static/home-bg.jpg'" mode="aspectFill" class="shop-logo" />
          <text class="shop-name">{{ s.name }}</text>
          <view class="shop-promo">
            <text class="promo-txt">0款商品</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 4. 分类栏 -->
    <view class="cat-tabs">
      <view v-for="c in cats" :key="c.key" class="cat-tab" :class="{on: activeCat===c.key}" @tap="activeCat=c.key">
        <text class="ct-name">{{ c.name }}</text>
        <text class="ct-sub">{{ c.sub }}</text>
      </view>
    </view>

    <!-- 5. 商品列表 -->
    <view class="grid">
      <view v-for="p in products" :key="p.id" class="card" @tap="goDetail(p.id)">
        <view class="img-wrap">
          <image :src="imgSrc(p)" mode="aspectFill" class="card-img" />
          <view v-if="p.condition" class="tag" :class="'tag-'+tagCls(p.condition)">{{ p.condition }}</view>
        </view>
        <view class="card-body">
          <text class="card-title">{{ p.title }}</text>
          <text class="card-desc" v-if="p.brand">{{ p.brand }} · {{ p.model }}</text>
          <text class="card-price">¥{{ fmt(p.price_fen) }}</text>
        </view>
      </view>
    </view>
  </view>
</Layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Layout from '@/components/Layout.vue'
import { request } from '@/utils/request'

const products = ref([])
const activeCat = ref('best')
const shopRecs = ref([
  { id: 1, name: '佛山市南海俊岸...' },
  { id: 2, name: '四川亿万星河科...' },
  { id: 3, name: '宁夏皇榜无人机...' },
  { id: 4, name: '毕节牧青...' },
])

const banners = ref([])

const activities = ref([
  { id:'rush',  name:'抢购',   emoji:'⚡', bg:'linear-gradient(135deg,#ff5252,#d32f2f)' },
  { id:'group', name:'拼团',   emoji:'👥', bg:'linear-gradient(135deg,#ec407a,#c2185b)' },
  { id:'coupon',name:'领券',   emoji:'🎟', bg:'linear-gradient(135deg,#ffb74d,#f57c00)' },
  { id:'bargain',name:'砍价',  emoji:'✂', bg:'linear-gradient(135deg,#66bb6a,#388e3c)' },
  { id:'super', name:'超级品牌',emoji:'👑', bg:'linear-gradient(135deg,#ab47bc,#7b1fa2)' },
])

const cats = ref([
  { key:'best',    name:'精选',     sub:'为您甄选' },
  { key:'train',   name:'培训机构', sub:'技能培训' },
  { key:'sales',   name:'产品销售', sub:'无人机' },
  { key:'service', name:'服务应用', sub:'业务服务' },
  { key:'parts',   name:'配件/零件', sub:'无人机' },
])

const onActivity = (a) => uni.showToast({ title: a.name + ' 活动', icon: 'none' })
const goMoreShops = () => uni.navigateTo({ url: '/pages/shops/index' })
const goShop = (id) => uni.navigateTo({ url: '/pages/shops/index' })

const loadProducts = async () => {
  try {
    const res = await request({ url: '/api/v1/home' })
    const data = res.data || res
    products.value = data.products || []
    if (data.banners?.length) banners.value = data.banners
    if (data.shops?.length) shopRecs.value = data.shops
  } catch {}
}
const goDetail = (id) => uni.navigateTo({ url: '/pages/mall/detail?id=' + id })
const imgSrc = (p) => {
  try { const arr = typeof p.images === 'string' ? JSON.parse(p.images) : p.images; if (arr[0]) return arr[0] } catch {}
  return '/static/home-bg.jpg'
}
const fmt = (f) => f ? (f / 100).toLocaleString('en-US', { minimumFractionDigits: 2 }) : '0.00'
const tagCls = (c) => { if (c.includes('官方')) return 'official'; if (c.includes('95')) return 'n95'; if (c.includes('98')) return 'n98'; return 'default' }

onMounted(loadProducts)
</script>

<style scoped>
.page { min-height: 100vh; background: #f2f5f7; padding-bottom: 60px; }

/* 轮播 */
.banner { width: 100%; height: 160px; }
.banner-img { width: 100%; height: 100%; display: block; }

/* 优惠活动 */
.activity-row { display: flex; background: #fff; padding: 16px 0; margin-bottom: 8px; }
.act-item { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 6px; }
.act-icon { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 20px; color: #fff; box-shadow: 0 2px 6px rgba(0,0,0,.08); }
.act-name { font-size: 12px; color: #333; }

/* 通用 section */
.section { background: #fff; margin-bottom: 8px; padding: 14px 12px 16px; }
.sec-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.sec-title { display: flex; flex-direction: column; align-items: center; flex: 1; }
.sec-zh { font-size: 16px; font-weight: 700; color: #1a1a1a; }
.sec-en { font-size: 10px; color: #999; letter-spacing: 1px; margin-top: 2px; }
.sec-more { font-size: 12px; color: #999; }

/* 店铺推荐 */
.shop-row { white-space: nowrap; }
.shop-card { display: inline-block; width: 88px; margin-right: 10px; text-align: center; }
.shop-logo { width: 72px; height: 72px; border-radius: 8px; background: #f0f0f0; }
.shop-name { font-size: 11px; color: #333; display: block; margin-top: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.shop-promo { display: inline-block; margin-top: 4px; padding: 1px 8px; border: 1px solid #ff3b30; border-radius: 10px; }
.promo-txt { font-size: 10px; color: #ff3b30; }

/* 分类栏 */
.cat-tabs { display: flex; background: #fff; border-top: 1px solid #f0f0f0; border-bottom: 1px solid #f0f0f0; margin-bottom: 8px; padding: 10px 0; }
.cat-tab { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 2px; border-right: 1px solid #f0f0f0; }
.cat-tab:last-child { border-right: none; }
.cat-tab.on .ct-name { color: #ff3b30; position: relative; }
.cat-tab.on .ct-name::before { content: ''; position: absolute; left: 50%; bottom: -8px; transform: translateX(-50%); width: 22px; height: 3px; background: #ff3b30; border-radius: 2px; }
.ct-name { font-size: 14px; font-weight: 600; color: #1a1a1a; }
.ct-sub { font-size: 10px; color: #999; }

/* 商品网格 */
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; padding: 0 10px; }
.card { background: #fff; border-radius: 10px; overflow: hidden; }
.img-wrap { position: relative; aspect-ratio: 1; background: #f7f8fa; }
.card-img { width: 100%; height: 100%; display: block; }
.tag { position: absolute; top: 6px; left: 6px; font-size: 10px; padding: 2px 6px; border-radius: 3px; color: #fff; font-weight: 600; }
.tag-official { background: #34c759; } .tag-n95 { background: #ff9500; } .tag-n98 { background: #ff3b30; } .tag-default { background: #8e8e93; }
.card-body { padding: 8px; }
.card-title { font-size: 13px; font-weight: 600; color: #1a1a1a; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; height: 34px; }
.card-desc { font-size: 11px; color: #999; display: block; margin: 2px 0; }
.card-price { font-size: 16px; font-weight: 700; color: #ff3b30; }
</style>
