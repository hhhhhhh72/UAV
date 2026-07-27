<template>
  <div class="admin-page">
    <div class="page-header"><h2>资源管理</h2><el-button type="primary" @click="handleAdd">新增资源</el-button></div>
    <DataToolbar v-model="searchForm" :filters="filterConfig" @search="handleSearch" @reset="handleReset" />
    <el-table v-loading="loading" :data="tableData" border stripe>
      <el-table-column prop="id" label="ID" width="140" />
      <el-table-column prop="name" label="名称" min-width="180" />
      <el-table-column prop="res_type" label="类型" width="120" />
      <el-table-column prop="model" label="型号/规格" width="160" />
      <el-table-column prop="location" label="位置" width="160" />
      <el-table-column prop="price_fen" label="价格(分)" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><el-tag :type="statusMap[row.status]">{{ row.status }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }"><el-button size="small" @click="handleEdit(row)">编辑</el-button><el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button></template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total,prev,pager,next,jumper" @change="loadData" />
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑资源' : '新增资源'" width="500px" destroy-on-close>
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="类型"><el-input v-model="form.res_type" /></el-form-item>
        <el-form-item label="规格"><el-input v-model="form.specs" /></el-form-item>
        <el-form-item label="位置"><el-input v-model="form.location" /></el-form-item>
        <el-form-item label="价格(分)"><el-input-number v-model="form.price_fen" /></el-form-item>
        <el-form-item label="预约信息"><el-input v-model="form.booking_info" type="textarea" rows="2" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option label="可用" value="available" /><el-option label="不可用" value="unavailable" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog.visible=false">取消</el-button><el-button type="primary" @click="handleSubmit" :loading="dialog.loading">提交</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import DataToolbar from '../components/DataToolbar.vue'
import axios from '@/utils/http'
const loading = ref(false); const tableData = ref([]); const searchForm = ref({})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 }); const filterConfig = []
const statusMap = { available: 'success', unavailable: 'danger' }
const dialog = reactive({ visible: false, isEdit: false, loading: false })
const form = reactive({ id: '', name: '', res_type: '', specs: '', location: '', price_fen: 0, booking_info: '', status: 'available' })
const loadData = async () => { loading.value = true; try { const r = await axios.get('/api/v1/admin/resources', { params: { ...searchForm.value, page: pagination.page, page_size: pagination.pageSize } }); tableData.value = r.data.data || []; pagination.total = r.data.total || 0 } catch { ElMessage.error('load failed') } finally { loading.value = false } }
const handleSearch = () => { pagination.page = 1; loadData() }
const handleReset = () => { searchForm.value = {}; handleSearch() }
const handleAdd = () => { Object.assign(form, { id: '', name: '', res_type: '', specs: '', location: '', price_fen: 0, booking_info: '', status: 'available' }); dialog.isEdit = false; dialog.visible = true }
const handleEdit = (row) => { Object.assign(form, { ...row }); dialog.isEdit = true; dialog.visible = true }
const handleSubmit = async () => {
  if (!form.name) { ElMessage.warning('name required'); return }; dialog.loading = true
  try {
    dialog.isEdit ? await axios.put(`/api/v1/admin/industry-resources/${form.id}`, { ...form }) : await axios.post('/api/v1/admin/industry-resources', { ...form })
    ElMessage.success(dialog.isEdit ? 'updated' : 'created'); dialog.visible = false; loadData()
  } catch (e) { ElMessage.error(e?.response?.data?.message || 'failed') } finally { dialog.loading = false }
}
const handleDelete = (row) => { ElMessageBox.confirm(`delete ${row.name}?`, 'Warning', { type: 'warning' }).then(async () => { await axios.delete(`/api/v1/admin/industry-resources/${id}`, id = row.id); ElMessage.success('deleted'); loadData() }).catch(() => {}) }
onMounted(loadData)
</script>
<style scoped>.admin-page { padding: 20px; } .page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; } .page-header h2 { margin: 0; font-size: 20px; }</style>
