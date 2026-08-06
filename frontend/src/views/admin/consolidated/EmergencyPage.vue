<template>
  <div class="admin-page">
    <div class="page-header"><h2>应急协同</h2></div>
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
const tab = ref('resources')
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
