<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索成果名称..."
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
          v-model="filterParams.field"
          placeholder="领域"
          clearable
          style="width: 140px"
          @change="onSearchSubmit"
        >
          <el-option label="全部" value="" />
          <el-option label="无人机平台" value="无人机平台" />
          <el-option label="飞控系统" value="飞控系统" />
          <el-option label="导航与定位" value="导航与定位" />
          <el-option label="通信链路" value="通信链路" />
          <el-option label="载荷与传感器" value="载荷与传感器" />
          <el-option label="能源动力" value="能源动力" />
          <el-option label="人工智能" value="人工智能" />
          <el-option label="新材料" value="新材料" />
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
        <el-table-column prop="title" label="成果名称" min-width="180">
          <template #default="{ row }">
            <span class="cell-name">{{ row.title || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="field" label="领域" min-width="120" />
        <el-table-column prop="field" label="领域" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.field || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="stage" label="阶段" width="100">
          <template #default="{ row }">
            <el-tag :type="stageTag(row.stage)" size="small">{{ stageLabel(row.stage) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="achieve_type" label="成果类型" width="120" />
        <el-table-column prop="created_at" label="提交时间" width="170" sortable="custom">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
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
          <el-empty description="暂无成果数据" />
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

    <el-dialog v-model="detailVisible" title="成果详情" width="640px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="成果名称" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="领域">{{ currentItem.field || '-' }}</el-descriptions-item>
          <el-descriptions-item label="成果类型">{{ currentItem.achieve_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="所处阶段">
            <el-tag :type="stageTag(currentItem.stage)" size="small">{{ stageLabel(currentItem.stage) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="成果描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="附件" :span="2">{{ (currentItem.attachments || []).length }} 份</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <el-dialog v-model="formVisible" :title="formEdit?'编辑成果':'新增成果'" width="560px" @close="resetForm">
      <el-form :model="form" label-width="80px">
        <el-form-item label="成果名称"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="领域"><el-input v-model="form.field" /></el-form-item>
        <el-form-item label="成果类型"><el-input v-model="form.achieve_type" placeholder="如：专利 / 样机 / 技术方案" /></el-form-item>
        <el-form-item label="阶段">
          <el-select v-model="form.stage" style="width:100%">
            <el-option label="实验室" value="lab" /><el-option label="中试" value="pilot" />
            <el-option label="产业化" value="industrialization" /><el-option label="上市" value="launched" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="附件资料">
          <div style="width:100%">
            <div v-for="(at, i) in form.attachments" :key="i" style="display:flex;gap:6px;margin-bottom:6px">
              <el-input v-model="at.name" placeholder="附件名" style="width:40%" />
              <el-input v-model="at.size" placeholder="大小" style="width:20%" />
              <el-input v-model="at.url" placeholder="/uploads/xxx.pdf" style="flex:1" />
              <el-button type="danger" :icon="Delete" circle size="small" @click="form.attachments.splice(i,1)" />
            </div>
            <el-button type="primary" plain size="small" @click="form.attachments.push({name:'',size:'',url:''})">+ 添加附件</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible=false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="formLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('achievements')

const stageLabel = (s) => ({ lab: '实验室', pilot: '中试', industrialization: '产业化', launched: '上市' }[s] || s || '-')
const stageTag = (s) => ({ lab: 'info', pilot: 'warning', industrialization: 'success', launched: '' }[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { field: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',title:'',field:'',stage:'lab',achieve_type:'',description:''})
const resetForm=()=>Object.assign(form,{id:'',title:'',field:'',stage:'lab',achieve_type:'',description:'',attachments:[]})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,r);formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.title){ElMessage.warning('请输入成果名称');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete=(r)=>{ElMessageBox.confirm(`确定删除成果"${r.title}"吗？`,'提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{})}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-name { font-weight: 500; color: var(--el-text-color-primary); }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
