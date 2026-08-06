<template>
  <div class="admin-page">
    <div class="page-header"><h2>运营推广</h2></div>
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="活动管理" key="events"><EventList /></a-tab-pane>
      <a-tab-pane title="品牌管理" key="portfolios"><PortfolioList /></a-tab-pane>
      <a-tab-pane title="展会管理" key="exhibitions"><ExhibitionList /></a-tab-pane>
      <a-tab-pane title="报告管理" key="reports"><ReportList /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import EventList from '../events/EventList.vue'
import PortfolioList from '../portfolios/PortfolioList.vue'
import ExhibitionList from '../exhibitions/ExhibitionList.vue'
import ReportList from '../reports/ReportList.vue'
const tab = ref('events')
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
