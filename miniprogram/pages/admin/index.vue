<template>
  <view class="admin-page">
    <view class="tab-header">
      <view 
        v-for="(tab, index) in tabs" 
        :key="index" 
        class="tab-item" 
        :class="{ active: activeTab === index }"
        @tap="activeTab = index; fetchTabData(index)"
      >
        {{ tab }}
      </view>
    </view>

    <view class="content-body">
      <!-- 订单管理 -->
      <view v-if="activeTab === 0" class="data-list">
        <view v-if="orders.length === 0" class="empty">暂无申请数据</view>
        <view 
          v-for="item in orders" 
          :key="item.id || item.orderNo" 
          class="card"
          @tap="showOrderDetail(item)"
        >
          <view class="card-title">
            <text class="name">{{ item.serviceName }}</text>
            <text class="status" :class="getStatusClass(item.status)">{{ item.status || '待处理' }}</text>
          </view>
          <view class="card-info">
            <view class="info-line">申请人：{{ item.contactName || item.traineeName || item.name || '匿名' }}</view>
            <view class="info-line">电话：{{ item.contactPhone || item.traineePhone || item.phone }}</view>
            <view class="info-line date">时间：{{ item.applyTime || '未知' }}</view>
          </view>
        </view>
      </view>

      <!-- 企业审核 -->
      <view v-if="activeTab === 1" class="data-list">
        <!-- 统计卡片 -->
        <view class="stats-row">
          <view class="stat-card"><text class="stat-num">{{ enterpriseStats.total }}</text><text class="stat-label">总申请</text></view>
          <view class="stat-card"><text class="stat-num orange">{{ enterpriseStats.pending }}</text><text class="stat-label">待审核</text></view>
          <view class="stat-card"><text class="stat-num green">{{ enterpriseStats.approved }}</text><text class="stat-label">已通过</text></view>
          <view class="stat-card"><text class="stat-num red">{{ enterpriseStats.rejected }}</text><text class="stat-label">已驳回</text></view>
        </view>

        <!-- 状态筛选 -->
        <view class="filter-bar">
          <view
            v-for="opt in statusFilterOptions" :key="opt.value"
            class="filter-item" :class="{ active: selectedStatus === opt.value }"
            @tap="selectedStatus = opt.value"
          >{{ opt.label }}</view>
        </view>

        <view v-if="filteredEnterprise.length === 0" class="empty">暂无企业入驻申请</view>
        <view
          v-for="item in filteredEnterprise"
          :key="item.id"
          class="card"
          @tap="showEnterpriseDetail(item)"
        >
          <view class="card-title">
            <text class="name">{{ item.companyName || '未知企业' }}</text>
            <text class="status" :class="getEnterpriseStatusClass(item.status)">{{ item.statusText || item.status || '待审核' }}</text>
          </view>
          <view class="card-info">
            <view class="info-line">联系人：{{ item.contactName || '-' }}</view>
            <view class="info-line">电话：{{ item.contactPhone || '-' }}</view>
            <view class="info-line date">申请时间：{{ item.applyTime || '未知' }}</view>
          </view>
        </view>
      </view>

      <!-- 用户管理 -->
      <view v-if="activeTab === 2" class="data-list">
        <view v-if="users.length === 0" class="empty">暂无用户数据</view>
        <view 
          v-for="u in users" 
          :key="u.id" 
          class="card user-card"
        >
          <view class="user-main">
            <view class="user-avatar">👤</view>
            <view class="user-detail">
              <view class="user-name">
                {{ u.name || '未命名' }}
                <text class="role-tag" :class="u.role">{{ getRoleLabel(u.role) }}</text>
              </view>
              <view class="user-phone">{{ u.phone }}</view>
            </view>
          </view>
          <button 
            class="action-btn" 
            size="mini" 
            :type="u.role === 'admin' || u.role === 'dsl_admin' ? 'default' : 'primary'"
            @tap="toggleRole(u)"
          >
            {{ u.role === 'admin' || u.role === 'dsl_admin' ? '设为用户' : '设为管理' }}
          </button>
        </view>
      </view>
    </view>

    <!-- 订单操作弹窗 -->
    <view class="modal" v-if="showModal" @tap="showModal = false">
      <view class="modal-content" @tap.stop>
        <view class="modal-title">订单详情</view>
        <view class="detail-rows" v-if="currentOrder">
          <view class="detail-row"><text class="dlabel">服务：</text><text>{{ currentOrder.serviceName }}</text></view>
          <view class="detail-row"><text class="dlabel">申请人：</text><text>{{ currentOrder.contactName || currentOrder.traineeName || currentOrder.name || '匿名' }}</text></view>
          <view class="detail-row"><text class="dlabel">电话：</text><text>{{ currentOrder.contactPhone || currentOrder.traineePhone || currentOrder.phone || '-' }}</text></view>
          <view class="detail-row"><text class="dlabel">时间：</text><text>{{ currentOrder.applyTime || '-' }}</text></view>
          <view class="detail-row"><text class="dlabel">状态：</text><text>{{ currentOrder.status || '待处理' }}</text></view>
          <view class="detail-row" v-if="currentOrder.competitionRoleText"><text class="dlabel">角色：</text><text>{{ currentOrder.competitionRoleText }}</text></view>
          <view class="detail-row" v-if="currentOrder.companyName"><text class="dlabel">单位：</text><text>{{ currentOrder.companyName }}</text></view>
          <view class="detail-row" v-if="currentOrder.remark"><text class="dlabel">备注：</text><text>{{ currentOrder.remark }}</text></view>
        </view>
        <view class="modal-title" style="margin-top: 16px;">修改状态</view>
        <view class="status-options">
          <view 
            v-for="opt in statusOptions" 
            :key="opt" 
            class="opt-item"
            @tap="updateStatus(opt)"
          >
            {{ opt }}
          </view>
        </view>
        <button class="close-btn" @tap="showModal = false">取消</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { request, getStoredUser } from '../../utils/request'

const tabs = ['订单管理', '企业审核', '用户管理']
const activeTab = ref(0)
const orders = ref([])
const users = ref([])
const showModal = ref(false)
const currentOrder = ref(null)
const statusOptions = ['待处理', '处理中', '已联系', '已完成', '已取消']
// enterprise filter state

const statusFilterOptions = [
  { label: '全部', value: 'all' },
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已驳回', value: 'rejected' }
]

const getRoleLabel = (role) => {
  if (role === 'admin') return '管理员'
  if (role === 'dsl_admin') return '平台管理'
  return '用户'
}

const enterpriseList = ref([])
const selectedStatus = ref('all')

const enterpriseStats = computed(() => {
  const stats = { total: 0, pending: 0, approved: 0, rejected: 0 }
  enterpriseList.value.forEach(item => {
    stats.total++
    if (item.status === 'pending') stats.pending++
    else if (item.status === 'approved') stats.approved++
    else if (item.status === 'rejected') stats.rejected++
  })
  return stats
})

const filteredEnterprise = computed(() => {
  if (selectedStatus.value === 'all') return enterpriseList.value
  return enterpriseList.value.filter(i => i.status === selectedStatus.value)
})

const getEnterpriseStatusClass = (status) => {
  if (status === 'approved') return 'success'
  if (status === 'pending') return 'warning'
  if (status === 'rejected') return 'default'
  return 'default'
}

const showEnterpriseDetail = (item) => {
  currentOrder.value = item
  showModal.value = true
}

const fetchOrders = async () => {
  try {
    const res = await request({ url: '/api/list', data: { role: 'admin' } })
    const list = Array.isArray(res) ? res : (res?.data || [])
    orders.value = list
  } catch (e) {
    const all = uni.getStorageSync('mock_applications') || []
    orders.value = all
  }

  // 加载企业入驻申请
  try {
    const res = await request({ url: '/api/enterprise/list' })
    enterpriseList.value = Array.isArray(res) ? res : (res?.data || [])
  } catch (e) {
    enterpriseList.value = []
  }
}

const fetchUsers = async () => {
  try {
    const res = await request({ url: '/api/users' })
    users.value = Array.isArray(res) ? res : (res?.data || [])
  } catch (e) {
    users.value = [
      { id: 1, name: '管理员', phone: '13800000000', role: 'admin' },
      { id: 2, name: '张三', phone: '13911112222', role: 'user' },
      { id: 3, name: '李四', phone: '13733334444', role: 'user' }
    ]
  }
}

const fetchTabData = (idx) => {
  if (idx === 0 || idx === 1) fetchOrders()
  if (idx === 2) fetchUsers()
}

onMounted(() => {
  fetchOrders()
})

const getStatusClass = (status) => {
  if (!status) return 'default'
  if (status.includes('完成') || status.includes('成功')) return 'success'
  if (status.includes('处理')) return 'warning'
  if (status.includes('联系')) return 'primary'
  return 'default'
}

const showOrderDetail = (item) => {
  currentOrder.value = item
  showModal.value = true
}

const updateStatus = async (newStatus) => {
  if (!currentOrder.value) return

  try {
    await request({
      url: '/api/update',
      method: 'POST',
      data: { id: currentOrder.value.id || currentOrder.value.orderNo, status: newStatus }
    })
  } catch (e) { /* fallback to local */ }

  currentOrder.value.status = newStatus
  
  const all = uni.getStorageSync('mock_applications') || []
  const idx = all.findIndex(a => a.orderNo === currentOrder.value.orderNo)
  if (idx > -1) {
    all[idx].status = newStatus
    uni.setStorageSync('mock_applications', all)
  }
  
  showModal.value = false
  uni.showToast({ title: '修改成功' })
}

const toggleRole = async (u) => {
  const newRole = (u.role === 'admin' || u.role === 'dsl_admin') ? 'user' : 'admin'
  try {
    await request({
      url: '/api/user/role',
      method: 'POST',
      data: { userId: u.id, role: newRole }
    })
  } catch (e) { /* fallback */ }
  u.role = newRole
  uni.showToast({ title: '权限已更新' })
}
</script>

<style scoped>
.admin-page { min-height: 100vh; background: #f7f8fa; }
.tab-header { display: flex; background: #fff; position: sticky; top: 0; z-index: 10; border-bottom: 1px solid #eee; }
.tab-item { flex: 1; text-align: center; padding: 14px 0; font-size: 14px; color: #646566; position: relative; }
.tab-item.active { color: #667eea; font-weight: bold; }
.tab-item.active::after { content: ''; position: absolute; bottom: 0; left: 30%; right: 30%; height: 3px; background: #667eea; border-radius: 2px; }

.content-body { padding: 12px; }
.empty { text-align: center; padding: 100px 0; color: #969799; font-size: 14px; }

.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; margin-bottom: 12px; }
.stat-card { background: #fff; border-radius: 10px; padding: 12px 8px; text-align: center; }
.stat-num { font-size: 20px; font-weight: bold; color: #667eea; display: block; }
.stat-num.red { color: #ee0a24; }
.stat-num.orange { color: #ff976a; }
.stat-num.green { color: #07c160; }
.stat-label { font-size: 11px; color: #969799; margin-top: 4px; display: block; }

.filter-bar { display: flex; gap: 8px; margin-bottom: 12px; overflow-x: auto; }
.filter-item { padding: 6px 14px; background: #fff; border-radius: 20px; font-size: 13px; color: #646566; white-space: nowrap; border: 1px solid #ebedf0; }
.filter-item.active { background: #667eea; color: #fff; border-color: #667eea; }

.card { background: #fff; border-radius: 12px; padding: 16px; margin-bottom: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.02); }
.card-title { display: flex; justify-content: space-between; margin-bottom: 12px; border-bottom: 1px solid #f2f3f5; padding-bottom: 10px; }
.card-title .name { font-size: 16px; font-weight: bold; }
.status { font-size: 12px; padding: 2px 8px; border-radius: 4px; }
.status.success { background: #e8f9f0; color: #07c160; }
.status.warning { background: #fff7e8; color: #ff976a; }
.status.primary { background: #eef5ff; color: #667eea; }
.status.default { background: #f7f8fa; color: #969799; }

.card-info { font-size: 13px; color: #646566; }
.info-line { margin-bottom: 6px; }
.info-line.date { color: #969799; font-size: 12px; }

.user-card { display: flex; justify-content: space-between; align-items: center; }
.user-main { display: flex; gap: 12px; align-items: center; }
.user-avatar { width: 40px; height: 40px; background: #f0f2f5; border-radius: 20px; display: flex; align-items: center; justify-content: center; font-size: 20px; }
.user-name { font-size: 15px; font-weight: bold; display: flex; align-items: center; gap: 6px; }
.user-phone { font-size: 13px; color: #969799; }
.role-tag { font-size: 10px; padding: 1px 4px; border-radius: 3px; font-weight: normal; }
.role-tag.admin { background: #fff2f0; color: #ff4d4f; border: 1px solid #ffa39e; }
.role-tag.dsl_admin { background: #f0f0ff; color: #667eea; border: 1px solid #b0b4ff; }
.role-tag.user { background: #e6f7ff; color: #1890ff; border: 1px solid #91d5ff; }

.action-btn { font-size: 12px; border-radius: 8px; }

.modal { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); display: flex; align-items: flex-end; z-index: 100; }
.modal-content { background: #fff; width: 100%; border-radius: 16px 16px 0 0; padding: 24px 16px; max-height: 80vh; overflow-y: auto; }
.modal-title { font-size: 17px; font-weight: bold; text-align: center; margin-bottom: 16px; }
.detail-rows { margin-bottom: 12px; }
.detail-row { display: flex; padding: 8px 0; border-bottom: 1px solid #f7f8fa; font-size: 14px; }
.dlabel { color: #969799; width: 70px; flex-shrink: 0; }
.status-options { display: flex; flex-direction: column; gap: 10px; margin-bottom: 20px; }
.opt-item { background: #f7f8fa; padding: 14px; text-align: center; border-radius: 10px; font-size: 15px; }
.opt-item:active { background: #ebedf0; }
.close-btn { border-radius: 99px; font-size: 15px; }
</style>
