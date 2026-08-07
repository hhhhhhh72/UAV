<template>
  <div class="admin-page">
    <div class="page-header"><h2>会员管理</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="用户管理" key="users"><UserList /></a-tab-pane>
      <a-tab-pane title="企业管理" key="enterprises"><EnterpriseList /></a-tab-pane>
      <a-tab-pane title="商家管理" key="shops"><ShopList /></a-tab-pane>
      <a-tab-pane title="专家管理" key="experts"><ExpertList /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import UserList from '../users/UserList.vue'
import EnterpriseList from '../enterprises/EnterpriseList.vue'
import ShopList from '../shops/ShopList.vue'
import ExpertList from '../experts/ExpertList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('users')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '平台用户', path: 'total_users' },
    { label: '待审企业', path: 'pending_enterprises' },
    { label: '认证专家', path: 'modules.industry.experts' },
    { label: '需求总数', path: 'total_demands' }
  ],
  trend: { title: '近 12 月用户增长', key: 'user' },
  pie: null
}
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
