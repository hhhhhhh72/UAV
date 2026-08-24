<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 导航：返回 + 标题 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">发布赛事</view>
      <view class="pub-nav-action" :style="{ marginRight: capsuleGap + 'px' }"></view>
    </view>

    <view class="pub-form-intro">
      <view class="pub-form-intro-h2">发布赛事</view>
      <view class="pub-form-intro-p">赛事提交后由协会审核，通过后公开展示在赛事列表</view>
    </view>

    <!-- 基本信息 -->
    <view class="pub-section">
      <view class="pub-section-title">赛事信息</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">赛事名称<text class="pub-required">*</text></view>
          <input class="pub-input" v-model="form.title" placeholder="例如：2026 全国无人机职业技能大赛" placeholder-class="pub-placeholder" />
        </view>
        <view class="pub-field">
          <view class="pub-field-label">赛事类别</view>
          <picker mode="selector" :range="categoryOptions" @change="onCategoryChange">
            <view class="pub-select-field" :class="{ placeholder: !form.category }">
              <text>{{ form.category || '选择赛事类别' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </picker>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">举办地点</view>
          <input class="pub-input" v-model="form.location" placeholder="例如：重庆国际博览中心" placeholder-class="pub-placeholder" />
        </view>
        <view class="pub-field">
          <view class="pub-field-label">主办单位</view>
          <input class="pub-input" v-model="form.sponsor" placeholder="填写主办方全称" placeholder-class="pub-placeholder" />
        </view>
        <view class="pub-field">
          <view class="pub-field-label">开始日期</view>
          <picker mode="date" :value="form.start_date" @change="(e) => form.start_date = e.detail.value">
            <view class="pub-select-field" :class="{ placeholder: !form.start_date }">
              <text>{{ form.start_date || '选择开始日期' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </picker>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">结束日期</view>
          <picker mode="date" :value="form.end_date" @change="(e) => form.end_date = e.detail.value">
            <view class="pub-select-field" :class="{ placeholder: !form.end_date }">
              <text>{{ form.end_date || '选择结束日期' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </picker>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">报名截止</view>
          <picker mode="date" :value="form.deadline" @change="(e) => form.deadline = e.detail.value">
            <view class="pub-select-field" :class="{ placeholder: !form.deadline }">
              <text>{{ form.deadline || '选择报名截止日期' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </picker>
        </view>
      </view>
    </view>

    <!-- 报名信息 -->
    <view class="pub-section">
      <view class="pub-section-title">报名信息</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">报名费(元)</view>
          <input class="pub-input" type="digit" v-model="form.feeYuan" placeholder="0 表示免费，单位：元" placeholder-class="pub-placeholder" />
          <text class="pub-field-hint">收费赛事请同时填写报名截止日期</text>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">队伍名额</view>
          <input class="pub-input" type="number" v-model="form.maxTeams" placeholder="例如：50" placeholder-class="pub-placeholder" />
        </view>
        <view class="pub-field">
          <view class="pub-field-label">赛事简介</view>
          <textarea class="pub-input pub-input--textarea" v-model="form.description" placeholder="赛事规则、参赛要求、奖项设置等" placeholder-class="pub-placeholder" :maxlength="500" />
        </view>
      </view>
    </view>

    <!-- 海报图 -->
    <view class="pub-section">
      <view class="pub-section-title">海报图</view>
      <view class="pub-form-card">
        <view class="pub-upload-row">
          <view v-if="poster" class="pub-photo">
            <image :src="poster" mode="aspectFill" class="pub-photo-img" />
          </view>
          <view class="pub-add-photo" hover-class="pub-fade" @tap="choosePoster">＋</view>
        </view>
        <view class="pub-upload-tip">建议上传赛事海报，首图将作为赛事列表封面</view>
      </view>
    </view>

    <!-- 提交 -->
    <view class="pub-sticky">
      <view class="pub-btn pub-btn--ghost" hover-class="pub-btn--active" @tap="goBack">取消</view>
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="submit">提交审核</view>
    </view>

    <view v-if="toast" class="pub-toast">{{ toast }}</view>
  </view>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useSafeTop } from '../../../utils/safeTop'
import { request, authStorage, BASE_URL, getStoredUser } from '../../../utils/request'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const categoryOptions = ['竞技赛事', '职业技能', '行业应用', '科普教育', '展演交流', '其他']
const form = reactive({
  title: '', category: '', location: '', sponsor: '',
  start_date: '', end_date: '', deadline: '',
  feeYuan: '', maxTeams: '', description: '',
})
const poster = ref('')
const submitting = ref(false)
const toast = ref('')

function showToast(msg) {
  toast.value = msg
  setTimeout(() => { toast.value = '' }, 2000)
}

const onCategoryChange = (e) => {
  form.category = categoryOptions[Number(e.detail.value)] || ''
}

const choosePoster = () => {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: (res) => {
      poster.value = res.tempFilePaths && res.tempFilePaths[0]
    },
  })
}

// 上传海报到服务器，返回 /uploads/{file_id}
async function uploadPoster() {
  const token = authStorage.getAccessToken()
  if (!poster.value) return ''
  const data = await new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + '/api/v1/files/upload',
      filePath: poster.value,
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
  if (!fid) throw new Error('upload response missing file_id')
  return '/uploads/' + fid
}

async function submit() {
  if (!form.title.trim()) { showToast('请输入赛事名称'); return }
  if (submitting.value) return
  submitting.value = true
  try {
    const posterUrl = poster.value ? await uploadPoster() : ''
    const created = await request({
      url: '/api/v1/competitions',
      method: 'POST',
      data: {
        title: form.title.trim(),
        category: form.category,
        location: form.location.trim(),
        sponsor: form.sponsor.trim(),
        start_date: form.start_date,
        end_date: form.end_date,
        deadline: form.deadline,
        // 后端 Competition.Fee 单位为元（非分），直接传元
        fee: Math.round(Number(form.feeYuan) || 0),
        max_teams: Number(form.maxTeams) || 0,
        description: form.description.trim(),
        poster: posterUrl,
      },
    })
    if (!created || !created.id) throw new Error('create competition failed')
    uni.showToast({ title: '已提交审核', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (e) {
    showToast('发布失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

function goBack() {
  uni.navigateBack({ fail: () => uni.redirectTo({ url: '/pkg-eco/pages/competitions/list' }) })
}

onLoad(() => {
  initSafeTop()
  // 登录/角色守卫：未登录重定向登录页；非企业账号提示并返回
  if (!authStorage.getAccessToken()) {
    uni.redirectTo({ url: '/pages/login/index' })
    return
  }
  const u = getStoredUser()
  if (!(u && (u.role === 'enterprise' || u.user_type === 'enterprise'))) {
    uni.showToast({ title: '仅企业账号可发布赛事', icon: 'none' })
    setTimeout(() => goBack(), 800)
    return
  }
})
</script>

<style scoped>
/* 与发布中心二级表单（pages/publish/form.vue）同一套视觉体系 */
.pub-page { min-height: 100vh; background: #F4F6F8; padding-bottom: calc(170rpx + env(safe-area-inset-bottom)); box-sizing: border-box; }
.pub-nav { position: relative; display: flex; align-items: center; height: 88rpx; padding: 0 24rpx; }
.pub-back { width: 64rpx; height: 64rpx; border-radius: 50%; background: #F5F8FC; display: flex; align-items: center; justify-content: center; font-size: 40rpx; color: #17212B; }
.pub-nav-title { position: absolute; left: 0; right: 0; text-align: center; font-size: 32rpx; font-weight: 700; color: #17212B; }
.pub-nav-action { position: absolute; right: 24rpx; font-size: 26rpx; color: #0A66C2; }
.pub-form-intro { padding: 8rpx 32rpx 24rpx; }
.pub-form-intro-h2 { font-size: 40rpx; font-weight: 800; color: #17212B; }
.pub-form-intro-p { margin-top: 8rpx; font-size: 24rpx; color: #667085; }
.pub-section { padding: 0 24rpx 24rpx; }
.pub-section-title { font-size: 28rpx; font-weight: 700; color: #344054; margin-bottom: 16rpx; }
.pub-section-note { font-size: 22rpx; color: #98A2B3; margin-bottom: 12rpx; }
.pub-form-card { background: #fff; border-radius: 24rpx; padding: 8rpx 24rpx; }
.pub-field { padding: 20rpx 0; border-bottom: 1rpx solid #F0F2F5; }
.pub-field:last-child { border-bottom: none; }
.pub-field-label { font-size: 26rpx; color: #344054; margin-bottom: 12rpx; }
.pub-required { color: #E5484D; margin-left: 4rpx; }
.pub-input { width: 100%; font-size: 28rpx; color: #17212B; }
.pub-input--textarea { min-height: 140rpx; line-height: 1.6; }
.pub-placeholder { color: #B0B8C4; }
.pub-select-field { display: flex; align-items: center; justify-content: space-between; font-size: 28rpx; color: #17212B; }
.pub-select-field.placeholder { color: #B0B8C4; }
.pub-arrow { color: #C0C6CF; font-size: 30rpx; }
.pub-field-hint { display: block; margin-top: 8rpx; font-size: 22rpx; color: #98A2B3; }
.pub-upload-row { display: flex; gap: 16rpx; padding: 12rpx 0; flex-wrap: wrap; }
.pub-photo { width: 160rpx; height: 160rpx; border-radius: 16rpx; overflow: hidden; }
.pub-photo-img { width: 100%; height: 100%; }
.pub-add-photo { width: 160rpx; height: 160rpx; border: 2rpx dashed #C8CDD6; border-radius: 16rpx; display: flex; align-items: center; justify-content: center; font-size: 56rpx; color: #98A2B3; background: #FAFBFC; }
.pub-upload-tip { font-size: 22rpx; color: #98A2B3; padding-bottom: 16rpx; }
.pub-sticky { position: fixed; left: 0; right: 0; bottom: 0; display: flex; gap: 16rpx; padding: 16rpx 24rpx calc(env(safe-area-inset-bottom) + 16rpx); background: #fff; box-shadow: 0 -4rpx 24rpx rgba(15, 23, 42, 0.06); }
.pub-btn { flex: 1; height: 88rpx; border-radius: 50rpx; display: flex; align-items: center; justify-content: center; font-size: 30rpx; font-weight: 700; }
.pub-btn--ghost { background: #F1F3F7; color: #344054; }
.pub-btn--primary { background: #0A66C2; color: #fff; box-shadow: 0 8rpx 24rpx rgba(10, 102, 194, 0.32); }
.pub-btn--active { opacity: 0.85; transform: scale(0.98); }
.pub-fade { opacity: 0.85; }
.pub-toast { position: fixed; left: 50%; bottom: 180rpx; transform: translateX(-50%); background: rgba(23, 33, 43, 0.85); color: #fff; font-size: 26rpx; padding: 16rpx 32rpx; border-radius: 999rpx; z-index: 99; }
</style>
