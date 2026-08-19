<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="test-sites/bookings"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      :batch-delete="false"
    >
      <template #time="{ record }">
        <span class="time-text">{{ formatTime(record.start_time) }} ~ {{ formatTime(record.end_time).split(' ')[1] }}</span>
      </template>
      <template #purpose="{ record }">
        <span>{{ record.purpose || '-' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <template v-if="record.status === 'pending'">
            <a-button type="text" status="success" size="small" @click="review(record, 'approved')">通过</a-button>
            <a-button type="text" status="danger" size="small" @click="review(record, 'rejected')">驳回</a-button>
          </template>
          <template v-else>
            <a-button v-if="record.status === 'approved'" type="text" status="success" size="small" @click="review(record, 'completed')">完成</a-button>
          </template>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无预约记录" />
      </template>
    </CrudList>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import axios from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()

const formatTime = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusLabel = { pending: '待审核', approved: '已通过', rejected: '已驳回', completed: '已完成' }
const statusTag = (s) => ({ pending: 'orange', approved: 'green', rejected: 'red', completed: 'gray' }[s] || 'gray')

// 预约审核走专用端点 + 状态机按行流转，无合适的批量业务动作
const batchActions = []

const searchFields = [
  { key: 'keyword', label: '关键词', type: 'input', width: 220, placeholder: '搜索联系人/用途...' },
  { key: 'status', label: '状态', type: 'select', width: 140, placeholder: '审核状态', options: [
    { value: 'pending', label: '待审核' },
    { value: 'approved', label: '已通过' },
    { value: 'rejected', label: '已驳回' },
    { value: 'completed', label: '已完成' }
  ]}
]

const columns = [
  { title: '场地ID', dataIndex: 'site_id', minWidth: 180 },
  { title: '预约时间', dataIndex: 'start_time', slotName: 'time', width: 170 },
  { title: '联系人', dataIndex: 'contact_name', width: 100 },
  { title: '联系电话', dataIndex: 'contact_phone', width: 130 },
  { title: '用途', dataIndex: 'purpose', slotName: 'purpose', minWidth: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 120, fixed: 'right' },
]

// 审核：通过 / 驳回 / 完成（专用端点）
const review = async (row, status) => {
  try {
    await axios.post(`/api/v1/admin/test-sites/bookings/${row.id}/review`, { status, note: '' })
    Message.success({ approved: '已通过', rejected: '已驳回', completed: '已完成' }[status])
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.time-text { color: var(--color-text-2); font-size: 12px; }
</style>
