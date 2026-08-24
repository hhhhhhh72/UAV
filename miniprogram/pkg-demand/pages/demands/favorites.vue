<template>
  <view class="fav-page">
    <!-- 头部 -->
    <view class="page-header" :style="headerStyle">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">我的收藏</text>
      <view class="head-action" :style="{ marginRight: capsuleGap + 'px' }" @tap="goHall">
        <text class="head-action-text">需求大厅</text>
      </view>
    </view>

    <!-- 类型筛选 -->
    <view class="mine-filters">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="t in kindTabs"
            :key="t.value"
            class="filter-chip"
            :class="{ active: favType === t.value }"
            @tap="favType = t.value"
          >{{ t.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">收藏记录</text>
      <text class="list-count">共 {{ visibleCards.length }} 条</text>
    </view>

    <!-- 加载中骨架 -->
    <view v-if="loading" class="post-list">
      <view v-for="i in 3" :key="i" class="mine-card">
        <view class="skl" style="width: 96rpx; height: 36rpx"></view>
        <view class="skl post-title-skl"></view>
        <view class="skl" style="width: 320rpx; height: 26rpx"></view>
      </view>
    </view>

    <!-- 空状态 -->
    <view v-else-if="visibleCards.length === 0" class="state-panel">
      <view class="state-mark">♡</view>
      <text class="state-title">{{ loadError ? '加载失败' : emptyTitle }}</text>
      <text class="state-desc">{{ loadError ? '网络异常，请稍后重试' : emptyDesc }}</text>
      <view class="state-btn" @tap="loadError ? fetchAll() : goHall">{{ loadError ? '重新加载' : '去逛逛' }}</view>
    </view>

    <!-- 收藏列表 -->
    <view v-else class="post-list">
      <view v-for="post in visibleCards" :key="post.type + '-' + post.id" class="mine-card" hover-class="mine-card--active" @tap="goDetail(post)">
        <view class="tag-row">
          <text class="tag" :class="typeTagClass(post.type)">{{ post.label }}</text>
          <text class="tag" :class="statusTagClass(post.statusKey)">{{ post.status }}</text>
        </view>
        <text class="post-title">{{ post.title }}</text>
        <view class="post-meta">
          <text v-for="(m, i) in post.meta" :key="i" class="post-meta-item">{{ m }}</text>
        </view>
        <view class="mine-action-row">
          <view class="action-link" @tap.stop="goDetail(post)">查看详情</view>
          <view class="action-link danger" @tap.stop="unfavorite(post)">取消收藏</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { safeNavigateTo, safeBack } from '../../../utils/nav'
import { request, getErrorMessage, authStorage } from '../../../utils/request'
import { bizTypeLabel } from '../../../utils/enums'
import { useSafeTop } from '../../../utils/safeTop'

const favType = ref('all')
const cards = ref([])
const loading = ref(false)
const loadError = ref(false)

// 自定义导航：状态栏留白 + 右上角避让微信胶囊
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)
const headerStyle = computed(() => ({
  paddingTop: (topPad.value || statusBarH.value) + 'px',
  height: (56 + (topPad.value || statusBarH.value)) + 'px',
}))

/* ===== 类型筛选 ===== */
const kindTabs = [
  { label: '全部', value: 'all' },
  { label: '需求', value: 'demand' },
  { label: '商品', value: 'product' },
  { label: '服务', value: 'service' },
  { label: '课程', value: 'course' },
]

const visibleCards = computed(() =>
  favType.value === 'all' ? cards.value : cards.value.filter((c) => c.type === favType.value)
)

const emptyTitle = computed(() => {
  const map = { all: '还没有收藏', demand: '还没有收藏需求', product: '还没有收藏商品', service: '还没有收藏服务', course: '还没有收藏课程' }
  return map[favType.value] || '还没有收藏'
})
const emptyDesc = computed(() => {
  const map = {
    all: '在需求大厅、商城、培训里看到感兴趣的内容，点「收藏」就能在这里找到',
    demand: '在需求大厅看到感兴趣的需求，点「收藏」就能在这里找到',
    product: '在商城看到感兴趣的设备，点「收藏」就能在这里找到',
    service: '在需求大厅的服务能力里点「收藏」就能在这里找到',
    course: '在培训课程里点「收藏」就能在这里找到',
  }
  return map[favType.value] || ''
})

function typeTagClass(t) {
  if (t === 'demand') return 'blue'
  if (t === 'product') return 'orange'
  if (t === 'service') return 'green'
  return 'purple' // course
}
function statusTagClass(key) {
  if (key === 'rejected' || key === 'removed' || key === 'cancelled' || key === 'off') return 'red'
  if (key === 'pending' || key === 'draft') return 'orange'
  if (key === 'completed' || key === 'sold' || key === 'full') return 'gray'
  return 'green'
}

/* ===== 数据归一化 ===== */
const fmtMoney = (fen) => {
  const yuan = (Number(fen) || 0) / 100
  return String(Math.round(yuan)).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}
const normalizeList = (res) => {
  if (Array.isArray(res)) return res
  if (res && Array.isArray(res.data)) return res.data
  return []
}

const DEMAND_STATUS = { pending: '待审核', published: '已上架', completed: '已结束', cancelled: '已下架', rejected: '未通过' }
const PRODUCT_STATUS = { pending: '待审核', listed: '在售', sold: '已售', removed: '已下架' }

function demandToCard(d) {
  return {
    id: d.id,
    type: 'demand',
    label: '需求',
    title: d.title || '未命名需求',
    status: DEMAND_STATUS[d.status] || d.status || '',
    statusKey: d.status || 'published',
    meta: [bizTypeLabel(d.biz_type) || '其他', d.district || '重庆', d.budget_fen ? '预算 ' + fmtMoney(d.budget_fen) + ' 元' : '预算可协商'],
    raw: d,
  }
}
function productToCard(p) {
  return {
    id: p.id,
    type: 'product',
    label: '商品设备',
    title: p.title || '未命名商品',
    status: PRODUCT_STATUS[p.status] || '在售',
    statusKey: p.status || 'listed',
    meta: [p.brand || '品牌待定', p.model || '型号待定', p.price_fen ? fmtMoney(p.price_fen) + ' 元' : '面议'],
    raw: p,
  }
}
function serviceToCard(s) {
  return {
    id: s.id,
    type: 'service',
    label: '服务能力',
    title: s.title || '未命名服务',
    status: s.status === 'published' ? '已发布' : (s.status === 'off' ? '已下架' : '审核中'),
    statusKey: s.status === 'published' ? 'published' : (s.status === 'off' ? 'off' : 'pending'),
    meta: [s.category || '', s.region || '', s.price_fen ? fmtMoney(s.price_fen) + ' 元/' + (s.unit || '次') : ''].filter(Boolean),
    raw: s,
  }
}
function courseToCard(c) {
  return {
    id: c.id,
    type: 'course',
    label: '培训课程',
    title: c.title || '未命名课程',
    status: c.status === 'published' ? '已发布' : (c.status === 'draft' ? '草稿' : '审核中'),
    statusKey: c.status === 'published' ? 'published' : (c.status === 'draft' ? 'draft' : 'pending'),
    meta: [c.org_name || '', c.district || ''].filter(Boolean),
    raw: c,
  }
}

/* ===== 数据加载 ===== */
const fetchAll = async () => {
  loading.value = true
  loadError.value = false
  try {
    const [demandsRes, productsRes, servicesRes, coursesRes] = await Promise.all([
      request({ url: '/api/v1/demands/favorites/mine' }).catch(() => []),
      request({ url: '/api/v1/products/favorites/mine' }).catch(() => []),
      request({ url: '/api/v1/service-listings/favorites/mine' }).catch(() => []),
      request({ url: '/api/v1/training-courses/favorites/mine' }).catch(() => []),
    ])
    const out = []
    normalizeList(demandsRes).forEach((d) => out.push(demandToCard(d)))
    normalizeList(productsRes).forEach((p) => out.push(productToCard(p)))
    normalizeList(servicesRes).forEach((s) => out.push(serviceToCard(s)))
    normalizeList(coursesRes).forEach((c) => out.push(courseToCard(c)))
    // 按收藏时间倒序：各接口本身按收藏时间倒序返回，合并后按日期粗略排序
    out.sort((a, b) => String(b.raw.created_at || '').localeCompare(String(a.raw.created_at || '')))
    cards.value = out
  } catch (e) {
    loadError.value = true
    cards.value = []
  } finally {
    loading.value = false
  }
}

onLoad(() => {
  initSafeTop()
  // 未登录不进收藏页：引导登录
  if (!authStorage.getAccessToken()) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  fetchAll()
})
onPullDownRefresh(() => {
  fetchAll().finally(() => uni.stopPullDownRefresh())
})

/* ================= 跳转 ================= */

const goDetail = (post) => {
  if (post.type === 'product') {
    safeNavigateTo('/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(post.id))
    return
  }
  if (post.type === 'course') {
    safeNavigateTo('/pkg-talent/pages/training/enroll?id=' + encodeURIComponent(post.id))
    return
  }
  // 需求 / 服务 共用详情页（按 id 分流）
  safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(post.id))
}
const goHall = () => safeNavigateTo('/pages/demands/index')
const goBack = () => safeBack()

/* ================= 操作 ================= */

async function unfavorite(post) {
  const confirm = await new Promise((resolve) => {
    uni.showModal({ title: '取消收藏', content: '确定不再收藏这条内容吗？', success: (r) => resolve(r.confirm) })
  })
  if (!confirm) return
  const base = {
    demand: '/api/v1/demands/',
    product: '/api/v1/products/',
    service: '/api/v1/service-listings/',
    course: '/api/v1/training-courses/',
  }[post.type]
  if (!base) return
  try {
    await request({
      url: base + encodeURIComponent(post.id) + '/favorite',
      method: 'POST',
      data: { favorite: false },
    })
    cards.value = cards.value.filter((x) => !(x.type === post.type && x.id === post.id))
    uni.showToast({ title: '已取消收藏', icon: 'none' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}
</script>

<style scoped>
.fav-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* 头部 */
.page-header {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; text-align: center; }
.head-action { padding: 14rpx; }
.head-action-text { color: #0A66C2; font-size: 26rpx; font-weight: 600; }

/* 筛选 */
.mine-filters {
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 20rpx 24rpx;
}
.filter-scroll { white-space: nowrap; }
.filter-inner { display: inline-flex; gap: 12rpx; }
.filter-chip {
  display: inline-flex;
  align-items: center;
  height: 56rpx;
  padding: 0 20rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #344054;
  font-size: 24rpx;
  box-sizing: border-box;
}
.filter-chip.active {
  color: #0A66C2;
  border-color: #B9D6EF;
  background: #EAF3FB;
  font-weight: 650;
}

/* 列表 */
.list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

.post-list { padding: 0 32rpx 32rpx; }
.mine-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 26rpx;
  border: 1px solid #EEF1F4;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.mine-card + .mine-card { margin-top: 20rpx; }
.mine-card--active { opacity: .8; }

.tag-row { display: flex; gap: 10rpx; }
.tag {
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.purple { color: #7B61D1; background: #F0EDFF; }
.tag.red { color: #D92D20; background: #FEF3F2; }
.tag.gray { color: #667085; background: #F1F3F5; }

.post-title {
  display: block;
  font-size: 28rpx;
  line-height: 1.45;
  color: #17212B;
  font-weight: 700;
  margin: 16rpx 0 8rpx;
}
.post-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx 20rpx;
}
.post-meta-item { font-size: 22rpx; color: #667085; }

.mine-action-row {
  display: flex;
  gap: 32rpx;
  border-top: 1px solid #EEF1F4;
  margin-top: 22rpx;
  padding-top: 20rpx;
}
.action-link { color: #0A66C2; font-size: 24rpx; font-weight: 600; }
.action-link.danger { color: #D92D20; }

/* 骨架 */
.skl {
  border-radius: 8rpx;
  background: #EDF0F3;
}
.post-title-skl {
  height: 32rpx;
  margin: 18rpx 0;
  width: 70%;
}

/* 空状态 */
.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}
</style>
