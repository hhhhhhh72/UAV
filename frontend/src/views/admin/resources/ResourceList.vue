<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="industry-resources"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增资源"
      @add="openForm()"
    >
      <template #resType="{ record }">
        <a-tag :color="resTypeColor[record.res_type] || 'gray'" size="small">{{ resTypeLabel[record.res_type] || record.res_type }}</a-tag>
      </template>
      <template #price="{ record }">
        <span class="cell-amount">¥{{ formatPriceFen(record.price_fen) }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusColor[record.status] || 'gray'" size="small">{{ statusLabel[record.status] || record.status }}</a-tag>
      </template>
      <template #vis="{ record }">
        <a-tag :color="visColor[record.visibility_level] || 'gray'" size="small">{{ visLabel[record.visibility_level] || '公开' }}</a-tag>
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
    <a-modal v-model:visible="detailVisible" title="资源详情" :width="'min(640px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID" :span="2">{{ currentItem.id }}</a-descriptions-item>
          <a-descriptions-item label="资源名称">{{ currentItem.name }}</a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="resTypeColor[currentItem.res_type] || 'gray'" size="small">{{ resTypeLabel[currentItem.res_type] || currentItem.res_type }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="型号/规格">{{ currentItem.model || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="可见级别">{{ currentItem.visibility_level || '-' }}</a-descriptions-item>
          <a-descriptions-item label="费用">¥{{ formatPriceFen(currentItem.price_fen) }}</a-descriptions-item>
          <a-descriptions-item label="预约信息">{{ currentItem.booking_info || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusColor[currentItem.status] || 'gray'" size="small">{{ statusLabel[currentItem.status] || currentItem.status }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="预约方式" :span="2">{{ currentItem.booking_info || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑资源' : '新增资源'" :width="'min(560px, 94vw)'" :mask-closable="false" @before-cancel="beforeCancel" destroy-on-close>
      <a-form :model="form" layout="vertical">
        <a-form-item label="资源名称" required><a-input v-model="form.name" :aria-required="true" placeholder="输入名称" style="width: 100%" /></a-form-item>
        <a-form-item label="资源类型" required>
          <a-select v-model="form.res_type" :aria-required="true" style="width: 100%">
            <a-option value="drone">无人机</a-option>
            <a-option value="airport">机场</a-option>
            <a-option value="test_site">试飞场地</a-option>
            <a-option value="flying_field">试飞场地</a-option>
            <a-option value="test_base">测试基地</a-option>
            <a-option value="other">其他</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="型号规格"><a-input v-model="form.model" style="width: 100%" /></a-form-item>
        <a-form-item label="规格参数"><a-input v-model="form.specs" style="width: 100%" /></a-form-item>
        <a-form-item label="地区"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="费用(元)">
          <a-input-number v-model="form.priceYuan" :min="0" :hide-button="true" style="width: 100%" placeholder="单位：元" />
        </a-form-item>
        <a-form-item label="可见级别">
          <a-select v-model="form.visibility_level" style="width: 100%">
            <a-option value="public">公开（政府访客可见）</a-option>
            <a-option value="member">会员可见</a-option>
            <a-option value="partner">副会长单位可见</a-option>
            <a-option value="admin">仅协会管理员</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="available">可用</a-option>
            <a-option value="in_use">使用中</a-option>
            <a-option value="maintenance">维护中</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="预约方式"><a-input v-model="form.booking_info" placeholder="如：工作日 9-18 点，需提前 3 天" style="width: 100%" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="cancelForm">取消</a-button>
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
const api = useAdminApi('industry-resources')

const resTypeLabel = { drone: '无人机', airport: '机场', test_site: '试飞场地', flying_field: '试飞场地', test_base: '测试基地', other: '其他' }
const resTypeColor = { drone: 'green', airport: 'orangered', test_site: 'orangered', flying_field: 'gray', test_base: 'arcoblue', other: 'arcoblue' }
const statusLabel = { available: '可用', in_use: '使用中', maintenance: '维护中' }
const visLabel = { public: '公开', member: '会员', partner: '副会长单位', admin: '仅管理员' }
const visColor = { public: 'gray', member: 'arcoblue', partner: 'orangered', admin: 'red' }
const statusColor = { available: 'green', in_use: 'orangered', maintenance: 'red' }

const formatPriceFen = (fen) => {
  if (fen == null) return '-'
  const yuan = fen / 100
  return yuan.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// 批量动作：设为可用 / 设为维护中——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'available', label: '设为可用', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'available' }) },
  { key: 'maintenance', label: '设为维护中', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'maintenance' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索资源名称', width: 220 },
  { key: 'res_type', label: '资源类型', type: 'select', width: 160, options: [
    { value: '', label: '全部' },
    { value: 'drone', label: '无人机' },
    { value: 'airport', label: '机场' },
    { value: 'test_site', label: '试飞场地' },
    { value: 'test_base', label: '测试基地' },
    { value: 'other', label: '其他' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '资源名称', dataIndex: 'name', minWidth: 180 },
  { title: '类型', dataIndex: 'res_type', slotName: 'resType', width: 120 },
  { title: '型号/规格', dataIndex: 'model', width: 160 },
  { title: '地区', dataIndex: 'location', width: 140 },
  { title: '费用', dataIndex: 'price_fen', slotName: 'price', width: 120, align: 'right' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 110 },
  { title: '可见级别', dataIndex: 'visibility_level', slotName: 'vis', width: 130 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

// ---- Add/Edit form ----
const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', res_type: '', model: '', specs: '', location: '', priceYuan: null, status: 'available', booking_info: '', visibility_level: 'public' })
const resetForm = () => Object.assign(form, { id: '', name: '', res_type: '', model: '', specs: '', location: '', priceYuan: null, status: 'available', booking_info: '', visibility_level: 'public' })
// 未保存守卫：Esc/遮罩/X/取消 关闭前若有改动则确认，避免输入全丢
let formSnapshot = ''
const confirmDiscard = (onOk) => {
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk,
  })
}
// Arco 2.x：@before-cancel 为同步守卫（返回 false 阻止关闭），Esc/X/遮罩均走此路径
const beforeCancel = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  confirmDiscard(() => { formVisible.value = false })
  return false
}
// footer 取消按钮也走守卫
const cancelForm = () => {
  if (JSON.stringify(form) === formSnapshot) { formVisible.value = false; return }
  confirmDiscard(() => { formVisible.value = false })
}

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    // 显式映射可编辑字段：不整行回传，避免把只读字段（id/created_at 等）带回提交
    Object.assign(form, {
      id: row.id, name: row.name || '', res_type: row.res_type || '',
      model: row.model || '', specs: row.specs || '', location: row.location || '',
      status: row.status || 'available', booking_info: row.booking_info || '',
      visibility_level: row.visibility_level || 'public',
      priceYuan: row.price_fen != null ? Math.round(row.price_fen) / 100 : null,
    })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.name) { Message.warning('请输入资源名称'); return }
  formLoading.value = true
  try {
    // 只提交可编辑字段；费用为空传 null，不伪造 0
    const payload = {
      name: form.name,
      res_type: form.res_type,
      model: form.model || '',
      specs: form.specs || '',
      location: form.location || '',
      booking_info: form.booking_info || '',
      visibility_level: form.visibility_level || 'public',
      status: form.status,
      price_fen: (form.priceYuan == null || form.priceYuan === '') ? null : Math.round(form.priceYuan * 100),
    }
    if (formEdit.value) {
      await api.update(form.id, payload); Message.success('更新成功')
    } else {
      await api.create(payload); Message.success('创建成功')
    }
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleDelete = (row) => {
  Modal.confirm({
    title: '删除资源',
    content: `确定删除「${row.name}」吗？删除后不可恢复`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); crudRef.value?.reload() }
      catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.cell-amount { color: #E96012; font-weight: 500; }
</style>
