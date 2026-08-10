<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="transformations"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增转化"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || record.achievement_id || '-' }}</span>
      </template>
      <template #stage="{ record }">
        <a-tag :color="stageTag(record.stage)" size="small">{{ stageLabel(record.stage) }}</a-tag>
      </template>
      <template #progress="{ record }">
        <span class="cell-title">{{ record.progress || '-' }}</span>
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
        <a-empty description="暂无转化数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="转化详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="转化标题" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="当前阶段">
            <a-tag :color="stageTag(currentItem.stage)" size="small">{{ stageLabel(currentItem.stage) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="负责人ID">{{ currentItem.owner_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="完成进度">{{ currentItem.progress ?? '-' }}%</a-descriptions-item>
          <a-descriptions-item label="关联成果ID">{{ currentItem.achievement_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="里程碑" :span="2">
            <a-space v-if="Array.isArray(currentItem.milestones) && currentItem.milestones.length" direction="vertical" :size="4" style="width: 100%;">
              <div v-for="(ms, idx) in currentItem.milestones" :key="idx" class="ms-item">
                <a-tag :color="ms.completed ? 'green' : 'gray'" size="small">{{ ms.completed ? '已完成' : '未完成' }}</a-tag>
                <span>{{ ms.name || ms.date || '里程碑' }}</span>
              </div>
            </a-space>
            <span v-else>-</span>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑转化' : '新增转化'" :width="560" @close="resetForm">
      <a-form :model="form" layout="vertical">
        <a-form-item label="标题"><a-input v-model="form.title" style="width: 100%" /></a-form-item>
        <a-form-item label="成果ID"><a-input v-model="form.achievement_id" placeholder="关联成果 ID" style="width: 100%" /></a-form-item>
        <a-form-item label="合作方ID"><a-input v-model="form.partner_id" placeholder="关联合作企业 ID" style="width: 100%" /></a-form-item>
        <a-form-item label="阶段" :extra="formEdit ? '' : '新转化默认从「实验室」阶段开始，创建后可编辑推进'">
          <a-select v-model="form.stage" :disabled="!formEdit" style="width: 100%">
            <a-option value="lab">实验室</a-option>
            <a-option value="pilot">中试</a-option>
            <a-option value="industrialized">产业化</a-option>
            <a-option value="listed">上市</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="进度说明"><a-input v-model="form.progress" type="textarea" :autosize="{ minRows: 3 }" style="width: 100%" /></a-form-item>
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
const api = useAdminApi('transformations')
const formEdit = ref(false)
const formVisible = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', achievement_id: '', partner_id: '', stage: 'lab', progress: '' })

const stageLabel = (s) => ({ lab: '实验室', pilot: '中试', industrialized: '产业化', listed: '上市' }[s] || s || '-')
const stageTag = (s) => ({ lab: 'gray', pilot: 'orangered', industrialized: 'green', listed: 'arcoblue' }[s] || 'gray')
const statusTag = (s) => ({ active: 'green', completed: 'arcoblue', cancelled: 'red', ongoing: 'orangered' }[s] || 'gray')
const statusLabel = (s) => ({ active: '进行中', completed: '已完成', cancelled: '已取消', ongoing: '进行中' }[s] || (s || '-'))

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}

// 批量动作：标记完成 / 标记取消——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'complete', label: '标记完成', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'completed' }) },
  { key: 'cancel', label: '标记取消', status: 'danger', api: (row) => api.update(row.id, { ...row, status: 'cancelled' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索成果名称', width: 220 },
  { key: 'stage', label: '阶段', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'lab', label: '实验室' },
    { value: 'pilot', label: '中试' },
    { value: 'industrialized', label: '产业化' },
    { value: 'listed', label: '上市' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '转化标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '当前阶段', dataIndex: 'stage', slotName: 'stage', width: 100 },
  { title: '负责人ID', dataIndex: 'owner_id', width: 110 },
  { title: '进度说明', dataIndex: 'progress', slotName: 'progress', minWidth: 160 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, row)
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}
const resetForm = () => Object.assign(form, { id: '', title: '', achievement_id: '', partner_id: '', achievement_title: '', stage: 'lab', target_date: '', description: '' })
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入标题'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleDelete = (row) => {
  Modal.confirm({
    title: '删除转化记录',
    content: `确定删除转化记录 "${row.achievement_title}" 吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); crudRef.value?.reload() }
      catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

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

