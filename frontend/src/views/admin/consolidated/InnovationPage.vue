<template>
  <div class="admin-page">
    <div class="page-header"><h2>产学研</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="成果管理" key="achievements"><AchievementList /></a-tab-pane>
      <a-tab-pane title="研发难题" key="challenges"><ChallengeList /></a-tab-pane>
      <a-tab-pane title="课题攻关" key="projects"><ProjectList /></a-tab-pane>
      <a-tab-pane title="测试场地" key="testsites"><TestSiteList /></a-tab-pane>
      <a-tab-pane title="场地预约" key="bookings"><BookingList /></a-tab-pane>
      <a-tab-pane title="成果转化" key="transformations"><TransList /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import AchievementList from '../achievements/AchievementList.vue'
import ChallengeList from '../challenges/ChallengeList.vue'
import ProjectList from '../projects/ProjectList.vue'
import TestSiteList from '../testsites/TestSiteList.vue'
import TransList from '../transformations/TransList.vue'
import BookingList from '../bookings/BookingList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('achievements')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '科技成果', path: 'modules.industry.achievements' },
    { label: '研发难题', path: 'modules.industry.rd_challenges' },
    { label: '课题攻关', path: 'modules.industry.research_projects' },
    { label: '测试场地', path: 'modules.industry.test_sites' }
  ],
  trend: null,
  pie: { title: '需求状态分布', key: 'status_dist' }
}
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
