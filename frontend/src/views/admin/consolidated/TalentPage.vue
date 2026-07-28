<template>
  <div class="admin-page"><div class="page-header"><h2>人才教育</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="培训课程" name="training"><CourseList /></el-tab-pane>
      <el-tab-pane label="证书管理" name="certs"><CertList /></el-tab-pane>
      <el-tab-pane label="赛事管理" name="competition"><CompetitionList /></el-tab-pane>
      <el-tab-pane label="职位管理" name="jobs"><JobList /></el-tab-pane>
      <el-tab-pane label="院校管理" name="colleges"><CollegeList /></el-tab-pane>
      <el-tab-pane label="研学管理" name="study"><StudyList /></el-tab-pane>
    </el-tabs>
  </div>
</template><script setup>import { ref } from 'vue'; import CourseList from '../training/CourseList.vue'; import CertList from '../training/CertList.vue'; import CompetitionList from '../competition/CompetitionList.vue'; import JobList from '../jobs/JobList.vue'; import CollegeList from '../colleges/CollegeList.vue'; import StudyList from '../study/StudyList.vue'; const tab = ref('training')</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
