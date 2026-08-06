<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索资源名称..." clearable style="width:220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.res_type" clearable style="width:140px" @change="onSearchSubmit">
          <el-option label="全部类型" value="" />
          <el-option label="无人机" value="drone" />
          <el-option label="通信" value="comm" />
          <el-option label="照明" value="light" />
          <el-option label="运输" value="transport" />
          <el-option label="其他" value="other" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto"><el-button type="success" @click="handleAdd">新增资源</el-button></div>
      </div>
    </div>
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="name" label="资源名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="res_type" label="类型" width="100"><template #default="{row}"><el-tag :type="typeTag(row.res_type)" size="small">{{ typeLabel(row.res_type) }}</el-tag></template></el-table-column>
        <el-table-column prop="specs" label="规格" width="140" />
        <el-table-column prop="quantity" label="数量" width="70" align="center" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{row}"><el-tag :type="statusTag(row.status)" size="small">{{ row.status || '-' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{row}">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无应急资源" /></template>
      </el-table>
    </div>
    <div class="pagination-wrap" v-if="total>0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>
    <el-dialog v-model="detailVisible" title="应急资源详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="资源名称" :span="2">{{ currentItem.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型"><el-tag :type="typeTag(currentItem.res_type)" size="small">{{ typeLabel(currentItem.res_type) }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="规格">{{ currentItem.specs || '-' }}</el-descriptions-item>
          <el-descriptions-item label="数量">{{ currentItem.quantity || 0 }}</el-descriptions-item>
          <el-descriptions-item label="位置">{{ currentItem.location || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ currentItem.contact_info || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态"><el-tag :type="statusTag(currentItem.status)" size="small">{{ currentItem.status || '-' }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ currentItem.notes || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
    <el-dialog v-model="formVisible" :title="formEdit?'编辑应急资源':'新增应急资源'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16"><el-col :span="14"><el-form-item label="资源名称" required><el-input v-model="form.name"/></el-form-item></el-col><el-col :span="10"><el-form-item label="类型"><el-select v-model="form.res_type" style="width:100%"><el-option label="无人机" value="drone"/><el-option label="通信" value="comm"/><el-option label="照明" value="light"/><el-option label="运输" value="transport"/><el-option label="其他" value="other"/></el-select></el-form-item></el-col></el-row>
        <el-row :gutter="16"><el-col :span="12"><el-form-item label="规格"><el-input v-model="form.specs"/></el-form-item></el-col><el-col :span="12"><el-form-item label="数量"><el-input-number v-model="form.quantity" :min="0" style="width:100%"/:controls="false"></el-form-item></el-col></el-row>
        <el-form-item label="位置"><el-input v-model="form.location"/></el-form-item>
        <el-form-item label="联系人信息"><el-input v-model="form.contact_info" placeholder="姓名 / 电话，如：张工 13800138000"/></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option label="可用" value="available"/><el-option label="使用中" value="in_use"/><el-option label="维护中" value="maintenance"/></el-select></el-form-item>
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
const api = useAdminApi('emergency-resources')
const typeLabel = (t) => ({ drone:'无人机', comm:'通信', light:'照明', transport:'运输', other:'其他' }[t] || t || '-')
const typeTag = (t) => ({ drone:'success', comm:'warning', light:'info', transport:'', other:'' }[t] || 'info')
const statusTag = (s) => ({ available:'success', in_use:'warning', maintenance:'danger' }[s] || 'info')
const { listData, loading, total, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({ apiFunction: api.list, idKey: 'id', defaultParams: { res_type: '' } })
const detailVisible = ref(false); const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',name:'',res_type:'drone',specs:'',quantity:0,location:'',contact_info:'',status:'available'})
const resetForm=()=>Object.assign(form,{id:'',name:'',res_type:'drone',specs:'',quantity:0,location:'',contact_info:'',status:'available'})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,{...r,quantity:r.quantity||0});formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.name){ElMessage.warning('请输入资源名称');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete=(r)=>{ElMessageBox.confirm('确定删除?','提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{})}
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
