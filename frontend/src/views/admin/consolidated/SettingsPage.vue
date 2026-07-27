<template>
  <div class="admin-page"><div class="page-header"><h2>系统设置</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="服务配置" name="config"><ServiceConfigList /></el-tab-pane>
      <el-tab-pane label="消息通知" name="messages"><NotifyList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script>import ServiceConfigList from '../config/ServiceConfigList.vue'; import NotifyList from '../messages/NotifyList.vue'; export default {components:{ServiceConfigList,NotifyList}}</script>
<script setup>var tab = $ref('config')</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
