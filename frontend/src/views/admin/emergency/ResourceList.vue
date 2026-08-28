<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="emergency-resources"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增资源"
      @add="openForm()"
    >
      <template #name="{ record }">
        <span class="cell-title">{{ record.name || '-' }}</span>
      </template>
      <template #type="{ record }">
        <a-tag :color="typeTag(record.res_type)" size="small">{{ typeLabel(record.res_type) }}</a-tag>
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
        <a-empty description="暂无应急资源" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="应急资源详情" :width="'min(600px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="资源名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="typeTag(currentItem.res_type)" size="small">{{ typeLabel(currentItem.res_type) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="规格">{{ currentItem.specs || '-' }}</a-descriptions-item>
          <a-descriptions-item label="数量">{{ currentItem.quantity ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="位置">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="联系人">{{ revealPII ? (currentItem.contact_info || '-') : maskContact(currentItem.contact_info) }}</a-descriptions-item>
          <a-descriptions-item label="敏感信息" :span="2">
            <a-checkbox v-model="revealPII" :disabled="!currentItem">显示完整联系人信息（谨慎操作）</a-checkbox>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑应急资源' : '新增应急资源'" :width="'min(560px, 94vw)'" :mask-closable="false" :unmount-on-close="true" @before-cancel="beforeCancel">
      <a-form :model="form" layout="vertical">
        <a-form-item label="资源名称" required><a-input v-model="form.name" :aria-required="true" style="width: 100%" /></a-form-item>
        <a-form-item label="类型">
          <a-select v-model="form.res_type" style="width: 100%">
            <a-option value="drone">无人机</a-option>
            <a-option value="comm">通信</a-option>
            <a-option value="light">照明</a-option>
            <a-option value="transport">运输</a-option>
            <a-option value="other">其他</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="规格"><a-input v-model="form.specs" style="width: 100%" /></a-form-item>
        <a-form-item label="数量"><a-input-number v-model="form.quantity" :min="0" hide-button style="width: 100%" /></a-form-item>
        <a-form-item label="位置"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="联系人信息"><a-input v-model="form.contact_info" placeholder="姓名 / 电话，如：张工 13800138000" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="available">可用</a-option>
            <a-option value="in_use">使用中</a-option>
            <a-option value="maintenance">维护中</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="cancelForm">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">保存</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { maskContact } from '@/utils/mask'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('emergency-resources')

const typeLabel = (t) => ({ drone: '无人机', comm: '通信', light: '照明', transport: '运输', other: '其他' }[t] || t || '-')
const typeTag = (t) => ({ drone: 'green', comm: 'orange', light: 'gray', transport: 'arcoblue', other: 'arcoblue' }[t] || 'gray')
const statusLabel = (s) => ({ available: '可用', in_use: '使用中', maintenance: '维护中' }[s] || s || '-')
const statusTag = (s) => ({ available: 'green', in_use: 'orange', maintenance: 'red' }[s] || 'gray')

// 批量动作：设为可用 / 设为维护中——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'available', label: '设为可用', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'available' }) },
  { key: 'maintenance', label: '设为维护中', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'maintenance' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索资源名称...', width: 220 },
  { key: 'res_type', label: '类型', type: 'select', options: [
    { value: '', label: '全部类型' },
    { value: 'drone', label: '无人机' },
    { value: 'comm', label: '通信' },
    { value: 'light', label: '照明' },
    { value: 'transport', label: '运输' },
    { value: 'other', label: '其他' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '资源名称', dataIndex: 'name', slotName: 'name', minWidth: 160 },
  { title: '类型', dataIndex: 'res_type', slotName: 'type', width: 100 },
  { title: '规格', dataIndex: 'specs', width: 140 },
  { title: '数量', dataIndex: 'quantity', width: 70, align: 'center' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const revealPII = ref(false)
const showDetail = (r) => { revealPII.value = false; currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', res_type: 'drone', specs: '', quantity: null, location: '', contact_info: '', status: 'available' })
const resetForm = () => Object.assign(form, { id: '', name: '', res_type: 'drone', specs: '', quantity: null, location: '', contact_info: '', status: 'available' })

// 未保存守卫：Esc/遮罩/X/取消 关闭前若有改动则确认，避免输入全丢
let formSnapshot = ''
const confirmDiscard = (onOk) => {
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk,
  })
}
// Arco 2.x：@before-cancel 为同步守卫（返回 false 阻止关闭），Esc/X/遮罩均走此路径
const beforeCancel = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  confirmDiscard(() => { formVisible.value = false })
  return false
}
// footer 取消按钮也走守卫
const cancelForm = () => {
  if (JSON.stringify(form) === formSnapshot) { formVisible.value = false; return }
  confirmDiscard(() => { formVisible.value = false })
}

const openForm = (r) => {
  resetForm()
  if (r) {
    formEdit.value = true
    // 显式映射可编辑字段：不整行回传，避免把只读字段（id/created_at 等）带回提交
    Object.assign(form, {
      id: r.id, name: r.name || '', res_type: r.res_type || 'drone', specs: r.specs || '',
      quantity: r.quantity ?? null, location: r.location || '', contact_info: r.contact_info || '',
      status: r.status || 'available',
    })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入资源名称'); return }
  formLoading.value = true
  try {
    // 只提交可编辑字段；数量为空传 null，不伪造 0
    const p = {
      name: form.name,
      res_type: form.res_type,
      specs: form.specs || '',
      quantity: form.quantity ?? null,
      location: form.location || '',
      contact_info: form.contact_info || '',
      status: form.status,
    }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (r) => {
  Modal.confirm({
    title: '提示',
    content: `确定删除应急资源「${r.name || r.id}」吗？删除后不可恢复`,
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
</style>
