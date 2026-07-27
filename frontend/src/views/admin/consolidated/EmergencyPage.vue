<template>
  <div class="admin-page"><div class="page-header"><h2>应急协同</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="应急资源" name="resources"><EmergencyResourceList /></el-tab-pane>
      <el-tab-pane label="应急调度" name="dispatches"><DispatchList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script>import EmergencyResourceList from '../emergency/ResourceList.vue'; import DispatchList from '../emergency/DispatchList.vue'; export default {components:{EmergencyResourceList,DispatchList}}</script>
<script setup>
import { ref } from 'vue'
const tab = ref('resources')
</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
