<template>
  <div class="admin-page">
    <div class="page-header">
      <h2>{{ title }}</h2>
      <el-button type="primary" @click="handleAdd">新增</el-button>
    </div>
    <DataToolbar v-model="searchForm" :filters="filterConfig" @search="handleSearch" @reset="handleReset" />
    <el-table v-loading="loading" :data="tableData" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><el-tag :type="statusMap[row.status]">{{ row.status }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total, prev, pager, next, jumper" @change="loadData" />
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
const statusMap = { active: 'success', pending: 'warning', inactive: 'info' }
const loadData = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/admin/events', { params: { ...searchForm.value, page: pagination.page, page_size: pagination.pageSize } })
    tableData.value = res.data.data || []
    pagination.total = res.data.total || 0
  } catch (e) { ElMessage.error('load failed') } finally { loading.value = false }
}
const handleSearch = () => { pagination.page = 1; loadData() }
const handleReset = () => { searchForm.value = {}; handleSearch() }
const handleAdd = () => { ElMessage.info('TODO: add dialog') }
const handleEdit = (row) => { ElMessage.info('TODO: edit ' + row.id) }
const handleDelete = (row) => {
  ElMessageBox.confirm('confirm delete?', 'Warning', { confirmButtonText: 'Delete', cancelButtonText: 'Cancel', type: 'warning' })
    .then(() => ElMessage.success('TODO: call delete API'))
    .catch(() => {})
}
onMounted(loadData)
</script>
<style scoped>
.admin-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
</style>
