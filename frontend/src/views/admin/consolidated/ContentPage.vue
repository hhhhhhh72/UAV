<template>
  <div class="admin-page">
    <div class="page-header"><h2>内容审核</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="案例管理" key="cases"><CaseList /></a-tab-pane>
      <a-tab-pane title="合规文档" key="compliance"><ComplianceList /></a-tab-pane>
      <a-tab-pane title="资源台账" key="resources"><ResourceList /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import CaseList from '../cases/CaseList.vue'
import ComplianceList from '../compliance/ComplianceList.vue'
import ResourceList from '../resources/ResourceList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('cases')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '企业案例', path: 'modules.industry.cases' },
    { label: '资源台账', path: 'modules.industry.industry_resources' },
    { label: '专家智库', path: 'modules.industry.experts' },
    { label: '品牌展示', path: 'modules.industry.portfolios' }
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
