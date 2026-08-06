<template>
  <div class="admin-page">
    <div class="page-header"><h2>人才教育</h2></div>
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="培训课程" key="training"><CourseList /></a-tab-pane>
      <a-tab-pane title="证书管理" key="certs"><CertList /></a-tab-pane>
      <a-tab-pane title="赛事管理" key="competition"><CompetitionList /></a-tab-pane>
      <a-tab-pane title="职位管理" key="jobs"><JobList /></a-tab-pane>
      <a-tab-pane title="院校管理" key="colleges"><CollegeList /></a-tab-pane>
      <a-tab-pane title="研学管理" key="study"><StudyList /></a-tab-pane>
      <a-tab-pane title="飞手认证" key="pilots"><PilotList /></a-tab-pane>
      <a-tab-pane title="报名记录" key="enrollments"><EnrollmentList /></a-tab-pane>
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
const tab = ref('training')
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
