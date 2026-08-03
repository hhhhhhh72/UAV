<template>
  <div class="enterprise-list-page">
    <!-- 搜索过滤区 -->
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索企业名称..."
          clearable
          style="width: 240px"
          @keyup.enter="onSearchSubmit"
          @clear="onSearchSubmit"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-select
          v-model="filterParams.status"
          placeholder="审核状态"
          clearable
          style="width: 150px"
          @change="onSearchSubmit"
        >
          <el-option label="待审核" value="submitted" />
          <el-option label="已通过" value="approved" />
          <el-option label="已驳回" value="rejected" />
          <el-option label="草稿" value="draft" />
          <el-option label="需补充" value="supplement_required" />
        </el-select>

        <el-button type="primary" @click="onSearchSubmit">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <!-- 批量操作栏 -->
    <div class="batch-bar" v-if="selectedIds.length > 0">
      <span class="batch-info">已选择 <b>{{ selectedIds.length }}</b> 项</span>
      <el-button type="success" :icon="Check" @click="batchReview('approved')">批量通过</el-button>
      <el-button type="danger" :icon="CloseBold" @click="batchReview('rejected')">批量驳回</el-button>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrap">
      <el-table
        v-loading="loading"
        :data="listData"
        row-key="id"
        stripe
        border
        style="width: 100%"
        @selection-change="onSelectChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="40" />

        <el-table-column prop="name" label="企业名称" min-width="180" sortable="custom">
          <template #default="{ row }">
            <span class="cell-name">{{ row.name || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="account_name" label="对公账户" width="160" />
        <el-table-column prop="contact_person" label="联系人" width="100" />
        <el-table-column prop="contact_phone" label="联系电话" width="130" />

        <el-table-column prop="status" label="审核状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="提交时间" width="160" sortable="custom">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <template v-if="row.status === 'submitted'">
              <el-divider direction="vertical" />
              <el-button link type="success" size="small" @click="handleReview(row, 'approved')">通过</el-button>
              <el-button link type="danger" size="small" @click="handleReview(row, 'rejected')">驳回</el-button>
            </template>
          </template>
        </el-table-column>

        <template #empty>
          <el-empty description="暂无企业数据" />
        </template>
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
    <el-dialog
      v-model="detailVisible"
      title="企业详情"
      width="640px"
      :close-on-click-modal="false"
    >
      <template v-if="currentEnterprise">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="企业名称" :span="2">{{ currentEnterprise.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="对公账户">{{ currentEnterprise.account_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ currentEnterprise.contact_person || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ currentEnterprise.contact_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核状态">
            <el-tag :type="statusTagType(currentEnterprise.status)" size="small">{{ statusLabel(currentEnterprise.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="协会会员">{{ currentEnterprise.is_member ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatDate(currentEnterprise.created_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentEnterprise.license_url" label="营业执照" :span="2">
            <el-image
              :src="currentEnterprise.license_url"
              style="width: 200px; cursor: pointer"
              :preview-src-list="[currentEnterprise.license_url]"
              fit="contain"
            />
          </el-descriptions-item>
        </el-descriptions>

        <!-- 审核操作 -->
        <div v-if="currentEnterprise.status === 'submitted'" class="review-actions">
          <el-divider />
          <el-button type="success" @click="handleReview(currentEnterprise, 'approved')">审核通过</el-button>
          <el-button type="danger" @click="handleReview(currentEnterprise, 'rejected')">驳回</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search, Check, CloseBold } from '@element-plus/icons-vue'
import { showSuccessToast, showFailToast } from '@/utils/feedback'
import { useListRequest } from '@/hooks/useListRequest'
import { getEnterpriseList, reviewEnterprise, batchReviewEnterprise } from '@/api/admin/enterprise'

// --- 状态映射 ---
const statusLabel = (s) => ({
  draft: '草稿',
  submitted: '待审核',
  supplement_required: '需补充',
  approved: '已通过',
  rejected: '已驳回'
}[s] || s || '-')

const statusTagType = (s) => ({
  approved: 'success',
  rejected: 'danger',
  submitted: 'warning',
  supplement_required: 'warning',
  draft: 'info'
}[s] || 'info')

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// --- 列表数据 ---
const {
  listData,
  loading,
  total,
  selectedIds,
  filterParams,
  loadData,
  onSearchSubmit,
  onSortChange,
  onSelectChange,
  onPageChange,
  onBatchAction,
  resetParams
} = useListRequest({
  apiFunction: getEnterpriseList,
  idKey: 'id',
  defaultParams: { status: 'submitted' }
})

// --- 详情弹窗 ---
const detailVisible = ref(false)
const currentEnterprise = ref(null)

const showDetail = (ent) => {
  currentEnterprise.value = { ...ent }
  detailVisible.value = true
}

// --- 审核操作 ---
const handleReview = async (ent, action) => {
  try {
    await reviewEnterprise(ent.id, { action, reason: '' })
    showSuccessToast(action === 'approved' ? '审核通过' : '已驳回')
    if (currentEnterprise.value) {
      currentEnterprise.value.status = action === 'approved' ? 'approved' : 'rejected'
    }
    detailVisible.value = false
    loadData()
  } catch (error) {
    showFailToast(error?.response?.data?.message || error?.message || '操作失败')
  }
}

const batchReview = (action) => {
  onBatchAction(action, batchReviewEnterprise)
}

onMounted(loadData)
</script>

<style scoped>
.enterprise-list-page {
  max-width: 1400px;
  margin: 0 auto;
}

.search-bar {
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}

.search-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.batch-bar {
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-5);
  border-radius: 8px;
  padding: 10px 16px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.batch-info {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-right: auto;
}

.table-wrap {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  overflow: hidden;
}

.cell-name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}

.review-actions {
  text-align: center;
  padding-top: 16px;
}

@media (max-width: 767px) {
  .search-bar { padding: 12px; }
  .search-row { flex-direction: column; align-items: stretch; }
  .table-wrap { overflow-x: auto; }
}
</style>
