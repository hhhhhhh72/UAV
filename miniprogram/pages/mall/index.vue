<template>
  <Layout :current="1">
    <view class="mall-page">
      <!-- 顶部导航：标题 + 搜索（对齐服务页原型） -->
      <view class="top-bar" :style="{ paddingTop: (statusBarH || 24) + 'px' }">
        <view class="nav-row">
          <view class="nav-title-block">
            <text class="nav-title">服务</text>
            <text class="nav-sub">供给能力展示</text>
          </view>
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
              <view class="skel-line w60" />
            </view>
          </view>
        </view>

        <!-- 空态（整个分区无供给） -->
        <view v-else-if="!errorMsg && allEmpty" class="state-box">
          <view class="s-ico"><view class="s-box" /></view>
          <text class="s-title">{{ emptyTitle }}</text>
          <text v-if="!keyword" class="state-note">会员供给上架中，或切换其他分区</text>
          <u-button v-if="!keyword" type="primary" size="small" round @tap="goDemand">去需求大厅发需求</u-button>
        </view>

        <!-- 失败态 -->
        <view v-else-if="errorMsg" class="state-box">
          <view class="s-ico"><view class="s-fail" /></view>
          <text class="s-title">供给列表加载失败</text>
          <text class="state-note">网络开小差了，请检查后重试</text>
          <u-button type="primary" size="small" round @tap="loadAll">重新加载</u-button>
        </view>

        <!-- 有数据：按分区渲染 -->
        <template v-else>
          <!-- 设备分区：整机/零件 二级切换 + 双列商品卡 -->
          <view v-if="activeZone === 'device'" class="group">
            <view class="sub-tabs">
              <view
                v-for="t in deviceTabs"
                :key="t.key"
                class="st"
                :class="{ on: deviceTab === t.key }"
                @tap="deviceTab = t.key"
              >
                <text>{{ t.title }}</text>
                <text class="st-count">{{ typeCount(t.types) }}</text>
              </view>
            </view>

            <!-- 当前子分类空（另一分类可能有数据） -->
            <view v-if="deviceEmpty" class="state-box mini">
              <text class="s-title">{{ keyword ? '未找到相关供给' : '该分类暂无供给' }}</text>
              <text v-if="!keyword" class="state-note">会员供给上架中，或切换其他分类</text>
            </view>

            <view v-else class="grid">
              <view v-for="p in deviceItems" :key="p.id" class="prod" @tap="goDetail(p.id)">
                <view class="prod-img">
                  <image v-if="imgSrc(p)" :src="imgSrc(p)" mode="aspectFill" class="prod-img-inner" />
                  <view v-else class="prod-img-ph"><u-icon name="plus" size="36rpx" color="#0A66C2" /></view>
                  <text class="prod-tag">{{ typeShort(p.prod_type) }}</text>
                </view>
                <view class="prod-body">
                  <text class="prod-title">{{ p.title }}</text>
                  <view class="prod-meta">
                    <text class="prod-meta-l">{{ metaText(p) }}</text>
                    <text v-if="p.views" class="prod-views">{{ p.views }} 次浏览</text>
                  </view>
                  <view class="prod-foot">
                    <text class="prod-price">¥{{ fmt(p.price_fen) }}</text>
                    <text v-if="condLabel(p)" class="prod-cert">{{ condLabel(p) }}</text>
                  </view>
                </view>
              </view>
            </view>
          </view>

          <!-- 服务/场地分区：按子类型分组 -->
          <template v-else>
            <view v-for="g in zoneGroups" :key="g.key" class="group">
              <view class="zone-head">
                <view class="zone-dot" />
                <text class="zone-title">{{ g.title }}</text>
                <text class="zone-more" @tap="toastMore(g.title)">查看全部 ›</text>
              </view>

              <!-- 服务分区：横卡（联系对接） -->
              <view v-if="activeZone === 'service'" class="hlist">
                <template v-for="p in g.items" :key="p.id">
                  <view class="hcard" @tap="goDetail(p.id)">
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
                </template>
              </view>

              <!-- 场地分区：横卡（预约） -->
              <view v-else class="hlist">
                <template v-for="p in g.items" :key="p.id">
                  <view class="hcard" @tap="goDetail(p.id)">
                    <view class="hcard-img">
                      <image v-if="imgSrc(p)" :src="imgSrc(p)" mode="aspectFill" class="hcard-img-inner" />
                      <view v-else class="hcard-img-ph"><u-icon name="location" size="30rpx" color="#0A66C2" /></view>
                    </view>
                    <view class="hcard-info">
                      <text class="hcard-title">{{ p.title }}</text>
                      <view class="site-meta">
                        <text v-if="p.brand || p.model">{{ p.brand || p.model }}</text>
                        <text v-if="condLabel(p)">{{ condLabel(p) }}</text>
                      </view>
                      <view class="hcard-foot">
                        <text class="hcard-price">{{ p.price_fen ? '¥' + fmt(p.price_fen) + ' /小时' : '面议' }}</text>
                        <text class="hcard-btn" @tap.stop="goBooking">预约 ›</text>
                      </view>
                    </view>
                  </view>
                </template>
              </view>
            </view>
          </template>
          <view class="more-tip">— 已加载全部 —</view>
        </template>
      </scroll-view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onReachBottom } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { request, BASE_URL } from '@/utils/request'
import { productTypeLabel as typeLabel, productTypeShort as typeShort } from '@/utils/enums'

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

// 设备分区二级切换：整机 / 零件
const deviceTab = ref('drone')
const deviceTabs = [
  { key: 'drone', title: '整机', types: ['drone'] },
  { key: 'part', title: '零件', types: ['part'] },
]

// 服务/场地分区 → 子类型分组（需求②-2 供给能力展示，对齐服务页原型）
// 服务/场地分区 → 子类型分组（需求②-2 供给能力展示；培训归金刚区→培训认证模块，不在此重复）
const groupDefs = {
  service: [
    { key: 'aerial', title: '航拍服务', types: ['aerial'] },
    { key: 'calibration', title: '检测标定', types: ['calibration'] },
    { key: 'airspace', title: '空域协调', types: ['airspace'] },
    { key: 'repair', title: '维修服务', types: ['repair'] },
  ],
  site: [{ key: 'test_fly', title: '试飞测试场地', types: ['test_fly'] }],
}

const condLabel = (p) => (p.condition === 'used' ? '已检修' : p.condition === 'new' ? '全新' : p.seller_name || '')

const kwMatch = (p, kw) => !kw || (p.title || '').toLowerCase().includes(kw) || (p.brand || '').toLowerCase().includes(kw)

// 设备分区：当前子分类可见供给
const deviceItems = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  const t = deviceTabs.find((t) => t.key === deviceTab.value)
  return products.value.filter((p) => t && t.types.includes(p.prod_type) && kwMatch(p, kw))
})
const deviceEmpty = computed(() => deviceItems.value.length === 0)

// 服务/场地分区：各子类型分组（关键词即时过滤，空组隐藏）
const zoneGroups = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  const defs = groupDefs[activeZone.value] || []
  return defs
    .map((g) => ({ ...g, items: products.value.filter((p) => g.types.includes(p.prod_type) && kwMatch(p, kw)) }))
    .filter((g) => g.items.length)
})

const typeCount = (types) => products.value.filter((p) => types.includes(p.prod_type)).length
const allEmpty = computed(() =>
  activeZone.value === 'device'
    ? deviceTabs.every((t) => typeCount(t.types) === 0)
    : zoneGroups.value.length === 0,
)
const emptyTitle = computed(() => (keyword.value.trim() ? '未找到相关供给' : '该分区暂无供给'))
const metaText = (p) => [p.brand, p.model].filter(Boolean).join(' · ') || p.seller_name || ''

const switchZone = (key) => { activeZone.value = key }
const onSearch = () => { /* 前端即时过滤，列表自动生效 */ }
const clearSearch = () => { keyword.value = '' }
const toastMore = (t) => { uni.showToast({ title: t + '已全部展示', icon: 'none' }) }

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
// 场地卡「预约」→ 场地预约页（①台账 + ③预约 同一页面承载，跳过商品详情缩短路径）
const goBooking = () => {
  uni.navigateTo({ url: '/pages/testsites/book' })
}

onMounted(() => {
  try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 24 } catch (e) {}
  loadAll()
})

onReachBottom(() => { /* 全量加载，无需翻页 */ })
</script>

<style scoped>
.mall-page { height: 100vh; display: flex; flex-direction: column; background: var(--color-bg); }

/* 顶部导航：深蓝底 + 标题 + 搜索 */
.top-bar {
  background: var(--color-primary-deep);
  padding: 8px 12px 12px;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
}
.top-bar::after {
  content: ''; position: absolute; top: -60%; right: -10%;
  width: 360rpx; height: 360rpx; border-radius: 50%;
  background: radial-gradient(circle, rgba(29,212,168,.12), transparent 65%);
  pointer-events: none;
}
.nav-row { display: flex; align-items: center; gap: 16rpx; position: relative; z-index: 1; }
.nav-title-block { flex-shrink: 0; display: flex; flex-direction: column; }
.nav-title { color: #fff; font-size: 34rpx; font-weight: 700; line-height: 1.3; }
.nav-sub { font-size: 20rpx; color: rgba(255,255,255,.65); line-height: 1.4; }
.search-box {
  flex: 1; display: flex; align-items: center; gap: 8rpx; min-width: 0;
  height: 40px; background: #fff; border-radius: 6px; padding: 0 12rpx;
  box-shadow: 0 2px 8px rgba(0,0,0,.12);
}
.search-input { flex: 1; font-size: 24rpx; color: var(--color-text); }
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
.group { padding: 0 12px; }

/* 设备分区二级切换（整机/零件） */
.sub-tabs { display: flex; gap: 16rpx; margin: 20rpx 0; }
.st {
  flex: 1; height: 56rpx; display: flex; align-items: center; justify-content: center; gap: 8rpx;
  border-radius: 6px; background: var(--color-divider); color: var(--color-text-secondary);
  font-size: 26rpx; font-weight: 600;
}
.st.on { background: var(--color-primary); color: #fff; }
.st-count { font-size: 20rpx; opacity: .75; }

/* 分区组标题 */
.zone-head { display: flex; align-items: center; gap: 12rpx; margin: 20rpx 0 16rpx; }
.zone-dot { width: 16rpx; height: 16rpx; border-radius: 4rpx; background: var(--color-primary); }
.zone-title { font-size: 26rpx; font-weight: 700; color: var(--color-text); }
.zone-more { font-size: 22rpx; color: var(--color-text-placeholder); margin-left: auto; }

/* 设备卡（双列，详细版） */
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20rpx; }
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
.prod-meta { display: flex; justify-content: space-between; align-items: center; gap: 8rpx; font-size: 20rpx; color: var(--color-text-secondary); margin-top: 6rpx; }
.prod-meta-l { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
.prod-views { flex-shrink: 0; }
.prod-foot { display: flex; justify-content: space-between; align-items: center; margin-top: 6rpx; }
.prod-price { font-size: 26rpx; font-weight: 700; color: var(--color-accent-deep); }
.prod-cert { font-size: 18rpx; color: var(--color-success); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 55%; }

/* 服务/场地横卡 */
.hlist { margin-bottom: 12rpx; }
.hcard { display: flex; gap: 12px; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 3px 12px rgba(16,24,40,0.05); margin-bottom: 20rpx; }
.hcard-img { width: 252rpx; height: 168rpx; position: relative; flex-shrink: 0; }
.hcard-img-inner { width: 100%; height: 100%; }
.hcard-img-ph { width: 100%; height: 100%; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; }
.hcard-info { flex: 1; padding: 16rpx 20rpx; min-width: 0; display: flex; flex-direction: column; }
.hcard-title { font-size: 26rpx; font-weight: 700; color: var(--color-text); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.hcard-type { display: inline-block; font-size: 20rpx; padding: 2px 8px; border-radius: 4px; background: var(--color-accent-light); color: var(--color-accent-deep); font-weight: 600; margin-top: 8rpx; align-self: flex-start; }
.site-meta { display: flex; gap: 16rpx; font-size: 20rpx; color: var(--color-text-secondary); margin-top: 8rpx; }
.hcard-foot { display: flex; justify-content: space-between; align-items: center; margin-top: auto; padding-top: 8rpx; }
.hcard-price { font-size: 22rpx; color: var(--color-accent-deep); font-weight: 700; }
.hcard-cta { font-size: 22rpx; color: var(--color-primary); font-weight: 600; padding: 4rpx 16rpx; border: 1rpx solid var(--color-primary); border-radius: 6px; }
.hcard-btn { font-size: 22rpx; color: #fff; background: var(--color-primary); padding: 4rpx 16rpx; border-radius: 6px; font-weight: 600; }

/* 骨架 */
.skel-list { padding: 12px; }
.skel-row { display: flex; gap: 12px; background: #fff; border-radius: 8px; padding: 12px; margin-bottom: 20rpx; }
.skel-img { width: 252rpx; height: 168rpx; border-radius: 6px; background: var(--color-divider); flex-shrink: 0; }
.skel-body { flex: 1; padding-top: 4px; }
.skel-line { height: 24rpx; border-radius: 4px; background: var(--color-divider); margin-bottom: 12rpx; }
.skel-line.w60 { width: 60%; }

/* 状态 */
.state-box { padding: 100rpx 40rpx; display: flex; flex-direction: column; align-items: center; gap: 16rpx; }
.state-box.mini { padding: 60rpx 40rpx; }
.s-ico { width: 128rpx; height: 128rpx; border-radius: 50%; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; }
.s-box { width: 56rpx; height: 40rpx; border: 3rpx solid var(--color-primary); border-radius: 6rpx; position: relative; margin-top: 12rpx; }
.s-box::before {
  content: ''; position: absolute; top: -14rpx; left: 12rpx;
  width: 28rpx; height: 12rpx;
  border: 3rpx solid var(--color-primary); border-bottom: none; border-radius: 4rpx 4rpx 0 0;
  background: var(--color-primary-light);
}
.s-fail { width: 8rpx; height: 40rpx; border-radius: 4rpx; background: var(--color-primary); position: relative; }
.s-fail::after {
  content: ''; position: absolute; bottom: -20rpx; left: 50%; transform: translateX(-50%);
  width: 8rpx; height: 8rpx; border-radius: 50%; background: var(--color-primary);
}
.s-title { font-size: 26rpx; color: var(--color-text-secondary); }
.state-note { font-size: 22rpx; color: var(--color-text-placeholder); }
.more-tip { text-align: center; font-size: 22rpx; color: var(--color-text-placeholder); padding: 16rpx 0; }
</style>
