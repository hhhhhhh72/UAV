<template>
  <div class="demand-list-page">
    <!-- 搜索过滤区 -->
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索需求标题..."
          clearable
          style="width: 240px"
          @keyup.enter="onSearchSubmit"
          @clear="onSearchSubmit"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="filterParams.status" clearable style="width: 140px" @change="onSearchSubmit">
          <el-option label="全部" value="all" />
          <el-option label="待审核" value="pending" />
          <el-option label="已发布" value="published" />
          <el-option label="已匹配" value="matched" />
          <el-option label="已完成" value="completed" />
          <el-option label="已取消" value="cancelled" />
          <el-option label="已驳回" value="rejected" />
        </el-select>

        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <!-- 批量操作栏 -->
    <div class="batch-bar" v-if="selectedIds.length > 0">
      <span class="batch-info">已选择 <b>{{ selectedIds.length }}</b> 项</span>
      <el-button type="success" :icon="Check" @click="batchApprove">批量通过</el-button>
      <el-button type="danger" :icon="CloseBold" @click="batchReject">批量驳回</el-button>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrap">
      <el-table
        v-loading="loading"
        :data="listData"
        row-key="id"
        stripe border
        @selection-change="onSelectChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="40" />

        <el-table-column prop="title" label="需求标题" min-width="200" sortable="custom">
          <template #default="{ row }">
            <span class="cell-title">{{ row.title || '无标题' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="publisher_name" label="发布者" width="130" />
        <el-table-column prop="biz_type" label="业务类型" width="120">
          <template #default="{ row }">{{ bizTypeLabel(row.biz_type) }}</template>
        </el-table-column>

        <el-table-column prop="district" label="地区" width="110" />

        <el-table-column prop="budget_fen" label="预算" width="110" sortable="custom" align="right">
          <template #default="{ row }">
            {{ row.budget_fen ? '¥' + Number(row.budget_fen / 100).toLocaleString() : '-' }}
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="提交时间" width="160" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <template v-if="row.status === 'pending'">
              <el-divider direction="vertical" />
              <el-button link type="success" size="small" @click="handleApprove(row)">通过</el-button>
              <el-button link type="danger" size="small" @click="handleReject(row)">驳回</el-button>
            </template>
          </template>
        </el-table-column>

        <template #empty><el-empty description="暂无需求数据" /></template>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="需求详情" width="600px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发布者">{{ currentItem.publisher_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="业务类型">{{ bizTypeLabel(currentItem.biz_type) }}</el-descriptions-item>
          <el-descriptions-item label="地区">{{ currentItem.district || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="预算">{{ currentItem.budget_fen ? '¥' + (currentItem.budget_fen / 100).toLocaleString() : '-' }}</el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatDate(currentItem.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="currentItem.status === 'pending'" class="review-actions">
          <el-divider />
          <el-button type="success" @click="handleApprove(currentItem)">审核通过</el-button>
          <el-button type="danger" @click="handleReject(currentItem)">驳回</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search, Check, CloseBold } from '@element-plus/icons-vue'
import { showSuccessToast, showFailToast } from 'vant'
import { useListRequest } from '@/hooks/useListRequest'
import { getDemandList, approveDemand, rejectDemand } from '@/api/admin/demand'

const bizTypeLabel = (t) => ({
  aerial_photo: '航拍摄影', mapping: '测绘', inspection: '巡检',
  agriculture: '植保', logistics: '物流配送', training: '培训',
  competition: '赛事', other: '其他'
}[t] || t || '-')

const statusLabel = (s) => ({
  pending: '待审核', published: '已发布', matched: '已匹配',
  completed: '已完成', cancelled: '已取消', rejected: '已驳回'
}[s] || s || '-')

const statusTagType = (s) => ({
  published: 'success', matched: '', completed: 'success',
  pending: 'warning', cancelled: 'info', rejected: 'danger'
}[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: getDemandList,
  idKey: 'id',
  defaultParams: { status: 'pending' }
})

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (d) => { currentItem.value = d; detailVisible.value = true }

const handleApprove = async (item) => {
  try {
    await approveDemand(item.id)
    showSuccessToast('审核通过')
    item.status = 'published'
    detailVisible.value = false
    loadData()
  } catch (e) { showFailToast(e?.response?.data?.message || '操作失败') }
}

const handleReject = async (item) => {
  try {
    await rejectDemand(item.id, '')
    showSuccessToast('已驳回')
    item.status = 'rejected'
    detailVisible.value = false
    loadData()
  } catch (e) { showFailToast(e?.response?.data?.message || '操作失败') }
}

const batchApprove = () => {
  selectedIds.value.forEach(id => approveDemand(id).catch(() => {}))
  showSuccessToast('批量通过已提交')
  loadData()
}

const batchReject = () => {
  selectedIds.value.forEach(id => rejectDemand(id, '').catch(() => {}))
  showSuccessToast('批量驳回已提交')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.demand-list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.batch-bar { background: var(--el-color-primary-light-9); border: 1px solid var(--el-color-primary-light-5); border-radius: 8px; padding: 10px 16px; margin-bottom: 16px; display: flex; align-items: center; gap: 12px; }
.batch-info { font-size: 13px; color: var(--el-text-color-secondary); margin-right: auto; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-title { font-weight: 500; color: var(--el-text-color-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; max-width: 300px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.review-actions { text-align: center; padding-top: 16px; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
