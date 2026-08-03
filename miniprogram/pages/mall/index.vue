<template>
  <Layout :current="1">
    <view class="mall-page">
      <!-- 顶部：搜索 -->
      <view class="top-bar" :style="{ paddingTop: (statusBarH || 24) + 'px' }">
        <view class="search-box">
          <u-icon name="search" size="30rpx" color="#98A2B3" />
          <input
            class="search-input"
            v-model="keyword"
            placeholder="搜索型号 / 品牌 / 配件"
            placeholder-class="search-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <text v-if="keyword" class="search-clear" @tap="clearSearch">×</text>
        </view>
      </view>

      <!-- 左右布局：左分类轨 + 右商品区 -->
      <view class="cate-layout">
        <!-- 左分类栏 -->
        <scroll-view scroll-y class="cate-side">
          <view
            v-for="c in cats"
            :key="c.key"
            class="cate-item"
            :class="{ on: activeCat === c.key }"
            @tap="selectCat(c.key)"
          >
            <view class="cate-icon" :class="{ on: activeCat === c.key }">{{ c.icon }}</view>
            <text class="cate-name">{{ c.name }}</text>
          </view>
        </scroll-view>

        <!-- 右商品区 -->
        <scroll-view scroll-y class="goods-area" :show-scrollbar="false">
          <!-- 加载态 -->
          <view v-if="loading" class="grid">
            <view v-for="i in 4" :key="i" class="card skeleton">
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

          <!-- 商品列表 -->
          <view v-else class="goods-list">
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
                <view class="card-contact" hover-class="btn-press" @tap.stop="goDetail(p.id)">
                  <text class="contact-ico">电</text>
                  <text class="contact-txt">联系</text>
                </view>
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
            <view v-if="loadingMore" class="more-tip">
              <u-loading size="24rpx" />
              <text>加载中...</text>
            </view>
            <view v-else-if="!hasMore && products.length > 0" class="more-tip">没有更多了</view>
          </view>
        </scroll-view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { onReachBottom } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request, BASE_URL } from '@/utils/request'

const statusBarH = ref(24)
const products = ref([])
const loading = ref(false)
const loadingMore = ref(false)
const hasMore = ref(true)
const errorMsg = ref('')
const activeCat = ref('')
const page = ref(1)

const cats = [
  { key: '', name: '全部', icon: '全' },
  { key: 'drone', name: '整机', icon: '机' },
  { key: 'part', name: '配件', icon: '配' },
  { key: 'repair', name: '维修', icon: '修' },
]

const keyword = ref('')

const selectCat = (key) => {
  activeCat.value = key
  page.value = 1
  hasMore.value = true
  loadProducts(key, keyword.value)
}

const onSearch = () => {
  page.value = 1
  hasMore.value = true
  loadProducts(activeCat.value, keyword.value.trim())
}

const clearSearch = () => {
  keyword.value = ''
  page.value = 1
  hasMore.value = true
  loadProducts(activeCat.value, '')
}

const loadProducts = async (cat, kw) => {
  loading.value = true
  errorMsg.value = ''
  try {
    const data = { page: 1, page_size: 10 }
    if (cat) data.prod_type = cat
    if (kw) data.keyword = kw
    const res = await request({ url: '/api/v1/products', data })
    products.value = Array.isArray(res) ? res : []
    hasMore.value = products.value.length >= (res?.total || products.value.length)
  } catch (e) {
    errorMsg.value = '加载失败'
  } finally {
    loading.value = false
  }
}

// 触底加载更多
const loadMore = async () => {
  if (loading.value || loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const data = { page: page.value + 1, page_size: 10 }
    if (activeCat.value) data.prod_type = activeCat.value
    if (keyword.value.trim()) data.keyword = keyword.value.trim()
    const res = await request({ url: '/api/v1/products', data })
    const list = Array.isArray(res) ? res : []
    products.value = products.value.concat(list)
    page.value += 1
    hasMore.value = list.length >= 10
  } catch (e) {
    /* 静默：下次触底重试 */
  } finally {
    loadingMore.value = false
  }
}

// 图片 URL：相对路径（/uploads/...）拼后端地址，完整 URL 原样
const fullUrl = (u) => (u && u.startsWith('http') ? u : BASE_URL + (u || ''))
const imgSrc = (p) => {
  try {
    const arr = typeof p.images === 'string' ? JSON.parse(p.images) : p.images
    if (arr && arr[0]) return fullUrl(arr[0])
  } catch (e) {}
  return ''
}
const fmt = (f) => (f ? (f / 100).toLocaleString('en-US') : '0')

const goDetail = (id) => {
  uni.navigateTo({ url: '/pages/mall/detail?id=' + encodeURIComponent(id) })
}

onMounted(() => {
  try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 24 } catch (e) {}
  loadProducts('')
})

onReachBottom(loadMore)
</script>

<style scoped>
.mall-page { height: 100vh; display: flex; flex-direction: column; background: var(--color-bg); }

/* 顶部搜索 */
.top-bar { background: var(--color-primary); padding: 8px 12px 12px; flex-shrink: 0; }
.search-box {
  display: flex; align-items: center; gap: 8px;
  height: 40px; background: #fff; border-radius: 8px; padding: 0 12px;
}
.search-input { flex: 1; font-size: 26rpx; color: var(--color-text); }
.search-ph { color: #98A2B3; }
.search-clear { width: 32rpx; height: 32rpx; display: flex; align-items: center; justify-content: center; color: #c8c9cc; font-size: 28rpx; }

/* 左右布局 */
.cate-layout { flex: 1; display: flex; min-height: 0; }

/* 左分类栏（Tigshop 分类页风格） */
.cate-side { width: 88px; background: var(--color-bg); flex-shrink: 0; height: 100%; }
.cate-item {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 18rpx 0; gap: 6rpx;
}
.cate-icon {
  width: 48rpx; height: 48rpx; border-radius: 12rpx;
  background: var(--color-primary-light); color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-size: 22rpx; font-weight: 600;
}
.cate-item.on { background: #fff; position: relative; }
.cate-item.on::before { content: ''; position: absolute; left: 0; top: 30%; bottom: 30%; width: 6rpx; border-radius: 3rpx; background: var(--color-primary); }
.cate-item.on .cate-icon { background: var(--color-primary); color: #fff; }
.cate-name { font-size: 22rpx; color: var(--color-text-secondary); }
.cate-item.on .cate-name { color: var(--color-primary); font-weight: 600; }

/* 右商品区 */
.goods-area { flex: 1; height: 100%; min-width: 0; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; padding: 12px; }
.goods-list { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; padding: 12px; }

/* 商品卡（Tigshop 风格：图内边距 + 联系快捷按钮） */
.card { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 3px 12px rgba(16,24,40,0.05); }
.img-wrap { position: relative; background: #fff; padding: 10rpx; }
.card-img { width: 100%; aspect-ratio: 1/1; border-radius: 6rpx; display: block; }
.img-ph { width: 100%; aspect-ratio: 1/1; border-radius: 6rpx; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; }
.tag {
  position: absolute; left: 18rpx; top: 18rpx;
  font-size: 20rpx; padding: 2px 8px; border-radius: 4px;
}
.tag-new { background: var(--color-primary-light); color: var(--color-primary); }
.tag-used { background: var(--color-primary-light); color: var(--color-text-secondary); }
/* 卡上联系按钮（悬浮图右下角） */
.card-contact {
  position: absolute; right: 18rpx; bottom: 18rpx;
  display: flex; align-items: center; gap: 4rpx;
  background: var(--color-primary); color: #fff;
  padding: 8rpx 14rpx; border-radius: 6rpx;
}
.contact-ico { font-size: 20rpx; line-height: 1; }
.contact-txt { font-size: 20rpx; line-height: 1; }
.btn-press { transform: scale(.95); }

.card-body { padding: 0 12rpx 12rpx; }
.card-title {
  font-size: 26rpx; font-weight: 700; color: var(--color-text); line-height: 1.4;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.card-desc { display: block; font-size: 22rpx; color: var(--color-text-secondary); margin: 6px 0 4px; }
.card-foot { display: flex; align-items: baseline; justify-content: space-between; gap: 4px; }
.price { font-size: 28rpx; font-weight: 700; color: var(--color-warning); white-space: nowrap; }
.seller { font-size: 20rpx; color: var(--color-text-placeholder); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 50%; }

/* 骨架 */
.skeleton .s-img { aspect-ratio: 1/1; background: var(--color-divider); margin: 10rpx; border-radius: 6rpx; }
.skeleton .s-line { height: 24rpx; background: var(--color-divider); border-radius: 4px; margin: 10px 10px 0; }
.skeleton .s-line.short { width: 60%; margin-bottom: 12px; }

/* 状态 */
.state-box { padding: 100rpx 40rpx; display: flex; flex-direction: column; align-items: center; gap: 16rpx; }
.state-note { font-size: 22rpx; color: var(--color-text-placeholder); }

/* 加载更多提示 */
.more-tip {
  grid-column: 1 / -1;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 20rpx 0;
  font-size: 22rpx; color: var(--color-text-placeholder);
}
</style>
