<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索项目名称..."
          clearable
          style="width: 220px"
          @keyup.enter="onSearchSubmit"
          @clear="onSearchSubmit"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-select
          v-model="filterParams.status"
          placeholder="状态"
          clearable
          style="width: 140px"
          @change="onSearchSubmit"
        >
          <el-option label="全部" value="" />
          <el-option label="招募中" value="recruiting" />
          <el-option label="进行中" value="ongoing" />
          <el-option label="已完成" value="completed" />
        </el-select>

        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left: auto">
          <el-button type="success" @click="handleAdd">新增</el-button>
        </div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table
        v-loading="loading"
        :data="listData"
        row-key="id"
        stripe
        border
        @selection-change="onSelectChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="title" label="项目名称" min-width="200">
          <template #default="{ row }">
            <span class="cell-name">{{ row.title || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="lead_org" label="牵头单位" min-width="160" />
        <el-table-column prop="field" label="领域" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.field || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="budget_fen" label="预算" width="130" sortable="custom">
          <template #default="{ row }">
            <span class="cell-amount">{{ formatMoney(row.budget_fen) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="member_count" label="参与人数" width="100" sortable="custom" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无项目数据" />
        </template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

    <el-dialog v-model="detailVisible" title="项目详情" width="640px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="项目名称" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="牵头单位">{{ currentItem.lead_org || '-' }}</el-descriptions-item>
          <el-descriptions-item label="领域">{{ currentItem.field || '-' }}</el-descriptions-item>
          <el-descriptions-item label="预算">{{ formatMoney(currentItem.budget_fen) }}</el-descriptions-item>
          <el-descriptions-item label="参与人数">{{ currentItem.member_count ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="项目目标" :span="2">{{ currentItem.objectives || '-' }}</el-descriptions-item>
          <el-descriptions-item label="时间计划" :span="2">{{ currentItem.timeline || '-' }}</el-descriptions-item>
          <el-descriptions-item label="参与要求" :span="2">{{ currentItem.requirements || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('research-projects')

const statusLabel = (s) => ({ recruiting: '招募中', ongoing: '进行中', completed: '已完成' }[s] || s || '-')
const statusTag = (s) => ({ recruiting: 'warning', ongoing: '', completed: 'success' }[s] || 'info')

const formatMoney = (fen) => {
  if (fen == null) return '-'
  const yuan = Number(fen) / 100
  return '¥' + yuan.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { status: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',title:'',lead_org:'',field:'',budget_fen:0,member_count:0,status:'recruiting',objectives:'',timeline:'',requirements:''})
const resetForm=()=>Object.assign(form,{id:'',title:'',lead_org:'',field:'',budget_fen:0,member_count:0,status:'recruiting',objectives:'',timeline:'',requirements:''})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,{...r,budget_fen:r.budget_fen||0,member_count:r.member_count||0});formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.title){ElMessage.warning('请输入项目名称');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete=(r)=>{ElMessageBox.confirm(`确定删除项目"${r.title}"吗？`,'提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{})}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-name { font-weight: 500; color: var(--el-text-color-primary); }
.cell-amount { font-weight: 600; color: var(--el-color-danger); }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
