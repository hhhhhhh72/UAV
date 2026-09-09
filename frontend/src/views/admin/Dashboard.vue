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
      <a-col :span="16" class="main-col">
        <a-card class="general-card main-card" :bordered="false">
          <template #title>运营趋势</template>
          <template #extra>
            <span class="card-extra">{{ rangeStart }} → {{ rangeEnd }} · 按{{ bucketUnit }}统计</span>
          </template>

          <!-- 指标切换：一次只看一条曲线，避免量级差把曲线压扁；四个窗口合计并排可比 -->
          <div class="metric-tabs">
            <button
              v-for="m in TREND_SERIES"
              :key="m.key"
              type="button"
              class="metric-tab"
              :class="{ 'metric-tab--active': m.key === activeMetric }"
              :style="m.key === activeMetric ? { color: m.color, borderColor: m.color, background: tint(m.color) } : {}"
              @click="activeMetric = m.key"
            >
              <span class="dot" :style="{ background: m.color }"></span>
              <span class="metric-tab-name">{{ m.name }}</span>
              <span class="metric-tab-total">{{ totals[m.key] }}</span>
            </button>
          </div>

          <div class="trend-chart">
            <a-spin :loading="loading" style="display: block; height: 100%;">
              <v-chart :option="trendOption" autoresize />
            </a-spin>
          </div>
        </a-card>
      </a-col>
      <a-col :span="8" class="side-col">
        <a-card class="general-card side-card" title="模块数据" :bordered="false">
          <div class="side-chart"><v-chart :option="barChartOption" autoresize /></div>
        </a-card>
        <a-card class="general-card side-card" title="需求类型分布" :bordered="false">
          <div class="side-chart"><v-chart :option="radarChartOption" autoresize /></div>
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
const activeMetric = ref('demand')

const formatBucket = (date, bucket) => {
  if (!date) return ''
  if (bucket === 'month') return String(date).slice(5) + '月'
  return String(date).slice(5)
}

const bucketLabels = computed(() => (trendsDetail.value.demand || []).map(d => formatBucket(d.date, stats.value.bucket)))
const rangeStart = computed(() => bucketLabels.value[0] || '')
const rangeEnd = computed(() => bucketLabels.value[bucketLabels.value.length - 1] || '')

const totals = computed(() => {
  const out = {}
  for (const s of TREND_SERIES) {
    out[s.key] = (trendsDetail.value[s.key] || []).reduce((sum, d) => sum + (d.count || 0), 0)
  }
  return out
})

/* hex → rgba：面积渐变用同一品牌色，只调透明度 */
const tint = (hex, alpha = 0.08) => {
  const h = String(hex).replace('#', '')
  const n = parseInt(h.length === 3 ? h.split('').map(c => c + c).join('') : h, 16)
  return 'rgba(' + ((n >> 16) & 255) + ',' + ((n >> 8) & 255) + ',' + (n & 255) + ',' + alpha + ')'
}

/* 单指标面积折线：一次一条曲线，纵轴按当前指标自适应，读得清 */
const trendOption = computed(() => {
  const m = TREND_SERIES.find(s => s.key === activeMetric.value) || TREND_SERIES[0]
  const data = (trendsDetail.value[m.key] || []).map(d => d.count || 0)
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'line', lineStyle: { color: '#D0D5DD' } } },
    grid: { left: 4, right: 16, top: 18, bottom: 4, containLabel: true },
    xAxis: {
      type: 'category', boundaryGap: false, data: bucketLabels.value,
      axisLine: { lineStyle: { color: '#E4E7EC' } }, axisTick: { show: false },
      axisLabel: { color: '#667085', fontSize: 11, hideOverlap: true }
    },
    yAxis: {
      type: 'value', minInterval: 1, min: 0,
      axisLine: { show: false }, axisTick: { show: false },
      axisLabel: { color: '#667085', fontSize: 11 },
      splitLine: { lineStyle: { color: '#F2F3F5' } }
    },
    series: [{
      name: m.name,
      type: 'line',
      data,
      smooth: false,
      symbol: 'circle',
      symbolSize: 6,
      showSymbol: data.length <= 31,
      itemStyle: { color: m.color, borderColor: '#fff', borderWidth: 2 },
      lineStyle: { width: 2.5, color: m.color },
      areaStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [{ offset: 0, color: tint(m.color, 0.3) }, { offset: 1, color: tint(m.color, 0.02) }]
        }
      }
    }]
  }
})

// 模块数据：横向圆角柱
const barChartOption = computed(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: '3%', right: '6%', bottom: '3%', top: 10, containLabel: true },
  xAxis: { type: 'value', splitLine: { lineStyle: { color: '#F2F3F5' } }, axisLabel: { color: '#667085', fontSize: 11 } },
  yAxis: { type: 'category', data: ['平台用户', '内容帖子', '待审企业', '需求总数'], axisLine: { show: false }, axisTick: { show: false }, axisLabel: { color: '#667085', fontSize: 11 } },
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
    radar: {
      indicator, radius: '62%',
      axisName: { color: '#667085', fontSize: 11 },
      splitLine: { lineStyle: { color: '#F2F3F5' } },
      splitArea: { show: false },
      axisLine: { lineStyle: { color: '#F2F3F5' } }
    },
    series: [
      {
        type: 'radar',
        data: [
          {
            value: top.length ? top.map(([, v]) => Number(v) || 0) : [0],
            name: '需求类型',
            itemStyle: { color: '#5B8FF9' },
            areaStyle: { opacity: 0.12, color: '#5B8FF9' }
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
  legend: { bottom: 0, icon: 'circle', textStyle: { color: '#667085', fontSize: 11 } },
  series: [{
    type: 'pie',
    radius: ['60%', '80%'],
    label: { show: true, formatter: '{d}%', color: '#344054', fontSize: 11 },
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
/* 主区高度 = 右栏两卡自然高度之和（2×224 + 16 间距），主卡图表再吃掉剩余空间 */
.main-row {
  min-height: 464px;
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
.main-col {
  display: flex;
}
.main-card {
  display: flex;
  flex-direction: column;
  width: 100%;
}
.main-card :deep(.arco-card-body) {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  gap: 12px;
}
.card-extra {
  font-size: 12px;
  color: var(--color-text-3);
}
/* 指标切换 */
.metric-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.metric-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border: 1px solid var(--color-border-2);
  border-radius: 999px;
  background: var(--color-bg-2);
  color: var(--color-text-2);
  font-size: 13px;
  line-height: 20px;
  cursor: pointer;
  transition: border-color 0.16s ease, color 0.16s ease, background 0.16s ease;
}
.metric-tab:hover {
  border-color: var(--color-border-3);
}
.metric-tab-name {
  font-weight: 500;
}
.metric-tab-total {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.trend-chart {
  flex: 1;
  min-height: 320px;
}
.side-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.side-card :deep(.arco-card-body) {
  padding: 12px 16px 16px;
}
/* 固定高度：百分比高度在 flex 容器里可能在挂载时解析为 0，导致 ECharts 初始化出 0 高画布 */
.side-chart {
  height: 150px;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
</style>
