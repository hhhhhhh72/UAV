<template>
  <div class="admin-page"><div class="page-header"><h2>内容审核</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="案例管理" name="cases"><CaseList /></el-tab-pane>
      <el-tab-pane label="合规文档" name="compliance"><ComplianceList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script>import CaseList from '../cases/CaseList.vue'; import ComplianceList from '../compliance/ComplianceList.vue'; export default {components:{CaseList,ComplianceList}}</script>
<script setup>var tab = $ref('cases')</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
