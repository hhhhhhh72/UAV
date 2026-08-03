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
      <u-empty v-if="!loadingBids && bids.length === 0" description="暂无竞标记录" />
      <u-cell-group v-else inset>
        <u-cell
          v-for="item in bids"
          :key="item.id"
          :label="'报价: ' + (item.price || item.amount || '--')"
          is-link
          @click="viewDetail('bid', item)"
        >
          <template #title>
            <view class="cell-title-row">
              <text class="cell-title-text">{{ item.demandTitle || item.title || '竞标' + item.id }}</text>
              <u-tag :type="getStatusType(item.status)" size="mini">{{ getStatusText(item.status) }}</u-tag>
            </view>
          </template>
        </u-cell>
      </u-cell-group>
      <view v-if="loadingBids" class="tab-loading"><u-loading size="24rpx" /><text>加载中...</text></view>
    </view>

    <view v-else-if="activeTab === 2" class="tab-content">
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

    <view v-else class="tab-content">
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

const tabTitles = ['我的需求', '我的竞标', '我的合同', '我的订单']

const demands = ref([])
const bids = ref([])
const contracts = ref([])
const orders = ref([])

const loadingDemands = ref(false)
const loadingBids = ref(false)
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

const fetchBids = async () => {
  const user = getStoredUser()
  if (!user) { bids.value = []; return }
  loadingBids.value = true
  try {
    const res = await request({ url: '/api/v1/demands/bids/mine' })
    const list = res?.data || res || []
    bids.value = Array.isArray(list) ? list : []
  } catch (e) {
    // silent fail
    bids.value = []
  } finally {
    loadingBids.value = false
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
  else if (index === 1) fetchBids()
  else if (index === 2) fetchContracts()
  else if (index === 3) fetchOrders()
}

onShow(() => {
  // Load active tab data on show
  if (activeTab.value === 0) fetchDemands()
  else if (activeTab.value === 1) fetchBids()
  else if (activeTab.value === 2) fetchContracts()
  else if (activeTab.value === 3) fetchOrders()
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
  uni.showToast({ title: '详情 - 即将上线', icon: 'none', duration: 1500 })
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
