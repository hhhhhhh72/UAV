<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="jobs"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增职位"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #salary="{ record }">
        <span>{{ record.salary_fen ? '¥' + (record.salary_fen / 100).toLocaleString('zh-CN') : '-' }}</span>
      </template>
      <template #jobType="{ record }">
        <a-tag :color="jobTypeTag(record.job_type)" size="small">{{ record.job_type || '-' }}</a-tag>
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
        <a-empty description="暂无职位数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="职位详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="职位名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="公司ID">{{ currentItem.enterprise_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="薪资">{{ currentItem.salary_fen ? '¥' + (currentItem.salary_fen / 100).toLocaleString('zh-CN') : '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="jobTypeTag(currentItem.job_type)" size="small">{{ currentItem.job_type || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发布时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑职位' : '新增职位'" :width="560" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical">
        <a-form-item label="职位名称" required><a-input v-model="form.title" style="width: 100%" /></a-form-item>
        <a-form-item label="类型">
          <a-select v-model="form.job_type" style="width: 100%">
            <a-option value="全职">全职</a-option>
            <a-option value="兼职">兼职</a-option>
            <a-option value="实习">实习</a-option>
            <a-option value="外包">外包</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="地区"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="published">招聘中</a-option>
            <a-option value="closed">已关闭</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="薪资(元)">
          <a-input-number v-model="form.salary" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
        </a-form-item>
        <a-form-item label="职位描述"><a-input v-model="form.description" type="textarea" :rows="2" style="width: 100%" /></a-form-item>
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
const api = useAdminApi('jobs')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const jobTypeTag = (t) => ({ '全职': 'arcoblue', '兼职': 'orangered', '实习': 'gray', 'contract': 'gray' }[t] || 'gray')

const statusTag = (s) => ({ published: 'green', closed: 'gray' }[s] || 'gray')
const statusLabel = { published: '招聘中', closed: '已关闭' }

// 批量动作：批量发布 / 批量关闭——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'close', label: '批量关闭', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索职位标题...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'published', label: '招聘中' },
    { value: 'closed', label: '已关闭' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '职位名称', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '公司ID', dataIndex: 'enterprise_id', width: 160 },
  { title: '地区', dataIndex: 'location', width: 120 },
  { title: '薪资', dataIndex: 'salary_fen', slotName: 'salary', width: 130 },
  { title: '类型', dataIndex: 'job_type', slotName: 'jobType', width: 100 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', location: '', salary: 0, job_type: '全职', status: 'published', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', location: '', salary: 0, job_type: '全职', status: 'published', description: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id, title: row.title || '', location: row.location || '',
      salary: row.salary_fen ? Math.round(row.salary_fen / 100 * 100) / 100 : 0,
      job_type: row.job_type || '全职', status: row.status || 'published', description: row.description || ''
    })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入职位名称'); return }
  formLoading.value = true
  try {
    const p = {
      id: form.id, title: form.title, location: form.location, job_type: form.job_type,
      status: form.status, description: form.description,
      salary_fen: Math.round((Number(form.salary) || 0) * 100)
    }
    if (formEdit.value) {
      await api.update(form.id, p)
      Message.success('更新成功')
    } else {
      await api.create(p)
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
    title: '删除职位',
    content: '确定删除该职位吗？删除后不可恢复',
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
.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}
</style>
