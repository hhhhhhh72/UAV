<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索成果名称..."
          clearable
          style="width: 220px"
          @keyup.enter="onSearchSubmit"
          @clear="onSearchSubmit"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-select
          v-model="filterParams.stage"
          placeholder="阶段"
          clearable
          style="width: 140px"
          @change="onSearchSubmit"
        >
          <el-option label="全部" value="" />
          <el-option label="实验室" value="lab" />
          <el-option label="中试" value="pilot" />
          <el-option label="产业化" value="industrialization" />
          <el-option label="上市" value="launched" />
        </el-select>

        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left: auto">
          <el-button type="success" @click="handleAdd">新增</el-button>
        </div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table
        v-loading="loading"
        :data="listData"
        row-key="id"
        stripe
        border
        @selection-change="onSelectChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="achievement_title" label="成果名称" min-width="200">
          <template #default="{ row }">
            <span class="cell-name">{{ row.achievement_title || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="stage" label="当前阶段" width="100">
          <template #default="{ row }">
            <el-tag :type="stageTag(row.stage)" size="small">{{ stageLabel(row.stage) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="owner" label="负责人" width="100" />
        <el-table-column prop="start_date" label="开始日期" width="130" sortable="custom">
          <template #default="{ row }">
            {{ formatDate(row.start_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="target_date" label="目标完成" width="130" sortable="custom">
          <template #default="{ row }">
            {{ formatDate(row.target_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无转化数据" />
        </template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

    <el-dialog v-model="detailVisible" title="转化详情" width="640px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="成果名称" :span="2">{{ currentItem.achievement_title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="当前阶段">
            <el-tag :type="stageTag(currentItem.stage)" size="small">{{ stageLabel(currentItem.stage) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="负责人">{{ currentItem.owner || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</el-descriptions-item>
          <el-descriptions-item label="目标完成">{{ formatDate(currentItem.target_date) }}</el-descriptions-item>
          <el-descriptions-item label="进展记录" :span="2">{{ currentItem.progress_notes || '-' }}</el-descriptions-item>
          <el-descriptions-item label="里程碑" :span="2">{{ currentItem.milestones || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('transformations')

const stageLabel = (s) => ({ lab: '实验室', pilot: '中试', industrialization: '产业化', launched: '上市' }[s] || s || '-')
const stageTag = (s) => ({ lab: 'info', pilot: 'warning', industrialization: 'success', launched: '' }[s] || 'info')
const statusTag = (s) => ({}[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { stage: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
const handleAdd = () => { ElMessage.info('TODO: 新增转化记录表单') }
const handleEdit = (row) => { ElMessage.info('TODO: 编辑转化记录 ' + row.id) }
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除转化记录 "${row.achievement_title}" 吗？`, '提示', { type: 'warning' }).then(async () => {
    try { await api.delete(row.id); ElMessage.success('已删除'); loadData() } catch { ElMessage.error('删除失败') }
  }).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-name { font-weight: 500; color: var(--el-text-color-primary); }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
