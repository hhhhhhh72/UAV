<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="rd-challenges"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增难题"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #field="{ record }">
        <a-tag size="small">{{ record.field || '-' }}</a-tag>
      </template>
      <template #reward="{ record }">
        <span class="cell-amount">{{ formatMoney(record.budget_fen) }}</span>
      </template>
      <template #deadline="{ record }">
        <span class="time-text">{{ formatDate(record.deadline) }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无难题数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="难题详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="难题标题" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发布企业ID">{{ currentItem.poster_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="领域">{{ currentItem.field || '-' }}</a-descriptions-item>
          <a-descriptions-item label="悬赏金额">{{ formatMoney(currentItem.budget_fen) }}</a-descriptions-item>
          <a-descriptions-item label="截止日期">{{ formatDate(currentItem.deadline) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="难题描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑难题' : '新增难题'" :width="560" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical">
        <a-form-item label="难题名称"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="领域"><a-input v-model="form.field" placeholder="如：飞控系统" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="悬赏金额(分)"><a-input-number v-model="form.budget_fen" :min="0" style="width: 100%" placeholder="单位：分" /></a-form-item>
        <a-form-item label="截止日期"><a-date-picker v-model="form.deadline" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
        <a-form-item v-if="formEdit" label="状态" :extra="formEdit ? '' : '新建默认征集中，创建后可在编辑中调整'">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="open">征集中</a-option>
            <a-option value="closed">已关闭</a-option>
            <a-option value="resolved">已解决</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('rd-challenges')

const statusLabel = (s) => ({ open: '征集中', in_progress: '进行中', closed: '已关闭', resolved: '已解决', published: '已发布' }[s] || s || '-')
const statusTag = (s) => ({ open: 'orangered', in_progress: 'arcoblue', closed: 'gray', resolved: 'green', published: 'orangered' }[s] || 'gray')

const formatMoney = (fen) => {
  if (fen == null) return '-'
  const yuan = Number(fen) / 100
  return '¥' + yuan.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 批量动作：批量关闭 / 批量标记解决——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'close', label: '批量关闭', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) },
  { key: 'resolve', label: '批量标记解决', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'resolved' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索难题标题', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'in_progress', label: '进行中' },
    { value: 'closed', label: '已关闭' },
    { value: 'resolved', label: '已解决' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '难题标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '发布企业ID', dataIndex: 'poster_id', minWidth: 140 },
  { title: '领域', dataIndex: 'field', slotName: 'field', width: 120 },
  { title: '悬赏金额', dataIndex: 'budget_fen', slotName: 'reward', width: 130, align: 'right' },
  { title: '截止日期', dataIndex: 'deadline', slotName: 'deadline', width: 130 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', field: '', budget_fen: 0, deadline: '', status: 'open', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', field: '', budget_fen: 0, deadline: '', status: 'open', description: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id, title: row.title || '', field: row.field || '',
      budget_fen: row.budget_fen || 0, deadline: row.deadline || '',
      status: row.status || 'open', description: row.description || ''
    })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入难题标题'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    if (formEdit.value) {
      await api.update(form.id, p)
      Message.success('更新成功')
    } else {
      await api.create(p)
      Message.success('创建成功')
    }
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除难题',
    content: `确定删除难题"${row.title}"吗？`,
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
.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.cell-amount { color: #E96012; font-weight: 500; }

.time-text { color: #86909C; font-size: 12px; }
</style>
