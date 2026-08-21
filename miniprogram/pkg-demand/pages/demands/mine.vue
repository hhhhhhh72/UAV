<template>
  <view class="mine-page">
    <!-- 头部 -->
    <view class="page-header" :style="headerStyle">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">我的发布</text>
    </view>

    <!-- 类型筛选 -->
    <view class="mine-filters">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="t in kindOptions"
            :key="t.value"
            class="filter-chip"
            :class="{ active: mineType === t.value }"
            @tap="mineType = t.value"
          >{{ t.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 状态筛选 -->
    <view class="mine-filters status">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="s in statusOptions"
            :key="s"
            class="filter-chip"
            :class="{ active: mineStatus === s }"
            @tap="mineStatus = s"
          >{{ s }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">发布记录</text>
      <text class="list-count">共 {{ filteredPosts.length }} 条</text>
    </view>

    <!-- 空状态 -->
    <view v-if="filteredPosts.length === 0" class="state-panel">
      <view class="state-mark">⌁</view>
      <text class="state-title">{{ loadError ? '加载失败' : '暂无符合条件的发布' }}</text>
      <text class="state-desc">{{ loadError ? '网络异常，请稍后重试' : '发布的需求、服务、商品、课程都在这里查看' }}</text>
      <view class="state-btn" @tap="loadError ? fetchMine() : resetMineFilter">{{ loadError ? '重新加载' : '清除筛选' }}</view>
    </view>

    <!-- 发布记录 -->
    <view v-else class="post-list">
      <view v-for="post in filteredPosts" :key="post.id" class="mine-card" hover-class="mine-card--active" @tap="goDetail(post)">
        <view class="tag-row">
          <text class="tag" :class="typeTagClass(post.type)">{{ post.label }}</text>
          <text class="tag" :class="statusTagClass(post.statusKey)">{{ post.status }}</text>
        </view>
        <text class="post-title">{{ post.title }}</text>
        <view class="post-meta">
          <text v-for="(m, i) in post.meta" :key="i" class="post-meta-item">{{ m }}</text>
        </view>
        <view class="mine-action-row" v-if="post.source === 'backend' && post.type === 'demand'">
          <template v-if="post.statusKey === 'rejected'">
            <view class="action-link" @tap.stop="republish(post)">重新发布</view>
          </template>
          <template v-else-if="post.statusKey === 'published'">
            <view class="action-link" @tap.stop="goIntents(post.id)">查看意向</view>
            <view class="action-link" @tap.stop="completePost(post)">标记完成</view>
            <view class="action-link danger" @tap.stop="closePost(post)">下架</view>
          </template>
          <template v-else-if="post.statusKey === 'pending'">
            <view class="action-link" @tap.stop="toastPending">查看审核进度</view>
          </template>
        </view>
        <view class="mine-action-row" v-else-if="post.source === 'local' && post.statusKey === 'live'">
          <view class="action-link" @tap.stop="goDetail(post)">查看详情</view>
          <view class="action-link danger" @tap.stop="localOffShelf(post)">下架</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { safeNavigateTo } from '../../../utils/nav'
import { request, getErrorMessage, authStorage } from '../../../utils/request'
import { getPosts, upsertPost, KIND_ORDER, KIND_LABEL } from '../../../utils/publishData'
import { bizTypeLabel } from '../../../utils/enums'
import { useSafeTop } from '../../../utils/safeTop'

const mineType = ref('')
const mineStatus = ref('全部')
const posts = ref([])
const loadError = ref(false)

// 自定义导航：状态栏留白 + 右上角避让微信胶囊（JS 方式）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)
const headerStyle = computed(() => ({
  paddingTop: (topPad.value || statusBarH.value) + 'px',
  height: (56 + (topPad.value || statusBarH.value)) + 'px',
}))

// 类型筛选：全部 + 四类发布内容
const kindOptions = KIND_ORDER.map((k) => ({ value: k === 'all' ? '' : k, label: KIND_LABEL[k] }))

// 通用状态筛选（跨 需求/服务/商品/课程 归一化分组）
const statusOptions = ['全部', '已发布', '审核中', '草稿', '已下架', '已结束', '未通过']
const STATUS_GROUP = {
  live: '已发布', published: '已发布', listed: '已发布',
  pending: '审核中', draft: '草稿',
  removed: '已下架', cancelled: '已下架',
  completed: '已结束', rejected: '未通过',
}
const PRODUCT_STATUS = { pending: '待审核', listed: '在售', sold: '已售', removed: '已下架' }
const DEMAND_STATUS = { pending: '待审核', published: '已上架', completed: '已结束', cancelled: '已下架', rejected: '未通过' }

const filteredPosts = computed(() => {
  return posts.value.filter(
    (p) =>
      (mineType.value === '' || p.type === mineType.value) &&
      (mineStatus.value === '全部' || (STATUS_GROUP[p.statusKey] || p.status || '') === mineStatus.value)
  )
})

function typeTagClass(t) {
  if (t === 'demand') return 'blue'
  if (t === 'product') return 'orange'
  if (t === 'service') return 'green'
  return 'purple' // course
}
function statusTagClass(key) {
  if (key === 'rejected' || key === 'removed' || key === 'cancelled') return 'red'
  if (key === 'pending' || key === 'draft') return 'orange'
  if (key === 'completed' || key === 'sold') return 'gray'
  return 'green'
}

/* ================= 数据加载 ================= */

const fmtMoney = (fen) => {
  const yuan = (Number(fen) || 0) / 100
  return String(Math.round(yuan)).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}
const formatDate = (iso) => {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return String(iso).slice(0, 10)
  const m = d.getMonth() + 1
  const day = d.getDate()
  return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
}
const normalizeList = (res) => {
  if (Array.isArray(res)) return res
  if (res && Array.isArray(res.data)) return res.data
  return []
}

// 后端需求 → 统一卡片
function demandToCard(d) {
  return {
    id: d.id,
    type: 'demand',
    source: 'backend',
    label: '需求',
    title: d.title || '未命名需求',
    status: DEMAND_STATUS[d.status] || d.status || '',
    statusKey: d.status || 'published',
    meta: [bizTypeLabel(d.biz_type) || '其他', d.district || '重庆', d.budget_fen ? '预算 ' + fmtMoney(d.budget_fen) + ' 元' : '预算可协商'],
    date: formatDate(d.created_at),
    raw: d,
  }
}

// 后端商品 → 统一卡片
function productToCard(p) {
  return {
    id: p.id,
    type: 'product',
    source: 'backend',
    label: '商品设备',
    title: p.title || '未命名商品',
    status: PRODUCT_STATUS[p.status] || '在售',
    statusKey: p.status || 'listed',
    meta: [p.brand || '品牌待定', p.model || '型号待定', p.price_fen ? fmtMoney(p.price_fen) + ' 元' : '面议'],
    date: formatDate(p.created_at),
    raw: p,
  }
}

// 后端服务能力 → 统一卡片
function serviceToCard(s) {
  return {
    id: s.id,
    type: 'service',
    source: 'backend',
    label: '服务能力',
    title: s.title || '未命名服务',
    status: s.status === 'published' ? '已发布' : (s.status === 'off' ? '已下架' : '审核中'),
    statusKey: s.status === 'published' ? 'published' : (s.status === 'off' ? 'removed' : 'pending'),
    meta: [s.category || '', s.region || ''].filter(Boolean),
    date: formatDate(s.created_at),
    raw: s,
  }
}

// 后端课程 → 统一卡片
function courseToCard(c) {
  return {
    id: c.id,
    type: 'course',
    source: 'backend',
    label: '培训课程',
    title: c.title || '未命名课程',
    status: c.status === 'published' ? '已发布' : (c.status === 'draft' ? '草稿' : '审核中'),
    statusKey: c.status === 'published' ? 'published' : (c.status === 'draft' ? 'draft' : 'pending'),
    meta: [c.org_name || '', c.district || ''].filter(Boolean),
    date: formatDate(c.created_at),
    raw: c,
  }
}

// 本地发布记录 → 统一卡片（商品 backendId 非空的由后端商品统一展示，这里跳过）
function localToCard(p) {
  return {
    id: p.id,
    type: p.type,
    source: 'local',
    label: p.label || p.type,
    title: p.title,
    status: p.status || '',
    statusKey: p.statusKey || 'live',
    meta: p.meta || [],
    date: p.date || '',
    backendId: p.backendId || '',
    raw: p,
  }
}

const fetchMine = async () => {
  loadError.value = false
  try {
    const [demandsRes, productsRes, servicesRes, coursesRes] = await Promise.all([
      request({ url: '/api/v1/demands?mine=1&page_size=100' }).catch(() => []),
      request({ url: '/api/v1/products?mine=1&page_size=100' }).catch(() => []),
      request({ url: '/api/v1/service-listings?mine=1&page_size=100' }).catch(() => []),
      request({ url: '/api/v1/training-courses?mine=1&page_size=100' }).catch(() => []),
    ])
    const cards = []
    // 后端记录（权威，含本地缓存被清后的记录；四类全部走后端）
    normalizeList(productsRes).forEach((p) => cards.push(productToCard(p)))
    normalizeList(demandsRes).forEach((d) => cards.push(demandToCard(d)))
    normalizeList(servicesRes).forEach((s) => cards.push(serviceToCard(s)))
    normalizeList(coursesRes).forEach((c) => cards.push(courseToCard(c)))
    // 本地发布：有 backendId 且后端已返回的跳过（需求/服务/课程同商品规则，防同一发布重复展示）
    const backendIds = new Set(cards.map((c) => String(c.id)))
    getPosts().forEach((p) => {
      if (p.backendId && backendIds.has(String(p.backendId))) return
      cards.push(localToCard(p))
    })
    posts.value = cards
  } catch (e) {
    loadError.value = true
    posts.value = []
  }
}

onLoad((options) => {
  initSafeTop()
  if (options && options.status && statusOptions.includes(options.status)) {
    mineStatus.value = options.status
  }
  // 未登录不进"我的发布"：后端 mine=1 未登录只返回空列表，这里提前拦截引导登录
  if (!authStorage.getAccessToken()) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  fetchMine()
})
onPullDownRefresh(() => {
  fetchMine().finally(() => uni.stopPullDownRefresh())
})

const resetMineFilter = () => {
  mineType.value = ''
  mineStatus.value = '全部'
}

/* ================= 跳转 ================= */

const goDetail = (post) => {
  if (post.source === 'backend') {
    if (post.type === 'product') {
      safeNavigateTo('/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(post.id))
      return
    }
    if (post.type === 'demand') {
      safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(post.id))
      return
    }
    if (post.type === 'service') {
      safeNavigateTo('/pages/demands/index')
      return
    }
    if (post.type === 'course') {
      safeNavigateTo('/pkg-talent/pages/training/courses')
      return
    }
  }
  // 本地记录 → 本地详情
  safeNavigateTo('/pages/publish/detail?id=' + encodeURIComponent(post.id))
}
const goIntents = (id) => safeNavigateTo('/pkg-demand/pages/demands/intents?demandId=' + encodeURIComponent(id))
const goBack = () => uni.navigateBack()

// 重新发布：后端无按 id 复制接口，原内容不会带入，先确认再进发布页新建
function republish(post) {
  uni.showModal({
    title: '重新发布',
    content: '重新发布将新建一条需求，原内容不会自动带入，需重新填写。确定继续？',
    success: (r) => {
      if (r.confirm) safeNavigateTo('/pkg-demand/pages/demands/publish')
    },
  })
}

/* ================= 操作 ================= */

// 本地下架：仅改状态（需求大厅/各列表只并入 live 记录），保留在"我的发布"
async function localOffShelf(post) {
  const confirm = await new Promise((resolve) => {
    uni.showModal({ title: '下架发布', content: '下架后将从对应列表移除，可在筛选「已下架」中找回，确定下架？', success: (r) => resolve(r.confirm) })
  })
  if (!confirm) return
  upsertPost(Object.assign({}, post.raw, { statusKey: 'removed', status: '已下架' }))
  post.statusKey = 'removed'
  post.status = '已下架'
  uni.showToast({ title: '已下架', icon: 'none' })
}

async function completePost(post) {
  try {
    await request({ url: '/api/v1/demands/' + encodeURIComponent(post.id) + '/complete', method: 'POST' })
    post.status = '已结束'
    post.statusKey = 'completed'
    uni.showToast({ title: '已标记完成', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

async function closePost(post) {
  const confirm = await new Promise((resolve) => {
    uni.showModal({ title: '下架需求', content: '下架后需求将从需求大厅移除，确定下架？', success: (r) => resolve(r.confirm) })
  })
  if (!confirm) return
  try {
    await request({ url: '/api/v1/demands/' + encodeURIComponent(post.id) + '/cancel', method: 'POST' })
    post.status = '已下架'
    post.statusKey = 'cancelled'
    uni.showToast({ title: '已下架', icon: 'none' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

const toastPending = () => {
  uni.showToast({ title: '审核进度：等待协会审核', icon: 'none' })
}
</script>

<style scoped>
.mine-page {
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
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-action { padding: 14rpx; }
.head-action-text { color: #0A66C2; font-size: 26rpx; font-weight: 600; }

/* 筛选 */
.mine-filters {
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 20rpx 24rpx;
}
.mine-filters.status { border-bottom: 0; padding-top: 4rpx; }
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
