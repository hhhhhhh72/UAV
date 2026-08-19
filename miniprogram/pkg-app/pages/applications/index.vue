<template>
  <view class="applications-page">
    <u-nav-bar title="我的业务" show-back @back="goBack" />

    <!-- Tabs -->
    <u-sticky>
      <u-tabs v-model:active="activeTab" :titles="tabTitles" @change="onTabChange" />
    </u-sticky>

    <!-- 未登录引导 -->
    <view v-if="showLogin" class="login-guide">
      <text class="login-guide-title">登录后查看我的业务</text>
      <text class="login-guide-desc">我的需求、合同、订单一览</text>
      <view class="login-guide-btn" hover-class="login-guide-btn-hover" @tap="goLogin">
        <text>去登录</text>
      </view>
    </view>

    <template v-else>
    <view v-if="activeTab === 0" class="tab-content">
      <view v-if="errorMsg && demands.length === 0" class="tab-error">
        <text class="tab-error-text">{{ errorMsg }}</text>
        <view class="tab-error-btn" hover-class="login-guide-btn-hover" @tap="fetchDemands">重试</view>
      </view>
      <u-empty v-else-if="!loadingDemands && demands.length === 0" description="暂无需求" />
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
      <view v-if="errorMsg && contracts.length === 0" class="tab-error">
        <text class="tab-error-text">{{ errorMsg }}</text>
        <view class="tab-error-btn" hover-class="login-guide-btn-hover" @tap="fetchContracts">重试</view>
      </view>
      <u-empty v-else-if="!loadingContracts && contracts.length === 0" description="暂无合同" />
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
      <view v-if="errorMsg && orders.length === 0" class="tab-error">
        <text class="tab-error-text">{{ errorMsg }}</text>
        <view class="tab-error-btn" hover-class="login-guide-btn-hover" @tap="fetchOrders">重试</view>
      </view>
      <u-empty v-else-if="!loadingOrders && orders.length === 0" description="暂无订单" />
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

    <!-- 有旧数据时加载失败横幅（保留旧数据，可重试/下拉刷新） -->
    <view v-if="errorMsg && activeList.length > 0" class="tab-error-banner">
      <text>{{ errorMsg }}</text>
      <text class="tab-error-banner-retry" @tap="retry">重试</text>
    </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../../utils/request'

const activeTab = ref(0)

const tabTitles = ['我的需求', '我的合同', '我的订单']

const demands = ref([])
const contracts = ref([])
const orders = ref([])

const loadingDemands = ref(false)
const loadingContracts = ref(false)
const loadingOrders = ref(false)

// P1 修复：未登录引导 + 加载失败提示
const showLogin = ref(false)
const errorMsg = ref('')

const fetchDemands = async () => {
  const user = getStoredUser()
  if (!user) {
    showLogin.value = true
    errorMsg.value = ''
    demands.value = []
    return
  }
  showLogin.value = false
  loadingDemands.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/demands', data: { mine: 1 } })
    const list = res?.data || res || []
    demands.value = Array.isArray(list) ? list : []
  } catch (e) {
    // 失败保留旧数据：空列表展示错误态+重试；有旧数据时 toast 提示不清空
    errorMsg.value = '加载失败，请稍后重试'
    if (demands.value.length > 0) {
      uni.showToast({ title: '加载失败，请下拉重试', icon: 'none' })
    }
  } finally {
    loadingDemands.value = false
  }
}

const fetchContracts = async () => {
  const user = getStoredUser()
  if (!user) {
    showLogin.value = true
    errorMsg.value = ''
    contracts.value = []
    return
  }
  showLogin.value = false
  loadingContracts.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/contracts' })
    const list = res?.data || res || []
    contracts.value = Array.isArray(list) ? list : []
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
    if (contracts.value.length > 0) {
      uni.showToast({ title: '加载失败，请下拉重试', icon: 'none' })
    }
  } finally {
    loadingContracts.value = false
  }
}

const fetchOrders = async () => {
  const user = getStoredUser()
  if (!user) {
    showLogin.value = true
    errorMsg.value = ''
    orders.value = []
    return
  }
  showLogin.value = false
  loadingOrders.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/trade-orders/mine' })
    const list = res?.data || res || []
    orders.value = Array.isArray(list) ? list : []
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
    if (orders.value.length > 0) {
      uni.showToast({ title: '加载失败，请下拉重试', icon: 'none' })
    }
  } finally {
    loadingOrders.value = false
  }
}

// 当前 Tab 的列表（错误横幅判断是否有旧数据可保留）
const activeList = computed(() => {
  if (activeTab.value === 0) return demands.value
  if (activeTab.value === 1) return contracts.value
  return orders.value
})

const retry = () => {
  if (activeTab.value === 0) fetchDemands()
  else if (activeTab.value === 1) fetchContracts()
  else fetchOrders()
}

const goLogin = () => {
  uni.navigateTo({ url: '/pages/login/index' })
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

/* P1 修复：未登录引导 + 加载失败态 */
.login-guide {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80px 24px 40px;
}
.login-guide-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text);
}
.login-guide-desc {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
}
.login-guide-btn {
  margin-top: 20px;
  padding: 10px 44px;
  border-radius: 50rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}
.login-guide-btn-hover {
  opacity: 0.85;
}
.tab-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60px 24px 30px;
}
.tab-error-text {
  font-size: 14px;
  color: var(--color-text-secondary);
}
.tab-error-btn {
  margin-top: 16px;
  padding: 8px 32px;
  border-radius: 50rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 13px;
}
.tab-error-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin: 0 12px 12px;
  padding: 10px 16px;
  border-radius: 12rpx;
  background: #FEF0EF;
  font-size: 13px;
  color: #B42318;
}
.tab-error-banner-retry {
  color: var(--color-primary);
  font-weight: 600;
}
</style>
