<template>
  <view class="publish-page">
    <u-nav-bar title="发布招聘" show-back @back="goBack" />

    <!-- 非企业账号：引导入驻 -->
    <view v-if="!canPost" class="gate">
      <view class="gate-ico">聘</view>
      <text class="gate-title">仅企业账号可发布招聘</text>
      <text class="gate-note">发布招聘需要企业身份，请先完成企业入驻后重试。</text>
      <u-button type="primary" size="medium" round @tap="goRegister">去企业入驻</u-button>
    </view>

    <template v-else>
      <u-cell-group inset>
        <view class="form-wrap">
          <u-field
            v-model="form.title"
            label="职位名称"
            placeholder="如：无人机飞手"
          />
          <u-field v-model="form.location" label="工作地点" placeholder="如：重庆·渝北" />

          <view class="field-row">
            <u-field
              v-model="form.salary"
              label="薪资"
              placeholder="月薪（元）"
              type="digit"
            />
            <text class="unit">元/月</text>
          </view>

          <u-field
            v-model="form.description"
            label="职位描述"
            placeholder="岗位职责 / 任职要求"
            type="textarea"
          />
        </view>
      </u-cell-group>

      <view class="submit-wrap">
        <u-button type="primary" size="large" round :loading="submitting" @tap="submit">
          发布招聘
        </u-button>
        <text class="submit-note">发布后为草稿状态，可联系协会或在我的招聘中发布上线。</text>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { request, getStoredUser } from '@/utils/request'

const goBack = () => uni.navigateBack()
const goRegister = () => uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })

const user = getStoredUser()
const canPost = computed(() => user && (user.role === 'enterprise' || user.role === 'platform_admin' || user.role === 'association_admin'))

const form = ref({ title: '', location: '', salary: '', description: '' })
const submitting = ref(false)

const submit = async () => {
  if (!form.value.title) return uni.showToast({ title: '请输入职位名称', icon: 'none' })
  submitting.value = true
  try {
    await request({
      url: '/api/v1/jobs',
      method: 'POST',
      data: {
        title: form.value.title,
        location: form.value.location,
        salary_fen: Math.round((Number(form.value.salary) || 0) * 100),
        description: form.value.description,
      },
    })
    uni.showToast({ title: '发布成功（草稿）', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '发布失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.publish-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 40rpx; }
.form-wrap { padding: 24rpx 0; }
.field-row { position: relative; }
.unit { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-secondary); font-size: 26rpx; z-index: 2; }
.submit-wrap { padding: 32rpx 24rpx; display: flex; flex-direction: column; align-items: center; gap: 20rpx; }
.submit-note { font-size: 22rpx; color: var(--color-text-placeholder); }
.gate { padding: 120rpx 48rpx; display: flex; flex-direction: column; align-items: center; gap: 24rpx; }
.gate-ico { width: 120rpx; height: 120rpx; border-radius: 50%; background: var(--color-primary-light); color: var(--color-primary); font-size: 48rpx; font-weight: 700; display: flex; align-items: center; justify-content: center; }
.gate-title { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.gate-note { font-size: 24rpx; color: var(--color-text-secondary); text-align: center; line-height: 1.6; }
</style>
