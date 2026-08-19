<template>
  <div class="page">
    <!-- Tab 切换（合规文档 / 团体标准） -->
    <a-card :bordered="false" class="tab-card">
      <a-tabs v-model:active-key="activeTab">
        <a-tab-pane title="合规文档" key="docs" />
        <a-tab-pane title="团体标准" key="standards" />
      </a-tabs>
    </a-card>

    <!-- 文档列表 -->
    <CrudList
      v-if="activeTab === 'docs'"
      ref="crudRef"
      resource="compliance-docs"
      :columns="docsColumns"
      :search-fields="docsSearchFields"
      :batch-actions="docsBatchActions"
      creatable
      add-label="新增文档"
      @add="handleAdd()"
      @sorter-change="handleSorterChange"
    >
      <template #category="{ record }">
        <a-tag :color="categoryColor[record.category] || 'gray'" size="small">{{ record.category }}</a-tag>
      </template>
      <template #publishDate="{ record }">
        <span class="time-text">{{ formatDate(record.publish_date) }}</span>
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
    </CrudList>

    <!-- 标准列表 -->
    <CrudList
      v-else
      ref="crudRef"
      resource="compliance-standards"
      :columns="stdColumns"
      :search-fields="stdSearchFields"
      :batch-actions="stdBatchActions"
      creatable
      add-label="新增标准"
      @add="handleAdd()"
      @sorter-change="handleSorterChange"
    >
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
    </CrudList>

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
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑' : '新增'" :width="560" :on-before-cancel="guardClose" @cancel="formVisible = false">
      <!-- 文档表单 -->
      <a-form v-if="activeTab === 'docs'" :model="docForm" layout="vertical" class="dialog-form">
        <a-form-item label="文档标题" required>
          <a-input v-model="docForm.title" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="分类">
          <a-select v-model="docForm.category" style="width: 100%">
            <a-option v-for="c in ['政策', '法规', '标准', '指南']" :key="c" :label="c" :value="c" />
          </a-select>
        </a-form-item>
        <a-form-item label="发布机构">
          <a-input v-model="docForm.publisher" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="发布日期">
          <a-date-picker v-model="docForm.publish_date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="docForm.status" style="width: 100%">
            <a-option label="待审核" value="pending" />
            <a-option label="已发布" value="published" />
            <a-option label="草稿" value="draft" />
          </a-select>
        </a-form-item>
        <a-form-item label="摘要">
          <a-input v-model="docForm.summary" type="textarea" :auto-size="{ minRows: 2, maxRows: 6 }" style="width: 100%" />
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
      <a-form v-else :model="stdForm" layout="vertical" class="dialog-form">
        <a-form-item label="分类">
          <a-select v-model="stdForm.category" style="width: 100%">
            <a-option v-for="c in ['国家标准', '行业标准', '团体标准', '企业标准']" :key="c" :label="c" :value="c" />
          </a-select>
        </a-form-item>
        <a-form-item label="标准编号">
          <a-input v-model="stdForm.standard_no" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="标准名称" required>
          <a-input v-model="stdForm.title" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="发布机构">
          <a-input v-model="stdForm.publisher" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="生效日期">
          <a-date-picker v-model="stdForm.effective_date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="stdForm.status" style="width: 100%">
            <a-option label="待审核" value="pending" />
            <a-option label="已发布" value="published" />
            <a-option label="草稿" value="draft" />
          </a-select>
        </a-form-item>
        <a-form-item label="适用范围">
          <a-input v-model="stdForm.scope" type="textarea" :auto-size="{ minRows: 2, maxRows: 6 }" style="width: 100%" />
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
        <a-button @click="handleCancel">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="handleFormSubmit">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import axios, { getAuthHeader } from '@/utils/http'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
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

const categoryColor = { '政策': 'orange', '法规': 'red', '标准': 'green', '指南': 'gray', '国家标准': 'purple', '行业标准': 'blue', '团体标准': 'green', '企业标准': 'gray' }
const statusColor = { 'published': 'green', 'draft': 'gray', 'archived': 'orange', 'pending': 'orange' }
const statusLabel = { 'published': '已发布', 'draft': '草稿', 'archived': '已下架', 'pending': '待审核' }

// 文档列表配置
const docsSearchFields = [
  // 后端 ListDocs 仅支持 category 过滤，keyword 无效已移除
  { key: 'category', label: '分类', type: 'select', width: 160, placeholder: '分类筛选', options: [
    { value: '', label: '全部' },
    { value: '政策', label: '政策' },
    { value: '法规', label: '法规' },
    { value: '标准', label: '标准' },
    { value: '指南', label: '指南' }
  ]}
]

const docsColumns = [
  { title: 'ID', dataIndex: 'id', width: 160, sortable: true },
  { title: '文档标题', dataIndex: 'title', minWidth: 200 },
  { title: '分类', dataIndex: 'category', slotName: 'category', width: 100 },
  { title: '发布机构', dataIndex: 'publisher', width: 160 },
  { title: '发布日期', dataIndex: 'publish_date', slotName: 'publishDate', width: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

// 批量动作：批量发布 / 批量下架——传完整行数据避免清空其他字段
const docsBatchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => docsApi.update(row.id, { ...row, status: 'published' }) },
  { key: 'archive', label: '批量下架', status: 'warning', api: (row) => docsApi.update(row.id, { ...row, status: 'archived' }) }
]

// 标准列表配置
const stdSearchFields = [
  // 后端 ListStandards 仅支持 category 过滤，keyword 无效已移除
  { key: 'category', label: '分类', type: 'select', width: 160, placeholder: '分类筛选', options: [
    { value: '', label: '全部' },
    { value: '国家标准', label: '国家标准' },
    { value: '行业标准', label: '行业标准' },
    { value: '团体标准', label: '团体标准' },
    { value: '企业标准', label: '企业标准' }
  ]}
]

const stdColumns = [
  { title: 'ID', dataIndex: 'id', width: 160, sortable: true },
  { title: '分类', dataIndex: 'category', slotName: 'category', width: 100 },
  { title: '标准编号', dataIndex: 'standard_no', width: 180 },
  { title: '标准名称', dataIndex: 'title', minWidth: 200 },
  { title: '发布机构', dataIndex: 'publisher', width: 160 },
  { title: '生效日期', dataIndex: 'effective_date', slotName: 'effectiveDate', width: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const stdBatchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => standardsApi.update(row.id, { ...row, status: 'published' }) },
  { key: 'archive', label: '批量下架', status: 'warning', api: (row) => standardsApi.update(row.id, { ...row, status: 'archived' }) }
]

// a-table 排序（Arco direction → useListRequest 的 order 语义）
const handleSorterChange = (dataIndex, direction) => {
  const order = direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  crudRef.value?.onSortChange({ prop: dataIndex, order })
}

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

// ---- Add/Edit form ----
const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const docForm = reactive({ id: '', title: '', category: '政策', publisher: '', publish_date: '', status: 'published', summary: '', file_url: '' })
const stdForm = reactive({ id: '', category: '团体标准', standard_no: '', title: '', publisher: '', effective_date: '', status: 'published', scope: '', file_url: '' })

// 未保存守卫：按当前 Tab 的表单做快照比对 + Modal.confirm；
// X/遮罩/Esc 走 onBeforeCancel，footer 取消按钮也走守卫
let formSnapshot = ''
const activeForm = () => (activeTab.value === 'docs' ? docForm : stdForm)
const takeSnapshot = () => { formSnapshot = JSON.stringify(activeForm()) }
const guardClose = () => {
  if (JSON.stringify(activeForm()) === formSnapshot) return true
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
  return false
}
const handleCancel = () => {
  if (guardClose()) formVisible.value = false
}

const beforeUpload = (file) => {
  const maxSize = 20 * 1024 * 1024 // 20MB for documents
  if (file.size > maxSize) { Message.error('文件不能超过 20MB'); return false }
  return true
}

// 附件上传（/api/v1/upload 返回相对 URL，按当前 Tab 写入对应表单）
const uploadFile = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  try {
    const res = await axios.post('/api/v1/upload', fd, { headers: getAuthHeader() })
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
const resetStdForm = () => Object.assign(stdForm, { id: '', category: '团体标准', standard_no: '', title: '', publisher: '', effective_date: '', status: 'published', scope: '', file_url: '' })
const handleAdd = () => {
  if (activeTab.value === 'docs') resetDocForm(); else resetStdForm()
  formEdit.value = false
  takeSnapshot()
  formVisible.value = true
}
const handleEdit = (row) => {
  if (activeTab.value === 'docs') Object.assign(docForm, row)
  else Object.assign(stdForm, row)
  formEdit.value = true
  takeSnapshot()
  formVisible.value = true
}
const handleFormSubmit = async () => {
  const api = activeTab.value === 'docs' ? docsApi : standardsApi
  const payload = activeTab.value === 'docs' ? { ...docForm } : { ...stdForm }
  if (!payload.title) { Message.warning('请输入标题'); return }
  formLoading.value = true
  try {
    if (formEdit.value) { await api.update(payload.id, payload); Message.success('更新成功') }
    else { await api.create(payload); Message.success('创建成功') }
    takeSnapshot()
    formVisible.value = false
    crudRef.value?.reload()
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
        crudRef.value?.reload()
      } catch { Message.error('删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.tab-card { margin-bottom: 16px; }

.time-text { color: #86909C; font-size: 12px; }

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }

.download-link { color: #165DFF; text-decoration: none; }
.download-link:hover { text-decoration: underline; }

.file-upload { display: inline-block; margin-right: 8px; }
.file-info { display: inline-flex; align-items: center; gap: 8px; font-size: 12px; color: #666; }
</style>
