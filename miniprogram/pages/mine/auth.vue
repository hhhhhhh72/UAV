<template>
  <view class="auth-page">
    <view class="status-card">
      <view class="status-icon">!</view>
      <view class="status-text">实名认证建设中</view>
      <view class="status-tip">根据监管要求，使用低空服务需完成实名登记。该功能正在对接权威核验渠道，上线前暂时无法提交认证。</view>
    </view>

    <view class="notice-box">
      <view class="notice-title">敬请期待</view>
      <view class="notice-text">功能上线后，您可在此提交真实姓名与身份证号完成实名登记。提交的信息将受到严格保护，仅用于身份核验。</view>
    </view>
  </view>
</template>

<script setup>
// 实名认证：后端暂无实名登记接口（User 模型无身份证字段），
// 此前的前端伪造"认证成功"实现已下线，避免出现无法真实核验的"已认证"假状态。
// 待接入权威实名核验接口后再恢复提交表单。
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding: 20px 16px;
}

.status-card {
  background: #fff;
  border-radius: 8px;
  padding: 30px 20px;
  text-align: center;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.02);
}

.status-icon {
  width: 60px;
  height: 60px;
  line-height: 60px;
  border-radius: 50%;
  background: #ffe4a8;
  color: #d48806;
  font-size: 32px;
  font-weight: bold;
  margin: 0 auto 16px;
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
  line-height: 1.6;
}

.notice-box {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
}

.notice-title {
  font-size: 15px;
  font-weight: 600;
  color: #0a66c2;
  display: block;
  margin-bottom: 8px;
}

.notice-text {
  font-size: 14px;
  color: #666;
  line-height: 1.7;
}
</style>
