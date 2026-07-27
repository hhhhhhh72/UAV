<template>
  <div class="admin-page"><div class="page-header"><h2>交易管理</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="需求管理" name="demands"><DemandList /></el-tab-pane>
      <el-tab-pane label="订单管理" name="orders"><OrderList /></el-tab-pane>
      <el-tab-pane label="评价管理" name="reviews"><ReviewList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import DemandList from '../demands/DemandList.vue'
import OrderList from '../orders/OrderList.vue'
import ReviewList from '../reviews/ReviewList.vue'
const tab = ref('demands')
</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
