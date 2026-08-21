<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="机构不存在"
      @retry="fetchDetail"
    >
      <template v-if="detail">
        <!-- ═══ ① 全屏 Hero（250px，机构实景图 + 遮罩）═══ -->
        <view class="hero">
          <image v-if="heroImage(detail)" :src="heroImage(detail)" mode="aspectFill" class="hero-img" lazy-load />
          <view v-else class="hero-fallback">
            <view class="drone-svg">
              <view class="drone-prop p1" /><view class="drone-prop p2" /><view class="drone-prop p3" /><view class="drone-prop p4" />
              <view class="drone-arm a1" /><view class="drone-arm a2" />
              <view class="drone-body" />
              <view class="drone-gimbal" />
            </view>
          </view>

          <view class="hero-mask" />
          <view class="hero-highlight" />

          <!-- 顶部导航（毛玻璃） -->
          <view class="hero-nav">
            <view class="nav-back" hover-class="nav-press" :hover-stay-time="100" @click="goBack">
              <text class="nav-back-icon">‹</text>
            </view>
            <view class="nav-capsule">
              <view class="capsule-dot" />
              <view class="capsule-divider" />
              <view class="capsule-arrow" />
            </view>
          </view>

          <!-- 左上状态徽章 -->
          <view class="status-badge">
            <view class="status-dot" />
            <text class="status-text">{{ statusText(detail) }}</text>
          </view>

          <!-- Hero 底部信息区：机构名 + 课程周期 -->
          <view class="hero-bottom">
            <text class="hero-title">{{ orgNameOf(detail) }}</text>
            <view class="hero-meta-row">
              <view class="meta-ico meta-ico--cal"><view class="cal-top" /><view class="cal-body"><view class="cal-line l1" /><view class="cal-line l2" /><view class="cal-line l3" /></view></view>
              <text class="hero-meta-text">{{ coursePeriod(detail) }}</text>
            </view>
          </view>
        </view>

        <!-- ═══ ② 白色内容区 ═══ -->
        <view class="content">
          <!-- 评分卡 -->
          <view class="section">
            <view class="rating-card">
              <view class="rating-left">
                <text class="rating-score">{{ ratingOf(detail) }}<text class="rating-total">/5.0</text></text>
                <view class="rating-stars">
                  <text v-for="n in 5" :key="n" class="star" :class="{ 'star--on': n <= starCount(detail) }">★</text>
                </view>
              </view>
              <view class="rating-divider" />
              <view class="rating-right">
                <view class="r-stat"><text class="r-key">累计评价</text><text class="r-val">{{ detail.review_count || 0 }} 条</text></view>
                <view class="r-stat"><text class="r-key">通过考试</text><text class="r-val">{{ passRateOf(detail) === '—' ? '—' : passRateOf(detail) + '%' }}</text></view>
                <view class="r-stat"><text class="r-key">机构年限</text><text class="r-val">{{ yearsOf(detail) === '—' ? '—' : yearsOf(detail) + ' 年' }}</text></view>
              </view>
            </view>
          </view>

          <!-- 标签分组卡 -->
          <view class="section">
            <view class="group-card">
              <view class="group-block">
                <text class="group-title">证书类型</text>
                <view class="group-tags">
                  <view v-for="ct in certTypeTags(detail)" :key="ct" class="g-tag g-tag--cert">
                    <view class="g-tag-ico g-tag-ico--cert"><view class="tag-drone" /></view>
                    <text class="g-tag-text">{{ ct }}</text>
                  </view>
                </view>
              </view>
              <view class="group-divider" v-if="featureTags(detail).length > 0" />
              <view class="group-block" v-if="featureTags(detail).length > 0">
                <text class="group-title">机构服务</text>
                <view class="group-tags">
                  <view v-for="(ft, i) in featureTags(detail)" :key="ft" class="g-tag" :class="'g-tag--c' + (i % 5)">
                    <text class="g-tag-text">{{ ft }}</text>
                  </view>
                </view>
              </view>
            </view>
          </view>

          <!-- 培训参考价卡 -->
          <view class="section">
            <view class="section-head">
              <text class="section-title">培训参考价</text>
              <text class="section-sub">元 / 人 · 仅供参考</text>
            </view>
            <view class="price-list" v-if="priceList(detail).length > 0">
              <view
                v-for="(p, i) in priceList(detail)"
                :key="i"
                class="price-item"
                :class="{ 'price-item--hot': i === 0, 'price-item--active': activePriceIndex === i }"
                hover-class="price-press"
                :hover-stay-time="100"
                @click="handlePriceTap(p, i)"
              >
                <view v-if="i === 0" class="price-hot-badge">热销</view>
                <view class="price-left">
                  <text class="price-name">{{ p.name }}</text>
                  <view class="price-desc">
                    <text v-for="(d, di) in priceIncludes(i)" :key="di" class="inc-item">{{ d }}</text>
                  </view>
                </view>
                <view class="price-right">
                  <view v-if="activePriceIndex === i" class="price-check" />
                  <text class="price-symbol">¥</text>
                  <text class="price-value" :class="{ 'price-value--hot': i === 0 }">{{ p.price }}</text>
                  <text class="price-unit">/{{ p.unit || '人' }}</text>
                </view>
              </view>
            </view>
            <view v-else class="price-empty">费用面议</view>
          </view>

          <!-- 机构简介卡 -->
          <view v-if="orgIntro(detail)" class="section">
            <text class="section-title">机构简介</text>
            <view class="intro-card">
              <text class="intro-text">{{ orgIntro(detail) }}</text>
            </view>
          </view>

          <!-- 联系信息卡 -->
          <view class="section">
            <text class="section-title">联系信息</text>
            <view class="contact-card">
              <view class="contact-item" hover-class="contact-press" :hover-stay-time="100" @click="openMap">
                <view class="contact-icon contact-icon--blue"><view class="ci-pin" /></view>
                <view class="contact-body">
                  <text class="contact-key">地址</text>
                  <text class="contact-val">{{ detail.location || '暂无' }}</text>
                </view>
                <view class="contact-arrow">›</view>
              </view>
              <view class="contact-item" hover-class="contact-press" :hover-stay-time="100" @click="callPhone">
                <view class="contact-icon contact-icon--green"><view class="ci-phone" /></view>
                <view class="contact-body">
                  <text class="contact-key">电话</text>
                  <text class="contact-val contact-val--link">{{ detail.phone || detail.contact_phone || '400-116-0851' }}</text>
                </view>
                <view class="contact-arrow">›</view>
              </view>
              <view class="contact-item" hover-class="contact-press" :hover-stay-time="100">
                <view class="contact-icon contact-icon--orange"><view class="ci-clock" /></view>
                <view class="contact-body">
                  <text class="contact-key">营业时间</text>
                  <text class="contact-val">{{ detail.business_hours || '周一至周日 09:00-18:00' }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 培训资格证卡 -->
          <view class="section">
            <text class="section-title">培训资格证</text>
            <view class="cert-card">
              <view class="cert-watermark">资</view>
              <view class="cert-verified"><view class="cert-check" /><text class="cert-verified-text">已认证</text></view>
              <view class="cert-center">
                <view class="cert-seal"><text class="cert-seal-char">资</text></view>
                <text class="cert-name">民用无人机驾驶员训练机构合格证</text>
              </view>
              <view class="cert-upload" hover-class="cert-upload-press" :hover-stay-time="100" @click="previewCert">
                <view class="cert-upload-ico" />
                <text class="cert-upload-text">上传证书图</text>
              </view>
              <view class="cert-link" @click="previewCert">
                <view class="cert-check cert-check--link" />
                <text class="cert-link-text">点击查看完整证书</text>
              </view>
            </view>
          </view>

          <!-- 培训环境 -->
          <view class="section section--last">
            <text class="section-title">培训环境</text>
            <view class="env-grid">
              <view
                v-for="(e, i) in envPlaceholders"
                :key="i"
                class="env-cell"
                @click="previewEnv(i)"
              >
                <image
                  v-if="envImages(detail)[i]"
                  :src="envImages(detail)[i]"
                  class="env-cell-img"
                  mode="aspectFill"
                  :style="{ opacity: imgLoaded.env[i] ? 1 : 0 }"
                  @load="onImgLoad('env', i)"
                />
                <view v-else class="env-cell-fallback">
                  <view class="env-cell-icon">{{ e.icon }}</view>
                </view>
                <view class="env-mask" />
                <text class="env-label">{{ e.name }}</text>
              </view>
            </view>
          </view>

          <view class="bottom-space" />
        </view>
      </template>
    </StateView>

    <!-- ═══ ③ 底部固定操作栏 ═══ -->
    <view v-if="detail" class="bottom-bar">
      <!-- 收藏（真实接口，登录后可用） -->
      <view class="btn-fav" :class="{ on: isFav }" hover-class="btn-fav-press" @click="toggleFav">
        <text class="fav-heart">{{ isFav ? '♥' : '♡' }}</text>
        <text class="fav-label">{{ isFav ? '已收藏' : '收藏' }}</text>
      </view>
      <view class="bottom-left">
        <text class="fee-label">培训参考价</text>
        <view class="fee-price">
          <text v-if="minPrice(detail) !== '面议'" class="fee-symbol">¥</text>
          <text class="fee-value">{{ minPrice(detail) }}</text>
          <text v-if="minPrice(detail) !== '面议'" class="fee-unit">起/人</text>
        </view>
      </view>
      <view class="bottom-actions">
        <view class="btn-outline" hover-class="btn-outline-press" :hover-stay-time="100" @click="handleConsult">
          <view class="btn-phone-ico" />
          <text class="btn-outline-text">联系咨询</text>
        </view>
        <view class="btn-primary" hover-class="btn-primary-press" :hover-stay-time="100" @click="handleEnroll">
          <text class="btn-primary-text">立即报名</text>
        </view>
      </view>
    </view>

    <!-- ═══ ④ 自定义 Toast ═══ -->
    <view v-if="toast.show" class="custom-toast" :class="{ 'custom-toast--out': toast.hide }">
      <view class="toast-icon"><view class="toast-check" /></view>
      <text class="toast-text">{{ toast.msg }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, authStorage } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
const activePriceIndex = ref(0) // 培训参考价选中项（默认第一个热销）
const imgLoaded = reactive({ banner: false, certificate: false, env: {} })

/* Toast */
const toast = ref({ show: false, hide: false, msg: '' })
let toastTimer = null
let toastOutTimer = null

/* 环境占位数据 */
const envPlaceholders = computed(function () {
  return [
    { icon: '景', name: '实操场地', color: 'blue' },
    { icon: '室', name: '理论教室', color: 'purple' },
    { icon: '飞', name: '模拟训练', color: 'orange' },
  ]
})

/* ===== 状态 ===== */
function statusText(item) {
  var map = { recruiting: '招生中', full: '已满', urgent: '名额紧张', upcoming: '即将开课' }
  return map[item.status] || '招生中'
}

/* ===== 数据映射 ===== */
function heroImage(item) {
  const u = item.banner || item.cover_image || item.image
  return u ? u : ''
}
function certificateImage(item) {
  const u = item.certificate || item.certificate_url
  return u ? u : ''
}
function orgNameOf(item) {
  return item.org_name || item.enterprise_name || item.name || '未知机构'
}
function courseTitleOf(item) {
  return item.title || item.name || '未知课程'
}
function initShort(item) {
  const n = orgNameOf(item) || ''
  if (!n || n === '未知机构') return '培'
  const strip = n.replace(/培训中心|飞行学院|分校|服务中心|培训基地|学院|中心|学校/gi, '')
  const base = strip || n
  return base.charAt(0)
}

/* 课程周期（Hero 副信息） */
function coursePeriod(item) {
  var s = fmtDate(item.start_date)
  var e = fmtDate(item.end_date)
  if (s === '待定' && e === '待定') return '课程周期待定'
  return e && e !== s ? s + ' ~ ' + e : s
}
function fmtDate(d) {
  if (!d) return '待定'
  if (String(d).indexOf('.') >= 0 || String(d).indexOf('年') >= 0) return String(d)
  return String(d).slice(0, 10)
}

/* 评分（缺失显示 —，不编造默认分） */
function ratingOf(item) {
  return item.rating != null ? item.rating : '—'
}
function starCount(item) {
  var r = Number(item.rating)
  return isNaN(r) ? 0 : Math.round(r)
}
/* 通过率（缺失显示 —） */
function passRateOf(item) {
  return item.pass_rate != null ? item.pass_rate : '—'
}
/* 机构年限（缺失显示 —） */
function yearsOf(item) {
  if (item.years != null) return item.years
  if (item.establish_year) return Math.max(1, new Date().getFullYear() - Number(item.establish_year))
  return '—'
}

/* 证书类型标签 */
function certTypeTags(item) {
  if (Array.isArray(item.course_types) && item.course_types.length > 0) return item.course_types.slice(0, 2)
  const ct = item.cert_type || 'CAAC'
  return [ct + '视距内', ct + '超视距']
}

function featureTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  return []
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
  // 有真实价格字段才展示单条；缺失返回空（模板显示"费用面议"），不编造超视距价
  if (item.price != null || item.price_fen != null) {
    const price = item.price != null ? item.price : (item.price_fen ? (item.price_fen / 100) : 0)
    const ct = item.cert_type || '课程'
    return [{ name: ct, price: price, unit: '人' }]
  }
  return []
}
/* 价格内容清单 */
function priceIncludes(i) {
  return i === 0 ? ['含教材', '含考证', '含复训'] : ['含保险', '1v1 教练']
}
function minPrice(item) {
  var arr = priceList(item)
  if (arr.length === 0) return '面议'
  var min = arr[0].price
  for (var i = 1; i < arr.length; i++) if (arr[i].price < min) min = arr[i].price
  return min
}

function orgIntro(item) {
  const intro = item.intro || item.description || ''
  return intro || ''
}

function envImages(item) {
  if (Array.isArray(item.environment) && item.environment.length > 0) return item.environment
  if (Array.isArray(item.env_images)) return item.env_images
  if (Array.isArray(item.images)) return item.images
  return []
}

/* ===== 数据获取（按 id 单查 + storage 缓存，删除 mock 兜底） ===== */
async function fetchDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    // 优先读取列表页经 storage 传入的完整数据
    var cached = uni.getStorageSync('training_course_detail')
    if (cached && String(cached.id) === String(id.value)) {
      detail.value = cached
      loading.value = false
      loadFavState()
      return
    }
    const res = await request({ url: '/api/v1/training-courses/' + encodeURIComponent(id.value) })
    const data = (res && res.data) || res || null
    detail.value = data && data.id ? data : null
    if (!detail.value) errorMsg.value = '机构不存在'
    else loadFavState()
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

/* ===== 课程收藏（真实接口，登录后可用） ===== */
const isFav = ref(false)
const favBusy = ref(false)

async function toggleFav() {
  const item = detail.value
  if (!item || !item.id) return
  if (!authStorage.getAccessToken()) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  if (favBusy.value) return
  favBusy.value = true
  try {
    const next = !isFav.value
    await request({
      url: '/api/v1/training-courses/' + encodeURIComponent(item.id) + '/favorite',
      method: 'POST',
      data: { favorite: next },
    })
    isFav.value = next
    showCustomToast(next ? '已收藏，可在「我的收藏」查看' : '已取消收藏')
  } catch (e) {
    showCustomToast('操作失败，请重试')
  } finally {
    favBusy.value = false
  }
}

// 进入页面时回显收藏状态（登录后）
async function loadFavState() {
  if (!authStorage.getAccessToken() || !detail.value || !detail.value.id) return
  try {
    const res = await request({ url: '/api/v1/training-courses/favorites/mine' })
    const list = Array.isArray(res) ? res : (res && res.data) || []
    if (Array.isArray(list)) {
      isFav.value = list.some((d) => (typeof d === 'string' ? d : d && d.id) === detail.value.id)
    }
  } catch (e) { /* 忽略：保持未收藏 */ }
}

/* ===== 交互 ===== */
function onImgLoad(key, idx) {
  if (idx !== undefined) imgLoaded.env[idx] = true
  else imgLoaded[key] = true
}
function goBack() { uni.navigateBack({ delta: 1 }) }
function openMap() {
  const addr = (detail.value && detail.value.location) || ''
  showCustomToast(addr ? '导航到：' + addr : '暂无地址信息')
}
function callPhone() {
  const phone = (detail.value && (detail.value.phone || detail.value.contact_phone)) || ''
  if (phone) uni.makePhoneCall({ phoneNumber: phone })
  else showCustomToast('暂无联系电话')
}
function handleConsult() { showCustomToast('已提交咨询，客服稍后联系') }
function handleEnroll() {
  uni.navigateTo({ url: '/pkg-talent/pages/training/register?id=' + encodeURIComponent(id.value) })
}
function handlePriceTap(p, i) {
  // 切换选中项（高亮态带动画过渡）
  if (activePriceIndex.value !== i) activePriceIndex.value = i
  showCustomToast(p.name + ' · ¥' + p.price)
}
function previewCert() {
  const url = certificateImage(detail.value)
  if (url) uni.previewImage({ urls: [url], current: url })
  else showCustomToast('证书图待上传')
}
function previewEnv(idx) {
  const imgs = envImages(detail.value)
  if (imgs.length > 0) uni.previewImage({ urls: imgs, current: imgs[idx] || imgs[0] })
}

function showCustomToast(msg) {
  clearTimeout(toastTimer)
  clearTimeout(toastOutTimer)
  toast.value = { show: true, hide: false, msg: msg }
  toastTimer = setTimeout(function () {
    toast.value.hide = true
    toastOutTimer = setTimeout(function () {
      toast.value.show = false
    }, 200)
  }, 2000)
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
  --ease: cubic-bezier(0.2, 0.8, 0.2, 1);
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
  padding-left: constant(safe-area-inset-left);
  padding-left: env(safe-area-inset-left);
  padding-right: constant(safe-area-inset-right);
  padding-right: env(safe-area-inset-right);
  overflow-x: hidden;
}

/* ═══ ① 全屏 Hero（250px） ═══ */
.hero {
  position: relative;
  width: 100%;
  height: 500rpx;
  overflow: hidden;
}
.hero-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
.hero-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #0a5897 0%, #074D92 100%);
}
.drone-svg { position: relative; width: 160rpx; height: 120rpx; opacity: 0.9; }
.drone-prop {
  position: absolute;
  width: 44rpx; height: 44rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.65);
  border-radius: 50%;
  box-sizing: border-box;
}
.drone-prop.p1 { left: 0; top: 0; }
.drone-prop.p2 { right: 0; top: 0; }
.drone-prop.p3 { left: 0; bottom: 0; }
.drone-prop.p4 { right: 0; bottom: 0; }
.drone-arm {
  position: absolute;
  left: 50%; top: 50%;
  width: 116rpx; height: 3rpx;
  background: rgba(255, 255, 255, 0.4);
}
.drone-arm.a1 { transform: translate(-50%, -50%) rotate(-45deg); }
.drone-arm.a2 { transform: translate(-50%, -50%) rotate(45deg); }
.drone-body {
  position: absolute;
  left: 50%; top: 50%;
  width: 56rpx; height: 36rpx;
  margin: -18rpx 0 0 -28rpx;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 10rpx;
}
.drone-gimbal {
  position: absolute;
  left: 50%; top: 50%;
  width: 24rpx; height: 24rpx;
  margin: 20rpx 0 0 -12rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.65);
  border-radius: 50%;
  box-sizing: border-box;
}

.hero-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(7, 77, 146, 0.35) 0%, rgba(7, 77, 146, 0.10) 30%, rgba(7, 77, 146, 0.65) 100%);
}
.hero-highlight {
  position: absolute;
  left: 80%;
  top: 10%;
  width: 900rpx;
  height: 400rpx;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.18) 0%, transparent 70%);
  transform: translate(-50%, -50%);
  pointer-events: none;
}

.hero-nav {
  position: absolute;
  top: var(--status-bar-height);
  left: 0;
  right: 0;
  padding: 8rpx 24rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 5;
}
.nav-back {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.18);
  border: 1rpx solid rgba(255, 255, 255, 0.24);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 180ms var(--ease);
}
.nav-press { background: rgba(255, 255, 255, 0.32); }
.nav-back-icon { font-size: 40rpx; color: #ffffff; font-weight: 300; line-height: 1; }
.nav-capsule {
  width: 176rpx;
  height: 60rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.24);
  border-radius: 999rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  background: rgba(255, 255, 255, 0.16);
}
.capsule-dot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #ffffff; }
.capsule-divider { width: 1rpx; height: 28rpx; background: rgba(255, 255, 255, 0.4); }
.capsule-arrow {
  width: 0; height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 8rpx solid #ffffff;
}

.status-badge {
  position: absolute;
  top: calc(var(--status-bar-height) + 80rpx);
  left: 32rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 16rpx;
  border-radius: 6rpx;
  background: #F97316;
  box-shadow: 0 4rpx 10rpx rgba(249, 115, 22, 0.32);
  z-index: 4;
}
.status-text { font-size: 20rpx; font-weight: 600; color: #ffffff; }
.status-dot {
  width: 10rpx; height: 10rpx;
  border-radius: 50%;
  background: #ffffff;
  position: relative;
}
.status-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: #ffffff;
  animation: badgeRing 1.4s ease-out infinite;
}

.hero-bottom {
  position: absolute;
  left: 32rpx;
  right: 32rpx;
  bottom: 24rpx;
  z-index: 3;
}
.hero-title {
  display: block;
  font-size: 40rpx;
  font-weight: 760;
  color: #ffffff;
  line-height: 1.35;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.32);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero-meta-row { display: flex; align-items: center; gap: 8rpx; margin-top: 12rpx; }
.hero-meta-text { font-size: 24rpx; color: rgba(255, 255, 255, 0.85); }

.meta-ico { width: 28rpx; height: 28rpx; flex-shrink: 0; position: relative; }
.meta-ico--cal {
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  border-radius: 4rpx;
  box-sizing: border-box;
}
.meta-ico--cal::before,
.meta-ico--cal::after {
  content: '';
  position: absolute;
  top: 4rpx;
  width: 2rpx; height: 5rpx;
  background: rgba(255, 255, 255, 0.9);
}
.meta-ico--cal::before { left: 6rpx; }
.meta-ico--cal::after { right: 6rpx; }
.meta-ico--cal .cal-top {
  position: absolute;
  left: 3rpx; right: 3rpx; top: 8rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.9);
}
.meta-ico--cal .cal-line {
  position: absolute;
  left: 5rpx; right: 5rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.9);
  opacity: 0.6;
}
.meta-ico--cal .cal-line.l1 { top: 14rpx; }
.meta-ico--cal .cal-line.l2 { top: 19rpx; }
.meta-ico--cal .cal-line.l3 { top: 24rpx; }

/* ═══ ② 白色内容区 ═══ */
.content {
  position: relative;
  background: #F4F6F8;
  border-radius: 28rpx 28rpx 0 0;
  margin-top: -28rpx;
  padding: 28rpx 24rpx 0;
  box-shadow: 0 -16rpx 48rpx rgba(7, 77, 146, 0.12);
  animation: contentIn 400ms var(--ease) 100ms both;
}
.section { margin-bottom: 28rpx; }
.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 760;
  color: #17212B;
  margin-bottom: 14rpx;
  padding-left: 16rpx;
  border-left: 6rpx solid #0A66C2;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14rpx;
  padding-left: 16rpx;
  border-left: 6rpx solid #0A66C2;
}
.section-head .section-title { margin-bottom: 0; padding-left: 0; border-left: none; }
.section-sub { font-size: 20rpx; color: #98A2B3; }

/* 评分卡 */
.rating-card {
  display: flex;
  align-items: center;
  background: #ffffff;
  border-radius: 24rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  padding: 28rpx;
}
.rating-left {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  padding-right: 28rpx;
  flex-shrink: 0;
}
.rating-score { font-size: 48rpx; font-weight: 760; color: #E96012; line-height: 1; }
.rating-total { font-size: 22rpx; font-weight: 500; color: #98A2B3; }
.rating-stars { display: flex; gap: 2rpx; }
.star { font-size: 22rpx; color: #E4E7EC; }
.star--on { color: #F5B301; }
.rating-divider {
  width: 1rpx;
  height: 96rpx;
  background: #E4E7EC;
  margin: 0 28rpx;
}
.rating-right { flex: 1; display: flex; flex-direction: column; gap: 12rpx; }
.r-stat { display: flex; align-items: center; justify-content: space-between; }
.r-key { font-size: 22rpx; color: #667085; }
.r-val { font-size: 22rpx; font-weight: 760; color: #0A66C2; }

/* 标签分组卡 */
.group-card {
  background: #ffffff;
  border-radius: 20rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  padding: 24rpx;
}
.group-title {
  display: block;
  font-size: 22rpx;
  font-weight: 600;
  color: #667085;
  letter-spacing: 0.4px;
  margin-bottom: 12rpx;
}
.group-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }
.g-tag {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 6rpx;
  font-size: 22rpx;
  font-weight: 600;
}
.g-tag--cert { background: #F3F0FF; color: #7C3AED; }
.g-tag--c0 { background: #E9F7F0; color: #168A55; }
.g-tag--c1 { background: #FEF6E7; color: #B54708; }
.g-tag--c2 { background: #EAF3FB; color: #0A66C2; }
.g-tag--c3 { background: #FFF0E6; color: #E96012; }
.g-tag--c4 { background: #EAF3FB; color: #0A66C2; }
.g-tag-ico { width: 26rpx; height: 26rpx; position: relative; }
.tag-drone {
  position: absolute;
  left: 50%; top: 50%;
  width: 12rpx; height: 12rpx;
  margin: -6rpx 0 0 -6rpx;
  border: 2rpx solid currentColor;
  border-radius: 3rpx;
  box-sizing: border-box;
}
.g-tag-ico--cert::before {
  content: '';
  position: absolute;
  left: -4rpx; top: 50%;
  width: 8rpx; height: 8rpx;
  margin-top: -4rpx;
  border: 2rpx solid currentColor;
  border-radius: 50%;
  box-sizing: border-box;
}
.group-divider { height: 1rpx; background: #EEF1F4; margin: 20rpx 0; }

/* 培训参考价卡 */
.price-list { display: flex; flex-direction: column; gap: 16rpx; }
.price-empty {
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 20rpx;
  padding: 28rpx;
  font-size: 26rpx;
  color: #667085;
  text-align: center;
}
.price-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  background: #ffffff;
  border: 1rpx solid #EEF1F4;
  border-radius: 20rpx;
  padding: 24rpx;
  transition: border-color 200ms var(--ease), background 200ms var(--ease), transform 180ms var(--ease), box-shadow 180ms var(--ease);
}
/* 热销项：仅保留角标，边框/底色与其他项一致，高亮全归选中态 */
.price-item--hot {
  border-color: #EEF1F4;
}
/* 选中项：橙色描边 + 浅橙底 + 高亮投影（点击切换平滑过渡） */
.price-item--active {
  border-color: #F97316;
  background: linear-gradient(180deg, #FFF0E6 0%, #ffffff 100%);
  box-shadow: 0 4rpx 14rpx rgba(249, 115, 22, 0.16);
}
/* 选中打勾 */
.price-check {
  width: 16rpx;
  height: 9rpx;
  border-left: 3rpx solid #E96012;
  border-bottom: 3rpx solid #E96012;
  transform: rotate(-45deg) translate(1rpx, -1rpx);
  margin-right: 6rpx;
}
.price-press { transform: scale(0.985); box-shadow: 0 6px 18px rgba(16, 24, 40, 0.08); }
.price-hot-badge {
  position: absolute;
  top: 0;
  right: 16rpx;
  padding: 4rpx 14rpx;
  border-radius: 0 0 8rpx 8rpx;
  background: #F97316;
  color: #ffffff;
  font-size: 18rpx;
  font-weight: 600;
}
.price-left { flex: 1; min-width: 0; }
.price-name { display: block; font-size: 28rpx; font-weight: 720; color: #17212B; }
.price-desc { display: flex; flex-wrap: wrap; gap: 4rpx 12rpx; margin-top: 10rpx; }
.inc-item { font-size: 20rpx; color: #667085; position: relative; }
.inc-item + .inc-item::before {
  content: '·';
  position: absolute;
  left: -8rpx;
  color: #98A2B3;
}
.price-right { display: flex; align-items: baseline; flex-shrink: 0; }
.price-symbol { font-size: 22rpx; font-weight: 700; color: #E96012; }
.price-value { font-size: 36rpx; font-weight: 760; color: #17212B; line-height: 1; }
.price-value--hot { color: #E96012; }
.price-unit { font-size: 20rpx; color: #98A2B3; margin-left: 4rpx; }

/* 机构简介卡 */
.intro-card {
  background: #ffffff;
  border-radius: 20rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  padding: 24rpx;
}
.intro-text {
  display: block;
  font-size: 22rpx;
  color: #667085;
  line-height: 1.6;
  white-space: pre-line;
}

/* 联系信息卡 */
.contact-card {
  background: #ffffff;
  border-radius: 20rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05);
  overflow: hidden;
}
.contact-item {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 24rpx;
  transition: background 180ms var(--ease);
}
.contact-item + .contact-item { border-top: 1rpx solid #EEF1F4; }
.contact-press { background: #F4F8FC; }
.contact-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.contact-icon--blue { background: #EAF3FB; }
.contact-icon--green { background: #E9F7F0; }
.contact-icon--orange { background: #FFF0E6; }
.ci-pin {
  width: 24rpx; height: 24rpx;
  border: 3rpx solid #0A66C2;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  box-sizing: border-box;
}
.ci-phone {
  width: 26rpx; height: 20rpx;
  border: 3rpx solid #168A55;
  border-radius: 4rpx;
  position: relative;
}
.ci-phone::before {
  content: '';
  position: absolute;
  left: 50%; top: -6rpx;
  width: 10rpx; height: 4rpx;
  background: #168A55;
  margin-left: -5rpx;
}
.ci-clock {
  width: 26rpx; height: 26rpx;
  border: 3rpx solid #E96012;
  border-radius: 50%;
  box-sizing: border-box;
  position: relative;
}
.ci-clock::before {
  content: '';
  position: absolute;
  left: 50%; top: 4rpx;
  width: 3rpx; height: 10rpx;
  background: #E96012;
  margin-left: -1.5rpx;
}
.contact-body { flex: 1; min-width: 0; }
.contact-key { display: block; font-size: 20rpx; color: #98A2B3; }
.contact-val {
  display: block;
  font-size: 26rpx;
  font-weight: 720;
  color: #17212B;
  margin-top: 4rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.contact-val--link { color: #0A66C2; }
.contact-arrow { font-size: 32rpx; color: #98A2B3; }

/* 培训资格证卡 */
.cert-card {
  position: relative;
  background: linear-gradient(135deg, #FEF6E7 0%, #fff7ed 50%, #FEF6E7 100%);
  border: 1rpx dashed #B54708;
  border-radius: 20rpx;
  padding: 36rpx;
  overflow: hidden;
}
.cert-watermark {
  position: absolute;
  right: -30rpx;
  top: -20rpx;
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
  background: rgba(181, 71, 8, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 72rpx;
  font-weight: 760;
  color: rgba(181, 71, 8, 0.2);
}
.cert-verified {
  position: absolute;
  top: 16rpx;
  right: 16rpx;
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 6rpx 14rpx;
  border-radius: 6rpx;
  background: #168A55;
  box-shadow: 0 4rpx 10rpx rgba(22, 138, 85, 0.28);
  z-index: 2;
}
.cert-check {
  width: 16rpx;
  height: 9rpx;
  border-left: 3rpx solid #ffffff;
  border-bottom: 3rpx solid #ffffff;
  transform: rotate(-45deg) translate(1rpx, -1rpx);
}
.cert-verified-text { font-size: 20rpx; font-weight: 600; color: #ffffff; }
.cert-center { display: flex; flex-direction: column; align-items: center; gap: 12rpx; position: relative; z-index: 1; }
.cert-seal {
  width: 72rpx;
  height: 72rpx;
  border-radius: 16rpx;
  background: #B54708;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cert-seal-char { font-size: 32rpx; font-weight: 700; color: #ffffff; }
.cert-name {
  display: block;
  font-size: 28rpx;
  font-weight: 720;
  color: #17212B;
  text-align: center;
}
.cert-upload {
  margin-top: 24rpx;
  height: 80rpx;
  border: 1rpx dashed #B54708;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  transition: background 180ms var(--ease);
}
.cert-upload-press { background: rgba(181, 71, 8, 0.06); }
.cert-upload-ico {
  width: 20rpx; height: 20rpx;
  border: 2rpx solid #B54708;
  border-radius: 4rpx;
  position: relative;
}
.cert-upload-ico::before {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  width: 10rpx; height: 2rpx;
  background: #B54708;
  margin: -1rpx 0 0 -5rpx;
}
.cert-upload-ico::after {
  content: '';
  position: absolute;
  left: 50%; top: 50%;
  width: 2rpx; height: 10rpx;
  background: #B54708;
  margin: -5rpx 0 0 -1rpx;
}
.cert-upload-text { font-size: 24rpx; font-weight: 600; color: #B54708; }
.cert-link {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  margin-top: 16rpx;
}
.cert-check--link { border-color: #0A66C2; }
.cert-link-text { font-size: 22rpx; font-weight: 600; color: #0A66C2; }

/* 培训环境 */
.env-grid {
  display: flex;
  gap: 16rpx;
}
.env-cell {
  flex: 1;
  position: relative;
  height: 180rpx; /* 固定高度，避免 flex 宽度下 padding 比例失真 */
  border-radius: 16rpx;
  overflow: hidden;
}
.env-cell-img {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  pointer-events: none; /* 图片不拦截点击，穿透到容器 @click */
  transition: opacity 240ms ease-out;
}
.env-cell-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #0a5897 0%, #074D92 100%);
  pointer-events: none;
}
.env-cell-icon { font-size: 44rpx; font-weight: 700; color: rgba(255, 255, 255, 0.9); }
.env-mask {
  position: absolute;
  left: 0; right: 0; bottom: 0;
  height: 50%;
  background: linear-gradient(180deg, rgba(7, 77, 146, 0) 50%, rgba(7, 77, 146, 0.4) 100%);
  pointer-events: none;
}
.env-label {
  position: absolute;
  left: 50%;
  bottom: 12rpx;
  transform: translateX(-50%);
  padding: 2rpx 12rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.88);
  font-size: 20rpx;
  font-weight: 500;
  color: #344054;
  white-space: nowrap;
  pointer-events: none;
}
.bottom-space { height: 20rpx; }

/* ═══ ③ 底部固定操作栏 ═══ */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid #EEF1F4;
  padding: 16rpx 24rpx calc(16rpx + env(safe-area-inset-bottom));
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  z-index: 50;
}
.bottom-left { display: flex; align-items: baseline; gap: 8rpx; flex-shrink: 0; }

/* 收藏按钮（底部栏最左，图标+文字纵向） */
.btn-fav {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2rpx;
  padding: 0 12rpx;
  margin-right: 4rpx;
  flex-shrink: 0;
  border-radius: 10rpx;
  transition: opacity 160ms var(--ease);
}
.btn-fav-press { opacity: 0.65; }
.fav-heart { font-size: 34rpx; line-height: 1.1; color: #667085; }
.fav-label { font-size: 18rpx; color: #667085; }
.btn-fav.on .fav-heart { color: #E96012; }
.btn-fav.on .fav-label { color: #E96012; font-weight: 600; }
.fee-label { font-size: 20rpx; color: #98A2B3; }
.fee-price { display: flex; align-items: baseline; }
.fee-symbol { font-size: 22rpx; font-weight: 700; color: #E96012; }
.fee-value {
  font-size: 44rpx;
  font-weight: 760;
  color: #E96012;
  line-height: 1;
  animation: priceIn 500ms var(--ease) both;
}
.fee-unit { font-size: 20rpx; color: #98A2B3; margin-left: 4rpx; }
.bottom-actions { display: flex; gap: 12rpx; flex: 1; justify-content: flex-end; }
.btn-outline {
  display: flex;
  align-items: center;
  gap: 8rpx;
  height: 76rpx;
  padding: 0 24rpx;
  border: 1rpx solid #0A66C2;
  border-radius: 10rpx;
  background: transparent;
  transition: transform 180ms var(--ease), background 180ms var(--ease);
}
.btn-outline-press { background: #EAF3FB; }
.btn-phone-ico {
  width: 20rpx;
  height: 20rpx;
  border: 2rpx solid #0A66C2;
  border-radius: 4rpx;
  position: relative;
}
.btn-phone-ico::before {
  content: '';
  position: absolute;
  left: 50%; top: -3rpx;
  width: 6rpx; height: 3rpx;
  background: #0A66C2;
  margin-left: -3rpx;
}
.btn-outline-text { font-size: 24rpx; font-weight: 700; color: #0A66C2; }
.btn-primary {
  height: 76rpx;
  padding: 0 32rpx;
  border-radius: 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #F97316;
  box-shadow: 0 4rpx 10rpx rgba(249, 115, 22, 0.32);
  transition: transform 180ms var(--ease), background 180ms var(--ease);
}
.btn-primary:active { background: #E96012; }
.btn-primary-press { transform: scale(0.97); }
.btn-primary-text { font-size: 28rpx; font-weight: 700; color: #ffffff; }

/* ═══ ④ 自定义 Toast ═══ */
.custom-toast {
  position: fixed;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  z-index: 999;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 32rpx;
  background: rgba(16, 24, 40, 0.92);
  border-radius: 10rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.24);
  animation: toastIn 250ms var(--ease) both;
  max-width: 70vw;
}
.custom-toast--out { animation: toastOut 200ms ease both; }
.toast-icon {
  width: 32rpx; height: 32rpx;
  border-radius: 50%;
  background: rgba(91, 255, 176, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.toast-check {
  width: 16rpx; height: 9rpx;
  border-left: 3rpx solid #5BFFB0;
  border-bottom: 3rpx solid #5BFFB0;
  transform: rotate(-45deg) translate(1rpx, -1rpx);
}
.toast-text { font-size: 26rpx; color: #ffffff; font-weight: 500; line-height: 1.4; }

/* ═══ 动画 ═══ */
@keyframes contentIn {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes badgeRing {
  0% { transform: scale(1); opacity: 0.8; }
  80% { transform: scale(2.4); opacity: 0; }
  100% { transform: scale(2.4); opacity: 0; }
}
@keyframes priceIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes toastIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% - 20rpx)); }
  to { opacity: 1; transform: translate(-50%, -50%); }
}
@keyframes toastOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* ═══ 减少动态效果支持 ═══ */
@media (prefers-reduced-motion: reduce) {
  .content,
  .status-dot::after,
  .fee-value,
  .custom-toast {
    animation: none !important;
    transition: none !important;
  }
}
</style>
