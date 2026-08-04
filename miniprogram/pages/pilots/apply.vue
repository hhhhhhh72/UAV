<template>
  <view class="apply-page">
    <u-nav-bar title="申请认证飞手" show-back @back="goBack" />

    <!-- 头部品牌区 -->
    <view class="hero">
      <view class="hero-glow" />
      <text class="hero-title">申请认证飞手</text>
      <text class="hero-sub">提交资质信息，经协会审核通过后展示在认证飞手名录</text>
      <view class="hero-steps">
        <view v-for="(s, i) in steps" :key="i" class="step">
          <view class="step-dot">{{ i + 1 }}</view>
          <text class="step-label">{{ s }}</text>
        </view>
      </view>
    </view>

    <!-- 个人信息 -->
    <view class="section">
      <view class="section-title">
        <view class="section-dot" />
        <text>个人信息</text>
        <text class="section-note">用于资质审核</text>
      </view>
      <u-cell-group inset>
        <view class="form-wrap">
          <view class="field-row">
            <u-field v-model="form.real_name" label="真实姓名" placeholder="请输入真实姓名" />
            <text class="required">*</text>
          </view>
          <view class="field-row">
            <u-field v-model="form.id_card" label="身份证号" placeholder="请输入身份证号" type="idcard" />
            <text class="required">*</text>
          </view>
        </view>
      </u-cell-group>
    </view>

    <!-- 资质信息 -->
    <view class="section">
      <view class="section-title">
        <view class="section-dot" />
        <text>资质信息</text>
        <text class="section-note">选填，展示在名录</text>
      </view>
      <u-cell-group inset>
        <view class="form-wrap">
          <view class="field-row">
            <u-field v-model="form.flight_hours" label="飞行时长" placeholder="累计飞行小时" type="digit" />
            <text class="unit">小时</text>
          </view>
          <u-field v-model="form.bio" label="擅长领域" placeholder="如：电力巡检 / 测绘航拍" />
        </view>
      </u-cell-group>
    </view>

    <!-- 证书自动关联 -->
    <view class="section">
      <view class="section-title">
        <view class="section-dot" />
        <text>证书认证</text>
        <text class="section-note">自动关联</text>
      </view>
      <view class="cert-card">
        <template v-if="loadingCerts">
          <u-loading size="24rpx" />
          <text class="cert-loading">读取证书中...</text>
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
    <view class="privacy">
      <text class="privacy-tag">隐私</text>
      <text class="privacy-text">身份证号加密存储，名录中自动脱敏展示，仅协会审核可见</text>
    </view>

    <view class="submit-wrap">
      <u-button type="primary" size="large" round :loading="submitting" @tap="submit">
        提交申请
      </u-button>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

const goBack = () => uni.navigateBack()
const steps = ['提交申请', '协会审核', '名录展示']
const form = ref({ real_name: '', id_card: '', flight_hours: '', bio: '' })
const submitting = ref(false)

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
  if (!form.value.real_name.trim()) return uni.showToast({ title: '请输入真实姓名', icon: 'none' })
  if (!form.value.id_card.trim()) return uni.showToast({ title: '请输入身份证号', icon: 'none' })
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

onLoad(loadCerts)
</script>

<style scoped>
.apply-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 40rpx; }

/* 头部品牌区 */
.hero {
  position: relative;
  overflow: hidden;
  padding: 36rpx 32rpx 40rpx;
  background: linear-gradient(135deg, var(--color-primary-deep) 0%, var(--color-primary) 100%);
}
.hero-glow {
  position: absolute;
  top: -120rpx;
  right: -80rpx;
  width: 360rpx;
  height: 360rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(29,212,168,.15), transparent 65%);
  pointer-events: none;
}
.hero-title { font-size: 40rpx; font-weight: 700; color: #fff; display: block; position: relative; }
.hero-sub { font-size: 24rpx; color: rgba(255,255,255,.7); margin-top: 12rpx; display: block; line-height: 1.6; position: relative; }
.hero-steps { display: flex; gap: 24rpx; margin-top: 28rpx; position: relative; }
.step { display: flex; align-items: center; gap: 10rpx; }
.step-dot {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: rgba(255,255,255,.18);
  color: #fff;
  font-size: 22rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.step-label { font-size: 22rpx; color: rgba(255,255,255,.8); }

/* 分区 */
.section { margin-top: 28rpx; }
.section-title { display: flex; align-items: center; gap: 12rpx; padding: 0 32rpx 16rpx; }
.section-dot { width: 12rpx; height: 12rpx; border-radius: 4rpx; background: var(--color-primary); }
.section-title text:nth-child(2) { font-size: 28rpx; font-weight: 700; color: var(--color-text); }
.section-note { font-size: 22rpx; color: var(--color-text-placeholder); margin-left: auto; }
.form-wrap { padding: 24rpx 0; }
.field-row { position: relative; }
.required { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-danger); font-size: 28rpx; z-index: 2; }
.unit { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-secondary); font-size: 26rpx; z-index: 2; }

/* 证书卡 */
.cert-card {
  margin: 0 24rpx;
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 28rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  box-shadow: 0 3px 12px rgba(16,24,40,.05);
}
.cert-loading { font-size: 24rpx; color: var(--color-text-secondary); display: flex; align-items: center; gap: 12rpx; }
.cert-head { display: flex; align-items: baseline; gap: 16rpx; }
.cert-num { font-size: 44rpx; font-weight: 800; color: var(--color-success); }
.cert-desc { font-size: 22rpx; color: var(--color-text-secondary); line-height: 1.6; flex: 1; }
.cert-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }
.cert-tag {
  font-size: 20rpx;
  padding: 4rpx 16rpx;
  border-radius: 4px;
  background: var(--color-success-soft, #E9F7F0);
  color: var(--color-success);
  font-weight: 600;
}
.cert-empty { font-size: 24rpx; color: var(--color-text-secondary); line-height: 1.6; }

/* 隐私说明 + 提交 */
.privacy { display: flex; align-items: flex-start; gap: 12rpx; padding: 24rpx 40rpx 0; }
.privacy-tag {
  font-size: 20rpx;
  padding: 2rpx 12rpx;
  border-radius: 4px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: 600;
  flex-shrink: 0;
}
.privacy-text { font-size: 22rpx; color: var(--color-text-placeholder); line-height: 1.6; }
.submit-wrap { padding: 40rpx 32rpx 0; }
</style>
