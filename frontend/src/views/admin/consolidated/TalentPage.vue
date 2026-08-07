<template>
  <div class="admin-page">
    <div class="page-header"><h2>人才教育</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="培训课程" key="training"><CourseList v-if="tab === 'training'" /></a-tab-pane>
      <a-tab-pane title="证书管理" key="certs"><CertList v-if="tab === 'certs'" /></a-tab-pane>
      <a-tab-pane title="赛事管理" key="competition"><CompetitionList v-if="tab === 'competition'" /></a-tab-pane>
      <a-tab-pane title="职位管理" key="jobs"><JobList v-if="tab === 'jobs'" /></a-tab-pane>
      <a-tab-pane title="院校管理" key="colleges"><CollegeList v-if="tab === 'colleges'" /></a-tab-pane>
      <a-tab-pane title="研学管理" key="study"><StudyList v-if="tab === 'study'" /></a-tab-pane>
      <a-tab-pane title="飞手认证" key="pilots"><PilotList v-if="tab === 'pilots'" /></a-tab-pane>
      <a-tab-pane title="报名记录" key="enrollments"><EnrollmentList v-if="tab === 'enrollments'" /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import CourseList from '../training/CourseList.vue'
import CertList from '../training/CertList.vue'
import CompetitionList from '../competition/CompetitionList.vue'
import JobList from '../jobs/JobList.vue'
import CollegeList from '../colleges/CollegeList.vue'
import StudyList from '../study/StudyList.vue'
import PilotList from '../pilots/PilotList.vue'
import EnrollmentList from '../enrollments/EnrollmentList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('training')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard */
const overview = {
  metrics: [
    { label: '培训课程', path: 'modules.talent.training_courses' },
    { label: '培训证书', path: 'modules.talent.certificates' },
    { label: '合作院校', path: 'modules.talent.colleges' },
    { label: '招聘岗位', path: 'modules.talent.jobs' },
    { label: '研学线路', path: 'modules.talent.study_tours' }
  ],
  trend: { title: '近 12 月用户增长', key: 'user' },
  pie: null
}
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
