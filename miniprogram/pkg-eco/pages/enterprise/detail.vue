<template>
  <view class="page">
    <!-- 骨架与 pkg-eco 其它详情页一致：StateView 统一承载 加载 / 失败 / 空态 -->
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="企业不存在或暂未公开"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- 顶部导航（对齐培训详情：独立导航条，白圆钮返回 + 居中标题 + 右侧分享） -->
        <view class="detail-nav" :style="{ paddingTop: statusBarHeight + 'px' }">
          <view class="detail-nav-back" hover-class="detail-nav-back--press" :hover-stay-time="120" aria-role="button" aria-label="返回" @click="goBack">
            <text>‹</text>
          </view>
          <text class="detail-nav-title">企业详情</text>
          <button class="share-btn" open-type="share" hover-class="detail-nav-back--press" :hover-stay-time="120">
            <text class="share-text">分享</text>
          </button>
        </view>

        <!-- ① Hero（对齐培训详情：内嵌圆角卡片，封面 + 蒙层 + 状态徽章 + 信息贴底） -->
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
            <view class="hero-fallback-icon"><text class="hero-fallback-char">{{ ent.name ? ent.name.charAt(0) : '企' }}</text></view>
          </view>
          <view class="hero-mask" />
          <!-- 状态徽章（对齐培训页：Hero 左上，白底胶囊 + 彩色文字） -->
          <view class="status-badge"><text class="status-text">{{ ent.is_member ? '协会会员' : '入驻企业' }}</text></view>
          <!-- Hero 底部信息区（对齐培训页 hero-bottom：标题 + 副标 + meta 行） -->
          <view class="hero-bottom">
            <text class="hero-title">{{ ent.name || '企业' }}</text>
            <text class="hero-org">{{ firstCategory() }}</text>
            <view class="hero-meta-row">
              <view class="meta-ico meta-ico--pin"><view class="pin-dot" /><view class="pin-tail" /></view>
              <text class="hero-meta-text">{{ ent.address || '地区待公开' }}</text>
            </view>
          </view>
        </view>

        <!-- ② 内容区（对齐培训详情页：直接在页面底色上排布卡片） -->
        <view class="content">
          <!-- 企业信息：决策信息前置（对齐培训详情 course-info-card） -->
          <view class="card info-card">
            <view class="info-head">
              <text class="info-head-title">企业信息</text>
              <text class="info-head-status">{{ ent.is_member ? '协会会员' : '已入驻' }}</text>
            </view>
            <view class="info-grid">
              <view class="info-item"><text class="info-label">企业规模</text><text class="info-value">{{ ent.scale || '—' }}</text></view>
              <view class="info-item"><text class="info-label">成立时间</text><text class="info-value">{{ ent.founded_at || '—' }}</text></view>
              <view class="info-item"><text class="info-label">所在地区</text><text class="info-value">{{ ent.address || '—' }}</text></view>
              <view class="info-item"><text class="info-label">营业时间</text><text class="info-value">{{ ent.business_hours || '—' }}</text></view>
            </view>
          </view>

          <!-- 平台背书（对齐培训详情 platform-endorse：首屏信任） -->
          <view class="card endorse">
            <view class="pe-mark"><view class="pe-check" /></view>
            <view class="pe-body">
              <text class="pe-title">协会平台审核 · 企业资质已核验</text>
              <text class="pe-sub">入驻资料经协会审核后公示</text>
            </view>
          </view>

          <!-- 标签分组卡（对齐培训详情 group-card） -->
          <view class="card group-card">
            <view class="group-block">
              <text class="group-title">行业分类</text>
              <view class="group-tags">
                <view v-for="(c, i) in categoryList(ent)" :key="c" class="g-tag" :class="'g-tag--c' + (i % 4)">
                  <text class="g-tag-text">{{ c }}</text>
                </view>
                <text v-if="categoryList(ent).length === 0" class="g-tag-empty">未填写</text>
              </view>
            </view>
            <view v-if="tagList(ent).length" class="group-divider" />
            <view v-if="tagList(ent).length" class="group-block">
              <text class="group-title">核心能力</text>
              <view class="group-tags">
                <view v-for="t in tagList(ent)" :key="t" class="g-tag g-tag--plain">
                  <text class="g-tag-text">{{ t }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 企业简介 -->
          <view class="section-head"><text class="section-title">企业简介</text></view>
          <view class="card intro-card"><text class="desc-text">{{ ent.description || '该企业暂未填写简介' }}</text></view>

          <text class="foot-note">信息由企业提交，经协会审核后公示</text>
        </view>
      </template>
    </StateView>

    <!-- ③ 底部固定操作栏（对齐培训详情 bottom-bar） -->
    <view v-if="detail" class="bottom-bar">
      <view class="btn-fav" :class="{ on: isFav }" hover-class="btn-fav-press" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" @click="toggleFav">
        <view class="fav-heart" :class="{ 'fav-heart--on': isFav }" />
        <text class="fav-label">{{ isFav ? '已收藏' : '收藏' }}</text>
      </view>
      <view class="bottom-actions">
        <button class="btn-outline" open-type="share" hover-class="btn-outline-press" hover-start-time="0" hover-stay-time="300" aria-label="转发">
          <text class="btn-outline-text">转发</text>
        </button>
        <view
          class="btn-primary"
          :class="{ 'btn-primary--disabled': !canContact }"
          hover-class="btn-primary-press"
          :hover-stay-time="100"
          aria-role="button"
          aria-label="联系企业"
          @click="onContact"
        >
          <view class="btn-phone-ico" />
          <text class="btn-primary-text">联系企业</text>
        </view>
      </view>
    </view>
  </view>
</template>
<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import StateView from '../../../components/StateView.vue'
import { request, BASE_URL } from '../../../utils/request'

// 状态机与 pkg-eco 其它详情页（赛事/院校）一致：
//   loading / errorMsg / detail 三个量交给 StateView 决定显示什么，页面只负责填。
//   detail 为 null 即空态（404 未审核或不存在），不再是手搓的 state-panel。
const loading = ref(true)
const errorMsg = ref('')
const detail = ref(null)
const statusBarHeight = ref(20)
const heroImgError = ref(false)
const isFav = ref(false)
let detailId = ''

// 模板里沿用 ent.xxx 的写法，同时保持 ent 只读（图片出错用 heroImgError 标记，不写回数据）
const ent = computed(() => detail.value || {})

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
const canContact = computed(() => !!(detail.value && (detail.value.contact_phone || detail.value.contact || detail.value.phone)))
const onContact = () => {
  const phone = detail.value && (detail.value.contact_phone || detail.value.contact || detail.value.phone)
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
const firstCategory = () => categoryList(ent.value)[0] || '入驻企业'

const formatDate = (d) => {
  if (!d) return '—'
  const dt = new Date(d)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

async function loadDetail() {
  if (!detailId) { loading.value = false; return }
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/enterprises/public/detail?id=' + encodeURIComponent(detailId) })
    detail.value = res || null
    heroImgError.value = false
    isFav.value = favs.has(detailId)
  } catch (e) {
    // 404（未审核/不存在）→ detail 保持 null，走 StateView 的空态（文案已说明原因）；
    // 其余（网络/服务端）→ 可重试的失败态
    const status = (e && e.statusCode) || 0
    if (status !== 404) errorMsg.value = (e && e.message) || '加载失败'
  } finally {
    loading.value = false
  }
}

onShareAppMessage(() => ({
  title: detail.value ? '入驻企业：' + (detail.value.name || '') : '入驻企业',
  path: '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(detailId),
}))

onLoad(async (query) => {
  // 状态栏高度同步取得，首帧即正确（与培训详情 .detail-nav 的 paddingTop 写法一致）
  try {
    statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
  detailId = (query && query.id) || ''
  loadDetail()
})
</script>

<style scoped>
/* 视觉基准 = 培训详情页（pkg-talent/pages/training/enroll.vue）最终态，
   骨架与同目录的赛事/院校详情一致：StateView + .detail-nav + .card。
   每条数值都取自该页，改之前请先对一遍，别凭手感调。 */
.page {
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
  overflow-x: hidden;
}

/* ═══ 顶部导航（对齐培训详情 .detail-nav：白圆钮返回 + 居中标题 + 右侧分享） ═══ */
.detail-nav {
  position: relative;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-left: 24rpx;
  padding-right: 24rpx;
  box-sizing: content-box;
  background: #F5F8FC;
}
.detail-nav-back {
  width: 60rpx;
  height: 60rpx;
  flex: 0 0 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #ffffff;
  box-shadow: 0 6rpx 16rpx rgba(31, 89, 169, 0.13);
}
.detail-nav-back--press { transform: scale(0.94); opacity: 0.86; }
.detail-nav-back text { margin-top: -4rpx; color: #1A3353; font-size: 42rpx; line-height: 1; }
.detail-nav-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  max-width: 56%;
  display: block;
  color: #17212B;
  font-size: 34rpx;
  font-weight: 700;
  text-align: center;
  line-height: 88rpx;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
/* 右侧分享（button 去默认样式） */
.share-btn {
  height: 60rpx;
  margin: 0;
  padding: 0 8rpx;
  line-height: 60rpx;
  text-align: center;
  background: transparent;
  border: none;
  font-size: 24rpx;
}
.share-btn::after { border: none; }
.share-text { color: #0A66C2; font-weight: 600; }

/* ═══ ① Hero（对齐培训详情：内嵌圆角卡片） ═══ */
.hero {
  position: relative;
  width: auto;
  height: 348rpx;
  margin: 0 24rpx;
  border-radius: 24rpx;
  overflow: hidden;
  box-shadow: 0 14rpx 34rpx rgba(31, 89, 169, 0.2);
}
.hero-img { width: 100%; height: 100%; }
.hero-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(145deg, #163C66 0%, #0A66C2 100%);
}
.hero-fallback-icon {
  width: 112rpx;
  height: 112rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 30rpx;
  background: rgba(255, 255, 255, 0.14);
}
.hero-fallback-char { font-size: 52rpx; font-weight: 700; color: rgba(255, 255, 255, 0.92); }
.hero-mask { background: linear-gradient(180deg, rgba(4, 30, 68, 0.08) 0%, rgba(4, 30, 68, 0.05) 34%, rgba(4, 30, 68, 0.8) 100%); position: absolute; inset: 0; }
.status-badge { position: absolute; top: 18rpx; left: 18rpx; padding: 7rpx 16rpx; border-radius: 999rpx; background: rgba(255, 255, 255, 0.92); }
.status-text { color: #0A66C2; font-size: 20rpx; font-weight: 650; }
.hero-bottom { position: absolute; left: 24rpx; right: 24rpx; bottom: 24rpx; }
.hero-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.3;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.18);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero-org { display: block; margin-top: 6rpx; font-size: 24rpx; color: rgba(255, 255, 255, 0.8); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.hero-meta-row { display: flex; align-items: center; gap: 8rpx; margin-top: 10rpx; }
.hero-meta-text { font-size: 24rpx; color: rgba(255, 255, 255, 0.85); }
/* 定位图标：CSS 画（项目规范不用 emoji/Unicode 图形，对齐培训详情的 meta-ico 写法） */
.meta-ico { position: relative; width: 24rpx; height: 24rpx; flex-shrink: 0; }
.pin-dot { position: absolute; left: 5rpx; top: 1rpx; width: 15rpx; height: 15rpx; box-sizing: border-box; border: 2rpx solid rgba(255, 255, 255, 0.9); border-radius: 50%; }
.pin-tail { position: absolute; left: 9rpx; top: 13rpx; width: 7rpx; height: 7rpx; box-sizing: border-box; border-right: 2rpx solid rgba(255, 255, 255, 0.9); border-bottom: 2rpx solid rgba(255, 255, 255, 0.9); transform: rotate(45deg); }

/* ═══ ② 内容区（对齐培训详情：页面底色上排布卡片） ═══ */
.content { position: relative; background: transparent; border-radius: 0; margin-top: 0; padding: 24rpx 24rpx 0; box-shadow: none; }
/* 通用卡片（对齐培训详情页：白卡 + 圆角 + 1rpx 边框 + 柔和投影） */
.card { position: relative; background: #ffffff; border: 1rpx solid #EEF1F4; border-radius: 16rpx; box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05); }

/* 企业信息四宫格（对齐培训详情 course-info-card 的最终态） */
.info-card { margin: 0 0 34rpx; padding: 22rpx; border-color: #E8EDF3; border-radius: 16rpx; background: #F8FAFC; box-shadow: none; }
.info-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18rpx; }
.info-head-title { font-size: 29rpx; font-weight: 700; color: #17212B; }
.info-head-status { padding: 5rpx 12rpx; border-radius: 999rpx; background: #EAF3FB; color: #0A66C2; font-size: 20rpx; font-weight: 600; }
.info-grid { display: flex; flex-wrap: wrap; overflow: hidden; border: 1rpx solid #E6ECF3; border-radius: 12rpx; background: #ffffff; }
.info-item { width: 50%; min-width: 0; min-height: 102rpx; display: flex; flex-direction: column; justify-content: center; gap: 8rpx; padding: 14rpx 16rpx; box-sizing: border-box; }
.info-item:nth-child(odd) { border-right: 1rpx solid #E6ECF3; }
.info-item:nth-child(-n + 2) { border-bottom: 1rpx solid #E6ECF3; }
.info-label { font-size: 20rpx; color: #8A94A4; }
.info-value { overflow: hidden; font-size: 22rpx; font-weight: 600; color: #344054; line-height: 1.35; white-space: nowrap; text-overflow: ellipsis; }

/* 平台背书（逐值对齐培训详情 .platform-endorse） */
.endorse { display: flex; align-items: center; gap: 18rpx; margin: 0 0 20rpx; padding: 20rpx 24rpx; background: #E7F1FC; border: 2rpx solid #C9DFF5; border-radius: 16rpx; }
.pe-mark { width: 68rpx; height: 68rpx; flex: 0 0 68rpx; border-radius: 50%; background: #0A66C2; display: flex; align-items: center; justify-content: center; }
.pe-check { width: 34rpx; height: 18rpx; border-left: 5rpx solid #fff; border-bottom: 5rpx solid #fff; transform: rotate(-45deg) translate(1rpx, -3rpx); }
.pe-body { display: flex; flex-direction: column; gap: 4rpx; }
.pe-title { font-size: 30rpx; font-weight: 700; color: #0A66C2; line-height: 1.3; }
.pe-sub { font-size: 20rpx; color: #4A6E94; line-height: 1.3; }

/* 标签分组卡（逐值对齐培训详情 .group-card） */
.group-card { margin: 0 0 20rpx; padding: 24rpx; border-radius: 20rpx; }
.group-title { display: block; font-size: 22rpx; font-weight: 600; color: #667085; letter-spacing: 0.4px; margin-bottom: 12rpx; }
.group-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }
.g-tag { display: inline-flex; align-items: center; padding: 6rpx 14rpx; border-radius: 6rpx; }
.g-tag-text { font-size: 22rpx; font-weight: 600; }
.g-tag--c0 { background: #E9F7F0; }
.g-tag--c0 .g-tag-text { color: #168A55; }
.g-tag--c1 { background: #FEF6E7; }
.g-tag--c1 .g-tag-text { color: #B54708; }
.g-tag--c2 { background: #EAF3FB; }
.g-tag--c2 .g-tag-text { color: #0A66C2; }
.g-tag--c3 { background: #FFF0E6; }
.g-tag--c3 .g-tag-text { color: #E96012; }
.g-tag--plain { background: #F4F6F8; }
.g-tag--plain .g-tag-text { color: #667085; }
.g-tag-empty { font-size: 22rpx; color: #98A2B3; }
.group-divider { height: 1rpx; background: #EEF1F4; margin: 20rpx 0; }

/* 章节标题（对齐培训详情 .section-head 最终态：无左侧蓝条，纯 30rpx 粗体） */
.section-head { display: flex; align-items: center; justify-content: space-between; margin: 0 0 16rpx; }
.section-title { font-size: 30rpx; font-weight: 700; color: #17212B; }
.intro-card { margin: 0; padding: 24rpx; border-radius: 20rpx; }
.desc-text { display: block; font-size: 26rpx; color: #344054; line-height: 1.7; }
.foot-note { display: block; margin: 24rpx 0 0; text-align: center; font-size: 20rpx; color: #98A2B3; }

/* ═══ ③ 底部固定操作栏（逐值对齐培训详情 .bottom-bar） ═══ */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid #E8EDF3;
  box-shadow: 0 -6rpx 18rpx rgba(16, 24, 40, 0.06);
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  z-index: 50;
}
/* 收藏：图标 + 文字纵向（对齐培训详情 .btn-fav） */
.btn-fav { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2rpx; padding: 0 12rpx; margin-right: 4rpx; flex-shrink: 0; border-radius: 10rpx; }
.btn-fav-press { opacity: 0.65; }
.fav-heart { width: 34rpx; height: 32rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%2398A2B3'%3E%3Cpath d='M12 21s-7.5-4.9-9.7-9.2C.6 8.4 2.6 4.5 6.3 4.5c2.2 0 3.9 1.2 4.7 3h2c.8-1.8 2.5-3 4.7-3 3.7 0 5.7 3.9 4 7.3C19.5 16.1 12 21 12 21z'/%3E%3C/svg%3E"); background-size: contain; background-repeat: no-repeat; background-position: center; }
.fav-heart--on { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23E96012'%3E%3Cpath d='M12 21s-7.5-4.9-9.7-9.2C.6 8.4 2.6 4.5 6.3 4.5c2.2 0 3.9 1.2 4.7 3h2c.8-1.8 2.5-3 4.7-3 3.7 0 5.7 3.9 4 7.3C19.5 16.1 12 21 12 21z'/%3E%3C/svg%3E"); }
.fav-label { font-size: 18rpx; color: #667085; }
.btn-fav.on .fav-label { color: #E96012; font-weight: 600; }
.bottom-actions { display: flex; gap: 12rpx; flex: 1; justify-content: flex-end; }
.btn-outline { display: flex; align-items: center; gap: 8rpx; height: 76rpx; padding: 0 24rpx; border: 1rpx solid #0A66C2; border-radius: 12rpx; background: transparent; margin: 0; line-height: normal; }
.btn-outline::after { border: none; }
.btn-outline-press { background: #EAF3FB; }
.btn-outline-text { font-size: 24rpx; font-weight: 700; color: #0A66C2; }
.btn-primary { height: 76rpx; padding: 0 32rpx; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; gap: 8rpx; background: #0A66C2; }
.btn-primary:active { background: #0759AA; }
.btn-primary-press { transform: scale(0.97); }
.btn-primary-text { font-size: 28rpx; font-weight: 700; color: #ffffff; }
.btn-primary--disabled { background: #C9CDD4; }
.btn-phone-ico { width: 28rpx; height: 28rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23FFFFFF' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1 1 .4 2 .7 2.9a2 2 0 0 1-.5 2.1L8.1 9.9a16 16 0 0 0 6 6l1.2-1.2a2 2 0 0 1 2.1-.5c.9.3 1.9.6 2.9.7A2 2 0 0 1 22 16.9z'/%3E%3C/svg%3E"); background-size: contain; background-repeat: no-repeat; background-position: center; }
</style>