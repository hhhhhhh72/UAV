<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="enterprises"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      :batch-actions="batchActions"
      :batch-delete="false"
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
    </CrudList>

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
import { ref } from 'vue'
import { showSuccessToast, showFailToast } from '@/utils/feedback'
import { reviewEnterprise } from '@/api/admin/enterprise'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const defaultParams = { status: 'all' }

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

// 批量动作：批量通过/批量驳回（传完整行数据，逐个调用审核接口）
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: (row) => reviewEnterprise(row.id, { action: 'approved', reason: '' }) },
  { key: 'reject', label: '批量驳回', status: 'danger', api: (row) => reviewEnterprise(row.id, { action: 'rejected', reason: '' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索企业名称...', width: 240 },
  { key: 'status', label: '审核状态', type: 'select', options: [
    { value: 'all', label: '全部状态' },
    { value: 'submitted', label: '待审核' },
    { value: 'approved', label: '已通过' },
    { value: 'rejected', label: '已驳回' },
    { value: 'draft', label: '草稿' },
    { value: 'supplement_required', label: '需补充' }
  ]}
]

const columns = [
  { title: '企业名称', dataIndex: 'name', slotName: 'name', minWidth: 180 },
  { title: '对公账户', dataIndex: 'account_name', width: 160 },
  { title: '法人', dataIndex: 'legal_person', width: 100 },
  { title: '联系电话', dataIndex: 'contact_phone', width: 130 },
  { title: '审核状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '提交时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
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
    crudRef.value?.reload()
  } catch (error) {
    showFailToast(error?.response?.data?.message || error?.message || '操作失败')
  }
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.cell-name { font-weight: 500; color: var(--color-text-1); }

.time-text { color: #86909C; font-size: 12px; }

.license-img { cursor: pointer; }

.review-actions {
  text-align: center;
  padding-top: 16px;
}
</style>
