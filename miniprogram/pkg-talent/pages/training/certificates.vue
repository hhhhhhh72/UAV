<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="我的证书" show-back :fixed="true" @back="goBack" />

    <!-- Banner -->
    <view class="banner">
      <view class="banner-icon">证</view>
      <view class="banner-info">
        <text class="banner-title">我的证书</text>
        <text class="banner-sub">协会认证培训考核 · 证书图片随时预览</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 -->
      <view class="ir">
        <text>共 <text class="irn">{{ list.length }}</text> 项证书</text>
      </view>

      <!-- 骨架 -->
      <view v-if="loading" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w80"></view>
            <view class="sk-l w60"></view>
          </view>
        </view>
      </view>

      <!-- 错误 -->
      <view v-else-if="errorMsg && !list.length" class="st">
        <u-empty :description="errorMsg">
          <view class="stb" @tap="fetchList">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="!list.length" class="st">
        <u-empty description="暂无证书">
          <text class="sth">证书为线下考核后由协会颁发，完成培训课程后请联系管理员</text>
          <view class="stb" @tap="goCourses">去逛逛培训课程</view>
        </u-empty>
      </view>

      <!-- 列表：状态徽章 + 标题 + 描述 + 元信息 + 证书缩略图 -->
      <view v-else class="cl">
        <view
          v-for="item in list"
          :key="item.id"
          class="card"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="viewCert(item)"
        >
          <view class="c-main">
            <view class="c-badges">
              <text class="c-tag" :style="typeStyle(item.cert_type)">{{ typeLabel(item.cert_type) }}</text>
              <text class="c-st" :class="statusCls(item.status)">{{ statusLabel(item.status) }}</text>
            </view>
            <text class="ct">{{ typeFull(item.cert_type) }}</text>
            <text v-if="item.cert_number" class="c-desc">编号：{{ item.cert_number }}</text>
            <view class="c-meta">
              <text v-if="item.issue_date">发证 {{ dateText(item.issue_date) }}</text>
              <text v-if="item.issue_date && (item.expire_date || item.expiry_date)" class="c-dot">·</text>
              <text v-if="item.expire_date || item.expiry_date">至 {{ dateText(item.expire_date || item.expiry_date) }}</text>
            </view>
          </view>
          <image v-if="certImage(item)" class="c-thumb" :src="certImage(item)" mode="aspectFill" />
        </view>
      </view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onPageScroll } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'

/* 证书类型 → 徽章短名（兼容后端枚举 caac/utc_dji/gov_level 与展示形 CAAC/UTC/人社） */
const CERT_TYPE_SHORT = {
  caac: 'CAAC', utc_dji: 'UTC', gov_level: '人社',
  aopa: 'AOPA', asfc: 'ASFC',
}
/* 证书类型 → 标题全名（沿用 pilots/detail、register 的产品措辞；未知类型回退短名） */
const CERT_TYPE_FULL = {
  caac: 'CAAC 执照', utc_dji: '大疆 UTC 认证', gov_level: '人社等级证书',
}
/* 证书类型 → tag 配色（深色字 + 浅底，与研发难题广场领域标签同构） */
const CERT_TYPE_STYLE = {
  caac: { color: '#0d47a1', bg: '#E3EDF9' },
  utc_dji: { color: '#4a148c', bg: '#F0E9F7' },
  gov_level: { color: '#B54708', bg: '#FDEEE4' },
  aopa: { color: '#004d40', bg: '#E4F2EF' },
  asfc: { color: '#1a237e', bg: '#E7E9F4' },
}
const CERT_TYPE_STYLE_DEFAULT = { color: '#344054', bg: '#EEF1F4' }

/* 状态 → 文案 / 状态色（对齐挑战广场语义：有效/通过=绿、审核中=蓝、过期/驳回=灰、吊销=红） */
const STATUS_LABEL = {
  active: '有效', expired: '已过期', pending: '审核中', revoked: '已吊销',
  approved: '已通过', rejected: '已驳回',
}
const STATUS_CLS = {
  active: 'st-open', approved: 'st-open',
  pending: 'st-pending',
  expired: 'st-closed', rejected: 'st-closed',
  revoked: 'st-err',
}

const loading = ref(false)
const errorMsg = ref('')
const list = ref([])
const statusBarHeight = ref(20)
const showBt = ref(false)
const { noMotion, checkMotion } = useReduceMotion()

const norm = (t) => String(t || '').toLowerCase()
const typeLabel = (t) => CERT_TYPE_SHORT[norm(t)] || t || '通用'
const typeFull = (t) => CERT_TYPE_FULL[norm(t)] || CERT_TYPE_SHORT[norm(t)] || t || '通用证书'
const typeStyle = (t) => CERT_TYPE_STYLE[norm(t)] || CERT_TYPE_STYLE_DEFAULT
const statusLabel = (s) => STATUS_LABEL[s] || s || '未知'
const statusCls = (s) => STATUS_CLS[s] || 'st-closed'
const dateText = (iso) => (iso ? String(iso).slice(0, 10) : '—')

/* 证书图：兼容 image_url / certificate_url / image / certificate 四类字段 */
const certImage = (item) => item && (item.image_url || item.certificate_url || item.image || item.certificate)

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/certificates/mine' })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    list.value = items
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function viewCert(item) {
  // 有证书图则全屏预览；无图如实提示
  const url = certImage(item)
  if (url) {
    uni.previewImage({ urls: [url], current: url })
  } else {
    uni.showToast({ title: '暂无证书图片', icon: 'none' })
  }
}

function goBack() { uni.navigateBack() }
function goCourses() { uni.navigateTo({ url: '/pkg-talent/pages/training/courses' }) }
function scrollToTop() { uni.pageScrollTo({ scrollTop: 0, duration: 300 }) }

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  if (!requireLogin()) return
  fetchList()
})

onPullDownRefresh(() => {
  fetchList().then(function () {
    uni.stopPullDownRefresh()
  })
})

onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
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

/* ===== Banner（对齐研发难题广场：蓝渐变 + 圆角图标 + 标题/副标题） ===== */
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
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: block; }

/* ===== 白色板块 ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }

/* ===== 列表卡片（白上白：灰描边 + 极淡灰投影浮起；无左缘色条） ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 12px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}
.tap-scale { transform: scale(0.95); opacity: 0.9; }
.c-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 7px;
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
.c-st.st-open { color: #0B6B41; background: #E9F7F0; }
.c-st.st-pending { color: #0A66C2; background: #EAF3FB; }
.c-st.st-closed { color: #5D6B82; background: #EEF1F4; }
.c-st.st-err { color: #B42318; background: #FDECEC; }
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
  flex-wrap: wrap;
}
.c-dot { color: #DDE1E6; }
/* 证书缩略图：圆角小图，点击整卡预览原图 */
.c-thumb {
  flex: none;
  width: 60px;
  height: 60px;
  border-radius: 8px;
  background: #F4F6F8;
}

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
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; text-align: center; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 回到顶部 ===== */
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
  z-index: 35;
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(0.16, 1, 0.3, 1);
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; }

/* ===================== 动效规范（对齐研发难题广场） =====================
   白名单：仅 transform / opacity（小尺寸颜色过渡允许）
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1) + ios-decel cubic-bezier(.32,.72,0,1)
   数量：列表入场仅错峰首屏 6 项，其余静置
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 列表入场：前 6 项每 20ms 依次淡入上移（backwards 填充 → 延迟期不闪跳） */
.card { animation: none; }
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* Banner 内部微编排：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms，总 340ms ≤ 400ms */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; }
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }

/* Banner 单次扫光（非循环装饰：100ms 起播 280ms 线性，380ms 内收完） */
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

/* 2) 交互反馈：卡片按压（快进慢出） */
.card { transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.card.tap-scale { transition-duration: .1s; transition-timing-function: linear; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }

/* 骨架呼吸（加载中环境光；循环动画 1.4s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */
.page.no-motion .card,
.page.no-motion .banner,
.page.no-motion .ir { animation: none; }
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner::before,
.page.no-motion .banner::after { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .tap-scale { transform: none !important; }
.page.no-motion .stb:active,
.page.no-motion .bt:active { transform: none; }
</style>
