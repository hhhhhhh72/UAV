<template>
  <div class="admin-page">
    <div class="page-header"><h2>系统设置</h2></div>
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="服务配置" key="config"><ServiceConfigList v-if="tab === 'config'" /></a-tab-pane>
      <a-tab-pane title="消息通知" key="messages"><NotifyList v-if="tab === 'messages'" /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import ServiceConfigList from '../config/ServiceConfigList.vue'
import NotifyList from '../messages/NotifyList.vue'
const tab = ref('config')
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
