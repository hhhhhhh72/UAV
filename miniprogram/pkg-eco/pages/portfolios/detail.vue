<template>
  <view class="page" :style="{ '--status-bar': statusBarHeight + 'px', paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="品牌详情" show-back :fixed="true" @back="goBack" />

    <!-- 加载骨架 -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- 错误 -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <template v-else-if="d">
      <scroll-view scroll-y class="body">
        <!-- 品牌 Hero：封面图（cover_url）或渐变占位；无视频不显示播放按钮 -->
        <view class="d-hero" :class="d.grad">
          <image v-if="d.cover" :src="d.cover" mode="aspectFill" class="d-hero-img" />
          <view class="glow"></view>
          <view v-if="d.videoList.length" class="play-btn" hover-class="tap-fade" @tap="openPlayer(d.videoList[0])">
            <view class="pulse"></view>
            <text>▶</text>
          </view>
          <view class="hero-badge">{{ d.catLabel }} · {{ d.verified ? '已认证' : '认证中' }}</view>
        </view>

        <!-- 上浮信息卡 -->
        <view class="d-info">
          <view class="d-top">
            <view class="d-logo">
              <image v-if="d.logo" :src="d.logo" mode="aspectFill" class="d-logo-img" />
              <text v-else>{{ d.logoText }}</text>
            </view>
            <view class="d-name-wrap">
              <text class="d-name">{{ d.name }}</text>
              <view class="d-cert">
                <text v-if="d.verified" class="d-badge db-ok">✓ 已认证</text>
                <text v-else class="d-badge db-ing">认证中</text>
                <text v-if="d.isVip" class="d-badge db-vip">理事单位</text>
                <text v-if="d.featured" class="d-badge db-vip">✦ 精选品牌</text>
              </view>
            </view>
          </view>
          <text class="d-desc">{{ d.desc || '暂无品牌简介' }}</text>
          <view v-if="d.caseCount > 0 || d.videoCount > 0 || d.views > 0" class="d-stat">
            <view class="d-si"><text class="d-sv">{{ d.caseCount }}</text><text class="d-sl">品牌案例</text></view>
            <view class="d-si"><text class="d-sv">{{ d.videoCount }}</text><text class="d-sl">宣传视频</text></view>
            <view class="d-si"><text class="d-sv">{{ fmt(d.views) }}</text><text class="d-sl">品牌浏览</text></view>
          </view>
        </view>

        <!-- 企业信息 -->
        <view class="d-sec">
          <view class="d-sh"><view class="d-sd"></view><text class="d-sht">企业信息</text></view>
          <view class="d-row">
            <view class="d-ic"><text>域</text></view>
            <view class="d-it"><text class="d-il">主营领域</text><text class="d-iv">{{ (d.fields || []).join(' · ') || '—' }}</text></view>
          </view>
          <view class="d-row">
            <view class="d-ic ic-orange"><text>址</text></view>
            <view class="d-it"><text class="d-il">所在地</text><text class="d-iv">{{ d.location || '—' }}</text></view>
          </view>
          <view class="d-row">
            <view class="d-ic ic-green"><text>龄</text></view>
            <view class="d-it"><text class="d-il">成立时间</text><text class="d-iv">{{ d.founded || '—' }}</text></view>
          </view>
          <view class="d-row">
            <view class="d-ic ic-purple"><text>证</text></view>
            <view class="d-it"><text class="d-il">资质认证</text><text class="d-iv">{{ (d.certs || []).join(' · ') || '—' }}</text></view>
          </view>
        </view>

        <!-- 产品展示 -->
        <view class="d-sec">
          <view class="d-sh"><view class="d-sd"></view><text class="d-sht">产品展示<text class="more">共 {{ d.products.length }} 款</text></text></view>
          <view v-if="d.products.length" class="p-grid">
            <view v-for="(p, i) in d.products" :key="i" class="p-card" hover-class="tap-fade">
              <view class="p-cv" :class="p.grad"><text class="p-char">{{ p.char }}</text><text class="p-tag">{{ d.catLabel }}</text></view>
              <view class="p-bd">
                <text class="p-name">{{ p.name }}</text>
                <text v-if="p.meta" class="p-meta">{{ p.meta }}</text>
              </view>
            </view>
          </view>
          <text v-else class="empty-tip">暂无产品展示</text>
        </view>

        <!-- 荣誉资质 -->
        <view class="d-sec">
          <view class="d-sh"><view class="d-sd"></view><text class="d-sht">荣誉资质<text class="more">共 {{ d.honors.length }} 项</text></text></view>
          <view v-if="d.honors.length" class="h-grid">
            <view v-for="(h, i) in d.honors" :key="i" class="h-card" hover-class="tap-fade">
              <view class="h-ic"><text>🏆</text></view>
              <view class="h-txt">
                <text class="h-t">{{ h.name }}</text>
                <text v-if="h.meta" class="h-d">{{ h.meta }}</text>
              </view>
            </view>
          </view>
          <text v-else class="empty-tip">暂无荣誉资质</text>
        </view>

        <!-- 宣传视频 -->
        <view class="d-sec">
          <view class="d-sh"><view class="d-sd"></view><text class="d-sht">宣传视频<text class="more">共 {{ d.videoList.length }} 条</text></text></view>
          <view v-if="d.videoList.length" class="v-grid">
            <view
              v-for="(v, i) in d.videoList"
              :key="i"
              class="v-card"
              :class="v.grad"
              hover-class="tap-fade"
              @tap="openPlayer(v)"
            >
              <view class="vplay"><text>▶</text></view>
              <text class="vtime">{{ v.duration }}</text>
              <text class="vt">{{ v.title }}</text>
            </view>
          </view>
          <text v-else class="empty-tip">暂无宣传视频</text>
        </view>

        <!-- 品牌案例 -->
        <view class="d-sec">
          <view class="d-sh"><view class="d-sd"></view><text class="d-sht">品牌案例<text class="more">查看全部 ›</text></text></view>
          <view v-if="d.cases.length" class="c-grid">
            <view
              v-for="(c, i) in d.cases"
              :key="i"
              class="c-card"
              hover-class="tap-fade"
            >
              <view class="c-cv" :class="c.grad"><text class="c-ctag">{{ c.tag }}</text></view>
              <view class="c-bd">
                <text class="c-t">{{ c.title }}</text>
                <text class="c-d">{{ c.meta }}</text>
              </view>
            </view>
          </view>
          <text v-else class="empty-tip">暂无品牌案例</text>
        </view>
        <view style="height: 120rpx"></view>
      </scroll-view>

      <!-- 底部操作栏 -->
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" @tap="toggleFav">
          <text class="bit">{{ isFav ? '♥' : '♡' }}</text>
        </view>
        <view class="bo" @tap="onWechat">微信</view>
        <view class="bp" @tap="onPhone">📞 电话咨询</view>
      </view>
    </template>

    <!-- 视频播放浮层 -->
    <BrandVideoPlayer
      :show="playerShow"
      :title="playerTitle"
      :duration="playerDuration"
      @close="closePlayer"
    />

    <!-- toast -->
    <view class="toast" :class="{ show: toastShow }">{{ toastText }}</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage, BASE_URL } from '@/utils/request'
import BrandVideoPlayer from '../../components/BrandVideoPlayer.vue'
import { MOCK_BRANDS, CATEGORY_MAP } from '@/utils/mockBrands'

const id = ref('')
const d = ref(null)
const loading = ref(true)
const err = ref(false)
const statusBarHeight = ref(20)
const isFav = ref(false)

// 视频播放
const playerShow = ref(false)
const playerTitle = ref('')
const playerDuration = ref('03:00')

// toast
const toastShow = ref(false)
const toastText = ref('')
let toastTimer = null

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
const gradCls = (g) => 'gd-' + String(g || 'gd1').replace(/^gd-?/, '')
const fmt = (n) => (n >= 10000 ? (n / 10000).toFixed(1) + 'w' : n >= 1000 ? (n / 1000).toFixed(1) + 'k' : String(n))

// 相对路径（存库格式）→ 完整 URL
const resolveUrl = (u) => {
  if (!u) return ''
  if (u.indexOf('http') === 0) return u
  return BASE_URL + u
}

// 后端字段 → 详情展示字段（优雅降级）
const buildDetail = (it) => {
  const catKey = keyOf(it.category || it.industry)
  const videos = (it.video_list || it.videos || []).map((v) => ({
    title: v.title || v.name || '宣传视频',
    duration: v.duration || '03:00',
    grad: gradCls(v.grad || it.grad),
  }))
  const cases = (it.cases || it.case_list || []).map((c) => ({
    title: c.title || '',
    tag: c.tag || c.category || '案例',
    meta: c.meta || c.date || '',
    grad: gradCls(c.grad || it.grad),
  }))
  const products = (it.products || []).map((p) => {
    const name = typeof p === 'string' ? p : (p && (p.name || p.title)) || ''
    return {
      name,
      char: name ? String(name).charAt(0) : '品',
      meta: typeof p === 'object' && p.meta ? p.meta : '',
      grad: gradCls(it.grad),
    }
  })
  const honors = (it.honors || []).map((h) => {
    const name = typeof h === 'string' ? h : (h && (h.name || h.title)) || ''
    return { name, meta: typeof h === 'object' && h.meta ? h.meta : '' }
  })
  return {
    id: it.id,
    name: it.name || it.company_name || '',
    catKey,
    catLabel: CATEGORY_MAP[catKey] || it.category || '品牌',
    logo: resolveUrl(it.logo_url || ''),
    cover: resolveUrl(it.cover_url || ''),
    logoText: it.logo_text || (it.name ? String(it.name).charAt(0) : '牌'),
    verified: it.status === 'published', // 已公示 = 协会已认证
    isVip: !!it.is_vip,
    featured: !!it.featured,
    views: 0, // 后端暂无统计字段
    videoCount: videos.length,
    caseCount: cases.length,
    grad: gradCls(it.grad),
    desc: it.desc || it.description || '',
    fields: it.fields || [],
    location: it.location || '',
    founded: it.founded || '',
    certs: it.certs || [],
    products,
    honors,
    videoList: videos,
    cases,
  }
}

// 接口替换点：GET /api/v1/portfolios/:id；失败/为空时回退 mock
const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    const res = await request({ url: '/api/v1/portfolios/' + encodeURIComponent(id.value) })
    const it = (res && res.data) || res
    if (it && it.id) {
      d.value = buildDetail(it)
      return
    }
    throw new Error('empty')
  } catch {
    // 后端暂无按 id 单查接口：优先使用列表页写入的缓存，其次演示数据
    const cached = uni.getStorageSync('portfolio_cache_' + id.value)
    if (cached && cached.id) {
      d.value = buildDetail(cached)
    } else if (String(id.value).startsWith('demo-')) {
      if (import.meta.env.DEV) {
        useMock()
      } else {
        err.value = true
      }
    } else {
      err.value = true
    }
  } finally {
    loading.value = false
  }
}

const useMock = () => {
  const found = MOCK_BRANDS.find((x) => x.id === id.value)
  d.value = buildDetail(found || MOCK_BRANDS[0])
}

// ===== 交互 =====
const openPlayer = (v) => {
  playerTitle.value = (v && v.title) || '品牌宣传视频'
  playerDuration.value = (v && v.duration) || '03:00'
  playerShow.value = true
}
const closePlayer = () => {
  playerShow.value = false
}

const toggleFav = () => {
  const token = authStorage.getAccessToken()
  if (!token && !String(id.value).startsWith('demo-')) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  isFav.value = !isFav.value
  showToast(isFav.value ? '已收藏' : '已取消收藏')
}

const onWechat = () => showToast('已复制企业微信')
const onPhone = () => showToast('正在呼叫企业电话…')

const showToast = (msg) => {
  toastText.value = msg
  toastShow.value = true
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastShow.value = false }, 1600)
}

const goBack = () => uni.navigateBack()

onLoad((options) => {
  if (options?.id) id.value = decodeURIComponent(options.id)
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  fetchData()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}
.tap-fade { opacity: .72; }
.body { height: calc(100vh - 76px - env(safe-area-inset-bottom)); }

/* ===== Hero ===== */
.d-hero { position: relative; height: 432rpx; overflow: hidden; }
.d-hero-img { position: absolute; inset: 0; width: 100%; height: 100%; }
.glow { position: absolute; inset: 0; }
.glow::after { content: ''; position: absolute; top: -30%; right: -10%; width: 440rpx; height: 440rpx; border-radius: 50%; background: radial-gradient(circle, rgba(255,255,255,.12) 0%, transparent 70%); }
.play-btn {
  position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%);
  width: 120rpx; height: 120rpx; border-radius: 50%;
  background: rgba(255,255,255,.22); border: 2px solid rgba(255,255,255,.75);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 40rpx;
  box-shadow: 0 8px 24px rgba(0,0,0,.3);
  z-index: 2;
}
.pulse { position: absolute; inset: -4rpx; border-radius: 50%; border: 2px solid rgba(255,255,255,.5); animation: pulse 2s ease-out infinite; }
@keyframes pulse { 0% { transform: scale(1); opacity: 1; } 100% { transform: scale(1.55); opacity: 0; } }
.hero-badge {
  position: absolute; left: 32rpx; bottom: 28rpx; z-index: 3;
  display: inline-block; font-size: 22rpx; padding: 6rpx 20rpx; border-radius: 8rpx;
  font-weight: 600; background: rgba(255,255,255,.92); color: #0A66C2;
}

/* ===== 信息卡 ===== */
.d-info {
  position: relative; z-index: 5; margin: -68rpx 24rpx 0;
  background: #fff; border: 1px solid #EEF1F4; border-radius: 24rpx;
  padding: 32rpx; box-shadow: 0 4px 16px rgba(0,0,0,.05);
  animation: riseIn .4s cubic-bezier(.2,.9,.3,1) both;
}
@keyframes riseIn { from { opacity: 0; transform: translateY(14px); } to { opacity: 1; transform: translateY(0); } }
.d-top { display: flex; align-items: center; gap: 24rpx; margin-bottom: 24rpx; }
.d-logo {
  width: 104rpx; height: 104rpx; border-radius: 28rpx; background: #EAF3FB;
  display: flex; align-items: center; justify-content: center;
  font-size: 44rpx; font-weight: 700; color: #0A66C2; flex: none;
  overflow: hidden;
}
.d-logo-img { width: 100%; height: 100%; }
.d-name-wrap { flex: 1; min-width: 0; }
.d-name { font-size: 36rpx; font-weight: 700; color: #17212B; line-height: 1.3; display: block; }
.d-cert { margin-top: 12rpx; display: flex; gap: 12rpx; flex-wrap: wrap; }
.d-badge { font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 8rpx; font-weight: 500; }
.db-ok { color: #168A55; background: #E9F7F0; }
.db-ing { color: #E96012; background: #FFF0E6; }
.db-vip { color: #7A5AF8; background: #F6F4FF; }
.d-desc { font-size: 26rpx; color: #667085; line-height: 1.7; border-top: 1px solid #EBEDF0; padding-top: 24rpx; display: block; }
.d-stat { display: flex; padding: 24rpx 0 8rpx; border-top: 1px solid #EBEDF0; margin-top: 24rpx; }
.d-si { flex: 1; text-align: center; position: relative; }
.d-si + .d-si::before { content: ''; position: absolute; left: 0; top: 6px; bottom: 6px; width: .5px; background: #F0F0F0; }
.d-sv { font-size: 30rpx; font-weight: 700; color: #17212B; display: block; }
.d-sl { font-size: 21rpx; color: #98A2B3; margin-top: 4rpx; display: block; }

/* ===== 区块 ===== */
.d-sec { margin: 20rpx 24rpx 0; padding: 28rpx 32rpx; background: #fff; border: 1px solid #EEF1F4; border-radius: 24rpx; animation: riseIn .4s cubic-bezier(.2,.9,.3,1) both; }
.d-sec:nth-of-type(2) { animation-delay: .05s; }
.d-sec:nth-of-type(3) { animation-delay: .1s; }
.d-sec:nth-of-type(4) { animation-delay: .15s; }
.d-sh { display: flex; align-items: center; gap: 16rpx; margin-bottom: 24rpx; }
.d-sd { width: 8rpx; height: 32rpx; background: #0A66C2; border-radius: 4rpx; flex-shrink: 0; }
.d-sht { font-size: 30rpx; font-weight: 700; color: #17212B; display: flex; align-items: center; width: 100%; }
.more { margin-left: auto; font-size: 22rpx; color: #98A2B3; font-weight: 400; }
.d-row { display: flex; align-items: flex-start; gap: 20rpx; padding: 18rpx 0; border-bottom: .5px solid #F5F5F5; }
.d-row:last-child { border-bottom: none; }
.d-ic {
  width: 60rpx; height: 60rpx; border-radius: 16rpx; background: #EAF3FB; color: #0A66C2;
  display: flex; align-items: center; justify-content: center; flex: none;
  font-size: 24rpx; font-weight: 600;
}
.d-ic.ic-orange { background: #FFF0E6; color: #E96012; }
.d-ic.ic-green { background: #E9F7F0; color: #168A55; }
.d-ic.ic-purple { background: #F6F4FF; color: #7A5AF8; }
.d-it { flex: 1; min-width: 0; }
.d-il { font-size: 21rpx; color: #98A2B3; margin-bottom: 2rpx; display: block; }
.d-iv { font-size: 27rpx; color: #17212B; font-weight: 500; line-height: 1.4; display: block; word-break: break-all; }

/* ===== 视频网格 ===== */
.v-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; }
.v-card { position: relative; padding-top: 62.5%; border-radius: 20rpx; overflow: hidden; }
.vplay {
  position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%);
  width: 64rpx; height: 64rpx; border-radius: 50%;
  background: rgba(255,255,255,.9); display: flex; align-items: center; justify-content: center;
  color: #0A66C2; font-size: 24rpx;
}
.vtime { position: absolute; right: 12rpx; bottom: 12rpx; font-size: 18rpx; color: #fff; background: rgba(0,0,0,.55); padding: 2rpx 12rpx; border-radius: 8rpx; }
.vt {
  position: absolute; left: 16rpx; right: 16rpx; bottom: 12rpx;
  font-size: 20rpx; color: #fff; text-shadow: 0 1px 4px rgba(0,0,0,.5);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

/* ===== 案例横滑 ===== */
.c-grid { display: flex; gap: 16rpx; overflow-x: auto; }
.c-card { flex: none; width: 360rpx; border-radius: 20rpx; overflow: hidden; border: 1px solid #EEF1F4; background: #fff; }
.c-cv { padding-top: 56.25%; position: relative; }
.c-ctag { position: absolute; top: 12rpx; left: 12rpx; font-size: 18rpx; color: #fff; background: rgba(0,0,0,.45); padding: 4rpx 16rpx; border-radius: 8rpx; }
.c-bd { padding: 16rpx 20rpx; }
.c-t { font-size: 24rpx; font-weight: 600; color: #17212B; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.c-d { font-size: 20rpx; color: #98A2B3; margin-top: 8rpx; display: block; }

/* ===== 产品展示 ===== */
.p-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; }
.p-card { background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; overflow: hidden; transition: transform .2s cubic-bezier(.25,.8,.3,1); }
.p-card:active { transform: scale(.96); }
.p-cv { position: relative; padding-top: 72%; display: flex; align-items: center; justify-content: center; }
.p-char { font-size: 60rpx; font-weight: 800; color: rgba(255,255,255,.95); text-shadow: 0 2px 10px rgba(0,0,0,.25); }
.p-tag { position: absolute; top: 16rpx; left: 16rpx; font-size: 19rpx; color: #fff; background: rgba(0,0,0,.38); padding: 4rpx 16rpx; border-radius: 12rpx; }
.p-bd { padding: 18rpx 22rpx; }
.p-name { font-size: 25rpx; font-weight: 600; color: #17212B; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.p-meta { font-size: 20rpx; color: #98A2B3; margin-top: 6rpx; display: block; }

/* ===== 荣誉资质 ===== */
.h-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; }
.h-card { display: flex; align-items: center; gap: 16rpx; background: #FAFBFD; border: 1px solid #EEF1F4; border-radius: 20rpx; padding: 20rpx; transition: transform .2s cubic-bezier(.25,.8,.3,1); }
.h-card:active { transform: scale(.96); }
.h-ic { width: 60rpx; height: 60rpx; border-radius: 16rpx; background: #FFF4E5; display: flex; align-items: center; justify-content: center; font-size: 30rpx; flex: none; }
.h-txt { flex: 1; min-width: 0; }
.h-t { font-size: 24rpx; font-weight: 600; color: #17212B; line-height: 1.35; display: block; }
.h-d { font-size: 19rpx; color: #98A2B3; margin-top: 4rpx; display: block; }

.empty-tip { font-size: 24rpx; color: #98A2B3; display: block; }

/* ===== 底部操作栏 ===== */
.bb {
  position: fixed; left: 0; right: 0; bottom: 0;
  display: flex; align-items: center; padding: 20rpx 24rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff; border-top: .5px solid #F0F0F0;
  box-shadow: 0 -2px 12px rgba(0,0,0,.04); gap: 20rpx; z-index: 50;
}
.bi { width: 84rpx; height: 84rpx; border-radius: 50%; background: #F4F6F8; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.bi.fv { color: #ff3b30; }
.bit { font-size: 40rpx; line-height: 1; }
.bp {
  flex: 1; height: 88rpx; border-radius: 16rpx; background: #0A66C2; color: #fff;
  font-size: 30rpx; font-weight: 600; display: flex; align-items: center; justify-content: center;
  box-shadow: 0 2px 8px rgba(10,102,194,.35);
}
.bp:active { transform: scale(.97); }
.bo {
  height: 88rpx; border-radius: 16rpx; border: 1.5px solid #0A66C2; background: #fff; color: #0A66C2;
  font-size: 28rpx; font-weight: 600; padding: 0 32rpx; display: flex; align-items: center; flex-shrink: 0;
}
.bo:active { background: #EAF3FB; }

/* ===== toast ===== */
.toast {
  position: fixed; top: 50%; left: 50%; transform: translate(-50%,-50%) scale(.85);
  background: rgba(23,33,43,.92); color: #fff; font-size: 26rpx;
  padding: 24rpx 44rpx; border-radius: 20rpx;
  opacity: 0; pointer-events: none; transition: all .28s cubic-bezier(.2,1.2,.4,1); z-index: 100;
}
.toast.show { opacity: 1; transform: translate(-50%,-50%) scale(1); }

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
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 40rpx; min-height: 600rpx; }
.stb { padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; margin-top: 24rpx; }

/* 封面渐变 */
.gd-1 { background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5); }
.gd-2 { background: linear-gradient(135deg,#004d40,#00695c 60%,#26a69a); }
.gd-3 { background: linear-gradient(135deg,#e65100,#ef6c00 60%,#fb8c00); }
.gd-4 { background: linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc); }
.gd-5 { background: linear-gradient(135deg,#263238,#37474f 60%,#607d8b); }
.gd-6 { background: linear-gradient(135deg,#b71c1c,#c62828 60%,#e57373); }
.gd-7 { background: linear-gradient(135deg,#1a237e,#283593 60%,#5c6bc0); }
.gd-8 { background: linear-gradient(135deg,#004d40,#00695c 60%,#4db6ac); }
</style>
