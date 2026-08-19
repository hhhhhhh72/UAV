<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="portfolios"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增品牌"
      @add="openForm()"
    >
      <template #name="{ record }">
        <span class="cell-title">{{ record.name || '-' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button v-if="record.status === 'pending'" type="text" status="success" size="small" @click="handleApprove(record)">通过</a-button>
          <a-button v-if="record.status === 'pending'" type="text" status="danger" size="small" @click="handleReject(record)">驳回</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无品牌数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="品牌详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="品牌名称">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="审核状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Logo URL" :span="2">{{ currentItem.logo_url || '-' }}</a-descriptions-item>
          <a-descriptions-item label="封面图 URL" :span="2">{{ currentItem.cover_url || '-' }}</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="荣誉" :span="2">{{ (Array.isArray(currentItem.honors) ? currentItem.honors.join('、') : (currentItem.honors || '-')) }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑品牌' : '新增品牌'" :width="560" :on-before-cancel="beforeClose">
      <a-form :model="form" layout="vertical">
        <a-form-item label="品牌名称" required><a-input v-model="form.name" style="width: 100%" /></a-form-item>
        <a-form-item label="Logo URL"><a-input v-model="form.logo_url" style="width: 100%" /></a-form-item>
        <a-form-item label="封面图 URL"><a-input v-model="form.cover_url" style="width: 100%" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 2 }" style="width: 100%" /></a-form-item>
        <a-form-item label="荣誉"><a-input v-model="form.honorsText" type="textarea" :autosize="{ minRows: 2 }" placeholder="多个荣誉用逗号分隔" style="width: 100%" /></a-form-item>
        <a-form-item label="审核状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="published">已发布</a-option>
            <a-option value="rejected">已驳回</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="cancelForm">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('portfolios')

const statusLabel = (s) => ({ pending: '待审核', published: '已发布', rejected: '已驳回' }[s] || s || '-')
const statusTag = (s) => ({ pending: 'orangered', published: 'green', rejected: 'red' }[s] || 'gray')

// 批量动作：批量通过 / 批量驳回——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'reject', label: '批量驳回', status: 'danger', api: (row) => api.update(row.id, { ...row, status: 'rejected' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', type: 'input', width: 240, placeholder: '搜索企业或品牌名称' },
  { key: 'status', label: '状态', type: 'select', width: 140, options: [
    { value: '', label: '全部状态' },
    { value: 'draft', label: '草稿' },
    { value: 'published', label: '已发布' },
    { value: 'rejected', label: '已驳回' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '品牌名称', dataIndex: 'name', slotName: 'name', minWidth: 160, ellipsis: true, tooltip: true },
  { title: '审核状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 220, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', logo_url: '', cover_url: '', description: '', honorsText: '', status: 'draft' })
const resetForm = () => Object.assign(form, { id: '', name: '', logo_url: '', cover_url: '', description: '', honorsText: '', status: 'draft' })
const openForm = (r) => {
  resetForm()
  if (r) {
    formEdit.value = true
    Object.assign(form, { id: r.id, name: r.name || '', logo_url: r.logo_url || '', cover_url: r.cover_url || '', description: r.description || '', honorsText: Array.isArray(r.honors) ? r.honors.join('、') : (r.honors || ''), status: r.status || 'draft' })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.name) { Message.warning('请输入品牌名称'); return }
  formLoading.value = true
  try {
    const p = { id: form.id, name: form.name, logo_url: form.logo_url, cover_url: form.cover_url, description: form.description, status: form.status, honors: String(form.honorsText || '').split(/[,，、]/).map(x => x.trim()).filter(Boolean) }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
// 未保存守卫：Esc/点 X/点遮罩/底部取消 关闭前比对快照，有改动先确认
// 注意：Arco 2.58 无 beforeClose prop（beforeClose 只是 emits 事件），
// 需用 on-before-cancel 拦截用户关闭（X/ESC/遮罩）；底部取消按钮走 cancelForm。
let formSnapshot = ''
const confirmDiscard = () => {
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
}
const cancelForm = () => {
  if (JSON.stringify(form) === formSnapshot) { formVisible.value = false; return }
  confirmDiscard()
}
const beforeClose = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  confirmDiscard()
  return false
}
const handleApprove = async (r) => {
  // 传完整行：后端 update 是全字段覆盖，只传 status 会清空 name/logo/描述等
  try { await api.update(r.id, { ...r, status: 'published' }); Message.success('已发布'); crudRef.value?.reload() }
  catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
}
const handleReject = async (r) => {
  try { await api.update(r.id, { ...r, status: 'rejected' }); Message.success('已驳回'); crudRef.value?.reload() }
  catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除品牌',
    content: '确定删除该品牌吗？',
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
