<template>
  <div class="multi-dimension-container">
    <a-row :gutter="16" style="margin-bottom: 16px; display: flex; align-items: stretch;">
      <a-col :span="16">
        <a-card class="general-card" title="平台用户增长" :bordered="false" style="height: 100%;">
          <a-row :gutter="16" style="margin-bottom: 24px;">
            <a-col :span="6">
              <div class="overview-item">
                <div class="overview-title">需求总数</div>
                <div class="overview-value"><icon-list style="color: #5B8FF9;" /> {{ stats.totalDemands.toLocaleString() }}</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="overview-item">
                <div class="overview-title">待审企业</div>
                <div class="overview-value"><icon-edit style="color: #F6903D;" /> {{ stats.pendingEnterprises.toLocaleString() }}</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="overview-item">
                <div class="overview-title">内容帖子</div>
                <div class="overview-value"><icon-eye style="color: #78D3F8;" /> {{ stats.totalPosts.toLocaleString() }}</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="overview-item">
                <div class="overview-title">平台用户</div>
                <div class="overview-value"><icon-user style="color: #9270CA;" /> {{ stats.totalUsers.toLocaleString() }}</div>
              </div>
            </a-col>
          </a-row>
          <div style="height: 320px;">
            <v-chart :option="overviewLineOption" autoresize />
          </div>
        </a-card>
      </a-col>
      <a-col :span="8" style="display: flex; flex-direction: column; gap: 16px;">
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

const stats = ref({
  totalDemands: 0, pendingEnterprises: 0, totalPosts: 0,
  totalUsers: 0, pendingReports: 0,
  orderTrend: [], competitionByRole: [], userGrowth: [], statusDist: {}
})
const trendsDetail = ref({ demand: [], post: [], user: [], message: [] })

// 平台用户增长月度柱状图（与需求状态饼图不重复，数据用 trends_detail.user）
const overviewLineOption = computed(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: { type: 'category', data: (stats.value.orderTrend || []).map(d => (d.date || '').slice(5)) },
  yAxis: { type: 'value', minInterval: 1 },
  series: [{
    name: '平台用户',
    type: 'bar',
    barWidth: '40%',
    data: (trendsDetail.value.user || []).map(d => d.count || 0),
    itemStyle: { color: '#9270CA', borderRadius: [4, 4, 0, 0] }
  }]
}));

// 照抄 Arco Pro barChartOption：横向圆角柱
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

// 照抄 Arco Pro radarChartOption：雷达图
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

// 需求状态分布：多段环形饼（照抄 Arco Pro 环形饼样式：60-80% 半径 + {d}% 标签）
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
        competitionByRole: d.category_dist || [],
        userGrowth: d.trends || [],
        statusDist: d.status_dist || {}
      }
      trendsDetail.value = d.trends_detail || { demand: [], post: [], user: [], message: [] }
    }
  } catch (err) {
    showFailToast('获取统计数据失败')
  }
}

onMounted(fetchStats)
</script>

<style scoped>
/* 照抄 Arco Pro multi-dimension 全部样式 */
.multi-dimension-container {
  display: flex;
  flex-direction: column;
}
.general-card {
  border-radius: 4px;
}
.overview-item {
  display: flex;
  flex-direction: column;
}
.overview-title {
  color: var(--color-text-3);
  font-size: 14px;
}
.overview-value {
  color: var(--color-text-1);
  font-size: 24px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}
.kpi-card .kpi-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-1);
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.legend-box {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-top: 16px;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-2);
}
.legend-item .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
</style>
