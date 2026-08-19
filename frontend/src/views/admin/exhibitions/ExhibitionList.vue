<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="exhibitions"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增展会"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #startDate="{ record }">
        <span class="time-text">{{ formatDate(record.start_date) }}</span>
      </template>
      <template #endDate="{ record }">
        <span class="time-text">{{ formatDate(record.end_date) }}</span>
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
        <a-empty description="暂无展会数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="展会详情" :width="'min(600px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="展会名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束日期">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="展位数">{{ currentItem.booth_count || 0 }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="组织方" :span="2">{{ currentItem.organizer || '-' }}</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="展位费" :span="2">{{ currentItem.booth_price_fen ? '¥' + (currentItem.booth_price_fen / 100).toLocaleString() : '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑展会' : '新增展会'" :width="'min(560px, 94vw)'" :mask-closable="false" :unmount-on-close="true" :on-before-cancel="guardClose">
      <a-form :model="form" layout="vertical">
        <a-form-item label="展会名称" required><a-input v-model="form.title" style="width: 100%" :aria-required="true" /></a-form-item>
        <a-form-item label="展位数"><a-input-number v-model="form.booth_count" :min="0" hide-button style="width: 100%" /></a-form-item>
        <a-form-item label="地点"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="分类"><a-input v-model="form.category" placeholder="如：行业展会" maxlength="50" style="width: 100%" /></a-form-item>
        <a-form-item label="组织方"><a-input v-model="form.organizer" style="width: 100%" /></a-form-item>
        <a-form-item label="封面图URL"><a-input v-model="form.cover_url" placeholder="展会封面图片地址" style="width: 100%" /></a-form-item>
        <a-form-item label="开始日期"><a-date-picker v-model="form.start_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
        <a-form-item label="结束日期"><a-date-picker v-model="form.end_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="recruiting">招募中</a-option>
            <a-option value="underway">进行中</a-option>
            <a-option value="ended">已结束</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="展位费(元)">
          <a-input-number v-model="form.boothPriceYuan" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
        </a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="2" style="width: 100%" /></a-form-item>
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
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('exhibitions')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}

const statusTag = (s) => ({ draft: 'gray', recruiting: 'orange', underway: 'green', ended: 'gray' }[s] || 'gray')
const statusLabel = { draft: '草稿', recruiting: '招募中', underway: '进行中', ended: '已结束' }

// 批量动作：设为招募中 / 标记结束——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'recruit', label: '设为招募中', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'recruiting' }) },
  { key: 'end', label: '标记结束', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'ended' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索展会名称...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部状态' },
    { value: 'draft', label: '草稿' },
    { value: 'recruiting', label: '招募中' },
    { value: 'underway', label: '进行中' },
    { value: 'ended', label: '已结束' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '展会名称', dataIndex: 'title', slotName: 'title', minWidth: 180 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '开始日期', dataIndex: 'start_date', slotName: 'startDate', width: 120 },
  { title: '结束日期', dataIndex: 'end_date', slotName: 'endDate', width: 120 },
  { title: '展位数', dataIndex: 'booth_count', width: 80, align: 'center' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', category: '', location: '', cover_url: '', start_date: '', end_date: '', booth_count: 0, organizer: '', boothPriceYuan: null, status: 'draft', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', category: '', location: '', cover_url: '', start_date: '', end_date: '', booth_count: 0, organizer: '', boothPriceYuan: null, status: 'draft', description: '' })

// 未保存守卫：formSnapshot 快照比对 + Modal.confirm；
// X/遮罩/Esc 走 on-before-cancel（onBeforeCancel 返回 false 阻止关闭），footer 取消按钮也走守卫
let formSnapshot = ''
const takeSnapshot = () => { formSnapshot = JSON.stringify(form) }
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
const handleCancel = () => {
  if (guardClose()) formVisible.value = false
}

const openForm = (r) => {
  resetForm()
  if (r) {
    formEdit.value = true
    // 显式映射可写字段，避免只读/统计字段混入表单后被全量回传；
    // booth_price_fen(分) → boothPriceYuan(元) 回显
    Object.assign(form, {
      id: r.id,
      title: r.title || '',
      category: r.category || '',
      location: r.location || '',
      cover_url: r.cover_url || '',
      start_date: r.start_date || '',
      end_date: r.end_date || '',
      booth_count: r.booth_count ?? 0,
      organizer: r.organizer || '',
      status: r.status || 'draft',
      description: r.description || '',
      boothPriceYuan: r.booth_price_fen ? Math.round(r.booth_price_fen / 100 * 100) / 100 : null,
    })
  } else {
    formEdit.value = false
  }
  takeSnapshot()
  formVisible.value = true
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入展会名称'); return }
  formLoading.value = true
  try {
    // 白名单 payload：只回传可写字段；
    // boothPriceYuan(元) → booth_price_fen(分)，清空（null/空）时提交 null（"未设置"），不写 0 分
    const p = {
      title: form.title,
      category: form.category,
      location: form.location,
      cover_url: form.cover_url,
      start_date: form.start_date,
      end_date: form.end_date,
      booth_count: form.booth_count,
      organizer: form.organizer,
      status: form.status,
      description: form.description,
      booth_price_fen: form.boothPriceYuan == null || form.boothPriceYuan === '' ? null : Math.round(Number(form.boothPriceYuan) * 100),
    }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    takeSnapshot()
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (r) => {
  Modal.confirm({
    title: '提示',
    content: `确定删除展会「${r.title}」吗？删除后不可恢复`,
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
