<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="cases"
      :api-function="fetchCases"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      :batch-delete="false"
      creatable
      add-label="新增案例"
      @add="createCase()"
      @sorter-change="handleSorterChange"
    >
      <template #cover="{ record }">
        <div class="case-thumb">
          <img v-if="(record.images || [])[0]" :src="normalizeMediaUrl(record.images[0])" :alt="record.title" />
        </div>
      </template>
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '未命名案例' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="caseStatusColor[record.status]" size="small">{{ caseStatusLabel[record.status] || record.status }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="editCase(record)">编辑</a-button>
          <a-divider direction="vertical" />
          <a-button type="text" status="danger" size="small" @click="onDeleteCase(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无案例数据" />
      </template>
    </CrudList>

    <!-- 案例编辑弹窗（对齐 CaseEntry v1 模型） -->
    <a-modal
      v-model:visible="showCaseEditPopup"
      :title="currentCase?.id ? '编辑案例' : '新增案例'"
      :width="720"
      :footer="false"
      :on-before-cancel="beforeClose"
    >
      <template v-if="currentCase">
        <a-form :model="currentCase" layout="vertical" class="dialog-form">
          <a-divider orientation="left">基本信息</a-divider>
          <a-form-item label="分类" required>
            <a-input ref="categoryRef" v-model="currentCase.category" placeholder="如：物流配送 / 测绘巡检 / 应急救援" allow-clear style="width: 100%" />
          </a-form-item>
          <a-form-item label="标题" required>
            <a-input v-model="currentCase.title" placeholder="请输入标题" allow-clear style="width: 100%" />
          </a-form-item>
          <a-form-item label="简介">
            <a-input v-model="currentCase.description" type="textarea" :auto-size="{ minRows: 2, maxRows: 6 }" placeholder="请输入简介" style="width: 100%" />
          </a-form-item>
          <a-form-item label="客户名称">
            <a-input v-model="currentCase.clientName" placeholder="如：重庆市某区应急管理局" allow-clear style="width: 100%" />
          </a-form-item>
          <a-form-item label="成果">
            <a-input v-model="currentCase.result" type="textarea" :auto-size="{ minRows: 2, maxRows: 4 }" placeholder="项目成果/数据（可选）" style="width: 100%" />
          </a-form-item>

          <a-divider orientation="left">封面图片</a-divider>
          <a-form-item label="封面">
            <a-upload
              class="cover-upload"
              :show-file-list="false"
              :custom-request="uploadCover"
              :before-upload="beforeUpload"
              accept="image/*"
            >
              <img v-if="(currentCase.images || [])[0]" :src="normalizeMediaUrl(currentCase.images[0])" class="cover-preview" />
              <a-button v-else type="primary">点击上传</a-button>
            </a-upload>
            <a-button v-if="(currentCase.images || [])[0]" size="small" style="margin-top: 8px" @click="currentCase.images = []">清除</a-button>
          </a-form-item>

          <a-divider orientation="left">审核状态</a-divider>
          <a-form-item label="状态">
            <a-select v-model="currentCase.status" style="width: 100%">
              <a-option label="待审核" value="pending" />
              <a-option label="已发布" value="published" />
              <a-option label="已下架" value="archived" />
            </a-select>
          </a-form-item>
        </a-form>

        <div class="modal-footer">
          <a-space>
            <a-button @click="cancelForm">取消</a-button>
            <a-button v-if="currentCase?.id" status="danger" @click="onDeleteCase(currentCase)">删除案例</a-button>
            <a-button type="primary" @click="onSaveCase">保存</a-button>
          </a-space>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import axios from '@/utils/http'
import CrudList from '../components/CrudList.vue'
import { normalizeMediaUrl } from '../composables/useMedia'

const crudRef = ref()

// --- 案例列表（CaseEntry v1 模型：/api/v1/admin/cases） ---
const fetchCases = async (params) => {
  try {
    const res = await axios.get('/api/v1/admin/cases', { params })
    return {
      data: Array.isArray(res.data?.data) ? res.data.data : [],
      total: res.data?.total || 0
    }
  } catch (error) {
    Message.error('获取案例数据失败')
    return { data: [], total: 0 }
  }
}

// 批量动作（PUT 全字段覆盖 / DELETE）
const batchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => axios.put(`/api/v1/admin/cases/${row.id}`, { ...row, status: 'published' }) },
  { key: 'archive', label: '批量下架', status: 'warning', api: (row) => axios.put(`/api/v1/admin/cases/${row.id}`, { ...row, status: 'archived' }) },
  { key: 'delete', label: '批量删除', status: 'danger', api: (row) => axios.delete(`/api/v1/admin/cases/${row.id}`) }
]

const searchFields = computed(() => [
  { key: 'category', label: '分类', type: 'input', width: 200, placeholder: '输入分类名称（精确匹配）' }
])

const columns = [
  { title: '封面', dataIndex: 'images', slotName: 'cover', width: 90 },
  { title: '标题', dataIndex: 'title', slotName: 'title', minWidth: 180, sortable: true },
  { title: '分类', dataIndex: 'category', width: 110 },
  { title: '客户', dataIndex: 'clientName', width: 140 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

// a-table 排序（Arco direction → useListRequest 的 order 语义）
const handleSorterChange = (dataIndex, direction) => {
  const order = direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  crudRef.value?.onSortChange({ prop: dataIndex, order })
}

// --- 案例编辑 ---
const showCaseEditPopup = ref(false)
const categoryRef = ref()
const currentCase = ref(null)
const caseStatusColor = { pending: 'orange', published: 'green', archived: 'gray' }
const caseStatusLabel = { pending: '待审核', published: '已发布', archived: '已下架' }

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

// 封面图片上传（/api/v1/upload 返回 { url }）
// 注意：Arco custom-request 的参数是 fileItem，原生 File 在 fileItem.file 上
const uploadCover = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  try {
    const res = await axios.post('/api/v1/upload', fd)
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    if (currentCase.value) currentCase.value.images = [url]
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error('上传失败')
  }
}

const createCase = () => {
  currentCase.value = {
    title: '', category: '', description: '', images: [], clientName: '', result: '', status: 'pending'
  }
  caseSnapshot = JSON.stringify(currentCase.value)
  showCaseEditPopup.value = true
}

const editCase = (caseItem) => {
  currentCase.value = JSON.parse(JSON.stringify(caseItem))
  if (!Array.isArray(currentCase.value.images)) currentCase.value.images = []
  caseSnapshot = JSON.stringify(currentCase.value)
  showCaseEditPopup.value = true
}

const onSaveCase = async () => {
  if (!currentCase.value) return
  if (!currentCase.value.title?.trim()) {
    Message.error('标题不能为空')
    return
  }
  if (!String(currentCase.value.category || '').trim()) {
    Message.error('分类不能为空')
    categoryRef.value && categoryRef.value.focus && categoryRef.value.focus()
    return
  }
  Message.loading('保存中...', 0)
  try {
    if (currentCase.value.id) {
      await axios.put(`/api/v1/admin/cases/${currentCase.value.id}`, currentCase.value)
    } else {
      await axios.post('/api/v1/admin/cases', currentCase.value)
    }
    Message.clear()
    Message.success('保存成功')
    caseSnapshot = JSON.stringify(currentCase.value)
    showCaseEditPopup.value = false
    crudRef.value?.reload()
  } catch (error) {
    Message.clear()
    Message.error(error?.response?.data?.message || '保存失败')
  }
}

// 未保存守卫：Esc/点 X/点遮罩/底部取消 关闭前比对快照，有改动先确认
// 注意：Arco 2.58 无 beforeClose prop（beforeClose 只是 emits 事件），
// 需用 on-before-cancel 拦截用户关闭（X/ESC/遮罩）；底部取消按钮走 cancelForm。
let caseSnapshot = ''
const confirmDiscard = () => {
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { showCaseEditPopup.value = false },
  })
}
const cancelForm = () => {
  if (JSON.stringify(currentCase.value) === caseSnapshot) { showCaseEditPopup.value = false; return }
  confirmDiscard()
}
const beforeClose = () => {
  if (JSON.stringify(currentCase.value) === caseSnapshot) return true
  confirmDiscard()
  return false
}

const onDeleteCase = (caseItem) => {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这个案例吗？删除后无法恢复。',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await axios.delete(`/api/v1/admin/cases/${caseItem.id}`)
        Message.success('删除成功')
        if (currentCase.value?.id === caseItem.id) showCaseEditPopup.value = false
        crudRef.value?.reload()
      } catch (error) {
        Message.error('删除失败')
      }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.case-thumb { width: 56px; height: 56px; border-radius: 6px; overflow: hidden; background: #F7F8FA; }
.case-thumb img { width: 100%; height: 100%; object-fit: cover; display: block; }
.cell-title { font-weight: 500; }

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #EEF1F4;
}

.cover-upload { display: inline-block; margin-right: 8px; }
.cover-preview { width: 160px; height: 100px; object-fit: cover; border: 1px solid #ddd; border-radius: 4px; cursor: pointer; }
</style>
