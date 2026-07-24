<template>
  <div class="medical-order-list">
    <DataToolbar>
      <template #filters>
        <van-dropdown-menu>
          <van-dropdown-item v-model="filters.status" :options="statusOptions" @change="fetchOrders" />
        </van-dropdown-menu>
        <van-search v-model="filters.keyword" placeholder="订单号/寄件人/收件人" show-action @search="fetchOrders" @clear="fetchOrders" style="flex:1; padding: 0;">
          <template #action>
            <span @click="fetchOrders">搜索</span>
          </template>
        </van-search>
      </template>
      <template #actions>
        <van-button type="default" size="small" icon="replay" @click="fetchOrders">刷新</van-button>
      </template>
    </DataToolbar>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <span class="stat-value">{{ stats.pending }}</span>
        <span class="stat-label">待接单</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats.delivering }}</span>
        <span class="stat-label">配送中</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats.exception }}</span>
        <span class="stat-label">异常</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats.today }}</span>
        <span class="stat-label">今日新增</span>
      </div>
    </div>

    <!-- 订单列表 -->
    <van-empty v-if="!loading && orders.length === 0" description="暂无订单数据" />

    <van-list v-model:loading="loadingMore" :finished="finished" finished-text="没有更多了" @load="loadMore">
      <van-cell-group v-for="order in orders" :key="order.id" inset style="margin-bottom: 12px; border-radius: 10px;">
        <van-cell :border="false">
          <template #title>
            <div class="order-header">
              <span class="order-no">{{ order.order_no }}</span>
              <van-tag :type="getStatusType(order.status)" size="medium">{{ getStatusLabel(order.status) }}</van-tag>
            </div>
          </template>
          <template #label>
            <div class="order-info">
              <div class="route-line">
                <span class="from">{{ order.sender_name }} · {{ order.pickup_pad_name || '取件点' }}</span>
                <span class="arrow">→</span>
                <span class="to">{{ order.receiver_name }} · {{ order.delivery_pad_name || '送达点' }}</span>
              </div>
              <div class="meta-line">
                <span>{{ order.item_type_label || order.item_type }}</span>
                <span>{{ order.urgency_label || order.urgency }}</span>
                <span>¥{{ order.estimated_fee || '-' }}</span>
                <span>{{ formatTime(order.created_at) }}</span>
              </div>
            </div>
          </template>
        </van-cell>
        <van-cell :border="false">
          <template #title>
            <div class="order-actions">
              <van-button v-if="order.status === 'pending'" type="primary" size="small" @click="handleAccept(order)">接单</van-button>
              <van-button v-if="order.status === 'accepted'" type="primary" size="small" @click="handlePickup(order)">标记取件</van-button>
              <van-button v-if="order.status === 'pickup'" type="primary" size="small" @click="handleDeliver(order)">开始配送</van-button>
              <van-button v-if="order.status === 'delivering'" type="success" size="small" @click="handleDelivered(order)">已送达</van-button>
              <van-button v-if="order.status === 'delivered'" type="success" size="small" @click="handleComplete(order)">确认完成</van-button>
              <van-button v-if="['pending','accepted','pickup','delivering'].includes(order.status)" type="danger" size="small" plain @click="handleException(order)">异常</van-button>
              <van-button size="small" plain @click="showDetail(order)">详情</van-button>
            </div>
          </template>
        </van-cell>
      </van-cell-group>
    </van-list>

    <!-- 订单详情弹窗 -->
    <van-popup v-model:show="detailVisible" position="bottom" :style="{ height: '85%' }" round>
      <div class="detail-popup" v-if="currentOrder">
        <div class="detail-header">
          <h3>订单详情</h3>
          <van-tag :type="getStatusType(currentOrder.status)" size="large">{{ getStatusLabel(currentOrder.status) }}</van-tag>
        </div>

        <van-cell-group title="基本信息" inset>
          <van-cell title="订单号" :value="currentOrder.order_no" />
          <van-cell title="紧急程度" :value="currentOrder.urgency_label || currentOrder.urgency" />
          <van-cell title="创建时间" :value="formatTime(currentOrder.created_at)" />
        </van-cell-group>

        <van-cell-group title="寄件信息" inset>
          <van-cell title="寄件人" :value="currentOrder.sender_name" />
          <van-cell title="联系电话" :value="currentOrder.sender_phone" is-link :url="'tel:' + currentOrder.sender_phone" />
          <van-cell title="取件点" :value="currentOrder.pickup_pad_name" />
        </van-cell-group>

        <van-cell-group title="收件信息" inset>
          <van-cell title="收件人" :value="currentOrder.receiver_name" />
          <van-cell title="联系电话" :value="currentOrder.receiver_phone" is-link :url="'tel:' + currentOrder.receiver_phone" />
          <van-cell title="送达点" :value="currentOrder.delivery_pad_name" />
        </van-cell-group>

        <van-cell-group title="物品信息" inset>
          <van-cell title="物品类型" :value="currentOrder.item_type_label || currentOrder.item_type" />
          <van-cell title="物品名称" :value="currentOrder.item_name || '-'" />
          <van-cell title="重量" :value="currentOrder.weight ? currentOrder.weight + 'kg' : '-'" />
          <van-cell title="温控要求" :value="currentOrder.temperature_control ? '是' : '否'" />
          <van-cell v-if="currentOrder.remark" title="备注" :value="currentOrder.remark" />
        </van-cell-group>

        <van-cell-group title="时间线" inset v-if="currentOrder.timeline && currentOrder.timeline.length">
          <van-cell v-for="(item, idx) in currentOrder.timeline" :key="idx" :title="item.action_label || item.action" :value="formatTime(item.time)" :label="item.remark || ''" />
        </van-cell-group>

        <div class="detail-actions">
          <van-button v-if="currentOrder.status === 'pending'" type="primary" block @click="handleAccept(currentOrder); detailVisible = false;">接单</van-button>
          <van-button v-if="currentOrder.status === 'accepted'" type="primary" block @click="handlePickup(currentOrder); detailVisible = false;">标记取件</van-button>
          <van-button v-if="currentOrder.status === 'pickup'" type="primary" block @click="handleDeliver(currentOrder); detailVisible = false;">开始配送</van-button>
          <van-button v-if="currentOrder.status === 'delivering'" type="success" block @click="handleDelivered(currentOrder); detailVisible = false;">已送达</van-button>
          <van-button v-if="currentOrder.status === 'delivered'" type="success" block @click="handleComplete(currentOrder); detailVisible = false;">确认完成</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 异常备注弹窗 -->
    <van-dialog v-model:show="exceptionVisible" title="标记异常" show-cancel-button @confirm="confirmException">
      <van-field v-model="exceptionRemark" type="textarea" rows="3" placeholder="请输入异常说明" style="margin: 16px;" />
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { showSuccessToast, showFailToast, showLoadingToast } from 'vant'
import axios from '@/utils/http'
import DataToolbar from '../components/DataToolbar.vue'

const statusOptions = [
  { text: '全部状态', value: '' },
  { text: '待接单', value: 'pending' },
  { text: '已接单', value: 'accepted' },
  { text: '待取件', value: 'pickup' },
  { text: '配送中', value: 'delivering' },
  { text: '已送达', value: 'delivered' },
  { text: '已完成', value: 'completed' },
  { text: '已取消', value: 'cancelled' },
  { text: '异常', value: 'exception' }
]

const statusMap = {
  pending: '待接单',
  accepted: '已接单',
  pickup: '待取件',
  delivering: '配送中',
  delivered: '已送达',
  completed: '已完成',
  cancelled: '已取消',
  exception: '异常'
}

const statusTypeMap = {
  pending: 'warning',
  accepted: 'primary',
  pickup: 'primary',
  delivering: 'primary',
  delivered: 'success',
  completed: 'success',
  cancelled: 'default',
  exception: 'danger'
}

const filters = reactive({ status: '', keyword: '' })
const orders = ref([])
const loading = ref(false)
const loadingMore = ref(false)
const finished = ref(false)
const page = ref(1)
const pageSize = 20

const stats = reactive({ pending: 0, delivering: 0, exception: 0, today: 0 })

const detailVisible = ref(false)
const currentOrder = ref(null)
const exceptionVisible = ref(false)
const exceptionRemark = ref('')
const exceptionOrderId = ref(null)

const getStatusLabel = (status) => statusMap[status] || status
const getStatusType = (status) => statusTypeMap[status] || 'default'

const formatTime = (t) => {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const fetchOrders = async () => {
  loading.value = true
  page.value = 1
  finished.value = false
  try {
    const params = { page: 1, pageSize }
    if (filters.status) params.status = filters.status
    if (filters.keyword) params.keyword = filters.keyword
    const res = await axios.get('/api/medical/orders', { params })
    if (res.data?.success) {
      orders.value = res.data.data.list || res.data.data || []
      const total = res.data.data.total || orders.value.length
      if (orders.value.length >= total) finished.value = true
      computeStats()
    }
  } catch (err) {
    showFailToast('获取订单失败')
  } finally {
    loading.value = false
  }
}

const loadMore = async () => {
  if (finished.value) return
  page.value++
  try {
    const params = { page: page.value, pageSize }
    if (filters.status) params.status = filters.status
    if (filters.keyword) params.keyword = filters.keyword
    const res = await axios.get('/api/medical/orders', { params })
    if (res.data?.success) {
      const list = res.data.data.list || res.data.data || []
      orders.value.push(...list)
      if (list.length < pageSize) finished.value = true
    }
  } catch (err) {
    // ignore
  } finally {
    loadingMore.value = false
  }
}

const computeStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  stats.pending = orders.value.filter(o => o.status === 'pending').length
  stats.delivering = orders.value.filter(o => ['accepted', 'pickup', 'delivering'].includes(o.status)).length
  stats.exception = orders.value.filter(o => o.status === 'exception').length
  stats.today = orders.value.filter(o => o.created_at && o.created_at.slice(0, 10) === today).length
}

const showDetail = (order) => {
  currentOrder.value = order
  detailVisible.value = true
}

const handleAccept = async (order) => {
  try {
    const res = await axios.post(`/api/medical/orders/${order.id}/accept`)
    if (res.data?.success) {
      showSuccessToast('接单成功')
      fetchOrders()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

const handlePickup = async (order) => {
  try {
    const res = await axios.post(`/api/medical/orders/${order.id}/pickup`)
    if (res.data?.success) {
      showSuccessToast('已标记取件')
      fetchOrders()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

const handleDeliver = async (order) => {
  try {
    const res = await axios.post(`/api/medical/orders/${order.id}/deliver`)
    if (res.data?.success) {
      showSuccessToast('开始配送')
      fetchOrders()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

const handleDelivered = async (order) => {
  try {
    const res = await axios.post(`/api/medical/orders/${order.id}/delivered`)
    if (res.data?.success) {
      showSuccessToast('已送达')
      fetchOrders()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

const handleComplete = async (order) => {
  try {
    const res = await axios.post(`/api/medical/orders/${order.id}/complete`)
    if (res.data?.success) {
      showSuccessToast('已完成')
      fetchOrders()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

const handleException = (order) => {
  exceptionOrderId.value = order.id
  exceptionRemark.value = ''
  exceptionVisible.value = true
}

const confirmException = async () => {
  try {
    const res = await axios.post(`/api/medical/orders/${exceptionOrderId.value}/exception`, { remark: exceptionRemark.value })
    if (res.data?.success) {
      showSuccessToast('已标记异常')
      fetchOrders()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

onMounted(fetchOrders)
</script>

<style scoped>
.medical-order-list {
  padding-bottom: 20px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 14px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.stat-value {
  display: block;
  font-size: 24px;
  font-weight: 700;
  color: var(--accent-color, #0071e3);
}

.stat-label {
  display: block;
  font-size: 12px;
  color: var(--text-secondary, #86868b);
  margin-top: 4px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-no {
  font-weight: 600;
  font-size: 14px;
}

.order-info {
  margin-top: 6px;
}

.route-line {
  font-size: 13px;
  margin-bottom: 4px;
}

.route-line .arrow {
  margin: 0 6px;
  color: var(--accent-color, #0071e3);
}

.meta-line {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: var(--text-secondary, #86868b);
}

.order-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.detail-popup {
  padding: 20px;
  overflow-y: auto;
  height: 100%;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.detail-header h3 {
  margin: 0;
  font-size: 18px;
}

.detail-actions {
  margin-top: 20px;
  padding-bottom: 30px;
}

.detail-actions .van-button {
  margin-bottom: 10px;
}

@media (max-width: 767px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
