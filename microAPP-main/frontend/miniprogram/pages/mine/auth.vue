<template>
  <view class="auth-page">
    <view class="status-card" :class="{ verified: isVerified }">
      <view class="status-icon">{{ isVerified ? '✓' : '!' }}</view>
      <view class="status-text">{{ isVerified ? '已完成实名认证' : '未完成实名认证' }}</view>
      <view class="status-tip" v-if="!isVerified">根据监管要求，使用低空服务需完成实名登记</view>
    </view>

    <view class="form-box" v-if="!isVerified">
      <view class="section-title">身份信息</view>
      <view class="input-group">
        <view class="input-item">
          <text class="label">姓名</text>
          <input class="input" v-model="form.realName" placeholder="请输入真实姓名" />
        </view>
        <view class="input-item">
          <text class="label">身份证号</text>
          <input class="input" v-model="form.idCard" type="idcard" placeholder="请输入身份证号码" />
        </view>
      </view>

      <view class="section-title">手机验证</view>
      <view class="input-group">
        <view class="input-item">
          <text class="label">手机号</text>
          <input class="input" v-model="form.phone" type="number" placeholder="请输入手机号" />
        </view>
        <view class="input-item">
          <text class="label">验证码</text>
          <input class="input" v-model="form.code" type="number" placeholder="请输入验证码" />
          <text class="code-btn" @tap="sendCode">{{ codeText }}</text>
        </view>
      </view>

      <view class="notice">
        <text class="notice-icon">ℹ️</text>
        <text class="notice-text">您的信息将受到严格保护，仅用于身份核验。</text>
      </view>

      <button class="submit-btn" type="primary" @tap="handleSubmit">提交认证</button>
    </view>

    <view class="info-box" v-else>
      <view class="info-item">
        <text class="label">真实姓名</text>
        <text class="value">{{ maskName(form.realName) }}</text>
      </view>
      <view class="info-item">
        <text class="label">身份证号</text>
        <text class="value">{{ maskIdCard(form.idCard) }}</text>
      </view>
      <view class="info-item">
        <text class="label">认证时间</text>
        <text class="value">{{ authTime }}</text>
      </view>
    </view>

  </view>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

const isVerified = ref(false)
const codeText = ref('获取验证码')
const authTime = ref('')
const form = reactive({
  realName: '',
  idCard: '',
  phone: '',
  code: ''
})

onMounted(() => {
  const user = JSON.parse(uni.getStorageSync('user') || '{}')
  if (user.isAuth) {
    isVerified.value = true
    form.realName = user.realName || '张**'
    form.idCard = user.idCard || '3303***********1234'
    authTime.value = user.authTime || '2025-12-24'
  }
})

const sendCode = () => {
  if (codeText.value !== '获取验证码') return
  if (!form.phone) return uni.showToast({ title: '请输入手机号', icon: 'none' })
  
  uni.showToast({ title: '验证码已发送' })
  let count = 60
  const timer = setInterval(() => {
    count--
    if (count <= 0) {
      clearInterval(timer)
      codeText.value = '获取验证码'
    } else {
      codeText.value = `${count}s`
    }
  }, 1000)
}

const handleSubmit = () => {
  if (!form.realName || !form.idCard || !form.phone || !form.code) {
    return uni.showToast({ title: '请填写完整信息', icon: 'none' })
  }

  uni.showLoading({ title: '认证中...' })
  
  setTimeout(() => {
    const user = JSON.parse(uni.getStorageSync('user') || '{}')
    user.isAuth = true
    user.realName = form.realName
    user.idCard = form.idCard
    user.authTime = new Date().toISOString().split('T')[0]
    
    uni.setStorageSync('user', JSON.stringify(user))
    isVerified.value = true
    authTime.value = user.authTime
    
    uni.hideLoading()
    uni.showToast({ title: '认证成功' })
  }, 1500)
}

const maskName = (name) => {
  if (!name) return ''
  return name.charAt(0) + '*'.repeat(name.length - 1)
}

const maskIdCard = (id) => {
  if (!id) return ''
  return id.substring(0, 4) + '***********' + id.substring(id.length - 4)
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding: 20px 16px;
}

.status-card {
  background: #fff;
  border-radius: 16px;
  padding: 30px 20px;
  text-align: center;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.02);
}

.status-icon {
  width: 60px;
  height: 60px;
  line-height: 60px;
  border-radius: 30px;
  background: #ff9900;
  color: #fff;
  font-size: 32px;
  font-weight: bold;
  margin: 0 auto 16px;
}

.verified .status-icon {
  background: #07c160;
}

.status-text {
  font-size: 18px;
  font-weight: bold;
  color: #323233;
  margin-bottom: 8px;
}

.status-tip {
  font-size: 14px;
  color: #969799;
}

.section-title {
  font-size: 14px;
  color: #969799;
  margin: 0 0 10px 4px;
}

.input-group {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 20px;
}

.input-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f2f3f5;
}

.input-item:last-child {
  border-bottom: none;
}

.label {
  width: 80px;
  font-size: 15px;
  color: #323233;
}

.input {
  flex: 1;
  font-size: 15px;
}

.code-btn {
  font-size: 14px;
  color: #2f7ef7;
  padding-left: 12px;
  border-left: 1px solid #f2f3f5;
}

.notice {
  display: flex;
  gap: 6px;
  padding: 0 4px;
  margin-bottom: 30px;
}

.notice-icon {
  font-size: 14px;
}

.notice-text {
  font-size: 12px;
  color: #969799;
}

.submit-btn {
  height: 48px;
  line-height: 48px;
  border-radius: 24px;
  background-color: #2f7ef7 !important;
  font-weight: bold;
}

.info-box {
  background: #fff;
  border-radius: 12px;
  padding: 8px 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f2f3f5;
}

.info-item:last-child {
  border-bottom: none;
}

.info-item .label {
  color: #969799;
}

.info-item .value {
  color: #323233;
  font-weight: 500;
}
</style>

