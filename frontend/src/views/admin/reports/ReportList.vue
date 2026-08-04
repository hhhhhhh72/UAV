<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索报告标题..." clearable style="width:220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.report_type" clearable style="width:140px" @change="onSearchSubmit">
          <el-option label="全部类型" value="" />
          <el-option label="白皮书" value="whitepaper" />
          <el-option label="调研报告" value="research" />
          <el-option label="行业分析" value="analysis" />
          <el-option label="其他" value="other" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto"><el-button type="success" @click="handleAdd">新增报告</el-button></div>
      </div>
    </div>
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="title" label="报告标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="category" label="类型" width="110"><template #default="{row}"><el-tag :type="typeTag(row.category)" size="small">{{ typeLabel(row.category) }}</el-tag></template></el-table-column>
        <el-table-column prop="period" label="报告期" width="130" />
        <el-table-column prop="author" label="作者" width="140" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{row}"><el-tag :type="statusTag(row.status)" size="small">{{ row.status || '-' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{row}">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无报告数据" /></template>
      </el-table>
    </div>
    <div class="pagination-wrap" v-if="total>0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>
    <el-dialog v-model="detailVisible" title="报告详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="报告标题" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型"><el-tag :type="typeTag(currentItem.report_type)" size="small">{{ typeLabel(currentItem.report_type) }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="发布机构">{{ currentItem.publisher || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发布日期">{{ formatDate(currentItem.publish_date) }}</el-descriptions-item>
          <el-descriptions-item label="作者">{{ currentItem.authors || '-' }}</el-descriptions-item>
          <el-descriptions-item label="摘要" :span="2">{{ currentItem.abstract || '-' }}</el-descriptions-item>
          <el-descriptions-item label="文件" :span="2">
            <a v-if="currentItem.file_url" :href="currentItem.file_url" target="_blank">下载报告</a>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
    <el-dialog v-model="formVisible" :title="formEdit?'编辑报告':'新增报告'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16"><el-col :span="16"><el-form-item label="报告标题" required><el-input v-model="form.title"/></el-form-item></el-col><el-col :span="8"><el-form-item label="类型"><el-select v-model="form.category" style="width:100%"><el-option label="白皮书" value="whitepaper"/><el-option label="调研报告" value="research"/><el-option label="行业分析" value="analysis"/><el-option label="其他" value="other"/></el-select></el-form-item></el-col></el-row>
        <el-row :gutter="16"><el-col :span="12"><el-form-item label="报告期"><el-input v-model="form.period" placeholder="如 2026-Q2"/></el-form-item></el-col><el-col :span="12"><el-form-item label="作者"><el-input v-model="form.author"/></el-form-item></el-col></el-row>
        <el-form-item label="摘要"><el-input v-model="form.summary" type="textarea" rows="3"/></el-form-item>
        <el-form-item label="文件URL"><el-input v-model="form.file_url"/></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option label="草稿" value="draft"/><el-option label="已发布" value="published"/></el-select></el-form-item>
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
const api = useAdminApi('industry-reports')
const formatDate = (d) => { if(!d) return '-'; const dt=new Date(d); const p=n=>String(n).padStart(2,'0'); return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())}` }
const typeLabel = (t) => ({ whitepaper:'白皮书', research:'调研报告', analysis:'行业分析', other:'其他' }[t] || t || '-')
const typeTag = (t) => ({ whitepaper:'warning', research:'success', analysis:'', other:'info' }[t] || 'info')
const statusTag = (s) => ({ published:'success', draft:'warning' }[s] || 'info')
const { listData, loading, total, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({ apiFunction: api.list, idKey: 'id', defaultParams: { report_type: '' } })
const detailVisible = ref(false); const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',title:'',category:'whitepaper',period:'',author:'',summary:'',file_url:'',status:'draft'})
const resetForm=()=>Object.assign(form,{id:'',title:'',report_type:'whitepaper',publisher:'',publish_date:'',authors:'',abstract:'',file_url:'',status:'draft'})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,r);formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.title){ElMessage.warning('请输入报告标题');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
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
