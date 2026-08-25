<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- Skeleton -->
    <view v-if="loading" class="skw">
      <!-- 骨架镜像刊头终态结构：bar + tag + 双行标题 + 副行（替换 16/9 英雄图暗示） -->
      <view class="sk-h" :style="{ paddingTop: sbh + 'px' }">
        <view class="sk-h-bar"></view>
        <view class="sk-h-line sk-h-tag"></view>
        <view class="sk-h-line sk-h-title"></view>
        <view class="sk-h-line sk-h-title"></view>
        <view class="sk-h-line sk-h-sub"></view>
      </view>
      <view class="sk-cover"></view>
      <view class="sk-stats"><view class="sk-stat"></view><view class="sk-stat"></view><view class="sk-stat"></view></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- Error -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- Empty -->
    <view v-else-if="!d" class="st">
      <u-empty description="该成果已下架或不存在">
        <view class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <!-- Content -->
    <template v-else>
      <!-- 刊头（navigationStyle: custom）：全标题 + 领域/类型 tag + 副行 -->
      <view class="mh" :style="{ paddingTop: sbh + 'px' }">
        <view class="mh-bar">
          <view class="mh-back" aria-role="button" aria-label="返回上一页" @tap="goBack"></view>
          <text class="mh-title">成果详情</text>
          <view class="mh-side"></view>
        </view>
        <view class="mh-hero">
          <view class="mh-tags">
            <text v-if="stageLabel" class="mh-tag mh-stage" :class="stageTagCls">{{ stageLabel }}</text>
            <text v-if="d.field" class="mh-tag">{{ d.field }}</text>
            <text v-if="typeLabel" class="mh-tag">{{ typeLabel }}</text>
          </view>
          <text class="mh-big">{{ d.title }}</text>
          <text class="mh-sub">{{ '发布于 ' + d.date }}</text>
        </view>
      </view>

      <!-- 图位（列表页第一视位层级：领域色底 → 图片终态；坏图 @error 回退单字+刊名；已转化角标同列表 conv-badge；
           无图时压缩为小图位（此前 4/3 大片占位色块观感突兀） -->
      <view class="cov" :class="{ 'cov--noimg': !d.img || imgFailed }" :style="{ background: d.tone.bg }">
        <image v-if="d.img && !imgFailed" class="cov-img" :src="d.img" mode="aspectFill" @error="imgFailed = true" />
        <block v-else>
          <text class="cov-ic" :style="{ color: d.tone.fg }">{{ d.ic }}</text>
          <text class="cov-name" :style="{ color: d.tone.fg }">{{ d.f }}</text>
        </block>
        <text v-if="statusLabel === '已转化'" class="conv-badge">✓ 已转化</text>
      </view>

      <!-- Stats Row（3 格：仅保留刊头/基本信息未重复的事实——来源/浏览/收藏；发布日期与类型已在刊头与基本信息各有一处） -->
      <view class="sec stats" :class="{ vis: vis }">
        <view class="si"><text class="sv sv-e">{{ d.poster_name || '—' }}</text><text class="sl">来源</text></view>
        <view class="si"><text class="sv">{{ fmtNum(d.views) }}</text><text class="sl">浏览</text></view>
        <view class="si"><text class="sv">{{ fmtNum(d.favs + (isFav ? 1 : 0)) }}</text><text class="sl">收藏</text></view>
      </view>

      <!-- Description -->
      <view class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">成果描述</text></view>
        <text class="sb">{{ d.description || '暂无描述' }}</text>
      </view>

      <!-- Basic Info -->
      <view class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">基本信息</text></view>
        <view class="it">
          <view v-if="d.poster_name" class="ir2"><text class="ik">来源</text><text class="iv">{{ d.poster_name }}</text></view>
          <view v-if="statusLabel" class="ir2"><text class="ik">成果状态</text><text class="iv">{{ statusLabel }}</text></view>
          <view v-if="d.field" class="ir2"><text class="ik">所属领域</text><text class="iv">{{ d.field }}</text></view>
          <view v-if="d.achieve_type" class="ir2"><text class="ik">成果类型</text><text class="iv">{{ typeLabel }}</text></view>
          <view v-if="d.contact_info" class="ir2"><text class="ik">联系方式</text><text class="iv">{{ maskContact(d.contact_info) }}</text></view>
          <view v-if="d.created_at" class="ir2"><text class="ik">创建时间</text><text class="iv">{{ d.date }}</text></view>
          <view v-if="d.updated_at && d.updated_at !== d.created_at" class="ir2"><text class="ik">更新时间</text><text class="iv">{{ (d.updated_at || '').slice(0, 10) }}</text></view>
        </view>
      </view>

      <!-- 成果阶段（常驻轨道：以 d.stage 驱动，四段跑道永不缺席；转化记录只做深链钻取，不在页内复述里程碑） -->
      <view v-if="showTrack" class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">成果阶段</text></view>
        <view class="tr-flow">
          <view class="tr-track">
            <view class="tr-base"></view>
            <view class="tr-prog" :style="{ transform: flowReady ? 'scaleX(' + flowPctOf(d.stage) + ')' : 'scaleX(0)' }"></view>
            <view class="tr-stages">
              <view
                v-for="(st, si) in stages"
                :key="si"
                class="tr-stage"
                :class="{ done: stageRank >= si + 1, cur: stageRank === si + 1 }"
              >
                <view class="tr-dot"></view>
                <text class="tr-stage-name">{{ st }}</text>
              </view>
            </view>
          </view>
          <view class="tr-meta">
            <text>{{ stageRank ? '第 ' + stageRank + '/4 阶段' : '成果阶段待确认' }}</text>
          </view>
        </view>
        <!-- 转化记录：加载失败原地重试；有记录做深链；已转化但记录未整理 = 诚实占位（rail 不因记录缺失而消失） -->
        <view v-if="transErr" class="tr-err" @tap="fetchTransformations(id)">转化进展加载失败，点此重试</view>
        <template v-else>
          <view v-if="transforms.length" class="tr-go" @tap="goTrack"><text>查看转化进展 ›</text></view>
          <view v-else-if="statusLabel === '已转化'" class="tr-ongoing">转化进展整理中</view>
        </template>
      </view>

      <!-- 附件资料 -->
      <view v-if="d.attachments && d.attachments.length" class="sec" :class="{ vis: vis }">
        <view class="sh"><view class="sd"></view><text class="sht">附件资料</text></view>
        <view v-for="(at, i) in d.attachments" :key="i" class="at-row" :class="{ dl: dl.i === i }" @tap="downloadAt(at, i)">
          <view class="at-ic"><text>附</text></view>
          <view class="at-info">
            <text class="at-name">{{ at.name || '附件' }}</text>
            <text class="at-size">{{ dl.i === i ? (dl.pct >= 0 ? '下载中 ' + dl.pct + '%' : '下载中…') : (at.size || '') }}</text>
          </view>
          <text v-if="dl.i === i" class="at-btn at-cancel" @tap.stop="cancelDownload">取消</text>
          <text v-else class="at-btn" :class="{ dis: dl.i >= 0 }">下载</text>
        </view>
      </view>

      <view style="height: 160rpx"></view>

      <!-- 底部操作栏（常驻不随滚动撤走；成果转化入口只在转化区块内一处） -->
      <view class="bb">
        <view class="bi" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" :class="{ fv: isFav }" @tap="toggleFav"><text class="bit"></text></view>
        <button class="bo" open-type="share" hover-class="bo-hover" hover-start-time="0" hover-stay-time="300" aria-label="转发">转发</button>
        <view class="bp" @tap="onContact">联系对接</view>
      </view>
      <view class="fp" :class="favPop ? (favHide ? 'hide' : '') : 'hide'"></view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onUnload, onShareAppMessage } from '@dcloudio/uni-app'
import { request, BASE_URL } from '@/utils/request'
import { maskContact, revealContactCopy } from '@/utils/contactMask'
import { useReduceMotion } from '@/utils/motion'
import { MOCK_ACHIEVEMENTS, MOCK_TRANSFORMS_BY_ACH, ACH_TYPE_LABEL, STAGE_SHORT, STAGE_RANK, FIELD_TONE, TONE_DEFAULT, FIELD_ICON } from '@/utils/mockAchievements'

const goBack = () => uni.navigateBack()

const id = ref('')
const d = ref(null)
const imgFailed = ref(false) // 图位坏图 → 回退单字+刊名版式（同列表页 @error 契约）
const loading = ref(true)
const err = ref(false)
const vis = ref(false)
const isFav = ref(false)
const favPop = ref(false)
const favHide = ref(false)
const sbh = ref(20) // 自定义导航：状态栏高度（真机按系统信息覆盖）

const statusLabel = computed(() => {
  const s = (d.value?.status || '').toLowerCase()
  if (s === 'hot') return '热门'
  if (s === 'transformed') return '已转化'
  return ''
})
const fmtNum = (n) => (n >= 1e4 ? (n / 1e4).toFixed(1) + 'w' : n >= 1e3 ? (n / 1e3).toFixed(1) + 'k' : String(n || 0))
const stageLabel = computed(() => {
  const s = (d.value?.stage || '').toLowerCase()
  return STAGE_SHORT[s] || d.value?.stage || ''
})
// 刊头阶段徽章四段签名色（同列表页 badge 变体：实验室灰 / 中试琥珀 / 产业化绿 / 已上市深蓝）
const STAGE_TAG_CLS = { lab: 'cl-la', laboratory: 'cl-la', pilot: 'cl-pi', industrialization: 'cl-in', industrialized: 'cl-in', listed: 'cl-li' }
const stageTagCls = computed(() => STAGE_TAG_CLS[(d.value?.stage || '').toLowerCase()] || '')
const transforms = ref([])
const stages = ['实验室', '中试', '产业化', '已上市'] // 阶段字面同系统：已上市
// 常驻轨道：阶段 = 成果的成熟度宣言，以 d.stage 驱动（转化记录只做深链钻取）
const stageRank = computed(() => {
  const r = STAGE_RANK[(d.value?.stage || '').toLowerCase()] || 0
  return r
})
// 常驻条件：有阶段信息或任何转化信号即渲染；完全没有时省略空轨（避免全灰跑道 + 待确认的噪声）
const showTrack = computed(() => !!(d.value?.stage || transforms.value.length || transErr.value || statusLabel.value === '已转化'))
const flowReady = ref(false)
const flowPctOf = (s) => {
  const r = STAGE_RANK[(s || '').toLowerCase()] || 0
  return r > 0 ? (r - 1) / 3 : 0
}
const applyTransforms = (list) => {
  transforms.value = list
  flowReady.value = false
  setTimeout(() => { flowReady.value = true }, 300)
}
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
const transErr = ref(false)
const fetchTransformations = async (aid) => {
  transErr.value = false
  try {
    const res = await withTimeout(request({ url: '/api/v1/transformations', data: { achievement_id: aid } }))
    applyTransforms(Array.isArray(res) ? res : (res?.data || []))
  } catch (e) { transErr.value = true; applyTransforms([]) }
}
const typeLabel = computed(() => {
  const t = (d.value?.achieve_type || '').toLowerCase()
  return ACH_TYPE_LABEL[t] || d.value?.achieve_type || '未分类'
})

const imgSrc = (images) => {
  let arr = images
  if (typeof images === 'string') {
    try { arr = JSON.parse(images) } catch { return '' }
  }
  if (!Array.isArray(arr) || !arr.length) return ''
  const u = arr[0]
  return u ? (u.startsWith('http') ? u : BASE_URL + u) : ''
}

// ===== 数据获取 =====
// 请求超时兜底：挂起/断网时 10s 内转为错误态，避免永久骨架（生产环境不回落演示数据）
const withTimeout = (p, ms = 10000) => new Promise((resolve, reject) => {
  const timer = setTimeout(() => reject(new Error('请求超时，请重试')), ms)
  p.then(
    (v) => { clearTimeout(timer); resolve(v) },
    (e) => { clearTimeout(timer); reject(e) }
  )
})

// 接口替换点：GET /api/v1/achievements/{id} + GET /api/v1/transformations?achievement_id=
const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    const res = await withTimeout(request({ url: '/api/v1/achievements/' + encodeURIComponent(id.value) }))
    const item = res?.data || res
    if (item) {
      applyDetail(item)
      fetchTransformations(item.id)
    } else {
      err.value = true
    }
  } catch {
    useMock()
  } finally {
    loading.value = false
  }
  setTimeout(() => { vis.value = true }, 200)
}

const applyDetail = (item) => {
  d.value = {
    id: item.id,
    poster_name: item.poster_name || '', // 来源单位（与列表页同一契约：无则省略，不编造）
    title: item.title || '',
    achieve_type: item.achieve_type || '',
    description: item.description || '',
    field: item.field || '',
    tone: FIELD_TONE[item.field] || TONE_DEFAULT, // 图位领域色（与列表页同源，见 utils/mockAchievements.js）
    ic: FIELD_ICON[item.field] || (item.field ? item.field.charAt(0) : '果'),
    f: item.field || '其他',
    stage: item.stage || '',
    images: item.images || [],
    img: imgSrc(item.images),
    attachments: item.attachments || [],
    contact_info: item.contact_info || '',
    status: item.status || '',
    created_at: item.created_at || '',
    updated_at: item.updated_at || '',
    date: (item.created_at || '').slice(0, 10),
    views: item.views || 0,
    favs: item.favs || 0
  }
}

// 演示数据回退（仅 demo- 前缀 id）
const useMock = () => {
  // P2-6：生产环境绝不回退演示数据，失败即诚实错误态
  if (process.env.NODE_ENV === 'production') { err.value = true; return }
  const mock = MOCK_ACHIEVEMENTS.find((x) => x.id === id.value)
  if (mock) {
    applyDetail(mock)
    applyTransforms(MOCK_TRANSFORMS_BY_ACH[id.value] || [])
  } else {
    err.value = true
  }
}

// ===== 收藏统一：本地单键增量过渡（与列表页共享 fav_ach_set） =====
// 接口替换点：POST/DELETE /api/v1/favorites/{achievement_id} + GET /api/v1/favorites/mine
// 落地后切换为后端计数（《科技成果库-后端改动清单》§1），删除本段
const favSet = ref(new Set())
const loadFavs = () => {
  try {
    const raw = uni.getStorageSync('fav_ach_set')
    favSet.value = new Set(Array.isArray(raw) ? raw : [])
  } catch (e) { favSet.value = new Set() }
}
const saveFavs = () => {
  uni.setStorageSync('fav_ach_set', Array.from(favSet.value))
}
const toggleFav = () => {
  if (favSet.value.has(id.value)) favSet.value.delete(id.value)
  else favSet.value.add(id.value)
  saveFavs()
  isFav.value = favSet.value.has(id.value)
  if (isFav.value) {
    favPop.value = true
    favHide.value = false
    uni.showToast({ title: '已收藏', icon: 'success', duration: 1200 })
    setTimeout(() => { favHide.value = true }, 600)
    setTimeout(() => { favPop.value = false; favHide.value = false }, 1000)
  } else {
    uni.showToast({ title: '已取消收藏', icon: 'none', duration: 1200 })
  }
}
const onContact = () => {
  // 信任门槛：展示已脱敏，完整值经确认后复制（见 utils/contactMask.js）
  if (!revealContactCopy(d.value?.contact_info)) {
    uni.showToast({ title: '暂未公开联系方式', icon: 'none', duration: 1500 })
  }
}

// 转发：分享卡片（同研发难题 onShareAppMessage）
onShareAppMessage(() => ({
  title: d.value ? '科技成果：' + d.value.title : '低空经济生态服务平台 · 科技成果库',
  path: '/pkg-eco/pages/achievements/detail?id=' + encodeURIComponent(id.value),
}))

// 进入成果转化页（详情 → 转化；入口仅在转化区块有数据时渲染）
const goTrack = () => {
  const t = transforms.value[0]
  if (!t) return
  uni.navigateTo({ url: '/pkg-eco/pages/transformations/track?achievement_id=' + encodeURIComponent(id.value) + '&id=' + encodeURIComponent(t.id || '') })
}

// ===== 附件下载（反馈闭环）：行级状态 + 进度百分比 + 移动网络确认 + 失败真实落点 =====
const dl = ref({ i: -1, pct: -1 }) // i: 下载中的行；pct: -1 = 进度不可得（只显示「下载中…」）
const confirming = ref(false) // 确认弹窗防重入
let dlTask = null

const parseBytes = (size) => {
  if (typeof size === 'number') return size
  const m = String(size || '').match(/([\d.]+)\s*(B|KB|MB|GB|TB)/i)
  if (!m) return 0
  const n = parseFloat(m[1])
  const unit = { b: 1, kb: 1024, mb: 1024 ** 2, gb: 1024 ** 3, tb: 1024 ** 4 }[m[2].toLowerCase()] || 1
  return Math.round(n * unit)
}
const fmtBytes = (n) => {
  if (!n) return ''
  if (n >= 1024 ** 3) return (n / 1024 ** 3).toFixed(1) + ' GB'
  if (n >= 1024 ** 2) return (n / 1024 ** 2).toFixed(1) + ' MB'
  if (n >= 1024) return (n / 1024).toFixed(0) + ' KB'
  return n + ' B'
}

const downloadAt = (at, i) => {
  if (!at?.url) { uni.showToast({ title: '附件地址缺失', icon: 'none' }); return }
  if (dl.value.i >= 0 || confirming.value) { uni.showToast({ title: '请等待当前下载完成', icon: 'none' }); return }
  const url = at.url.startsWith('http') ? at.url : BASE_URL + at.url
  const bytes = parseBytes(at.size)
  uni.getNetworkType({
    success: (nt) => {
      const net = (nt.networkType || '').toLowerCase()
      if (net === 'none') { uni.showToast({ title: '当前无网络，请联网后重试', icon: 'none' }); return }
      // 移动网络 + 大附件：先确认流量，不静默走蜂窝
      if (net !== 'wifi' && bytes > 10 * 1024 * 1024) {
        confirming.value = true
        uni.showModal({
          title: '继续下载？',
          content: '当前为移动网络，附件约 ' + fmtBytes(bytes) + '，可能消耗较多流量。',
          cancelText: '取消',
          confirmText: '继续下载',
          success: (r) => {
            confirming.value = false
            if (r.confirm) startDownload(at, url, i)
          },
          fail: () => { confirming.value = false },
        })
        return
      }
      startDownload(at, url, i)
    },
    fail: () => startDownload(at, url, i), // 拿不到网络状态时不拦截
  })
}

const startDownload = (at, url, i) => {
  dl.value = { i, pct: -1 }
  dlTask = uni.downloadFile({
    url,
    success: (res) => {
      dl.value = { i: -1, pct: -1 }
      if (res.statusCode === 200) {
        uni.openDocument({
          filePath: res.tempFilePath,
          showMenu: true,
          fail: () => {
            // 预览不了 ≠ 下载失败：给复制链接的真实落点，不再丢一句「请用浏览器打开」
            uni.showModal({
              title: '无法直接预览',
              content: '该格式暂不支持小程序内预览，可复制链接后用浏览器或电脑打开。',
              cancelText: '取消',
              confirmText: '复制链接',
              success: (r) => {
                if (!r.confirm) return
                uni.setClipboardData({ data: url, success: () => uni.showToast({ title: '链接已复制', icon: 'success' }) })
              },
            })
          },
        })
      } else {
        uni.showToast({ title: '下载失败（' + res.statusCode + '）', icon: 'none' })
      }
    },
    fail: (e) => {
      dl.value = { i: -1, pct: -1 }
      // 用户主动取消（errMsg 含 abort）不弹错误
      if (String(e?.errMsg || '').indexOf('abort') >= 0) return
      uni.showToast({ title: '下载失败，请重试', icon: 'none' })
    },
  })
  if (dlTask && typeof dlTask.onProgressUpdate === 'function') {
    dlTask.onProgressUpdate((res) => {
      if (res.progress > 0) dl.value = { i, pct: res.progress }
    })
  }
}

const cancelDownload = () => {
  if (dlTask && typeof dlTask.abort === 'function') dlTask.abort()
  dl.value = { i: -1, pct: -1 }
  dlTask = null
}

// 离开页面时中断在途下载（状态/内存卫生）
onUnload(() => {
  if (dlTask && typeof dlTask.abort === 'function') dlTask.abort()
  dlTask = null
})

onLoad((options) => {
  if (options?.id) id.value = decodeURIComponent(options.id)
  try { sbh.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 保持默认 20 */ }
  checkMotion()
  loadFavs()
  isFav.value = favSet.value.has(id.value)
  fetchData()
})
</script>

<style>
page { background: var(--color-bg); }
</style>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ===== 刊头（navigationStyle: custom）：深蓝渐变 + 全标题 + 领域/类型 tag ===== */
.mh { background: linear-gradient(160deg, #0a3a6b, #074d92); }
.mh-bar { position: relative; display: flex; align-items: center; justify-content: center; height: 88rpx; }
.mh-back { position: absolute; left: 8rpx; top: 0; width: 88rpx; height: 88rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23ffffff' stroke-width='2.2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M15 4l-8 8 8 8'/%3E%3C/svg%3E"); background-size: 44rpx; background-repeat: no-repeat; background-position: center; transition: opacity .2s ease; }
.mh-back:active { opacity: .6; }
.mh-title { font-size: 32rpx; font-weight: 600; color: #fff; } /* 对齐 u-nav-bar 标题（系统字体 32rpx/600） */
.mh-side { position: absolute; right: 0; width: 88rpx; height: 88rpx; }
.mh-hero { padding: 20rpx 40rpx 44rpx; }
.mh-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 20rpx; }
.mh-tag { font-size: 20rpx; font-weight: 700; padding: 6rpx 16rpx; border-radius: 8rpx; background: rgba(255,255,255,.16); color: #fff; }
/* 阶段徽章（刊头第一视觉位，四段签名色）：浅底深字同列表页 badge 变体 */
.mh-stage.cl-la { background: #EEF1F4; color: #5D6B82; }
.mh-stage.cl-pi { background: #FFF4E5; color: #B45309; }
.mh-stage.cl-in { background: #E9F7F0; color: #1F7A48; }
.mh-stage.cl-li { background: #EAF3FB; color: #074D92; }
.mh-big { display: block; font-family: Georgia, "Songti SC", "STSong", SimSun, serif; font-size: 40rpx; font-weight: 700; letter-spacing: 4rpx; line-height: 1.5; color: #fff; } /* 白字在深蓝渐变上 ≥7:1，不加投影（深度靠色阶，见 DESIGN.md 轻影规则） */
.mh-sub { display: block; margin-top: 16rpx; font-size: 24rpx; color: rgba(255,255,255,.78); }

/* ===== 图位（列表页第一视位层级：领域色底 → 图片终态；坏图回退单字+刊名；对齐组卡面） ===== */
.cov { position: relative; margin: 0 32rpx 24rpx; aspect-ratio: 4/3; min-height: 240rpx; display: flex; align-items: center; justify-content: center; border: 2rpx solid #E4E7EC; border-radius: 20rpx; overflow: hidden; box-shadow: 0 2rpx 6rpx rgba(10,30,60,.04), 0 12rpx 32rpx rgba(10,30,60,.05); }
/* 无图/坏图回退：压缩为紧凑图位（领域色 + 单字 + 刊名纵向排布），不做 4/3 大片占位 */
.cov--noimg { aspect-ratio: auto; min-height: 0; padding: 36rpx 0 30rpx; flex-direction: column; gap: 12rpx; }
.cov--noimg .cov-ic { font-size: 44rpx; }
.cov--noimg .cov-name { position: static; font-family: Georgia, "Songti SC", "STSong", SimSun, serif; font-size: 20rpx; letter-spacing: 2rpx; }
.cov-img { position: absolute; left: 0; top: 0; width: 100%; height: 100%; display: block; } /* 图片脱离布局：异步加载完成不改变图位高度（同列表页防晃动） */
.cov-ic { font-size: 40rpx; font-weight: 800; } /* 图位单字：对齐列表页 cov-ic 40rpx 先例 */
.cov-name { position: absolute; left: 24rpx; bottom: 20rpx; font-family: Georgia, "Songti SC", "STSong", SimSun, serif; font-size: 20rpx; letter-spacing: 2rpx; } /* 领域刊名（衬线微刻字，DESIGN.md 封面例外） */
.conv-badge { position: absolute; right: 16rpx; top: 16rpx; font-size: 24rpx; font-weight: 700; padding: 4rpx 14rpx; border-radius: 8rpx; background: #E9F7F0; color: #0B6B41; } /* 已转化角标 12px：同列表页徽章基准 */

/* ===== Sections（保留 .vis 上浮动画；对齐组卡片：描边 + 20rpx + 蓝调双层） ===== */
.sec { margin: 0 32rpx 24rpx; padding: 32rpx; background: #fff; border: 2rpx solid #E4E7EC; border-radius: 20rpx; box-shadow: 0 2rpx 6rpx rgba(10,30,60,.04), 0 12rpx 32rpx rgba(10,30,60,.05); opacity: 0; transform: translateY(20rpx); transition: opacity .3s cubic-bezier(0.16, 1, 0.3, 1), transform .3s cubic-bezier(0.16, 1, 0.3, 1); }
.sec.vis { opacity: 1; transform: translateY(0); }
.sh { display: flex; align-items: center; gap: 16rpx; margin-bottom: 24rpx; }
.sd { width: 8rpx; height: 36rpx; background: var(--color-primary); border-radius: 2rpx; flex-shrink: 0; }
.sht { font-size: 32rpx; font-weight: 700; color: var(--color-text); }
.sb { font-size: 30rpx; color: var(--color-text-secondary); line-height: 1.75; white-space: pre-wrap; } /* 正文 15px：同研发难题详情 .p */

/* ===== Stats（保留） ===== */
.stats { display: flex; padding: 32rpx 0; }
.si { flex: 1; text-align: center; position: relative; }
.si + .si::before { content: ''; position: absolute; left: 0; top: 16rpx; bottom: 16rpx; width: 1rpx; background: var(--color-divider); }
.sv { font-size: 32rpx; font-weight: 700; color: var(--color-text); display: block; }
.sv-e { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; } /* 来源单位长名截断（stats 行重组后保留） */
.sl { font-size: 24rpx; color: #667085; margin-top: 4rpx; display: block; }

/* ===== Info Table（保留） ===== */
.it { display: flex; flex-direction: column; }
.ir2 { display: flex; padding: 24rpx 0; border-bottom: 1rpx solid var(--color-divider); }
.ir2:last-child { border-bottom: none; }
.ik { width: 140rpx; flex-shrink: 0; font-size: 24rpx; color: #667085; }
.iv { flex: 1; font-size: 30rpx; color: var(--color-text); word-break: break-all; } /* 值 15px：同研发难题详情 .iv */
/* 四段签名色同列表页 badge 文字色：实验室灰 / 中试琥珀 / 产业化绿 / 已上市深蓝 */
.iv.cl-la { color: #5D6B82; font-weight: 600; }
.iv.cl-pi { color: #B45309; font-weight: 600; }
.iv.cl-in { color: #1F7A48; font-weight: 600; }
.iv.cl-li { color: #074D92; font-weight: 600; }

/* ===== 附件资料（保留） ===== */
.at-row { display: flex; align-items: center; gap: 16rpx; padding: 20rpx 0; border-bottom: 1rpx solid var(--color-divider); }
.at-row:last-child { border-bottom: none; }
.at-ic { width: 56rpx; height: 56rpx; border-radius: 12rpx; background: var(--color-primary-light); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.at-ic text { font-size: 24rpx; color: var(--color-primary); font-weight: 600; }
.at-info { flex: 1; min-width: 0; }
.at-name { display: block; font-size: 28rpx; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.at-size { display: block; font-size: 20rpx; color: #667085; margin-top: 4rpx; }
.at-btn { flex-shrink: 0; padding: 8rpx 20rpx; border-radius: 8rpx; background: var(--color-primary); color: #fff; font-size: 24rpx; }
.at-btn.dis { background: var(--color-text-placeholder); } /* 有下载在途：其余行禁点（占位灰实底，不做半透明） */
.at-btn.at-cancel { background: #fff; border: 2rpx solid var(--color-text-placeholder); color: var(--color-text-placeholder); }
.at-row.dl { background: var(--color-primary-light); border-radius: 12rpx; }

/* ===== 成果阶段（常驻轨道，同 track.vue；转化记录深链到成果转化页） ===== */
/* 流程：虚线轨道 + 已完成进度动画；间距用 flex/gap 自适应，不写死 */
.tr-flow { position: relative; display: flex; flex-direction: column; gap: 16rpx; }
.tr-track { position: relative; padding: 6rpx 0 10rpx; }
.tr-base { position: absolute; left: 12.5%; right: 12.5%; top: 10rpx; border-top: 2rpx dashed #DDE1E6; z-index: 0; }
.tr-prog { position: absolute; left: 12.5%; top: 9rpx; height: 4rpx; background: var(--color-primary); border-radius: 2rpx; z-index: 1; transform-origin: left center; transform: scaleX(0); transition: transform .4s cubic-bezier(0.16, 1, 0.3, 1); }
.tr-stages { display: flex; justify-content: space-between; position: relative; z-index: 2; }
.tr-stage { display: flex; flex-direction: column; align-items: center; gap: 8rpx; width: 25%; }
.tr-dot { width: 20rpx; height: 20rpx; border-radius: 50%; background: #fff; border: 4rpx solid var(--color-divider); box-sizing: border-box; transition: background .3s ease, border-color .3s ease, box-shadow .3s ease; }
.tr-stage.done .tr-dot { background: var(--color-primary); border-color: var(--color-primary); }
/* 当前点：实心 + 光晕呼吸脉冲（用户点名恢复的循环动画；环只在 6rpx↔14rpx 间透明渐变，实心点不闪灭，状态恒可读；no-motion 门内静止见 .page.no-motion .tr-dot） */
.tr-stage.cur .tr-dot { background: var(--color-primary); border-color: var(--color-primary); box-shadow: 0 0 0 6rpx var(--color-primary-light); animation: trCurPulse 2.2s ease-in-out infinite; }
@keyframes trCurPulse { 0%, 100% { box-shadow: 0 0 0 6rpx #E8F2FC; } 50% { box-shadow: 0 0 0 14rpx rgba(232, 242, 252, 0); } }
.tr-stage-name { font-size: 20rpx; color: #667085; }
.tr-stage.done .tr-stage-name, .tr-stage.cur .tr-stage-name { color: var(--color-primary); font-weight: 600; }
.tr-meta { display: flex; justify-content: space-between; align-items: center; font-size: 24rpx; color: #667085; }
.tr-cur { color: var(--color-primary); font-weight: 600; }
.tr-go { margin-top: 20rpx; padding: 20rpx 0 4rpx; border-top: 1rpx solid var(--color-border); text-align: center; font-size: 26rpx; color: var(--color-primary); font-weight: 600; } /* 28→26rpx：同研发难题 13px 行档 */
.tr-go:active { opacity: .7; }
.tr-err { padding: 24rpx 0; text-align: center; font-size: 26rpx; color: #667085; } /* 可交互错误行：家族灰（≥4.5:1），与 tr-ongoing 同为状态行族，不再用占位灰 */
.tr-err:active { opacity: .7; }
.tr-ongoing { margin-top: 20rpx; padding: 20rpx 0 4rpx; border-top: 1rpx solid var(--color-border); text-align: center; font-size: 26rpx; color: #667085; }

/* ===== 底部操作栏（常驻：主 CTA 不随滚动撤走，同 track.vue 行为） ===== */
.bb { position: sticky; bottom: 0; z-index: 50; background: #fff; border-top: 1rpx solid var(--color-divider); display: flex; align-items: center; padding: 20rpx 32rpx; gap: 20rpx; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); box-shadow: 0 -2rpx 12rpx rgba(0,0,0,.04); }
.bi { width: 88rpx; height: 88rpx; border-radius: 16rpx; background: #F4F6F8; display: flex; align-items: center; justify-content: center; flex-shrink: 0; } /* 方形 88rpx 圆角 16rpx：同研发难题 .bi（原圆形） */
.bi:active { transform: scale(.9); background: #EAF3FB; }
.bi.fv:active { background: #FDECEC; } /* 收藏态按压：浅红底（同研发难题） */
.bit { width: 40rpx; height: 40rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23667085' stroke-width='2' stroke-linejoin='round'%3E%3Cpath d='M12 2.5l2.9 6.1 6.6.9-4.8 4.6 1.2 6.6L12 17.6l-5.9 3.1 1.2-6.6-4.8-4.6 6.6-.9z'/%3E%3C/svg%3E"); background-size: contain; background-repeat: no-repeat; } /* 心形 20px：同研发难题 .bit（原 44rpx） */
/* 收藏态：实心红心 #ff3b30（同研发难题收藏态；灰描边 = 未收藏，与列表页灰星同家族） */
.bi.fv .bit { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23ff3b30'%3E%3Cpath d='M12 2.5l2.9 6.1 6.6.9-4.8 4.6 1.2 6.6L12 17.6l-5.9 3.1 1.2-6.6-4.8-4.6 6.6-.9z'/%3E%3C/svg%3E"); }
.bp { flex: 1; height: 88rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 28rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; }
.bp:active { transform: scale(.97); }
/* 「转发」文字按钮（同研发难题 .bo）：白底蓝描边，44px 高 13px/600 */
.bo {
  height: 88rpx;
  border-radius: 16rpx;
  border: 2rpx solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  font-size: 26rpx;
  font-weight: 600;
  padding: 0 32rpx;
  margin: 0;
  line-height: 1;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.bo::after { border: none; }
.bo-hover { transform: scale(.95); background: #F4F8FC; }

/* ===== 收藏弹星（SVG 绘制；ios-pop 弹簧仅 transform；锚定收藏按钮中心：栏 padding 20 + 按钮半高 44 = 64rpx，translateY(50%) 抵消星自身半高——修复残留 -50% 偏移） ===== */
.fp { position: fixed; left: 76rpx; bottom: calc(64rpx + env(safe-area-inset-bottom)); width: 96rpx; height: 96rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23ff3b30'%3E%3Cpath d='M12 2.5l2.9 6.1 6.6.9-4.8 4.6 1.2 6.6L12 17.6l-5.9 3.1 1.2-6.6-4.8-4.6 6.6-.9z'/%3E%3C/svg%3E"); background-size: contain; background-repeat: no-repeat; transform: translate(-50%, 50%) scale(0); z-index: 100; pointer-events: none; transition: transform .3s cubic-bezier(0.34, 1.8, 0.64, 1), opacity .3s ease; opacity: 0; } /* 收藏弹星随收藏态同色（红心） */
.fp:not(.hide) { transform: translate(-50%, 50%) scale(1); opacity: 1; }
.fp.hide { transform: translate(-50%, 50%) scale(1.2); opacity: 0; }

/* ===== Skeleton（保留） ===== */
/* 骨架镜像刊头终态结构（bar + tag + 双行标题 + 副行），替换 16/9 英雄图暗示；paddingTop 与真刊头一致，交换零跳动 */
.sk-h { box-sizing: border-box; padding: 0 40rpx 44rpx; background: linear-gradient(160deg, #0a3a6b, #074d92); animation: shimmer 1.5s ease; }
.sk-h-bar { height: 88rpx; }
.sk-h-line { background: rgba(255,255,255,.3); border-radius: 8rpx; margin-bottom: 20rpx; animation: shimmer 1.5s ease; }
.sk-h-tag { height: 34rpx; width: 40%; }
.sk-h-title { height: 56rpx; width: 90%; }
.sk-h-sub { height: 24rpx; width: 60%; margin-bottom: 0; }
.sk-cover { margin: 0 32rpx 24rpx; aspect-ratio: 4/3; background: #EDF0F3; border-radius: 20rpx; animation: shimmer 1.5s ease; } /* 骨架镜像图位卡（同终态结构，交换零跳动） */
.sk-stats { display: flex; margin: 0 32rpx 24rpx; padding: 32rpx 0; gap: 16rpx; background: #fff; border: 2rpx solid #E4E7EC; border-radius: 20rpx; box-shadow: 0 2rpx 6rpx rgba(10,30,60,.04), 0 12rpx 32rpx rgba(10,30,60,.05); }
.sk-stat { flex: 1; height: 120rpx; background: #EDF0F3; border-radius: 16rpx; animation: shimmer 1.5s ease; }
.sk-sec { margin: 0 32rpx 24rpx; padding: 32rpx; background: #fff; border: 2rpx solid #E4E7EC; border-radius: 20rpx; }
.sk-l { height: 28rpx; background: #EDF0F3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s ease; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== State（保留） ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 40rpx; min-height: 800rpx; }
.stb { min-height: 88rpx; padding: 16rpx 48rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 26rpx; font-weight: 500; display: flex; align-items: center; justify-content: center; } /* 按钮 13px：同研发难题 .stb */
.stb:active { opacity: .8; }

/* ===================== UI/UX 体验优化（仅新增/修改 wxss，不动模板/数据/逻辑） ===================== */
/* 动画统一 200-400ms；优先 transform/opacity；生产级轻量克制 */

/* 1) 入场动画：附件行、进度行依次淡入（20ms 错开，≤6 项；backwards） */
.sec .at-row { animation: uiRowIn .3s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.sec .at-row:nth-child(2) { animation-delay: 0ms; }
.sec .at-row:nth-child(3) { animation-delay: 20ms; }
@keyframes uiRowIn { from { opacity: 0; transform: translateX(10rpx); } to { opacity: 1; transform: translateX(0); } }

/* 2) 交互反馈：列表行/按钮按压轻微缩放 + 透明度（200ms） */
.at-row { transition: transform .2s ease, opacity .2s ease; }
.at-row:active { transform: scale(.99); opacity: .8; }
.tr-go { transition: transform .2s ease, opacity .2s ease; }
.tr-go:active { transform: scale(.98); opacity: .75; }
.bi { transition: transform .2s ease, color .2s ease, background .2s ease; }
.bo { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; } /* ios-pop（同研发难题 .bo） */
.bo:active { transform: scale(.95); background: #F4F8FC; transition: transform .08s linear; }
.bp { transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease; }
.bp:active { transform: scale(.97); opacity: .92; }
.stb { transition: transform .2s ease, opacity .2s ease; }
.stb:active { transform: scale(.95); opacity: .85; }

/* 4) 层级加固：底部操作栏置顶于内容之上；收藏弹心最高层 */
.bb { z-index: 60; }
.fp { z-index: 100; }

/* ===== 【对齐组范式】主行动按钮 gloss：内高光 + 内厚度 + 品牌蓝晕（如需回退删除本块即可） ===== */
.bp { box-shadow: 0 4rpx 14rpx rgba(10,102,194,.25), inset 0 2rpx 0 rgba(255,255,255,.2), inset 0 -4rpx 10rpx rgba(7,77,146,.18); }
.stb { box-shadow: 0 4rpx 14rpx rgba(10,102,194,.25), inset 0 2rpx 0 rgba(255,255,255,.2), inset 0 -4rpx 10rpx rgba(7,77,146,.18); }

/* ===== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈（同 list.vue） ===== */
.page.no-motion .sec { transition: none; animation: none; }
.page.no-motion .at-row { transition: none; animation: none; }
.page.no-motion .tr-prog { transition: none; }
.page.no-motion .tr-dot { transition: none; animation: none; }
.page.no-motion .tr-go { transition: none; }
.page.no-motion .mh-back { transition: none; }
.page.no-motion .bi { transition: none; }
.page.no-motion .bo { transition: none; }
.page.no-motion .bp { transition: none; }
.page.no-motion .stb { transition: none; }
.page.no-motion .fp { transition: none; }
.page.no-motion .bi:active,
.page.no-motion .bo:active,
.page.no-motion .bp:active,
.page.no-motion .stb:active,
.page.no-motion .tr-go:active,
.page.no-motion .tr-err:active { transform: none; }
</style>
