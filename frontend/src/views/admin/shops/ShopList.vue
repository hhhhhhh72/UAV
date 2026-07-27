<template>
  <div class="admin-page">
    <div class="page-header">
      <h2>商家管理</h2>
      <el-button type="primary" @click="handleAdd">新增商家</el-button>
    </div>

    <DataToolbar v-model="searchForm" :filters="filterConfig" @search="handleSearch" @reset="handleReset" />

    <el-table v-loading="loading" :data="tableData" border stripe>
      <el-table-column prop="id" label="ID" width="120" />
      <el-table-column prop="name" label="商家名称" min-width="180" />
      <el-table-column prop="license_url" label="营业执照" width="100">
        <template #default="{ row }">
          <el-image v-if="row.license_url" :src="row.license_url" style="width:48px;height:48px" fit="cover" />
          <span v-else class="text-muted">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="account_name" label="对公账户" width="160" />
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusMap[row.status]">{{ statusText[row.status] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="is_member" label="会员" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_member ? 'success' : 'info'">{{ row.is_member ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="入驻时间" width="170">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      layout="total, prev, pager, next, jumper"
      @change="loadData"
    />

    <!-- Add/Edit Dialog -->
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑商家' : '新增商家'" width="500px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="商家名称" required>
          <el-input v-model="form.name" placeholder="输入商家名称" />
        </el-form-item>
        <el-form-item label="营业执照">
          <el-input v-model="form.license_url" placeholder="营业执照图片 URL" />
        </el-form-item>
        <el-form-item label="对公账户">
          <el-input v-model="form.account_name" placeholder="对公账户名称" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="待审核" value="pending" />
            <el-option label="已批准" value="approved" />
            <el-option label="已驳回" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item label="协会会员">
          <el-switch v-model="form.is_member" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
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
const statusMap = { pending: 'warning', approved: 'success', rejected: 'danger' }
const statusText = { pending: '待审核', approved: '已批准', rejected: '已驳回' }

const dialog = reactive({ visible: false, isEdit: false, loading: false })
const form = reactive({ id: '', name: '', license_url: '', account_name: '', status: 'pending', is_member: false })

const formatDate = (d) => {
  if (!d) return '—'
  return new Date(d).toLocaleString('zh-CN')
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/admin/shops', {
      params: { page: pagination.page, page_size: pagination.pageSize, ...searchForm.value }
    })
    tableData.value = res.data.data || []
    pagination.total = res.data.total || 0
  } catch (e) {
    ElMessage.error('加载失败')
  } finally { loading.value = false }
}

const handleSearch = () => { pagination.page = 1; loadData() }
const handleReset = () => { searchForm.value = {}; handleSearch() }

const handleAdd = () => {
  dialog.isEdit = false
  dialog.visible = true
  Object.assign(form, { id: '', name: '', license_url: '', account_name: '', status: 'pending', is_member: false })
}

const handleEdit = (row) => {
  dialog.isEdit = true
  dialog.visible = true
  Object.assign(form, {
    id: row.id, name: row.name || '', license_url: row.license_url || '',
    account_name: row.account_name || '', status: row.status || 'pending', is_member: !!row.is_member
  })
}

const handleSubmit = async () => {
  if (!form.name) { ElMessage.warning('请输入商家名称'); return }
  dialog.loading = true
  try {
    if (dialog.isEdit) {
      await axios.put(`/api/v1/admin/shops/${form.id}`, {
        name: form.name, license_url: form.license_url, account_name: form.account_name,
        status: form.status, is_member: form.is_member
      })
      ElMessage.success('更新成功')
    } else {
      await axios.post('/api/v1/admin/shops', {
        name: form.name, license_url: form.license_url, account_name: form.account_name,
        status: form.status, is_member: form.is_member
      })
      ElMessage.success('创建成功')
    }
    dialog.visible = false
    loadData()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  } finally { dialog.loading = false }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确认删除商家「${row.name}」？`, '提示', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' })
    .then(async () => {
      await axios.delete(`/api/v1/admin/shops/${row.id}`)
      ElMessage.success('删除成功')
      loadData()
    })
    .catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.admin-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
.text-muted { color: #999; }
</style>
