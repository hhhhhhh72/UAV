<template>
  <view class="page">
    <!-- ① 自定义导航栏（白底 + 返回 + 标题 + 胶囊占位） -->
    <view class="nav-wrap" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="nav-bar">
        <view class="nav-back" hover-class="nav-press" :hover-stay-time="100" @click="goBack">
          <text class="nav-back-icon">‹</text>
        </view>
        <text class="nav-title">赛事列表</text>
        <view class="nav-capsule">
          <view class="capsule-dot" />
          <view class="capsule-divider" />
          <view class="capsule-arrow" />
        </view>
      </view>
      <view class="nav-meta">
        <view class="meta-sync">
          <view class="sync-dot" />
          <text class="sync-text">已同步 · 中国 AOPA 赛事平台</text>
        </view>
        <view class="meta-all">
          <text class="meta-all-text">全部</text>
          <view class="meta-all-arrow" />
        </view>
      </view>
    </view>

    <!-- ② Tab + 搜索粘性区 -->
    <view class="sticky-head">
      <scroll-view scroll-x class="tabs-scroll" :show-scrollbar="false">
        <view class="tabs-inner">
          <view
            v-for="t in tabs"
            :key="t.key"
            class="tab-item"
            :class="{ on: currentTab === t.key }"
            @click="switchTab(t.key)"
          >
            <view class="tab-ico" :class="'tab-ico--' + t.ico" />
            <text class="tab-label">{{ t.label }}</text>
            <text v-if="tabCount(t.key)" class="tab-count">{{ tabCount(t.key) }}</text>
          </view>
        </view>
      </scroll-view>

      <view class="search-row">
        <view class="search-box" :class="{ focus: searchFocus }">
          <view class="search-ico"><view class="search-ring" /><view class="search-bar-ico" /></view>
          <input
            class="search-input"
            v-model="keyword"
            placeholder="搜索赛事名称 / 主办方 / 地点"
            placeholder-class="search-ph"
            confirm-type="search"
            @focus="searchFocus = true"
            @blur="searchFocus = false"
            @input="onSearch"
            @confirm="doSearch"
          />
          <text v-show="!keyword && !searchFocus" class="search-kbd">⌘K</text>
          <view v-if="keyword" class="search-clear" @click="clearKeyword"><text class="search-clear-x">×</text></view>
        </view>
        <view class="search-btn" hover-class="filter-press" :hover-stay-time="100" @click="doSearch">
          <text class="search-btn-text">搜索</text>
        </view>
      </view>
    </view>

    <!-- ③ 副标题行：实时统计 + 排序入口 -->
    <view class="sub-bar">
      <text class="sub-text">共 {{ filteredList.length }} 场 · {{ sortLabel }}</text>
      <view class="sort-pill" hover-class="sort-press" :hover-stay-time="100" @click="openSortSheet">
        <text class="sort-text">排序</text>
        <view class="sort-arrow" />
      </view>
    </view>

    <!-- ④ 卡片列表 -->
    <view class="content">
      <StateView
        class="state-fill"
        :loading="loading"
        :error="!!errorMsg"
        :empty="!loading && !errorMsg && filteredList.length === 0"
        empty-text="暂无赛事"
        @retry="loadData(true)"
      >
        <scroll-view class="list-scroll" scroll-y :show-scrollbar="false">
          <view
            v-for="(item, i) in filteredList"
            :key="item.id"
            class="card"
            :style="{ animationDelay: (i * 80) + 'ms' }"
            hover-class="card-press"
            :hover-stay-time="120"
            @click="goDetail(item)"
          >
            <!-- 顶部横幅 16:9 -->
            <view class="banner" :class="'banner--' + bannerType(item)">
              <!-- 真实海报图（有则显示） -->
              <image
                v-if="coverOf(item)"
                :src="coverOf(item)"
                class="banner-img"
                mode="aspectFill"
                lazy-load
                @load="onPosterLoad(item.id)"
                :style="{ opacity: imgLoaded[item.id] ? 1 : 0 }"
              />
              <!-- 无图兜底：深色渐变 + 线性无人机轮廓 -->
              <view v-else class="banner-fallback">
                <view class="drone-svg">
                  <view class="drone-prop p1" /><view class="drone-prop p2" /><view class="drone-prop p3" /><view class="drone-prop p4" />
                  <view class="drone-arm a1" /><view class="drone-arm a2" />
                  <view class="drone-body" />
                  <view class="drone-gimbal" />
                </view>
              </view>

              <!-- 底部渐变遮罩（保证徽章可读） -->
              <view class="banner-mask" />

              <!-- 左上状态徽章（五态） -->
              <view class="status-badge" :class="'badge--' + normStatus(item)">
                <view v-if="normStatus(item) === 'ongoing'" class="badge-dot" />
                <text class="badge-text">{{ statusText(item) }}</text>
              </view>

              <!-- 右上倒计时胶囊 -->
              <view class="countdown" :class="'cd--' + normStatus(item)">
                <view class="cd-ico" />
                <text class="cd-text">{{ countdownText(item) }}</text>
              </view>

              <!-- 底部信息行：级别徽标 + 叠层头像报名数 -->
              <view class="banner-foot">
                <view class="level-badge" :class="'level--' + levelOf(item)">
                  <text class="level-text">{{ levelText(item) }}</text>
                </view>
                <view class="reg-row">
                  <view class="avatars">
                    <view class="avatar av--1" />
                    <view class="avatar av--2" />
                    <view class="avatar av--3" />
                  </view>
                  <text class="reg-count">{{ regCount(item) }} 报名</text>
                </view>
              </view>
            </view>

            <!-- 卡片内容 -->
            <view class="card-info">
              <text class="card-title">{{ item.title || item.name || '未知赛事' }}</text>

              <view class="meta-grid">
                <view class="meta-cell">
                  <view class="meta-ico meta-ico--cal"><view class="cal-top" /><view class="cal-body"><view class="cal-line l1" /><view class="cal-line l2" /><view class="cal-line l3" /></view></view>
                  <text class="meta-text">{{ dateRange(item) }}</text>
                </view>
                <view class="meta-cell">
                  <view class="meta-ico meta-ico--loc"><view class="loc-pin" /></view>
                  <text class="meta-text ellipsis">{{ item.location || '待定' }}</text>
                </view>
                <view class="meta-cell meta-cell--wide">
                  <view class="meta-ico meta-ico--org"><view class="org-head" /><view class="org-body" /></view>
                  <text class="meta-text ellipsis">主办：{{ item.organizer || item.sponsor || '待定' }}</text>
                </view>
              </view>

              <view class="card-tags">
                <view v-for="t in compTags(item)" :key="t" class="pill" :class="tagCls(t)">
                  <view v-if="tagIco(t)" class="pill-ico" :class="'pill-ico--' + tagIco(t)" />
                  <text class="pill-text">{{ t }}</text>
                </view>
              </view>
            </view>

            <!-- 卡片底栏 -->
            <view class="card-foot">
              <view class="price-box">
                <template v-if="isFree(item)">
                  <text class="price-free">免费</text>
                </template>
                <template v-else>
                  <text class="price-symbol">¥</text>
                  <text class="price-num">{{ feeText(item) }}</text>
                  <text v-if="origFee(item)" class="price-orig">¥{{ origFee(item) }}</text>
                  <text class="price-unit">/人</text>
                </template>
              </view>
              <view
                class="cta-btn"
                :class="ctaFor(item).cls"
                :disabled="ctaFor(item).disabled"
                hover-class="cta-press"
                :hover-stay-time="100"
                @click.stop="onCta(item)"
              >
                <text class="cta-text">{{ ctaFor(item).text }}</text>
              </view>
            </view>
          </view>

          <view v-if="filteredList.length > 0 && !hasMore" class="load-more-wrap">
            <text class="no-more">没有更多了</text>
          </view>
          <view style="height:48rpx" />
        </scroll-view>
      </StateView>
    </view>

    <!-- ⑤ 自定义 Toast -->
    <view v-if="toast.show" class="custom-toast" :class="{ 'custom-toast--out': toast.hide }">
      <view class="toast-icon"><view class="toast-check" /></view>
      <text class="toast-text">{{ toast.msg }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

// 状态栏高度：微信端 CSS 变量 --status-bar-height 不生效，须 JS 读取动态设置
const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 20

/* ===== 展示态 ===== */
const currentTab = ref('all')
const keyword = ref('')
const searchFocus = ref(false)
const sortKey = ref('dateAsc')

/* ===== 数据 ===== */
const loading = ref(false)
const errorMsg = ref('')
const allList = ref([])   // 全量（已加载）
const hasMore = ref(false)

/* 图片加载 */
const imgLoaded = ref({})

/* Toast */
const toast = ref({ show: false, hide: false, msg: '' })
let toastTimer = null
let toastOutTimer = null

/* ===== Tab 定义 ===== */
const tabs = [
  { key: 'all', label: '全部赛事', ico: 'all' },
  { key: 'enrolling', label: '报名中', ico: 'signal' },
  { key: 'upcoming', label: '即将开始', ico: 'cal' },
  { key: 'ongoing', label: '进行中', ico: 'clock' },
  { key: 'closed', label: '已结束', ico: 'dot' },
]

/* ===== 前端过滤 + 排序（Tab/搜索/排序都不动后端） ===== */
const filteredList = computed(function () {
  var base = allList.value.slice()
  // 1) 分类 Tab 过滤
  if (currentTab.value !== 'all') {
    base = base.filter(function (c) { return normStatus(c) === currentTab.value })
  }
  // 2) 关键词搜索（名称/主办方/地点）
  var kw = keyword.value.trim()
  if (kw) {
    var lower = kw.toLowerCase()
    base = base.filter(function (c) {
      var t = ((c.title || '') + ' ' + (c.organizer || c.sponsor || '') + ' ' + (c.location || '')).toLowerCase()
      return t.indexOf(lower) >= 0
    })
  }
  // 3) 排序
  sortList(base)
  return base
})

/* 排序规则 */
const SORTERS = {
  dateAsc: function (a, b) { return cmpDate(a, b, 1) },    // 开赛时间正序
  dateDesc: function (a, b) { return cmpDate(a, b, -1) },  // 开赛时间倒序
  feeAsc: function (a, b) { return feeOf(a) - feeOf(b) },  // 价格从低到高
  feeDesc: function (a, b) { return feeOf(b) - feeOf(a) }, // 价格从高到低
  regDesc: function (a, b) { return regNum(b) - regNum(a) }, // 报名人数从多到少
}
const SORT_LABELS = {
  dateAsc: '开赛时间正序',
  dateDesc: '开赛时间倒序',
  feeAsc: '价格从低到高',
  feeDesc: '价格从高到低',
  regDesc: '报名人数最多',
}

const sortLabel = computed(function () { return SORT_LABELS[sortKey.value] || '开赛时间正序' })

function sortList(arr) {
  var sorter = SORTERS[sortKey.value]
  if (sorter) arr.sort(sorter)
}

/* 按开始时间比较（日期空则排最后） */
function cmpDate(a, b, dir) {
  var ta = dateStamp(a.start_date), tb = dateStamp(b.start_date)
  if (ta == null && tb == null) return 0
  if (ta == null) return 1
  if (tb == null) return -1
  return (ta - tb) * dir
}
function dateStamp(s) {
  var d = parseDate(s)
  return d ? d.getTime() : null
}

function tabCount(key) {
  if (key === 'all') return allList.value.length || 0
  return allList.value.filter(function (c) { return normStatus(c) === key }).length
}

/* ===== 状态归一化（五态） ===== */
function normStatus(item) {
  var s = item.status
  if (s === 'open' || s === 'enrolling' || s === 'registration_open') return 'enrolling'
  if (s === 'upcoming' || s === 'not_started') return 'upcoming'
  if (s === 'ongoing' || s === 'live') return 'ongoing'
  if (s === 'full' || s === 'deadline') return 'closed'
  return s || 'enrolling'
}

function statusText(item) {
  var map = { enrolling: '报名中', upcoming: '即将开始', ongoing: '进行中', closed: '已结束' }
  return map[normStatus(item)] || '报名中'
}

function isClosed(item) {
  return normStatus(item) === 'closed'
}

/* ===== 倒计时 ===== */
function countdownText(item) {
  var st = normStatus(item)
  if (st === 'ongoing') return '进行中'
  if (st === 'closed') return '已闭幕'
  var d = parseDate(item.start_date)
  if (!d) return '距开赛'
  var diff = Math.ceil((d.getTime() - Date.now()) / 86400000)
  if (diff < 0) return '已闭幕'
  return '距开赛 ' + diff + ' 天'
}

function parseDate(s) {
  if (!s) return null
  // 支持 "2026.09.15" / "2026-09-15" / "2026/09/15"
  var m = String(s).match(/(\d{4})[.\-\/](\d{1,2})[.\-\/](\d{1,2})/)
  if (!m) return null
  return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
}

/* ===== 级别徽标 ===== */
function levelOf(item) {
  var t = (item.title || '') + ' ' + (item.level || '') + ' ' + (item.tags || []).join(' ')
  if (t.indexOf('国家级') >= 0 || t.indexOf('全国') >= 0 || t.indexOf('国际') >= 0) return 'national'
  if (t.indexOf('省级') >= 0 || t.indexOf('西南') >= 0 || t.indexOf('华南') >= 0) return 'province'
  return 'city'
}

function levelText(item) {
  var map = { national: '国家级', province: '省级', city: '城市级' }
  return map[levelOf(item)]
}

/* ===== 报名人数 + 叠层头像 ===== */
function regNum(item) {
  return Number(item.registration_count || item.participant_count || item.reg_count || 0)
}
function regCount(item) {
  return String(regNum(item))
}

/* ===== 价格 ===== */
function feeOf(item) {
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  if (item.price != null) return item.price
  return 0
}
function feeText(item) {
  return String(feeOf(item)).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}
function origFee(item) {
  var o = item.original_fee
  if (o == null && item.original_price != null) o = item.original_price
  return o && o > feeOf(item) ? o : null
}
function isFree(item) {
  return feeOf(item) <= 0
}

/* ===== 横幅类型（深色渐变） ===== */
function bannerType(item) {
  var st = normStatus(item)
  if (st === 'closed') return 'closed'
  var t = item.title || ''
  if (t.indexOf('竞速') >= 0 || t.indexOf('FPV') >= 0) return 'racing'
  if (t.indexOf('物流') >= 0 || t.indexOf('配送') >= 0) return 'logistics'
  if (t.indexOf('植保') >= 0) return 'agri'
  if (t.indexOf('巡检') >= 0 || t.indexOf('电力') >= 0) return 'inspect'
  return 'default'
}

/* ===== CTA 按钮三态 ===== */
function ctaFor(item) {
  var st = normStatus(item)
  if (st === 'enrolling') return { text: '立即报名', cls: 'cta--primary', disabled: false }
  if (st === 'upcoming') return { text: '报名提醒', cls: 'cta--blue', disabled: false }
  if (st === 'ongoing') return { text: '查看直播', cls: 'cta--blue', disabled: false }
  return { text: '查看回顾', cls: 'cta--disabled', disabled: true }
}

/* ===== 数据映射 ===== */
function dateRange(item) {
  var s = fmtDate(item.start_date)
  var e = fmtDate(item.end_date)
  if (s === '待定' && e === '待定') return '日期待定'
  return e && e !== s ? s + ' - ' + e : s
}

function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags.slice(0, 3)
  var tags = []
  if (item.category) tags.push(item.category)
  if (tags.length === 0) tags = ['多旋翼', '国家级']
  return tags
}

function tagCls(tag) {
  var lv = ['国家级', '国际赛', '国际', '省级', '城市级']
  if (lv.indexOf(tag) >= 0) return 'pill--level'
  if (tag.indexOf('竞速') >= 0 || tag.indexOf('FPV') >= 0) return 'pill--warn'
  if (tag.indexOf('物流') >= 0 || tag.indexOf('植保') >= 0) return 'pill--ok'
  return 'pill--model'
}

/* 标签前线性图标：巡检/电力=闪电，物流=纸箱，其余=无人机 */
function tagIco(tag) {
  if (tag.indexOf('巡检') >= 0 || tag.indexOf('电力') >= 0 || tag.indexOf('能源') >= 0) return 'bolt'
  if (tag.indexOf('物流') >= 0 || tag.indexOf('配送') >= 0 || tag.indexOf('运输') >= 0) return 'box'
  return 'drone'
}

function introOf(item) {
  return item.intro || item.description || item.summary || ''
}

/** 封面图 URL：兼容 poster / cover / image / banner */
function coverOf(item) {
  var u = item.poster || item.cover || item.image || item.banner
  return u ? u : ''
}

function onPosterLoad(id) {
  imgLoaded.value[id] = true
}

/* ===== 数据获取（一次拉全量，前端过滤） ===== */
async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { loading.value = true }
  errorMsg.value = ''

  try {
    var params = { page: 1, page_size: 100 }
    if (keyword.value) params.keyword = keyword.value

    var res = await request({ url: '/api/v1/competitions', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []

    if (items.length === 0) {
      items = getMockList()
    }
    allList.value = items
    hasMore.value = false
  } catch (e) {
    allList.value = getMockList()
    hasMore.value = false
  } finally {
    loading.value = false
  }
}

function getMockList() {
  return [
    {
      id: 'comp-1', title: '2026 全国无人机职业技能大赛', status: 'enrolling',
      location: '深圳 · 宝安国际会展中心', organizer: '中国航空器拥有者及驾驶员协会',
      start_date: '2026-09-15', end_date: '2026-09-18',
      tags: ['多旋翼', '固定翼', '国家级'], fee: 380, original_fee: 480, registration_count: 128,
      poster: '',
    },
    {
      id: 'comp-2', title: '首届西南无人机 FPV 竞速挑战赛', status: 'enrolling',
      location: '成都 · 天府新区无人机竞速基地', organizer: '四川省航空运动协会',
      start_date: '2026-10-01', end_date: '2026-10-03',
      tags: ['竞速FPV', '多旋翼', '省级'], fee: 280, original_fee: 360, registration_count: 56,
      poster: '',
    },
    {
      id: 'comp-3', title: '低空物流配送实战演练赛', status: 'ongoing',
      location: '杭州 · 余杭未来科技城', organizer: '杭州市低空经济产业协会',
      start_date: '2026-08-01', end_date: '2026-08-15',
      tags: ['物流配送', '多旋翼', '城市级'], fee: 0, registration_count: 256,
      poster: '/static/home/demand-lift.jpg',
    },
    {
      id: 'comp-4', title: '第十届植保无人机飞防作业大赛', status: 'upcoming',
      location: '郑州 · 黄河农场', organizer: '河南省植保技术推广站',
      start_date: '2026-11-01', end_date: '2026-11-02',
      tags: ['植保飞防', '多旋翼', '国家级'], fee: 580, original_fee: 720, registration_count: 340,
      poster: '',
    },
    {
      id: 'comp-5', title: '2025 长三角无人机巡检技能赛', status: 'closed',
      location: '苏州 · 工业园区', organizer: '长三角低空经济发展联盟',
      start_date: '2025-12-05', end_date: '2025-12-07',
      tags: ['电力巡检', '固定翼', '省级'], fee: 220, registration_count: 89,
      poster: '/static/home/hero-inspection.jpg',
    },
  ]
}

/* ===== Tab / 搜索 / 排序（纯前端） ===== */
function switchTab(tab) {
  if (currentTab.value === tab) return
  currentTab.value = tab
}

var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { /* 输入驱动 computed 实时过滤 */ }, 0)
}

function doSearch() {
  clearTimeout(searchTimer)
}

function clearKeyword() {
  keyword.value = ''
}

/* 排序选项（ActionSheet） */
function openSortSheet() {
  var labels = ['按开赛时间正序', '按开赛时间倒序', '价格从低到高', '价格从高到低', '报名人数最多']
  uni.showActionSheet({
    itemList: labels,
    success: function (res) {
      var keys = ['dateAsc', 'dateDesc', 'feeAsc', 'feeDesc', 'regDesc']
      sortKey.value = keys[res.tapIndex] || 'dateAsc'
    },
  })
}

/* ===== 交互 ===== */
function goDetail(item) {
  // 后端无公开单查接口，详情页只能靠"拉全量按 id 匹配"。
  // 先把完整卡片数据存 storage，详情页优先读取，保证跳转后内容与卡片一致。
  uni.setStorageSync('competition_detail', item)
  uni.navigateTo({ url: '/pkg-eco/pages/competitions/detail?id=' + encodeURIComponent(item.id) })
}

function onCta(item) {
  var st = normStatus(item)
  if (st === 'enrolling') {
    uni.navigateTo({ url: '/pkg-eco/pages/competitions/register?id=' + encodeURIComponent(item.id) })
    return
  }
  if (st === 'upcoming') { showCustomToast('已设置开赛提醒'); return }
  if (st === 'ongoing') { showCustomToast('直播功能即将上线'); return }
  showCustomToast('赛事已结束，回顾制作中')
}

function showCustomToast(msg) {
  clearTimeout(toastTimer)
  clearTimeout(toastOutTimer)
  toast.value = { show: true, hide: false, msg: msg }
  toastTimer = setTimeout(function () {
    toast.value.hide = true
    toastOutTimer = setTimeout(function () {
      toast.value.show = false
    }, 200)
  }, 2000)
}

function goBack() {
  uni.navigateBack({ delta: 1 })
}

onLoad(function () { loadData(true) })

onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
  min-height: 100vh;
  background: #F4F6F8;
  display: flex;
  flex-direction: column;
  padding-left: constant(safe-area-inset-left);
  padding-left: env(safe-area-inset-left);
  padding-right: constant(safe-area-inset-right);
  padding-right: env(safe-area-inset-right);
  overflow-x: hidden;
}

/* ═══ ① 自定义导航栏 ═══ */
.nav-wrap {
  background: #ffffff;
  padding: var(--status-bar-height) 0 0;
  position: relative;
  z-index: 5;
  border-bottom: 1rpx solid #EEF1F4;
}
.nav-bar {
  display: flex;
  align-items: center;
  height: 88rpx;
  padding: 0 24rpx;
}
.nav-back {
  width: 64rpx;
  height: 64rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.nav-press { background: #F4F6F8; }
.nav-back-icon { font-size: 40rpx; color: #344054; font-weight: 300; line-height: 1; }
.nav-title { flex: 1; text-align: center; font-size: 34rpx; font-weight: 760; color: #17212B; }
.nav-capsule {
  width: 176rpx;
  height: 60rpx;
  border: 1rpx solid #E4E7EC;
  border-radius: 999rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  flex-shrink: 0;
}
.capsule-dot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #98A2B3; }
.capsule-divider { width: 1rpx; height: 28rpx; background: #E4E7EC; }
.capsule-arrow {
  width: 0;
  height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 8rpx solid #344054;
}
.nav-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24rpx 16rpx;
}
.meta-sync { display: flex; align-items: center; gap: 8rpx; }
.sync-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #168A55;
  box-shadow: 0 0 0 0 rgba(22, 138, 85, 0.5);
  animation: syncPulse 1.8s ease-out infinite;
}
.sync-text { font-size: 20rpx; color: #667085; }
.meta-all { display: flex; align-items: center; gap: 6rpx; }
.meta-all-text { font-size: 22rpx; font-weight: 600; color: #344054; }
.meta-all-arrow {
  width: 0;
  height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 7rpx solid #98A2B3;
}

/* ═══ ② Tab + 搜索粘性区 ═══ */
.sticky-head {
  background: #ffffff;
  position: sticky;
  top: 0;
  z-index: 4;
  border-bottom: 1rpx solid #EEF1F4;
}
.tabs-scroll { white-space: nowrap; padding: 16rpx 0 8rpx; }
.tabs-inner { display: inline-flex; gap: 12rpx; padding: 0 24rpx; }
.tab-item {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  height: 60rpx;
  padding: 0 24rpx;
  border-radius: 999rpx;
  background: #F4F6F8;
  color: #344054;
  transition: background 180ms var(--ease), color 180ms var(--ease), box-shadow 180ms var(--ease);
}
.tab-item.on {
  background: #0A66C2;
  color: #ffffff;
  box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28);
}
.tab-label { font-size: 24rpx; font-weight: 600; }
.tab-count { font-size: 20rpx; font-weight: 600; opacity: 0.75; }
.tab-ico { width: 24rpx; height: 24rpx; position: relative; flex-shrink: 0; }
.tab-ico--all { border: 2rpx solid currentColor; border-radius: 4rpx; box-sizing: border-box; }
.tab-ico--signal::before {
  content: '';
  position: absolute;
  left: 6rpx; bottom: 0;
  width: 2rpx; height: 10rpx;
  background: currentColor;
  border-radius: 2rpx;
  transform: rotate(-30deg);
}
.tab-ico--signal::after {
  content: '';
  position: absolute;
  right: 6rpx; bottom: 0;
  width: 2rpx; height: 10rpx;
  background: currentColor;
  border-radius: 2rpx;
  transform: rotate(30deg);
}
.tab-ico--cal { border: 2rpx solid currentColor; border-radius: 4rpx; box-sizing: border-box; }
.tab-ico--cal::after {
  content: '';
  position: absolute;
  left: 4rpx; right: 4rpx; top: 6rpx; bottom: 0;
  border-top: 1rpx solid currentColor;
}
.tab-ico--clock { border: 2rpx solid currentColor; border-radius: 50%; box-sizing: border-box; }
.tab-ico--clock::after {
  content: '';
  position: absolute;
  left: 10rpx; top: 4rpx;
  width: 2rpx; height: 8rpx;
  background: currentColor;
}
.tab-ico--dot { border: 3rpx solid currentColor; border-radius: 50%; box-sizing: border-box; }

/* 搜索行 */
.search-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 8rpx 24rpx 16rpx;
}
.search-box {
  flex: 1;
  height: 76rpx;
  background: #F4F6F8;
  border: 1rpx solid #EDF0F5;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  padding: 0 20rpx;
  gap: 10rpx;
  transition: background 200ms var(--ease), border-color 200ms var(--ease), box-shadow 200ms var(--ease);
}
.search-box.focus {
  background: #ffffff;
  border-color: #0A66C2;
  box-shadow: 0 0 0 3px rgba(10, 102, 194, 0.12);
}
.search-ico { position: relative; width: 26rpx; height: 26rpx; flex-shrink: 0; }
.search-ring {
  width: 16rpx;
  height: 16rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50%;
  transition: border-color 200ms var(--ease);
}
.search-bar-ico {
  position: absolute;
  right: 0;
  bottom: 2rpx;
  width: 8rpx;
  height: 3rpx;
  background: #98A2B3;
  transform: rotate(45deg);
  transform-origin: right center;
  transition: background 200ms var(--ease);
}
.search-box.focus .search-ring { border-color: #0A66C2; }
.search-box.focus .search-bar-ico { background: #0A66C2; }
.search-input { flex: 1; font-size: 26rpx; color: #17212B; }
.search-ph { color: #98A2B3; }
.search-kbd {
  font-size: 18rpx;
  color: #98A2B3;
  border: 1rpx solid #E4E7EC;
  border-radius: 4rpx;
  padding: 2rpx 10rpx;
  transition: opacity 200ms var(--ease);
}
.search-box.focus .search-kbd { opacity: 0; }
.search-clear {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #EDF0F5;
  display: flex;
  align-items: center;
  justify-content: center;
}
.search-clear-x { font-size: 28rpx; color: #667085; line-height: 1; }
.search-btn {
  height: 76rpx;
  padding: 0 28rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0A66C2;
  flex-shrink: 0;
  transition: transform 180ms var(--ease), background 180ms var(--ease);
}
.search-btn:active { background: #074D92; }
.search-btn-text { font-size: 24rpx; font-weight: 700; color: #ffffff; }

/* ═══ ③ 副标题行 ═══ */
.sub-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 32rpx 4rpx;
}
.sub-text { font-size: 22rpx; color: #667085; }
.sort-pill { display: flex; align-items: center; gap: 6rpx; padding: 8rpx 16rpx; border-radius: 6rpx; }
.sort-press { background: #F4F6F8; }
.sort-text { font-size: 22rpx; font-weight: 600; color: #344054; }
.sort-arrow {
  width: 0;
  height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 7rpx solid #98A2B3;
}

/* ═══ ④ 卡片列表 ═══ */
.content { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.state-fill { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.list-scroll { padding: 2rpx 24rpx 0; flex: 1; min-height: 0; box-sizing: border-box; }

.card {
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 10px;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  overflow: hidden;
  margin-bottom: 24rpx;
  animation: cardIn 460ms var(--ease) both;
  transition: transform 180ms var(--ease), box-shadow 180ms var(--ease);
}
.card-press {
  transform: scale(0.985);
  box-shadow: 0 6px 18px rgba(16, 24, 40, 0.08);
}

/* 横幅 16:9 */
.banner { position: relative; width: 100%; height: 0; padding-bottom: 56.25%; overflow: hidden; }
.banner--default { background: linear-gradient(160deg, #0a5897 0%, #074D92 100%); }
.banner--inspect { background: linear-gradient(160deg, #0f2f5c 0%, #0a5897 100%); }
.banner--racing { background: linear-gradient(160deg, #1c4d3c 0%, #0f3a2d 100%); }
.banner--logistics { background: linear-gradient(160deg, #b45309 0%, #92400e 100%); }
.banner--agri { background: linear-gradient(160deg, #4a5d23 0%, #38471a 100%); }
.banner--closed { background: linear-gradient(160deg, #55606e 0%, #3d4652 100%); }

.banner-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity 240ms ease-out;
}
.banner-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 线性无人机轮廓 */
.drone-svg { position: relative; width: 120rpx; height: 90rpx; opacity: 0.9; }
.drone-prop {
  position: absolute;
  width: 34rpx;
  height: 34rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.65);
  border-radius: 50%;
  box-sizing: border-box;
}
.drone-prop.p1 { left: 0; top: 0; }
.drone-prop.p2 { right: 0; top: 0; }
.drone-prop.p3 { left: 0; bottom: 0; }
.drone-prop.p4 { right: 0; bottom: 0; }
.drone-arm {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 86rpx;
  height: 3rpx;
  background: rgba(255, 255, 255, 0.4);
  transform-origin: center;
}
.drone-arm.a1 { transform: translate(-50%, -50%) rotate(-45deg); }
.drone-arm.a2 { transform: translate(-50%, -50%) rotate(45deg); }
.drone-body {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 44rpx;
  height: 28rpx;
  margin: -14rpx 0 0 -22rpx;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 8rpx;
}
.drone-gimbal {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 18rpx;
  height: 18rpx;
  margin: 16rpx 0 0 -9rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.65);
  border-radius: 50%;
  box-sizing: border-box;
}

.banner-mask {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 40%;
  background: linear-gradient(180deg, rgba(7, 77, 146, 0) 30%, rgba(7, 77, 146, 0.55) 100%);
  pointer-events: none;
}

/* 状态徽章（五态） */
.status-badge {
  position: absolute;
  top: 12rpx;
  left: 12rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 6rpx;
}
.badge-text { font-size: 20rpx; font-weight: 600; }
.badge--enrolling { background: rgba(255, 255, 255, 0.92); color: #E96012; }
.badge--upcoming { background: rgba(255, 255, 255, 0.92); color: #0A66C2; }
.badge--ongoing { background: #168A55; color: #ffffff; }
.badge--ongoing .badge-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: currentColor;
  position: relative;
}
.badge--ongoing .badge-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: currentColor;
  animation: badgeRing 1.4s ease-out infinite;
}
.badge--closed { background: rgba(16, 24, 40, 0.55); color: #ffffff; }

/* 倒计时胶囊 */
.countdown {
  position: absolute;
  top: 12rpx;
  right: 12rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 6rpx;
  background: rgba(7, 77, 146, 0.85);
  color: #ffffff;
}
.cd-ico {
  width: 20rpx;
  height: 20rpx;
  border: 2rpx solid currentColor;
  border-radius: 50%;
  box-sizing: border-box;
  position: relative;
  flex-shrink: 0;
}
.cd-ico::after {
  content: '';
  position: absolute;
  left: 8rpx;
  top: 3rpx;
  width: 2rpx;
  height: 6rpx;
  background: currentColor;
}
.cd-text { font-size: 18rpx; font-weight: 600; }
.cd--ongoing { background: #168A55; }
.cd--closed { background: rgba(16, 24, 40, 0.55); }

/* 底部信息行 */
.banner-foot {
  position: absolute;
  left: 12rpx;
  right: 12rpx;
  bottom: 10rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.level-badge {
  padding: 4rpx 12rpx;
  border-radius: 6rpx;
  background: rgba(255, 255, 255, 0.88);
  transition: transform 180ms cubic-bezier(0.2, 0.8, 0.2, 1);
}
/* hover 轻微上浮（与倒计时胶囊拉开距离的细节感） */
.card:active .level-badge { transform: translateY(-4rpx); }
.level-text { font-size: 18rpx; font-weight: 600; }
.level--national .level-text { color: #E96012; }
.level--province .level-text { color: #0A66C2; }
.level--city .level-text { color: #168A55; }

.reg-row { display: flex; align-items: center; gap: 8rpx; }
.avatars { display: flex; }
.avatar {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #ffffff;
  box-sizing: border-box;
}
.avatar + .avatar { margin-left: -10rpx; }
.av--1 { background: #0A66C2; }
.av--2 { background: #F97316; }
.av--3 { background: #168A55; }
.reg-count { font-size: 18rpx; font-weight: 500; color: #ffffff; text-shadow: 0 1rpx 2rpx rgba(0, 0, 0, 0.3); }

/* 卡片内容 */
.card-info { padding: 20rpx 28rpx 4rpx; }
.card-title {
  display: block;
  font-size: 32rpx;
  font-weight: 720;
  color: #17212B;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.meta-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx 0;
  margin-top: 14rpx;
}
.meta-cell {
  width: 50%;
  display: flex;
  align-items: center;
  gap: 8rpx;
  min-width: 0;
}
.meta-cell--wide { width: 100%; }
.meta-ico { width: 28rpx; height: 28rpx; flex-shrink: 0; position: relative; }
/* 圆角矩形日历：外框 + 顶部挂环 + 中部横线 + 底部刻度 */
.meta-ico--cal {
  border: 2rpx solid #98A2B3;
  border-radius: 4rpx;
  box-sizing: border-box;
}
.meta-ico--cal::before,
.meta-ico--cal::after {
  content: '';
  position: absolute;
  top: 4rpx;
  width: 2rpx;
  height: 5rpx;
  background: #98A2B3;
}
.meta-ico--cal::before { left: 6rpx; }
.meta-ico--cal::after { right: 6rpx; }
.meta-ico--cal .cal-top {
  position: absolute;
  left: 3rpx; right: 3rpx; top: 8rpx;
  height: 2rpx;
  background: #98A2B3;
}
.meta-ico--cal .cal-line {
  position: absolute;
  left: 5rpx; right: 5rpx;
  height: 2rpx;
  background: #98A2B3;
  opacity: 0.6;
}
.meta-ico--cal .cal-line.l1 { top: 14rpx; }
.meta-ico--cal .cal-line.l2 { top: 19rpx; }
.meta-ico--cal .cal-line.l3 { top: 24rpx; }
/* 定位针：泪滴轮廓 + 中心实心点 */
.meta-ico--loc { width: 22rpx; }
.meta-ico--loc .loc-pin {
  position: absolute;
  left: 50%;
  top: 4rpx;
  width: 14rpx;
  height: 14rpx;
  margin-left: -7rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  box-sizing: border-box;
}
.meta-ico--loc .loc-pin::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 4rpx;
  height: 4rpx;
  margin: -2rpx 0 0 -2rpx;
  border-radius: 50%;
  background: #98A2B3;
}
/* 机构轮廓：圆头 + 肩线 */
.meta-ico--org { width: 24rpx; }
.meta-ico--org .org-head {
  position: absolute;
  left: 50%;
  top: 2rpx;
  width: 8rpx;
  height: 8rpx;
  margin-left: -4rpx;
  border: 2rpx solid #98A2B3;
  border-radius: 50%;
  box-sizing: border-box;
}
.meta-ico--org .org-body {
  position: absolute;
  left: 50%;
  top: 13rpx;
  width: 14rpx;
  height: 9rpx;
  margin-left: -7rpx;
  border: 2rpx solid #98A2B3;
  border-top: none;
  border-radius: 0 0 8rpx 8rpx;
  box-sizing: border-box;
}
.meta-text { font-size: 22rpx; color: #667085; }
.meta-text.ellipsis {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 12rpx;
}
/* 标签：11px 线性图标 + 10px 文字 */
.pill {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  padding: 4rpx 12rpx;
  border-radius: 4rpx;
  font-size: 20rpx;
  font-weight: 600;
}
.pill-ico { width: 22rpx; height: 22rpx; flex-shrink: 0; position: relative; }
/* 无人机（多旋翼默认） */
.pill-ico--drone::before {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  width: 10rpx; height: 10rpx;
  margin: -5rpx 0 0 -5rpx;
  border: 2rpx solid currentColor;
  border-radius: 3rpx;
  box-sizing: border-box;
}
.pill-ico--drone::after {
  content: '';
  position: absolute;
  left: 2rpx; top: 50%;
  width: 6rpx; height: 6rpx;
  margin-top: -3rpx;
  border: 1.5rpx solid currentColor;
  border-radius: 50%;
  box-sizing: border-box;
  box-shadow: 12rpx 0 0 -1rpx currentColor;
}
/* 闪电（巡检/电力） */
.pill-ico--bolt {
  width: 12rpx;
  height: 20rpx;
  margin: 0 auto;
  background: currentColor;
  clip-path: polygon(55% 0, 100% 0, 40% 55%, 65% 55%, 0 100%, 25% 45%, 0 45%);
  opacity: 0.85;
}
/* 纸箱（物流） */
.pill-ico--box {
  border: 2rpx solid currentColor;
  border-radius: 2rpx;
  box-sizing: border-box;
}
.pill-ico--box::before {
  content: '';
  position: absolute;
  left: -2rpx; right: -2rpx; top: 6rpx;
  height: 2rpx;
  background: currentColor;
}
.pill--model { background: #EAF3FB; color: #0A66C2; }
.pill--warn { background: #FEF6E7; color: #B54708; }
.pill--ok { background: #E9F7F0; color: #168A55; }
.pill--level { background: #FFF0E6; color: #E96012; }
.pill {
  animation: pillIn 360ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
}
.pill:nth-child(2) { animation-delay: 60ms; }
.pill:nth-child(3) { animation-delay: 120ms; }
@keyframes pillIn {
  from { opacity: 0; transform: scale(0.8); }
  to { opacity: 1; transform: scale(1); }
}

/* 卡片底栏 */
.card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 28rpx 20rpx;
}
.price-box { display: flex; align-items: baseline; }
.price-symbol { font-size: 22rpx; font-weight: 700; color: #E96012; }
.price-num {
  font-size: 44rpx;
  font-weight: 760;
  color: #E96012;
  line-height: 1;
  animation: priceIn 500ms var(--ease) both;
}
.price-orig {
  font-size: 20rpx;
  color: #98A2B3;
  margin-left: 8rpx;
  text-decoration: line-through;
}
.price-unit { font-size: 20rpx; color: #98A2B3; margin-left: 4rpx; }
.price-free { font-size: 44rpx; font-weight: 760; color: #168A55; line-height: 1; }

.cta-btn {
  height: 64rpx;
  padding: 0 28rpx;
  border-radius: 6rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 180ms var(--ease), background 180ms var(--ease);
}
.cta-press { transform: scale(0.97); }
.cta-text { font-size: 24rpx; font-weight: 700; color: #ffffff; }
.cta--primary { background: #F97316; box-shadow: 0 4rpx 10rpx rgba(249, 115, 22, 0.32); }
.cta--primary:active { background: #E96012; }
.cta--blue { background: #0A66C2; box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.24); }
.cta--blue:active { background: #074D92; }
.cta--disabled { background: #EEF1F4; opacity: 0.5; pointer-events: none; }
.cta--disabled .cta-text { color: #667085; }

/* 加载更多 */
.load-more-wrap { text-align: center; padding: 24rpx 0; }
.no-more { font-size: 22rpx; color: #98A2B3; }

/* ═══ ⑤ 自定义 Toast ═══ */
.custom-toast {
  position: fixed;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  z-index: 999;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 32rpx;
  background: rgba(16, 24, 40, 0.92);
  border-radius: 10rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.24);
  animation: toastIn 250ms var(--ease) both;
  max-width: 70vw;
}
.custom-toast--out { animation: toastOut 200ms ease both; }
.toast-icon {
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: rgba(91, 255, 176, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.toast-check {
  width: 16rpx;
  height: 9rpx;
  border-left: 3rpx solid #5BFFB0;
  border-bottom: 3rpx solid #5BFFB0;
  transform: rotate(-45deg) translate(1rpx, -1rpx);
}
.toast-text { font-size: 26rpx; color: #ffffff; font-weight: 500; line-height: 1.4; }

/* ═══ 动画 ═══ */
@keyframes cardIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes badgeRing {
  0% { transform: scale(1); opacity: 0.8; }
  80% { transform: scale(2.4); opacity: 0; }
  100% { transform: scale(2.4); opacity: 0; }
}
@keyframes syncPulse {
  0% { box-shadow: 0 0 0 0 rgba(22, 138, 85, 0.5); }
  70% { box-shadow: 0 0 0 12rpx rgba(22, 138, 85, 0); }
  100% { box-shadow: 0 0 0 0 rgba(22, 138, 85, 0); }
}
@keyframes priceIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes toastIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% - 20rpx)); }
  to { opacity: 1; transform: translate(-50%, -50%); }
}
@keyframes toastOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* ═══ 减少动态效果支持 ═══ */
@media (prefers-reduced-motion: reduce) {
  .card,
  .sync-dot,
  .badge--ongoing .badge-dot::after,
  .price-num,
  .pill,
  .custom-toast {
    animation: none !important;
    transition: none !important;
  }
}
</style>
