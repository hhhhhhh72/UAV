<template>
  <Layout :current="2">
    <view class="applications-page">
      <view v-if="applications.length === 0" class="empty-wrap">
        <view class="empty-title">暂无申请记录</view>
        <view class="empty-desc">去服务大厅提交一个申请吧</view>
        <button class="primary-btn" type="primary" @tap="goHome">去申请服务</button>
      </view>

      <view v-else class="application-list">
        <view
          v-for="app in applications"
          :key="app.id"
          class="application-card"
          @tap="viewDetail(app)"
        >
          <view class="card-header">
            <view class="service-info">
              <view class="service-name">{{ app.serviceName }}</view>
              <view class="status-tag" :class="statusClass(app.status)">{{ app.status }}</view>
            </view>
            <view class="apply-time">{{ app.applyTime }}</view>
          </view>

          <view class="card-content">
            <view class="info-row">
              <text class="label">申请编号：</text>
              <text class="value">{{ app.applyNo }}</text>
            </view>
            <view class="info-row">
              <text class="label">联系人：</text>
              <text class="value">{{ app.contactName }}</text>
            </view>
            <view class="info-row">
              <text class="label">联系电话：</text>
              <text class="value">{{ app.contactPhone }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import Layout from '@/components/Layout.vue'
import { getStoredUser, request } from '../../utils/request'

const applications = ref([])

const fetchData = async () => {
  const user = getStoredUser()
  if (!user) {
    applications.value = []
    return
  }

  try {
    const res = await request({ url: '/api/list', data: { userId: user.id } })
    const list = Array.isArray(res) ? res : (res?.data || [])
    applications.value = list.map((item) => ({
      id: item.id,
      applyNo: item.orderNo || item.id,
      serviceName: item.serviceName || '未知服务',
      status: item.status || '待处理',
      contactName: item.contactName || item.traineeName || item.name,
      contactPhone: item.contactPhone || item.traineePhone || item.phone,
      applyTime: item.applyTime || new Date(item.createTime).toLocaleString()
    }))
  } catch (error) {
    const mock = uni.getStorageSync('mock_applications') || []
    applications.value = mock.filter((a) => a.userId === user.id)
  }
}

onShow(() => {
  fetchData()
})

const goHome = () => {
  uni.switchTab({ url: '/pages/home/index' })
}

const statusClass = (status) => {
  if (status.includes('完成') || status.includes('成功')) return 'success'
  if (status.includes('处理')) return 'primary'
  if (status.includes('联系')) return 'warning'
  return 'default'
}

const viewDetail = (app) => {
  uni.showModal({
    title: '申请详情',
    content: `编号：${app.applyNo}\n服务：${app.serviceName}\n联系人：${app.contactName}\n电话：${app.contactPhone}\n状态：${app.status}`,
    showCancel: false
  })
}
</script>

<style scoped>
.applications-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 12px;
}
.empty-wrap {
  padding-top: 100px;
  text-align: center;
}
.empty-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}
.empty-desc {
  font-size: 14px;
  color: #969799;
  margin-bottom: 24px;
}
.primary-btn {
  width: 200px;
  border-radius: 999px;
  background-color: #2f7ef7;
}
.application-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f2f3f5;
}
.service-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.service-name {
  font-size: 16px;
  font-weight: 600;
}
.status-tag {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
}
.status-tag.success { background: #e8f9f0; color: #07c160; }
.status-tag.primary { background: #eef5ff; color: #2f7ef7; }
.status-tag.warning { background: #fff7e8; color: #ff976a; }
.status-tag.default { background: #f7f8fa; color: #969799; }
.apply-time { font-size: 12px; color: #969799; }
.card-content { display: flex; flex-direction: column; gap: 8px; }
.info-row { display: flex; font-size: 13px; }
.info-row .label { color: #969799; width: 80px; }
.info-row .value { color: #646566; flex: 1; }
</style>
