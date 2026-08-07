<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="competitions"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新建赛事"
      @add="showCreate()"
    >
      <template #startDate="{ record }">
        <span class="time-text">{{ formatDate(record.start_date) }}</span>
      </template>
      <template #regCount="{ record }">
        <span>{{ record.reg_count || 0 }} / {{ record.max_teams || 0 }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTagType(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无赛事" />
      </template>
    </CrudList>

    <!-- 详情弹窗（含修改状态区） -->
    <a-modal v-model:visible="detailVisible" title="赛事详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="赛事名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类别">{{ currentItem.category || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="报名/名额">{{ currentItem.reg_count || 0 }} / {{ currentItem.max_teams || 0 }}</a-descriptions-item>
          <a-descriptions-item label="主办方">{{ currentItem.sponsor || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item v-if="currentItem.description" label="简介" :span="2">{{ currentItem.description }}</a-descriptions-item>
        </a-descriptions>

        <div class="review-actions">
          <a-divider />
          <span style="margin-right: 12px;">修改状态：</span>
          <a-select v-model="newStatus" style="width: 140px;">
            <a-option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</a-option>
          </a-select>
          <a-button type="primary" @click="onUpdateStatus">更新</a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('competitions')

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '报名中', value: 'open' },
  { label: '已结束', value: 'closed' }
]
const statusLabel = (s) => statusOptions.find(o => o.value === s)?.label || s || '-'
const statusTagType = (s) => ({ open: 'green', closed: 'gray', draft: 'orangered' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 批量动作：开始报名 / 结束报名——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'open', label: '开始报名', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'open' }) },
  { key: 'close', label: '结束报名', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部状态' },
    ...statusOptions
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 180 },
  { title: '赛事名称', dataIndex: 'title', minWidth: 160 },
  { title: '类别', dataIndex: 'category', width: 100 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '开始时间', dataIndex: 'start_date', slotName: 'startDate', width: 170 },
  { title: '报名/名额', dataIndex: 'reg_count', slotName: 'regCount', width: 110 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const newStatus = ref('draft')

const showDetail = (item) => {
  currentItem.value = { ...item }
  newStatus.value = item.status || 'draft'
  detailVisible.value = true
}

const showCreate = () => {
  Message.info('请使用小程序/直接调用 POST /api/v1/admin/competitions 创建赛事')
}

const onUpdateStatus = async () => {
  if (!currentItem.value) return
  try {
    // 传完整行：后端 update 是全字段覆盖，只传 status 会清空标题/类别/地点等
    await api.update(currentItem.value.id, { ...currentItem.value, status: newStatus.value })
    currentItem.value.status = newStatus.value
    Message.success('状态已更新')
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '更新失败') }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除赛事',
    content: `确定删除「${row.title}」吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}
</script>

<style scoped>
.time-text { color: #86909C; font-size: 12px; }

.review-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 16px;
  gap: 8px;
}
</style>
