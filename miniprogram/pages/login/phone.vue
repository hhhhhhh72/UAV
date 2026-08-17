<template>
  <view class="page">
    <!-- 返回 -->
    <view class="nav-bar">
      <text class="back-btn" @tap="goBack">←</text>
    </view>

    <!-- 标题 -->
    <view class="welcome">
      <text class="welcome-title">手机号登录</text>
      <text class="welcome-sub">未注册的手机号验证码登录后将自动创建账号</text>
    </view>

    <!-- Tab 切换：密码 / 验证码 -->
    <view class="tabs">
      <view class="tab-item" :class="{ active: mode === 'password' }" @tap="mode = 'password'">
        <text class="tab-text">密码登录</text>
      </view>
      <view class="tab-item" :class="{ active: mode === 'code' }" @tap="mode = 'code'">
        <text class="tab-text">验证码登录</text>
      </view>
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

      <!-- 密码登录 -->
      <view class="field" v-if="mode === 'password'">
        <text class="field-label">密码</text>
        <view class="input-box">
          <input
            class="input"
            v-model="password"
            placeholder="请输入密码"
            :password="!showPwd"
            placeholder-style="color:#C8C9CC;"
          />
          <view class="pwd-toggle" @tap="showPwd = !showPwd">
            <view v-if="!showPwd" class="icon-eye"></view>
            <view v-else class="icon-eye-off"></view>
          </view>
        </view>
        <view class="field-extra">
          <text class="forgot" @tap="forgotPwd">忘记密码？</text>
        </view>
      </view>

      <!-- 验证码登录 -->
      <view class="field" v-if="mode === 'code'">
        <text class="field-label">验证码</text>
        <view class="input-box">
          <input
            class="input"
            v-model="code"
            placeholder="请输入6位验证码"
            type="number"
            maxlength="6"
            placeholder-style="color:#C8C9CC;"
          />
          <view class="send-btn" :class="{ disabled: countdown > 0 }" @tap="sendCode">
            <text class="send-btn-text">{{ countdown > 0 ? countdown + 's 后重发' : '获取验证码' }}</text>
          </view>
        </view>
      </view>

      <view class="submit-btn" :class="{ loading: loading }" @tap="doLogin">
        <text v-if="!loading">登录</text>
        <text v-else>登录中...</text>
      </view>
    </view>

    <!-- 协议 -->
    <view class="agreement" @tap="agreed = !agreed">
      <view class="checkbox" :class="{ checked: agreed }">
        <text v-if="agreed" class="check-mark">✓</text>
      </view>
      <text class="agreement-text">
        我已阅读并同意<text class="agreement-link" @tap.stop="showDoc('terms')">《用户协议》</text>和<text class="agreement-link" @tap.stop="showDoc('privacy')">《隐私政策》</text>
      </text>
    </view>

    <view class="bottom">
      <text class="bottom-text">还没有账号？</text>
      <text class="bottom-link" @tap="goRegister">立即注册</text>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onUnload } from '@dcloudio/uni-app'
import { request, authStorage, getErrorMessage } from '@/utils/request'

const mode = ref('password') // password | code
const phone = ref('')
const password = ref('')
const code = ref('')
const showPwd = ref(false)
const agreed = ref(false)
const loading = ref(false)
const countdown = ref(0)
let timer = null

onUnload(() => {
  if (timer) clearInterval(timer)
})

const goBack = () => uni.navigateBack()
const goRegister = () => uni.navigateTo({ url: '/pages/register/index' })
const showDoc = (type) => uni.navigateTo({ url: '/pages/agreement/index?type=' + type })
const forgotPwd = () => uni.showToast({ title: '请联系管理员重置密码', icon: 'none' })

const validPhone = () => {
  if (!/^1[3-9]\d{9}$/.test(phone.value)) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return false
  }
  return true
}

// 发送验证码（60s 倒计时）
const sendCode = () => {
  if (countdown.value > 0) return
  if (!validPhone()) return
  request({
    url: '/api/auth/send-code',
    method: 'POST',
    data: { phone: phone.value },
  })
    .then((res) => {
      // 开发态后端回显 dev_code，方便调试时直接填入
      const devCode = res?.data?.dev_code || res?.dev_code
      uni.showToast({ title: devCode ? '验证码已发送：' + devCode : '验证码已发送', icon: 'none' })
      countdown.value = 60
      timer = setInterval(() => {
        countdown.value -= 1
        if (countdown.value <= 0) clearInterval(timer)
      }, 1000)
    })
    .catch(() => uni.showToast({ title: '发送失败，请重试', icon: 'none' }))
}

const finishLogin = (res) => {
  authStorage.setTokens(res.accessToken, res.refreshToken)
  uni.setStorageSync('user', JSON.stringify(res.user))
  uni.showToast({ title: '登录成功', icon: 'success' })
  setTimeout(() => {
    loading.value = false
    uni.switchTab({ url: '/pages/home/index' })
  }, 600)
}

const doLogin = async () => {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意《用户协议》和《隐私政策》', icon: 'none' })
    return
  }
  if (!validPhone()) return
  if (mode.value === 'password' && !password.value) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return
  }
  if (mode.value === 'code' && code.value.length !== 6) {
    uni.showToast({ title: '请输入6位验证码', icon: 'none' })
    return
  }

  loading.value = true
  try {
    const url = mode.value === 'password' ? '/api/auth/login' : '/api/auth/login-code'
    const data = mode.value === 'password'
      ? { phone: phone.value, password: password.value }
      : { phone: phone.value, code: code.value }
    const res = await request({ url, method: 'POST', data })
    if (res.accessToken) {
      finishLogin(res)
    } else {
      loading.value = false
      uni.showToast({ title: res.message || '登录失败，请重试', icon: 'none' })
    }
  } catch (e) {
    loading.value = false
    const msg = getErrorMessage(e) || '网络错误'
    uni.showToast({ title: String(msg).substring(0, 30), icon: 'none' })
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  padding: 0 32rpx;
  box-sizing: border-box;
}

.nav-bar { height: 96rpx; display: flex; align-items: center; }
.back-btn { font-size: 44rpx; color: var(--color-text); }

/* ---- 标题 ---- */
.welcome { margin-bottom: 32rpx; }
.welcome-title {
  display: block; font-size: 56rpx; font-weight: 700;
  color: var(--color-text); letter-spacing: -1rpx;
}
.welcome-sub {
  display: block; font-size: 24rpx; color: var(--color-text-secondary);
  margin-top: 12rpx;
}

/* ---- Tab ---- */
.tabs {
  display: flex;
  border-bottom: 2rpx solid var(--color-divider);
  margin-bottom: 32rpx;
}
.tab-item {
  position: relative;
  padding: 16rpx 8rpx;
  margin-right: 48rpx;
}
.tab-item.active .tab-text { color: var(--color-primary); font-weight: 700; }
.tab-item.active::after {
  content: '';
  position: absolute; left: 0; right: 0; bottom: -2rpx;
  height: 6rpx; border-radius: 6rpx; background: var(--color-primary);
}
.tab-text { font-size: 30rpx; color: var(--color-text-secondary); }

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
  background: #fafafa; border-radius: 16rpx;
  border: 2rpx solid transparent;
}
.input-box:focus-within { border-color: var(--color-primary); background: #ffffff; }
.input { flex: 1; font-size: 30rpx; color: var(--color-text); background: transparent; }
.field-extra { display: flex; justify-content: flex-end; margin-top: 12rpx; }
.forgot { font-size: 24rpx; color: var(--color-primary); font-weight: 600; }

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

/* 获取验证码 */
.send-btn { flex-shrink: 0; margin-left: 16rpx; padding: 12rpx 20rpx; }
.send-btn-text { font-size: 26rpx; font-weight: 600; color: var(--color-primary); }
.send-btn.disabled .send-btn-text { color: var(--color-text-placeholder); }

/* 登录按钮 */
.submit-btn {
  height: 100rpx; border-radius: 16rpx;
  background: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8rpx 32rpx rgba(10, 102, 194, 0.25);
  margin-top: 32rpx;
}
.submit-btn:active { opacity: 0.9; transform: scale(0.98); }
.submit-btn.loading { opacity: 0.7; pointer-events: none; }
.submit-btn text {
  font-size: 32rpx; font-weight: 700; color: #ffffff; letter-spacing: 4rpx;
}

/* ---- 协议 ---- */
.agreement {
  display: flex; align-items: center; justify-content: center; gap: 10rpx;
  padding: 16rpx 0;
}
.checkbox {
  width: 32rpx; height: 32rpx; border-radius: 50%;
  border: 2rpx solid #D0D5DD; background: #ffffff;
  display: flex; align-items: center; justify-content: center;
  box-sizing: border-box;
}
.checkbox.checked { background: var(--color-primary); border-color: var(--color-primary); }
.check-mark { font-size: 20rpx; color: #ffffff; }
.agreement-text { font-size: 24rpx; color: var(--color-text-secondary); }
.agreement-link { font-size: 24rpx; color: var(--color-primary); }

/* ---- 底部 ---- */
.bottom {
  display: flex; justify-content: center; align-items: center; gap: 10rpx;
  padding: 24rpx 0 calc(24rpx + env(safe-area-inset-bottom));
}
.bottom-text { font-size: 28rpx; color: var(--color-text-secondary); }
.bottom-link { font-size: 28rpx; color: var(--color-primary); font-weight: 700; }
</style>
