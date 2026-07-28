<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索企业或品牌名称..." clearable style="width:240px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.status" clearable style="width:140px" @change="onSearchSubmit">
          <el-option label="全部状态" value="" />
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已驳回" value="rejected" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto"><el-button type="success" @click="handleAdd">新增品牌</el-button></div>
      </div>
    </div>
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="company" label="企业名称" width="160" />
        <el-table-column prop="brand_name" label="品牌名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="industry" label="行业" width="120" />
        <el-table-column prop="portfolio_type" label="展示类型" width="110" />
        <el-table-column prop="status" label="审核状态" width="100">
          <template #default="{row}"><el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{row}">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button v-if="row.status==='pending'" link type="success" size="small" @click="handleApprove(row)">通过</el-button>
            <el-button v-if="row.status==='pending'" link type="danger" size="small" @click="handleReject(row)">驳回</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无品牌数据" /></template>
      </el-table>
    </div>
    <div class="pagination-wrap" v-if="total>0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>
    <el-dialog v-model="detailVisible" title="品牌详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="企业名称">{{ currentItem.company || '-' }}</el-descriptions-item>
          <el-descriptions-item label="品牌名称">{{ currentItem.brand_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="行业">{{ currentItem.industry || '-' }}</el-descriptions-item>
          <el-descriptions-item label="展示类型">{{ currentItem.portfolio_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核状态"><el-tag :type="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="Logo" :span="2">{{ currentItem.logo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="封面图" :span="2">{{ currentItem.cover_image || '-' }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="荣誉" :span="2">{{ currentItem.honors || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
    <el-dialog v-model="formVisible" :title="formEdit?'编辑品牌':'新增品牌'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16"><el-col :span="12"><el-form-item label="企业名称"><el-input v-model="form.company"/></el-form-item></el-col><el-col :span="12"><el-form-item label="品牌名称" required><el-input v-model="form.brand_name"/></el-form-item></el-col></el-row>
        <el-row :gutter="16"><el-col :span="12"><el-form-item label="行业"><el-input v-model="form.industry"/></el-form-item></el-col><el-col :span="12"><el-form-item label="展示类型"><el-input v-model="form.portfolio_type"/></el-form-item></el-col></el-row>
        <el-form-item label="Logo URL"><el-input v-model="form.logo"/></el-form-item>
        <el-form-item label="封面图 URL"><el-input v-model="form.cover_image"/></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="2"/></el-form-item>
        <el-form-item label="荣誉"><el-input v-model="form.honors" type="textarea" rows="2"/></el-form-item>
        <el-form-item label="审核状态"><el-select v-model="form.status"><el-option label="待审核" value="pending"/><el-option label="已通过" value="approved"/><el-option label="已驳回" value="rejected"/></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="formVisible=false">取消</el-button><el-button type="primary" @click="submitForm" :loading="formLoading">提交</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'
const api = useAdminApi('portfolios')
const statusLabel = (s) => ({ pending:'待审核', approved:'已通过', rejected:'已驳回' }[s] || s || '-')
const statusTag = (s) => ({ pending:'warning', approved:'success', rejected:'danger' }[s] || 'info')
const { listData, loading, total, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({ apiFunction: api.list, idKey: 'id', defaultParams: { status: '' } })
const detailVisible = ref(false); const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',company:'',brand_name:'',industry:'',portfolio_type:'',logo:'',cover_image:'',description:'',honors:'',status:'pending'})
const resetForm=()=>Object.assign(form,{id:'',company:'',brand_name:'',industry:'',portfolio_type:'',logo:'',cover_image:'',description:'',honors:'',status:'pending'})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,r);formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.brand_name){ElMessage.warning('请输入品牌名称');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleApprove = async (r) => { try { await api.update(r.id, { status: 'approved' }); ElMessage.success('已通过'); loadData() } catch { ElMessage.error('操作失败') } }
const handleReject = async (r) => { try { await api.update(r.id, { status: 'rejected' }); ElMessage.success('已驳回'); loadData() } catch { ElMessage.error('操作失败') } }
const handleDelete = (r) => { ElMessageBox.confirm('确定删除?','提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{}) }
onMounted(loadData)
</script>

<style scoped>
.list-page { max-width:1400px; margin:0 auto }
.search-bar { background:#fff; border-radius:8px; padding:16px 20px; margin-bottom:16px; box-shadow:0 1px 3px rgba(0,0,0,.06) }
.search-row { display:flex; align-items:center; gap:12px; flex-wrap:wrap }
.table-wrap { background:#fff; border-radius:8px; box-shadow:0 1px 3px rgba(0,0,0,.06); overflow:hidden }
.pagination-wrap { display:flex; justify-content:flex-end; margin-top:16px; background:#fff; border-radius:8px; padding:16px 20px; box-shadow:0 1px 3px rgba(0,0,0,.06) }
@media(max-width:767px){ .search-bar{padding:12px} .search-row{flex-direction:column;align-items:stretch} .table-wrap{overflow-x:auto} }
</style>
