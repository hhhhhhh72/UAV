<template>
  <view class="page">
    <view class="profile-card">
      <image class="avatar" src="/static/logo.png" mode="aspectFit" />
      <view class="profile-info">
        <text class="name">{{ user.name || '未登录' }}</text>
        <text class="role">{{ roleLabel }}</text>
      </view>
    </view>
    <view class="menu-group">
      <view class="menu-item" @click="go('/pages/mine/profile')"><text>🏢 企业入驻</text><text class="arrow">&gt;</text></view>
      <view class="menu-item" @click="go('/pages/applications/index')"><text>📋 我的需求</text><text class="arrow">&gt;</text></view>
      <view class="menu-item" @click="showToast('证书')"><text>🎖️ 我的证书</text><text class="arrow">&gt;</text></view>
      <view class="menu-item" @click="showToast('合同')"><text>📄 我的合同</text><text class="arrow">&gt;</text></view>
    </view>
    <view class="menu-group">
      <view class="menu-item" @click="showToast('钱包')"><text>💰 钱包余额</text><text class="arrow">&gt;</text></view>
      <view class="menu-item" @click="go('/pages/mine/auth')"><text>✅ 实名认证</text><text class="arrow">&gt;</text></view>
      <view class="menu-item" @click="go('/pages/admin/index')"><text>⚙️ 管理后台</text><text class="arrow">&gt;</text></view>
    </view>
  </view>
</template>
<script>
export default {
  data() { return { user: {} } },
  computed: {
    roleLabel() { const m={platform_admin:'平台管理员',association_admin:'协会管理员',enterprise:'企业用户',individual:'个人用户'}; return m[this.user.role]||'个人用户' }
  },
  onShow() {
    try { const u=uni.getStorageSync('user'); if(u) this.user=typeof u==='string'?JSON.parse(u):u } catch(e) {}
  },
  methods: {
    go(url) { uni.navigateTo({url}) },
    showToast(msg) { uni.showToast({title:msg+' — 即将上线',icon:'none'}) }
  }
}
</script>
<style scoped>
.page{min-height:100vh;background:#f5f5f5}
.profile-card{display:flex;align-items:center;padding:24px 16px;background:linear-gradient(135deg,#1565C0,#1E88E5);gap:14px}
.avatar{width:60px;height:60px;border-radius:30px;background:#fff}
.profile-info{flex:1}
.name{color:#fff;font-size:18px;font-weight:bold;display:block}
.role{color:rgba(255,255,255,0.8);font-size:13px;margin-top:4px;display:block}
.menu-group{background:#fff;margin:12px;border-radius:12px;overflow:hidden}
.menu-item{display:flex;justify-content:space-between;align-items:center;padding:14px 16px;border-bottom:1px solid #f0f0f0;font-size:15px}
.arrow{color:#c8c9cc}
</style>
