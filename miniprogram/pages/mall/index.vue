<template>
  <Layout :current="1">
    <view class="mall-page">
      <!-- 顶部：搜索 + 标题 -->
      <view class="top-bar" :style="{ paddingTop: (statusBarH || 24) + 'px' }">
        <view class="search-box" @tap="focusSearch">
          <u-icon name="search" size="30rpx" color="#98A2B3" />
          <text class="search-hint">搜索无人机型号 / 配件</text>
        </view>
      </view>

      <!-- 分类轨道 -->
      <view class="cat-bar">
        <scroll-view scroll-x :show-scrollbar="false" class="cat-scroll">
          <view class="cat-inner">
            <view
              v-for="c in cats"
              :key="c.key"
              class="cat-tab"
              :class="{ on: activeCat === c.key }"
              @tap="selectCat(c.key)"
            >{{ c.name }}</view>
          </view>
        </scroll-view>
      </view>

      <!-- 加载态 -->
      <view v-if="loading" class="grid">
        <view v-for="i in 6" :key="i" class="card skeleton">
          <view class="s-img" />
          <view class="s-line" />
          <view class="s-line short" />
        </view>
      </view>

      <!-- 空态 -->
      <view v-else-if="!errorMsg && products.length === 0" class="state-box">
        <u-empty description="暂无在售商品" />
        <text class="state-note">商品即将上架，或切换分类查看</text>
      </view>

      <!-- 失败态 -->
      <view v-else-if="errorMsg" class="state-box">
        <u-empty description="商品列表加载失败" />
        <u-button type="primary" size="small" round @click="loadProducts(activeCat)">重新加载</u-button>
      </view>

      <!-- 商品双列 -->
      <view v-else class="grid">
        <view
          v-for="p in products"
          :key="p.id"
          class="card"
          @tap="goDetail(p.id)"
        >
          <view class="img-wrap">
            <image v-if="imgSrc(p)" :src="imgSrc(p)" mode="aspectFill" class="card-img" />
            <view v-else class="img-ph">
              <u-icon name="plus" size="36rpx" color="#0A66C2" />
            </view>
            <text v-if="p.condition" class="tag" :class="p.condition === 'used' ? 'tag-used' : 'tag-new'">
              {{ p.condition === 'used' ? '二手' : '全新' }}
            </text>
          </view>
          <view class="card-body">
            <text class="card-title">{{ p.title }}</text>
            <text v-if="p.brand || p.model" class="card-desc">{{ p.brand || '' }}{{ p.brand && p.model ? ' · ' : '' }}{{ p.model || '' }}</text>
            <view class="card-foot">
              <text class="price">¥{{ fmt(p.price_fen) }}</text>
              <text class="seller">{{ p.seller_name || '' }}</text>
            </view>
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

const statusBarH = ref(24)
const products = ref([])
const loading = ref(false)
const errorMsg = ref('')
const activeCat = ref('')

const cats = [
  { key: '', name: '全部' },
  { key: 'drone', name: '整机' },
  { key: 'part', name: '配件' },
  { key: 'repair', name: '维修服务' },
]

const selectCat = (key) => {
  activeCat.value = key
  loadProducts(key)
}

const loadProducts = async (cat) => {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({
      url: '/api/v1/products',
      data: cat ? { prod_type: cat } : {},
    })
    products.value = Array.isArray(res) ? res : []
  } catch (e) {
    errorMsg.value = '加载失败'
  } finally {
    loading.value = false
  }
}

const focusSearch = () => {
  uni.navigateTo({ url: '/pages/search/index' })
}
const goDetail = (id) => {
  uni.navigateTo({ url: '/pages/mall/detail?id=' + encodeURIComponent(id) })
}
const imgSrc = (p) => {
  try {
    const arr = typeof p.images === 'string' ? JSON.parse(p.images) : p.images
    if (arr && arr[0]) return arr[0]
  } catch (e) {}
  return ''
}
const fmt = (f) => (f ? (f / 100).toLocaleString('en-US') : '0')

onMounted(() => {
  try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 24 } catch (e) {}
  loadProducts('')
})
</script>

<style scoped>
.mall-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 20px; }

/* 顶部搜索 */
.top-bar { background: var(--color-primary); padding: 8px 12px 12px; }
.search-box {
  display: flex; align-items: center; gap: 8px;
  height: 40px; background: #fff; border-radius: 8px; padding: 0 12px;
}
.search-hint { font-size: 26rpx; color: #98A2B3; }

/* 分类轨道 */
.cat-bar { background: #fff; padding: 10px 0; border-bottom: 1rpx solid var(--color-divider); }
.cat-scroll { white-space: nowrap; }
.cat-inner { display: inline-flex; padding: 0 12px; gap: 8px; }
.cat-tab {
  height: 30px; padding: 0 14px; border-radius: 6px;
  border: 1rpx solid var(--color-border);
  display: flex; align-items: center;
  font-size: 24rpx; color: var(--color-text-secondary);
  background: #fff;
}
.cat-tab.on { background: var(--color-primary); border-color: var(--color-primary); color: #fff; font-weight: 600; }

/* 商品双列 */
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; padding: 12px; }
.card { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 3px 12px rgba(16,24,40,0.05); }
.img-wrap { position: relative; aspect-ratio: 4/3; background: var(--color-primary-light); }
.card-img { width: 100%; height: 100%; }
.img-ph { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; }
.tag {
  position: absolute; left: 8px; top: 8px;
  font-size: 20rpx; padding: 2px 8px; border-radius: 4px;
}
.tag-new { background: var(--color-primary-light); color: var(--color-primary); }
.tag-used { background: var(--color-primary-light); color: var(--color-text-secondary); }
.card-body { padding: 8px 10px 10px; }
.card-title {
  font-size: 26rpx; font-weight: 700; color: var(--color-text); line-height: 1.4;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.card-desc { display: block; font-size: 22rpx; color: var(--color-text-secondary); margin: 6px 0 4px; }
.card-foot { display: flex; align-items: baseline; justify-content: space-between; gap: 4px; }
.price { font-size: 28rpx; font-weight: 700; color: var(--color-warning); white-space: nowrap; }
.seller { font-size: 20rpx; color: var(--color-text-placeholder); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 50%; }

/* 骨架 */
.skeleton .s-img { aspect-ratio: 4/3; background: var(--color-divider); }
.skeleton .s-line { height: 24rpx; background: var(--color-divider); border-radius: 4px; margin: 10px 10px 0; }
.skeleton .s-line.short { width: 60%; margin-bottom: 12px; }

/* 状态 */
.state-box { padding: 100rpx 40rpx; display: flex; flex-direction: column; align-items: center; gap: 16rpx; }
.state-note { font-size: 22rpx; color: var(--color-text-placeholder); }
</style>
