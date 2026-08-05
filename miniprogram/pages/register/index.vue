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

    <!-- 标题 -->
    <view class="welcome">
      <text class="welcome-title">创建新账户</text>
      <text class="welcome-sub">加入无人机产业服务平台</text>
    </view>

    <!-- 表单 -->
    <view class="form">
      <view class="field">
        <text class="field-label">姓名</text>
        <view class="input-box">
          <input
            class="input"
            v-model="name"
            placeholder="请输入您的姓名"
            placeholder-style="color:#C8C9CC;"
          />
        </view>
      </view>

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
        <text class="field-label">设置密码</text>
        <view class="input-box">
          <input
            class="input"
            v-model="password"
            placeholder="请设置6-20位密码"
            :password="!showPwd"
            placeholder-style="color:#C8C9CC;"
          />
          <view class="pwd-toggle" @tap="showPwd = !showPwd">
            <view v-if="!showPwd" class="icon-eye"></view>
            <view v-else class="icon-eye-off"></view>
          </view>
        </view>
      </view>

      <view class="field">
        <text class="field-label">确认密码</text>
        <view class="input-box">
          <input
            class="input"
            v-model="confirm"
            placeholder="请再次输入密码"
            :password="!showConfirm"
            placeholder-style="color:#C8C9CC;"
          />
          <view class="pwd-toggle" @tap="showConfirm = !showConfirm">
            <view v-if="!showConfirm" class="icon-eye"></view>
            <view v-else class="icon-eye-off"></view>
          </view>
        </view>
      </view>

      <view class="agreement" @tap="agreed = !agreed">
        <view class="checkbox" :class="{ checked: agreed }">
          <text v-if="agreed" class="check-mark">✓</text>
        </view>
        <text class="agreement-text">
          我已阅读并同意<text class="agreement-link" @tap.stop="showTerms">《服务协议》</text>和<text class="agreement-link" @tap.stop="showPrivacy">《隐私政策》</text>
        </text>
      </view>

      <view class="submit-btn" :class="{ loading: loading }" @tap="doRegister">
        <text v-if="!loading">注册</text>
        <text v-else>注册中...</text>
      </view>
    </view>

    <view class="bottom">
      <text class="bottom-text">已经有账号？</text>
      <text class="bottom-link" @tap="goLogin">去登录</text>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { request } from '@/utils/request'

const name = ref('')
const phone = ref('')
const password = ref('')
const confirm = ref('')
const showPwd = ref(false)
const showConfirm = ref(false)
const agreed = ref(false)
const loading = ref(false)

const goBack = () => uni.navigateBack()
const goLogin = () => uni.navigateTo({ url: '/pages/login/index' })
const showTerms = () => uni.showToast({ title: '服务协议详情', icon: 'none' })
const showPrivacy = () => uni.showToast({ title: '隐私政策详情', icon: 'none' })

const doRegister = async () => {
  if (!name.value || !phone.value || !password.value) {
    uni.showToast({ title: '请填写完整', icon: 'none' })
    return
  }
  if (!/^1[3-9]\d{9}$/.test(phone.value)) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }
  if (password.value.length < 6) {
    uni.showToast({ title: '密码至少6位', icon: 'none' })
    return
  }
  if (password.value !== confirm.value) {
    uni.showToast({ title: '两次密码输入不一致', icon: 'none' })
    return
  }
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意服务协议', icon: 'none' })
    return
  }
  loading.value = true
  try {
    const res = await request({
      url: '/api/auth/register',
      method: 'POST',
      data: { name: name.value, phone: phone.value, password: password.value }
    })
    if (res.success) {
      uni.showToast({ title: '注册成功', icon: 'success' })
      setTimeout(() => {
        loading.value = false
        uni.navigateBack()
      }, 800)
    } else {
      loading.value = false
      uni.showToast({ title: res.message || '注册失败', icon: 'none' })
    }
  } catch (e) {
    loading.value = false
    const msg = e?.data?.error?.message || e?.message || e?.errMsg || '网络错误'
    uni.showToast({ title: String(msg).substring(0, 30), icon: 'none' })
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh; background: #ffffff;
  display: flex; flex-direction: column; padding: 0 32rpx;
}

.nav-bar { height: 96rpx; display: flex; align-items: center; }
.back-btn { font-size: 44rpx; color: #1a1a1a; }

/* ---- 品牌 ---- */
.brand {
  display: flex; align-items: center; gap: 16rpx;
  padding: 16rpx 0 40rpx;
}
.brand-mark {
  width: 68rpx; height: 68rpx; border-radius: 16rpx;
  background: #0A66C2;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.brand-mark-text { font-size: 36rpx; font-weight: 700; color: #ffffff; }
.brand-name { font-size: 26rpx; font-weight: 500; color: #8E8E93; }

/* ---- 标题 ---- */
.welcome { margin-bottom: 40rpx; }
.welcome-title {
  display: block; font-size: 56rpx; font-weight: 700;
  color: #1a1a1a; letter-spacing: -1rpx;
}
.welcome-sub {
  display: block; font-size: 28rpx; color: #8E8E93; margin-top: 8rpx;
}

/* ---- 表单 ---- */
.form { flex: 1; }
.field { margin-bottom: 24rpx; }
.field-label {
  display: block; font-size: 26rpx; font-weight: 600;
  color: #1a1a1a; margin-bottom: 12rpx;
}
.input-box {
  display: flex; align-items: center;
  height: 104rpx; padding: 0 28rpx;
  background: #fafafa; border-radius: 16rpx;
  border: 2rpx solid transparent; transition: border-color 0.2s;
}
.input-box:focus-within { border-color: #0A66C2; background: #ffffff; }
.input { flex: 1; font-size: 30rpx; color: #1a1a1a; background: transparent; }

/* 密码切换 */
.pwd-toggle {
  width: 48rpx; height: 48rpx;
  display: flex; align-items: center; justify-content: center;
}
.icon-eye {
  width: 36rpx; height: 28rpx; border: 3rpx solid #8E8E93; border-radius: 50%;
  position: relative;
}
.icon-eye::after {
  content: ''; position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 10rpx; height: 10rpx; background: #8E8E93; border-radius: 50%;
}
.icon-eye-off {
  width: 36rpx; height: 3rpx; background: #8E8E93; transform: rotate(-45deg);
}

/* 协议勾选 */
.agreement {
  display: flex; align-items: flex-start; gap: 10rpx;
  margin: 12rpx 0 36rpx;
}
.checkbox {
  width: 36rpx; height: 36rpx; border-radius: 8rpx; border: 3rpx solid #ddd;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; margin-top: 2rpx;
}
.checkbox.checked { background: #0A66C2; border-color: #0A66C2; }
.check-mark { font-size: 24rpx; color: #ffffff; }
.agreement-text { font-size: 24rpx; color: #8E8E93; line-height: 1.6; }
.agreement-link { color: #0A66C2; font-weight: 500; }

/* 注册按钮 */
.submit-btn {
  height: 100rpx; border-radius: 16rpx;
  background: #0A66C2;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8rpx 32rpx rgba(10, 102, 194, 0.25);
}
.submit-btn:active { opacity: 0.9; transform: scale(0.98); }
.submit-btn.loading { opacity: 0.7; pointer-events: none; }
.submit-btn text {
  font-size: 32rpx; font-weight: 700; color: #ffffff; letter-spacing: 4rpx;
}

/* 底部 */
.bottom {
  display: flex; justify-content: center; align-items: center; gap: 10rpx;
  padding: 40rpx 0 calc(40rpx + env(safe-area-inset-bottom));
}
.bottom-text { font-size: 28rpx; color: #8E8E93; }
.bottom-link { font-size: 28rpx; color: #0A66C2; font-weight: 700; }
</style>
