<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="messages"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      creatable
      add-label="发送通知"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #isRead="{ record }">
        <a-tag :color="record.is_read ? 'gray' : 'orange'" size="small">{{ record.is_read ? '已读' : '未读' }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="通知详情" :width="640" :footer="false" :unmount-on-close="true">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID" :span="2">{{ currentItem.id }}</a-descriptions-item>
          <a-descriptions-item label="标题" :span="2">{{ currentItem.title }}</a-descriptions-item>
          <a-descriptions-item label="消息类型">
            <a-tag :color="typeColor[currentItem.msg_type] || 'gray'" size="small">{{ currentItem.msg_type || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="接收者">{{ currentItem.to_user || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发送时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="阅读时间">{{ formatDate(currentItem.read_at) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusColor[currentItem.status] || 'gray'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="内容" :span="2">
            <div style="white-space: pre-wrap; line-height: 1.6;">{{ currentItem.content || '-' }}</div>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（发送/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑通知' : '发送通知'" :width="560" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="horizontal">
        <a-form-item label="消息标题" required><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="接收者"><a-input v-model="form.receiver_id" placeholder="留空 = 广播给所有管理员" /></a-form-item>
        <a-form-item label="消息内容" required><a-input v-model="form.content" type="textarea" :rows="5" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">发送</a-button>
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
const api = useAdminApi('messages')
const defaultParams = { msg_type: '' }

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return '-'
  const p = n => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate()) + ' ' + p(dt.getHours()) + ':' + p(dt.getMinutes())
}

const typeColor = { '系统通知': 'orange', '活动提醒': 'green', '审核结果': 'red', '其他': 'gray' }
const statusLabel = { 'unread': '未读', 'read': '已读' }
const statusColor = { 'unread': 'orange', 'read': 'gray' }

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索标题...', width: 220 },
  { key: 'msg_type', label: '类型', type: 'select', options: [
    { value: '', label: '全部' },
    { value: '系统通知', label: '系统通知' },
    { value: '活动提醒', label: '活动提醒' },
    { value: '审核结果', label: '审核结果' },
    { value: '其他', label: '其他' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '接收者', dataIndex: 'receiver_id', width: 140 },
  { title: '发送时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '状态', dataIndex: 'is_read', slotName: 'isRead', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', receiver_id: '', content: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', receiver_id: '', content: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, { title: row.title || '', receiver_id: row.receiver_id || '', content: row.content || '' })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '发送失败'

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入消息标题'); return }
  formLoading.value = true
  try {
    const p = { sender_id: 'system', receiver_id: form.receiver_id || '', title: form.title, content: form.content }
    await api.create(p)
    Message.success(form.receiver_id ? '发送成功' : '已广播给所有管理员')
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: '确定删除该通知吗？',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); crudRef.value?.reload() } catch { Message.error('删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.time-text { color: #86909C; font-size: 12px; }
</style>
