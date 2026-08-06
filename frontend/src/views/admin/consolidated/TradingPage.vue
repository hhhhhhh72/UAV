<template>
  <div class="admin-page">
    <div class="page-header"><h2>交易管理</h2></div>
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="需求管理" key="demands"><DemandList /></a-tab-pane>
      <a-tab-pane title="订单管理" key="orders"><OrderList /></a-tab-pane>
      <a-tab-pane title="评价管理" key="reviews"><ReviewList /></a-tab-pane>
      <a-tab-pane title="商品管理" key="products"><ProductList /></a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DemandList from '../demands/DemandList.vue'
import OrderList from '../orders/OrderList.vue'
import ReviewList from '../reviews/ReviewList.vue'
import ProductList from '../products/ProductList.vue'

const tab = ref('demands')
</script>

<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
