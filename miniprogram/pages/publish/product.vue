<template>
  <view class="publish-page">
    <u-nav-bar title="发布商品" show-back @back="goBack" />

    <u-cell-group inset>
      <view class="form-wrap">
        <u-field
          v-model="form.title"
          label="标题"
          placeholder="请输入商品标题"
        />

        <view class="field-row" @tap="showTypePicker = true">
          <u-field
            v-model="typeText"
            label="供给类型"
            placeholder="请选择类型"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <u-field v-model="form.brand" label="品牌" placeholder="如：渝航智能" />
        <u-field v-model="form.model" label="型号" placeholder="如：X6-28L" />

        <view class="field-row" @tap="showCondPicker = true">
          <u-field
            v-model="condText"
            label="成色"
            placeholder="请选择成色"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <view class="field-row">
          <u-field
            v-model="form.price"
            label="价格"
            placeholder="请输入价格"
            type="digit"
          />
          <text class="unit">元</text>
        </view>

        <u-field
          v-model="form.description"
          label="描述"
          placeholder="请描述商品详情"
          type="textarea"
        />
      </view>
    </u-cell-group>

    <view class="submit-wrap">
      <u-button type="primary" size="large" round :loading="submitting" @tap="submit">
        发布商品
      </u-button>
    </view>

    <u-picker
      :show="showTypePicker"
      title="请选择供给类型"
      :columns="typeNames"
      @confirm="onTypeConfirm"
      @update:show="showTypePicker = $event"
    />
    <u-picker
      :show="showCondPicker"
      title="请选择成色"
      :columns="['全新', '二手']"
      @confirm="onCondConfirm"
      @update:show="showCondPicker = $event"
    />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { request } from '@/utils/request'
import { productTypeLabel } from '@/utils/enums'

const goBack = () => uni.navigateBack()

// 供给 6 类（功能方案修订版：整机/零部件/航拍/试飞测试/检测标定/空域协调；维修为商城补充类）
const TYPES = [
  { value: 'drone', label: '整机' },
  { value: 'part', label: '零件' },
  { value: 'repair', label: '维修服务' },
  { value: 'aerial', label: '航拍服务' },
  { value: 'test_fly', label: '试飞测试' },
  { value: 'calibration', label: '检测标定' },
  { value: 'airspace', label: '空域协调' },
]
const typeNames = TYPES.map((t) => t.label)

const form = ref({ title: '', type: '', brand: '', model: '', condition: '', price: '', description: '' })
const showTypePicker = ref(false)
const showCondPicker = ref(false)
const submitting = ref(false)

const typeText = computed(() => (form.value.type ? productTypeLabel(form.value.type) : ''))
const condText = computed(() => (form.value.condition === 'used' ? '二手' : form.value.condition === 'new' ? '全新' : ''))

// u-picker confirm 回调直接回传选中的字符串（单列）
const onTypeConfirm = (v) => {
  const hit = TYPES.find((t) => t.label === v)
  if (hit) form.value.type = hit.value
}
const onCondConfirm = (v) => {
  form.value.condition = v === '全新' ? 'new' : 'used'
}

const submit = async () => {
  if (!form.value.title) return uni.showToast({ title: '请输入商品标题', icon: 'none' })
  if (!form.value.type) return uni.showToast({ title: '请选择供给类型', icon: 'none' })
  if (!form.value.price) return uni.showToast({ title: '请输入价格', icon: 'none' })
  submitting.value = true
  try {
    const p = await request({
      url: '/api/v1/products',
      method: 'POST',
      data: {
        title: form.value.title,
        prod_type: form.value.type,
        brand: form.value.brand,
        model: form.value.model,
        condition: form.value.condition || 'new',
        price_fen: Math.round(Number(form.value.price) * 100),
        description: form.value.description,
      },
    })
    uni.showToast({ title: '发布成功', icon: 'success' })
    setTimeout(() => {
      if (p && p.id) uni.redirectTo({ url: '/pages/mall/detail?id=' + encodeURIComponent(p.id) })
      else uni.navigateBack()
    }, 800)
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
.field-arrow { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-placeholder); font-size: 30rpx; z-index: 2; }
.unit { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-secondary); font-size: 26rpx; z-index: 2; }
.submit-wrap { padding: 32rpx 24rpx; }
</style>
