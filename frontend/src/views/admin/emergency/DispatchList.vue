<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="emergency-dispatches"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      creatable
      add-label="新建调度"
      @add="openForm()"
    >
      <template #eventDesc="{ record }">
        <span class="cell-title">{{ record.event_desc || '-' }}</span>
      </template>
      <template #startTime="{ record }">
        <span class="time-text">{{ formatDate(record.start_time) }}</span>
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
        <a-empty description="暂无调度记录" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="调度详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="任务名称" :span="2">{{ currentItem.event_desc || '-' }}</a-descriptions-item>
          <a-descriptions-item label="调度资源">{{ currentItem.resource_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="位置">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="调度时间">{{ formatDate(currentItem.start_time) }}</a-descriptions-item>
          <a-descriptions-item label="结束时间">{{ formatDate(currentItem.end_time) }}</a-descriptions-item>
          <a-descriptions-item label="指挥员">{{ currentItem.commander || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item v-if="currentItem.result" label="处理结果" :span="2">{{ currentItem.result }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑调度' : '新建调度'" :width="560" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="horizontal">
        <a-row :gutter="16">
          <a-col :span="14">
            <a-form-item label="任务名称" required><a-input v-model="form.event_desc" /></a-form-item>
          </a-col>
          <a-col :span="10">
            <a-form-item label="处理结果"><a-input v-model="form.result" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="调度资源"><a-input v-model="form.resource_id" placeholder="资源 ID" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="位置"><a-input v-model="form.location" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="调度时间"><a-input v-model="form.start_time" placeholder="YYYY-MM-DD HH:mm" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="结束时间"><a-input v-model="form.end_time" placeholder="YYYY-MM-DD HH:mm" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="指挥员"><a-input v-model="form.commander" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="dispatched">已调度</a-option>
            <a-option value="in_progress">执行中</a-option>
            <a-option value="completed">已完成</a-option>
            <a-option value="cancelled">已取消</a-option>
          </a-select>
        </a-form-item>
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
const api = useAdminApi('emergency-dispatches')
const defaultParams = { status: '' }

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusLabel = (s) => ({ dispatched: '已调度', in_progress: '执行中', completed: '已完成', cancelled: '已取消' }[s] || s || '-')
const statusTag = (s) => ({ dispatched: 'orange', in_progress: 'arcoblue', completed: 'green', cancelled: 'red' }[s] || 'gray')

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索任务名称...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部状态' },
    { value: 'dispatched', label: '已调度' },
    { value: 'in_progress', label: '执行中' },
    { value: 'completed', label: '已完成' },
    { value: 'cancelled', label: '已取消' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '任务名称', dataIndex: 'event_desc', slotName: 'eventDesc', minWidth: 180 },
  { title: '调度资源', dataIndex: 'resource_id', width: 140 },
  { title: '调度时间', dataIndex: 'start_time', slotName: 'startTime', width: 160 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', event_desc: '', result: '', resource_id: '', start_time: '', end_time: '', location: '', commander: '', status: 'dispatched' })
const resetForm = () => Object.assign(form, { id: '', event_desc: '', result: '', resource_id: '', start_time: '', end_time: '', location: '', commander: '', status: 'dispatched' })

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

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.event_desc) { Message.warning('请输入任务名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (r) => {
  Modal.confirm({
    title: '提示',
    content: '确定删除?',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(r.id); Message.success('已删除'); crudRef.value?.reload() } catch { Message.error('删除失败') }
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

.time-text { color: #86909C; font-size: 12px; }
</style>
