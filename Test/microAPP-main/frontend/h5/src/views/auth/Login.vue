<template>
  <div class="login-page">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="circle circle-1"></div>
      <div class="circle circle-2"></div>
      <div class="circle circle-3"></div>
    </div>

    <!-- Logo和标题 -->
    <div class="header-section">
      <div class="logo-container">
        <van-icon name="guide-o" size="60" color="#667eea" />
      </div>
      <h1 class="app-title">低空综合服务平台</h1>
      <p class="app-slogan">专业 · 高效 · 安全</p>
    </div>

    <!-- 登录表单 -->
    <div class="form-container">
      <van-form @submit="onLogin">
        <van-field
          v-model="phone"
          type="tel"
          placeholder="请输入手机号"
          :rules="[{ pattern: /^1\d{10}$/, message: '请输入正确的手机号' }]"
        >
          <template #left-icon>
            <van-icon name="phone-o" />
          </template>
        </van-field>

        <van-field
          v-model="code"
          type="digit"
          placeholder="请输入验证码"
        >
          <template #left-icon>
            <van-icon name="shield-o" />
          </template>
          <template #button>
            <van-button
              size="small"
              type="primary"
              plain
              :disabled="countdown > 0"
              @click.prevent="sendCode"
            >
              {{ countdown > 0 ? `${countdown}秒后重发` : '获取验证码' }}
            </van-button>
          </template>
        </van-field>

        <van-button
          round
          block
          type="primary"
          native-type="submit"
          class="login-button"
        >
          登录 / 注册
        </van-button>
      </van-form>

      <div class="footer-tip">
        <p>登录即表示同意用户协议和隐私政策</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showSuccessToast, showLoadingToast } from 'vant'

const router = useRouter()
const phone = ref('')
const code = ref('')
const countdown = ref(0)

const sendCode = () => {
  if (!phone.value) {
    showToast('请输入手机号')
    return
  }
  if (!/^1\d{10}$/.test(phone.value)) {
    showToast('请输入正确的手机号')
    return
  }

  showSuccessToast('验证码已发送')
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
    }
  }, 1000)
}

const onLogin = () => {
  showLoadingToast({
    message: '登录中...',
    forbidClick: true,
    duration: 1000
  })

  setTimeout(() => {
    showSuccessToast('登录成功')
    router.replace('/home')
  }, 1000)
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #f5f5f5;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.bg-decoration {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 60%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  z-index: 0;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}

.circle-1 {
  width: 300px;
  height: 300px;
  top: -100px;
  right: -100px;
}

.circle-2 {
  width: 200px;
  height: 200px;
  top: 20%;
  left: -50px;
}

.circle-3 {
  width: 150px;
  height: 150px;
  top: 40%;
  right: 20%;
}

.header-section {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: 60px 20px 40px;
  color: #fff;
}

.logo-container {
  width: 90px;
  height: 90px;
  background: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.app-title {
  font-size: 26px;
  font-weight: bold;
  margin-bottom: 8px;
}

.app-slogan {
  font-size: 14px;
  opacity: 0.9;
}

.form-container {
  position: relative;
  z-index: 1;
  background: #fff;
  margin: 0 16px;
  padding: 24px 20px;
  border-radius: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.form-container :deep(.van-cell) {
  padding: 16px 12px;
  margin-bottom: 12px;
  background: #f7f8fa;
  border-radius: 8px;
}

.login-button {
  margin-top: 24px;
  height: 44px;
  font-size: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

.footer-tip {
  text-align: center;
  margin-top: 20px;
}

.footer-tip p {
  font-size: 12px;
  color: #969799;
}
</style>

