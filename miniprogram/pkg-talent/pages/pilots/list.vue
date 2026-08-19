<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="认证飞手" show-back :fixed="true" @back="goBack" />

    <!-- 右下角浮动申请按钮（避开微信胶囊，单次弹性入场） -->
    <view class="apply-fab" hover-class="apply-fab-hover" :hover-stay-time="80" @tap="applyPilot">
      <view class="fab-icon"><text class="fab-icon-char">飞</text></view>
      <text class="fab-text">{{ applyText }}</text>
    </view>

    <!-- ① 统计横幅（真实聚合） -->
    <view class="stats-banner">
      <view class="stat-cell">
        <view class="stat-icon"><view class="icon-cert" /></view>
        <text class="stat-num" :class="{ 'stat-anim': loaded }">{{ displayStats.totalCerts }}</text>
        <text class="stat-label">证书</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-cell">
        <view class="stat-icon"><view class="icon-plane" /></view>
        <text class="stat-num" :class="{ 'stat-anim': loaded }">{{ displayStats.totalHours }}</text>
        <text class="stat-label">飞行小时</text>
      </view>
      <template v-if="displayStats.hasRating">
        <view class="stat-divider" />
        <view class="stat-cell">
          <view class="stat-icon"><view class="star-shape" /></view>
          <text class="stat-num" :class="{ 'stat-anim': loaded }">{{ displayStats.avgRating }}</text>
          <text class="stat-label">平均评分</text>
        </view>
      </template>
    </view>

    <!-- ② 搜索框（白上白：描边 + 双层投影；CSS 画放大镜；右侧"搜索"文字按钮） -->
    <view class="sbar">
      <view class="b-search">
        <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
        <input
          class="b-sinp"
          v-model="searchText"
          placeholder="搜索认证飞手姓名/编号"
          placeholder-class="b-ph"
          confirm-type="search"
          @confirm="onSearch"
        />
        <text v-if="searchText" class="b-sclr" @tap="clearSearch">×</text>
        <view class="b-sep" />
        <text class="b-sbtn" @tap="onSearch">搜索</text>
      </view>
    </view>

    <!-- ③ 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 位飞手</text>
      <text class="ir-hint">{{ searchText ? '搜索结果' : '协会认证' }}</text>
    </view>

    <!-- ④ 骨架屏：首次加载 -->
    <view v-if="loading && !list.length" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w60"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- ⑤ 错误态 -->
    <view v-else-if="!loading && errorMsg" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑥ 空态 -->
    <view v-else-if="!loading && !list.length" class="st">
      <u-empty description="暂无认证飞手">
        <text class="sth">成为协会认证飞手，即可展示在此名录</text>
        <view class="stb" @tap="applyPilot">申请认证</view>
      </u-empty>
    </view>

    <!-- ⑦ 飞手卡片列表 -->
    <view v-else class="cl">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="goDetail(item)"
      >
        <!-- 卡片头部：头像 + 认证徽章 + 名字/编号 + 评分 -->
        <view class="card-head">
          <view class="avatar-wrap">
            <image
              v-if="item.avatar"
              :src="item.avatar"
              mode="aspectFill"
              class="avatar"
              lazy-load
            />
            <view v-else class="avatar avatar-fallback" :style="{ background: avatarBg(item.real_name) }">
              <text class="avatar-char">{{ firstChar(item.real_name) }}</text>
            </view>
            <view class="cert-badge" />
          </view>
          <view class="head-main">
            <view class="name-row">
              <text class="name">{{ item.real_name || '认证飞手' }}</text>
              <text class="pilot-id">{{ idLabel(item) }}</text>
            </view>
            <view class="sub-row">
              <text class="cert-count">{{ (item.cert_ids || []).length }} 项认证</text>
            </view>
          </view>
          <view class="rating-wrap">
            <view class="star"><view class="star-shape" /></view>
            <text class="rating-num">{{ ratingText(item) }}</text>
            <text class="rating-sub">{{ ratingSub(item) }}</text>
          </view>
        </view>

        <!-- 数据行（两行四列） -->
        <view class="data-grid">
          <view class="data-item">
            <view class="data-icon data-icon-blue"><view class="icon-cert" /></view>
            <view class="data-body">
              <text class="data-label">证书</text>
              <text class="data-value">{{ (item.cert_ids || []).length }}</text>
            </view>
          </view>
          <view class="data-item">
            <view class="data-icon data-icon-purple"><view class="icon-plane" /></view>
            <view class="data-body">
              <text class="data-label">飞行</text>
              <text class="data-value">{{ item.flight_hours || 0 }} 小时</text>
            </view>
          </view>
          <view class="data-item">
            <view class="data-icon data-icon-green"><view class="icon-check" /></view>
            <view class="data-body">
              <text class="data-label">作业</text>
              <text class="data-value">{{ item.completed_jobs || 0 }}</text>
            </view>
          </view>
          <view class="data-item">
            <view class="data-icon data-icon-orange"><view class="icon-target" /></view>
            <view class="data-body">
              <text class="data-label">擅长</text>
              <text class="data-value ellipsis">{{ mainSkill(item) }}</text>
            </view>
          </view>
        </view>

        <!-- 作业类型标签（扁平 tint） -->
        <view v-if="jobTags(item).length > 0" class="tag-row">
          <text
            v-for="(t, ti) in shownTags(item)"
            :key="ti"
            class="job-tag"
            :style="tagStyle(t)"
          >{{ t }}</text>
          <text v-if="jobTags(item).length > 3" class="more-tag">+{{ jobTags(item).length - 3 }}</text>
        </view>

        <!-- 底部：查看飞手档案（整行可点击） -->
        <view class="card-footer" hover-class="footer-hover" :hover-stay-time="100" @tap.stop="goDetail(item)">
          <text class="card-hint">查看飞手档案</text>
          <view class="footer-chev" />
        </view>
      </view>

      <!-- 列表底部 -->
      <view class="list-footer">
        <text class="footer-line" />
        <text class="footer-text">没有更多了</text>
        <text class="footer-line" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'

const searchText = ref('')
const list = ref([])
const loading = ref(false)
const loaded = ref(false)
const errorMsg = ref('')
const statusBarHeight = ref(20)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画全关
const goBack = () => uni.navigateBack()

// 右上按钮文案：随我的认证状态变化，入口永远存在
const applyText = ref('申请认证')
const refreshMineLabel = async () => {
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = res && res.data ? res.data : res
    if (mine && mine.id) {
      applyText.value = { pending: '审核中', approved: '我的档案', rejected: '重新申请' }[mine.status] || '申请认证'
    } else {
      applyText.value = '申请认证'
    }
  } catch (e) { applyText.value = '申请认证' }
}

// ── 统计数字滚动计数 ─────────────────────
const displayStats = ref({ totalCerts: 0, totalHours: 0, avgRating: '—', hasRating: false })
const countUp = (target, duration = 800) => {
  const from = { totalCerts: 0, totalHours: 0 }
  const to = { totalCerts: target.totalCerts, totalHours: target.totalHours }
  const start = Date.now()
  // 微信小程序无 requestAnimationFrame，用 setTimeout 模拟帧（16ms）
  const tick = () => {
    const p = Math.min(1, (Date.now() - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3) // easeOutCubic
    displayStats.value = {
      totalCerts: Math.round(from.totalCerts + (to.totalCerts - from.totalCerts) * ease),
      totalHours: Math.round(from.totalHours + (to.totalHours - from.totalHours) * ease),
      avgRating: target.avgRating,
      hasRating: target.hasRating,
    }
    if (p < 1) setTimeout(tick, 16)
  }
  tick()
}

// ── 统计横幅（真实聚合）──────────────────────
const stats = computed(() => {
  const totalCerts = list.value.reduce((s, p) => s + (p.cert_ids || []).length, 0)
  const totalHours = list.value.reduce((s, p) => s + (p.flight_hours || 0), 0)
  const rated = list.value.filter((p) => p.rating > 0)
  const avg = rated.length ? (rated.reduce((s, p) => s + p.rating, 0) / rated.length) : 0
  const hasRating = avg > 0
  return {
    totalCerts,
    totalHours,
    hasRating,
    // 无真实评分时不展示评分项（横幅由 hasRating 控制显隐）
    avgRating: hasRating ? avg.toFixed(1) : '—',
  }
})

// ── 头像兜底：姓名首字 + 姓名哈希选渐变（每人不同）──
const AVATAR_GRADIENTS = [
  'linear-gradient(135deg,#0A1F44,#1E5EFF)',
  'linear-gradient(135deg,#6D28D9,#DB2777)',
  'linear-gradient(135deg,#0EA5E9,#06B6D4)',
  'linear-gradient(135deg,#FF8E3C,#F97316)',
  'linear-gradient(135deg,#00C896,#34c759)',
]
const firstChar = (name) => String(name || '飞').charAt(0)
const avatarBg = (name) => {
  const n = String(name || '')
  if (!n) return AVATAR_GRADIENTS[0]
  let h = 0
  for (let i = 0; i < n.length; i++) h = (h * 31 + n.charCodeAt(i)) >>> 0
  return AVATAR_GRADIENTS[h % AVATAR_GRADIENTS.length]
}

// ── 卡片字段映射 ──────────────────────────
// 编号：后端无编号字段，展示"协会认证 · N 项证书"
const idLabel = (item) => {
  const n = (item.cert_ids || []).length
  return `协会认证 · ${n} 项证书`
}
// 评分：rating>0 显示实际评分；无评分显示"—"（不伪造满分）
const ratingText = (item) => (item.rating > 0 ? item.rating.toFixed(1) : '—')
// 评价数：≥10 才显示"XX 人评价"，否则显示"暂无评价"（弱化）
const ratingSub = (item) => {
  const n = item.completed_jobs || 0
  return n >= 10 ? `${n} 人评价` : '暂无评价'
}
// 擅长（替代地址）：bio 拆出的第一个标签
const bioList = (bio) => String(bio || '').split(/[/，,、\s]+/).filter(Boolean)
const mainSkill = (item) => bioList(item.bio)[0] || '综合'

// ── 作业类型标签（扁平 tint：浅底深字，无描边）──────────────
const JOB_TAG_MAP = [
  { key: ['电力巡检', '巡检'], color: '#0A66C2', bg: '#EAF3FB' },
  { key: ['测绘', '航拍', '拍摄'], color: '#6941C6', bg: '#F0E9F7' },
  { key: ['植保', '喷洒'], color: '#0B6B41', bg: '#E9F7F0' },
  { key: ['应急', '救援', '侦察'], color: '#B42318', bg: '#FDECEC' },
  { key: ['吊运', '吊装', '实操'], color: '#C2410C', bg: '#FFF4EC' },
  { key: ['物流', '运输', '投送'], color: '#0E7090', bg: '#E6F4F7' },
  { key: ['航拍', '宣传'], color: '#C11574', bg: '#FCE7F3' },
]
const matchTag = (tag) => {
  for (const m of JOB_TAG_MAP) {
    if (m.key.some((k) => tag.includes(k))) return m
  }
  return { color: '#5D6B82', bg: '#EEF1F4' }
}
const jobTags = (item) => bioList(item.bio)
const shownTags = (item) => jobTags(item).slice(0, 3)
const tagStyle = (t) => {
  const m = matchTag(t)
  return { color: m.color, background: m.bg }
}

// ── 数据加载 ────────────────────────────
const goDetail = (item) => {
  uni.setStorageSync('pilot_detail', item)
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(item.id) })
}
const onSearch = () => fetchData()
const clearSearch = () => { searchText.value = ''; fetchData() }

const fetchData = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const kw = searchText.value.trim()
    const res = await request({ url: '/api/v1/certified-pilots', data: { page: 1, page_size: 100, keyword: kw } })
    list.value = (Array.isArray(res) ? res : (res.data || []))
  } catch {
    list.value = []
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
    loaded.value = true
    // 统计数字滚动动画（数据就绪后触发）
    countUp(stats.value)
  }
}

// ---- 申请认证 / 我的状态 ----
const applyPilot = async () => {
  if (!requireLogin()) return
  let mine = null
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    mine = res && res.data ? res.data : res
  } catch (e) {}
  if (mine && mine.id) {
    const label = { pending: '待审核', approved: '已认证', rejected: '未通过' }[mine.status] || mine.status
    // 已认证 → 查看我的档案（detail 页按 id 回源）；驳回 → 重新提交（后端支持覆盖重提）；待审核 → 仅提示
    if (mine.status === 'approved') {
      uni.showModal({
        title: '我的飞手认证',
        content: `当前状态：${label}，已展示在认证飞手名录中`,
        confirmText: '查看档案',
        success: (r) => {
          if (r.confirm) {
            uni.removeStorageSync('pilot_detail')
            uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(mine.id) })
          }
        },
      })
      return
    }
    if (mine.status === 'rejected') {
      uni.showModal({
        title: '我的飞手认证',
        content: `当前状态：${label}，可修改资料后重新提交\n${mine.real_name || ''}`,
        confirmText: '重新提交',
        success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' }) },
      })
      return
    }
    uni.showModal({ title: '我的飞手认证', content: `当前状态：${label}\n${mine.real_name || ''}`, showCancel: false, confirmText: '知道了' })
    return
  }
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    if (sys && sys.statusBarHeight) statusBarHeight.value = sys.statusBarHeight
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  fetchData()
  refreshMineLabel()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 80px; /* 给右下角浮动按钮留空间 */
}

/* ═══ 右下角浮动申请按钮 ═══ */
.apply-fab {
  position: fixed;
  right: 16px;
  bottom: calc(24px + env(safe-area-inset-bottom));
  z-index: 60;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px 7px 7px;
  background: #0A66C2;
  border-radius: 999px;
  box-shadow: 0 6px 18px rgba(10, 102, 194, 0.28);
  animation: fab-in 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.apply-fab-hover {
  transform: scale(0.94);
  opacity: 0.92;
}
.fab-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}
.fab-icon-char {
  font-size: 14px;
  font-weight: 700;
  color: #fff;
}
.fab-text {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}
/* 单次弹性入场：上浮 + 缩放回弹（无循环装饰动画） */
@keyframes fab-in {
  from { opacity: 0; transform: translateY(30px) scale(0.6); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

/* ═══ ① 统计横幅 ═══ */
.stats-banner {
  display: flex;
  align-items: center;
  margin: 10px 12px 2px;
  background: #F4F8FC;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  padding: 14px 10px;
}
.stat-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.stat-icon {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-num {
  font-size: 20px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.2;
}
.stat-anim {
  animation: statPop 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}
.stat-label {
  font-size: 11px;
  color: #667085;
}
.stat-divider {
  width: 1px;
  height: 36px;
  background: #E4E7EC;
}

/* ═══ ② 搜索框：白上白——纯白填充 + 灰描边 + 双层投影 ═══ */
.sbar { padding: 8px 12px 6px; background: #fff; }
.b-search {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05);
  display: flex;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.b-search-ic { position: relative; width: 15px; height: 15px; flex: none; }
.ic-ring {
  width: 9px; height: 9px;
  border: 1.5px solid #98A2B3;
  border-radius: 50%;
  position: absolute; top: 0; left: 0;
}
.ic-bar {
  position: absolute; right: 0; bottom: 1px;
  width: 5px; height: 1.5px;
  background: #98A2B3;
  transform: rotate(45deg);
}
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; }
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ═══ ③ 信息行 ═══ */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
  animation: fadeUp 0.25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { font-size: 12px; color: #98A2B3; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ═══ ④ 骨架屏 ═══ */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 4px 12px 12px; }
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
.sk-tag { width: 40px; height: 40px; border-radius: 8px; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.55; } }

/* ═══ ⑤⑥ 空 / 错误 ═══ */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ═══ ⑦ 飞手卡片 ═══ */
.cl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px 12px; }
.card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  transition: transform 0.35s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.15s ease;
}
.card:nth-child(-n+6) { animation: cardIn 0.22s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.tap-scale { transform: scale(0.97); opacity: 0.9; }

/* 7.1 卡片头部 */
.card-head {
  display: flex;
  align-items: center;
  gap: 10px;
}
.avatar-wrap {
  position: relative;
  flex-shrink: 0;
}
.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
}
.avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-char {
  font-size: 18px;
  font-weight: 700;
  color: #ffffff;
}
.cert-badge {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #00C896;
  border: 2px solid #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-badge::after {
  content: '';
  position: absolute;
  left: 4.5px;
  top: 3.5px;
  width: 5px;
  height: 8px;
  border: solid #fff;
  border-width: 0 1.5px 1.5px 0;
  transform: rotate(45deg);
}
.head-main {
  flex: 1;
  min-width: 0;
}
.name-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.name {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.pilot-id {
  font-size: 11px;
  color: #98A2B3;
  flex-shrink: 0;
}
.sub-row {
  margin-top: 4px;
}
.cert-count {
  font-size: 11px;
  color: #667085;
}
.rating-wrap {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  flex-shrink: 0;
}
.star {
  display: flex;
  align-items: center;
  justify-content: center;
}
.rating-num {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  margin-top: 2px;
}
.rating-sub {
  font-size: 10px;
  color: #98A2B3;
  margin-top: 2px;
}

/* 7.2 数据行 */
.data-grid {
  display: flex;
  flex-wrap: wrap;
  background: #F7F8FA;
  border-radius: 8px;
  padding: 8px 10px;
  gap: 4px 0;
}
.data-item {
  width: 50%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 4px;
  box-sizing: border-box;
}
.data-icon {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.data-icon-blue { background: #EAF3FB; }
.data-icon-purple { background: #F0E9F7; }
.data-icon-green { background: #E9F7F0; }
.data-icon-orange { background: #FFF4EC; }
.data-body {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.data-label {
  font-size: 10px;
  color: #667085;
}
.data-value {
  font-size: 13px;
  font-weight: 700;
  color: #17212B;
}
.data-value.ellipsis {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 90px;
}

/* ═══ CSS 绘制的符号图标（禁 emoji 规范）═══ */

/* 证书：圆角方块 + 中缝 */
.icon-cert {
  width: 8px;
  height: 10px;
  border-radius: 2px;
  background: #0A66C2;
  position: relative;
}
.icon-cert::after {
  content: '';
  position: absolute;
  left: 1.5px;
  right: 1.5px;
  top: 4px;
  height: 1px;
  background: #EAF3FB;
  box-shadow: 0 2px 0 #EAF3FB;
}

/* 纸飞机：clip-path 三角翼 + 尾翼（非 emoji） */
.icon-plane {
  width: 9px;
  height: 5px;
  background: #6941C6;
  clip-path: polygon(100% 0, 0 50%, 100% 100%);
  position: relative;
  transform: rotate(-8deg);
}
.icon-plane::after {
  content: '';
  position: absolute;
  right: -2px;
  top: 2px;
  width: 4px;
  height: 4px;
  background: #6941C6;
  clip-path: polygon(50% 0, 100% 100%, 0 100%);
}

/* 对勾：作业完成 */
.icon-check {
  width: 7px;
  height: 12px;
  border: solid #0B6B41;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  margin-top: -3px;
}

/* 目标：圆环 + 中心点 */
.icon-target {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid #C2410C;
  position: relative;
}
.icon-target::after {
  content: '';
  position: absolute;
  left: 2px;
  top: 2px;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: #C2410C;
}

/* 星星：clip-path 五角（评分/统计横幅共用） */
.star-shape {
  width: 11px;
  height: 11px;
  background: #F79009;
  clip-path: polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%);
}

/* 7.3 作业标签（扁平 tint：浅底深字，无描边无动画） */
.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 2px;
}
.job-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  border: none;
  max-width: 140px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.more-tag {
  font-size: 11px;
  font-weight: 600;
  color: #5D6B82;
  background: #EEF1F4;
  padding: 2px 8px;
  border-radius: 4px;
}

/* 7.4 卡片底部：整行可点击 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  margin-top: 2px;
  padding-top: 10px;
  border-top: 1px solid #F0F1F3;
}
.footer-hover {
  opacity: 0.7;
}
.card-hint {
  font-size: 13px;
  color: #0A66C2;
  font-weight: 600;
}
.footer-chev {
  width: 6px;
  height: 6px;
  border-top: 1.5px solid #0A66C2;
  border-right: 1.5px solid #0A66C2;
  transform: rotate(45deg);
}

/* ═══ 列表底部 ═══ */
.list-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px 0 8px;
}
.footer-line {
  width: 48px;
  height: 1px;
  background: #EDF0F3;
}
.footer-text {
  font-size: 12px;
  color: #98A2B3;
}

/* ═══ 微动效（单次入场，无循环装饰动画）═══ */
@keyframes statPop {
  from {
    transform: scale(0.8);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

/* ═══ 减弱动效（无障碍） ═══ */
.page.no-motion .card,
.page.no-motion .ir,
.page.no-motion .stats-banner,
.page.no-motion .stat-num { animation: none; }
.page.no-motion .sk-tag,
.page.no-motion .sk-l { animation: none; }
.page.no-motion .apply-fab { animation: none; }
</style>
