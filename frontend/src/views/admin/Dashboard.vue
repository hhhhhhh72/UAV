<template>
  <div class="dashboard">
    <!-- 顶部 -->
    <div class="top-bar">
      <div>
        <h1 class="top-title">数据看板</h1>
        <span class="top-date">{{ today }}</span>
      </div>
      <span class="top-badge"><i class="dot"></i>实时更新</span>
    </div>

    <!-- 指标卡片 -->
    <div class="metrics-grid">
      <MetricCard label="需求总数" :value="stats.totalDemands" sub="累计发布" value-color="#0A66C2" />
      <MetricCard label="待审企业" :value="stats.pendingEnterprises" sub="企业入驻" value-color="#f59e0b" />
      <MetricCard label="内容帖子" :value="stats.totalPosts" sub="社区" value-color="#10b981" />
      <MetricCard label="平台用户" :value="stats.totalUsers" sub="注册用户" value-color="#6366f1" />
      <MetricCard v-if="isPlatformAdmin || isAssociationAdmin" label="待处举报" :value="stats.pendingReports" sub="待处理" value-color="#ef4444" />
    </div>

    <!-- 总览图表 -->
    <div class="charts-row">
      <div class="chart-card chart-wide">
        <div class="chart-head"><span>平台数据总览</span><span class="chart-hint">{{ roleLabel }}</span></div>
        <v-chart :option="overviewOption" autoresize class="chart" />
      </div>
      <div class="chart-card chart-wide">
        <div class="chart-head"><span>需求分布</span></div>
        <v-chart :option="pieOption" autoresize class="chart" />
      </div>
    </div>

    <!-- 第二行图表 -->
    <div class="charts-row">
      <div class="chart-card chart-wide">
        <div class="chart-head"><span>需求状态分布</span></div>
        <v-chart :option="statusOption" autoresize class="chart" />
      </div>
      <div class="chart-card chart-wide">
        <div class="chart-head"><span>用户增长</span></div>
        <v-chart :option="userGrowthOption" autoresize class="chart" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import axios from '@/utils/http'
import { showFailToast } from 'vant'
import MetricCard from './components/MetricCard.vue'
import { useAuth } from './composables/useAuth'

const { userRole, isPlatformAdmin, isAssociationAdmin } = useAuth()
const roleLabel = computed(() => isAssociationAdmin.value ? '协会' : '所有服务')

use([CanvasRenderer, LineChart, PieChart, BarChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

const today = new Date().toLocaleDateString('zh-CN', { year:'numeric', month:'long', day:'numeric', weekday:'long' })

const stats = ref({
  totalDemands: 0, pendingEnterprises: 0, totalPosts: 0,
  totalUsers: 0, pendingReports: 0,
  orderTrend: [], competitionByRole: {}, userGrowth: [], statusDist: {}
})

const COLORS = {
  blue: '#0A66C2', green: '#10b981', orange: '#f59e0b',
  red: '#ef4444', purple: '#6366f1', teal: '#14b8a6', gray: '#868e96'
}

const growthOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 48, right: 16, top: 16, bottom: 24 },
  xAxis: {
    type: 'category',
    data: stats.value.userGrowth.map(d => d.date.slice(2)),
    axisLabel: { color: '#868e96', fontSize: 10 },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value', minInterval: 1,
    splitLine: { lineStyle: { color: '#f0f0f2' } },
    axisLabel: { color: '#868e96', fontSize: 10 }
  },
  series: [{
    name: '新增用户', data: stats.value.userGrowth.map(d => d.count),
    type: 'line', smooth: true, symbol: 'circle', symbolSize: 5,
    lineStyle: { color: COLORS.purple, width: 2 },
    itemStyle: { color: COLORS.purple },
    areaStyle: { color: { type:'linear',x:0,y:0,x2:0,y2:1,
      colorStops: [{offset:0,color:'rgba(99,102,241,.12)'},{offset:1,color:'rgba(99,102,241,.01)'}] }}
  }]
}))

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
  const cats = stats.value.competitionByRole
  const colorList = [COLORS.blue, COLORS.orange, COLORS.green, COLORS.purple, COLORS.teal]
  const data = Object.entries(cats).map(([name, value], i) => ({
    value, name, itemStyle: { color: colorList[i % colorList.length] }
  }))
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { color: '#868e96', fontSize: 11 } },
    series: [{ type: 'pie', radius: ['45%','70%'], center: ['50%','42%'],
      label: { show: false }, data }]
  }
})

const userGrowthOption = computed(() => ({
  ...growthOption.value,
  ...{ series: [{ ...growthOption.value.series[0], name: '新增用户' }] }
}))

const statusOption = computed(() => {
  const dist = stats.value.statusDist
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { color: '#868e96', fontSize: 11 } },
    series: [{
      type: 'pie', radius: ['45%','70%'], center: ['50%','42%'], label: { show: false },
      data: Object.entries(dist).map(([name, value]) => ({
        value, name,
        itemStyle: { color: name.includes('已发布') ? COLORS.blue : COLORS.gray }
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
        competitionByRole: d.category_dist || {},
        userGrowth: d.trends || [],
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
.dashboard { max-width: 1200px; margin: 0 auto; }

.top-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.top-title { font-size: 22px; font-weight: 700; margin: 0; color: #1a1a1a; }
.top-date { font-size: 13px; color: #868e96; margin-top: 2px; display: block; }
.top-badge { font-size: 12px; color: #10b981; display: flex; align-items: center; gap: 6px; }
.dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #10b981; }

.metrics-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 14px; margin-bottom: 20px; }
.metric-value { font-size: 24px; }

.charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.chart-wide { grid-column: span 1; }
.chart-narrow { grid-column: span 1; }
.chart-full { grid-column: span 2; }

.chart-card { background: #fff; border-radius: 12px; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,.08); }
.chart-head { display: flex; justify-content: space-between; align-items: center; font-size: 14px; font-weight: 600; color: #1d1d1f; margin-bottom: 12px; }
.chart-hint { font-size: 11px; color: #868e96; font-weight: 400; }
.chart { width: 100%; height: 260px; }

.chart-title { font-size: 14px; font-weight: 600; color: #1d1d1f; margin: 0 0 12px 0; }

@media (max-width: 1024px) {
  .metrics-grid { grid-template-columns: repeat(3, 1fr); gap: 12px; }
}
@media (max-width: 767px) {
  .metrics-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; margin-bottom: 14px; }
  .charts-row { grid-template-columns: 1fr; gap: 12px; margin-bottom: 12px; }
  .chart-full { grid-column: span 1; }
  .chart { height: 200px; }
  .chart-card { padding: 14px; }
}
</style>
