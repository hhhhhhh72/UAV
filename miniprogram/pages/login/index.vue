<template>
  <view class="page">
    <!-- 返回 -->
    <view class="nav-bar">
      <text class="back-btn" @tap="goBack">←</text>
    </view>

    <!-- 品牌标识 -->
    <view class="brand">
      <view class="brand-mark">
        <text class="brand-mark-text">U</text>
      </view>
      <text class="brand-name">无人机产业综合服务平台</text>
    </view>

    <!-- 欢迎语 -->
    <view class="welcome">
      <text class="welcome-title">欢迎回来</text>
      <text class="welcome-sub">登录您的账号</text>
    </view>

    <!-- 表单 -->
    <view class="form">
      <view class="field">
        <text class="field-label">手机号</text>
        <view class="input-box">
          <input
            class="input"
            v-model="phone"
            placeholder="请输入手机号"
            type="number"
            maxlength="11"
            placeholder-style="color:#C8C9CC;"
          />
        </view>
      </view>

      <view class="field">
        <text class="field-label">密码</text>
        <view class="input-box">
          <input
            class="input"
            v-model="password"
            placeholder="请输入密码"
            :type="showPwd ? 'text' : 'password'"
            placeholder-style="color:#C8C9CC;"
          />
          <view class="pwd-toggle" @tap="showPwd = !showPwd">
            <view v-if="!showPwd" class="icon-eye"></view>
            <view v-else class="icon-eye-off"></view>
          </view>
        </view>
      </view>

      <view class="options">
        <view class="remember" @tap="remember = !remember">
          <view class="checkbox" :class="{ checked: remember }">
            <text v-if="remember" class="check-mark">✓</text>
          </view>
          <text class="remember-text">记住我</text>
        </view>
        <text class="forgot" @tap="forgotPwd">忘记密码？</text>
      </view>

      <view class="submit-btn" :class="{ loading: loading }" @tap="doLogin">
        <text v-if="!loading">登录</text>
        <text v-else>登录中...</text>
      </view>

      <!-- 微信快捷登录 -->
      <!-- #ifdef MP-WEIXIN -->
      <view class="divider">
        <view class="divider-line"></view>
        <text class="divider-text">其他方式登录</text>
        <view class="divider-line"></view>
      </view>

      <button
        class="wechat-btn"
        open-type="getPhoneNumber"
        @getphonenumber="handlePhone"
      >
        <view class="wechat-icon-wrap">
          <text class="wechat-icon">◎</text>
        </view>
        <text>微信一键登录</text>
      </button>
      <!-- #endif -->
    </view>

    <view class="bottom">
      <text class="bottom-text">还没有账号？</text>
      <text class="bottom-link" @tap="goRegister">立即注册</text>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request, authStorage } from '@/utils/request'

const phone = ref('')
const password = ref('')
const showPwd = ref(false)
const remember = ref(true)

const loading = ref(false)

const goBack = () => uni.navigateBack()
const goRegister = () => uni.navigateTo({ url: '/pages/register/index' })
const forgotPwd = () => uni.showToast({ title: '请联系管理员重置密码', icon: 'none' })

const doLogin = async () => {
  if (!phone.value || !password.value) {
    uni.showToast({ title: '请填写完整', icon: 'none' })
    return
  }
  loading.value = true
  try {
    const res = await request({
      url: '/api/auth/login',
      method: 'POST',
      data: { phone: phone.value, password: password.value }
    })
    if (res.accessToken) {
      authStorage.setTokens(res.accessToken, res.refreshToken)
      uni.setStorageSync('user', JSON.stringify(res.user))
      if (remember.value) uni.setStorageSync('savedPhone', phone.value)
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        loading.value = false
        uni.switchTab({ url: '/pages/home/index' })
      }, 600)
    } else {
      loading.value = false
      uni.showToast({ title: '账号或密码错误', icon: 'none' })
    }
  } catch {
    loading.value = false
    uni.showToast({ title: '网络错误，请重试', icon: 'none' })
  }
}

const handlePhone = () => {
  uni.switchTab({ url: '/pages/home/index' })
}

// 恢复已保存手机号
const saved = uni.getStorageSync('savedPhone')
if (saved) phone.value = saved
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  padding: 0 32rpx;
}

.nav-bar { height: 96rpx; display: flex; align-items: center; }
.back-btn { font-size: 44rpx; color: var(--color-text); }

/* ---- 品牌 ---- */
.brand {
  display: flex; align-items: center; gap: 16rpx;
  padding: 16rpx 0 40rpx;
}
.brand-mark {
  width: 68rpx; height: 68rpx; border-radius: 20rpx;
  background: linear-gradient(135deg, #0A66C2, #1DD4A8);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.brand-mark-text { font-size: 36rpx; font-weight: 700; color: #ffffff; }
.brand-name { font-size: 26rpx; font-weight: 500; color: var(--color-text-secondary); }

/* ---- 欢迎语 ---- */
.welcome { margin-bottom: 40rpx; }
.welcome-title {
  display: block; font-size: 56rpx; font-weight: 700;
  color: var(--color-text); letter-spacing: -1rpx;
}
.welcome-sub {
  display: block; font-size: 28rpx; color: var(--color-text-secondary);
  margin-top: 8rpx;
}

/* ---- 表单 ---- */
.form { flex: 1; }
.field { margin-bottom: 24rpx; }
.field-label {
  display: block; font-size: 26rpx; font-weight: 600;
  color: var(--color-text); margin-bottom: 12rpx;
}
.input-box {
  display: flex; align-items: center;
  height: 104rpx; padding: 0 28rpx;
  background: #fafafa; border-radius: 24rpx;
  border: 2rpx solid transparent; transition: border-color 0.2s;
}
.input-box:focus-within { border-color: var(--color-primary); background: #ffffff; }
.input { flex: 1; font-size: 30rpx; color: var(--color-text); background: transparent; }

/* 密码切换 */
.pwd-toggle {
  width: 48rpx; height: 48rpx;
  display: flex; align-items: center; justify-content: center;
}
.icon-eye {
  width: 36rpx; height: 28rpx; border: 3rpx solid var(--color-text-secondary); border-radius: 50%;
  position: relative;
}
.icon-eye::after {
  content: ''; position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 10rpx; height: 10rpx; background: var(--color-text-secondary); border-radius: 50%;
}
.icon-eye-off {
  width: 36rpx; height: 3rpx; background: var(--color-text-secondary); transform: rotate(-45deg);
}

/* 记住我 + 忘记密码 */
.options {
  display: flex; align-items: center; justify-content: space-between;
  margin: 16rpx 0 36rpx;
}
.remember { display: flex; align-items: center; gap: 10rpx; }
.checkbox {
  width: 36rpx; height: 36rpx; border-radius: 8rpx; border: 3rpx solid #ddd;
  display: flex; align-items: center; justify-content: center;
}
.checkbox.checked { background: var(--color-primary); border-color: var(--color-primary); }
.check-mark { font-size: 24rpx; color: #ffffff; }
.remember-text { font-size: 26rpx; color: var(--color-text); font-weight: 500; }
.forgot { font-size: 26rpx; color: var(--color-primary); font-weight: 600; }

/* 登录按钮 */
.submit-btn {
  height: 100rpx; border-radius: 50rpx;
  background: linear-gradient(135deg, var(--color-primary), #1677D4);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8rpx 32rpx rgba(10, 102, 194, 0.25);
}
.submit-btn:active { opacity: 0.9; transform: scale(0.98); }
.submit-btn.loading { opacity: 0.7; pointer-events: none; }
.submit-btn text {
  font-size: 32rpx; font-weight: 700; color: #ffffff; letter-spacing: 4rpx;
}

/* 分割线 */
.divider {
  display: flex; align-items: center; gap: 20rpx;
  margin: 48rpx 0 36rpx;
}
.divider-line { flex: 1; height: 1rpx; background: var(--color-divider); }
.divider-text { font-size: 24rpx; color: var(--color-text-secondary); font-weight: 500; }

/* 微信登录 */
.wechat-btn {
  display: flex; align-items: center; justify-content: center; gap: 16rpx;
  height: 96rpx; background: #ffffff; border: 2rpx solid var(--color-divider);
  border-radius: 50rpx; font-size: 28rpx; font-weight: 600; color: var(--color-text);
}
.wechat-btn::after { border: none; }
.wechat-icon-wrap {
  width: 40rpx; height: 40rpx;
  display: flex; align-items: center; justify-content: center;
}
.wechat-icon { font-size: 36rpx; color: #1DD4A8; }

/* 底部 */
.bottom {
  display: flex; justify-content: center; align-items: center; gap: 10rpx;
  padding: 40rpx 0 calc(40rpx + env(safe-area-inset-bottom));
}
.bottom-text { font-size: 28rpx; color: var(--color-text-secondary); }
.bottom-link { font-size: 28rpx; color: var(--color-primary); font-weight: 700; }
</style>
