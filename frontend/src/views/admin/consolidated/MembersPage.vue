<template>
  <div class="admin-page">
    <div class="page-header"><h2>会员管理</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="用户管理" name="users"><UserList /></el-tab-pane>
      <el-tab-pane label="企业管理" name="enterprises"><EnterpriseList /></el-tab-pane>
      <el-tab-pane label="商家管理" name="shops"><ShopList /></el-tab-pane>
      <el-tab-pane label="专家管理" name="experts"><ExpertList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup>
var tab = $ref('users')
</script>
<script>
import UserList from '../users/UserList.vue'
import EnterpriseList from '../enterprises/EnterpriseList.vue'
import ShopList from '../shops/ShopList.vue'
import ExpertList from '../experts/ExpertList.vue'
export default { components: { UserList, EnterpriseList, ShopList, ExpertList } }
</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}.page-header h2{margin:0;font-size:20px}</style>
