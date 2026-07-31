<template>
  <div class="admin-page">
    <div class="page-header">
      <h2>专家管理</h2>
      <el-button type="primary" @click="handleAdd">新增专家</el-button>
    </div>
    <DataToolbar v-model="searchForm" :filters="filterConfig" @search="handleSearch" @reset="handleReset" />
    <el-table v-loading="loading" :data="tableData" border stripe>
      <el-table-column prop="id" label="ID" width="140" />
      <el-table-column prop="name" label="姓名" width="100" />
      <el-table-column prop="title" label="职称" width="120" />
      <el-table-column prop="org" label="所属单位" min-width="180" />
      <el-table-column prop="field" label="领域" width="120" />
      <el-table-column label="标签" width="200">
        <template #default="{ row }">
          <el-tag v-for="t in row.tags" :key="t" size="small" style="margin:2px">{{ t }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusMap[row.status]">{{ statusLabel[row.status] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total,prev,pager,next,jumper" @change="loadData" />

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑专家' : '新增专家'" width="500px" destroy-on-close>
      <el-form :model="form" label-width="80px">
        <el-form-item label="姓名" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="职称"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.org" /></el-form-item>
        <el-form-item label="领域"><el-input v-model="form.field" /></el-form-item>
        <el-form-item label="简介"><el-input v-model="form.bio" type="textarea" rows="3" /></el-form-item>
        <el-form-item label="头像">
          <el-upload class="avatar-upload" :action="uploadUrl" :headers="uploadHeaders" :show-file-list="false" :on-success="onUploadSuccess" :before-upload="beforeUpload" accept="image/*">
            <el-avatar v-if="form.avatar_url" :src="form.avatar_url" :size="80" shape="square" />
            <el-button v-else type="primary" plain>点击上传</el-button>
          </el-upload>
          <el-button v-if="form.avatar_url" size="small" style="margin-top:8px" @click="form.avatar_url=''">清除</el-button>
        </el-form-item>
        <el-form-item label="标签"><el-input v-model="tagsInput" placeholder="逗号分隔" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="待审核" value="pending" />
            <el-option label="已发布" value="published" />
            <el-option label="已下架" value="archived" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible=false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="dialog.loading">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import DataToolbar from '../components/DataToolbar.vue'
import axios from '@/utils/http'

const loading = ref(false)
const tableData = ref([])
const searchForm = ref({})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const filterConfig = []
const statusMap = { pending: 'warning', published: 'success', archived: 'info' }
const statusLabel = { pending: '待审核', published: '已发布', archived: '已下架' }
const tagsInput = ref('')
const dialog = reactive({ visible: false, isEdit: false, loading: false })
const form = reactive({ id: '', name: '', title: '', org: '', field: '', bio: '', avatar_url: '', status: 'pending', tags: [] })
const uploadUrl = '/api/v1/upload'
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}` }

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { ElMessage.error('只能上传图片文件'); return false }
  if (!isLt5M) { ElMessage.error('图片不能超过 5MB'); return false }
  return true
}

const onUploadSuccess = (res) => {
  form.avatar_url = res?.data?.url || res?.url || ''
  ElMessage.success('上传成功')
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/admin/experts', { params: { ...searchForm.value, page: pagination.page, page_size: pagination.pageSize } })
    const items = Array.isArray(res.data) ? res.data : (res.data?.data || [])
    tableData.value = items
    pagination.total = res.data?.total || items.length
  } catch (e) { ElMessage.error('load failed') } finally { loading.value = false }
}

const resetForm = () => Object.assign(form, { id: '', name: '', title: '', org: '', field: '', bio: '', avatar_url: '', status: 'pending', tags: [] })
const handleSearch = () => { pagination.page = 1; loadData() }
const handleReset = () => { searchForm.value = {}; handleSearch() }
const handleAdd = () => { resetForm(); tagsInput.value = ''; dialog.isEdit = false; dialog.visible = true }
const handleEdit = (row) => {
  Object.assign(form, { ...row, tags: row.tags || [] })
  tagsInput.value = (row.tags || []).join(',')
  dialog.isEdit = true; dialog.visible = true
}
const handleSubmit = async () => {
  if (!form.name) { ElMessage.warning('name required'); return }
  form.tags = tagsInput.value.split(',').map(s=>s.trim()).filter(Boolean)
  dialog.loading = true
  try {
    if (dialog.isEdit) {
      await axios.put(`/api/v1/admin/experts/${form.id}`, { ...form })
      ElMessage.success('updated')
    } else {
      await axios.post('/api/v1/admin/experts', { name: form.name, title: form.title, org: form.org, field: form.field, bio: form.bio, avatar_url: form.avatar_url, tags: form.tags, status: form.status })
      ElMessage.success('created')
    }
    dialog.visible = false; loadData()
  } catch (e) { ElMessage.error(e?.response?.data?.message || 'failed') } finally { dialog.loading = false }
}
const handleDelete = (row) => {
  ElMessageBox.confirm(`delete ${row.name}?`, 'Warning', { confirmButtonText: 'Delete', cancelButtonText: 'Cancel', type: 'warning' })
    .then(async () => { await axios.delete(`/api/v1/admin/experts/${row.id}`); ElMessage.success('deleted'); loadData() })
    .catch(() => {})
}
onMounted(loadData)
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
.avatar-upload { display: inline-block; cursor: pointer; }
</style>
