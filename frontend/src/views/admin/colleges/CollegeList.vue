<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索院校名称..." clearable style="width: 220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.status" clearable style="width: 160px" @change="onSearchSubmit">
          <el-option label="全部" value="" />
          <el-option label="合作中" value="active" />
          <el-option label="待合作" value="pending" />
          <el-option label="已终止" value="closed" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto">
          <el-button type="success" @click="handleAdd">新增</el-button>
        </div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="name" label="院校名称" min-width="200">
          <template #default="{ row }">
            <span class="cell-title">{{ row.name || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="college_type" label="类型" width="120">
          <template #default="{ row }">{{ row.college_type || '-' }}</template>
        </el-table-column>
        <el-table-column prop="location" label="地区" width="120" />
        <el-table-column prop="major" label="特色专业" min-width="160">
          <template #default="{ row }">{{ row.major || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="合作状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无院校数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>

    <el-dialog v-model="detailVisible" title="院校详情" width="600px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="院校名称" :span="2">{{ currentItem.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentItem.college_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地区">{{ currentItem.location || '-' }}</el-descriptions-item>
          <el-descriptions-item label="合作状态">
            <el-tag :type="statusTag(currentItem.status)" size="small">{{ currentItem.status || '-' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="官网" :span="2">
            <span v-if="currentItem.website">{{ currentItem.website }}</span>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="特色专业" :span="2">{{ currentItem.majors || currentItem.major || '-' }}</el-descriptions-item>
          <el-descriptions-item label="简介" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系方式">{{ currentItem.contact || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(currentItem.created_at) }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <el-dialog v-model="formVisible" :title="formEdit?'编辑院校':'新增院校'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="14"><el-form-item label="院校名称" required><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="10"><el-form-item label="类型"><el-select v-model="form.college_type" style="width:100%"><el-option label="本科" value="本科" /><el-option label="专科" value="专科" /><el-option label="职校" value="职校" /><el-option label="研究机构" value="研究机构" /></el-select></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="地区"><el-input v-model="form.location" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="合作状态"><el-select v-model="form.status" style="width:100%"><el-option label="合作中" value="active" /><el-option label="洽谈中" value="pending" /><el-option label="已结束" value="closed" /></el-select></el-form-item></el-col>
        </el-row>
        <el-form-item label="特色专业"><el-input v-model="form.major" /></el-form-item>
        <el-form-item label="官网"><el-input v-model="form.website" /></el-form-item>
        <el-form-item label="联系方式"><el-input v-model="form.contact" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="2" /></el-form-item>
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

const api = useAdminApi('colleges')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ 'active': 'success', 'pending': 'warning', 'closed': 'info' }[s] || 'info')

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible=ref(false); const formEdit=ref(false); const formLoading=ref(false)
const form=reactive({id:'',name:'',college_type:'',location:'',major:'',website:'',contact:'',status:'pending',description:''})
const resetForm=()=>Object.assign(form,{id:'',name:'',college_type:'',location:'',major:'',website:'',contact:'',status:'pending',description:''})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,r);formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.name){ElMessage.warning('请输入院校名称');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete=(r)=>{ElMessageBox.confirm('确定删除该院校?','提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{})}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.cell-title { font-weight: 500; color: var(--el-text-color-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; max-width: 300px; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
