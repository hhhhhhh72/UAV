<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="events"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增活动"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #type="{ record }">
        <a-tag :color="typeTag(record.event_type)" size="small">{{ record.event_type || '-' }}</a-tag>
      </template>
      <template #startTime="{ record }">
        <span class="time-text">{{ formatDate(record.start_time) }}</span>
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
        <a-empty description="暂无活动数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="活动详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="活动名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">{{ currentItem.event_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始时间">{{ formatDate(currentItem.start_time) }}</a-descriptions-item>
          <a-descriptions-item label="结束时间">{{ formatDate(currentItem.end_time) }}</a-descriptions-item>
          <a-descriptions-item label="名额">{{ currentItem.max_attendees || '-' }}</a-descriptions-item>
          <a-descriptions-item label="已报名">{{ currentItem.reg_count || 0 }} 人</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑活动' : '新增活动'" :width="560" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical">
        <a-form-item label="活动名称" required><a-input v-model="form.title" style="width: 100%" /></a-form-item>
        <a-form-item label="类型">
          <a-select v-model="form.event_type" style="width: 100%">
            <a-option value="论坛">论坛</a-option>
            <a-option value="走访">走访</a-option>
            <a-option value="沙龙">沙龙</a-option>
            <a-option value="培训">培训</a-option>
            <a-option value="其他">其他</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="地点"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="封面图URL"><a-input v-model="form.cover_url" placeholder="活动封面图片地址" style="width: 100%" /></a-form-item>
        <a-form-item label="开始时间"><a-input v-model="form.start_time" placeholder="YYYY-MM-DD HH:mm" style="width: 100%" /></a-form-item>
        <a-form-item label="结束时间"><a-input v-model="form.end_time" placeholder="YYYY-MM-DD HH:mm" style="width: 100%" /></a-form-item>
        <a-form-item label="名额"><a-input-number v-model="form.max_attendees" :min="0" hide-button style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="published">已发布</a-option>
            <a-option value="ongoing">进行中</a-option>
            <a-option value="ended">已结束</a-option>
            <a-option value="cancelled">已取消</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="3" style="width: 100%" /></a-form-item>
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
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('events')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const typeTag = (t) => ({ '论坛': 'green', '走访': 'arcoblue', '沙龙': 'orange', '培训': 'gray', '其他': 'arcoblue' }[t] || 'gray')
const statusTag = (s) => ({ published: 'orange', ongoing: 'green', ended: 'gray', cancelled: 'red' }[s] || 'gray')
const statusLabel = { published: '已发布', ongoing: '进行中', ended: '已结束', cancelled: '已取消' }

// 批量动作：批量发布 / 批量结束——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'end', label: '批量结束', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'ended' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索活动名称...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部状态' },
    { value: 'published', label: '已发布' },
    { value: 'ongoing', label: '进行中' },
    { value: 'ended', label: '已结束' },
    { value: 'cancelled', label: '已取消' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '活动名称', dataIndex: 'title', slotName: 'title', minWidth: 180 },
  { title: '类型', dataIndex: 'event_type', slotName: 'type', width: 100 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '开始时间', dataIndex: 'start_time', slotName: 'startTime', width: 160 },
  { title: '名额', dataIndex: 'max_attendees', width: 70, align: 'center' },
  { title: '已报名', dataIndex: 'reg_count', width: 80, align: 'center' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', event_type: '论坛', location: '', cover_url: '', start_time: '', end_time: '', max_attendees: 0, status: 'published', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', event_type: '论坛', location: '', cover_url: '', start_time: '', end_time: '', max_attendees: 0, status: 'published', description: '' })
const openForm = (r) => {
  resetForm()
  if (r) {
    formEdit.value = true
    Object.assign(form, r)
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入活动名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (r) => {
  Modal.confirm({
    title: '提示',
    content: '确定删除该活动?',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(r.id); Message.success('已删除'); crudRef.value?.reload() } catch { Message.error('删除失败') }
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
