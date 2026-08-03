<template>
<view class="page" v-if="product.id">
  <!-- 图区 -->
  <view class="img-box">
    <swiper v-if="images.length" :current="curImg" @change="e=>curImg=e.detail.current" circular>
      <swiper-item v-for="(img,i) in images" :key="i">
        <image :src="img" mode="aspectFill" class="img-swiper" @click="preview(i)" />
      </swiper-item>
    </swiper>
    <image v-else src="/static/home-bg.jpg" mode="aspectFill" class="img-swiper" />
    <view class="img-dots">
      <view v-for="(_,i) in (images.length||1)" :key="i" class="dot" :class="{on:i===curImg}" />
    </view>
    <view class="img-back" @tap="goBack">←</view>
  </view>

  <!-- 价格区 -->
  <view class="price-card">
    <view class="price-row">
      <view class="price-left">
        <text class="price-symbol">¥</text>
        <text class="price-num">{{ priceInt }}</text>
        <text class="price-decimal">.{{ priceDec }}</text>
      </view>
      <text class="price-tag" :class="tagClass">{{ product.condition || '商家发布' }}</text>
    </view>
    <text class="price-title">{{ product.title }}</text>
    <text class="price-desc" v-if="product.description">{{ product.description }}</text>
  </view>

  <!-- 规格区 -->
  <view class="spec-card">
    <view class="spec-head">商品参数</view>
    <view class="spec-grid">
      <view class="spec-item" v-if="product.brand">
        <text class="spec-label">品牌</text><text class="spec-val">{{ product.brand }}</text>
      </view>
      <view class="spec-item" v-if="product.model">
        <text class="spec-label">型号</text><text class="spec-val">{{ product.model }}</text>
      </view>
      <view class="spec-item">
        <text class="spec-label">成色</text><text class="spec-val">{{ product.condition === 'used' ? '二手' : (product.condition === 'new' ? '全新' : '商家发布') }}</text>
      </view>
      <view class="spec-item">
        <text class="spec-label">类型</text><text class="spec-val">{{ typeLabel(product.prod_type) }}</text>
      </view>
      <view class="spec-item">
        <text class="spec-label">浏览</text><text class="spec-val">{{ product.views || 0 }}次</text>
      </view>
      <view class="spec-item" v-if="product.seller_name">
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
    <text class="shop-btn">联系卖家</text>
  </view>

  <!-- 图文详情 -->
  <view class="detail-card">
    <view class="detail-head">商品详情</view>
    <view class="detail-body">
      <text class="detail-text">该商品由{{ product.seller_name || '认证商家' }}发布，支持{{ product.condition || '平台检测' }}。</text>
      <text class="detail-text" v-if="product.description">{{ product.description }}</text>
    </view>
  </view>

  <!-- 底部栏 -->
  <view class="bottom">
    <view class="bottom-left">
      <view class="bottom-fav" @tap="fav">♥ 收藏</view>
      <view class="bottom-share" @tap="share">↗ 分享</view>
    </view>
    <view class="bottom-cart" @tap="addCart">加入购物袋</view>
    <view class="bottom-buy" @tap="buy">立即购买</view>
  </view>
</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

const product = ref({})
const images = ref([])
const curImg = ref(0)

const priceInt = computed(() => Math.floor((product.value.price_fen||0) / 100))
const priceDec = computed(() => String((product.value.price_fen||0) % 100).padStart(2,'0'))
const tagClass = computed(() => {
  const c = product.value.condition || ''
  if (c === 'new') return 'tag-green'
  if (c === 'used') return 'tag-orange'
  return 'tag-gray'
})
const typeLabel = (t) => ({ drone: '整机', part: '配件', repair: '维修服务' }[t] || '商品')

onLoad((opts) => {
  if (!opts.id) return;
  (async () => {
    try {
      const p = await request({ url: '/api/v1/products/' + encodeURIComponent(opts.id) })
      if (p && p.id) {
        product.value = p
        try { images.value = typeof p.images === 'string' ? JSON.parse(p.images) : (p.images || []) } catch { images.value = [] }
      }
    } catch {}
  })()
})

const preview = (i) => {
  uni.previewImage({ current: i, urls: images.value })
}
const goBack = () => uni.navigateBack()
const contactShop = () => uni.showToast({ title: '联系卖家', icon: 'none' })
const fav = () => uni.showToast({ title: '已收藏', icon: 'success' })
const share = () => uni.showToast({ title: '分享', icon: 'none' })
const addCart = () => uni.showToast({ title: '已加入购物袋', icon: 'success' })
const buy = () => uni.showToast({ title: '下单功能开发中', icon: 'none' })
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: 60px; }

.img-box { position: relative; height: 320px; background: var(--color-text); }
.img-swiper { width: 100%; height: 100%; display: block; }
.img-dots { position: absolute; bottom: 12px; left: 0; right: 0; display: flex; justify-content: center; gap: 6px; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: rgba(255,255,255,.35); }
.dot.on { background: #fff; width: 16px; border-radius: 3px; }
.img-back { position: absolute; top: 12px; left: 12px; width: 32px; height: 32px; border-radius: 50%; background: rgba(0,0,0,.35); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 16px; }

.price-card { margin: -8px 10px 8px; background: #fff; border-radius: 12px; padding: 14px; position: relative; z-index: 1; }
.price-row { display: flex; justify-content: space-between; align-items: flex-end; }
.price-left { display: flex; align-items: baseline; }
.price-symbol { font-size: 18px; font-weight: 700; color: var(--color-warning); }
.price-num { font-size: 30px; font-weight: 700; color: var(--color-warning); line-height: 1; }
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
.spec-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px 16px; }
.spec-item { display: flex; justify-content: space-between; }
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
.bottom-cart { flex: 1; height: 38px; border-radius: 19px; background: #fff; border: 1px solid var(--color-warning); color: var(--color-warning); display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 500; }
.bottom-buy { flex: 1; height: 38px; border-radius: 19px; background: linear-gradient(135deg,var(--color-warning),var(--color-warning)); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 600; }
</style>
