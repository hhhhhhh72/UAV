<template>
  <div class="admin-page">
    <van-nav-bar
      title="后台管理"
      left-arrow
      @click-left="$router.push('/')"
      fixed
      placeholder
    >
      <template #right>
        <van-icon name="wap-home-o" size="18" @click="$router.push('/')" />
      </template>
    </van-nav-bar>

    <van-tabs v-model:active="activeTab" sticky offset-top="46">
      <van-tab title="企业审核">
        <EnterpriseReview />
      </van-tab>
      <van-tab title="需求审核">
        <DemandReview />
      </van-tab>
      <van-tab title="评价审核">
        <ReviewManage />
      </van-tab>
      <van-tab title="举报处理">
        <ReportManage />
      </van-tab>
      <van-tab title="用户管理">
        <UserManage />
      </van-tab>
      <van-tab title="平台配置">
        <ConfigManage />
      </van-tab>
    </van-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import EnterpriseReview from './enterprises/EnterpriseReview.vue'
import DemandReview from './demands/DemandReview.vue'
import ReviewManage from './reviews/ReviewManage.vue'
import ReportManage from './reports/ReportManage.vue'
import UserManage from './users/UserManage.vue'
import ConfigManage from './config/ConfigManage.vue'

const activeTab = ref(0)
</script>

<style scoped>
.admin-page {
  background-color: #f7f8fa;
  min-height: 100vh;
  padding-bottom: 20px;
}
</style>
