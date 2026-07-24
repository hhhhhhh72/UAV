<template>
  <view class="profile-page">
    <view class="section">
      <view class="list-item avatar-item" @tap="chooseAvatar">
        <text class="label">头像</text>
        <view class="right">
          <image v-if="form.avatar" :src="form.avatar" mode="aspectFill" class="avatar" />
          <view v-else class="default-avatar">👤</view>
          <text class="arrow">›</text>
        </view>
      </view>
      <view class="list-item">
        <text class="label">昵称</text>
        <input class="input" v-model="form.name" placeholder="请输入昵称" />
      </view>
      <view class="list-item">
        <text class="label">手机号</text>
        <input class="input" v-model="form.phone" type="number" placeholder="请输入手机号" />
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

const form = ref({
  name: '',
  phone: '',
  avatar: '',
  isAuth: false
})

onMounted(() => {
  const user = JSON.parse(uni.getStorageSync('user') || '{}')
  form.value = {
    name: user.name || '',
    phone: user.phone || '',
    avatar: user.avatar || '',
    isAuth: !!user.isAuth
  }
})

const chooseAvatar = () => {
  uni.chooseImage({
    count: 1,
    success: (res) => {
      form.value.avatar = res.tempFilePaths[0]
    }
  })
}

const goAuth = () => {
  uni.navigateTo({ url: '/pages/mine/auth' })
}

const handleSave = () => {
  if (!form.value.name) return uni.showToast({ title: '请输入昵称', icon: 'none' })
  
  const user = JSON.parse(uni.getStorageSync('user') || '{}')
  const updatedUser = { ...user, ...form.value }
  uni.setStorageSync('user', JSON.stringify(updatedUser))
  
  uni.showToast({ title: '保存成功' })
  setTimeout(() => uni.navigateBack(), 1500)
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
.avatar { width: 44px; height: 44px; border-radius: 22px; }
.default-avatar { width: 44px; height: 44px; border-radius: 22px; background: #f0f2f5; display: flex; align-items: center; justify-content: center; font-size: 24px; }
.arrow { font-size: 18px; color: #ccc; }
.status-text { font-size: 14px; color: #969799; }
.status-text.verified { color: #07c160; }

.save-btn-wrap { padding: 30px 16px; }
.save-btn { border-radius: 99px; font-weight: bold; background-color: #2f7ef7 !important; }
</style>

