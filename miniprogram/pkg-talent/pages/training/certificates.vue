<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="我的证书" show-back :fixed="true" @back="goBack" />

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 + 归档证书入口 -->
      <view class="ir">
        <text>共 <text class="irn">{{ list.length }}</text> 项证书</text>
        <view class="ir-btn" hover-class="ir-btn--hover" @tap="openApply">＋ 归档证书</view>
      </view>

      <!-- 骨架 -->
      <view v-if="loading" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w80"></view>
            <view class="sk-l w60"></view>
          </view>
        </view>
      </view>

      <!-- 错误 -->
      <view v-else-if="errorMsg && !list.length" class="st">
        <u-empty :description="errorMsg">
          <view class="stb" @tap="fetchList">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="!list.length" class="st">
        <u-empty description="暂无证书">
          <text class="sth">证书由协会在线下考核后颁发，完成培训课程后可联系管理员；已有纸质证书（CAAC/AOPA/UTC 等）可点上方「归档证书」提交登记，协会审核通过后计入档案</text>
          <view class="stb" @tap="goCourses">去逛逛培训课程</view>
        </u-empty>
      </view>

      <!-- 列表：状态徽章 + 标题 + 描述 + 元信息 + 证书缩略图 -->
      <view v-else class="cl">
        <!-- 到期提醒条：30 天内到期/已过证书一屏提示，引导续证（纯站内，无推送依赖） -->
        <view v-if="expiringHint" class="cert-hint" :class="{ 'cert-hint--expired': expiredCount > 0 }">
          <text class="cert-hint-mark">!</text>
          <text class="cert-hint-text">{{ expiringHint }}</text>
        </view>
        <view
          v-for="item in list"
          :key="item.id"
          class="card"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="viewCert(item)"
        >
          <view class="c-main">
            <view class="c-badges">
              <text class="c-tag" :style="typeStyle(item.cert_type)">{{ typeLabel(item.cert_type) }}</text>
              <text class="c-st" :class="statusCls(item.status)">{{ statusLabel(item.status) }}</text>
<template v-if="warnBadge(item)">
<text class="c-st" :class="warnBadge(item).cls">{{ warnBadge(item).tag }}</text>
</template>
            </view>
            <text class="ct">{{ typeFull(item.cert_type) }}</text>
            <text v-if="item.cert_number" class="c-desc">编号：{{ item.cert_number }}</text>
            <view class="c-meta">
              <text v-if="item.issue_date">发证 {{ dateText(item.issue_date) }}</text>
              <text v-if="item.issue_date && (item.expire_date || item.expiry_date)" class="c-dot">·</text>
              <text v-if="item.expire_date || item.expiry_date">至 {{ dateText(item.expire_date || item.expiry_date) }}</text>
            </view>
          </view>
          <image v-if="certImage(item)" class="c-thumb" :src="certImage(item)" mode="aspectFill" />
        </view>
      </view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>

    <!-- 归档证书 弹层（标准 u-popup：遮罩/圆角/上滑动画与全局弹层一致） -->
    <u-popup :show="applyShow" round position="bottom" @close="closeApply">
      <view class="l-sheet" @tap.stop>
        <view class="l-head">
          <text class="l-title">归档证书</text>
          <text class="l-x" @tap="closeApply">×</text>
        </view>
        <view class="l-body">
          <view class="l-field">
            <text class="l-label">证书类型 <text class="l-req">*</text></text>
            <picker mode="selector" :range="certTypeLabels" :value="certTypeIdx" @change="onCertType">
              <view class="l-input" :class="{ empty: !af.cert_type }">{{ certTypeIdx >= 0 ? certTypeLabels[certTypeIdx] : '选择类型（CAAC / AOPA / UTC / ASFC / 人社）' }}</view>
            </picker>
          </view>
          <view class="l-field">
            <text class="l-label">证书编号 <text class="l-req">*</text></text>
            <input v-model="af.cert_number" class="l-input" placeholder="证书上的编号（用于查重）" />
          </view>
          <view class="l-field">
            <text class="l-label">证书等级（选填）</text>
            <input v-model="af.level" class="l-input" placeholder="如 III / 初级 / 高级" />
          </view>
          <view class="l-field">
            <text class="l-label">发证机构（选填）</text>
            <input v-model="af.issuer_org" class="l-input" placeholder="如 中国民用航空局" />
          </view>
          <view class="l-field-row">
            <view class="l-field half">
              <text class="l-label">发证日期 <text class="l-req">*</text></text>
              <picker mode="date" :value="af.issue_date" start="2000-01-01" :end="todayStr" @change="onIssueDate">
                <view class="l-input" :class="{ empty: !af.issue_date }">{{ af.issue_date || '选择日期' }}</view>
              </picker>
            </view>
            <view class="l-field half">
              <text class="l-label">有效期至（选填）</text>
              <picker mode="date" :value="af.expire_date" start="2000-01-01" @change="onExpireDate">
                <view class="l-input" :class="{ empty: !af.expire_date }">{{ af.expire_date || '长期有效' }}</view>
              </picker>
            </view>
          </view>
          <view class="l-field">
            <text class="l-label">证书照片（选填，建议上传）</text>
            <view class="l-upload" hover-class="l-upload--hover" @tap="chooseCertImg">
              <image v-if="af.image_url" :src="af.image_url" mode="aspectFill" class="l-upload-img" />
              <template v-else>
                <text class="l-upload-plus">＋</text>
                <text class="l-upload-tip">上传证书照片</text>
              </template>
            </view>
          </view>
          <text class="l-note">提交后由协会审核，审核通过且未过期即计入飞手/导师认证的有效证书。</text>
        </view>
        <view class="l-btn" hover-class="l-btn--hover" @tap="submitCert">{{ submittingCert ? '提交中...' : '提交申请' }}</view>
      </view>
    </u-popup>

    <!-- 证书文字详情（无证书图时的兜底：此前只弹一句"暂无证书图片"，等于点不开） -->
    <u-popup :show="detailShow" round position="bottom" @close="closeDetail">
      <view class="d-sheet" @tap.stop>
        <view class="d-head">
          <text class="d-title">证书详情</text>
          <text class="d-x" @tap="closeDetail">×</text>
        </view>
        <view class="d-body" v-if="activeCert">
          <view class="d-row"><text class="d-k">证书类型</text><text class="d-v">{{ typeFull(activeCert.cert_type) }}</text></view>
          <view class="d-row"><text class="d-k">证书编号</text><text class="d-v">{{ activeCert.cert_number || '未填写' }}</text></view>
          <view class="d-row" v-if="activeCert.level"><text class="d-k">等级</text><text class="d-v">{{ activeCert.level }}</text></view>
          <view class="d-row" v-if="activeCert.issuer_org"><text class="d-k">发证机构</text><text class="d-v">{{ activeCert.issuer_org }}</text></view>
          <view class="d-row"><text class="d-k">发证日期</text><text class="d-v">{{ activeCert.issue_date ? dateText(activeCert.issue_date) : '未填写' }}</text></view>
          <view class="d-row"><text class="d-k">有效期</text><text class="d-v">{{ activeCert.expire_date ? '至 ' + dateText(activeCert.expire_date) : '长期有效' }}</text></view>
          <view class="d-row"><text class="d-k">状态</text><text class="d-v">{{ statusLabel(activeCert.status) }}</text></view>
          <view class="d-row" v-if="activeCert.review_note"><text class="d-k">审核备注</text><text class="d-v">{{ activeCert.review_note }}</text></view>
          <text class="d-tip">该证书未上传照片，以上为登记信息；需补图请联系协会。</text>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh, onPageScroll } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage, getErrorMessage } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'
import { useReduceMotion } from '../../../utils/motion'

/* 证书类型 → 徽章短名（兼容后端枚举 caac/utc_dji/gov_level 与展示形 CAAC/UTC/人社） */
const CERT_TYPE_SHORT = {
  caac: 'CAAC', utc_dji: 'UTC', gov_level: '人社',
  aopa: 'AOPA', asfc: 'ASFC',
}
/* 证书类型 → 标题全名（沿用 pilots/detail、register 的产品措辞；未知类型回退短名） */
const CERT_TYPE_FULL = {
  caac: 'CAAC 执照', utc_dji: '大疆 UTC 认证', gov_level: '人社等级证书',
}
/* 证书类型 → tag 配色（深色字 + 浅底，与研发难题广场领域标签同构） */
const CERT_TYPE_STYLE = {
  caac: { color: '#0d47a1', bg: '#E3EDF9' },
  utc_dji: { color: '#4a148c', bg: '#F0E9F7' },
  gov_level: { color: '#B54708', bg: '#FDEEE4' },
  aopa: { color: '#004d40', bg: '#E4F2EF' },
  asfc: { color: '#1a237e', bg: '#E7E9F4' },
}
const CERT_TYPE_STYLE_DEFAULT = { color: '#344054', bg: '#EEF1F4' }

/* 状态 → 文案 / 状态色（对齐挑战广场语义：有效/通过=绿、审核中=蓝、过期/驳回=灰、吊销=红） */
const STATUS_LABEL = {
  active: '有效', expired: '已过期', pending: '审核中', revoked: '已吊销',
  approved: '已通过', rejected: '已驳回',
}
const STATUS_CLS = {
  active: 'st-open', approved: 'st-open',
  pending: 'st-pending',
  expired: 'st-closed', rejected: 'st-closed',
  revoked: 'st-err',
}

const loading = ref(false)
const errorMsg = ref('')
const list = ref([])
const detailShow = ref(false)   // 无图证书的文字详情弹层（BUG-002）
const activeCert = ref(null)
const statusBarHeight = ref(20)
const showBt = ref(false)
const { noMotion, checkMotion } = useReduceMotion()

const norm = (t) => String(t || '').toLowerCase()
const typeLabel = (t) => CERT_TYPE_SHORT[norm(t)] || t || '通用'
const typeFull = (t) => CERT_TYPE_FULL[norm(t)] || CERT_TYPE_SHORT[norm(t)] || t || '通用证书'
const typeStyle = (t) => CERT_TYPE_STYLE[norm(t)] || CERT_TYPE_STYLE_DEFAULT
const statusLabel = (s) => STATUS_LABEL[s] || s || '未知'
const statusCls = (s) => STATUS_CLS[s] || 'st-closed'
const dateText = (iso) => (iso ? String(iso).slice(0, 10) : '—')

/* 后端对"没填有效期"会下发哨兵日期（NULL 落成 Go 零值时间 → 0001-01-01T…）。
   直接拿它算天数会判成"已过期"，卡片还会显示"至 0001-01-01"（BUG-001）。
   这里统一按"早于 2000-01-01 = 没填"清空，展示与到期提醒都按"长期有效"处理。 */
const realDate = (v) => {
  const s = String(v || '').slice(0, 10)
  const y = Number(s.slice(0, 4))
  return /^\d{4}-\d{2}-\d{2}$/.test(s) && y >= 2000 ? v : ''
}

/* 到期提醒（站内）：到期日期 30 天内 → 即将到期（橙），已过 → 已过期（红）。
   以日期为准而非 status——后端审核态不随日期自动流转，日期提醒才不失效。 */
const EXPIRING_DAYS = 30
const daysToExpire = (item) => {
  const d = item && (item.expire_date || item.expiry_date || '')
  if (!d) return null
  const t = new Date(String(d).slice(0, 10).replace(/-/g, '/'))
  if (isNaN(t.getTime())) return null
  return Math.ceil((t.getTime() - Date.now()) / 86400000)
}
const warnBadge = (item) => {
  const n = daysToExpire(item)
  if (n === null || n === undefined) return null
  if (n < 0) return { tag: '已过期', cls: 'st-err' }
  if (n <= EXPIRING_DAYS) return { tag: n === 0 ? '今天到期' : '即将到期', cls: 'st-warn' }
  return null
}
const expiredCount = computed(() => list.value.filter((it) => { const n = daysToExpire(it); return n !== null && n < 0 }).length)
const expiringCount = computed(() => list.value.filter((it) => { const n = daysToExpire(it); return n !== null && n >= 0 && n <= EXPIRING_DAYS }).length)
const expiringHint = computed(() => {
  if (expiredCount.value > 0 && expiringCount.value > 0) return `您的 ${expiredCount.value} 本证书已过期、${expiringCount.value} 本将在 30 天内到期，请尽快联系协会续证`
  if (expiredCount.value > 0) return `您的 ${expiredCount.value} 本证书已过期，请尽快联系协会续证`
  if (expiringCount.value > 0) return `您的 ${expiringCount.value} 本证书将在 30 天内到期，请注意续证`
  return ''
})

/* 证书图：兼容 image_url / certificate_url / image / certificate 四类字段 */
const certImage = (item) => item && (item.image_url || item.certificate_url || item.image || item.certificate)

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/certificates/mine' })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    list.value = items.map((it) => Object.assign({}, it, {
      issue_date: realDate(it.issue_date),
      expire_date: realDate(it.expire_date || it.expiry_date),
      expiry_date: '',
    }))
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function viewCert(item) {
  // 有证书图 → 全屏预览；无图 → 弹文字详情（BUG-002：此前只弹一句"暂无证书图片"，等于点不开）
  const url = certImage(item)
  if (url) {
    uni.previewImage({ urls: [url], current: url })
    return
  }
  activeCert.value = item
  detailShow.value = true
}

function closeDetail() {
  detailShow.value = false
  activeCert.value = null
}

function goBack() { uni.navigateBack() }

/* ================= 归档证书 ================= */
const CERT_TYPE_OPTS = [
  { key: 'caac', label: 'CAAC 民航局执照' },
  { key: 'utc_dji', label: '大疆 UTC 认证' },
  { key: 'aopa', label: 'AOPA 执照' },
  { key: 'asfc', label: 'ASFC 执照' },
  { key: 'gov_level', label: '人社等级证书' },
]
const certTypeLabels = CERT_TYPE_OPTS.map((o) => o.label)
const applyShow = ref(false)
const submittingCert = ref(false)
const certTypeIdx = ref(-1)
const af = ref({ cert_type: '', cert_number: '', level: '', issuer_org: '', image_url: '', issue_date: '', expire_date: '' })
const todayStr = new Date().toISOString().slice(0, 10)

function openApply() {
  if (!requireLogin()) return
  applyShow.value = true
}
function closeApply() { applyShow.value = false }
function onCertType(e) {
  certTypeIdx.value = Number(e.detail.value)
  af.value.cert_type = CERT_TYPE_OPTS[certTypeIdx.value] ? CERT_TYPE_OPTS[certTypeIdx.value].key : ''
}
function onIssueDate(e) { af.value.issue_date = e.detail.value }
function onExpireDate(e) { af.value.expire_date = e.detail.value }

// 证书照片上传：uni.uploadFile → /api/v1/files/upload → /uploads/{file_id}
function chooseCertImg() {
  uni.chooseImage({
    count: 1,
    sourceType: ['album', 'camera'],
    success: (res) => uploadCertImg(res.tempFilePaths[0]),
  })
}
const uploadCertImg = async (filePath) => {
  const token = authStorage.getAccessToken()
  uni.showLoading({ title: '上传中...' })
  try {
    const data = await new Promise((resolve, reject) => {
      uni.uploadFile({
        url: BASE_URL + '/api/v1/files/upload',
        filePath,
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
    if (!fid) {
      uni.showToast({ title: '上传失败，请重试', icon: 'none' })
      return
    }
    af.value.image_url = '/uploads/' + fid
  } catch (e) {
    uni.showToast({ title: '上传失败，请重试', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

async function submitCert() {
  if (submittingCert.value) return
  if (!af.value.cert_type) return uni.showToast({ title: '请选择证书类型', icon: 'none' })
  if (!af.value.cert_number.trim()) return uni.showToast({ title: '请填写证书编号', icon: 'none' })
  if (!af.value.issue_date) return uni.showToast({ title: '请选择发证日期', icon: 'none' })
  submittingCert.value = true
  try {
    await request({ url: '/api/v1/certificates', method: 'POST', data: { ...af.value } })
    uni.showModal({
      title: '已提交',
      content: '证书已提交归档，协会审核通过后计入有效证书',
      showCancel: false,
      confirmText: '知道了',
      success: () => { closeApply(); fetchList() },
    })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '提交失败，请重试', icon: 'none', duration: 2500 })
  } finally {
    submittingCert.value = false
  }
}
function goCourses() { uni.navigateTo({ url: '/pkg-talent/pages/training/courses' }) }
function scrollToTop() { uni.pageScrollTo({ scrollTop: 0, duration: 300 }) }

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  if (!requireLogin()) return
  fetchList()
})

onPullDownRefresh(() => {
  fetchList().then(function () {
    uni.stopPullDownRefresh()
  })
})

onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})
</script>

<style>
page {
  background: #F5F8FC;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #fff; /* 整体设计统一：白底页 + 描边卡（与政策资讯/成果案例列表一致，此前 #F5F8FC 偏离） */
  padding-bottom: 40px;
}

/* ===== 白色板块 ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }

/* ===== 列表卡片（白上白：灰描边 + 极淡灰投影浮起；无左缘色条） ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 12px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}
.tap-scale { transform: scale(0.95); opacity: 0.9; }
.c-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 7px;
}
.c-badges { display: flex; gap: 6px; }
.c-tag, .c-st {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.c-st.st-open { color: #0B6B41; background: #E9F7F0; }
.c-st.st-pending { color: #0A66C2; background: #EAF3FB; }
.c-st.st-closed { color: #5D6B82; background: #EEF1F4; }
.c-st.st-err { color: #B42318; background: #FDECEC; }
.c-st.st-warn { color: #B54708; background: #FFF4E5; }

/* 到期提醒条 */
.cert-hint { display: flex; align-items: flex-start; gap: 12rpx; padding: 20rpx 24rpx; margin-bottom: 20rpx; border-radius: 12rpx; background: #FFF6E5; border: 1rpx solid #F5CD8A; }
.cert-hint--expired { background: #FEF3F2; border-color: #F0A99F; }
.cert-hint-mark { flex-shrink: 0; width: 36rpx; height: 36rpx; border-radius: 50%; background: #B54708; color: #fff; font-size: 24rpx; font-weight: 700; text-align: center; line-height: 36rpx; }
.cert-hint--expired .cert-hint-mark { background: #D92D20; }
.cert-hint-text { flex: 1; font-size: 24rpx; color: #7A3E0D; line-height: 1.6; }
.cert-hint--expired .cert-hint-text { color: #B42318; }
.ct {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-desc {
  font-size: 12.5px;
  color: #667085;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #667085;
  flex-wrap: wrap;
}
.c-dot { color: #DDE1E6; }
/* 证书缩略图：圆角小图，点击整卡预览原图 */
.c-thumb {
  flex: none;
  width: 60px;
  height: 60px;
  border-radius: 8px;
  background: #F4F6F8;
}

/* ===== 骨架 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; text-align: center; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 回到顶部 ===== */
.bt {
  position: fixed;
  bottom: 90px;
  right: 16px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 35;
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(0.16, 1, 0.3, 1);
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; }

/* ===================== 动效规范（对齐研发难题广场） =====================
   白名单：仅 transform / opacity（小尺寸颜色过渡允许）
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1) + ios-decel cubic-bezier(.32,.72,0,1)
   数量：列表入场仅错峰首屏 6 项，其余静置
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 列表入场：前 6 项每 20ms 依次淡入上移（backwards 填充 → 延迟期不闪跳） */
.card { animation: none; }
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 2) 交互反馈：卡片按压（快进慢出） */
.card { transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.card.tap-scale { transition-duration: .1s; transition-timing-function: linear; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }

/* 骨架呼吸（加载中环境光；循环动画 1.4s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .tap-scale { transform: none !important; }
.page.no-motion .stb:active,
.page.no-motion .bt:active { transform: none; }

/* ===== 证书文字详情弹层（无图兜底） ===== */
.d-sheet { width: 100%; background: #fff; border-radius: 24rpx 24rpx 0 0; padding-bottom: calc(24rpx + env(safe-area-inset-bottom)); }
.d-head { display: flex; align-items: center; justify-content: space-between; padding: 32rpx 32rpx 16rpx; }
.d-title { font-size: 32rpx; font-weight: 700; color: #17212B; }
.d-x { font-size: 40rpx; color: #667085; line-height: 1; padding: 0 8rpx; }
.d-body { padding: 0 32rpx 8rpx; }
.d-row { display: flex; align-items: flex-start; gap: 24rpx; padding: 20rpx 0; border-bottom: 1rpx solid #F0F1F3; }
.d-row:last-of-type { border-bottom: none; }
.d-k { flex: none; width: 140rpx; font-size: 26rpx; color: #667085; }
.d-v { flex: 1; min-width: 0; font-size: 28rpx; color: #17212B; font-weight: 500; word-break: break-all; }
.d-tip { display: block; margin-top: 16rpx; font-size: 24rpx; color: #667085; line-height: 1.6; }

/* ===== 归档证书入口 ===== */
.ir { display: flex; align-items: center; justify-content: space-between; }
.ir-btn {
  font-size: 26rpx;
  font-weight: 600;
  color: #0A66C2;
  padding: 10rpx 20rpx;
  border: 2rpx solid #BFD8F2;
  border-radius: 999rpx;
  background: #F0F7FE;
}
.ir-btn--hover { transform: scale(.96); opacity: .85; }

/* ===== 归档证书弹层（容器为 u-popup：遮罩/圆角/动画由组件提供，此处只管内容与间距） ===== */
.l-sheet {
  width: 100%;
  background: #fff;
  padding: 28rpx 28rpx calc(28rpx + env(safe-area-inset-bottom));
  max-height: 82vh;
  overflow-y: auto;
  box-sizing: border-box;
}
.l-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.l-title { font-size: 32rpx; font-weight: 700; color: #17212B; }
.l-x { font-size: 40rpx; color: #667085; padding: 0 8rpx; }
.l-body { display: flex; flex-direction: column; gap: 20rpx; }
.l-field-row { display: flex; gap: 16rpx; }
.l-field.half { flex: 1; min-width: 0; }
.l-label { display: block; font-size: 24rpx; font-weight: 600; color: #344054; margin-bottom: 10rpx; }
.l-req { color: #D92D20; }
.l-input {
  height: 84rpx;
  line-height: 84rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #17212B;
  background: #FAFAFA;
  border: 2rpx solid #E4E7EC;
  border-radius: 14rpx;
  box-sizing: border-box;
}
.l-input.empty { color: #98A2B3; }
.l-upload {
  width: 220rpx; height: 160rpx;
  border: 2rpx dashed #D0D5DD;
  border-radius: 14rpx;
  background: #FAFAFA;
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8rpx;
  overflow: hidden;
}
.l-upload--hover { opacity: .8; }
.l-upload-plus { font-size: 44rpx; color: #0A66C2; line-height: 1; }
.l-upload-tip { font-size: 22rpx; color: #98A2B3; }
.l-upload-img { width: 100%; height: 100%; }
.l-note { font-size: 22rpx; color: #98A2B3; line-height: 1.6; }
.l-btn {
  margin-top: 28rpx;
  height: 88rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 30rpx;
  font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.l-btn--hover { opacity: .9; }
</style>
