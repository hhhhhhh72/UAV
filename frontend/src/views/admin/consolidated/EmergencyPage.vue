<template>
  <div class="admin-page">
    <div class="page-header"><h2>应急协同</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="应急资源" key="resources"><EmergencyResourceList /></a-tab-pane>
      <a-tab-pane title="应急调度" key="dispatches"><DispatchList /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import EmergencyResourceList from '../emergency/ResourceList.vue'
import DispatchList from '../emergency/DispatchList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('resources')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '应急资源', path: 'modules.events.emergency_resources' },
    { label: '调度记录', path: 'modules.events.emergency_dispatches' }
  ],
  trend: { title: '近 12 月需求趋势', key: 'demand' },
  pie: { title: '需求状态分布', key: 'status_dist' }
}
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
