<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="competitions"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新建赛事"
      @add="showCreate()"
    >
      <template #startDate="{ record }">
        <span class="time-text">{{ formatDate(record.start_date) }}</span>
      </template>
      <template #regCount="{ record }">
        <span>{{ record.reg_count || 0 }} / {{ record.max_teams ? record.max_teams : '不限' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTagType(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无赛事" />
      </template>
    </CrudList>

    <!-- 详情弹窗（含修改状态区） -->
    <a-modal v-model:visible="detailVisible" title="赛事详情" :width="'min(600px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="赛事名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类别">{{ currentItem.category || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="报名/名额">{{ currentItem.reg_count || 0 }} / {{ currentItem.max_teams ? currentItem.max_teams : '不限' }}</a-descriptions-item>
          <a-descriptions-item label="主办方">{{ currentItem.sponsor || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item v-if="currentItem.poster" label="海报" :span="2">
            <a-image :src="currentItem.poster" :width="160" alt="赛事海报" :preview-props="{ src: currentItem.poster }" />
          </a-descriptions-item>
          <a-descriptions-item v-if="currentItem.description" label="简介" :span="2">{{ currentItem.description }}</a-descriptions-item>
        </a-descriptions>

        <div class="review-actions">
          <a-divider />
          <span style="margin-right: 12px;">修改状态：</span>
          <a-select v-model="newStatus" style="width: 140px;">
            <a-option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</a-option>
          </a-select>
          <a-button type="primary" @click="onUpdateStatus">更新</a-button>
        </div>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑赛事' : '新建赛事'" :width="'min(600px, 94vw)'" :mask-closable="false" :on-before-cancel="guardClose">
      <a-form :model="form" layout="vertical">
        <a-form-item label="赛事名称" required><a-input v-model="form.title" style="width: 100%" :aria-required="true" /></a-form-item>
        <a-form-item label="海报图">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="uploadRequest" accept="image/*" :before-upload="beforeUpload">
            <a-avatar v-if="form.poster" :image-url="form.poster" :size="80" shape="square" />
            <a-button v-else type="outline">点击上传</a-button>
          </a-upload>
        </a-form-item>
        <a-form-item label="类别"><a-input v-model="form.category" style="width: 100%" /></a-form-item>
        <a-form-item label="地点"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="主办方"><a-input v-model="form.sponsor" style="width: 100%" /></a-form-item>
        <a-form-item label="开始日期">
          <a-date-picker v-model="form.start_date" value-format="YYYY-MM-DD" placeholder="选择开始日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="结束日期">
          <a-date-picker v-model="form.end_date" value-format="YYYY-MM-DD" placeholder="选择结束日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="报名截止">
          <a-date-picker v-model="form.deadline" value-format="YYYY-MM-DD" placeholder="选择报名截止日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="报名费(元)">
          <a-input-number v-model="form.feeYuan" :min="0" hide-button style="width: 100%" placeholder="单位：元，0 表示免费" />
        </a-form-item>
        <a-form-item label="队伍名额">
          <a-input-number v-model="form.max_teams" :min="0" hide-button style="width: 100%" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="赛事简介"><a-input v-model="form.description" type="textarea" :rows="3" style="width: 100%" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="handleCancel">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">保存</a-button>
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
import { useAdminApi } from '@/api/admin/common'
import axios, { getAuthHeader } from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('competitions')
const uploadUrl = '/api/v1/upload'

const beforeUpload = (item) => {
  const file = item?.file || item
  const isImage = !!file.type && file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

// 海报上传：动态读取最新 accessToken
const uploadRequest = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  try {
    const res = await axios.post(uploadUrl, fd, { headers: getAuthHeader() })
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    form.poster = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error(e?.response?.data?.error?.message || e?.response?.data?.message || '上传失败')
  }
}

const statusOptions = [
  { label: '审核中', value: 'pending' },
  { label: '草稿', value: 'draft' },
  { label: '报名中', value: 'enrolling' },
  { label: '已下架', value: 'closed' }
]
const statusLabel = (s) => statusOptions.find(o => o.value === s)?.label || s || '-'
const statusTagType = (s) => ({ enrolling: 'green', pending: 'orangered', closed: 'gray', draft: 'gray' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 批量动作：开始报名 / 下架——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'open', label: '开始报名', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'enrolling' }) },
  { key: 'close', label: '下架', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部状态' },
    ...statusOptions
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 180 },
  { title: '赛事名称', dataIndex: 'title', minWidth: 160 },
  { title: '类别', dataIndex: 'category', width: 100 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '开始时间', dataIndex: 'start_date', slotName: 'startDate', width: 170 },
  { title: '报名/名额', dataIndex: 'reg_count', slotName: 'regCount', width: 110 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const newStatus = ref('draft')

const showDetail = (item) => {
  currentItem.value = { ...item }
  newStatus.value = item.status || 'draft'
  detailVisible.value = true
}

const showCreate = () => {
  openForm(null)
}

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', category: '', poster: '', location: '', sponsor: '', start_date: '', end_date: '', deadline: '', feeYuan: null, max_teams: null, status: 'draft', description: '' })

const resetForm = () => Object.assign(form, { id: '', title: '', category: '', poster: '', location: '', sponsor: '', start_date: '', end_date: '', deadline: '', feeYuan: null, max_teams: null, status: 'draft', description: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id, title: row.title || '', category: row.category || '', poster: row.poster || '',
      location: row.location || '', sponsor: row.sponsor || '',
      start_date: row.start_date ? String(row.start_date).slice(0, 10) : '',
      end_date: row.end_date ? String(row.end_date).slice(0, 10) : '',
      deadline: row.deadline ? String(row.deadline).slice(0, 10) : '',
      feeYuan: row.fee ?? null,
      max_teams: row.max_teams ?? null,
      status: row.status || 'draft', description: row.description || ''
    })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入赛事名称'); return }
  // 日期先后校验：报名截止 ≤ 开始 ≤ 结束
  if (form.start_date && form.end_date && form.start_date > form.end_date) { Message.warning('开始日期不能晚于结束日期'); return }
  if (form.deadline && form.start_date && form.deadline > form.start_date) { Message.warning('报名截止日期不能晚于开始日期'); return }
  formLoading.value = true
  try {
    const p = {
      title: form.title, category: form.category, poster: form.poster, location: form.location,
      sponsor: form.sponsor, start_date: form.start_date, end_date: form.end_date, deadline: form.deadline,
      fee: form.feeYuan ?? 0, max_teams: form.max_teams ?? 0,
      status: form.status, description: form.description
    }
    if (formEdit.value) {
      await api.update(form.id, p)
      Message.success('更新成功')
    } else {
      await api.create(p)
      Message.success('创建成功')
    }
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}

const onUpdateStatus = async () => {
  if (!currentItem.value) return
  try {
    // 传完整行：后端 update 是全字段覆盖，只传 status 会清空标题/类别/地点等
    await api.update(currentItem.value.id, { ...currentItem.value, status: newStatus.value })
    currentItem.value.status = newStatus.value
    Message.success('状态已更新')
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '更新失败') }
}

// 未保存守卫：X/Esc/遮罩/取消 关闭前，表单有改动则确认（onBeforeCancel 返回 false 阻断关闭）
let formSnapshot = ''
const guardClose = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
  return false
}
// 底部取消按钮：走守卫，确认无改动/放弃修改后才真正关闭
const handleCancel = () => { if (guardClose()) formVisible.value = false }

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除赛事',
    content: `确定删除「${row.title}」吗？`,
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
.time-text { color: var(--color-text-2); font-size: 12px; }

.review-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 16px;
  gap: 8px;
}
</style>
