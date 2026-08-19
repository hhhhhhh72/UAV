<template>
  <view class="actd-page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="活动详情" show-back :fixed="true" @back="goBack" />

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

    <!-- 空 -->
    <view v-else-if="!d" class="st">
      <u-empty description="该活动已下架或不存在">
        <view class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <template v-else>
      <view>
        <!-- Hero -->
        <view class="hero" :style="{ background: heroBg }">
          <view class="hero-glow"></view>
          <text class="hero-ic">{{ d.char }}</text>
          <view class="hero-foot">
            <view class="hero-badge">{{ d.stLabel }}</view>
            <view class="hero-title">{{ d.t }}</view>
          </view>
        </view>

        <!-- 信息卡 -->
        <view class="info-card">
          <view class="tag-row">
            <text class="tag">{{ d.catLabel }}</text>
          </view>
          <view class="stat-row">
            <view class="si"><text class="sv">{{ d.dateShort }}</text><text class="sl">活动日期</text></view>
            <view class="si"><text class="sv">{{ d.total ? d.total + '人' : '不限' }}</text><text class="sl">活动规模</text></view>
            <view class="si"><text class="sv">{{ d.reg + '人' }}</text><text class="sl">已报名</text></view>
          </view>
          <view class="info-row">
            <view class="info-ic"><text>时</text></view>
            <view class="info-txt">
              <text class="info-label">活动时间</text>
              <text class="info-value">{{ d.timeText }}</text>
            </view>
          </view>
          <view class="info-row">
            <view class="info-ic ic-orange"><text>地</text></view>
            <view class="info-txt">
              <text class="info-label">活动地点</text>
              <text class="info-value">{{ d.loc }}</text>
            </view>
          </view>
          <view class="info-row">
            <view class="info-ic ic-green"><text>主</text></view>
            <view class="info-txt">
              <text class="info-label">主办单位</text>
              <text class="info-value">{{ d.organizer || '重庆市无人机产业协会' }}</text>
            </view>
          </view>
          <view class="info-row">
            <view class="info-ic"><text>止</text></view>
            <view class="info-txt">
              <text class="info-label">报名截止</text>
              <text class="info-value" :class="{ 'cl-su': d.deadline }">{{ d.deadline || '额满即止' }}</text>
            </view>
          </view>
        </view>

        <!-- 活动简介 -->
        <view class="sec">
          <view class="sh"><view class="sd"></view><text class="sht">活动简介</text></view>
          <text class="sb">{{ d.description || '暂无简介' }}</text>
        </view>

        <!-- 议程安排 -->
        <view class="sec">
          <view class="sh"><view class="sd"></view><text class="sht">议程安排</text></view>
          <view v-if="d.agenda.length">
            <view v-for="(a, i) in d.agenda" :key="i" class="agenda-item">
              <text class="at">{{ a.time || a.t }}</text>
              <text class="att">{{ a.title || a.name }}</text>
            </view>
          </view>
          <text v-else class="sb dim">暂无议程，详情以现场安排为准</text>
        </view>
        <view style="height: 90px"></view>
      </view>

      <!-- 底部操作栏 -->
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" @tap="toggleFav">
          <text class="bit">{{ isFav ? '♥' : '♡' }}</text>
        </view>
        <view class="bo" @tap="onShare">分享</view>
        <view
          class="bp"
          :class="{ disabled: d.status === 'end' }"
          @tap="goSignup"
        >{{ d.status === 'end' ? '已结束' : '立即报名' }}</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage } from '@/utils/request'
import { dateOf, timeOf } from '@/utils/eventTime'

const TYPE_LABEL = { forum: '行业论坛', salon: '主题沙龙', visit: '企业考察', expo: '行业展会' }
const TYPE_CHAR = { forum: '论', salon: '沙', visit: '察', expo: '展' }
const TYPE_BG = {
  forum: 'linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5)',
  salon: 'linear-gradient(135deg,#e65100,#ef6c00 60%,#fb8c00)',
  visit: 'linear-gradient(135deg,#004d40,#00695c 60%,#26a69a)',
  expo: 'linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc)',
}
const STATUS_LABEL = { open: '报名中', soon: '即将开始', end: '已结束' }

const id = ref('')
const d = ref(null)
const loading = ref(true)
const err = ref(false)
const isFav = ref(false)
const statusBarHeight = ref(20)

const heroBg = computed(() => TYPE_BG[d.value?.catKey] || TYPE_BG.forum)

const keyOf = (c) => {
  const s = String(c || '').toLowerCase()
  if (TYPE_LABEL[s]) return s
  if (s.includes('论坛')) return 'forum'
  if (s.includes('沙龙')) return 'salon'
  if (s.includes('考察')) return 'visit'
  if (s.includes('展')) return 'expo'
  return 'forum'
}
const statusOf = (it) => {
  const s = String(it.status || '').toLowerCase()
  if (s === 'open' || s === '报名中' || s === 'published') return 'open'
  if (s === 'soon' || s === '即将开始' || s === 'upcoming') return 'soon'
  if (s === 'end' || s === '已结束' || s === 'finished' || s === 'closed') return 'end'
  const dd = dateOf(it.start_date || it.event_date || it.start_time || '')
  if (dd && dd < today()) return 'end'
  return 'open'
}
const pad = (n) => (n < 10 ? '0' + n : '' + n)
const today = () => {
  const now = new Date()
  return now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate())
}
const fmt = (key) => {
  if (!key) return ''
  const p = key.split('-')
  return p[1] + '-' + p[2]
}

const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    // API 优先（实时数据）；接口未部署/网络失败时读列表页透传快照兜底
    let it = null
    try {
      const res = await request({ url: '/api/v1/events/' + encodeURIComponent(id.value) })
      it = (res && res.data) || res
    } catch (e) { it = null }
    if (!it || !it.id) {
      try { it = uni.getStorageSync('act_detail_' + id.value) || null } catch (e) { it = null }
    } else {
      try { uni.removeStorageSync('act_detail_' + id.value) } catch (e) { /* 清理失败不影响展示 */ }
    }
    if (it && it.id) {
      const catKey = keyOf(it.category || it.activity_type || it.field || it.event_type)
      const rawTime = it.start_date || it.event_date || it.start_time || ''
      const date = dateOf(rawTime)
      const timeTxt = it.time_text || it.time || timeOf(rawTime)
      const deadline = (it.register_deadline || it.deadline || '').slice(0, 10)
      let agenda = it.agenda || it.schedule || []
      if (!Array.isArray(agenda)) agenda = []
      d.value = {
        id: it.id,
        t: it.title || '',
        catKey,
        char: TYPE_CHAR[catKey] || '活',
        catLabel: TYPE_LABEL[catKey] || it.category || '协会活动',
        status: statusOf(it),
        stLabel: STATUS_LABEL[statusOf(it)] || '报名中',
        dateShort: date ? fmt(date) : '待定',
        timeText: (date ? fmt(date) + (timeTxt ? ' ' + timeTxt : '') : timeTxt) || '时间待定',
        loc: it.location || it.address || '地点待定',
        organizer: it.organizer || '',
        reg: it.reg_count ?? it.registered ?? 0,
        total: it.quota || it.capacity || it.max_attendees || 0,
        deadline: deadline ? fmt(deadline) : '',
        description: it.description || '',
        agenda,
      }
    } else {
      err.value = true
    }
  } catch {
    err.value = true
  } finally {
    loading.value = false
  }
}

const toggleFav = () => {
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  // 收藏接口未就绪：不做本地假成功，保持原状态并如实提示
  uni.showToast({ title: '收藏功能即将开放', icon: 'none', duration: 1200 })
}
const onShare = () => {
  uni.showToast({ title: '分享功能待开放', icon: 'none', duration: 1200 })
}
const goSignup = () => {
  if (d.value.status === 'end') return
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  uni.navigateTo({ url: '/pkg-eco/pages/activities/signup?id=' + encodeURIComponent(d.value.id) })
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

<style>
page { background: #F4F6F8; }
</style>
<style scoped>
.actd-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== Hero ===== */
.hero { position: relative; height: 216px; overflow: hidden; }
.hero-glow { position: absolute; inset: 0; }
.hero-glow::after { content: ''; position: absolute; top: -30%; right: -10%; width: 220px; height: 220px; border-radius: 50%; background: radial-gradient(circle, rgba(255,255,255,.12) 0%, transparent 70%); }
.hero-ic { position: absolute; right: 28px; bottom: 36px; font-size: 88px; font-weight: 700; color: rgba(255,255,255,.92); text-shadow: 0 6px 18px rgba(0,0,0,.22); animation: float 3.2s ease-in-out infinite; }
@keyframes float { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-7px); } }
.hero-foot { position: absolute; left: 16px; right: 16px; bottom: 14px; z-index: 3; }
.hero-badge { display: inline-block; font-size: 11px; padding: 3px 10px; border-radius: 4px; font-weight: 600; background: rgba(255,255,255,.92); color: #0A66C2; margin-bottom: 8px; }
.hero-title { font-size: 19px; font-weight: 700; color: #fff; line-height: 1.35; text-shadow: 0 2px 8px rgba(0,0,0,.25); }

/* ===== 信息卡 ===== */
.info-card { position: relative; z-index: 5; margin: -34px 12px 0; background: #fff; border: 1px solid #EEF1F4; border-radius: 10px; padding: 16px 16px 6px; box-shadow: 0 4px 16px rgba(0,0,0,.05); }
.tag-row { display: flex; gap: 6px; margin-bottom: 14px; }
.tag { font-size: 10px; padding: 2px 8px; border-radius: 4px; font-weight: 500; color: #0A66C2; background: #EAF3FB; }
.stat-row { display: flex; padding: 10px 0 12px; border-top: 1px solid #EBEDF0; }
.si { flex: 1; text-align: center; position: relative; }
.si + .si::before { content: ''; position: absolute; left: 0; top: 6px; bottom: 6px; width: .5px; background: #F0F0F0; }
.sv { font-size: 17px; font-weight: 700; color: #17212B; display: block; }
.sl { font-size: 11px; color: #98A2B3; margin-top: 2px; display: block; }
.info-row { display: flex; align-items: flex-start; gap: 10px; padding: 9px 0; border-bottom: .5px solid #F5F5F5; }
.info-row:last-child { border-bottom: none; }
.info-ic { width: 30px; height: 30px; border-radius: 8px; background: #EAF3FB; color: #0A66C2; display: flex; align-items: center; justify-content: center; flex: none; font-size: 13px; font-weight: 600; }
.info-ic.ic-orange { background: #FFF0E6; color: #E96012; }
.info-ic.ic-green { background: #E9F7F0; color: #168A55; }
.info-txt { flex: 1; min-width: 0; }
.info-label { font-size: 10.5px; color: #98A2B3; margin-bottom: 1px; display: block; }
.info-value { font-size: 13.5px; color: #17212B; font-weight: 500; line-height: 1.4; display: block; }
.info-value.cl-su { color: #168A55; }

/* ===== 区块 ===== */
.sec { margin: 10px 12px 0; padding: 14px 16px; background: #fff; border: 1px solid #EEF1F4; border-radius: 10px; }
.sh { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.sd { width: 4px; height: 16px; background: #0A66C2; border-radius: 2px; flex-shrink: 0; }
.sht { font-size: 15px; font-weight: 700; color: #17212B; }
.sb { font-size: 13.5px; color: #667085; line-height: 1.75; white-space: pre-wrap; display: block; }
.sb.dim { color: #98A2B3; }

/* ===== 议程 ===== */
.agenda-item { position: relative; padding: 0 0 14px 22px; }
.agenda-item::before { content: ''; position: absolute; left: 6px; top: 5px; width: 10px; height: 10px; border-radius: 50%; background: #0A66C2; border: 2px solid #fff; box-shadow: 0 0 0 2px #EAF3FB; }
.agenda-item::after { content: ''; position: absolute; left: 10px; top: 17px; bottom: -3px; width: 2px; background: #EBEDF0; }
.agenda-item:last-child::after { display: none; }
.at { font-size: 11px; color: #0A66C2; font-weight: 600; display: block; }
.att { font-size: 13.5px; color: #17212B; font-weight: 500; margin-top: 2px; display: block; }

/* ===== 底部操作栏 ===== */
.bb { position: fixed; left: 0; right: 0; bottom: 0; display: flex; align-items: center; padding: 10px 12px; padding-bottom: calc(10px + env(safe-area-inset-bottom)); background: #fff; border-top: .5px solid #F0F0F0; box-shadow: 0 -2px 12px rgba(0,0,0,.04); gap: 10px; z-index: 50; }
.bi { width: 42px; height: 42px; border-radius: 50%; background: #F4F6F8; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.bi.fv { color: #ff3b30; }
.bit { font-size: 20px; line-height: 1; }
.bp { flex: 1; height: 44px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 15px; font-weight: 600; display: flex; align-items: center; justify-content: center; box-shadow: 0 2px 8px rgba(10,102,194,.35); }
.bp.disabled { background: #98A2B3; box-shadow: none; }
.bo { height: 44px; border-radius: 8px; border: 1.5px solid #0A66C2; background: #fff; color: #0A66C2; font-size: 14px; font-weight: 600; padding: 0 16px; display: flex; align-items: center; flex-shrink: 0; }

/* ===== 骨架 ===== */
.skw { padding-top: 10px; }
.sk-h { height: 216px; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.sk-sec { margin: 12px; padding: 16px; background: #fff; border-radius: 10px; }
.sk-l { height: 28rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 120px 40rpx; min-height: 500px; }
.stb { padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; margin-top: 24rpx; }
</style>
