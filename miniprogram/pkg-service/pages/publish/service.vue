<template>
  <view class="publish-page">
    <u-nav-bar title="发布服务能力" show-back @back="goBack" />

    <u-cell-group inset>
      <view class="form-wrap">
        <u-field
          v-model="form.title"
          label="标题"
          placeholder="请输入服务名称"
        />

        <view class="field-row" @tap="showCatPicker = true">
          <u-field
            v-model="catText"
            label="服务分类"
            placeholder="请选择分类"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <u-field
          v-model="form.range"
          label="服务范围"
          placeholder="例如 重庆市内及周边"
        />

        <view class="field-row" @tap="showQuotePicker = true">
          <u-field
            v-model="quoteText"
            label="报价方式"
            placeholder="请选择报价方式"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <u-field
          v-model="form.qualification"
          label="设备与资质"
          placeholder="例如 M350 RTK、热成像载荷、保险及行业资质"
          type="textarea"
        />

        <view class="field-row">
          <u-field
            v-model="form.contact"
            label="联系人"
            placeholder="姓名或企业对接人"
          />
          <text class="required">*</text>
        </view>
      </view>
    </u-cell-group>

    <view class="submit-wrap">
      <u-button type="primary" size="large" round :loading="submitting" @tap="submit">
        发布服务能力
      </u-button>
    </view>

    <u-picker
      :show="showCatPicker"
      title="请选择服务分类"
      :columns="catNames"
      @confirm="onCatConfirm"
      @update:show="showCatPicker = $event"
    />
    <u-picker
      :show="showQuotePicker"
      title="请选择报价方式"
      :columns="quoteOptions"
      @confirm="onQuoteConfirm"
      @update:show="showQuotePicker = $event"
    />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { authStorage, getStoredUser } from '@/utils/request'
import { HALL_CATEGORIES, getPosts, savePosts } from '@/utils/hallData'

const goBack = () => uni.navigateBack()

// 服务分类与大厅服务 Tab 一致（去掉“全部”）
const CATS = (HALL_CATEGORIES.service || []).filter((c) => c !== '全部')
const catNames = CATS

const quoteOptions = ['按项目报价', '按天报价', '面议']

const form = ref({ title: '', category: '', range: '', quote: '', qualification: '', contact: '' })
const showCatPicker = ref(false)
const showQuotePicker = ref(false)
const submitting = ref(false)

const catText = computed(() => form.value.category)
const quoteText = computed(() => form.value.quote)

const onCatConfirm = (v) => { form.value.category = v }
const onQuoteConfirm = (v) => { form.value.quote = v }

const submit = async () => {
  if (!form.value.title.trim()) return uni.showToast({ title: '请输入服务名称', icon: 'none' })
  if (!form.value.category) return uni.showToast({ title: '请选择服务分类', icon: 'none' })
  if (!form.value.contact.trim()) return uni.showToast({ title: '请填写联系人', icon: 'none' })
  submitting.value = true
  try {
    // 后端暂未提供服务发布接口（/api/v1/services 仅有 GET），本期与需求一致走本地“我的发布”
    const posts = getPosts()
    posts.unshift({
      id: 'm' + Date.now(),
      type: '服务',
      title: form.value.title,
      status: '待审核',
      date: '刚刚',
    })
    savePosts(posts)
    uni.showToast({ title: '已提交审核，请等待管理员审核', icon: 'none' })
    setTimeout(goBack, 800)
  } catch (e) {
    uni.showToast({ title: '发布失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.publish-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 40rpx; }
.form-wrap { padding: 24rpx 0; }
.field-row { position: relative; }
.field-arrow { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-placeholder); font-size: 30rpx; z-index: 2; }
.required { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-danger); font-size: 28rpx; z-index: 2; }
.submit-wrap { padding: 32rpx 24rpx; }
</style>
