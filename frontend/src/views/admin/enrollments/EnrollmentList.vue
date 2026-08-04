<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索姓名/电话..." clearable style="width: 200px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column prop="id_card" label="身份证号" width="190" />
        <el-table-column prop="course_id" label="课程ID" min-width="180">
          <template #default="{ row }">{{ row.course_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="gender" label="性别" width="70" />
        <el-table-column prop="education" label="学历" width="80" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enrolled' ? 'success' : 'info'" size="small">{{ row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="报名时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无报名记录" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>

    <!-- 报名详情（含证件资料） -->
    <el-dialog v-model="detailVisible" title="报名详情" width="640px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="姓名" :span="2">{{ currentItem.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ currentItem.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ currentItem.id_card || '-' }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ currentItem.gender || '-' }}</el-descriptions-item>
          <el-descriptions-item label="生日">{{ currentItem.birthday || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ currentItem.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="学历">{{ currentItem.education || '-' }}</el-descriptions-item>
          <el-descriptions-item label="从业经验" :span="2">{{ currentItem.experience || '-' }}</el-descriptions-item>
          <el-descriptions-item label="课程ID" :span="2">{{ currentItem.course_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="证件照" :span="1">
            <el-image v-if="currentItem.photo_url" :src="fullUrl(currentItem.photo_url)" :preview-src-list="[fullUrl(currentItem.photo_url)]" fit="cover" style="width: 64px; height: 64px; border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="身份证照片">
            <el-image v-if="currentItem.id_card_image" :src="fullUrl(currentItem.id_card_image)" :preview-src-list="[fullUrl(currentItem.id_card_image)]" fit="cover" style="width: 64px; height: 64px; border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="无犯罪证明" :span="2">{{ currentItem.no_crime || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('enrollments')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}
const fullUrl = (u) => (u && u.startsWith('http') ? u : (import.meta.env.VITE_API_TARGET || 'http://localhost:8080') + (u || ''))

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: {},
})

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
</style>
