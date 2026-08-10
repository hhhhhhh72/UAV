<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="training-courses"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增课程"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #price="{ record }">
        <span>{{ record.price_fen ? '¥' + (record.price_fen / 100).toLocaleString() : '-' }}</span>
      </template>
      <template #maxStudents="{ record }">
        <span>{{ record.max_students ?? '-' }}</span>
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
        <a-empty description="暂无课程数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="课程详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="课程名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证书类型">{{ certTypeLabel(currentItem.cert_type) }}</a-descriptions-item>
          <a-descriptions-item label="价格">{{ currentItem.price_fen ? '¥' + (currentItem.price_fen / 100).toLocaleString() : '-' }}</a-descriptions-item>
          <a-descriptions-item label="名额">{{ currentItem.max_students ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="已报名">{{ currentItem.enrolled_count ?? 0 }} 人</a-descriptions-item>
          <a-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束日期">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑课程' : '新增课程'" :width="560" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical">
        <a-form-item label="课程名称" required><a-input v-model="form.title" style="width: 100%" /></a-form-item>
        <a-form-item label="证书类型">
          <a-select v-model="form.cert_type" style="width: 100%">
            <a-option value="caac">CAAC 执照</a-option>
            <a-option value="utc_dji">大疆 UTC</a-option>
            <a-option value="gov_level">人社等级</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="上课地点"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="published">已发布</a-option>
            <a-option value="closed">已关闭</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="价格(元)">
          <a-input-number v-model="form.priceYuan" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
        </a-form-item>
        <a-form-item label="名额">
          <a-input-number v-model="form.max_students" :min="0" hide-button style="width: 100%" placeholder="招生人数" />
        </a-form-item>
        <a-form-item label="开始日期">
          <a-date-picker v-model="form.start_date" value-format="YYYY-MM-DD" placeholder="选择开班日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="结束日期">
          <a-date-picker v-model="form.end_date" value-format="YYYY-MM-DD" placeholder="选择结课日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="课程描述"><a-input v-model="form.description" type="textarea" :rows="3" style="width: 100%" /></a-form-item>
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
const api = useAdminApi('training-courses')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ published: 'green', draft: 'orangered', closed: 'gray' }[s] || 'gray')
const statusLabel = { published: '已发布', draft: '草稿', closed: '已关闭' }
const certTypeLabel = (t) => ({ caac: 'CAAC 执照', utc_dji: '大疆 UTC', gov_level: '人社等级' }[t] || t || '-')

// 批量动作：批量发布 / 批量关闭——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量发布', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'close', label: '批量关闭', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索课程标题...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'draft', label: '草稿' },
    { value: 'published', label: '已发布' },
    { value: 'closed', label: '已关闭' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '课程名称', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '价格', dataIndex: 'price_fen', slotName: 'price', width: 110, align: 'right' },
  { title: '名额', dataIndex: 'max_students', slotName: 'maxStudents', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (d) => { currentItem.value = d; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', cert_type: 'caac', priceYuan: null, max_students: null, location: '', start_date: '', end_date: '', status: 'draft', description: '' })

const resetForm = () => Object.assign(form, { id: '', title: '', cert_type: 'caac', priceYuan: null, max_students: null, location: '', start_date: '', end_date: '', status: 'draft', description: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      ...row,
      priceYuan: row.price_fen ? Math.round(row.price_fen / 100 * 100) / 100 : null,
      max_students: row.max_students ?? null
    })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入课程名称'); return }
  formLoading.value = true
  try {
    const payload = { ...form }
    payload.price_fen = Math.round((form.priceYuan || 0) * 100)
    payload.max_students = payload.max_students ?? 0
    delete payload.priceYuan
    if (formEdit.value) {
      await api.update(form.id, payload)
      Message.success('更新成功')
    } else {
      await api.create(payload)
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
    title: '删除课程',
    content: `确定删除该课程吗？（${row.title || row.id}）`,
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
