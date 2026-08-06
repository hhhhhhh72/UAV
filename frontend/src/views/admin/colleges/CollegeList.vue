<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="colleges"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增院校"
      @add="openForm()"
    >
      <template #name="{ record }">
        <span class="cell-title">{{ record.name || '-' }}</span>
      </template>
      <template #coopType="{ record }">
        <span>{{ coopTypeLabel[record.coop_type] || coopTypeLabel.both }}</span>
      </template>
      <template #majors="{ record }">
        <span>{{ arrText(record.majors) }}</span>
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
        <a-empty description="暂无院校数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="院校详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="院校名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="分域">{{ coopTypeLabel[currentItem.coop_type] || coopTypeLabel.both }}</a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.region || '-' }}</a-descriptions-item>
          <a-descriptions-item label="合作状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="特色专业" :span="2">{{ arrText(currentItem.majors) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="实训设施" :span="2">{{ arrText(currentItem.facilities) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="简介" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑院校' : '新增院校'" :width="560" @cancel="formVisible = false">
      <a-form :model="form" :label-col-flex="90">
        <a-row :gutter="16">
          <a-col :span="14">
            <a-form-item label="院校名称" required><a-input v-model="form.name" /></a-form-item>
          </a-col>
          <a-col :span="10">
            <a-form-item label="分域">
              <a-select v-model="form.coop_type" style="width: 100%">
                <a-option value="research">科研合作</a-option>
                <a-option value="talent">人才培养</a-option>
                <a-option value="both">综合</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="地区"><a-input v-model="form.region" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="合作状态">
              <a-select v-model="form.status" style="width: 100%">
                <a-option value="active">合作中</a-option>
                <a-option value="pending">洽谈中</a-option>
                <a-option value="closed">已结束</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Logo 地址"><a-input v-model="form.logo_url" placeholder="院校 Logo 图片地址" /></a-form-item>
        <a-form-item label="特色专业"><a-input v-model="form.majorsText" placeholder="多个专业用逗号分隔，如：无人机应用技术,测绘地理信息" /></a-form-item>
        <a-form-item label="实训设施"><a-input v-model="form.facilitiesText" placeholder="多个设施用逗号分隔，如：实训基地,联合实验室" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="2" /></a-form-item>
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
const api = useAdminApi('colleges')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ active: 'green', pending: 'orangered', closed: 'gray' }[s] || 'gray')
const statusLabel = { active: '合作中', pending: '待合作', closed: '已终止' }

// 批量动作：批量合作中 / 批量终止合作——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'activate', label: '批量合作中', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'active' }) },
  { key: 'close', label: '批量终止合作', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索院校名称...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'active', label: '合作中' },
    { value: 'pending', label: '待合作' },
    { value: 'closed', label: '已终止' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '院校名称', dataIndex: 'name', slotName: 'name', minWidth: 200 },
  { title: '分域', dataIndex: 'coop_type', slotName: 'coopType', width: 100 },
  { title: '地区', dataIndex: 'region', width: 120 },
  { title: '特色专业', dataIndex: 'majors', slotName: 'majors', minWidth: 160 },
  { title: '合作状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const coopTypeLabel = { research: '科研合作', talent: '人才培养', both: '综合' }
const arrText = (v) => (Array.isArray(v) ? v.join('、') : (v || ''))
const splitArr = (s) => String(s || '').split(/[,，、]/).map(x => x.trim()).filter(Boolean)
const form = reactive({ id: '', name: '', coop_type: 'both', region: '', logo_url: '', majorsText: '', facilitiesText: '', status: 'pending', description: '' })
const resetForm = () => Object.assign(form, { id: '', name: '', coop_type: 'both', region: '', logo_url: '', majorsText: '', facilitiesText: '', status: 'pending', description: '' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id, name: row.name || '', coop_type: row.coop_type || 'both', region: row.region || '',
      logo_url: row.logo_url || '', majorsText: arrText(row.majors), facilitiesText: arrText(row.facilities),
      status: row.status || 'pending', description: row.description || ''
    })
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入院校名称'); return }
  formLoading.value = true
  try {
    const p = {
      id: form.id, name: form.name, coop_type: form.coop_type, region: form.region,
      logo_url: form.logo_url, status: form.status, description: form.description,
      majors: splitArr(form.majorsText), facilities: splitArr(form.facilitiesText)
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
    title: '删除院校',
    content: '确定删除该院校吗？删除后不可恢复',
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
