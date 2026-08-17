<template>
<view class="page" v-if="product.id">
  <!-- ═══════ 图区 ═══════ -->
  <view class="img-box">
    <swiper v-if="images.length" class="img-swiper-box" :current="curImg" @change="e=>curImg=e.detail.current" circular autoplay :interval="3500" :duration="400">
      <swiper-item v-for="(img,i) in images" :key="i">
        <image :src="img" mode="aspectFill" class="img-swiper" :class="{ 'img-loaded': loadedImgs[i] }" @load="onImgLoad(i)" @click="preview(i)" />
      </swiper-item>
    </swiper>
    <image v-else src="/static/home/demand-lift.jpg" mode="aspectFill" class="img-swiper" :class="{ 'img-loaded': true }" />
    <!-- 返回按钮 -->
    <view class="img-back" hover-class="back-press" @tap="goBack">
      <view class="back-arrow"></view>
    </view>
    <!-- 图片计数 -->
    <view v-if="images.length > 1" class="img-count">{{ curImg + 1 }}/{{ images.length }}</view>
    <!-- 圆点指示器 -->
    <view v-if="images.length > 1" class="img-dots">
      <view v-for="(_,i) in images.length" :key="i" class="dot" :class="{ on: i === curImg }" />
    </view>
  </view>

  <!-- ═══════ 锚点导航（商品/参数/图文，吸顶） ═══════ -->
  <view class="anchor-nav" :class="{ 'anchor-nav--sticky': anchorSticky }">
    <view
      v-for="a in anchors"
      :key="a.id"
      class="anchor-item"
      :class="{ on: activeAnchor === a.id }"
      @tap="scrollToAnchor(a.id)"
    >{{ a.name }}</view>
  </view>

  <!-- ═══════ 价格区 ═══════ -->
  <view id="anchor-product" class="card price-card">
    <view class="price-row">
      <view class="price-left">
        <text class="price-symbol">¥</text>
        <text class="price-num">{{ priceDisplay }}</text>
        <text class="price-decimal">.{{ priceDec }}</text>
      </view>
      <text class="price-tag" :class="tagClass">{{ product.condition === 'used' ? '二手' : (product.condition === 'new' ? '全新' : '商家发布') }}</text>
    </view>
    <text class="price-title">{{ product.title }}</text>
    <text class="price-desc" v-if="product.description">{{ product.description }}</text>
    <view class="price-meta">
      <text class="meta-item">已浏览 {{ product.views || 0 }} 次</text>
      <text class="meta-dot">·</text>
      <text class="meta-item" :class="product.status === 'listed' ? 'on' : 'off'">{{ statusLabel }}</text>
    </view>
  </view>

  <!-- ═══════ 服务保障行 ═══════ -->
  <view class="card assurance">
    <view class="assurance-item">
      <view class="assurance-dot blue"></view>
      <text class="assurance-txt">平台认证</text>
    </view>
    <view class="assurance-item">
      <view class="assurance-dot green"></view>
      <text class="assurance-txt">商家直供</text>
    </view>
    <view class="assurance-item">
      <view class="assurance-dot orange"></view>
      <text class="assurance-txt">{{ product.condition === 'used' ? '成色如描述' : '支持验机' }}</text>
    </view>
  </view>

  <!-- ═══════ 参数区 ═══════ -->
  <view id="anchor-params" class="card spec-card">
    <view class="card-head">商品参数</view>
    <view class="spec-list">
      <view class="spec-row" v-if="product.brand">
        <text class="spec-label">品牌</text><text class="spec-val">{{ product.brand }}</text>
      </view>
      <view class="spec-row" v-if="product.model">
        <text class="spec-label">型号</text><text class="spec-val">{{ product.model }}</text>
      </view>
      <view class="spec-row">
        <text class="spec-label">成色</text><text class="spec-val">{{ product.condition === 'used' ? '二手' : (product.condition === 'new' ? '全新' : '商家发布') }}</text>
      </view>
      <view class="spec-row">
        <text class="spec-label">类型</text><text class="spec-val">{{ typeLabel(product.prod_type) }}</text>
      </view>
      <view class="spec-row" v-if="product.seller_name">
        <text class="spec-label">卖家</text><text class="spec-val">{{ product.seller_name }}</text>
      </view>
    </view>
  </view>

  <!-- ═══════ 商家 ═══════ -->
  <view class="card shop-card" v-if="product.seller_name" @tap="contactShop">
    <view class="shop-ava">{{ product.seller_name[0] }}</view>
    <view class="shop-info">
      <text class="shop-name">{{ product.seller_name }}</text>
      <text class="shop-status">平台认证商家</text>
    </view>
    <view class="shop-btn" hover-class="btn-press">联系卖家</view>
  </view>

  <!-- ═══════ 图文详情 ═══════ -->
  <view id="anchor-detail" class="card detail-card">
    <view class="card-head">商品详情</view>
    <view>
      <text class="detail-text">该商品由{{ product.seller_name || '认证商家' }}发布，{{ product.condition === 'new' ? '全新未拆封，支持平台检测。' : (product.condition === 'used' ? '二手设备，成色如描述所述。' : '商品信息以页面展示为准。') }}</text>
      <text class="detail-text" v-if="product.description">{{ product.description }}</text>
    </view>
  </view>

  <!-- ═══════ 底部操作栏 ═══════ -->
  <view class="bottom">
    <view class="bottom-left">
      <view class="bottom-fav" :class="{ 'fav-active': isFav }" hover-class="btn-press" @tap="toggleFav">
        <text class="fav-ico" :class="{ 'heart-pop': favAnim }">♥</text>
        <text>收藏</text>
      </view>
      <view class="bottom-share" hover-class="btn-press" @tap="onShare">
        <text class="share-ico" :class="{ 'share-pop': shareAnim }">↗</text>
        <text>分享</text>
      </view>
    </view>
    <view class="bottom-contact" hover-class="btn-press" @tap="contactShop">联系卖家</view>
    <view class="bottom-buy" hover-class="btn-press" @tap="buy">{{ product.prod_type === 'test_fly' ? '预约试飞' : '立即购买' }}</view>
  </view>
</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPageScroll } from '@dcloudio/uni-app'
import { productTypeLabel } from '@/utils/enums'
import { request, getStoredUser } from '../../../utils/request'
import { fullImgUrl } from '../../../utils/hallData'

const product = ref({})
const images = ref([])
const curImg = ref(0)
const loadedImgs = ref([])

const priceInt = computed(() => Math.floor((product.value.price_fen||0) / 100))
const priceDec = computed(() => String((product.value.price_fen||0) % 100).padStart(2,'0'))
// 价格数字滚动：0 → 实际值（700ms 缓出）
const priceDisplay = ref(0)
let priceTimer = null
const animatePrice = (target) => {
  if (priceTimer) clearInterval(priceTimer)
  const steps = 18
  const perStep = Math.max(1, Math.round(target / steps))
  let current = 0
  priceTimer = setInterval(() => {
    current += perStep
    if (current >= target) {
      priceDisplay.value = target
      clearInterval(priceTimer)
      priceTimer = null
    } else {
      priceDisplay.value = current
    }
  }, 700 / steps)
}
const onImgLoad = (i) => {
  loadedImgs.value[i] = true
}
const tagClass = computed(() => {
  const c = product.value.condition || ''
  if (c === 'new') return 'tag-green'
  if (c === 'used') return 'tag-orange'
  return 'tag-gray'
})
const typeLabel = (t) => productTypeLabel(t) || '商品'
const statusLabel = computed(() => {
  const s = product.value.status || 'listed'
  if (s === 'sold') return '已售'
  if (s === 'removed') return '已下架'
  return '在售'
})

// 锚点导航（商品/参数/图文）
const anchors = [
  { id: 'anchor-product', name: '商品' },
  { id: 'anchor-params', name: '参数' },
  { id: 'anchor-detail', name: '图文' },
]
const activeAnchor = ref('anchor-product')
const anchorSticky = ref(false)
// 各锚点相对页面顶部的绝对位置（测量一次，滚动后不失效）
const anchorAbs = {}
let lastScrollTop = 0
// 手动点击后的锁定锚点：滚动动画期间 onPageScroll 不覆盖高亮（修复切换跳回）
let pendingAnchor = null
const NAV_OFFSET_PX = uni.upx2px(84) // 吸顶导航高度 + 间距补偿（84rpx）

const scrollToAnchor = (id) => {
  pendingAnchor = id
  activeAnchor.value = id
  // 解锁延时必须大于滚动动画时长（260ms）——complete 回调在微信端会提前触发，
  // 动画中途解锁会被 onPageScroll 的高亮重算抢回原锚点
  setTimeout(() => { if (pendingAnchor === id) pendingAnchor = null }, 360)
  const abs = anchorAbs[id]
  if (abs !== undefined) {
    uni.pageScrollTo({ scrollTop: Math.max(0, abs - NAV_OFFSET_PX), duration: 260 })
    return
  }
  // 兜底：实时测量（绝对位置 = 相对视口 top + 当前 scrollTop）
  const q = uni.createSelectorQuery()
  q.select('#' + id).boundingClientRect((rect) => {
    if (!rect) return
    anchorAbs[id] = rect.top + lastScrollTop
    uni.pageScrollTo({ scrollTop: Math.max(0, anchorAbs[id] - NAV_OFFSET_PX), duration: 260 })
  })
  q.exec()
}
// 高亮当前区块（最后一个 abs <= scrollTop + 阈值 的锚点）
const recomputeHighlight = () => {
  let current = anchors[0].id
  anchors.forEach(a => {
    const off = anchorAbs[a.id]
    if (off !== undefined && off <= lastScrollTop + 100) current = a.id
  })
  activeAnchor.value = current
}
// 数据加载后测量各区块位置
const measureAnchors = () => {
  setTimeout(() => {
    const q = uni.createSelectorQuery()
    anchors.forEach(a => {
      q.select('#' + a.id).boundingClientRect(rect => {
        if (rect) anchorAbs[a.id] = rect.top + lastScrollTop
      })
    })
    q.exec()
  }, 300)
}
onPageScroll((e) => {
  lastScrollTop = e.scrollTop
  anchorSticky.value = e.scrollTop > 320
  if (pendingAnchor) return // 滚动动画期间保持点击的高亮
  recomputeHighlight()
})

onLoad((opts) => {
  if (!opts.id) return;
  (async () => {
    try {
      const p = await request({ url: '/api/v1/products/' + encodeURIComponent(opts.id) })
      if (p && p.id) {
        product.value = p
        // 图片处理与列表页一致（fullImgUrl）：/static/ 原样使用，/uploads/ 拼后端地址
        try {
          const arr = typeof p.images === 'string' ? JSON.parse(p.images) : (p.images || [])
          images.value = arr.map(fullImgUrl).filter(Boolean)
        } catch { images.value = [] }
        loadedImgs.value = images.value.map(() => false)
        animatePrice(priceInt.value)
        measureAnchors()
      }
    } catch {}
  })()
})

const preview = (i) => {
  uni.previewImage({ current: i, urls: images.value })
}
const goBack = () => uni.navigateBack()
const contactShop = () => uni.showToast({ title: '已复制卖家联系方式', icon: 'none' })

// 收藏：状态切换（灰心 → 红心）+ 心跳动画
const isFav = ref(false)
const favAnim = ref(false)
const toggleFav = () => {
  isFav.value = !isFav.value
  favAnim.value = true
  setTimeout(() => { favAnim.value = false }, 450)
  uni.showToast({ title: isFav.value ? '已收藏' : '已取消收藏', icon: 'none' })
}

// 分享：弹跳反馈
const shareAnim = ref(false)
const onShare = () => {
  shareAnim.value = true
  setTimeout(() => { shareAnim.value = false }, 450)
  uni.showToast({ title: '分享', icon: 'none' })
}
// 主行动按钮：试飞测试供给 → 测试场地列表（②展示卡 → ③预约入口打通）；其余为立即购买（下单闭环）
const buy = async () => {
  if (product.value.prod_type === 'test_fly') {
    uni.navigateTo({ url: '/pkg-service/pages/testsites/list' })
    return
  }
  // 面议商品（无定价）不支持下单，引导联系卖家
  const fen = product.value.price_fen || 0
  if (fen <= 0) {
    uni.showToast({ title: '该商品为面议报价，请联系卖家', icon: 'none' })
    return
  }
  // 登录校验（未登录跳登录页）
  if (!getStoredUser()) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  uni.showLoading({ title: '下单中...' })
  try {
    const o = await request({
      url: '/api/v1/trade-orders',
      method: 'POST',
      data: { product_id: product.value.id, seller_id: product.value.seller_id || '', amount_fen: fen },
    })
    uni.hideLoading()
    uni.showToast({ title: '下单成功', icon: 'success' })
    // 跳订单详情：新订单为待付款（支付接口接入前保持该状态）
    setTimeout(() => {
      uni.redirectTo({ url: '/pages/orders/detail?id=' + encodeURIComponent(o.id) + '&type=product' })
    }, 600)
  } catch (e) {
    uni.hideLoading()
    const msg = (e && e.data && e.data.error && e.data.error.message) || '下单失败，请稍后重试'
    uni.showToast({ title: msg, icon: 'none' })
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(140rpx + env(safe-area-inset-bottom));
}

/* ═══════ 图区 ═══════ */
.img-box {
  position: relative;
  height: 660rpx;
  background: linear-gradient(160deg, #074D92 0%, #0A66C2 55%, #1DD4A8 130%);
  overflow: hidden;
}
.img-swiper-box {
  width: 100%;
  height: 100%;
}
.img-swiper {
  width: 100%;
  height: 100%;
  display: block;
  opacity: 0;
  transition: opacity .35s ease;
}
.img-swiper.img-loaded { opacity: 1; }
.img-back {
  position: absolute;
  top: calc(16rpx + var(--status-bar-height));
  left: 24rpx;
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 4rpx 16rpx rgba(7, 77, 146, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 5;
}
.back-arrow {
  width: 20rpx;
  height: 20rpx;
  border-left: 4rpx solid #17212B;
  border-bottom: 4rpx solid #17212B;
  transform: rotate(45deg);
  margin-left: 6rpx;
}
.img-count {
  position: absolute;
  right: 24rpx;
  bottom: 28rpx;
  padding: 6rpx 18rpx;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 22rpx;
  border-radius: 24rpx;
  z-index: 5;
}
.img-dots {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 36rpx;
  display: flex;
  justify-content: center;
  gap: 10rpx;
  z-index: 4;
}
.dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.35);
  transition: all .25s;
}
.dot.on {
  background: #fff;
  width: 32rpx;
  border-radius: 6rpx;
}

/* ═══════ 锚点导航 ═══════ */
.anchor-nav {
  display: flex;
  background: #fff;
  padding: 0 24rpx;
  z-index: 20;
}
.anchor-nav--sticky {
  position: sticky;
  top: 0;
  box-shadow: 0 4rpx 16rpx rgba(16, 24, 40, 0.06);
}
.anchor-item {
  position: relative;
  padding: 22rpx 32rpx 20rpx;
  font-size: 26rpx;
  color: #667085;
}
.anchor-item.on {
  color: #0A66C2;
  font-weight: 700;
}
.anchor-item.on::after {
  content: '';
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  bottom: 0;
  width: 48rpx;
  height: 8rpx;
  border-radius: 4rpx;
  background: #0A66C2;
}

/* ═══════ 通用卡片 ═══════ */
.card {
  margin: 0 24rpx 20rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx;
  box-shadow: 0 4rpx 16rpx rgba(7, 77, 146, 0.06);
}
.card-head {
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 8rpx;
}

/* ═══════ 价格区 ═══════ */
.price-card {
  margin-top: 20rpx;
}
.price-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.price-left {
  display: flex;
  align-items: baseline;
  color: #E84C3D;
}
.price-symbol { font-size: 32rpx; font-weight: 700; }
.price-num { font-size: 56rpx; font-weight: 800; line-height: 1; }
.price-decimal { font-size: 28rpx; font-weight: 600; }
.price-tag {
  font-size: 22rpx;
  padding: 6rpx 16rpx;
  border-radius: 24rpx;
  font-weight: 600;
}
.tag-green { background: #E8F7EF; color: #16A34A; }
.tag-orange { background: #FDF1E7; color: #E46426; }
.tag-gray { background: #F0F3F6; color: #667085; }
.price-title {
  display: block;
  margin-top: 16rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
}
.price-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #667085;
  line-height: 1.6;
}
.price-meta {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 16rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #EEF1F4;
}
.meta-item {
  font-size: 22rpx;
  color: #98A2B3;
}
.meta-item.on { color: #0A66C2; font-weight: 600; }
.meta-item.off { color: #E84C3D; font-weight: 600; }
.meta-dot { color: #D0D5DD; font-size: 22rpx; }

/* ═══════ 服务保障行 ═══════ */
.assurance {
  display: flex;
  align-items: center;
  padding: 24rpx 28rpx;
}
.assurance-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
}
.assurance-item + .assurance-item {
  border-left: 1rpx solid #EEF1F4;
}
.assurance-dot {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
}
.assurance-dot.blue { background: #0A66C2; }
.assurance-dot.green { background: #1DD4A8; }
.assurance-dot.orange { background: #F79009; }
.assurance-txt {
  font-size: 24rpx;
  color: #344054;
  font-weight: 500;
}

/* ═══════ 参数区 ═══════ */
.spec-list {
  display: flex;
  flex-direction: column;
}
.spec-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #F2F4F7;
}
.spec-row:last-child { border-bottom: none; }
.spec-label { font-size: 24rpx; color: #98A2B3; }
.spec-val { font-size: 26rpx; color: #17212B; font-weight: 600; }

/* ═══════ 商家卡 ═══════ */
.shop-card {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx 28rpx;
}
.shop-ava {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  background: linear-gradient(140deg, #0A66C2, #074D92);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 700;
  flex-shrink: 0;
}
.shop-info { flex: 1; min-width: 0; }
.shop-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.shop-status {
  display: inline-block;
  margin-top: 8rpx;
  font-size: 20rpx;
  color: #0A66C2;
  background: #EAF3FD;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}
.shop-btn {
  flex-shrink: 0;
  padding: 14rpx 28rpx;
  border-radius: 32rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  font-weight: 600;
  transition: transform .18s;
}

/* ═══════ 图文详情 ═══════ */
.detail-text {
  display: block;
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #344054;
  line-height: 1.9;
}

/* ═══════ 底部操作栏 ═══════ */
.bottom {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  display: flex;
  align-items: center;
  gap: 16rpx;
  border-top: 1rpx solid #EEF1F4;
  box-shadow: 0 -4rpx 16rpx rgba(0, 0, 0, 0.04);
  z-index: 30;
}
.bottom-left {
  display: flex;
  gap: 8rpx;
}
.bottom-fav, .bottom-share {
  width: 88rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 20rpx;
  color: #667085;
  line-height: 1.4;
}
.fav-ico {
  font-size: 34rpx;
  line-height: 1.2;
  color: #98A2B3;
  transition: color .2s;
}
.fav-active .fav-ico { color: #E84C3D; }
.share-ico { font-size: 32rpx; line-height: 1.2; }
.bottom-contact {
  flex: 1;
  height: 80rpx;
  border-radius: 40rpx;
  background: #fff;
  border: 2rpx solid #0A66C2;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
  transition: transform .18s;
}
.bottom-buy {
  flex: 1.2;
  height: 80rpx;
  border-radius: 40rpx;
  background: linear-gradient(135deg, #0A66C2, #074D92);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
  transition: transform .18s;
}

/* ═══════ 按压反馈 ═══════ */
.btn-press { transform: scale(.96); }
.back-press { transform: scale(.9); }

/* ═══════ 收藏 / 分享交互动画 ═══════ */
@keyframes heart-pop {
  0% { transform: scale(1); }
  40% { transform: scale(1.4); }
  70% { transform: scale(.88); }
  100% { transform: scale(1); }
}
.heart-pop { animation: heart-pop .42s ease; }
@keyframes share-pop {
  0% { transform: scale(1) translateY(0); }
  40% { transform: scale(1.3) translateY(-4px); }
  100% { transform: scale(1) translateY(0); }
}
.share-pop { animation: share-pop .42s ease; }

/* ═══════ 入场动画（仅 transform/opacity） ═══════ */
@keyframes hero-zoom {
  from { opacity: 0; transform: scale(1.08); }
  to { opacity: 1; transform: scale(1); }
}
@keyframes card-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes tag-pop {
  0% { opacity: 0; transform: scale(.7); }
  70% { transform: scale(1.06); }
  100% { opacity: 1; transform: scale(1); }
}
@keyframes bar-up {
  from { opacity: 0; transform: translateY(100%); }
  to { opacity: 1; transform: translateY(0); }
}
.img-box { animation: hero-zoom .4s ease-out both; }
.price-card { animation: card-up .26s ease-out both; }
.price-tag { animation: tag-pop .3s ease-out .1s both; }
.assurance { animation: card-up .26s ease-out .05s both; }
.spec-card { animation: card-up .26s ease-out .1s both; }
.shop-card { animation: card-up .26s ease-out .15s both; }
.detail-card { animation: card-up .26s ease-out .2s both; }
.bottom { animation: bar-up .28s ease-out .18s both; }

/* 尊重“减少动态效果”设置 */
@media (prefers-reduced-motion: reduce) {
  .img-box, .price-card, .price-tag, .assurance, .spec-card, .shop-card, .detail-card, .bottom { animation: none; }
  .img-swiper { opacity: 1; transition: none; }
  .heart-pop, .share-pop { animation: none; }
}
</style>
