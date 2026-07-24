<template>
  <div class="dashboard">
    <!-- Metric Cards -->
    <div class="metrics-grid">
      <MetricCard label="需求总数" :value="stats.totalDemands" sub="累计发布" />
      <MetricCard label="待审企业" :value="stats.pendingEnterprises" value-color="#ff9f0a" sub="企业入驻" />
      <MetricCard label="内容帖子" :value="stats.totalPosts" value-color="#0071e3" sub="社区" />
      <MetricCard label="平台用户" :value="stats.totalUsers" value-color="#34c759" sub="注册用户" />
      <MetricCard v-if="isPlatformAdmin || isAssociationAdmin" label="待处举报" :value="stats.pendingReports" value-color="#5856d6" />
    </div>

    <!-- Overview Chart -->
    <div class="charts-row">
      <div class="chart-card chart-full">
        <h3 class="chart-title">平台数据总览</h3>
        <v-chart :option="overviewOption" autoresize class="chart" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import axios from '@/utils/http'
import { showFailToast } from 'vant'
import MetricCard from './components/MetricCard.vue'
import { useAuth } from './composables/useAuth'

const { userRole, isPlatformAdmin, isAssociationAdmin } = useAuth()

const roleLabel = computed(() => {
  if (isAssociationAdmin.value) return '协会'
  return '所有服务'
})

use([CanvasRenderer, LineChart, PieChart, BarChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

const stats = ref({
  totalDemands: 0,
  pendingEnterprises: 0,
  totalPosts: 0,
  totalUsers: 0,
  pendingReports: 0,
  orderTrend: [],
  competitionByRole: {},
  userGrowth: [],
  statusDist: {}
})

const COLORS = {
  blue: '#0071e3',
  green: '#34c759',
  orange: '#ff9f0a',
  red: '#ff3b30',
  purple: '#5856d6',
  teal: '#5ac8fa',
  gray: '#86868b'
}

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 40, right: 16, top: 16, bottom: 24 },
  xAxis: {
    type: 'category',
    data: stats.value.orderTrend.map(d => d.date.slice(5)),
    axisLine: { lineStyle: { color: '#e5e5e7' } },
    axisLabel: { color: '#86868b', fontSize: 11 },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value',
    minInterval: 1,
    splitLine: { lineStyle: { color: '#f0f0f2' } },
    axisLabel: { color: '#86868b', fontSize: 11 }
  },
  series: [{
    data: stats.value.orderTrend.map(d => d.count),
    type: 'line',
    smooth: true,
    symbol: 'circle',
    symbolSize: 6,
    lineStyle: { color: COLORS.blue, width: 2 },
    itemStyle: { color: COLORS.blue },
    areaStyle: {
      color: {
        type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: 'rgba(0, 113, 227, 0.15)' },
          { offset: 1, color: 'rgba(0, 113, 227, 0.01)' }
        ]
      }
    }
  }]
}))

const pieOption = computed(() => {
  const r = stats.value.competitionByRole
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { color: '#86868b', fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '42%'],
      avoidLabelOverlap: true,
      label: { show: false },
      data: [
        { value: r.athlete, name: '运动员', itemStyle: { color: COLORS.blue } },
        { value: r.coach, name: '教练员', itemStyle: { color: COLORS.orange } },
        { value: r.referee, name: '裁判员', itemStyle: { color: COLORS.green } },
        { value: r.club, name: '俱乐部', itemStyle: { color: COLORS.purple } }
      ]
    }]
  }
})

const userGrowthOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 40, right: 16, top: 16, bottom: 24 },
  xAxis: {
    type: 'category',
    data: stats.value.userGrowth.map(d => d.month.slice(5) + '月'),
    axisLine: { lineStyle: { color: '#e5e5e7' } },
    axisLabel: { color: '#86868b', fontSize: 11 },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value',
    minInterval: 1,
    splitLine: { lineStyle: { color: '#f0f0f2' } },
    axisLabel: { color: '#86868b', fontSize: 11 }
  },
  series: [{
    data: stats.value.userGrowth.map(d => d.count),
    type: 'bar',
    barWidth: '50%',
    itemStyle: { color: COLORS.teal, borderRadius: [4, 4, 0, 0] }
  }]
}))

const statusOption = computed(() => {
  const colorMap = {
    '待处理': COLORS.orange,
    '处理中': COLORS.blue,
    '已完成': COLORS.green,
    '已取消': COLORS.gray
  }
  const dist = stats.value.statusDist
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { color: '#86868b', fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '42%'],
      label: { show: false },
      data: Object.entries(dist).map(([name, value]) => ({
        value,
        name,
        itemStyle: { color: colorMap[name] || COLORS.gray }
      }))
    }]
  }
})

const overviewOption = computed(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: 60, right: 16, top: 16, bottom: 24 },
  xAxis: {
    type: 'category',
    data: ['需求总数', '待审企业', '内容帖子', '平台用户', '待处举报'],
    axisLabel: { color: '#86868b', fontSize: 11 }
  },
  yAxis: { type: 'value', minInterval: 1, splitLine: { lineStyle: { color: '#f0f0f2' } } },
  series: [{
    type: 'bar', barWidth: '40%',
    data: [
      { value: stats.value.totalDemands, itemStyle: { color: COLORS.blue } },
      { value: stats.value.pendingEnterprises, itemStyle: { color: COLORS.orange } },
      { value: stats.value.totalPosts, itemStyle: { color: COLORS.green } },
      { value: stats.value.totalUsers, itemStyle: { color: COLORS.teal } },
      { value: stats.value.pendingReports, itemStyle: { color: COLORS.red } }
    ]
  }]
}))

const fetchStats = async () => {
  try {
    const res = await axios.get('/api/v1/admin/dashboard')
    const d = res.data
    if (d) {
      stats.value = {
        totalDemands: d.total_demands ?? 0,
        pendingEnterprises: d.pending_enterprises ?? 0,
        totalPosts: d.total_posts ?? 0,
        totalUsers: d.total_users ?? 0,
        pendingReports: d.pending_reports ?? 0,
        orderTrend: d.trends || [],
        competitionByRole: d.type_dist || {},
        userGrowth: [],
        statusDist: d.status_dist || {}
      }
    }
  } catch (err) {
    console.error(err)
    showFailToast('获取统计数据失败')
  }
}

onMounted(fetchStats)
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.charts-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.chart-wide {
  grid-column: span 1;
}

.chart-narrow {
  grid-column: span 1;
}

.chart-full {
  grid-column: span 2;
}

.chart-card {
  background: var(--bg-primary, #fff);
  border-radius: var(--card-radius, 12px);
  padding: 20px;
  box-shadow: var(--card-shadow, 0 1px 3px rgba(0, 0, 0, 0.08));
}

.chart-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color, #1d1d1f);
  margin: 0 0 12px 0;
}

.chart {
  width: 100%;
  height: 240px;
}

/* Tablet */
@media (max-width: 1024px) {
  .metrics-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }
}

/* Mobile */
@media (max-width: 767px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    margin-bottom: 14px;
  }
  .charts-row {
    grid-template-columns: 1fr;
    gap: 12px;
    margin-bottom: 12px;
  }
  .chart-full {
    grid-column: span 1;
  }
  .chart {
    height: 200px;
  }
  .chart-card {
    padding: 14px;
  }
}
</style>
