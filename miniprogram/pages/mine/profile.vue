<template>
  <view class="profile-page">
    <view class="section">
      <view class="list-item avatar-item" @tap="chooseAvatar">
        <text class="label">头像</text>
        <view class="right">
          <image v-if="form.avatar" :src="form.avatar" mode="aspectFill" class="avatar" />
          <view v-else class="default-avatar">我</view>
          <text class="arrow">›</text>
        </view>
      </view>
      <view class="list-item">
        <text class="label">昵称</text>
        <input class="input" v-model="form.name" placeholder="请输入昵称" />
      </view>
    </view>

    <view class="section">
      <view class="list-item" @tap="goAuth">
        <text class="label">实名状态</text>
        <view class="right">
          <text class="status-text" :class="{ verified: form.isAuth }">{{ form.isAuth ? '已认证' : '未认证' }}</text>
          <text class="arrow">›</text>
        </view>
      </view>
    </view>

    <view class="save-btn-wrap">
      <button class="save-btn" type="primary" @tap="handleSave">保存修改</button>
    </view>

  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { request, BASE_URL, authStorage } from '../../utils/request'

const form = ref({
  name: '',
  avatar: '',
  isAuth: false
})

onMounted(() => {
  const user = JSON.parse(uni.getStorageSync('user') || '{}')
  form.value = {
    name: user.name || '',
    avatar: user.avatar || user.avatar_url || '',
    isAuth: !!user.isAuth
  }
})

// 选择头像后立即上传 /api/v1/upload，成功后保存服务器 URL
const chooseAvatar = () => {
  uni.chooseImage({
    count: 1,
    success: (res) => {
      const tempPath = res.tempFilePaths[0]
      uni.showLoading({ title: '上传中...', mask: true })
      const token = authStorage.getAccessToken()
      uni.uploadFile({
        url: BASE_URL + '/api/v1/upload',
        filePath: tempPath,
        name: 'file',
        header: token ? { Authorization: `Bearer ${token}` } : {},
        success: (upRes) => {
          uni.hideLoading()
          if (upRes.statusCode >= 200 && upRes.statusCode < 300) {
            const body = JSON.parse(upRes.data || '{}')
            const url = body.data?.url || body.url
            if (url) {
              form.value.avatar = url
              uni.showToast({ title: '头像已上传' })
            } else {
              uni.showToast({ title: '上传失败', icon: 'none' })
            }
          } else {
            uni.showToast({ title: '上传失败', icon: 'none' })
          }
        },
        fail: () => {
          uni.hideLoading()
          uni.showToast({ title: '上传失败', icon: 'none' })
        }
      })
    }
  })
}

const goAuth = () => {
  uni.navigateTo({ url: '/pages/mine/auth' })
}

const handleSave = async () => {
  if (!form.value.name) return uni.showToast({ title: '请输入昵称', icon: 'none' })

  uni.showLoading({ title: '保存中...', mask: true })
  try {
    // 昵称/头像保存到服务端（users.name / users.avatar_url）
    await request({
      url: '/api/v1/me',
      method: 'PATCH',
      data: { name: form.value.name, avatar_url: form.value.avatar }
    })
    // 同步本地缓存
    const user = JSON.parse(uni.getStorageSync('user') || '{}')
    const updatedUser = { ...user, name: form.value.name, avatar: form.value.avatar }
    uni.setStorageSync('user', JSON.stringify(updatedUser))
    uni.hideLoading()
    uni.showToast({ title: '保存成功' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '保存失败，请重试', icon: 'none' })
  }
}
</script>

<style scoped>
.profile-page { min-height: 100vh; background: #f7f8fa; padding: 12px 0; }
.section { background: #fff; margin-bottom: 12px; }
.list-item { display: flex; align-items: center; justify-content: space-between; padding: 16px; border-bottom: 1px solid #f2f3f5; }
.list-item:last-child { border-bottom: none; }
.label { font-size: 15px; color: #323233; width: 80px; }
.input { flex: 1; text-align: right; font-size: 15px; color: #646566; }
.right { display: flex; align-items: center; gap: 8px; }
.avatar { width: 44px; height: 44px; border-radius: 8px; }
.default-avatar { width: 44px; height: 44px; border-radius: 8px; background: #f0f2f5; display: flex; align-items: center; justify-content: center; font-size: 24px; }
.arrow { font-size: 18px; color: #ccc; }
.status-text { font-size: 14px; color: #969799; }
.status-text.verified { color: #07c160; }

.save-btn-wrap { padding: 30px 16px; }
.save-btn { border-radius: 8px; font-weight: bold; background-color: #2f7ef7 !important; }
</style>

