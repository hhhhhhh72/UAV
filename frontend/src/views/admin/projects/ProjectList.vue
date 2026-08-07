<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="research-projects"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增项目"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #field="{ record }">
        <a-tag size="small">{{ record.field || '-' }}</a-tag>
      </template>
      <template #budget="{ record }">
        <span class="cell-amount">{{ formatMoney(record.budget_fen) }}</span>
      </template>
      <template #members="{ record }">{{ Array.isArray(record.members) ? record.members.join('、') : (record.members || '-') }}</template>
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
        <a-empty description="暂无项目数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="项目详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="项目名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="牵头单位">{{ currentItem.lead_org || '-' }}</a-descriptions-item>
          <a-descriptions-item label="领域">{{ currentItem.field || '-' }}</a-descriptions-item>
          <a-descriptions-item label="预算">{{ formatMoney(currentItem.budget_fen) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束日期">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="参与单位" :span="2">{{ Array.isArray(currentItem.members) ? currentItem.members.join('、') : (currentItem.members || '-') }}</a-descriptions-item>
          <a-descriptions-item label="里程碑" :span="2">{{ currentItem.milestones || '-' }}</a-descriptions-item>
          <a-descriptions-item v-if="currentItem.description" label="描述" :span="2">{{ currentItem.description }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑课题' : '新增课题'" :width="560" @close="resetForm">
      <a-form :model="form" layout="vertical">
        <a-form-item label="课题名称"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="牵头单位"><a-input v-model="form.lead_org" /></a-form-item>
        <a-form-item label="领域"><a-input v-model="form.field" /></a-form-item>
        <a-form-item label="参与单位"><a-input v-model="form.membersInput" placeholder="多个单位用逗号分隔" /></a-form-item>
        <a-form-item label="预算(分)"><a-input-number v-model="form.budget_fen" :min="0" style="width: 100%" /></a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="开始日期"><a-date-picker v-model="form.start_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="结束日期"><a-date-picker v-model="form.end_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="里程碑"><a-input v-model="form.milestones" placeholder="阶段目标，如：方案设计→样机测试" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="active">进行中</a-option>
            <a-option value="completed">已完成</a-option>
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
const api = useAdminApi('research-projects')

// 统一状态为 active(进行中)/completed(已完成)，兼容历史数据 planning/recruiting/ongoing
const statusLabel = (s) => ({ active: '进行中', planning: '规划中', recruiting: '进行中', ongoing: '进行中', completed: '已完成' }[s] || s || '-')
const statusTag = (s) => ({ active: 'arcoblue', planning: 'gray', recruiting: 'orangered', ongoing: 'arcoblue', completed: 'green' }[s] || 'gray')
const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}

const formatMoney = (fen) => {
  if (fen == null) return '-'
  const yuan = Number(fen) / 100
  return '¥' + yuan.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

// 批量动作：设为进行中 / 标记完成——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'start', label: '设为进行中', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'active' }) },
  { key: 'complete', label: '标记完成', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'completed' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索项目名称', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'active', label: '进行中' },
    { value: 'completed', label: '已完成' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '项目名称', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '牵头单位', dataIndex: 'lead_org', minWidth: 160 },
  { title: '领域', dataIndex: 'field', slotName: 'field', width: 120 },
  { title: '预算', dataIndex: 'budget_fen', slotName: 'budget', width: 130, align: 'right' },
  { title: '参与单位', dataIndex: 'members', slotName: 'members', minWidth: 140 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', lead_org: '', field: '', budget_fen: 0, milestones: '', start_date: '', end_date: '', membersInput: '', status: 'active', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', lead_org: '', field: '', budget_fen: 0, start_date: '', end_date: '', membersInput: '', status: 'active', description: '' })
const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id, title: row.title || '', lead_org: row.lead_org || '', field: row.field || '',
      budget_fen: row.budget_fen || 0, milestones: row.milestones || '', status: row.status || 'active', description: row.description || '',
      start_date: (row.start_date || '').slice(0, 10), end_date: (row.end_date || '').slice(0, 10),
      membersInput: Array.isArray(row.members) ? row.members.join('、') : (row.members || '')
    })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入项目名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    // members: 逗号/顿号分隔 → 数组（后端 []string）
    p.members = (form.membersInput || '').split(/[,、]/).map(s => s.trim()).filter(Boolean)
    delete p.membersInput
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除项目',
    content: `确定删除项目"${r.title}"吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(r.id); Message.success('已删除'); crudRef.value?.reload() }
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

.cell-amount { color: #E96012; font-weight: 500; }
</style>
