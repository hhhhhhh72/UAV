<template>
  <div class="page">
    <!-- Tab 切换 + 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-tabs v-model:active-key="activeTab" @change="onTabChange" class="list-tabs">
        <a-tab-pane title="合规文档" key="docs" />
        <a-tab-pane title="团体标准" key="standards" />
      </a-tabs>
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input
              v-model="filterParams.keyword"
              placeholder="搜索关键词..."
              allow-clear
              style="width: 220px"
              @press-enter="onSearchSubmit"
              @clear="onSearchSubmit"
            />
          </a-form-item>
          <a-form-item label="分类" class="form-item">
            <a-select v-model="filterParams.category" style="width: 160px" allow-clear placeholder="分类筛选" @change="onSearchSubmit">
              <a-option label="全部" value="" />
              <a-option v-if="activeTab === 'docs'" label="政策" value="政策" />
              <a-option v-if="activeTab === 'docs'" label="法规" value="法规" />
              <a-option v-if="activeTab === 'docs'" label="标准" value="标准" />
              <a-option v-if="activeTab === 'docs'" label="指南" value="指南" />
              <a-option v-if="activeTab === 'standards'" label="国家标准" value="国家标准" />
              <a-option v-if="activeTab === 'standards'" label="行业标准" value="行业标准" />
              <a-option v-if="activeTab === 'standards'" label="团体标准" value="团体标准" />
              <a-option v-if="activeTab === 'standards'" label="企业标准" value="企业标准" />
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>搜索</a-button>
          <a-button @click="resetParams">重置</a-button>
          <div style="margin-left: auto"><a-button type="primary" @click="handleAdd">新增</a-button></div>
        </a-space>
      </a-form>
    </a-card>

    <!-- 数据表格 -->
    <a-card :bordered="false">
      <a-table
        :columns="tableColumns"
        :data="listData"
        :loading="loading"
        row-key="id"
        :pagination="false"
        :row-selection="rowSelection"
        @sorter-change="handleSorterChange"
      >
        <template #category="{ record }">
          <a-tag :color="categoryColor[record.category] || 'gray'" size="small">{{ record.category }}</a-tag>
        </template>
        <template #publishDate="{ record }">
          <span class="time-text">{{ formatDate(record.publish_date) }}</span>
        </template>
        <template #effectiveDate="{ record }">
          <span class="time-text">{{ formatDate(record.effective_date) }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusColor[record.status] || 'gray'" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无数据" />
        </template>
      </a-table>

      <div class="pagination-wrap" v-if="total > 0">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" :title="activeTab === 'docs' ? '文档详情' : '标准详情'" :width="640" :footer="false">
      <template v-if="currentItem && activeTab === 'docs'">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID" :span="2">{{ currentItem.id }}</a-descriptions-item>
          <a-descriptions-item label="文档标题" :span="2">{{ currentItem.title }}</a-descriptions-item>
          <a-descriptions-item label="分类">
            <a-tag :color="categoryColor[currentItem.category] || 'gray'" size="small">{{ currentItem.category }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发布机构">{{ currentItem.publisher || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发布日期">{{ formatDate(currentItem.publish_date) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusColor[currentItem.status] || 'gray'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="摘要" :span="2">{{ currentItem.summary || '-' }}</a-descriptions-item>
          <a-descriptions-item label="附件" :span="2">
            <a v-if="currentItem.file_url" :href="currentItem.file_url" target="_blank" download class="download-link">下载文件</a>
            <span v-else>-</span>
          </a-descriptions-item>
        </a-descriptions>
      </template>
      <template v-if="currentItem && activeTab === 'standards'">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID" :span="2">{{ currentItem.id }}</a-descriptions-item>
          <a-descriptions-item label="标准编号">{{ currentItem.standard_no || '-' }}</a-descriptions-item>
          <a-descriptions-item label="标准名称">{{ currentItem.title }}</a-descriptions-item>
          <a-descriptions-item label="发布机构">{{ currentItem.publisher || '-' }}</a-descriptions-item>
          <a-descriptions-item label="生效日期">{{ formatDate(currentItem.effective_date) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusColor[currentItem.status] || 'gray'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="适用范围" :span="2">{{ currentItem.scope || '-' }}</a-descriptions-item>
          <a-descriptions-item label="附件" :span="2">
            <a v-if="currentItem.file_url" :href="currentItem.file_url" target="_blank" download class="download-link">下载文件</a>
            <span v-else>-</span>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增 / 编辑表单弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑' : '新增'" :width="560">
      <!-- 文档表单 -->
      <a-form v-if="activeTab === 'docs'" :model="docForm" layout="horizontal" class="dialog-form">
        <a-form-item label="文档标题" required>
          <a-input v-model="docForm.title" allow-clear />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="分类">
              <a-select v-model="docForm.category" style="width: 100%">
                <a-option v-for="c in ['政策', '法规', '标准', '指南']" :key="c" :label="c" :value="c" />
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="发布机构">
              <a-input v-model="docForm.publisher" allow-clear />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="发布日期">
              <a-date-picker v-model="docForm.publish_date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态">
              <a-select v-model="docForm.status" style="width: 100%">
                <a-option label="待审核" value="pending" />
                <a-option label="已发布" value="published" />
                <a-option label="草稿" value="draft" />
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="摘要">
          <a-input v-model="docForm.summary" type="textarea" :auto-size="{ minRows: 2, maxRows: 6 }" />
        </a-form-item>
        <a-form-item label="附件文件">
          <a-upload
            class="file-upload"
            :show-file-list="false"
            :custom-request="uploadFile"
            :before-upload="beforeUpload"
          >
            <a-button type="primary">点击上传文件</a-button>
          </a-upload>
          <div v-if="docForm.file_url" class="file-info">
            已上传：{{ docForm.file_url.split('/').pop() }}
            <a-button size="small" @click="docForm.file_url = ''">清除</a-button>
          </div>
        </a-form-item>
      </a-form>
      <!-- 标准表单 -->
      <a-form v-else :model="stdForm" layout="horizontal" class="dialog-form">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="标准编号">
              <a-input v-model="stdForm.standard_no" allow-clear />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="标准名称" required>
              <a-input v-model="stdForm.title" allow-clear />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="发布机构">
              <a-input v-model="stdForm.publisher" allow-clear />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="生效日期">
              <a-date-picker v-model="stdForm.effective_date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="状态">
          <a-select v-model="stdForm.status" style="width: 200px">
            <a-option label="待审核" value="pending" />
            <a-option label="已发布" value="published" />
            <a-option label="草稿" value="draft" />
          </a-select>
        </a-form-item>
        <a-form-item label="适用范围">
          <a-input v-model="stdForm.scope" type="textarea" :auto-size="{ minRows: 2, maxRows: 6 }" />
        </a-form-item>
        <a-form-item label="附件文件">
          <a-upload
            class="file-upload"
            :show-file-list="false"
            :custom-request="uploadFile"
            :before-upload="beforeUpload"
          >
            <a-button type="primary">点击上传文件</a-button>
          </a-upload>
          <div v-if="stdForm.file_url" class="file-info">
            已上传：{{ stdForm.file_url.split('/').pop() }}
            <a-button size="small" @click="stdForm.file_url = ''">清除</a-button>
          </div>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="handleFormSubmit">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import axios from '@/utils/http'
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

const categoryColor = { '政策': 'orange', '法规': 'red', '标准': 'green', '指南': 'gray' }
const statusColor = { 'published': 'green', 'draft': 'gray', 'archived': 'orange', 'pending': 'orange' }
const statusLabel = { 'published': '已发布', 'draft': '草稿', 'archived': '已下架', 'pending': '待审核' }

const currentApiFn = computed(() => activeTab.value === 'docs' ? docsApi.list : standardsApi.list)

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: currentApiFn,
  idKey: 'id',
  defaultParams: { category: '' }
})

const onTabChange = () => {
  resetParams()
}

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

// a-table 排序（Arco direction → useListRequest 的 order 语义）
const handleSorterChange = (dataIndex, direction) => {
  const order = direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  onSortChange({ prop: dataIndex, order })
}

// 按当前 Tab 生成列定义
const tableColumns = computed(() => {
  const idCol = { title: 'ID', dataIndex: 'id', width: 160, sortable: true }
  const statusCol = { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 }
  const actionsCol = { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
  if (activeTab.value === 'docs') {
    return [
      idCol,
      { title: '文档标题', dataIndex: 'title', minWidth: 200 },
      { title: '分类', dataIndex: 'category', slotName: 'category', width: 100 },
      { title: '发布机构', dataIndex: 'publisher', width: 160 },
      { title: '发布日期', dataIndex: 'publish_date', slotName: 'publishDate', width: 120 },
      statusCol, actionsCol
    ]
  }
  return [
    idCol,
    { title: '标准编号', dataIndex: 'standard_no', width: 180 },
    { title: '标准名称', dataIndex: 'title', minWidth: 200 },
    { title: '发布机构', dataIndex: 'publisher', width: 160 },
    { title: '生效日期', dataIndex: 'effective_date', slotName: 'effectiveDate', width: 120 },
    statusCol, actionsCol
  ]
})

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

// ---- Add/Edit form ----
const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const docForm = reactive({ id: '', title: '', category: '政策', publisher: '', publish_date: '', status: 'published', summary: '', file_url: '' })
const stdForm = reactive({ id: '', standard_no: '', title: '', publisher: '', effective_date: '', status: 'published', scope: '', file_url: '' })

const beforeUpload = (file) => {
  const maxSize = 20 * 1024 * 1024 // 20MB for documents
  if (file.size > maxSize) { Message.error('文件不能超过 20MB'); return false }
  return true
}

// 附件上传（/api/v1/upload 返回相对 URL，按当前 Tab 写入对应表单）
const uploadFile = async ({ file, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await axios.post('/api/v1/upload', fd)
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    if (activeTab.value === 'docs') docForm.file_url = url
    else stdForm.file_url = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error('上传失败')
  }
}

const resetDocForm = () => Object.assign(docForm, { id: '', title: '', category: '政策', publisher: '', publish_date: '', status: 'published', summary: '', file_url: '' })
const resetStdForm = () => Object.assign(stdForm, { id: '', standard_no: '', title: '', publisher: '', effective_date: '', status: 'published', scope: '', file_url: '' })
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
  if (!payload.title) { Message.warning('请输入标题'); return }
  formLoading.value = true
  try {
    if (formEdit.value) { await api.update(payload.id, payload); Message.success('更新成功') }
    else { await api.create(payload); Message.success('创建成功') }
    formVisible.value = false; loadData()
  } catch (e) { console.error('[compliance]', e?.response?.status, e?.response?.data || e); Message.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: '确定删除该项吗？',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        const api = activeTab.value === 'docs' ? docsApi : standardsApi
        await api.delete(row.id)
        Message.success('已删除')
        loadData()
      } catch { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.list-tabs { margin-bottom: 8px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }

.download-link { color: #165DFF; text-decoration: none; }
.download-link:hover { text-decoration: underline; }

.file-upload { display: inline-block; margin-right: 8px; }
.file-info { display: inline-flex; align-items: center; gap: 8px; font-size: 12px; color: #666; }
</style>
