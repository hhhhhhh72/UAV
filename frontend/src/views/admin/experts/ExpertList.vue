<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="experts"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增专家"
      @add="openForm()"
    >
      <template #tags="{ record }">
        <a-tag v-for="t in record.tags" :key="t" size="small" style="margin: 2px">{{ t }}</a-tag>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusMap[record.status] || 'gray'" size="small">{{ statusLabel[record.status] || record.status }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无专家数据" />
      </template>
    </CrudList>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑专家' : '新增专家'" :width="500" destroy-on-close>
      <a-form :model="form" layout="vertical">
        <a-form-item label="姓名" required><a-input v-model="form.name" /></a-form-item>
        <a-form-item label="职称"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="单位"><a-input v-model="form.org" /></a-form-item>
        <a-form-item label="领域"><a-input v-model="form.field" /></a-form-item>
        <a-form-item label="简介"><a-input v-model="form.bio" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="头像">
          <a-upload
            class="avatar-upload"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :show-file-list="false"
            accept="image/*"
            :before-upload="beforeUpload"
            @success="onUploadSuccess"
          >
            <a-avatar v-if="form.avatar_url" :src="form.avatar_url" :size="80" shape="square" />
            <a-button v-else type="outline">点击上传</a-button>
          </a-upload>
          <a-button v-if="form.avatar_url" size="small" style="margin-top: 8px" @click="form.avatar_url = ''">清除</a-button>
        </a-form-item>
        <a-form-item label="标签"><a-input v-model="tagsInput" placeholder="逗号分隔" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="pending">待审核</a-option>
            <a-option value="published">已发布</a-option>
            <a-option value="archived">已下架</a-option>
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
const api = useAdminApi('experts')

const statusMap = { pending: 'orangered', published: 'green', archived: 'gray' }
const statusLabel = { pending: '待审核', published: '已发布', archived: '已下架' }
const tagsInput = ref('')
const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', title: '', org: '', field: '', bio: '', avatar_url: '', status: 'pending', tags: [] })
const uploadUrl = '/api/v1/upload'
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}` }

const columns = [
  { title: 'ID', dataIndex: 'id', width: 140 },
  { title: '姓名', dataIndex: 'name', width: 100 },
  { title: '职称', dataIndex: 'title', width: 120 },
  { title: '所属单位', dataIndex: 'org', minWidth: 180 },
  { title: '领域', dataIndex: 'field', width: 120 },
  { title: '标签', slotName: 'tags', width: 200 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

// 批量动作：批量发布 / 批量下架——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'archive', label: '批量下架', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'archived' }) }
]

// 专家列表仅按领域筛选（仅 field 参数）
const searchFields = [
  { key: 'field', label: '领域', type: 'select', width: 160, options: [
    { value: '', label: '全部' },
    { value: '无人机平台', label: '无人机平台' },
    { value: '飞控系统', label: '飞控系统' },
    { value: '导航与定位', label: '导航与定位' },
    { value: '通信链路', label: '通信链路' },
    { value: '载荷与传感器', label: '载荷与传感器' },
    { value: '能源动力', label: '能源动力' },
    { value: '人工智能', label: '人工智能' },
    { value: '新材料', label: '新材料' }
  ]}
]

const beforeUpload = (item) => {
  const file = item?.file || item
  const isImage = !!file.type && file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

const onUploadSuccess = (fileItem) => {
  const res = fileItem?.response || {}
  form.avatar_url = res?.data?.url || res?.url || ''
  Message.success('上传成功')
}

const resetForm = () => Object.assign(form, { id: '', name: '', title: '', org: '', field: '', bio: '', avatar_url: '', status: 'pending', tags: [] })
const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, { ...row, tags: row.tags || [] })
    tagsInput.value = (row.tags || []).join(',')
  } else {
    formEdit.value = false
    tagsInput.value = ''
  }
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.name) { Message.warning('请输入姓名'); return }
  form.tags = tagsInput.value.split(',').map(s => s.trim()).filter(Boolean)
  formLoading.value = true
  try {
    if (formEdit.value) {
      await api.update(form.id, { ...form })
      Message.success('更新成功')
    } else {
      await api.create({ name: form.name, title: form.title, org: form.org, field: form.field, bio: form.bio, avatar_url: form.avatar_url, tags: form.tags, status: form.status })
      Message.success('创建成功')
    }
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (row) => {
  Modal.confirm({
    title: '删除专家',
    content: `确定删除专家"${row.name}"吗？`,
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

.avatar-upload { display: inline-block; cursor: pointer; }
</style>
