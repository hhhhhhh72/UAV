<template>
  <div class="certification-status-page">
    <van-nav-bar title="认证状态" left-arrow @click-left="$router.back()" fixed placeholder />

    <div class="status-container" v-if="!loading">
      <!-- 未认证 -->
      <div v-if="!certification || certification.status === 'none'" class="status-card">
        <van-icon name="info-o" size="60" color="#999" />
        <p class="status-text">您尚未提交认证</p>
        <p class="status-desc">医疗配送需通过身份登记后方可使用</p>
        <van-button type="primary" round @click="$router.push('/medical/certification')">立即认证</van-button>
      </div>

      <!-- 待审核 -->
      <div v-else-if="certification.status === 'pending'" class="status-card">
        <van-icon name="clock-o" size="60" color="#faad14" />
        <p class="status-text pending">审核中</p>
        <p class="status-desc">您的认证申请已提交，预计24小时内完成审核</p>
        <div class="info-list">
          <p><span>姓名：</span>{{ certification.real_name }}</p>
          <p><span>手机：</span>{{ certification.phone }}</p>
          <p><span>机构：</span>{{ certification.org_name }}</p>
          <p><span>提交时间：</span>{{ formatTime(certification.created_at) }}</p>
        </div>
      </div>

      <!-- 已通过 -->
      <div v-else-if="certification.status === 'approved'" class="status-card">
        <van-icon name="checked" size="60" color="#52c41a" />
        <p class="status-text approved">认证已通过</p>
        <p class="status-desc">您已完成身份登记，可正常使用医疗配送服务</p>
        <div class="info-list">
          <p><span>姓名：</span>{{ certification.real_name }}</p>
          <p><span>机构：</span>{{ certification.org_name }}</p>
          <p><span>通过时间：</span>{{ formatTime(certification.review?.reviewed_at) }}</p>
        </div>
        <div class="quick-actions">
          <van-button type="primary" round @click="$router.push('/medical/order/create')">去下单</van-button>
          <van-button plain round @click="$router.push('/medical/orders')">我的寄件</van-button>
          <van-button plain round @click="$router.push('/medical/received')">寄给我的</van-button>
        </div>
      </div>

      <!-- 已驳回 -->
      <div v-else-if="certification.status === 'rejected'" class="status-card">
        <van-icon name="close" size="60" color="#ff4d4f" />
        <p class="status-text rejected">认证未通过</p>
        <p class="status-desc">驳回原因：{{ certification.review?.reject_reason || '未知' }}</p>
        <van-button type="primary" round @click="$router.push('/medical/certification')">重新提交</van-button>
      </div>
    </div>

    <div v-else class="loading-container">
      <van-loading size="36px">加载中...</van-loading>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMedicalStore } from '@/stores/medical'

const store = useMedicalStore()
const loading = ref(true)
const certification = ref(null)

onMounted(async () => {
  try {
    await store.fetchCertificationStatus()
    certification.value = store.certification
  } catch (e) {
    console.error('获取认证状态异常:', e)
    certification.value = null
  } finally {
    loading.value = false
  }
})

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style scoped>
.certification-status-page {
  min-height: 100vh;
  background: #f5f5f5;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

/* 固定导航栏居中约束 */
.certification-status-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.status-container {
  padding: 40px 20px;
}
.status-card {
  background: #fff;
  border-radius: 12px;
  padding: 40px 24px;
  text-align: center;
}
.status-text {
  font-size: 20px;
  font-weight: 600;
  margin: 16px 0 8px;
}
.status-text.pending { color: #faad14; }
.status-text.approved { color: #52c41a; }
.status-text.rejected { color: #ff4d4f; }
.status-desc {
  font-size: 14px;
  color: #666;
  margin-bottom: 24px;
}
.info-list {
  text-align: left;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}
.info-list p {
  font-size: 14px;
  color: #333;
  margin: 8px 0;
}
.info-list span {
  color: #999;
}
.quick-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}
.loading-container {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
</style>
