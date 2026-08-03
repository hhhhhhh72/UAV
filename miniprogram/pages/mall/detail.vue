<template>
<view class="page" v-if="product.id">
  <!-- 图区 -->
  <view class="img-box">
    <swiper v-if="images.length" :current="curImg" @change="e=>curImg=e.detail.current" circular>
      <swiper-item v-for="(img,i) in images" :key="i">
        <image :src="img" mode="aspectFill" class="img-swiper" :class="{ 'img-loaded': loadedImgs[i] }" @load="onImgLoad(i)" @click="preview(i)" />
      </swiper-item>
    </swiper>
    <image v-else src="/static/home-bg.jpg" mode="aspectFill" class="img-swiper" :class="{ 'img-loaded': true }" />
    <view class="img-dots">
      <view v-for="(_,i) in (images.length||1)" :key="i" class="dot" :class="{on:i===curImg}" />
    </view>
    <view class="img-back" hover-class="back-press" @tap="goBack">←</view>
  </view>

  <!-- 锚点导航（Tigshop 风格：商品/参数/图文 滚动定位） -->
  <view class="anchor-nav" :class="{ 'anchor-nav--sticky': anchorSticky }">
    <view
      v-for="a in anchors"
      :key="a.id"
      class="anchor-item"
      :class="{ on: activeAnchor === a.id }"
      @tap="scrollToAnchor(a.id)"
    >{{ a.name }}</view>
  </view>

  <!-- 价格区 -->
  <view id="anchor-product" class="price-card">
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
  </view>

  <!-- 参数区（Tigshop 逐行信息风格） -->
  <view id="anchor-params" class="spec-card">
    <view class="spec-head">商品参数</view>
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
      <view class="spec-row">
        <text class="spec-label">浏览量</text><text class="spec-val">{{ product.views || 0 }}次</text>
      </view>
      <view class="spec-row" v-if="product.seller_name">
        <text class="spec-label">卖家</text><text class="spec-val">{{ product.seller_name }}</text>
      </view>
    </view>
  </view>

  <!-- 商家 -->
  <view class="shop-card" v-if="product.seller_name" @tap="contactShop">
    <view class="shop-ava">{{ product.seller_name[0] }}</view>
    <view class="shop-info">
      <text class="shop-name">{{ product.seller_name }}</text>
      <text class="shop-status">平台商家</text>
    </view>
    <text class="shop-btn" hover-class="btn-press">联系卖家</text>
  </view>

  <!-- 图文详情 -->
  <view id="anchor-detail" class="detail-card">
    <view class="detail-head">商品详情</view>
    <view class="detail-body">
      <text class="detail-text">该商品由{{ product.seller_name || '认证商家' }}发布，支持{{ product.condition || '平台检测' }}。</text>
      <text class="detail-text" v-if="product.description">{{ product.description }}</text>
    </view>
  </view>

  <!-- 底部栏（Tigshop 风格：收藏/分享 + 联系卖家 + 主行动） -->
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
    <view class="bottom-contact" hover-class="btn-press" @tap="contactShop">
      <text class="contact-ico">电</text>
      <text class="contact-txt">联系卖家</text>
    </view>
    <view class="bottom-buy" hover-class="btn-press" @tap="buy">咨询报价</view>
  </view>
</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPageScroll } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../utils/request'

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
const typeLabel = (t) => ({ drone: '整机', part: '配件', repair: '维修服务' }[t] || '商品')

// 锚点导航（商品/参数/图文）
const anchors = [
  { id: 'anchor-product', name: '商品' },
  { id: 'anchor-params', name: '参数' },
  { id: 'anchor-detail', name: '图文' },
]
const activeAnchor = ref('anchor-product')
const anchorSticky = ref(false)
const anchorOffsets = {}
const scrollToAnchor = (id) => {
  uni.pageScrollTo({ selector: '#' + id, duration: 260 })
  activeAnchor.value = id
}
// 数据加载后测量各区块位置
const measureAnchors = () => {
  setTimeout(() => {
    const q = uni.createSelectorQuery()
    anchors.forEach(a => {
      q.select('#' + a.id).boundingClientRect(rect => {
        if (rect) anchorOffsets[a.id] = rect.top
      })
    })
    q.exec()
  }, 300)
}
onPageScroll((e) => {
  const top = e.scrollTop
  anchorSticky.value = top > 320
  // 高亮当前区块（最后一个 offset <= scrollTop + 导航高度 的锚点）
  let current = 'anchor-product'
  anchors.forEach(a => {
    const off = anchorOffsets[a.id]
    if (off !== undefined && off <= top + 100) current = a.id
  })
  activeAnchor.value = current
})

onLoad((opts) => {
  if (!opts.id) return;
  (async () => {
    try {
      const p = await request({ url: '/api/v1/products/' + encodeURIComponent(opts.id) })
      if (p && p.id) {
        product.value = p
        // 图片相对路径（/uploads/...）拼后端地址供小程序加载
        try {
          const arr = typeof p.images === 'string' ? JSON.parse(p.images) : (p.images || [])
          images.value = arr.map(u => (u && u.startsWith('http') ? u : BASE_URL + u))
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
const buy = () => uni.showToast({ title: '咨询报价功能开发中', icon: 'none' })
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: 60px; }

/* ===== 详情页动画系统（仅 transform/opacity，260ms 内，层次化） ===== */

/* 图区：淡入 + 轻微缩放（1.03→1） */
@keyframes hero-zoom {
  from { opacity: 0; transform: scale(1.03); }
  to { opacity: 1; transform: scale(1); }
}
/* 卡片：淡入上移（位移 16px，比基础更明显） */
@keyframes card-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
/* 价格标签：弹入（放大回位） */
@keyframes tag-pop {
  0% { opacity: 0; transform: scale(.7); }
  70% { transform: scale(1.06); }
  100% { opacity: 1; transform: scale(1); }
}
/* 底部栏：从下方滑入 */
@keyframes bar-up {
  from { opacity: 0; transform: translateY(100%); }
  to { opacity: 1; transform: translateY(0); }
}

.img-box { animation: hero-zoom .3s ease-out both; }

/* 图片加载完成淡入（避免白屏闪烁） */
.img-swiper { opacity: 0; transition: opacity .35s ease, transform .35s ease; }
.img-swiper.img-loaded { opacity: 1; }
/* 返回按钮：图区入场后从左侧滑入 */
@keyframes back-in {
  from { opacity: 0; transform: translateX(-24px); }
  to { opacity: 1; transform: translateX(0); }
}
.img-back { animation: back-in .3s ease-out .28s both; }
.price-card { animation: card-up .26s ease-out both; }
.price-tag { animation: tag-pop .3s ease-out .1s both; }
.spec-card { animation: card-up .26s ease-out .07s both; }
.shop-card { animation: card-up .26s ease-out .14s both; }
.detail-card { animation: card-up .26s ease-out .21s both; }
.bottom { animation: bar-up .28s ease-out .18s both; }

/* 尊重“减少动态效果”设置 */
@media (prefers-reduced-motion: reduce) {
  .img-box, .price-card, .price-tag, .spec-card, .shop-card, .detail-card, .bottom, .img-back { animation: none; }
  .img-swiper { opacity: 1; transition: none; }
  .heart-pop, .share-pop { animation: none; }
}

.img-box { position: relative; height: 320px; background: var(--color-text); }

/* 锚点导航（Tigshop 风格：吸顶定位） */
.anchor-nav {
  display: flex; background: #fff; border-bottom: 1px solid var(--color-divider);
  padding: 0 12px; z-index: 20;
}
.anchor-nav--sticky { position: sticky; top: 0; box-shadow: 0 2px 8px rgba(16,24,40,.06); }
.anchor-item {
  position: relative; padding: 22rpx 20rpx; font-size: 26rpx; color: var(--color-text-secondary);
}
.anchor-item.on { color: var(--color-primary); font-weight: 600; }
.anchor-item.on::after {
  content: ''; position: absolute; left: 50%; transform: translateX(-50%);
  bottom: 0; width: 40rpx; height: 6rpx; border-radius: 3rpx; background: var(--color-primary);
}
.img-swiper { width: 100%; height: 100%; display: block; }
.img-dots { position: absolute; bottom: 12px; left: 0; right: 0; display: flex; justify-content: center; gap: 6px; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: rgba(255,255,255,.35); }
.dot.on { background: #fff; width: 16px; border-radius: 3px; }
.img-back { position: absolute; top: 12px; left: 12px; width: 32px; height: 32px; border-radius: 50%; background: rgba(0,0,0,.35); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 16px; }

.price-card { margin: -8px 10px 8px; background: #fff; border-radius: 12px; padding: 14px; position: relative; z-index: 1; }
.price-row { display: flex; justify-content: space-between; align-items: flex-end; }
.price-left { display: flex; align-items: baseline; }
.price-symbol { font-size: 18px; font-weight: 700; color: var(--color-warning); }
.price-num { font-size: 30px; font-weight: 700; color: var(--color-warning); line-height: 1; transition: opacity .12s; }
.price-decimal { font-size: 16px; font-weight: 600; color: var(--color-warning); }
.price-tag { font-size: 11px; padding: 2px 8px; border-radius: 4px; font-weight: 600; }
.tag-green { background: var(--color-success); color: #fff; }
.tag-orange { background: var(--color-warning); color: #fff; }
.tag-red { background: var(--color-warning); color: #fff; }
.tag-gray { background: var(--color-text-secondary); color: #fff; }
.price-title { font-size: 16px; font-weight: 600; color: var(--color-text); display: block; margin-top: 8px; line-height: 1.4; }
.price-desc { font-size: 13px; color: var(--color-text-secondary); display: block; margin-top: 4px; line-height: 1.4; }

.spec-card { margin: 0 10px 8px; background: #fff; border-radius: 12px; padding: 14px; }
.spec-head { font-size: 14px; font-weight: 600; color: var(--color-text); margin-bottom: 10px; }
/* 参数逐行信息（Tigshop cart-item 风格） */
.spec-list { display: flex; flex-direction: column; }
.spec-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 22rpx 0; border-bottom: 1px solid var(--color-divider);
}
.spec-row:last-child { border-bottom: none; }
.spec-label { font-size: 13px; color: var(--color-text-secondary); }
.spec-val { font-size: 13px; color: var(--color-text); font-weight: 500; }

.shop-card { margin: 0 10px 8px; background: #fff; border-radius: 12px; padding: 12px 14px; display: flex; align-items: center; gap: 10px; }
.shop-ava { width: 40px; height: 40px; border-radius: 50%; background: linear-gradient(135deg,var(--color-primary-light),var(--color-primary-light)); display: flex; align-items: center; justify-content: center; font-size: 16px; font-weight: 600; color: var(--color-primary); flex-shrink: 0; }
.shop-info { flex: 1; }
.shop-name { font-size: 14px; font-weight: 500; display: block; }
.shop-status { font-size: 11px; color: var(--color-success); }
.shop-btn { padding: 6px 14px; border-radius: 14px; background: linear-gradient(135deg,var(--color-primary),var(--color-primary)); color: #fff; font-size: 13px; flex-shrink: 0; }

.detail-card { margin: 0 10px 8px; background: #fff; border-radius: 12px; padding: 14px; }
.detail-head { font-size: 14px; font-weight: 600; color: var(--color-text); margin-bottom: 8px; }
.detail-text { font-size: 13px; color: var(--color-text-secondary); line-height: 1.8; display: block; }

.bottom { position: fixed; bottom: 0; left: 0; right: 0; background: #fff; padding: 8px 12px; display: flex; align-items: center; gap: 8px; border-top: 1px solid var(--color-divider); box-shadow: 0 -2px 8px rgba(0,0,0,.03); }
.bottom-left { display: flex; gap: 2px; }
.bottom-fav, .bottom-share { width: 44px; display: flex; flex-direction: column; align-items: center; font-size: 10px; color: var(--color-text-secondary); line-height: 1.3; }
.bottom-contact {
  flex: 1.1; height: 38px; border-radius: 19px;
  background: #fff; border: 1px solid var(--color-primary); color: var(--color-primary);
  display: flex; align-items: center; justify-content: center; gap: 4px;
  font-size: 14px; font-weight: 500; transition: transform .18s;
}
.contact-ico { font-size: 15px; line-height: 1; }
.contact-txt { line-height: 1; }
.bottom-buy {
  flex: 1.1; height: 38px; border-radius: 19px;
  background: linear-gradient(135deg,var(--color-primary),var(--color-primary));
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 14px; font-weight: 600; transition: transform .18s;
}

/* ===== 按压反馈（hover-class，小程序标准方案） ===== */
.btn-press { transform: scale(.97); }
.back-press { transform: scale(.92); }
.shop-btn, .img-back { transition: transform .18s; }

/* ===== 收藏 / 分享交互动画 ===== */
.bottom-fav, .bottom-share { transition: transform .18s; }
.fav-ico { display: inline-block; font-size: 18px; line-height: 1.2; color: var(--color-text-secondary); transition: color .2s; }
.fav-active .fav-ico { color: var(--color-danger); }
.share-ico { display: inline-block; font-size: 16px; line-height: 1.2; }
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
@media (prefers-reduced-motion: reduce) {
  .heart-pop, .share-pop { animation: none; }
}
</style>
