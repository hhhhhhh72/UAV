<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="certificates"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增证书"
      @add="openForm()"
    >
      <template #certNo="{ record }">
        <span class="cell-mono">{{ record.cert_number || '-' }}</span>
      </template>
      <template #issueDate="{ record }">
        <span class="time-text">{{ formatDate(record.issue_date) }}</span>
      </template>
      <template #expireDate="{ record }">
        <span class="time-text">{{ formatDate(record.expire_date) }}</span>
      </template>
      <template #image="{ record }">
        <a-image
          v-if="record.image_url"
          :src="fullUrl(record.image_url)"
          :preview="true"
          width="44"
          height="44"
          fit="cover"
          style="border-radius: 4px; cursor: pointer;"
        />
        <span v-else>-</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无证书数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="证书详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="证书编号">{{ currentItem.cert_number || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证书类型">{{ currentItem.cert_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="持有人ID">{{ currentItem.user_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="签发日期">{{ formatDate(currentItem.issue_date) }}</a-descriptions-item>
          <a-descriptions-item label="有效期至">{{ formatDate(currentItem.expire_date) }}</a-descriptions-item>
          <a-descriptions-item label="等级">{{ currentItem.level || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发证机构" :span="2">{{ currentItem.issuer_org || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑证书' : '新增证书'" :width="560" @cancel="formVisible = false">
      <a-form :model="form" :label-col-flex="90">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="证书编号" required><a-input v-model="form.cert_number" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="证书类型"><a-input v-model="form.cert_type" placeholder="caac / utc_dji / gov_level" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="等级"><a-input v-model="form.level" placeholder="如：CAAC Ⅲ类" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="发证机构"><a-input v-model="form.issuer_org" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="签发日期">
              <a-date-picker v-model="form.issue_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="有效期至">
              <a-date-picker v-model="form.expire_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="pending">待审核</a-option>
            <a-option value="approved">已通过</a-option>
            <a-option value="rejected">已驳回</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('certificates')

// 相对路径图片补全后端地址（vite/nginx 已代理 /uploads）
const fullUrl = (u) => (u && u.startsWith('http') ? u : (import.meta.env.VITE_API_TARGET || 'http://localhost:8080') + (u || ''))

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ approved: 'green', pending: 'orangered', rejected: 'red' }[s] || 'gray')
const statusLabel = { approved: '已通过', pending: '待审核', rejected: '已驳回' }

// 批量动作：批量通过 / 批量驳回——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'approved' }) },
  { key: 'reject', label: '批量驳回', status: 'danger', api: (row) => api.update(row.id, { ...row, status: 'rejected' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索持有人或证书编号...', width: 260 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'pending', label: '待审核' },
    { value: 'approved', label: '已通过' },
    { value: 'rejected', label: '已驳回' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '证书编号', dataIndex: 'cert_number', slotName: 'certNo', width: 180 },
  { title: '持有人ID', dataIndex: 'user_id', minWidth: 140 },
  { title: '证书类型', dataIndex: 'cert_type', width: 120 },
  { title: '签发日期', dataIndex: 'issue_date', slotName: 'issueDate', width: 120 },
  { title: '有效期至', dataIndex: 'expire_date', slotName: 'expireDate', width: 120 },
  { title: '等级', dataIndex: 'level', width: 80 },
  { title: '发证机构', dataIndex: 'issuer_org', minWidth: 140 },
  { title: '证书图片', dataIndex: 'image_url', slotName: 'image', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', cert_number: '', cert_type: '', level: '', issuer_org: '', issue_date: '', expire_date: '', status: 'pending' })
const resetForm = () => Object.assign(form, { id: '', cert_number: '', cert_type: '', level: '', issuer_org: '', issue_date: '', expire_date: '', status: 'pending' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, { ...row })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.cert_number) { Message.warning('请输入证书编号'); return }
  formLoading.value = true
  try {
    if (formEdit.value) {
      await api.update(form.id, { ...form })
      Message.success('更新成功')
    } else {
      await api.create({ ...form })
      Message.success('创建成功')
    }
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除证书',
    content: '确定删除该证书吗？删除后不可恢复',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}
</script>

<style scoped>
.cell-mono { font-family: 'Courier New', monospace; font-size: 13px; }
.time-text { color: #86909C; font-size: 12px; }
</style>
