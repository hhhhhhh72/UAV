<template>
  <div class="admin-page"><div class="page-header"><h2>产学研</h2></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="成果管理" name="achievements"><AchievementList /></el-tab-pane>
      <el-tab-pane label="研发难题" name="challenges"><ChallengeList /></el-tab-pane>
      <el-tab-pane label="课题攻关" name="projects"><ProjectList /></el-tab-pane>
      <el-tab-pane label="测试场地" name="testsites"><TestSiteList /></el-tab-pane>
      <el-tab-pane label="成果转化" name="transformations"><TransList /></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script>import AchievementList from '../achievements/AchievementList.vue'; import ChallengeList from '../challenges/ChallengeList.vue'; import ProjectList from '../projects/ProjectList.vue'; import TestSiteList from '../testsites/TestSiteList.vue'; import TransList from '../transformations/TransList.vue'; export default {components:{AchievementList,ChallengeList,ProjectList,TestSiteList,TransList}}</script>
<script setup>
import { ref } from 'vue'
const tab = ref('achievements')
</script>
<style scoped>.admin-page{padding:20px}.page-header{margin-bottom:20px}h2{margin:0;font-size:20px}</style>
