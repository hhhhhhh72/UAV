<template>
  <Layout :current="0">
    <view class="home-page">
      <!-- 1. 蓝顶栏 + 搜索 -->
      <view class="top-bar" :style="{ paddingTop: statusBarH + 'px' }">
        <view class="top-location" @tap="handleLocation">
          <text class="loc-text">全国</text>
          <text class="loc-arrow">▼</text>
        </view>
        <view class="top-search" @tap="handleSearchClick">
          <image class="search-icon" src="/static/icons/search.svg" mode="aspectFit" />
          <text class="search-hint">大家都在搜：吊运项目</text>
          <text class="search-btn">搜索</text>
        </view>
      </view>

      <!-- 2. 横幅广告 -->
      <view class="hero-banner">
        <image class="hero-img" src="/static/home-bg.jpg" mode="aspectFill" />
        <view class="hero-overlay">
          <view class="hero-center">
            <text class="hero-title">免费发布信息</text>
            <view class="hero-tags">
              <text class="ht">吊运</text><text class="ht">植保打药</text><text class="ht">租赁</text><text class="ht">航拍表演</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 3. 数据概览 -->
      <view class="stats-bar">
        <text class="stats-text">📢 浏览：669万 ｜ 发布：848 ｜ 商家：105</text>
        <text class="stats-help">帮助</text>
      </view>

      <!-- 4. 功能矩阵 (3×4 + 浮动分享键) -->
      <view class="func-grid-wrapper">
        <view class="func-grid">
          <view v-for="(f, i) in functions" :key="i" class="func-item" @tap="handleFunc(f)">
            <image v-if="f.icon" :src="f.icon" mode="aspectFit" class="func-icon" />
            <text v-else class="func-emoji">{{ f.emoji }}</text>
            <text class="func-label">{{ f.name }}</text>
          </view>
          <view class="func-item func-share" @tap="handleShare">
            <image src="/static/icons/share.svg" mode="aspectFit" class="func-icon" />
            <text class="func-label">分享</text>
          </view>
        </view>
      </view>

      <!-- 5. 同城公告 -->
      <view class="notice-bar">
        <text class="notice-tag">同城公告</text>
        <swiper class="notice-swipe" vertical autoplay circular interval="3000">
          <swiper-item v-for="(n, i) in notices" :key="i">
            <text class="notice-text">{{ n }}</text>
          </swiper-item>
        </swiper>
        <text class="notice-more">></text>
      </view>

      <!-- 6. 会员/合伙人入口 -->
      <view class="vouch-section">
        <view class="vouch-row member" @tap="navigateTo('/pages/mine/index')">
          <view>
            <text class="vouch-title">加入会员</text>
            <text class="vouch-sub">加入会员更优惠</text>
          </view>
          <image class="vouch-avatar" src="/static/icons/member.svg" mode="aspectFit" />
        </view>
        <view class="vouch-row partner" @tap="navigateTo('/pages/enterprise/register')">
          <view>
            <text class="vouch-title">同城合伙人</text>
            <text class="vouch-sub">加入同城合伙人</text>
          </view>
          <view class="vouch-right">
            <image class="vouch-avatar-sm" src="/static/icons/handshake.svg" mode="aspectFit" />
            <text class="vouch-badge">V</text>
          </view>
        </view>
      </view>

      <!-- 7. 本地商家 -->
      <view class="shop-section">
        <view class="shop-header">
          <text class="shop-title">本地商家</text>
          <text class="shop-more" @tap="navigateTo('/pages/shops/index')">全部 ></text>
        </view>
        <scroll-view scroll-x class="shop-scroll" :show-scrollbar="false">
          <view 
            v-for="(s, i) in shops" :key="i"
            class="shop-card"
            @tap="navigateTo('/pages/services/detail?id=' + s.id)"
          >
            <image class="shop-img" :src="s.logo || '/static/home-bg.jpg'" mode="aspectFill" />
            <text class="shop-name">{{ s.name }}</text>
            <text class="shop-desc">{{ s.desc }}</text>
          </view>
        </scroll-view>
      </view>

      <!-- 8. 更多商家 -->
      <view class="more-shops" @tap="navigateTo('/pages/shops/index')">
        <text>更多商家</text>
        <text class="ml-arrow">></text>
      </view>

    </view>
  </Layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Layout from '@/components/Layout.vue'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'
import { request } from '../../utils/request'

const notices = ref(['飞行须知：请保持安全高度', '无人机登记政策已更新', '欢迎加入同城合伙人计划'])

const functions = ref([
  { name: '吊运独家', emoji: '🚁', path: '/pages/demands/list?type=lift' },
  { name: '买卖租赁', emoji: '🛒', path: '/pages/demands/list?type=trade' },
  { name: '考证培训', emoji: '📜', path: '/pages/training/courses' },
  { name: '植保运输', emoji: '🌾', path: '/pages/demands/list?type=plant' },
  { name: '清洗无人机', emoji: '💧', path: '/pages/demands/list?type=clean' },
  { name: '二手交易', emoji: '🔄', path: '/pages/demands/list?type=secondhand' },
  { name: '免费互助', emoji: '🤝', path: '/pages/demands/list?type=help' },
  { name: '商家入驻', emoji: '🏪', path: '/pages/enterprise/register' },
  { name: '社区交流', emoji: '💬', path: '/pages/community/list' },
  { name: '求职招聘', emoji: '💼', path: '/pages/jobs/list' },
  { name: '需求大厅', emoji: '📋', path: '/pages/demands/list' },
  // 第12个位置留给分享按钮
])

const shops = ref([
  { id: '1', name: '大疆授权店', logo: '', desc: '无人机销售维修' },
  { id: '2', name: '飞手之家', logo: '', desc: '航拍培训基地' },
  { id: '3', name: '天行植保', logo: '', desc: '农业植保服务' },
  { id: '4', name: '极飞科技中心', logo: '', desc: '智能农业方案' },
])

const statusBarH = ref(0)

const handleLocation = () => uni.showToast({ title: '城市选择开发中', icon: 'none' })
const handleSearchClick = () => safeSwitchTab('/pages/services/index')
const handleFunc = (f) => {
  if (f.path.startsWith('/')) safeNavigateTo(f.path)
  else safeSwitchTab(f.path)
}
const handleShare = () => uni.showShareMenu()
const navigateTo = (p) => safeNavigateTo(p)

onMounted(() => {
  const sys = uni.getSystemInfoSync()
  statusBarH.value = (sys.statusBarHeight || 20) + 6
  // 胶囊按钮宽度，避免搜索栏被遮挡
  // #ifdef MP-WEIXIN
  const cap = uni.getMenuButtonBoundingClientRect()
  if (cap) {
    // 让搜索栏右边界不超出胶囊左边界
  }
  // #endif
})
</script>

<style scoped>
.home-page { min-height: 100vh; background: #f2f5f7; padding-bottom: 20px; }

/* 1. 蓝顶栏 */
.top-bar {
  display: flex; align-items: center; gap: 10px;
  padding: 0 14px 10px; background: #1989fa;
}
.top-location { display: flex; align-items: center; gap: 2px; min-width: 52px; }
.loc-text { font-size: 15px; font-weight: 600; color: #fff; }
.loc-arrow { font-size: 10px; color: rgba(255,255,255,0.7); }
.top-search {
  flex: 1; display: flex; align-items: center;
  background: rgba(255,255,255,0.2); border-radius: 20px; padding: 8px 12px; gap: 6px;
}
.search-icon { width: 14px; height: 14px; opacity: 0.6; filter: brightness(0) invert(1); }
.search-hint { flex: 1; font-size: 13px; color: rgba(255,255,255,0.6); }
.search-btn { font-size: 13px; color: #fff; font-weight: 500; padding-left: 8px; border-left: 1px solid rgba(255,255,255,0.3); }

/* 2. 横幅 */
.hero-banner { position: relative; height: 160px; overflow: hidden; }
.hero-img { width: 100%; height: 100%; display: block; }
.hero-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(180deg, rgba(0,0,0,0.2) 0%, rgba(0,0,0,0.4) 100%);
  display: flex; align-items: center; justify-content: center;
}
.hero-center { text-align: center; }
.hero-title { font-size: 26px; font-weight: 700; color: #fff; display: block; margin-bottom: 10px; text-shadow: 0 2px 8px rgba(0,0,0,0.3); }
.hero-tags { display: flex; gap: 6px; justify-content: center; flex-wrap: wrap; }
.ht { padding: 4px 12px; background: rgba(255,255,255,0.2); color: #fff; font-size: 11px; border-radius: 10px; }

/* 3. 数据条 */
.stats-bar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 14px; background: #fff; border-bottom: 1px solid #eee;
}
.stats-text { font-size: 12px; color: #666; }
.stats-help { font-size: 12px; color: #1989fa; font-weight: 500; }

/* 4. 功能矩阵 */
.func-grid-wrapper { background: #fff; padding: 16px 10px 8px; position: relative; }
.func-grid {
  display: grid; grid-template-columns: repeat(4, 1fr); row-gap: 16px; text-align: center;
}
.func-item { display: flex; flex-direction: column; align-items: center; gap: 6px; padding: 4px 0; }
.func-icon { width: 36px; height: 36px; }
.func-emoji { font-size: 28px; }
.func-label { font-size: 12px; color: #333; }
.func-share {
  background: linear-gradient(135deg, #07c160, #06ad56);
  border-radius: 50%; width: 44px; height: 44px; margin: 0 auto;
  display: flex; align-items: center; justify-content: center;
}
.func-share .func-icon { width: 20px; height: 20px; filter: brightness(0) invert(1); }
.func-share .func-label { font-size: 10px; color: #07c160; margin-top: 2px; }

/* 5. 同城公告 */
.notice-bar {
  display: flex; align-items: center; gap: 8px; margin: 10px 12px;
  background: #fff; border-radius: 10px; padding: 10px 14px;
}
.notice-tag { font-size: 13px; color: #ff6b35; font-weight: 600; white-space: nowrap; border-right: 1px solid #eee; padding-right: 10px; }
.notice-swipe { flex: 1; height: 20px; }
.notice-text { font-size: 12px; color: #666; line-height: 20px; }
.notice-more { font-size: 14px; color: #ccc; }

/* 6. 会员/合伙人 */
.vouch-section { margin: 0 12px 10px; display: flex; gap: 10px; }
.vouch-row {
  flex: 1; display: flex; align-items: center; justify-content: space-between;
  border-radius: 12px; padding: 14px;
}
.vouch-row.member { background: linear-gradient(135deg, #e8f8ee, #d4f2e2); }
.vouch-row.partner { background: linear-gradient(135deg, #fff3e8, #ffe8d4); }
.vouch-title { font-size: 15px; font-weight: 600; color: #333; display: block; margin-bottom: 4px; }
.vouch-sub { font-size: 11px; color: #999; }
.vouch-avatar { width: 44px; height: 44px; border-radius: 50%; background: #c8e6c9; }
.vouch-right { display: flex; align-items: center; gap: 6px; }
.vouch-avatar-sm { width: 36px; height: 36px; border-radius: 50%; background: #ffe0b2; }
.vouch-badge { font-size: 14px; font-weight: 700; color: #ff6b35; }

/* 7. 本地商家 */
.shop-section { margin: 0 12px 10px; background: #fff; border-radius: 12px; padding: 16px; }
.shop-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.shop-title { font-size: 16px; font-weight: 600; color: #1a1a1a; }
.shop-more { font-size: 13px; color: #999; }
.shop-scroll { white-space: nowrap; }
.shop-card {
  display: inline-block; width: 100px; margin-right: 12px; text-align: center; vertical-align: top;
}
.shop-img { width: 80px; height: 80px; border-radius: 12px; display: block; margin: 0 auto 8px; background: #e8f2fc; }
.shop-name { font-size: 13px; font-weight: 500; color: #333; display: block; }
.shop-desc { font-size: 11px; color: #999; }

/* 8. 更多商家 */
.more-shops {
  text-align: center; padding: 14px; margin: 0 12px; background: #fff; border-radius: 10px;
  font-size: 14px; color: #666;
}
.ml-arrow { color: #ccc; margin-left: 4px; }
</style>
