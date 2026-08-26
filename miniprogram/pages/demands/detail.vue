<template>
  <view class="detail-page">
    <!-- 头部 -->
    <view class="page-header" :style="headerStyle">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">{{ detailTitle }}</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 加载状态 -->
    <view v-if="state === 'loading'" class="skeleton-wrap">
      <view class="skeleton skeleton-title"></view>
      <view class="skeleton skeleton-line"></view>
      <view class="skeleton skeleton-block"></view>
    </view>

    <!-- 错误状态 -->
    <view v-else-if="state === 'error'" class="state-panel">
      <view class="state-mark err">!</view>
      <text class="state-title">内容加载失败</text>
      <text class="state-desc">请检查网络后重新加载</text>
      <view class="state-btn" @tap="loadDetail">重新加载</view>
    </view>

    <view v-else-if="item" class="detail-body">
      <!-- 状态 / 分类 / 标题 -->
      <view class="detail-hero">
        <view class="tag-row">
          <text class="tag blue">{{ item.cat }}</text>
          <text class="tag" :class="statusTagClass">{{ item.status }}</text>
          <text v-if="isEndedItem" class="tag red">已停止对接</text>
        </view>
        <text class="detail-title">{{ item.title }}</text>
        <view class="detail-sub">
          <text>{{ item.region }}</text>
          <text>{{ item.time }}</text>
        </view>
        <view class="detail-grid">
          <view class="grid-cell">
            <text class="grid-value price">{{ item.price }}</text>
            <text class="grid-label">{{ item.unit }}</text>
          </view>
          <view class="grid-cell">
            <text class="grid-value">{{ deadlineText }}</text>
            <text class="grid-label">{{ deadlineLabel }}</text>
          </view>
          <view class="grid-cell">
            <text class="grid-value">{{ shortCompany }}</text>
            <text class="grid-label">发布企业</text>
          </view>
        </view>
      </view>

      <!-- 描述 -->
      <view class="detail-section">
        <text class="section-title">{{ descTitle }}</text>
        <text class="desc-text">{{ item.desc }}</text>
      </view>

      <!-- 字段化信息 -->
      <view class="detail-section">
        <text class="section-title">{{ fieldsTitle }}</text>
        <view class="detail-fields">
          <view v-for="(f, i) in item.fields" :key="i" class="field-cell">
            <text class="field-label">{{ f[0] }}</text>
            <text class="field-value">{{ f[1] }}</text>
          </view>
        </view>
      </view>

      <!-- 附件 / 案例 -->
      <view class="detail-section">
        <text class="section-title">{{ mediaTitle }}</text>
        <scroll-view scroll-x class="media-row" :show-scrollbar="false">
          <view class="media-inner">
            <image
              v-for="(img, i) in mediaImages"
              :key="i"
              :src="img"
              mode="aspectFill"
              class="media-img"
              @tap="previewImage(i)"
            />
          </view>
        </scroll-view>
        <view v-if="attachmentNames.length" class="attach-list">
          <view v-for="(a, i) in attachmentNames" :key="i" class="attach-item">
            <view class="attach-icon">▣</view>
            <text class="attach-name">{{ a }}</text>
          </view>
        </view>
      </view>

      <!-- 发布企业（有已认证企业才显示认证声明；无则按个人发布展示，杜绝虚假认证） -->
      <view class="detail-section">
        <text class="section-title">发布方</text>
        <view class="company-row">
          <view class="company-avatar">
            <text>{{ companyInitial }}</text>
          </view>
          <view class="company-copy">
            <text class="company-name">{{ publisherName }}</text>
            <text v-if="publisherEnterprise" class="company-tag">已完成企业认证 · 信息经平台审核</text>
            <text v-else class="company-tag company-tag--plain">个人发布 · 信息由发布者提供</text>
          </view>
          <view v-if="publisherEnterprise" class="company-link" @tap="toastCompany">
            <text>企业主页 ›</text>
          </view>
        </view>
      </view>

      <!-- 推荐 -->
      <view class="detail-section">
        <view class="recommend-head">
          <text class="section-title">为你推荐</text>
          <view class="recommend-more" @tap="goMatches">
            <text>查看全部</text>
            <text class="more-arrow">›</text>
          </view>
        </view>
        <scroll-view scroll-x class="recommend-row" :show-scrollbar="false">
          <view class="recommend-inner">
            <view
              v-for="r in recommendItems"
              :key="r.id"
              class="recommend-card"
              hover-class="tap-fade"
              @tap="goDetail(r)"
            >
              <text class="recommend-title">{{ r.title }}</text>
              <text class="recommend-meta">{{ r.region }} · {{ r.price }}</text>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 底部操作栏（贴合屏幕底部，含安全区） -->
      <view class="fixed-actions">
        <view class="action-secondary" @tap="onFavorite">
          <text class="fav-icon" :class="{ on: favorited }">{{ favorited ? '♥' : '♡' }}</text>
          <text>{{ favorited ? '已收藏' : '收藏' }}</text>
        </view>
        <button class="action-secondary share-btn" open-type="share">
          <text class="share-icon">↗</text>
          <text>分享</text>
        </button>
        <view v-if="isEndedItem" class="action-primary disabled">
          <text>该信息已结束</text>
        </view>
        <view v-else-if="isMyDemand" class="action-primary disabled">
          <text>这是您发布的需求</text>
        </view>
        <view v-else-if="intented" class="action-primary disabled" @tap="onIntent">
          <text>已登记</text>
        </view>
        <view v-else class="action-primary" @tap="onIntent">
          <text>登记对接</text>
        </view>
      </view>
    </view>

    <!-- ═══════ 登录 / 认证 / 登记意向 弹层 ═══════ -->
    <u-popup :show="sheet.show" position="bottom" round @close="closeSheet">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">{{ sheet.title }}</text>
          <view class="sheet-close" @tap="closeSheet"><text class="sheet-x">×</text></view>
        </view>

        <view class="sheet-body">
          <!-- 登录引导 -->
          <template v-if="sheet.kind === 'login'">
            <text class="sheet-desc">发布、收藏和登记对接需要登录，登录后完成企业认证或飞手认证（任一）即可建立正式对接。</text>
            <view class="sheet-actions">
              <view class="ghost-btn" @tap="closeSheet">暂不登录</view>
              <view class="primary-btn" @tap="goLogin">去登录</view>
            </view>
          </template>

          <!-- 认证引导 -->
          <template v-else-if="sheet.kind === 'cert'">
            <text class="sheet-desc">为保障供需双方信息真实，登记对接前需完成企业认证或飞手认证（任一即可）。</text>
            <view class="sheet-actions">
              <view class="ghost-btn" @tap="closeSheet">稍后认证</view>
            </view>
            <view class="sheet-actions">
              <view class="primary-btn" @tap="goPilotCert">我是个人飞手 · 去飞手认证</view>
              <view class="secondary-btn" @tap="goCert">我是企业 · 去企业认证</view>
            </view>
          </template>

          <!-- 登记意向 -->
          <template v-else-if="sheet.kind === 'intent'">
            <text class="sheet-desc">发布方将看到对接主体、联系人和对接说明，联系方式不会在公开页面展示。</text>
            <text class="field-label">对接主体</text>
            <view class="field-static">{{ enterpriseName }}</view>
            <text class="field-label">联系人 <text class="req">*</text></text>
            <input v-model="intentForm.name" class="field" placeholder="请输入联系人" />
            <text class="field-label">联系电话 <text class="req">*</text></text>
            <input v-model="intentForm.phone" class="field" type="number" placeholder="请输入联系电话" />
            <text class="field-label">能力说明 / 备注 <text class="req">*</text></text>
            <textarea v-model="intentForm.note" class="textarea" placeholder="简要说明可提供的能力、档期或合作意向"></textarea>
            <view class="agree-row" @tap="intentForm.agree = !intentForm.agree">
              <view class="agree-box" :class="{ on: intentForm.agree }">
                <text v-if="intentForm.agree" class="agree-check">✓</text>
              </view>
              <text class="agree-text">我已阅读并同意向发布方提供以上对接信息。</text>
            </view>
            <view class="sheet-actions">
              <view class="ghost-btn" @tap="closeSheet">取消</view>
              <view class="primary-btn" :class="{ disabled: submitting }" @tap="submitIntent">
                <text>{{ submitting ? '提交中...' : '提交意向' }}</text>
              </view>
            </view>
          </template>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import { safeNavigateTo, safeBack } from '../../utils/nav'
import {
  IMG_SOLAR, IMG_LIFT, IMG_HERO, isEnded, normalizeDemand, normalizeService,
  getKindItems, isLoggedIn, currentUserName, saveSentIntents, getSentIntents,
  publishPostToCard,
} from '../../utils/hallData'
import { getPosts } from '../../utils/publishData'
import { useSafeTop } from '../../utils/safeTop'

const item = ref(null)
const state = ref('loading') // loading | ready | error
const favorited = ref(false)
const favoriting = ref(false)
const submitting = ref(false)

// 自定义导航：状态栏留白 + 右上角避让微信胶囊（JS 方式）
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)
const headerStyle = computed(() => ({
  paddingTop: (topPad.value || statusBarH.value) + 'px',
  height: (56 + (topPad.value || statusBarH.value)) + 'px',
}))
let postId = ''

const isEndedItem = computed(() => (item.value ? isEnded(item.value) : false))

const detailTitle = computed(() => {
  if (!item.value) return '详情'
  return item.value.type === '需求' ? '需求详情' : item.value.type === '服务' ? '服务能力详情' : '商品设备详情'
})

const statusTagClass = computed(() => {
  if (!item.value) return 'green'
  return isEnded(item.value) ? 'gray' : item.value.type === '商品' ? 'orange' : 'green'
})

const deadlineText = computed(() => {
  if (!item.value) return ''
  if (item.value.type === '需求') return item.value.deadline
  return item.value.type === '服务' ? '可预约' : '可对接'
})
const deadlineLabel = computed(() => {
  if (!item.value) return ''
  return item.value.type === '需求' ? '截止时间' : item.value.type === '服务' ? '服务档期' : '供货状态'
})

const shortCompany = computed(() => {
  if (!item.value) return ''
  return item.value.company.replace('有限公司', '')
})
const companyInitial = computed(() => (item.value ? (item.value.company || '企').slice(0, 1) : '企'))
// 发布方展示：有已认证企业显示企业名，否则显示发布者名（个人）
const publisherEnterprise = computed(() => (item.value && item.value.publisher_enterprise) || null)
const publisherName = computed(() => {
  if (!item.value) return ''
  return (publisherEnterprise.value && publisherEnterprise.value.name) || item.value.company || '平台用户'
})

const descTitle = computed(() => {
  if (!item.value) return ''
  return item.value.type === '需求' ? '需求描述' : item.value.type === '服务' ? '服务能力' : '商品说明'
})
const fieldsTitle = computed(() => {
  if (!item.value) return ''
  return item.value.type === '需求' ? '作业与交付信息' : item.value.type === '服务' ? '能力与资质信息' : '设备与保障信息'
})
const mediaTitle = computed(() => {
  if (!item.value) return ''
  return item.value.type === '需求' ? '图片资料' : item.value.type === '服务' ? '作业案例' : '实拍与资料'
})

const mediaImages = computed(() => {
  if (!item.value) return []
  // 真实数据：展示全部上传图片；模拟/兜底数据：展示封面图
  if (Array.isArray(item.value.images) && item.value.images.length) return item.value.images
  const first = item.value.image || IMG_SOLAR
  const second = first === IMG_SOLAR ? IMG_HERO : IMG_SOLAR
  return [first, second]
})
// 附件名列表：原为写死的模拟文件名（作业技术要求.pdf 等），实际从未真实上传，
// 属误导性死数据，已移除；模板 v-if="attachmentNames.length" 保证空时不渲染
const attachmentNames = computed(() => [])

const recommendItems = computed(() => {
  if (!item.value) return []
  const pool = item.value.type === '需求' ? getKindItems('supply', 'service') : getKindItems('demand')
  return pool.slice(0, 2)
})

/* ================= 数据加载 ================= */
async function loadDetail() {
  state.value = 'loading'
  const fallback = findMock(postId)
  if (!postId && fallback) {
    item.value = fallback
    state.value = 'ready'
    return
  }
  // 服务能力无独立详情接口时的最终兜底：从服务列表按 id 匹配
  const applyServiceFallback = async () => {
    const svc = await fetchServiceByList(postId)
    if (svc) {
      item.value = svc
      state.value = 'ready'
      return true
    }
    return false
  }
  // 本地已上架发布（发布页打通展示）：接口找不到时按 id 匹配本地内容
  const applyLocal = () => {
    const local = getPosts()
      .filter((p) => p.statusKey === 'live')
      .map(publishPostToCard)
      .find((c) => c && String(c.id) === String(postId))
    if (local) {
      item.value = local
      state.value = 'ready'
      return true
    }
    return false
  }
  try {
    const res = await request({ url: '/api/v1/demands/' + encodeURIComponent(postId) })
    const d = (res && res.data) || res
    const normalized = normalizeDemand(d)
    if (normalized) {
      item.value = normalized
      state.value = 'ready'
      return
    }
    // 详情接口无此 id（服务/商品）：先试服务详情接口，再列表匹配兜底
    if (await applyServiceDetail()) return
    if (await applyServiceFallback()) return
    if (applyLocal()) return
    if (fallback) {
      item.value = fallback
      state.value = 'ready'
      return
    }
    state.value = 'error'
  } catch (e) {
    if (await applyServiceDetail()) return
    if (await applyServiceFallback()) return
    if (applyLocal()) return
    if (fallback) {
      item.value = fallback
      state.value = 'ready'
    } else {
      state.value = 'error'
    }
  }
}

// 服务能力公开详情：GET /api/v1/service-listings/{id}
async function applyServiceDetail() {
  try {
    const res = await request({ url: '/api/v1/service-listings/' + encodeURIComponent(postId) })
    const svc = normalizeService((res && res.data) || res)
    if (svc) {
      item.value = svc
      state.value = 'ready'
      return true
    }
  } catch (err) {
    // 404 等情况继续走列表兜底
  }
  return false
}

async function fetchServiceByList(id) {
  try {
    const res = await request({ url: '/api/v1/service-listings', data: { page: 1, page_size: 100 } })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    const found = items.find((s) => String(s.id) === String(id))
    return found ? normalizeService(found) : null
  } catch (err) {
    return null
  }
}

function findMock(id) {
  const all = [...getKindItems('demand'), ...getKindItems('supply', 'service'), ...getKindItems('supply', 'product')]
  return all.find((i) => i.id === id) || null
}

/* ================= 导航 ================= */
const goBack = () => safeBack()
const goDetail = (r) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(r.id))
const goMatches = () => safeNavigateTo('/pkg-demand/pages/demands/matches')
const previewImage = (i) => uni.previewImage({ urls: mediaImages.value, current: mediaImages.value[i] })

// 微信原生分享（右上角菜单与底部 open-type="share" 按钮共用）
onShareAppMessage(() => {
  const it = item.value || {}
  return {
    title: '需求：' + (it.title || '无人机服务需求'),
    path: '/pages/demands/detail?id=' + encodeURIComponent(demandId),
  }
})
// 企业主页：仅当发布者有已认证企业时可跳转（后端 publisher_enterprise 兜底）
const toastCompany = () => {
  const ent = item.value && item.value.publisher_enterprise
  if (ent && ent.id) {
    uni.navigateTo({ url: '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(ent.id) })
  } else {
    uni.showToast({ title: '该发布者为个人，暂无企业主页', icon: 'none' })
  }
}

/* ================= 收藏 / 登记对接 ================= */
// 收藏走真实接口，按内容类型选端点：需求 / 服务能力 / 商品（登录后可用）
const favBaseUrls = {
  '需求': '/api/v1/demands/',
  '服务': '/api/v1/service-listings/',
  '商品': '/api/v1/products/',
}
const favListUrls = {
  '需求': '/api/v1/demands/favorites/mine',
  '服务': '/api/v1/service-listings/favorites/mine',
  '商品': '/api/v1/products/favorites/mine',
}
const onFavorite = async () => {
  if (!isLoggedIn()) {
    openSheet('login')
    return
  }
  const base = favBaseUrls[item.value && item.value.type]
  if (!base) return
  if (favoriting.value) return
  favoriting.value = true
  try {
    const next = !favorited.value
    await request({
      url: base + encodeURIComponent(postId) + '/favorite',
      method: 'POST',
      data: { favorite: next },
    })
    favorited.value = next
    uni.showToast({ title: next ? '已收藏' : '已取消收藏', icon: 'none' })
  } catch (e) {
    uni.showToast({ title: '操作失败，请重试', icon: 'none' })
  } finally {
    favoriting.value = false
  }
}

// 进入页面时加载收藏状态（登录后，内容类型确定后调用）
const loadFavoriteState = async () => {
  if (!isLoggedIn() || !postId || !item.value) return
  const url = favListUrls[item.value.type]
  if (!url) return
  try {
    const res = await request({ url })
    const list = Array.isArray(res) ? res : (res && res.data) || []
    // 接口返回收藏对象数组；兼容旧版纯 ID 数组
    favorited.value = Array.isArray(list) && list.some(d => (typeof d === 'string' ? d : d && d.id) === postId)
  } catch (e) { /* 忽略：保持未收藏 */ }
}

const onIntent = async () => {
  if (!isLoggedIn()) {
    openSheet('login')
    return
  }
  // 自己发布的需求不可登记对接（本地记录/后端 is_mine 双路径拦截）
  if (isMyDemand.value) {
    uni.showToast({ title: '不能登记自己发布的需求', icon: 'none' })
    return
  }
  // 已登记过该需求（历史任意状态）→ 不再开放重复登记
  if (intented.value) {
    uni.showToast({ title: '已登记过该需求的对接意向', icon: 'none' })
    return
  }
  if (!(await isAnyCertified())) {
    openSheet('cert')
    return
  }
  openSheet('intent')
}

// 是否自己发布的需求：后端 is_mine 标记（真实需求）或本地记录（post- 前缀，本地记录均为本人发布）
const isMyDemand = computed(() => {
  if (item.value && item.value.is_mine) return true
  return !!(postId && (postId.indexOf('post-') === 0 || postId.indexOf('local-') === 0))
})

// 我的意向记录里该需求是否存在"待处理"意向——存在则隐藏登记入口；
// 已关闭（含取消登记）/已洽谈的不阻塞再次登记，与后端防重复规则一致
const intented = ref(false)
const checkIntented = async () => {
  if (!isLoggedIn()) return
  try {
    const res = await request({ url: '/api/v1/intents/mine' })
    const data = Array.isArray(res) ? res : (res && res.data) || []
    intented.value = data.some((it) => it.demand_id === postId && it.status === 'pending')
  } catch (e) {
    /* 拉取失败不阻塞页面 */
  }
}

// 企业认证：真实检查（/api/v1/enterprises 是否存在 approved 记录），不再读 mock hall_certified
const isEnterpriseCertified = async () => {
  try {
    const res = await request({ url: '/api/v1/enterprises' })
    const data = (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    return items.some((e) => e.status === 'approved')
  } catch (e) {
    return false
  }
}

// 飞手认证：真实检查（/api/v1/certified-pilots/mine status=approved，个人飞手走此通道）
const isPilotCertified = async () => {
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const p = res || {}
    return p.status === 'approved'
  } catch (e) {
    return false
  }
}

// 对接认证门槛：企业认证或飞手认证任一通过即可登记对接（个人飞手不强制企业主体）
const isAnyCertified = async () => (await isEnterpriseCertified()) || (await isPilotCertified())

/* ================= 会话弹层 ================= */
const sheet = ref({ show: false, kind: '', title: '' })
const intentForm = ref({ name: '', phone: '', note: '', agree: false })

const enterpriseName = computed(() => currentUserName())

function openSheet(kind) {
  const titles = {
    login: '登录后继续',
    cert: '完成认证',
    intent: '登记对接意向',
  }
  if (kind === 'intent') {
    const u = currentUserName()
    intentForm.value.name = u === '微信用户' ? '' : u
    // 不预填假手机号：真实号码由用户填写（placeholder 已提示），避免误导发布方
    intentForm.value.phone = ''
    intentForm.value.note = ''
    intentForm.value.agree = false
  }
  sheet.value = { show: true, kind, title: titles[kind] || '' }
}

const closeSheet = () => { sheet.value = { show: false, kind: '', title: '' } }

const goLogin = () => {
  closeSheet()
  uni.navigateTo({ url: '/pages/login/index' })
}

const goCert = () => {
  closeSheet()
  uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })
}

const goPilotCert = () => {
  closeSheet()
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

const submitIntent = async () => {
  if (submitting.value) return
  if (!intentForm.value.name.trim()) {
    uni.showToast({ title: '请填写联系人', icon: 'none' })
    return
  }
  if (!intentForm.value.phone.trim()) {
    uni.showToast({ title: '请填写联系电话', icon: 'none' })
    return
  }
  if (!intentForm.value.note.trim()) {
    uni.showToast({ title: '请填写能力说明', icon: 'none' })
    return
  }
  if (!intentForm.value.agree) {
    uni.showToast({ title: '请确认信息授权', icon: 'none' })
    return
  }
  // 兜底：自己发布的需求（本地记录）不可自登记，直接阻断不落本地
  if (isMyDemand.value) {
    uni.showToast({ title: '不能登记自己发布的需求', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    // 登记后端意向。真实后端需求 id 是数字字符串（normalizeDemand 的 String(d.id)），
    // 非本地前缀（post-*/local-* 与 hallData mock d1/s1/p1）一律走后端 POST，
    // 提交失败明确提示；仅本地演示内容落本地存储兜底
    const isLocalDemo = !postId ||
      postId.indexOf('post-') === 0 ||
      postId.indexOf('local-') === 0 ||
      !!findMock(postId)
    const isReal = !isLocalDemo
    let backendOk = false
    if (isReal) {
      try {
        await request({
          url: '/api/v1/demands/' + encodeURIComponent(postId) + '/intents',
          method: 'POST',
          data: {
            intentor_name: intentForm.value.name,
            contact: intentForm.value.phone,
            remark: intentForm.value.note,
          },
        })
        backendOk = true
      } catch (e) {
        let msg = ''
        try {
          if (e && e.data && e.data.error && e.data.error.message) msg = e.data.error.message
          else if (e && e.message) msg = e.message
        } catch { /* ignore */ }
        uni.showToast({ title: msg || '提交失败，请稍后重试', icon: 'none' })
        submitting.value = false
        return
      }
    }
    if (!backendOk) {
      // 仅本地演示内容写入本地存储（后端成功时由 /api/v1/intents/mine 承载）
      const sent = getSentIntents()
      sent.unshift({
        id: 'sent' + Date.now(),
        name: currentUserName(),
        initial: (currentUserName() || '云').slice(0, 1),
        target: item.value ? item.value.title : '',
        note: intentForm.value.note.trim(),
        status: '待处理',
        createdAt: '刚刚',
      })
      saveSentIntents(sent)
    }
    uni.showToast({ title: backendOk ? '对接意向已提交' : '意向已保存到本地', icon: 'success' })
    closeSheet()
  } finally {
    submitting.value = false
  }
}

/* ================= 生命周期 ================= */
onLoad((options) => {
  initSafeTop()
  postId = (options && options.id) || ''
  loadDetail()
  checkIntented()
  // 收藏状态依赖内容类型（需求/服务/商品），内容就绪后自动加载
})
watch(state, (v) => {
  if (v === 'ready') loadFavoriteState()
})
</script>

<style scoped>
.detail-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(150rpx + env(safe-area-inset-bottom));
}

.tap-fade { opacity: 0.85; }

/* ═══════ 头部 ═══════ */
.page-header {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn {
  width: 72rpx;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; text-align: center; }
/* 右侧与返回钮同宽占位：标题区域左右对称，文字真正屏幕居中 */
.head-spacer { width: 72rpx; flex-shrink: 0; }

/* ═══════ 内容区块 ═══════ */
.detail-body { padding-bottom: 8rpx; }

.detail-hero {
  padding: 36rpx 32rpx 28rpx;
  background: #fff;
  border-bottom: 16rpx solid #F4F6F8;
}
.tag-row { display: flex; gap: 10rpx; align-items: center; }
.tag {
  max-width: 260rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.gray { color: #667085; background: #F1F3F5; }
.tag.red { color: #D92D20; background: #FEF3F2; }

.detail-title {
  display: block;
  font-size: 42rpx;
  line-height: 1.35;
  margin: 18rpx 0 22rpx;
  font-weight: 760;
  color: #17212B;
}
.detail-sub {
  display: flex;
  gap: 20rpx;
  color: #667085;
  font-size: 24rpx;
}
.detail-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  margin-top: 30rpx;
  padding-top: 24rpx;
  border-top: 1px solid #EEF1F4;
  gap: 16rpx;
}
.grid-cell { min-width: 0; }
.grid-value {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.grid-value.price { color: #E96012; }
.grid-label {
  display: block;
  margin-top: 8rpx;
  font-size: 20rpx;
  color: #667085;
}

/* ═══════ 分段区块 ═══════ */
.detail-section {
  padding: 32rpx;
  background: #fff;
  border-bottom: 16rpx solid #F4F6F8;
}
.section-title { display: block; font-size: 30rpx; font-weight: 700; color: #17212B; margin-bottom: 20rpx; }
.desc-text {
  color: #344054;
  font-size: 26rpx;
  line-height: 1.8;
  white-space: pre-line;
}

/* 字段表 */
.detail-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1px;
  background: #EEF1F4;
  border: 1px solid #EEF1F4;
  border-radius: 14rpx;
  overflow: hidden;
}
.field-cell {
  padding: 20rpx;
  background: #fff;
  min-height: 114rpx;
  box-sizing: border-box;
}
.field-label { display: block; color: #667085; font-size: 20rpx; margin-bottom: 8rpx; }
.field-value { font-size: 24rpx; line-height: 1.35; color: #17212B; font-weight: 600; }

/* 媒体 */
.media-row { white-space: nowrap; }
.media-inner { display: inline-flex; gap: 16rpx; }
.media-img {
  width: 236rpx;
  height: 160rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  background: #E8F2FC;
}
.attach-list { margin-top: 16rpx; display: flex; flex-direction: column; gap: 12rpx; }
.attach-item {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx;
  border: 1px solid #EEF1F4;
  border-radius: 12rpx;
  background: #FAFBFC;
}
.attach-icon { color: #0A66C2; font-size: 26rpx; }
.attach-name { color: #344054; font-size: 24rpx; }

/* 发布企业 */
.company-row {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 22rpx;
  border: 1px solid #E4E7EC;
  border-radius: 14rpx;
}
.company-avatar {
  width: 72rpx;
  height: 72rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  color: #0A66C2;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 32rpx;
}
.company-copy { flex: 1; min-width: 0; }
.company-name { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; }
.company-tag { display: block; font-size: 22rpx; color: #667085; margin-top: 6rpx; }
.company-tag--plain { color: #98A2B3; }
.company-link { color: #0A66C2; font-size: 24rpx; white-space: nowrap; }

/* 推荐 */
.recommend-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.recommend-head .section-title { margin-bottom: 0; }
.recommend-more { display: flex; align-items: center; gap: 4rpx; color: #0A66C2; font-size: 24rpx; font-weight: 600; }
.more-arrow { font-size: 28rpx; }
.recommend-row { white-space: nowrap; }
.recommend-inner { display: inline-flex; gap: 16rpx; }
.recommend-card {
  min-width: 364rpx;
  padding: 22rpx;
  background: #F8FBFE;
  border: 1px solid #DCEBF8;
  border-radius: 14rpx;
}
.recommend-title {
  display: block;
  font-size: 24rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.recommend-meta {
  display: block;
  color: #667085;
  font-size: 22rpx;
  margin-top: 12rpx;
}

/* ═══════ 底部操作栏 ═══════ */
.fixed-actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom));
  display: flex;
  gap: 20rpx;
  background: #fff;
  border-top: 1px solid #EEF1F4;
  z-index: 19;
  box-shadow: 0 -2px 10px rgba(16, 24, 40, 0.05);
}
.action-secondary,
.action-primary {
  height: 86rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 26rpx;
}
.action-secondary {
  width: 168rpx;
  flex-shrink: 0;
  color: #344054;
  background: #fff;
  border: 1px solid #E4E7EC;
  gap: 8rpx;
}
.fav-icon { font-size: 30rpx; line-height: 1; }
.fav-icon.on { color: #E96012; }
.share-icon { font-size: 28rpx; line-height: 1; }
/* 分享走原生 button open-type="share"：清除微信 button 默认边框/背景，与收藏按钮同款 */
.share-btn {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  line-height: 84rpx;
  color: #344054;
  background: #fff;
  border: 1px solid #E4E7EC;
  gap: 8rpx;
}
.share-btn::after { border: none; }
.action-primary {
  flex: 1;
  color: #fff;
  background: #0A66C2;
}
.action-primary.disabled {
  color: #667085;
  background: #E9EDF1;
}

/* ═══════ 骨架屏 ═══════ */
.skeleton-wrap { padding: 32rpx; display: flex; flex-direction: column; gap: 24rpx; }
.skeleton {
  background: linear-gradient(90deg, #E9EDF1 25%, #F5F7F9 37%, #E9EDF1 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s infinite;
  border-radius: 12rpx;
}
.skeleton-title { height: 44rpx; width: 70%; }
.skeleton-line { height: 24rpx; width: 90%; }
.skeleton-block { height: 240rpx; }
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: 0 0; }
}

/* ═══════ 状态面板 ═══════ */
.state-panel {
  min-height: 620rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-mark.err { color: #D92D20; background: #FEF3F2; }
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}

/* ═══════ 弹层 ═══════ */
.sheet { padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); }
.sheet-head {
  display: flex;
  align-items: center;
  padding: 28rpx 32rpx 20rpx;
}
.sheet-title { flex: 1; font-size: 32rpx; font-weight: 700; color: #17212B; }
.sheet-close { width: 56rpx; height: 56rpx; display: flex; align-items: center; justify-content: center; }
.sheet-x { font-size: 40rpx; color: #98A2B3; line-height: 1; }
.sheet-body { padding: 0 32rpx 24rpx; }
.sheet-desc {
  display: block;
  font-size: 26rpx;
  color: #667085;
  line-height: 1.7;
  margin-bottom: 8rpx;
}
.sheet-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 32rpx;
}
.sheet-actions > view {
  flex: 1;
  height: 84rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
}
.ghost-btn { border: 1px solid #E4E7EC; background: #fff; color: #344054; }
.primary-btn { background: #0A66C2; color: #fff; }
.secondary-btn { border: 1px solid #0A66C2; background: #fff; color: #0A66C2; }
.primary-btn.disabled { opacity: 0.6; }

.field-label {
  display: block;
  color: #344054;
  font-size: 24rpx;
  font-weight: 650;
  margin: 24rpx 0 14rpx;
}
.req { color: #D92D20; font-style: normal; }
.field-static {
  height: 80rpx;
  line-height: 80rpx;
  padding: 0 20rpx;
  background: #F4F6F8;
  border-radius: 12rpx;
  font-size: 26rpx;
  color: #17212B;
}
.field,
.textarea {
  width: 100%;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #17212B;
  font-size: 26rpx;
  padding: 0 20rpx;
  box-sizing: border-box;
}
.field { height: 84rpx; }
.textarea {
  height: 168rpx;
  padding-top: 20rpx;
  line-height: 1.6;
}
.agree-row {
  display: flex;
  gap: 14rpx;
  align-items: flex-start;
  margin: 24rpx 0 8rpx;
}
.agree-box {
  width: 36rpx;
  height: 36rpx;
  flex-shrink: 0;
  border: 2rpx solid #D0D5DD;
  border-radius: 6rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 2rpx;
}
.agree-box.on { background: #0A66C2; border-color: #0A66C2; }
.agree-check { color: #fff; font-size: 24rpx; line-height: 1; }
.agree-text {
  font-size: 22rpx;
  color: #667085;
  line-height: 1.5;
}
</style>
