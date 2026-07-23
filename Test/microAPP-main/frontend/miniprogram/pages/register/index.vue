<template>
  <view class="register-page">
    <view class="logo-wrap">
      <view class="register-avatar">
        <text class="avatar-text">+</text>
      </view>
      <view class="title">创建账号</view>
    </view>

    <view class="form-box">
      <view class="input-item">
        <text class="label">手机号</text>
        <input class="input" v-model="form.phone" type="number" placeholder="请输入手机号" />
      </view>
      <view class="input-item">
        <text class="label">姓名</text>
        <input class="input" v-model="form.name" placeholder="请输入姓名（选填）" />
      </view>
      <view class="input-item">
        <text class="label">密码</text>
        <input class="input" v-model="form.password" password placeholder="请设置登录密码" />
      </view>
      <view class="input-item">
        <text class="label">确认</text>
        <input class="input" v-model="form.confirmPassword" password placeholder="请再次输入密码" />
      </view>

      <button class="reg-btn" type="primary" @tap="handleRegister" :loading="loading">注册</button>
      
      <view class="login-link" @tap="goLogin">
        已有账号？<text class="blue">立即登录</text>
      </view>

      <view class="sso-tip">
        <text>从畅行温州平台跳转将自动注册登录</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request, authStorage } from '../../utils/request'

const loading = ref(false)

const form = ref({
  phone: '',
  name: '',
  password: '',
  confirmPassword: ''
})

const handleRegister = async () => {
  if (!form.value.phone) {
    return uni.showToast({ title: '请填写手机号', icon: 'none' })
  }
  if (!/^1[3-9]\d{9}$/.test(form.value.phone)) {
    return uni.showToast({ title: '手机号格式不正确', icon: 'none' })
  }
  if (!form.value.password || form.value.password.length < 6) {
    return uni.showToast({ title: '密码至少6位', icon: 'none' })
  }
  if (form.value.password !== form.value.confirmPassword) {
    return uni.showToast({ title: '两次密码不一致', icon: 'none' })
  }

  loading.value = true
  try {
    const res = await request({
      url: '/api/auth/register',
      method: 'POST',
      data: {
        phone: form.value.phone,
        name: form.value.name || `User${form.value.phone.slice(-4)}`,
        password: form.value.password
      }
    })

    if (!res?.success) {
      throw new Error(res?.message || '注册失败')
    }

    uni.setStorageSync('user', JSON.stringify(res.user))
    authStorage.setTokens(res.accessToken, res.refreshToken)

    uni.showToast({ title: '注册成功' })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/home/index' })
    }, 1000)
  } catch (error) {
    console.error(error)
    const msg = error?.data?.message || error?.message || '注册失败'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    loading.value = false
  }
}

const goLogin = () => {
  uni.navigateBack()
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  background: #fff;
  padding: 80px 30px 30px;
}

.logo-wrap {
  text-align: center;
  margin-bottom: 40px;
}

.register-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #f0f2ff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.avatar-text {
  font-size: 36px;
  color: #667eea;
}

.title {
  font-size: 22px;
  font-weight: bold;
  color: #323233;
}

.form-box {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.input-item {
  display: flex;
  align-items: center;
  background: #f7f8fa;
  padding: 14px 16px;
  border-radius: 12px;
}

.label {
  width: 50px;
  font-size: 15px;
  color: #323233;
  flex-shrink: 0;
}

.input {
  flex: 1;
  font-size: 15px;
}

.reg-btn {
  margin-top: 12px;
  height: 48px;
  line-height: 48px;
  border-radius: 24px;
  background-color: #667eea !important;
  font-size: 17px;
  font-weight: bold;
}

.login-link {
  text-align: center;
  font-size: 14px;
  color: #969799;
}

.blue {
  color: #667eea;
}

.sso-tip {
  text-align: center;
  font-size: 13px;
  color: #c8c9cc;
  margin-top: 20px;
}
</style>
