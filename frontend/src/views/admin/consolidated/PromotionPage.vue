<template>
  <div class="admin-page"><div class="page-header"><h2>运营推广</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="活动管理" name="events"><EventList /></el-tab-pane>
      <el-tab-pane label="品牌管理" name="portfolios"><PortfolioList /></el-tab-pane>
      <el-tab-pane label="展会管理" name="exhibitions"><ExhibitionList /></el-tab-pane>
      <el-tab-pane label="报告管理" name="reports"><ReportList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script>import EventList from '../events/EventList.vue'; import PortfolioList from '../portfolios/PortfolioList.vue'; import ExhibitionList from '../exhibitions/ExhibitionList.vue'; import ReportList from '../reports/ReportList.vue'; export default {components:{EventList,PortfolioList,ExhibitionList,ReportList}}</script>
<script setup>
import { ref } from 'vue'
const tab = ref('events')
</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
