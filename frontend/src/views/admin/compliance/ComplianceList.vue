<template>
  <div class="admin-page"><div class="page-header"><h2>合规文档</h2><el-button type="primary" @click="handleAdd">新增</el-button></div>
    <DataToolbar v-model="searchForm" :filters="filterConfig" @search="handleSearch" @reset="handleReset" />
    <el-tabs v-model="tab" @tab-change="loadData">
      <el-tab-pane label="合规文档" name="docs" /><el-tab-pane label="团体标准" name="standards" />
    </el-tabs>
    <el-table v-loading="loading" :data="tableData" border stripe>
      <el-table-column prop="id" label="ID" width="140" />
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column prop="status" label="状态" width="100"><template #default="{row}"><el-tag :type="statusMap[row.status]">{{row.status}}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="200"><template #default="{row}"><el-button size="small" @click="handleEdit(row)">编辑</el-button><el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button></template></el-table-column>
    </el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total,prev,pager,next,jumper" @change="loadData" />
  </div>
</template>
<script setup>
import { ref, reactive, onMounted } from 'vue'; import { ElMessage, ElMessageBox } from 'element-plus'; import DataToolbar from '../components/DataToolbar.vue'; import axios from '@/utils/http'
const tab = ref('docs'); const loading = ref(false); const tableData = ref([]); const searchForm = ref({})
const pagination = reactive({page:1,pageSize:20,total:0}); const filterConfig=[]
const statusMap = {published:'success',draft:'info',archived:'warning'}
const loadData = async () => {
  loading.value=true; try {
    const ep = tab.value==='docs'?'/api/v1/admin/compliance':'/api/v1/compliance-standards'
    const r = await axios.get(ep,{params:{...searchForm.value,page:pagination.page,page_size:pagination.pageSize}})
    tableData.value=r.data.data||[]; pagination.total=r.data.total||0
  } catch {ElMessage.error('load failed')} finally {loading.value=false}
}
const handleSearch=()=>{pagination.page=1;loadData()}
const handleReset=()=>{searchForm.value={};handleSearch()}
const handleAdd=()=>{ElMessage.info('TODO')}
const handleEdit=(r)=>{ElMessage.info('TODO: '+r.id)}
const handleDelete=(r)=>{ElMessageBox.confirm('Delete?','',{type:'warning'}).then(()=>ElMessage.info('TODO')).catch(()=>{})}
onMounted(loadData)
</script>
<style scoped>.admin-page{padding:20px}.page-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}h2{margin:0;font-size:20px}</style>
