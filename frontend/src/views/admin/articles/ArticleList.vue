<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="articles"
      :api-function="fetchArticles"
      :columns="columns"
      :search-fields="searchFields"
      :batch-delete="false"
      creatable
      add-label="发布资讯"
      @add="openCreate()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '未命名' }}</span>
      </template>
      <template #category="{ record }">
        <a-tag size="small" color="arcoblue">{{ categoryLabel(record.category) }}</a-tag>
      </template>
      <template #status="{ record }">
        <a-tag size="small" :color="record.status === 'published' ? 'green' : 'orange'">
          {{ record.status === 'published' ? '已发布' : '草稿' }}
        </a-tag>
      </template>
      <template #actions="{ record }">
        <a-button type="text" size="small" @click="openEdit(record)">编辑</a-button>
        <a-button
          v-if="record.status !== 'published'"
          type="text"
          size="small"
          @click="publishArticle(record)"
        >发布</a-button>
      </template>
      <template #empty>
        <a-empty description="暂无资讯，点击右上角「发布资讯」添加" />
      </template>
    </CrudList>

    <!-- 发布资讯弹窗 -->
    <a-modal
      v-model:visible="showPopup"
      :title="editingId ? '编辑资讯' : '发布资讯'"
      :width="640"
      :footer="false"
    >
      <a-form :model="form" layout="vertical" class="dialog-form">
        <a-form-item label="分类" required>
          <a-select v-model="form.category" placeholder="选择资讯分类" style="width: 100%">
            <a-option v-for="opt in CATEGORY_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="标题" required>
          <a-input v-model="form.title" placeholder="请输入资讯标题" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="来源">
          <a-input v-model="form.source" placeholder="如：重庆市无人机产业协会" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="正文" required>
          <a-textarea v-model="form.content" :auto-size="{ minRows: 6, maxRows: 14 }" placeholder="请输入资讯正文（列表摘要将自动截取前 100 字）" style="width: 100%" />
        </a-form-item>
      </a-form>

      <div class="modal-footer">
        <a-space>
          <a-button @click="showPopup = false">取消</a-button>
          <a-button type="primary" @click="onSave">保存草稿</a-button>
          <a-button type="primary" status="success" @click="onSaveAndPublish">保存并发布</a-button>
        </a-space>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import axios from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()

/* 分类与小程序政策资讯页保持一致（GET /api/v1/articles?category=） */
const CATEGORY_OPTIONS = [
  { value: 'low_altitude_policy', label: '低空经济' },
  { value: 'uav_regulation', label: '无人机法规' },
  { value: 'airspace_management', label: '空域管理' },
  { value: 'subsidy_policy', label: '补贴政策' },
  { value: 'industry_standard', label: '行业标准' },
  { value: 'drone_knowledge', label: '无人机知识' },
]
const CATEGORY_MAP = Object.fromEntries(CATEGORY_OPTIONS.map((o) => [o.value, o.label]))

const categoryLabel = (v) => CATEGORY_MAP[v] || v || '其他'

// --- 列表（公开读接口 /api/v1/articles） ---
const fetchArticles = async (params) => {
  try {
    const res = await axios.get('/api/v1/articles', { params })
    return {
      data: Array.isArray(res.data?.data) ? res.data.data : [],
      total: res.data?.total || 0
    }
  } catch (error) {
    Message.error('获取资讯数据失败')
    return { data: [], total: 0 }
  }
}

const searchFields = computed(() => [
  { key: 'category', label: '分类', type: 'select', width: 200, options: CATEGORY_OPTIONS, placeholder: '全部' }
])

const columns = [
  { title: '标题', dataIndex: 'title', slotName: 'title', minWidth: 240 },
  { title: '分类', dataIndex: 'category', slotName: 'category', width: 110 },
  { title: '来源', dataIndex: 'source', width: 180, ellipsis: true },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '发布时间', dataIndex: 'created_at', width: 170 },
  { title: '操作', slotName: 'actions', width: 90, fixed: 'right' },
]

// --- 发布 ---
const showPopup = ref(false)
const form = ref(createEmptyForm())

function createEmptyForm() {
  return { title: '', content: '', category: '', source: '' }
}

const editingId = ref('')

const openCreate = () => {
  editingId.value = ''
  form.value = createEmptyForm()
  showPopup.value = true
}

const openEdit = (item) => {
  editingId.value = item.id
  form.value = {
    title: item.title || '',
    content: item.content || '',
    category: item.category || '',
    source: item.source || ''
  }
  showPopup.value = true
}

const validate = () => {
  if (!form.value.title?.trim()) { Message.error('标题不能为空'); return false }
  if (!form.value.category) { Message.error('请选择分类'); return false }
  if (!form.value.content?.trim()) { Message.error('正文不能为空'); return false }
  return true
}

const buildPayload = () => ({
  title: form.value.title.trim(),
  content: form.value.content,
  category: form.value.category,
  source: form.value.source?.trim() || ''
})

const onSave = async () => {
  if (!validate()) return
  Message.loading('保存中...', 0)
  try {
    if (editingId.value) {
      await axios.put(`/api/v1/articles/${editingId.value}`, buildPayload())
      Message.success('已保存')
    } else {
      await axios.post('/api/v1/articles', buildPayload())
      Message.success('草稿已保存')
    }
    Message.clear()
    showPopup.value = false
    crudRef.value?.reload()
  } catch (error) {
    Message.clear()
    Message.error(error?.response?.data?.message || '保存失败')
  }
}

const onSaveAndPublish = async () => {
  if (!validate()) return
  Message.loading('发布中...', 0)
  try {
    if (editingId.value) {
      // 编辑已发布/草稿文章：更新内容，状态保持不变
      await axios.put(`/api/v1/articles/${editingId.value}`, buildPayload())
    } else {
      const res = await axios.post('/api/v1/articles', buildPayload())
      const id = res?.data?.data?.id || res?.data?.id
      if (id) await axios.post(`/api/v1/articles/${id}/publish`)
    }
    Message.clear()
    Message.success('已发布')
    showPopup.value = false
    crudRef.value?.reload()
  } catch (error) {
    Message.clear()
    Message.error(error?.response?.data?.message || '发布失败')
  }
}

// 草稿 → 发布
const publishArticle = (item) => {
  Modal.confirm({
    title: '确认发布',
    content: `确定要发布「${item.title}」吗？发布后小程序端即可见。`,
    okText: '发布',
    cancelText: '取消',
    onOk: async () => {
      try {
        await axios.post(`/api/v1/articles/${item.id}/publish`)
        Message.success('发布成功')
        crudRef.value?.reload()
      } catch (error) {
        Message.error('发布失败')
      }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }
.cell-title { font-weight: 500; }
.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }
.modal-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #EEF1F4;
}
</style>
