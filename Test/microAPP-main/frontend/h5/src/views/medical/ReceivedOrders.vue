<template>
  <div class="orders-page">
    <van-nav-bar title="寄给我的" left-arrow @click-left="$router.back()" fixed placeholder />

    <van-tabs v-model:active="activeTab" sticky @change="onTabChange">
      <van-tab title="全部" name="" />
      <van-tab title="配送中" name="delivering" />
      <van-tab title="待签收" name="delivered" />
      <van-tab title="已签收" name="completed" />
    </van-tabs>

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list v-model:loading="loading" :finished="finished" @load="loadMore" finished-text="没有更多了">
        <div class="order-card" v-for="order in orders" :key="order.id" @click="goDetail(order.id)">
          <div class="order-header">
            <span class="order-no">{{ order.order_no }}</span>
            <van-tag :type="statusTagType(order.status)" size="medium">{{ order.status_label }}</van-tag>
          </div>
          <div class="order-route">
            <span>{{ order.route.departure_name }}</span>
            <van-icon name="arrow" />
            <span>{{ order.route.arrival_name }}</span>
          </div>
          <div class="order-info">
            <span>{{ order.item.type_label }} | {{ order.item.weight }}kg</span>
            <span class="sender-info">寄件人: {{ order.sender.name }}</span>
          </div>
          <div class="order-footer">
            <span class="time">{{ formatTime(order.created_at) }}</span>
            <div class="actions">
              <van-button v-if="order.status === 'delivered'" size="small" type="success" @click.stop="confirmReceipt(order)">确认签收</van-button>
            </div>
          </div>
        </div>
        <van-empty v-if="!loading && !orders.length" description="暂无收件订单" />
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { useMedicalStore } from '@/stores/medical'

const router = useRouter()
const store = useMedicalStore()

const orders = ref([])
const activeTab = ref('')
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

onMounted(() => fetchOrders())

async function fetchOrders() {
  loading.value = true
  try {
    const res = await store.fetchReceivedOrders({ status: activeTab.value || undefined, page: page.value, limit: 20 })
    if (res?.success) {
      orders.value = res.data
      finished.value = res.data.length >= res.total
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function onTabChange() {
  page.value = 1
  finished.value = false
  fetchOrders()
}

async function onRefresh() {
  page.value = 1
  finished.value = false
  await fetchOrders()
  refreshing.value = false
}

function loadMore() {
  if (finished.value) return
  page.value++
  fetchOrders()
}

function goDetail(id) {
  router.push(`/medical/orders/${id}`)
}

function statusTagType(status) {
  const map = { pending: 'warning', accepted: 'primary', pickup: 'primary', delivering: 'primary', delivered: 'success', completed: 'success', cancelled: 'default', exception: 'danger' }
  return map[status] || 'default'
}

function formatTime(time) {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function confirmReceipt(order) {
  try {
    await showConfirmDialog({ title: '确认签收', message: `确认已收到订单 ${order.order_no} 的配送物品？` })
    const res = await store.confirmReceipt(order.id)
    if (res?.success) {
      showSuccessToast('签收成功')
      fetchOrders()
    } else {
      showFailToast(res?.message || '签收失败')
    }
  } catch (e) {
    if (e !== 'cancel') showFailToast(e.response?.data?.message || '签收失败')
  }
}
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: #f5f5f5;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

/* 固定导航栏居中约束 */
.orders-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

/* Sticky tabs 居中约束 */
.orders-page :deep(.van-sticky--fixed) {
  left: 50% !important;
  transform: translateX(-50%);
  width: 100% !important;
  max-width: var(--page-max-width);
}

.order-card {
  background: #fff;
  margin: 12px;
  border-radius: 10px;
  padding: 16px;
}
.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.order-no {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}
.order-route {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #555;
  margin-bottom: 8px;
}
.order-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #888;
  margin-bottom: 10px;
}
.sender-info {
  color: #666;
}
.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #f0f0f0;
  padding-top: 10px;
}
.order-footer .time {
  font-size: 12px;
  color: #999;
}
.actions {
  display: flex;
  gap: 8px;
}
</style>
