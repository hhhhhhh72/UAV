<template>
  <view class="applications-page">
    <van-nav-bar title="我的业务" fixed placeholder left-arrow @click-left="goBack" />

    <!-- Tabs -->
    <van-tabs v-model:active="activeTab" @change="onTabChange" sticky>
        <van-tab title="我的需求">
          <view class="tab-content">
            <van-empty v-if="!loadingDemands && demands.length === 0" description="暂无需求" image="search" />
            <van-cell-group v-else inset>
              <van-cell
                v-for="item in demands"
                :key="item.id"
                :title="item.title || item.serviceName || '需求' + item.id"
                :label="item.description || item.content || ''"
                :value="getStatusText(item.status)"
                is-link
                @tap="viewDetail('demand', item)"
              >
                <template #title>
                  <view class="cell-title-row">
                    <text class="cell-title-text">{{ item.title || item.serviceName || '需求' + item.id }}</text>
                    <van-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</van-tag>
                  </view>
                </template>
              </van-cell>
            </van-cell-group>
            <van-loading v-if="loadingDemands" size="24" class="tab-loading">加载中...</van-loading>
          </view>
        </van-tab>

        <van-tab title="我的竞标">
          <view class="tab-content">
            <van-empty v-if="!loadingBids && bids.length === 0" description="暂无竞标记录" image="search" />
            <van-cell-group v-else inset>
              <van-cell
                v-for="item in bids"
                :key="item.id"
                :title="item.demandTitle || item.title || '竞标' + item.id"
                :label="'报价: ' + (item.price || item.amount || '--')"
                :value="getStatusText(item.status)"
                is-link
                @tap="viewDetail('bid', item)"
              >
                <template #title>
                  <view class="cell-title-row">
                    <text class="cell-title-text">{{ item.demandTitle || item.title || '竞标' + item.id }}</text>
                    <van-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</van-tag>
                  </view>
                </template>
              </van-cell>
            </van-cell-group>
            <van-loading v-if="loadingBids" size="24" class="tab-loading">加载中...</van-loading>
          </view>
        </van-tab>

        <van-tab title="我的合同">
          <view class="tab-content">
            <van-empty v-if="!loadingContracts && contracts.length === 0" description="暂无合同" image="search" />
            <van-cell-group v-else inset>
              <van-cell
                v-for="item in contracts"
                :key="item.id"
                :title="item.title || item.contractNo || '合同' + item.id"
                :label="item.partyB || item.counterparty || ''"
                :value="getStatusText(item.status)"
                is-link
                @tap="viewDetail('contract', item)"
              >
                <template #title>
                  <view class="cell-title-row">
                    <text class="cell-title-text">{{ item.title || item.contractNo || '合同' + item.id }}</text>
                    <van-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</van-tag>
                  </view>
                </template>
              </van-cell>
            </van-cell-group>
            <van-loading v-if="loadingContracts" size="24" class="tab-loading">加载中...</van-loading>
          </view>
        </van-tab>

        <van-tab title="我的订单">
          <view class="tab-content">
            <van-empty v-if="!loadingOrders && orders.length === 0" description="暂无订单" image="search" />
            <van-cell-group v-else inset>
              <van-cell
                v-for="item in orders"
                :key="item.id"
                :title="item.productName || item.title || '订单' + item.id"
                :label="'金额: ' + (item.amount || item.price || '--')"
                :value="getStatusText(item.status)"
                is-link
                @tap="viewDetail('order', item)"
              >
                <template #title>
                  <view class="cell-title-row">
                    <text class="cell-title-text">{{ item.productName || item.title || '订单' + item.id }}</text>
                    <van-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</van-tag>
                  </view>
                </template>
              </van-cell>
            </van-cell-group>
            <van-loading v-if="loadingOrders" size="24" class="tab-loading">加载中...</van-loading>
          </view>
        </van-tab>
      </van-tabs>
    </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../utils/request'

const activeTab = ref(0)

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
    console.warn('Failed to load demands:', e)
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
    console.warn('Failed to load bids:', e)
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
    console.warn('Failed to load contracts:', e)
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
    console.warn('Failed to load orders:', e)
    orders.value = []
  } finally {
    loadingOrders.value = false
  }
}

const onTabChange = (e) => {
  const index = e.detail ? e.detail.index : e
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
  background: #f7f8fa;
  min-height: 100vh;
}

.tab-content {
  padding: 12px 0;
  min-height: 200px;
}

.tab-loading {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.cell-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cell-title-text {
  font-size: 14px;
  font-weight: 500;
  color: #323233;
}
</style>
