<template>
  <view class="login-page">
    <!-- 手机号授权弹窗 -->
    <!-- #ifdef MP-WEIXIN -->
    <view class="phone-modal" v-if="showPhoneAuth" @tap.stop>
      <view class="modal-mask" @tap="skipPhoneAuth"></view>
      <view class="modal-content">
        <view class="modal-title">绑定手机号</view>
        <view class="modal-desc">绑定手机号后可使用更多功能</view>
        <button class="phone-auth-btn" open-type="getPhoneNumber" @getphonenumber="handleGetPhone">
          授权微信手机号
        </button>
        <view class="skip-btn" @tap="skipPhoneAuth">
          <text>暂时跳过</text>
        </view>
      </view>
    </view>
    <!-- #endif -->

    <view class="logo-wrap">
      <view class="login-avatar">
        <text class="avatar-text">👤</text>
      </view>
      <view class="title">低空综合服务平台</view>
    </view>

    <view class="form-box">
      <view class="input-item">
        <text class="label">账号</text>
        <input class="input" v-model="phone" placeholder="请输入手机号/用户名" />
      </view>
      <view class="input-item">
        <text class="label">密码</text>
        <input class="input" v-model="password" password placeholder="请输入密码" />
      </view>

      <button class="login-btn" type="primary" @tap="handleLogin" :loading="loading">登录</button>
      
      <view class="action-links">
        <text class="link-text" @tap="goRegister">还没有账号？立即注册</text>
      </view>

      <view class="divider">
        <view class="line"></view>
        <text class="text">其他方式</text>
        <view class="line"></view>
      </view>

      <!-- #ifdef MP-WEIXIN -->
      <button class="wechat-btn" @tap="handleWechatLogin" :loading="wxLoading">
        微信一键登录
      </button>
      <!-- #endif -->
      <!-- #ifndef MP-WEIXIN -->
      <view class="sso-tip">
        <text>从畅行温州平台进入将自动登录</text>
      </view>
      <!-- #endif -->
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request, authStorage } from '../../utils/request'

const phone = ref('')
const password = ref('')
const loading = ref(false)
const wxLoading = ref(false)
const showPhoneAuth = ref(false)

function navigateAfterLogin(user) {
  if (user.role === 'admin' || user.role === 'dsl_admin') {
    setTimeout(() => uni.navigateTo({ url: '/pages/admin/index' }), 800)
  } else {
    setTimeout(() => uni.switchTab({ url: '/pages/home/index' }), 800)
  }
}

const handleLogin = async () => {
  if (!phone.value || !password.value) {
    uni.showToast({ title: '请填写账号和密码', icon: 'none' })
    return
  }

  loading.value = true
  try {
    const res = await request({
      url: '/api/auth/login',
      method: 'POST',
      data: {
        phone: phone.value,
        username: phone.value,
        password: password.value
      }
    })

    if (!res?.success) {
      throw new Error(res?.message || '登录失败')
    }

    uni.setStorageSync('user', JSON.stringify(res.user))
    authStorage.setTokens(res.accessToken, res.refreshToken)
    uni.showToast({ title: '登录成功' })
    navigateAfterLogin(res.user)
  } catch (error) {
    console.error(error)
    const msg = error?.data?.message || error?.message || '登录失败'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    loading.value = false
  }
}

const goRegister = () => {
  uni.navigateTo({ url: '/pages/register/index' })
}

const handleWechatLogin = () => {
  wxLoading.value = true

  uni.login({
    provider: 'weixin',
    success: async (loginRes) => {
      try {
        const res = await request({
          url: '/api/auth/wx-login',
          method: 'POST',
          data: { code: loginRes.code }
        })

        if (!res?.success) {
          throw new Error(res?.message || '微信登录失败')
        }

        uni.setStorageSync('user', JSON.stringify(res.user))
        authStorage.setTokens(res.accessToken, res.refreshToken)

        if (res.isNewUser || !res.user.phone) {
          wxLoading.value = false
          showPhoneAuth.value = true
        } else {
          wxLoading.value = false
          uni.showToast({ title: '登录成功' })
          navigateAfterLogin(res.user)
        }
      } catch (e) {
        wxLoading.value = false
        uni.showToast({ title: e?.message || '微信登录失败', icon: 'none' })
      }
    },
    fail: () => {
      wxLoading.value = false
      uni.showToast({ title: '微信授权失败', icon: 'none' })
    }
  })
}

const handleGetPhone = async (e) => {
  if (e.detail.errMsg !== 'getPhoneNumber:ok') {
    skipPhoneAuth()
    return
  }

  uni.showLoading({ title: '绑定中...' })
  try {
    const res = await request({
      url: '/api/auth/wx-phone',
      method: 'POST',
      data: { code: e.detail.code }
    })

    if (res?.success) {
      uni.setStorageSync('user', JSON.stringify(res.user))
    }
  } catch (err) {
    console.warn('手机号绑定失败:', err?.message)
  }

  showPhoneAuth.value = false
  uni.hideLoading()
  uni.showToast({ title: '登录成功' })

  const stored = uni.getStorageSync('user')
  const user = stored ? JSON.parse(stored) : { role: 'user' }
  navigateAfterLogin(user)
}

const skipPhoneAuth = () => {
  showPhoneAuth.value = false
  uni.showToast({ title: '登录成功' })

  const stored = uni.getStorageSync('user')
  const user = stored ? JSON.parse(stored) : { role: 'user' }
  navigateAfterLogin(user)
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #fff;
  padding: 80px 30px 30px;
  position: relative;
}

.logo-wrap {
  text-align: center;
  margin-bottom: 50px;
}

.login-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #f5f6fa;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.avatar-text {
  font-size: 36px;
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

.login-btn {
  margin-top: 12px;
  height: 48px;
  line-height: 48px;
  border-radius: 24px;
  background-color: #667eea !important;
  font-size: 17px;
  font-weight: bold;
}

.action-links {
  text-align: center;
}

.link-text {
  font-size: 14px;
  color: #667eea;
}

.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 10px 0;
}

.divider .line {
  flex: 1;
  height: 1px;
  background: #ebedf0;
}

.divider .text {
  font-size: 12px;
  color: #969799;
}

.wechat-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 48px;
  border-radius: 24px;
  background-color: #07c160 !important;
  font-size: 16px;
  color: #fff;
}

.wechat-btn::after {
  border: none;
}

.sso-tip {
  text-align: center;
  font-size: 13px;
  color: #969799;
}

.phone-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
}

.modal-content {
  position: relative;
  width: 80%;
  max-width: 320px;
  background: #fff;
  border-radius: 16px;
  padding: 32px 24px;
  text-align: center;
}

.modal-title {
  font-size: 18px;
  font-weight: bold;
  color: #323233;
  margin-bottom: 8px;
}

.modal-desc {
  font-size: 14px;
  color: #969799;
  margin-bottom: 24px;
}

.phone-auth-btn {
  height: 44px;
  line-height: 44px;
  border-radius: 22px;
  background-color: #07c160 !important;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 12px;
}

.phone-auth-btn::after {
  border: none;
}

.skip-btn {
  padding: 8px;
}

.skip-btn text {
  font-size: 14px;
  color: #969799;
}
</style>
