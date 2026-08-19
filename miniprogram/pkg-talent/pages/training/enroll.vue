<template>
  <view class="page">
    <!-- 骨架屏：loading 时显示 -->
    <view v-if="loading" class="skeleton-list">
      <view class="sk-block" v-for="n in 4" :key="n">
        <view class="sk-title"></view>
        <view class="sk-line w80"></view>
        <view class="sk-line w60"></view>
      </view>
    </view>

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="机构不存在"
      @retry="fetchDetail"
    >
      <template v-if="detail">
        <!-- ====== Hero 区：真实图片 + 三段蒙层 + 机构校徽 ====== -->
        <view class="hero">
          <!-- 类型渐变兜底底（图片缺失时可见） -->
          <view class="hero-fallback">
            <view class="hero-deco">
              <view class="deco-grid" />
              <view class="deco-radar" />
              <view class="deco-star s1" />
              <view class="deco-star s2" />
            </view>
            <view class="hero-placeholder">
              <text class="hero-cam">＋</text>
              <text class="hero-placeholder-text">机构实景图</text>
            </view>
          </view>

          <!-- 真实图片（有则覆盖渐变） -->
          <image
            v-if="heroImage(detail)"
            :src="heroImage(detail)"
            mode="aspectFill"
            class="hero-img"
            :style="{ opacity: imgLoaded.banner ? 1 : 0 }"
            @load="onImgLoad('banner')"
          />

          <!-- 三段渐变蒙层（顶暗→中透→底暗，让按钮/校徽可读） -->
          <view class="hero-mask" />

          <!-- 底部过渡蒙层（图片 → 白卡自然衔接） -->
          <view class="hero-bottom-fade" />

          <!-- 左上角返回按钮（玻璃感） -->
          <view class="back-btn" hover-class="press-feedback" :hover-stay-time="120" @click="goBack">
            <text class="back-icon">‹</text>
          </view>

          <!-- 右上角状态徽章 -->
          <view class="hero-status">{{ statusText(detail) }}</view>

          <!-- 底部机构校徽（半嵌在图片底部边缘） -->
          <view class="hero-badge">
            <text class="hero-badge-char">{{ initShort(detail) }}</text>
          </view>
        </view>

        <!-- ====== 主卡片 ====== -->
        <view class="page-main">
        <view class="main-card">
          <!-- 评分 + 主标题(机构) + 副标题(课程) -->
          <view class="org-head">
            <view class="rating-box">
              <text class="rating-star">★</text>
              <text class="rating-score">{{ detail.rating || '5.0' }}</text>
              <text class="rating-reviews">{{ detail.review_count || 128 }} 人评价</text>
            </view>
            <view class="org-name">{{ orgNameOf(detail) }}</view>
            <view class="course-title">{{ courseTitleOf(detail) }}</view>
          </view>

          <!-- 课程类型标签 -->
          <view class="course-type-tags">
            <view
              v-for="ct in courseTypes(detail)"
              :key="ct"
              class="course-type-tag"
            >{{ ct }}</view>
          </view>

          <!-- 特色标签 -->
          <view class="feature-tags">
            <view
              v-for="ft in featureTags(detail)"
              :key="ft"
              class="feature-tag"
            >{{ ft }}</view>
          </view>

          <!-- 培训参考价（主次分明） -->
          <view class="section-block">
            <view class="section-title">培训参考价</view>
            <view class="price-subtitle">元 / 人 · 仅供参考，签约以机构确认为准</view>
            <view class="price-list">
              <view
                v-for="(p, i) in priceList(detail)"
                :key="i"
                class="price-item"
                :class="{ 'price-item--main': i === 0 }"
              >
                <view v-if="i === 0" class="price-hot">热销</view>
                <view class="price-bar" :class="i === 0 ? 'bar--main' : 'bar--sub'" />
                <view class="price-info">
                  <text class="price-name">{{ p.name }}</text>
                  <text class="price-desc">{{ i === 0 ? '含教材 + 考证 + 复训' : '含保险 + 1v1 教练' }}</text>
                </view>
                <view class="price-right">
                  <text class="price-symbol">¥</text>
                  <text class="price-value" :class="{ 'price-value--main': i === 0 }">{{ p.price }}</text>
                  <text class="price-unit">/{{ p.unit || '人' }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 机构简介 -->
          <view class="section-block">
            <view class="section-title">机构简介</view>
            <view class="org-intro">{{ orgIntro(detail) }}</view>
          </view>

          <!-- 联系信息（详情列表） -->
          <view class="section-block">
            <view class="section-title">联系信息</view>
            <view class="contact-list">
              <view class="contact-item" @click="openMap">
                <view class="contact-icon contact-icon--blue"><text class="contact-icon-text">址</text></view>
                <view class="contact-content">
                  <text class="contact-label">地址</text>
                  <text class="contact-value">{{ detail.location || '暂无' }}</text>
                </view>
                <text class="contact-arrow">›</text>
              </view>
              <view class="contact-item" @click="callPhone">
                <view class="contact-icon contact-icon--green"><text class="contact-icon-text">话</text></view>
                <view class="contact-content">
                  <text class="contact-label">电话</text>
                  <text class="contact-value contact-value--link">{{ detail.phone || detail.contact_phone || '400-116-0851' }}</text>
                </view>
                <text class="contact-arrow">›</text>
              </view>
              <view class="contact-item">
                <view class="contact-icon contact-icon--orange"><text class="contact-icon-text">时</text></view>
                <view class="contact-content">
                  <text class="contact-label">营业时间</text>
                  <text class="contact-value">{{ detail.business_hours || '周一至周日 09:00-18:00' }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 培训资格证（真实证书图） -->
          <view class="section-block">
            <view class="section-title">培训资格证</view>
            <view class="cert-image-wrap" @click="previewCert">
              <!-- 真实证书图（有则显示，aspectFill） -->
              <image
                v-if="certificateImage(detail)"
                :src="certificateImage(detail)"
                mode="aspectFill"
                lazy-load
                class="cert-image"
                :style="{ opacity: imgLoaded.certificate ? 1 : 0 }"
                @load="onImgLoad('certificate')"
              />
              <!-- 无图兜底：黄色占位（保留原设计，不出白块） -->
              <view v-else class="certificate-placeholder">
                <view class="cert-seal">✦</view>
                <text class="cert-title">民用无人机驾驶员训练机构合格证</text>
                <view class="cert-upload">
                  <text class="cert-cam">＋</text>
                  <text class="cert-upload-text">上传证书图</text>
                </view>
              </view>
              <!-- 右上角"已认证"绿标 -->
              <view class="cert-verified">已认证</view>
            </view>
            <view class="cert-tip">点击查看完整证书 ›</view>
          </view>

          <!-- 培训环境（三图网格） -->
          <view class="section-block">
            <view class="section-title">培训环境</view>
            <view class="env-grid">
              <view
                v-for="(e, i) in envPlaceholders"
                :key="i"
                class="env-cell"
                @click="previewEnv(i)"
              >
                <!-- 真实环境图（有则显示） -->
                <image
                  v-if="envImages(detail)[i]"
                  :src="envImages(detail)[i]"
                  class="env-cell-img"
                  mode="aspectFill"
                  :style="{ opacity: imgLoaded.env[i] ? 1 : 0 }"
                  @load="onImgLoad('env', i)"
                />
                <!-- 无图兜底：类型渐变色块 + 色条 + 标题 -->
                <view v-else class="env-cell-fallback" :class="'env-cell-fallback--' + e.color">
                  <view class="env-cell-bar" :class="'env-cell-bar--' + e.color" />
                  <text class="env-cell-icon">{{ e.icon }}</text>
                  <text class="env-cell-name">{{ e.name }}</text>
                </view>
              </view>
            </view>
          </view>

          <view class="bottom-placeholder" />
        </view>
        </view>
      </template>
    </StateView>

    <!-- 底部双按钮 -->
    <view v-if="detail" class="bottom-action-bar">
      <view class="action-btn consult-btn" hover-class="press-feedback" :hover-stay-time="120" @click="handleConsult">
        <text class="action-text">联系咨询</text>
      </view>
      <view class="action-btn enroll-btn" hover-class="press-feedback" :hover-stay-time="120" @click="handleEnroll">
        <text class="action-text">立即报名</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'
import StateView from '../../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
const imgLoaded = reactive({ banner: false, certificate: false, env: {} })

/* 环境占位数据 */
const envPlaceholders = computed(function () {
  return [
    { icon: '景', name: '培训场地实景', desc: '户外实操场地，含起降区与空域', color: 'blue' },
    { icon: '室', name: '理论教室环境', desc: '多媒体理论教室，可容纳 50 人', color: 'purple' },
    { icon: '飞', name: '户外飞行训练', desc: '标准飞行训练场，全天候开放', color: 'orange' },
  ]
})

/* 状态文字 */
function statusText(item) {
  var map = { recruiting: '招生中', full: '已满', urgent: '名额紧张', upcoming: '即将开课' }
  return map[item.status] || '招生中'
}

/* === 数据映射 === */

/** Hero 实景图：兼容 banner / cover_image / image */
function heroImage(item) {
  const u = item.banner || item.cover_image || item.image
  return u ? u : ''
}

/** 培训资格证图：兼容 certificate / certificate_url */
function certificateImage(item) {
  const u = item.certificate || item.certificate_url
  return u ? u : ''
}

/** 主标题 = 机构名；课程名降为副标题 */
function orgNameOf(item) {
  return item.org_name || item.enterprise_name || item.name || '未知机构'
}

function courseTitleOf(item) {
  return item.title || item.name || '未知课程'
}

/** 机构简称（校徽首字） */
function initShort(item) {
  const n = orgNameOf(item) || ''
  if (!n || n === '未知机构') return '培'
  // 取机构名第一个字符，跳过后缀关键词（如"学院/培训"）
  const strip = n.replace(/培训中心|飞行学院|分校|服务中心|培训基地|学院|中心|学校/gi, '')
  const base = strip || n
  return base.charAt(0)
}

function courseTypes(item) {
  if (Array.isArray(item.course_types) && item.course_types.length > 0) return item.course_types
  const ct = item.cert_type || 'CAAC'
  return [ct + '视距内', ct + '超视距']
}

function featureTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  const tags = []
  if (item.district) tags.push(item.district)
  else tags.push('花溪区')
  if (item.scale) tags.push(item.scale)
  else tags.push('规模大')
  tags.push('包住')
  tags.push('拿证快')
  tags.push('专业教培')
  return tags
}

function priceList(item) {
  if (Array.isArray(item.prices) && item.prices.length > 0) return item.prices
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    return item.courses.map(function (c) {
      return {
        name: c.name || c.title || c.cert_type || '课程',
        price: c.price != null ? c.price : (c.price_fen ? (c.price_fen / 100) : 0),
        unit: c.unit || '人',
      }
    })
  }
  const price = item.price != null ? item.price : (item.price_fen ? (item.price_fen / 100) : 5800)
  const ct = item.cert_type || 'CAAC'
  return [
    { name: ct + '视距内', price: price, unit: '人' },
    { name: ct + '超视距', price: Math.round(price * 1.5), unit: '人' },
  ]
}

function orgIntro(item) {
  const intro = item.intro || item.description || ''
  if (intro && intro.length > 40) return intro
  return '1、构建"能力培养-场景应用-生态共建"全链条服务。即搭建"考证培训—实景应用—企业赋能"闭环\n\n2、差异化课程设计-垂直场景深度绑定。慧飞行6大行业课程设计-植保、吊运、航测、航拍、巡检、应急消防\n\n3、从培训到销售、维修、维护、保养、保险、飞行服务、二手交易、覆盖用户全生命周期价值。'
}

function envImages(item) {
  if (Array.isArray(item.environment) && item.environment.length > 0) return item.environment
  if (Array.isArray(item.env_images)) return item.env_images
  if (Array.isArray(item.images)) return item.images
  return []
}

/* === 数据获取 === */
async function fetchDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/training-courses/' + encodeURIComponent(id.value) })
    detail.value = res
    if (!res) errorMsg.value = '机构不存在'
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

/* === 交互 === */
function onImgLoad(key, idx) {
  if (idx !== undefined) imgLoaded.env[idx] = true
  else imgLoaded[key] = true
}

function goBack() { uni.navigateBack({ delta: 1 }) }

function openMap() {
  const addr = (detail.value && detail.value.location) || ''
  uni.showToast({ title: addr ? '导航到：' + addr : '暂无地址信息', icon: 'none' })
}

function callPhone() {
  const phone = (detail.value && (detail.value.phone || detail.value.contact_phone)) || ''
  if (phone) uni.makePhoneCall({ phoneNumber: phone })
  else uni.showToast({ title: '暂无联系电话', icon: 'none' })
}

function handleConsult() {
  uni.showToast({ title: '已提交咨询，客服稍后联系', icon: 'none' })
}

function handleEnroll() {
  // 报名前登录守卫：未登录先引导登录，避免进入报名表单后 401 误报"报名失败"
  if (!requireLogin()) return
  // 隐私二次确认：报名表单将收集姓名与身份证信息，进入前明确用途
  uni.showModal({
    title: '信息授权确认',
    content: '报名培训需提交您的姓名与身份证信息，用于课程报名与证书核发，仅限本次报名使用，受隐私政策保护。',
    confirmText: '同意并继续',
    cancelText: '暂不报名',
    success: (res) => {
      if (res.confirm) {
        uni.navigateTo({ url: '/pkg-talent/pages/training/register?id=' + encodeURIComponent(id.value) })
      }
    },
  })
}

function previewCert() {
  const url = certificateImage(detail.value)
  if (url) uni.previewImage({ urls: [url], current: url })
}

function previewEnv(idx) {
  const imgs = envImages(detail.value)
  if (imgs.length > 0) uni.previewImage({ urls: imgs, current: imgs[idx] || imgs[0] })
}

onLoad(function (options) {
  id.value = options.id || ''
  fetchDetail()
})

onPullDownRefresh(function () {
  fetchDetail().then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page {
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.25, 0.46, 0.45, 0.94);
  min-height: 100vh;
  background: #fafafa;
}

/* ===== 动效 ===== */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}
.page-main { animation: pageIn var(--anim-slow) var(--ease-out) both; }

@keyframes blockIn {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}
.main-card .section-block { animation: blockIn var(--anim-base) var(--ease-out) both; }
.main-card .section-block:nth-of-type(1) { animation-delay: 80ms; }
.main-card .section-block:nth-of-type(2) { animation-delay: 160ms; }
.main-card .section-block:nth-of-type(3) { animation-delay: 240ms; }
.main-card .section-block:nth-of-type(4) { animation-delay: 320ms; }

@keyframes badgePulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(10, 102, 194, 0.3); }
  50% { box-shadow: 0 0 0 8rpx rgba(10, 102, 194, 0); }
}

.press-feedback { transform: scale(0.98); opacity: 0.85; }

/* 骨架屏 */
.skeleton-list { padding: 24rpx 32rpx; }
.sk-block { background: #ffffff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.03); }
.sk-title { width: 50%; height: 32rpx; border-radius: 6rpx; margin-bottom: 20rpx; animation: shimmer 1.2s linear infinite; background: linear-gradient(90deg, #f0f1f3 25%, #f8fafc 50%, #f0f1f3 75%); background-size: 200% 100%; }
.sk-line { height: 24rpx; border-radius: 4rpx; margin-bottom: 12rpx; animation: shimmer 1.2s linear infinite; background: linear-gradient(90deg, #f0f1f3 25%, #f8fafc 50%, #f0f1f3 75%); background-size: 200% 100%; }
.sk-line.w80 { width: 80%; }
.sk-line.w60 { width: 60%; }
@keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

@media (prefers-reduced-motion: reduce) {
  .page-main, .section-block, .press-feedback, .sk-title, .sk-line { animation: none !important; transition: none !important; }
}

/* ===== ① Hero ===== */
.hero {
  height: 360rpx;
  background: linear-gradient(135deg, #074D92 0%, #0A66C2 100%);
  position: relative;
  overflow: hidden;
  padding: 32rpx;
}

/* 渐变兜底层（图片缺失时显示） */
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
.deco-radar::before, .deco-radar::after {
  content: '';
  position: absolute;
  inset: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}
.deco-radar::after { inset: 90rpx; border: 2rpx solid rgba(255, 255, 255, 0.08); }

.deco-star { position: absolute; border-radius: 50%; background: rgba(255, 255, 255, 0.5); animation: twinkle 2.5s ease-in-out infinite; }
.s1 { left: 60rpx; top: 100rpx; width: 6rpx; height: 6rpx; animation-delay: 0s; }
.s2 { left: 200rpx; top: 60rpx; width: 8rpx; height: 8rpx; animation-delay: 0.6s; }

@keyframes twinkle { 0%, 100% { opacity: 0.2; } 50% { opacity: 0.8; } }

/* 真实图片 */
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity var(--anim-base) ease-out;
}

/* 三段渐变蒙层：顶暗→中透→底暗 */
.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg,
    rgba(10, 31, 68, 0.55) 0%,
    rgba(10, 31, 68, 0.1) 50%,
    rgba(10, 31, 68, 0.85) 100%);
  pointer-events: none;
  z-index: 5;
}

/* 底部过渡蒙层：图片底部 → 白卡自然衔接 */
.hero-bottom-fade {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 80rpx;
  background: linear-gradient(180deg, transparent 0%, rgba(10, 31, 68, 0.4) 100%);
  pointer-events: none;
  z-index: 6;
}

/* 左上角返回按钮（玻璃感） */
.back-btn {
  position: absolute;
  top: 32rpx;
  left: 32rpx;
  z-index: 20;
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

.hero-status {
  position: absolute;
  top: 40rpx;
  right: 32rpx;
  padding: 8rpx 20rpx;
  background: linear-gradient(135deg, #F97316, #E96012);
  color: #ffffff;
  font-size: 22rpx;
  font-weight: 600;
  border-radius: 999rpx;
  z-index: 20;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.2);
}

/* 底部机构校徽：圆形白底 + 机构首字，半嵌在图片底部边缘 */
.hero-badge {
  position: absolute;
  left: 32rpx;
  bottom: -40rpx;
  width: 96rpx;
  height: 96rpx;
  background: #ffffff;
  border-radius: 50%;
  border: 2rpx solid rgba(255, 255, 255, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 20;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.15);
}

.hero-badge-char {
  font-size: 36rpx;
  font-weight: 700;
  color: #0A66C2;
}

.hero-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
}

.hero-cam {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  border: 1rpx solid rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 40rpx;
  font-weight: 300;
}

.hero-placeholder-text { font-size: 24rpx; color: rgba(255, 255, 255, 0.75); letter-spacing: 2rpx; }

/* ===== ② 主卡片 ===== */
.main-card {
  background: #ffffff;
  border-radius: 32rpx 32rpx 0 0;
  margin-top: -40rpx;
  padding: 40rpx 32rpx 0;
  position: relative;
  z-index: 2;
}

/* 机构头 */
.org-head { margin-bottom: 16rpx; }

.rating-box { display: flex; align-items: center; gap: 8rpx; margin-bottom: 12rpx; }
.rating-star { font-size: 28rpx; color: #FFB300; }
.rating-score { font-size: 30rpx; font-weight: 700; color: #17212B; }
.rating-reviews { font-size: 22rpx; color: #c8c9cc; }

/* 主标题 = 机构名 */
.org-name { font-size: 36rpx; font-weight: 700; color: #17212B; line-height: 1.4; margin-bottom: 6rpx; }

/* 副标题 = 课程名 */
.course-title { font-size: 28rpx; color: #969799; line-height: 1.5; margin-bottom: 16rpx; }

/* 课程类型标签 */
.course-type-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 16rpx; }
.course-type-tag {
  padding: 6rpx 18rpx;
  border: 1rpx solid rgba(10, 102, 194, 0.3);
  border-radius: 999rpx;
  color: #0A66C2;
  font-size: 24rpx;
  font-weight: 500;
  background: rgba(10, 102, 194, 0.06);
}

/* 特色标签 */
.feature-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 32rpx; }
.feature-tag {
  padding: 6rpx 18rpx;
  background: rgba(52, 199, 89, 0.08);
  border: 1rpx solid rgba(52, 199, 89, 0.2);
  color: #34c759;
  font-size: 24rpx;
  border-radius: 999rpx;
  font-weight: 500;
}

/* Section */
.section-block { margin-top: 36rpx; padding-top: 28rpx; border-top: 1rpx solid #ebedf0; }
.section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  padding-left: 16rpx;
  border-left: 6rpx solid #0A66C2;
  line-height: 1.3;
  margin-bottom: 12rpx;
}
.price-subtitle { font-size: 22rpx; color: #c8c9cc; margin-bottom: 20rpx; padding-left: 16rpx; }

/* 价格列表（主次分明） */
.price-list { display: flex; flex-direction: column; gap: 16rpx; }

.price-item {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 24rpx;
  background: #fafafa;
  border-radius: 12rpx;
  position: relative;
  overflow: hidden;
}

.price-item--main {
  background: #ffffff;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.08);
}

.price-hot {
  position: absolute;
  top: 0;
  right: 0;
  padding: 4rpx 16rpx;
  background: linear-gradient(135deg, #F97316, #E96012);
  color: #ffffff;
  font-size: 20rpx;
  font-weight: 600;
  border-radius: 0 12rpx 0 12rpx;
}

.price-bar { width: 6rpx; height: 56rpx; border-radius: 3rpx; flex-shrink: 0; }
.bar--main { background: linear-gradient(135deg, #F97316, #E96012); }
.bar--sub { background: linear-gradient(135deg, #8B5CF6, #DB2777); }

.price-info { flex: 1; }
.price-name { font-size: 30rpx; color: #17212B; font-weight: 600; display: block; margin-bottom: 4rpx; }
.price-desc { font-size: 22rpx; color: #c8c9cc; }

.price-right { display: flex; align-items: baseline; }
.price-symbol { font-size: 24rpx; color: #E96012; font-weight: 600; }
.price-value { font-size: 32rpx; color: #17212B; font-weight: 700; margin: 0 4rpx; }
.price-value--main { font-size: 44rpx; color: #E96012; font-weight: 800; }
.price-unit { font-size: 22rpx; color: #c8c9cc; }

/* 机构简介 */
.org-intro { font-size: 28rpx; color: #17212B; line-height: 1.8; white-space: pre-line; }

/* 联系信息（详情列表） */
.contact-list { background: #fafafa; border-radius: 16rpx; padding: 0 24rpx; }

.contact-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #ebedf0;
}
.contact-item:last-child { border-bottom: none; }

.contact-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.contact-icon--blue { background: linear-gradient(135deg, #0A66C2, #0A66C2); }
.contact-icon--green { background: linear-gradient(135deg, #34c759, #22C55E); }
.contact-icon--orange { background: linear-gradient(135deg, #F97316, #E96012); }
.contact-icon-text { font-size: 28rpx; color: #ffffff; font-weight: 600; }

.contact-content { flex: 1; }
.contact-label { font-size: 22rpx; color: #969799; display: block; margin-bottom: 4rpx; }
.contact-value { font-size: 28rpx; color: #17212B; }
.contact-value--link { color: #0A66C2; font-weight: 600; }
.contact-arrow { color: #c8c9cc; font-size: 32rpx; }

/* 培训资格证（真实证书图） */
.cert-image-wrap {
  position: relative;
  width: 100%;
  height: 320rpx;
  border-radius: 12rpx;
  border: 1rpx solid rgba(10, 31, 68, 0.08);
  overflow: hidden;
}

.cert-image {
  width: 100%;
  height: 100%;
  display: block;
  transition: opacity var(--anim-base) ease-out;
}

/* 无图兜底：黄色占位 */
.certificate-placeholder {
  width: 100%;
  height: 100%;
  border-radius: 12rpx;
  background: linear-gradient(135deg, #FFFBEB, #FEF3C7);
  border: 1rpx solid #FDE68A;
  padding: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  box-sizing: border-box;
}

.cert-seal { font-size: 40rpx; color: #D97706; }
.cert-title { font-size: 26rpx; font-weight: 600; color: #92400E; text-align: center; }

.cert-upload {
  width: 200rpx;
  height: 80rpx;
  border: 2rpx dashed #F59E0B;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

.cert-cam { font-size: 28rpx; color: #D97706; font-weight: 300; }
.cert-upload-text { font-size: 22rpx; color: #D97706; }

/* 右上角"已认证"绿标 */
.cert-verified {
  position: absolute;
  top: 16rpx;
  right: 16rpx;
  padding: 4rpx 12rpx;
  background: rgba(52, 199, 89, 0.9);
  color: #ffffff;
  font-size: 20rpx;
  font-weight: 600;
  border-radius: 999rpx;
  z-index: 2;
}

/* 下方提示文字 */
.cert-tip {
  margin-top: 12rpx;
  text-align: center;
  font-size: 22rpx;
  color: #0A66C2;
  text-decoration: underline;
}

/* 培训环境（三图网格） */
.env-grid {
  display: flex;
  gap: 12rpx;
}

.env-cell {
  flex: 1;
  height: 200rpx;
  border-radius: 12rpx;
  overflow: hidden;
  position: relative;
}

.env-cell-img {
  width: 100%;
  height: 100%;
  display: block;
  transition: opacity var(--anim-base) ease-out;
}

.env-cell-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  position: relative;
  border-radius: 12rpx;
}

.env-cell-fallback--blue { background: linear-gradient(135deg, #074D92, #0A66C2); }
.env-cell-fallback--purple { background: linear-gradient(135deg, #8B5CF6, #A78BFA); }
.env-cell-fallback--orange { background: linear-gradient(135deg, #F97316, #E96012); }

/* 顶部色条 */
.env-cell-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 8rpx;
}
.env-cell-bar--blue { background: #0A66C2; }
.env-cell-bar--purple { background: #8B5CF6; }
.env-cell-bar--orange { background: #F97316; }

.env-cell-icon { font-size: 40rpx; color: #ffffff; font-weight: 600; }
.env-cell-name { font-size: 22rpx; color: rgba(255, 255, 255, 0.9); font-weight: 500; }

.bottom-placeholder { height: 180rpx; }

/* 底部双按钮 */
.bottom-action-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.95);
  z-index: 100;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
  display: flex;
  gap: 20rpx;
}

.action-btn {
  flex: 1;
  height: 96rpx;
  border-radius: 50rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.enroll-btn {
  background: linear-gradient(135deg, #074D92 0%, #0A66C2 100%);
  box-shadow: 0 8rpx 24rpx rgba(10, 102, 194, 0.3);
  animation: badgePulse 2s ease-in-out infinite;
}

.consult-btn {
  border: 2rpx solid #0A66C2;
  background: #ffffff;
}

.action-text { color: #ffffff; font-size: 32rpx; font-weight: 600; letter-spacing: 2rpx; }
.consult-btn .action-text { color: #0A66C2; }
</style>
