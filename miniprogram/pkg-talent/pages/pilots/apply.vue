<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">申请认证飞手</view>
    </view>

    <!-- 表单头部 -->
    <view class="pub-form-intro">
      <view class="pub-form-intro-h2">申请认证飞手</view>
      <view class="pub-form-intro-p">提交资质信息，经协会审核通过后展示在认证飞手名录</view>
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

    <!-- 证书自动关联 -->
    <view class="pub-section">
      <view class="pub-section-title">证书认证</view>
      <view class="pub-section-note">自动关联</view>
      <view class="cert-card">
        <template v-if="loadingCerts">
          <view class="cert-loading">
            <u-loading size="24rpx" />
            <text>读取证书中...</text>
          </view>
        </template>
        <template v-else-if="approvedCerts.length">
          <view class="cert-head">
            <text class="cert-num">{{ approvedCerts.length }}</text>
            <text class="cert-desc">项已认证证书将自动关联至您的飞手档案，审核时一并核验</text>
          </view>
          <view class="cert-tags">
            <text v-for="c in approvedCerts" :key="c.id" class="cert-tag">{{ certTypeLabel(c.cert_type) }}</text>
          </view>
        </template>
        <template v-else>
          <text class="cert-empty">暂无可关联的已认证证书，提交后仍可审核（证书可后续补充）</text>
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
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage } from '../../../utils/request'
import { requireLogin } from '../../../utils/nav'
import { useSafeTop } from '../../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

const goBack = () => uni.navigateBack()
const form = ref({ real_name: '', id_card: '', flight_hours: '', bio: '', avatar: '', region: '' })
const submitting = ref(false)

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
    form.value.avatar = '/uploads/' + fid
    avatarPreview.value = filePath
  } catch (e) {
    uni.showToast({ title: '上传失败，请重试', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

// 证书自动关联：读取我的已认证证书（approved）
const loadingCerts = ref(false)
const approvedCerts = ref([])
const certTypeLabel = (t) => ({ caac: 'CAAC 执照', utc_dji: '大疆 UTC', gov_level: '人社等级' }[t] || t || '证书')

const loadCerts = async () => {
  loadingCerts.value = true
  try {
    const res = await request({ url: '/api/v1/certificates/mine' })
    const list = Array.isArray(res) ? res : ((res && res.data) || [])
    approvedCerts.value = list.filter((c) => c.status === 'approved')
  } catch (e) {} finally {
    loadingCerts.value = false
  }
}

const submit = async () => {
  if (!requireLogin()) return
  if (submitting.value) return
  if (!form.value.real_name.trim()) return uni.showToast({ title: '请输入真实姓名', icon: 'none' })
  if (!form.value.id_card.trim()) return uni.showToast({ title: '请输入身份证号', icon: 'none' })
  if (!/^\d{17}[\dXx]$/.test(form.value.id_card.trim())) return uni.showToast({ title: '身份证号格式不正确', icon: 'none' })
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
    uni.showToast({ title: (e && e.message) || '申请失败，请重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

// 已认证 / 审核中直接提示（驳回状态放行，可重提）；已认证给「查看档案」入口
onLoad(async () => {
  initSafeTop()
  loadCerts()
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = res && res.data ? res.data : res
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
  background: #E8F2FC;
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
  color: #315776;
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
  color: #315776;
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
.cert-empty {
  font-size: 12px;
  color: #315776;
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
  background: #E8F2FC;
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
