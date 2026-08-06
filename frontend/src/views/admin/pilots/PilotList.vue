<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="certified-pilots"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
    >
      <template #certCount="{ record }">
        <span>{{ (record.cert_ids || []).length }} 项</span>
      </template>
      <template #bio="{ record }">
        <span>{{ record.bio || '-' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <template v-if="record.status === 'pending'">
            <a-button type="text" status="success" size="small" @click="handleApprove(record)">通过</a-button>
            <a-button type="text" status="danger" size="small" @click="handleReject(record)">驳回</a-button>
          </template>
          <template v-else>
            <a-button v-if="record.status === 'approved'" type="text" status="danger" size="small" @click="handleReject(record)">撤销</a-button>
            <a-button v-if="record.status === 'rejected'" type="text" status="success" size="small" @click="handleApprove(record)">恢复通过</a-button>
          </template>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无飞手申请" />
      </template>
    </CrudList>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import axios from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusLabel = { pending: '待审核', approved: '已认证', rejected: '已驳回' }
const statusTag = (s) => ({ pending: 'orangered', approved: 'green', rejected: 'red' }[s] || 'gray')

// 批量动作：批量通过 / 批量驳回——走专用审核端点
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: (row) => axios.post(`/api/v1/admin/certified-pilots/${row.id}/approve`) },
  { key: 'reject', label: '批量驳回', status: 'danger', api: (row) => axios.post(`/api/v1/admin/certified-pilots/${row.id}/reject`) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索姓名...', width: 200 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'pending', label: '待审核' },
    { value: 'approved', label: '已认证' },
    { value: 'rejected', label: '已驳回' }
  ]}
]

const columns = [
  { title: '姓名', dataIndex: 'real_name', minWidth: 100 },
  { title: '身份证号', dataIndex: 'id_card', width: 200 },
  { title: '证书', dataIndex: 'cert_ids', slotName: 'certCount', width: 80, align: 'center' },
  { title: '时长(h)', dataIndex: 'flight_hours', width: 80, align: 'center' },
  { title: '擅长领域', dataIndex: 'bio', slotName: 'bio', minWidth: 140 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '申请时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' }
]

// 审核：通过 / 驳回（专用端点）
const setStatus = async (row, action, tip) => {
  try {
    await axios.post(`/api/v1/admin/certified-pilots/${row.id}/${action}`)
    Message.success(tip)
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}
const handleApprove = (row) => setStatus(row, 'approve', '已通过，飞手进入公开名录')
const handleReject = (row) => {
  Modal.confirm({
    title: '驳回申请',
    content: `确定驳回 ${row.real_name} 的飞手认证申请？`,
    okText: '驳回',
    cancelText: '取消',
    onOk: () => setStatus(row, 'reject', '已驳回')
  })
}
</script>

<style scoped>
.time-text { color: #86909C; font-size: 12px; }
</style>
