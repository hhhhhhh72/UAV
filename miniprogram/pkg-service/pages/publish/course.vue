<template>
  <view class="publish-page">
    <u-nav-bar title="发布培训课程" show-back @back="goBack" />

    <u-cell-group inset>
      <view class="form-wrap">
        <u-field
          v-model="form.title"
          label="课程标题"
          placeholder="例如 CAAC 民航局多旋翼驾驶员执照班"
        />

        <view class="field-row" @tap="showTypePicker = true">
          <u-field
            v-model="typeText"
            label="课程类型"
            placeholder="请选择证书类型"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <view class="field-row" @tap="showDistrictPicker = true">
          <u-field
            v-model="districtText"
            label="所属区县"
            placeholder="请选择区县"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <u-field v-model="form.location" label="培训地点" placeholder="例如 渝北区金开大道 68 号" />
        <u-field v-model="form.duration_days" label="培训天数" placeholder="例如 25 天" type="digit" />

        <view class="field-row">
          <u-field
            v-model="form.price"
            label="课程价格"
            placeholder="请输入价格"
            type="digit"
          />
          <text class="unit">元</text>
        </view>

        <u-field
          v-model="form.description"
          label="课程介绍"
          placeholder="请介绍课程内容、颁发证书与招生对象"
          type="textarea"
        />
      </view>
    </u-cell-group>

    <view class="submit-wrap">
      <u-button type="primary" size="large" round :loading="submitting" @tap="submit">
        发布课程
      </u-button>
    </view>

    <u-picker
      :show="showTypePicker"
      title="请选择课程类型"
      :columns="typeNames"
      @confirm="onTypeConfirm"
      @update:show="showTypePicker = $event"
    />
    <u-picker
      :show="showDistrictPicker"
      title="请选择重庆区县"
      :columns="districtOptions"
      @confirm="onDistrictConfirm"
      @update:show="showDistrictPicker = $event"
    />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { request } from '@/utils/request'

const goBack = () => uni.navigateBack()

// 证书类型与后端 CertType 契约一致（internal/domain/models.go）
const TYPES = [
  { value: 'caac', label: 'CAAC 民航局执照' },
  { value: 'aopa', label: 'AOPA 执照' },
  { value: 'utc_dji', label: '大疆 UTC 证书' },
  { value: 'gov_level', label: '职业技能等级' },
]
const typeNames = TYPES.map((t) => t.label)

const districtOptions = [
  '渝中区', '江北区', '南岸区', '渝北区', '沙坪坝区', '九龙坡区', '大渡口区', '北碚区', '巴南区',
  '两江新区', '高新区', '万州区', '涪陵区', '黔江区', '长寿区', '江津区', '合川区', '永川区',
  '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区',
  '武隆区', '城口县', '丰都县', '垫江县', '忠县', '云阳县', '奉节县', '巫山县', '巫溪县',
  '石柱县', '秀山县', '酉阳县', '彭水县',
]

const form = ref({ title: '', cert_type: '', district: '', location: '', duration_days: '', price: '', description: '' })
const showTypePicker = ref(false)
const showDistrictPicker = ref(false)
const submitting = ref(false)

const typeText = computed(() => (form.value.cert_type ? TYPES.find((t) => t.value === form.value.cert_type)?.label : ''))
const districtText = computed(() => form.value.district)

const onTypeConfirm = (v) => {
  const hit = TYPES.find((t) => t.label === v)
  if (hit) form.value.cert_type = hit.value
}
const onDistrictConfirm = (v) => { form.value.district = v }

const submit = async () => {
  if (!form.value.title.trim()) return uni.showToast({ title: '请输入课程标题', icon: 'none' })
  if (!form.value.cert_type) return uni.showToast({ title: '请选择课程类型', icon: 'none' })
  if (!form.value.district) return uni.showToast({ title: '请选择所属区县', icon: 'none' })
  if (!form.value.price) return uni.showToast({ title: '请输入课程价格', icon: 'none' })
  submitting.value = true
  try {
    await request({
      url: '/api/v1/training-courses',
      method: 'POST',
      data: {
        title: form.value.title,
        cert_type: form.value.cert_type,
        district: form.value.district,
        location: form.value.location,
        duration_days: form.value.duration_days ? Number(form.value.duration_days) : 0,
        price_fen: Math.round(Number(form.value.price) * 100),
        description: form.value.description,
      },
    })
    uni.showToast({ title: '发布成功', icon: 'success' })
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
.field-arrow { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-placeholder); font-size: 30rpx; z-index: 2; }
.unit { position: absolute; right: 28rpx; top: 50%; transform: translateY(-50%); color: var(--color-text-secondary); font-size: 26rpx; z-index: 2; }
.submit-wrap { padding: 32rpx 24rpx; }
</style>
