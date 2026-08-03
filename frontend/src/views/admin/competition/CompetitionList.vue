<template>
  <div class="competition-page">
    <div class="search-bar">
      <div class="search-row">
        <el-select v-model="filterParams.status" clearable style="width: 130px" @change="onSearchSubmit">
          <el-option label="全部状态" value="" />
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">查询</el-button>
        <el-button type="success" @click="showCreate">新建赛事</el-button>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="id" label="ID" width="180" />
        <el-table-column prop="title" label="赛事名称" min-width="160" />
        <el-table-column prop="category" label="类别" width="100" />
        <el-table-column prop="location" label="地点" width="120" />
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ formatDate(row.start_date) }}</template>
        </el-table-column>
        <el-table-column label="报名/名额" width="110">
          <template #default="{ row }">{{ row.reg_count || 0 }} / {{ row.max_teams || 0 }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无赛事" /></template>
      </el-table>
    </div>

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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="赛事详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="赛事名称" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类别">{{ currentItem.category || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地点">{{ currentItem.location || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开始">{{ formatDate(currentItem.start_date) }}</el-descriptions-item>
          <el-descriptions-item label="结束">{{ formatDate(currentItem.end_date) }}</el-descriptions-item>
          <el-descriptions-item label="报名/名额">{{ currentItem.reg_count || 0 }} / {{ currentItem.max_teams || 0 }}</el-descriptions-item>
          <el-descriptions-item label="主办方">{{ currentItem.sponsor || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="currentItem.description" label="简介" :span="2">{{ currentItem.description }}</el-descriptions-item>
        </el-descriptions>

        <div class="review-actions">
          <el-divider />
          <span style="margin-right: 12px;">修改状态：</span>
          <el-select v-model="newStatus" style="width: 140px;">
            <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
          <el-button type="primary" @click="onUpdateStatus">更新</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { showToast, showSuccessToast, showConfirmDialog } from 'vant'
import { useListRequest } from '@/hooks/useListRequest'
import { getCompetitionList, updateCompetition, deleteCompetition } from '@/api/admin/competition'

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '报名中', value: 'open' },
  { label: '已结束', value: 'closed' }
]
const statusLabel = (s) => statusOptions.find(o => o.value === s)?.label || s || '-'
const statusTagType = (s) => ({ open: 'success', closed: 'info', draft: 'warning' }[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, filterParams, loadData, onSearchSubmit } = useListRequest({
  apiFunction: getCompetitionList,
  idKey: 'id',
  defaultParams: { status: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)
const newStatus = ref('draft')

const showDetail = (item) => {
  currentItem.value = { ...item }
  newStatus.value = item.status || 'draft'
  detailVisible.value = true
}

const showCreate = () => {
  showToast('请使用小程序/直接调用 POST /api/v1/admin/competitions 创建赛事')
}

const onUpdateStatus = async () => {
  if (!currentItem.value) return
  try {
    await updateCompetition(currentItem.value.id, { status: newStatus.value })
    currentItem.value.status = newStatus.value
    showSuccessToast('状态已更新')
    loadData()
  } catch (e) { showToast('更新失败') }
}

const onDelete = async (row) => {
  try {
    await showConfirmDialog({ title: '删除赛事', message: `确定删除「${row.title}」吗？` })
  } catch (e) { return }
  try {
    await deleteCompetition(row.id)
    showSuccessToast('已删除')
    loadData()
  } catch (e) { showToast('删除失败') }
}

onMounted(loadData)
</script>

<style scoped>
.competition-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
