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
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="通知详情" :width="'min(640px, 94vw)'" :footer="false" :unmount-on-close="true">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID" :span="2">{{ currentItem.id }}</a-descriptions-item>
          <a-descriptions-item label="标题" :span="2">{{ currentItem.title }}</a-descriptions-item>
          <a-descriptions-item label="发送者">{{ currentItem.sender_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="接收者">{{ currentItem.receiver_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="资源类型">{{ currentItem.resource_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="资源ID">{{ currentItem.resource_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发送时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="currentItem.is_read ? 'gray' : 'orange'" size="small">{{ currentItem.is_read ? '已读' : '未读' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="内容" :span="2">
            <div style="white-space: pre-wrap; line-height: 1.6;">{{ currentItem.content || '-' }}</div>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（发送/广播） -->
    <a-modal v-model:visible="formVisible" title="发送通知" :width="'min(560px, 94vw)'" :mask-closable="false" :unmount-on-close="true" :on-before-cancel="beforeCancel">
      <a-form :model="form" layout="vertical">
        <a-form-item label="消息标题" required><a-input v-model="form.title" :aria-required="true" style="width: 100%" /></a-form-item>
        <a-form-item label="接收者"><a-input v-model="form.receiver_id" placeholder="留空 = 广播给全部用户" style="width: 100%" /></a-form-item>
        <a-form-item label="消息内容" required><RichEditor v-model="form.content" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="beforeCancel">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">发送</a-button>
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
import CrudList from '../components/CrudList.vue'
import RichEditor from '@/components/RichEditor.vue'

const crudRef = ref()
const api = useAdminApi('messages')
const defaultParams = {}

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return '-'
  const p = n => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate()) + ' ' + p(dt.getHours()) + ':' + p(dt.getMinutes())
}

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索标题/内容...', width: 220 }
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
const formLoading = ref(false)
const form = reactive({ id: '', title: '', receiver_id: '', content: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', receiver_id: '', content: '' })

// 仅支持发送/广播：后端没有消息更新接口（PUT /admin/messages/{id} 实为标记已读），
// 原"编辑"会走创建逻辑重复入库，故列表不提供编辑入口，避免产生重复消息。
const openForm = () => {
  resetForm()
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

// 未保存守卫：Esc/点 X/点取消关闭前，若表单有改动则确认，避免输入全丢
let formSnapshot = ''
const beforeCancel = () => {
  if (JSON.stringify(form) === formSnapshot) { formVisible.value = false; return true }
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
  return false
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '发送失败'

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入消息标题'); return }
  formLoading.value = true
  try {
    const p = { sender_id: 'system', receiver_id: form.receiver_id || '', title: form.title, content: form.content }
    await api.create(p)
    Message.success(form.receiver_id ? '发送成功' : '已广播给全部用户')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: `确定删除通知「${row.title || row.id}」吗？删除后不可恢复`,
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

.time-text { color: var(--color-text-2); font-size: 12px; }
</style>
