<template>
  <div class="biz-overview">
    <a-alert v-if="error" class="load-error" type="warning" closable @close="error = ''">
      统计数据加载失败，请稍后刷新重试
    </a-alert>
    <a-row :gutter="16" class="metric-row">
      <a-col v-for="m in metrics" :key="m.label" :span="Math.floor(24 / metrics.length)" :xs="12" :sm="12" :md="24 / metrics.length">
        <a-card class="metric-card" :bordered="false">
          <a-statistic :title="m.label" :value="getValue(m)" :precision="loading ? 0 : (m.precision || 0)" :value-style="{ fontWeight: 600 }">
            <template #suffix v-if="m.unit && !loading"><span class="metric-unit">{{ m.unit }}</span></template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>
    <a-row :gutter="16" v-if="trend || pie">
      <a-col :span="trend && pie ? 16 : 24" :xs="24" v-if="trend">
        <a-card class="chart-card" :bordered="false">
          <template #title>{{ trend.title }}</template>
          <div ref="trendRef" class="chart-box" v-show="!loading"></div>
          <a-skeleton v-if="loading" class="chart-skeleton" :animation="true" />
        </a-card>
      </a-col>
      <a-col :span="trend && pie ? 8 : 24" :xs="24" v-if="pie">
        <a-card class="chart-card" :bordered="false">
          <template #title>{{ pie.title }}</template>
          <div ref="pieRef" class="chart-box" v-show="!loading"></div>
          <a-skeleton v-if="loading" class="chart-skeleton" :animation="true" />
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, shallowRef, watch, nextTick } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import axios from '@/utils/http'

echarts.use([LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const props = defineProps({
  /** [{ label, path, unit, precision }] — path 为 dashboard 响应中的点号取值路径 */
  metrics: { type: Array, required: true },
  /** { title, key } — key 对应 dashboard.trends_detail */
  trend: { type: Object, default: null },
  /** { title, key } — key 对应 dashboard.category_dist / status_dist */
  pie: { type: Object, default: null }
})

const trendRef = ref(null)
const pieRef = ref(null)
const trendChart = shallowRef(null)
const pieChart = shallowRef(null)
const dashboard = ref(null)

const loading = ref(true)
const error = ref('')

const getValue = (m) => {
  if (loading.value) return '-'
  if (!dashboard.value) return '--'
  const v = m.path.split('.').reduce((o, k) => (o == null ? 0 : o[k]), dashboard.value) ?? 0
  return m.divide ? v / m.divide : v
}

const loadData = async () => {
  loading.value = true
  error.value = ''
  try {
    const d = await axios.get('/api/v1/admin/dashboard').then(r => r.data)
    dashboard.value = d
  } catch (e) {
    error.value = e?.message || '加载失败'
  } finally {
    loading.value = false
    // 等 v-show 容器可见后再初始化图表（隐藏时 init 会得到 0 尺寸）
    nextTick(renderCharts)
  }
}

const renderCharts = () => {
  if (!dashboard.value) return
  if (props.trend && trendRef.value) {
    const rows = dashboard.value.trends_detail?.[props.trend.key] || []
    trendChart.value?.dispose()
    trendChart.value = echarts.init(trendRef.value)
    trendChart.value.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 48, right: 24, top: 24, bottom: 28 },
      xAxis: { type: 'category', data: rows.map(r => (r.date || '').slice(5)) },
      yAxis: { type: 'value', minInterval: 1 },
      series: [{
        type: 'line', smooth: true, symbolSize: 6,
        data: rows.map(r => r.count),
        lineStyle: { width: 2, color: '#165DFF' },
        itemStyle: { color: '#165DFF' },
        areaStyle: { color: 'rgba(22,93,255,0.08)' }
      }]
    })
  }
  if (props.pie && pieRef.value) {
    const dist = dashboard.value[props.pie.key] || {}
    pieChart.value?.dispose()
    pieChart.value = echarts.init(pieRef.value)
    pieChart.value.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      legend: { bottom: 0, icon: 'circle', itemWidth: 8, itemHeight: 8 },
      series: [{
        type: 'pie', radius: ['40%', '65%'], center: ['50%', '45%'],
        label: { show: false },
        data: Object.entries(dist).filter(([, v]) => v > 0).map(([name, value]) => ({ name, value }))
      }]
    })
  }
}

const onResize = () => { trendChart.value?.resize(); pieChart.value?.resize() }

onMounted(() => {
  loadData()
  window.addEventListener('resize', onResize)
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  trendChart.value?.dispose()
  pieChart.value?.dispose()
})
watch(() => [props.trend, props.pie], renderCharts)
</script>

<style scoped>
.biz-overview {
  margin-bottom: 16px;
}

.load-error {
  margin-bottom: 12px;
}

.chart-skeleton {
  height: 260px;
}

.metric-row {
  margin-bottom: 16px;
}

.metric-card {
  border-radius: 4px;
}

.metric-card :deep(.arco-statistic-title) {
  color: var(--color-text-2);
  font-size: 13px;
}

.metric-unit {
  font-size: 13px;
  color: var(--color-text-3);
  font-weight: 400;
  margin-left: 2px;
}

.chart-card {
  border-radius: 4px;
}

.chart-card :deep(.arco-card-header) {
  font-size: 14px;
  font-weight: 600;
}

.chart-box {
  height: 260px;
}
</style>
