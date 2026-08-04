<template>
  <view class="apply-page">
    <u-nav-bar title="申请认证飞手" show-back @back="goBack" />

    <u-cell-group inset>
      <view class="form-wrap">
        <u-field
          v-model="form.real_name"
          label="真实姓名"
          placeholder="请输入真实姓名"
        />
        <u-field
          v-model="form.id_card"
          label="身份证号"
          placeholder="用于资质审核，加密存储"
          type="idcard"
        />
        <view class="field-row">
          <u-field
            v-model="form.flight_hours"
            label="飞行时长"
            placeholder="累计飞行小时（选填）"
            type="digit"
          />
          <text class="unit">小时</text>
        </view>
        <u-field
          v-model="form.bio"
          label="擅长领域"
          placeholder="如：电力巡检 / 测绘航拍（选填）"
        />
      </view>
    </u-cell-group>

    <view class="note">证书将自动关联您已认证的证书（无需手动填写），审核通过后展示在认证飞手名录。</view>

    <view class="submit-wrap">
      <u-button type="primary" size="large" round :loading="submitting" @tap="submit">
        提交申请
      </u-button>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request } from '../../utils/request'

const goBack = () => uni.navigateBack()
const form = ref({ real_name: '', id_card: '', flight_hours: '', bio: '' })
const submitting = ref(false)

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
</script>

<style scoped>
.apply-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 40rpx; }
.form-wrap { padding: 24rpx 0; }
.field-row { position: relative; }
.unit { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-secondary); font-size: 26rpx; z-index: 2; }
.note { padding: 0 32rpx; font-size: 22rpx; color: var(--color-text-placeholder); line-height: 1.6; }
.submit-wrap { padding: 32rpx 24rpx; }
</style>
