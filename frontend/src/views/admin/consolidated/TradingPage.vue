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
      <div class="chart-row">
        <div class="chart-card">
          <div class="chart-title">需求月度趋势</div>
          <v-chart :option="trendOption" autoresize class="chart" />
        </div>
        <div class="chart-card">
          <div class="chart-title">需求状态分布</div>
          <v-chart :option="pieOption" autoresize class="chart" />
        </div>
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
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import axios from '@/utils/http'
import DemandList from '../demands/DemandList.vue'
import OrderList from '../orders/OrderList.vue'
import ReviewList from '../reviews/ReviewList.vue'
import ProductList from '../products/ProductList.vue'

use([CanvasRenderer, LineChart, PieChart, TooltipComponent, LegendComponent, GridComponent])

const tab = ref('demands')
const dash = ref({ totalDemands: 0, completed: 0, offlineAmount: '0.00', rate: 0 })
const trends = ref([])
const statusDist = ref({})

const STATUS_LABEL = { pending: '待审核', published: '已公开', completed: '已完成', cancelled: '已取消', rejected: '已驳回' }

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 48, right: 16, top: 16, bottom: 24 },
  xAxis: {
    type: 'category',
    data: trends.value.map(d => String(d.date || '').slice(5)),
    axisLabel: { color: '#868e96', fontSize: 10 },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value', minInterval: 1,
    splitLine: { lineStyle: { color: '#f0f0f2' } },
    axisLabel: { color: '#868e96', fontSize: 10 }
  },
  series: [{
    name: '需求数',
    data: trends.value.map(d => d.count || 0),
    type: 'line', smooth: true, symbol: 'circle', symbolSize: 5,
    lineStyle: { color: '#0A66C2', width: 2 },
    itemStyle: { color: '#0A66C2' },
    areaStyle: { color: 'rgba(10,102,194,.08)' }
  }]
}))

const pieOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
  legend: { bottom: 0, textStyle: { fontSize: 10, color: '#868e96' } },
  series: [{
    type: 'pie',
    radius: ['42%', '68%'],
    center: ['50%', '44%'],
    avoidLabelOverlap: true,
    itemStyle: { borderRadius: 4, borderColor: '#fff', borderWidth: 1 },
    label: { show: false },
    data: Object.entries(statusDist.value).map(([k, v]) => ({
      name: STATUS_LABEL[k] || k,
      value: v
    }))
  }]
}))

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
    trends.value = d.trends || []
    statusDist.value = d.status_dist || {}
  } catch (e) { /* 看板失败不影响列表 */ }
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
.chart-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.chart-card { background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,.06); }
.chart-title { font-size: 14px; font-weight: 600; color: var(--el-text-color-primary); margin-bottom: 8px; }
.chart { height: 220px; }
@media (max-width: 900px) { .metric-row, .chart-row { grid-template-columns: 1fr 1fr; } }
</style>
