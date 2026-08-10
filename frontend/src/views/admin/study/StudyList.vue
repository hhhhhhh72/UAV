<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="study-tours"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增研学"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无研学项目数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="研学项目详情" :width="600" :footer="false" :mask-closable="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="项目名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="目的地">{{ currentItem.destination || '-' }}</a-descriptions-item>
          <a-descriptions-item label="时长">{{ currentItem.duration || '-' }}</a-descriptions-item>
          <a-descriptions-item label="名额">{{ currentItem.capacity ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="项目介绍" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑研学项目' : '新增研学项目'" :width="560" :mask-closable="false" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical">
        <a-form-item label="项目名称" required><a-input v-model="form.title" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="active">进行中</a-option>
            <a-option value="closed">已结束</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="目的地"><a-input v-model="form.destination" style="width: 100%" /></a-form-item>
        <a-form-item label="时长"><a-input v-model="form.duration" placeholder="如: 3天2晚" style="width: 100%" /></a-form-item>
        <a-form-item label="名额"><a-input-number v-model="form.capacity" :min="0" hide-button style="width: 100%" /></a-form-item>
        <a-form-item label="项目介绍"><a-input v-model="form.description" type="textarea" :rows="2" style="width: 100%" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">提交</a-button>
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
const api = useAdminApi('study-tours')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ active: 'green', closed: 'gray', draft: 'orangered' }[s] || 'gray')
const statusLabel = { active: '进行中', closed: '已结束', draft: '草稿' }

// 批量动作：批量上架（active）/ 批量结束（closed）——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '开始招募', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'active' }) },
  { key: 'close', label: '结束招募', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索研学项目...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'active', label: '进行中' },
    { value: 'closed', label: '已结束' },
    { value: 'draft', label: '草稿' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '研学项目', dataIndex: 'title', slotName: 'title', minWidth: 220 },
  { title: '目的地', dataIndex: 'destination', width: 140 },
  { title: '时长', dataIndex: 'duration', width: 100 },
  { title: '名额', dataIndex: 'capacity', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (d) => { currentItem.value = d; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', destination: '', duration: '', capacity: 0, status: 'draft', description: '' })

const resetForm = () => Object.assign(form, { id: '', title: '', destination: '', duration: '', capacity: 0, status: 'draft', description: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, { ...row, capacity: row.capacity ?? 0 })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入项目名称'); return }
  formLoading.value = true
  try {
    if (formEdit.value) {
      await api.update(form.id, { ...form })
      Message.success('更新成功')
    } else {
      await api.create({ ...form })
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
    title: '删除研学项目',
    content: `确定删除该研学项目吗？（${row.title || row.id}）`,
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
</style>
