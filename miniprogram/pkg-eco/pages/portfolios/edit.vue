<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar :title="isEdit ? '编辑品牌' : '创建品牌'" show-back :fixed="true" @back="goBack" />

    <view class="form-card">
      <view class="form-group">
        <view class="form-label">品牌名称 <text class="required">*</text></view>
        <input class="form-input" v-model="form.name" maxlength="30" placeholder="请输入品牌 / 企业名称" placeholder-class="ph" />
      </view>

      <view class="form-group">
        <view class="form-label">品牌简介</view>
        <textarea class="form-textarea" v-model="form.description" maxlength="300" placeholder="一句话介绍企业核心能力、主营方向" placeholder-class="ph" />
      </view>

      <view class="form-group">
        <view class="form-label">产品展示 <text class="hint">点击标签可删除</text></view>
        <view v-if="form.products.length" class="tag-list">
          <view v-for="(p, i) in form.products" :key="i" class="tag-chip" @tap="removeAt('products', i)">{{ p }}<text class="tag-x">✕</text></view>
        </view>
        <view class="tag-add">
          <input class="tag-input" v-model="productInput" placeholder="输入产品 / 解决方案名称" placeholder-class="ph" confirm-type="done" @confirm="addProduct" />
          <view class="tag-add-btn" hover-class="tap-fade" @tap="addProduct">添加</view>
        </view>
      </view>

      <view class="form-group">
        <view class="form-label">荣誉资质 <text class="hint">点击标签可删除</text></view>
        <view v-if="form.honors.length" class="tag-list">
          <view v-for="(h, i) in form.honors" :key="i" class="tag-chip chip-amber" @tap="removeAt('honors', i)">{{ h }}<text class="tag-x">✕</text></view>
        </view>
        <view class="tag-add">
          <input class="tag-input" v-model="honorInput" placeholder="输入认证 / 荣誉名称" placeholder-class="ph" confirm-type="done" @confirm="addHonor" />
          <view class="tag-add-btn" hover-class="tap-fade" @tap="addHonor">添加</view>
        </view>
      </view>

      <view class="form-group">
        <view class="form-label">联系方式</view>
        <input class="form-input" v-model="form.contact_info" maxlength="60" placeholder="微信号 / 联系电话（展示在品牌详情）" placeholder-class="ph" />
      </view>
    </view>

    <view class="notice">
      <text class="notice-title">提交须知</text>
      <text class="notice-line">· 提交后进入协会审核，审核通过后对外展示</text>
      <text class="notice-line">· 审核中 / 已发布均可继续修改，修改后重新进入审核</text>
    </view>

    <view class="submit-bar">
      <view class="submit-btn" :class="{ disabled: submitting }" hover-class="tap-fade" @tap="submit">
        {{ submitting ? '提交中...' : (isEdit ? '保存修改' : '提交审核') }}
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage } from '@/utils/request'

const statusBarHeight = ref(20)
const isEdit = ref(false)
const submitting = ref(false)
const id = ref('')
const productInput = ref('')
const honorInput = ref('')
const form = ref({ name: '', logo_url: '', cover_url: '', description: '', contact_info: '', products: [], honors: [], status: 'draft' })

const addAt = (listKey, input, inputKey) => {
  const val = (input || '').trim()
  if (!val) return
  if (form.value[listKey].includes(val)) {
    uni.showToast({ title: '已存在', icon: 'none' })
    return
  }
  form.value[listKey].push(val)
  form.value[inputKey] = ''
}
const addProduct = () => addAt('products', productInput.value, 'productInput')
const addHonor = () => addAt('honors', honorInput.value, 'honorInput')
const removeAt = (listKey, i) => { form.value[listKey].splice(i, 1) }

const fillForm = (it) => {
  form.value = {
    name: it.name || '',
    logo_url: it.logo_url || '',
    cover_url: it.cover_url || '',
    description: it.desc || it.description || '',
    contact_info: it.contact_info || '',
    products: (it.products || []).map(String),
    honors: (it.honors || []).map(String),
    status: it.status || 'draft',
  }
}

// 编辑模式回填：优先使用「我的品牌」写入的缓存，其次拉取 mine 接口
const fetchMineAndFill = async () => {
  try {
    const res = await request({ url: '/api/v1/portfolios/mine' })
    const list = Array.isArray(res) ? res : (res && res.data) || []
    const found = list.find((x) => x.id === id.value)
    if (found) fillForm(found)
  } catch { /* 保持表单为空，用户可重新填写 */ }
}

const submit = async () => {
  if (!form.value.name.trim()) {
    uni.showToast({ title: '请填写品牌名称', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: form.value.name.trim(),
      logo_url: form.value.logo_url || '',
      cover_url: form.value.cover_url || '',
      description: form.value.description.trim(),
      contact_info: form.value.contact_info.trim(),
      products: form.value.products.filter((x) => String(x).trim()),
      honors: form.value.honors.filter((x) => String(x).trim()),
    }
    if (isEdit.value) {
      payload.status = form.value.status || 'draft'
      await request({ url: '/api/v1/portfolios/' + encodeURIComponent(id.value), method: 'PUT', data: payload })
    } else {
      await request({ url: '/api/v1/portfolios', method: 'POST', data: payload })
    }
    uni.showToast({ title: isEdit.value ? '保存成功' : '提交成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch {
    uni.showToast({ title: '提交失败，请重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

const goBack = () => uni.navigateBack()

onLoad((options) => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(() => uni.navigateTo({ url: '/pages/login/index' }), 500)
    return
  }
  if (options && options.id) {
    id.value = decodeURIComponent(options.id)
    isEdit.value = true
  }
  const cached = uni.getStorageSync('portfolio_edit_cache')
  if (cached && cached.id) {
    fillForm(cached)
    uni.removeStorageSync('portfolio_edit_cache')
  } else if (isEdit.value) {
    fetchMineAndFill()
  }
})
</script>

<style scoped>
.page { min-height: 100vh; background: #F4F6F8; padding-bottom: 140rpx; }

/* ===== 表单 ===== */
.form-card { background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; margin: 24rpx; padding: 32rpx; }
.form-group { margin-bottom: 32rpx; }
.form-group:last-child { margin-bottom: 0; }
.form-label { display: flex; align-items: center; font-size: 26rpx; font-weight: 600; color: #344054; margin-bottom: 12rpx; }
.required { color: #D92D20; font-size: 24rpx; margin-left: 4rpx; }
.hint { font-size: 20rpx; color: #98A2B3; font-weight: 400; margin-left: 12rpx; }
.form-input { width: 100%; box-sizing: border-box; min-height: 88rpx; padding: 20rpx 24rpx; border: 1px solid #E4E7EC; border-radius: 16rpx; font-size: 28rpx; color: #17212B; background: #fff; }
.form-textarea { width: 100%; box-sizing: border-box; min-height: 160rpx; padding: 20rpx 24rpx; border: 1px solid #E4E7EC; border-radius: 16rpx; font-size: 28rpx; color: #17212B; background: #fff; }
.ph { color: #98A2B3; }

/* ===== 动态标签 ===== */
.tag-list { display: flex; flex-wrap: wrap; gap: 16rpx; margin-bottom: 16rpx; }
.tag-chip {
  display: inline-flex; align-items: center; gap: 8rpx;
  font-size: 24rpx; color: #0A66C2; background: #EAF3FB;
  padding: 10rpx 20rpx; border-radius: 12rpx;
}
.tag-chip.chip-amber { color: #E96012; background: #FFF0E6; }
.tag-x { font-size: 20rpx; opacity: .7; }
.tag-add { display: flex; gap: 16rpx; align-items: center; }
.tag-input { flex: 1; min-width: 0; height: 80rpx; padding: 0 24rpx; border: 1px solid #E4E7EC; border-radius: 16rpx; font-size: 26rpx; color: #17212B; background: #fff; }
.tag-add-btn {
  flex: none; height: 80rpx; padding: 0 36rpx; border-radius: 16rpx;
  background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
}

/* ===== 须知 ===== */
.notice { background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; margin: 0 24rpx; padding: 24rpx 28rpx; }
.notice-title { display: block; font-size: 26rpx; font-weight: 600; color: #0A66C2; margin-bottom: 12rpx; }
.notice-line { display: block; font-size: 24rpx; color: #667085; line-height: 1.7; }

/* ===== 提交栏 ===== */
.submit-bar { position: fixed; left: 0; right: 0; bottom: 0; padding: 20rpx 24rpx; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); background: #fff; border-top: 1px solid #EEF1F4; }
.submit-btn { height: 92rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 30rpx; font-weight: 600; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.35); }
.submit-btn.disabled { background: #98A2B3; box-shadow: none; }
</style>
