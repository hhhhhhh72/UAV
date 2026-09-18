<template>
  <view class="ent-detail-page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <!-- 统一导航（与课题/活动/难题详情同款 u-nav-bar：状态栏避让 + 44px 栏高 + 标题 32rpx） -->
    <u-nav-bar title="企业详情" show-back :fixed="true" @back="goBack" />

    <!-- ① Hero：封面 + **统一半透明纯色遮罩** + 身份区。
         规范要求图片叠字只用纯色遮罩、禁止装饰性渐变（visual-foundations.md:68），
         所以这里不再用三层渐变，改成品牌深蓝 #074D92 的单层压暗。
         身份信息（Logo/名称/地址）只在这里出现一次，避免与下方区块重复。 -->
    <view class="hero">
      <image
        v-if="ent.cover_image && !heroImgError"
        :src="resolveUrl(ent.cover_image)"
        mode="aspectFill"
        class="hero-img"
        lazy-load
        @error="heroImgError = true"
      />
      <view v-else class="hero-fallback">
        <text class="hero-fallback-char">{{ ent.name ? ent.name.charAt(0) : '企' }}</text>
      </view>
      <view class="hero-scrim" />
      <view class="hero-bottom">
        <view class="hero-logo">
          <image v-if="ent.logo" :src="resolveUrl(ent.logo)" mode="aspectFill" class="hero-logo-img" @error="ent.logo = ''" />
          <text v-else class="hero-logo-char">{{ ent.name ? ent.name.charAt(0) : '企' }}</text>
        </view>
        <view class="hero-id">
          <text class="hero-title">{{ ent.name || '企业' }}</text>
          <view class="hero-chips">
            <text class="chip chip-verified">协会认证</text>
            <text v-if="ent.is_member" class="chip chip-member">协会会员</text>
          </view>
          <text class="hero-addr">{{ ent.address || '地区待公开' }}</text>
        </view>
      </view>
    </view>

    <!-- 加载骨架 -->
    <view v-if="loading" class="skeleton-wrap">
      <view class="skeleton-card"></view>
      <view class="skeleton-card"></view>
    </view>

    <!-- 404 空态 -->
    <view v-else-if="notFound" class="state-panel">
      <view class="state-mark">
        <view class="state-building">
          <view class="state-win state-win-1" />
          <view class="state-win state-win-2" />
        </view>
      </view>
      <text class="state-title">企业不存在或暂未公开</text>
      <text class="state-desc">仅完成入驻审核的企业在此公示</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="goBack">
        <text>返回列表</text>
      </view>
    </view>

    <!-- 错误态 -->
    <view v-else-if="err" class="state-panel">
      <view class="state-mark">
        <view class="state-building">
          <view class="state-win state-win-1" />
          <view class="state-win state-win-2" />
        </view>
      </view>
      <text class="state-title">加载失败</text>
      <text class="state-desc">请检查网络后重试</text>
      <view class="state-btn" hover-class="tap-fade" hover-stay-time="120" @tap="loadDetail">
        <text>重新加载</text>
      </view>
    </view>

    <!-- ② 内容区：**一整块白色业务表面 + 分隔线**，不给每个 section 加阴影
         （visual-foundations.md:59）。此前是 5 张各自带阴影的浮起卡片。 -->
    <template v-else>
      <view class="surface">
        <!-- 企业信息前置：详情页先给决策信息（SKILL.md:50），
             此前规模/成立/入驻被压在页面最底部的「基本信息」里。 -->
        <view class="sec sec-first">
          <view class="sec-head"><view class="sec-bar" /><text class="sec-title">企业信息</text></view>
          <view class="fact-grid">
            <view class="fact">
              <text class="fact-label">企业规模</text>
              <text class="fact-value">{{ ent.scale || '—' }}</text>
            </view>
            <view class="fact">
              <text class="fact-label">成立时间</text>
              <text class="fact-value">{{ ent.founded_at || '—' }}</text>
            </view>
            <view class="fact">
              <text class="fact-label">营业时间</text>
              <text class="fact-value">{{ ent.business_hours || '—' }}</text>
            </view>
            <view class="fact">
              <text class="fact-label">入驻时间</text>
              <text class="fact-value">{{ formatDate(ent.created_at) }}</text>
            </view>
          </view>
        </view>

        <view class="sec">
          <view class="sec-head"><view class="sec-bar" /><text class="sec-title">行业分类</text></view>
          <view class="tag-wrap">
            <text v-for="c in categoryList(ent)" :key="c" class="tag tag-blue">{{ c }}</text>
            <text v-if="categoryList(ent).length === 0" class="tag-empty">未填写</text>
          </view>
        </view>

        <view class="sec">
          <view class="sec-head"><view class="sec-bar" /><text class="sec-title">企业简介</text></view>
          <text class="desc-text">{{ ent.description || '该企业暂未填写简介' }}</text>
        </view>

        <view v-if="tagList(ent).length" class="sec sec-last">
          <view class="sec-head"><view class="sec-bar" /><text class="sec-title">核心能力</text></view>
          <view class="tag-wrap">
            <text v-for="t in tagList(ent)" :key="t" class="tag tag-plain">{{ t }}</text>
          </view>
        </view>
      </view>

      <text class="foot-note">信息由企业提交，经协会审核后公示</text>

      <!-- 底部操作栏：收藏 + 分享 + 主操作（主按钮用品牌蓝，橙只留给发布/预算/价值） -->
      <view class="bb-space"></view>
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" @tap="toggleFav">
          <view class="heart" />
        </view>
        <button class="bo" open-type="share" hover-class="bo-hover" hover-start-time="0" hover-stay-time="300" aria-label="转发">分享</button>
        <view class="bp" :class="{ disabled: !canContact }" @tap="onContact">联系企业</view>
      </view>
    </template>
  </view>
</template>
<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const loading = ref(true)
const notFound = ref(false)
const err = ref(false)
const ent = ref({})
const statusBarHeight = ref(20)
const heroImgError = ref(false)
const isFav = ref(false)
let detailId = ''

// 收藏（本地持久化，按企业 id 记忆）
const FAV_KEY = 'ent_favs'
const loadFavs = () => {
  try {
    const v = uni.getStorageSync(FAV_KEY)
    return Array.isArray(v) ? new Set(v) : new Set()
  } catch (e) { return new Set() }
}
const favs = loadFavs()
const saveFavs = () => {
  try { uni.setStorageSync(FAV_KEY, Array.from(favs)) } catch (e) { /* 忽略 */ }
}
const toggleFav = () => {
  if (!detailId) return
  if (favs.has(detailId)) { favs.delete(detailId); isFav.value = false }
  else { favs.add(detailId); isFav.value = true }
  saveFavs()
  uni.showToast({ title: isFav.value ? '已收藏' : '已取消收藏', icon: 'none' })
}

// 联系企业：公开响应里可能无电话（脱敏/未公开）——没有就给去向，不装可点
const canContact = computed(() => !!(ent.value && (ent.value.contact_phone || ent.value.contact || ent.value.phone)))
const onContact = () => {
  const phone = ent.value && (ent.value.contact_phone || ent.value.contact || ent.value.phone)
  if (phone) {
    uni.makePhoneCall({ phoneNumber: String(phone).replace(/\s/g, '') })
    return
  }
  uni.showModal({
    title: '联系方式暂未公开',
    content: '该企业暂未公开联系方式，可稍后重试，或通过协会秘书处联系。',
    confirmText: '知道了',
    showCancel: false,
  })
}

const goBack = () => {
  // 栈感知：有上级页面则返回，否则兜底回列表页
  const pages = getCurrentPages()
  if (pages.length > 1) uni.navigateBack()
  else uni.redirectTo({ url: '/pkg-eco/pages/enterprise/list' })
}

const splitTags = (str) => {
  if (!str) return []
  return String(str).split(',').map((t) => t.trim()).filter(Boolean)
}
// 相对路径（存库格式）→ 完整 URL（预览格式）
const resolveUrl = (u) => {
  if (!u) return ''
  if (u.indexOf('http') === 0) return u
  return BASE_URL + u
}
const categoryList = (e) => splitTags(e.industry_category)
const tagList = (e) => splitTags(e.capability_tags)

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

async function loadDetail() {
  loading.value = true
  notFound.value = false
  err.value = false
  try {
    const res = await request({ url: '/api/v1/enterprises/public/detail?id=' + encodeURIComponent(detailId) })
    ent.value = res || {}
    heroImgError.value = false
    isFav.value = favs.has(detailId)
  } catch (e) {
    // 404（未审核/不存在）走空态；网络/服务端错误走错误态 + 重试
    const status = (e && e.statusCode) || 0
    if (status === 404) notFound.value = true
    else err.value = true
  } finally {
    loading.value = false
  }
}

onShareAppMessage(() => ({
  title: ent.value ? '入驻企业：' + (ent.value.name || '') : '入驻企业',
  path: '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(detailId),
}))

onLoad(async (query) => {
  // 状态栏高度已在 setup 同步取得（与 u-nav-bar 组件一致，首帧即正确，不再闪跳）
  try {
    statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
  detailId = query && query.id
  if (!detailId) {
    notFound.value = true
    loading.value = false
    return
  }
  loadDetail()
})
</script>

<style scoped>
/* 全部数值取自 .claude/skills/design-uav-miniprogram-prototypes/references/tokens.json：
   圆角上限 8px(16rpx)、标签 4px(8rpx)、图标底与紧凑控件 6px(12rpx)；
   板块标题 17px(34rpx)、正文 11–13px、辅助 9–11px；页面左右边距 12px(24rpx)；
   卡片阴影只有可点击业务实体才用 0 3px 12px rgba(16,24,40,.05)。 */
.ent-detail-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(150rpx + env(safe-area-inset-bottom));
  overflow-x: hidden;
}

.tap-fade { opacity: 0.85; }

/* ═══════ ① Hero：封面 + 纯色遮罩 + 身份区 ═══════ */
.hero {
  position: relative;
  width: 100%;
  height: 420rpx;
  overflow: hidden;
  background: #074D92;
}
.hero-img { position: absolute; inset: 0; width: 100%; height: 100%; }
.hero-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #074D92;
}
.hero-fallback-char { font-size: 96rpx; font-weight: 760; color: rgba(255, 255, 255, 0.22); }
/* 统一半透明纯色遮罩：图片叠字只用纯色，不用装饰性渐变 */
.hero-scrim { position: absolute; inset: 0; background: rgba(7, 77, 146, 0.58); }
.hero-bottom {
  position: absolute;
  left: 24rpx;
  right: 24rpx;
  bottom: 36rpx;
  display: flex;
  align-items: flex-end;
  gap: 20rpx;
}
.hero-logo {
  width: 112rpx;
  height: 112rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  overflow: hidden;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}
.hero-logo-img { width: 100%; height: 100%; }
.hero-logo-char { font-size: 44rpx; font-weight: 760; color: #0A66C2; }
.hero-id { flex: 1; min-width: 0; }
.hero-title {
  font-size: 34rpx;
  font-weight: 760;
  color: #ffffff;
  line-height: 1.28;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.3);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero-chips { display: flex; align-items: center; gap: 8rpx; margin-top: 10rpx; }
.chip {
  font-size: 20rpx;
  font-weight: 600;
  line-height: 1.4;
  padding: 3rpx 10rpx;
  border-radius: 8rpx;
}
.chip-verified { background: #E9F7F0; color: #168A55; }
.chip-member { background: #EAF3FB; color: #0A66C2; }
.hero-addr {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.82);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ═══════ ② 内容：一整块白色表面 + 分隔线（不给每个 section 加阴影） ═══════ */
.surface {
  position: relative;
  margin-top: -20rpx;
  background: #fff;
  border-radius: 16rpx 16rpx 0 0;
  padding: 0 24rpx;
}
.sec { padding: 24rpx 0; border-bottom: 1rpx solid #EEF1F4; }
.sec-first { padding-top: 28rpx; }
.sec-last { border-bottom: none; }
.sec-head { display: flex; align-items: center; gap: 10rpx; margin-bottom: 16rpx; }
.sec-bar { width: 6rpx; height: 28rpx; border-radius: 3rpx; background: #0A66C2; }
.sec-title { font-size: 34rpx; font-weight: 760; color: #17212B; }

/* 关键事实 2×2：详情页先给决策信息 */
.fact-grid { display: flex; flex-wrap: wrap; }
.fact { width: 50%; padding: 10rpx 0; box-sizing: border-box; }
.fact-label { display: block; font-size: 22rpx; color: #98A2B3; line-height: 1.4; }
.fact-value {
  display: block;
  margin-top: 6rpx;
  font-size: 26rpx;
  font-weight: 600;
  color: #344054;
  line-height: 1.4;
  word-break: break-all;
}

.tag-wrap { display: flex; flex-wrap: wrap; gap: 12rpx; }
.tag { border-radius: 8rpx; padding: 6rpx 14rpx; font-size: 22rpx; line-height: 1.4; }
.tag-blue { color: #0A66C2; background: #EAF3FB; border: 1rpx solid #E4E7EC; }
.tag-plain { color: #667085; background: #F4F6F8; border: 1rpx solid #EEF1F4; }
.tag-empty { font-size: 22rpx; color: #98A2B3; }

.desc-text { display: block; font-size: 26rpx; color: #344054; line-height: 1.7; }

.foot-note {
  display: block;
  margin: 20rpx 0 0;
  padding: 0 24rpx;
  text-align: center;
  font-size: 20rpx;
  color: #98A2B3;
}

/* ═══════ 底部操作栏 ═══════ */
.bb-space { height: 150rpx; }
.bb {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1rpx solid #E4E7EC;
  box-shadow: 0 -5px 16px rgba(16, 24, 40, 0.04);
  z-index: 60;
}
/* 纯图标按钮：视觉 42px、触控 ≥40px */
.bi {
  width: 84rpx;
  height: 84rpx;
  border-radius: 12rpx;
  background: #fff;
  border: 1rpx solid #E4E7EC;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
/* 线性心形（2px 描边，非实心填充）：选中态用品牌蓝 —— 蓝色代表「选择」 */
.heart { position: relative; width: 36rpx; height: 32rpx; }
.heart::before,
.heart::after {
  content: '';
  position: absolute;
  top: 0;
  width: 18rpx;
  height: 28rpx;
  box-sizing: border-box;
  border: 4rpx solid #667085;
  border-bottom: none;
  border-radius: 9rpx 9rpx 0 0;
  background: transparent;
}
.heart::before { left: 18rpx; transform: rotate(-45deg); transform-origin: 0 100%; }
.heart::after { left: 0; transform: rotate(45deg); transform-origin: 100% 100%; }
.bi.fv .heart::before,
.bi.fv .heart::after { border-color: #0A66C2; }
.bo {
  height: 84rpx;
  border-radius: 14rpx;
  border: 2rpx solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  font-size: 26rpx;
  font-weight: 600;
  padding: 0 34rpx;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  margin: 0;
}
.bo::after { border: none; }
.bo-hover { opacity: 0.8; }
.bp {
  flex: 1;
  height: 84rpx;
  border-radius: 14rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 28rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.bp.disabled { background: #98A2B3; }

/* ═══════ 骨架屏（shimmer 仅作加载反馈） ═══════ */
.skeleton-wrap {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 24rpx;
  margin-top: -20rpx;
}
.skeleton-card {
  height: 200rpx;
  border-radius: 16rpx;
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 空态 / 错误态 ═══════ */
.state-panel {
  min-height: 640rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx 48rpx;
  text-align: center;
}
.state-mark {
  width: 120rpx;
  height: 120rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 16rpx;
  background: #EAF3FB;
}
/* CSS 楼宇图标（非 emoji），纯色不渐变 */
.state-building {
  width: 52rpx;
  height: 44rpx;
  position: relative;
  background: #0A66C2;
  border-radius: 4rpx 4rpx 2rpx 2rpx;
}
.state-building::after {
  content: '';
  position: absolute;
  left: 14rpx;
  bottom: -8rpx;
  width: 24rpx;
  height: 8rpx;
  background: #0A66C2;
  border-radius: 0 0 4rpx 4rpx;
}
.state-win {
  position: absolute;
  top: 12rpx;
  width: 8rpx;
  height: 12rpx;
  background: #fff;
  border-radius: 2rpx;
}
.state-win-1 { left: 12rpx; }
.state-win-2 { right: 12rpx; }
.state-title { font-size: 30rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; line-height: 1.5; }
.state-btn {
  margin-top: 32rpx;
  padding: 18rpx 56rpx;
  border-radius: 14rpx;
  background: #0A66C2;
  font-size: 26rpx;
  font-weight: 600;
  color: #fff;
}
</style>