<template>
  <div class="admin-page">
    <div class="page-header"><h2>运营推广</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="活动管理" key="events"><EventList v-if="tab === 'events'" /></a-tab-pane>
      <a-tab-pane title="品牌管理" key="portfolios"><PortfolioList v-if="tab === 'portfolios'" /></a-tab-pane>
      <a-tab-pane title="展会管理" key="exhibitions"><ExhibitionList v-if="tab === 'exhibitions'" /></a-tab-pane>
      <a-tab-pane title="报告管理" key="reports"><ReportList v-if="tab === 'reports'" /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import EventList from '../events/EventList.vue'
import PortfolioList from '../portfolios/PortfolioList.vue'
import ExhibitionList from '../exhibitions/ExhibitionList.vue'
import ReportList from '../reports/ReportList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('events')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '活动总数', path: 'modules.events.events' },
    { label: '赛事总数', path: 'modules.events.competitions' },
    { label: '展会排期', path: 'modules.events.exhibitions' },
    { label: '行业报告', path: 'modules.industry.industry_reports' }
  ],
  trend: { title: '近 12 月社区内容', key: 'post' },
  pie: null
}
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
