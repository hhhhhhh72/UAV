<template>
  <div class="list-page">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <el-tab-pane label="合规文档" name="docs" />
      <el-tab-pane label="团体标准" name="standards" />
    </el-tabs>
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索关键词..." clearable style="width: 220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.category" clearable placeholder="分类筛选" style="width: 160px" @change="onSearchSubmit">
          <el-option label="全部" value="" />
          <el-option v-if="activeTab === 'docs'" label="政策" value="政策" />
          <el-option v-if="activeTab === 'docs'" label="法规" value="法规" />
          <el-option v-if="activeTab === 'docs'" label="标准" value="标准" />
          <el-option v-if="activeTab === 'docs'" label="指南" value="指南" />
          <el-option v-if="activeTab === 'standards'" label="国家标准" value="国家标准" />
          <el-option v-if="activeTab === 'standards'" label="行业标准" value="行业标准" />
          <el-option v-if="activeTab === 'standards'" label="团体标准" value="团体标准" />
          <el-option v-if="activeTab === 'standards'" label="企业标准" value="企业标准" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto"><el-button type="success" @click="handleAdd">新增</el-button></div>
      </div>
    </div>
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange" @sort-change="onSortChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <template v-if="activeTab === 'docs'">
          <el-table-column prop="title" label="文档标题" min-width="200" />
          <el-table-column prop="category" label="分类" width="100">
            <template #default="{ row }"><el-tag :type="categoryColor[row.category] || 'info'">{{ row.category }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="publisher" label="发布机构" width="160" />
          <el-table-column prop="publish_date" label="发布日期" width="120">
            <template #default="{ row }">{{ formatDate(row.publish_date) }}</template>
          </el-table-column>
        </template>
        <template v-if="activeTab === 'standards'">
          <el-table-column prop="standard_no" label="标准编号" width="180" />
          <el-table-column prop="title" label="标准名称" min-width="200" />
          <el-table-column prop="publisher" label="发布机构" width="160" />
          <el-table-column prop="effective_date" label="生效日期" width="120">
            <template #default="{ row }">{{ formatDate(row.effective_date) }}</template>
          </el-table-column>
        </template>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusColor[row.status] || 'info'">{{ statusLabel[row.status] || row.status || '-' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无数据" /></template>
      </el-table>
    </div>
    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>
    <el-dialog v-model="detailVisible" :title="activeTab === 'docs' ? '文档详情' : '标准详情'" width="640px" destroy-on-close>
      <template v-if="currentItem && activeTab === 'docs'">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID" :span="2">{{ currentItem.id }}</el-descriptions-item>
          <el-descriptions-item label="文档标题" :span="2">{{ currentItem.title }}</el-descriptions-item>
          <el-descriptions-item label="分类"><el-tag :type="categoryColor[currentItem.category] || 'info'" size="small">{{ currentItem.category }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="发布机构">{{ currentItem.publisher || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发布日期">{{ formatDate(currentItem.publish_date) }}</el-descriptions-item>
          <el-descriptions-item label="状态"><el-tag :type="statusColor[currentItem.status] || 'info'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="摘要" :span="2">{{ currentItem.summary || '-' }}</el-descriptions-item>
          <el-descriptions-item label="附件" :span="2">
            <a v-if="currentItem.file_url" :href="currentItem.file_url" target="_blank" download class="download-link">下载文件</a>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
      </template>
      <template v-if="currentItem && activeTab === 'standards'">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID" :span="2">{{ currentItem.id }}</el-descriptions-item>
          <el-descriptions-item label="标准编号">{{ currentItem.standard_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="标准名称">{{ currentItem.title }}</el-descriptions-item>
          <el-descriptions-item label="发布机构">{{ currentItem.publisher || '-' }}</el-descriptions-item>
          <el-descriptions-item label="生效日期">{{ formatDate(currentItem.effective_date) }}</el-descriptions-item>
          <el-descriptions-item label="状态"><el-tag :type="statusColor[currentItem.status] || 'info'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="适用范围" :span="2">{{ currentItem.scope || '-' }}</el-descriptions-item>
          <el-descriptions-item label="附件" :span="2">
            <a v-if="currentItem.file_url" :href="currentItem.file_url" target="_blank" download class="download-link">下载文件</a>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <!-- Add/Edit Form Dialog -->
    <el-dialog v-model="formVisible" :title="formEdit ? '编辑' : '新增'" width="560px" destroy-on-close>
      <!-- Doc Form -->
      <el-form v-if="activeTab === 'docs'" :model="docForm" label-width="90px">
        <el-form-item label="文档标题" required><el-input v-model="docForm.title" /></el-form-item>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="分类"><el-select v-model="docForm.category" style="width:100%"><el-option v-for="c in ['政策','法规','标准','指南']" :key="c" :label="c" :value="c" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="发布机构"><el-input v-model="docForm.publisher" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="发布日期"><el-date-picker v-model="docForm.publish_date" type="date" placeholder="选择日期" format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="状态"><el-select v-model="docForm.status" style="width:100%"><el-option label="待审核" value="pending" /><el-option label="已发布" value="published" /><el-option label="草稿" value="draft" /></el-select></el-form-item></el-col>
        </el-row>
        <el-form-item label="摘要"><el-input v-model="docForm.summary" type="textarea" rows="2" /></el-form-item>
        <el-form-item label="附件文件">
          <el-upload class="file-upload" :action="uploadUrl" :headers="uploadHeaders" :show-file-list="false"
            :on-success="onDocFileSuccess" :before-upload="beforeUpload" accept="*">
            <el-button type="primary" plain>点击上传文件</el-button>
          </el-upload>
          <div v-if="docForm.file_url" style="margin-top:8px;font-size:12px;color:#666">
            已上传：{{ docForm.file_url.split('/').pop() }}
            <el-button size="small" @click="docForm.file_url=''">清除</el-button>
          </div>
        </el-form-item>
      </el-form>
      <!-- Standard Form -->
      <el-form v-else :model="stdForm" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="标准编号"><el-input v-model="stdForm.standard_no" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="标准名称" required><el-input v-model="stdForm.title" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="发布机构"><el-input v-model="stdForm.publisher" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="生效日期"><el-date-picker v-model="stdForm.effective_date" type="date" placeholder="选择日期" format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="状态"><el-select v-model="stdForm.status" style="width:200px"><el-option label="待审核" value="pending" /><el-option label="已发布" value="published" /><el-option label="草稿" value="draft" /></el-select></el-form-item>
        <el-form-item label="适用范围"><el-input v-model="stdForm.scope" type="textarea" rows="2" /></el-form-item>
        <el-form-item label="附件文件">
          <el-upload class="file-upload" :action="uploadUrl" :headers="uploadHeaders" :show-file-list="false"
            :on-success="onStdFileSuccess" :before-upload="beforeUpload" accept="*">
            <el-button type="primary" plain>点击上传文件</el-button>
          </el-upload>
          <div v-if="stdForm.file_url" style="margin-top:8px;font-size:12px;color:#666">
            已上传：{{ stdForm.file_url.split('/').pop() }}
            <el-button size="small" @click="stdForm.file_url=''">清除</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleFormSubmit" :loading="formLoading">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const activeTab = ref('docs')
const docsApi = useAdminApi('compliance-docs')
const standardsApi = useAdminApi('compliance-standards')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return '-'
  const p = n => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate())
}

const categoryColor = { '政策': 'warning', '法规': 'danger', '标准': 'success', '指南': 'info' }
const statusColor = { 'published': 'success', 'draft': 'info', 'archived': 'warning', 'pending': 'warning' }
const statusLabel = { 'published': '已发布', 'draft': '草稿', 'archived': '已下架', 'pending': '待审核' }

const currentApiFn = computed(() => activeTab.value === 'docs' ? docsApi.list : standardsApi.list)

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: currentApiFn,
  idKey: 'id',
  defaultParams: { category: '' }
})

const onTabChange = () => {
  resetParams()
}

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

// ---- Add/Edit form ----
const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const docForm = reactive({ id:'', title:'', category:'政策', publisher:'', publish_date:'', status:'published', summary:'', file_url:'' })
const stdForm = reactive({ id:'', standard_no:'', title:'', publisher:'', effective_date:'', status:'published', scope:'', file_url:'' })
const uploadUrl = '/api/v1/upload'
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}` }

const beforeUpload = (file) => {
  const maxSize = 20 * 1024 * 1024 // 20MB for documents
  if (file.size > maxSize) { ElMessage.error('文件不能超过 20MB'); return false }
  return true
}

const onDocFileSuccess = (res) => { docForm.file_url = res?.data?.url || res?.url || ''; ElMessage.success('上传成功') }
const onStdFileSuccess = (res) => { stdForm.file_url = res?.data?.url || res?.url || ''; ElMessage.success('上传成功') }
const resetDocForm = () => Object.assign(docForm, { id:'', title:'', category:'政策', publisher:'', publish_date:'', status:'published', summary:'', file_url:'' })
const resetStdForm = () => Object.assign(stdForm, { id:'', standard_no:'', title:'', publisher:'', effective_date:'', status:'published', scope:'', file_url:'' })
const handleAdd = () => {
  if (activeTab.value === 'docs') resetDocForm(); else resetStdForm()
  formEdit.value = false; formVisible.value = true
}
const handleEdit = (row) => {
  if (activeTab.value === 'docs') Object.assign(docForm, row)
  else Object.assign(stdForm, row)
  formEdit.value = true; formVisible.value = true
}
const handleFormSubmit = async () => {
  const api = activeTab.value === 'docs' ? docsApi : standardsApi
  const payload = activeTab.value === 'docs' ? { ...docForm } : { ...stdForm }
  if (!payload.title) { ElMessage.warning('请输入标题'); return }
  formLoading.value = true
  try {
    if (formEdit.value) { await api.update(payload.id, payload); ElMessage.success('更新成功') }
    else { await api.create(payload); ElMessage.success('创建成功') }
    formVisible.value = false; loadData()
  } catch (e) { console.error('[compliance]', e?.response?.status, e?.response?.data || e); ElMessage.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (row) => {
  ElMessageBox.confirm('确定删除该项吗？', '提示', { type: 'warning' }).then(async () => {
    try { const api = activeTab.value === 'docs' ? docsApi : standardsApi; await api.delete(row.id); ElMessage.success('已删除'); loadData() } catch { ElMessage.error('删除失败') }
  }).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.download-link { color: #409eff; text-decoration: none; }
.download-link:hover { text-decoration: underline; }
.file-upload { display: inline-block; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
