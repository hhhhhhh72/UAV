<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">发布预览</view>
      <view class="pub-nav-action" hover-class="pub-fade" :style="{ marginRight: capsuleGap + 'px' }" @tap="goBack">编辑</view>
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
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { TYPES, computePreviewMeta, makePost, upsertPost, loadFormState, clearFormState } from '../../utils/publishData'
import { useSafeTop } from '../../utils/safeTop'
import { request, authStorage, BASE_URL } from '../../utils/request'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const type = ref('')
const values = ref({})
const photoCount = ref(0)
const resumeId = ref('')
// 已选图片的真实临时路径（预览展示 + 发布时上传到服务器）
const photoList = ref([])
const showConfirm = ref(false)
const showSuccess = ref(false)
const toast = ref('')
const toastTimer = ref(null)
// 发布防重入：防止连点「确认发布」产生重复数据（弹窗关闭前二次点击）
const submitting = ref(false)
// 商品发布成功后的后端商品 id（写入本地记录 backendId，供列表跳转商品详情）
const backendProductId = ref('')

const typeConfig = computed(() => TYPES[type.value] || null)

const title = computed(() => values.value.title || '待命名发布内容')
const metaList = computed(() => computePreviewMeta(type.value, values.value))
const copyText = computed(() => values.value.description || '暂未填写补充说明。发布前仍可返回编辑，发布后将展示在供需大厅对应列表。')

// 课程无大厅分类，发布文案与需求/服务/商品区分
const isCourse = computed(() => type.value === 'course')
const confirmText = computed(() => {
  if (isCourse.value) return typeConfig.value.name + '发布后将公开展示，可在「我的发布」中管理。'
  if (type.value === 'product') return '提交后由协会审核，通过后才上架到低空商城，可在「我的发布」中查看审核状态。'
  return '发布后将立即上架到供需大厅，可在「我的发布」中随时下架或编辑。'
})
const successText = computed(() => {
  if (isCourse.value) return '内容已保存并公开展示，可在「我的发布」中查看状态。'
  if (type.value === 'product') return '商品已提交审核，协会通过后上架到低空商城，可在「我的发布」中查看状态。'
  return '内容已上架到供需大厅，可在「我的发布」中管理。'
})

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => { toast.value = '' }, 2200)
}

function goBack() {
  uni.navigateBack()
}

function openConfirm() {
  showConfirm.value = true
}

async function submitPublish() {
  const t = typeConfig.value
  if (!t) return
  if (submitting.value) return
  submitting.value = true
  // 商品发布要求已登录（未登录时后端 401 会跳登录页；这里提前拦截给明确提示）
  if (type.value === 'product' && !authStorage.getAccessToken()) {
    submitting.value = false
    showToast('请先登录后再发布商品')
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  // 商品发布：先写入后端商城（POST /api/v1/products），成功才置 live 并记录 backendId。
  // 后端商品在商城/需求大厅展示并支持下单；失败则不发布，避免本地假上架。
  if (type.value === 'product') {
    try {
      // 先上传已选图片（POST /api/v1/files/upload → /uploads/{file_id}），
      // 图片上传失败即发布失败，不静默丢图；无图则正常发布
      const images = photoList.value.length ? await uploadImages(photoList.value) : []
      const created = await createBackendProduct(values.value, images)
      if (!created || !created.id) throw new Error('create product failed')
      backendProductId.value = created.id
    } catch (e) {
      submitting.value = false
      showToast('发布失败，请稍后重试')
      return
    }
  }
  const post = makePost({
    id: resumeId.value || '',
    type: type.value,
    values: Object.assign({}, values.value),
    photoCount: photoCount.value,
    // 商品走后端审核：提交后为"待审核"（通过后由后端商品统一展示，本地记录仅留发布根）
    statusKey: type.value === 'product' ? 'pending' : 'live',
    status: type.value === 'product' ? '待审核' : '已发布',
    date: '刚刚发布',
    note: type.value === 'product' ? '商品已提交审核，协会通过后上架到低空商城，可在「我的发布」中查看状态。' : '内容已上架，可在「我的发布」中随时下架或编辑。',
  })
  if (backendProductId.value) post.backendId = backendProductId.value
  upsertPost(post)
  showConfirm.value = false
  showSuccess.value = true
  submitting.value = false
}

/* ── 商品发布 → 后端商城（字段映射） ── */

// 逐张上传本地图片到服务器，返回 /uploads/{file_id} 可访问路径（与证件上传同模式）
async function uploadImages(paths) {
  const token = authStorage.getAccessToken()
  const urls = []
  for (const p of paths) {
    const data = await new Promise((resolve, reject) => {
      uni.uploadFile({
        url: BASE_URL + '/api/v1/files/upload',
        filePath: p,
        name: 'file',
        header: { Authorization: 'Bearer ' + token },
        success: (r) => {
          if (r.statusCode >= 200 && r.statusCode < 300) {
            try { resolve(JSON.parse(r.data)) } catch (e) { reject(e) }
          } else {
            reject(new Error('upload failed ' + r.statusCode))
          }
        },
        fail: reject,
      })
    })
    const fid = data && (data.file_id || (data.data && data.data.file_id))
    if (!fid) throw new Error('upload response missing file_id')
    urls.push('/uploads/' + fid)
  }
  return urls
}

// POST /api/v1/products，返回创建的后端商品
async function createBackendProduct(v, images) {
  const fen = Math.round((Number(v.price) || 0) * 100)
  return request({
    url: '/api/v1/products',
    method: 'POST',
    data: {
      title: String(v.title || '').trim(),
      description: String(v.description || '').trim(),
      brand: splitBrand(v.brand),
      model: splitModel(v.brand),
      condition: String(v.condition || '').indexOf('二手') === 0 ? 'used' : 'new',
      prod_type: mapProdType(v.productType),
      price_fen: fen,
      images: images || [],
    },
  })
}

// 表单商品类型（中文）→ 后端 prod_type 枚举（drone / part / repair）
function mapProdType(t) {
  if (t === '整机') return 'drone'
  if (t === '维修服务') return 'repair'
  return 'part' // 零部件 / 载荷设备 / 租赁设备
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
