<template>
  <div class="admin-page">
    <div class="page-header"><h2>交易管理</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="需求管理" key="demands"><DemandList v-if="tab === 'demands'" /></a-tab-pane>
      <a-tab-pane title="订单管理" key="orders"><OrderList v-if="tab === 'orders'" /></a-tab-pane>
      <a-tab-pane title="评价管理" key="reviews"><ReviewList v-if="tab === 'reviews'" /></a-tab-pane>
      <a-tab-pane title="商品管理" key="products"><ProductList v-if="tab === 'products'" /></a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DemandList from '../demands/DemandList.vue'
import OrderList from '../orders/OrderList.vue'
import ReviewList from '../reviews/ReviewList.vue'
import ProductList from '../products/ProductList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('demands')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '需求总数', path: 'total_demands' },
    { label: '线下成交额(万元)', path: 'offline_amount_total', divide: 1000000, precision: 2 },
    { label: '待审企业', path: 'pending_enterprises' },
    { label: '社区帖子', path: 'total_posts' }
  ],
  trend: { title: '近 12 月需求趋势', key: 'demand' },
  pie: { title: '需求类型分布', key: 'category_dist' }
}
</script>

<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
