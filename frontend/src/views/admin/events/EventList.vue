<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索活动名称..." clearable style="width:220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.status" clearable style="width:140px" @change="onSearchSubmit">
          <el-option label="全部状态" value="" />
          <el-option label="即将开始" value="upcoming" />
          <el-option label="进行中" value="ongoing" />
          <el-option label="已结束" value="ended" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto"><el-button type="success" @click="handleAdd">新增活动</el-button></div>
      </div>
    </div>
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="title" label="活动名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="event_type" label="类型" width="100"><template #default="{row}"><el-tag size="small" :type="typeTag(row.event_type)">{{ row.event_type || '-' }}</el-tag></template></el-table-column>
        <el-table-column prop="location" label="地点" width="120" />
        <el-table-column prop="start_time" label="开始时间" width="160" sortable="custom"><template #default="{row}">{{ formatDate(row.start_time) }}</template></el-table-column>
        <el-table-column prop="capacity" label="名额" width="70" align="center" />
        <el-table-column prop="reg_count" label="已报名" width="80" align="center" />
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
        <template #empty><el-empty description="暂无活动数据" /></template>
      </el-table>
    </div>
    <div class="pagination-wrap" v-if="total>0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>
    <el-dialog v-model="detailVisible" title="活动详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="活动名称" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentItem.event_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地点">{{ currentItem.location || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatDate(currentItem.start_time) }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ formatDate(currentItem.end_time) }}</el-descriptions-item>
          <el-descriptions-item label="名额">{{ currentItem.capacity || '-' }}</el-descriptions-item>
          <el-descriptions-item label="已报名">{{ currentItem.reg_count || 0 }}</el-descriptions-item>
          <el-descriptions-item label="组织方" :span="2">{{ currentItem.organizer || '-' }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
    <el-dialog v-model="formVisible" :title="formEdit?'编辑活动':'新增活动'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16"><el-col :span="16"><el-form-item label="活动名称" required><el-input v-model="form.title"/></el-form-item></el-col><el-col :span="8"><el-form-item label="类型"><el-select v-model="form.event_type" style="width:100%"><el-option label="论坛" value="论坛"/><el-option label="走访" value="走访"/><el-option label="沙龙" value="沙龙"/><el-option label="培训" value="培训"/><el-option label="其他" value="其他"/></el-select></el-form-item></el-col></el-row>
        <el-row :gutter="16"><el-col :span="12"><el-form-item label="地点"><el-input v-model="form.location"/></el-form-item></el-col><el-col :span="12"><el-form-item label="组织方"><el-input v-model="form.organizer"/></el-form-item></el-col></el-row>
        <el-row :gutter="16"><el-col :span="12"><el-form-item label="开始时间"><el-input v-model="form.start_time" placeholder="YYYY-MM-DD HH:mm"/></el-form-item></el-col><el-col :span="12"><el-form-item label="结束时间"><el-input v-model="form.end_time" placeholder="YYYY-MM-DD HH:mm"/></el-form-item></el-col></el-row>
        <el-row :gutter="16"><el-col :span="8"><el-form-item label="名额"><el-input-number v-model="form.capacity" :min="0" style="width:100%"/></el-form-item></el-col><el-col :span="8"><el-form-item label="状态"><el-select v-model="form.status" style="width:100%"><el-option label="即将开始" value="upcoming"/><el-option label="进行中" value="ongoing"/><el-option label="已结束" value="ended"/><el-option label="已取消" value="cancelled"/></el-select></el-form-item></el-col></el-row>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="3"/></el-form-item>
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
const api = useAdminApi('events')
const formatDate = (d) => { if(!d) return '-'; const dt=new Date(d); const p=n=>String(n).padStart(2,'0'); return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}` }
const typeTag = (t) => ({ '论坛':'success', '走访':'', '沙龙':'warning', '培训':'info', '其他':'' }[t] || 'info')
const statusTag = (s) => ({ upcoming:'warning', ongoing:'success', ended:'info', cancelled:'danger' }[s] || 'info')
const { listData, loading, total, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({ apiFunction: api.list, idKey: 'id', defaultParams: { status: '' } })
const detailVisible = ref(false); const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',title:'',event_type:'论坛',location:'',start_time:'',end_time:'',capacity:0,organizer:'',status:'upcoming',description:''})
const resetForm=()=>Object.assign(form,{id:'',title:'',event_type:'论坛',location:'',start_time:'',end_time:'',capacity:0,organizer:'',status:'upcoming',description:''})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,{...r,capacity:r.capacity||0});formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.title){ElMessage.warning('请输入活动名称');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete=(r)=>{ElMessageBox.confirm('确定删除该活动?','提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{})}
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
