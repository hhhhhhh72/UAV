<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="品牌展示" show-back :fixed="true" @back="goBack">
      <template #right>
        <view class="mine-entry" aria-role="button" aria-label="我的品牌" hover-class="tap-fade" @tap="goMine"><text>我</text></view>
      </template>
    </u-nav-bar>

    <!-- 搜索（对齐原型 .sbar） -->
    <view class="sbar">
      <view class="sbox">
        <u-icon name="search" size="28rpx" color="#98A2B3" />
        <input
          class="sinp"
          v-model="q"
          placeholder="搜索品牌名称、主营领域"
          placeholder-class="ph"
          confirm-type="search"
          @input="onSearch"
        />
        <view v-if="q" class="sclr" aria-role="button" aria-label="清除搜索" @tap="clearSearch">×</view>
      </view>
    </view>

    <!-- 精选横幅（接口替换点：featured 推荐位接口 / 列表 featured=true 字段） -->
    <view class="feature-wrap" v-if="featured.length">
      <swiper
        class="feature-swiper"
        circular
        autoplay
        :interval="3500"
        :duration="400"
        @change="onSlide"
      >
        <swiper-item v-for="f in featured" :key="f.id">
          <view class="feature-card" :class="'gd-' + f.grad" hover-class="tap-fade" @tap="goDetailById(f.id)">
            <view class="f-inner">
              <view class="f-logo"><text>{{ f.logoText }}</text></view>
              <view class="f-txt">
                <text class="f-tag">{{ f.tag }}</text>
                <text class="f-name">{{ f.name }}</text>
                <text class="f-sub">{{ f.sub }}</text>
              </view>
              <view v-if="f.hasVideo" class="f-play"><text>▶</text></view>
            </view>
          </view>
        </swiper-item>
      </swiper>
      <view class="f-dots">
        <view v-for="(_, i) in featured" :key="i" class="f-dot" :class="{ on: sIdx === i }"></view>
      </view>
    </view>

    <!-- 分类 pills -->
    <scroll-view scroll-x :show-scrollbar="false" class="tabs-scroll">
      <view class="tabs">
        <view
          v-for="c in CATS"
          :key="c.k"
          class="tab"
          :class="{ act: cat === c.k }"
          @tap="pickCat(c.k)"
        >{{ c.l }}</view>
      </view>
    </scroll-view>

    <!-- 信息行：共 N 家 + 排序 -->
    <view class="ir">
      <text>共 <text class="irn">{{ totalCount }}</text> 家会员品牌</text>
      <view class="irs-wrap">
        <text class="irs" aria-role="button" aria-label="切换排序" @tap="toggleSort">{{ sortLabel }} ▾</text>
        <view v-if="showSort" class="spop" @tap.stop>
          <view
            v-for="s in SORTS"
            :key="s.v"
            class="sp-opt"
            :class="{ act: sort === s.v }"
            @tap="pickSort(s.v)"
          >
            <text>{{ s.l }}</text><text v-if="sort === s.v" class="chk">✓</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 列表内容 -->
    <view v-if="loading" class="sk-grid">
      <view v-for="i in 4" :key="'sk' + i" class="sk-card">
        <view class="sk-cv"></view>
        <view class="sk-bd"><view class="sk-l w80"></view><view class="sk-l w50"></view></view>
      </view>
      <view class="sk-tip">正在加载品牌数据…</view>
    </view>

    <view v-else-if="loadError" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" aria-role="button" aria-label="重新加载" @tap="retryLoad">重新加载</view>
      </u-empty>
    </view>

    <view v-else-if="!list.length" class="st">
      <u-empty description="暂无匹配品牌">
        <text class="sth">试试调整分类或搜索关键词</text>
        <view class="stb" aria-role="button" aria-label="清除筛选" @tap="resetAll">清除筛选</view>
      </u-empty>
    </view>

    <view v-else class="grid">
      <BrandCard
        v-for="b in list"
        :key="b.id"
        :item="b"
        :grad-class="b.grad"
        @click="goDetail"
      />
    </view>

    <view v-if="list.length && !loadError" class="lm">
      <text v-if="loadingMore">正在加载更多…</text>
      <text v-else-if="!hasMore">没有更多了</text>
    </view>

    <view class="mock-note" v-if="mockMode && isDev">当前为演示数据 · 接口就绪后自动切换</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom, onUnload } from '@dcloudio/uni-app'
import { request, BASE_URL } from '@/utils/request'
import BrandCard from '../../components/BrandCard.vue'
import { MOCK_BRANDS, MOCK_FEATURED, CATEGORY_MAP } from '@/utils/mockBrands'

// ===== 静态配置 =====
const CATS = [
  { k: 'all', l: '全部' },
  ...Object.entries(CATEGORY_MAP).map(([k, l]) => ({ k, l })),
]
const SORTS = [
  { v: 'latest', l: '最新入驻' },
  { v: 'views', l: '最多浏览' },
  { v: 'video', l: '视频优先' },
]
const SORT_LABEL = { views: '最多浏览', latest: '最新入驻', video: '视频优先' }
const PAGE_SIZE = 100
const SEARCH_DEBOUNCE_MS = 250 // 搜索防抖：停顿 250ms 后再请求（对齐 challenges）
const isDev = import.meta.env.DEV // 演示数据仅在开发环境回退

// ===== 状态 =====
const statusBarHeight = ref(20)
const q = ref('')
const cat = ref('all')
const sort = ref('latest') // 默认「最新入驻」：新入驻品牌可见性优先（2026-08-14 产品决策）
const showSort = ref(false)
const sIdx = ref(0)
const loading = ref(true)
const mockMode = ref(false)
const loadError = ref(false)
const loadingMore = ref(false)
const page = ref(1)
const hasMore = ref(false)
const totalCount = ref(0)
const list = ref([])
const fullList = ref([])
const featured = ref([])
const rawById = new Map()
let searchTimer = null
let fetchSeq = 0 // 请求序号：reset 使在途的分页响应作废，防旧结果覆盖新筛选

// ===== 数据映射（后端字段 → 展示字段，优雅降级） =====
const keyOf = (c) => {
  const s = String(c || '').toLowerCase()
  if (CATEGORY_MAP[s]) return s
  if (s.includes('整机') || s.includes('制造')) return 'drone'
  if (s.includes('零部件') || s.includes('配件')) return 'part'
  if (s.includes('飞控')) return 'flight_ctrl'
  if (s.includes('载荷')) return 'payload'
  if (s.includes('运营') || s.includes('服务')) return 'operator'
  if (s.includes('院校') || s.includes('学院') || s.includes('研究') || s.includes('学')) return 'college'
  if (s.includes('机场') || s.includes('通航')) return 'airport'
  if (s.includes('检测') || s.includes('机构')) return 'inspector'
  return 'drone'
}
// 相对路径（存库格式）→ 完整 URL
const resolveUrl = (u) => {
  if (!u) return ''
  if (u.indexOf('http') === 0) return u
  return BASE_URL + u
}

// 后端 MemberPortfolio：id/enterprise_id/name/logo_url/cover_url/description/
// products/honors/contact_info/status/created_at —— 前端按真实字段映射
const mapItem = (it) => {
  const honor = (it.honors && it.honors[0]) || ''
  const catKey = keyOf(it.category || honor) || 'brand'
  return {
    id: it.id,
    name: it.name || it.company_name || '',
    catKey,
    catLabel: CATEGORY_MAP[catKey] || honor || '会员品牌',
    char: it.name ? String(it.name).charAt(0) : '牌',
    logo: resolveUrl(it.logo_url || ''),
    cover: resolveUrl(it.cover_url || ''),
    logoText: '',
    verified: it.status === 'published', // 已公示 = 协会已认证
    hasVideo: false, // 后端暂无视频字段
    views: 0,
    videoCount: 0,
    grad: 'gd-' + String(it.grad || 'gd1').replace(/^gd-?/, ''),
  }
}
const mapFeatured = (f) => ({
  id: f.id,
  tag: f.tag || '精选推荐',
  name: f.name || '',
  sub: f.sub || '',
  logoText: f.logo_text || (f.name ? String(f.name).charAt(0) : '牌'),
  hasVideo: !!f.has_video,
  grad: String(f.grad || 'gd1').replace(/^gd-?/, ''),
})

// ===== 过滤与排序 =====
// 真实数据：q/category/sort 作为接口参数由后端执行；演示数据：前端本地过滤（同逻辑）
const applyFilter = () => {
  if (!mockMode.value) {
    list.value = fullList.value
    return
  }
  const kw = (q.value || '').trim().toLowerCase()
  let items = fullList.value.slice()
  if (kw) items = items.filter((x) => (x.name + ' ' + x.catLabel).toLowerCase().includes(kw))
  if (cat.value !== 'all') items = items.filter((x) => x.catKey === cat.value)
  if (sort.value === 'views') items.sort((a, b) => b.views - a.views)
  else if (sort.value === 'video') items.sort((a, b) => (b.hasVideo ? 1 : 0) - (a.hasVideo ? 1 : 0) || b.views - a.views)
  // latest：保持返回顺序（mock 中即入驻先后）
  list.value = items
  totalCount.value = items.length
}

const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新入驻')

// ===== 数据获取 =====
// ===== 接口替换点 =====
// GET /api/v1/portfolios (data: { q, category, sort, page, page_size })
// 后端就绪后自动使用真实数据；开发环境接口失败/为空时回退演示数据，生产环境绝不回退
const fetchList = async (reset = true) => {
  const seq = reset ? ++fetchSeq : fetchSeq
  if (reset) {
    page.value = 1
    hasMore.value = false
    loading.value = true
    loadError.value = false
    mockMode.value = false
  } else {
    loadingMore.value = true
  }
  try {
    const params = { page: page.value, page_size: PAGE_SIZE }
    if (q.value.trim()) params.q = q.value.trim()
    if (cat.value !== 'all') params.category = cat.value
    if (sort.value) params.sort = sort.value
    const res = await request({ url: '/api/v1/portfolios', data: params })
    if (seq !== fetchSeq) return // 在途期间发生了 reset，丢弃过期结果
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    const total = (data && data.total) != null ? data.total : items.length
    items.forEach((it) => rawById.set(it.id, it))
    if (reset) {
      // 精选位只在首屏请求刷新；接口未返回时生产留空（整块隐藏），开发回退演示精选
      // 注意：request.js 的分页信封 {data:[...], total} 会丢弃顶层 featured——
      // 契约需后端以对象信封 {data:{items,total,featured}} 返回，或将 featured 挂到数据数组上
      const feats = data && Array.isArray(data.featured) ? data.featured : []
      featured.value = (feats.length ? feats : (import.meta.env.DEV ? MOCK_FEATURED : [])).map(mapFeatured)
      if (items.length) {
        fullList.value = items.map(mapItem)
        totalCount.value = total
      } else if (import.meta.env.DEV) {
        useMock()
      } else {
        fullList.value = []
        totalCount.value = 0 // 真实空结果：走空态，不混入演示数据
      }
    } else {
      fullList.value = fullList.value.concat(items.map(mapItem))
      totalCount.value = total
    }
    hasMore.value = fullList.value.length < total
    applyFilter()
  } catch (e) {
    if (seq !== fetchSeq) return
    if (reset) {
      // 失败回退仅限开发环境；生产：首次加载走错误态，已有数据保留并提示
      if (import.meta.env.DEV) {
        useMock()
      } else if (!fullList.value.length) {
        loadError.value = true
      } else {
        uni.showToast({ title: '加载失败，已显示上次数据', icon: 'none' })
      }
    } else {
      page.value = Math.max(1, page.value - 1) // 失败回滚页码：下次触底重试同一页，不跳页
      uni.showToast({ title: '加载失败，请稍后重试', icon: 'none' })
    }
  } finally {
    if (seq === fetchSeq) {
      loading.value = false
      loadingMore.value = false
    }
  }
}

const useMock = () => {
  MOCK_BRANDS.forEach((m) => rawById.set(m.id, m))
  fullList.value = MOCK_BRANDS.map(mapItem)
  featured.value = MOCK_FEATURED.map(mapFeatured)
  mockMode.value = true
  hasMore.value = false
  loadError.value = false
}

// ===== 交互 =====
const onSlide = (e) => { sIdx.value = e.detail.current }
// 演示数据本地过滤即时生效；真实数据走防抖后的服务端查询
const onSearch = () => {
  clearTimeout(searchTimer)
  if (mockMode.value) searchTimer = setTimeout(applyFilter, SEARCH_DEBOUNCE_MS)
  else searchTimer = setTimeout(() => fetchList(true), SEARCH_DEBOUNCE_MS)
}
const clearSearch = () => {
  q.value = ''
  clearTimeout(searchTimer)
  if (mockMode.value) applyFilter()
  else fetchList(true)
}
const pickCat = (k) => {
  cat.value = k
  if (mockMode.value) applyFilter()
  else fetchList(true)
}
const toggleSort = () => { showSort.value = !showSort.value }
const pickSort = (v) => {
  sort.value = v
  showSort.value = false
  if (mockMode.value) applyFilter()
  else fetchList(true)
}
const resetAll = () => {
  q.value = ''
  cat.value = 'all'
  sort.value = 'latest'
  if (mockMode.value) applyFilter()
  else fetchList(true)
}
const retryLoad = () => fetchList(true)
const loadMore = async () => {
  page.value += 1 // 失败时由 fetchList 的 catch 回滚
  await fetchList(false)
}

const cacheRaw = (id) => {
  const raw = rawById.get(id)
  if (raw) uni.setStorageSync('portfolio_cache_' + id, raw)
}
const goDetail = (item) => {
  cacheRaw(item.id)
  uni.navigateTo({ url: '/pkg-eco/pages/portfolios/detail?id=' + encodeURIComponent(item.id) })
}
const goDetailById = (id) => {
  cacheRaw(id)
  uni.navigateTo({ url: '/pkg-eco/pages/portfolios/detail?id=' + encodeURIComponent(id) })
}
const goMine = () => {
  uni.navigateTo({ url: '/pkg-eco/pages/portfolios/mine' })
}
const goBack = () => uni.navigateBack()

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  fetchList(true)
})
onPullDownRefresh(async () => {
  await fetchList(true)
  uni.stopPullDownRefresh()
})
onReachBottom(() => {
  if (!loadingMore.value && hasMore.value) loadMore()
})
onUnload(() => { clearTimeout(searchTimer) })
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}
.tap-fade { opacity: .72; }

/* ===== 搜索 ===== */
.sbar { display: flex; align-items: center; padding: 24rpx 24rpx 20rpx; background: #fff; }
.sbox {
  flex: 1; display: flex; align-items: center;
  background: #F4F6F8; border: 1px solid #EEF1F4; border-radius: 16rpx;
  padding: 24rpx 24rpx; gap: 16rpx; /* 40 + 48 = 88rpx 触控达标 */
}
.sinp { flex: 1; font-size: 28rpx; color: #17212B; height: 40rpx; line-height: 40rpx; min-width: 0; }
.ph { color: #bbb; }
/* 热区 ≥88×88rpx（含最差行高假设）：padding 外扩 + 负 margin 补偿布局，视觉不变 */
.sclr { color: #98A2B3; font-size: 32rpx; padding: 25rpx 28rpx; margin: -25rpx -28rpx; flex-shrink: 0; }

/* ===== 精选横幅 ===== */
.feature-wrap { position: relative; margin: 24rpx 24rpx 8rpx; }
.feature-swiper { height: 280rpx; border-radius: 24rpx; overflow: hidden; }
.feature-card {
  position: relative; width: 100%; height: 100%;
  border-radius: 24rpx; overflow: hidden;
}
.f-inner {
  position: absolute; inset: 0;
  display: flex; align-items: center; gap: 28rpx;
  padding: 0 36rpx;
  background: linear-gradient(90deg, rgba(8,34,62,.55), rgba(8,34,62,0) 75%);
}
.f-logo {
  width: 104rpx; height: 104rpx; border-radius: 28rpx;
  background: rgba(255,255,255,.94);
  display: flex; align-items: center; justify-content: center;
  font-size: 44rpx; font-weight: 700; color: #17212B;
  flex: none; box-shadow: 0 4px 12px rgba(0,0,0,.2);
}
.f-txt { flex: 1; min-width: 0; }
.f-tag {
  display: inline-block; font-size: 19rpx; color: #fff;
  background: rgba(255,255,255,.22); padding: 4rpx 16rpx; border-radius: 8rpx;
  margin-bottom: 12rpx;
}
.f-name { font-size: 32rpx; font-weight: 700; color: #fff; line-height: 1.3; margin-bottom: 8rpx; display: block; }
.f-sub { font-size: 22rpx; color: rgba(255,255,255,.8); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: block; }
.f-play {
  position: absolute; right: 28rpx; bottom: 24rpx;
  width: 60rpx; height: 60rpx; border-radius: 50%;
  background: rgba(255,255,255,.95);
  display: flex; align-items: center; justify-content: center;
  color: #0A66C2; font-size: 22rpx; box-shadow: 0 2px 8px rgba(0,0,0,.25);
}
.f-dots { position: absolute; bottom: 20rpx; left: 50%; transform: translateX(-50%); display: flex; gap: 10rpx; }
.f-dot { width: 10rpx; height: 10rpx; border-radius: 50%; background: #fff; opacity: .4; transition: all .25s; }
.f-dot.on { opacity: 1; width: 28rpx; border-radius: 6rpx; }

/* ===== 分类 ===== */
/* 我的品牌入口（导航栏右侧） */
.mine-entry {
  position: relative;
  width: 72rpx; height: 72rpx; border-radius: 50%;
  background: #0A66C2; color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 26rpx; font-weight: 600;
  box-shadow: 0 2px 8px rgba(10,102,194,.3);
}
/* 触控热区扩至 88rpx（72rpx 圆钮 + 伪元素外扩） */
.mine-entry::after { content: ''; position: absolute; top: -8rpx; right: -8rpx; bottom: -8rpx; left: -8rpx; }

.tabs-scroll { background: #fff; white-space: nowrap; }
.tabs { display: inline-flex; gap: 16rpx; padding: 20rpx 24rpx 4rpx; }
.tab {
  position: relative;
  flex: none; padding: 14rpx 32rpx; border-radius: 16rpx;
  font-size: 25rpx; color: #344054;
  background: #F4F6F8; border: 1px solid #EEF1F4;
  transition: all .2s cubic-bezier(.25,.8,.3,1);
}
/* 触控热区扩至 88rpx：伪元素外扩，视觉与布局不变（背景 pill 不被撑大） */
.tab::after { content: ''; position: absolute; top: -16rpx; bottom: -16rpx; left: 0; right: 0; }
.tab.act { background: #0A66C2; border-color: #0A66C2; color: #fff; font-weight: 600; box-shadow: 0 3px 10px rgba(10,102,194,.3); }

/* ===== 信息行 ===== */
.ir { display: flex; justify-content: space-between; align-items: center; padding: 20rpx 32rpx 16rpx; font-size: 24rpx; color: #667085; position: relative; z-index: 20; }
.irn { color: #0A66C2; font-weight: 600; }
.irs-wrap { position: relative; }
.irs { color: #0A66C2; font-weight: 500; padding: 30rpx 16rpx; margin: -30rpx -16rpx; border-radius: 12rpx; } /* 热区 ≥88rpx（最差行高假设下 88.8），视觉不变 */
.irs:active { background: #EAF3FB; }
.spop {
  position: absolute; top: 60rpx; right: 0; z-index: 90;
  background: #fff; border-radius: 20rpx; box-shadow: 0 8px 28px rgba(0,0,0,.14);
  padding: 12rpx; min-width: 264rpx; animation: popIn .16s ease;
}
@keyframes popIn { from { opacity: 0; transform: translateY(-8rpx); } to { opacity: 1; transform: translateY(0); } }
.sp-opt { padding: 30rpx 24rpx; border-radius: 14rpx; font-size: 26rpx; color: #17212B; display: flex; align-items: center; justify-content: space-between; } /* 热区 ≥88rpx（最差行高假设下 91） */
.sp-opt:active { background: #F4F6F8; }
.sp-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }
.chk { color: #0A66C2; font-size: 24rpx; }

/* ===== 网格 ===== */
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; padding: 8rpx 24rpx 48rpx; }

/* ===== 骨架 ===== */
.sk-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; padding: 8rpx 24rpx; }
.sk-card { background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; overflow: hidden; }
.sk-cv { padding-top: 75%; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.sk-bd { padding: 20rpx; }
.sk-l { height: 24rpx; background: #f0f1f3; border-radius: 12rpx; margin-bottom: 16rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w50 { width: 50%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }
.sk-tip { grid-column: 1 / -1; text-align: center; padding: 12rpx; font-size: 22rpx; color: #98A2B3; }

/* ===== 空态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.sth { font-size: 24rpx; color: #ccc; margin: 24rpx 0; display: block; }
.stb { padding: 30rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; } /* 热区 ≥88rpx（最差行高假设下 91） */

/* ===== 加载更多 ===== */
.lm { text-align: center; padding: 4rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }

/* ===== 演示数据提示 ===== */
.mock-note { text-align: center; padding: 0 0 24rpx; font-size: 20rpx; color: #98A2B3; }

/* 封面渐变（占位，真实数据以 images 替换） */
.gd-1 { background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5); }
.gd-2 { background: linear-gradient(135deg,#004d40,#00695c 60%,#26a69a); }
.gd-3 { background: linear-gradient(135deg,#e65100,#ef6c00 60%,#fb8c00); }
.gd-4 { background: linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc); }
.gd-5 { background: linear-gradient(135deg,#263238,#37474f 60%,#607d8b); }
.gd-6 { background: linear-gradient(135deg,#b71c1c,#c62828 60%,#e57373); }
.gd-7 { background: linear-gradient(135deg,#1a237e,#283593 60%,#5c6bc0); }
.gd-8 { background: linear-gradient(135deg,#004d40,#00695c 60%,#4db6ac); }
</style>
