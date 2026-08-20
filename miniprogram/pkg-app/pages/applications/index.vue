<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }" @tap="closeAll">
    <u-nav-bar title="我的业务" show-back :fixed="true" @back="goBack" />

    <!-- 固定头部：搜索 + Tab 胶囊（一体吸顶） -->
    <view class="sticky-head" :style="{ top: (statusBarHeight + 44) + 'px' }" @tap.stop>

      <!-- 搜索框 -->
      <view class="sbar">
        <view class="b-search">
          <image class="b-search-ic" src="/static/home/icons/search.svg" mode="aspectFit" />
          <input class="b-sinp" v-model="q" placeholder="搜索标题、关键词" placeholder-class="b-ph" confirm-type="search" @input="onSearch" />
          <text v-if="q" class="b-sclr" aria-role="button" aria-label="清除搜索" @tap="clearSearch">×</text>
          <view class="b-sep"></view>
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- Tab 胶囊（需求/合同/订单）：沿用筛选胶囊选中态 -->
      <view class="fbar">
        <view v-for="(t, i) in TAB_OPTS" :key="t" class="fpill" :class="{ on: activeTab === i }" @tap="switchTab(i)">
          <text class="fpv">{{ t }}</text>
        </view>
      </view>
    </view>

    <!-- Banner -->
    <view class="banner">
      <view class="banner-icon">业</view>
      <view class="banner-info">
        <text class="banner-title">我的业务，全程可视</text>
        <text class="banner-sub">需求发布 · 合同履约 · 订单交易 一站式跟踪</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 + 排序 -->
      <view class="ir">
        <text>共 <text class="irn">{{ shownList.length }}</text> 项{{ tabLabel }}</text>
        <view class="irs-wrap">
          <text class="irs" @tap.stop="toggleSort">{{ sortLabel }} ▾</text>
          <view v-if="showSort" class="spop" :class="{ closing: sortClosing }" @tap.stop>
            <view v-for="s in SORTS" :key="s.v" class="sp-opt" :class="{ act: sort === s.v }" @tap="pickSort(s.v)">
              <text>{{ s.l }}</text><text v-if="sort === s.v" class="chk">✓</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 未登录引导 -->
      <view v-if="showLogin" class="st">
        <u-empty description="登录后查看我的业务">
          <text class="sth">需求、合同、订单一览</text>
          <view class="stb" @tap="goLogin">去登录</view>
        </u-empty>
      </view>

      <!-- 骨架 -->
      <view v-else-if="loading && !activeList.length" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-tag sk-w28"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w80"></view>
            <view class="sk-l w60"></view>
          </view>
        </view>
      </view>

      <!-- 错误（无旧数据） -->
      <view v-else-if="err && !shownList.length" class="st">
        <u-empty description="加载失败，请检查网络">
          <view class="stb" @tap="retry">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="!shownList.length" class="st">
        <u-empty :description="'暂无' + tabLabel">
          <text class="sth">{{ searching ? '试试调整搜索关键词' : '发布或交易后，内容自动同步到这里' }}</text>
          <view v-if="searching" class="stb" @tap="clearSearch">清除搜索</view>
        </u-empty>
      </view>

      <!-- 列表：纯文字卡片（类别/状态 tag + 标题 + 描述 + 元信息） -->
      <view v-else class="cl" :class="{ replay }">
        <view
          v-for="x in shownList"
          :key="x.key"
          class="card"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="viewDetail(x.type, x.raw)"
        >
          <view class="c-top">
            <view class="c-badges">
              <text class="c-tag" :style="{ color: x.tagC, background: x.tagBg }">{{ x.cat }}</text>
              <text class="c-st" :class="x.stCls">{{ x.stLabel }}</text>
            </view>
            <view v-if="x.valText" class="c-val"><text class="lb">{{ x.valLabel }}</text><text class="vl">{{ x.valText }}</text></view>
          </view>
          <text class="ct">{{ x.t }}</text>
          <text v-if="x.d" class="c-desc">{{ x.d }}</text>
          <view class="c-meta">
            <text>{{ x.meta1 }}</text>
            <text class="c-dot">·</text>
            <text class="c-dl">{{ x.meta2 }}</text>
          </view>
        </view>
      </view>

      <view v-if="shownList.length" class="lm">— 没有更多了 —</view>
    </view>

    <!-- 有旧数据时加载失败横幅（保留旧数据，可重试） -->
    <view v-if="!showLogin && err && activeList.length" class="err-banner">
      <text>{{ errorMsg }}</text>
      <text class="err-banner-retry" @tap="retry">重试</text>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { onLoad, onShow, onPageScroll, onUnload } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../../utils/request'
import { useReduceMotion } from '../../../utils/motion'

const SEARCH_DEBOUNCE_MS = 250 // 搜索防抖：击键停顿 250ms 后筛选，防每键整表重渲染
const SORT_CLOSE_MS = 170 // 排序弹层退场 150ms + 缓冲

const TAB_OPTS = ['需求', '合同', '订单']
const TYPE_NAMES = ['demand', 'contract', 'order']

/* 类别配色：左缘色条 + 类别 tag（与参考页领域配色同构，对比度 ≥4.5:1） */
const CAT_TAG = [
  { tagC: '#0d47a1', tagBg: '#E3EDF9' }, // 需求 · 蓝
  { tagC: '#004d40', tagBg: '#E4F2EF' }, // 合同 · 绿
  { tagC: '#B54708', tagBg: '#FDEEE4' }, // 订单 · 琥珀
]
const SORTS = [
  { v: 'latest', l: '最新发布' },
  { v: 'oldest', l: '较早发布' },
  { v: 'status', l: '状态优先' },
]
const SORT_LABEL = { latest: '最新发布', oldest: '较早发布', status: '状态优先' }

const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关

const activeTab = ref(0)
const q = ref('')
const kw = ref('') // 防抖后的搜索关键词（筛选语义用，输入框展示用 q）
const sort = ref('latest')
const showSort = ref(false)
const sortClosing = ref(false) // 排序弹层退场动画中
const replay = ref(false) // 列表轻淡入重播开关：搜索/排序后开启
const showBt = ref(false)
const statusBarHeight = ref(20)
let searchT = null // 搜索防抖定时器（onUnload 清理）
let sortT = null // 排序弹层退场定时器（onUnload 清理）

const demands = ref([])
const contracts = ref([])
const orders = ref([])

const loadingDemands = ref(false)
const loadingContracts = ref(false)
const loadingOrders = ref(false)

// P1 修复：未登录引导 + 加载失败提示
const showLogin = ref(false)
const errorMsg = ref('')

const fetchDemands = async () => {
  const user = getStoredUser()
  if (!user) {
    showLogin.value = true
    errorMsg.value = ''
    demands.value = []
    return
  }
  showLogin.value = false
  loadingDemands.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/demands', data: { mine: 1, page: 1, page_size: 100 } })
    const list = res?.data || res || []
    demands.value = Array.isArray(list) ? list : []
  } catch (e) {
    // 失败保留旧数据：空列表展示错误态+重试；有旧数据时 toast 提示不清空
    errorMsg.value = '加载失败，请稍后重试'
    if (demands.value.length > 0) {
      uni.showToast({ title: '加载失败，请下拉重试', icon: 'none' })
    }
  } finally {
    loadingDemands.value = false
  }
}

// 合同接口仅企业/协会/平台管理员可用（后端 GET /api/v1/contracts 仅 enterprise/admin 路由）：
// 个人用户直接置空态「暂无合同」，不发必然 403 的请求。
const CONTRACT_ROLES = ['enterprise', 'association_admin', 'platform_admin']

const fetchContracts = async () => {
  const user = getStoredUser()
  if (!user) {
    showLogin.value = true
    errorMsg.value = ''
    contracts.value = []
    return
  }
  const role = user.role || user.user_type || ''
  if (!CONTRACT_ROLES.includes(role)) {
    showLogin.value = false
    errorMsg.value = ''
    contracts.value = []
    return
  }
  showLogin.value = false
  loadingContracts.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/contracts' })
    const list = res?.data || res || []
    contracts.value = Array.isArray(list) ? list : []
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
    if (contracts.value.length > 0) {
      uni.showToast({ title: '加载失败，请下拉重试', icon: 'none' })
    }
  } finally {
    loadingContracts.value = false
  }
}

const fetchOrders = async () => {
  const user = getStoredUser()
  if (!user) {
    showLogin.value = true
    errorMsg.value = ''
    orders.value = []
    return
  }
  showLogin.value = false
  loadingOrders.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/trade-orders/mine' })
    const list = res?.data || res || []
    orders.value = Array.isArray(list) ? list : []
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
    if (orders.value.length > 0) {
      uni.showToast({ title: '加载失败，请下拉重试', icon: 'none' })
    }
  } finally {
    loadingOrders.value = false
  }
}

/* ===== 派生 ===== */
const activeList = computed(() => {
  if (activeTab.value === 0) return demands.value
  if (activeTab.value === 1) return contracts.value
  return orders.value
})
const loading = computed(() => {
  if (activeTab.value === 0) return loadingDemands.value
  if (activeTab.value === 1) return loadingContracts.value
  return loadingOrders.value
})
const err = computed(() => !!errorMsg.value)
const tabLabel = computed(() => TAB_OPTS[activeTab.value])
const sortLabel = computed(() => SORT_LABEL[sort.value] || '最新发布')
const searching = computed(() => !!kw.value)

/* 无重排轻淡入：仅离散操作调用（搜索防抖落定/排序切换） */
const revealList = () => {
  if (noMotion.value) return
  replay.value = false
  nextTick(() => { replay.value = true })
}

/* ===== 数据映射 ===== */
const fmtWan = (fen) => {
  const wan = Number(fen) / 100 / 10000
  return '¥' + (wan % 1 === 0 ? wan : wan.toFixed(1)) + '万'
}
const fmtYuan = (fen) => {
  const n = Number(fen) / 100
  return '¥' + n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
const idTail = (id) => {
  const s = String(id || '')
  const d = s.replace(/\D/g, '')
  return d ? d.slice(-8) : s.slice(-8) || '--'
}
const mapItem = (ti, it) => {
  const cat = CAT_TAG[ti]
  const created = String(it.created_at || it.createdAt || '')
  const dt = created.slice(0, 10)
  let t = ''
  let d = ''
  let valLabel = ''
  let valText = ''
  if (ti === 0) {
    t = it.title || it.serviceName || '需求' + it.id
    d = it.description || it.content || ''
    const fen = it.budget_fen != null ? it.budget_fen : it.budgetFen
    if (fen != null && Number(fen) > 0) { valLabel = '预算'; valText = fmtWan(fen) }
  } else if (ti === 1) {
    // 后端 Contract 无 title/contractNo/partyB 字段，编号缺失时用兜底文案
    t = it.title || it.contractNo || '合同详情'
    d = it.partyB || it.counterparty || ''
  } else {
    t = it.product_name || it.productName || it.title || '订单' + it.id
    const fen = it.amount_fen != null ? it.amount_fen : (it.amountFen != null ? it.amountFen : (it.amount != null ? it.amount : it.price))
    if (fen != null && Number(fen) > 0) { valLabel = '金额'; valText = fmtYuan(fen) }
  }
  return {
    key: TYPE_NAMES[ti] + ':' + it.id,
    type: TYPE_NAMES[ti],
    raw: it,
    cat: TAB_OPTS[ti],
    tagC: cat.tagC,
    tagBg: cat.tagBg,
    stLabel: getStatusText(it.status),
    stCls: getStatusCls(it.status),
    t,
    d,
    valLabel,
    valText,
    meta1: '发布于 ' + (dt || '时间未知'),
    meta2: '编号 ' + idTail(it.id),
    created,
  }
}

/* 状态优先排序：待处理 → 处理中 → 已完成 → 已取消/已拒绝 → 未知 */
const statusRank = (status) => {
  const t = getStatusType(status)
  if (t === 'warning') return 0
  if (t === 'primary') return 1
  if (t === 'success') return 2
  if (t === 'danger') return 3
  return 4
}

/* 过滤 + 排序（客户端，作用于当前 tab 全量数据） */
const shownList = computed(() => {
  let items = activeList.value.map((it) => mapItem(activeTab.value, it))
  if (kw.value) {
    const k = kw.value.trim().toLowerCase()
    if (k) items = items.filter((x) => (x.t + ' ' + x.d + ' ' + x.stLabel + ' ' + x.cat).toLowerCase().includes(k))
  }
  if (sort.value === 'oldest') items.sort((a, b) => String(a.created).localeCompare(String(b.created)))
  else if (sort.value === 'status') items.sort((a, b) => statusRank(a.raw.status) - statusRank(b.raw.status) || String(b.created).localeCompare(String(a.created)))
  else items.sort((a, b) => String(b.created).localeCompare(String(a.created)))
  return items
})

/* ===== 交互 ===== */
const startCloseSort = () => {
  if (sortClosing.value) return
  sortClosing.value = true
  clearTimeout(sortT)
  sortT = setTimeout(() => {
    showSort.value = false
    sortClosing.value = false
    sortT = null
  }, SORT_CLOSE_MS)
}
const toggleSort = () => {
  if (showSort.value) { startCloseSort(); return }
  showSort.value = true
}
const pickSort = (v) => { sort.value = v; startCloseSort(); revealList() }
const closeAll = () => {
  // 页面空白处点击：打开的排序弹层走退场动画收起
  if (showSort.value) startCloseSort()
}
/* 搜索防抖：击键/点搜索均走 250ms 停顿后筛选（点 × 清除不防抖，跟手优先） */
const onSearch = () => {
  clearTimeout(searchT)
  searchT = setTimeout(() => { kw.value = q.value.trim(); revealList() }, SEARCH_DEBOUNCE_MS)
}
const clearSearch = () => { clearTimeout(searchT); q.value = ''; kw.value = ''; revealList() }
const switchTab = (i) => {
  if (activeTab.value === i) return
  activeTab.value = i
  closeAll()
  // 与原 onTabChange 一致：切换即加载对应 tab 数据
  if (i === 0) fetchDemands()
  else if (i === 1) fetchContracts()
  else fetchOrders()
}
const retry = () => {
  if (activeTab.value === 0) fetchDemands()
  else if (activeTab.value === 1) fetchContracts()
  else fetchOrders()
}

const goLogin = () => {
  uni.navigateTo({ url: '/pages/login/index' })
}

const getStatusType = (status) => {
  if (!status) return 'default'
  if (status === 'pending' || status === '待处理' || status === 'draft' || status === '草稿' || status === 'aftersale' || status === '售后中') return 'warning'
  if (status === 'processing' || status === '处理中' || status === 'active' || status === '进行中' || status === 'paid' || status === '已付款' || status === 'shipped' || status === '已发货') return 'primary'
  if (status === 'sent' || status === 'signing' || status === '已发送' || status === '签署中') return 'primary'
  if (status === 'completed' || status === '已完成' || status === 'done' || status === 'published' || status === '已上架') return 'success'
  if (status === 'signed' || status === '已签署') return 'success'
  if (status === 'cancelled' || status === '已取消' || status === 'rejected' || status === '已拒绝') return 'danger'
  if (status === 'voided' || status === '已作废' || status === 'expired' || status === '已到期') return 'danger'
  return 'default'
}

const getStatusText = (status) => {
  if (!status) return '未知'
  const map = {
    pending: '待处理',
    processing: '处理中',
    completed: '已完成',
    done: '已完成',
    cancelled: '已取消',
    rejected: '已拒绝',
    active: '进行中',
    draft: '草稿',
    published: '已上架',
    paid: '已付款',
    shipped: '已发货',
    aftersale: '售后中',
    // 合同状态（后端 ContractStatus）
    sent: '已发送',
    signing: '签署中',
    signed: '已签署',
    voided: '已作废',
    expired: '已到期',
  }
  return map[status] || status
}

/* 状态 tag 视觉类：正常绿或蓝 / 错误红 / 未知灰 */
const getStatusCls = (status) => {
  const t = getStatusType(status)
  if (t === 'warning') return 'st-pending'
  if (t === 'primary') return 'st-proc'
  if (t === 'success') return 'st-done'
  if (t === 'danger') return 'st-bad'
  return 'st-mute'
}

const viewDetail = (type, item) => {
  const titleMap = { demand: '需求详情', contract: '合同详情', order: '订单详情' }
  const lines = []
  if (item.title || item.serviceName) lines.push('标题：' + (item.title || item.serviceName))
  if (item.description || item.content) lines.push('描述：' + (item.description || item.content))
  if (item.partyB || item.counterparty) lines.push('对方：' + (item.partyB || item.counterparty))
  if (item.amount != null || item.price != null) lines.push('金额：' + (item.amount || item.price))
  if (item.status) lines.push('状态：' + getStatusText(item.status))
  if (item.created_at) lines.push('时间：' + String(item.created_at).slice(0, 10))
  uni.showModal({
    title: titleMap[type] || '详情',
    content: lines.length ? lines.join('\n') : '暂无更多信息',
    showCancel: false,
    confirmText: '知道了',
  })
}

const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })
const goBack = () => {
  uni.navigateBack()
}

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
})

onShow(() => {
  // Load active tab data on show
  if (activeTab.value === 0) fetchDemands()
  else if (activeTab.value === 1) fetchContracts()
  else if (activeTab.value === 2) fetchOrders()
})

onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})

onUnload(() => {
  // 页面卸载清除所有定时器，防回调泄漏
  clearTimeout(searchT)
  clearTimeout(sortT)
})
</script>

<style>
page {
  background: #fff;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 40px;
}

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 极淡灰投影，从白底上"浮"起 ===== */
.sbar {
  padding: 12px 12px 8px;
  background: #fff;
}
.b-search {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05); /* 双层投影：接触阴影贴地 + 环境阴影弥散浮起 */
  display: flex;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.b-search-ic { width: 15px; height: 15px; flex: none; }
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; } /* 热区扩大：视觉 × 外扩 10px，点击不脱靶 */
/* 小红书风格搜索按钮：无底色文字 + 左侧细竖杠分隔 */
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== 固定头部 ===== */
.sticky-head {
  position: sticky;
  z-index: 40;
  background: #fff;
}

/* ===== Tab 胶囊条（需求/合同/订单，沿用筛选胶囊样式） ===== */
.fbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px 12px;
  background: #fff;
}
.fpill {
  flex: 1;
  min-width: 0;
  min-height: 40px; /* 触控目标：34px→40px（Tab 切换高频操作） */
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease, opacity .2s ease;
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpill:active { opacity: .75; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ===== Banner（参考页 projects 风格） ===== */
.banner {
  margin: 12px 14px;
  padding: 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #0A66C2 0%, #074D92 100%);
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
  position: relative;
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}
.banner::after {
  content: '';
  position: absolute;
  top: -30%;
  right: -20%;
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.08) 0%, transparent 70%);
}
.banner-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}
.banner-info { flex: 1; min-width: 0; position: relative; z-index: 1; }
.banner-title { font-size: 14px; font-weight: 600; margin-bottom: 4px; display: block; line-height: 1.3; color: #fff; }
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: block; } /* 白 95%：蓝底上 ≥4.5:1 达标 */

/* ===== 白色板块（信息行 + 列表）：与页面同底，融入不分块 ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 信息行：白底浮条（共 N 项 + 排序），从白底上"浮"起 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 4px 12px 8px;
  padding: 10px 12px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 8px 28px rgba(16, 24, 40, 0.12);
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }
.irs-wrap { position: relative; }
.irs { color: #0A66C2; font-weight: 500; padding: 8px 4px 8px 12px; } /* 热区扩大：6px→8px 纵向 */
.spop {
  position: absolute;
  top: 32px;
  right: 0;
  z-index: 90;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 8px 28px rgba(16, 24, 40, 0.12);
  padding: 6px;
  min-width: 140px;
  animation: spopIn .22s cubic-bezier(.32, .72, 0, 1); /* ios-decel：下拉流体减速，越到终点越柔 */
}
.spop.closing {
  animation: spopOut .15s ease-in forwards;
}
@keyframes spopIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes spopOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-4px); }
}
.sp-opt {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  color: #17212B;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sp-opt.act { color: #0A66C2; font-weight: 600; background: #EAF3FB; }
.chk { color: #0A66C2; font-size: 12px; }

/* ===== 列表项：纯文字卡片（白上白：灰描边 + 极淡灰投影浮起，窄缝 8px；类别/状态 tag 为视觉锚点） ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
  display: flex;
  flex-direction: column;
  gap: 7px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC; /* 描边提级：低端设备投影失效时仍与白底可分辨 */
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06); /* 卡片浮层：大偏移 + 宽模糊 + 低透明，柔和环境阴影 */
}
.tap-scale { transform: scale(0.95); opacity: 0.9; }
.c-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.c-badges { display: flex; gap: 6px; }
.c-tag, .c-st {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.c-tag { color: #074D92; background: #EAF3FB; } /* 兜底色；实际按类别由 mapItem 传入 */
.c-st.st-pending { color: #0A66C2; background: #EAF3FB; } /* 待处理 · 蓝 */
.c-st.st-proc { color: #0B6B41; background: #E9F7F0; } /* 处理中/进行中 · 绿 */
.c-st.st-done { color: #0B6B41; background: #E9F7F0; } /* 已完成 · 绿 */
.c-st.st-bad { color: #B42318; background: #FDECEC; } /* 已取消/已拒绝 · 红 */
.c-st.st-mute { color: #5D6B82; background: #EEF1F4; } /* 未知 · 灰 */
.ct {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-desc {
  font-size: 12.5px;
  color: #667085;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #667085;
}
.c-dot { color: #DDE1E6; }
.c-dl { color: #667085; font-weight: 500; }
.c-val { display: flex; align-items: baseline; gap: 3px; color: #C2410C; } /* 预算/金额：暖色强调，与参考页悬赏同构 */
.c-val .lb { font-size: 12px; font-weight: 500; }
.c-val .vl { font-size: 18px; font-weight: 800; }

/* ===== 骨架 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-tag.sk-w28 { width: 28px; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 遮罩 / 加载更多 / 返回顶部 / 错误横幅 ===== */
.lm { text-align: center; padding: 12px; font-size: 12px; color: #667085; }
.err-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin: 12px 14px 0;
  padding: 10px 16px;
  border-radius: 10px;
  background: #FDECEC;
  font-size: 13px;
  color: #B42318;
}
.err-banner-retry { color: #0A66C2; font-weight: 600; }
.bt {
  position: fixed;
  bottom: 90px;
  right: 16px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 35; /* 低于 sticky-head(40)，高于页面内容 */
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(0.16, 1, 0.3, 1); /* ios-pop：出现/隐藏弹簧收尾，返回顶部"弹"出来 */
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; } /* 按压即时到位 */

/* ===================== 动效规范（对齐参考页全局动画规范） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许——仅重绘不重排）
   禁参与动画：top/left/width/height/margin、box-shadow/filter（低端安卓掉帧）
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1)（按压/弹出类 transform）+
        ios-decel cubic-bezier(.32,.72,0,1)（浮层流体减速）；其余进场 ease-out / 退场 ease-in
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 0) 列表入场：前 6 项每 20ms 依次淡入上移（首屏可见范围；80ms 起 + 100ms 错峰 + 220ms 动画 = 400ms ≤ 400ms）
   backwards 填充 → 延迟期保持隐藏不闪跳 */
.card { animation: none; }
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
/* 列表刷新（搜索/排序后）：前 4 项轻淡入——2px 上移 + 半透明起，180ms + 30ms 错峰 ≤270ms；
   说明"这组内容因操作而更新"，无大位移不抢戏；nth-child 优先级高于入场错峰（同源覆盖） */
.cl.replay .card:nth-child(-n+4) { animation: listFade .18s ease-out backwards; }
.cl.replay .card:nth-child(1) { animation-delay: 0ms; }
.cl.replay .card:nth-child(2) { animation-delay: 30ms; }
.cl.replay .card:nth-child(3) { animation-delay: 60ms; }
.cl.replay .card:nth-child(4) { animation-delay: 90ms; }
@keyframes listFade { from { opacity: .3; transform: translateY(2px); } to { opacity: 1; transform: translateY(0); } }
/* 卡片按压（快进慢出）：hover-start-time=0 按下立即触发；按下 .1s linear 直接到位，松手 .35s ios-pop 弹簧回位 */
.card { transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; } /* ios-pop */
.card.tap-scale { transition-duration: .1s; transition-timing-function: linear; }

/* Banner 内部微编排：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms，总 340ms ≤ 400ms；
   全部单次动画非循环；banner 单次扫光（::before 默认态 translateX(-150%) 隐藏，动画结束复位） */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; }
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }
.banner::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  background: linear-gradient(100deg, transparent 0%, rgba(255, 255, 255, 0.22) 50%, transparent 100%);
  transform: translateX(-150%) skewX(-20deg);
  animation: shineOnce .28s linear 100ms backwards;
  pointer-events: none;
}
@keyframes shineOnce {
  from { transform: translateX(-150%) skewX(-20deg); }
  to { transform: translateX(320%) skewX(-20deg); }
}
/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；循环动画 1.4s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 1) 交互反馈：可点元素按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位；opacity/background 150-200ms） */
.irs { transition: opacity .2s ease, transform .3s cubic-bezier(0.16, 1, 0.3, 1); } /* ios-pop */
.irs:active { opacity: .7; transform: scale(.95); transition: transform .08s linear; }
.sp-opt { transition: background .2s ease, color .2s ease; }
.sp-opt:active { background: #F4F8FC; }
.b-sclr:active { opacity: .6; }
.b-sbtn { transition: opacity .2s ease; }
.b-sbtn:active { opacity: .5; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; } /* ios-pop：松手弹簧回位 */
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }
.err-banner-retry { transition: opacity .2s ease; }
.err-banner-retry:active { opacity: .6; }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .card,
.page.no-motion .banner,
.page.no-motion .ir { animation: none; } /* 装饰入场全关 */
.page.no-motion .cl.replay .card { animation: none; } /* 列表刷新轻淡入关闭（覆盖高优先级重播规则） */
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner::before,
.page.no-motion .banner::after { animation: none; } /* banner 内部微编排/扫光/装饰圆全关 */
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; } /* 循环呼吸关 */
.page.no-motion .spop { animation: spopFadeIn .2s ease-out; }
.page.no-motion .spop.closing { animation: spopFadeOut .15s ease-in forwards; }
.page.no-motion .tap-scale { transform: none !important; } /* 按压缩放关闭，保留 opacity 反馈 */
.page.no-motion .irs:active,
.page.no-motion .stb:active,
.page.no-motion .bt:active { transform: none; } /* 按压微缩放关闭，保留颜色/透明度反馈 */
@keyframes spopFadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes spopFadeOut { from { opacity: 1; } to { opacity: 0; } }
</style>
