<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="院校不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- ① Hero：真实校园全景图 + 三段蒙层 + 校徽半嵌 -->
        <view class="hero">
          <!-- 兜底层（图片缺失时可见） -->
          <view class="hero-fallback">
            <view class="hero-deco">
              <view class="deco-grid" />
              <view class="deco-radar" />
              <view class="deco-star s1" />
              <view class="deco-star s2" />
            </view>
          </view>

          <!-- 真实校园全景图（有则覆盖渐变） -->
          <image
            v-if="heroImage(detail)"
            :src="heroImage(detail)"
            mode="aspectFill"
            class="hero-img"
            lazy-load
            :style="{ opacity: imgLoaded.hero ? 1 : 0 }"
            @load="onHeroLoad"
          />

          <!-- 三段渐变蒙层：顶暗→中透→底暗 -->
          <view class="hero-mask-top" />
          <view class="hero-mask-bottom" />

          <!-- 导航层 -->
          <view class="hero-nav">
            <view class="back-btn" hover-class="press-feedback" :hover-stay-time="120" @click="goBack">
              <text class="back-icon">‹</text>
            </view>
            <view class="hero-action" @click="handleShare"><text class="share-icon">↗</text></view>
          </view>

          <!-- 圆形校徽（半嵌在图片底部边缘） -->
          <view class="hero-emblem"><text class="hero-emblem-char">{{ initShort(detail) }}</text></view>

          <!-- 标题 + 标签（叠在底部蒙层上） -->
          <view class="hero-content">
            <text class="hero-name">{{ detail.name || detail.title || '未知院校' }}</text>
            <text class="hero-location">{{ detail.city || '' }} · {{ detail.level_tags || detail.levelTags || (detail.tags || ['无人机专业']).join(' · ') }}</text>
            <view class="hero-tags">
              <text v-for="t in compTags(detail)" :key="t" class="hero-tag">{{ t }}</text>
            </view>
          </view>
        </view>

        <view class="main-card">
          <!-- ② 数据条（4 格） -->
          <view class="stats-row">
            <view class="stat" v-for="s in statsData(detail)" :key="s.label">
              <text class="stat-num">{{ s.value }}</text>
              <text class="stat-label">{{ s.label }}</text>
            </view>
          </view>

          <!-- ③ 院校简介 -->
          <view class="section-title">院校简介</view>
          <view class="intro-text">{{ detail.intro || detail.description || '暂无简介' }}</view>

          <!-- ④ 无人机相关专业 -->
          <view v-if="majorsList(detail).length > 0" class="section-block">
            <view class="section-title">无人机相关专业</view>
            <view class="major-list">
              <view v-for="m in majorsList(detail)" :key="m.name" class="major-item">
                <view class="major-info">
                  <text class="major-name">{{ m.name }}</text>
                  <text class="major-meta">{{ m.degree || '本科' }} · {{ m.duration || 4 }}年制{{ m.key ? ' · ' + m.key : '' }}</text>
                </view>
                <view v-if="m.flagship" class="flagship-tag">王牌</view>
              </view>
            </view>
          </view>

          <!-- ⑤ 合作企业 -->
          <view v-if="partnerList(detail).length > 0" class="section-block">
            <view class="section-title">合作企业</view>
            <scroll-view class="partner-scroll" scroll-x :show-scrollbar="false">
              <view v-for="p in partnerList(detail)" :key="p.name" class="partner-card">
                <view class="partner-icon"><text class="partner-icon-text">{{ p.icon || '企' }}</text></view>
                <text class="partner-name">{{ p.name }}</text>
                <text class="partner-type">{{ p.type || '合作单位' }}</text>
              </view>
            </scroll-view>
          </view>

          <!-- ⑥ 校园环境（4 图网格，按教学场景配色） -->
          <view class="section-block">
            <view class="section-title">校园环境</view>
            <view class="env-grid">
              <view
                v-for="(e, i) in envScenes"
                :key="e.name"
                class="env-item"
                @click="previewEnv(i)"
              >
                <!-- 真实校园场景图 -->
                <image
                  v-if="envImages(detail)[i]"
                  :src="envImages(detail)[i]"
                  class="env-img"
                  mode="aspectFill"
                  lazy-load
                  :style="{ opacity: imgLoaded['env' + i] ? 1 : 0 }"
                  @load="onEnvLoad(i)"
                />
                <!-- 无图兜底：按场景类型配色 -->
                <view v-else class="env-fallback" :class="'env-fallback--' + e.color">
                  <text class="env-char">{{ e.char }}</text>
                </view>
                <!-- 顶部色条 + 左下角标题 -->
                <view class="env-bar" :class="'env-bar--' + e.color" />
                <view class="env-label">
                  <view class="env-label-dot" :class="'env-label-dot--' + e.color" />
                  <text class="env-label-text">{{ e.name }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- ⑦ 底部按钮 -->
          <view class="bottom-bar">
            <view class="btn-outline" hover-class="press-feedback" :hover-stay-time="120" @click="callPhone">联系电话</view>
            <view v-if="detail.website" class="btn-primary" hover-class="press-feedback" :hover-stay-time="120" @click="openWebsite">访问官网</view>
          </view>
          <view class="bottom-spacer" />
        </view>
      </template>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
const imgLoaded = ref({ hero: false })

/** Hero 校园全景图：兼容 cover / image / campus_image / cover_image */
/** 生产库院校 cover 为空，按 id 兜底映射本地校园图（与列表页同一套） */
const COVER_FALLBACK = {
  'college-1': '/static/colleges/buaa-library.jpg',
  'college-2': '/static/colleges/nuaa-yufeng.jpg',
  'college-3': '/static/colleges/npu-building.jpg',
  'college-4': '/static/colleges/chengdu-building.jpg',
  'college-5': '/static/colleges/changsha-tiaoma.jpg',
  'college-6': '/static/colleges/cauc-scenery.jpg',
}
function heroImage(item) {
  const u = item.cover || item.image || item.campus_image || item.cover_image
  return u ? u : COVER_FALLBACK[item.id] || ''
}

function onHeroLoad() {
  imgLoaded.value.hero = true
}

/** 校园环境场景（景/学/飞/室，按教学场景类型配色） */
const envScenes = [
  { name: '景', char: '景', color: 'scenic' },
  { name: '学', char: '学', color: 'class' },
  { name: '飞', char: '飞', color: 'fly' },
  { name: '室', char: '室', color: 'lab' },
]

/** 校园环境图：兼容 photos / environment / env_images */
function envImages(item) {
  if (Array.isArray(item.photos) && item.photos.length > 0) return item.photos
  if (Array.isArray(item.environment)) return item.environment
  if (Array.isArray(item.env_images)) return item.env_images
  return []
}

function onEnvLoad(i) {
  imgLoaded.value['env' + i] = true
}

function previewEnv(idx) {
  const imgs = envImages(detail.value)
  if (imgs.length > 0) uni.previewImage({ urls: imgs, current: imgs[idx] || imgs[0] })
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags.slice(0, 3)
  if (Array.isArray(item.specialties)) return item.specialties
  return ['飞行器设计', '无人机工程']
}

function initShort(item) {
  if (item.short_name || item.shortName) return item.short_name || item.shortName
  var name = item.name || ''
  return name.charAt(0) || '院'
}

function statsData(item) {
  return [
    { label: '无人机专业', value: item.majorCount || item.major_count || 6 },
    { label: '合作企业', value: item.partnerCount || item.partner_count || 28 },
    { label: '在读学生', value: (item.studentCount || item.student_count || '320') + '+' },
    { label: '硕博导师', value: item.teacherCount || item.teacher_count || 12 },
  ]
}

function majorsList(item) {
  // 优先专业对象数组（majors_detail，扩展列）；majors 基础列若是字符串数组则跳过
  if (Array.isArray(item.majors_detail) && item.majors_detail.length > 0) return item.majors_detail
  if (Array.isArray(item.majors)) {
    var objList = item.majors.filter(function (m) { return m && typeof m === 'object' })
    if (objList.length > 0) return objList
  }
  return [
    { name: '飞行器设计与工程', degree: '本科', duration: 4, key: '国家级特色专业', flagship: true },
    { name: '无人机系统工程', degree: '本科', duration: 4, key: '', flagship: false },
    { name: '飞行器控制与信息工程', degree: '硕士', duration: 3, key: '', flagship: false },
  ]
}

function partnerList(item) {
  if (Array.isArray(item.partners) && item.partners.length > 0) return item.partners
  return [
    { icon: '机', name: '大疆创新', type: '联合实验室' },
    { icon: '航', name: '中航工业', type: '实习基地' },
    { icon: '天', name: '航天科技', type: '合作研究' },
    { icon: '装', name: '亿航智能', type: '人才输送' },
  ]
}

function previewPhotos(idx) {
  if (detail.value && Array.isArray(detail.value.photos)) {
    uni.previewImage({ urls: detail.value.photos, current: idx })
  }
}

function callPhone() {
  var phone = (detail.value && detail.value.phone) || '010-82310000'
  uni.makePhoneCall({ phoneNumber: phone })
}

function handleShare() {
  uni.showToast({ title: '分享功能开发中', icon: 'none' })
}

function websiteHost(u) {
  if (typeof u !== 'string') return ''
  var m = /^https?:\/\/([^\/?#]+)/.exec(u)
  return m ? m[1].replace(/:\d+$/, '').toLowerCase() : ''
}

function openWebsite() {
  var w = detail.value && detail.value.website
  if (!w) return
  // 显式声明业务白名单：仅放行该院校官网域名，webview 页据此校验 https + 域名
  var host = websiteHost(w) || 'api.cqnarc.cn'
  uni.navigateTo({
    url: '/pages/webview/index?url=' + encodeURIComponent(w) + '&allowed_domains=' + encodeURIComponent(host),
  })
}

function goBack() { uni.navigateBack({ delta: 1 }) }

async function loadDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    var res = await request({ url: '/api/v1/colleges/' + encodeURIComponent(id.value) })
    detail.value = res
    if (!res) errorMsg.value = '院校不存在'
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onLoad(function (options) {
  id.value = options.id || ''
  loadDetail()
})
</script>

<style scoped>
.page {
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.25, 0.46, 0.45, 0.94);
  min-height: 100vh;
  background: linear-gradient(180deg, #f5f6f8 0%, #E8F2FC 100%);
  padding-bottom: env(safe-area-inset-bottom);
}

/* ================================================================= */
/* ① 深空蓝 Hero                                                       */
/* ================================================================= */
.hero {
  height: 260rpx;
  background: linear-gradient(135deg, #074D92 0%, #0A66C2 100%);
  position: relative;
  overflow: hidden;
  padding: 88rpx 32rpx 40rpx;
}

/* 兜底层（图片缺失时显示装饰渐变） */
.hero-fallback {
  position: absolute;
  inset: 0;
}

.hero-deco { position: absolute; inset: 0; pointer-events: none; }

.deco-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.12) 2rpx, transparent 2rpx);
  background-size: 40rpx 40rpx;
  opacity: 0.6;
}

.deco-radar {
  position: absolute;
  right: -80rpx;
  top: -80rpx;
  width: 300rpx;
  height: 300rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.15);
  border-radius: 50%;
}

.deco-radar::before,
.deco-radar::after {
  content: '';
  position: absolute;
  inset: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

.deco-radar::after { inset: 90rpx; border: 2rpx solid rgba(255, 255, 255, 0.08); }

.deco-star {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  animation: twinkle 2.5s ease-in-out infinite;
}
.s1 { left: 60rpx; top: 80rpx; width: 6rpx; height: 6rpx; animation-delay: 0s; }
.s2 { left: 200rpx; top: 120rpx; width: 8rpx; height: 8rpx; animation-delay: 0.6s; }

/* 真实校园全景图 */
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity var(--anim-base) ease-out;
}

/* 三段渐变蒙层：顶暗→中透→底暗 */
.hero-mask-top {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 90rpx;
  background: linear-gradient(180deg, rgba(10, 31, 68, 0.65) 0%, rgba(10, 31, 68, 0) 100%);
  pointer-events: none;
  z-index: 1;
}

.hero-mask-bottom {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 120rpx;
  background: linear-gradient(0deg, rgba(10, 31, 68, 0.85) 0%, rgba(10, 31, 68, 0) 100%);
  pointer-events: none;
  z-index: 1;
}

.hero-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 2;
}

.back-btn {
  width: 88rpx;
  height: 88rpx;
  background: rgba(255, 255, 255, 0.15);
  border: 1rpx solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-icon { color: #ffffff; font-size: 44rpx; font-weight: 300; }

.hero-action {
  width: 88rpx;
  height: 88rpx;
  background: rgba(255, 255, 255, 0.15);
  border: 1rpx solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.share-icon { color: #ffffff; font-size: 36rpx; font-weight: 300; }

/* 圆形校徽：半嵌在图片底部边缘 */
.hero-emblem {
  position: absolute;
  left: 32rpx;
  bottom: -24rpx;
  width: 80rpx;
  height: 80rpx;
  background: #ffffff;
  border: 4rpx solid rgba(10, 102, 194, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
}

.hero-emblem-char {
  font-size: 34rpx;
  font-weight: 700;
  color: #0A66C2;
}

.hero-content {
  position: absolute;
  left: 32rpx;
  right: 32rpx;
  bottom: 32rpx;
  z-index: 2;
}

.hero-name {
  color: #ffffff;
  font-size: 40rpx;
  font-weight: 700;
  line-height: 1.2;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.3);
  display: block;
  margin-bottom: 8rpx;
}

.hero-location {
  color: rgba(255, 255, 255, 0.75);
  font-size: 24rpx;
  display: block;
  margin-bottom: 16rpx;
}

.hero-tags { display: flex; flex-wrap: wrap; gap: 12rpx; }

.hero-tag {
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.15);
  color: #ffffff;
  border: 1rpx solid rgba(255, 255, 255, 0.3);
}

/* ================================================================= */
/* 主卡片                                                             */
/* ================================================================= */
.main-card {
  background: #ffffff;
  border-radius: 32rpx 32rpx 0 0;
  margin-top: -32rpx;
  padding: 32rpx 32rpx 32rpx;
  position: relative;
  z-index: 2;
  animation: pageIn var(--anim-slow) var(--ease-out) both;
}

/* ② 数据条（4 格） */
.stats-row { display: flex; gap: 12rpx; margin-bottom: 36rpx; }
.stat { flex: 1; padding: 24rpx 8rpx; background: #fafafa; border-radius: 16rpx; text-align: center; }
.stat-num { font-size: 40rpx; font-weight: 700; color: #17212B; display: block; }
.stat-label { font-size: 22rpx; color: #969799; display: block; margin-top: 6rpx; }

/* Section */
.section-block { margin-top: 36rpx; }
.section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  padding-left: 20rpx;
  border-left: 6rpx solid #0A66C2;
  line-height: 1.3;
  margin-bottom: 20rpx;
}

/* ③ 简介 */
.intro-text {
  font-size: 28rpx;
  color: #17212B;
  line-height: 1.8;
  white-space: pre-line;
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(10, 31, 68, 0.04);
}

/* ④ 专业 */
.major-list { display: flex; flex-direction: column; gap: 16rpx; }

.major-item {
  padding: 24rpx;
  background: #fafafa;
  border-radius: 12rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.major-info { flex: 1; min-width: 0; }
.major-name { font-size: 28rpx; font-weight: 600; color: #17212B; display: block; }
.major-meta { font-size: 24rpx; color: #969799; display: block; margin-top: 6rpx; }

.flagship-tag {
  font-size: 22rpx;
  color: #E96012;
  background: #FFF4E6;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  font-weight: 600;
  flex-shrink: 0;
}

/* ⑤ 合作企业 */
.partner-scroll { display: flex; gap: 16rpx; white-space: nowrap; padding-bottom: 4rpx; }
.partner-card {
  padding: 24rpx 28rpx;
  background: #fafafa;
  border-radius: 16rpx;
  text-align: center;
  flex-shrink: 0;
  min-width: 160rpx;
  display: inline-block;
}

.partner-icon {
  width: 72rpx;
  height: 72rpx;
  margin: 0 auto 8rpx;
  background: rgba(10, 102, 194, 0.1);
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.partner-icon-text { font-size: 32rpx; color: #0A66C2; font-weight: 600; }
.partner-name { font-size: 26rpx; font-weight: 500; color: #17212B; display: block; }
.partner-type { font-size: 22rpx; color: #969799; display: block; margin-top: 4rpx; }

/* ⑥ 校园环境（4 图网格） */
.env-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16rpx;
}

.env-item {
  position: relative;
  height: 168rpx;
  border-radius: 12rpx;
  overflow: hidden;
}

.env-img {
  width: 100%;
  height: 100%;
  display: block;
  transition: opacity var(--anim-base) ease-out;
}

/* 兜底：按教学场景类型配色 */
.env-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.env-fallback--scenic { background: linear-gradient(135deg, #059669, #10B981); }
.env-fallback--class { background: linear-gradient(135deg, #074D92, #0A66C2); }
.env-fallback--fly { background: linear-gradient(135deg, #F97316, #E96012); }
.env-fallback--lab { background: linear-gradient(135deg, #8B5CF6, #A78BFA); }

.env-char { font-size: 56rpx; color: #ffffff; font-weight: 600; opacity: 0.85; }

/* 顶部色条 */
.env-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6rpx;
}
.env-bar--scenic { background: #10B981; }
.env-bar--class { background: #0A66C2; }
.env-bar--fly { background: #E96012; }
.env-bar--lab { background: #8B5CF6; }

/* 左下角标题 */
.env-label {
  position: absolute;
  left: 10rpx;
  bottom: 10rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  background: rgba(255, 255, 255, 0.9);
  padding: 4rpx 10rpx;
  border-radius: 999rpx;
}

.env-label-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
}
.env-label-dot--scenic { background: #10B981; }
.env-label-dot--class { background: #0A66C2; }
.env-label-dot--fly { background: #E96012; }
.env-label-dot--lab { background: #8B5CF6; }

.env-label-text { font-size: 20rpx; color: #17212B; font-weight: 500; }

/* ⑦ 底部 */
.bottom-bar {
  display: flex;
  gap: 24rpx;
  border-top: 1rpx solid #ebedf0;
  padding-top: 24rpx;
  margin-top: 36rpx;
}

.btn-outline {
  flex: 1;
  height: 88rpx;
  border-radius: 50rpx;
  border: 2rpx solid #0A66C2;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 500;
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.btn-primary {
  flex: 1;
  height: 88rpx;
  border-radius: 50rpx;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
  box-shadow: 0 8rpx 24rpx rgba(10, 102, 194, 0.3);
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.bottom-spacer { height: calc(40rpx + env(safe-area-inset-bottom)); }

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.2; }
  50%      { opacity: 0.8; }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .hero-content, .main-card, .btn-primary, .btn-outline {
    animation: none !important;
    transition: none !important;
  }
}
</style>
