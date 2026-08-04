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
            placeholder="搜索整机 / 航拍 / 试飞场地"
            placeholder-class="search-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <text v-if="keyword" class="search-clear" @tap="clearSearch">×</text>
        </view>
      </view>

      <!-- 三大分区（设备/服务/场地 — 供给能力展示） -->
      <view class="zones">
        <view
          v-for="z in zones"
          :key="z.key"
          class="zone"
          :class="{ on: activeZone === z.key }"
          @tap="switchZone(z.key)"
        >{{ z.name }}</view>
      </view>

      <scroll-view scroll-y class="content-area" :show-scrollbar="false">
        <!-- 加载态 -->
        <view v-if="loading" class="skel-list">
          <view v-for="i in 4" :key="i" class="skel-row">
            <view class="skel-img" />
            <view class="skel-body">
              <view class="skel-line w60" />
              <view class="skel-line" />
            </view>
          </view>
        </view>

        <!-- 空态 -->
        <view v-else-if="!errorMsg && zoneItems.length === 0" class="state-box">
          <u-empty description="该分区暂无供给" />
          <text class="state-note">会员供给上架中，或切换其他分区</text>
          <u-button type="primary" size="small" round @tap="goDemand">去需求大厅发需求</u-button>
        </view>

        <!-- 失败态 -->
        <view v-else-if="errorMsg" class="state-box">
          <u-empty description="供给列表加载失败" />
          <u-button type="primary" size="small" round @click="loadAll">重新加载</u-button>
        </view>

        <!-- 设备分区：双列商品卡 -->
        <view v-else-if="activeZone === 'device'" class="grid">
          <view v-for="p in zoneItems" :key="p.id" class="prod" @tap="goDetail(p.id)">
            <view class="prod-img">
              <image v-if="imgSrc(p)" :src="imgSrc(p)" mode="aspectFill" class="prod-img-inner" />
              <view v-else class="prod-img-ph"><u-icon name="plus" size="36rpx" color="#0A66C2" /></view>
              <text class="prod-tag">{{ typeShort(p.prod_type) }}</text>
            </view>
            <view class="prod-body">
              <text class="prod-title">{{ p.title }}</text>
              <view class="prod-foot">
                <text class="prod-price">¥{{ fmt(p.price_fen) }}</text>
                <text v-if="p.seller_name" class="prod-cert">{{ p.seller_name }}</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 服务分区：横卡（对接） -->
        <view v-else-if="activeZone === 'service'" class="hlist">
          <view v-for="p in zoneItems" :key="p.id" class="hcard" @tap="goDetail(p.id)">
            <view class="hcard-img">
              <image v-if="imgSrc(p)" :src="imgSrc(p)" mode="aspectFill" class="hcard-img-inner" />
              <view v-else class="hcard-img-ph"><u-icon name="plus" size="30rpx" color="#E96012" /></view>
            </view>
            <view class="hcard-info">
              <text class="hcard-title">{{ p.title }}</text>
              <text class="hcard-type">{{ typeLabel(p.prod_type) }}</text>
              <view class="hcard-foot">
                <text class="hcard-price">{{ p.price_fen ? '¥' + fmt(p.price_fen) : '面议' }}</text>
                <text class="hcard-cta">联系对接 ›</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 场地分区：横卡（预约） -->
        <view v-else class="hlist">
          <view v-for="p in zoneItems" :key="p.id" class="hcard" @tap="goDetail(p.id)">
            <view class="hcard-img">
              <image v-if="imgSrc(p)" :src="imgSrc(p)" mode="aspectFill" class="hcard-img-inner" />
              <view v-else class="hcard-img-ph"><u-icon name="location" size="30rpx" color="#0A66C2" /></view>
            </view>
            <view class="hcard-info">
              <text class="hcard-title">{{ p.title }}</text>
              <text class="hcard-type">{{ typeLabel(p.prod_type) }}</text>
              <view class="hcard-foot">
                <text class="hcard-price">{{ p.price_fen ? '¥' + fmt(p.price_fen) + ' /小时' : '面议' }}</text>
                <text class="hcard-btn">预约 ›</text>
              </view>
            </view>
          </view>
        </view>

        <view v-if="!loading && !errorMsg && zoneItems.length" class="more-tip">— 已加载全部 —</view>
      </scroll-view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onReachBottom } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request, BASE_URL } from '@/utils/request'

const statusBarH = ref(24)
const products = ref([])
const loading = ref(false)
const errorMsg = ref('')
const keyword = ref('')
const activeZone = ref('device')

const zones = [
  { key: 'device', name: '设备' },
  { key: 'service', name: '服务' },
  { key: 'site', name: '场地' },
]
// 分区 → prod_type 映射（需求②-2 供给能力展示 7 类）
const zoneTypes = {
  device: ['drone', 'part'],
  service: ['aerial', 'calibration', 'airspace', 'repair'],
  site: ['test_fly'],
}

const typeLabel = (t) => ({ drone: '整机', part: '零部件', repair: '维修服务', aerial: '航拍服务', test_fly: '试飞测试', calibration: '检测标定', airspace: '空域协调' }[t] || t || '')
const typeShort = (t) => ({ drone: '整机', part: '配件', test_fly: '试飞' }[t] || typeLabel(t))

// 当前分区可见供给（按分区类型 + 关键词过滤）
const zoneItems = computed(() => {
  const types = zoneTypes[activeZone.value] || []
  const kw = keyword.value.trim().toLowerCase()
  return products.value.filter(p => {
    if (!types.includes(p.prod_type)) return false
    if (kw && !((p.title || '').toLowerCase().includes(kw) || (p.brand || '').toLowerCase().includes(kw))) return false
    return true
  })
})

const switchZone = (key) => { activeZone.value = key }

const onSearch = () => { /* 前端过滤 zoneItems 即时生效 */ }
const clearSearch = () => { keyword.value = '' }

const loadAll = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/products', data: { page: 1, page_size: 50 } })
    products.value = Array.isArray(res) ? res : []
  } catch (e) {
    errorMsg.value = '加载失败'
  } finally {
    loading.value = false
  }
}

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
const goDemand = () => {
  uni.navigateTo({ url: '/pages/demands/list' })
}

onMounted(() => {
  try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 24 } catch (e) {}
  loadAll()
})

onReachBottom(() => { /* 全量加载，无需翻页 */ })
</script>

<style scoped>
.mall-page { height: 100vh; display: flex; flex-direction: column; background: var(--color-bg); }

/* 顶部搜索 */
.top-bar { background: var(--color-primary); padding: 8px 12px 12px; flex-shrink: 0; }
.search-box {
  display: flex; align-items: center; gap: 8px;
  height: 40px; background: #fff; border-radius: 8px; padding: 0 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,.12);
}
.search-input { flex: 1; font-size: 26rpx; color: var(--color-text); }
.search-ph { color: #98A2B3; }
.search-clear { width: 32rpx; height: 32rpx; display: flex; align-items: center; justify-content: center; color: #c8c9cc; font-size: 28rpx; }

/* 三大分区 tab */
.zones { display: flex; background: #fff; border-bottom: 1rpx solid var(--color-divider); flex-shrink: 0; }
.zone {
  flex: 1; text-align: center; padding: 28rpx 0 24rpx;
  font-size: 28rpx; color: var(--color-text-secondary); position: relative; font-weight: 500;
}
.zone.on { color: var(--color-primary); font-weight: 700; }
.zone.on::after {
  content: ''; position: absolute; left: 50%; transform: translateX(-50%);
  bottom: 0; width: 64rpx; height: 6rpx; border-radius: 3rpx; background: var(--color-primary);
}

/* 内容区 */
.content-area { flex: 1; min-height: 0; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; padding: 12px; }
.hlist { padding: 12px; }

/* 设备卡（双列） */
.prod { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 3px 12px rgba(16,24,40,0.05); }
.prod-img { aspect-ratio: 4/3; position: relative; background: #fff; }
.prod-img-inner { width: 100%; height: 100%; }
.prod-img-ph { width: 100%; height: 100%; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; }
.prod-tag {
  position: absolute; left: 8px; top: 8px; z-index: 1;
  font-size: 18rpx; padding: 2px 8px; border-radius: 4px;
  background: rgba(255,255,255,.9); color: var(--color-primary); font-weight: 600;
}
.prod-body { padding: 8px 10px 10px; }
.prod-title { font-size: 24rpx; font-weight: 700; color: var(--color-text); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; min-height: 66rpx; }
.prod-foot { display: flex; justify-content: space-between; align-items: center; margin-top: 6rpx; }
.prod-price { font-size: 26rpx; font-weight: 700; color: var(--color-warning); }
.prod-cert { font-size: 18rpx; color: var(--color-success); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 55%; }

/* 服务/场地横卡 */
.hcard { display: flex; gap: 12px; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 3px 12px rgba(16,24,40,0.05); margin-bottom: 10px; }
.hcard-img { width: 252rpx; height: 168rpx; position: relative; flex-shrink: 0; }
.hcard-img-inner { width: 100%; height: 100%; }
.hcard-img-ph { width: 100%; height: 100%; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; }
.hcard-info { flex: 1; padding: 16rpx 20rpx; min-width: 0; display: flex; flex-direction: column; }
.hcard-title { font-size: 26rpx; font-weight: 700; color: var(--color-text); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.hcard-type { display: inline-block; font-size: 20rpx; padding: 2px 8px; border-radius: 4px; background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; margin-top: 8rpx; align-self: flex-start; }
.hcard-foot { display: flex; justify-content: space-between; align-items: center; margin-top: auto; padding-top: 8rpx; }
.hcard-price { font-size: 22rpx; color: var(--color-text-secondary); }
.hcard-price { color: var(--color-warning); font-weight: 700; }
.hcard-cta { font-size: 22rpx; color: var(--color-primary); font-weight: 600; }
.hcard-btn { font-size: 22rpx; color: #fff; background: var(--color-primary); padding: 6rpx 20rpx; border-radius: 6px; font-weight: 600; }

/* 骨架 */
.skel-list { padding: 12px; }
.skel-row { display: flex; gap: 12px; background: #fff; border-radius: 8px; padding: 12px; margin-bottom: 10px; }
.skel-img { width: 252rpx; height: 168rpx; border-radius: 6px; background: var(--color-divider); flex-shrink: 0; }
.skel-body { flex: 1; padding-top: 4px; }
.skel-line { height: 24rpx; border-radius: 4px; background: var(--color-divider); margin-bottom: 12rpx; }
.skel-line.w60 { width: 60%; }

/* 状态 */
.state-box { padding: 100rpx 40rpx; display: flex; flex-direction: column; align-items: center; gap: 16rpx; }
.state-note { font-size: 22rpx; color: var(--color-text-placeholder); }
.more-tip { text-align: center; font-size: 22rpx; color: var(--color-text-placeholder); padding: 16rpx 0; }
</style>
