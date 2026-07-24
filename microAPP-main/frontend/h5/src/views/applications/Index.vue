<template>
  <div class="applications-page page-container">
    <van-nav-bar title="我的申请" fixed placeholder />

    <div class="content-wrapper">
      <!-- 空状态 -->
      <van-empty
        v-if="applications.length === 0"
        description="暂无申请记录"
        image="search"
      >
        <van-button round type="primary" @click="$router.push('/home')">
          去申请服务
        </van-button>
      </van-empty>

      <!-- 申请列表 -->
      <div v-else class="application-list">
        <div
          v-for="app in applications"
          :key="app.id"
          class="application-card card"
          @click="viewDetail(app)"
        >
          <div class="card-header">
            <div class="service-info">
              <h3 class="service-name">{{ app.serviceName }}</h3>
              <van-tag :type="getStatusType(app.status)" size="small">
                {{ getStatusText(app.status) }}
              </van-tag>
            </div>
            <div class="apply-time">{{ app.applyTime }}</div>
          </div>

          <div class="card-content">
            <div class="info-row">
              <span class="label">申请编号：</span>
              <span class="value">{{ app.applyNo }}</span>
            </div>
            <div class="info-row">
              <span class="label">联系人：</span>
              <span class="value">{{ app.contactName }}</span>
            </div>
            <div class="info-row">
              <span class="label">联系电话：</span>
              <span class="value">{{ app.contactPhone }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showDialog, showFailToast } from 'vant'
import axios, { authStorage } from '@/utils/http'

// 申请数据
const applications = ref([])

const fetchData = async () => {
  try {
    const userStr = localStorage.getItem('user');
    const user = userStr ? JSON.parse(userStr) : null;
    const accessToken = authStorage.getAccessToken()
    // If no user logged in, return empty list (or handle as needed)
    if (!user || !accessToken) {
        applications.value = [];
        return;
    }

    const params = { userId: user.id };
    const res = await axios.get('/api/list', { params })
    applications.value = res.data.map(item => ({
      id: item.id,
      applyNo: item.orderNo || item.id,
      serviceName: item.serviceName || '未知服务',
      status: item.status || '待处理',
      contactName: item.contactName,
      contactPhone: item.contactPhone,
      applyTime: item.applyTime || new Date(item.createTime).toLocaleString()
    }))
  } catch (error) {
    showFailToast('获取申请记录失败')
    console.error(error)
  }
}

onMounted(() => {
  fetchData()
})

const getStatusType = (status) => {
  // 简单的状态映射
  if (status.includes('完成') || status.includes('成功')) return 'success'
  if (status.includes('处理')) return 'primary'
  if (status.includes('联系')) return 'warning'
  return 'default'
}

const getStatusText = (status) => {
  return status
}

const viewDetail = (app) => {
  showDialog({
    title: '申请详情',
    message: `
      申请编号：${app.applyNo}
      服务名称：${app.serviceName}
      联系人：${app.contactName}
      联系电话：${app.contactPhone}
      申请时间：${app.applyTime}
      状态：${app.status}
    `.trim()
  })
}
</script>

<style scoped>
.application-card {
  margin-bottom: 12px;
  cursor: pointer;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebedf0;
}

.service-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.service-name {
  font-size: 16px;
  font-weight: 500;
  color: #323233;
}

.apply-time {
  font-size: 12px;
  color: #969799;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  font-size: 13px;
}

.info-row .label {
  color: #969799;
  min-width: 80px;
}

.info-row .value {
  color: #646566;
  flex: 1;
}
</style>

