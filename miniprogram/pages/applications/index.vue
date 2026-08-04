<template>
  <view class="applications-page">
    <u-nav-bar title="我的业务" show-back @back="goBack" />

    <!-- Tabs -->
    <u-sticky>
      <u-tabs v-model:active="activeTab" :titles="tabTitles" @change="onTabChange" />
    </u-sticky>

    <view v-if="activeTab === 0" class="tab-content">
      <u-empty v-if="!loadingDemands && demands.length === 0" description="暂无需求" />
      <u-cell-group v-else inset>
        <u-cell
          v-for="item in demands"
          :key="item.id"
          :label="item.description || item.content || ''"
          is-link
          @click="viewDetail('demand', item)"
        >
          <template #title>
            <view class="cell-title-row">
              <text class="cell-title-text">{{ item.title || item.serviceName || '需求' + item.id }}</text>
              <u-tag :type="getStatusType(item.status)" size="mini">{{ getStatusText(item.status) }}</u-tag>
            </view>
          </template>
        </u-cell>
      </u-cell-group>
      <view v-if="loadingDemands" class="tab-loading"><u-loading size="24rpx" /><text>加载中...</text></view>
    </view>

    <view v-else-if="activeTab === 1" class="tab-content">
      <u-empty v-if="!loadingContracts && contracts.length === 0" description="暂无合同" />
      <u-cell-group v-else inset>
        <u-cell
          v-for="item in contracts"
          :key="item.id"
          :label="item.partyB || item.counterparty || ''"
          is-link
          @click="viewDetail('contract', item)"
        >
          <template #title>
            <view class="cell-title-row">
              <text class="cell-title-text">{{ item.title || item.contractNo || '合同' + item.id }}</text>
              <u-tag :type="getStatusType(item.status)" size="mini">{{ getStatusText(item.status) }}</u-tag>
            </view>
          </template>
        </u-cell>
      </u-cell-group>
      <view v-if="loadingContracts" class="tab-loading"><u-loading size="24rpx" /><text>加载中...</text></view>
    </view>

    <view v-else-if="activeTab === 2" class="tab-content">
      <u-empty v-if="!loadingOrders && orders.length === 0" description="暂无订单" />
      <u-cell-group v-else inset>
        <u-cell
          v-for="item in orders"
          :key="item.id"
          :label="'金额: ' + (item.amount || item.price || '--')"
          is-link
          @click="viewDetail('order', item)"
        >
          <template #title>
            <view class="cell-title-row">
              <text class="cell-title-text">{{ item.productName || item.title || '订单' + item.id }}</text>
              <u-tag :type="getStatusType(item.status)" size="mini">{{ getStatusText(item.status) }}</u-tag>
            </view>
          </template>
        </u-cell>
      </u-cell-group>
      <view v-if="loadingOrders" class="tab-loading"><u-loading size="24rpx" /><text>加载中...</text></view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../utils/request'

const activeTab = ref(0)

const tabTitles = ['我的需求', '我的合同', '我的订单']

const demands = ref([])
const contracts = ref([])
const orders = ref([])

const loadingDemands = ref(false)
const loadingContracts = ref(false)
const loadingOrders = ref(false)

const fetchDemands = async () => {
  const user = getStoredUser()
  if (!user) { demands.value = []; return }
  loadingDemands.value = true
  try {
    const res = await request({ url: '/api/v1/demands', data: { mine: 1 } })
    const list = res?.data || res || []
    demands.value = Array.isArray(list) ? list : []
  } catch (e) {
    // silent fail
    demands.value = []
  } finally {
    loadingDemands.value = false
  }
}

const fetchContracts = async () => {
  const user = getStoredUser()
  if (!user) { contracts.value = []; return }
  loadingContracts.value = true
  try {
    const res = await request({ url: '/api/v1/contracts' })
    const list = res?.data || res || []
    contracts.value = Array.isArray(list) ? list : []
  } catch (e) {
    // silent fail
    contracts.value = []
  } finally {
    loadingContracts.value = false
  }
}

const fetchOrders = async () => {
  const user = getStoredUser()
  if (!user) { orders.value = []; return }
  loadingOrders.value = true
  try {
    const res = await request({ url: '/api/v1/trade-orders/mine' })
    const list = res?.data || res || []
    orders.value = Array.isArray(list) ? list : []
  } catch (e) {
    // silent fail
    orders.value = []
  } finally {
    loadingOrders.value = false
  }
}

const onTabChange = (index) => {
  if (index === 0) fetchDemands()
  else if (index === 1) fetchContracts()
  else if (index === 2) fetchOrders()
}

onShow(() => {
  // Load active tab data on show
  if (activeTab.value === 0) fetchDemands()
  else if (activeTab.value === 1) fetchContracts()
  else if (activeTab.value === 2) fetchOrders()
})

const getStatusType = (status) => {
  if (!status) return 'default'
  if (status === 'pending' || status === '待处理') return 'warning'
  if (status === 'processing' || status === '处理中') return 'primary'
  if (status === 'completed' || status === '已完成' || status === 'done') return 'success'
  if (status === 'cancelled' || status === '已取消' || status === 'rejected' || status === '已拒绝') return 'danger'
  return 'default'
}

const getStatusText = (status) => {
  if (!status) return '未知'
  const map = {
    pending: '待处理',
    processing: '处理中',
    completed: '已完成',
    done: '已完成',
    cancelled: '已取消',
    rejected: '已拒绝',
    active: '进行中',
  }
  return map[status] || status
}

const viewDetail = (type, item) => {
  const titleMap = { demand: '需求详情', contract: '合同详情', order: '订单详情' }
  const lines = []
  if (item.title || item.serviceName) lines.push('标题：' + (item.title || item.serviceName))
  if (item.description || item.content) lines.push('描述：' + (item.description || item.content))
  if (item.partyB || item.counterparty) lines.push('对方：' + (item.partyB || item.counterparty))
  if (item.amount != null || item.price != null) lines.push('金额：' + (item.amount || item.price))
  if (item.status) lines.push('状态：' + getStatusText(item.status))
  if (item.created_at) lines.push('时间：' + String(item.created_at).slice(0, 10))
  uni.showModal({
    title: titleMap[type] || '详情',
    content: lines.length ? lines.join('\n') : '暂无更多信息',
    showCancel: false,
    confirmText: '知道了',
  })
}

const goBack = () => {
  uni.navigateBack()
}
</script>

<style scoped>
.applications-page {
  background: var(--color-bg);
  min-height: 100vh;
}

.tab-content {
  padding: 12px 0;
  min-height: 200px;
}

.tab-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.cell-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cell-title-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
}
</style>
