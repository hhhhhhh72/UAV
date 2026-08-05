<template>
<Layout :current="1">
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
      <text class="search-cat">店铺</text>
      <input class="search-input" v-model="keyword" placeholder="搜店铺" confirm-type="search" />
      <u-icon name="search" size="28rpx" color="#999999" />
    </view>

    <!-- 分类图标 -->
    <view class="cat-row">
      <view v-for="c in cats" :key="c.id" class="cat-item" @tap="pickCat(c.id)">
        <view class="cat-icon">{{ c.icon }}</view>
        <text class="cat-name">{{ c.name }}</text>
      </view>
    </view>

    <!-- 公告（真实数据：首页配置公告） -->
    <view class="news-bar">
      <text class="news-tag">协会公告</text>
      <text class="news-text" @tap="goJoin">{{ notices[0] || '欢迎入驻重庆市无人机产业协会 · 会员商家展示' }}</text>
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

    <!-- 店铺卡片（真实数据：home.shops） -->
    <view class="shop-list">
      <view v-for="s in filteredShops" :key="s.id" class="shop-card">
        <image :src="s.logo_url || '/static/home-bg.jpg'" mode="aspectFill" class="shop-logo" />
        <view class="shop-body">
          <view class="shop-row1">
            <text class="shop-name">{{ s.name }}</text>
            <text v-if="s.is_member" class="shop-badge">会员</text>
          </view>
          <view v-if="s.description" class="shop-tags">
            <text class="tag">{{ s.description }}</text>
          </view>
          <view class="shop-addr"><u-icon name="location" size="24rpx" color="#969799" /><text>{{ s.address || '暂无地址' }}</text></view>
          <view class="shop-row3">
            <text>{{ (s.views || 0) + ' 次浏览' }}</text>
          </view>
        </view>
        <view class="shop-right">
          <view class="shop-call" @tap.stop="callShop(s)">电</view>
        </view>
      </view>
      <view v-if="!filteredShops.length" class="empty">暂无商家</view>
    </view>
  </view>
</Layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import Layout from '@/components/Layout.vue'
import { request } from '@/utils/request'

const banners = ref([
  { bg: '#4fc3f7', title: '无人机商家入驻' },
  { bg: '#81c784', title: '优质商家推荐' },
  { bg: '#ff8a65', title: '限时入驻优惠' },
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
const notices = ref([])
const keyword = ref('')

// 搜索 + Tab 过滤：推荐=全部 / 新人=最新加入 / 附近=全部（无定位时展示全部）
const filteredShops = computed(() => {
  let list = shops.value
  const kw = keyword.value.trim().toLowerCase()
  if (kw) list = list.filter((s) => (s.name || '').toLowerCase().includes(kw))
  if (activeTab.value === 'new') {
    list = [...list].sort((a, b) => String(b.created_at || '').localeCompare(String(a.created_at || '')))
  }
  return list
})

const loadShops = async () => {
  try {
    const res = await request({ url:'/api/v1/home' })
    const data = res.data || res
    shops.value = data.shops || []
    notices.value = data.notices || []
  } catch {}
}
loadShops()

// 分类入口：可跳转的真实页面；暂无可跳页面时引导联系协会
const pickCat = (id) => {
  if (id === 'train') return uni.navigateTo({ url: '/pages/training/courses' })
  if (id === 'sale' || id === 'parts') return uni.navigateTo({ url: '/pages/mall/index' })
  uni.showToast({ title: '该分类商家请联系协会推荐', icon: 'none' })
}
const goJoin = () => uni.navigateTo({ url:'/pages/enterprise/register' })
const callShop = (s) => {
  const phone = s && (s.contact_phone || s.phone)
  if (!phone) {
    uni.showToast({ title: '该商家暂未提供联系电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({ phoneNumber: String(phone) })
}
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
.join-cta { position:fixed; right:12px; bottom:80px; width:54px; height:54px; border-radius:50%; background:var(--color-warning)); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:600; box-shadow:0 4px 12px rgba(255,107,53,.4); z-index:50; }
</style>
