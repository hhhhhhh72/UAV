<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="课程不存在或已下架"
      @retry="fetchDetail"
    >
      <template v-if="detail">
        <!-- 顶部导航与课程封面分层，避免文字与系统控件压在同一张图上 -->
        <view class="detail-nav" :style="{ paddingTop: statusBarH + 'px' }">
          <view class="detail-nav-back" hover-class="detail-nav-back--press" :hover-stay-time="100" @click="goBack">
            <text>‹</text>
          </view>
          <text class="detail-nav-title">培训详情</text>
          <view class="detail-nav-balance" />
        </view>

        <!-- ① 课程封面 -->
        <view class="hero">
          <image v-if="heroImage(detail)" :src="heroImage(detail)" mode="aspectFill" class="hero-img" lazy-load />
          <view v-else class="hero-fallback">
            <view class="hero-fallback-icon"><image src="/static/home/icons/training.svg" mode="aspectFit" /></view>
          </view>

          <view class="hero-mask" />
          <!-- 课程状态 -->
          <view class="status-badge">
            <text class="status-text">{{ statusText(detail) }}</text>
          </view>

          <!-- Hero 底部信息区：课程名（主）+ 机构名（副）+ 课程周期 -->
          <view class="hero-bottom">
            <text class="hero-title">{{ courseTitleOf(detail) }}</text>
            <text class="hero-org">{{ orgNameOf(detail) }}</text>
            <view class="hero-meta-row">
              <view class="meta-ico meta-ico--cal"><view class="cal-top" /><view class="cal-body"><view class="cal-line l1" /><view class="cal-line l2" /><view class="cal-line l3" /></view></view>
              <text class="hero-meta-text">{{ coursePeriod(detail) }}</text>
            </view>
          </view>
        </view>

        <!-- ═══ ② 白色内容区 ═══ -->
        <view class="content">
          <!-- 课程关键信息：只展示训练课程实体已有字段，报名决策前不必先翻找详情 -->
          <view class="course-info-card">
            <view class="course-info-head">
              <text class="course-info-title">课程信息</text>
              <text class="course-info-status">{{ statusText(detail) }}</text>
            </view>
            <view class="course-info-grid">
              <view class="course-info-item">
                <text class="course-info-label">开课时间</text>
                <text class="course-info-value">{{ coursePeriod(detail) }}</text>
              </view>
              <view class="course-info-item">
                <text class="course-info-label">培训时长</text>
                <text class="course-info-value">{{ courseDuration(detail) }}</text>
              </view>
              <view class="course-info-item">
                <text class="course-info-label">培训地点</text>
                <text class="course-info-value">{{ courseLocation(detail) }}</text>
              </view>
              <view class="course-info-item">
                <text class="course-info-label">报名情况</text>
                <text class="course-info-value">{{ seatText(detail) }}</text>
              </view>
            </view>
          </view>

          <!-- 平台背书（首屏信任）：协会平台审核 · 机构资质已核验 -->
          <view class="platform-endorse">
            <view class="pe-mark"><view class="pe-check" /></view>
            <view class="pe-body">
              <text class="pe-title">协会平台审核 · 机构资质已核验</text>
              <text class="pe-sub">平台认证机构 · 报名费用公开透明</text>
            </view>
          </view>

          <!-- 评价数据为空时不渲染，避免以“—”占据首屏 -->
          <view v-if="hasRatingInfo(detail)" class="section">
            <view class="rating-card">
              <view class="rating-left">
                <text class="rating-score">{{ ratingOf(detail) }}<text class="rating-total">/5.0</text></text>
                <view class="rating-stars">
                  <view v-for="n in 5" :key="n" class="star" :class="{ 'star--on': n <= starCount(detail) }" />
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
                <text class="group-title">服务与保障</text>
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

          <!-- 课程介绍卡 -->
          <view v-if="courseDescription(detail)" class="section">
            <text class="section-title">课程介绍</text>
            <view class="intro-card">
              <text class="intro-text">{{ courseDescription(detail) }}</text>
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
                  <text class="contact-val contact-val--link">{{ displayPhone }}</text>
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
              <!-- 认证徽章：有证书图 → 已认证；无图 → 待上传（不再无图也展示"已认证"） -->
              <view v-if="certificateImage(detail)" class="cert-verified"><view class="cert-check" /><text class="cert-verified-text">已认证</text></view>
              <view v-else class="cert-verified cert-verified--pending"><view class="cert-check" /><text class="cert-verified-text">证书待上传</text></view>
              <view class="cert-center">
                <view class="cert-seal"><text class="cert-seal-char">资</text></view>
                <text class="cert-name">民用无人机驾驶员训练机构合格证</text>
              </view>
              <!-- 上传证书图：仅机构/管理员可见（学生端不暴露管理动作） -->
              <view v-if="isManageRole()" class="cert-upload" hover-class="cert-upload-press" :hover-stay-time="100" @click="previewCert">
                <view class="cert-upload-ico" />
                <text class="cert-upload-text">上传证书图</text>
              </view>
              <view class="cert-link" @click="previewCert">
                <view class="cert-check cert-check--link" />
                <text class="cert-link-text">{{ certificateImage(detail) ? '点击查看完整证书' : '证书上传后即可查看' }}</text>
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
        <view class="fav-heart" :class="{ 'fav-heart--on': isFav }" />
        <text class="fav-label">{{ isFav ? '已收藏' : '收藏' }}</text>
      </view>
      <view class="bottom-left">
        <text class="fee-label">当前选择</text>
        <view class="fee-price">
          <text v-if="bottomPrice !== '面议'" class="fee-symbol">¥</text>
          <text class="fee-value">{{ bottomPrice }}</text>
          <text v-if="bottomPrice !== '面议'" class="fee-unit">/人</text>
        </view>
      </view>
      <view class="bottom-actions">
        <view class="btn-outline" hover-class="btn-outline-press" :hover-stay-time="100" @click="handleConsult">
          <view class="btn-phone-ico" />
          <text class="btn-outline-text">联系咨询</text>
        </view>
        <view
          class="btn-primary"
          :class="{ 'btn-primary--disabled': enrollDisabled() }"
          hover-class="btn-primary-press"
          :hover-stay-time="100"
          @click="enrollDisabled() ? onEnrollBlocked() : handleEnroll()"
        >
          <text class="btn-primary-text">{{ enrollLabel() }}</text>
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
import { safeBack } from '../../../utils/nav'
import { ref, reactive, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, authStorage } from '../../../utils/request'
import StateView from '../../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)
// 状态栏高度（Hero 导航避让：微信端 CSS var(--status-bar-height) 不可靠，改用 JS 值）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
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

/* 管理侧角色（机构/平台/协会管理员）：证书上传等管理动作仅其可见 */
function isManageRole() {
  try {
    const u = uni.getStorageSync('user')
    if (!u) return false
    const role = u.role || (u.user && u.user.role)
    return role === 'enterprise' || role === 'platform_admin' || role === 'association_admin'
  } catch (e) { return false }
}

/* 课程周期（Hero 副信息） */
function coursePeriod(item) {
  var s = fmtDate(item.start_date)
  var e = fmtDate(item.end_date)
  if (s === '待定' && e === '待定') return '课程周期待定'
  return e && e !== s ? s + ' ~ ' + e : s
}
function courseDuration(item) {
  var days = Number(item && item.duration_days)
  return days > 0 ? days + ' 天' : '时长待定'
}
function courseLocation(item) {
  var district = item && item.district
  var location = item && item.location
  if (district && location) return String(district) + ' · ' + String(location)
  return district || location || '地点待定'
}
function seatText(item) {
  var capacity = Number(item && item.max_students)
  var enrolled = Number(item && item.enrolled_count)
  if (capacity > 0) return '已报 ' + (enrolled > 0 ? enrolled : 0) + ' / ' + capacity
  if (item && item.remain != null && Number(item.remain) >= 0) return '剩余 ' + item.remain + ' 个名额'
  return '名额待定'
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
function hasRatingInfo(item) {
  return Number(item && item.rating) > 0 || Number(item && item.review_count) > 0 ||
    (item && item.pass_rate != null) || (item && (item.years != null || item.establish_year))
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
function minPrice(item) {
  var arr = priceList(item)
  if (arr.length === 0) return '面议'
  var min = arr[0].price
  for (var i = 1; i < arr.length; i++) if (arr[i].price < min) min = arr[i].price
  return min
}

function courseDescription(item) {
  const intro = item.intro || item.description || ''
  return intro || ''
}

function envImages(item) {
  if (Array.isArray(item.environment) && item.environment.length > 0) return item.environment
  if (Array.isArray(item.env_images)) return item.env_images
  if (Array.isArray(item.images)) return item.images
  return []
}

/* ===== 数据获取（按 id 单查 + storage 仅作首帧缓存，删除 mock 兜底） ===== */
async function fetchDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    // storage 缓存仅作首帧展示；无论是否命中都发起接口刷新（价格/状态以服务端为准，
    // 此前命中缓存直接返回且永不清除，管理员改价/改状态后本地展示陈旧）。
    var cached = uni.getStorageSync('training_course_detail')
    if (cached && String(cached.id) === String(id.value)) {
      detail.value = cached
      loadFavState()
    }
    uni.removeStorageSync('training_course_detail')
    const res = await request({ url: '/api/v1/training-courses/' + encodeURIComponent(id.value) })
    const data = (res && res.data) || res || null
    detail.value = data && data.id ? data : null
    if (!detail.value) errorMsg.value = '课程不存在或已下架'
    else loadFavState()
  } catch (e) {
    // 接口失败且已有缓存首帧时保留展示，不覆盖为错误态
    if (!detail.value) errorMsg.value = '加载失败，请稍后重试'
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
function goBack() { safeBack() }
function openMap() {
  // 地图真动作：有经纬度 → 原生导航；仅地址 → 复制地址（可直接粘贴到导航 App）；两者皆无 → 如实提示
  const item = detail.value
  const addr = (item && item.location) || ''
  const lat = Number(item && (item.latitude != null ? item.latitude : item.lat))
  const lng = Number(item && (item.longitude != null ? item.longitude : item.lng))
  if (lat && lng) {
    uni.openLocation({ latitude: lat, longitude: lng, name: orgNameOf(item), address: addr })
    return
  }
  if (addr) {
    uni.setClipboardData({
      data: addr,
      success: function () { showCustomToast('地址已复制，可在导航 App 中搜索') },
    })
    return
  }
  showCustomToast('暂无地址信息')
}
function callPhone() {
  // 显示与拨打同源：卡上展示什么号码就拨什么（无机构号码时拨平台热线，绝不"显示却拨不了"）
  uni.makePhoneCall({ phoneNumber: displayPhone.value })
}
// 展示/拨打统一口径：机构电话优先，缺失时平台热线 400-116-0851
const displayPhone = computed(() => {
  const item = detail.value
  const p = item && (item.phone || item.contact_phone)
  return String(p || '400-116-0851')
})
// 底栏价格 = 当前选中的档位（选高档不再显示最低价"起"；无档位显示面议）
const bottomPrice = computed(() => {
  const list = priceList(detail.value)
  if (!list.length) return '面议'
  const p = list[activePriceIndex.value]
  return p ? String(p.price) : String((list[0] && list[0].price) || '面议')
})
function handleConsult() {
  // 咨询真化：统一拨 displayPhone（机构号或平台热线，与联系卡展示一致）
  uni.makePhoneCall({ phoneNumber: displayPhone.value })
}
function handleEnroll() {
  // 价格档传递：选中档随跳转写入 storage，register 回填匹配项（不再重选）
  try {
    const list = priceList(detail.value)
    const p = list[activePriceIndex.value]
    uni.setStorageSync('training_course_price', p ? { name: p.name, price: p.price, unit: p.unit } : null)
  } catch (e) { /* 存储失败不阻断跳转 */ }
  uni.navigateTo({ url: '/pkg-talent/pages/training/register?id=' + encodeURIComponent(id.value) })
}
/* 报名状态门禁（与列表页 courses 一致）：已满/即将开课不可报名 */
function enrollDisabled() {
  const s = detail.value && detail.value.status
  return s === 'full' || s === 'upcoming'
}
function enrollLabel() {
  const s = detail.value && detail.value.status
  if (s === 'full') return '本期已满 · 下期可约'
  if (s === 'upcoming') return '即将开课'
  return '立即报名'
}
function onEnrollBlocked() {
  const s = detail.value && detail.value.status
  if (s === 'full') showCustomToast('本期名额已满，可关注下一期开班')
  else showCustomToast('本课程即将开放报名，敬请期待')
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
.status-badge {
  position: absolute;
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
.hero-org {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.78);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.hero-meta-row { display: flex; align-items: center; gap: 8rpx; margin-top: 10rpx; }
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

/* 课程信息：将日期、时长、地点、报名情况前置，避免详情首屏出现无数据评分 */
.course-info-card {
  margin-bottom: 28rpx;
  padding: 24rpx;
  border: 1rpx solid #E4EAF2;
  border-radius: 20rpx;
  background: #ffffff;
  box-shadow: 0 6rpx 18rpx rgba(16, 24, 40, 0.06);
}
.course-info-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18rpx; }
.course-info-title { font-size: 30rpx; font-weight: 760; color: #17212B; }
.course-info-status { padding: 5rpx 12rpx; border-radius: 999rpx; background: #FFF0E6; color: #E96012; font-size: 20rpx; font-weight: 600; }
.course-info-grid { display: flex; flex-wrap: wrap; overflow: hidden; border: 1rpx solid #EEF1F4; border-radius: 14rpx; }
.course-info-item { width: 50%; min-width: 0; min-height: 102rpx; display: flex; flex-direction: column; justify-content: center; gap: 8rpx; padding: 14rpx 16rpx; box-sizing: border-box; }
.course-info-item:nth-child(odd) { border-right: 1rpx solid #EEF1F4; }
.course-info-item:nth-child(-n + 2) { border-bottom: 1rpx solid #EEF1F4; }
.course-info-label { font-size: 20rpx; color: #7A8798; }
.course-info-value { overflow: hidden; font-size: 23rpx; font-weight: 600; color: #344054; line-height: 1.35; white-space: nowrap; text-overflow: ellipsis; }

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
.star {
  width: 28rpx;
  height: 28rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23D8DCE3'%3E%3Cpath d='M12 2.4l2.94 6.04 6.66.92-4.85 4.66 1.18 6.6L12 17.42l-5.93 3.2 1.18-6.6L2.4 9.36l6.66-.92z'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}
.star--on {
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23F5B301'%3E%3Cpath d='M12 2.4l2.94 6.04 6.66.92-4.85 4.66 1.18 6.6L12 17.42l-5.93 3.2 1.18-6.6L2.4 9.36l6.66-.92z'/%3E%3C/svg%3E");
}
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
/* 证书未上传：灰色待上传态（澄清：无图不再伪装"已认证"） */
.cert-verified--pending {
  background: #98A2B3;
  box-shadow: 0 4rpx 10rpx rgba(152, 162, 179, 0.28);
}
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
/* CSS 绘制心形（替代 ♥/♡ 字符，符合项目"不用 emoji/Unicode 图标"规范） */
.fav-heart {
  width: 34rpx;
  height: 32rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%2398A2B3'%3E%3Cpath d='M12 21s-7.5-4.9-9.7-9.2C.6 8.4 2.6 4.5 6.3 4.5c2.2 0 3.9 1.2 4.7 3h2c.8-1.8 2.5-3 4.7-3 3.7 0 5.7 3.9 4 7.3C19.5 16.1 12 21 12 21z'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}
.fav-heart--on {
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23E96012'%3E%3Cpath d='M12 21s-7.5-4.9-9.7-9.2C.6 8.4 2.6 4.5 6.3 4.5c2.2 0 3.9 1.2 4.7 3h2c.8-1.8 2.5-3 4.7-3 3.7 0 5.7 3.9 4 7.3C19.5 16.1 12 21 12 21z'/%3E%3C/svg%3E");
}
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
  width: 28rpx;
  height: 28rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%230A66C2' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1 1 .4 2 .7 2.9a2 2 0 0 1-.5 2.1L8.1 9.9a16 16 0 0 0 6 6l1.2-1.2a2 2 0 0 1 2.1-.5c.9.3 1.9.6 2.9.7A2 2 0 0 1 22 16.9z'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
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
/* 已满/即将开课：置灰不可用（与列表页 courses 状态门禁一致） */
.btn-primary--disabled {
  background: #C9CDD4;
  box-shadow: none;
}
.btn-primary--disabled .btn-primary-text { color: #ffffff; }

/* ═══ 视觉重整：对齐小程序现有的浅蓝底 + 高饱和蓝色卡片语言 ═══ */
.page { background: #EAF4FF; }
/* 平台背书（首屏信任条：协会平台审核） */
.platform-endorse {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin: 0 24rpx 20rpx;
  padding: 20rpx 24rpx;
  background: #E7F1FC;
  border: 2rpx solid #C9DFF5;
  border-radius: 20rpx;
}
.pe-mark {
  width: 48rpx;
  height: 48rpx;
  flex: 0 0 48rpx;
  border-radius: 50%;
  background: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
}
.pe-check {
  width: 24rpx;
  height: 14rpx;
  border-left: 4rpx solid #fff;
  border-bottom: 4rpx solid #fff;
  transform: rotate(-45deg) translate(1rpx, -2rpx);
}
.pe-body { display: flex; flex-direction: column; gap: 4rpx; }
.pe-title { font-size: 26rpx; font-weight: 700; color: #0A66C2; }
.pe-sub { font-size: 22rpx; color: #4A6E94; }
.detail-nav { position: relative; height: 88rpx; display: flex; align-items: center; justify-content: space-between; padding-left: 24rpx; padding-right: 24rpx; box-sizing: content-box; background: #EAF4FF; }
.detail-nav-back, .detail-nav-balance { width: 60rpx; height: 60rpx; flex: 0 0 60rpx; }
.detail-nav-back { display: flex; align-items: center; justify-content: center; border-radius: 50%; background: #ffffff; box-shadow: 0 6rpx 16rpx rgba(31, 89, 169, 0.13); }
.detail-nav-back--press { transform: scale(0.94); opacity: 0.86; }
.detail-nav-back text { margin-top: -4rpx; color: #1A3353; font-size: 42rpx; line-height: 1; }
.detail-nav-title { position: absolute; left: 50%; transform: translateX(-50%); max-width: 56%; display: block; color: #17212B; font-size: 34rpx; font-weight: 700; text-align: center; line-height: 88rpx; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.hero { width: auto; height: 348rpx; margin: 0 24rpx; border-radius: 24rpx; box-shadow: 0 14rpx 34rpx rgba(31, 89, 169, 0.2); }
.hero-fallback { background: linear-gradient(145deg, #163C66 0%, #0A66C2 100%); }
.hero-fallback-icon { width: 112rpx; height: 112rpx; display: flex; align-items: center; justify-content: center; border-radius: 30rpx; background: rgba(255, 255, 255, 0.14); }
.hero-fallback-icon image { width: 64rpx; height: 64rpx; }
.hero-mask { background: linear-gradient(180deg, rgba(4, 30, 68, 0.08) 0%, rgba(4, 30, 68, 0.05) 34%, rgba(4, 30, 68, 0.8) 100%); }
.status-badge { top: 18rpx; left: 18rpx; padding: 7rpx 14rpx; border-radius: 999rpx; background: rgba(255, 255, 255, 0.92); box-shadow: none; }
.status-text { color: #0A66C2; font-size: 20rpx; font-weight: 650; }
.hero-bottom { left: 24rpx; right: 24rpx; bottom: 24rpx; }
.hero-title { font-size: 36rpx; line-height: 1.3; text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.18); }
.hero-org { margin-top: 8rpx; color: rgba(255, 255, 255, 0.84); }
.hero-meta-row { margin-top: 12rpx; }
.content { margin-top: 0; padding: 24rpx 24rpx 0; border-radius: 0; background: transparent; box-shadow: none; animation: none; }
.section { margin-bottom: 28rpx; }
.section-title { margin-bottom: 16rpx; padding-left: 0; border-left: 0; font-size: 30rpx; letter-spacing: 0; }
.section-head { margin-bottom: 16rpx; padding-left: 0; border-left: 0; }
.course-info-card { margin-bottom: 34rpx; padding: 22rpx; border-color: #E8EDF3; border-radius: 16rpx; background: #F8FAFC; box-shadow: none; }
.course-info-title { font-size: 29rpx; }
.course-info-status { background: #EAF3FB; color: #0A66C2; }
.course-info-grid { border-color: #E6ECF3; border-radius: 12rpx; background: #ffffff; }
.course-info-item:nth-child(odd), .course-info-item:nth-child(-n + 2) { border-color: #E6ECF3; }
.course-info-label { color: #8A94A4; }
.course-info-value { font-size: 22rpx; color: #344054; }
.rating-card, .group-card, .intro-card, .contact-card { border: 1rpx solid #E8EDF3; border-radius: 16rpx; box-shadow: none; }
.rating-card { background: #F8FAFC; }
.group-card, .intro-card, .contact-card { background: #ffffff; }
.price-item { border-color: #E8EDF3; border-radius: 16rpx; padding: 22rpx; box-shadow: none; }
.price-item--active { border-color: #0A66C2; background: #F4F8FC; box-shadow: none; }
.price-hot-badge { background: #0A66C2; }
.price-check { border-color: #0A66C2; }
.price-value, .price-value--hot { color: #0A66C2; }
.price-symbol { color: #0A66C2; }
.price-empty { border-radius: 16rpx; background: #F8FAFC; }
.contact-item { padding: 22rpx 24rpx; }
.cert-card { border-radius: 16rpx; box-shadow: none; }
.env-cell { border-radius: 12rpx; }
.bottom-bar { border-top: 1rpx solid #E8EDF3; box-shadow: 0 -6rpx 18rpx rgba(16, 24, 40, 0.06); }
.fee-symbol, .fee-value { color: #0A66C2; animation: none; }
.btn-primary { border-radius: 12rpx; background: #0A66C2; box-shadow: none; }
.btn-primary:active { background: #0759AA; }
.btn-outline { border-radius: 12rpx; }

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
  .custom-toast {
    animation: none !important;
    transition: none !important;
  }
}
</style>
