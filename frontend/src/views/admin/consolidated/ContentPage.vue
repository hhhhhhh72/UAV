<template>
  <div class="admin-page">
    <div class="page-header"><h2>内容审核</h2></div>
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
const tab = ref('cases')
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
