<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="industry-reports"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增报告"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #category="{ record }">
        <a-tag :color="typeTag(record.category)" size="small">{{ typeLabel(record.category) }}</a-tag>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ record.status || '-' }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无报告数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="报告详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="报告标题" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="typeTag(currentItem.report_type)" size="small">{{ typeLabel(currentItem.report_type) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发布机构">{{ currentItem.publisher || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发布日期">{{ formatDate(currentItem.publish_date) }}</a-descriptions-item>
          <a-descriptions-item label="作者">{{ currentItem.authors || '-' }}</a-descriptions-item>
          <a-descriptions-item label="摘要" :span="2">{{ currentItem.abstract || '-' }}</a-descriptions-item>
          <a-descriptions-item label="文件" :span="2">
            <a v-if="currentItem.file_url" :href="currentItem.file_url" target="_blank">下载报告</a>
            <span v-else>-</span>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑报告' : '新增报告'" :width="560" destroy-on-close>
      <a-form :model="form" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="16"><a-form-item label="报告标题" required><a-input v-model="form.title" /></a-form-item></a-col>
          <a-col :span="8">
            <a-form-item label="类型">
              <a-select v-model="form.category" style="width: 100%">
                <a-option value="whitepaper">白皮书</a-option>
                <a-option value="research">调研报告</a-option>
                <a-option value="analysis">行业分析</a-option>
                <a-option value="other">其他</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12"><a-form-item label="报告期"><a-input v-model="form.period" placeholder="如 2026-Q2" /></a-form-item></a-col>
          <a-col :span="12"><a-form-item label="作者"><a-input v-model="form.author" /></a-form-item></a-col>
        </a-row>
        <a-form-item label="摘要"><a-input v-model="form.summary" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="文件URL"><a-input v-model="form.file_url" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="published">已发布</a-option>
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
const api = useAdminApi('industry-reports')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}
const typeLabel = (t) => ({ whitepaper: '白皮书', research: '调研报告', analysis: '行业分析', other: '其他' }[t] || t || '-')
const typeTag = (t) => ({ whitepaper: 'orangered', research: 'green', analysis: 'arcoblue', other: 'gray' }[t] || 'gray')
const statusTag = (s) => ({ published: 'green', draft: 'orangered' }[s] || 'gray')

// 批量动作：批量发布——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) }
]

// 搜索用 report_type（列表字段为 category，表单字段为 category，详情字段为 report_type——三处字段差异保持原样）
const searchFields = [
  { key: 'keyword', label: '关键词', type: 'input', width: 220, placeholder: '搜索报告标题' },
  { key: 'report_type', label: '类型', type: 'select', width: 140, options: [
    { value: '', label: '全部类型' },
    { value: 'whitepaper', label: '白皮书' },
    { value: 'research', label: '调研报告' },
    { value: 'analysis', label: '行业分析' },
    { value: 'other', label: '其他' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '报告标题', dataIndex: 'title', slotName: 'title', minWidth: 200, ellipsis: true, tooltip: true },
  { title: '类型', dataIndex: 'category', slotName: 'category', width: 110 },
  { title: '报告期', dataIndex: 'period', width: 130 },
  { title: '作者', dataIndex: 'author', width: 140 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', category: 'whitepaper', period: '', author: '', summary: '', file_url: '', status: 'draft' })
const resetForm = () => Object.assign(form, { id: '', title: '', category: 'whitepaper', period: '', author: '', summary: '', file_url: '', status: 'draft' })
const openForm = (r) => {
  resetForm()
  if (r) {
    formEdit.value = true
    Object.assign(form, { id: r.id, title: r.title || '', category: r.category || 'whitepaper', period: r.period || '', author: r.author || '', summary: r.summary || '', file_url: r.file_url || '', status: r.status || 'draft' })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入报告标题'); return }
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
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除报告',
    content: `确定删除报告"${r.title}"吗？`,
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
</style>
