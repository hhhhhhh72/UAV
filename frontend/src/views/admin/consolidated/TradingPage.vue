<template>
  <div class="admin-page">
    <div class="page-header"><h2>交易管理</h2></div>

    <!-- 交易看板 -->
    <div class="dashboard">
      <div class="metric-row">
        <div class="metric-card"><div class="m-num">{{ dash.totalDemands }}</div><div class="m-label">需求总数</div></div>
        <div class="metric-card"><div class="m-num ok">{{ dash.completed }}</div><div class="m-label">已成交需求</div></div>
        <div class="metric-card"><div class="m-num money">¥{{ dash.offlineAmount }}</div><div class="m-label">撮合成交额</div></div>
        <div class="metric-card"><div class="m-num">{{ dash.rate }}%</div><div class="m-label">撮合率</div></div>
      </div>
    </div>

    <el-tabs v-model="tab">
      <el-tab-pane label="需求管理" name="demands"><DemandList /></el-tab-pane>
      <el-tab-pane label="订单管理" name="orders"><OrderList /></el-tab-pane>
      <el-tab-pane label="评价管理" name="reviews"><ReviewList /></el-tab-pane>
      <el-tab-pane label="商品管理" name="products"><ProductList /></el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/http'
import DemandList from '../demands/DemandList.vue'
import OrderList from '../orders/OrderList.vue'
import ReviewList from '../reviews/ReviewList.vue'
import ProductList from '../products/ProductList.vue'

const tab = ref('demands')
const dash = ref({ totalDemands: 0, completed: 0, offlineAmount: '0.00', rate: 0 })

const loadDashboard = async () => {
  try {
    const res = await axios.get('/api/v1/admin/dashboard')
    const d = res.data || {}
    const total = d.total_demands || 0
    const completed = (d.status_dist && d.status_dist.completed) || 0
    dash.value = {
      totalDemands: total,
      completed,
      offlineAmount: ((d.offline_amount_total || 0) / 100).toLocaleString('zh-CN', { minimumFractionDigits: 2 }),
      rate: total ? Math.round((completed / total) * 100) : 0
    }
  } catch (e) { /* 指标卡失败不影响列表 */ }
}

onMounted(loadDashboard)
</script>

<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
h2 { margin: 0; font-size: 20px; }

/* 看板 */
.dashboard { margin-bottom: 16px; }
.metric-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 12px; }
.metric-card { background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,.06); }
.m-num { font-size: 26px; font-weight: 700; color: var(--el-text-color-primary); }
.m-num.ok { color: var(--el-color-success); }
.m-num.money { color: var(--el-color-warning); }
.m-label { font-size: 13px; color: var(--el-text-color-secondary); margin-top: 4px; }
@media (max-width: 900px) { .metric-row { grid-template-columns: 1fr 1fr; } }
</style>
