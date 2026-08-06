<template>
  <div class="page">
    <!-- 搜索过滤区 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索企业名称..." allow-clear style="width: 240px" @press-enter="onSearchSubmit">
              <template #prefix><icon-search /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="审核状态" class="form-item">
            <a-select v-model="filterParams.status" placeholder="审核状态" style="width: 150px" @change="onSearchSubmit">
              <a-option value="all">全部状态</a-option>
              <a-option value="submitted">待审核</a-option>
              <a-option value="approved">已通过</a-option>
              <a-option value="rejected">已驳回</a-option>
              <a-option value="draft">草稿</a-option>
              <a-option value="supplement_required">需补充</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
        </a-space>
      </a-form>

      <!-- 批量操作栏 -->
      <div class="batch-bar" v-if="selectedIds.length > 0">
        <span class="batch-info">已选择 <b>{{ selectedIds.length }}</b> 项</span>
        <a-button type="primary" status="success" size="small" @click="batchReview('approved')"><template #icon><icon-check /></template>批量通过</a-button>
        <a-button type="primary" status="danger" size="small" @click="batchReview('rejected')"><template #icon><icon-close /></template>批量驳回</a-button>
      </div>
    </a-card>

    <!-- 数据表格 -->
    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data="listData"
        :loading="loading"
        row-key="id"
        :pagination="false"
        :row-selection="rowSelection"
        @sorter-change="handleSortChange"
      >
        <template #name="{ record }">
          <span class="cell-name">{{ record.name || '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTagType(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <template v-if="record.status === 'submitted'">
              <a-divider direction="vertical" />
              <a-button type="text" status="success" size="small" @click="handleReview(record, 'approved')">通过</a-button>
              <a-button type="text" status="danger" size="small" @click="handleReview(record, 'rejected')">驳回</a-button>
            </template>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无企业数据" />
        </template>
      </a-table>

      <div class="pagination-wrap" v-if="total > 0">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50, 100]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="企业详情" :width="640" :footer="false" :mask-closable="false">
      <template v-if="currentEnterprise">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="企业名称" :span="2">{{ currentEnterprise.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="对公账户">{{ currentEnterprise.account_name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="法定代表人">{{ currentEnterprise.legal_person || '-' }}</a-descriptions-item>
          <a-descriptions-item label="联系电话">{{ currentEnterprise.contact_phone || '-' }}</a-descriptions-item>
          <a-descriptions-item label="信用代码" :span="2">{{ currentEnterprise.credit_code || '-' }}</a-descriptions-item>
          <a-descriptions-item label="产业分类">{{ currentEnterprise.industry_category || '-' }}</a-descriptions-item>
          <a-descriptions-item label="企业规模">{{ currentEnterprise.scale || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地址" :span="2">{{ currentEnterprise.address || '-' }}</a-descriptions-item>
          <a-descriptions-item label="企业简介" :span="2">{{ currentEnterprise.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="审核状态">
            <a-tag :color="statusTagType(currentEnterprise.status)" size="small">{{ statusLabel(currentEnterprise.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="协会会员">{{ currentEnterprise.is_member ? '是' : '否' }}</a-descriptions-item>
          <a-descriptions-item label="提交时间" :span="2">{{ formatDate(currentEnterprise.created_at) }}</a-descriptions-item>
          <a-descriptions-item v-if="currentEnterprise.license_url" label="营业执照" :span="2">
            <a-image
              :src="currentEnterprise.license_url"
              :width="200"
              :preview-props="{ srcList: [currentEnterprise.license_url] }"
              fit="contain"
              class="license-img"
            />
          </a-descriptions-item>
        </a-descriptions>

        <!-- 审核操作 -->
        <div v-if="currentEnterprise.status === 'submitted'" class="review-actions">
          <a-divider />
          <a-button type="primary" status="success" @click="handleReview(currentEnterprise, 'approved')">审核通过</a-button>
          <a-button type="primary" status="danger" @click="handleReview(currentEnterprise, 'rejected')">驳回</a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
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
  approved: 'green',
  rejected: 'red',
  submitted: 'orange',
  supplement_required: 'orange',
  draft: 'gray'
}[s] || 'gray')

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
  onBatchAction,
  resetParams
} = useListRequest({
  apiFunction: getEnterpriseList,
  idKey: 'id',
  defaultParams: { status: 'all' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

// Arco sorter-change → useListRequest.onSortChange（el-table 的 { prop, order } 形态）
const handleSortChange = (dataIndex, direction) => {
  onSortChange({
    prop: dataIndex,
    order: direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  })
}

const columns = [
  { title: '企业名称', dataIndex: 'name', slotName: 'name', minWidth: 180, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '对公账户', dataIndex: 'account_name', width: 160 },
  { title: '法人', dataIndex: 'legal_person', width: 100 },
  { title: '联系电话', dataIndex: 'contact_phone', width: 130 },
  { title: '审核状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '提交时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

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
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid #EEF1F4;
}

.batch-info {
  font-size: 13px;
  color: var(--color-text-2);
  margin-right: auto;
}

.cell-name { font-weight: 500; color: var(--color-text-1); }

.time-text { color: #86909C; font-size: 12px; }

.license-img { cursor: pointer; }

.review-actions {
  text-align: center;
  padding-top: 16px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
