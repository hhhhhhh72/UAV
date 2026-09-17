<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">我的发布</view>
      <view class="pub-nav-action" hover-class="pub-fade" :style="{ marginRight: capsuleGap + 'px' }" @tap="filterOpen = true">筛选</view>
    </view>

    <!-- 三张概览卡 -->
    <view class="pub-summary">
      <view class="pub-summary-item" hover-class="pub-summary-item--active" @tap="setTab('pending')">
        <text class="pub-summary-num">{{ counts.pending }}</text>
        <text class="pub-summary-label">审核中</text>
      </view>
      <view class="pub-summary-item" hover-class="pub-summary-item--active" @tap="setTab('live')">
        <text class="pub-summary-num">{{ counts.live }}</text>
        <text class="pub-summary-label">已发布</text>
      </view>
      <view class="pub-summary-item" hover-class="pub-summary-item--active" @tap="setTab('draft')">
        <text class="pub-summary-num">{{ counts.draft }}</text>
        <text class="pub-summary-label">待完善</text>
      </view>
    </view>

    <!-- 状态筛选胶囊 -->
    <scroll-view scroll-x class="pub-filter-scroll" :show-scrollbar="false">
      <view class="pub-filter-inner">
        <view
          v-for="tab in tabOrder"
          :key="tab"
          class="pub-filter-chip"
          :class="{ 'pub-filter-chip--active': postsTab === tab }"
          @tap="setTab(tab)"
        >{{ TAB_LABEL[tab] }}</view>
      </view>
    </scroll-view>

    <!-- 列表 / 空状态 -->
    <view class="pub-posts">
      <view v-if="filtered.length === 0" class="pub-empty">
        <view class="pub-empty-mark">⌁</view>
        <view class="pub-empty-title">暂无符合条件的发布</view>
        <view class="pub-empty-desc">更换筛选条件，或去发布新内容。</view>
      </view>
      <view
        v-for="post in filtered"
        :key="post.id"
        class="pub-post-card"
        hover-class="pub-post-card--active"
        @tap="openDetail(post)"
      >
        <view class="pub-post-top">
          <text class="pub-post-type">{{ post.label }}</text>
          <text class="pub-post-status" :class="statusClass(post.statusKey)">{{ post.status }}</text>
        </view>
        <view class="pub-post-title">{{ post.title }}</view>
        <view class="pub-post-meta">
          <text v-for="(m, i) in post.meta" :key="i" class="pub-meta-item">{{ m }}</text>
        </view>
        <view class="pub-post-foot">
          <text>{{ post.date }}</text>
          <text class="pub-post-foot-strong">{{ post.leads || post.note }}</text>
        </view>
      </view>
    </view>

    <!-- 类型筛选抽屉 -->
    <view v-if="filterOpen" class="pub-overlay" @tap="filterOpen = false">
      <view class="pub-sheet" @tap.stop>
        <view class="pub-grab"></view>
        <view class="pub-sheet-head">
          <view class="pub-sheet-head-title">筛选发布类型</view>
          <view class="pub-sheet-cancel" @tap="filterOpen = false">完成</view>
        </view>
        <view
          v-for="kind in kindOrder"
          :key="kind"
          class="pub-option"
          :class="{ 'pub-option--selected': postKind === kind }"
          @tap="pickKind(kind)"
        >
          <text>{{ KIND_LABEL[kind] }}</text>
          <text v-if="postKind === kind" class="pub-option-check">✓</text>
        </view>
      </view>
    </view>

    <!-- 底部黑色 toast -->
    <view v-if="toast" class="pub-toast">{{ toast }}</view>
  </view>
</template>

<script setup>
import { safeBack } from '../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import {
  getPosts, TAB_ORDER, TAB_LABEL, KIND_ORDER, KIND_LABEL,
  pruneStalePointers, MINE_PAGE_SIZE,
} from '../../utils/publishData'
import { request, requireLogin } from '../../utils/request'
import { useSafeTop } from '../../utils/safeTop'
import { SERVICE_PROD_TYPES } from '../../utils/enums'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const postsTab = ref('all')
const postKind = ref('all')
const filterOpen = ref(false)
const toast = ref('')
const toastTimer = ref(null)

const tabOrder = TAB_ORDER
const kindOrder = KIND_ORDER

const allPosts = ref([])

/* ── 后端记录 → 卡片结构（与本地 publishData 卡片字段对齐） ── */

const BIZ_LABELS = {
  cable_inspection: '工业巡检',
  plant_transport: '植保运输',
  spray_pesticide: '农药喷洒',
  clean_paint: '清洗保洁',
  trade_lease: '租赁服务',
  other: '其他服务',
}

const fmtDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return ''
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

// demand.status: pending/published/matched/completed/rejected
const mapDemand = (d) => {
  let statusKey = 'pending'
  let status = '待审核'
  switch (d.status) {
    case 'published': statusKey = 'live'; status = '招募中'; break
    case 'matched': statusKey = 'live'; status = '已匹配'; break
    case 'completed': statusKey = 'live'; status = '已完成'; break
    case 'rejected': statusKey = 'rejected'; status = '未通过'; break
  }
  return {
    id: d.id, type: 'demand', label: '需求', backend: true,
    statusKey, status,
    title: d.title || '',
    meta: [d.district, BIZ_LABELS[d.biz_type]].filter(Boolean),
    date: fmtDate(d.created_at),
    note: '',
  }
}

// 服务类目商品：合并后它们也是商品，但在「我的发布」里仍按"服务能力"展示。
// 列表来自 utils/enums.js（对应后端 service_listing.go:35-41），不在页面里另立一份。

// product 的两个维度必须**分开判**：check_status（审核）× status（上架）。
//
// 修复前这里只判 status 的 sold/off，而服务端既不会返回 off、驳回写的又是 removed，
// 于是**待审核和被驳回的商品在用户端全部显示成"在售"**——卖家以为已经上架了。
//   check_status=pending  → 审核中
//   check_status=rejected → 未通过（note 带驳回原因，卖家知道要改什么）
//   status=listed         → 在售
//   status=sold           → 已售
//   status=removed        → 已下架
const mapProduct = (p) => {
  let statusKey = 'live'
  let status = '在售'
  let note = ''
  if (p.check_status === 'rejected') {
    statusKey = 'rejected'
    status = '未通过'
    note = p.check_reason || ''
  } else if (p.check_status === 'pending') {
    statusKey = 'pending'
    status = '审核中'
  } else if (p.status === 'sold') {
    status = '已售'
  } else if (p.status === 'removed') {
    statusKey = 'rejected'
    status = '已下架'
  } else if (p.status !== 'listed') {
    // 兜底：审核已通过但上架状态异常（历史数据/空串）时不得当作在售
    statusKey = 'pending'
    status = '未上架'
  }
  const price = p.price_fen ? '¥' + (p.price_fen / 100) : ''
  return {
    // 服务类目商品仍显示为「服务能力」，与合并前的用户认知一致
    id: p.id, type: 'product', label: SERVICE_PROD_TYPES.indexOf(p.prod_type) >= 0 ? '服务能力' : '商品', backend: true,
    statusKey, status,
    title: p.title || '',
    meta: [p.prod_type ? (p.prod_type === 'drone' ? '整机' : p.prod_type === 'repair' ? '维修服务' : '零部件') : '', price].filter(Boolean),
    date: fmtDate(p.created_at),
    note,
  }
}

// service_listing.status: pending/published（off 表示已下架）
const mapService = (sl) => {
  let statusKey = 'pending'
  let status = '审核中'
  if (sl.status === 'published') { statusKey = 'live'; status = '已发布' }
  // 已下架同商品：归入非 live 状态（rejected「未通过」Tab），避免误标 live
  if (sl.status === 'off') { statusKey = 'rejected'; status = '已下架' }
  return {
    id: sl.id, type: 'service', label: '服务能力', backend: true,
    statusKey, status,
    title: sl.title || '',
    meta: [sl.category, sl.region].filter(Boolean),
    date: fmtDate(sl.created_at),
    note: '',
  }
}

// course.status: draft/published
const mapCourse = (c) => {
  // 区分 pending（审核中）/draft（草稿）：只有 published 才算「已发布」
  let statusKey = 'pending'
  let status = '审核中'
  if (c.status === 'published') { statusKey = 'live'; status = '已发布' }
  else if (c.status === 'draft') { statusKey = 'draft'; status = '草稿' }
  return {
    id: c.id, type: 'course', label: '培训课程', backend: true,
    statusKey, status,
    title: c.title || '',
    meta: [c.org_name, c.district].filter(Boolean),
    date: fmtDate(c.created_at),
    note: '',
  }
}

async function refresh() {
  let local = getPosts()
  // 后端：我的需求 + 商品 + 课程（mine=1；未登录返回空，不会泄露他人数据）
  //
  // 服务能力已并入商品表（migration 000110）：/api/v1/products 返回的**已经包含**服务类目，
  // 再单独拉一次 /service-listings 会让同一批数据出现两遍（且 id 相同，去重只能救一半）。
  let backend = []
  let backendOK = false
  try {
    const [dRes, pRes, cRes] = await Promise.all([
      request({ url: '/api/v1/demands', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/products', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/training-courses', data: { mine: '1', page: 1, page_size: 100 } }),
    ])
    const dList = Array.isArray(dRes) ? dRes : dRes?.data || []
    const pList = Array.isArray(pRes) ? pRes : pRes?.data || []
    const cList = Array.isArray(cRes) ? cRes : cRes?.data || []
    backend = [
      ...dList.map(mapDemand),
      ...pList.map(mapProduct),
      ...cList.map(mapCourse),
    ]
    backendOK = true
    // 后端已确认不存在的实体：本地死指针顺手清掉（三个接口都完整返回时才清），
    // 清完重读本地，避免拿清理前的快照继续合并。
    const pruned = pruneStalePointers(
      backend.map((b) => String(b.id)),
      { complete: [dList, pList, cList].every((l) => l.length < MINE_PAGE_SIZE) }
    )
    if (pruned > 0) local = getPosts()
  } catch (e) {
    // 后端不可用：保留本地记录展示，不阻塞页面
  }
  // 本地保留：草稿 + 没有 backendId 的纯本地记录。
  //
  // 带 backendId 的是"后端实体的本地指针"，后端拉取成功时一律不用它——
  // 后端那份才是真值。这同时解决两种情况：同一条数两遍，以及实体已被删除
  // （老指针的 id 不在后端集合里，早先按 backendIds.has 判断会把它留下来，
  // 于是删掉一条老需求后列表里还挂着一张点不开的幽灵卡片）。
  // 后端不可用时（backendOK=false）退化为纯本地口径，页面不至于空白。
  const localKeep = local.filter(
    (p) => p.statusKey === 'draft' || !p.backendId || !backendOK
  )
  allPosts.value = [...localKeep, ...backend]
}

const filtered = computed(() => {
  return allPosts.value.filter(
    (p) =>
      (postsTab.value === 'all' || p.statusKey === postsTab.value) &&
      (postKind.value === 'all' || p.type === postKind.value)
  )
})

const counts = computed(() => {
  const list = allPosts.value
  return {
    pending: list.filter((p) => p.statusKey === 'pending').length,
    live: list.filter((p) => p.statusKey === 'live').length,
    draft: list.filter((p) => p.statusKey === 'draft').length,
  }
})

function statusClass(key) {
  return 'status-' + key
}

function setTab(tab) {
  postsTab.value = tab
}

function pickKind(kind) {
  postKind.value = kind
  filterOpen.value = false
}

function openDetail(post) {
  // 后端记录 → 跳真实详情/列表；本地草稿 → 本地详情
  if (post.backend) {
    if (post.type === 'demand') {
      uni.navigateTo({ url: '/pages/demands/detail?id=' + encodeURIComponent(post.id) })
      return
    }
    if (post.type === 'product') {
      uni.navigateTo({ url: '/pkg-eco/pages/mall/detail?id=' + encodeURIComponent(post.id) })
      return
    }
    if (post.type === 'service') {
      uni.navigateTo({ url: '/pages/demands/index' })
      return
    }
    if (post.type === 'course') {
      uni.navigateTo({ url: '/pkg-talent/pages/training/courses' })
      return
    }
  }
  uni.navigateTo({
    url: '/pages/publish/detail?id=' + post.id +
      '&tab=' + postsTab.value + '&kind=' + postKind.value,
  })
}

function goBack() {
  safeBack()
}

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => { toast.value = '' }, 2200)
}

onLoad((options) => {
  initSafeTop()
  if (options && options.tab && TAB_LABEL[options.tab]) postsTab.value = options.tab
  if (options && options.kind && KIND_LABEL[options.kind]) postKind.value = options.kind
  refresh()
})

onShow(() => {
  // 未登录每次进入都拦截（从登录页返回继续跳转，必须登录后查看）
  if (!requireLogin('请先登录后查看我的发布')) return
  // 从详情页返回后刷新（撤回/下架/删除后列表状态变化）
  refresh()
})
</script>

<style scoped>
@import './pub-style.css';
.pub-fade { opacity: 0.6; }
.pub-filter-inner {
  display: inline-flex;
  gap: 8px;
}
</style>
