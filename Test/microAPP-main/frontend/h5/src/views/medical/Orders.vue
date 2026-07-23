<template>
  <div class="orders-page">
    <van-nav-bar title="我的配送订单" left-arrow @click-left="$router.back()" fixed placeholder />

    <van-tabs v-model:active="activeTab" sticky @change="onTabChange">
      <van-tab title="全部" name="" />
      <van-tab title="待接单" name="pending" />
      <van-tab title="配送中" name="delivering" />
      <van-tab title="已完成" name="completed" />
      <van-tab title="已取消" name="cancelled" />
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
            <van-tag :type="urgencyType(order.urgency)" plain size="small">{{ order.urgency_label }}</van-tag>
          </div>
          <div class="order-footer">
            <span class="time">{{ formatTime(order.created_at) }}</span>
            <div class="actions">
              <van-button v-if="order.status === 'pending'" size="small" type="warning" @click.stop="cancelOrder(order)">取消</van-button>
              <van-button v-if="order.status === 'completed' && !order.rating" size="small" type="primary" @click.stop="$router.push(`/medical/orders/${order.id}/rate`)">评价</van-button>
              <van-button v-if="['completed','cancelled'].includes(order.status)" size="small" plain @click.stop="reorder(order)">再次下单</van-button>
            </div>
          </div>
        </div>
        <van-empty v-if="!loading && !orders.length" description="暂无订单" />
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { useMedicalStore } from '@/stores/medical'

const router = useRouter()
const store = useMedicalStore()
const orders = computed(() => store.orders)

const activeTab = ref('')
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

onMounted(() => fetchOrders())

async function fetchOrders() {
  loading.value = true
  await store.fetchMyOrders({ status: activeTab.value || undefined, page: page.value, limit: 20 })
  finished.value = store.orders.length >= store.ordersPagination.total
  loading.value = false
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

function urgencyType(urgency) {
  const map = { normal: 'primary', urgent: 'warning', critical: 'danger' }
  return map[urgency] || 'primary'
}

function formatTime(time) {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function cancelOrder(order) {
  try {
    await showConfirmDialog({ message: '确定取消该订单？' })
    const res = await store.cancelOrder(order.id, { reason_type: 'not_needed' })
    if (res?.success) {
      showSuccessToast('已取消')
      fetchOrders()
    }
  } catch (e) { /* cancelled */ }
}

async function reorder(order) {
  const res = await store.getReorderData(order.id)
  if (res?.success) {
    sessionStorage.setItem('reorderData', JSON.stringify(res.data))
    router.push('/medical/order/create?reorder=1')
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
