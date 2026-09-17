<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">发布预览</view>
    </view>

    <!-- 预览卡片 -->
    <view class="pub-preview-card">
      <view class="pub-preview-type">{{ typeConfig.short }} · 发布预览</view>
      <view v-if="photoList.length" class="pub-preview-photos">
        <image v-for="(p, i) in photoList" :key="i" :src="p" mode="aspectFill" class="pub-preview-photo" />
      </view>
      <view class="pub-preview-title">{{ title }}</view>
      <view class="pub-preview-meta">
        <text v-for="(m, i) in metaList" :key="i">{{ m }}</text>
      </view>
      <view class="pub-preview-copy">{{ copyText }}</view>
    </view>

    <!-- 发布说明 -->
    <view class="pub-review-note">
      <text class="pub-review-note-b">发布说明</text>
      <text>发布后将展示在供需大厅对应列表；联系方式仅向登录并发起对接的用户展示。</text>
    </view>

    <!-- 发布前检查 -->
    <view class="pub-section">
      <view class="pub-section-title">发布前检查</view>
      <view class="pub-form-card">
        <view class="pub-check-row"><text class="pub-check-mark">✓</text><text>已填写必填项目</text></view>
        <view class="pub-check-row"><text class="pub-check-mark">✓</text><text>已确认发布内容真实性</text></view>
        <view class="pub-check-row">
          <text class="pub-check-mark">✓</text>
          <text>{{ photoCount ? '已添加 ' + photoCount + ' 张图片' : '可补充图片提升展示效果' }}</text>
        </view>
      </view>
    </view>

    <!-- 固定底部操作区 -->
    <view class="pub-sticky">
      <view class="pub-btn pub-btn--ghost" hover-class="pub-btn--active" @tap="goBack">返回修改</view>
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="openConfirm">确认发布</view>
    </view>

    <!-- 确认弹窗 -->
    <view v-if="showConfirm" class="pub-modal" @tap="showConfirm = false">
      <view class="pub-dialog" @tap.stop>
        <view class="pub-dialog-title">确认发布？</view>
        <view class="pub-dialog-text">{{ confirmText }}</view>
        <view class="pub-dialog-actions">
          <view class="pub-dialog-btn" @tap="showConfirm = false">再检查一下</view>
          <view class="pub-dialog-btn" @tap="submitPublish">确认发布</view>
        </view>
      </view>
    </view>

    <!-- 成功弹窗 -->
    <view v-if="showSuccess" class="pub-modal">
      <view class="pub-dialog pub-success">
        <view class="pub-success-mark">✓</view>
        <view class="pub-dialog-title">已发布成功</view>
        <view class="pub-dialog-text">{{ successText }}</view>
        <view class="pub-dialog-actions">
          <view class="pub-dialog-btn" @tap="goHome">返回发布首页</view>
          <view class="pub-dialog-btn" @tap="goMyPosts">查看我的发布</view>
        </view>
      </view>
    </view>

    <!-- 底部黑色 toast -->
    <view v-if="toast" class="pub-toast">{{ toast }}</view>
  </view>
</template>

<script setup>
import { safeBack } from '../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { TYPES, computePreviewMeta, makePost, upsertPost, loadFormState, clearFormState, PROD_TYPE_MAP } from '../../utils/publishData'
import { isEnterpriseCertified } from '../../utils/cert'
import { useSafeTop } from '../../utils/safeTop'
import { request, authStorage, requireLogin, BASE_URL, getStoredUser, getErrorMessage, uploadFileWithAuth } from '../../utils/request'

const { topPad, initSafeTop } = useSafeTop(true)

const type = ref('')
const values = ref({})
const photoCount = ref(0)
const resumeId = ref('')
// 非空表示"编辑已发布商品"：提交走 PATCH /api/v1/products/{id} 而不是 POST
const editBackendId = ref('')
// 已选图片的真实临时路径（预览展示 + 发布时上传到服务器）
const photoList = ref([])
// 详情图（详情区长图）。与 photoList（顶部图集）是两组独立的图，分开上传与提交。
const detailPhotoList = ref([])
// 需求附件临时路径（发布时上传）
const fileList = ref([])
const showConfirm = ref(false)
const showSuccess = ref(false)
const toast = ref('')
const toastTimer = ref(null)
// 发布防重入：防止连点「确认发布」产生重复数据（弹窗关闭前二次点击）
const submitting = ref(false)
// 商品发布成功后的后端商品 id（写入本地记录 backendId，供列表跳转商品详情）
const backendProductId = ref('')
// 需求发布成功后的后端需求 id（写入本地记录 backendId）
const backendDemandId = ref('')
// 服务能力/培训课程发布成功后的后端 id
const backendServiceId = ref('')
const backendCourseId = ref('')

// 表单证书类型（中文）→ 后端 cert_type 枚举（caac / utc_dji / gov_level）
const mapCertType = (t) => {
  if (t === '大疆 UTC 证书') return 'utc_dji'
  if (t === '职业技能等级') return 'gov_level'
  return 'caac' // CAAC 民航局执照 / AOPA 执照
}

const typeConfig = computed(() => TYPES[type.value] || null)

const title = computed(() => values.value.title || '待命名发布内容')
const metaList = computed(() => computePreviewMeta(type.value, values.value))
const copyText = computed(() => values.value.description || '暂未填写补充说明。发布前仍可返回编辑，发布后将展示在供需大厅对应列表。')

// 课程无大厅分类，发布文案与需求/服务/商品区分
const isCourse = computed(() => type.value === 'course')
const confirmText = computed(() => {
  if (isCourse.value) return typeConfig.value.name + '提交后由协会审核，通过后才公开展示，可在「我的发布」中管理。'
  if (type.value === 'product') return '提交后由协会审核，通过后才上架到低空商城，可在「我的发布」中查看审核状态。'
  if (type.value === 'demand') return '提交后由协会审核，通过后才公开展示到供需大厅，可在「我的发布」中查看审核状态。'
  return '发布后将立即上架到供需大厅，可在「我的发布」中随时下架或编辑。'
})
const successText = computed(() => {
  if (isCourse.value) return '课程已提交审核，协会通过后公开展示，可在「我的发布」中查看状态。'
  if (type.value === 'product') return '已提交审核，协会通过后上架到低空商城，可在「我的发布」中查看状态。'
  if (type.value === 'demand') return '需求已提交审核，协会通过后公开展示到供需大厅，可在「我的发布」中查看状态。'
  return '内容已上架到供需大厅，可在「我的发布」中管理。'
})

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => { toast.value = '' }, 2200)
}

function goBack() {
  safeBack()
}

function openConfirm() {
  showConfirm.value = true
}

async function submitPublish() {
  const t = typeConfig.value
  if (!t) return
  if (submitting.value) return
  submitting.value = true
  // 商品/服务发布要求已登录（未登录时后端 401 会跳登录页；这里提前拦截给明确提示）
  if (type.value === 'product' && !authStorage.getAccessToken()) {
    submitting.value = false
    showToast('请先登录后再发布')
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  // 需求发布同样要求登录：先提交后端（POST /api/v1/demands），成功才落本地
  if (type.value === 'demand' && !authStorage.getAccessToken()) {
    submitting.value = false
    showToast('请先登录后再发布需求')
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  // 商品发布：先写入后端商城（POST /api/v1/products），成功才置 pending 并记录 backendId。
  // 后端商品在商城/需求大厅展示并支持下单；失败则不发布，避免本地假上架。
  if (type.value === 'product') {
    // 企业认证门禁**前置到上传之前**：商品详情页对买家承诺「平台认证商家」，
    // 后端只放行已完成企业认证（approved）的账号。不前置的话，用户填完整张表、
    // 传完封面图与详情图才被拒，而且每次重试都会在服务端留下一组孤儿上传。
    if (!(await isEnterpriseCertified())) {
      submitting.value = false
      showToast('发布商品需先完成企业认证，请到「我的 → 企业认证」提交')
      return
    }
    try {
      // 先上传已选图片（POST /api/v1/files/upload → /uploads/{file_id}），
      // 图片上传失败即发布失败，不静默丢图；无图则正常发布
      const images = photoList.value.length ? await uploadImages(photoList.value) : []
      // 详情图与顶部图集分开上传（uploadImages 会跳过已在服务器上的 URL）
      const detailImgs = detailPhotoList.value.length ? await uploadImages(detailPhotoList.value) : []
      const created = await createBackendProduct(values.value, images, detailImgs)
      if (!created || !created.id) throw new Error('create product failed')
      backendProductId.value = created.id
    } catch (e) {
      submitting.value = false
      // 与需求分支同口径：透出后端真实拒绝原因（如"请先完成企业认证"），不吞成通用提示
      showToast(getErrorMessage(e) || '发布失败，请稍后重试')
      return
    }
  }
  // 需求发布：先提交后端（POST /api/v1/demands），成功才落本地记录。
  // 与商品同模式：失败则不发布，避免本地假上架。
  if (type.value === 'demand') {
    // 角色门禁已取消：四种角色（企业/个人/平台管理员/协会管理员）都能发需求，
    // 协会账号是「运营方 + 市场主体」二合一。此前的 association_admin 前置拦截
    // 是配合当时的策略加的，策略已变，一并撤掉——后端保留默认拒绝兜底未知角色。
    try {
      const images = photoList.value.length ? await uploadImages(photoList.value) : []
      const attachments = fileList.value.length ? await uploadImages(fileList.value) : []
      const created = await createBackendDemand(values.value, images, attachments)
      if (!created || !created.id) throw new Error('create demand failed')
      backendDemandId.value = created.id
    } catch (e) {
      submitting.value = false
      // 透出后端真实拒绝原因（如预算下限>上限），不再吞成通用提示
      showToast(getErrorMessage(e) || '发布失败，请稍后重试')
      return
    }
  }
  // 服务能力发布：POST /api/v1/service-listings（待审核，管理端通过后进入公开列表）
  if (type.value === 'service') {
    if (!authStorage.getAccessToken()) {
      submitting.value = false
      showToast('请先登录后再发布服务')
      uni.navigateTo({ url: '/pages/login/index' })
      return
    }
    try {
      const user = getStoredUser()
      const images = photoList.value.length ? await uploadImages(photoList.value) : []
      const created = await request({
        url: '/api/v1/service-listings',
        method: 'POST',
        data: {
          provider_name: (user && user.name) || '',
          title: String(values.value.title || '').trim(),
          category: String(values.value.category || '').trim(),
          description: [values.value.equipment, values.value.cert].filter(Boolean).join('\n'),
          region: String(values.value.range || '').trim(),
          price_fen: 0, // 报价方式见 unit；具体金额由需求方洽谈
          unit: String(values.value.quote || '面议'),
          image: images[0] || '',
        },
      })
      if (!created || !created.id) throw new Error('create service listing failed')
      backendServiceId.value = created.id
    } catch (e) {
      submitting.value = false
      showToast('发布失败，请稍后重试')
      return
    }
  }
  // 培训课程发布：POST /api/v1/training-courses（即时上架公开）
  if (type.value === 'course') {
    if (!authStorage.getAccessToken()) {
      submitting.value = false
      showToast('请先登录后再发布课程')
      uni.navigateTo({ url: '/pages/login/index' })
      return
    }
    try {
      const images = photoList.value.length ? await uploadImages(photoList.value) : []
      const created = await request({
        url: '/api/v1/training-courses',
        method: 'POST',
        data: {
          title: String(values.value.title || '').trim(),
          cert_type: mapCertType(values.value.certType),
          description: String(values.value.description || '').trim(),
          org_name: String(values.value.org || '').trim(),
          district: String(values.value.district || '').trim(),
          location: String(values.value.location || '').trim(),
          price_fen: Math.round((Number(values.value.price) || 0) * 100),
          duration_days: Number(values.value.duration) || 0,
          max_students: Number(values.value.quota) || 0,
          image: images[0] || '',
        },
      })
      if (!created || !created.id) throw new Error('create course failed')
      backendCourseId.value = created.id
    } catch (e) {
      submitting.value = false
      showToast('发布失败，请稍后重试')
      return
    }
  }
  // 四种类型均走后端：提交后为"待审核/已发布"（通过后由后端统一展示，本地记录仅留发布根）
  const reviewBackend = true
  const post = makePost({
    id: resumeId.value || '',
    type: type.value,
    values: Object.assign({}, values.value),
    photoCount: photoCount.value,
    statusKey: reviewBackend ? 'pending' : 'live',
    status: reviewBackend ? '待审核' : '已发布',
    date: '刚刚发布',
    note: type.value === 'product'
      ? '已提交审核，协会通过后上架到低空商城，可在「我的发布」中查看状态。'
      : type.value === 'demand'
        ? '需求已提交，协会审核通过后公开展示'
        : type.value === 'service'
          ? '服务已提交审核，协会通过后展示在生态服务'
          : '课程已提交审核，协会通过后展示在培训认证，可在「我的发布」中查看状态。',
  })
  if (backendProductId.value) post.backendId = backendProductId.value
  if (backendDemandId.value) post.backendId = backendDemandId.value
  if (backendServiceId.value) post.backendId = backendServiceId.value
  if (backendCourseId.value) post.backendId = backendCourseId.value
  upsertPost(post)
  showConfirm.value = false
  showSuccess.value = true
  submitting.value = false
}

/* ── 商品发布 → 后端商城（字段映射） ── */

// 该图是否已经在服务器上（编辑已发布商品时回填的是 /uploads/... 这类地址，
// 它们不能再上传一次——服务端没有这个文件，uni.uploadFile 会直接失败）。
//
// ⚠️ 这里**绝不能**用"以 http 开头"来判：微信小程序的本地临时文件路径就长这样。
// wx.chooseMedia/chooseImage 返回的 tempFilePath 形如 http://tmp/xxx.jpg（开发者工具、
// 部分机型），同样是 http:// 开头。早先那样判会把手机本地路径误认成服务器图、
// 直接跳过上传，最后把 http://tmp/xxx.jpg 原样存进数据库——生产库 drone_products
// 已出现该数据，管理端与任何浏览器都打不开。
// 所以判据只认**我们自己产出的地址形态**，认不出来的一律当本地文件走上传（失败会抛出，不静默）。
function isServerImage(u) {
  if (typeof u !== 'string') return false
  const s = u.trim()
  if (!s) return false
  // 站内相对路径：/uploads/... 、/static/...（协议相对 //host 不是我们的形态）
  if (s.charAt(0) === '/' && s.indexOf('//') !== 0) return true
  if (s.indexOf('uploads/') === 0) return true
  if (s.indexOf('http://') === 0 || s.indexOf('https://') === 0) {
    if (BASE_URL && s.indexOf(BASE_URL) === 0) return true
    // BASE_URL 变更过的历史数据：仍认 /uploads/、/static/ 形态的绝对地址
    return /^https?:\/\/[^/]+\/(uploads|static)\//.test(s)
  }
  return false
}

// 逐张上传本地图片到服务器，返回 /uploads/{file_id} 可访问路径（与证件上传同模式）
async function uploadImages(paths) {
  // 带鉴权上传：401 自动刷新 token 重试（raw uploadFile 不走 request 拦截器）
  const urls = []
  for (const p of paths) {
    if (isServerImage(p)) {
      urls.push(p)
      continue
    }
    urls.push(await uploadFileWithAuth(p))
  }
  // 出口兜底：上传完每一项都必须是**服务端地址**。
  // uploadFileWithAuth 自己失败会抛错，但"被 isServerImage 误判而跳过上传"不会——
  // 那条路径会把手机本地路径原样写进库（生产已发生过一次：images 里存了 http://tmp/xxx.jpg），
  // 而且不会有任何报错。宁可发布失败，也不能把打不开的地址存进去。
  const bad = urls.find((u) => !isServerImage(u))
  if (bad) throw new Error('图片未成功上传，请重试')
  return urls
}

// 表单交付方式（中文）→ 后端 delivery 枚举。
// 此前这个字段**采集了却在提交时丢弃**（和"可售数量"一样），
// 于是"要不要收货地址"只能按商品类型猜：卖家选了「自提」，买家仍被要求填地址。
function mapDelivery(d) {
  const s = String(d || '').trim()
  if (s === '自提') return 'pickup'
  if (s === '同城配送') return 'city'
  if (s === '物流发货') return 'logistics'
  if (s === '可协商') return 'negotiable'
  return ''
}

// 服务区域：选「其他」时取手填值，否则取预置范围。
// 空串合法（选填），实物商品留空。
function mapRegion(v) {
  const preset = String(v.region || '').trim()
  if (preset === '其他') return String(v.regionOther || '').trim()
  return preset
}

// 表单价格方式（中文）→ 后端 price_mode 枚举。
// 此前没有这个维度：售价留空 → (Number('') || 0) * 100 = 0 → 落库 price_fen=0，
// 与"卖家填了 0 元"完全一样，后端无法区分；设备区卡片还把它显示成 ¥0。
function mapPriceMode(m) {
  return String(m || '').trim() === '面议' ? 'negotiable' : 'fixed'
}

// 保存商品：新建走 POST /api/v1/products；编辑已发布商品走 PATCH /api/v1/products/{id}
// （后者会被服务端**退回待审核**——审核的对象是内容，内容变了原结论就失效）。
async function createBackendProduct(v, images, detailImages) {
  const mode = mapPriceMode(v.priceMode)
  // 面议商品的价格必须为 0（后端同样强制），避免 0 这个值同时表达两种含义。
  const fen = mode === 'negotiable' ? 0 : Math.round((Number(v.price) || 0) * 100)
  const editing = !!editBackendId.value
  // 发布类型决定提交哪一套字段，另一分支**显式置空**。
  // 表单在切分支时已经清过值（form.vue pickSegment），这里再兜一次底：
  // 恢复的历史草稿可能两个分支都留着值，不兜底就会写出
  // "服务商品带成色/物流发货" 这类自相矛盾的记录。
  const svc = String(v.bizKind || '').trim() === '服务能力'
  return request({
    url: editing ? '/api/v1/products/' + encodeURIComponent(editBackendId.value) : '/api/v1/products',
    method: editing ? 'PATCH' : 'POST',
    data: {
      title: String(v.title || '').trim(),
      description: String(v.description || '').trim(),
      // 成色/品牌型号：只有实物商品有。服务留空 —— 空成色落到 'new'，
      // 与后端 internal/service/trading.go:159-167 的归一规则一致。
      condition: svc ? '' : (String(v.condition || '').indexOf('二手') === 0 ? 'used' : 'new'),
      brand: svc ? '' : splitBrand(v.brand),
      model: svc ? '' : splitModel(v.brand),
      // 交付方式同理：服务的 OrderNeedsReceiver 走 prod_type 兜底
      //（trading.go:88 只认 drone/part），本就不需要收货地址。
      delivery: svc ? '' : mapDelivery(v.delivery),
      // 服务类字段：服务能力并入商品后由商品承载（空串合法，实物商品留空）
      category: svc ? String(v.category || '').trim() : '',
      region: svc ? mapRegion(v) : '',
      unit: svc ? String(v.unit || '').trim() : '',
      // 商品类型的两个下拉只在一侧出现，按分支取
      prod_type: mapProdType(svc ? v.serviceType : v.productType),
      price_mode: mode,
      price_fen: fen,
      images: images || [],
      // 详情图（详情区长图）：与顶部图集 images 是两组独立的图
      detail_images: detailImages || [],
    },
  })
}

// 表单商品类型（中文）→ 后端 prod_type 枚举（7 类，与 domain.ProductType 一一对应）。
// 映射表唯一一份在 utils/publishData.js 的 PROD_TYPE_MAP——它与发布页的选项列表同源，
// 曾在这里各存一份，改一边就漂。
//
// 修复前这里只映射 3 类：航拍服务/试飞测试/检测标定/空域协调**在小程序里根本发不出来**，
// 只能从管理端建。管理端表单早就是 7 类，两边口径不一致。
function mapProdType(t) {
  return PROD_TYPE_MAP[String(t || '').trim()] || 'part'
}

// POST /api/v1/demands，返回创建的后端需求
async function createBackendDemand(v, images, attachments) {
  const u = getStoredUser() || {}
  return request({
    url: '/api/v1/demands',
    method: 'POST',
    data: {
      publisher_name: String(u.name || u.phone || '').trim() || '微信用户',
      contact: String(v.contact || '').trim(),
      district: String(v.district || '').trim(),
      biz_type: mapBizType(v.biz),
      title: String(v.title || '').trim(),
      description: String(v.description || '').trim(),
      images: images || [],
      budget: Number(v.budget_max) || 0, // 预算上限（元，后端换算为 budget_fen；0=面议）
      budget_min: Number(v.budget_min) || 0, // 预算下限（元，后端换算为 budget_min_fen；0=不限）
      aircraft: String(v.aircraft || '').split(/[，,]/).map((x) => x.trim()).filter(Boolean),
      pilot_count: Number(v.pilot_count) || 0,
      attachments: attachments || [],
    },
  })
}

// 表单业务类型（中文）→ 后端 biz_type 枚举（与 utils/enums.js BIZ_TYPE_LABEL 一致）
function mapBizType(b) {
  const m = {
    '巡检': 'cable_inspection',
    '植保': 'plant_transport',
    // 后端 BizType 暂无「测绘/航拍/吊运」枚举，统一归入 other（需求仍可正常发布展示）
    '测绘': 'other',
    '航拍': 'other',
    '吊运': 'other',
    '其他': 'other',
  }
  return m[b] || 'other'
}

// 表单品牌/型号为合并输入（如「DJI / M350 RTK」），拆分到后端 brand/model
function splitBrand(b) {
  const s = String(b || '')
  const i = s.indexOf('/')
  return (i > 0 ? s.slice(0, i) : s).trim()
}
function splitModel(b) {
  const s = String(b || '')
  const i = s.indexOf('/')
  return i > 0 ? s.slice(i + 1).trim() : ''
}

function goHome() {
  clearFormState()
  uni.switchTab({ url: '/pages/publish/index' })
}

function goMyPosts() {
  clearFormState()
  uni.navigateTo({ url: '/pages/publish/my-posts?tab=live' })
}

onShow(() => {
  // 未登录每次进入都拦截（分享直达/深链/从登录页返回均继续跳转）
  requireLogin('请先登录后再发布')
})

onLoad((options) => {
  initSafeTop()
  // 从表单页带入的临时状态
  const state = options && options.state
  if (state) {
    try {
      const parsed = JSON.parse(decodeURIComponent(state))
      type.value = parsed.type || ''
      values.value = parsed.values || {}
      photoCount.value = parsed.photoCount || 0
      resumeId.value = parsed.resumeId || ''
      photoList.value = Array.isArray(parsed.photos) ? parsed.photos : []
      detailPhotoList.value = Array.isArray(parsed.detailPhotos) ? parsed.detailPhotos : []
      fileList.value = Array.isArray(parsed.files) ? parsed.files.filter(Boolean) : []
      editBackendId.value = parsed.editBackendId || ''
      return
    } catch (e) { /* fallthrough */ }
  }
  // 兜底：读取表单页写入的 storage
  const st = loadFormState()
  if (st) {
    type.value = st.type || ''
    values.value = st.values || {}
    photoCount.value = st.photoCount || 0
    resumeId.value = st.resumeId || ''
    photoList.value = Array.isArray(st.photos) ? st.photos : []
    detailPhotoList.value = Array.isArray(st.detailPhotos) ? st.detailPhotos : []
    editBackendId.value = st.editBackendId || ''
  }
})
</script>

<style scoped>
@import './pub-style.css';
.pub-fade { opacity: 0.6; }
.pub-review-note-b { color: #0A66C2; font-weight: 700; }
.pub-preview-photos {
  display: flex;
  gap: 12rpx;
  margin-top: 16rpx;
  overflow-x: auto;
}
.pub-preview-photo {
  width: 160rpx;
  height: 160rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  background: var(--color-primary-light);
}
</style>
