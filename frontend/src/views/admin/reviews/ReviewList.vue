<template>
  <div class="review-list-page">
    <!-- 搜索过滤区 -->
    <div class="search-bar">
      <div class="search-row">
        <el-select v-model="filterParams.status" clearable style="width: 130px" @change="onSearchSubmit">
          <el-option label="全部状态" value="" />
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>

        <el-select v-model="filterParams.section" clearable style="width: 130px" @change="onSearchSubmit">
          <el-option label="全部板块" value="" />
          <el-option label="研学" value="yanxue" />
          <el-option label="无人机销售" value="sale" />
          <el-option label="乐园" value="park" />
        </el-select>

        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索评价内容..."
          clearable style="width: 200px"
          @keyup.enter="onSearchSubmit" @clear="onSearchSubmit"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="reviewer_id" label="评价人ID" width="140" show-overflow-tooltip />
        <el-table-column label="评价对象" min-width="160">
          <template #default="{ row }">
            <el-tag :type="sectionTagType(row.target_type)" size="small">{{ targetTypeLabel(row.target_type) }}</el-tag>
            <span class="target-id">{{ row.target_id || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="评分" width="100">
          <template #default="{ row }">
            <span class="stars">{{ '★'.repeat(row.rating || 0) }}{{ '☆'.repeat(5 - (row.rating || 0)) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="评价内容" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="评价时间" width="160" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button link type="success" size="small" @click="handleStatus(row, 'approved')">通过</el-button>
              <el-button link type="warning" size="small" @click="handleStatus(row, 'rejected')">拒绝</el-button>
            </template>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>

        <template #empty><el-empty description="暂无评价数据" /></template>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50]"
        :total="total" layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData" @current-change="loadData"
      />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { showToast, showConfirmDialog } from '@/utils/feedback'
import { useListRequest } from '@/hooks/useListRequest'
import { getReviewList, updateReviewStatus, deleteReview } from '@/api/admin/review'

const targetTypeLabel = (t) => ({ demand: '需求', product: '商品', shop: '商家', job: '职位', course: '课程', venue: '场地' }[t] || t || '通用')
const statusLabel = (s) => ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' }[s] || s)
const sectionTagType = (t) => ({ demand: '', product: 'success', shop: 'warning', job: 'primary', course: '', venue: 'info' }[t] || 'info')
const statusTagType = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: getReviewList,
  idKey: 'id',
  defaultParams: { status: '', section: '', limit: 20 }
})

const handleStatus = async (item, status) => {
  try {
    await updateReviewStatus(item.id, status)
    item.status = status
    showToast(status === 'approved' ? '已通过' : '已拒绝')
  } catch (e) { showToast('操作失败') }
}

const handleDelete = (item) => {
  showConfirmDialog({ title: '确认删除', message: '删除后不可恢复' }).then(async () => {
    try {
      await deleteReview(item.id)
      listData.value = listData.value.filter(r => r.id !== item.id)
      showToast('已删除')
    } catch (e) { showToast('删除失败') }
  }).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.review-list-page { max-width: 1200px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.stars { color: #ffd21e; font-size: 14px; letter-spacing: 1px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
