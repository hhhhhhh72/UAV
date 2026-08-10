<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="shops"
      :columns="columns"
      :search-fields="searchFields"
      creatable
      add-label="新增商家"
      @add="openForm()"
    >
      <template #license="{ record }">
        <a-image v-if="record.license_url" :src="record.license_url" :width="48" :height="48" :preview="false" fit="cover" style="border-radius: 4px" />
        <span v-else class="text-muted">—</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusMap[record.status]" size="small">{{ statusText[record.status] || record.status }}</a-tag>
      </template>
      <template #member="{ record }">
        <a-tag :color="record.is_member ? 'green' : 'gray'" size="small">{{ record.is_member ? '是' : '否' }}</a-tag>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无商家数据" />
      </template>
    </CrudList>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑商家' : '新增商家'" :width="500" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical">
        <a-form-item label="商家名称" required><a-input v-model="form.name" placeholder="输入商家名称" style="width: 100%" /></a-form-item>
        <a-form-item label="营业执照">
          <a-upload
            class="license-upload"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :show-file-list="false"
            :before-upload="beforeUpload"
            accept="image/*"
            @success="onUploadSuccess"
          >
            <img v-if="form.license_url" :src="form.license_url" class="license-preview" />
            <a-button v-else type="primary" status="success">点击上传</a-button>
          </a-upload>
          <a-button v-if="form.license_url" size="small" style="margin-top: 8px" @click="form.license_url = ''">清除</a-button>
        </a-form-item>
        <a-form-item label="对公账户"><a-input v-model="form.account_name" placeholder="对公账户名称" style="width: 100%" /></a-form-item>
        <a-form-item label="联系电话"><a-input v-model="form.contact_phone" placeholder="商家联系电话" style="width: 100%" /></a-form-item>
        <a-form-item label="地址"><a-input v-model="form.address" placeholder="商家地址" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="pending">待审核</a-option>
            <a-option value="approved">已批准</a-option>
            <a-option value="rejected">已驳回</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="协会会员"><a-switch v-model="form.is_member" /></a-form-item>
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
const api = useAdminApi('shops')

const searchFields = []

const statusMap = { pending: 'orange', approved: 'green', rejected: 'red' }
const statusText = { pending: '待审核', approved: '已批准', rejected: '已驳回' }

const columns = [
  { title: 'ID', dataIndex: 'id', width: 120 },
  { title: '商家名称', dataIndex: 'name', minWidth: 180 },
  { title: '营业执照', dataIndex: 'license_url', slotName: 'license', width: 100 },
  { title: '对公账户', dataIndex: 'account_name', width: 160 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 110 },
  { title: '会员', dataIndex: 'is_member', slotName: 'member', width: 80 },
  { title: '入驻时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', license_url: '', account_name: '', contact_phone: '', address: '', status: 'pending', is_member: false })
const uploadUrl = '/api/v1/upload'
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}` }

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

// a-upload 成功回调：fileItem.response 为后端原始响应体
const onUploadSuccess = (fileItem) => {
  const res = fileItem.response || {}
  form.license_url = res?.data?.url || res?.url || ''
  Message.success('上传成功')
}

const formatDate = (d) => {
  if (!d) return '—'
  return new Date(d).toLocaleString('zh-CN')
}

const resetForm = () => Object.assign(form, { id: '', name: '', license_url: '', account_name: '', contact_phone: '', address: '', status: 'pending', is_member: false })

const openForm = (row) => {
  resetForm()
  if (row) {
    if (!row.id) { Message.warning('此商家无有效 ID，无法编辑。请使用新增功能创建新商家。'); return }
    formEdit.value = true
    Object.assign(form, {
      id: row.id || '',
      name: row.name || '',
      license_url: row.license_url || '',
      account_name: row.account_name || '',
      contact_phone: row.contact_phone || '',
      address: row.address || '',
      status: row.status || 'pending',
      is_member: !!row.is_member
    })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入商家名称'); return }
  formLoading.value = true
  try {
    const payload = {
      name: form.name, license_url: form.license_url, account_name: form.account_name,
      contact_phone: form.contact_phone, address: form.address,
      status: form.status, is_member: form.is_member
    }
    if (formEdit.value) {
      await api.update(form.id, payload)
      Message.success('更新成功')
    } else {
      await api.create(payload)
      Message.success('创建成功')
    }
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(errMsg(e))
  } finally { formLoading.value = false }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: `确认删除商家「${row.name}」？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('删除成功')
        crudRef.value?.reload()
      } catch (e) { Message.error(errMsg(e)) }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.text-muted { color: #C9CDD4; }

.time-text { color: #86909C; font-size: 12px; }

.license-upload { display: inline-block; }

.license-preview {
  width: 120px;
  height: 80px;
  object-fit: cover;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
}
</style>
