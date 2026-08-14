<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="展会详情" show-back :fixed="true" @back="goBack" />

    <!-- 加载骨架 -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- 错误 / 不存在 -->
    <view v-else-if="err || !d" class="st">
      <u-empty :description="err ? '加载失败，请检查网络' : '该展会已下架或不存在'">
        <view v-if="err" class="stb" @tap="fetchData">重新加载</view>
        <view v-else class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <template v-else>
      <view class="body">
        <!-- Hero -->
        <view class="hero" :class="d.grad">
          <view class="hero-glow"></view>
          <text class="hero-badge">{{ d.catLabel }} · {{ d.statusLabel }}</text>
          <text class="hero-char">{{ d.char }}</text>
          <text class="hero-title">{{ d.title }}</text>
        </view>

        <!-- 信息卡 -->
        <view class="info-card">
          <view class="stat-row">
            <view class="si"><text class="sv">{{ d.dateShort }}</text><text class="sl">展会时间</text></view>
            <view class="si"><text class="sv">{{ d.boothCount }}</text><text class="sl">展位总数</text></view>
            <view class="si"><text class="sv">{{ d.remaining }}</text><text class="sl">剩余展位</text></view>
          </view>
          <view class="info-row">
            <view class="info-ic"><text>时</text></view>
            <view class="info-txt"><text class="info-label">展会时间</text><text class="info-value">{{ d.timeText }}</text></view>
          </view>
          <view class="info-row">
            <view class="info-ic ic-orange"><text>地</text></view>
            <view class="info-txt"><text class="info-label">展会地点</text><text class="info-value">{{ d.location || '—' }}</text></view>
          </view>
          <view class="info-row">
            <view class="info-ic ic-green"><text>主</text></view>
            <view class="info-txt"><text class="info-label">主办单位</text><text class="info-value">{{ d.organizer || '重庆市无人机产业协会' }}</text></view>
          </view>
          <view class="info-row">
            <view class="info-ic ic-purple"><text>价</text></view>
            <view class="info-txt"><text class="info-label">展位价格</text><text class="info-value">{{ d.priceText || '待定' }}</text></view>
          </view>
        </view>

        <!-- 展会介绍 -->
        <view class="sec">
          <view class="sh"><view class="sd"></view><text class="sht">展会介绍</text></view>
          <text class="sb">{{ d.description || '暂无介绍' }}</text>
        </view>

        <!-- 展位平面示意图 -->
        <view class="sec">
          <view class="sh"><view class="sd"></view><text class="sht">展位平面示意图</text></view>
          <view v-if="d.cells.length" class="fp-map">
            <view class="fp-legend">
              <text><view class="fp-dot ok"></view>可选</text>
              <text><view class="fp-dot no"></view>已订</text>
            </view>
            <view class="fp-grid">
              <view v-for="c in d.cells" :key="c.no" class="fp-cell" :class="c.occupied ? 'no' : 'ok'">{{ c.no }}</view>
            </view>
            <text class="fp-note">{{ d.capped ? '图为部分展位示意，完整展位以现场平面图为准；' : '' }}申请提交后由协会审核分配。</text>
          </view>
          <text v-else class="sb dim">展位信息待主办方更新</text>
        </view>

        <!-- 展位动态 -->
        <view class="sec">
          <view class="sh"><view class="sd"></view><text class="sht">展位动态</text></view>
          <view class="booth-stat">
            <view class="bs-card"><text class="bs-v">{{ d.boothCount }}</text><text class="bs-l">展位总数</text></view>
            <view class="bs-card"><text class="bs-v green">{{ d.remaining }}</text><text class="bs-l">剩余展位</text></view>
            <view class="bs-card"><text class="bs-v gray">{{ d.priceText || '—' }}</text><text class="bs-l">标准展位价格</text></view>
          </view>
        </view>
        <view style="height: 110px"></view>
      </view>

      <!-- 底部操作栏 -->
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" @tap="toggleFav">
          <text class="bit">{{ isFav ? '♥' : '♡' }}</text>
          <text class="bil">{{ isFav ? '已收藏' : '收藏' }}</text>
        </view>
        <view class="bo" @tap="onContact">联系主办</view>
        <view class="bp" :class="{ disabled: !canApply }" @tap="goBooth">
          {{ d.status === 'ended' ? '已结束' : (d.status === 'underway' ? '进行中' : '申请展位') }}
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { MOCK_EXHIBITIONS, MOCK_BOOTHS_BY_EXPO, EXPO_CATEGORY_LABEL, EXPO_STATUS_LABEL, fmtRange, fmtFen, gradOfCategory, buildBoothCells } from '@/utils/mockExhibitions'

const id = ref('')
const d = ref(null)
const loading = ref(true)
const err = ref(false)
const isFav = ref(false)
const statusBarHeight = ref(20)
let toastTimer = null

const canApply = computed(() => !!d.value && (d.value.status === 'recruiting' || d.value.status === 'draft'))

// ===== 数据映射 =====
const buildDetail = (it, booths = []) => {
  const occupied = new Set()
  ;(booths || []).forEach((b) => {
    if (b && b.booth_number && b.status !== 'rejected') occupied.add(String(b.booth_number).toLowerCase())
  })
  const { cells, capped, total } = buildBoothCells(it.booth_count, occupied)
  const start = String(it.start_date || '')
  const end = String(it.end_date || '')
  return {
    id: it.id,
    title: it.title || '',
    catLabel: EXPO_CATEGORY_LABEL[it.category] || it.category || '展会',
    status: it.status || '',
    statusLabel: EXPO_STATUS_LABEL[it.status] || (it.status === 'draft' ? '未发布' : '未知'),
    location: it.location || '',
    organizer: it.organizer || '',
    dateShort: fmtRange(start, end),
    timeText: [start.slice(0, 10), end.slice(0, 10)].filter(Boolean).join(' 至 ') || '',
    boothCount: total || 0,
    remaining: Math.max(total - occupied.size, 0),
    priceText: fmtFen(it.booth_price_fen),
    description: it.desc || it.description || '',
    cover: it.cover_url || '',
    char: '展',
    grad: gradOfCategory(it.category),
    cells,
    capped,
  }
}

// ===== 数据获取 =====
// 接口替换点：GET /api/v1/exhibitions/{id} + GET /api/v1/exhibitions/{id}/booths
const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    const [expoRes, boothRes] = await Promise.all([
      request({ url: '/api/v1/exhibitions/' + encodeURIComponent(id.value) }),
      request({ url: '/api/v1/exhibitions/' + encodeURIComponent(id.value) + '/booths' }).catch(() => []),
    ])
    const it = expoRes && expoRes.id ? expoRes : null
    if (it) {
      d.value = buildDetail(it, Array.isArray(boothRes) ? boothRes : [])
      return
    }
    throw new Error('empty')
  } catch {
    // 回退：演示数据 / 列表缓存
    const cached = uni.getStorageSync('exhibition_cache_' + id.value)
    const mock = MOCK_EXHIBITIONS.find((x) => x.id === id.value)
    const src = mock || cached
    if (src) {
      const booths = MOCK_BOOTHS_BY_EXPO[id.value] || []
      d.value = buildDetail(src, booths)
    } else {
      err.value = true
    }
  } finally {
    loading.value = false
  }
}

// ===== 交互 =====
const toggleFav = () => {
  isFav.value = !isFav.value
  showToast(isFav.value ? '已收藏，可在「我的收藏」中查看' : '已取消收藏')
}
const onContact = () => showToast('已复制主办方联系方式')
const goBooth = () => {
  if (!canApply.value) return
  uni.navigateTo({ url: '/pkg-eco/pages/exhibitions/booth?id=' + encodeURIComponent(id.value) })
}
const showToast = (msg) => {
  uni.showToast({ title: msg, icon: 'none', duration: 1600 })
}
const goBack = () => uni.navigateBack()

onLoad((options) => {
  if (options && options.id) id.value = decodeURIComponent(options.id)
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  fetchData()
})
</script>

<style scoped>
.page { min-height: 100vh; background: #F4F6F8; padding-bottom: 120rpx; }

/* ===== Hero ===== */
.hero { position: relative; height: 460rpx; overflow: hidden; display: flex; align-items: flex-end; }
.hero-glow { position: absolute; inset: 0; background: radial-gradient(120% 90% at 78% 12%, rgba(255,255,255,.28), transparent 46%); }
.hero-badge { position: absolute; left: 32rpx; top: 32rpx; font-size: 22rpx; padding: 6rpx 20rpx; border-radius: 8rpx; font-weight: 600; background: rgba(255,255,255,.92); color: #0A66C2; }
.hero-char { position: absolute; left: 48rpx; bottom: 28rpx; font-size: 160rpx; font-weight: 800; color: rgba(255,255,255,.9); text-shadow: 0 3px 12px rgba(0,0,0,.28); line-height: 1; }
.hero-title { position: absolute; left: 32rpx; right: 32rpx; bottom: 32rpx; font-size: 38rpx; font-weight: 700; color: #fff; text-shadow: 0 2px 8px rgba(0,0,0,.3); line-height: 1.35; z-index: 3; }

/* ===== 信息卡 ===== */
.info-card { position: relative; z-index: 5; margin: -60rpx 24rpx 0; background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; padding: 28rpx 32rpx 12rpx; box-shadow: 0 8rpx 32rpx rgba(0,0,0,.05); }
.stat-row { display: flex; padding: 16rpx 0 20rpx; border-bottom: 1px solid #EBEDF0; }
.si { flex: 1; text-align: center; position: relative; }
.si + .si::before { content: ''; position: absolute; left: 0; top: 12rpx; bottom: 12rpx; width: 1rpx; background: #F0F0F0; }
.sv { font-size: 32rpx; font-weight: 700; color: #17212B; display: block; }
.sl { font-size: 20rpx; color: #98A2B3; margin-top: 4rpx; display: block; }
.info-row { display: flex; align-items: flex-start; gap: 20rpx; padding: 18rpx 0; border-bottom: 1rpx solid #F5F5F5; }
.info-row:last-child { border-bottom: none; }
.info-ic { width: 56rpx; height: 56rpx; border-radius: 12rpx; background: #EAF3FB; color: #0A66C2; display: flex; align-items: center; justify-content: center; flex: none; font-size: 24rpx; font-weight: 600; }
.info-ic.ic-orange { background: #FFF0E6; color: #E96012; }
.info-ic.ic-green { background: #E9F7F0; color: #168A55; }
.info-ic.ic-purple { background: #F6F4FF; color: #7A5AF8; }
.info-txt { flex: 1; min-width: 0; }
.info-label { font-size: 20rpx; color: #98A2B3; margin-bottom: 2rpx; display: block; }
.info-value { font-size: 26rpx; color: #17212B; font-weight: 500; line-height: 1.4; display: block; }

/* ===== 区块 ===== */
.sec { margin: 20rpx 24rpx 0; padding: 28rpx 32rpx; background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; }
.sh { display: flex; align-items: center; gap: 16rpx; margin-bottom: 24rpx; }
.sd { width: 8rpx; height: 30rpx; background: #0A66C2; border-radius: 4rpx; flex-shrink: 0; }
.sht { font-size: 30rpx; font-weight: 700; color: #17212B; }
.sb { font-size: 26rpx; color: #667085; line-height: 1.75; white-space: pre-wrap; display: block; }
.sb.dim { color: #98A2B3; }

/* ===== 展位平面示意图 ===== */
.fp-map { margin-top: 4rpx; border: 1px solid #EEF1F4; border-radius: 20rpx; padding: 24rpx; background: #FBFCFE; }
.fp-legend { display: flex; gap: 28rpx; margin-bottom: 20rpx; font-size: 20rpx; color: #667085; align-items: center; }
.fp-dot { display: inline-block; width: 18rpx; height: 18rpx; border-radius: 4rpx; margin-right: 6rpx; vertical-align: middle; }
.fp-dot.ok { background: #34C759; }
.fp-dot.no { background: #D8DDE3; }
.fp-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 12rpx; }
.fp-cell { aspect-ratio: 1.4; border-radius: 8rpx; display: flex; align-items: center; justify-content: center; font-size: 18rpx; font-weight: 600; color: #fff; }
.fp-cell.ok { background: linear-gradient(135deg,#2ea44f,#34c759); }
.fp-cell.no { background: #D8DDE3; color: #98A2B3; }
.fp-note { font-size: 20rpx; color: #98A2B3; margin-top: 20rpx; line-height: 1.6; display: block; }

/* ===== 展位动态 ===== */
.booth-stat { display: flex; gap: 20rpx; }
.bs-card { flex: 1; text-align: center; background: #FBFCFE; border: 1px solid #EEF1F4; border-radius: 16rpx; padding: 24rpx 16rpx; }
.bs-v { font-size: 34rpx; font-weight: 700; color: #0A66C2; display: block; }
.bs-v.green { color: #168A55; }
.bs-v.gray { color: #667085; }
.bs-l { font-size: 20rpx; color: #98A2B3; margin-top: 6rpx; display: block; }

/* ===== 底部操作栏 ===== */
.bb { position: fixed; left: 0; right: 0; bottom: 0; display: flex; align-items: center; padding: 16rpx 24rpx; padding-bottom: calc(16rpx + env(safe-area-inset-bottom)); background: #fff; border-top: 1rpx solid #F0F0F0; box-shadow: 0 -4rpx 24rpx rgba(0,0,0,.04); gap: 20rpx; z-index: 50; }
.bi { width: 96rpx; height: 84rpx; border-radius: 16rpx; background: #F4F6F8; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 4rpx; flex-shrink: 0; }
.bi.fv { color: #ff3b30; background: #FFF0F0; }
.bit { font-size: 36rpx; line-height: 1; }
.bil { font-size: 18rpx; color: #667085; }
.bi.fv .bil { color: #ff3b30; }
.bo { height: 84rpx; border-radius: 16rpx; border: 3rpx solid #0A66C2; background: #fff; color: #0A66C2; font-size: 26rpx; font-weight: 600; padding: 0 32rpx; display: flex; align-items: center; flex-shrink: 0; }
.bp { flex: 1; height: 84rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 30rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.35); }
.bp.disabled { background: #98A2B3; box-shadow: none; }

/* ===== 骨架 ===== */
.skw { padding-top: 20rpx; }
.sk-h { height: 432rpx; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.sk-sec { margin: 24rpx; padding: 32rpx; background: #fff; border-radius: 24rpx; }
.sk-l { height: 28rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 48rpx; min-height: 600rpx; }
.stb { padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; margin-top: 24rpx; }

/* 封面渐变 */
.gd-1 { background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5); }
.gd-2 { background: linear-gradient(135deg,#004d40,#00695c 60%,#26a69a); }
.gd-4 { background: linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc); }
.gd-5 { background: linear-gradient(135deg,#263238,#37474f 60%,#607d8b); }
</style>
