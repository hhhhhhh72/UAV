<template>
  <div class="login-page">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="circle circle-1"></div>
      <div class="circle circle-2"></div>
      <div class="circle circle-3"></div>
    </div>

    <!-- 居中卡片 -->
    <div class="login-card">
      <div class="login-header">
        <div class="logo-badge">
          <icon-send :size="28" :style="{ color: '#fff' }" />
        </div>
        <h1 class="login-title">无人机产业综合服务平台</h1>
        <p class="login-subtitle">管理后台 · 账号登录</p>
      </div>

      <!-- 账号密码登录 -->
      <a-form
        ref="formRef"
        :model="loginForm"
        :rules="loginRules"
        size="large"
        layout="vertical"
        @submit.prevent="onSubmit"
      >
        <a-form-item field="phone" hide-label>
          <a-input
            v-model="loginForm.phone"
            placeholder="请输入手机号"
            allow-clear
            @press-enter="onSubmit"
          >
            <template #prefix><icon-phone /></template>
          </a-input>
        </a-form-item>
        <a-form-item field="password" hide-label>
          <a-input-password
            v-model="loginForm.password"
            placeholder="请输入密码"
            allow-clear
            @press-enter="onSubmit"
          >
            <template #prefix><icon-lock /></template>
          </a-input-password>
        </a-form-item>
        <a-button
          class="submit-btn"
          type="primary"
          long
          :loading="loading"
          @click="onSubmit"
        >
          登 录
        </a-button>

        <!-- 仅本地开发环境显示（import.meta.env.DEV 在构建时被替换，生产自动剔除） -->
        <a-button
          v-if="isDev"
          class="dev-btn"
          long
          :loading="devLoading"
          @click="devQuickLogin"
        >
          开发者快捷登录
        </a-button>
      </a-form>

      <p class="footer-tip">登录即表示同意用户协议和隐私政策</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios, { authStorage } from '@/utils/http'
import { showFailToast, showSuccessToast } from '@/utils/feedback'

const router = useRouter()
const formRef = ref(null)
const loading = ref(false)

// 开发者快捷登录：仅本地 dev 环境（后端须 ADMIN_DEV_MODE=true，生产为 403）
const isDev = import.meta.env.DEV
const devLoading = ref(false)

const devQuickLogin = async () => {
  devLoading.value = true
  try {
    const res = await axios.post('/api/v1/admin/token', { role: 'platform_admin' })
    const data = res.data || {}
    const accessToken = data.access_token || data.accessToken
    if (!accessToken) {
      showFailToast('开发者登录失败：未获取到令牌')
      return
    }
    const user = data.user || { id: 'admin-dev', role: 'platform_admin' }
    saveSession(user, accessToken)
    router.push('/admin')
  } catch (error) {
    const message = error?.response?.data?.error?.message
    showFailToast(message || '开发者登录失败（请确认后端 ADMIN_DEV_MODE=true）')
  } finally {
    devLoading.value = false
  }
}

const loginForm = ref({
  phone: '',
  password: ''
})

const loginRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1\d{10}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 保存登录态：token 走 authStorage，用户信息仅存 { id, role }
const saveSession = (user, accessToken, refreshToken) => {
  authStorage.setTokens(accessToken, refreshToken)
  if (user && typeof user === 'object' && user.id) {
    localStorage.setItem('user', JSON.stringify({ id: user.id, role: user.role }))
  }
}

const afterLogin = (user, accessToken, refreshToken) => {
  saveSession(user, accessToken, refreshToken)
  showSuccessToast('登录成功')
  router.push('/admin')
}

// 账号密码登录：POST /api/auth/login
const onSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    const res = await axios.post('/api/auth/login', {
      phone: loginForm.value.phone,
      password: loginForm.value.password
    })
    const data = res.data || {}
    if (!data.success) {
      showFailToast('账号或密码错误')
      return
    }
    afterLogin(data.user, data.accessToken, data.refreshToken)
  } catch (error) {
    const message = error?.response?.data?.error?.message
    showFailToast(message || '账号或密码错误')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0a66c2 0%, #053a6e 100%);
  position: relative;
  overflow: hidden;
  padding: 24px;
}

.bg-decoration {
  position: absolute;
  inset: 0;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
}

.circle-1 {
  width: 340px;
  height: 340px;
  top: -120px;
  right: -80px;
}

.circle-2 {
  width: 240px;
  height: 240px;
  bottom: -80px;
  left: -60px;
}

.circle-3 {
  width: 120px;
  height: 120px;
  top: 18%;
  left: 14%;
  background: rgba(29, 212, 168, 0.12);
}

/* Arco 主题覆盖为品牌蓝 #0A66C2
   注意：必须用 RGB 分量格式（10,102,194），Arco 组件样式是
   rgb(var(--primary-6)) 包装；十六进制会生成 rgb(#0a66c2) 非法值
   → 按钮背景失效变白（登录按钮曾因此"消失"）。 */
.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 16px;
  padding: 40px 36px 28px;
  box-shadow: 0 16px 48px rgba(2, 32, 71, 0.35);
  --primary-1: 233,241,251;
  --primary-2: 194,216,240;
  --primary-3: 156,191,230;
  --primary-4: 116,166,221;
  --primary-5: 69,135,209;
  --primary-6: 10,102,194;
  --primary-7: 8,84,157;
  --primary-8: 6,63,116;
}

.login-header {
  text-align: center;
  margin-bottom: 28px;
}

.logo-badge {
  width: 56px;
  height: 56px;
  margin: 0 auto 14px;
  border-radius: 14px;
  background: linear-gradient(135deg, #0a66c2 0%, #1dd4a8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 20px rgba(10, 102, 194, 0.35);
}

.login-title {
  font-size: 20px;
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 6px;
}

.login-subtitle {
  font-size: 13px;
  color: #86909c;
}

/* 表单与按钮间距（Arco 默认 20px 过大，收紧保持紧凑形态） */
.login-card :deep(.arco-form-item) {
  margin-bottom: 18px;
}

.submit-btn {
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
}

.dev-btn {
  margin-top: 12px;
  height: 40px;
  font-size: 14px;
  border-radius: 8px;
  border-style: dashed;
  color: #0a66c2;
}

.footer-tip {
  margin-top: 24px;
  text-align: center;
  font-size: 12px;
  color: #c0c4cc;
}

@media (max-width: 480px) {
  .login-card {
    padding: 32px 24px 24px;
  }
}
</style>
