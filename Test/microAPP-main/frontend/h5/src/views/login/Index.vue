<template>
  <div class="login-page">
    <van-nav-bar
      title="登录"
      left-arrow
      @click-left="$router.back()"
    >
      <template #right>
        <van-icon name="wap-home-o" size="18" class="nav-home" @click="goHome" />
      </template>
    </van-nav-bar>

    <div class="login-container">
      <!-- SSO 自动登录状态 -->
      <div v-if="autoLoginStatus" class="auto-login-area">
        <div class="logo-area">
          <div class="login-avatar">
            <van-loading type="spinner" size="40" color="#667eea" />
          </div>
          <h2 class="app-title">低空综合服务平台</h2>
        </div>
        <div class="auto-login-tip">
          <span>正在验证登录信息...</span>
        </div>
      </div>

      <!-- 正常登录表单 -->
      <template v-else>
        <div class="logo-area">
          <div class="login-avatar" aria-label="default avatar">
            <van-icon name="contact" size="40" color="#8e8e93" />
          </div>
          <h2 class="app-title">低空综合服务平台</h2>
        </div>

        <van-form @submit="onPasswordLogin" class="login-form">
          <van-cell-group inset>
            <van-field
              v-model="loginForm.username"
              name="username"
              label="账号"
              placeholder="请输入手机号/用户名"
              :rules="[{ required: true, message: '请填写账号' }]"
            />
            <van-field
              v-model="loginForm.password"
              type="password"
              name="password"
              label="密码"
              placeholder="请输入密码"
              :rules="[{ required: true, message: '请填写密码' }]"
            />
          </van-cell-group>
          <div style="margin: 24px 16px;">
            <van-button round block type="primary" native-type="submit" :loading="loading">
              登录
            </van-button>
          </div>
        </van-form>

        <div class="action-links">
          <span @click="goRegister">还没有账号？立即注册</span>
        </div>

        <div class="wechat-login">
          <van-divider>其他登录方式</van-divider>
          <div class="wechat-btn-wrapper">
            <van-button
              round
              block
              type="primary"
              color="#07c160"
              icon="wechat"
              @click="onWechatLogin"
              :loading="wechatLoading"
            >
              微信授权登录
            </van-button>
          </div>
        </div>

        <div class="sso-tip">
          <p>从畅行温州平台进入将自动登录</p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showSuccessToast, showFailToast } from 'vant'
import axios, { authStorage } from '@/utils/http'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const autoLoginStatus = ref(false)
const wechatLoading = ref(false)

const loginForm = ref({
  username: '',
  password: ''
})

const goHome = () => {
  router.replace('/home')
}

const goRegister = () => {
  router.push('/register')
}

// 账号密码登录
const onPasswordLogin = async (values) => {
  loading.value = true
  try {
    const res = await axios.post('/api/auth/login', {
      phone: values.username,
      username: values.username,
      password: values.password
    })
    if (!res.data?.success) {
      throw new Error(res.data?.message || '登录失败')
    }
    localStorage.setItem('user', JSON.stringify(res.data.user))
    authStorage.setTokens(res.data.accessToken, res.data.refreshToken)
    showSuccessToast('登录成功')
    
    // 根据用户角色跳转
    const user = res.data.user
    if (user.role === 'admin' || user.role === 'dsl_admin' || user.role === 'study_admin') {
      router.push('/admin')
    } else {
      router.push('/home')
    }
  } catch (error) {
    console.error(error)
    showFailToast(error?.response?.data?.message || error?.message || '登录失败')
  } finally {
    loading.value = false
  }
}

// 微信授权登录
const onWechatLogin = async () => {
  wechatLoading.value = true
  try {
    const res = await axios.get('/api/auth/wechat-oauth-url', {
      params: {
        redirectUrl: window.location.origin + '/home'
      }
    })
    
    if (!res.data?.success || !res.data?.authUrl) {
      throw new Error(res.data?.message || '获取微信授权URL失败')
    }
    
    // 重定向到微信授权页面
    window.location.href = res.data.authUrl
  } catch (error) {
    console.error(error)
    showFailToast(error?.response?.data?.message || error?.message || '微信授权登录失败')
  } finally {
    wechatLoading.value = false
  }
}

// 处理微信授权回调
const handleWechatCallback = () => {
  const wechatAuth = route.query.wechat_auth
  const userData = route.query.user
  const tokensData = route.query.tokens
  
  if (wechatAuth === '1' && userData && tokensData) {
    try {
      // 解析用户信息
      const user = JSON.parse(atob(userData))
      const tokens = JSON.parse(atob(tokensData))
      
      // 保存到本地存储
      localStorage.setItem('user', JSON.stringify(user))
      authStorage.setTokens(tokens.accessToken, tokens.refreshToken)
      
      showSuccessToast('微信授权登录成功')
      
      // 根据用户角色跳转
      if (user.role === 'admin' || user.role === 'dsl_admin' || user.role === 'study_admin') {
        router.push('/admin')
      } else {
        router.push('/home')
      }
    } catch (error) {
      console.error('解析微信登录数据失败:', error)
      showFailToast('微信授权登录失败，请重试')
    }
  }
  
  // 处理错误
  const error = route.query.error
  if (error) {
    showFailToast('微信授权登录失败，请重试')
  }
}

// 授权码自动登录（SSO）- 从畅行温州平台跳转时自动触发
const ssoLogin = async (code) => {
  autoLoginStatus.value = true
  try {
    const res = await axios.post('/api/sso/login', { authcode: code })
    if (!res.data?.success) {
      throw new Error(res.data?.message || '授权登录失败')
    }
    localStorage.setItem('user', JSON.stringify(res.data.user))
    authStorage.setTokens(res.data.accessToken, res.data.refreshToken)
    showSuccessToast('登录成功')
    router.push('/home')
  } catch (error) {
    console.error(error)
    throw error
  } finally {
    autoLoginStatus.value = false
  }
}

onMounted(() => {
  // 检查是否是微信授权回调
  handleWechatCallback()
  
  // 检查 URL 中是否有授权码参数（从畅行温州平台跳转时自动携带）
  const code = route.query.authcode || route.query.jyauthcode
  if (typeof code === 'string' && code.trim()) {
    // 自动执行授权码登录，用户无需手动操作
    ssoLogin(code.trim()).catch((error) => {
      console.error('Auto SSO login failed:', error)
      showFailToast(error?.response?.data?.message || error?.message || '自动登录失败，请使用账号密码登录')
    })
  }
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #f7f8fa;
}

.nav-home {
  color: #1d1d1f;
}

.login-container {
  padding-top: 40px;
}

.logo-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 32px;
}

.login-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}

.app-title {
  margin-top: 16px;
  font-size: 20px;
  color: #333;
  font-weight: 600;
}

.login-form {
  margin-top: 16px;
}

.action-links {
  text-align: center;
  margin-top: 16px;
  color: #667eea;
  font-size: 14px;
  cursor: pointer;
}

.wechat-login {
  margin-top: 24px;
}

.wechat-btn-wrapper {
  padding: 0 16px;
}

.auto-login-area {
  padding-top: 60px;
}

.auto-login-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 24px;
  color: #666;
  font-size: 14px;
}

.sso-tip {
  margin-top: 32px;
  padding: 0 16px;
  text-align: center;
}

.sso-tip p {
  color: #969799;
  font-size: 13px;
  margin-top: 8px;
}
</style>
