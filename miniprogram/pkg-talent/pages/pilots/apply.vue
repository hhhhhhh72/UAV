<template>
  <view class="pub-page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <!-- 顶栏：与名录/详情同款 u-nav-bar（内建胶囊避让与冷启动回首页兜底） -->
    <u-nav-bar title="申请认证飞手" show-back :fixed="true" @back="goBack" />

    <!-- 表单头部 -->
    <view class="pub-form-intro">
      <view class="pub-form-intro-h2">申请认证飞手</view>
      <view class="pub-form-intro-p">提交资质信息，经协会审核通过后展示在认证飞手名录</view>
    </view>

    <!-- 上次被驳回：放行重提，但先说清原因 -->
    <view v-if="rejectedReason" class="pub-section">
      <view class="reject-notice">
        <view class="reject-notice-head">
          <text class="reject-notice-dot">!</text>
          <text class="reject-notice-title">上次申请未通过</text>
        </view>
        <text class="reject-notice-body">{{ rejectedReason }}</text>
        <text class="reject-notice-tip">补充或更正材料后，可直接在下方重新提交。</text>
      </view>
    </view>

    <!-- 个人信息 -->
    <view class="pub-section">
      <view class="pub-section-title">个人信息</view>
      <view class="pub-section-note">用于资质审核</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">头像</view>
          <view class="pub-upload-row avatar-row">
            <view v-if="form.avatar" class="pub-photo" @tap="chooseAvatar">
              <image :src="avatarPreview" mode="aspectFill" class="pub-photo-img" />
            </view>
            <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="chooseAvatar">＋</view>
            <text class="avatar-hint">选填，展示在名录</text>
          </view>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">真实姓名<text class="pub-required">*</text></view>
          <input
            v-model="form.real_name"
            class="pub-input"
            placeholder="请输入真实姓名"
            placeholder-class="pub-placeholder"
          />
        </view>
        <view class="pub-field">
          <view class="pub-field-label">身份证号<text class="pub-required">*</text></view>
          <input
            v-model="form.id_card"
            class="pub-input"
            type="idcard"
            placeholder="请输入身份证号"
            placeholder-class="pub-placeholder"
          />
        </view>
      </view>
    </view>

    <!-- 资质信息 -->
    <view class="pub-section">
      <view class="pub-section-title">资质信息</view>
      <view class="pub-section-note">选填，展示在名录</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">飞行时长</view>
          <input
            v-model="form.flight_hours"
            class="pub-input"
            type="digit"
            placeholder="累计飞行小时"
            placeholder-class="pub-placeholder"
          />
          <text class="pub-field-hint">小时</text>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">擅长领域</view>
          <input
            v-model="form.bio"
            class="pub-input"
            placeholder="如：电力巡检 / 测绘航拍"
            placeholder-class="pub-placeholder"
          />
        </view>
        <view class="pub-field">
          <view class="pub-field-label">所在地区</view>
          <picker :range="chongqingDistricts" @change="onDistrictChange">
            <view class="pub-select-field">
              <text :class="form.region ? 'pub-select-value' : 'pub-placeholder'">{{ form.region || '请选择区县（选填）' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </picker>
        </view>
      </view>
    </view>

    <!-- 证书材料：**随申请一并提交，协会审一次即同时裁定**（合并审核）。
         此前证书必须先单独提交、由管理端审核通过后才允许申请飞手认证——
         同一批证据走两道人工审核，而本页只能显示一句"请先提交证书"，没有填写位置。 -->
    <view class="pub-section">
      <view class="pub-section-title">证书材料</view>
      <view class="pub-section-note">随申请一并提交，协会审一次即同时裁定</view>
      <view class="cert-card">
        <view v-if="loadingCerts" class="cert-loading">
          <u-loading size="24rpx" />
          <text>读取证书中...</text>
        </view>
        <template v-else>
          <!-- 已备案且有效：协会已经审过，本次不必重复提交 -->
          <view v-if="approvedCerts.length" class="cert-approved">
            <text class="cert-approved-tip">已备案有效证书 {{ approvedCerts.length }} 张，无需重复提交</text>
            <view class="cert-tags">
              <text v-for="c in approvedCerts" :key="c.id" class="cert-tag">{{ certTypeLabel(c.cert_type) }}</text>
            </view>
            <text class="cert-manage" @tap="goMyCerts">管理已归档证书 ›</text>
          </view>

          <!-- 待随申请提交的证书 -->
          <view v-for="(cf, i) in certForms" :key="'cf' + i" class="cert-form">
            <view class="cert-form-head">
              <text class="cert-form-title">证书 {{ i + 1 }}</text>
              <text class="cert-form-del" @tap="removeCertForm(i)">删除</text>
            </view>
            <view class="cf-row">
              <text class="cf-label">类型<text class="cf-req">*</text></text>
              <picker class="cf-picker" mode="selector" :range="certTypeLabels" :value="certTypeIdx[i]" @change="onCertType(i, $event)">
                <view class="cf-input" :class="{ empty: !cf.cert_type }">{{ cf.cert_type ? certTypeLabel(cf.cert_type) : '请选择证书类型' }}</view>
              </picker>
            </view>
            <view class="cf-row">
              <text class="cf-label">编号<text class="cf-req">*</text></text>
              <input v-model="cf.cert_number" class="cf-input" placeholder="证书上的编号（用于查重）" />
            </view>
            <view class="cf-row">
              <text class="cf-label">发证机构</text>
              <input v-model="cf.issuer_org" class="cf-input" placeholder="如 中国民用航空局" />
            </view>
            <view class="cf-row">
              <text class="cf-label">发证日期<text class="cf-req">*</text></text>
              <picker class="cf-picker" mode="date" :value="cf.issue_date" start="2000-01-01" :end="todayStr" @change="onCertDate(i, 'issue_date', $event)">
                <view class="cf-input" :class="{ empty: !cf.issue_date }">{{ cf.issue_date || '请选择' }}</view>
              </picker>
            </view>
            <view class="cf-row">
              <text class="cf-label">有效期至</text>
              <picker class="cf-picker" mode="date" :value="cf.expire_date" start="2000-01-01" @change="onCertDate(i, 'expire_date', $event)">
                <view class="cf-input" :class="{ empty: !cf.expire_date }">{{ cf.expire_date || '长期有效' }}</view>
              </picker>
            </view>
            <view class="cf-row cf-row--upload">
              <text class="cf-label">证书照片</text>
              <view class="cf-upload" hover-class="pub-fade" @tap="chooseCertImage(i)">
                <image v-if="cf.image_url" :src="cf.image_url" mode="aspectFill" class="cf-upload-img" />
                <template v-else>
                  <text class="cf-upload-plus">＋</text>
                  <text class="cf-upload-tip">建议上传</text>
                </template>
              </view>
            </view>
          </view>

          <view class="cert-add" hover-class="pub-fade" @tap="addCertForm">＋ 添加证书</view>
          <text v-if="certGateTip" class="cert-gate-tip">{{ certGateTip }}</text>
        </template>
      </view>
    </view>

    <!-- 隐私说明 -->
    <view class="privacy-note">
      <text class="privacy-tag">隐私</text>
      <text class="privacy-text">身份证号加密存储，名录中自动脱敏展示，仅协会审核可见</text>
    </view>

    <!-- 固定底部操作区（与发布页同款） -->
    <view class="pub-sticky">
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="submit">
        {{ submitting ? '提交中...' : '提交申请' }}
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage, getErrorMessage, uploadFileWithAuth } from '../../../utils/request'
import { requireLogin, safeBack } from '../../../utils/nav'

// 顶栏改用 u-nav-bar（与名录/详情同一套）：页面只需让出「状态栏 + 44px 导航高度」
const statusBarHeight = ref(20)

const goBack = () => safeBack()
const form = ref({ real_name: '', id_card: '', flight_hours: '', bio: '', avatar: '', region: '' })
const submitting = ref(false)
// 上次申请的驳回原因（rejected 时展示；放行重提但必须说明原因）
const rejectedReason = ref('')

// 所在地区（重庆 38 个区县，与研学/培训列表一致）
const chongqingDistricts = ['渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区', '南岸区', '北碚区', '渝北区', '巴南区', '两江新区', '长寿区', '江津区', '合川区', '永川区', '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区', '武隆区', '万州区', '黔江区', '涪陵区', '奉节县', '云阳县', '忠县', '垫江县', '丰都县', '城口县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县']
const onDistrictChange = (e) => { form.value.region = chongqingDistricts[Number(e.detail.value)] || '' }

// 头像上传：uni.uploadFile → /api/v1/files/upload，保存为 /uploads/{file_id} 路径
const avatarPreview = ref('')
const chooseAvatar = () => {
  uni.chooseImage({
    count: 1,
    sourceType: ['album', 'camera'],
    success: (res) => uploadAvatar(res.tempFilePaths[0]),
  })
}
const uploadAvatar = async (filePath) => {
  if (!requireLogin()) return
  uni.showLoading({ title: '上传中...' })
  try {
    // 带鉴权上传（401 自动刷新重试）
    const url = await uploadFileWithAuth(filePath)
    form.value.avatar = url.startsWith('/uploads/') ? url : ('/uploads/' + url.split('/').pop())
    avatarPreview.value = filePath
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '上传失败，请重试', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

// 证书材料：已备案的（只读展示）+ 随本次申请提交的（可增删）
const loadingCerts = ref(false)
const approvedCerts = ref([])
const certTypeLabel = (t) => ({ caac: 'CAAC 执照', utc_dji: '大疆 UTC', aopa: 'AOPA 执照', asfc: 'ASFC 执照', gov_level: '人社等级' }[t] || t || '证书')

// 与「我的证书」页同一套类型选项，避免两处口径不一致
const CERT_TYPE_OPTS = [
  { key: 'caac', label: 'CAAC 民航局执照' },
  { key: 'utc_dji', label: '大疆 UTC 认证' },
  { key: 'aopa', label: 'AOPA 执照' },
  { key: 'asfc', label: 'ASFC 执照' },
  { key: 'gov_level', label: '人社等级证书' },
]
const certTypeLabels = CERT_TYPE_OPTS.map((o) => o.label)

const newCertForm = () => ({ cert_type: '', cert_number: '', level: '', issuer_org: '', image_url: '', issue_date: '', expire_date: '' })
const certForms = ref([])
const certTypeIdx = ref([])
const addCertForm = () => {
  certForms.value.push(newCertForm())
  certTypeIdx.value.push(-1)
}
const removeCertForm = (i) => {
  certForms.value.splice(i, 1)
  certTypeIdx.value.splice(i, 1)
}
const onCertType = (i, e) => {
  const idx = Number(e.detail.value)
  certTypeIdx.value[i] = idx
  const o = CERT_TYPE_OPTS[idx]
  if (certForms.value[i]) certForms.value[i].cert_type = o ? o.key : ''
}
const onCertDate = (i, field, e) => {
  if (certForms.value[i]) certForms.value[i][field] = e.detail.value
}
const chooseCertImage = (i) => {
  if (!requireLogin()) return
  uni.chooseImage({
    count: 1,
    sourceType: ['album', 'camera'],
    success: async (res) => {
      uni.showLoading({ title: '上传中...' })
      try {
        const url = await uploadFileWithAuth(res.tempFilePaths[0])
        if (certForms.value[i]) certForms.value[i].image_url = url
      } catch (e) {
        uni.showToast({ title: getErrorMessage(e) || '上传失败，请重试', icon: 'none' })
      } finally {
        uni.hideLoading()
      }
    },
  })
}

// 门禁提示：既没有已备案的有效证书，也没有待提交的证书
const certGateTip = computed(() => {
  if (loadingCerts.value) return ''
  if (approvedCerts.value.length > 0 || certForms.value.length > 0) return ''
  return '至少提交 1 张证书（CAAC / AOPA / 大疆 UTC 等）才能申请飞手认证'
})

const loadCerts = async () => {
  loadingCerts.value = true
  try {
    const res = await request({ url: '/api/v1/certificates/mine' })
    const list = Array.isArray(res) ? res : ((res && res.data) || [])
    approvedCerts.value = list.filter((c) => c.status === 'approved')
    // 没有已备案的有效证书 → 默认给一个待提交表单。
    // 合并审核后证书就在本页一起填，不再要求先去别处"提交证书再回来"。
    if (!approvedCerts.value.length && !certForms.value.length) addCertForm()
  } catch (e) {} finally {
    loadingCerts.value = false
  }
}

// 去「我的证书」查看/管理已归档证书（本页已可直接提交，这里只作补充入口）
const goMyCerts = () => {
  uni.navigateTo({ url: '/pkg-talent/pages/training/certificates' })
}

const submit = async () => {
  if (!requireLogin()) return
  if (submitting.value) return
  if (!form.value.real_name.trim()) return uni.showToast({ title: '请输入真实姓名', icon: 'none' })
  if (!form.value.id_card.trim()) return uni.showToast({ title: '请输入身份证号', icon: 'none' })
  if (!/^\d{17}[\dXx]$/.test(form.value.id_card.trim())) return uni.showToast({ title: '身份证号格式不正确', icon: 'none' })
  // 门禁：无证不批（后端同规则）。合并审核后，证书可以**随申请一并提交**，
  // 所以"本次带了证书"与"已有备案有效证书"二者其一即可，不再强制先去别处提交。
  if (!approvedCerts.value.length && !certForms.value.length) {
    return uni.showToast({ title: '请至少提交 1 张证书（CAAC / AOPA / 大疆 UTC 等）', icon: 'none', duration: 2500 })
  }
  // 逐张校验随申请提交的证书：字段与 POST /api/v1/certificates 一致
  const certs = []
  for (let i = 0; i < certForms.value.length; i++) {
    const c = certForms.value[i]
    const n = i + 1
    if (!c.cert_type) return uni.showToast({ title: '请选择第 ' + n + ' 张证书的类型', icon: 'none' })
    if (!String(c.cert_number || '').trim()) return uni.showToast({ title: '请填写第 ' + n + ' 张证书的编号', icon: 'none' })
    if (!c.issue_date) return uni.showToast({ title: '请选择第 ' + n + ' 张证书的发证日期', icon: 'none' })
    certs.push({
      cert_type: c.cert_type,
      cert_number: String(c.cert_number).trim(),
      level: String(c.level || '').trim(),
      issuer_org: String(c.issuer_org || '').trim(),
      image_url: c.image_url || '',
      issue_date: c.issue_date,
      expire_date: c.expire_date || '',
    })
  }
  // 已备案有效证书 + 本次提交的证书，至少要有一张
  if (!approvedCerts.value.length && !certs.length) {
    return uni.showToast({ title: '请至少提交 1 张证书', icon: 'none' })
  }
  submitting.value = true
  try {
    await request({
      url: '/api/v1/certified-pilots',
      method: 'POST',
      data: {
        real_name: form.value.real_name.trim(),
        id_card: form.value.id_card.trim(),
        flight_hours: Number(form.value.flight_hours) || 0,
        bio: form.value.bio.trim(),
        avatar: form.value.avatar,
        region: form.value.region,
        certs,
      },
    })
    uni.showModal({
      title: '申请已提交',
      content: '协会审核通过后，您将展示在认证飞手名录中',
      showCancel: false,
      confirmText: '知道了',
      success: () => uni.navigateBack(),
    })
  } catch (e) {
    // 提取后端 403/400 的明确提示（如"需要至少一张未过期的有效证书"），避免只显示"申请失败"
    uni.showToast({ title: getErrorMessage(e) || '申请失败，请重试', icon: 'none', duration: 2500 })
  } finally {
    submitting.value = false
  }
}

// 已认证 / 审核中直接提示（驳回状态放行，可重提）；已认证给「查看档案」入口
onLoad(async () => {
  try {
    const sys = uni.getSystemInfoSync()
    if (sys && sys.statusBarHeight) statusBarHeight.value = sys.statusBarHeight
  } catch (e) { /* 保持默认 20 */ }
  loadCerts()
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = res && res.data ? res.data : res
    // 被驳回：放行进表单允许重提，但把上次的驳回原因摆出来。
    // 此前这里只拦 pending/approved，驳回直接落到一张空白表单，
    // 用户不知道上次为什么没过、该改什么。
    if (mine && mine.id && mine.status === 'rejected') {
      rejectedReason.value = String(mine.reject_reason || '').trim()
    }
    if (mine && mine.id && (mine.status === 'pending' || mine.status === 'approved')) {
      if (mine.status === 'approved') {
        uni.showModal({
          title: '已通过认证',
          content: '您已通过飞手认证，展示在认证飞手名录中',
          confirmText: '查看档案',
          success: (r) => {
            uni.removeStorageSync('pilot_detail')
            uni.redirectTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(mine.id) })
          },
        })
        return
      }
      uni.showModal({
        title: '无需重复申请',
        content: '当前状态：待审核，请耐心等待协会审核',
        showCancel: false,
        confirmText: '知道了',
        success: () => uni.navigateBack(),
      })
    }
  } catch (e) { /* 未登录等：放行到表单 */ }
})
</script>

<style scoped>
@import '../../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }

/* 驳回提示条：与页面其他卡片同一套圆角/描边语言，用警示色区分 */
.reject-notice {
  padding: 20rpx 24rpx;
  border: 2rpx solid #F5C9A8;
  border-radius: 16rpx;
  background: #FFF6EF;
}
.reject-notice-head { display: flex; align-items: center; gap: 10rpx; margin-bottom: 10rpx; }
.reject-notice-dot {
  width: 32rpx;
  height: 32rpx;
  flex: 0 0 32rpx;
  border-radius: 50%;
  background: var(--color-accent-deep, #E96012);
  color: #fff;
  font-size: 22rpx;
  font-weight: 700;
  text-align: center;
  line-height: 32rpx;
}
.reject-notice-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.reject-notice-body { display: block; font-size: 24rpx; color: #B54708; line-height: 1.5; }
.reject-notice-tip { display: block; margin-top: 8rpx; font-size: 22rpx; color: #9A6A45; line-height: 1.5; }
.pub-form-intro-h2 {
  font-size: 20px;
  margin: 0 0 4px;
  color: #17212B;
}
.pub-form-intro-p {
  font-size: 12px;
  color: #667085;
  margin: 0;
  line-height: 1.5;
}
.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
}

/* 头像上传（嵌入 pub-field，去掉行内自带 padding） */
.avatar-row { padding: 0; }
.avatar-hint {
  color: #98A2B3;
  font-size: 11px;
  line-height: 1.45;
}

/* 证书自动关联卡（浅蓝底蓝字，对齐 pub 色板） */
.cert-card {
  background: #EAF3FB;
  border-radius: 9px;
  padding: 13px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.cert-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #17212B;
  font-size: 12px;
}
.cert-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.cert-num {
  font-size: 20px;
  font-weight: 750;
  color: #0A66C2;
}
.cert-desc {
  font-size: 12px;
  color: #17212B;
  line-height: 1.6;
  flex: 1;
}
.cert-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.cert-tag {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 5px;
  background: #fff;
  color: #0A66C2;
  font-weight: 700;
}
/* ── 已备案有效证书（只读展示） ── */
.cert-approved {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.cert-approved-tip {
  font-size: 12px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.5;
}

/* ── 随申请提交的证书：每张一张白卡，与浅蓝底区分开 ── */
.cert-form {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.cert-form-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.cert-form-title {
  font-size: 13px;
  font-weight: 750;
  color: #0A66C2;
}
.cert-form-del {
  font-size: 12px;
  color: #D92D20;
  padding: 4px 2px;
}
.cf-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.cf-label {
  width: 68px;
  flex: 0 0 68px;
  font-size: 12px;
  color: #5B6B7C;
}
.cf-req { color: #D92D20; margin-left: 2px; }
.cf-picker { flex: 1; min-width: 0; }
.cf-input {
  font-size: 13px;
  color: #17212B;
  border-bottom: 1px solid #E4EAF2;
  padding: 6px 0;
  min-height: 32px;
  box-sizing: border-box;
}
.cf-input.empty { color: #98A2B3; }
.cf-row--upload { align-items: flex-start; }
.cf-upload {
  width: 72px;
  height: 72px;
  flex: 0 0 72px;
  border-radius: 8px;
  background: #F0F5FA;
  border: 1px dashed #C9DFF5;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  overflow: hidden;
}
.cf-upload-img { width: 100%; height: 100%; }
.cf-upload-plus { font-size: 20px; color: #0A66C2; line-height: 1; }
.cf-upload-tip { font-size: 10px; color: #5B6B7C; }
.cert-add {
  align-self: flex-start;
  font-size: 13px;
  font-weight: 700;
  color: #0A66C2;
  padding: 8px 0;
}
.cert-manage {
  align-self: flex-start;
  font-size: 12px;
  color: #0A66C2;
  font-weight: 600;
  padding: 2px 0;
}
.cert-gate-tip {
  font-size: 12px;
  color: #B54708;
  line-height: 1.6;
}

/* 隐私说明 */
.privacy-note {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: 0 4px 13px;
}
.privacy-tag {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 5px;
  background: #EAF3FB;
  color: #0A66C2;
  font-weight: 700;
  flex-shrink: 0;
}
.privacy-text {
  font-size: 11px;
  color: #98A2B3;
  line-height: 1.6;
}
</style>
