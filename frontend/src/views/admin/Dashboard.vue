<template>
  <div class="admin-page multi-dimension-container">
    <div class="page-header">
      <div class="page-header-main">
        <h2>数据看板</h2>
        <span class="page-sub">平台运营总览 · 统计窗口：{{ rangeLabel }}</span>
      </div>
      <a-radio-group v-model="range" type="button" size="small" @change="fetchStats">
        <a-radio v-for="opt in RANGE_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</a-radio>
      </a-radio-group>
    </div>

    <!-- KPI 行 -->
    <a-row :gutter="16" class="kpi-row">
      <a-col :span="6">
        <a-card class="general-card kpi-card" :bordered="false">
          <div class="kpi-label"><icon-list style="color: #5B8FF9;" /> 需求总数</div>
          <div class="kpi-value">{{ stats.totalDemands.toLocaleString() }}</div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="general-card kpi-card" :bordered="false">
          <div class="kpi-label"><icon-edit style="color: #F6903D;" /> 待审企业</div>
          <div class="kpi-value">{{ stats.pendingEnterprises.toLocaleString() }}</div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="general-card kpi-card" :bordered="false">
          <div class="kpi-label"><icon-eye style="color: #78D3F8;" /> 内容帖子</div>
          <div class="kpi-value">{{ stats.totalPosts.toLocaleString() }}</div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="general-card kpi-card" :bordered="false">
          <div class="kpi-label"><icon-user style="color: #9270CA;" /> 平台用户</div>
          <div class="kpi-value">{{ stats.totalUsers.toLocaleString() }}</div>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" class="main-row">
      <a-col :span="16">
        <a-card class="general-card" :bordered="false" style="height: 100%;">
          <template #title>运营趋势</template>
          <template #extra>
            <span class="card-extra">{{ rangeStart }} → {{ rangeEnd }} · 按{{ bucketUnit }} · 各指标独立刻度</span>
          </template>
          <!-- 小多图：每个指标一张独立刻度的迷你趋势图。
               四条线画在同一张图上时，量级差异（成交 10 / 工单 1）会把小指标压成贴着零轴的直线，
               各自独立刻度后形状可比、量级由标题上的窗口合计给出 -->
          <a-spin :loading="loading" style="display: block;">
            <div class="trend-grid">
              <div v-for="m in TREND_SERIES" :key="m.key" class="trend-cell">
                <div class="trend-cell-head">
                  <span class="dot" :style="{ background: m.color }"></span>
                  <span class="trend-cell-name">{{ m.name }}</span>
                  <span class="trend-cell-total">{{ totals[m.key] }}</span>
                </div>
                <div class="trend-cell-chart">
                  <v-chart :option="miniOption(m)" autoresize />
                </div>
              </div>
            </div>
          </a-spin>
        </a-card>
      </a-col>
      <a-col :span="8" class="side-col">
        <a-card class="general-card" title="模块数据" :bordered="false" style="flex: 1;">
          <div style="height: 180px;">
            <v-chart :option="barChartOption" autoresize />
          </div>
        </a-card>
        <a-card class="general-card" title="需求类型分布" :bordered="false" style="flex: 1;">
          <div style="height: 180px;">
            <v-chart :option="radarChartOption" autoresize />
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-card class="general-card" title="需求状态分布" :bordered="false">
      <a-row :gutter="16" style="align-items: center;">
        <a-col :span="12">
          <div style="height: 280px;"><v-chart :option="statusPieOption" autoresize /></div>
        </a-col>
        <a-col :span="12">
          <a-table :data="statusTableData" :pagination="false" :bordered="false" class="status-table">
            <template #columns>
              <a-table-column title="状态" data-index="name" />
              <a-table-column title="数量" data-index="value" />
              <a-table-column title="占比" data-index="pct" />
            </template>
          </a-table>
        </a-col>
      </a-row>
    </a-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { LineChart, BarChart, RadarChart, PieChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import VChart from 'vue-echarts';
import axios from '@/utils/http';
import { showFailToast } from '@/utils/feedback';

use([CanvasRenderer, LineChart, BarChart, RadarChart, PieChart, GridComponent, TooltipComponent, LegendComponent]);

/* 时间窗口：与后端 ?range= 取值一致（7d/30d/90d 按日分桶，12m 按月分桶） */
const RANGE_OPTIONS = [
  { value: '7d', label: '近 7 天' },
  { value: '30d', label: '近 30 天' },
  { value: '90d', label: '近 90 天' },
  { value: '12m', label: '近 12 个月' }
]
const range = ref('30d')
const loading = ref(false)
const rangeLabel = computed(() => (RANGE_OPTIONS.find(o => o.value === range.value) || {}).label || '')
const bucketUnit = computed(() => (stats.value.bucket === 'month' ? '月' : '日'))

const stats = ref({
  totalDemands: 0, pendingEnterprises: 0, totalPosts: 0,
  totalUsers: 0, pendingReports: 0, bucket: 'day',
  orderTrend: [], competitionByRole: [], statusDist: {}
})
const trendsDetail = ref({ demand: [], post: [], user: [], message: [], article: [], enrollment: [], order: [], work_order: [] })

/* 业务趋势四条线：需求 / 报名 / 成交 / 工单（与后端 trends_detail 键一致） */
const TREND_SERIES = [
  { key: 'demand', name: '需求', color: '#5B8FF9' },
  { key: 'enrollment', name: '报名', color: '#9270CA' },
  { key: 'order', name: '成交', color: '#F6903D' },
  { key: 'work_order', name: '工单', color: '#61DDAA' }
]

const formatBucket = (date, bucket) => {
  if (!date) return ''
  if (bucket === 'month') return String(date).slice(5) + '月'
  return String(date).slice(5)
}

const bucketLabels = computed(() => (trendsDetail.value.demand || []).map(d => formatBucket(d.date, stats.value.bucket)))
const rangeStart = computed(() => bucketLabels.value[0] || '')
const rangeEnd = computed(() => bucketLabels.value[bucketLabels.value.length - 1] || '')

/* 窗口内合计：形状看走势，数字看总量 */
const totals = computed(() => {
  const out = {}
  for (const s of TREND_SERIES) {
    out[s.key] = (trendsDetail.value[s.key] || []).reduce((sum, d) => sum + (d.count || 0), 0)
  }
  return out
})

/* 迷你趋势图：单指标、独立刻度、无坐标轴文字，靠悬浮提示读数 */
const miniOption = (m) => {
  const data = (trendsDetail.value[m.key] || []).map(d => d.count || 0)
  const max = Math.max(1, ...data)
  return {
    tooltip: { trigger: 'axis', formatter: (ps) => (ps && ps[0] ? ps[0].axisValue + '：' + ps[0].data : '') },
    grid: { left: 2, right: 6, top: 6, bottom: 2, containLabel: false },
    xAxis: { type: 'category', boundaryGap: false, data: bucketLabels.value, axisLine: { show: false }, axisTick: { show: false }, axisLabel: { show: false } },
    yAxis: { type: 'value', max, min: 0, minInterval: 1, splitLine: { show: false }, axisLabel: { show: false }, axisLine: { show: false }, axisTick: { show: false } },
    series: [{
      type: 'line',
      smooth: false,
      symbol: 'none',
      data,
      lineStyle: { width: 2, color: m.color },
      areaStyle: { color: m.color, opacity: 0.12 }
    }]
  }
}

// 模块数据：横向圆角柱
const barChartOption = computed(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: '3%', right: '4%', bottom: '3%', top: 10, containLabel: true },
  xAxis: { type: 'value' },
  yAxis: { type: 'category', data: ['平台用户', '内容帖子', '待审企业', '需求总数'] },
  series: [
    {
      type: 'bar',
      barWidth: 10,
      data: [
        { value: stats.value.totalUsers, itemStyle: { color: '#9270CA', borderRadius: 5 } },
        { value: stats.value.totalPosts, itemStyle: { color: '#78D3F8', borderRadius: 5 } },
        { value: stats.value.pendingEnterprises, itemStyle: { color: '#F6903D', borderRadius: 5 } },
        { value: stats.value.totalDemands, itemStyle: { color: '#5B8FF9', borderRadius: 5 } }
      ]
    }
  ]
}));

// 需求类型：枚举 → 中文（与小程序 enums.js 一致）
const BIZ_LABEL = {
  cable_inspection: '巡检', plant_transport: '植保', spray_pesticide: '农药',
  trade_lease: '租赁', clean_paint: '清洗', other: '其他'
}

const radarChartOption = computed(() => {
  const cats = Array.isArray(stats.value.competitionByRole) ? stats.value.competitionByRole : Object.entries(stats.value.competitionByRole || {})
  const top = cats.slice(0, 6)
  const indicator = top.length ? top.map(([name]) => ({ name: BIZ_LABEL[name] || name, max: Math.max(...top.map(([, v]) => Number(v) || 1), 1) })) : [{ name: '暂无数据', max: 1 }]
  return {
    tooltip: {},
    radar: { indicator, radius: '60%' },
    series: [
      {
        type: 'radar',
        data: [
          {
            value: top.length ? top.map(([, v]) => Number(v) || 0) : [0],
            name: '需求类型',
            itemStyle: { color: '#5B8FF9' },
            areaStyle: { opacity: 0.1, color: '#5B8FF9' }
          }
        ]
      }
    ]
  }
});

const PIE_COLORS = ['#5B8FF9', '#9270CA', '#78D3F8', '#F6903D', '#61DDAA']
const statusPieData = computed(() => {
  const dist = stats.value.statusDist || {}
  return Object.entries(dist).map(([name, value], i) => ({
    name, value,
    itemStyle: { color: PIE_COLORS[i % PIE_COLORS.length] }
  }))
})

const statusPieOption = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0, icon: 'circle' },
  series: [{
    type: 'pie',
    radius: ['60%', '80%'],
    label: { show: true, formatter: '{d}%' },
    data: statusPieData.value
  }]
}))

const statusTableData = computed(() => {
  const dist = stats.value.statusDist || {}
  const total = Object.values(dist).reduce((s, v) => s + (v || 0), 0) || 1
  return Object.entries(dist).map(([name, value]) => ({
    name, value,
    pct: Math.round((value / total) * 100) + '%'
  }))
})

const fetchStats = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/admin/dashboard', { params: { range: range.value } })
    const d = res.data
    if (d) {
      stats.value = {
        totalDemands: d.total_demands ?? 0,
        pendingEnterprises: d.pending_enterprises ?? 0,
        totalPosts: d.total_posts ?? 0,
        totalUsers: d.total_users ?? 0,
        pendingReports: d.pending_reports ?? 0,
        bucket: d.bucket || 'day',
        orderTrend: d.trends || [],
        competitionByRole: d.category_dist || [],
        statusDist: d.status_dist || {}
      }
      trendsDetail.value = d.trends_detail || {}
    }
  } catch (err) {
    showFailToast('获取统计数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<style scoped>
/* Arco Pro multi-dimension 版式 */
.multi-dimension-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
}
.page-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.page-header-main {
  display: flex;
  align-items: baseline;
  gap: 12px;
  flex-wrap: wrap;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-1);
}
.page-sub {
  font-size: 13px;
  color: var(--color-text-3);
}
.kpi-row,
.main-row {
  margin: 0;
}
.kpi-card :deep(.arco-card-body) {
  padding: 16px;
}
.kpi-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-3);
  font-size: 13px;
}
.kpi-value {
  margin-top: 6px;
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-1);
  font-variant-numeric: tabular-nums;
}
.general-card {
  border-radius: 4px;
}
.side-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.card-extra {
  font-size: 12px;
  color: var(--color-text-3);
}
/* 小多图：2×2 网格，每格一个指标 */
.trend-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px 24px;
}
.trend-cell-head {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.trend-cell-name {
  font-size: 13px;
  color: var(--color-text-3);
}
.trend-cell-total {
  margin-left: auto;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-1);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}
.trend-cell-chart {
  height: 108px;
  margin-top: 4px;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
</style>
