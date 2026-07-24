<template>
  <div class="dashboard">
    <!-- Metric Cards -->
    <div class="metrics-grid">
      <MetricCard label="需求总数" :value="dashboard.totalDemands" sub="累计发布" />
      <MetricCard label="企业入驻数" :value="dashboard.totalEnterprises" value-color="#0071e3" sub="已认证企业" />
      <MetricCard label="培训认证数" :value="dashboard.totalCerts" value-color="#5856d6" sub="飞手/教员" />
      <MetricCard label="成交率" :value="dashboard.completionRate" value-color="#34c759" sub="近30天" />
    </div>

    <!-- Charts -->
    <div class="charts-row">
      <div class="chart-card chart-wide">
        <h3 class="chart-title">需求发布趋势（近30天）</h3>
        <v-chart :option="trendOption" autoresize class="chart" />
      </div>
      <div class="chart-card chart-narrow">
        <h3 class="chart-title">需求类型分布</h3>
        <v-chart :option="typeDistOption" autoresize class="chart" />
      </div>
    </div>

    <div class="charts-row">
      <div class="chart-card chart-wide">
        <h3 class="chart-title">企业审核状态分布</h3>
        <v-chart :option="statusDistOption" autoresize class="chart" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import http from '@/utils/http'
import { showFailToast } from 'vant'
import MetricCard from './components/MetricCard.vue'

use([CanvasRenderer, LineChart, PieChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

const dashboard = ref({
  totalDemands: 0,
  totalEnterprises: 0,
  totalCerts: 0,
  completionRate: '0%',
  trends: [],
  typeDist: {},
  statusDist: {}
})

const COLORS = {
  blue: '#0071e3',
  green: '#34c759',
  orange: '#ff9f0a',
  red: '#ff3b30',
  purple: '#5856d6',
  teal: '#5ac8fa',
  cyan: '#32d6d6',
  gray: '#86868b'
}

const TYPE_LABELS = {
  patrol: '巡检',
  plant: '植保',
  pesticide: '农药',
  lease: '租赁',
  clean: '清洗'
}

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 40, right: 16, top: 16, bottom: 24 },
  xAxis: {
    type: 'category',
    data: dashboard.value.trends.map(d => d.date ? d.date.slice(5) : ''),
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
    data: dashboard.value.trends.map(d => d.count || 0),
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

const TYPE_COLORS = {
  patrol: COLORS.blue,
  plant: COLORS.green,
  pesticide: COLORS.orange,
  lease: COLORS.purple,
  clean: COLORS.teal
}

const typeDistOption = computed(() => {
  const dist = dashboard.value.typeDist || {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, textStyle: { color: '#86868b', fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '42%'],
      avoidLabelOverlap: true,
      label: { show: false },
      data: Object.entries(dist).map(([key, value]) => ({
        value,
        name: TYPE_LABELS[key] || key,
        itemStyle: { color: TYPE_COLORS[key] || COLORS.gray }
      }))
    }]
  }
})

const STATUS_LABELS = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已驳回'
}

const STATUS_COLORS = {
  pending: COLORS.orange,
  approved: COLORS.green,
  rejected: COLORS.red
}

const statusDistOption = computed(() => {
  const dist = dashboard.value.statusDist || {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, textStyle: { color: '#86868b', fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '42%'],
      label: { show: false },
      data: Object.entries(dist).map(([key, value]) => ({
        value,
        name: STATUS_LABELS[key] || key,
        itemStyle: { color: STATUS_COLORS[key] || COLORS.gray }
      }))
    }]
  }
})

const fetchDashboard = async () => {
  try {
    const res = await http.get('/api/v1/admin/dashboard')
    if (res.data) {
      dashboard.value = res.data
    }
  } catch (err) {
    console.error(err)
    showFailToast('获取统计数据失败')
  }
}

onMounted(fetchDashboard)
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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
    grid-template-columns: repeat(2, 1fr);
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
  .chart {
    height: 200px;
  }
  .chart-card {
    padding: 14px;
  }
}
</style>
