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
        <!-- 顶部导航（对齐课程详情：独立导航条，返回 + 居中标题 + 分享） -->
        <view class="detail-nav" :style="{ paddingTop: statusBarH + 'px' }">
          <view class="detail-nav-back" hover-class="detail-nav-back--press" :hover-stay-time="120" aria-role="button" aria-label="返回" @click="goBack">
            <text>‹</text>
          </view>
          <text class="detail-nav-title">院校详情</text>
          <button class="share-btn" open-type="share" hover-class="detail-nav-back--press" :hover-stay-time="120">
            <text class="share-btn-text">分享</text>
          </button>
        </view>

        <!-- ① Hero（对齐课程详情：内嵌圆角卡片，封面 + 蒙层 + 信息贴底） -->
        <view class="hero">
          <image
            v-if="heroImage(detail)"
            :src="heroImage(detail)"
            mode="aspectFill"
            class="hero-img"
            lazy-load
            :style="{ opacity: imgLoaded.hero ? 1 : 0 }"
            @load="onHeroLoad"
          />

          <!-- 兜底层（图片缺失时可见） -->
          <view v-if="!heroImage(detail)" class="hero-fallback">
            <view class="hero-deco">
              <view class="deco-grid" />
              <view class="deco-radar" />
            </view>
          </view>

          <view class="hero-mask" />

          <!-- 底部信息（标题 + 城市层次 + 标签） -->
          <view class="hero-bottom">
            <text class="hero-name">{{ detail.name || detail.title || '未知院校' }}</text>
            <text class="hero-location">{{ heroLoc(detail) }}</text>
            <view class="hero-tags">
              <text v-for="t in compTags(detail)" :key="t" class="hero-tag">{{ t }}</text>
            </view>
          </view>
        </view>

        <view class="content">
          <!-- ② 数据条卡（4 格：彩色图标 + 数字） -->
          <view class="card stats-card">
            <view class="stats-row">
              <view class="stat" v-for="s in statsData(detail)" :key="s.label">
                <view class="stat-icon" :class="'stat-icon--' + s.color">
                  <view class="stat-ico" :class="'stat-ico--' + s.icon" />
                </view>
                <text class="stat-num">{{ s.value }}</text>
                <text class="stat-label">{{ s.label }}</text>
              </view>
            </view>
          </view>

          <!-- ③ 院校简介 -->
          <view class="section-block">
            <text class="section-title">院校简介</text>
            <view class="card intro-card">
              <rich-text v-if="((detail.intro || detail.description) || '').indexOf('<') >= 0" class="intro-text" :nodes="detail.intro || detail.description"></rich-text>
        <text v-else class="intro-text">{{ detail.intro || detail.description || '暂无简介' }}</text>
            </view>
          </view>

          <!-- ④ 无人机相关专业 -->
          <view v-if="majorsList(detail).length > 0" class="section-block">
            <text class="section-title">无人机相关专业</text>
            <view class="major-list">
              <view v-for="m in majorsList(detail)" :key="m.name" class="card major-item">
                <view class="major-left">
                  <view class="major-icon"><view class="major-ico" /></view>
                  <view class="major-info">
                    <text class="major-name">{{ m.name }}</text>
                    <text class="major-meta">{{ m.degree || '本科' }} · {{ m.duration || 4 }}年制{{ m.key ? ' · ' + m.key : '' }}</text>
                  </view>
                </view>
                <view v-if="m.flagship" class="flagship-tag">王牌</view>
              </view>
            </view>
          </view>

          <!-- ⑤ 合作企业 -->
          <view v-if="partnerList(detail).length > 0" class="section-block">
            <text class="section-title">合作企业</text>
            <scroll-view class="partner-scroll" scroll-x :show-scrollbar="false">
              <view v-for="p in partnerList(detail)" :key="p.name" class="card partner-card">
                <view class="partner-icon"></view>
                <text class="partner-name">{{ p.name }}</text>
                <text class="partner-type">{{ p.type || '合作单位' }}</text>
              </view>
            </scroll-view>
          </view>

          <!-- ⑥ 校园环境（4 图网格，按教学场景配色） -->
          <view class="section-block">
            <text class="section-title">校园环境</text>
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

          <view class="bottom-space" />
        </view>

        <!-- ⑦ 底部固定栏（联系电话 + 访问官网） -->
        <view class="bottom-bar">
          <view class="btn-outline" v-if="detail.phone" hover-class="press-feedback" :hover-stay-time="120" @click="callPhone">联系电话</view>
          <view v-if="detail.website" class="btn-primary" hover-class="press-feedback" :hover-stay-time="120" @click="openWebsite">访问官网</view>
        </view>
      </template>
    </StateView>
  </view>
</template>

<script setup>
import { safeBack } from '../../../utils/nav'
import { ref } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'

// 状态栏高度（Hero 导航避让）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
import StateView from '../../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
const imgLoaded = ref({ hero: false })

/** Hero 校园全景图：兼容 cover / image / campus_image / cover_image / logo_url */
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
  const u = item.cover || item.image || item.campus_image || item.cover_image || item.logo_url
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

/* 副标题：城市 · 层次标签（tag 为空时兜底"高校"，不出现孤立"·"） */
function heroLoc(item) {
  var parts = []
  if (item.city) parts.push(item.city)
  var lv = item.level_tags || item.levelTags || ''
  if (lv) parts.push(lv)
  else if (Array.isArray(item.tags) && item.tags.length) parts.push(item.tags.join(' · '))
  return parts.join(' · ') || '高校'
}

function statsData(item) {
  // 未填（0/空）显示 "—"，不伪造默认数字；兼容 snake/camel 两种字段名
  const num = (v) => { const n = Number(v); return n > 0 ? n : null }
  const val = (snake, camel) => num(item[snake] != null ? item[snake] : item[camel])
  const students = val('student_count', 'studentCount')
  return [
    { label: '无人机专业', value: val('major_count', 'majorCount') ?? '—', color: 'blue', icon: 'major' },
    { label: '合作企业', value: val('partner_count', 'partnerCount') ?? '—', color: 'purple', icon: 'partner' },
    { label: '在读学生', value: students != null ? students + '+' : '—', color: 'green', icon: 'student' },
    { label: '硕博导师', value: val('teacher_count', 'teacherCount') ?? '—', color: 'orange', icon: 'teacher' },
  ]
}

function majorsList(item) {
  // 优先专业对象数组（majors_detail，扩展列）；majors 基础列若是字符串数组则跳过
  if (Array.isArray(item.majors_detail) && item.majors_detail.length > 0) return item.majors_detail
  if (Array.isArray(item.majors)) {
    var objList = item.majors.filter(function (m) { return m && typeof m === 'object' })
    if (objList.length > 0) return objList
  }
  return []
}

function partnerList(item) {
  if (Array.isArray(item.partners) && item.partners.length > 0) return item.partners
  return []
}

function previewPhotos(idx) {
  if (detail.value && Array.isArray(detail.value.photos)) {
    uni.previewImage({ urls: detail.value.photos, current: idx })
  }
}

function callPhone() {
  var phone = detail.value && detail.value.phone
  if (!phone) return
  uni.makePhoneCall({ phoneNumber: phone })
}

/* 分享（导航"分享"按钮 / 右上角菜单） */
onShareAppMessage(function () {
  var d = (detail && detail.value) || {}
  return {
    title: '院校：' + (d.name || d.title || '无人机院校'),
    path: '/pkg-eco/pages/colleges/detail?id=' + encodeURIComponent(d.id || ''),
  }
})

function websiteHost(u) {
  if (typeof u !== 'string') return ''
  var m = /^https?:\/\/([^\/?#]+)/.exec(u)
  return m ? m[1].replace(/:\d+$/, '').toLowerCase() : ''
}

/** 官网跳转白名单：固定列表，仅放行 .edu.cn 后缀的教育机构官网域名（不由 URL 自行推导） */
const WEBSITE_ALLOW_SUFFIXES = ['.edu.cn']

function openWebsite() {
  var w = detail.value && detail.value.website
  if (!w) return
  var host = websiteHost(w)
  var allowed = !!host && WEBSITE_ALLOW_SUFFIXES.some(function (suffix) {
    return host.indexOf(suffix) === host.length - suffix.length
  })
  if (!allowed) {
    uni.showToast({ title: '该官网暂不支持跳转', icon: 'none' })
    return
  }
  uni.navigateTo({
    url: '/pages/webview/index?url=' + encodeURIComponent(w) + '&allowed_domains=' + encodeURIComponent(host),
  })
}

function goBack() { safeBack() }

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
  min-height: 100vh;
  background: #F5F8FC;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ================================================================= */
/* 顶部导航（对齐课程详情：白圆钮返回 + 居中标题 + 分享）                */
/* ================================================================= */
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
.share-btn {
  height: 60rpx;
  flex: 0 0 60rpx;
  margin: 0;
  padding: 0 8rpx;
  line-height: 60rpx;
  text-align: center;
  background: transparent;
  border: none;
  font-size: 24rpx;
}
.share-btn::after { border: none; }
.share-btn-text { color: #0A66C2; font-weight: 600; }

/* ================================================================= */
/* ① Hero（对齐课程详情：内嵌圆角卡片）                                */
/* ================================================================= */
.hero {
  position: relative;
  width: auto;
  height: 348rpx;
  margin: 0 24rpx;
  border-radius: 24rpx;
  overflow: hidden;
  background: linear-gradient(145deg, #074D92 0%, #0A66C2 100%);
  box-shadow: 0 14rpx 34rpx rgba(31, 89, 169, 0.2);
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

/* 真实校园全景图 */
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transition: opacity 240ms ease-out;
}

/* 单层渐变蒙层（对齐课程 hero-mask） */
.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(4, 30, 68, 0.08) 0%, rgba(4, 30, 68, 0.05) 34%, rgba(4, 30, 68, 0.8) 100%);
  pointer-events: none;
}

/* Hero 底部信息（标题 + 城市层次 + 标签） */
.hero-bottom {
  position: absolute;
  left: 24rpx;
  right: 24rpx;
  bottom: 24rpx;
  z-index: 3;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}
.hero-name {
  color: #ffffff;
  font-size: 36rpx;
  font-weight: 700;
  line-height: 1.3;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.32);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.hero-location {
  color: rgba(255, 255, 255, 0.78);
  font-size: 24rpx;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.hero-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }
.hero-tag {
  padding: 5rpx 12rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
  border: 1rpx solid rgba(255, 255, 255, 0.3);
}

/* ================================================================= */
/* 内容区（直接在页面底色排布，卡片语言对齐课程）                        */
/* ================================================================= */
.content {
  position: relative;
  margin-top: 20rpx;
  z-index: 2;
  padding: 0 24rpx;
}

/* 通用卡片：白底 + 16px 圆角 + 1rpx 边框 + 柔和投影 */
.card {
  position: relative;
  background: #ffffff;
  border: 1rpx solid #E8EDF3;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.06);
  overflow: hidden;
}

/* ② 数据条（4 格：彩色图标瓷片 + 数字） */
.stats-card { padding: 26rpx 8rpx; }
.stats-row { display: flex; }
.stat { flex: 1; min-width: 0; text-align: center; padding: 6rpx 6rpx; }
.stat + .stat { border-left: 1rpx solid #EEF1F4; }
.stat-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 16rpx;
  margin: 0 auto 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon--blue { background: #E8F2FC; }
.stat-icon--purple { background: #F3F0FF; }
.stat-icon--green { background: #E9F7F0; }
.stat-icon--orange { background: #FFF0E6; }
.stat-ico { width: 30rpx; height: 30rpx; background-size: contain; background-repeat: no-repeat; background-position: center; }
.stat-ico--major { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230A66C2' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M12 4L2 9l10 5 10-5-10-5z'/%3E%3Cpath d='M6 11.5V16c0 1.7 2.7 3 6 3s6-1.3 6-3v-4.5'/%3E%3Cpath d='M22 9v5'/%3E%3C/svg%3E"); }
.stat-ico--partner { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%237C3AED' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M3 21h18M5 21V8l7-4 7 4v13M9 21v-4h6v4'/%3E%3C/svg%3E"); }
.stat-ico--student { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230B6B41' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='9' cy='8' r='3.5'/%3E%3Cpath d='M3 20a6 6 0 0 1 12 0'/%3E%3Ccircle cx='17' cy='9' r='2.5'/%3E%3Cpath d='M14.5 20a5.5 5.5 0 0 1 6-5.4'/%3E%3C/svg%3E"); }
.stat-ico--teacher { background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23E96012' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='12' cy='7' r='3.5'/%3E%3Cpath d='M4.5 20a7.5 7.5 0 0 1 15 0'/%3E%3C/svg%3E"); }
.stat-num { font-size: 38rpx; font-weight: 700; color: #17212B; display: block; }
.stat-label { font-size: 21rpx; color: #7A8798; display: block; margin-top: 6rpx; }

/* Section */
.section-block { margin-top: 28rpx; }
.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.3;
  margin-bottom: 14rpx;
}

/* ③ 简介 */
.intro-card { padding: 24rpx; }
.intro-text {
  font-size: 28rpx;
  color: #344054;
  line-height: 1.8;
  white-space: pre-line;
  display: block;
}

/* ④ 专业 */
.major-list { display: flex; flex-direction: column; gap: 12rpx; }
.major-item {
  padding: 22rpx 24rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.major-left { display: flex; align-items: center; gap: 16rpx; flex: 1; min-width: 0; }
.major-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 14rpx;
  background: #E8F2FC;
  position: relative;
  flex-shrink: 0;
}
.major-ico {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230A66C2' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M4 19.5A2.5 2.5 0 0 1 6.5 17H20'/%3E%3Cpath d='M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z'/%3E%3C/svg%3E");
  background-size: 58% auto;
  background-repeat: no-repeat;
  background-position: center;
}
.major-info { flex: 1; min-width: 0; }
.major-name { font-size: 28rpx; font-weight: 600; color: #17212B; display: block; }
.major-meta { font-size: 23rpx; color: #7A8798; display: block; margin-top: 6rpx; }
.flagship-tag {
  font-size: 22rpx;
  color: #E96012;
  background: #FFF0E6;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  font-weight: 600;
  flex-shrink: 0;
}

/* ⑤ 合作企业 */
.partner-scroll { display: flex; gap: 12rpx; white-space: nowrap; padding-bottom: 4rpx; }
.partner-card {
  padding: 24rpx 24rpx;
  text-align: center;
  flex-shrink: 0;
  min-width: 160rpx;
  display: inline-block;
}
.partner-icon {
  width: 56rpx;
  height: 56rpx;
  margin: 0 auto 8rpx;
  background: linear-gradient(135deg, #0A66C2, #2889DD);
  border-radius: 14rpx;
  position: relative;
  box-shadow: 0 4rpx 12rpx rgba(10, 102, 194, 0.22);
}
.partner-icon::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23ffffff' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M3 21h18M5 21V8l7-4 7 4v13M9 21v-4h6v4'/%3E%3C/svg%3E");
  background-size: 58% auto;
  background-repeat: no-repeat;
  background-position: center;
}
.partner-name { font-size: 26rpx; font-weight: 500; color: #17212B; display: block; }
.partner-type { font-size: 22rpx; color: #98A2B3; display: block; margin-top: 4rpx; }

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

/* ⑦ 底部固定栏（联系电话 + 访问官网，对齐课程底部栏） */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 100;
  display: flex;
  gap: 16rpx;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid #E8EDF3;
  box-shadow: 0 -6rpx 18rpx rgba(16, 24, 40, 0.06);
}
.btn-outline {
  flex: 1;
  height: 76rpx;
  border-radius: 12px;
  border: 2rpx solid #A6C9EE;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
  box-sizing: border-box;
  transition: transform 180ms ease, opacity 160ms ease;
}
.btn-primary {
  flex: 1;
  height: 76rpx;
  border-radius: 12px;
  background: #0A66C2;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
  box-shadow: 0 4rpx 10rpx rgba(10, 102, 194, 0.28);
  transition: transform 180ms ease, opacity 160ms ease;
}
.btn-primary:active { background: #0759AA; }
.bottom-space { height: calc(140rpx + env(safe-area-inset-bottom)); }

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .hero-bottom, .content, .btn-primary, .btn-outline {
    animation: none !important;
    transition: none !important;
  }
}
</style>
