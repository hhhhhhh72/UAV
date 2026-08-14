<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">发布招聘</view>
    </view>

    <!-- 非企业账号：引导入驻 -->
    <view v-if="!canPost" class="pub-empty">
      <view class="pub-empty-mark">聘</view>
      <view class="pub-empty-title">仅企业账号可发布招聘</view>
      <view class="pub-empty-desc">发布招聘需要企业身份，请先完成企业入驻后重试。</view>
      <view class="pub-btn pub-btn--primary gate-btn" hover-class="pub-btn--active" @tap="goRegister">去企业入驻</view>
    </view>

    <template v-else>
      <view class="pub-form-intro">
        <view class="pub-form-intro-h2">发布招聘</view>
        <view class="pub-form-intro-p">发布后为草稿状态，可联系协会或在我的招聘中发布上线。</view>
      </view>

      <view class="pub-section">
        <view class="pub-section-title">职位信息</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">职位名称<text class="pub-required">*</text></view>
            <input
              v-model="form.title"
              class="pub-input"
              placeholder="如：无人机飞手"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">工作地点</view>
            <input
              v-model="form.location"
              class="pub-input"
              placeholder="如：重庆·渝北"
              placeholder-class="pub-placeholder"
            />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">薪资</view>
            <input
              v-model="form.salary"
              class="pub-input"
              type="digit"
              placeholder="月薪（元）"
              placeholder-class="pub-placeholder"
            />
            <text class="pub-field-hint">元/月</text>
          </view>
          <view class="pub-field">
            <view class="pub-field-label">职位描述</view>
            <textarea
              v-model="form.description"
              class="pub-input pub-input--textarea"
              placeholder="岗位职责 / 任职要求"
              placeholder-class="pub-placeholder"
            />
          </view>
        </view>
      </view>

      <!-- 固定底部操作区（与发布页同款） -->
      <view class="pub-sticky">
        <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="submit">
          {{ submitting ? '发布中...' : '发布招聘' }}
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { request, getStoredUser } from '@/utils/request'
import { useSafeTop } from '../../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

const goBack = () => uni.navigateBack()
const goRegister = () => uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })

const user = getStoredUser()
const canPost = computed(() => user && (user.role === 'enterprise' || user.role === 'platform_admin' || user.role === 'association_admin'))

const form = ref({ title: '', location: '', salary: '', description: '' })
const submitting = ref(false)
let backTimer = null

const submit = async () => {
  if (submitting.value) return
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
    backTimer = setTimeout(() => uni.navigateBack(), 800)
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '发布失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

onLoad(() => {
  initSafeTop()
})

onUnload(() => {
  if (backTimer) clearTimeout(backTimer)
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

/* 企业闸门按钮：居中自适应宽度（仿 resume.vue 的 retry-btn 模式） */
.gate-btn {
  flex: none;
  margin: 14px auto 0;
  padding: 0 22px;
}
</style>
