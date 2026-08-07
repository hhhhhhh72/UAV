<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="reviews"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      @sorter-change="handleSorterChange"
    >
      <template #target="{ record }">
        <a-tag :color="sectionTagColor(record.target_type)" size="small">{{ targetTypeLabel(record.target_type) }}</a-tag>
        <span class="target-id">{{ record.target_id || '-' }}</span>
      </template>
      <template #rating="{ record }">
        <span class="stars">{{ '★'.repeat(record.rating || 0) }}{{ '☆'.repeat(5 - (record.rating || 0)) }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTagColor(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <template v-if="record.status === 'pending'">
            <a-button type="text" status="success" size="small" @click="handleStatus(record, 'approved')">通过</a-button>
            <a-button type="text" status="warning" size="small" @click="handleStatus(record, 'rejected')">拒绝</a-button>
          </template>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无评价数据" />
      </template>
    </CrudList>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useAdminApi } from '@/api/admin/common'
import { updateReviewStatus } from '@/api/admin/review'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('reviews')

const targetTypeLabel = (t) => ({ demand: '需求', product: '商品', shop: '商家', job: '职位', course: '课程', venue: '场地' }[t] || t || '通用')
const statusLabel = (s) => ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' }[s] || s)
const sectionTagColor = (t) => ({ demand: 'arcoblue', product: 'green', shop: 'orange', job: 'arcoblue', course: 'arcoblue', venue: 'gray' }[t] || 'gray')
const statusTagColor = (s) => ({ pending: 'orange', approved: 'green', rejected: 'red' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 批量动作：批量通过 / 批量拒绝（专用端点 /api/v1/admin/reviews/{id}/approve|reject）
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: (row) => updateReviewStatus(row.id, 'approved') },
  { key: 'reject', label: '批量驳回', status: 'danger', api: (row) => updateReviewStatus(row.id, 'rejected') }
]

const searchFields = [
  { key: 'status', label: '状态', type: 'select', width: 130, options: [
    { value: '', label: '全部状态' },
    { value: 'pending', label: '待审核' },
    { value: 'approved', label: '已通过' },
    { value: 'rejected', label: '已拒绝' }
  ]}
]

const columns = [
  { title: '评价人ID', dataIndex: 'reviewer_id', width: 140, tooltip: true },
  { title: '评价对象', dataIndex: 'target_id', slotName: 'target', minWidth: 160 },
  { title: '评分', dataIndex: 'rating', slotName: 'rating', width: 100 },
  { title: '评价内容', dataIndex: 'content', minWidth: 200, tooltip: true },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '评价时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160, sortable: true },
  { title: '操作', slotName: 'actions', width: 180, fixed: 'right' },
]

// a-table 排序（Arco direction → useListRequest 的 order 语义）
const handleSorterChange = (dataIndex, direction) => {
  const order = direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  crudRef.value?.onSortChange({ prop: dataIndex, order })
}

const handleStatus = async (item, status) => {
  try {
    await updateReviewStatus(item.id, status)
    item.status = status
    Message.success(status === 'approved' ? '已通过' : '已拒绝')
  } catch (e) { Message.error('操作失败') }
}

const handleDelete = (item) => {
  Modal.confirm({
    title: '确认删除',
    content: '删除后不可恢复',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(item.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error('删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }

.target-id { margin-left: 6px; color: #86909C; }

.stars { color: #ffd21e; font-size: 14px; letter-spacing: 1px; }
.time-text { color: #86909C; font-size: 12px; }
</style>
