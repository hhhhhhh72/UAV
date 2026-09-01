<template>
  <div class="admin-page">
    <div class="page-header"><h2>资讯中心</h2></div>
    <BizOverview :metrics="overview.metrics" :trend="overview.trend" :pie="overview.pie" />
    <a-tabs v-model:active-key="tab">
      <a-tab-pane title="资讯管理" key="articles"><ArticleList v-if="tab === 'articles'" /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import ArticleList from '../articles/ArticleList.vue'
import BizOverview from '../components/BizOverview.vue'

const tab = ref('articles')

/* 业务概览：数据来自 GET /api/v1/admin/dashboard（后端已补 articles 计数与 article 趋势） */
const overview = {
  metrics: [
    { label: '资讯总数', path: 'modules.industry.articles' },
    { label: '已发布', path: 'modules.industry.articles_published' },
    { label: '草稿', path: 'modules.industry.articles_draft' },
    { label: '置顶', path: 'modules.industry.articles_pinned' }
  ],
  trend: { title: '近 12 月资讯发布', key: 'article' },
  pie: null
}
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: var(--color-text-1); }
</style>
