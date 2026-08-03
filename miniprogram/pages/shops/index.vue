<template>
<Layout :current="3">
  <view class="page">
    <!-- 轮播图 -->
    <swiper class="banner" autoplay circular interval="3000" indicator-dots indicator-color="rgba(255,255,255,.4)" indicator-active-color="#fff">
      <swiper-item v-for="(b,i) in banners" :key="i">
        <view class="banner-slide" :style="{ background: b.bg }">
          <text class="banner-txt">{{ b.title }}</text>
        </view>
      </swiper-item>
    </swiper>

    <!-- 搜索栏 -->
    <view class="search-row">
      <text class="search-cat">店铺 ▼</text>
      <input class="search-input" placeholder="搜店铺" />
      <u-icon name="search" size="28rpx" color="#999999" />
    </view>

    <!-- 分类��标 -->
    <view class="cat-row">
      <view v-for="c in cats" :key="c.id" class="cat-item" @tap="pickCat(c.id)">
        <view class="cat-icon">{{ c.icon }}</view>
        <text class="cat-name">{{ c.name }}</text>
      </view>
    </view>

    <!-- 头条 -->
    <view class="news-bar">
      <text class="news-tag">南国头条</text>
      <text class="news-text" @tap="goJoin">昭通驰通机械城租赁入驻成功，立即入驻 ›</text>
    </view>

    <!-- 分类 Tab -->
    <view class="cat-tabs">
      <view class="tab-item" :class="{on:activeTab==='rec'}" @tap="activeTab='rec'">
        <text class="t1">推荐</text>
        <text class="t2">为您推荐</text>
      </view>
      <view class="tab-item" :class="{on:activeTab==='new'}" @tap="activeTab='new'">
        <text class="t1">新人</text>
        <text class="t2">最新加入</text>
      </view>
      <view class="tab-item" :class="{on:activeTab==='near'}" @tap="activeTab='near'">
        <text class="t1">附近</text>
        <text class="t2">附近店铺</text>
      </view>
    </view>

    <!-- 入驻 CTA -->
    <view class="join-cta" @tap="goJoin">入驻</view>

    <!-- 店铺卡片 -->
    <view class="shop-list">
      <view v-for="s in shops" :key="s.id" class="shop-card">
        <image :src="s.logo_url || '/static/home-bg.jpg'" mode="aspectFill" class="shop-logo" />
        <view class="shop-body">
          <view class="shop-row1">
            <text class="shop-name">{{ s.name }}</text>
            <text class="shop-badge">官方</text>
          </view>
          <view class="shop-row2">
            <text class="shop-hours">营业 9:00-20:00</text>
          </view>
          <view class="shop-tags">
            <text class="tag">无人机生产</text>
            <text class="tag">无人机销售</text>
            <text class="tag">无人机吊运</text>
          </view>
          <view class="shop-addr"><u-icon name="location" size="24rpx" color="#969799" /><text>{{ s.district || s.description || '郑州市中原区亿达科技城三期' }}</text></view>
          <view class="shop-row3">
            <text class="shop-bat">充电20分,续航40分</text>
          </view>
        </view>
        <view class="shop-right">
          <view class="shop-call" @tap.stop="callShop">电</view>
          <text class="shop-views">{{ s.views || 6274 }}浏览</text>
        </view>
      </view>
      <view v-if="!shops.length" class="empty">暂无商家</view>
    </view>
  </view>
</Layout>
</template>

<script setup>
import { ref } from 'vue'
import Layout from '@/components/Layout.vue'
import { request } from '@/utils/request'

const banners = ref([
  { bg: 'linear-gradient(135deg,#4fc3f7,#0288d1)', title: '无人机商家入驻' },
  { bg: 'linear-gradient(135deg,#81c784,#388e3c)', title: '优质商家推荐' },
  { bg: 'linear-gradient(135deg,#ff8a65,#e64a19)', title: '限时入驻优惠' },
])

const cats = ref([
  { id:'train', name:'培训机构', icon:'培' },
  { id:'sale', name:'无人机销售', icon:'销' },
  { id:'app', name:'无人机应用', icon:'用' },
  { id:'parts', name:'无人机配件', icon:'配' },
  { id:'repair', name:'无人机维修', icon:'修' },
])
const activeTab = ref('rec')
const shops = ref([])

const loadShops = async () => {
  try {
    const res = await request({ url:'/api/v1/home' })
    shops.value = (res.data||res).shops || []
  } catch {}
}
loadShops()

const pickCat = (id) => uni.showToast({ title: id, icon: 'none' })
const goJoin = () => uni.navigateTo({ url:'/pages/enterprise/register' })
const callShop = () => uni.showToast({ title:'拨号开发中', icon:'none' })
</script>

<style scoped>
.page { min-height:100vh; background:var(--color-bg); padding-bottom:80px; }

/* Banner */
.banner { height:180px; }
.banner-slide { width:100%; height:100%; display:flex; flex-direction:column; align-items:center; justify-content:center; }
.banner-txt { color:#fff; font-size:22px; font-weight:700; text-shadow:0 2px 4px rgba(0,0,0,.2); }

/* Search */
.search-row { display:flex; align-items:center; margin:12px 12px; padding:10px 14px; background:#fff; border-radius:8px; border:1px solid #e8e8e8; }
.search-cat { font-size:13px; color:#666; padding-right:8px; border-right:1px solid #eee; }
.search-input { flex:1; padding:0 8px; font-size:14px; }
.search-icon { font-size:14px; color:#999; }

/* Cats */
.cat-row { display:grid; grid-template-columns:repeat(5,1fr); margin:14px 0; background:#fff; padding:14px 12px; }
.cat-item { display:flex; flex-direction:column; align-items:center; gap:4px; }
.cat-icon { width:40px; height:40px; border-radius:50%; background:var(--color-primary-light); display:flex; align-items:center; justify-content:center; font-size:20px; color:var(--color-primary); font-weight:600; }
.cat-name { font-size:11px; color:#333; }

/* News */
.news-bar { display:flex; align-items:center; gap:8px; background:#fff; padding:8px 12px; margin-bottom:4px; }
.news-tag { font-size:11px; padding:2px 6px; background:#fff3e0; color:var(--color-warning); border-radius:3px; font-weight:600; }
.news-text { font-size:13px; color:#1a1a1a; flex:1; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }

/* Tabs */
.cat-tabs { display:grid; grid-template-columns:repeat(3,1fr); background:#fff; padding:8px 12px; }
.tab-item { text-align:center; padding:6px; }
.t1 { font-size:14px; font-weight:600; color:#666; display:block; }
.t2 { font-size:10px; color:#999; display:block; margin-top:2px; }
.tab-item.on .t1 { color:var(--color-primary); }
.tab-item.on .t2 { color:var(--color-primary); }

/* Shop cards */
.shop-list { padding:8px 12px; }
.shop-card { display:flex; gap:10px; background:#fff; border-radius:10px; padding:12px; margin-bottom:8px; }
.shop-logo { width:64px; height:64px; border-radius:8px; background:#f5f5f5; flex-shrink:0; }
.shop-body { flex:1; min-width:0; }
.shop-row1 { display:flex; align-items:center; gap:6px; margin-bottom:3px; }
.shop-name { font-size:14px; font-weight:600; color:#1a1a1a; max-width:200px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.shop-badge { font-size:10px; padding:1px 5px; background:#fff8e1; color:var(--color-warning); border-radius:3px; }
.shop-row2 { font-size:11px; color:var(--color-warning); margin:2px 0; }
.shop-tags { display:flex; flex-wrap:wrap; gap:3px; margin:3px 0; }
.tag { font-size:10px; padding:1px 5px; background:#fff3e0; color:var(--color-warning); border-radius:3px; }
.shop-addr { font-size:11px; color:#666; display:flex; align-items:center; gap:4rpx; margin:3px 0; line-height:1.3; }
.shop-row3 { font-size:11px; color:var(--color-primary); }
.shop-right { display:flex; flex-direction:column; align-items:center; justify-content:space-between; gap:6px; }
.shop-call { width:36px; height:36px; border-radius:50%; background:var(--color-warning); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:600; }
.shop-views { font-size:10px; color:#999; }
.empty { text-align:center; padding:40px; color:#999; font-size:13px; }

/* Join CTA */
.join-cta { position:fixed; right:12px; bottom:80px; width:54px; height:54px; border-radius:50%; background:linear-gradient(135deg,var(--color-warning),var(--color-danger)); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:600; box-shadow:0 4px 12px rgba(255,107,53,.4); z-index:50; }
</style>
