<template>
  <view class="page">
    <!-- 返回 -->
    <view class="nav-bar" :style="navStyle">
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
      <text class="welcome-sub">登录后查看需求、对接意向与商城订单</text>
    </view>

    <!-- 登录区 -->
    <view class="login-area">
      <button class="wx-btn" :class="{ loading: loading }" @tap="doWxLogin">
        <view class="wx-icon">
          <text class="wx-icon-text">微</text>
        </view>
        <text v-if="!loading" class="wx-btn-text">微信一键登录</text>
        <text v-else class="wx-btn-text">登录中...</text>
      </button>

      <view class="other-login" @tap="goPhoneLogin">
        <text class="other-login-text">其他方式登录</text>
        <text class="other-login-arrow">›</text>
      </view>
    </view>

    <!-- 协议（滚动到底部固定） -->
    <view class="agreement" @tap="agreed = !agreed">
      <view class="checkbox" :class="{ checked: agreed }">
        <text v-if="agreed" class="check-mark">✓</text>
      </view>
      <text class="agreement-text">
        我已阅读并同意<text class="agreement-link" @tap.stop="showDoc('terms')">《用户协议》</text>和<text class="agreement-link" @tap.stop="showDoc('privacy')">《隐私政策》</text>
      </text>
    </view>
  </view>
</template>

<script setup>
import { safeBack } from '../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage, getErrorMessage } from '@/utils/request'

const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
// 自定义导航：状态栏留白（JS 方式，微信端 var(--status-bar-height) 不可靠）
const navStyle = computed(() => ({
  paddingTop: statusBarH.value + 'px',
  height: (48 + statusBarH.value) + 'px', // 原 96rpx = 48px + 状态栏
}))

const loading = ref(false)
const agreed = ref(false)

// 来自发布类页面（?from=publish）/ 企业入驻（?from=enterprise）：登录页自身显示提示——不依赖 toast 与跳转时序
const routeFrom = ref('')
onLoad((options) => {
  routeFrom.value = (options && options.from) || ''
  if (routeFrom.value === 'publish') {
    uni.showToast({ title: '请先登录后再使用发布功能', icon: 'none', duration: 1500 })
  } else if (routeFrom.value === 'enterprise') {
    uni.showToast({ title: '请先登录后再申请企业入驻', icon: 'none', duration: 1500 })
  }
})

const goBack = () => safeBack()

const goPhoneLogin = () => uni.navigateTo({ url: '/pages/login/phone' + (routeFrom.value ? '?from=' + routeFrom.value : '') })

const showDoc = (type) => uni.navigateTo({ url: '/pages/agreement/index?type=' + type })

// 微信一键登录：用户主动点击后才 wx.login + 后端换 token（自动注册语义在授权之后）
const doWxLogin = () => {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意《用户协议》和《隐私政策》', icon: 'none' })
    return
  }
  loading.value = true
  uni.login({
    provider: 'weixin',
    success: async (loginRes) => {
      try {
        const res = await request({
          url: '/api/v1/auth/wechat/login',
          method: 'POST',
          data: { code: loginRes.code }
        })
        if (res?.access_token && res?.user) {
          authStorage.setTokens(res.access_token, res.refresh_token)
          uni.setStorageSync('user', JSON.stringify(res.user))
          uni.showToast({ title: '登录成功', icon: 'success' })
          setTimeout(() => {
            loading.value = false
            // 带来源的流程（如企业入驻）：登录成功回到来源页（表单在页面栈中），不回首页丢流程
            if (routeFrom.value === 'enterprise') uni.navigateBack({ delta: 1 })
            else uni.switchTab({ url: '/pages/home/index' })
          }, 600)
        } else {
          loading.value = false
          uni.showToast({ title: '登录失败，请重试', icon: 'none' })
        }
      } catch (e) {
        loading.value = false
        uni.showToast({ title: getErrorMessage(e) || '网络错误，请重试', icon: 'none' })
      }
    },
    fail: () => {
      loading.value = false
      uni.showToast({ title: '微信登录失败', icon: 'none' })
    },
  })
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

.nav-bar {
  box-sizing: border-box;
  display: flex;
  align-items: center;
}
.back-btn { font-size: 44rpx; color: var(--color-text); }

/* ---- 品牌 ---- */
.brand {
  display: flex; align-items: center; gap: 16rpx;
  padding: 16rpx 0 48rpx;
}
.brand-mark {
  width: 68rpx; height: 68rpx; border-radius: 16rpx;
  background: #0A66C2;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.brand-mark-text { font-size: 36rpx; font-weight: 700; color: #ffffff; }
.brand-name { font-size: 26rpx; font-weight: 500; color: var(--color-text-secondary); }

/* ---- 欢迎语 ---- */
.welcome { margin-bottom: 48rpx; }
.welcome-title {
  display: block; font-size: 56rpx; font-weight: 700;
  color: var(--color-text); letter-spacing: -1rpx;
}
.welcome-sub {
  display: block; font-size: 28rpx; color: var(--color-text-secondary);
  margin-top: 12rpx;
}

/* ---- 登录区 ---- */
.login-area { flex: 1; }

.wx-btn {
  display: flex; align-items: center; justify-content: center; gap: 16rpx;
  height: 100rpx; border-radius: 16rpx;
  background: #07C160;
  box-shadow: 0 8rpx 32rpx rgba(7, 193, 96, 0.28);
  border: none;
  margin: 0;
}
.wx-btn::after { border: none; }
.wx-btn:active { opacity: 0.9; transform: scale(0.98); }
.wx-btn.loading { opacity: 0.7; pointer-events: none; }
.wx-icon {
  width: 44rpx; height: 44rpx; border-radius: 10rpx;
  background: #ffffff; display: flex; align-items: center; justify-content: center;
}
.wx-icon-text { font-size: 26rpx; font-weight: 700; color: #07C160; }
.wx-btn-text { font-size: 32rpx; font-weight: 700; color: #ffffff; letter-spacing: 2rpx; }

.other-login {
  display: flex; align-items: center; justify-content: center; gap: 6rpx;
  margin-top: 48rpx; padding: 16rpx 0;
}
.other-login-text { font-size: 28rpx; font-weight: 600; color: var(--color-primary); }
.other-login-arrow { font-size: 32rpx; font-weight: 300; color: var(--color-primary); }

/* ---- 协议 ---- */
.agreement {
  display: flex; align-items: center; justify-content: center; gap: 10rpx;
  padding: 24rpx 0 calc(24rpx + env(safe-area-inset-bottom));
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
</style>
